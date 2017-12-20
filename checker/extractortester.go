package checker

import (
	"bytes"
	"encoding/gob"
	"log"
	"reflect"
	"testing"
	"time"

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
			log.Printf("↓ #%d [%.2f%%] (%s checkpoint)", c2.EntryIndex, c2.Progress*100, humanize.IBytes(uint64(checkpointSize)))
			c = c2
			return savior.AfterSaveStop, nil
		}

		// log.Printf("↷ Skipping over checkpoint at #%d", checkpoint.EntryIndex)
		return savior.AfterSaveContinue, nil
	})

	startTime := time.Now()

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
				log.Printf("↻ #%d [%.2f%%]", c.EntryIndex, c.Progress*100)
			} else {
				log.Printf("↻ #%d [%.2f%%], %v @ %s", c.EntryIndex, c.Progress*100, reflect.TypeOf(c.SourceCheckpoint.Data), humanize.IBytes(uint64(c.SourceCheckpoint.Offset)))
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

	totalDuration := time.Since(startTime)
	log.Printf("Done in %s, total resumes: %d", totalDuration, numResumes)
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
