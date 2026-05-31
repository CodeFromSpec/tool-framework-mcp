// code-from-spec: ROOT/golang/tests/mcp_tools/write_file@M3flOA1tN3Sedg0zmtwzqDvQsBY
package mcpwritefile_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpwritefile"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// testChdir changes the working directory to dir and registers a cleanup
// that restores the original directory.
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

// testCreateNodeMd creates a _node.md file at the path corresponding to the
// given ROOT/ logical name, writing the given YAML frontmatter body.
// The file is placed under code-from-spec/<...>/_node.md relative to the cwd.
func testCreateNodeMd(t *testing.T, logicalName string, frontmatterYAML string) {
	t.Helper()
	cfsPath, err := logicalnames.LogicalNameToPath(logicalName)
	if err != nil {
		t.Fatalf("testCreateNodeMd: LogicalNameToPath(%q): %v", logicalName, err)
	}
	// Create directories
	dir := filepath.Dir(cfsPath.Value)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testCreateNodeMd: MkdirAll(%q): %v", dir, err)
	}
	content := "---\n" + frontmatterYAML + "---\n"
	if err := os.WriteFile(cfsPath.Value, []byte(content), 0644); err != nil {
		t.Fatalf("testCreateNodeMd: WriteFile(%q): %v", cfsPath.Value, err)
	}
}

// TestMCPWriteFile_WritesFileSuccessfully verifies that MCPWriteFile writes
// a file and returns "wrote <path>" when the path is declared in outputs.
func TestMCPWriteFile_WritesFileSuccessfully(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testCreateNodeMd(t, "ROOT/a", "outputs:\n  - id: code\n    path: output/file.go\n")

	result, err := mcpwritefile.MCPWriteFile("ROOT/a", "output/file.go", "package main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "wrote output/file.go" {
		t.Errorf("result = %q, want %q", result, "wrote output/file.go")
	}

	got, err := os.ReadFile(filepath.Join(tempDir, "output", "file.go"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "package main" {
		t.Errorf("file content = %q, want %q", string(got), "package main")
	}
}

// TestMCPWriteFile_CreatesIntermediateDirectories verifies that MCPWriteFile
// creates any missing intermediate directories.
func TestMCPWriteFile_CreatesIntermediateDirectories(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testCreateNodeMd(t, "ROOT/a", "outputs:\n  - id: code\n    path: deep/nested/dir/file.go\n")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "deep/nested/dir/file.go", "package main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(tempDir, "deep", "nested", "dir", "file.go"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "package main" {
		t.Errorf("file content = %q, want %q", string(got), "package main")
	}
}

// TestMCPWriteFile_OverwritesExistingFile verifies that MCPWriteFile overwrites
// a file that already exists on disk.
func TestMCPWriteFile_OverwritesExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testCreateNodeMd(t, "ROOT/a", "outputs:\n  - id: code\n    path: output/file.go\n")

	// Pre-create the file with old content.
	if err := os.MkdirAll(filepath.Join(tempDir, "output"), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "output", "file.go"), []byte("old"), 0644); err != nil {
		t.Fatalf("WriteFile (old): %v", err)
	}

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "output/file.go", "new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(tempDir, "output", "file.go"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "new" {
		t.Errorf("file content = %q, want %q", string(got), "new")
	}
}

// TestMCPWriteFile_InvalidLogicalName_ArtifactReference verifies that an
// ARTIFACT/ logical name returns an ErrUnsupportedReference error.
func TestMCPWriteFile_InvalidLogicalName_ArtifactReference(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := mcpwritefile.MCPWriteFile("ARTIFACT/x(y)", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("error = %v, want errors.Is ErrUnsupportedReference", err)
	}
}

// TestMCPWriteFile_InvalidLogicalName_WithQualifier verifies that a ROOT/
// logical name with a qualifier that doesn't resolve to an existing node
// returns an ErrUnreadableFrontmatter error.
func TestMCPWriteFile_InvalidLogicalName_WithQualifier(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := mcpwritefile.MCPWriteFile("ROOT/a(interface)", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrUnreadableFrontmatter) {
		t.Errorf("error = %v, want errors.Is ErrUnreadableFrontmatter", err)
	}
}

// TestMCPWriteFile_NonexistentNodeFile verifies that a missing _node.md
// returns ErrUnreadableFrontmatter.
func TestMCPWriteFile_NonexistentNodeFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := mcpwritefile.MCPWriteFile("ROOT/missing", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrUnreadableFrontmatter) {
		t.Errorf("error = %v, want errors.Is ErrUnreadableFrontmatter", err)
	}
}

// TestMCPWriteFile_NoOutputsDeclared verifies that a node with no outputs
// field returns ErrNoOutputs.
func TestMCPWriteFile_NoOutputsDeclared(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Empty frontmatter — no outputs field.
	testCreateNodeMd(t, "ROOT/a", "")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrNoOutputs) {
		t.Errorf("error = %v, want errors.Is ErrNoOutputs", err)
	}
}

// TestMCPWriteFile_PathNotInOutputs verifies that a path not listed in
// the node's outputs returns ErrPathNotInOutputs.
func TestMCPWriteFile_PathNotInOutputs(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testCreateNodeMd(t, "ROOT/a", "outputs:\n  - id: code\n    path: allowed/file.go\n")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "other/file.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrPathNotInOutputs) {
		t.Errorf("error = %v, want errors.Is ErrPathNotInOutputs", err)
	}
}

// TestMCPWriteFile_PathValidation_EmptyPath verifies that an empty path
// returns ErrPathIsEmpty (from pathutils).
func TestMCPWriteFile_PathValidation_EmptyPath(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testCreateNodeMd(t, "ROOT/a", "outputs:\n  - id: code\n    path: out.go\n")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrPathIsEmpty) {
		t.Errorf("error = %v, want errors.Is ErrPathIsEmpty", err)
	}
}

// TestMCPWriteFile_PathValidation_DirectoryTraversal verifies that a path
// containing ".." returns ErrDirectoryTraversal (from pathutils).
func TestMCPWriteFile_PathValidation_DirectoryTraversal(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testCreateNodeMd(t, "ROOT/a", "outputs:\n  - id: code\n    path: out.go\n")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "../../etc/passwd", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("error = %v, want errors.Is ErrDirectoryTraversal", err)
	}
}

// TestMCPWriteFile_PathValidation_BackslashInPath verifies that a path
// with backslashes returns ErrPathContainsBackslash (from pathutils).
func TestMCPWriteFile_PathValidation_BackslashInPath(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testCreateNodeMd(t, "ROOT/a", "outputs:\n  - id: code\n    path: out.go\n")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", `output\file.go`, "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrPathContainsBackslash) {
		t.Errorf("error = %v, want errors.Is ErrPathContainsBackslash", err)
	}
}

// Ensure the frontmatter package import is used (it may be referenced
// transitively; this declaration keeps the compiler satisfied).
var _ = frontmatter.ErrFileUnreadable
