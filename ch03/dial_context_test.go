package ch03

import (
	"context"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestDialContext(t *testing.T) {
    deadline := time.Now().Add(5 * time.Second)

    ctx, cancel := context.WithDeadline(context.Background(), deadline)

    defer cancel()

    // DialContext is a method on a Dialer
    var dialer net.Dialer
    dialer.Control = func(_, _ string, _ syscall.RawConn) error {
        // Sleep long enough to reach the context's deadline.
        time.Sleep(5 * time.Second + time.Millisecond)
        return nil
    }

    connection, err := dialer.DialContext(ctx, "tcp", "10.0.0.0:80")

    if err == nil {
        connection.Close()
        t.Fatal("connection did not time out")
    }

    // Attempt a cast to net.Error
    nErr, ok := err.(net.Error)

    // If the casting fails
    if !ok {
        t.Error(err)
    } else {
        if !nErr.Timeout() {
            t.Errorf("error is not a timeout: %v", err)
        }
    }

    // Sanity check to make sure that the reaching of the deadline canceled the context.
    if ctx.Err() != context.DeadlineExceeded {
        t.Errorf("expected deadline exceeded; actual: %v", ctx.Err())
    }
}
