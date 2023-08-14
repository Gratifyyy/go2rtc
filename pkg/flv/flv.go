package flv

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	TagAudio = 8
	TagVideo = 9
	TagData  = 18

	CodecAAC = 10
	CodecAVC = 7
)

// Transport - it is recommended to implement io.Closer
type Transport interface {
	ReadTag() (byte, uint32, []byte, error)
}

// NewTransport - it is recommended to use bufio.Reader
func NewTransport(rd io.Reader) (Transport, error) {
	c := &flv{rd: rd}
	if err := c.readHeader(); err != nil {
		return nil, err
	}
	return c, nil
}

type flv struct {
	rd io.Reader
}

func (c *flv) ReadTag() (byte, uint32, []byte, error) {
	// https://rtmp.veriskope.com/pdf/video_file_format_spec_v10.pdf
	b := make([]byte, 4+11)
	if _, err := io.ReadFull(c.rd, b); err != nil {
		return 0, 0, nil, err
	}

	b = b[4 : 4+11] // skip previous tag size

	tagType := b[0]
	size := uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	timeMS := uint32(b[4])<<16 | uint32(b[5])<<8 | uint32(b[6]) | uint32(b[7])<<24

	b = make([]byte, size)
	if _, err := io.ReadFull(c.rd, b); err != nil {
		return 0, 0, nil, err
	}

	return tagType, timeMS, b, nil
}

func (c *flv) Close() error {
	if closer, ok := c.rd.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (c *flv) readHeader() error {
	b := make([]byte, 9)
	if _, err := io.ReadFull(c.rd, b); err != nil {
		return err
	}

	if string(b[:3]) != "FLV" {
		return errors.New("flv: wrong header")
	}

	_ = b[4] // flags (skip because unsupported by Reolink cameras)

	if skip := binary.BigEndian.Uint32(b[5:]) - 9; skip > 0 {
		if _, err := io.ReadFull(c.rd, make([]byte, skip)); err != nil {
			return err
		}
	}

	return nil
}

func TimeToRTP(timeMS uint32, clockRate uint32) uint32 {
	return timeMS * clockRate / 1000
}