package gotftp

import (
	"encoding/binary"
	"fmt"
	"net"
)

const (
	maxPacketSize = 520
)

type Request struct {
	conn *net.PacketConn
	localAddr Addr
	remoteAddr Addr
	request TftpRequest
}

func ReadRequest(connections chan *net.PacketConn, requests chan *Request) {

	origBuf := make([]byte, maxPacketSize)
	maxPacketSizeError := toTftpErrorSlice(TftpError{0, "max packet size exceeded"})
	unknownOpcodeError := toTftpErrorSlice(TftpError{0, "unknown opcode"})
	parseError := toTftpErrorSlice(TftpError{0, "error parsing packet"})

	// TODO : better logging / instrumentation
	for conn := range connections {
		numBytes, addr, err := conn.ReadFrom(origBuf)
		if err != nil {
			conn.Close()
			continue
		}

		if numBytes > 516 {
			conn.WriteTo(maxPacketSizeError, addr)
			conn.Close()
			continue
		}

		buf := origBuf[:numBytes]

		opcode := binary.BigEndian.Uint16(buf[0:2])

		err := nil
		tftpRequest := nil

		switch opcode {
		case 1 || 2: ioRequest, err := parseIORequest(buf)
					 tftpRequest = TftpRequest(ioRequest)
		case 3: dataBlockRequest, err := parseDataBlock(buf)
				tftpRequest = TftpRequest(dataBlockRequest)
		case 4: ackRequest, err := parseAck(buf)
				tftpRequest = TftpRequest(ackRequest)
		case 5: errorRequest, err := parseTftpErrorSlice(buf)
				tftpRequest = TftpRequest(errorRequest)
		}

		if tftpRequest == nil {
			conn.WriteTo(unknownOpcodeError, addr)
			conn.Close()
			continue
		}

		if err != nil {
			conn.WriteTo(parseError, addr)
			conn.Close()
			continue
		}

		requests <- Request{conn.LocalAddr(), addr, tftpRequest}
	}

}

func WriteResponse(requests chan* Request, connections chan *net.PacketConn) {

	origBuf := make([]byte, maxPacketSize)

	// TODO : better logging / instrumentation
	for request := range requests {

		switch request.request.getType() {
		case 1:
		case 2:
		case 3:
		case 4:
		case 5:
		}

		connections <- request.conn
	}


}

//UDPServer starts a UDP server on the specified ip and port and passes received
//connections to the connections channel. If the UDP server receives more connections
//than it can handle it sends an error message to the client and closes the connection.
func UDPServer(ip string, port int, run *bool, connections chan *net.PacketConn) {
	fmt.Println("starting UDP Server on ", ip, port, *run)

	maxConnectionsError := toTftpErrorSlice(TftpError{0, "max conns reached"})

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
			go func() {
				// send error message
				conn.Write(maxConnectionsError)
				conn.Close()
			}()
		}

	}
}

func main() {
	run := true

	maxConnections := 1000
	connections := make(chan *net.PacketConn, maxConnections)

	UDPServer("127.0.0.1", 8000, &run, connections)
}

