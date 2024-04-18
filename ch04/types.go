package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)


const (
    BinaryType uint8 = iota + 1
    StringType

    // 10 MB
    MaxPayloadSize uint32 = 10 << 20
)


var ErrMaxPayloadSize = errors.New("maximum payload size exceeded")


type Payload interface {
    fmt.Stringer
    io.ReaderFrom
    io.WriterTo
    Bytes() []byte
}


type Binary []byte

func (m Binary) Bytes() []byte { return m }
func (m Binary) String() string { return string(m) }

func (m Binary) WriteTo(writer io.Writer) (int64, error) {
    // 1-byte type
    err := binary.Write(writer, binary.BigEndian, BinaryType)
    if err != nil {
	return 0, err
    }

    var headerSize int64 = 1

    // 4-byte size
    err = binary.Write(writer, binary.BigEndian, uint32(len(m)))
    if err != nil {
	return headerSize, err
    }
    headerSize += 4
    
    // Payload
    payloadSize, err := writer.Write(m)

    return headerSize + int64(payloadSize), err
}

func (m *Binary) ReadFrom(reader io.Reader) (int64, error) {
    var payloadType uint8    

    // 1-byte type
    err := binary.Read(reader, binary.BigEndian, &payloadType)
    if err != nil {
	return 0, err
    }

    var headerSize int64 = 1

    if payloadType != BinaryType {
	return headerSize, errors.New("invalid Binary")
    }

    var payloadSize uint32

    // 4-byte size
    err = binary.Read(reader, binary.BigEndian, &payloadSize)
    if err != nil {
	return headerSize, err
    }
    headerSize += 4

    if payloadSize > MaxPayloadSize {
	return headerSize, ErrMaxPayloadSize
    }

    *m = make([]byte, payloadSize)
    // Payload
    payloadRead, err := reader.Read(*m)

    return headerSize + int64(payloadRead), err
}


type String string

func (m String) Bytes() []byte { return []byte(m) }
func (m String) String() string { return string(m) }

func (m String) WriteTo(writer io.Writer) (int64, error) {
    // 1-byte type
    err := binary.Write(writer, binary.BigEndian, StringType)
    if err != nil {
	return 0, err
    }

    var headerSize int64 = 1

    // 4-byte size
    err = binary.Write(writer, binary.BigEndian, uint32(len(m)))
    if err != nil {
	return headerSize, err
    }

    headerSize += 4

    payloadSize, err := writer.Write([]byte(m))
    
    return headerSize + int64(payloadSize), err
}

func (m *String) ReadFrom(reader io.Reader) (int64, error) {
    var payloadType uint8

    // 1-byte type
    err := binary.Read(reader, binary.BigEndian, &payloadType)
    if err != nil {
	return 0, err
    }

    var headerSize int64 = 1
    
    if payloadType != StringType {
	return headerSize, errors.New("invalid String")
    }

    var payloadSize uint32

    // 4-byte size
    err = binary.Read(reader, binary.BigEndian, &payloadSize)
    if err != nil {
	return headerSize, err
    }

    headerSize += 4

    payloadBuffer := make([]byte, payloadSize)
    
    // Payload
    payloadRead, err := reader.Read(payloadBuffer)
    if err != nil {
	return headerSize, err
    }

    *m = String(payloadBuffer)

    return headerSize + int64(payloadRead), nil
}


func decode(reader io.Reader) (Payload, error) {
    var payloadType uint8
    err := binary.Read(reader, binary.BigEndian, &payloadType)
    if err != nil {
	return nil, err
    }

    var payload Payload

    switch payloadType {
    case BinaryType:
    	payload = new(Binary)
    case StringType:
	payload = new(String)
    default:
	return nil, errors.New("unknown type")
    }

    _, err = payload.ReadFrom(io.MultiReader(bytes.NewReader([]byte{payloadType}), reader))
    if err != nil {
	return nil, err
    }

    return payload, nil
}
