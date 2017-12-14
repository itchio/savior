package tarextractor_test

import (
	"testing"

	"github.com/itchio/savior"
	"github.com/stretchr/testify/assert"

	"github.com/itchio/savior/tarextractor"
)

func TestTar(t *testing.T) {
	ex := tarextractor.New()

	err := ex.Configure(&savior.ExtractorParams{
		LastCheckpoint: nil,
		OnProgress:     nil,
		Sink: &savior.Sink{
			Directory: ".",
		},
	})
	assert.NoError(t, err)

	// err = ex.Work()
	// assert.NoError(t, err)
}
