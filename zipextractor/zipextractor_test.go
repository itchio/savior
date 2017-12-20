package zipextractor_test

import (
	"archive/zip"
	"bytes"
	"log"
	"os"
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
	log.Printf("Testing .zip (%s)", humanize.IBytes(uint64(len(zipBytes))))
	testZipExtractor(t, zipBytes, sink)
}

func testZipExtractor(t *testing.T, source []byte, sink savior.Sink) {
	var c *savior.ExtractorCheckpoint

	sc := checker.NewTestSaveConsumer(3*1024*1024, func(checkpoint *savior.ExtractorCheckpoint) (savior.AfterSaveAction, error) {
		log.Printf("↓ Saving at entry %d", checkpoint.EntryIndex)
		c = deepcopy.Copy(checkpoint).(*savior.ExtractorCheckpoint)
		return savior.AfterSaveStop, nil
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
				log.Printf("↻ Resuming from entry %d", c.EntryIndex)
			} else {
				log.Printf("↻ Resuming from entry %d, source is at %s", c.EntryIndex, humanize.IBytes(uint64(c.SourceCheckpoint.Offset)))
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
			fh.Method = zip.Deflate
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

	return buf.Bytes()
}
