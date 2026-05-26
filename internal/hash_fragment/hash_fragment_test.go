// code-from-spec: ROOT/golang/internal/tools/hash_fragment/tests@ZSgxdbMC6Y2mJWtbwLHMqphNkgY

// Package hash_fragment contains tests for the hash_fragment MCP tool handler.
//
// These are internal tests (same package as the implementation) so they share
// the package namespace and can exercise the handler without any export gap.
//
// Each test uses t.TempDir() as an isolated project root and changes the
// working directory to it so that pathvalidation.ValidatePath(".", ...) anchors
// to the temp directory. The previous working directory is restored after each
// test.
//
// Test helper functions and types are prefixed with "test" to avoid collisions
// with unexported identifiers in the package under test.
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

// testLines is the canonical five-line content used across most tests.
// Index 0 = line 1, index 1 = line 2, …
var testLines = []string{
	"alpha",   // line 1
	"bravo",   // line 2
	"charlie", // line 3
	"delta",   // line 4
	"echo",    // line 5
}

// testWriteFile creates a file at relPath inside projectRoot, writing the
// provided lines joined with LF. It returns the relative path that was written
// (same as relPath) so callers can pass it straight to the handler.
//
// Cleanup is automatic because projectRoot should be a t.TempDir().
func testWriteFile(t *testing.T, projectRoot, relPath string, lines []string) string {
	t.Helper()
	abs := filepath.Join(projectRoot, relPath)
	// Create any intermediate directories.
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("testWriteFile MkdirAll: %v", err)
	}
	content := strings.Join(lines, "\n")
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatalf("testWriteFile WriteFile: %v", err)
	}
	return relPath
}

// testHashLines computes the expected SHA-1 hash for the provided lines joined
// with LF, encoded as a 27-character base64url string without padding.
//
// This mirrors exactly what HandleHashFragment does internally so tests can
// produce expected values without duplicating the algorithm manually.
func testHashLines(lines []string) string {
	joined := strings.Join(lines, "\n")
	sum := sha1.Sum([]byte(joined))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// testChdir changes the working directory to dir and registers a cleanup
// function that restores the original working directory after the test.
//
// The handler uses "." as the project root, which resolves against the process
// working directory — so tests must chdir into their temp project root.
func testChdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChdir Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChdir Chdir(%q): %v", dir, err)
	}
	t.Cleanup(func() {
		// Restore the original directory; ignore errors because the test has
		// already finished by the time cleanup runs.
		_ = os.Chdir(orig)
	})
}

// testCall invokes HandleHashFragment and returns the result plus the Go-level
// error. The returned Go error is expected to be nil for all normal cases; a
// non-nil Go error indicates a catastrophic handler failure.
func testCall(t *testing.T, path, lines string) (*mcp.CallToolResult, error) {
	t.Helper()
	args := HashFragmentArgs{Path: path, Lines: lines}
	result, _, goErr := HandleHashFragment(context.Background(), nil, args)
	return result, goErr
}

// testAssertSuccess asserts that the result is a non-error tool response and
// returns the text content.
func testAssertSuccess(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if result == nil {
		t.Fatal("testAssertSuccess: result is nil")
	}
	if result.IsError {
		t.Fatalf("testAssertSuccess: expected success but got tool error: %s",
			testResultText(result))
	}
	return testResultText(result)
}

// testAssertToolError asserts that the result has IsError: true and that its
// text content contains wantSubstr.
func testAssertToolError(t *testing.T, result *mcp.CallToolResult, wantSubstr string) {
	t.Helper()
	if result == nil {
		t.Fatal("testAssertToolError: result is nil")
	}
	if !result.IsError {
		t.Fatalf("testAssertToolError: expected tool error but got success: %s",
			testResultText(result))
	}
	text := testResultText(result)
	if !strings.Contains(text, wantSubstr) {
		t.Errorf("testAssertToolError: error message %q does not contain %q",
			text, wantSubstr)
	}
}

// testResultText extracts the text from the first TextContent entry in a
// CallToolResult.
func testResultText(result *mcp.CallToolResult) string {
	if result == nil || len(result.Content) == 0 {
		return ""
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		return ""
	}
	return tc.Text
}

// testAssertHash asserts that text is a 27-character string and that it equals
// the expected hash produced by testHashLines for the given lines.
func testAssertHash(t *testing.T, text string, expectedLines []string) {
	t.Helper()
	// The encoding always produces 27 characters for a 20-byte SHA-1 digest.
	if len(text) != 27 {
		t.Errorf("hash length = %d, want 27; text = %q", len(text), text)
	}
	want := testHashLines(expectedLines)
	if text != want {
		t.Errorf("hash = %q, want %q", text, want)
	}
}

// ---------------------------------------------------------------------------
// Happy-path tests
// ---------------------------------------------------------------------------

// TestHashFragment_HappyPath_MultipleLines verifies that the handler correctly
// hashes lines 2 through 4 (inclusive) of a multi-line file.
func TestHashFragment_HappyPath_MultipleLines(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	testWriteFile(t, root, "sample.txt", testLines)

	result, goErr := testCall(t, "sample.txt", "2-4")
	if goErr != nil {
		t.Fatalf("unexpected Go error: %v", goErr)
	}

	text := testAssertSuccess(t, result)

	// Lines 2-4 are indices 1-3: bravo, charlie, delta.
	testAssertHash(t, text, testLines[1:4])
}

// TestHashFragment_HappyPath_SingleLine verifies the handler hashes exactly
// one line when start == end (line 3).
func TestHashFragment_HappyPath_SingleLine(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	testWriteFile(t, root, "sample.txt", testLines)

	result, goErr := testCall(t, "sample.txt", "3-3")
	if goErr != nil {
		t.Fatalf("unexpected Go error: %v", goErr)
	}

	text := testAssertSuccess(t, result)

	// Line 3 only = "charlie" (index 2).
	testAssertHash(t, text, testLines[2:3])
}

// TestHashFragment_HappyPath_FirstLine verifies that the handler handles the
// first line of the file (Lines: "1-1").
func TestHashFragment_HappyPath_FirstLine(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	testWriteFile(t, root, "sample.txt", testLines)

	result, goErr := testCall(t, "sample.txt", "1-1")
	if goErr != nil {
		t.Fatalf("unexpected Go error: %v", goErr)
	}

	text := testAssertSuccess(t, result)

	// Line 1 only = "alpha" (index 0).
	testAssertHash(t, text, testLines[0:1])
}

// TestHashFragment_HappyPath_LastLine verifies that the handler handles the
// last line of a 5-line file (Lines: "5-5").
func TestHashFragment_HappyPath_LastLine(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	// Exactly 5 lines as specified.
	testWriteFile(t, root, "sample.txt", testLines)

	result, goErr := testCall(t, "sample.txt", "5-5")
	if goErr != nil {
		t.Fatalf("unexpected Go error: %v", goErr)
	}

	text := testAssertSuccess(t, result)

	// Line 5 only = "echo" (index 4).
	testAssertHash(t, text, testLines[4:5])
}

// ---------------------------------------------------------------------------
// Failure-case tests
// ---------------------------------------------------------------------------

// TestHashFragment_Failure_FileNotFound verifies that requesting a file that
// does not exist returns a tool error containing "file not found".
func TestHashFragment_Failure_FileNotFound(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	// Do NOT create the file — it must be absent.
	result, goErr := testCall(t, "nonexistent.go", "1-5")
	if goErr != nil {
		t.Fatalf("unexpected Go error: %v", goErr)
	}

	testAssertToolError(t, result, "file not found")
}

// TestHashFragment_Failure_InvalidLineRangeFormat_NotARange verifies that a
// non-numeric, non-range string returns a tool error containing "invalid line
// range".
func TestHashFragment_Failure_InvalidLineRangeFormat_NotARange(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	// Path doesn't matter; the range is parsed before the file is opened.
	// However, we provide a valid file to ensure the error originates from
	// range parsing, not file access.
	testWriteFile(t, root, "sample.txt", testLines)

	result, goErr := testCall(t, "sample.txt", "abc")
	if goErr != nil {
		t.Fatalf("unexpected Go error: %v", goErr)
	}

	testAssertToolError(t, result, "invalid line range")
}

// TestHashFragment_Failure_InvalidLineRangeFormat_StartGreaterThanEnd verifies
// that a range where start > end returns "invalid line range".
func TestHashFragment_Failure_InvalidLineRangeFormat_StartGreaterThanEnd(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	testWriteFile(t, root, "sample.txt", testLines)

	result, goErr := testCall(t, "sample.txt", "5-2")
	if goErr != nil {
		t.Fatalf("unexpected Go error: %v", goErr)
	}

	testAssertToolError(t, result, "invalid line range")
}

// TestHashFragment_Failure_LineRangeOutOfBounds verifies that requesting lines
// beyond the file's actual line count returns a tool error that contains both
// "invalid line range" and the file's actual line count.
func TestHashFragment_Failure_LineRangeOutOfBounds(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	// Create a file with only 3 lines.
	threeLines := []string{"one", "two", "three"}
	testWriteFile(t, root, "small.txt", threeLines)

	// Request lines 1-10, which exceeds the 3-line file.
	result, goErr := testCall(t, "small.txt", "1-10")
	if goErr != nil {
		t.Fatalf("unexpected Go error: %v", goErr)
	}

	text := testResultText(result)

	testAssertToolError(t, result, "invalid line range")

	// The error message must also contain the actual line count (3).
	wantCount := fmt.Sprintf("%d", len(threeLines))
	if !strings.Contains(text, wantCount) {
		t.Errorf("error message %q does not contain actual line count %q",
			text, wantCount)
	}
}

// TestHashFragment_Failure_EmptyPath verifies that an empty path triggers a
// path-validation tool error.
func TestHashFragment_Failure_EmptyPath(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	result, goErr := testCall(t, "", "1-5")
	if goErr != nil {
		t.Fatalf("unexpected Go error: %v", goErr)
	}

	// The path validation error message for empty path is "path is empty".
	testAssertToolError(t, result, "path is empty")
}

// TestHashFragment_Failure_PathTraversal verifies that a directory traversal
// attempt is rejected by path validation before any file access occurs.
func TestHashFragment_Failure_PathTraversal(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	result, goErr := testCall(t, "../../etc/passwd", "1-5")
	if goErr != nil {
		t.Fatalf("unexpected Go error: %v", goErr)
	}

	// Path validation must reject this; the exact message comes from
	// pathvalidation.ValidatePath — it contains "traversal" or "outside".
	// We check for a broad term that covers both possible messages.
	text := testResultText(result)
	if !result.IsError {
		t.Fatalf("expected tool error but got success: %s", text)
	}
	if !strings.Contains(text, "traversal") && !strings.Contains(text, "outside") {
		t.Errorf("path traversal error %q does not mention traversal or outside", text)
	}
}

// TestHashFragment_Failure_StartLineZero verifies that a range starting at
// line 0 is rejected as "invalid line range" (lines are 1-indexed).
func TestHashFragment_Failure_StartLineZero(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	testWriteFile(t, root, "sample.txt", testLines)

	result, goErr := testCall(t, "sample.txt", "0-5")
	if goErr != nil {
		t.Fatalf("unexpected Go error: %v", goErr)
	}

	testAssertToolError(t, result, "invalid line range")
}
