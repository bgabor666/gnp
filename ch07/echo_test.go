package echo

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"testing"
)


func TestEchoServerUnix(t *testing.T) {
    dir, err := ioutil.TempDir("", "echo_unix")
    if err != nil {
	t.Fatal(err)
    }
    defer func() {
	if rErr := os.RemoveAll(dir); rErr != nil {
	    t.Error(rErr)
	}
    }()

    ctx, cancel := context.WithCancel(context.Background())
    socket := filepath.Join(dir, fmt.Sprintf("%d.sock", os.Getpid()))
    rAddr, err := streamingEchoServer(ctx, "unix", socket)
    if err != nil {
	t.Fatal(err)
    }
    defer cancel()

    err = os.Chmod(socket, os.ModeSocket | 0666)
    if err != nil {
	t.Fatal(err)
    }

    connection, err := net.Dial("unix", rAddr.String())
    if err != nil {
	t.Fatal(err)
    }
    defer func() {
        _ = connection.Close()
    }()

    msg := []byte("ping")
    // write 3 "ping" messages
    for i := 0; i < 3; i++ {
        _, err = connection.Write(msg)
	if err != nil {
	    t.Fatal(err)
	}
    }

    buf := make([]byte, 1024)
    // read once from the server
    n, err := connection.Read(buf)
    if err != nil {
	t.Fatal(err)
    }

    expected := bytes.Repeat(msg, 3)
    if !bytes.Equal(expected, buf[:n]) {
	t.Fatalf("expected reply %q; actual reply %q", expected, buf[:n])
    }
}
