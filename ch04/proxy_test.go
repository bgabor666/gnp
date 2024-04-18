package main

import (
	"io"
	"net"
	"sync"
	"testing"
)


func proxy(from io.Reader, to io.Writer) error {
    fromWriter, fromIsWriter := from.(io.Writer)
    toReader, toIsreader := to.(io.Reader)

    if toIsreader && fromIsWriter {
        // Send replies since "from" ad "to" implement the necessary interfaces.
        go func() {
            _, _ = io.Copy(fromWriter, toReader)
        }()
    }

    _, err := io.Copy(to, from)

    return err
}


func TestProxy(t *testing.T) {
    var waitGroup sync.WaitGroup

    // Server listens for a "ping" message and responds with a "pong" message.
    // All other messages are echoed back to the client.
    server, err := net.Listen("tcp", "127.0.0.1:")
    if err != nil {
        t.Fatal(err)
    }

    waitGroup.Add(1)

    // Server side goroutine
    go func() {
        defer waitGroup.Done()

        for {
            conn, err := server.Accept()
            if err != nil {
                return
            }

            go func(c net.Conn) {
                defer c.Close()

                for {
                    buf := make([]byte, 1024)
                    n, err := c.Read(buf)
                    if err != nil {
                        if err != io.EOF {
                            t.Error(err)
                        }
                        return
                    }

                    switch msg := string(buf[:n]); msg {
                    case "ping":
                        _, err = c.Write([]byte("pong"))
                    default:
                        _, err = c.Write(buf[:n])
                    }

                    if err != nil {
                        if err != io.EOF {
                            t.Error(err)
                        }
                        return
                    }
                }
            }(conn)
        }
    }()

    // The proxyServer proxies messages from client connections to the server.
    // Replies from the server are proxied back to the clients.
    proxyServer, err := net.Listen("tcp", "127.0.0.1:")
    if err != nil {
        t.Fatal(err)
    }

    waitGroup.Add(1)

    // Proxy side goroutine
    go func() {
        defer waitGroup.Done()
        
        for {
            conn, err := proxyServer.Accept()
            if err != nil {
                return
            }

            go func(from net.Conn) {
                defer from.Close()

                to, err := net.Dial("tcp", server.Addr().String())
                if err != nil {
                    t.Error(err)
                    return
                }

                defer to.Close()

                err = proxy(from, to)
                if err != nil {
                    t.Error(err)
                }
            }(conn)
        }
    }()

    client, err := net.Dial("tcp", proxyServer.Addr().String())
    if err != nil {
        t.Fatal(err)
    }

    messages := []struct{ Message, Reply string }{
        {"ping", "pong"},
        {"pong", "pong"},
        {"echo", "echo"},
        {"ping", "pong"},
    }

    for i, message := range messages {
        _, err = client.Write([]byte(message.Message))
        if err != nil {
            t.Fatal(err)
        }

        buf := make([]byte, 1024)

        n, err := client.Read(buf)
        if err != nil {
            t.Fatal(err)
        }

        actual := string(buf[:n])
        t.Logf("%q -> proxy -> %q", message.Message, actual)

        if actual != message.Reply {
            t.Errorf("%d: expected reply: %q; actual: %q", i, message.Reply, actual)
        }
    }

    _ = client.Close()
    _ = proxyServer.Close()
    _ = server.Close()

    waitGroup.Wait()
}
