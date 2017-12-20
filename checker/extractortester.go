package checker

import (
	"bytes"
	"encoding/gob"
	"log"
	"reflect"
	"testing"

	humanize "github.com/dustin/go-humanize"
	"github.com/itchio/savior"
)

type MakeExtractorFunc func() savior.Extractor
type ShouldSaveFunc func() bool

func RunExtractorText(t *testing.T, makeExtractor MakeExtractorFunc, shouldSave ShouldSaveFunc) {
	var c *savior.ExtractorCheckpoint

	sc := NewTestSaveConsumer(3*1024*1024, func(checkpoint *savior.ExtractorCheckpoint) (savior.AfterSaveAction, error) {
		if shouldSave() {
			c2, checkpointSize := roundtripEThroughGob(t, checkpoint)
			log.Printf("↓ Saved at #%d (%s checkpoint)", c2.EntryIndex, humanize.IBytes(uint64(checkpointSize)))
			c = c2
			return savior.AfterSaveStop, nil
		}

		log.Printf("↷ Skipping over checkpoint at #%d", checkpoint.EntryIndex)
		return savior.AfterSaveContinue, nil
	})

	maxResumes := 24
	numResumes := 0
	for {
		if numResumes > maxResumes {
			t.Error("Too many resumes, something must be wrong")
			t.FailNow()
		}

		ex := makeExtractor()
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

func roundtripEThroughGob(t *testing.T, c *savior.ExtractorCheckpoint) (*savior.ExtractorCheckpoint, int) {
	saveBuf := new(bytes.Buffer)
	enc := gob.NewEncoder(saveBuf)
	err := enc.Encode(c)
	must(t, err)

	buflen := saveBuf.Len()

	c2 := &savior.ExtractorCheckpoint{}
	err = gob.NewDecoder(saveBuf).Decode(c2)
	must(t, err)

	return c2, buflen
}
