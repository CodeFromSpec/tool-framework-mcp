// code-from-spec: ROOT/golang/tests/os/list_files@gdLK-r9BddMnFTdsPT8g_zs1wlY
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

func TestListFiles_FlatDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("a.txt", []byte("a"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("b.txt", []byte("b"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("c.txt", []byte("c"), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

func TestListFiles_Recursive(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("dir/sub/deep", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("dir/alpha.txt", []byte("a"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("dir/sub/beta.txt", []byte("b"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("dir/sub/deep/gamma.txt", []byte("g"), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "dir"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

func TestListFiles_SortedAlphabetically(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("z.txt", []byte("z"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("a.txt", []byte("a"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("m.txt", []byte("m"), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

func TestListFiles_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.Mkdir("empty", 0755); err != nil {
		t.Fatal(err)
	}

	files, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "empty"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected 0 files, got %d", len(files))
	}
}

func TestListFiles_OnlySubdirectories(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("root/sub1", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("root/sub2", 0755); err != nil {
		t.Fatal(err)
	}

	files, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "root"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected 0 files, got %d", len(files))
	}
}

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

func TestListFiles_PathTraversal(t *testing.T) {
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

func TestListFiles_SymlinkOutsideRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks not reliably supported on Windows")
	}

	tempDir := t.TempDir()
	testChdir(t, tempDir)

	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "outside.txt")
	if err := os.WriteFile(outsideFile, []byte("outside"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.Mkdir("mydir", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("mydir/regular.txt", []byte("regular"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outsideFile, "mydir/link.txt"); err != nil {
		t.Skip("symlinks not supported on this platform")
	}

	_, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "mydir"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
		t.Errorf("expected ErrResolvesOutsideRoot, got: %v", err)
	}
}

func TestListFiles_WalkError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permissions cannot prevent traversal on Windows")
	}

	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("parent/restricted", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("parent/restricted/hidden.txt", []byte("hidden"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod("parent/restricted", 0000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod("parent/restricted", 0755)
	})

	_, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "parent"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, listfiles.ErrWalkError) {
		t.Errorf("expected ErrWalkError, got: %v", err)
	}
}
