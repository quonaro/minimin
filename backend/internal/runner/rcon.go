package runner

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

// RCONClient is a minimal Minecraft RCON TCP client.
type RCONClient struct {
	conn  net.Conn
	reqID int32
}

// DialRCON connects to addr and authenticates with password.
func DialRCON(addr, password string, timeout time.Duration) (*RCONClient, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}
	client := &RCONClient{conn: conn}
	if err := client.authenticate(password); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return client, nil
}

func (c *RCONClient) authenticate(password string) error {
	id, err := c.sendPacket(3, password)
	if err != nil {
		return err
	}
	respID, _, err := c.readPacket()
	if err != nil {
		return err
	}
	if respID == -1 {
		return fmt.Errorf("rcon authentication failed")
	}
	if respID != id {
		return fmt.Errorf("rcon auth response id mismatch")
	}
	return nil
}

// Execute sends a command and returns the server's text response.
func (c *RCONClient) Execute(cmd string) (string, error) {
	id, err := c.sendPacket(2, cmd)
	if err != nil {
		return "", err
	}
	for {
		respID, body, err := c.readPacket()
		if err != nil {
			return "", err
		}
		if respID == id {
			return body, nil
		}
	}
}

// Close tears down the underlying TCP connection.
func (c *RCONClient) Close() error {
	return c.conn.Close()
}

func (c *RCONClient) sendPacket(pktType int32, body string) (int32, error) {
	c.reqID++
	id := c.reqID
	buf := new(bytes.Buffer)
	payload := []byte(body)
	length := int32(4 + 4 + len(payload) + 1 + 1) // id + type + payload + 2 nulls
	_ = binary.Write(buf, binary.LittleEndian, length)
	_ = binary.Write(buf, binary.LittleEndian, id)
	_ = binary.Write(buf, binary.LittleEndian, pktType)
	buf.Write(payload)
	buf.WriteByte(0)
	buf.WriteByte(0)
	_, err := c.conn.Write(buf.Bytes())
	return id, err
}

func (c *RCONClient) readPacket() (int32, string, error) {
	var length int32
	if err := binary.Read(c.conn, binary.LittleEndian, &length); err != nil {
		return 0, "", err
	}
	data := make([]byte, length)
	if _, err := io.ReadFull(c.conn, data); err != nil {
		return 0, "", err
	}
	id := int32(binary.LittleEndian.Uint32(data[0:4]))
	_ = int32(binary.LittleEndian.Uint32(data[4:8])) // packet type
	payload := data[8:]
	if i := bytes.IndexByte(payload, 0); i >= 0 {
		payload = payload[:i]
	}
	return id, string(payload), nil
}
