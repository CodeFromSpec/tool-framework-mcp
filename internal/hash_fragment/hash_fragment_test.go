// code-from-spec: ROOT/golang/internal/tools/hash_fragment/tests@lQhJmtalB760DTWzZVwmeMSbjmg

// Package hash_fragment contains tests for the hash_fragment tool handler.
// Each test uses t.TempDir() as the working directory so the handler's
// path validation logic resolves paths against a controlled root.
package hash_fragment

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// testMakeFile writes content to a file inside dir and returns the relative
// path suitable for passing to the handler (just the filename).
func testMakeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("testMakeFile: %v", err)
	}
	return name
}

// testExpectedHash computes the expected 27-character base64url-encoded SHA-1
// of lines joined with LF, mirroring what the implementation should produce.
// lines is a slice where index 0 = line 1 of the file.
func testExpectedHash(lines []string) string {
	joined := strings.Join(lines, "\n")
	sum := sha1.Sum([]byte(joined))
	// base64url, no padding — trim trailing '=' to reach exactly 27 chars.
	encoded := base64.URLEncoding.EncodeToString(sum[:])
	return strings.TrimRight(encoded, "=")
}

// testCallHandler is a convenience wrapper that sets the working directory to
// dir, calls HandleHashFragment, then restores the original working directory.
func testCallHandler(t *testing.T, dir string, args HashFragmentArgs) (*mcp.CallToolResult, error) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("testCallHandler: getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testCallHandler: chdir: %v", err)
	}
	t.Cleanup(func() {
		// Restore working directory after the test regardless of outcome.
		if err := os.Chdir(orig); err != nil {
			t.Logf("testCallHandler cleanup: chdir back: %v", err)
		}
	})

	result, _, handlerErr := HandleHashFragment(context.Background(), &mcp.CallToolRequest{}, args)
	return result, handlerErr
}

// testText extracts the text from the first TextContent entry of a result.
func testText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if result == nil {
		t.Fatal("testText: result is nil")
	}
	if len(result.Content) == 0 {
		t.Fatal("testText: result.Content is empty")
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("testText: first content is %T, want *mcp.TextContent", result.Content[0])
	}
	return tc.Text
}

// testAssertSuccess checks that the result is a success (IsError is false) and
// returns the text so the caller can inspect it.
func testAssertSuccess(t *testing.T, result *mcp.CallToolResult, handlerErr error) string {
	t.Helper()
	if handlerErr != nil {
		t.Fatalf("handler returned unexpected Go error: %v", handlerErr)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if result.IsError {
		t.Fatalf("expected success result but got IsError=true, text: %s", testText(t, result))
	}
	return testText(t, result)
}

// testAssertToolError checks that the result is an MCP tool error and that its
// text contains the expected substring.
func testAssertToolError(t *testing.T, result *mcp.CallToolResult, handlerErr error, wantSubstr string) {
	t.Helper()
	if handlerErr != nil {
		t.Fatalf("handler returned unexpected Go error: %v", handlerErr)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if !result.IsError {
		t.Fatalf("expected tool error but got success, text: %s", testText(t, result))
	}
	got := testText(t, result)
	if !strings.Contains(got, wantSubstr) {
		t.Fatalf("error text %q does not contain expected substring %q", got, wantSubstr)
	}
}

// ---------------------------------------------------------------------------
// Happy-path tests
// ---------------------------------------------------------------------------

// TestHandleHashFragment_ValidRange verifies that the handler returns a
// 27-character base64url SHA-1 hash for a normal multi-line range.
func TestHandleHashFragment_ValidRange(t *testing.T) {
	dir := t.TempDir()

	// Known file content — lines are numbered 1-5 for easy reference.
	fileLines := []string{
		"line one",   // line 1
		"line two",   // line 2
		"line three", // line 3
		"line four",  // line 4
		"line five",  // line 5
	}
	relPath := testMakeFile(t, dir, "sample.txt", strings.Join(fileLines, "\n"))

	result, handlerErr := testCallHandler(t, dir, HashFragmentArgs{
		Path:  relPath,
		Lines: "2-4",
	})

	text := testAssertSuccess(t, result, handlerErr)

	// The hash must be exactly 27 characters long.
	if len(text) != 27 {
		t.Fatalf("expected 27-char hash, got %d chars: %q", len(text), text)
	}

	// Verify the hash matches the SHA-1 of lines 2–4 joined with LF.
	// fileLines is 0-indexed; lines 2-4 map to indices 1-3.
	want := testExpectedHash(fileLines[1:4])
	if text != want {
		t.Fatalf("hash mismatch: got %q, want %q", text, want)
	}
}

// TestHandleHashFragment_SingleLine verifies that "3-3" returns the SHA-1 of
// exactly line 3 (no extra LF from joining a single element).
func TestHandleHashFragment_SingleLine(t *testing.T) {
	dir := t.TempDir()

	fileLines := []string{
		"alpha",   // line 1
		"beta",    // line 2
		"gamma",   // line 3
		"delta",   // line 4
	}
	relPath := testMakeFile(t, dir, "single.txt", strings.Join(fileLines, "\n"))

	result, handlerErr := testCallHandler(t, dir, HashFragmentArgs{
		Path:  relPath,
		Lines: "3-3",
	})

	text := testAssertSuccess(t, result, handlerErr)

	if len(text) != 27 {
		t.Fatalf("expected 27-char hash, got %d chars: %q", len(text), text)
	}

	// Single element join produces no separator.
	want := testExpectedHash([]string{fileLines[2]})
	if text != want {
		t.Fatalf("hash mismatch: got %q, want %q", text, want)
	}
}

// TestHandleHashFragment_FirstLine verifies that "1-1" hashes only line 1.
func TestHandleHashFragment_FirstLine(t *testing.T) {
	dir := t.TempDir()

	fileLines := []string{
		"first line",
		"second line",
		"third line",
	}
	relPath := testMakeFile(t, dir, "first.txt", strings.Join(fileLines, "\n"))

	result, handlerErr := testCallHandler(t, dir, HashFragmentArgs{
		Path:  relPath,
		Lines: "1-1",
	})

	text := testAssertSuccess(t, result, handlerErr)

	if len(text) != 27 {
		t.Fatalf("expected 27-char hash, got %d chars: %q", len(text), text)
	}

	want := testExpectedHash([]string{fileLines[0]})
	if text != want {
		t.Fatalf("hash mismatch: got %q, want %q", text, want)
	}
}

// TestHandleHashFragment_LastLine verifies that "5-5" on a 5-line file hashes
// only the last line.
func TestHandleHashFragment_LastLine(t *testing.T) {
	dir := t.TempDir()

	fileLines := []string{
		"line 1",
		"line 2",
		"line 3",
		"line 4",
		"line 5",
	}
	relPath := testMakeFile(t, dir, "last.txt", strings.Join(fileLines, "\n"))

	result, handlerErr := testCallHandler(t, dir, HashFragmentArgs{
		Path:  relPath,
		Lines: "5-5",
	})

	text := testAssertSuccess(t, result, handlerErr)

	if len(text) != 27 {
		t.Fatalf("expected 27-char hash, got %d chars: %q", len(text), text)
	}

	want := testExpectedHash([]string{fileLines[4]})
	if text != want {
		t.Fatalf("hash mismatch: got %q, want %q", text, want)
	}
}

// ---------------------------------------------------------------------------
// Failure-case tests
// ---------------------------------------------------------------------------

// TestHandleHashFragment_FileNotFound verifies the "file not found" error path.
func TestHandleHashFragment_FileNotFound(t *testing.T) {
	dir := t.TempDir()

	result, handlerErr := testCallHandler(t, dir, HashFragmentArgs{
		Path:  "nonexistent.go",
		Lines: "1-5",
	})

	testAssertToolError(t, result, handlerErr, "file not found")
}

// TestHandleHashFragment_InvalidRangeNotARange verifies that a non-range string
// like "abc" produces an "invalid line range" error.
func TestHandleHashFragment_InvalidRangeNotARange(t *testing.T) {
	dir := t.TempDir()

	// The file does not need to exist; parsing should fail first.
	result, handlerErr := testCallHandler(t, dir, HashFragmentArgs{
		Path:  "any.txt",
		Lines: "abc",
	})

	testAssertToolError(t, result, handlerErr, "invalid line range")
}

// TestHandleHashFragment_InvalidRangeStartGreaterThanEnd verifies that "5-2"
// is rejected as an invalid line range.
func TestHandleHashFragment_InvalidRangeStartGreaterThanEnd(t *testing.T) {
	dir := t.TempDir()

	result, handlerErr := testCallHandler(t, dir, HashFragmentArgs{
		Path:  "any.txt",
		Lines: "5-2",
	})

	testAssertToolError(t, result, handlerErr, "invalid line range")
}

// TestHandleHashFragment_RangeOutOfBounds verifies that requesting lines
// beyond the file's actual line count produces an error that includes both
// "invalid line range" and the actual line count.
func TestHandleHashFragment_RangeOutOfBounds(t *testing.T) {
	dir := t.TempDir()

	fileLines := []string{"a", "b", "c"} // exactly 3 lines
	relPath := testMakeFile(t, dir, "short.txt", strings.Join(fileLines, "\n"))

	result, handlerErr := testCallHandler(t, dir, HashFragmentArgs{
		Path:  relPath,
		Lines: "1-10",
	})

	text := testText(t, result)

	// Must be a tool error.
	if !result.IsError {
		t.Fatalf("expected tool error but got success, text: %s", text)
	}
	if handlerErr != nil {
		t.Fatalf("handler returned unexpected Go error: %v", handlerErr)
	}

	// Must mention "invalid line range".
	if !strings.Contains(text, "invalid line range") {
		t.Fatalf("error text %q does not contain %q", text, "invalid line range")
	}

	// Must include the actual line count (3) so the agent knows what the file has.
	lineCountStr := fmt.Sprintf("%d", len(fileLines))
	if !strings.Contains(text, lineCountStr) {
		t.Fatalf("error text %q does not contain actual line count %q", text, lineCountStr)
	}
}

// TestHandleHashFragment_EmptyPath verifies that an empty path triggers path
// validation and returns a tool error.
func TestHandleHashFragment_EmptyPath(t *testing.T) {
	dir := t.TempDir()

	result, handlerErr := testCallHandler(t, dir, HashFragmentArgs{
		Path:  "",
		Lines: "1-5",
	})

	// The exact message wording is determined by ValidatePath, but it must be
	// a tool error (IsError=true).
	if handlerErr != nil {
		t.Fatalf("handler returned unexpected Go error: %v", handlerErr)
	}
	if result == nil || !result.IsError {
		t.Fatalf("expected tool error for empty path, got success or nil result")
	}
}

// TestHandleHashFragment_PathTraversal verifies that "../../etc/passwd" is
// rejected by path validation.
func TestHandleHashFragment_PathTraversal(t *testing.T) {
	dir := t.TempDir()

	result, handlerErr := testCallHandler(t, dir, HashFragmentArgs{
		Path:  "../../etc/passwd",
		Lines: "1-5",
	})

	if handlerErr != nil {
		t.Fatalf("handler returned unexpected Go error: %v", handlerErr)
	}
	if result == nil || !result.IsError {
		t.Fatalf("expected tool error for path traversal, got success or nil result")
	}
}

// TestHandleHashFragment_StartLineZero verifies that "0-5" is treated as an
// invalid line range (lines are 1-indexed).
func TestHandleHashFragment_StartLineZero(t *testing.T) {
	dir := t.TempDir()

	result, handlerErr := testCallHandler(t, dir, HashFragmentArgs{
		Path:  "any.txt",
		Lines: "0-5",
	})

	testAssertToolError(t, result, handlerErr, "invalid line range")
}
