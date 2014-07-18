package gotftp

import (
	"bytes"
	"errors"
	"encoding/binary"
)

const (
	_ = iota
	readOpcode uint16 = iota
	writeOpcode uint16 = iota
	dataBlockOpcode uint16 = iota
	ackOpcode uint16 = iota
	errorOpcode uint16 = iota
)

type TftpRequest interface {
	getType() uint16
}

type IORequest struct {
	isWrite bool
	filename string
	mode string
}

func (i IORequest) getType() uint16 {
	if i.isWrite {
		return writeOpcode
	}

	return readOpcode
}

func parseIORequest(byteSlice []byte) (IORequest, error) {
	if byteSlice == nil || len(byteSlice) < 4 {
		return IORequest{}, errors.New("byteSlice parameter was nil")
	}

	isWrite := false
	opcode := binary.BigEndian.Uint16(byteSlice[0:2])
	if opcode == readOpcode {
		isWrite = false
	} else if opcode == writeOpcode {
		isWrite = true
	} else {
		return IORequest{}, errors.New("Invalid opcode")
	}

	buffer := bytes.NewBuffer(byteSlice[2:len(byteSlice)])
	filenameBytes, err := buffer.ReadBytes((byte)(0))
	if err != nil {
		return IORequest{}, err
	}

	filenameBytesLength := len(filenameBytes) - 1
	if filenameBytesLength < 1 {
		return IORequest{}, errors.New("request does not contain a filename")
	}

	filename := string(filenameBytes[:filenameBytesLength])

	modeBytes, err := buffer.ReadBytes((byte)(0))
	if err != nil {
		return IORequest{}, err
	}

	modeBytesLength := len(modeBytes) - 1
	if modeBytesLength < 1 {
		return IORequest{}, errors.New("request does not contain a mode")
	}

	mode := string(modeBytes[:modeBytesLength])

	if mode != "octet" {
		return IORequest{}, errors.New("cannot support modes other than octet")
	}

	return IORequest{isWrite, filename, mode}, nil
}


type DataBlock struct {
	blockNumber uint16
	data []byte
}

func (d DataBlock) getType() uint16 {
	return dataBlockOpcode
}

func (d DataBlock) isFinal() bool {
	if len(d.data) < 512 {
		return true
	}

	return false
}

func parseDataBlock(byteSlice []byte) (DataBlock, error) {
	if byteSlice == nil || len(byteSlice) < 4 {
		return DataBlock{}, errors.New("byteSlice parameter was nil or number of bytes in byteSlice is less than 4 for a data block")
	}

	opcode := binary.BigEndian.Uint16(byteSlice[0:2])
	if opcode != dataBlockOpcode {
		return DataBlock{}, errors.New("Invalid opcode")
	}

	blockNumber := binary.BigEndian.Uint16(byteSlice[2:4])
	data := byteSlice[4:len(byteSlice)]

	return DataBlock{blockNumber, data}, nil
}


func dataBlockToSlice(dataBlock DataBlock, dataBlockSlice []byte) int {

	dataBlockLength := 4 + len(dataBlock.data)
	binary.BigEndian.PutUint16(dataBlockSlice[0:2], dataBlockOpcode)
	binary.BigEndian.PutUint16(dataBlockSlice[2:4], dataBlock.blockNumber)
	copy(dataBlockSlice[4:dataBlockLength], dataBlock.data)

	return dataBlockLength

}


type Ack struct {
	blockNumber uint16
}

func (a Ack) getType() uint16 {
	return ackOpcode
}

func parseAck(byteSlice []byte) (Ack, error) {
	if byteSlice == nil || len(byteSlice) < 4 {
		return Ack{}, errors.New("byteSlice parameter was nil or number of bytes in byteSlice is less than 4 for a data block")
	}

	opcode := binary.BigEndian.Uint16(byteSlice[0:2])

	if opcode != ackOpcode {
		return Ack{}, errors.New("Invalid opcode")
	}

	blockNumber := binary.BigEndian.Uint16(byteSlice[2:4])

	return Ack{blockNumber}, nil
}


func ackToSlice(ack Ack, ackSlice []byte) int {

	binary.BigEndian.PutUint16(ackSlice[0:2], ackOpcode)
	binary.BigEndian.PutUint16(ackSlice[2:4], ack.blockNumber)

	return 4

}

type TftpError struct {
	errorCode uint16
	errMsg string
}

func (e TftpError) getType() uint16 {
	return errorOpcode
}

func parseTftpErrorSlice(byteSlice []byte) (TftpError, error) {
	if byteSlice == nil || len(byteSlice) < 4 {
		return TftpError{}, errors.New("byteSlice parameter was nil")
	}

	opcode := binary.BigEndian.Uint16(byteSlice[0:2])
	if opcode != errorOpcode {
		return TftpError{}, errors.New("Invalid opcode")
	}

	errorCode := binary.BigEndian.Uint16(byteSlice[2:4])

	if len(byteSlice) == 4 {
		return TftpError{errorCode, ""}, nil
	}

	buffer := bytes.NewBuffer(byteSlice[4:len(byteSlice)])
	errorBytes, err := buffer.ReadBytes((byte)(0))
	if err != nil {
		return TftpError{}, err
	}

	errorBytesLength := len(errorBytes) - 1
	if errorBytesLength < 1 {
		return TftpError{errorCode, ""}, nil
	}

	errorString := string(errorBytes[:errorBytesLength])

	return TftpError{errorCode, errorString}, nil

}

func toTftpErrorSlice(tftpError TftpError, errorSlice []byte) int {
	errorLength :=  4 + len(tftpError.errMsg) + 1

	binary.BigEndian.PutUint16(errorSlice[0:2], errorOpcode)
	binary.BigEndian.PutUint16(errorSlice[2:4], tftpError.errorCode)

	copy(errorSlice[4:errorLength], []byte(tftpError.errMsg))

	return errorLength

}

