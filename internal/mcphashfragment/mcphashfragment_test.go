// code-from-spec: ROOT/golang/tests/mcp_tools/hash_fragment@19MJY07Im4YlusR5VWFTWKxuYQc
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

// testBase64urlSHA1 computes the base64url-encoded (no padding) SHA-1 digest of data.
func testBase64urlSHA1(data string) string {
	h := sha1.Sum([]byte(data))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// testFiveLineFile is the standard content used in multiple tests.
const testFiveLineFile = "alpha\nbravo\ncharlie\ndelta\necho\n"

func TestMCPHashFragment_ValidRange(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte(testFiveLineFile), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := mcphashfragment.MCPHashFragment("file.txt", "2-4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 27 {
		t.Errorf("expected 27-char hash, got %d: %q", len(got), got)
	}
	want := testBase64urlSHA1("bravo\ncharlie\ndelta\n")
	if got != want {
		t.Errorf("hash mismatch: got %q, want %q", got, want)
	}
}

func TestMCPHashFragment_SingleLineRange(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte(testFiveLineFile), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := mcphashfragment.MCPHashFragment("file.txt", "3-3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 27 {
		t.Errorf("expected 27-char hash, got %d: %q", len(got), got)
	}
	want := testBase64urlSHA1("charlie\n")
	if got != want {
		t.Errorf("hash mismatch: got %q, want %q", got, want)
	}
}

func TestMCPHashFragment_FirstLine(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte(testFiveLineFile), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := mcphashfragment.MCPHashFragment("file.txt", "1-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 27 {
		t.Errorf("expected 27-char hash, got %d: %q", len(got), got)
	}
	want := testBase64urlSHA1("alpha\n")
	if got != want {
		t.Errorf("hash mismatch: got %q, want %q", got, want)
	}
}

func TestMCPHashFragment_LastLine(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte(testFiveLineFile), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := mcphashfragment.MCPHashFragment("file.txt", "5-5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 27 {
		t.Errorf("expected 27-char hash, got %d: %q", len(got), got)
	}
	want := testBase64urlSHA1("echo\n")
	if got != want {
		t.Errorf("hash mismatch: got %q, want %q", got, want)
	}
}

func TestMCPHashFragment_Deterministic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte(testFiveLineFile), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got1, err := mcphashfragment.MCPHashFragment("file.txt", "2-4")
	if err != nil {
		t.Fatalf("first call unexpected error: %v", err)
	}
	got2, err := mcphashfragment.MCPHashFragment("file.txt", "2-4")
	if err != nil {
		t.Fatalf("second call unexpected error: %v", err)
	}
	if got1 != got2 {
		t.Errorf("non-deterministic: got %q then %q", got1, got2)
	}
}

func TestMCPHashFragment_FileNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := mcphashfragment.MCPHashFragment("nonexistent.go", "1-5")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestMCPHashFragment_InvalidRangeFormat(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte(testFiveLineFile), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := mcphashfragment.MCPHashFragment("file.txt", "abc")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcphashfragment.ErrInvalidLineRange) {
		t.Errorf("expected ErrInvalidLineRange, got: %v", err)
	}
}

func TestMCPHashFragment_StartGreaterThanEnd(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte(testFiveLineFile), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := mcphashfragment.MCPHashFragment("file.txt", "5-2")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcphashfragment.ErrInvalidLineRange) {
		t.Errorf("expected ErrInvalidLineRange, got: %v", err)
	}
}

func TestMCPHashFragment_StartLessThanOne(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte(testFiveLineFile), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := mcphashfragment.MCPHashFragment("file.txt", "0-5")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcphashfragment.ErrInvalidLineRange) {
		t.Errorf("expected ErrInvalidLineRange, got: %v", err)
	}
}

func TestMCPHashFragment_RangeOutOfBounds(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "line1\nline2\nline3\n"
	if err := os.WriteFile("file.txt", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := mcphashfragment.MCPHashFragment("file.txt", "1-10")
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
