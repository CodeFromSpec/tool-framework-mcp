// code-from-spec: ROOT/golang/internal/tools/write_file/tests@UkExZuroWpy57NBy1f1dXovzMKU

// Package write_file provides tests for the write_file MCP tool handler.
// Each test creates an isolated temp directory as the project root, builds
// a minimal spec tree in it, then calls HandleWriteFile directly.
package write_file

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// ----------------------------------------------------------------------------
// Test helpers
// ----------------------------------------------------------------------------

// testMakeSpecFile creates the spec directory structure and writes a minimal
// YAML frontmatter block for the given logical name under root.
//
// logicalName must start with "ROOT/". The spec file is placed at:
//
//	<root>/code-from-spec/<rest-of-logical-name>/spec.md
//
// frontmatter is the raw YAML string to embed (no surrounding "---" fences —
// they are added by this helper).
func testMakeSpecFile(t *testing.T, root, logicalName, frontmatter string) {
	t.Helper()

	// Convert "ROOT/a/b" → "a/b" then build the directory path.
	const prefix = "ROOT/"
	if !strings.HasPrefix(logicalName, prefix) {
		t.Fatalf("testMakeSpecFile: logicalName must start with %q, got %q", prefix, logicalName)
	}
	rel := strings.TrimPrefix(logicalName, prefix)
	dir := filepath.Join(root, "code-from-spec", filepath.FromSlash(rel))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("testMakeSpecFile: MkdirAll %s: %v", dir, err)
	}

	content := "---\n" + frontmatter + "\n---\n"
	specPath := filepath.Join(dir, "spec.md")
	if err := os.WriteFile(specPath, []byte(content), 0o644); err != nil {
		t.Fatalf("testMakeSpecFile: WriteFile %s: %v", specPath, err)
	}
}

// testChangeDir temporarily changes the working directory to dir for the
// duration of the test and restores it afterwards. The write_file handler
// resolves paths relative to the working directory.
func testChangeDir(t *testing.T, dir string) {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChangeDir: Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChangeDir: Chdir %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Fatalf("testChangeDir cleanup: Chdir %s: %v", original, err)
		}
	})
}

// testCallHandler is a thin wrapper that calls HandleWriteFile with a
// background context and no *mcp.CallToolRequest (nil is acceptable because
// the handler extracts all inputs from args).
func testCallHandler(t *testing.T, args WriteFileArgs) (string, bool) {
	t.Helper()
	result, _, err := HandleWriteFile(context.Background(), nil, args)
	if err != nil {
		// A non-nil Go error signals a catastrophic server failure, which
		// should never happen in normal operation.
		t.Fatalf("HandleWriteFile returned unexpected Go error: %v", err)
	}
	if result == nil {
		t.Fatal("HandleWriteFile returned nil result")
	}
	if len(result.Content) == 0 {
		t.Fatal("HandleWriteFile returned result with no content entries")
	}
	// The spec mandates TextContent as the first content entry.
	type textContent interface {
		GetText() string
	}
	tc, ok := result.Content[0].(textContent)
	if !ok {
		// Try the concrete type directly via type assertion on the interface.
		t.Fatalf("first content entry does not implement GetText(); type=%T", result.Content[0])
	}
	return tc.GetText(), result.IsError
}

// ----------------------------------------------------------------------------
// Happy path tests
// ----------------------------------------------------------------------------

// TestWritesFileSuccessfully verifies that the handler writes the file to disk
// and returns the expected success message.
func TestWritesFileSuccessfully(t *testing.T) {
	root := t.TempDir()
	testChangeDir(t, root)

	testMakeSpecFile(t, root, "ROOT/a",
		"outputs:\n  - id: file\n    path: output/file.go")

	text, isErr := testCallHandler(t, WriteFileArgs{
		LogicalName: "ROOT/a",
		Path:        "output/file.go",
		Content:     "package main",
	})

	if isErr {
		t.Fatalf("expected success, got tool error: %s", text)
	}
	if text != "wrote output/file.go" {
		t.Errorf("unexpected success message: %q (want %q)", text, "wrote output/file.go")
	}

	// Verify the file exists with the correct content.
	got, err := os.ReadFile(filepath.Join(root, "output", "file.go"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "package main" {
		t.Errorf("file content = %q, want %q", string(got), "package main")
	}
}

// TestCreatesIntermediateDirectories verifies that deeply nested directories
// are created automatically when they do not yet exist.
func TestCreatesIntermediateDirectories(t *testing.T) {
	root := t.TempDir()
	testChangeDir(t, root)

	testMakeSpecFile(t, root, "ROOT/a",
		"outputs:\n  - id: file\n    path: deep/nested/dir/file.go")

	text, isErr := testCallHandler(t, WriteFileArgs{
		LogicalName: "ROOT/a",
		Path:        "deep/nested/dir/file.go",
		Content:     "package main",
	})

	if isErr {
		t.Fatalf("expected success, got tool error: %s", text)
	}

	// Confirm the file was physically created.
	if _, err := os.Stat(filepath.Join(root, "deep", "nested", "dir", "file.go")); err != nil {
		t.Errorf("expected file to exist: %v", err)
	}
}

// TestOverwritesExistingFile verifies that writing to a path that already has
// content replaces that content entirely.
func TestOverwritesExistingFile(t *testing.T) {
	root := t.TempDir()
	testChangeDir(t, root)

	testMakeSpecFile(t, root, "ROOT/a",
		"outputs:\n  - id: file\n    path: output/file.go")

	// Pre-create the file with old content.
	outputDir := filepath.Join(root, "output")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, "file.go"), []byte("old content"), 0o644); err != nil {
		t.Fatalf("WriteFile (initial): %v", err)
	}

	text, isErr := testCallHandler(t, WriteFileArgs{
		LogicalName: "ROOT/a",
		Path:        "output/file.go",
		Content:     "new content",
	})

	if isErr {
		t.Fatalf("expected success, got tool error: %s", text)
	}

	got, err := os.ReadFile(filepath.Join(root, "output", "file.go"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "new content" {
		t.Errorf("file content = %q, want %q", string(got), "new content")
	}
}

// TestPathWithBackslashesNormalizedWindowsOnly verifies that a path supplied
// with backslash separators is normalised to forward-slash before it is
// matched against the outputs list. This behaviour is only relevant on
// Windows; on Linux/macOS a backslash is a valid filename character, not a
// path separator, so the test is skipped there.
func TestPathWithBackslashesNormalizedWindowsOnly(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("backslash normalisation only applies on Windows")
	}

	root := t.TempDir()
	testChangeDir(t, root)

	testMakeSpecFile(t, root, "ROOT/a",
		"outputs:\n  - id: file\n    path: output/file.go")

	// Supply the path with Windows-style backslashes.
	text, isErr := testCallHandler(t, WriteFileArgs{
		LogicalName: "ROOT/a",
		Path:        `output\file.go`,
		Content:     "package main",
	})

	if isErr {
		t.Fatalf("expected success after normalisation, got tool error: %s", text)
	}
	// After normalisation the success message must use forward slashes.
	if text != "wrote output/file.go" {
		t.Errorf("unexpected success message: %q (want %q)", text, "wrote output/file.go")
	}
}

// ----------------------------------------------------------------------------
// Failure case tests
// ----------------------------------------------------------------------------

// TestInvalidLogicalNamePrefix verifies that a logical name not starting with
// "ROOT/" is rejected as a tool error.
func TestInvalidLogicalNamePrefix(t *testing.T) {
	root := t.TempDir()
	testChangeDir(t, root)

	text, isErr := testCallHandler(t, WriteFileArgs{
		LogicalName: "ROOT/external/something",
		Path:        "some/file.go",
		Content:     "package main",
	})

	if !isErr {
		t.Fatalf("expected tool error for invalid prefix, got success: %s", text)
	}
}

// TestNonexistentLogicalName verifies that referencing a logical name whose
// spec file does not exist returns a tool error.
func TestNonexistentLogicalName(t *testing.T) {
	root := t.TempDir()
	testChangeDir(t, root)

	// Do NOT create a spec file for ROOT/nonexistent.
	text, isErr := testCallHandler(t, WriteFileArgs{
		LogicalName: "ROOT/nonexistent",
		Path:        "some/file.go",
		Content:     "package main",
	})

	if !isErr {
		t.Fatalf("expected tool error for nonexistent node, got success: %s", text)
	}
}

// TestPathNotInOutputs verifies that a path not listed in the node's outputs
// is rejected with a message containing "path not allowed" and the list of
// allowed paths.
func TestPathNotInOutputs(t *testing.T) {
	root := t.TempDir()
	testChangeDir(t, root)

	testMakeSpecFile(t, root, "ROOT/a",
		"outputs:\n  - id: file\n    path: allowed/file.go")

	text, isErr := testCallHandler(t, WriteFileArgs{
		LogicalName: "ROOT/a",
		Path:        "other/file.go",
		Content:     "package main",
	})

	if !isErr {
		t.Fatalf("expected tool error for unlisted path, got success: %s", text)
	}
	if !strings.Contains(text, "path not allowed") {
		t.Errorf("error message does not contain %q: %s", "path not allowed", text)
	}
	// The error should also list the allowed path so the agent knows what to use.
	if !strings.Contains(text, "allowed/file.go") {
		t.Errorf("error message does not list allowed path %q: %s", "allowed/file.go", text)
	}
}

// TestPathTraversalAttempt verifies that a path designed to escape the project
// root via ".." is caught by ValidatePath and returns a tool error.
func TestPathTraversalAttempt(t *testing.T) {
	root := t.TempDir()
	testChangeDir(t, root)

	// Register the traversal path in outputs — the outputs check must not
	// short-circuit before ValidatePath runs.
	testMakeSpecFile(t, root, "ROOT/a",
		"outputs:\n  - id: file\n    path: ../../etc/passwd")

	text, isErr := testCallHandler(t, WriteFileArgs{
		LogicalName: "ROOT/a",
		Path:        "../../etc/passwd",
		Content:     "malicious",
	})

	if !isErr {
		t.Fatalf("expected tool error for path traversal, got success: %s", text)
	}
}

// TestEmptyPath verifies that an empty path string is rejected immediately
// with a message containing "path is empty".
func TestEmptyPath(t *testing.T) {
	root := t.TempDir()
	testChangeDir(t, root)

	testMakeSpecFile(t, root, "ROOT/a",
		"outputs:\n  - id: file\n    path: some/file.go")

	text, isErr := testCallHandler(t, WriteFileArgs{
		LogicalName: "ROOT/a",
		Path:        "",
		Content:     "package main",
	})

	if !isErr {
		t.Fatalf("expected tool error for empty path, got success: %s", text)
	}
	if !strings.Contains(text, "path is empty") {
		t.Errorf("error message does not contain %q: %s", "path is empty", text)
	}
}

// TestSymlinkEscapingProjectRoot verifies that a symlink inside the temp dir
// that resolves to a target outside the project root is rejected by
// ValidatePath with a message containing "resolves outside project root".
func TestSymlinkEscapingProjectRoot(t *testing.T) {
	// Symlink creation may require elevated privileges on Windows; skip if
	// os.Symlink fails with a permission error.
	root := t.TempDir()
	outside := t.TempDir() // a directory that is NOT under root

	// Create a symlink at <root>/escape.go pointing to <outside>/target.go
	symlinkPath := filepath.Join(root, "escape.go")
	targetPath := filepath.Join(outside, "target.go")
	if err := os.WriteFile(targetPath, []byte(""), 0o644); err != nil {
		t.Fatalf("WriteFile target: %v", err)
	}
	if err := os.Symlink(targetPath, symlinkPath); err != nil {
		if strings.Contains(err.Error(), "privilege") ||
			strings.Contains(err.Error(), "permission") {
			t.Skip("insufficient privileges to create symlinks on this platform")
		}
		t.Fatalf("Symlink: %v", err)
	}

	testChangeDir(t, root)

	// Register the symlink path in outputs so the path-in-outputs check passes
	// and ValidatePath is the one that catches the escape.
	testMakeSpecFile(t, root, "ROOT/a",
		"outputs:\n  - id: file\n    path: escape.go")

	text, isErr := testCallHandler(t, WriteFileArgs{
		LogicalName: "ROOT/a",
		Path:        "escape.go",
		Content:     "malicious",
	})

	if !isErr {
		t.Fatalf("expected tool error for symlink escape, got success: %s", text)
	}
	if !strings.Contains(text, "resolves outside project root") {
		t.Errorf("error message does not contain %q: %s", "resolves outside project root", text)
	}
}
