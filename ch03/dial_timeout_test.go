package ch03

import (
    "net"
    "syscall"
    "testing"
    "time"
)

func DialTimeout(
    network,
    address string,
    timeout time.Duration,
) (net.Conn, error) {
    dialer := net.Dialer{
        Control: func(_, addr string, _ syscall.RawConn) error {
            return &net.DNSError{
                Err: "connection timed out",
                Name: addr,
                Server: "127.0.0.1",
                IsTimeout: true,
                IsTemporary: true,
            }
        },
        Timeout: timeout,
    }

    return dialer.Dial(network, address)
}

func TestDialTimeout(t *testing.T) {
    connection, err := DialTimeout("tcp", "10.0.0.1:http", 5 * time.Second)

    if err == nil {
        connection.Close()
        t.Fatal("connection did not time out")
    }

    nErr, ok := err.(net.Error)

    if !ok {
        t.Fatal(err)
    }

    if !nErr.Timeout() {
        t.Fatal("error is not a timout")
    }
}
