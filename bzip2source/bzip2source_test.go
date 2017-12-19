package bzip2source_test

import (
	"bytes"
	"log"
	"testing"

	"github.com/go-errors/errors"
	"github.com/itchio/savior"

	"github.com/dsnet/compress/bzip2"

	humanize "github.com/dustin/go-humanize"
	"github.com/itchio/savior/bzip2source"
	"github.com/itchio/savior/seeksource"
	"github.com/itchio/savior/semirandom"
	"github.com/stretchr/testify/assert"
)

func TestBzip2Source(t *testing.T) {
	reference := semirandom.Generate(4 * 1024 * 1024 /* 4 MiB of random data */)
	compressed, err := bzip2Compress(reference)
	assert.NoError(t, err)

	log.Printf("uncompressed size: %s", humanize.IBytes(uint64(len(reference))))
	log.Printf("  compressed size: %s", humanize.IBytes(uint64(len(compressed))))

	source := seeksource.FromBytes(compressed)
	bs := bzip2source.New(source, 256*1024 /* 128 KiB threshold */)

	savior.RunSourceTest(t, bs, reference)
}

func bzip2Compress(input []byte) ([]byte, error) {
	compressedBuf := new(bytes.Buffer)
	w, err := bzip2.NewWriter(compressedBuf, &bzip2.WriterConfig{Level: 2})
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	_, err = w.Write(input)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	err = w.Close()
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	return compressedBuf.Bytes(), nil
}
