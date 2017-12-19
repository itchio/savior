package tarextractor_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"

	"github.com/itchio/savior/semirandom"

	"github.com/itchio/arkive/tar"
	"github.com/itchio/savior"
	"github.com/stretchr/testify/assert"

	"github.com/itchio/savior/seeksource"
	"github.com/itchio/savior/tarextractor"
)

func must(t *testing.T, err error) {
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
}

func TestTar(t *testing.T) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	rng := rand.New(rand.NewSource(0xf617a899))
	for i := 0; i < 12; i++ {
		dataLen := rng.Int63n(512 * 1024)

		name := fmt.Sprintf("hello%d.bin", i)

		err := tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     name,
			Mode:     0644,
			Size:     int64(dataLen),
		})
		must(t, err)

		err = semirandom.Write(tw, dataLen, rng.Int63())
		must(t, err)
	}

	err := tw.Close()
	must(t, err)

	source := seeksource.FromBytes(buf.Bytes())

	ex := tarextractor.New()

	err = ex.Configure(&savior.ExtractorParams{
		LastCheckpoint: nil,
		OnProgress:     nil,
		Source:         source,
		Sink: &savior.FolderSink{
			Directory: "./ignored",
		},
	})
	must(t, err)

	err = ex.Work()
	must(t, err)
}
