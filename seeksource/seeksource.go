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
	if c != nil {
		_, err := ss.rs.Seek(c.Offset, io.SeekStart)
		if err != nil {
			return ss.offset, errors.Wrap(err, 0)
		}

		ss.offset = c.Offset
	}

	return ss.offset, nil
}

func (ss *seekSource) Read(buf []byte) (int, error) {
	n, err := ss.rs.Read(buf)
	ss.offset += int64(n)
	return n, err
}
