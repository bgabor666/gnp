package ch03

import (
	"io"
	"net"
	"testing"
	"time"
)

func TestDeadline(t *testing.T) {
    sync := make(chan struct{})

    listener, err := net.Listen("tcp", "127.0.0.1:")
    if err != nil {
        t.Fatal(err)
    }

    go func() {
        connection, err := listener.Accept()
        if err != nil {
            t.Log(err)
            return
        }

        defer func() {
            connection.Close()
            // Read from sync should not block due to early return
            close(sync)
        }()

        err = connection.SetDeadline(time.Now().Add(5 * time.Second))
        if err != nil {
            t.Error(err)
            return
        }

        buf := make([]byte, 1)
        // Blocked until remote sends data
        _, err = connection.Read(buf)
        nErr, ok := err.(net.Error)
        // If the cast fails or the error is not a timeout error
        if !ok || !nErr.Timeout() {
            t.Errorf("expected timeout error; actual: %v", err)
        }

        sync <- struct{}{}

        // Reset the deadline and wait for incoming data
        err = connection.SetDeadline(time.Now().Add(5 * time.Second))
        if err != nil {
            t.Error(err)
            return
        }

        _, err = connection.Read(buf)
        if err != nil {
            t.Error(err)
        }
    }()

    connection, err := net.Dial("tcp", listener.Addr().String())
    if err != nil {
        t.Fatal(err)
    }
    defer connection.Close()

    <-sync
    _, err = connection.Write([]byte("1"))
    if err != nil {
        t.Fatal(err)
    }

    buf := make([]byte, 1)
    // Blocked until remote node sends data (which it will not do)
    _, err = connection.Read(buf)
    if err != io.EOF {
        t.Errorf("expected server termination; actual: %v", err)
    }
}
