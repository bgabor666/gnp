package ch03

import (
    "context"
    "fmt"
    "io"
    "time"
)


func ExamplePinger() {
    ctx, cancel := context.WithCancel(context.Background())

    // In lieu (instead) of net.Conn
    reader, writer := io.Pipe()

    done := make(chan struct{})
    resetTimer := make(chan time.Duration, 1)

    // Initial ping interval
    resetTimer <- time.Second

    go func() {
        Pinger(ctx, writer, resetTimer)
        close(done)
    }()

    receivePing := func(duration time.Duration, reader io.Reader) {
        if duration >= 0 {
            fmt.Printf("resetting timer (%s)\n", duration)
            resetTimer <- duration
        }

        now := time.Now()
        buf := make([]byte, 1024)

        n, err := reader.Read(buf)
        if err != nil {
            fmt.Println(err)
        }

        fmt.Printf("received %q (%s)\n", buf[:n], time.Since(now).Round(100 * time.Millisecond))
    }

    for i, v := range []int64{0, 200, 300, 0, -1, -1, -1} {
        fmt.Printf("Run %d:\n", i + 1)
        receivePing(time.Duration(v) * time.Millisecond, reader)
    }

    cancel()

    // Ensure the pinger exits after canceling the context
    <-done

    // Output:
    // Run 1:
    // resetting timer (0s)
    // received "ping" (1s)
    // Run 2:
    // resetting timer (200ms)
    // received "ping" (200ms)
    // Run 3:
    // resetting timer (300ms)
    // received "ping" (300ms)
    // Run 4:
    // resetting timer (0s)
    // received "ping" (300ms)
    // Run 5:
    // received "ping" (300ms)
    // Run 6:
    // received "ping" (300ms)
    // Run 7:
    // received "ping" (300ms)
}
