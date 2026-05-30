// code-from-spec: ROOT/golang/tests/os/list_files@ZFekZcmQzdG1uyfnjQ36SZECkQs
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

// testChdir changes the working directory to dir for the duration of the test.
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

// testWriteFile creates a file at the given path (relative to cwd), creating
// parent directories as needed, and writes content to it.
func testWriteFile(t *testing.T, relPath string, content []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(relPath), 0o755); err != nil {
		t.Fatalf("testWriteFile: MkdirAll: %v", err)
	}
	if err := os.WriteFile(relPath, content, 0o644); err != nil {
		t.Fatalf("testWriteFile: WriteFile: %v", err)
	}
}

// testMkdir creates a directory at the given path relative to cwd.
func testMkdir(t *testing.T, relPath string) {
	t.Helper()
	if err := os.MkdirAll(relPath, 0o755); err != nil {
		t.Fatalf("testMkdir: %v", err)
	}
}

// testPathCfs returns a PathCfs with the given forward-slash value.
func testPathCfs(value string) *pathutils.PathCfs {
	return &pathutils.PathCfs{Value: value}
}

// TC-01: Lists files in a flat directory.
func TestListFiles_FlatDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	base := "flatdir"
	testWriteFile(t, filepath.Join(base, "a.txt"), []byte("a"))
	testWriteFile(t, filepath.Join(base, "b.txt"), []byte("b"))
	testWriteFile(t, filepath.Join(base, "c.txt"), []byte("c"))

	result, err := listfiles.ListFiles(testPathCfs("flatdir"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 files, got %d", len(result))
	}

	want := []string{
		"flatdir/a.txt",
		"flatdir/b.txt",
		"flatdir/c.txt",
	}
	for i, w := range want {
		if result[i].Value != w {
			t.Errorf("result[%d]: got %q, want %q", i, result[i].Value, w)
		}
	}
}

// TC-02: Lists files recursively.
func TestListFiles_RecursiveDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	base := "recdir"
	testWriteFile(t, filepath.Join(base, "alpha.txt"), []byte("alpha"))
	testWriteFile(t, filepath.Join(base, "sub", "beta.txt"), []byte("beta"))
	testWriteFile(t, filepath.Join(base, "sub", "deep", "gamma.txt"), []byte("gamma"))

	result, err := listfiles.ListFiles(testPathCfs("recdir"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 files, got %d", len(result))
	}

	want := []string{
		"recdir/alpha.txt",
		"recdir/sub/beta.txt",
		"recdir/sub/deep/gamma.txt",
	}
	for i, w := range want {
		if result[i].Value != w {
			t.Errorf("result[%d]: got %q, want %q", i, result[i].Value, w)
		}
	}
}

// TC-03: Results are sorted alphabetically.
func TestListFiles_SortedAlphabetically(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	base := "sortdir"
	// Create in reverse order to verify sorting is not insertion-order dependent.
	testWriteFile(t, filepath.Join(base, "z.txt"), []byte("z"))
	testWriteFile(t, filepath.Join(base, "a.txt"), []byte("a"))
	testWriteFile(t, filepath.Join(base, "m.txt"), []byte("m"))

	result, err := listfiles.ListFiles(testPathCfs("sortdir"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 files, got %d", len(result))
	}

	want := []string{
		"sortdir/a.txt",
		"sortdir/m.txt",
		"sortdir/z.txt",
	}
	for i, w := range want {
		if result[i].Value != w {
			t.Errorf("result[%d]: got %q, want %q", i, result[i].Value, w)
		}
	}
}

// TC-04: Empty directory returns empty list without error.
func TestListFiles_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	base := "emptydir"
	testMkdir(t, base)

	result, err := listfiles.ListFiles(testPathCfs("emptydir"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty list, got %d entries", len(result))
	}
}

// TC-05: Directory with only subdirectories returns empty list without error.
func TestListFiles_OnlySubdirectories(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	base := "subdirsonly"
	testMkdir(t, filepath.Join(base, "sub1"))
	testMkdir(t, filepath.Join(base, "sub2", "nested"))

	result, err := listfiles.ListFiles(testPathCfs("subdirsonly"))
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

	_, err := listfiles.ListFiles(testPathCfs("nonexistent/directory"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, listfiles.ErrDirectoryNotFound) {
		t.Errorf("expected ErrDirectoryNotFound, got: %v", err)
	}
}

// TC-07: Invalid PathCfs value (directory traversal) propagates validation error.
func TestListFiles_InvalidPathTraversal(t *testing.T) {
	_, err := listfiles.ListFiles(testPathCfs("../../outside"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// The error message should contain "directory traversal" or similar sentinel
	// propagated from PathCfsToOs.
	if !containsTraversalError(err) {
		t.Errorf("expected directory traversal error, got: %v", err)
	}
}

// containsTraversalError checks whether an error relates to directory traversal.
func containsTraversalError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	for _, substr := range []string{"traversal", "outside", "invalid"} {
		if containsSubstring(msg, substr) {
			return true
		}
	}
	return false
}

// containsSubstring checks if s contains substr (case-insensitive).
func containsSubstring(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			cs := s[i+j]
			ct := substr[j]
			if cs >= 'A' && cs <= 'Z' {
				cs += 'a' - 'A'
			}
			if ct >= 'A' && ct <= 'Z' {
				ct += 'a' - 'A'
			}
			if cs != ct {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// TC-08: Propagates conversion errors from PathOsToCfs (symlink outside root).
func TestListFiles_SymlinkOutsideRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping: symlinks may require elevated privileges on Windows")
	}

	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create an outside target directory (a separate temp dir not under cwd).
	outsideTarget := t.TempDir()

	base := "symlinkdir"
	testWriteFile(t, filepath.Join(base, "regular.txt"), []byte("regular"))

	// Create symlink inside base pointing to a directory outside the project root.
	symlinkPath := filepath.Join(base, "link_outside")
	if err := os.Symlink(outsideTarget, symlinkPath); err != nil {
		t.Skipf("skipping: could not create symlink: %v", err)
	}

	_, err := listfiles.ListFiles(testPathCfs("symlinkdir"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Expect an error related to resolving outside root, propagated from PathOsToCfs.
	if !containsSubstring(err.Error(), "outside") && !containsSubstring(err.Error(), "traversal") && !containsSubstring(err.Error(), "root") {
		t.Errorf("expected 'outside root' error, got: %v", err)
	}
}

// TC-09: Walk error on unreadable directory.
func TestListFiles_WalkError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping: directory permissions may not prevent traversal on Windows")
	}
	if os.Getuid() == 0 {
		t.Skip("skipping: running as superuser, permissions are not enforced")
	}

	tempDir := t.TempDir()
	testChdir(t, tempDir)

	base := "walkbase"
	restricted := filepath.Join(base, "restricted")
	testWriteFile(t, filepath.Join(restricted, "secret.txt"), []byte("secret"))

	// Remove read and execute permissions from the restricted directory.
	if err := os.Chmod(restricted, 0o000); err != nil {
		t.Fatalf("could not chmod restricted dir: %v", err)
	}
	t.Cleanup(func() {
		// Restore permissions so TempDir cleanup can delete the directory.
		_ = os.Chmod(restricted, 0o755)
	})

	_, err := listfiles.ListFiles(testPathCfs("walkbase"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, listfiles.ErrWalkError) {
		t.Errorf("expected ErrWalkError, got: %v", err)
	}
}
