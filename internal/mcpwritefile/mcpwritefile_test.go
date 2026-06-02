// code-from-spec: ROOT/golang/tests/mcp_tools/write_file@2U4XFZUrXDs6rjWTzVkpS3CuafM
package mcpwritefile_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpwritefile"
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

func testWriteNode(t *testing.T, logicalName string, frontmatter string) {
	t.Helper()
	segments := logicalName[len("ROOT"):]
	var dir string
	if segments == "" {
		dir = "code-from-spec"
	} else {
		dir = filepath.Join("code-from-spec", filepath.FromSlash(segments[1:]))
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testWriteNode mkdir %s: %v", logicalName, err)
	}
	content := "---\n" + frontmatter + "---\n\n# " + logicalName + "\n"
	if err := os.WriteFile(filepath.Join(dir, "_node.md"), []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNode write %s: %v", logicalName, err)
	}
}

func TestMCPWriteFile_WritesFileSuccessfully(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "output: output/file.go\n")

	result, err := mcpwritefile.MCPWriteFile("ROOT/a", "output/file.go", "package main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "wrote output/file.go" {
		t.Errorf("expected 'wrote output/file.go', got %q", result)
	}
	data, err := os.ReadFile(filepath.Join("output", "file.go"))
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}
	if string(data) != "package main" {
		t.Errorf("expected content 'package main', got %q", string(data))
	}
}

func TestMCPWriteFile_CreatesIntermediateDirectories(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "output: deep/nested/dir/file.go\n")

	result, err := mcpwritefile.MCPWriteFile("ROOT/a", "deep/nested/dir/file.go", "package main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "wrote deep/nested/dir/file.go" {
		t.Errorf("expected 'wrote deep/nested/dir/file.go', got %q", result)
	}
	if _, err := os.Stat(filepath.Join("deep", "nested", "dir", "file.go")); err != nil {
		t.Errorf("expected file to exist: %v", err)
	}
}

func TestMCPWriteFile_OverwritesExistingFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "output: output/file.go\n")

	if err := os.MkdirAll("output", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join("output", "file.go"), []byte("old"), 0644); err != nil {
		t.Fatalf("write initial: %v", err)
	}

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "output/file.go", "new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(filepath.Join("output", "file.go"))
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(data) != "new" {
		t.Errorf("expected content 'new', got %q", string(data))
	}
}

func TestMCPWriteFile_InvalidLogicalName_ArtifactReference(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := mcpwritefile.MCPWriteFile("ARTIFACT/x", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("expected ErrUnsupportedReference, got %v", err)
	}
}

func TestMCPWriteFile_NonexistentNodeFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")

	_, err := mcpwritefile.MCPWriteFile("ROOT/missing", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrUnreadableFrontmatter) {
		t.Errorf("expected ErrUnreadableFrontmatter, got %v", err)
	}
}

func TestMCPWriteFile_NoOutputDeclared(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrNoOutput) {
		t.Errorf("expected ErrNoOutput, got %v", err)
	}
}

func TestMCPWriteFile_PathNotInOutput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "output: allowed/file.go\n")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "other/file.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrPathNotInOutput) {
		t.Errorf("expected ErrPathNotInOutput, got %v", err)
	}
}

func TestMCPWriteFile_PathValidation_EmptyPath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "output: out.go\n")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrPathEmpty) {
		t.Errorf("expected ErrPathEmpty, got %v", err)
	}
}

func TestMCPWriteFile_PathValidation_Traversal(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "output: out.go\n")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "../../etc/passwd", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got %v", err)
	}
}

func TestMCPWriteFile_PathValidation_Backslash(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "output: out.go\n")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", `output\file.go`, "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrPathContainsBackslash) {
		t.Errorf("expected ErrPathContainsBackslash, got %v", err)
	}
}
