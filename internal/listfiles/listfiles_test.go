// code-from-spec: ROOT/golang/tests/os/list_files@sXQHM0Mdnduw-shmJIeKku-E3_k
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

// testChdir changes the working directory to dir and registers a cleanup
// to restore the original directory.
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

// testWriteFile creates a file at path (relative to cwd), creating parent
// directories as needed, and writes content to it.
func testWriteFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

// testPathValues extracts the Value strings from a slice of PathCfs.
func testPathValues(paths []*pathutils.PathCfs) []string {
	result := make([]string, len(paths))
	for i, p := range paths {
		result[i] = p.Value
	}
	return result
}

// TC-01: Lists files in a flat directory.
func TestListFiles_FlatDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	base := "base"
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatalf("mkdir base: %v", err)
	}
	testWriteFile(t, filepath.Join(base, "a.txt"), []byte("a"))
	testWriteFile(t, filepath.Join(base, "b.txt"), []byte("b"))
	testWriteFile(t, filepath.Join(base, "c.txt"), []byte("c"))

	result, err := listfiles.ListFiles(&pathutils.PathCfs{Value: base})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 files, got %d", len(result))
	}

	want := []string{
		filepath.ToSlash(filepath.Join(base, "a.txt")),
		filepath.ToSlash(filepath.Join(base, "b.txt")),
		filepath.ToSlash(filepath.Join(base, "c.txt")),
	}
	got := testPathValues(result)
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("result[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TC-02: Lists files recursively.
func TestListFiles_Recursive(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	base := "base"
	testWriteFile(t, filepath.Join(base, "alpha.txt"), []byte("alpha"))
	testWriteFile(t, filepath.Join(base, "sub", "beta.txt"), []byte("beta"))
	testWriteFile(t, filepath.Join(base, "sub", "deep", "gamma.txt"), []byte("gamma"))

	result, err := listfiles.ListFiles(&pathutils.PathCfs{Value: base})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 files, got %d", len(result))
	}

	want := []string{
		filepath.ToSlash(filepath.Join(base, "alpha.txt")),
		filepath.ToSlash(filepath.Join(base, "sub", "beta.txt")),
		filepath.ToSlash(filepath.Join(base, "sub", "deep", "gamma.txt")),
	}
	got := testPathValues(result)
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("result[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TC-03: Results are sorted alphabetically.
func TestListFiles_SortedAlphabetically(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	base := "base"
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatalf("mkdir base: %v", err)
	}
	// Create files in non-alphabetical order.
	testWriteFile(t, filepath.Join(base, "z.txt"), []byte("z"))
	testWriteFile(t, filepath.Join(base, "a.txt"), []byte("a"))
	testWriteFile(t, filepath.Join(base, "m.txt"), []byte("m"))

	result, err := listfiles.ListFiles(&pathutils.PathCfs{Value: base})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 files, got %d", len(result))
	}

	want := []string{
		filepath.ToSlash(filepath.Join(base, "a.txt")),
		filepath.ToSlash(filepath.Join(base, "m.txt")),
		filepath.ToSlash(filepath.Join(base, "z.txt")),
	}
	got := testPathValues(result)
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("result[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TC-04: Empty directory returns empty list without error.
func TestListFiles_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	base := "base"
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatalf("mkdir base: %v", err)
	}

	result, err := listfiles.ListFiles(&pathutils.PathCfs{Value: base})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty list, got %d files", len(result))
	}
}

// TC-05: Directory with only subdirectories returns empty list without error.
func TestListFiles_OnlySubdirectories(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	base := "base"
	if err := os.MkdirAll(filepath.Join(base, "sub1", "nested"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(base, "sub2"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	result, err := listfiles.ListFiles(&pathutils.PathCfs{Value: base})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty list, got %d files", len(result))
	}
}

// TC-06: Directory does not exist returns ErrDirectoryNotFound.
func TestListFiles_DirectoryNotFound(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "nonexistent-dir"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, listfiles.ErrDirectoryNotFound) {
		t.Errorf("expected ErrDirectoryNotFound, got: %v", err)
	}
}

// TC-07: Propagates validation errors from PathCfsToOs (directory traversal).
func TestListFiles_TraversalError(t *testing.T) {
	_, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "../../outside"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// The error should mention directory traversal — propagated from PathCfsToOs.
	// We check by string inspection since the sentinel is owned by pathutils.
	errStr := err.Error()
	if !containsString(errStr, "traversal") {
		t.Errorf("expected traversal error, got: %v", err)
	}
}

// TC-08: Propagates conversion errors from PathOsToCfs (resolves outside root).
func TestListFiles_SymlinkOutsideRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks require elevated privileges on Windows")
	}

	tempDir := t.TempDir()
	testChdir(t, tempDir)

	base := "base"
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatalf("mkdir base: %v", err)
	}

	// Create a regular file inside base.
	testWriteFile(t, filepath.Join(base, "regular.txt"), []byte("regular"))

	// Create a symlink inside base that points outside the project root (tempDir).
	outsideDir := t.TempDir()
	symlinkPath := filepath.Join(base, "outside_link")
	if err := os.Symlink(outsideDir, symlinkPath); err != nil {
		t.Skipf("symlink not supported: %v", err)
	}

	_, err := listfiles.ListFiles(&pathutils.PathCfs{Value: base})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errStr := err.Error()
	if !containsString(errStr, "resolves outside root") {
		t.Errorf("expected 'resolves outside root' error, got: %v", err)
	}
}

// TC-09: Walk error is returned when a subdirectory is not readable.
func TestListFiles_WalkError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permission restrictions behave differently on Windows")
	}
	if os.Getuid() == 0 {
		t.Skip("running as root; permission restrictions do not apply")
	}

	tempDir := t.TempDir()
	testChdir(t, tempDir)

	base := "base"
	restricted := filepath.Join(base, "restricted")
	if err := os.MkdirAll(restricted, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	testWriteFile(t, filepath.Join(restricted, "secret.txt"), []byte("secret"))

	// Remove read and execute permissions so the directory cannot be traversed.
	if err := os.Chmod(restricted, 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() {
		// Restore permissions so TempDir cleanup can remove the directory.
		_ = os.Chmod(restricted, 0o755)
	})

	_, err := listfiles.ListFiles(&pathutils.PathCfs{Value: base})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, listfiles.ErrWalkError) {
		t.Errorf("expected ErrWalkError, got: %v", err)
	}
}

// containsString reports whether s contains substr.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
