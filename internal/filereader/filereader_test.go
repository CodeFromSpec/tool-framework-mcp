// code-from-spec: ROOT/golang/tests/os/file_reader@uJ8D0S-6XxN_eqC_9uwubF9GkXw

package filereader

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

// testMakeCfsPath creates a PathCfs pointing at the given absolute os path
// by writing the file into a temp dir and returning a relative path from the
// project root. Because PathCfsToOs resolves against the working directory
// (project root), we write the file in the temp dir and construct a relative
// path that os.Open can reach. Instead, we bypass the CFS path validation by
// placing files inside t.TempDir() and using a small helper that creates a
// PathCfs with a value the implementation will resolve relative to the cwd.
//
// Since the tests run with the project root as cwd and t.TempDir() is an
// absolute path outside the project root, we cannot use a raw PathCfs
// directly. We therefore write test files under a subdirectory of the
// project root's temp-equivalent by using os.MkdirTemp inside t.TempDir()
// — but that still would be absolute. The cleanest approach is to write
// files into a relative directory under the working directory.
//
// To keep it simple: write files to t.TempDir(), compute a relative path
// from cwd, and use that as the PathCfs value so PathCfsToOs resolves it.
func testMakeCfsPath(t *testing.T, content string) (*pathutils.PathCfs, string) {
	t.Helper()

	// Use a path relative to cwd. We'll create a temp subdir under the
	// working directory so that the relative path stays within the project
	// root and passes CFS validation.
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	// Create a temp directory inside the project root.
	tmpDir, err := os.MkdirTemp(cwd, "filereader_test_*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	filePath := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Compute relative path from cwd.
	rel, err := filepath.Rel(cwd, filePath)
	if err != nil {
		t.Fatalf("Rel: %v", err)
	}
	// CFS paths use forward slashes.
	rel = filepath.ToSlash(rel)

	return &pathutils.PathCfs{Value: rel}, filePath
}

// testMakeNonExistentCfsPath returns a PathCfs pointing at a path that does
// not exist but is otherwise structurally valid (relative, no traversal).
func testMakeNonExistentCfsPath(t *testing.T) *pathutils.PathCfs {
	t.Helper()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	// Create a temp dir to anchor the relative path, then remove it so the
	// target does not exist.
	tmpDir, err := os.MkdirTemp(cwd, "filereader_missing_*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	// Remove immediately — we just want the relative path segment.
	os.RemoveAll(tmpDir)

	rel, err := filepath.Rel(cwd, filepath.Join(tmpDir, "ghost.txt"))
	if err != nil {
		t.Fatalf("Rel: %v", err)
	}
	return &pathutils.PathCfs{Value: filepath.ToSlash(rel)}
}

// ---------------------------------------------------------------------------
// Happy Path
// ---------------------------------------------------------------------------

func TestFileOpen_ReadsAllLines(t *testing.T) {
	cfsPath, _ := testMakeCfsPath(t, "alpha\nbeta\ngamma\n")

	reader, err := FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	wants := []string{"alpha", "beta", "gamma"}
	for _, want := range wants {
		got, err := FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	_, err = FileReadLine(reader)
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileReadLine_NormalizesCRLF(t *testing.T) {
	cfsPath, _ := testMakeCfsPath(t, "alpha\r\nbeta\r\n")

	reader, err := FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	for _, want := range []string{"alpha", "beta"} {
		got, err := FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFileOpen_NoTrailingNewline(t *testing.T) {
	cfsPath, _ := testMakeCfsPath(t, "alpha\nbeta")

	reader, err := FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	for _, want := range []string{"alpha", "beta"} {
		got, err := FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	_, err = FileReadLine(reader)
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileSkipLines_AdvancesReader(t *testing.T) {
	cfsPath, _ := testMakeCfsPath(t, "one\ntwo\nthree\nfour\nfive\n")

	reader, err := FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	FileSkipLines(reader, 2)

	got, err := FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if got != "three" {
		t.Errorf("got %q, want %q", got, "three")
	}
}

func TestFileSkipLines_PastEndOfFile(t *testing.T) {
	cfsPath, _ := testMakeCfsPath(t, "one\ntwo\n")

	reader, err := FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	FileSkipLines(reader, 10) // should not panic or error

	_, err = FileReadLine(reader)
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileReadLine_PreservesLeadingWhitespace(t *testing.T) {
	cfsPath, _ := testMakeCfsPath(t, "  alpha\n    beta\n")

	reader, err := FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	for _, want := range []string{"  alpha", "    beta"} {
		got, err := FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFileReadLine_PreservesTrailingWhitespace(t *testing.T) {
	cfsPath, _ := testMakeCfsPath(t, "alpha  \nbeta   \n")

	reader, err := FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	for _, want := range []string{"alpha  ", "beta   "} {
		got, err := FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFileReadLine_PreservesInternalWhitespace(t *testing.T) {
	cfsPath, _ := testMakeCfsPath(t, "alpha   beta\none\ttwo\n")

	reader, err := FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	for _, want := range []string{"alpha   beta", "one\ttwo"} {
		got, err := FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFileReadLine_PreservesEmptyLines(t *testing.T) {
	cfsPath, _ := testMakeCfsPath(t, "alpha\n\n\nbeta\n")

	reader, err := FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	wants := []string{"alpha", "", "", "beta"}
	for _, want := range wants {
		got, err := FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFileReadLine_PreservesNonASCII(t *testing.T) {
	cfsPath, _ := testMakeCfsPath(t, "café\n日本語\n🎉🚀\n")

	reader, err := FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	for _, want := range []string{"café", "日本語", "🎉🚀"} {
		got, err := FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

// ---------------------------------------------------------------------------
// Edge Cases
// ---------------------------------------------------------------------------

func TestFileOpen_EmptyFile(t *testing.T) {
	cfsPath, _ := testMakeCfsPath(t, "")

	reader, err := FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	_, err = FileReadLine(reader)
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileOpen_SingleLineNoNewline(t *testing.T) {
	cfsPath, _ := testMakeCfsPath(t, "hello")

	reader, err := FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	got, err := FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}

	_, err = FileReadLine(reader)
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Failure Cases
// ---------------------------------------------------------------------------

func TestFileOpen_FileDoesNotExist(t *testing.T) {
	cfsPath := testMakeNonExistentCfsPath(t)

	_, err := FileOpen(cfsPath)
	if !errors.Is(err, ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestFileReadLine_AfterClose(t *testing.T) {
	cfsPath, _ := testMakeCfsPath(t, "alpha\n")

	reader, err := FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	FileClose(reader)

	_, err = FileReadLine(reader)
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileSkipLines_AfterClose(t *testing.T) {
	cfsPath, _ := testMakeCfsPath(t, "alpha\n")

	reader, err := FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	FileClose(reader)

	// Should do nothing and not panic.
	FileSkipLines(reader, 1)
}
