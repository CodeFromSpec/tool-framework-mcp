// code-from-spec: ROOT/golang/tests/mcp_tools/hash_fragment@2et7uCfpfEIfmAPMPi6VAKWiThA
package mcphashfragment_test

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

// testComputeHash computes the expected SHA-1 base64url (no padding) hash
// of the given content string, which should already have \n appended per line.
func testComputeHash(content string) string {
	h := sha1.Sum([]byte(content))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// testMakeFile creates a file at the given relative path inside tempDir,
// writing the given content. Any necessary subdirectories are created.
func testMakeFile(t *testing.T, tempDir, relPath, content string) {
	t.Helper()
	fullPath := filepath.Join(tempDir, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("testMakeFile mkdir: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("testMakeFile write: %v", err)
	}
}

// testFiveLineContent is the standard 5-line file content used in multiple tests.
const testFiveLineContent = "alpha\nbravo\ncharlie\ndelta\necho\n"

func TestMCPHashFragment_ValidRange(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testMakeFile(t, tempDir, "testfile.txt", testFiveLineContent)

	hash, err := mcphashfragment.MCPHashFragment("testfile.txt", "2-4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hash) != 27 {
		t.Errorf("expected hash length 27, got %d: %q", len(hash), hash)
	}

	expected := testComputeHash("bravo\ncharlie\ndelta\n")
	if hash != expected {
		t.Errorf("expected hash %q, got %q", expected, hash)
	}
}

func TestMCPHashFragment_SingleLine(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testMakeFile(t, tempDir, "testfile.txt", testFiveLineContent)

	hash, err := mcphashfragment.MCPHashFragment("testfile.txt", "3-3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hash) != 27 {
		t.Errorf("expected hash length 27, got %d: %q", len(hash), hash)
	}

	expected := testComputeHash("charlie\n")
	if hash != expected {
		t.Errorf("expected hash %q, got %q", expected, hash)
	}
}

func TestMCPHashFragment_FirstLine(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testMakeFile(t, tempDir, "testfile.txt", testFiveLineContent)

	hash, err := mcphashfragment.MCPHashFragment("testfile.txt", "1-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hash) != 27 {
		t.Errorf("expected hash length 27, got %d: %q", len(hash), hash)
	}

	expected := testComputeHash("alpha\n")
	if hash != expected {
		t.Errorf("expected hash %q, got %q", expected, hash)
	}
}

func TestMCPHashFragment_LastLine(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testMakeFile(t, tempDir, "testfile.txt", testFiveLineContent)

	hash, err := mcphashfragment.MCPHashFragment("testfile.txt", "5-5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hash) != 27 {
		t.Errorf("expected hash length 27, got %d: %q", len(hash), hash)
	}

	expected := testComputeHash("echo\n")
	if hash != expected {
		t.Errorf("expected hash %q, got %q", expected, hash)
	}
}

func TestMCPHashFragment_Deterministic(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testMakeFile(t, tempDir, "testfile.txt", testFiveLineContent)

	hash1, err := mcphashfragment.MCPHashFragment("testfile.txt", "2-4")
	if err != nil {
		t.Fatalf("first call unexpected error: %v", err)
	}

	hash2, err := mcphashfragment.MCPHashFragment("testfile.txt", "2-4")
	if err != nil {
		t.Fatalf("second call unexpected error: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("expected deterministic results, got %q and %q", hash1, hash2)
	}
}

func TestMCPHashFragment_FileDoesNotExist(t *testing.T) {
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
	testMakeFile(t, tempDir, "testfile.txt", testFiveLineContent)

	_, err := mcphashfragment.MCPHashFragment("testfile.txt", "abc")
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
	testMakeFile(t, tempDir, "testfile.txt", testFiveLineContent)

	_, err := mcphashfragment.MCPHashFragment("testfile.txt", "5-2")
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
	testMakeFile(t, tempDir, "testfile.txt", testFiveLineContent)

	_, err := mcphashfragment.MCPHashFragment("testfile.txt", "0-5")
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
	testMakeFile(t, tempDir, "testfile.txt", "line1\nline2\nline3\n")

	_, err := mcphashfragment.MCPHashFragment("testfile.txt", "1-10")
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
	if !errors.Is(err, pathutils.ErrPathEmpty) {
		t.Errorf("expected ErrPathEmpty, got: %v", err)
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

// Ensure the imports for fmt and filepath are used — suppress unused import
// errors for helper references not used directly in test bodies.
var _ = fmt.Sprintf
var _ = filepath.Join
