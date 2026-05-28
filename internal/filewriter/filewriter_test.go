// code-from-spec: ROOT/golang/tests/os/file_writer@THvFl5BmwZuTxKtBNrkf8VQ2h5U

package filewriter_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filewriter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

// testChdir changes the working directory to dir and restores the original on cleanup.
func testChdir(t *testing.T, dir string) {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChdir: could not get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChdir: could not chdir to %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Errorf("testChdir cleanup: could not restore working directory: %v", err)
		}
	})
}

// testPath returns a *pathutils.PathCfs for the given relative path value.
func testPath(value string) *pathutils.PathCfs {
	return &pathutils.PathCfs{Value: value}
}

func TestFileWrite_WritesContentToNewFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	cfsPath := testPath("output.txt")
	content := "hello world"

	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "output.txt"))
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}
	if string(got) != content {
		t.Errorf("file content = %q, want %q", string(got), content)
	}
}

func TestFileWrite_OverwritesExistingFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	filePath := filepath.Join(dir, "target.txt")
	if err := os.WriteFile(filePath, []byte("old"), 0o644); err != nil {
		t.Fatalf("setup: could not write initial file: %v", err)
	}

	cfsPath := testPath("target.txt")
	err := filewriter.FileWrite(cfsPath, "new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}
	if string(got) != "new" {
		t.Errorf("file content = %q, want %q", string(got), "new")
	}
}

func TestFileWrite_CreatesIntermediateDirectories(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	cfsPath := testPath("a/b/c/file.txt")
	content := "nested content"

	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, subdir := range []string{"a", "a/b", "a/b/c"} {
		info, err := os.Stat(filepath.Join(dir, filepath.FromSlash(subdir)))
		if err != nil {
			t.Errorf("directory %q not found: %v", subdir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected %q to be a directory", subdir)
		}
	}

	got, err := os.ReadFile(filepath.Join(dir, "a", "b", "c", "file.txt"))
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}
	if string(got) != content {
		t.Errorf("file content = %q, want %q", string(got), content)
	}
}

func TestFileWrite_PreservesUTF8Content(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "café 日本語 🎉"
	cfsPath := testPath("utf8.txt")

	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "utf8.txt"))
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}
	if string(got) != content {
		t.Errorf("file bytes do not match UTF-8 encoding of content")
	}
}

func TestFileWrite_PreservesLineEndings(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "alpha\r\nbeta\r\n"
	cfsPath := testPath("crlf.txt")

	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "crlf.txt"))
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}
	if string(got) != content {
		t.Errorf("file content = %q, want %q (line endings must be preserved)", string(got), content)
	}
}

func TestFileWrite_WritesEmptyContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	cfsPath := testPath("empty.txt")

	err := filewriter.FileWrite(cfsPath, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(filepath.Join(dir, "empty.txt"))
	if err != nil {
		t.Fatalf("file not found: %v", err)
	}
	if info.Size() != 0 {
		t.Errorf("file size = %d, want 0", info.Size())
	}
}

func TestFileWrite_PropagatesValidationErrors(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	cfsPath := testPath("../../outside")

	err := filewriter.FileWrite(cfsPath, "should not be written")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) && !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
		t.Errorf("expected directory traversal error, got: %v", err)
	}

	// Verify no file was created outside the temp dir.
	_, statErr := os.Stat(filepath.Join(dir, "..", "..", "outside"))
	if statErr == nil {
		t.Error("file should not have been created outside the CFS root")
	}
}

func TestFileWrite_CannotCreateDirectory(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// Create a regular file named "a" so that "a/b/file.txt" cannot be created.
	if err := os.WriteFile(filepath.Join(dir, "a"), []byte("blocking file"), 0o644); err != nil {
		t.Fatalf("setup: could not create blocking file: %v", err)
	}

	cfsPath := testPath("a/b/file.txt")

	err := filewriter.FileWrite(cfsPath, "content")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, filewriter.ErrCannotCreateDirectory) {
		t.Errorf("expected ErrCannotCreateDirectory, got: %v", err)
	}
}

func TestFileWrite_CannotWriteFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// Create a directory at the target path so writing a file there fails.
	if err := os.MkdirAll(filepath.Join(dir, "target"), 0o755); err != nil {
		t.Fatalf("setup: could not create directory: %v", err)
	}

	cfsPath := testPath("target")

	err := filewriter.FileWrite(cfsPath, "content")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, filewriter.ErrCannotWriteFile) {
		t.Errorf("expected ErrCannotWriteFile, got: %v", err)
	}
}
