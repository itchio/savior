package zipextractor_test

import (
	"bytes"
	"log"
	"testing"

	humanize "github.com/dustin/go-humanize"
	"github.com/itchio/savior"
	"github.com/itchio/savior/checker"
	"github.com/itchio/savior/zipextractor"
	"github.com/stretchr/testify/assert"
)

func must(t *testing.T, err error) {
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
}

func TestZip(t *testing.T) {
	sink := checker.MakeTestSink()

	log.Printf("Making zip from checker.Sink...")
	zipBytes := checker.MakeZip(t, sink)

	makeZipExtractor := func() savior.Extractor {
		return zipextractor.New(bytes.NewReader(zipBytes), int64(len(zipBytes)), sink)
	}

	log.Printf("Testing .zip (%s), no resumes", humanize.IBytes(uint64(len(zipBytes))))
	checker.RunExtractorText(t, makeZipExtractor, func() bool {
		return false
	})

	log.Printf("Testing .zip (%s), every resume", humanize.IBytes(uint64(len(zipBytes))))
	checker.RunExtractorText(t, makeZipExtractor, func() bool {
		return true
	})

	log.Printf("Testing .zip (%s), every other resume", humanize.IBytes(uint64(len(zipBytes))))
	i := 0
	checker.RunExtractorText(t, makeZipExtractor, func() bool {
		i++
		return i%2 == 0
	})
}
