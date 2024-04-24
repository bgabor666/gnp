package echo

import (
	"context"
	"fmt"
	"net"
)


func echoServerUDP(ctx context.Context, addr string) (net.Addr, error) {
    server, err := net.ListenPacket("udp", addr)
    if err != nil {
	return nil, fmt.Errorf("binding to udp %s: %w", addr, err)
    }

    go func() {
        go func() {
            <-ctx.Done()
            _ = server.Close()
        }()

	buf := make([]byte, 1024)

	for {
	    // client to server
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
