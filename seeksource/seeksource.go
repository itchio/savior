package seeksource

import (
	"io"

	"github.com/go-errors/errors"
	"github.com/itchio/savior"
)

type seekSource struct {
	rs io.ReadSeeker

	offset int64
}

var _ savior.Source = (*seekSource)(nil)

func New(rs io.ReadSeeker) *seekSource {
	return &seekSource{
		rs: rs,
	}
}

func (ss *seekSource) Save() (*savior.SourceCheckpoint, error) {
	c := &savior.SourceCheckpoint{
		Offset: ss.offset,
	}
	return c, nil
}

func (ss *seekSource) Resume(c *savior.SourceCheckpoint) (int64, error) {
	if c == nil {
		ss.offset = 0
	} else {
		ss.offset = c.Offset
	}

	newOffset, err := ss.rs.Seek(ss.offset, io.SeekStart)
	if err != nil {
		return newOffset, errors.Wrap(err, 0)
	}

	return ss.offset, nil
}

func (ss *seekSource) Read(buf []byte) (int, error) {
	n, err := ss.rs.Read(buf)
	ss.offset += int64(n)
	return n, err
}

func (ss *seekSource) ReadByte() (byte, error) {
	buf := []byte{0}
	_, err := ss.Read(buf)
	return buf[0], err
}
