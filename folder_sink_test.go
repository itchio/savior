package savior_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/itchio/savior"
	"github.com/stretchr/testify/assert"
)

func Test_FolderSink(t *testing.T) {
	assert := assert.New(t)

	dir, err := ioutil.TempDir("", "foldersink-test")
	tmust(t, err)

	fs := &savior.FolderSink{
		Directory: dir,
	}

	entry := &savior.Entry{
		Kind:          savior.EntryKindFile,
		Mode:          0644,
		CanonicalPath: "secret",
		WriteOffset:   0,
	}

	{
		w, err := fs.GetWriter(entry)
		tmust(t, err)
		_, err = w.Write([]byte("foobar"))
		tmust(t, err)
		err = w.Close()
		tmust(t, err)
	}
	entry.WriteOffset = 1
	{
		w, err := fs.GetWriter(entry)
		tmust(t, err)
		_, err = w.Write([]byte("ee"))
		tmust(t, err)
		err = w.Close()
		tmust(t, err)
	}

	bs, err := ioutil.ReadFile(filepath.Join(dir, "secret"))
	tmust(t, err)

	s := string(bs)
	assert.EqualValues("fee", s)
}

func Test_FolderSinkIgnorePaths(t *testing.T) {
	assert := assert.New(t)

	dir, err := ioutil.TempDir("", "foldersink-test")
	tmust(t, err)

	fs := &savior.FolderSink{
		Directory: dir,
	}

	entries := []*savior.Entry{
		&savior.Entry{
			Kind:          savior.EntryKindFile,
			Mode:          0644,
			CanonicalPath: "Icon\r",
			WriteOffset:   0,
		},
		&savior.Entry{
			Kind:          savior.EntryKindFile,
			Mode:          0644,
			CanonicalPath: "Foobar",
			WriteOffset:   0,
		},
	}

	for _, entry := range entries {
		{
			w, err := fs.GetWriter(entry)
			tmust(t, err)
			_, err = w.Write([]byte("foobar"))
			tmust(t, err)
			err = w.Close()
			tmust(t, err)
		}
	}

	files, err := ioutil.ReadDir(dir)
	tmust(t, err)

	assert.Equal(1, len(files))
}

// tmust shows a complete error stack and fails a test immediately
// if err is non-nil
func tmust(t *testing.T, err error) {
	if err != nil {
		t.Helper()
		t.Errorf("%+v", err)
		t.FailNow()
	}
}
