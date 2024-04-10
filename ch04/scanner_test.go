package main

import (
	"bufio"
	"net"
	"reflect"
	"testing"
)

const payload = "The bigger the interface, the weaker the abstraction."

func TestScanner(t *testing.T) {
    listener, err := net.Listen("tcp", "127.0.0.1:")
    if err != nil {
	t.Fatal(err)
    }

    go func() {
	listenerConnection, err := listener.Accept()
	if err != nil {
	    t.Error(err)
	    return
	}
	defer listenerConnection.Close()

	_, err = listenerConnection.Write([]byte(payload))
	if err != nil {
	    t.Error(err)
	}
    }()

    dialerConnection, err := net.Dial("tcp", listener.Addr().String())
    if err != nil {
	t.Fatal(err)
    }
    defer dialerConnection.Close()

    scanner := bufio.NewScanner(dialerConnection)
    scanner.Split(bufio.ScanWords)

    var words []string

    for scanner.Scan() {
	words = append(words, scanner.Text())
    }

    // In case of EOF err will be nil.
    err = scanner.Err()
    if err != nil {
	t.Error(err)
    }

    expected := []string{
	"The",
	"bigger",
	"the",
	"interface,",
	"the",
	"weaker",
	"the",
	"abstraction.",
    }

    if !reflect.DeepEqual(words, expected) {
	t.Fatal("inaccurate scanned word list")
    }

    t.Logf("Scanned words: %#v", words)
}
