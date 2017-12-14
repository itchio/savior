package savior

import (
	"io"

	"github.com/go-errors/errors"
)

type SourceCheckpoint struct {
	Offset int64
	Data   interface{}
}

type Source interface {
	Save() (*SourceCheckpoint, error)
	Resume(checkpoint *SourceCheckpoint) (int64, error)

	io.Reader
}

func DiscardByRead(source Source, delta int64) error {
	buf := make([]byte, 4096)
	for delta > 0 {
		toRead := delta
		if toRead > int64(len(buf)) {
			toRead = int64(len(buf))
		}
		n, err := source.Read(buf[:toRead])
		if err != nil {
			return errors.Wrap(err, 0)
		}

		delta -= int64(n)
	}
	return nil
}

type NopSeeker struct {
	Offset int64
	Reader io.Reader
}

var _ io.ReadSeeker = (*NopSeeker)(nil)

func (ns *NopSeeker) Seek(offset int64, whence int) (int64, error) {
	return ns.Offset, nil
}

func (ns *NopSeeker) Read(buf []byte) (int, error) {
	return ns.Reader.Read(buf)
}
