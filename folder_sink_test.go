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
	wtest.Must(t, err)

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
		wtest.Must(t, err)
		_, err = w.Write([]byte("foobar"))
		wtest.Must(t, err)
		err = w.Close()
		wtest.Must(t, err)
	}
	entry.WriteOffset = 1
	{
		w, err := fs.GetWriter(entry)
		wtest.Must(t, err)
		_, err = w.Write([]byte("ee"))
		wtest.Must(t, err)
		err = w.Close()
		wtest.Must(t, err)
	}

	bs, err := ioutil.ReadFile(filepath.Join(dir, "secret"))
	wtest.Must(t, err)

	s := string(bs)
	assert.EqualValues("fee", s)
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
