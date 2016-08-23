package container_monitor

import (
	"log"
	"time"
	"gopkg.in/redis.v4"
)

// Container monitor struct. This monitor listens unix socket
// and writes to socket system info from container where this running.
type ContainerMonitor struct {
	close_channel chan bool          // Channel for close signal.
	info_factory  *SystemInfoFactory // System info factory.
	client *redis.Client
}

func NewContainerMonitor(redis_url string) *ContainerMonitor {
	return &ContainerMonitor{
		close_channel: make(chan bool),
		info_factory:  NewSystemInfoFactory(),
		client: redis.NewClient(&redis.Options{
			Addr:redis_url,
			Password:"",
			DB:0,
		}),
	}
}

// Runs the container monitor.
// Just starts listen of unix socket.
func (m *ContainerMonitor) Run() {
	pong,err:= m.client.Ping().Result()

	defer m.client.Close()
	if(err != nil){
		log.Printf("redis error: %s",err.Error())
		return
	}
	log.Printf("pong: %v",pong)
	for{
		select{
		case <-time.After(2 * time.Second):
		m.writeSystemInfo()
		case <- m.close_channel:
		return
		}
	}
}


func (m *ContainerMonitor) writeSystemInfo(){
 info:= m.info_factory.GetSystemInfo();
	log.Printf(string(info))
	err:= m.client.Set("system:info",info,0).Err()
	if(err != nil){
		log.Printf("WRITE DATA ERROR: %s",err.Error())
		m.Stop()
	}
}

func (m *ContainerMonitor) Stop() {
	m.close_channel <- true
}
