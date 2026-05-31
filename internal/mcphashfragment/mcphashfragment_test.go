// code-from-spec: ROOT/golang/tests/mcp_tools/hash_fragment@8XbuMC9yvzmGbA1bi4D6bmi4Mtg
package mcphashfragment_test

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcphashfragment"
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

// testHashLines computes the expected SHA-1 base64url hash of lines joined by "\n".
func testHashLines(lines ...string) string {
	h := sha1.New()
	for _, line := range lines {
		h.Write([]byte(line + "\n"))
	}
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// testWriteFile creates a file at path (relative to cwd) with the given content.
func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

// testFiveLineFile creates a temp-dir-based file with 5 known lines and returns
// the relative path. testChdir must have been called before this helper.
func testFiveLineFile(t *testing.T, name string) string {
	t.Helper()
	content := "alpha\nbravo\ncharlie\ndelta\necho\n"
	testWriteFile(t, name, content)
	return name
}

// ---------------------------------------------------------------------------
// Happy path tests
// ---------------------------------------------------------------------------

func TestMCPHashFragment_ValidRange(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	path := testFiveLineFile(t, "file.txt")

	result, err := mcphashfragment.MCPHashFragment(path, "2-4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 27 {
		t.Errorf("expected 27-char string, got %d chars: %q", len(result), result)
	}
	expected := testHashLines("bravo", "charlie", "delta")
	if result != expected {
		t.Errorf("hash mismatch: got %q, want %q", result, expected)
	}
}

func TestMCPHashFragment_SingleLine(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	path := testFiveLineFile(t, "file.txt")

	result, err := mcphashfragment.MCPHashFragment(path, "3-3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 27 {
		t.Errorf("expected 27-char string, got %d chars: %q", len(result), result)
	}
	expected := testHashLines("charlie")
	if result != expected {
		t.Errorf("hash mismatch: got %q, want %q", result, expected)
	}
}

func TestMCPHashFragment_FirstLine(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	path := testFiveLineFile(t, "file.txt")

	result, err := mcphashfragment.MCPHashFragment(path, "1-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 27 {
		t.Errorf("expected 27-char string, got %d chars: %q", len(result), result)
	}
	expected := testHashLines("alpha")
	if result != expected {
		t.Errorf("hash mismatch: got %q, want %q", result, expected)
	}
}

func TestMCPHashFragment_LastLine(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	path := testFiveLineFile(t, "file.txt")

	result, err := mcphashfragment.MCPHashFragment(path, "5-5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 27 {
		t.Errorf("expected 27-char string, got %d chars: %q", len(result), result)
	}
	expected := testHashLines("echo")
	if result != expected {
		t.Errorf("hash mismatch: got %q, want %q", result, expected)
	}
}

func TestMCPHashFragment_Deterministic(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	path := testFiveLineFile(t, "file.txt")

	result1, err := mcphashfragment.MCPHashFragment(path, "1-3")
	if err != nil {
		t.Fatalf("first call unexpected error: %v", err)
	}
	result2, err := mcphashfragment.MCPHashFragment(path, "1-3")
	if err != nil {
		t.Fatalf("second call unexpected error: %v", err)
	}
	if result1 != result2 {
		t.Errorf("results differ: %q vs %q", result1, result2)
	}
}

// ---------------------------------------------------------------------------
// Error case tests
// ---------------------------------------------------------------------------

func TestMCPHashFragment_FileNotExist(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := mcphashfragment.MCPHashFragment("nonexistent.go", "1-5")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestMCPHashFragment_InvalidRangeFormat(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	path := testFiveLineFile(t, "file.txt")

	_, err := mcphashfragment.MCPHashFragment(path, "abc")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcphashfragment.ErrInvalidLineRange) {
		t.Errorf("expected ErrInvalidLineRange, got: %v", err)
	}
}

func TestMCPHashFragment_StartGreaterThanEnd(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	path := testFiveLineFile(t, "file.txt")

	_, err := mcphashfragment.MCPHashFragment(path, "5-2")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcphashfragment.ErrInvalidLineRange) {
		t.Errorf("expected ErrInvalidLineRange, got: %v", err)
	}
}

func TestMCPHashFragment_StartLessThanOne(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	path := testFiveLineFile(t, "file.txt")

	_, err := mcphashfragment.MCPHashFragment(path, "0-5")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcphashfragment.ErrInvalidLineRange) {
		t.Errorf("expected ErrInvalidLineRange, got: %v", err)
	}
}

func TestMCPHashFragment_RangeOutOfBounds(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "small.txt", "line1\nline2\nline3\n")

	_, err := mcphashfragment.MCPHashFragment("small.txt", "1-10")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcphashfragment.ErrInvalidLineRange) {
		t.Errorf("expected ErrInvalidLineRange, got: %v", err)
	}
}

func TestMCPHashFragment_EmptyPath(t *testing.T) {
	_, err := mcphashfragment.MCPHashFragment("", "1-5")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrPathIsEmpty) {
		t.Errorf("expected ErrPathIsEmpty, got: %v", err)
	}
}

func TestMCPHashFragment_PathTraversal(t *testing.T) {
	_, err := mcphashfragment.MCPHashFragment("../../etc/passwd", "1-5")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got: %v", err)
	}
}
