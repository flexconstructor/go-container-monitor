package container_monitor

import (
	"log"
	"net"
)

// System monitor UNIX socket client.
type MonitorClient struct {
	socket_file string      // UNIX socket file name.
	SystemInfo  *SystemInfo // Current system info value object
	// instance.
	close_chan   chan bool          // Socket close channel.
	info_factory *SystemInfoFactory // System information factory.
}

// Returns new system monitor client instance.
func NewMonitorClient(socket_url string, systemInfo *SystemInfo) *MonitorClient {
	return &MonitorClient{
		socket_file:  socket_url,
		SystemInfo:   systemInfo,
		close_chan:   make(chan bool),
		info_factory: NewSystemInfoFactory(),
	}
}

// Run the socket connection.
func (c *MonitorClient) Run() {
	connection, err := net.Dial("unix", c.socket_file)
	if err != nil {
		log.Printf("Can not connect %v", err)
		return
	}
	defer connection.Close()
	go c.readFromSocket(connection)
	for {
		select {
		case <-c.close_chan:
			return
		}
	}
}

// Listens socket and read data from.
func (c *MonitorClient) readFromSocket(connection net.Conn) {
	buf := make([]byte, 3072)
	for {
		n, err := connection.Read(buf[:])
		if err != nil {
			c.Stop()
			return
		}
		system_info, err := c.info_factory.UnMarshalInfo(buf[0:n])
		if err != nil {
			log.Println("Can not unmarshal system info!")
		}
		c.SystemInfo = system_info
	}
}

// Write to close channel for close connection and listener.
func (c *MonitorClient) Stop() {
	c.close_chan <- true
}
