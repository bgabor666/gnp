package echo

import (
	"context"
	"net"
	"os"
)


func streamingEchoServer(ctx context.Context, network string, addr string) (net.Addr, error) {
    server, err := net.Listen(network, addr)
    if err != nil {
	return nil, err
    }

    go func() {
        go func() {
	    <-ctx.Done()
	    _ = server.Close()
        }()

	for {
	    connection, err := server.Accept()
	    if err != nil {
		return
	    }

	    go func() {
	        defer func()  {
	            _ = connection.Close()
                }()

		for {
		    buf := make([]byte, 1024)
		    n, err := connection.Read(buf)
		    if err != nil {
			return
		    }

		    _, err = connection.Write(buf[:n])
		    if err != nil {
			return
		    }
		}
	    }()
	}
    }()

    return server.Addr(), nil
}


func datagramEchoServer(ctx context.Context, network string, addr string) (net.Addr, error) {
    server, err := net.ListenPacket(network, addr)
    if err != nil {
	return nil, err
    }

    go func() {
        go func() {
            <-ctx.Done()
	    _ = server.Close()
	    if network == "unixgram" {
		_ = os.Remove(addr)
	    }
        }()

	buf := make([]byte, 1024)

	for {
	    n, clientAddr, err := server.ReadFrom(buf)
	    if err != nil {
		return
	    }

	    _, err = server.WriteTo(buf[:n], clientAddr)
	    if err != nil {
		return
	    }
	}
    }()

    return server.LocalAddr(), nil
}
