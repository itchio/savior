package flatesource_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/itchio/kompress/flate"
	"github.com/itchio/savior/flatesource"
	"github.com/itchio/savior/seeksource"
	"github.com/stretchr/testify/assert"
)

func TestFlateSource(t *testing.T) {
	inputString := "That is a nice fox"
	inputData := []byte(inputString)

	compressedBuf := new(bytes.Buffer)
	w, err := flate.NewWriter(compressedBuf, 9)
	assert.NoError(t, err)

	_, err = w.Write(inputData)
	assert.NoError(t, err)

	err = w.Close()
	assert.NoError(t, err)

	compressedData := compressedBuf.Bytes()
	source := seeksource.New(bytes.NewReader(compressedData))
	fs := flatesource.New(source, 256*1024 /* 256 KiB */)

	_, err = fs.Resume(nil)
	assert.NoError(t, err)

	decompressedBuf := new(bytes.Buffer)
	_, err = io.Copy(decompressedBuf, fs)
	assert.NoError(t, err)

	outputString := decompressedBuf.String()
	assert.EqualValues(t, inputString, outputString)
}
