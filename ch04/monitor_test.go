package main

import (
	"io"
	"log"
	"net"
	"os"
)

// Monitor embeds a log.Logger meant for logging network traffic.
type Monitor struct {
    *log.Logger
}


// Write implements the io.Writer interface.
func (monitor *Monitor) Write(traffic []byte) (int, error)  {
    return len(traffic), monitor.Output(2, string(traffic))
}


func ExampleMonitor() {
    monitor := &Monitor{Logger: log.New(os.Stdout, "monitor: ", 0)}

    listener, err := net.Listen("tcp", "127.0.0.1:")
    if err != nil {
        monitor.Fatal(err)
    }

    done := make(chan struct{})

    go func() {
        defer close(done)

        connection, err := listener.Accept()
        if err != nil {
            return
        }
        defer connection.Close()

        buffer := make([]byte, 1024)
        reader := io.TeeReader(connection, monitor)

        n, err := reader.Read(buffer)
        if err != nil && err != io.EOF {
            monitor.Println(err)
            return
        }

        writer := io.MultiWriter(connection, monitor)

        // echo the message
        _, err = writer.Write(buffer[:n])
        if err != nil && err != io.EOF {
            monitor.Println(err)
            return
        }
    }()

    client, err := net.Dial("tcp", listener.Addr().String())
    if err != nil {
        monitor.Fatal(err)
    }

    _, err = client.Write([]byte("Test\n"))
    if err != nil {
        monitor.Fatal(err)
    }

    _ = client.Close()
    <-done

    // Output:
    // monitor: Test
    // monitor: Test
}

