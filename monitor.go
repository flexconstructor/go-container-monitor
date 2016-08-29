package container_monitor

import (
	"gopkg.in/redis.v4"
	"log"
	"time"
)

// Container monitor struct. This monitor listens unix socket
// and writes to socket system info from container where this running.
type ContainerMonitor struct {
	close_channel chan bool          // Channel for close signal.
	info_factory  *SystemInfoFactory // System info factory.
	testID string
}

// Returns new ContainerMonitor instance.
//
// params: client *redis.Client   Instance of Redis client.
//         test_id string         Test ID.
func newContainerMonitor(client *redis.Client, test_id string) *ContainerMonitor {
	return &ContainerMonitor{
		close_channel: make(chan bool),
		info_factory:  NewSystemInfoFactory(client),
		testID:test_id,
	}
}

// Runs the container monitor.
// Just starts listen of unix socket.
func (m *ContainerMonitor) Run() {
	for{
		select{
		case <- time.After(time.Second *2):
		m.info_factory.UpdateSystemInfo(m.testID)
		case <- m.close_channel:
		return
		}
	}
}


// Close redis connection.
func (m *ContainerMonitor) Stop() {
	log.Println("stop test: %s",m.testID)
	m.close_channel <- true
}
