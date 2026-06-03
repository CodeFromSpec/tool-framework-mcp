// code-from-spec: ROOT/golang/tests/os/list_files@6NM1b68X0LUm3GQyARicqm2wJgs
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

	if err := os.MkdirAll("mydir", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		if err := os.WriteFile(filepath.Join("mydir", name), []byte("x"), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}

	result, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "mydir"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 files, got %d", len(result))
	}
	expected := []string{"mydir/a.txt", "mydir/b.txt", "mydir/c.txt"}
	for i, f := range result {
		if f.Value != expected[i] {
			t.Errorf("result[%d] = %q, want %q", i, f.Value, expected[i])
		}
	}
}

func TestListFiles_Recursive(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll(filepath.Join("dir", "sub", "deep"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	files := map[string]string{
		filepath.Join("dir", "alpha.txt"):            "a",
		filepath.Join("dir", "sub", "beta.txt"):      "b",
		filepath.Join("dir", "sub", "deep", "gamma.txt"): "c",
	}
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}

	result, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "dir"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 files, got %d", len(result))
	}
	expected := []string{"dir/alpha.txt", "dir/sub/beta.txt", "dir/sub/deep/gamma.txt"}
	for i, f := range result {
		if f.Value != expected[i] {
			t.Errorf("result[%d] = %q, want %q", i, f.Value, expected[i])
		}
	}
}

func TestListFiles_SortedAlphabetically(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("sortdir", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for _, name := range []string{"z.txt", "a.txt", "m.txt"} {
		if err := os.WriteFile(filepath.Join("sortdir", name), []byte("x"), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}

	result, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "sortdir"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 files, got %d", len(result))
	}
	expected := []string{"sortdir/a.txt", "sortdir/m.txt", "sortdir/z.txt"}
	for i, f := range result {
		if f.Value != expected[i] {
			t.Errorf("result[%d] = %q, want %q", i, f.Value, expected[i])
		}
	}
}

func TestListFiles_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("emptydir", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	result, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "emptydir"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty list, got %d items", len(result))
	}
}

func TestListFiles_OnlySubdirectories(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll(filepath.Join("topdir", "sub1"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join("topdir", "sub2"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	result, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "topdir"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty list, got %d items", len(result))
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
		t.Skip("symlinks may not be supported on this platform")
	}

	tempDir := t.TempDir()
	testChdir(t, tempDir)

	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "outside.txt")
	if err := os.WriteFile(outsideFile, []byte("outside"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := os.MkdirAll("linkdir", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join("linkdir", "regular.txt"), []byte("x"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.Symlink(outsideFile, filepath.Join("linkdir", "link.txt")); err != nil {
		t.Skip("symlinks not supported on this platform")
	}

	_, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "linkdir"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
		t.Errorf("expected ErrResolvesOutsideRoot, got: %v", err)
	}
}

func TestListFiles_WalkError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permission restrictions may not prevent traversal on this platform")
	}

	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll(filepath.Join("parentdir", "restricted"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join("parentdir", "restricted", "file.txt"), []byte("x"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.Chmod(filepath.Join("parentdir", "restricted"), 0000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(filepath.Join("parentdir", "restricted"), 0755)
	})

	_, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "parentdir"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, listfiles.ErrWalkError) {
		t.Errorf("expected ErrWalkError, got: %v", err)
	}
}
