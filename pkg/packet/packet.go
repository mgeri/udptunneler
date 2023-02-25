package packet

import (
	"encoding/binary"
	"fmt"
	"net"
)

/*
### Packet Header
A packet has a 1 byte header specifying the type of payload it contains.

Type: uint8 => type of the packet (0x00 = HEARTBEAT, 0x01 = DATAGRAM)

### Packet Type 0x00 = HEARTBEAT
This type of packet has no payload. It is sent by the client to the server and helps ensure both ends of the connection know if the other end is alive.

### Packet Type 0x01 = DATAGRAM
This packet encapsulates the datagram observed by the client. Here is its complete description:

Datagram Length: uint16 => number of bytes of the datagram packet
UDP Channel Address: uint32 => destination address of the multicast group which the client joined to receive that datagram
UDP Channel: Port uint16 => destination port of the multicast group which the client joined to receive that datagram
Datagram Packet: variable []byte => actual datagram received by the client from the multicast channel
*/

const (
	TypeHeartbeat uint8 = 0x01
	TypeDatagram  uint8 = 0x02
)

const (
	HeartbeatPacketHeaderLen = 1
	DatagramPacketHeaderLen  = 1 + 2 + 4 + 2
)

type Packet interface {
	Decode([]byte) error
	Encode([]byte) error
	Length() int
}

type Heartbeat struct {
	Type uint8
}

func (p *Heartbeat) Decode(buffer []byte) error {
	if buffer[0] != TypeHeartbeat {
		return fmt.Errorf("invalid packet type [%d]", buffer[0])
	}
	p.Type = TypeHeartbeat
	return nil
}

func (p *Heartbeat) Encode(buffer []byte) error {
	buffer[0] = TypeHeartbeat
	return nil
}

func (p *Heartbeat) Length() int {
	return HeartbeatPacketHeaderLen
}

type Datagram struct {
	Type           uint8
	DatagramLength uint16
	UdpIP          net.IP
	UdpPort        uint16
	DatagramPacket []byte
}

func (p *Datagram) Decode(buffer []byte) error {
	if buffer[0] != TypeDatagram {
		return fmt.Errorf("invalid packet type [%d]", buffer[0])
	}
	p.Type = TypeDatagram
	p.DatagramLength = binary.LittleEndian.Uint16(buffer[1:3])
	p.UdpIP = net.IPv4(buffer[3], buffer[4], buffer[5], buffer[6])
	p.UdpPort = binary.LittleEndian.Uint16(buffer[7:9])
	p.DatagramPacket = buffer[9:]
	return nil
}

func (p *Datagram) Encode(buffer []byte) error {
	if p.DatagramPacket != nil {
		if len(p.DatagramPacket) != int(p.DatagramLength) {
			return fmt.Errorf("invalid datagram length [%d], expected [%d]", len(p.DatagramPacket), p.DatagramLength)
		}
	}
	buffer[0] = TypeDatagram
	binary.LittleEndian.PutUint16(buffer[1:3], p.DatagramLength)
	copy(buffer[3:7], p.UdpIP.To4())
	binary.LittleEndian.PutUint16(buffer[7:9], p.UdpPort)
	if (p.DatagramPacket != nil) && (len(p.DatagramPacket) > 0) {
		copy(buffer[DatagramPacketHeaderLen:], p.DatagramPacket)
	}
	return nil
}

func (p *Datagram) Length() int {
	return DatagramPacketHeaderLen + int(p.DatagramLength)
}

func Decode(buffer []byte) (Packet, error) {
	pktType := buffer[0]

	switch pktType {
	case TypeHeartbeat:
		p := Heartbeat{}
		err := p.Decode(buffer)
		if err != nil {
			return nil, err
		}
		return &p, nil
	case TypeDatagram:
		p := Datagram{}
		err := p.Decode(buffer)
		if err != nil {
			return nil, err
		}
		return &p, nil
	default:
		return nil, fmt.Errorf("unknown packet type [%d]", pktType)
	}
}

func Encode(p Packet, buffer []byte) error {
	err := p.Encode(buffer)
	return err
}
