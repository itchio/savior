package bzip2source_test

import (
	"log"
	"testing"

	humanize "github.com/dustin/go-humanize"
	"github.com/itchio/savior/bzip2source"
	"github.com/itchio/savior/checker"
	"github.com/itchio/savior/seeksource"
	"github.com/itchio/savior/semirandom"
	"github.com/stretchr/testify/assert"
)

func TestBzip2Source(t *testing.T) {
	reference := semirandom.Bytes(4 * 1024 * 1024 /* 4 MiB of random data */)
	compressed, err := checker.Bzip2Compress(reference)
	assert.NoError(t, err)

	log.Printf("uncompressed size: %s", humanize.IBytes(uint64(len(reference))))
	log.Printf("  compressed size: %s", humanize.IBytes(uint64(len(compressed))))

	source := seeksource.FromBytes(compressed)
	bs := bzip2source.New(source, 256*1024 /* 128 KiB threshold */)

	checker.RunSourceTest(t, bs, reference)
}
