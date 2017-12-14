package flatesource_test

import (
	"bytes"
	"io"
	"log"
	"testing"

	humanize "github.com/dustin/go-humanize"
	"github.com/itchio/kompress/flate"
	"github.com/itchio/savior/flatesource"
	"github.com/itchio/savior/seeksource"
	"github.com/stretchr/testify/assert"
)

func TestFlateSource(t *testing.T) {
	inputString := "That is a nice fox"
	var inputData = []byte(inputString)
	const maxSize = 12 * 1024 * 1024
	for len(inputData) < maxSize {
		inputData = append(inputData, inputData...)
	}

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
	fs := flatesource.New(source, 1*1024*1024 /* 1 MiB */)

	_, err = fs.Resume(nil)
	assert.NoError(t, err)

	decompressedBuf := new(bytes.Buffer)

	buf := make([]byte, maxSize/30)
	i := 0
	for {
		n, readErr := fs.Read(buf)

		_, err := decompressedBuf.Write(buf[:n])
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

		i++
		if i%20 == 0 {
			c, err := fs.Save()
			assert.NoError(t, err)
			if err != nil {
				t.FailNow()
			}

			if c != nil {
				log.Printf("got checkpoint at %s", humanize.IBytes(uint64(c.Offset)))

				newOffset, err := fs.Resume(c)
				assert.NoError(t, err)
				if err != nil {
					t.FailNow()
				}

				log.Printf("resumed at %d", newOffset)
				decompressedBuf.Truncate(int(newOffset))
			}
		}
	}

	assert.EqualValues(t, len(inputData), decompressedBuf.Len())
	assert.EqualValues(t, inputData, decompressedBuf.Bytes())
}
