package ch03

import (
    "context"
    "io"
    "net"
    "testing"
    "time"
)

func TestPingerAdvanceDeadline(t *testing.T) {
    done := make(chan struct{})
    listener, err := net.Listen("tcp", "127.0.0.1:")
    if err != nil {
        t.Fatal(err)
    }

    begin := time.Now()

    go func() {
        defer func() {
            close(done)
        }()

        connection, err := listener.Accept()
        if err != nil {
            t.Log(err)
            return
        }

        ctx, cancel := context.WithCancel(context.Background())
        defer func() {
            cancel()
            connection.Close()
        }()

        resetTimer := make(chan time.Duration, 1)
        resetTimer <- time.Second

        go Pinger(ctx, connection, resetTimer)

        err = connection.SetDeadline(time.Now().Add(5 * time.Second))
        if err != nil {
            t.Error(err)
            return
        }

        buf := make([]byte, 1024)

        for {
            n, err := connection.Read(buf)
            if err != nil {
                return
            }

            t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])

            resetTimer <- 0

            err = connection.SetDeadline(time.Now().Add(5 * time.Second))
            if err != nil {
                t.Error(err)
                return
            }
        }
    }()

    clientConnection, err := net.Dial("tcp", listener.Addr().String())
    if err != nil {
        t.Fatal(err)
    }
    defer clientConnection.Close()

    clientBuffer := make([]byte, 1024)

    // Read up to four pings
    for i := 0; i < 4; i++ {
        n, err := clientConnection.Read(clientBuffer)
        if err != nil {
            t.Fatal(err)
        }
        t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), clientBuffer[:n])
    }

    // Should reset the ping timer
    _, err = clientConnection.Write([]byte("PONG!!!"))
    if err != nil {
        t.Fatal(err)
    }

    // Read up to four more pings
    for i := 0; i < 4; i++ {
        n, err := clientConnection.Read(clientBuffer)
        if err != nil {
            if err != io.EOF {
                t.Fatal(err)
            }
            break
        }
        t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), clientBuffer[:n])
    }

    <-done

    end := time.Since(begin).Truncate(time.Second)
    t.Logf("[%s] done", end)

    if end != 9 * time.Second {
        t.Fatalf("expected EOF at 9 seconds; actual %s", end)
    }
}
