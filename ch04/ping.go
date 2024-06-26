package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)


var (
    count = flag.Int("c", 3, "number of pings: <= 0 means forever")
    interval = flag.Duration("i", time.Second, "interval between pings")
    timeout = flag.Duration("W", 5*time.Second, "time to wait for a reply")
)


func init() {
    flag.Usage = func() {
        fmt.Printf("Usage: %s [options] host:port\nOptions:\n", os.Args[0])
        flag.PrintDefaults()
    }
}


func main() {
    flag.Parse()

    if flag.NArg() != 1 {
        fmt.Print("host:port is required\n\n")
        flag.Usage()
        os.Exit(1)
    }

    target := flag.Arg(0)
    fmt.Println("PING", target)

    if *count <= 0 {
        fmt.Println("CTRL+C to stop.")
    }

    msg := 0

    for (*count <= 0) || (msg < *count) {
        msg++
        fmt.Print(msg, " ")

        start := time.Now()
        connection, err := net.DialTimeout("tcp", target, *timeout)
        duration := time.Since(start)

        if err != nil {
            fmt.Printf("fail in %s: %v\n", duration, err)
            if netErr, ok := err.(net.Error); !ok || !netErr.Temporary() {
                os.Exit(1)
            }
        } else {
            _ = connection.Close()
            fmt.Println(duration)
        }

        time.Sleep(*interval)
    }
}
