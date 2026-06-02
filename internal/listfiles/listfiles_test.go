// code-from-spec: ROOT/golang/tests/os/list_files@t34DcAhMcLlFQf1HDDYLCYUugMs
package listfiles_test

import (
	"errors"
	"os"
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
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.MkdirAll("mydir", 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		if err := os.WriteFile("mydir/"+name, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	files, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "mydir"})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"mydir/a.txt", "mydir/b.txt", "mydir/c.txt"}
	if len(files) != len(want) {
		t.Fatalf("got %d files, want %d", len(files), len(want))
	}
	for i, f := range files {
		if f.Value != want[i] {
			t.Errorf("files[%d] = %q, want %q", i, f.Value, want[i])
		}
	}
}

func TestListFiles_Recursive(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.MkdirAll("dir/sub/deep", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("dir/alpha.txt", []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("dir/sub/beta.txt", []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("dir/sub/deep/gamma.txt", []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "dir"})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"dir/alpha.txt", "dir/sub/beta.txt", "dir/sub/deep/gamma.txt"}
	if len(files) != len(want) {
		t.Fatalf("got %d files, want %d: %v", len(files), len(want), files)
	}
	for i, f := range files {
		if f.Value != want[i] {
			t.Errorf("files[%d] = %q, want %q", i, f.Value, want[i])
		}
	}
}

func TestListFiles_SortedAlphabetically(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.MkdirAll("sortdir", 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"z.txt", "a.txt", "m.txt"} {
		if err := os.WriteFile("sortdir/"+name, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	files, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "sortdir"})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"sortdir/a.txt", "sortdir/m.txt", "sortdir/z.txt"}
	if len(files) != len(want) {
		t.Fatalf("got %d files, want %d", len(files), len(want))
	}
	for i, f := range files {
		if f.Value != want[i] {
			t.Errorf("files[%d] = %q, want %q", i, f.Value, want[i])
		}
	}
}

func TestListFiles_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.MkdirAll("emptydir", 0755); err != nil {
		t.Fatal(err)
	}

	files, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "emptydir"})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Errorf("expected empty list, got %d files", len(files))
	}
}

func TestListFiles_OnlySubdirectories(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.MkdirAll("parent/sub1", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("parent/sub2", 0755); err != nil {
		t.Fatal(err)
	}

	files, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "parent"})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Errorf("expected empty list, got %d files", len(files))
	}
}

func TestListFiles_DirectoryDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "nonexistent"})
	if !errors.Is(err, listfiles.ErrDirectoryNotFound) {
		t.Errorf("expected ErrDirectoryNotFound, got %v", err)
	}
}

func TestListFiles_PropagatesValidationErrors(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "../../outside"})
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got %v", err)
	}
}

func TestListFiles_PropagatesConversionErrors(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks not reliably supported on Windows")
	}

	dir := t.TempDir()
	testChdir(t, dir)

	outside := t.TempDir()
	if err := os.WriteFile(outside+"/external.txt", []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll("linkdir", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("linkdir/normal.txt", []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside+"/external.txt", "linkdir/symlink.txt"); err != nil {
		t.Skip("cannot create symlink:", err)
	}

	_, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "linkdir"})
	if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
		t.Errorf("expected ErrResolvesOutsideRoot, got %v", err)
	}
}

func TestListFiles_WalkError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permissions cannot prevent traversal on Windows")
	}

	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.MkdirAll("parent/restricted", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("parent/restricted/file.txt", []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod("parent/restricted", 0000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.Chmod("parent/restricted", 0755)
	})

	_, err := listfiles.ListFiles(&pathutils.PathCfs{Value: "parent"})
	if !errors.Is(err, listfiles.ErrWalkError) {
		t.Errorf("expected ErrWalkError, got %v", err)
	}
}
