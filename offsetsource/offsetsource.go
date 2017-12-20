package offsetsource

import (
	"errors"
	"io"

	"github.com/itchio/savior"
)

type offsetSource struct {
	r io.Reader

	startOffset int64
	offset      int64
	totalBytes  int64
}

var _ savior.Source = (*offsetSource)(nil)

func New(r io.Reader, startOffset int64, totalBytes int64) savior.Source {
	return &offsetSource{
		r:           r,
		startOffset: startOffset,
		offset:      startOffset,
		totalBytes:  totalBytes,
	}
}

func (ofs *offsetSource) Save() (*savior.SourceCheckpoint, error) {
	return nil, errors.New("offsetSource can't Save()")
}

func (ofs *offsetSource) Resume(c *savior.SourceCheckpoint) (int64, error) {
	return ofs.offset, errors.New("offsetSource can't Resume()")
}

func (ofs *offsetSource) Read(buf []byte) (int, error) {
	n, err := ofs.r.Read(buf)
	ofs.offset += int64(n)
	return n, err
}

func (ofs *offsetSource) ReadByte() (byte, error) {
	buf := []byte{0}
	_, err := ofs.Read(buf)
	return buf[0], err
}

func (ofs *offsetSource) Progress() float64 {
	if ofs.totalBytes > 0 {
		return float64(ofs.offset) / float64(ofs.totalBytes)
	}

	return 0
}
