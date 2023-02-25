package frame

import (
	"encoding/binary"
	"github.com/bytedance/gopkg/lang/mcache"
	"io"
)

/*
Frame: frameHeader + framePayload(packet)

frameHeader: uint16, 2 bytes => length of frame payload (Packet)
framePayload: Packet
*/
const (
	FrameHeaderLen = 2
)

type FramePayload []byte

type StreamFrameCodec interface {
	Encode(io.Writer, FramePayload) error
	Decode(io.Reader) (FramePayload, error)
}

type frameCodec struct{}

func NewFrameCodec() StreamFrameCodec {
	return &frameCodec{}
}

func (p *frameCodec) Encode(w io.Writer, framePayload FramePayload) error {
	var f = framePayload
	var totalLen = uint16(len(framePayload)) + FrameHeaderLen

	err := binary.Write(w, binary.LittleEndian, &totalLen)
	if err != nil {
		return err
	}

	// make sure all data will be written to outbound stream
	for {
		n, err := w.Write(f) // write the frame payload to outbound stream
		if err != nil {
			return err
		}
		if n >= len(f) {
			break
		}
		if n < len(f) {
			f = f[n:]
		}
	}
	return nil
}

func (p *frameCodec) Decode(r io.Reader) (FramePayload, error) {
	var totalLen uint16
	err := binary.Read(r, binary.LittleEndian, &totalLen)
	if err != nil {
		return nil, err
	}

	buf := mcache.Malloc(int(totalLen - FrameHeaderLen))
	_, err = io.ReadFull(r, buf)

	if err != nil {
		return nil, err
	}
	return buf, nil
}
