package main

import (
	"crypto/rand"
	"io"
	"net"
	"testing"
)

func TestReadIntoBuffer(t *testing.T) {
    // 16 MB buffer for sending the payload
    payload := make([]byte, 1<<24)

    // Generate a random payload
    _, err := rand.Read(payload)
    if err != nil {
        t.Fatal(err)
    }

    listener, err := net.Listen("tcp", "127.0.0.1:")
    if err != nil {
        t.Fatal(err)
    }

    go func() {
        serverConnection, err := listener.Accept()
        if err != nil {
            t.Log(err)
            return
        }
        defer serverConnection.Close()

        _, err = serverConnection.Write(payload)
        if err != nil {
            t.Error(err)
        }
    }()

    clientConnection, err := net.Dial("tcp", listener.Addr().String())
    if err != nil {
        t.Fatal(err)
    }

    // 512 KB buffer for receiving the payload
    buf := make([]byte, 1<<19)

    for {
        n, err := clientConnection.Read(buf)
        if err != nil {
            if err != io.EOF {
                t.Error()
            }
            break
        }

        // buf[:n] is the data read from conn
        t.Logf("read %d bytes", n)
    }

    clientConnection.Close()
}
