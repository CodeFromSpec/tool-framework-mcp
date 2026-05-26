// code-from-spec: ROOT/golang/internal/tools/write_file/tests@no4nLtVdbowuN4SpEEtOWB8Btrk

// Tests for the write_file MCP tool handler.
//
// Each test creates an isolated temp directory as the project root and changes
// the process working directory to it, so that os.Getwd() inside the handler
// resolves to the temp dir. A minimal spec tree is created under
// "code-from-spec/<path>/_node.md" with the required frontmatter.
//
// Test helpers follow the project convention: all helper functions and types
// are prefixed with "test" to prevent name collisions with unexported
// identifiers in the package under test.
package write_file

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// --------------------------------------------------------------------------
// Test helpers
// --------------------------------------------------------------------------

// testSetupRoot creates a fresh temp dir and changes the process working
// directory to it. A cleanup function is registered to restore the original
// working directory when the test completes.
//
// Returns the absolute path of the temp dir (the project root for this test).
func testSetupRoot(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("testSetupRoot: Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testSetupRoot: Chdir(%s): %v", dir, err)
	}
	t.Cleanup(func() {
		// Restore the original working directory so subsequent tests are not
		// affected. Note that t.TempDir() cleans up the directory itself.
		if err := os.Chdir(orig); err != nil {
			t.Errorf("testSetupRoot cleanup: Chdir(%s): %v", orig, err)
		}
	})
	return dir
}

// testWriteNode creates the "_node.md" spec file for a given logical name
// inside the current working directory (which must be the project root).
//
// frontmatterBody is raw YAML placed between "---" delimiters. Example:
//
//	"outputs:\n  - id: file\n    path: output/file.go"
//
// The logical name must start with "ROOT/".
func testWriteNode(t *testing.T, logicalName string, frontmatterBody string) {
	t.Helper()

	if !strings.HasPrefix(logicalName, "ROOT/") {
		t.Fatalf("testWriteNode: logical name must start with ROOT/, got %q", logicalName)
	}

	// Derive the node file path: "ROOT/a/b" -> "code-from-spec/a/b/_node.md"
	rel := strings.TrimPrefix(logicalName, "ROOT/")
	nodePath := filepath.Join("code-from-spec", filepath.FromSlash(rel), "_node.md")

	if err := os.MkdirAll(filepath.Dir(nodePath), 0o755); err != nil {
		t.Fatalf("testWriteNode: MkdirAll(%s): %v", filepath.Dir(nodePath), err)
	}

	content := "---\n" + frontmatterBody + "\n---\n"
	if err := os.WriteFile(nodePath, []byte(content), 0o644); err != nil {
		t.Fatalf("testWriteNode: WriteFile(%s): %v", nodePath, err)
	}
}

// testFirstText extracts the text from the first Content entry of a
// CallToolResult and fails the test if no such entry exists or if the entry
// is not a *mcp.TextContent.
func testFirstText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if result == nil {
		t.Fatal("testFirstText: result is nil")
	}
	if len(result.Content) == 0 {
		t.Fatal("testFirstText: result has no content entries")
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("testFirstText: first content entry is %T, want *mcp.TextContent", result.Content[0])
	}
	return tc.Text
}

// testCall is a convenience wrapper that invokes HandleWriteFile and
// immediately fails the test on an unexpected Go-level error.
func testCall(
	t *testing.T,
	logicalName, path, content string,
) *mcp.CallToolResult {
	t.Helper()
	result, _, goErr := HandleWriteFile(context.Background(), nil, WriteFileArgs{
		LogicalName: logicalName,
		Path:        path,
		Content:     content,
	})
	if goErr != nil {
		t.Fatalf("testCall: unexpected Go error: %v", goErr)
	}
	if result == nil {
		t.Fatal("testCall: result is nil")
	}
	return result
}

// --------------------------------------------------------------------------
// Happy Path Tests
// --------------------------------------------------------------------------

// TestWriteFile_Success verifies the basic happy path:
//   - The file is written to disk with the provided content.
//   - The success result carries the text "wrote <path>".
func TestWriteFile_Success(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "ROOT/a", "outputs:\n  - id: file\n    path: output/file.go")

	result := testCall(t, "ROOT/a", "output/file.go", "package main")

	// Must not be an error result.
	if result.IsError {
		t.Fatalf("expected success result, got tool error: %s", testFirstText(t, result))
	}

	// The success message must identify the file that was written.
	text := testFirstText(t, result)
	if text != "wrote output/file.go" {
		t.Errorf("success text = %q, want %q", text, "wrote output/file.go")
	}

	// The file must exist on disk with the exact content provided.
	got, err := os.ReadFile("output/file.go")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "package main" {
		t.Errorf("file content = %q, want %q", string(got), "package main")
	}
}

// TestWriteFile_CreatesIntermediateDirectories verifies that the handler
// creates all missing parent directories before writing.
func TestWriteFile_CreatesIntermediateDirectories(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "ROOT/a", "outputs:\n  - id: file\n    path: deep/nested/dir/file.go")

	result := testCall(t, "ROOT/a", "deep/nested/dir/file.go", "package deep")

	if result.IsError {
		t.Fatalf("expected success, got tool error: %s", testFirstText(t, result))
	}

	// The file must exist even though the directories did not exist beforehand.
	if _, err := os.Stat("deep/nested/dir/file.go"); os.IsNotExist(err) {
		t.Error("expected deep/nested/dir/file.go to exist but it does not")
	}
}

// TestWriteFile_OverwritesExistingFile verifies that the handler replaces the
// full content of an existing file rather than appending or failing.
func TestWriteFile_OverwritesExistingFile(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "ROOT/a", "outputs:\n  - id: file\n    path: output/file.go")

	// Pre-create the output directory and file.
	if err := os.MkdirAll("output", 0o755); err != nil {
		t.Fatalf("setup MkdirAll: %v", err)
	}
	if err := os.WriteFile("output/file.go", []byte("old content"), 0o644); err != nil {
		t.Fatalf("setup WriteFile: %v", err)
	}

	result := testCall(t, "ROOT/a", "output/file.go", "new content")

	if result.IsError {
		t.Fatalf("expected success, got tool error: %s", testFirstText(t, result))
	}

	// The file must contain only the new content.
	got, err := os.ReadFile("output/file.go")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "new content" {
		t.Errorf("file content = %q, want %q", string(got), "new content")
	}
}

// TestWriteFile_BackslashNormalization verifies that Windows-style backslash
// separators in the path are normalized to forward slashes before the outputs
// list comparison, so a path like "output\file.go" matches the declared
// "output/file.go".
//
// This test is intentionally skipped on non-Windows platforms because a
// backslash is a valid filename character on Linux/macOS — not a separator —
// so the normalization semantics only apply on Windows.
func TestWriteFile_BackslashNormalization(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("backslash path separator normalization is only meaningful on Windows")
	}

	testSetupRoot(t)
	testWriteNode(t, "ROOT/a", "outputs:\n  - id: file\n    path: output/file.go")

	// Provide the path using the Windows backslash separator.
	result := testCall(t, "ROOT/a", `output\file.go`, "package main")

	if result.IsError {
		t.Fatalf("expected success after backslash normalization, got tool error: %s",
			testFirstText(t, result))
	}

	// The result must indicate success (contain "wrote").
	text := testFirstText(t, result)
	if !strings.Contains(text, "wrote") {
		t.Errorf("unexpected success text: %q", text)
	}

	// The file must have been written at the correct OS path.
	if _, err := os.Stat(filepath.Join("output", "file.go")); os.IsNotExist(err) {
		t.Error("expected output/file.go to exist but it does not")
	}
}

// --------------------------------------------------------------------------
// Failure Case Tests
// --------------------------------------------------------------------------

// TestWriteFile_InvalidLogicalNamePrefix verifies that a logical name that
// does not start with "ROOT/" is rejected immediately with a tool error.
func TestWriteFile_InvalidLogicalNamePrefix(t *testing.T) {
	testSetupRoot(t)

	// "SOMETHING/" is not a recognized ROOT/ prefix.
	result := testCall(t, "SOMETHING/external/something", "some/file.go", "package x")

	if !result.IsError {
		t.Fatalf("expected tool error for invalid logical name prefix, got success: %s",
			testFirstText(t, result))
	}
}

// TestWriteFile_NonexistentLogicalName verifies that a ROOT/ reference whose
// node file does not exist on disk produces a tool error (not a Go panic or
// server crash).
func TestWriteFile_NonexistentLogicalName(t *testing.T) {
	testSetupRoot(t)
	// Do NOT call testWriteNode — the node must not exist.

	result := testCall(t, "ROOT/nonexistent", "some/file.go", "package x")

	if !result.IsError {
		t.Fatalf("expected tool error for nonexistent node, got success: %s",
			testFirstText(t, result))
	}
}

// TestWriteFile_PathNotInOutputs verifies that providing a path that is not
// declared in the node's outputs list produces a tool error. The error message
// must be actionable: it must indicate the path is not allowed and list the
// allowed paths.
func TestWriteFile_PathNotInOutputs(t *testing.T) {
	testSetupRoot(t)
	// The node declares only "allowed/file.go".
	testWriteNode(t, "ROOT/a", "outputs:\n  - id: file\n    path: allowed/file.go")

	result := testCall(t, "ROOT/a", "other/file.go", "package x")

	if !result.IsError {
		t.Fatalf("expected tool error for path not in outputs, got success: %s",
			testFirstText(t, result))
	}

	// The error message must contain "path not allowed" (case-insensitive) so
	// the agent understands what went wrong.
	text := testFirstText(t, result)
	lower := strings.ToLower(text)
	if !strings.Contains(lower, "path not") {
		t.Errorf("expected error text to contain %q, got: %q", "path not", text)
	}

	// The error message should identify the rejected path or the node.
	if !strings.Contains(text, "other/file.go") && !strings.Contains(text, "ROOT/a") {
		t.Errorf("expected error text to identify the rejected path or node, got: %q", text)
	}
}

// TestWriteFile_PathTraversal verifies that a directory traversal attempt is
// rejected by ValidatePath. We include the traversal path in the outputs list
// so that the only rejection is from ValidatePath (not the outputs check),
// confirming the path safety layer fires independently.
func TestWriteFile_PathTraversal(t *testing.T) {
	testSetupRoot(t)
	// Include the traversal path in outputs to isolate the ValidatePath check.
	testWriteNode(t, "ROOT/a", "outputs:\n  - id: file\n    path: ../../etc/passwd")

	result := testCall(t, "ROOT/a", "../../etc/passwd", "malicious")

	if !result.IsError {
		t.Fatalf("expected tool error for path traversal, got success: %s",
			testFirstText(t, result))
	}

	// The error must mention traversal or escaping the project root.
	text := testFirstText(t, result)
	lower := strings.ToLower(text)
	if !strings.Contains(lower, "traversal") && !strings.Contains(lower, "outside") {
		t.Errorf("expected traversal/outside mention in error, got: %q", text)
	}
}

// TestWriteFile_EmptyPath verifies that an empty path string is rejected with
// an error message containing "path is empty" (the canonical message from
// ValidatePath).
func TestWriteFile_EmptyPath(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "ROOT/a", "outputs:\n  - id: file\n    path: some/file.go")

	result := testCall(t, "ROOT/a", "", "package x")

	if !result.IsError {
		t.Fatalf("expected tool error for empty path, got success: %s",
			testFirstText(t, result))
	}

	text := testFirstText(t, result)
	if !strings.Contains(strings.ToLower(text), "path is empty") {
		t.Errorf("expected \"path is empty\" in error text, got: %q", text)
	}
}

// TestWriteFile_SymlinkEscapingProjectRoot verifies that a symlink inside the
// project root that points to a file outside the project root is rejected.
//
// Skipped on Windows because creating symlinks there typically requires
// elevated privileges (SeCreateSymbolicLinkPrivilege) that may not be
// available in CI or developer environments.
func TestWriteFile_SymlinkEscapingProjectRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation on Windows requires elevated privileges; skipping")
	}

	// Create a target file outside the project root in a separate temp dir.
	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "secret.txt")
	if err := os.WriteFile(outsideFile, []byte("secret"), 0o644); err != nil {
		t.Fatalf("setup: WriteFile(outsideFile): %v", err)
	}

	// Switch the working directory to the project root AFTER creating the
	// outside dir, so Chdir points to the project root.
	testSetupRoot(t)

	// Create a symlink "link.txt" -> outsideFile inside the project root.
	if err := os.Symlink(outsideFile, "link.txt"); err != nil {
		t.Fatalf("setup: Symlink(%s, link.txt): %v", outsideFile, err)
	}

	// Include the symlink path in the outputs list so the check reaches
	// ValidatePath rather than failing at the outputs comparison.
	testWriteNode(t, "ROOT/a", "outputs:\n  - id: file\n    path: link.txt")

	result := testCall(t, "ROOT/a", "link.txt", "overwrite attempt")

	if !result.IsError {
		t.Fatalf("expected tool error for symlink escaping project root, got success: %s",
			testFirstText(t, result))
	}

	text := testFirstText(t, result)
	if !strings.Contains(strings.ToLower(text), "outside") {
		t.Errorf("expected \"outside\" in error text, got: %q", text)
	}
}
