package ch03

import (
    "context"
    "io"
    "time"
)


const defaultPingInterval = 30 * time.Second


func Pinger(ctx context.Context, writer io.Writer, reset <-chan time.Duration) {
    var interval time.Duration
    
    select {
    case <-ctx.Done():
        return
    // Pulled initial interval off reset channel
    case interval = <-reset:
    default:
    }

    if interval <= 0 {
        interval = defaultPingInterval
    }

    timer := time.NewTimer(interval)
    defer func() {
        if !timer.Stop() {
            <-timer.C
        }
    }()

    for {
        select {
        case <-ctx.Done():
            return
        case newInterval := <-reset:
            if !timer.Stop() {
                <-timer.C
            }
            if newInterval > 0 {
                interval = newInterval
            }
        case <-timer.C:
            if _, err := writer.Write([]byte("ping")); err != nil {
                // Track and act on consecutive timeouts here
                return
            }
        }

        _ = timer.Reset(interval)
    }
}