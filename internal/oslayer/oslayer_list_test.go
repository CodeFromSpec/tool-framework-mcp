// code-from-spec: SPEC/golang/tests/oslayer/list@0HLYFmnyPoU1kOq2kpR0BhUqe3c
package oslayer_test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
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

func TestListAllFiles_FlatDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("a.txt", []byte("a"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile("b.txt", []byte("b"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile("c.txt", []byte("c"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := oslayer.ListAllFiles(".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 files, got %d: %v", len(got), got)
	}
	expected := []oslayer.CfsPath{"a.txt", "b.txt", "c.txt"}
	for i, p := range expected {
		if got[i] != p {
			t.Errorf("got[%d] = %q, want %q", i, got[i], p)
		}
	}
}

func TestListAllFiles_Recursive(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("dir/sub/deep", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile("dir/alpha.txt", []byte("a"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile("dir/sub/beta.txt", []byte("b"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile("dir/sub/deep/gamma.txt", []byte("g"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := oslayer.ListAllFiles("dir")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 files, got %d: %v", len(got), got)
	}
	expected := []oslayer.CfsPath{"dir/alpha.txt", "dir/sub/beta.txt", "dir/sub/deep/gamma.txt"}
	for i, p := range expected {
		if got[i] != p {
			t.Errorf("got[%d] = %q, want %q", i, got[i], p)
		}
	}
}

func TestListAllFiles_SortedAlphabetically(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("z.txt", []byte("z"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile("a.txt", []byte("a"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile("m.txt", []byte("m"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := oslayer.ListAllFiles(".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 files, got %d: %v", len(got), got)
	}
	expected := []oslayer.CfsPath{"a.txt", "m.txt", "z.txt"}
	for i, p := range expected {
		if got[i] != p {
			t.Errorf("got[%d] = %q, want %q", i, got[i], p)
		}
	}
}

func TestListAllFiles_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("emptydir", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	got, err := oslayer.ListAllFiles("emptydir")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty list, got %v", got)
	}
}

func TestListAllFiles_HiddenFilesIncluded(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile(".hidden", []byte("h"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile("visible.txt", []byte("v"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := oslayer.ListAllFiles(".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(got), got)
	}
	expected := []oslayer.CfsPath{".hidden", "visible.txt"}
	for i, p := range expected {
		if got[i] != p {
			t.Errorf("got[%d] = %q, want %q", i, got[i], p)
		}
	}
}

func TestListAllFiles_SymlinkToFileWithinRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks not reliably supported on this platform")
	}

	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("real.txt", []byte("r"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.Symlink(filepath.Join(tempDir, "real.txt"), filepath.Join(tempDir, "link.txt")); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	got, err := oslayer.ListAllFiles(".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(got), got)
	}
	expected := []oslayer.CfsPath{"link.txt", "real.txt"}
	for i, p := range expected {
		if got[i] != p {
			t.Errorf("got[%d] = %q, want %q", i, got[i], p)
		}
	}
}

func TestListAllFiles_OnlySubdirectories(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("dir/sub1", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.MkdirAll("dir/sub2", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	got, err := oslayer.ListAllFiles("dir")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty list, got %v", got)
	}
}

func TestListAllFiles_DirectoryDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := oslayer.ListAllFiles("nonexistent/dir")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, oslayer.ErrDirectoryNotFound) {
		t.Errorf("expected ErrDirectoryNotFound, got: %v", err)
	}
}

func TestListAllFiles_InvalidCfsPath(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := oslayer.ListAllFiles("../../outside")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, oslayer.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got: %v", err)
	}
}

func TestListAllFiles_SymlinkOutsideRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks not reliably supported on this platform")
	}

	tempDir := t.TempDir()
	testChdir(t, tempDir)

	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "outside.txt")
	if err := os.WriteFile(outsideFile, []byte("o"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := os.MkdirAll("dir", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile("dir/regular.txt", []byte("r"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.Symlink(outsideFile, filepath.Join(tempDir, "dir/link.txt")); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	_, err := oslayer.ListAllFiles("dir")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, oslayer.ErrResolvesOutsideRoot) {
		t.Errorf("expected ErrResolvesOutsideRoot, got: %v", err)
	}
}

func TestListAllFiles_WalkError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permissions cannot reliably prevent traversal on Windows")
	}

	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("dir/restricted", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile("dir/restricted/file.txt", []byte("f"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.Chmod("dir/restricted", 0000); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(filepath.Join(tempDir, "dir/restricted"), 0755)
	})

	_, err := oslayer.ListAllFiles("dir")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, oslayer.ErrWalkError) {
		t.Errorf("expected ErrWalkError, got: %v", err)
	}
}
