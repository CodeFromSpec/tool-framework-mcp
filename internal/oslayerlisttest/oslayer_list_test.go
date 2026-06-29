// code-from-spec: SPEC/golang/tests/oslayer/list@EcOFM29uzzeF24txkMqo6_X0jjg
package oslayerlisttest_test

import (
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

	results, err := oslayer.ListAllFiles(".")
	if err != nil {
		t.Fatalf("ListAllFiles: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 files, got %d: %v", len(results), results)
	}
	expected := []oslayer.CfsPath{"a.txt", "b.txt", "c.txt"}
	for i, p := range results {
		if p != expected[i] {
			t.Errorf("results[%d] = %q, want %q", i, p, expected[i])
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
	if err := os.WriteFile("dir/sub/deep/gamma.txt", []byte("c"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	results, err := oslayer.ListAllFiles("dir")
	if err != nil {
		t.Fatalf("ListAllFiles: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 files, got %d: %v", len(results), results)
	}
	expected := []oslayer.CfsPath{"dir/alpha.txt", "dir/sub/beta.txt", "dir/sub/deep/gamma.txt"}
	for i, p := range results {
		if p != expected[i] {
			t.Errorf("results[%d] = %q, want %q", i, p, expected[i])
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

	results, err := oslayer.ListAllFiles(".")
	if err != nil {
		t.Fatalf("ListAllFiles: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 files, got %d: %v", len(results), results)
	}
	expected := []oslayer.CfsPath{"a.txt", "m.txt", "z.txt"}
	for i, p := range results {
		if p != expected[i] {
			t.Errorf("results[%d] = %q, want %q", i, p, expected[i])
		}
	}
}

func TestListAllFiles_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.Mkdir("emptydir", 0755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	results, err := oslayer.ListAllFiles("emptydir")
	if err != nil {
		t.Fatalf("ListAllFiles: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected empty list, got %d: %v", len(results), results)
	}
}

func TestListAllFiles_HiddenFilesIncluded(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile(".hidden", []byte("hidden"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile("visible.txt", []byte("visible"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	results, err := oslayer.ListAllFiles(".")
	if err != nil {
		t.Fatalf("ListAllFiles: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(results), results)
	}
	expected := []oslayer.CfsPath{".hidden", "visible.txt"}
	for i, p := range results {
		if p != expected[i] {
			t.Errorf("results[%d] = %q, want %q", i, p, expected[i])
		}
	}
}

func TestListAllFiles_SymlinkToFileWithinRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks not reliably supported on this platform")
	}

	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("real.txt", []byte("real"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.Symlink("real.txt", "link.txt"); err != nil {
		t.Skip("symlinks not supported: " + err.Error())
	}

	results, err := oslayer.ListAllFiles(".")
	if err != nil {
		t.Fatalf("ListAllFiles: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(results), results)
	}
	expected := []oslayer.CfsPath{"link.txt", "real.txt"}
	for i, p := range results {
		if p != expected[i] {
			t.Errorf("results[%d] = %q, want %q", i, p, expected[i])
		}
	}
}

func TestListAllFiles_DirectoryWithOnlySubdirectories(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("parent/child1/grandchild", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.Mkdir("parent/child2", 0755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	results, err := oslayer.ListAllFiles("parent")
	if err != nil {
		t.Fatalf("ListAllFiles: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected empty list, got %d: %v", len(results), results)
	}
}

func TestListAllFiles_DirectoryDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := oslayer.ListAllFiles("nonexistent/dir")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !isErr(err, oslayer.ErrDirectoryNotFound) {
		t.Errorf("expected ErrDirectoryNotFound, got %v", err)
	}
}

func TestListAllFiles_PropagatesValidationErrors(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := oslayer.ListAllFiles("../../outside")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !isErr(err, oslayer.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got %v", err)
	}
}

func TestListAllFiles_PropagatesConversionErrors(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks not reliably supported on this platform")
	}

	tempDir := t.TempDir()
	testChdir(t, tempDir)

	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "outside.txt")
	if err := os.WriteFile(outsideFile, []byte("outside"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := os.Mkdir("mydir", 0755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	if err := os.WriteFile("mydir/regular.txt", []byte("regular"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.Symlink(outsideFile, "mydir/link_outside.txt"); err != nil {
		t.Skip("symlinks not supported: " + err.Error())
	}

	_, err := oslayer.ListAllFiles("mydir")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !isErr(err, oslayer.ErrResolvesOutsideRoot) {
		t.Errorf("expected ErrResolvesOutsideRoot, got %v", err)
	}
}

func TestListAllFiles_WalkError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permission restrictions not effective on this platform")
	}

	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("parent/restricted", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile("parent/restricted/file.txt", []byte("data"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.Chmod("parent/restricted", 0000); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod("parent/restricted", 0755)
	})

	_, err := oslayer.ListAllFiles("parent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !isErr(err, oslayer.ErrWalkError) {
		t.Errorf("expected ErrWalkError, got %v", err)
	}
}

func isErr(err, target error) bool {
	if err == nil {
		return false
	}
	type unwrapper interface {
		Unwrap() error
	}
	type multiUnwrapper interface {
		Unwrap() []error
	}
	if err == target {
		return true
	}
	if u, ok := err.(unwrapper); ok {
		return isErr(u.Unwrap(), target)
	}
	if mu, ok := err.(multiUnwrapper); ok {
		for _, e := range mu.Unwrap() {
			if isErr(e, target) {
				return true
			}
		}
	}
	return false
}
