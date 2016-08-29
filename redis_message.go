package container_monitor

import (
	"encoding/json"
	"log"
)

const (
	START_COMMAND       = "start_test"         // Stress test started.
	STOP_COMMAND        = "stop_test"          // Stress test stopped.
	STRESS_TEST_CHANNEL = "stress_test_client" // Redis channel name.
)

// Redis pub/sub message
type redisMessage struct {
	Command string
	TestID  string
}

// Returns new instance of Redis pub/sub message.
//
// params: command string   Kind of command.
//         test_id string   Stress test ID.
func newRedisMessage(command string, test_id string) *redisMessage {
	return &redisMessage{
		Command: command,
		TestID:  test_id,
	}
}

// Encodes Redis pub/sub message instance to JSON string.
//
// param: message *redisMessage   Instance of Redis pub/sub message.
// return JSON string or error instance.
func marshalRedisMessage(message *redisMessage) (string, error) {
	message_string, err := json.Marshal(message)
	if err != nil {
		log.Printf("Can not marshal redis message %s", err.Error())
		return err.Error(), err
	}
	return string(message_string), nil
}

// Decodes Redis pub/sub message JSON string to instance of *redisMessage
//
// param: message_string string   JSON message string.
// return: Instance of *redisMessage or error instance.
func unmarshalRedisMessage(message_string string) (*redisMessage, error) {
	mess := &redisMessage{}
	err := json.Unmarshal([]byte(message_string), mess)
	if err != nil {
		log.Printf("Can not unmarshal redis message %s", err.Error())
		return nil, err
	}
	return mess, nil
}
