package util

import (
	"errors"
	"fmt"
	"net"
)

func SetReceiveBuffer(c net.PacketConn, receiveBufferSize int) error {
	conn, ok := c.(interface{ SetReadBuffer(int) error })
	if !ok {
		return errors.New("connection doesn't allow setting of receive buffer size. Not a *net.UDPConn")
	}
	if err := conn.SetReadBuffer(receiveBufferSize); err != nil {
		return fmt.Errorf("failed to increase receive buffer size: %w", err)
	}
	return nil
}
