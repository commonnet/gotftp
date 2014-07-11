package gotftp
/*
import (
	"fmt"
	"net"
)

const (
	MaxPacketSize = 520
)

func ParseRequest(connections chan *net.PacketConn) {
	// TODO : need to buffer pool, for this implementation simply create a new buffer every time you process a connection.
	// TODO : better logging / instrumentation
	for conn := range connections {
		buf := make([]byte, MaxPacketSize)

		numBytes, addr, err := conn.ReadFrom(buf)
		if err != nil {
			conn.Close()
		}

		if numBytes > 516 {
			// TODO : send error
			conn.Close()
		}

		switch buf[1] {
		case 1: // parse read-request
		case 2: // parse write-request
		case 3: // parse data-block
		case 4: // parse ack
		case 5: // parse error code
		default: // TODO : send error
		}

	}
}

//UDPServer starts a UDP server on the specified ip and port and passes received
//connections to the connections channel.
func UDPServer(ip string, port int, run *bool, connections chan *net.PacketConn) {
	fmt.Println("starting UDP Server on ", ip, port, *run)

	addr := net.UDPAddr{
		Port: port,
		IP: net.ParseIP(ip),
	}

	for (*run) {
		conn, err := net.ListenUDP("udp", &addr)

		if err != nil {
			fmt.Println("error occurred while listening for udp connections", err)
			conn.Close()
		}

		fmt.Println("received connection from ", conn.RemoteAddr().String())
		select {
		case connections <- conn:
			// can process/queue this connection
		default:
			// TODO : hit the max limit, read from the connection and send error response, for now just close the connection.
			conn.Close()
		}

	}
}

func main() {
	run := true

	maxConnections := 1000
	connections := make(chan *net.PacketConn, maxConnections)

	UDPServer("127.0.0.1", 8000, &run, connections)
}
*/
