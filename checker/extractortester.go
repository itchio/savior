package checker

import (
	"log"
	"reflect"
	"testing"

	humanize "github.com/dustin/go-humanize"
	"github.com/itchio/savior"
	"github.com/mohae/deepcopy"
)

type MakeExtractorFunc func() savior.Extractor
type ShouldSaveFunc func() bool

func RunExtractorText(t *testing.T, makeExtractor MakeExtractorFunc, shouldSave ShouldSaveFunc) {
	var c *savior.ExtractorCheckpoint

	sc := NewTestSaveConsumer(3*1024*1024, func(checkpoint *savior.ExtractorCheckpoint) (savior.AfterSaveAction, error) {
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
