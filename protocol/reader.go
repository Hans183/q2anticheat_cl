package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Reader reads binary protocol messages (little-endian, 2-byte length prefix)
type Reader struct {
	r       io.Reader
	buf     []byte
	msgLen  uint16
	pos     int
	lastErr error
}

// NewReader creates a new protocol reader
func NewReader(r io.Reader) *Reader {
	return &Reader{
		r:   r,
		buf: make([]byte, 0, 65536),
	}
}

// ReadMessage reads a complete message from the stream
// Format: [uint16 length][payload...]
func (r *Reader) ReadMessage() ([]byte, error) {
	// Read 2-byte length prefix (little-endian, matching q2pro's WL16)
	var msgLen uint16
	if err := binary.Read(r.r, binary.LittleEndian, &msgLen); err != nil {
		return nil, fmt.Errorf("read length: %w", err)
	}

	if msgLen == 0 {
		return nil, errors.New("zero-length message")
	}

	// Read payload
	buf := make([]byte, msgLen)
	if _, err := io.ReadFull(r.r, buf); err != nil {
		return nil, fmt.Errorf("read payload: %w", err)
	}

	r.buf = buf
	r.msgLen = msgLen
	r.pos = 0
	r.lastErr = nil

	return buf, nil
}

// ReadByte reads a single byte from the current message
func (r *Reader) ReadByte() (byte, error) {
	if r.pos >= len(r.buf) {
		return 0, io.ErrUnexpectedEOF
	}
	b := r.buf[r.pos]
	r.pos++
	return b, nil
}

// ReadUint16 reads a uint16 (little-endian) from the current message
func (r *Reader) ReadUint16() (uint16, error) {
	if r.pos+2 > len(r.buf) {
		return 0, io.ErrUnexpectedEOF
	}
	v := binary.LittleEndian.Uint16(r.buf[r.pos:])
	r.pos += 2
	return v, nil
}

// ReadUint32 reads a uint32 (little-endian) from the current message
func (r *Reader) ReadUint32() (uint32, error) {
	if r.pos+4 > len(r.buf) {
		return 0, io.ErrUnexpectedEOF
	}
	v := binary.LittleEndian.Uint32(r.buf[r.pos:])
	r.pos += 4
	return v, nil
}

// ReadBytes reads n bytes from the current message
func (r *Reader) ReadBytes(n int) ([]byte, error) {
	if r.pos+n > len(r.buf) {
		return nil, io.ErrUnexpectedEOF
	}
	data := r.buf[r.pos : r.pos+n]
	r.pos += n
	return data, nil
}

// ReadString reads a length-prefixed string (1-byte length + data)
func (r *Reader) ReadString() (string, error) {
	length, err := r.ReadByte()
	if err != nil {
		return "", err
	}

	data, err := r.ReadBytes(int(length))
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Remaining returns the number of unread bytes in the current message
func (r *Reader) Remaining() int {
	return len(r.buf) - r.pos
}

// Err returns the last error
func (r *Reader) Err() error {
	return r.lastErr
}
