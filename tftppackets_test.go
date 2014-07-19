package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestIORequestBasicParseSuccess(t *testing.T) {

	expectedMode := "octet"
	expectedFilename := "abc"
	byteSlice := []byte{0,1,'a','b','c',0,'o','c','t','e','t',0}

	ioRequest, err := parseIORequest(byteSlice)
	if err != nil {
		t.Error(err)
	}

	if ioRequest.isWrite != false {
		t.Error("Expected read request but was parsed as write request")
	}

	if ioRequest.filename != expectedFilename {
		t.Error(fmt.Sprintf("Expected %s for filename but was %s", expectedFilename, ioRequest.filename))
	}

	if ioRequest.mode != expectedMode {
		t.Error(fmt.Sprintf("Expected %s for mode but was %s", expectedMode, ioRequest.mode))
	}

}

func TestIORequestOpcodeParseFailure(t *testing.T) {

	byteSlice := []byte{0,3,'a','b','c',0,'o','c','t','e','t',0}

	_, err := parseIORequest(byteSlice)
	if err == nil {
		t.Error("Expected error on opcode 3 but didn't get an error")
	}

}

func TestIORequestFilenameParseFailure(t *testing.T) {

	byteSlice := []byte{0,3, 0,'o','c','t','e','t',0}

	_, err := parseIORequest(byteSlice)
	if err == nil {
		t.Error("Expected error on zero length filename but didn't get an error")
	}

}

func TestIORequestNoDelimiterFilenameParseFailure(t *testing.T) {

	byteSlice := []byte{0,3,'o','c','t','e','t'}

	_, err := parseIORequest(byteSlice)
	if err == nil {
		t.Error("Expected error on no delimiter after filename but didn't get an error")
	}

}


func TestIORequestNoFilenameParseFailure(t *testing.T) {

	byteSlice := []byte{0,3}

	_, err := parseIORequest(byteSlice)
	if err == nil {
		t.Error("Expected error on no filename but didn't get an error")
	}

}

func TestIORequestModeParseFailure(t *testing.T) {

	byteSlice := []byte{0,3,'a','b','c',0, 0}

	_, err := parseIORequest(byteSlice)
	if err == nil {
		t.Error("Expected error on zero length mode but didn't get an error")
	}

}

func TestIORequestNoModeParseFailure(t *testing.T) {

	byteSlice := []byte{0,3,'a','b','c',0}

	_, err := parseIORequest(byteSlice)
	if err == nil {
		t.Error("Expected error on missing mode but didn't get an error")
	}

}

func TestIORequestInvalidModeParseFailure(t *testing.T) {

	byteSlice := []byte{0,3,'a','b','c',0, 'm', 'a', 'i', 'l', 0}

	_, err := parseIORequest(byteSlice)
	if err == nil {
		t.Error("Expected error on invalid mode but didn't get an error")
	}

}

func TestDataBlockBasicParseSuccess(t *testing.T) {

	byteSlice := []byte{0,3, 0,10, 0,0,0,1}

	dataBlock, err := parseDataBlock(byteSlice)
	if err != nil {
		t.Error(err)
	}

	if dataBlock.blockNumber != 10 {
		t.Error(fmt.Sprintf("Expected block number 10, got %s", dataBlock.blockNumber))
	}

	if bytes.Compare(byteSlice[4:len(byteSlice)], dataBlock.data) != 0 {
		t.Error("data block parse error, data mismatch")
	}

	if !dataBlock.isFinal() {
		t.Error("data block length is less than 512 but was not considered as final")
	}
}

func TestDataBlockOpcodeParseFailure(t *testing.T) {

	byteSlice := []byte{0,4}

	_, err := parseDataBlock(byteSlice)
	if err == nil {
		t.Error("Expected error parsing opcode, didn't get an error")
	}

}

func TestDataBlockIncompleteOpcodeParseFailure(t *testing.T) {

	byteSlice := []byte{0}

	_, err := parseDataBlock(byteSlice)
	if err == nil {
		t.Error("Expected error parsing incomplete opcode, didn't get an error")
	}

}

func TestDataBlockIncompleteBlockNumberParseFailure(t *testing.T) {

	byteSlice := []byte{0,3, 0}

	_, err := parseDataBlock(byteSlice)
	if err == nil {
		t.Error("Expected error parsing incomplete block number, didn't get an error")
	}

}
func TestDataBlockBasicToSliceSuccess(t *testing.T) {

	var blockNumber uint16 = 1
	dataBytes := []byte{0,1,2}
	expectedDataBlockBytes := []byte{0,3,0,1,0,1,2}

	dataBlock := DataBlock{blockNumber, dataBytes}

	dataBlockLength :=  4 + len(dataBlock.data)
	dataBlockSlice := make([]byte, dataBlockLength)
	dataBlockToSlice(dataBlock, dataBlockSlice)

	if bytes.Compare(dataBlockSlice, expectedDataBlockBytes) != 0 {
		t.Error("serialized data block does not match the expected data block bytes")
	}

}

func TestAckBasicParseSuccess(t *testing.T) {

	byteSlice := []byte{0,4, 0,10}

	ack, err := parseAck(byteSlice)
	if err != nil {
		t.Error(err)
	}

	if ack.blockNumber != 10 {
		t.Error(fmt.Sprintf("Expected block number 10, got %s", ack.blockNumber))
	}

}

func TestAckOpcodeParseFailure(t *testing.T) {

	byteSlice := []byte{0,5}

	_, err := parseAck(byteSlice)
	if err == nil {
		t.Error("Expected error parsing opcode, didn't get an error")
	}

}

func TestAckIncompleteOpcodeParseFailure(t *testing.T) {

	byteSlice := []byte{0}

	_, err := parseAck(byteSlice)
	if err == nil {
		t.Error("Expected error parsing incomplete opcode, didn't get an error")
	}

}

func TestAckIncompleteBlockNumberParseFailure(t *testing.T) {

	byteSlice := []byte{0,3, 0}

	_, err := parseAck(byteSlice)
	if err == nil {
		t.Error("Expected error parsing incomplete block number, didn't get an error")
	}

}
func TestAckBasicToSliceSuccess(t *testing.T) {

	var blockNumber uint16 = 1
	expectedAckBytes := []byte{0,4,0,1}

	ack := Ack{blockNumber}

	ackLength := 4
	ackSlice := make([]byte, ackLength)
	ackToSlice(ack, ackSlice)

	if bytes.Compare(ackSlice, expectedAckBytes) != 0 {
		t.Error("serialized ack does not match the expected data block bytes")
	}

}

func TestTftpErrorBasicParseSuccess(t *testing.T) {

	byteSlice := []byte{0,5, 0,10, 'a', 'b', 'c', 0}

	tftpErr, err := parseTftpErrorSlice(byteSlice)
	if err != nil {
		t.Error(err)
	}

	if tftpErr.errorCode != 10 {
		t.Error(fmt.Sprintf("Expected error code 10, got %s", tftpErr.errorCode))
	}

}

func TestTftpOpcodeParseFailure(t *testing.T) {

	byteSlice := []byte{0,6}

	_, err := parseTftpErrorSlice(byteSlice)
	if err == nil {
		t.Error("Expected error parsing opcode, didn't get an error")
	}

}

func TestTftpIncompleteOpcodeParseFailure(t *testing.T) {

	byteSlice := []byte{0}

	_, err := parseTftpErrorSlice(byteSlice)
	if err == nil {
		t.Error("Expected error parsing incomplete opcode, didn't get an error")
	}

}

func TestTftpIncompleteErrorCodeParseFailure(t *testing.T) {

	byteSlice := []byte{0,3, 0}

	_, err := parseTftpErrorSlice(byteSlice)
	if err == nil {
		t.Error("Expected error parsing incomplete block number, didn't get an error")
	}

}

func TestTftpIncompleteErrorMsgParseFailure(t *testing.T) {

	byteSlice := []byte{0,3, 0, 1, 'a', 'b'}

	_, err := parseTftpErrorSlice(byteSlice)
	if err == nil {
		t.Error("Expected error parsing incomplete error string, didn't get an error")
	}

}

func TestTftpNoErrorMsgParseSucces(t *testing.T) {

	byteSlice := []byte{0,5, 0, 1, 0}

	tftpErr, err := parseTftpErrorSlice(byteSlice)
	if err != nil {
		t.Error(err)
	}

	if tftpErr.errMsg != "" {
		t.Error("invalid error string")
	}

	if tftpErr.errorCode != 1 {
		t.Error("invalid error code")
	}
}

func TestTftpNoErrorMsgParseSucces2(t *testing.T) {

	byteSlice := []byte{0,5, 0, 1}

	tftpErr, err := parseTftpErrorSlice(byteSlice)
	if err != nil {
		t.Error(err)
	}

	if tftpErr.errMsg != "" {
		t.Error("invalid error string")
	}

	if tftpErr.errorCode != 1 {
		t.Error("invalid error code")
	}
}

func TestTftpErrorBasicToSliceSuccess(t *testing.T) {

	var errorCode uint16 = 1
	expectedErrorBytes := []byte{0,5,0,1, 'a', 'b', 'c', 0}
	errorMsg := "abc"

	tftpErr := TftpError{errorCode, errorMsg}

	errorLength :=  4 + len(tftpErr.errMsg) + 1
	errorSlice := make([]byte, errorLength)

	toTftpErrorSlice(tftpErr, errorSlice)

	if bytes.Compare(errorSlice, expectedErrorBytes) != 0 {
		t.Error("serialized ack does not match the expected data block bytes")
	}

}
