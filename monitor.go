package container_monitor

import (
	"log"
	"net"
	"time"
)

// Container monitor struct. This monitor listens unix socket
// and writes to socket system info from container where this running.
type ContainerMonitor struct {
	socket_file   string             // UNIX socket file path.
	close_channel chan bool          // Channel for close signal.
	info_factory  *SystemInfoFactory // System info factory.
}

func NewContainerMonitor(socket_file_name string) *ContainerMonitor {
	return &ContainerMonitor{
		socket_file:   socket_file_name,
		close_channel: make(chan bool),
		info_factory:  NewSystemInfoFactory(),
	}
}

// Runs the container monitor.
// Just starts listen of unix socket.
func (m *ContainerMonitor) Run() {
	connection, err := net.Listen("unix", m.socket_file)
	if err != nil {
		log.Printf("listen error %v", err)
		return
	}
	defer connection.Close()
	go m.listenConnection(connection)

	if err != nil {
		log.Printf("Dial error %v", err)
		return
	}
	for {
		select {
		case <-m.close_channel:
			return
		}
	}
}

// Listens established connection and starts timer for send information for
// client.
func (m *ContainerMonitor) listenConnection(listener net.Listener) {
	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Printf("can not acept connection %v", err)
			m.Stop()
			return
		}
		for {
			select {
			case <-m.close_channel:
				connection.Close()
				return
			case <-time.After(2 * time.Second):
				_, err := connection.Write(m.info_factory.GetSystemInfo())
				if err != nil {
					log.Printf("Can not write message to socket: %v", err)
					m.Stop()
					return
				}
			}
		}
	}
}

func (m *ContainerMonitor) Stop() {
	m.close_channel <- true
}
