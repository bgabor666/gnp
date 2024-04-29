package tftp

import (
	"bytes"
	"errors"
	"log"
	"net"
	"time"
)


type Server struct {
    Payload []byte // the payload served for all read requests
    Retries uint8 // the number of times to retry a failed transmission
    Timeout time.Duration // the duration to wait for an acknowledgment
}


func (server Server) ListenAndServe(address string) error {
    connection, err := net.ListenPacket("udp", address)
    if err != nil {
	return err
    }
    defer func()  {
        _ = connection.Close()
    }()

    log.Printf("Listening on %s ...\n", connection.LocalAddr())

    return server.Serve(connection)
}

func (server Server) Serve(connection net.PacketConn) error {
    if connection == nil {
	return errors.New("nil connection")
    }

    if server.Payload == nil {
	return errors.New("payload is required")
    }

    if server.Retries == 0 {
	server.Retries = 10
    }

    if server.Timeout == 0 {
	server.Timeout = 6 * time.Second
    }

    var rrq ReadReq

    for {
	buf := make([]byte, DatagramSize)

	_, addr, err := connection.ReadFrom(buf)
	if err != nil {
	    return err
	}

	err = rrq.UnmarshalBinary(buf)
	if err != nil {
	    log.Printf("[%s] bad request: %v", addr, err)
	    continue
	}

	go server.handle(addr.String(), rrq)
    }
}

func (server Server) handle(clientAddr string, rrq ReadReq) {
    log.Printf("[%s] requested file: %s", clientAddr, rrq.Filename)

    connection, err := net.Dial("udp", clientAddr)
    if err != nil {
	log.Printf("[%s] dial: %v", clientAddr, err)
	return
    }
    defer func()  {
        _ = connection.Close()
    }()

    var (
	ackPkt Ack
	errPkt TFTPError
	dataPkt = Data{Payload: bytes.NewReader(server.Payload)}
	buf = make([]byte, DatagramSize)
    )

NEXTPACKET:
    for n := DatagramSize; n == DatagramSize; {
	data, err := dataPkt.MarshalBinary()
	if err != nil {
	    log.Printf("[%s] preparing data packet: %v", clientAddr, err)
	    return
	}

    RETRY:
	for i := server.Retries; i > 0; i-- {
	    // send the data packet
	    n, err = connection.Write(data)
	    if err != nil {
		log.Printf("[%s] write: %v", clientAddr, err)
		return
	    }

	    // wait for the client's ACK packet
	    _ = connection.SetReadDeadline(time.Now().Add(server.Timeout))

	    _, err = connection.Read(buf)
	    if err != nil {
		if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
		    continue RETRY
		}

		log.Printf("[%s] waiting for ACK: %v", clientAddr, err)
		return
	    }

	    switch {
	    case ackPkt.UnmarshalBinary(buf) == nil:
	        if uint16(ackPkt) == dataPkt.Block {
		    // received ACK, send next data packet
		    continue NEXTPACKET
		}
	    case errPkt.UnmarshalBinary(buf) == nil:
		log.Printf("[%s] received error: %v", clientAddr, errPkt.Message)
		return
	    default:
		log.Printf("[%s] bad packet", clientAddr)
	    }
	}

	log.Printf("[%s] exhausted retries", clientAddr)
	return
    }

    log.Printf("[%s] sent %d blocks", clientAddr, dataPkt.Block)
}
