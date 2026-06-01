package bluetooth

import (
	"encoding/binary"
	"fmt"
)

const maxFrameSize = 4096

// FrameEncoder reassembles length-prefixed BLE chunks into complete message payloads.
type FrameEncoder struct {
	buf []byte
}

// Append ingests a BLE chunk and returns any fully received payloads.
func (f *FrameEncoder) Append(chunk []byte) ([][]byte, error) {
	f.buf = append(f.buf, chunk...)
	var out [][]byte
	for {
		if len(f.buf) < 2 {
			break
		}
		n := int(binary.BigEndian.Uint16(f.buf[:2]))
		if n > maxFrameSize {
			f.buf = nil
			return nil, fmt.Errorf("frame too large: %d", n)
		}
		if len(f.buf) < 2+n {
			break
		}
		payload := make([]byte, n)
		copy(payload, f.buf[2:2+n])
		out = append(out, payload)
		f.buf = f.buf[2+n:]
	}
	return out, nil
}

// EncodeFrame prefixes a payload with its 16-bit length.
func EncodeFrame(payload []byte) ([]byte, error) {
	if len(payload) > maxFrameSize {
		return nil, fmt.Errorf("payload exceeds %d bytes", maxFrameSize)
	}
	out := make([]byte, 2+len(payload))
	binary.BigEndian.PutUint16(out[:2], uint16(len(payload)))
	copy(out[2:], payload)
	return out, nil
}

// WriteChunks splits a frame into BLE-friendly chunks for WriteWithoutResponse.
func WriteChunks(frame []byte, chunkSize int, write func([]byte) error) error {
	if chunkSize <= 0 {
		chunkSize = 20
	}
	for len(frame) > 0 {
		n := chunkSize
		if len(frame) < n {
			n = len(frame)
		}
		part := frame[:n]
		frame = frame[n:]
		if err := write(part); err != nil {
			return err
		}
	}
	return nil
}
