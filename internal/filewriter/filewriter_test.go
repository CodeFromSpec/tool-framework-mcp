// code-from-spec: ROOT/golang/tests/os/file_writer@gAknePkypZelu0LBqnB-7Zk7CiI
package filewriter_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filewriter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// testChdir changes the working directory to dir for the duration of the test,
// restoring the original directory on cleanup.
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

// TC-1: Writes content to a new file.
func TestFileWrite_WritesContentToNewFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := &pathutils.PathCfs{Value: "newfile.txt"}
	content := "hello world"

	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got, err := os.ReadFile("newfile.txt")
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}

	if string(got) != content {
		t.Errorf("expected content %q, got %q", content, string(got))
	}
}

// TC-2: Overwrites an existing file.
func TestFileWrite_OverwritesExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("existing.txt", []byte("old"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "existing.txt"}

	err := filewriter.FileWrite(cfsPath, "new")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got, err := os.ReadFile("existing.txt")
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}

	if string(got) != "new" {
		t.Errorf("expected content %q, got %q", "new", string(got))
	}
}

// TC-3: Creates intermediate directories.
func TestFileWrite_CreatesIntermediateDirectories(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := &pathutils.PathCfs{Value: "a/b/c/file.txt"}
	content := "nested content"

	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got, err := os.ReadFile("a/b/c/file.txt")
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}

	if string(got) != content {
		t.Errorf("expected content %q, got %q", content, string(got))
	}
}

// TC-4: Preserves UTF-8 content.
func TestFileWrite_PreservesUTF8Content(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := &pathutils.PathCfs{Value: "utf8file.txt"}
	content := "café 日本語 🎉"

	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	gotBytes, err := os.ReadFile("utf8file.txt")
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}

	expected := []byte(content)
	if string(gotBytes) != string(expected) {
		t.Errorf("expected bytes %v, got %v", expected, gotBytes)
	}
}

// TC-5: Preserves line endings as received.
func TestFileWrite_PreservesLineEndings(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := &pathutils.PathCfs{Value: "crlf.txt"}
	content := "alpha\r\nbeta\r\n"

	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	gotBytes, err := os.ReadFile("crlf.txt")
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}

	expected := []byte(content)
	if string(gotBytes) != string(expected) {
		t.Errorf("expected CRLF content %q, got %q", string(expected), string(gotBytes))
	}
}

// TC-6: Writes empty content.
func TestFileWrite_WritesEmptyContent(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := &pathutils.PathCfs{Value: "empty.txt"}

	err := filewriter.FileWrite(cfsPath, "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	info, err := os.Stat("empty.txt")
	if err != nil {
		t.Fatalf("could not stat file: %v", err)
	}

	if info.Size() != 0 {
		t.Errorf("expected file size 0, got %d", info.Size())
	}
}

// TC-7: Propagates validation errors from PathCfsToOs.
func TestFileWrite_PropagatesDirectoryTraversalError(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := &pathutils.PathCfs{Value: "../../outside/file.txt"}

	err := filewriter.FileWrite(cfsPath, "some content")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if !errors.Is(err, pathutils.ErrDirectoryTraversal) && !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
		t.Errorf("expected ErrDirectoryTraversal or ErrResolvesOutsideRoot, got: %v", err)
	}

	// Verify no file was created outside the temp dir.
	_, statErr := os.Stat("../../outside/file.txt")
	if statErr == nil {
		t.Error("file should not have been created")
	}
}

// TC-8: Cannot create directory.
func TestFileWrite_CannotCreateDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create a regular file named "a" where a directory "a" would need to be created.
	if err := os.WriteFile("a", []byte("blocking file"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// FileWrite needs to create directory "a" to satisfy path "a/b/file.txt",
	// but "a" is already a file.
	cfsPath := &pathutils.PathCfs{Value: "a/b/file.txt"}

	err := filewriter.FileWrite(cfsPath, "content")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if !errors.Is(err, filewriter.ErrCannotCreateDirectory) {
		t.Errorf("expected ErrCannotCreateDirectory, got: %v", err)
	}

	// Verify no target file was created.
	_, statErr := os.Stat("a/b/file.txt")
	if statErr == nil {
		t.Error("target file should not have been created")
	}
}

// TC-9: Cannot write file.
func TestFileWrite_CannotWriteFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create a directory at the exact path where FileWrite would write a file.
	if err := os.Mkdir("mydir", 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Attempt to write to "mydir" which is actually a directory.
	cfsPath := &pathutils.PathCfs{Value: "mydir"}

	err := filewriter.FileWrite(cfsPath, "content")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if !errors.Is(err, filewriter.ErrCannotWriteFile) {
		t.Errorf("expected ErrCannotWriteFile, got: %v", err)
	}

	// Verify the directory still exists and is not modified.
	info, statErr := os.Stat("mydir")
	if statErr != nil {
		t.Fatalf("directory should still exist: %v", statErr)
	}
	if !info.IsDir() {
		t.Error("expected mydir to remain a directory")
	}
}
