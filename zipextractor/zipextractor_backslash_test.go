package zipextractor_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/itchio/arkive/zip"
	"github.com/itchio/savior"
	"github.com/itchio/savior/zipextractor"
	"github.com/stretchr/testify/assert"
)

func makeBackslashZip(t *testing.T, names map[string]string) []byte {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	for name, contents := range names {
		if contents == "" {
			_, err := zw.CreateHeader(&zip.FileHeader{Name: name})
			must(t, err)
			continue
		}
		w, err := zw.Create(name)
		must(t, err)
		_, err = w.Write([]byte(contents))
		must(t, err)
	}
	must(t, zw.Close())
	return buf.Bytes()
}

func TestZipNormalizeBackslashes(t *testing.T) {
	zipBytes := makeBackslashZip(t, map[string]string{
		`dir\sub\`:         "",
		`dir\sub\file.txt`: "hello",
	})

	_, err := zipextractor.New(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	assert.True(t, errors.Is(err, zip.ErrInsecurePath))

	ex, err := zipextractor.NewWithParams(bytes.NewReader(zipBytes), int64(len(zipBytes)), zipextractor.Params{
		NormalizeBackslashes: true,
	})
	must(t, err)

	var kinds = map[string]savior.EntryKind{}
	for _, entry := range ex.Entries() {
		kinds[entry.CanonicalPath] = entry.Kind
	}
	assert.EqualValues(t, savior.EntryKindFile, kinds["dir/sub/file.txt"])

	dir := t.TempDir()
	_, err = ex.Resume(nil, &savior.FolderSink{Directory: dir})
	must(t, err)

	contents, err := os.ReadFile(filepath.Join(dir, "dir", "sub", "file.txt"))
	must(t, err)
	assert.Equal(t, "hello", string(contents))
}

func TestZipNormalizeBackslashesStillRejectsTraversal(t *testing.T) {
	zipBytes := makeBackslashZip(t, map[string]string{
		`..\evil.txt`: "nope",
	})

	_, err := zipextractor.NewWithParams(bytes.NewReader(zipBytes), int64(len(zipBytes)), zipextractor.Params{
		NormalizeBackslashes: true,
	})
	assert.True(t, errors.Is(err, zip.ErrInsecurePath))
}
