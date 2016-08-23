package container_monitor

import (
	"gopkg.in/redis.v4"
	"log"
)

// System monitor UNIX socket client.
type MonitorClient struct {
	SystemInfo   *SystemInfo // Current system info value object
	client       *redis.Client
	info_factory *SystemInfoFactory // System information factory.
}

// Returns new system monitor client instance.
func NewMonitorClient(redis_url string, systemInfo *SystemInfo) *MonitorClient {
	return &MonitorClient{
		SystemInfo:   systemInfo,
		info_factory: NewSystemInfoFactory(),
		client: redis.NewClient(&redis.Options{
			Addr:     redis_url,
			Password: "",
			DB:       0,
		}),
	}
}

// Run the socket connection.
func (c *MonitorClient) Update() {

	pong, err := c.client.Ping().Result()
	if err != nil {
		log.Printf("redis error: %s", err.Error())
		return
	}
	log.Printf("pong %v", pong)

	result := c.client.Get("system:info")
	if result.Err() != nil {
		log.Printf("READ ERR: %s", result.Err().Error())
		return
	}
	bytes, err := result.Bytes()
	if err != nil {
		log.Printf("parse result error: %s", err.Error())
	}
	log.Printf("RESULT: %s", result.String())
	system_info, err := c.info_factory.UnMarshalInfo(bytes)
	if err != nil {
		log.Println("Can not unmarshal system info!")
	}
	c.SystemInfo = system_info
}

// Close redis connection.
func (c *MonitorClient) Stop() {
	c.client.Close()
}
