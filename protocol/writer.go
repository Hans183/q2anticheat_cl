package protocol

import (
	"encoding/binary"
	"io"
)

// Writer writes binary protocol messages (little-endian, 2-byte length prefix)
type Writer struct {
	w   io.Writer
	buf []byte
}

// NewWriter creates a new protocol writer
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:   w,
		buf: make([]byte, 0, 65536),
	}
}

// BeginMessage starts a new message. Returns a builder that can be written to.
func (w *Writer) BeginMessage() *MessageBuilder {
	return &MessageBuilder{
		w:    w,
		buf:  make([]byte, 0, 1024),
		pos:  0,
	}
}

// MessageBuilder accumulates bytes for a single message
type MessageBuilder struct {
	w    *Writer
	buf  []byte
	pos  int
}

// WriteByte writes a single byte
func (mb *MessageBuilder) WriteByte(b byte) {
	mb.buf = append(mb.buf, b)
}

// WriteUint16 writes a uint16 (little-endian)
func (mb *MessageBuilder) WriteUint16(v uint16) {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], v)
	mb.buf = append(mb.buf, b[:]...)
}

// WriteUint32 writes a uint32 (little-endian)
func (mb *MessageBuilder) WriteUint32(v uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	mb.buf = append(mb.buf, b[:]...)
}

// WriteBytes writes raw bytes
func (mb *MessageBuilder) WriteBytes(data []byte) {
	mb.buf = append(mb.buf, data...)
}

// WriteString writes a length-prefixed string (1-byte length + data)
func (mb *MessageBuilder) WriteString(s string) {
	if len(s) > 255 {
		s = s[:255]
	}
	mb.buf = append(mb.buf, byte(len(s)))
	mb.buf = append(mb.buf, s...)
}

// Flush sends the complete message with 2-byte length prefix (little-endian)
func (mb *MessageBuilder) Flush() error {
	var header [2]byte
	binary.LittleEndian.PutUint16(header[:], uint16(len(mb.buf)))

	// Write length prefix
	if _, err := mb.w.w.Write(header[:]); err != nil {
		return err
	}

	// Write payload
	if _, err := mb.w.w.Write(mb.buf); err != nil {
		return err
	}

	return nil
}
