// code-from-spec: ROOT/golang/tests/os/file_writer@MaIMf_zDoAmECTHa3wB3NYTC7KM
package filewriter_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filewriter"
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

func TestFileWrite_WritesContentToNewFile(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	cfsPath := &pathutils.PathCfs{Value: "newfile.txt"}
	content := "hello world"

	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile("newfile.txt")
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}
	if string(got) != content {
		t.Errorf("content mismatch: got %q, want %q", string(got), content)
	}
}

func TestFileWrite_OverwritesExistingFile(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	if err := os.WriteFile("existing.txt", []byte("old"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "existing.txt"}

	err := filewriter.FileWrite(cfsPath, "new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile("existing.txt")
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}
	if string(got) != "new" {
		t.Errorf("content mismatch: got %q, want %q", string(got), "new")
	}
}

func TestFileWrite_CreatesIntermediateDirectories(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	cfsPath := &pathutils.PathCfs{Value: "a/b/c/file.txt"}

	err := filewriter.FileWrite(cfsPath, "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join("a", "b", "c", "file.txt"))
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("content mismatch: got %q, want %q", string(got), "hello")
	}
}

func TestFileWrite_PreservesUTF8Content(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	cfsPath := &pathutils.PathCfs{Value: "utf8.txt"}
	content := "café 日本語 🎉"

	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile("utf8.txt")
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}
	if string(got) != content {
		t.Errorf("content mismatch: got %q, want %q", string(got), content)
	}
}

func TestFileWrite_PreservesLineEndingsAsReceived(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	cfsPath := &pathutils.PathCfs{Value: "crlf.txt"}
	content := "alpha\r\nbeta\r\n"

	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile("crlf.txt")
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}
	if string(got) != content {
		t.Errorf("line endings not preserved: got %q, want %q", string(got), content)
	}
}

func TestFileWrite_WritesEmptyContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	cfsPath := &pathutils.PathCfs{Value: "empty.txt"}

	err := filewriter.FileWrite(cfsPath, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat("empty.txt")
	if err != nil {
		t.Fatalf("file does not exist: %v", err)
	}
	if info.Size() != 0 {
		t.Errorf("expected empty file, got size %d", info.Size())
	}
}

func TestFileWrite_PropagatesValidationErrors(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	cfsPath := &pathutils.PathCfs{Value: "../../outside"}

	err := filewriter.FileWrite(cfsPath, "content")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Verify no file was created at the traversal target.
	// The error should mention directory traversal.
	if !containsSubstring(err.Error(), "traversal") {
		t.Errorf("expected directory traversal error, got: %v", err)
	}
}

func TestFileWrite_CannotCreateDirectory(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	// Create a regular file named "a" so it conflicts with the directory "a/".
	if err := os.WriteFile("a", []byte("block"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "a/b/file.txt"}

	err := filewriter.FileWrite(cfsPath, "content")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filewriter.ErrCannotCreateDirectory) {
		t.Errorf("expected ErrCannotCreateDirectory, got: %v", err)
	}
}

func TestFileWrite_CannotWriteFile(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	// Create a directory at the target path so writing to it as a file fails.
	if err := os.Mkdir("target", 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "target"}

	err := filewriter.FileWrite(cfsPath, "content")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filewriter.ErrCannotWriteFile) {
		t.Errorf("expected ErrCannotWriteFile, got: %v", err)
	}
}

// containsSubstring reports whether s contains substr.
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
