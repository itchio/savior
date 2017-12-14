package savior

import (
	"io"
	"log"
	"testing"

	humanize "github.com/dustin/go-humanize"
	"github.com/itchio/savior/checker"
	"github.com/stretchr/testify/assert"
)

func RunSourceTest(t *testing.T, source Source, reference []byte) {
	_, err := source.Resume(nil)
	assert.NoError(t, err)

	output := checker.New(reference)
	totalCheckpoints := 0

	buf := make([]byte, 16*1024)
	for {
		n, readErr := source.Read(buf)

		_, err := output.Write(buf[:n])
		assert.NoError(t, err)
		if err != nil {
			t.FailNow()
		}

		if readErr != nil {
			if readErr == io.EOF {
				break
			}

			assert.NoError(t, readErr)
			t.FailNow()
		}

		c, err := source.Save()
		assert.NoError(t, err)
		if err != nil {
			t.FailNow()
		}

		if c != nil {
			totalCheckpoints++
			log.Printf("%s ↓ made checkpoint", humanize.IBytes(uint64(c.Offset)))

			newOffset, err := source.Resume(c)
			assert.NoError(t, err)
			if err != nil {
				t.FailNow()
			}

			log.Printf("%s ↻ resumed", humanize.IBytes(uint64(newOffset)))
			_, err = output.Seek(newOffset, io.SeekStart)
			if err != nil {
				assert.NoError(t, err)
				t.FailNow()
			}
		}
	}

	log.Printf("→ %d checkpoints total", totalCheckpoints)
	assert.True(t, totalCheckpoints > 0, "had at least one checkpoint")
}
