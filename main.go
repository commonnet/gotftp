package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	maxIOrequestBufSize = 1024
	maxDataBufSize = 512
	maxDataBlockSize = 520
)

type Config interface {
	getFSRoot() string
	getFSTmp() string
	getTftpIP() string
	getTftpPort() int
}

type TftpConfig struct {
	fsroot string
	fstmp string
	ip string
	port int
}

func (t TftpConfig) getFSRoot() string {
	return t.fsroot
}

func (t TftpConfig) getFSTmp() string {
	return t.fstmp
}

func (t TftpConfig) getTftpIP() string {
	return t.ip
}

func (t TftpConfig) getTftpPort() int {
	return t.port
}

type Connection interface {
	WriteTo([]byte) (numBytes int, err error)
	ReadFrom([]byte) (numBytes int, err error)
}

type UDPConnection struct {
	addr net.Addr
	conn *net.UDPConn
	writeTimeout uint64
	readTimeout uint64
}

func (u *UDPConnection) WriteTo(buf []byte) (numBytes int, err error) {
	u.conn.SetWriteDeadline(time.Now().Add(time.Duration(u.writeTimeout)))
	numBytes, err = u.conn.WriteTo(buf, u.addr)
	return numBytes, err
}

func (u *UDPConnection) ReadFrom(buf []byte) (numBytes int, err error) {
	u.conn.SetReadDeadline(time.Now().Add(time.Duration(u.readTimeout)))
	numBytes, _, err = u.conn.ReadFrom(buf)
	return numBytes, err
}

type Session struct {
	connection *UDPConnection
	ioRequest IORequest
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
	if err != nil {
		return err
	}

	defer file.Close()

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

	fmt.Println("read complete! ", readRequest.filename)
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

	final := false

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

		if final {
			break
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

		if dataBlock.isFinal() {
			final = true
			file.Close()

			oldFilename := fmt.Sprintf("%s%s", config.getFSTmp(), writeRequest.filename)
			newFilename := fmt.Sprintf("%s%s", config.getFSRoot(), writeRequest.filename)
			fmt.Printf("renaming %s to %s", oldFilename, newFilename)
			os.Rename(oldFilename, newFilename)
		}

	}

	fmt.Println("transfer complete! ", writeRequest.filename)
	return nil
}

//HandleConnection dequeues a session from the session channel and processes
//the IORequest corresponding to the session.
func HandleConnection(sessions chan* Session, config Config, run *bool) {

	error := make([]byte, 48)
	errorLength := toTftpErrorSlice(TftpError{0, "illegal request"}, error)


	for session := range sessions {

		if (session.ioRequest.isWrite) {
			err := processWriteRequest(session.connection, session.ioRequest, config)
			if err != nil {
				fmt.Println(err)
				session.connection.WriteTo(error[:errorLength])
			}
		} else {
			err := processReadRequest(session.connection, session.ioRequest, config)
			if err != nil {
				fmt.Println(err)
				session.connection.WriteTo(error[:errorLength])
			}
		}

		if (!*run) {
			break
		}
	}

}

//UDPServer starts a UDP server on the specified ip and port and passes received
//connections to the connections channel. If the UDP server receives more connections
//than it can handle it sends an error message to the client and closes the connection.
func UDPServer(sessions chan* Session, config Config, run *bool) {
	fmt.Println("starting UDP Server on ", config.getTftpIP(), config.getTftpPort(), *run)

	ioRequestBuf := make([]byte, maxIOrequestBufSize)

	error := make([]byte, 48)
	errorLength := toTftpErrorSlice(TftpError{0, "illegal request"}, error)

	addr := net.UDPAddr{
		Port: config.getTftpPort(),
		IP: net.ParseIP(config.getTftpIP()),
	}

	for (*run) {
		conn, err := net.ListenUDP("udp", &addr)
		if err != nil {
			fmt.Println("error occurred while listening for udp connections", err)
			conn.Close()

			continue
		}

		for {
			numBytes, addr, err := conn.ReadFrom(ioRequestBuf)
			if err != nil {
				fmt.Println("error while reading from the tftp listener port")
				conn.Close()
				break
			}

			addrServ := net.UDPAddr{
				Port: 0,
				IP: net.ParseIP(config.getTftpIP()),
			}

			connServ, err := net.ListenUDP("udp", &addrServ)
			if err != nil {
				fmt.Println("error occurred while listening on udp child socket for ", addr)
				connServ.Close()
				conn.Close()

				break
			}

			connection := &UDPConnection{addr, connServ, 8e9, 8e9}

			ioRequest, err := parseIORequest(ioRequestBuf[:numBytes])
			if err != nil {
				fmt.Println(err, addr, ioRequest.filename)
				connServ.WriteTo(error[:errorLength], addr)
				connServ.Close()
				break
			}

			session := &Session{connection, ioRequest}

			select {
			case sessions <- session:
				fmt.Printf("Processing session for remote: %s, local: %s, filename: %s, write: %v, mode: %s\n", addr, connServ.LocalAddr(), ioRequest.filename, ioRequest.isWrite, ioRequest.mode)
			default:
				fmt.Println("Rejecting session for remote: %s, local: %s, filename: %s, write: %v, mode: %s\n", addr, connServ.LocalAddr(), ioRequest.filename, ioRequest.isWrite, ioRequest.mode)
				go func() {
					connServ.WriteTo(error[:errorLength], addr)
					connServ.Close()
				}()
			}

			if !(*run) {
				break
			}
		}
	}
}

func usage(val int) {
	fmt.Println("./main <file system root> <file system tmp> <interface ip> <port>")
	os.Exit(val)
}

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return false, err
}

func main() {
	if len(os.Args) < 4 {
		usage(1)
	}

	run := true
	port, err := strconv.Atoi(os.Args[4])
	if err != nil {
		panic(err)
	}

	config := TftpConfig{fsroot:os.Args[1], fstmp:os.Args[2], ip:os.Args[3], port:port}

	dirExists, _ := exists(config.getFSRoot())
	if !dirExists {
		err := os.Mkdir(config.getFSRoot(), os.ModeDir | 0777)
		if err != nil {
			panic(err)
		}
	}

	dirExists, _ = exists(config.getFSTmp())
	if !dirExists {
		err := os.Mkdir(config.getFSTmp(), os.ModeDir | 0777)
		if err != nil {
			panic(err)
		}
	}

	sessions := make(chan *Session, 100)

	for i:=0; i<10; i++ {
		go HandleConnection(sessions, config, &run)
	}
	UDPServer(sessions, config, &run)
}
