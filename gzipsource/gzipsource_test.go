package gzipsource_test

import (
	"log"
	"testing"

	"github.com/itchio/headway/united"
	"github.com/itchio/savior"
	"github.com/itchio/savior/checker"
	"github.com/itchio/savior/gzipsource"
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

		gs := gzipsource.New(ss)
		_, err = gs.Read([]byte{})
		assert.Error(t, err)
		assert.True(t, errors.Cause(err) == savior.ErrUninitializedSource)

		_, err = gs.ReadByte()
		assert.Error(t, err)
		assert.True(t, errors.Cause(err) == savior.ErrUninitializedSource)
	}
}

func Test_Checkpoints(t *testing.T) {
	reference := semirandom.Bytes(4 * 1024 * 1024 /* 4 MiB of random data */)
	compressed, err := checker.GzipCompress(reference)
	assert.NoError(t, err)

	log.Printf("uncompressed size: %s", united.FormatBytes(int64(len(reference))))
	log.Printf("  compressed size: %s", united.FormatBytes(int64(len(compressed))))

	source := seeksource.FromBytes(compressed)
	gs := gzipsource.New(source)

	checker.RunSourceTest(t, gs, reference)
}
