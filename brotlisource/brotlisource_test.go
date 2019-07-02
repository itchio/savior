package brotlisource_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/itchio/headway/united"
	"github.com/itchio/savior"
	"github.com/itchio/savior/brotlisource"
	"github.com/itchio/savior/checker"
	"github.com/itchio/savior/fullyrandom"
	"github.com/itchio/savior/seeksource"
	"github.com/itchio/savior/semirandom"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func Test_Uninitialized(t *testing.T) {
	{
		ss := seeksource.FromBytes(nil)
		_, err := ss.Resume(nil)
		assert.NoError(t, err)

		bs := brotlisource.New(ss)
		_, err = bs.Read([]byte{})
		assert.Error(t, err)
		assert.True(t, errors.Cause(err) == savior.ErrUninitializedSource)

		_, err = bs.ReadByte()
		assert.Error(t, err)
		assert.True(t, errors.Cause(err) == savior.ErrUninitializedSource)
	}
}

type sample struct {
	name string
	data []byte
}

const dataSize = 16 * 1024 * 1024

func Test_Checkpoints(t *testing.T) {
	samples := []sample{
		{
			name: "zero",
			data: make([]byte, dataSize),
		},
		{
			name: "semirandom",
			data: semirandom.Bytes(dataSize),
		},
		{
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

				log.Printf("uncompressed size: %s", united.FormatBytes(int64(len(reference))))
				log.Printf("  compressed size: %s", united.FormatBytes(int64(len(compressed))))

				source := seeksource.FromBytes(compressed)
				bs := brotlisource.New(source)

				checker.RunSourceTest(t, bs, reference)
			})
		}
	}
}
