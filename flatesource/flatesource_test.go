package flatesource_test

import (
	"bytes"
	"io"
	"log"
	"testing"

	humanize "github.com/dustin/go-humanize"
	"github.com/itchio/kompress/flate"
	"github.com/itchio/savior/checker"
	"github.com/itchio/savior/flatesource"
	"github.com/itchio/savior/seeksource"
	"github.com/itchio/savior/semirandom"
	"github.com/stretchr/testify/assert"
)

func TestFlateSource(t *testing.T) {
	inputData := semirandom.Generate(4 * 1024 * 1024 /* 4 MiB */)

	compressedBuf := new(bytes.Buffer)
	w, err := flate.NewWriter(compressedBuf, 9)
	assert.NoError(t, err)

	_, err = w.Write(inputData)
	assert.NoError(t, err)

	err = w.Close()
	assert.NoError(t, err)

	log.Printf("uncompressed size: %s", humanize.IBytes(uint64(len(inputData))))
	log.Printf("  compressed size: %s", humanize.IBytes(uint64(compressedBuf.Len())))

	compressedData := compressedBuf.Bytes()
	source := seeksource.New(bytes.NewReader(compressedData))
	fs := flatesource.New(source, 256*1024 /* 256 KiB */)

	_, err = fs.Resume(nil)
	assert.NoError(t, err)

	output := checker.New(inputData)
	totalCheckpoints := 0

	buf := make([]byte, 32*1024)
	for {
		n, readErr := fs.Read(buf)

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

		c, err := fs.Save()
		assert.NoError(t, err)
		if err != nil {
			t.FailNow()
		}

		if c != nil {
			totalCheckpoints++
			log.Printf("%s ↓ made checkpoint", humanize.IBytes(uint64(c.Offset)))

			newOffset, err := fs.Resume(c)
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
