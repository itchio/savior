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
	testTarVariants(t, ".tar", int64(len(tarBytes)), source, sink)

	log.Printf("Compressing with gzip...")
	gzipBytes, err := checker.GzipCompress(tarBytes)
	must(t, err)
	gzipSource := gzipsource.New(seeksource.FromBytes(gzipBytes), 1*1024*1024)
	testTarVariants(t, ".tar.gz", int64(len(gzipBytes)), gzipSource, sink)

	log.Printf("Compressing with bzip2...")
	bzip2Bytes, err := checker.Bzip2Compress(tarBytes)
	must(t, err)
	bzip2Source := bzip2source.New(seeksource.FromBytes(bzip2Bytes), 1*1024*1024)
	testTarVariants(t, ".tar.bz2", int64(len(bzip2Bytes)), bzip2Source, sink)
}

func testTarVariants(t *testing.T, ext string, size int64, source savior.Source, sink savior.Sink) {
	makeExtractor := func() savior.Extractor {
		return tarextractor.New(source, sink)
	}

	log.Printf("Testing .tar (%s), no resumes", humanize.IBytes(uint64(size)))
	checker.RunExtractorText(t, makeExtractor, func() bool {
		return false
	})

	log.Printf("Testing .tar (%s), all resumes", humanize.IBytes(uint64(size)))
	checker.RunExtractorText(t, makeExtractor, func() bool {
		return true
	})

	log.Printf("Testing .tar (%s), every other resume", humanize.IBytes(uint64(size)))
	i := 0
	checker.RunExtractorText(t, makeExtractor, func() bool {
		i++
		return i%2 == 0
	})
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
