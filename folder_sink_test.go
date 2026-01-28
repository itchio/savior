package savior_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/itchio/savior"
	"github.com/stretchr/testify/assert"
)

func Test_FolderSink(t *testing.T) {
	assert := assert.New(t)

	dir, err := os.MkdirTemp("", "foldersink-test")
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

	bs, err := os.ReadFile(filepath.Join(dir, "secret"))
	tmust(t, err)

	s := string(bs)
	assert.EqualValues("fee", s)
}

func Test_FolderSinkIgnorePaths(t *testing.T) {
	assert := assert.New(t)

	dir, err := os.MkdirTemp("", "foldersink-test")
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

	files, err := os.ReadDir(dir)
	tmust(t, err)

	assert.Equal(1, len(files))
}

func Test_FolderSinkPathTraversal(t *testing.T) {
	dir, err := ioutil.TempDir("", "foldersink-traversal-test")
	tmust(t, err)
	defer os.RemoveAll(dir)

	fs := &savior.FolderSink{
		Directory: dir,
	}

	// Test cases for path traversal attempts
	testCases := []struct {
		name          string
		canonicalPath string
		shouldFail    bool
	}{
		{"normal path", "subdir/file.txt", false},
		{"simple traversal", "../outside.txt", true},
		{"deep traversal", "../../../etc/passwd", true},
		{"traversal in middle", "foo/../../outside.txt", true},
		{"dot path", "./inside.txt", false},
		{"double dot in name", "foo..bar/file.txt", false}, // this is a valid filename
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			entry := &savior.Entry{
				Kind:          savior.EntryKindFile,
				Mode:          0644,
				CanonicalPath: tc.canonicalPath,
			}

			w, err := fs.GetWriter(entry)
			if tc.shouldFail {
				assert.Error(t, err, "expected error for path: %s", tc.canonicalPath)
				assert.True(t, errors.Is(err, savior.ErrPathTraversal), "expected ErrPathTraversal for path: %s", tc.canonicalPath)
			} else {
				assert.NoError(t, err, "unexpected error for path: %s", tc.canonicalPath)
				if w != nil {
					w.Close()
				}
			}
		})
	}
}

func Test_FolderSinkMkdirPathTraversal(t *testing.T) {
	dir, err := ioutil.TempDir("", "foldersink-mkdir-test")
	tmust(t, err)
	defer os.RemoveAll(dir)

	fs := &savior.FolderSink{
		Directory: dir,
	}

	// Test Mkdir with path traversal
	entry := &savior.Entry{
		Kind:          savior.EntryKindDir,
		Mode:          0755,
		CanonicalPath: "../escaped-dir",
	}

	err = fs.Mkdir(entry)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, savior.ErrPathTraversal))

	// Verify directory was not created outside
	_, err = os.Stat(filepath.Join(filepath.Dir(dir), "escaped-dir"))
	assert.True(t, os.IsNotExist(err), "directory should not have been created outside sink")
}

func Test_FolderSinkSymlinkPathTraversal(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink tests not applicable on Windows")
	}

	dir, err := ioutil.TempDir("", "foldersink-symlink-test")
	tmust(t, err)
	defer os.RemoveAll(dir)

	fs := &savior.FolderSink{
		Directory: dir,
	}

	// Test cases for symlink target validation
	testCases := []struct {
		name       string
		entryPath  string
		linkTarget string
		shouldFail bool
	}{
		{"valid relative symlink", "link.txt", "target.txt", false},
		{"valid subdir symlink", "subdir/link.txt", "../other.txt", false}, // stays within dir
		{"absolute symlink target", "link.txt", "/etc/passwd", true},
		{"escaping symlink target", "link.txt", "../../../etc/passwd", true},
		{"escaping from subdir", "subdir/link.txt", "../../outside.txt", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			entry := &savior.Entry{
				Kind:          savior.EntryKindSymlink,
				Mode:          0777,
				CanonicalPath: tc.entryPath,
			}

			err := fs.Symlink(entry, tc.linkTarget)
			if tc.shouldFail {
				assert.Error(t, err, "expected error for symlink %s -> %s", tc.entryPath, tc.linkTarget)
				assert.True(t, errors.Is(err, savior.ErrPathTraversal), "expected ErrPathTraversal")
			} else {
				assert.NoError(t, err, "unexpected error for symlink %s -> %s", tc.entryPath, tc.linkTarget)
			}
		})
	}
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
