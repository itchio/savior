package tarextractor_test

import (
	"log"
	"testing"

	humanize "github.com/dustin/go-humanize"
	"github.com/itchio/savior/bzip2source"
	"github.com/itchio/savior/checker"
	"github.com/itchio/savior/gzipsource"

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
	tarBytes := checker.MakeTar(t, sink)
	source := seeksource.FromBytes(tarBytes)
	testTarVariants(t, ".tar", int64(len(tarBytes)), source, sink)

	log.Printf("Compressing with gzip...")
	gzipBytes, err := checker.GzipCompress(tarBytes)
	must(t, err)
	gzipSource := gzipsource.New(seeksource.FromBytes(gzipBytes))
	testTarVariants(t, ".tar.gz", int64(len(gzipBytes)), gzipSource, sink)

	log.Printf("Compressing with bzip2...")
	bzip2Bytes, err := checker.Bzip2Compress(tarBytes)
	must(t, err)
	bzip2Source := bzip2source.New(seeksource.FromBytes(bzip2Bytes))
	testTarVariants(t, ".tar.bz2", int64(len(bzip2Bytes)), bzip2Source, sink)
}

func testTarVariants(t *testing.T, ext string, size int64, source savior.Source, sink *checker.Sink) {
	makeExtractor := func() savior.Extractor {
		return tarextractor.New(source)
	}

	log.Printf("Testing .tar (%s), no resumes", humanize.IBytes(uint64(size)))
	checker.RunExtractorText(t, makeExtractor, sink, func() bool {
		return false
	})

	log.Printf("Testing .tar (%s), all resumes", humanize.IBytes(uint64(size)))
	checker.RunExtractorText(t, makeExtractor, sink, func() bool {
		return true
	})

	log.Printf("Testing .tar (%s), every other resume", humanize.IBytes(uint64(size)))
	i := 0
	checker.RunExtractorText(t, makeExtractor, sink, func() bool {
		i++
		return i%2 == 0
	})
}
