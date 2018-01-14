package brotlisource_test

import (
	"fmt"
	"log"
	"testing"

	humanize "github.com/dustin/go-humanize"
	"github.com/itchio/savior/brotlisource"
	"github.com/itchio/savior/checker"
	"github.com/itchio/savior/fullyrandom"
	"github.com/itchio/savior/seeksource"
	"github.com/itchio/savior/semirandom"
	"github.com/stretchr/testify/assert"
)

type sample struct {
	name string
	data []byte
}

const dataSize = 16 * 1024 * 1024

func Test_BrotliSource(t *testing.T) {
	samples := []sample{
		sample{
			name: "zero",
			data: make([]byte, dataSize),
		},
		sample{
			name: "semirandom",
			data: semirandom.Bytes(dataSize),
		},
		sample{
			name: "fullyrandom",
			data: fullyrandom.Bytes(dataSize),
		},
	}

	qualities := []int{
		1,
		2,
		3,
		4,
		5,
		6,
		7,
		8,
		9,
	}

	for _, sample := range samples {
		for _, quality := range qualities {
			t.Run(fmt.Sprintf("%s-q%d", sample.name, quality), func(t *testing.T) {
				reference := sample.data
				compressed, err := checker.BrotliCompress(reference, quality)
				assert.NoError(t, err)

				log.Printf("uncompressed size: %s", humanize.IBytes(uint64(len(reference))))
				log.Printf("  compressed size: %s", humanize.IBytes(uint64(len(compressed))))

				source := seeksource.FromBytes(compressed)
				bs := brotlisource.New(source)

				checker.RunSourceTest(t, bs, reference)
			})
		}
	}
}