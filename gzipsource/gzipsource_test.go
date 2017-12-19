package gzipsource_test

import (
	"bytes"
	"log"
	"testing"

	"github.com/go-errors/errors"
	"github.com/itchio/kompress/gzip"
	"github.com/itchio/savior"

	humanize "github.com/dustin/go-humanize"
	"github.com/itchio/savior/gzipsource"
	"github.com/itchio/savior/seeksource"
	"github.com/itchio/savior/semirandom"
	"github.com/stretchr/testify/assert"
)

func TestGzipSource(t *testing.T) {
	reference := semirandom.Generate(4 * 1024 * 1024 /* 4 MiB of random data */)
	compressed, err := gzipCompress(reference)
	assert.NoError(t, err)

	log.Printf("uncompressed size: %s", humanize.IBytes(uint64(len(reference))))
	log.Printf("  compressed size: %s", humanize.IBytes(uint64(len(compressed))))

	source := seeksource.FromBytes(compressed)
	gs := gzipsource.New(source, 256*1024 /* 256 KiB threshold */)

	savior.RunSourceTest(t, gs, reference)
}

func gzipCompress(input []byte) ([]byte, error) {
	compressedBuf := new(bytes.Buffer)
	w, err := gzip.NewWriterLevel(compressedBuf, 9)
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
