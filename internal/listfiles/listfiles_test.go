// code-from-spec: ROOT/golang/tests/os/list_files@JRFnBuC-27_1xmlPihKytK5Bayg

package listfiles_test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/listfiles"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// testChdir changes the working directory to dir and restores it on cleanup.
func testChdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChdir: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("testChdir cleanup: %v", err)
		}
	})
}

// testMakeDir creates a directory (and any parents) relative to the current
// working directory.
func testMakeDir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("testMakeDir: %v", err)
	}
}

// testWriteFile creates a file with the given content at path (relative to cwd).
func testWriteFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

// testPathCfs constructs a PathCfs for use in tests.
func testPathCfs(value string) *pathutils.PathCfs {
	return &pathutils.PathCfs{Value: value}
}

// testResultValues extracts the Value strings from a slice of PathCfs pointers.
func testResultValues(paths []*pathutils.PathCfs) []string {
	out := make([]string, len(paths))
	for i, p := range paths {
		out[i] = p.Value
	}
	return out
}

// TC-01: Lists files in a flat directory.
func TestListFiles_FlatDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "mydir/a.txt", []byte("a"))
	testWriteFile(t, "mydir/b.txt", []byte("b"))
	testWriteFile(t, "mydir/c.txt", []byte("c"))

	result, err := listfiles.ListFiles(testPathCfs("mydir"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 files, got %d", len(result))
	}

	got := testResultValues(result)
	want := []string{
		filepath.ToSlash("mydir/a.txt"),
		filepath.ToSlash("mydir/b.txt"),
		filepath.ToSlash("mydir/c.txt"),
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("result[%d]: got %q, want %q", i, got[i], w)
		}
	}
}

// TC-02: Lists files recursively.
func TestListFiles_Recursive(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "dir/alpha.txt", []byte("a"))
	testWriteFile(t, "dir/sub/beta.txt", []byte("b"))
	testWriteFile(t, "dir/sub/deep/gamma.txt", []byte("g"))

	result, err := listfiles.ListFiles(testPathCfs("dir"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 files, got %d", len(result))
	}

	got := testResultValues(result)
	want := []string{
		"dir/alpha.txt",
		"dir/sub/beta.txt",
		"dir/sub/deep/gamma.txt",
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("result[%d]: got %q, want %q", i, got[i], w)
		}
	}
}

// TC-03: Results are sorted alphabetically.
func TestListFiles_SortedAlphabetically(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create files in a non-alphabetical order to verify sorting.
	testWriteFile(t, "sortdir/z.txt", []byte("z"))
	testWriteFile(t, "sortdir/a.txt", []byte("a"))
	testWriteFile(t, "sortdir/m.txt", []byte("m"))

	result, err := listfiles.ListFiles(testPathCfs("sortdir"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 files, got %d", len(result))
	}

	got := testResultValues(result)
	want := []string{
		"sortdir/a.txt",
		"sortdir/m.txt",
		"sortdir/z.txt",
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("result[%d]: got %q, want %q", i, got[i], w)
		}
	}
}

// TC-04: Empty directory returns an empty list.
func TestListFiles_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testMakeDir(t, "emptydir")

	result, err := listfiles.ListFiles(testPathCfs("emptydir"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty list, got %d entries", len(result))
	}
}

// TC-05: Directory with only subdirectories returns an empty list.
func TestListFiles_OnlySubdirectories(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testMakeDir(t, "dirsonly/sub1")
	testMakeDir(t, "dirsonly/sub2/nested")

	result, err := listfiles.ListFiles(testPathCfs("dirsonly"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty list, got %d entries", len(result))
	}
}

// TC-06: Directory does not exist returns ErrDirectoryNotFound.
func TestListFiles_DirectoryNotFound(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := listfiles.ListFiles(testPathCfs("nonexistent/path"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, listfiles.ErrDirectoryNotFound) {
		t.Errorf("expected ErrDirectoryNotFound, got: %v", err)
	}
}

// TC-07: Propagates validation errors from PathCfsToOs (directory traversal).
func TestListFiles_DirectoryTraversal(t *testing.T) {
	_, err := listfiles.ListFiles(testPathCfs("../../outside"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got: %v", err)
	}
}

// TC-08: Propagates conversion errors from PathOsToCfs (resolves outside root).
// Skipped on platforms where symlinks are not supported.
func TestListFiles_ResolvesOutsideRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping: symlink creation may require elevated privileges on Windows")
	}

	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create a directory and a regular file inside it.
	testWriteFile(t, "symlinkdir/real.txt", []byte("real"))

	// Create a target file outside the project root (i.e., outside tempDir).
	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "outside.txt")
	if err := os.WriteFile(outsideFile, []byte("outside"), 0644); err != nil {
		t.Fatalf("creating outside file: %v", err)
	}

	// Create a symlink inside symlinkdir that points to the outside file.
	symlinkPath := filepath.Join(tempDir, "symlinkdir", "link.txt")
	if err := os.Symlink(outsideFile, symlinkPath); err != nil {
		t.Skipf("skipping: could not create symlink: %v", err)
	}

	_, err := listfiles.ListFiles(testPathCfs("symlinkdir"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
		t.Errorf("expected ErrResolvesOutsideRoot, got: %v", err)
	}
}

// TC-09: Walk error when subdirectory permissions prevent traversal.
// Skipped on platforms where directory permissions cannot prevent traversal.
func TestListFiles_WalkError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping: directory permissions may not prevent traversal on Windows")
	}
	if os.Getuid() == 0 {
		t.Skip("skipping: root user bypasses permission checks")
	}

	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create a subdirectory whose contents cannot be read.
	testMakeDir(t, "walkdir/restricted")
	testWriteFile(t, "walkdir/restricted/secret.txt", []byte("secret"))

	// Remove read and execute permissions from the subdirectory.
	if err := os.Chmod(filepath.Join(tempDir, "walkdir", "restricted"), 0000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() {
		// Restore permissions so TempDir cleanup can remove the directory.
		_ = os.Chmod(filepath.Join(tempDir, "walkdir", "restricted"), 0755)
	})

	_, err := listfiles.ListFiles(testPathCfs("walkdir"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, listfiles.ErrWalk) {
		t.Errorf("expected ErrWalk, got: %v", err)
	}
}
