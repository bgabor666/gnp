package main

import (
	"bytes"
	"encoding/binary"
	"net"
	"reflect"
	"testing"
)

func TestPayloads(t *testing.T) {
    binary1 := Binary("Clear is better than clever.")
    binary2 := Binary("Don't panic.")
    string1 := String("Errors are values.")
    payloads := []Payload{&binary1, &binary2, &string1}

    listener, err := net.Listen("tcp", "127.0.0.1:")
    if err != nil {
        t.Fatal(err)
    }

    go func() {
        listenerConnection, err := listener.Accept()
        if err != nil {
            t.Error(err)
            return
        }
        defer listenerConnection.Close()

        for _, payload := range payloads {
            _, err = payload.WriteTo(listenerConnection)
            if err != nil {
                t.Error(err)
                break
            }
        }
    }()

    dialerConnection, err := net.Dial("tcp", listener.Addr().String())
    if err != nil {
        t.Fatal(err)
    }
    defer dialerConnection.Close()

    for i := 0; i < len(payloads); i++ {
        actual, err := decode(dialerConnection)
        if err != nil {
            t.Fatal(err)
        }

        if expected := payloads[i]; !reflect.DeepEqual(expected, actual) {
            t.Errorf("value mismatch: %v != %v", expected, actual)
            continue
        }

        t.Logf("[%T] %[1]q", actual)
    }
}


func TestMaxPayloadSize(t *testing.T) {
    buf := new(bytes.Buffer)
    err := buf.WriteByte(BinaryType)
    if err != nil {
        t.Fatal(err)
    }

    // 1 GB
    err = binary.Write(buf, binary.BigEndian, uint32(1<<30))
    if err != nil {
        t.Fatal(err)
    }

    var b Binary
    _, err = b.ReadFrom(buf)
    if err != ErrMaxPayloadSize {
        t.Fatalf("expected ErrMaxPayloadSize; actual: %v", err)
    }
}
