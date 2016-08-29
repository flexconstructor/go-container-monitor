package container_monitor

import (
	"gopkg.in/redis.v4"
	"log"
)

type RedisListener struct {
	Client  *redis.Client     // Redis client
	monitor *ContainerMonitor // Container monitor.
}

// Returns new instance of Redis listener.
// props: r_url    string   Redis server URL.
//        password string   Redis server password.
//        db       int      Redis server Data Base ID.
func NewRedisListener(r_url string, password string, db int) *RedisListener {
	return &RedisListener{
		Client: redis.NewClient(&redis.Options{
			Addr:     r_url,
			Password: password,
			DB:       db,
		}),
	}
}

// Listens Redis pub/sub channel.
func (l *RedisListener) Listen() {
	defer l.Client.Close()
	pubsub, err := l.Client.Subscribe(STRESS_TEST_CHANNEL)
	if err != nil {
		log.Printf("CLIENT SUBSCRIBE ERROR: %s", err.Error())
		return
	}
	defer pubsub.Unsubscribe(STRESS_TEST_CHANNEL)
	for {
		err := l.ping()
		if err != nil {
			log.Printf("CLIENT PING ERROR: %s", err.Error())
			return
		}
		mess, err := pubsub.ReceiveMessage()
		if err == nil {
			l.readRedisMessage(mess)
		}
	}
}

// Calls redis pub/sub channel.
func (l *RedisListener) Call(test_id string, command string) {
	mess := newRedisMessage(command, test_id)
	message_string, err := marshalRedisMessage(mess)
	if err != nil {
		log.Printf("can not create start redis message %s", err.Error())
		return
	}
	l.Client.Publish(STRESS_TEST_CHANNEL, message_string)
}

// Closes listener
func (l *RedisListener) Close() {
	l.Client.Close()
}

// Reads Redis pub/sub messages.
func (l *RedisListener) readRedisMessage(mess *redis.Message) {
	message, err := unmarshalRedisMessage(mess.Payload)
	if err != nil {
		log.Printf("Can not unmarshall redis message: %s", err.Error())
		return
	}
	if message.Command == START_COMMAND {
		l.startTest(message.TestID)
	} else {
		l.stopTest(message.TestID)
	}
}

// Starts gathering information about the container system.
func (l *RedisListener) startTest(test_id string) {
	if l.monitor != nil {
		log.Println("ERROR: last test not finiched!")
		return
	}
	l.monitor = newContainerMonitor(l.Client, test_id)
	go l.monitor.Run()
}

// Stops gathering information about the container system.
func (l *RedisListener) stopTest(test_id string) {
	if l.monitor == nil {
		log.Println("ERROR: test not started!")
		return
	}
	l.monitor.Stop()
	l.monitor = nil
}

// Pings redis pub/sub channel.
func (l *RedisListener) ping() error {
	err := l.Client.Ping().Err()
	if err != nil {
		return err
	}
	return nil
}
