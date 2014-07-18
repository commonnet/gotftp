package gotftp

/*
import (
	"fmt"
	"net"
	"os"
	"time"
)
/*
const (
	maxUDPPacketSize = 65507
	maxPacketSize = 520
	maxRequestSize = 516
	maxDataBufSize = 512
	ackSize = 4
	readTimeout = 8e9
	writeTimeout = 8e9
)

const (
	ioRequestState = iota
	readState = iota
	writeState = iota
)

const (
	fsroot = "/home/ubuntu/root/"
	fstmp = "/home/ubuntu/tmp/"
)

func ProcessReadRequest(conn *net.PacketConn, addr net.Addr, ioRequest IORequest) (error) {
	dataBlockNumber := 0
	dataBuf := make([]byte, maxDataBufSize)
	ackBuf := make([]byte, ackSize)

	file, err := os.Open(fmt.Sprintf("%s%s", fsroot, ioRequest.filename))
	if err != nil {
		sendError(conn, addr, fileNotFoundError, err.String())
		conn.Close()
	}

	retries := 3

	for {
		numBytes, err := file.ReadAt(dataBuf, int64(dataBlockNumber))
		if err != nil {
			sendError(conn, addr, ioerror, err.String())
			conn.Close()
		}

		dataBlock := DataBlock{dataBlockNumber, dataBuf[:numBytes]}
		dataBlockResponseSlice := dataBlockToSlice(dataBlock)

		success := false
		for i:=0; i<3 && !success; i:=i+1 {
			conn.SetWriteDeadline(time.Now().Add(time.Duration(writeTimeout)))
			numBytes, err := conn.WriteTo(dataBlockResponseSlice, addr)
			if numBytes == len(dataBlockResponseSlice) && err == nil {
				success := true
			}
		}

		if !success {
			sendError(conn, addr, ioerror, err.String())
			conn.Close()
		}

		conn.SetReadDeadline(time.Now().Add(time.Duration(readTimeout)))
		numBytes, _, err := conn.ReadFrom(ackBuf)
		if err != nil {
			sendError(conn, addr, ioerror, err.String())
			conn.Close()
		}

		ack, err := parseAck(ackBuf)
		if err != nil {

		}
	}
}

//ProcessIORequest reads a read/write tftp request from the connection
func ProcessIORequest(conn *net.PacketConn, buf []byte) (IORequest, net.Addr) {

	conn.SetReadDeadline(time.Now().Add(time.Duration(readTimeout)))
	numBytes, addr, err := conn.ReadFrom(buf)
	if err != nil {
		sendError(conn, addr, parseError, err.String())
		conn.Close()
	}

	ioRequest, err := parseIORequest(buf[:numBytes])
	if err != nil {
		sendError(conn, addr, parseError, err.String())
		conn.Close()
	}

	return ioRequest, addr

}

//ProcessRequest is a simple state machine per connection that reads
//a request from the udp connection and sends a response.
//process request tracks a bunch of state:
// 1. Read/Write request(IORequest) that triggered the session.
// 2. The last data block number.
// 3. The last response packet.
//
//state machine.
// iorequeststate
// readstate
// writestate
func ProcessRequest(connections chan *net.PacketConn) {

	ioRequestBuf := make([]byte, maxUDPPacketSize)
	responseBuf := make([]byte, maxPacketSize)
	origBuf := make([]byte, maxPacketSize)

	dataBlockNumber := 0
	ioRequest := nil
	addr := nil

	// TODO : better logging / instrumentation
	for conn := range connections {
		currentState := ioRequestState

		for {
			switch currentState {
			case ioRequestState:
				ioRequest, addr := ProcessIORequest(conn, ioRequestBuf)
				if ioRequest.isWrite {
					currentState := writeState
				} else {
					currentState := readState
				}
			case readState:
				err := ProcessReadRequest(conn, addr, ioRequest)
			case writeState:
				err := ProcessWriteRequest(conn, addr, ioRequest)
			default:
				// don't know how I got here, send error and close connection.
			}

			if err != nil {
				break
			}
		}
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


/*
for {
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

			switch opcode {
			case 1 || 2: ioRequest, err := parseIORequest(buf)
						 processIORequest(ioRequest, conn, addr)
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

				maxPacketSizeError := toTftpErrorSlice(TftpError{0, "max packet size exceeded"})
	unknownOpcodeError := toTftpErrorSlice(TftpError{0, "unknown opcode"})
	parseError := toTftpErrorSlice(TftpError{0, "error parsing packet"})
 */

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

const (
	maxDataBufSize = 512
	maxDataBlockSize = 520
)

type Config interface {
	getFSRoot() string
	getFSTmp() string
}

type TftpConfig struct {
	fsroot string
	fstmp string
}

func (t TftpConfig) getFSRoot() string {
	return t.fsroot
}

func (t TftpConfig) getFSTmp() string {
	return t.fstmp
}

type Connection interface {
	WriteTo([]byte) (numBytes int, err error)
	ReadFrom([]byte) (numBytes int, err error)
}

type UDPConnection struct {
	addr net.Addr
	conn net.UDPConn
	writeTimeout int
	readTimeout int
}

func (u *UDPConnection) writeTo(buf []byte) (numBytes int, err error) {
	u.conn.SetWriteDeadline(time.Now().Add(time.Duration(u.writeTimeout)))
	numBytes, err = u.conn.WriteTo(buf, u.addr)
	return numBytes, err
}

func (u *UDPConnection) readFrom(buf []byte) (numBytes int, err error) {
	u.conn.SetReadDeadline(time.Now().Add(time.Duration(u.readTimeout)))
	numBytes, _, err = u.conn.ReadFrom(buf)
	return numBytes, err
}

/*
Read State Machine:

1. Incoming Connection.
2. Read Request contains file name / mode.
3. Send DataBlockNumber i.
4. Receive Ack DataBlockNumber i. On timeout re-send DataBlockNumber i. if retries > x, send error, close conn.
5. If remaining data, Goto Step 3, else exit.
*/
func processReadRequest(conn Connection, readRequest IORequest, config Config) error {

	dataBlockNumber := uint16(1)

	dataBuf := make([]byte, maxDataBufSize)
	dataBlockBuf := make([]byte, maxDataBlockSize)

	ackBuf := make([]byte, 4)


	file, err := os.Open(fmt.Sprintf("%s%s", config.getFSRoot(), readRequest.filename))
	defer file.Close()

	if err != nil {
		return err
	}

	for {
		numBytes, err := file.ReadAt(dataBuf, int64((dataBlockNumber-1) * maxDataBufSize))
		if err != nil && err != io.EOF {
			return err
		}

		dataBlock := DataBlock{dataBlockNumber, dataBuf[:numBytes]}
		numBytes = dataBlockToSlice(dataBlock, dataBlockBuf)

		_, err = conn.WriteTo(dataBlockBuf[:numBytes])

		if err != nil {
			return err
		}

		_, err = conn.ReadFrom(ackBuf)

		if err != nil {
			return err
		}

		ack, err := parseAck(ackBuf)
		if err != nil {
			return err
		}

		if ack.blockNumber != uint16(dataBlockNumber) {
			return errors.New(fmt.Sprintf("expected datablock %d got %d", dataBlockNumber, ack.blockNumber))
		}

		dataBlockNumber = dataBlockNumber+1

		if numBytes < 516 {
			break
		}
	}
	return nil
}

/*
Write State Machine:

1. Incoming Connection.
2. Write Request contains file name / mode.
3. Send Ack DataBlockNumber i.
4. Receive DataBlockNumber i+1. On timeout re-send Ack DataBlockNumber i. if retries > x, send error, close conn.
5. If datablock length < 512, Goto step 3 and exit, else Goto Step 3 and repeat.
*/
func processWriteRequest(conn Connection, writeRequest IORequest, config Config) error {

	dataBlockNumber := uint16(0)

	dataBlockBuf := make([]byte, maxDataBlockSize)

	ackBuf := make([]byte, 4)


	file, err := os.Create(fmt.Sprintf("%s%s", config.getFSTmp(), writeRequest.filename))
	defer file.Close()

	if err != nil {
		return err
	}

	for {
		ack := Ack{dataBlockNumber}
		ackToSlice(ack, ackBuf)

		numBytes, err := conn.WriteTo(ackBuf)

		dataBlockNumber = dataBlockNumber+1

		if err != nil {
			return err
		}

		if numBytes != 4 {
			return errors.New("unable to write complete ack response")
		}

		numBytes, err = conn.ReadFrom(dataBlockBuf)
		if err != nil {
			return err
		}

		dataBlock, err := parseDataBlock(dataBlockBuf[:numBytes])
		if err != nil {
			return err
		}

		if dataBlock.blockNumber != uint16(dataBlockNumber) {
			return errors.New(fmt.Sprintf("expected datablock %d got %d", dataBlockNumber, dataBlock.blockNumber))
		}

		numBytes, err = file.Write(dataBlock.data)
		if err != nil {
			return err
		}

		if len(dataBlock.data) < maxDataBufSize {
			file.Close()
			os.Rename(fmt.Sprintf("%s%s", config.getFSTmp(), writeRequest.filename), fmt.Sprintf("%s%s", config.getFSRoot(), writeRequest.filename))
			break
		}
	}
	return nil
}

