package jy901

import "encoding/binary"

const (
	FrameTypeAngles = byte(0x53)
)

type Frame struct {
	Type    byte
	Payload []byte
	Sum     byte
}

// IsValid calculate and compare the checksum
func (f Frame) IsValid() bool {
	var sum byte
	sum += FrameHead
	sum += f.Type
	for _, b := range f.Payload {
		sum += b
	}
	return f.Sum == sum
}

func (f Frame) GetAngles(rol, pit, yaw *float64) {
	*rol = float64(extractUint16(f.Payload)) / 32768
	*pit = float64(extractUint16(f.Payload[2:])) / 32768
	*yaw = float64(extractUint16(f.Payload[4:])) / 32768
}

func extractUint16(buf []byte) int16 {
	return int16(binary.LittleEndian.Uint16(buf))
}
