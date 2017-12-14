package tarextractor

import (
	"io"

	"github.com/go-errors/errors"
	"github.com/itchio/arkive/tar"
	"github.com/itchio/savior"
)

type tarExtractor struct {
	source savior.Source
	sink   *savior.Sink
}

var _ savior.Extractor = (*tarExtractor)(nil)

func New() savior.Extractor {
	return &tarExtractor{}
}

func (te *tarExtractor) Configure(params *savior.ExtractorParams) error {
	te.source = params.Source
	te.sink = params.Sink
	return nil
}

func (te *tarExtractor) Work() error {
	sr, err := tar.NewSaverReader(te.source)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	for {
		hdr, err := sr.Next()
		if err != nil {
			if err == io.EOF {
				// we done!
				return nil
			}
			return errors.Wrap(err, 0)
		}

		entry := &savior.Entry{
			CanonicalPath:    hdr.Name,
			UncompressedSize: hdr.Size,
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			entry.Kind = savior.EntryKindDir
		case tar.TypeSymlink:
			entry.Kind = savior.EntryKindSymlink
		case tar.TypeReg:
			entry.Kind = savior.EntryKindFile
		default:
			// let's just ignore that one..
			continue
		}

		// if we were resuming, it'd be here

		switch entry.Kind {
		case savior.EntryKindDir:
			err := te.sink.Mkdir(entry)
			if err != nil {
				return errors.Wrap(err, 0)
			}
		case savior.EntryKindSymlink:
			err := te.sink.Symlink(entry, hdr.Linkname)
			if err != nil {
				return errors.Wrap(err, 0)
			}
		case savior.EntryKindFile:
			err := (func() error {
				w, err := te.sink.GetWriter(entry)
				if err != nil {
					return errors.Wrap(err, 0)
				}
				defer w.Close()

				_, err = io.Copy(w, sr)
				if err != nil {
					return errors.Wrap(err, 0)
				}
				return nil
			})()
			if err != nil {
				return errors.Wrap(err, 0)
			}
		}
	}
}
