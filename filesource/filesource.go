package filesource

import (
	"context"

	"github.com/itchio/httpkit/eos"
	"github.com/itchio/httpkit/eos/option"
	"github.com/itchio/savior"
	"github.com/itchio/savior/seeksource"
)

func OpenPaused(ctx context.Context, name string, opts ...option.Option) (savior.FileSource, error) {
	f, err := eos.Open(ctx, name, opts...)
	if err != nil {
		return nil, err
	}

	ss := seeksource.FromFile(f)
	fs := &fileSource{
		SeekSource: ss,
		f:          f,
	}
	return fs, nil
}

func Open(ctx context.Context, name string, opts ...option.Option) (savior.FileSource, error) {
	s, err := OpenPaused(ctx, name, opts...)
	if err != nil {
		return nil, err
	}

	_, err = s.Resume(nil)
	if err != nil {
		return nil, err
	}

	return s, nil
}

type fileSource struct {
	savior.SeekSource

	f eos.File
}

func (fs *fileSource) Close() error {
	return fs.f.Close()
}
