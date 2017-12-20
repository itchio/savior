package zipextractor

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-errors/errors"
	"github.com/itchio/arkive/zip"
	"github.com/itchio/savior"
)

type zipExtractor struct {
	source savior.Source
	sink   savior.Sink

	reader     io.ReaderAt
	readerSize int64
	zr         *zip.Reader

	sc savior.SaveConsumer
}

var _ savior.Extractor = (*zipExtractor)(nil)

func New(reader io.ReaderAt, readerSize int64, sink savior.Sink) savior.Extractor {
	return &zipExtractor{
		reader:     reader,
		readerSize: readerSize,
		sink:       sink,
		sc:         savior.NopSaveConsumer(),
	}
}

func (ze *zipExtractor) SetSaveConsumer(sc savior.SaveConsumer) {
	ze.sc = sc
}

func (ze *zipExtractor) Resume(checkpoint *savior.ExtractorCheckpoint) error {
	if checkpoint != nil {
		return errors.New("zipextractor: can't resume with checkpoint yet (stub)")
	}

	var err error
	ze.zr, err = zip.NewReader(ze.reader, ze.readerSize)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	for _, zf := range ze.zr.File {
		err := func() error {

			entry := &savior.Entry{
				CanonicalPath:    filepath.ToSlash(zf.Name),
				CompressedSize:   int64(zf.CompressedSize64),
				UncompressedSize: int64(zf.UncompressedSize64),
				Mode:             zf.Mode(),
			}

			info := zf.FileInfo()

			if info.IsDir() {
				entry.Kind = savior.EntryKindDir
			} else if entry.Mode&os.ModeSymlink > 0 {
				entry.Kind = savior.EntryKindSymlink
			} else {
				entry.Kind = savior.EntryKindFile
			}

			switch entry.Kind {
			case savior.EntryKindDir:
				err := ze.sink.Mkdir(entry)
				if err != nil {
					return errors.Wrap(err, 0)
				}
			case savior.EntryKindSymlink:
				rc, err := zf.Open()
				if err != nil {
					return errors.Wrap(err, 0)
				}

				defer rc.Close()

				linkname, err := ioutil.ReadAll(rc)
				if err != nil {
					return errors.Wrap(err, 0)
				}

				err = ze.sink.Symlink(entry, string(linkname))
				if err != nil {
					return errors.Wrap(err, 0)
				}
			case savior.EntryKindFile:
				// TODO: we actually want to be able to stop/resume here
				rc, err := zf.Open()
				if err != nil {
					return errors.Wrap(err, 0)
				}

				defer rc.Close()

				writer, err := ze.sink.GetWriter(entry)
				if err != nil {
					return errors.Wrap(err, 0)
				}

				_, err = io.Copy(writer, rc)
				if err != nil {
					return errors.Wrap(err, 0)
				}
			}

			return nil
		}()
		if err != nil {
			return errors.Wrap(err, 0)
		}
	}

	return nil
}
