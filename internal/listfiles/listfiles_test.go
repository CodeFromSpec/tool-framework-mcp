// code-from-spec: ROOT/golang/tests/os/list_files@Aq9ChrZjyhKUjE0vadfl_fQHSUc
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

// TC-01: Lists files in a flat directory
func TestListFiles_FlatDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("a.txt", []byte("a"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile("b.txt", []byte("b"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile("c.txt", []byte("c"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	files, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "."})
	if err != nil {
		t.Fatalf("ListFiles: unexpected error: %v", err)
	}
	if len(files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(files))
	}

	expected := []string{"a.txt", "b.txt", "c.txt"}
	for i, f := range files {
		if f.Value != expected[i] {
			t.Errorf("files[%d]: got %q, want %q", i, f.Value, expected[i])
		}
	}
}

// TC-02: Lists files recursively
func TestListFiles_Recursive(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll(filepath.Join("dir", "sub", "deep"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join("dir", "alpha.txt"), []byte("a"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile(filepath.Join("dir", "sub", "beta.txt"), []byte("b"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile(filepath.Join("dir", "sub", "deep", "gamma.txt"), []byte("g"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	files, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "dir"})
	if err != nil {
		t.Fatalf("ListFiles: unexpected error: %v", err)
	}
	if len(files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(files))
	}

	expected := []string{"dir/alpha.txt", "dir/sub/beta.txt", "dir/sub/deep/gamma.txt"}
	for i, f := range files {
		if f.Value != expected[i] {
			t.Errorf("files[%d]: got %q, want %q", i, f.Value, expected[i])
		}
	}
}

// TC-03: Results are sorted alphabetically
func TestListFiles_SortedAlphabetically(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create files in non-alphabetical order
	if err := os.WriteFile("z.txt", []byte("z"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile("a.txt", []byte("a"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile("m.txt", []byte("m"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	files, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "."})
	if err != nil {
		t.Fatalf("ListFiles: unexpected error: %v", err)
	}
	if len(files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(files))
	}

	expected := []string{"a.txt", "m.txt", "z.txt"}
	for i, f := range files {
		if f.Value != expected[i] {
			t.Errorf("files[%d]: got %q, want %q", i, f.Value, expected[i])
		}
	}
}

// TC-04: Empty directory
func TestListFiles_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	files, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "."})
	if err != nil {
		t.Fatalf("ListFiles: unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected empty list, got %d files", len(files))
	}
}

// TC-05: Directory with only subdirectories (no files at any level)
func TestListFiles_OnlySubdirectories(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll(filepath.Join("sub1", "sub2"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	files, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "."})
	if err != nil {
		t.Fatalf("ListFiles: unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected empty list, got %d files", len(files))
	}
}

// TC-06: Directory does not exist
func TestListFiles_DirectoryNotFound(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "nonexistent/dir"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, listfiles.ErrDirectoryNotFound) {
		t.Errorf("expected ErrDirectoryNotFound, got: %v", err)
	}
}

// TC-07: Propagates validation errors from PathCfsToOs (directory traversal)
func TestListFiles_DirectoryTraversal(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "../../outside"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got: %v", err)
	}
}

// TC-08: Propagates conversion errors from PathOsToCfs (symlink outside root)
func TestListFiles_SymlinkOutsideRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks may not be supported on this platform")
	}

	// Create an external target outside the project root
	externalDir := t.TempDir()
	externalFile := filepath.Join(externalDir, "external.txt")
	if err := os.WriteFile(externalFile, []byte("external"), 0o644); err != nil {
		t.Fatalf("WriteFile external: %v", err)
	}

	projectDir := t.TempDir()
	testChdir(t, projectDir)

	if err := os.WriteFile("regular.txt", []byte("regular"), 0o644); err != nil {
		t.Fatalf("WriteFile regular: %v", err)
	}

	// Create a symlink pointing outside the project root
	if err := os.Symlink(externalFile, "symlink_outside.txt"); err != nil {
		t.Skipf("symlink creation failed (not supported?): %v", err)
	}

	_, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "."})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
		t.Errorf("expected ErrResolvesOutsideRoot, got: %v", err)
	}
}

// TC-09: Walk error (unreadable subdirectory)
func TestListFiles_WalkError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permissions may not prevent traversal on Windows")
	}

	// Skip if running as root (root can bypass directory permissions)
	if os.Getuid() == 0 {
		t.Skip("running as root; directory permissions cannot prevent traversal")
	}

	tempDir := t.TempDir()
	testChdir(t, tempDir)

	subDir := "subdir"
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "file.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Remove all permissions from the subdirectory
	if err := os.Chmod(subDir, 0o000); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() {
		// Restore permissions so TempDir cleanup can remove files
		_ = os.Chmod(subDir, 0o755)
	})

	_, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "."})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, listfiles.ErrWalkError) {
		t.Errorf("expected ErrWalkError, got: %v", err)
	}
}
