// code-from-spec: SPEC/golang/test/cases/oslayer/list@USRoA3jPYEYWeeV-SHa0fr2KmcM
package oslayerlisttest

import (
	"errors"
	"os"
	"runtime"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"
)

func TestListAllFiles_FlatDirectory(t *testing.T) {
	testutils.Chdir(t)

	if err := os.MkdirAll("flat", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		if err := os.WriteFile("flat/"+name, []byte("x"), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	results, err := oslayer.ListAllFiles("flat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 files, got %d", len(results))
	}
	expected := []oslayer.CfsPath{"flat/a.txt", "flat/b.txt", "flat/c.txt"}
	for i, e := range expected {
		if results[i] != e {
			t.Errorf("index %d: expected %q, got %q", i, e, results[i])
		}
	}
}

func TestListAllFiles_Recursive(t *testing.T) {
	testutils.Chdir(t)

	dirs := []string{"dir/sub/deep"}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	files := map[string]string{
		"dir/alpha.txt":          "x",
		"dir/sub/beta.txt":       "x",
		"dir/sub/deep/gamma.txt": "x",
	}
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	results, err := oslayer.ListAllFiles("dir")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 files, got %d", len(results))
	}
	expected := []oslayer.CfsPath{
		"dir/alpha.txt",
		"dir/sub/beta.txt",
		"dir/sub/deep/gamma.txt",
	}
	for i, e := range expected {
		if results[i] != e {
			t.Errorf("index %d: expected %q, got %q", i, e, results[i])
		}
	}
}

func TestListAllFiles_SortedAlphabetically(t *testing.T) {
	testutils.Chdir(t)

	if err := os.MkdirAll("sorted", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for _, name := range []string{"z.txt", "a.txt", "m.txt"} {
		if err := os.WriteFile("sorted/"+name, []byte("x"), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	results, err := oslayer.ListAllFiles("sorted")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 files, got %d", len(results))
	}
	expected := []oslayer.CfsPath{"sorted/a.txt", "sorted/m.txt", "sorted/z.txt"}
	for i, e := range expected {
		if results[i] != e {
			t.Errorf("index %d: expected %q, got %q", i, e, results[i])
		}
	}
}

func TestListAllFiles_EmptyDirectory(t *testing.T) {
	testutils.Chdir(t)

	if err := os.MkdirAll("empty", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	results, err := oslayer.ListAllFiles("empty")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected empty list, got %d items", len(results))
	}
}

func TestListAllFiles_HiddenFilesIncluded(t *testing.T) {
	testutils.Chdir(t)

	if err := os.MkdirAll("hidden", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for _, name := range []string{".hidden", "visible.txt"} {
		if err := os.WriteFile("hidden/"+name, []byte("x"), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	results, err := oslayer.ListAllFiles("hidden")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 files, got %d", len(results))
	}
	expected := []oslayer.CfsPath{"hidden/.hidden", "hidden/visible.txt"}
	for i, e := range expected {
		if results[i] != e {
			t.Errorf("index %d: expected %q, got %q", i, e, results[i])
		}
	}
}

func TestListAllFiles_SymlinkToFileWithinRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks not reliably supported on this platform")
	}

	testutils.Chdir(t)

	if err := os.MkdirAll("syms", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("syms/real.txt", []byte("x"), 0644); err != nil {
		t.Fatalf("write real.txt: %v", err)
	}
	if err := os.Symlink("real.txt", "syms/link.txt"); err != nil {
		t.Skipf("symlink creation failed, skipping: %v", err)
	}

	results, err := oslayer.ListAllFiles("syms")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 files, got %d", len(results))
	}
	expected := []oslayer.CfsPath{"syms/link.txt", "syms/real.txt"}
	for i, e := range expected {
		if results[i] != e {
			t.Errorf("index %d: expected %q, got %q", i, e, results[i])
		}
	}
}

func TestListAllFiles_OnlySubdirectories(t *testing.T) {
	testutils.Chdir(t)

	if err := os.MkdirAll("nodirs/a/b", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	results, err := oslayer.ListAllFiles("nodirs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected empty list, got %d items", len(results))
	}
}

func TestListAllFiles_DirectoryNotFound(t *testing.T) {
	testutils.Chdir(t)

	_, err := oslayer.ListAllFiles("nonexistent/dir")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, oslayer.ErrDirectoryNotFound) {
		t.Errorf("expected ErrDirectoryNotFound, got %v", err)
	}
}

func TestListAllFiles_InvalidCfsPath(t *testing.T) {
	testutils.Chdir(t)

	_, err := oslayer.ListAllFiles("../../outside")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, oslayer.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got %v", err)
	}
}

func TestListAllFiles_SymlinkOutsideRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks not reliably supported on this platform")
	}

	testutils.Chdir(t)

	if err := os.MkdirAll("outsidelink", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("outsidelink/regular.txt", []byte("x"), 0644); err != nil {
		t.Fatalf("write regular.txt: %v", err)
	}

	outsideDir := t.TempDir()
	outsideFile := outsideDir + "/outside.txt"
	if err := os.WriteFile(outsideFile, []byte("y"), 0644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	if err := os.Symlink(outsideFile, "outsidelink/escape.txt"); err != nil {
		t.Skipf("symlink creation failed, skipping: %v", err)
	}

	_, err := oslayer.ListAllFiles("outsidelink")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, oslayer.ErrResolvesOutsideRoot) {
		t.Errorf("expected ErrResolvesOutsideRoot, got %v", err)
	}
}

func TestListAllFiles_WalkError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permission restrictions may not prevent traversal on this platform")
	}

	testutils.Chdir(t)

	if err := os.MkdirAll("restricted/sub", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("restricted/sub/file.txt", []byte("x"), 0644); err != nil {
		t.Fatalf("write file.txt: %v", err)
	}
	if err := os.Chmod("restricted/sub", 0000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod("restricted/sub", 0755)
	})

	_, err := oslayer.ListAllFiles("restricted")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, oslayer.ErrWalkError) {
		t.Errorf("expected ErrWalkError, got %v", err)
	}
}
