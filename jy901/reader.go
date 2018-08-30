package jy901

import (
	"bufio"
	"io"
)

const (
	FrameHead byte = 0x55

	FramePayloadLength = 8
)

type FrameReader interface {
	NextFrame() (Frame, error)
}

func NewFrameReader(r io.Reader) FrameReader {
	return &frameReader{
		Reader: bufio.NewReader(r),
	}
}

type frameReader struct {
	*bufio.Reader
}

func (r *frameReader) NextFrame() (f Frame, err error) {
retry:
	// skipping 0x00, seek for FrameHead
	for {
		var b byte
		if b, err = r.ReadByte(); err != nil {
			return
		}
		if b == FrameHead {
			break
		}
	}
	// read TYPE, PAYLOAD, CHECKSUM
	buf := make([]byte, 1+FramePayloadLength+1, 1+FramePayloadLength+1)
	if _, err = r.Read(buf); err != nil {
		return
	}
	// extract Frame
	f.Type = buf[0]
	f.Payload = buf[1 : 1+FramePayloadLength]
	f.Sum = buf[1+FramePayloadLength]
	// check checksum
	if !f.IsValid() {
		goto retry
	}
	return
}
