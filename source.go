package savior

import (
	"encoding/gob"
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
	io.ByteReader
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

func init() {
	gob.Register(&SourceCheckpoint{})
}
