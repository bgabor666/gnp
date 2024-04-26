package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"
)


const (
    // The maximum supported datagram size
    DatagramSize = 516

    // The DatagramSize minus a 4-byte header
    BlockSize = DatagramSize - 4
)


type OpCode uint16

const (
    OpRRQ OpCode = iota + 1
    _
    OpData
    OpAck
    OpErr
)


type ErrCode uint16

const (
    ErrUnknown ErrCode = iota
    ErrNotFound
    ErrAccessViolation
    ErrDiskFull
    ErrIllegalOp
    ErrUnknownID
    ErrFileExists
    ErrNoUser
)


// Read request packet structure
//
// # 2 bytes # n bytes  # 1 byte # n bytes # 1 byte #
// ##################################################
// # OpCode  # Filename #    0   #   Mode  #   0    #
// ##################################################


type ReadReq struct {
    Filename string
    Mode string
}

// Although not used by our server, a client would make use of this method.
func (req ReadReq) MarshalBinary() ([]byte, error) {
    mode := "octet"
    if req.Mode != "" {
        mode = req.Mode
    }

    // OpCode + filename + 0 byte + mode + 0 byte
    packetLength := 2 + 2 + len(req.Filename) + 1 + len(req.Mode) + 1

    buf := new(bytes.Buffer)
    buf.Grow(packetLength)

    // write OpCode
    err := binary.Write(buf, binary.BigEndian, OpRRQ)
    if err != nil {
	return nil, err
    }

    // write filename
    _, err = buf.WriteString(req.Filename)
    if err != nil {
	return nil, err
    }

    // write 0 byte
    err = buf.WriteByte(0)
    if err != nil {
	return nil, err
    }

    // write mode
    _, err = buf.WriteString(mode)
    if err != nil {
	return nil, err
    }

    // write 0 byte
    err = buf.WriteByte(0)
    if err != nil {
	return nil, err
    }

    return buf.Bytes(), nil
}


func (req *ReadReq) UnmarshalBinary(packet []byte) error {
    buf := bytes.NewBuffer(packet)

    var code OpCode

    // read operation code
    err := binary.Read(buf, binary.BigEndian, &code)
    if err != nil {
	return err
    }

    if code != OpRRQ {
	return errors.New("invalid RRQ")
    }

    // read filename
    req.Filename, err = buf.ReadString(0)
    if err != nil {
	return errors.New("invalid RRQ")
    }

    // remove the 0-byte
    req.Filename = strings.TrimRight(req.Filename, "\x00")
    if len(req.Filename) == 0 {
	return errors.New("invalid RRQ")
    }

    // read mode
    req.Mode, err = buf.ReadString(0)
    if err != nil {
	return errors.New("invalid RRQ")
    }

    // remove the 0-byte
    req.Mode = strings.TrimRight(req.Mode, "\x00")
    if len(req.Mode) == 0 {
	return errors.New("invalid RRQ")
    }

    // enforce octet mode
    actual := strings.ToLower(req.Mode)
    if actual != "octet" {
	return errors.New("only binary transfers supported")
    }

    return nil
}


// Data packet structure
//
// # 2 bytes #    2 bytes   # n bytes #
// ####################################
// # OpCode  # Block number # Payload #
// ####################################

type Data struct {
    Block uint16
    Payload io.Reader
}

func (data *Data) MarshalBinary() ([]byte, error) {
    buf := new(bytes.Buffer)
    buf.Grow(DatagramSize)

    // block numbers increment from 1
    data.Block++

    // write operation code
    err := binary.Write(buf, binary.BigEndian, OpData)
    if err != nil {
	return nil, err
    }

    // write up to BlockSize worth of bytes
    _, err = io.CopyN(buf, data.Payload, BlockSize)
    if err != nil && err != io.EOF {
	return nil, err
    }

    return buf.Bytes(), nil
}

func (data *Data) UnmarshalBinary(packet []byte) error {
    if packetLength := len(packet); packetLength < 4 || packetLength > DatagramSize {
	return errors.New("invalid DATA")
    }

    var opcode OpCode

    // read the OpCode
    err := binary.Read(bytes.NewReader(packet[:2]), binary.BigEndian, &opcode)
    if err != nil || opcode != OpData {
	return errors.New("invalid DATA")
    }

    // read the block number
    err = binary.Read(bytes.NewReader(packet[2:4]), binary.BigEndian, &data.Block)
    if err != nil {
	return errors.New("invalid DATA")
    }

    // read the remaining bytes to payload
    data.Payload = bytes.NewBuffer(packet[4:])

    return nil
}
