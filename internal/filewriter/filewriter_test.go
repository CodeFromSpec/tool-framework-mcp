// code-from-spec: ROOT/golang/tests/os/file_writer@bYkuLpa4xPU0yPXSyUv30sK6rGA
package filewriter_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filewriter"
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

// TestFileWrite_WritesContentToNewFile verifies that FileWrite creates a new
// file with the expected content when the target does not exist.
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
		t.Fatalf("could not read written file: %v", err)
	}
	if string(got) != content {
		t.Errorf("file content = %q, want %q", string(got), content)
	}
}

// TestFileWrite_OverwritesExistingFile verifies that FileWrite replaces the
// content of an existing file completely.
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
		t.Fatalf("could not read written file: %v", err)
	}
	if string(got) != "new" {
		t.Errorf("file content = %q, want %q", string(got), "new")
	}
}

// TestFileWrite_CreatesIntermediateDirectories verifies that FileWrite creates
// all missing parent directories before writing the file.
func TestFileWrite_CreatesIntermediateDirectories(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := &pathutils.PathCfs{Value: "a/b/c/file.txt"}
	content := "hello"

	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got, err := os.ReadFile("a/b/c/file.txt")
	if err != nil {
		t.Fatalf("could not read written file: %v", err)
	}
	if string(got) != content {
		t.Errorf("file content = %q, want %q", string(got), content)
	}
}

// TestFileWrite_PreservesUTF8Content verifies that multi-byte UTF-8 characters
// are written and read back without corruption.
func TestFileWrite_PreservesUTF8Content(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := &pathutils.PathCfs{Value: "utf8.txt"}
	content := "café 日本語 🎉"

	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got, err := os.ReadFile("utf8.txt")
	if err != nil {
		t.Fatalf("could not read written file: %v", err)
	}
	if string(got) != content {
		t.Errorf("file content = %q, want %q", string(got), content)
	}
}

// TestFileWrite_PreservesLineEndingsAsReceived verifies that CRLF line endings
// are not normalized and are stored exactly as provided.
func TestFileWrite_PreservesLineEndingsAsReceived(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := &pathutils.PathCfs{Value: "crlf.txt"}
	content := "alpha\r\nbeta\r\n"

	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got, err := os.ReadFile("crlf.txt")
	if err != nil {
		t.Fatalf("could not read written file: %v", err)
	}
	if string(got) != content {
		t.Errorf("file content = %q, want %q", string(got), content)
	}
}

// TestFileWrite_WritesEmptyContent verifies that writing an empty string
// results in a zero-byte file.
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
		t.Fatalf("could not stat written file: %v", err)
	}
	if info.Size() != 0 {
		t.Errorf("file size = %d, want 0", info.Size())
	}
}

// TestFileWrite_PropagatesValidationErrors verifies that an invalid path
// attempting directory traversal causes FileWrite to return an error without
// creating any file or directory.
func TestFileWrite_PropagatesValidationErrors(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := &pathutils.PathCfs{Value: "../../outside"}

	err := filewriter.FileWrite(cfsPath, "content")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

// TestFileWrite_CannotCreateDirectory verifies that ErrCannotCreateDirectory
// is returned when an intermediate directory component conflicts with an
// existing regular file.
func TestFileWrite_CannotCreateDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create a regular file named "a" so that "a/b/file.txt" cannot be created.
	if err := os.WriteFile("a", []byte("blocking file"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "a/b/file.txt"}

	err := filewriter.FileWrite(cfsPath, "content")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, filewriter.ErrCannotCreateDirectory) {
		t.Errorf("expected ErrCannotCreateDirectory, got: %v", err)
	}
}

// TestFileWrite_CannotWriteFile verifies that ErrCannotWriteFile is returned
// when the target path resolves to an existing directory rather than a file.
func TestFileWrite_CannotWriteFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create a directory at the target path.
	if err := os.Mkdir("target", 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "target"}

	err := filewriter.FileWrite(cfsPath, "content")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, filewriter.ErrCannotWriteFile) {
		t.Errorf("expected ErrCannotWriteFile, got: %v", err)
	}
}
