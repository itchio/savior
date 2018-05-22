package bzip2source_test

import (
	"log"
	"testing"

	"github.com/itchio/httpkit/progress"
	"github.com/itchio/savior"
	"github.com/itchio/savior/bzip2source"
	"github.com/itchio/savior/checker"
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

		bs := bzip2source.New(ss)
		_, err = bs.Read([]byte{})
		assert.Error(t, err)
		assert.True(t, errors.Cause(err) == savior.ErrUninitializedSource)

		_, err = bs.ReadByte()
		assert.Error(t, err)
		assert.True(t, errors.Cause(err) == savior.ErrUninitializedSource)
	}
}
func Test_Checkpoints(t *testing.T) {
	reference := semirandom.Bytes(4 * 1024 * 1024 /* 4 MiB of random data */)
	compressed, err := checker.Bzip2Compress(reference)
	assert.NoError(t, err)

	log.Printf("uncompressed size: %s", progress.FormatBytes(int64(len(reference))))
	log.Printf("  compressed size: %s", progress.FormatBytes(int64(len(compressed))))

	source := seeksource.FromBytes(compressed)
	bs := bzip2source.New(source)

	checker.RunSourceTest(t, bs, reference)
}
