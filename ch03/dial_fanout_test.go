package ch03

import (
    "context"
    "net"
    "sync"
    "testing"
    "time"
)


func TestDialContextCancelFanout(t *testing.T) {
    ctx, cancel := context.WithDeadline(
        context.Background(),
        time.Now().Add(10 * time.Second),
    )

    listener, err := net.Listen("tcp", "127.0.0.1:")
    if err != nil {
        t.Fatal(err)
    }
    defer listener.Close()

    go func()  {
        // Only accepts a single connection and closes it after a successful handshake.
        connection, err := listener.Accept()
        if err == nil {
            connection.Close()
        }
    }()

    dial := func(
        ctx context.Context,
        address string,
        response chan int,
        id int,
        waitgroup *sync.WaitGroup,
    ) {
        defer waitgroup.Done()

        var dialer net.Dialer

        connection, err := dialer.DialContext(ctx, "tcp", address)
        if err != nil {
            return
        }
        connection.Close()

        select {
        case <-ctx.Done():
        case response <-id:
        }
    }

    res := make(chan int)
    var waitgroup sync.WaitGroup

    for i := 0; i < 10; i++ {
        waitgroup.Add(1)
        go dial(ctx, listener.Addr().String(), res, i+1, &waitgroup)
    }

    response := <-res
    cancel()
    waitgroup.Wait()
    close(res)

    if ctx.Err() != context.Canceled {
        t.Errorf("expected canceled context; actual: %s", ctx.Err())
    }

    t.Logf("dialer %d retrieved the resource", response)
}
