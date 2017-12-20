package tarextractor

import (
	"io"

	"github.com/go-errors/errors"
	"github.com/itchio/arkive/tar"
	"github.com/itchio/savior"
)

type tarExtractor struct {
	source savior.Source
	sink   savior.Sink

	sc savior.SaveConsumer
}

var _ savior.Extractor = (*tarExtractor)(nil)

func New(source savior.Source, sink savior.Sink) savior.Extractor {
	return &tarExtractor{
		source: source,
		sink:   sink,
	}
}

func (te *tarExtractor) SetSaveConsumer(sc savior.SaveConsumer) {
	te.sc = sc
}

func (te *tarExtractor) Resume(checkpoint *savior.ExtractorCheckpoint) error {
	if checkpoint != nil {
		return errors.New("tarextractor: cannot resume from checkpoint yet (stub)")
	}

	_, err := te.source.Resume(nil)
	if err != nil {
		return errors.Wrap(err, 0)
	}

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
			savior.Debugf(`tar: extracting dir %s`, entry.CanonicalPath)
			err := te.sink.Mkdir(entry)
			if err != nil {
				return errors.Wrap(err, 0)
			}
		case savior.EntryKindSymlink:
			savior.Debugf(`tar: extracting symlink %s`, entry.CanonicalPath)
			err := te.sink.Symlink(entry, hdr.Linkname)
			if err != nil {
				return errors.Wrap(err, 0)
			}
		case savior.EntryKindFile:
			savior.Debugf(`tar: extracting file %s`, entry.CanonicalPath)
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
