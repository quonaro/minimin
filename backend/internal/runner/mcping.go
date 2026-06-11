package runner

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"
)

// PingServer performs a lightweight Minecraft Server List Ping on the given host:port.
// It returns true if the server responds with a valid status JSON containing a version object.
// TryPingServer attempts to ping a Minecraft server from either the host (dev mode)
// or from inside the minimin Docker container. It first tries 127.0.0.1:hostPort,
// then mc-srv-<serverID>:25565 (internal Docker network port).
func TryPingServer(serverID string, hostPort uint16, timeout time.Duration) (bool, error) {
	candidates := []struct {
		host string
		port uint16
	}{
		{"127.0.0.1", hostPort},
		{fmt.Sprintf("mc-srv-%s", serverID), 25565},
	}
	var lastErr error
	for _, c := range candidates {
		ok, err := PingServer(c.host, c.port, timeout)
		if ok {
			slog.Debug("ping succeeded", "server_id", serverID, "host", c.host, "port", c.port)
			return true, nil
		}
		lastErr = err
		slog.Debug("ping failed", "server_id", serverID, "host", c.host, "port", c.port, "error", err)
	}
	return false, lastErr
}

func PingServer(host string, port uint16, timeout time.Duration) (bool, error) {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false, fmt.Errorf("dial %s: %w", addr, err)
	}
	defer func() { _ = conn.Close() }()

	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return false, fmt.Errorf("set deadline: %w", err)
	}

	// Handshake packet (nextState = 1 for status)
	if err := sendHandshake(conn, host, port); err != nil {
		return false, fmt.Errorf("send handshake: %w", err)
	}

	// Status request packet (empty payload, ID 0x00)
	if err := sendPacket(conn, 0x00, nil); err != nil {
		return false, fmt.Errorf("send status request: %w", err)
	}

	// Read response length
	length, err := readVarint(conn)
	if err != nil {
		return false, fmt.Errorf("read response length: %w", err)
	}
	if length <= 0 {
		return false, fmt.Errorf("invalid response length %d", length)
	}

	// Read full payload
	payload := make([]byte, length)
	if _, err := io.ReadFull(conn, payload); err != nil {
		return false, fmt.Errorf("read payload (%d bytes): %w", length, err)
	}

	// Parse packet ID and JSON string
	buf := bytes.NewBuffer(payload)
	packetID, err := readVarintBuf(buf)
	if err != nil {
		return false, fmt.Errorf("read packet id: %w", err)
	}
	if packetID != 0x00 {
		return false, fmt.Errorf("unexpected packet ID %d", packetID)
	}

	jsonStr, err := readString(buf)
	if err != nil {
		return false, fmt.Errorf("read json string: %w", err)
	}

	// Quick validation: must be valid JSON and contain version info
	var status struct {
		Version json.RawMessage `json:"version"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &status); err != nil {
		return false, fmt.Errorf("invalid status json: %w", err)
	}
	if status.Version == nil {
		return false, fmt.Errorf("status json missing version field")
	}
	return true, nil
}

func sendHandshake(w io.Writer, host string, port uint16) error {
	var buf bytes.Buffer
	writeVarintBuf(&buf, -1) // protocol version
	writeString(&buf, host)
	binary.Write(&buf, binary.BigEndian, port)
	writeVarintBuf(&buf, 1) // nextState = status
	return sendPacket(w, 0x00, buf.Bytes())
}

func sendPacket(w io.Writer, packetID int32, data []byte) error {
	var buf bytes.Buffer
	writeVarintBuf(&buf, packetID)
	if len(data) > 0 {
		buf.Write(data)
	}
	pkt := buf.Bytes()
	var lengthBuf bytes.Buffer
	writeVarintBuf(&lengthBuf, int32(len(pkt)))
	if _, err := w.Write(lengthBuf.Bytes()); err != nil {
		return err
	}
	_, err := w.Write(pkt)
	return err
}

func writeVarintBuf(buf *bytes.Buffer, value int32) {
	u := uint32(value)
	for {
		tmp := u & 0x7F
		u >>= 7
		if u != 0 {
			tmp |= 0x80
		}
		buf.WriteByte(byte(tmp))
		if u == 0 {
			break
		}
	}
}

func readVarint(r io.Reader) (int32, error) {
	var result int32
	var numRead int
	b := make([]byte, 1)
	for {
		if _, err := r.Read(b); err != nil {
			return 0, err
		}
		val := int32(b[0])
		result |= (val & 0x7F) << (7 * numRead)
		numRead++
		if numRead > 5 {
			return 0, fmt.Errorf("varint too big")
		}
		if val&0x80 == 0 {
			break
		}
	}
	return result, nil
}

func readVarintBuf(buf *bytes.Buffer) (int32, error) {
	var result int32
	var numRead int
	for {
		b, err := buf.ReadByte()
		if err != nil {
			return 0, err
		}
		val := int32(b)
		result |= (val & 0x7F) << (7 * numRead)
		numRead++
		if numRead > 5 {
			return 0, fmt.Errorf("varint too big")
		}
		if val&0x80 == 0 {
			break
		}
	}
	return result, nil
}

func writeString(buf *bytes.Buffer, s string) {
	writeVarintBuf(buf, int32(len(s)))
	buf.WriteString(s)
}

func readString(buf *bytes.Buffer) (string, error) {
	length, err := readVarintBuf(buf)
	if err != nil {
		return "", err
	}
	if length < 0 || length > int32(buf.Len()) {
		return "", fmt.Errorf("invalid string length %d", length)
	}
	return string(buf.Next(int(length))), nil
}
