// code-from-spec: SPEC/golang/tests/os/list_files@5uDqyjFkGojJdOdQj-D3rOIaiUo
package listfiles_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/listfiles"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
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

func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

func testPaths(results []*pathutils.PathCfs) []string {
	out := make([]string, len(results))
	for i, r := range results {
		out[i] = r.Value
	}
	return out
}

func TestListFiles_TC01_FlatDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "flat/a.txt", "a")
	testWriteFile(t, "flat/b.txt", "b")
	testWriteFile(t, "flat/c.txt", "c")

	results, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "flat"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	expected := []string{"flat/a.txt", "flat/b.txt", "flat/c.txt"}
	got := testPaths(results)
	for i, e := range expected {
		if got[i] != e {
			t.Errorf("index %d: expected %q, got %q", i, e, got[i])
		}
	}
}

func TestListFiles_TC02_RecursiveDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "dir/alpha.txt", "alpha")
	testWriteFile(t, "dir/sub/beta.txt", "beta")
	testWriteFile(t, "dir/sub/deep/gamma.txt", "gamma")

	results, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "dir"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	expected := []string{"dir/alpha.txt", "dir/sub/beta.txt", "dir/sub/deep/gamma.txt"}
	got := testPaths(results)
	for i, e := range expected {
		if got[i] != e {
			t.Errorf("index %d: expected %q, got %q", i, e, got[i])
		}
	}
}

func TestListFiles_TC03_SortedAlphabetically(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "sorted/z.txt", "z")
	testWriteFile(t, "sorted/a.txt", "a")
	testWriteFile(t, "sorted/m.txt", "m")

	results, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "sorted"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	expected := []string{"sorted/a.txt", "sorted/m.txt", "sorted/z.txt"}
	got := testPaths(results)
	for i, e := range expected {
		if got[i] != e {
			t.Errorf("index %d: expected %q, got %q", i, e, got[i])
		}
	}
}

func TestListFiles_TC04_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("emptydir", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	results, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "emptydir"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty list, got %d results", len(results))
	}
}

func TestListFiles_TC05_OnlySubdirectories(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("topdir/sub1", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll("topdir/sub2", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	results, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "topdir"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty list, got %d results", len(results))
	}
}

func TestListFiles_TC06_DirectoryDoesNotExist(t *testing.T) {
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

func TestListFiles_TC07_PropagatesValidationError(t *testing.T) {
	_, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "../../outside"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got: %v", err)
	}
}

func TestListFiles_TC08_PropagatesConversionError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks require elevated privileges on Windows")
	}

	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "symlinkdir/regular.txt", "content")

	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "outside.txt")
	if err := os.WriteFile(outsideFile, []byte("outside"), 0644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}

	symlinkPath := filepath.Join(tempDir, "symlinkdir", "link.txt")
	if err := os.Symlink(outsideFile, symlinkPath); err != nil {
		t.Skipf("symlink not supported: %v", err)
	}

	_, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "symlinkdir"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
		t.Errorf("expected ErrResolvesOutsideRoot, got: %v", err)
	}
}

func TestListFiles_TC09_WalkError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permission restrictions behave differently on Windows")
	}
	if os.Getuid() == 0 {
		t.Skip("running as root; permission restrictions do not apply")
	}

	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("walkdir/restricted", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	testWriteFile(t, fmt.Sprintf("walkdir/restricted/hidden.txt"), "hidden")

	if err := os.Chmod(filepath.Join(tempDir, "walkdir/restricted"), 0000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(filepath.Join(tempDir, "walkdir/restricted"), 0755)
	})

	_, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "walkdir"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, listfiles.ErrWalkError) {
		t.Errorf("expected ErrWalkError, got: %v", err)
	}
}
