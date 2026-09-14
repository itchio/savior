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

type symlinkSink struct {
	fs *savior.FolderSink
}

func newSymlinkSink(t *testing.T) (*symlinkSink, string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("symlink tests not applicable on Windows")
	}
	parent := t.TempDir()
	dir := filepath.Join(parent, "dest")
	tmust(t, os.Mkdir(dir, 0o755))
	return &symlinkSink{fs: &savior.FolderSink{Directory: dir}}, parent
}

func (s *symlinkSink) symlink(path, target string) error {
	return s.fs.Symlink(&savior.Entry{Kind: savior.EntryKindSymlink, Mode: 0777, CanonicalPath: path}, target)
}

func (s *symlinkSink) mkdir(path string) error {
	return s.fs.Mkdir(&savior.Entry{Kind: savior.EntryKindDir, Mode: 0755, CanonicalPath: path})
}

func (s *symlinkSink) file(path, contents string) error {
	w, err := s.fs.GetWriter(&savior.Entry{Kind: savior.EntryKindFile, Mode: 0644, CanonicalPath: path})
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(contents)); err != nil {
		return err
	}
	return w.Close()
}

func assertTraversal(t *testing.T, err error, what string) {
	t.Helper()
	assert.True(t, errors.Is(err, savior.ErrPathTraversal), "%s: expected ErrPathTraversal, got %v", what, err)
}

// Only "dest" itself may exist next to the sink directory.
func assertNothingEscaped(t *testing.T, parent string) {
	t.Helper()
	entries, err := os.ReadDir(parent)
	tmust(t, err)
	assert.Equal(t, 1, len(entries), "nothing may be written next to dest")
}

func Test_FolderSinkSymlinkChain(t *testing.T) {
	s, parent := newSymlinkSink(t)

	// Each entry is lexically inside dest, but "d" is an alias for dest,
	// so "d/e" -> ".." would point at dest's parent.
	tmust(t, s.symlink("d", "."))
	assertTraversal(t, s.symlink("d/e", ".."), "d/e -> ..")
	assertTraversal(t, s.file("d/e/escaped.txt", "x"), "file under symlinked parent")
	assertTraversal(t, s.mkdir("d/e/escaped-dir"), "dir under symlinked parent")
	assertTraversal(t, s.file("d/direct.txt", "x"), "file directly under symlinked dir")

	assertNothingEscaped(t, parent)
}

func Test_FolderSinkSymlinkForwardReference(t *testing.T) {
	s, parent := newSymlinkSink(t)

	// "e" -> "d/.." is lexically dest, and "d" does not exist yet so
	// nothing can be resolved through it. Once "d" -> "." lands, "e"
	// points at dest's parent. The sink cannot see that at creation time;
	// what it guarantees is that nothing is written through "e".
	tmust(t, s.symlink("e", "d/.."))
	tmust(t, s.symlink("d", "."))
	assertTraversal(t, s.file("e/escaped.txt", "x"), "file through forward-referenced link")
	assertTraversal(t, s.mkdir("e/escaped-dir"), "dir through forward-referenced link")
	assertTraversal(t, s.symlink("e/escaped-link", "x"), "link through forward-referenced link")

	assertNothingEscaped(t, parent)
}

func Test_FolderSinkSymlinkLeafReplacement(t *testing.T) {
	t.Run("ordinary", func(t *testing.T) {
		s, parent := newSymlinkSink(t)
		dir := s.fs.Directory

		tmust(t, s.mkdir("sub"))
		tmust(t, s.file("sub/target.txt", "target"))

		// A leaf symlink may be replaced by a file, a directory, or another
		// link; the old link is removed, never written through.
		tmust(t, s.symlink("x", "sub"))
		tmust(t, s.file("x", "plain"))
		bs, err := os.ReadFile(filepath.Join(dir, "x"))
		tmust(t, err)
		assert.Equal(t, "plain", string(bs))
		bs, err = os.ReadFile(filepath.Join(dir, "sub", "target.txt"))
		tmust(t, err)
		assert.Equal(t, "target", string(bs), "replacing the link must not touch its target")

		tmust(t, s.symlink("y", "sub"))
		tmust(t, s.mkdir("y"))
		st, err := os.Lstat(filepath.Join(dir, "y"))
		tmust(t, err)
		assert.True(t, st.IsDir())

		tmust(t, s.symlink("z", "sub"))
		tmust(t, s.symlink("z", "sub/target.txt"))
		link, err := os.Readlink(filepath.Join(dir, "z"))
		tmust(t, err)
		assert.Equal(t, "sub/target.txt", link)

		assertNothingEscaped(t, parent)
	})

	t.Run("outside", func(t *testing.T) {
		s, parent := newSymlinkSink(t)
		dir := s.fs.Directory

		// A leaf that already points outside (left by an older extractor, or by
		// the forward-reference ordering) is replaced, not followed.
		tmust(t, os.Symlink("..", filepath.Join(dir, "out")))
		tmust(t, s.file("out", "plain"))
		st, err := os.Lstat(filepath.Join(dir, "out"))
		tmust(t, err)
		assert.True(t, st.Mode().IsRegular())
		tmust(t, os.Symlink("..", filepath.Join(dir, "out-dir")))
		tmust(t, s.mkdir("out-dir"))
		st, err = os.Lstat(filepath.Join(dir, "out-dir"))
		tmust(t, err)
		assert.True(t, st.IsDir())
		// but writing through it is still refused
		tmust(t, os.Symlink("..", filepath.Join(dir, "out-parent")))
		assertTraversal(t, s.file("out-parent/escaped.txt", "x"), "file through outside-pointing leaf")

		assertNothingEscaped(t, parent)
	})

	t.Run("cyclic", func(t *testing.T) {
		s, parent := newSymlinkSink(t)
		dir := s.fs.Directory

		// Cyclic links must neither loop nor be written through.
		tmust(t, s.symlink("loop", "loop"))
		assertTraversal(t, s.file("loop/escaped.txt", "x"), "file through self-referential link")
		tmust(t, s.file("loop", "plain"))
		tmust(t, s.symlink("ping", "pong"))
		tmust(t, s.symlink("pong", "ping"))
		assertTraversal(t, s.mkdir("ping/escaped-dir"), "dir through cyclic link")
		tmust(t, s.mkdir("ping"))
		st, err := os.Lstat(filepath.Join(dir, "ping"))
		tmust(t, err)
		assert.True(t, st.IsDir())

		assertNothingEscaped(t, parent)
	})
}

func Test_FolderSinkSymlinkAppBundle(t *testing.T) {
	s, parent := newSymlinkSink(t)
	dir := s.fs.Directory
	framework := "Game.app/Contents/Frameworks/SDL2.framework"

	// Layout of a macOS .app with an embedded framework, in the order
	// archivers usually emit it: real files under Versions/A, a Current
	// alias, and top-level leaf links through the alias.
	tmust(t, s.mkdir("Game.app/Contents/MacOS"))
	tmust(t, s.file("Game.app/Contents/MacOS/Game", "binary"))
	tmust(t, s.file("Game.app/Contents/Info.plist", "plist"))
	tmust(t, s.mkdir(framework+"/Versions/A/Resources"))
	tmust(t, s.file(framework+"/Versions/A/SDL2", "dylib"))
	tmust(t, s.file(framework+"/Versions/A/Resources/Info.plist", "plist"))
	tmust(t, s.symlink(framework+"/Versions/Current", "A"))
	tmust(t, s.symlink(framework+"/SDL2", "Versions/Current/SDL2"))
	tmust(t, s.symlink(framework+"/Resources", "Versions/Current/Resources"))
	tmust(t, s.symlink("Game.app/Contents/MacOS/libSDL2.dylib", "../Frameworks/SDL2.framework/SDL2"))

	bs, err := os.ReadFile(filepath.Join(dir, framework+"/SDL2"))
	tmust(t, err)
	assert.Equal(t, "dylib", string(bs))
	bs, err = os.ReadFile(filepath.Join(dir, "Game.app/Contents/MacOS/libSDL2.dylib"))
	tmust(t, err)
	assert.Equal(t, "dylib", string(bs))

	// Writing through the Current alias is the one thing refused.
	assertTraversal(t, s.file(framework+"/Versions/Current/extra", "x"), "write through Current alias")

	assertNothingEscaped(t, parent)
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
