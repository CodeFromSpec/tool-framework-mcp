// code-from-spec: ROOT/golang/tests/os/list_files@mXR4qGtA5I9_BaszUhOrPhn3VTw

package listfiles_test

import (
	"errors"
	"os"
	"runtime"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/listfiles"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

// testChdir changes the working directory to dir and restores the original on cleanup.
func testChdir(t *testing.T, dir string) {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChdir: get original working dir: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChdir: chdir to %q: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Errorf("testChdir cleanup: restore working dir: %v", err)
		}
	})
}

// testWriteFile creates a file at path (relative to the current working dir)
// with the given content. Parent directories are created as needed.
func testWriteFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath_dir(path), 0o755); err != nil {
		t.Fatalf("testWriteFile: mkdirall for %q: %v", path, err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("testWriteFile: write %q: %v", path, err)
	}
}

// filepath_dir returns the directory portion of path, or "." if none.
func filepath_dir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return "."
}

// testPath constructs a PathCfs from the given value.
func testPath(value string) *pathutils.PathCfs {
	return &pathutils.PathCfs{Value: value}
}

// testExtractValues returns the Value field from each PathCfs in the slice.
func testExtractValues(paths []*pathutils.PathCfs) []string {
	result := make([]string, len(paths))
	for i, p := range paths {
		result[i] = p.Value
	}
	return result
}

// testStringsEqual returns true if a and b contain the same strings in the same order.
func testStringsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// TC-01: Lists files in a flat directory
// ---------------------------------------------------------------------------

func TestListFiles_FlatDirectory(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	dir := "mydir"
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	testWriteFile(t, "mydir/a.txt", []byte("a"))
	testWriteFile(t, "mydir/b.txt", []byte("b"))
	testWriteFile(t, "mydir/c.txt", []byte("c"))

	files, err := listfiles.ListFiles(testPath("mydir"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := testExtractValues(files)
	want := []string{"mydir/a.txt", "mydir/b.txt", "mydir/c.txt"}

	if !testStringsEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// ---------------------------------------------------------------------------
// TC-02: Lists files recursively
// ---------------------------------------------------------------------------

func TestListFiles_Recursive(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	if err := os.MkdirAll("dir/sub/deep", 0o755); err != nil {
		t.Fatalf("mkdirall: %v", err)
	}

	testWriteFile(t, "dir/alpha.txt", []byte("alpha"))
	testWriteFile(t, "dir/sub/beta.txt", []byte("beta"))
	testWriteFile(t, "dir/sub/deep/gamma.txt", []byte("gamma"))

	files, err := listfiles.ListFiles(testPath("dir"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := testExtractValues(files)
	want := []string{"dir/alpha.txt", "dir/sub/beta.txt", "dir/sub/deep/gamma.txt"}

	if !testStringsEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// ---------------------------------------------------------------------------
// TC-03: Results are sorted alphabetically
// ---------------------------------------------------------------------------

func TestListFiles_SortedAlphabetically(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	if err := os.Mkdir("sorted", 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Create files in non-alphabetical order.
	testWriteFile(t, "sorted/z.txt", []byte("z"))
	testWriteFile(t, "sorted/a.txt", []byte("a"))
	testWriteFile(t, "sorted/m.txt", []byte("m"))

	files, err := listfiles.ListFiles(testPath("sorted"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := testExtractValues(files)
	want := []string{"sorted/a.txt", "sorted/m.txt", "sorted/z.txt"}

	if !testStringsEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// ---------------------------------------------------------------------------
// TC-04: Empty directory
// ---------------------------------------------------------------------------

func TestListFiles_EmptyDirectory(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	if err := os.Mkdir("empty", 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	files, err := listfiles.ListFiles(testPath("empty"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("expected empty list, got %v", testExtractValues(files))
	}
}

// ---------------------------------------------------------------------------
// TC-05: Directory with only subdirectories
// ---------------------------------------------------------------------------

func TestListFiles_OnlySubdirectories(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	if err := os.MkdirAll("parent/sub1", 0o755); err != nil {
		t.Fatalf("mkdirall: %v", err)
	}
	if err := os.MkdirAll("parent/sub2", 0o755); err != nil {
		t.Fatalf("mkdirall: %v", err)
	}

	files, err := listfiles.ListFiles(testPath("parent"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("expected empty list, got %v", testExtractValues(files))
	}
}

// ---------------------------------------------------------------------------
// TC-06: Directory does not exist
// ---------------------------------------------------------------------------

func TestListFiles_DirectoryNotFound(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	_, err := listfiles.ListFiles(testPath("nonexistent"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, listfiles.ErrDirectoryNotFound) {
		t.Errorf("expected ErrDirectoryNotFound, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TC-07: Propagates validation errors from PathCfsToOs
// ---------------------------------------------------------------------------

func TestListFiles_PathValidationError(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	_, err := listfiles.ListFiles(testPath("../../outside"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TC-08: Propagates conversion errors from PathOsToCfs (symlink outside root)
// ---------------------------------------------------------------------------

func TestListFiles_SymlinkOutsideRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks not reliably supported on Windows")
	}

	// Create an external target outside the project root.
	external := t.TempDir()
	externalFile := external + "/outside.txt"
	if err := os.WriteFile(externalFile, []byte("outside"), 0o644); err != nil {
		t.Fatalf("write external file: %v", err)
	}

	// Set up the project root.
	projectRoot := t.TempDir()
	testChdir(t, projectRoot)

	if err := os.Mkdir("linkdir", 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Create a regular file inside the dir.
	testWriteFile(t, "linkdir/normal.txt", []byte("normal"))

	// Create a symlink inside linkdir pointing to the external file.
	if err := os.Symlink(externalFile, "linkdir/outsidelink.txt"); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	_, err := listfiles.ListFiles(testPath("linkdir"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
		t.Errorf("expected ErrResolvesOutsideRoot, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TC-09: Walk error due to unreadable subdirectory
// ---------------------------------------------------------------------------

func TestListFiles_WalkError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permissions may not prevent traversal on Windows")
	}

	root := t.TempDir()
	testChdir(t, root)

	if err := os.MkdirAll("walktest/secret", 0o755); err != nil {
		t.Fatalf("mkdirall: %v", err)
	}

	// Place a file in the secret subdir before locking it.
	testWriteFile(t, "walktest/secret/hidden.txt", []byte("hidden"))

	// Remove read+execute permissions so the directory cannot be traversed.
	if err := os.Chmod("walktest/secret", 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() {
		// Restore permissions so TempDir cleanup can remove the directory.
		_ = os.Chmod("walktest/secret", 0o755)
	})

	_, err := listfiles.ListFiles(testPath("walktest"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, listfiles.ErrWalk) {
		t.Errorf("expected ErrWalk, got: %v", err)
	}
}
