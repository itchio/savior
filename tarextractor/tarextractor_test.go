package tarextractor_test

import (
	"bytes"
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
	helloData := semirandom.Generate(8 * 1024 * 1024)

	err := tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     "hello.bin",
		Mode:     0644,
		Size:     int64(len(helloData)),
	})
	must(t, err)

	_, err = tw.Write(helloData)
	must(t, err)

	err = tw.Close()
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
