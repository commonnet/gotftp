package main

import (
	"bytes"
	"io"
	"testing"
	"os"
	"fmt"
	"hash/crc32"
	"io/ioutil"
)

type MockConnection struct {
	file *os.File
	t *testing.T

	input []byte
	output []byte
	inputLength int
	outputLength int

	err error
	handle func(*testing.T, *os.File, []byte, []byte) (int)
}

func (m *MockConnection) WriteTo(bytes []byte) (numBytes int, err error) {
	if m.err != nil {
		return 0, m.err
	}

	copy(m.output, bytes)
	m.outputLength = len(bytes)

	return m.outputLength, nil
}

func (m *MockConnection) ReadFrom(bytes []byte) (numBytes int, err error) {
	if m.err != nil {
		return 0, m.err
	}

	numBytes = m.handle(m.t, m.file, m.output[:m.outputLength], m.input)
	copy(bytes, m.input[:numBytes])

	return numBytes, nil
}

func InitTest(config Config) {
	dirExists, _ := Exists(config.GetFSRoot())
	if !dirExists {
		err := os.Mkdir(config.GetFSRoot(), os.ModeDir | 0777)
		if err != nil {
			panic(err)
		}
	}

	dirExists, _ = Exists(config.GetFSTmp())
	if !dirExists {
		err := os.Mkdir(config.GetFSTmp(), os.ModeDir | 0777)
		if err != nil {
			panic(err)
		}
	}
}

func CloseTest(config Config) {
	err := os.RemoveAll(config.GetFSRoot())
	if err != nil {
		panic(err)
	}

	err = os.RemoveAll(config.GetFSTmp())
	if err != nil {
		panic(err)
	}
}

func TestProcessReadRequest(t *testing.T) {
	config := TftpConfig{"/tmp/fsroot/", "/tmp/fstmp/", "127.0.0.1", 8000}

	InitTest(config)
	defer CloseTest(config)

	ioRequest := IORequest{isWrite:false, filename:"test.txt", mode:"octet"}

	fname := fmt.Sprintf("%s%s", config.GetFSRoot(),ioRequest.filename)
	CreateTestFile(fname, 512*5+256)

	file, err := os.Open(fname)
	if err != nil {
		t.Error(err)
	}

	connection := &MockConnection{file, t, make([]byte, 520), make([]byte, 520), 0, 0, nil, ReadHandler}

	err = ProcessReadRequest(connection, ioRequest, config)
	if err != nil {
		t.Error(err)
	}

	file.Close()
}

func ReadHandler(t *testing.T, f *os.File, dataBlockBytes []byte, ackBytes []byte) int {

	dataBlock, err := ParseDataBlock(dataBlockBytes)
	if err != nil {
		t.Error(err)
	}

	buf := make([]byte, 512)
	numBytes, err:= f.ReadAt(buf, int64(dataBlock.blockNumber-1)*512)
	if err != nil && err != io.EOF {
		t.Error(err)
	}

	if bytes.Compare(dataBlock.data, buf[:numBytes]) != 0 {
		t.Error("file bytes doesn't match the data block returned")
	}

	ack := Ack{dataBlock.blockNumber}
	AckToSlice(ack, ackBytes)

	return 4
}

func CreateTestFile(filename string, length int) {
	file, err := os.Create(filename)
	defer file.Close()

	if err != nil {
		panic(err)
	}

	buf := make([]byte, length)
	for i:=0; i<len(buf); i=i+1 {
		buf[i] = byte(i)
	}

	file.Write(buf)
}

func TestProcessWriteRequest(t *testing.T) {
	config := TftpConfig{"/tmp/fsroot/", "/tmp/fstmp/", "127.0.0.1", 8000}

	InitTest(config)
	defer CloseTest(config)

	ioRequest := IORequest{isWrite:true, filename:"test.txt", mode:"octet"}

	fname := fmt.Sprintf("%s%s", config.GetFSTmp(), "test-expected.txt")
	CreateTestFile(fname, 512*5+256)

	file, err := os.Open(fname)
	if err != nil {
		t.Error(err)
	}

	connection := &MockConnection{file, t, make([]byte, 520), make([]byte, 520), 0, 0, nil, WriteHandler}

	err = ProcessWriteRequest(connection, ioRequest, config)
	if err != nil {
		t.Error(err)
	}

	file.Close()

	hashExpected, _ := GetHash(fname)
	hashActual, _ := GetHash(fmt.Sprintf("%s%s", config.GetFSRoot(), "test.txt"))

	if hashExpected != hashActual {
		t.Error("files mismatched while writing")
	}
}

func WriteHandler(t *testing.T, f *os.File, ackBytes []byte, dataBlockBytes []byte) int {

	ack, err := ParseAck(ackBytes)
	if err != nil {
		t.Error(err)
	}

	buf := make([]byte, 512)
	numBytes, err:= f.ReadAt(buf, int64(ack.blockNumber)*512)
	if err != nil && err != io.EOF {
		t.Error(err)
	}

	dataBlock := DataBlock{ack.blockNumber+1, buf[:numBytes]}
	numBytes = DataBlockToSlice(dataBlock, dataBlockBytes)

	return numBytes

}

func GetHash(filename string) (uint32, error) {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, err
	}
	h := crc32.NewIEEE()
	h.Write(bs)
	return h.Sum32(), nil
}
