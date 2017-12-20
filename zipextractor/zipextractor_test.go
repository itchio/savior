package zipextractor_test

import (
	"archive/zip"
	"bytes"
	"log"
	"os"
	"reflect"
	"testing"

	humanize "github.com/dustin/go-humanize"
	"github.com/itchio/savior"
	"github.com/itchio/savior/checker"
	"github.com/itchio/savior/zipextractor"
	"github.com/mohae/deepcopy"
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
	zipBytes := makeZip(t, sink)

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

type shouldSaveFunc func() bool

func testZipExtractor(t *testing.T, source []byte, sink savior.Sink, shouldSave shouldSaveFunc) {
	var c *savior.ExtractorCheckpoint

	sc := checker.NewTestSaveConsumer(3*1024*1024, func(checkpoint *savior.ExtractorCheckpoint) (savior.AfterSaveAction, error) {
		if shouldSave() {
			log.Printf("↓ Saving at #%d", checkpoint.EntryIndex)
			c = deepcopy.Copy(checkpoint).(*savior.ExtractorCheckpoint)
			return savior.AfterSaveStop, nil
		} else {
			log.Printf("↷ Skipping over checkpoint at #%d", checkpoint.EntryIndex)
			return savior.AfterSaveContinue, nil
		}
	})

	maxResumes := 24
	numResumes := 0
	for {
		if numResumes > maxResumes {
			t.Error("Too many resumes, something must be wrong")
			t.FailNow()
		}

		ex := zipextractor.New(bytes.NewReader(source), int64(len(source)), sink)
		ex.SetSaveConsumer(sc)

		if c == nil {
			log.Printf("→ First resume")
		} else {
			if c.SourceCheckpoint == nil {
				log.Printf("↻ Resuming from #%d", c.EntryIndex)
			} else {
				log.Printf("↻ Resuming from #%d, source is at %s :: %v", c.EntryIndex, humanize.IBytes(uint64(c.SourceCheckpoint.Offset)), reflect.TypeOf(c.SourceCheckpoint.Data))
			}
		}
		err := ex.Resume(c)
		if err != nil {
			if err == savior.StopErr {
				numResumes++
				continue
			}
			must(t, err)
		}

		// yay, we did it!
		break
	}

	log.Printf("Total resumes: %d", numResumes)
}

func makeZip(t *testing.T, sink *checker.Sink) []byte {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	shouldCompress := true
	numDeflate := 0
	numStore := 0

	for _, item := range sink.Items {
		fh := &zip.FileHeader{
			Name: item.Entry.CanonicalPath,
		}

		switch item.Entry.Kind {
		case savior.EntryKindDir:
			fh.SetMode(os.ModeDir | 0755)
			_, err := zw.CreateHeader(fh)
			must(t, err)
		case savior.EntryKindFile:
			fh.SetMode(0644)
			if shouldCompress {
				fh.Method = zip.Deflate
				numDeflate++
			} else {
				fh.Method = zip.Store
				numStore++
			}
			shouldCompress = !shouldCompress
			writer, err := zw.CreateHeader(fh)
			must(t, err)

			_, err = writer.Write(item.Data)
			must(t, err)
		case savior.EntryKindSymlink:
			fh.SetMode(os.ModeSymlink | 0644)
			writer, err := zw.CreateHeader(fh)
			must(t, err)

			_, err = writer.Write([]byte(item.Linkname))
			must(t, err)
		}
	}

	err := zw.Close()
	must(t, err)

	log.Printf("Made zip with %d deflate files, %d store files", numDeflate, numStore)

	return buf.Bytes()
}
