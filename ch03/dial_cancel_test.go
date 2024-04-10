package ch03

import (
	"context"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestDialContextCancel(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    sync := make(chan struct{})

    go func() {
        defer func()  {
            sync <- struct{}{}
        }()

        var dialer net.Dialer
        dialer.Control = func(_, _ string, _ syscall.RawConn) error {
            time.Sleep(time.Second)
            return nil
        }

        connection, err := dialer.DialContext(ctx, "tcp", "10.0.0.1:80")

        if err != nil {
            t.Log(err)
            return
        }

        connection.Close()
        t.Error("connection did not time out")
    }()

    cancel()
    <-sync

    if ctx.Err() != context.Canceled {
        t.Errorf("expected canceled context; actual: %q", ctx.Err())
    }
}
