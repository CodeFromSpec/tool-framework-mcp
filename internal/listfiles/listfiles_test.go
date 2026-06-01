// code-from-spec: ROOT/golang/tests/os/list_files@QMm75BJzL7HqJ46y_YItQ3Ga6-E
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

// testMakeDir creates a directory (and all parents) under the working directory.
func testMakeDir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("testMakeDir: %v", err)
	}
}

// testWriteFile creates a file with the given content under the working directory.
func testWriteFile(t *testing.T, path string, content []byte) {
	t.Helper()
	dir := filepath.Dir(path)
	if dir != "." {
		testMakeDir(t, dir)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

// testCfsValues extracts the string values from a slice of PathCfs pointers.
func testCfsValues(paths []*pathutils.PathCfs) []string {
	vals := make([]string, len(paths))
	for i, p := range paths {
		vals[i] = p.Value
	}
	return vals
}

// testStringSliceEqual checks that two string slices have equal length and contents.
func testStringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestListFiles_FlatDirectory verifies that ListFiles returns files in a flat
// directory in alphabetical order.
func TestListFiles_FlatDirectory(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "a.txt", []byte("a"))
	testWriteFile(t, "b.txt", []byte("b"))
	testWriteFile(t, "c.txt", []byte("c"))

	result, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 files, got %d", len(result))
	}
	got := testCfsValues(result)
	want := []string{"a.txt", "b.txt", "c.txt"}
	if !testStringSliceEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

// TestListFiles_Recursive verifies that ListFiles descends into subdirectories.
func TestListFiles_Recursive(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "dir/alpha.txt", []byte("alpha"))
	testWriteFile(t, "dir/sub/beta.txt", []byte("beta"))
	testWriteFile(t, "dir/sub/deep/gamma.txt", []byte("gamma"))

	result, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "dir"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 files, got %d", len(result))
	}
	got := testCfsValues(result)
	want := []string{"dir/alpha.txt", "dir/sub/beta.txt", "dir/sub/deep/gamma.txt"}
	if !testStringSliceEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

// TestListFiles_SortedAlphabetically verifies that results are sorted regardless
// of the order files are created.
func TestListFiles_SortedAlphabetically(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	// Create in non-alphabetical order.
	testWriteFile(t, "flat/z.txt", []byte("z"))
	testWriteFile(t, "flat/a.txt", []byte("a"))
	testWriteFile(t, "flat/m.txt", []byte("m"))

	result, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "flat"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 files, got %d", len(result))
	}
	got := testCfsValues(result)
	want := []string{"flat/a.txt", "flat/m.txt", "flat/z.txt"}
	if !testStringSliceEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

// TestListFiles_EmptyDirectory verifies that ListFiles returns an empty slice
// for a directory with no files.
func TestListFiles_EmptyDirectory(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMakeDir(t, "empty")

	result, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "empty"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %v", testCfsValues(result))
	}
}

// TestListFiles_OnlySubdirectories verifies that ListFiles returns an empty
// slice when the directory tree contains only subdirectories and no files.
func TestListFiles_OnlySubdirectories(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMakeDir(t, "parent/child1")
	testMakeDir(t, "parent/child2")

	result, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "parent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %v", testCfsValues(result))
	}
}

// TestListFiles_DirectoryNotFound verifies that ListFiles returns
// ErrDirectoryNotFound for a path that does not exist.
func TestListFiles_DirectoryNotFound(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "nonexistent/dir"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, listfiles.ErrDirectoryNotFound) {
		t.Errorf("expected ErrDirectoryNotFound, got: %v", err)
	}
}

// TestListFiles_PropagatesDirectoryTraversal verifies that ListFiles propagates
// the ErrDirectoryTraversal error from PathUtils when given a path that escapes
// the project root.
func TestListFiles_PropagatesDirectoryTraversal(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "../../outside"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got: %v", err)
	}
}

// TestListFiles_PropagatesResolvesOutsideRoot verifies that ListFiles propagates
// ErrResolvesOutsideRoot when a symlink inside the directory points outside the
// project root.
func TestListFiles_PropagatesResolvesOutsideRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks require elevated privileges on Windows")
	}

	tmp := t.TempDir()
	testChdir(t, tmp)

	testMakeDir(t, "symlinkdir")
	// Create a target file outside the project root (use os.TempDir which is
	// guaranteed to be outside our temp working dir).
	outside := filepath.Join(os.TempDir(), "outside_target.txt")
	if err := os.WriteFile(outside, []byte("outside"), 0o644); err != nil {
		t.Fatalf("failed to create outside target: %v", err)
	}
	t.Cleanup(func() { os.Remove(outside) })

	// Create a regular file so the directory is not empty.
	testWriteFile(t, "symlinkdir/regular.txt", []byte("regular"))

	// Create a symlink pointing outside the project root.
	if err := os.Symlink(outside, "symlinkdir/link.txt"); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	_, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "symlinkdir"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
		t.Errorf("expected ErrResolvesOutsideRoot, got: %v", err)
	}
}

// TestListFiles_WalkError verifies that ListFiles returns ErrWalkError when a
// subdirectory cannot be read due to permission restrictions.
func TestListFiles_WalkError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permission restrictions behave differently on Windows")
	}
	if os.Getuid() == 0 {
		t.Skip("running as root; permission restrictions do not apply")
	}

	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "walkdir/sub/file.txt", []byte("content"))

	// Deny read access to the subdirectory.
	subPath := filepath.Join(tmp, "walkdir", "sub")
	if err := os.Chmod(subPath, 0o000); err != nil {
		t.Fatalf("failed to chmod subdir: %v", err)
	}
	// Restore permissions for cleanup.
	t.Cleanup(func() {
		os.Chmod(subPath, 0o755) //nolint:errcheck
	})

	_, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "walkdir"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, listfiles.ErrWalkError) {
		t.Errorf("expected ErrWalkError, got: %v", err)
	}
}
