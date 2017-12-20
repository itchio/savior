package tarextractor_test

import (
	"bytes"
	"log"
	"testing"

	humanize "github.com/dustin/go-humanize"
	"github.com/itchio/savior/bzip2source"
	"github.com/itchio/savior/checker"
	"github.com/itchio/savior/gzipsource"

	"github.com/itchio/arkive/tar"
	"github.com/itchio/savior"
	"github.com/stretchr/testify/assert"

	"github.com/itchio/savior/seeksource"
	"github.com/itchio/savior/tarextractor"
)

func must(t *testing.T, err error) {
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
}

func TestTar(t *testing.T) {
	sink := checker.MakeTestSink()

	log.Printf("Making tar from checker.Sink...")
	tarBytes := makeTar(t, sink)
	source := seeksource.FromBytes(tarBytes)
	log.Printf("Testing .tar (%s)", humanize.IBytes(uint64(len(tarBytes))))
	testTarExtractor(t, source, sink)

	log.Printf("Compressing with gzip...")
	gzipBytes, err := checker.GzipCompress(tarBytes)
	must(t, err)
	gzipSource := gzipsource.New(seeksource.FromBytes(gzipBytes), 1*1024*1024)
	log.Printf("Testing .tar.gz (%s)", humanize.IBytes(uint64(len(gzipBytes))))
	testTarExtractor(t, gzipSource, sink)

	log.Printf("Compressing with bzip2...")
	bzip2Bytes, err := checker.Bzip2Compress(tarBytes)
	must(t, err)
	bzip2Source := bzip2source.New(seeksource.FromBytes(bzip2Bytes), 1*1024*1024)
	log.Printf("Testing .tar.bz2 (%s)", humanize.IBytes(uint64(len(bzip2Bytes))))
	testTarExtractor(t, bzip2Source, sink)
}

func testTarExtractor(t *testing.T, source savior.Source, sink savior.Sink) {
	ex := tarextractor.New(source, sink)

	err := ex.Resume(nil)
	must(t, err)
}

func makeTar(t *testing.T, sink *checker.Sink) []byte {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	for _, item := range sink.Items {
		switch item.Entry.Kind {
		case savior.EntryKindDir:
			must(t, tw.WriteHeader(&tar.Header{
				Name:     item.Entry.CanonicalPath,
				Typeflag: tar.TypeDir,
				Mode:     0755,
			}))
		case savior.EntryKindFile:
			must(t, tw.WriteHeader(&tar.Header{
				Name:     item.Entry.CanonicalPath,
				Typeflag: tar.TypeReg,
				Size:     int64(len(item.Data)),
				Mode:     0644,
			}))

			_, err := tw.Write(item.Data)
			must(t, err)
		case savior.EntryKindSymlink:
			must(t, tw.WriteHeader(&tar.Header{
				Name:     item.Entry.CanonicalPath,
				Typeflag: tar.TypeSymlink,
				Mode:     0644,
				Linkname: item.Linkname,
			}))
		}
	}

	err := tw.Close()
	must(t, err)

	return buf.Bytes()
}
