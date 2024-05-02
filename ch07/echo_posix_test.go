//go:build darwin || linux
// +build darwin linux

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


func TestEchoServerUnixDatagram(t *testing.T) {
    dir, err := ioutil.TempDir("", "echo_unixgram")
    if err != nil {
	t.Fatal(err)
    }
    defer func() {
	if rErr := os.RemoveAll(dir); rErr != nil {
	    t.Error(rErr)
	}
    }()

    ctx, cancel := context.WithCancel(context.Background())

    serverSocket := filepath.Join(dir, fmt.Sprintf("server%d.sock", os.Getpid()))
    serverAddr, err := datagramEchoServer(ctx, "unixgram", serverSocket)
    if err != nil {
	t.Fatal(err)
    }
    defer cancel()

    err = os.Chmod(serverSocket, os.ModeSocket | 0622)
    if err != nil {
	t.Fatal(err)
    }

    clientSocket := filepath.Join(dir, fmt.Sprintf("client%d.sock", os.Getpid()))
    client, err := net.ListenPacket("unixgram", clientSocket)
    if err != nil {
	t.Fatal(err)
    }
    defer func() {
        _ = client.Close()
    }()

    err = os.Chmod(clientSocket, os.ModeSocket | 0622)
    if err != nil {
	t.Fatal(err)
    }

    msg := []byte("ping")
    for i := 0; i < 3; i++ {
        _, err = client.WriteTo(msg, serverAddr)
	if err != nil {
	    t.Fatal(err)
	}
    }

    buf := make([]byte, 1024)
    for i := 0; i < 3; i++ {
	n, addr, err := client.ReadFrom(buf)
	if err != nil {
	    t.Fatal(err)
	}

	if addr.String() != serverAddr.String() {
	    t.Fatalf("received reply from %q instead of %q", addr, serverAddr)
	}

	if !bytes.Equal(msg, buf[:n]) {
	    t.Fatalf("expected reply %q; actual reply %q", msg, buf[:n])
	}
    }
}
