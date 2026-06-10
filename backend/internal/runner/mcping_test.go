package runner

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"
)

func TestPingServerSuccess(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer func() { _ = ln.Close() }()

	go func() {
		conn, acceptErr := ln.Accept()
		if acceptErr != nil {
			return
		}
		defer func() { _ = conn.Close() }()

		// Discard handshake
		length1, _ := readVarint(conn)
		if length1 > 0 {
			_ = discardN(conn, int(length1))
		}
		// Discard status request
		length2, _ := readVarint(conn)
		if length2 > 0 {
			_ = discardN(conn, int(length2))
		}

		// Send status response (length-prefixed JSON string)
		var respBuf bytes.Buffer
		writeString(&respBuf, `{"version":{"name":"1.20.1"}}`)
		_ = sendPacket(conn, 0x00, respBuf.Bytes())
	}()

	addr := ln.Addr().(*net.TCPAddr)
	ok, err := PingServer("127.0.0.1", uint16(addr.Port), 2*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected PingServer to return true")
	}
}

func TestPingServerUnreachable(t *testing.T) {
	ok, err := PingServer("127.0.0.1", 0, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected PingServer to return false for unreachable port")
	}
}

func discardN(r net.Conn, n int) error {
	buf := make([]byte, n)
	_, err := io.ReadFull(r, buf)
	return err
}
