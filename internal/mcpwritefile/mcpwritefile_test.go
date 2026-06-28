// code-from-spec: SPEC/golang/tests/mcp_tools/write_file@b6qs86jGTMEq0Y3TJG1JSrxkGl8
package mcpwritefile_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcpwritefile"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
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

func testWriteNodeFile(t *testing.T, relPath string, content string) {
	t.Helper()
	if err := os.MkdirAll(relPath[:len(relPath)-len("_node.md")], 0755); err != nil {
		t.Fatalf("testWriteNodeFile mkdir: %v", err)
	}
	if err := os.WriteFile(relPath, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNodeFile write: %v", err)
	}
}

func TestMCPWriteFile_TC01_WritesFileSuccessfully(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "# SPEC\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "---\noutput: output/file.go\n---\n# SPEC/a\n")

	result, err := mcpwritefile.MCPWriteFile("SPEC/a", "output/file.go", "package main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "wrote output/file.go" {
		t.Errorf("result = %q, want %q", result, "wrote output/file.go")
	}

	data, err := os.ReadFile("output/file.go")
	if err != nil {
		t.Fatalf("file not found: %v", err)
	}
	if string(data) != "package main" {
		t.Errorf("file content = %q, want %q", string(data), "package main")
	}
}

func TestMCPWriteFile_TC02_CreatesIntermediateDirectories(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "# SPEC\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "---\noutput: deep/nested/dir/file.go\n---\n# SPEC/a\n")

	result, err := mcpwritefile.MCPWriteFile("SPEC/a", "deep/nested/dir/file.go", "package main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty result")
	}

	if _, err := os.Stat("deep/nested/dir/file.go"); err != nil {
		t.Errorf("file not found: %v", err)
	}
}

func TestMCPWriteFile_TC03_OverwritesExistingFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "# SPEC\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "---\noutput: output/file.go\n---\n# SPEC/a\n")

	if err := os.MkdirAll("output", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("output/file.go", []byte("old"), 0644); err != nil {
		t.Fatalf("write existing file: %v", err)
	}

	result, err := mcpwritefile.MCPWriteFile("SPEC/a", "output/file.go", "new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty result")
	}

	data, err := os.ReadFile("output/file.go")
	if err != nil {
		t.Fatalf("file not found: %v", err)
	}
	if string(data) != "new" {
		t.Errorf("file content = %q, want %q", string(data), "new")
	}
}

func TestMCPWriteFile_TC04_InvalidLogicalName_ArtifactReference(t *testing.T) {
	_, err := mcpwritefile.MCPWriteFile("ARTIFACT/x", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("expected ErrUnsupportedReference, got: %v", err)
	}
}

func TestMCPWriteFile_TC05_InvalidLogicalName_WithQualifier(t *testing.T) {
	_, err := mcpwritefile.MCPWriteFile("SPEC/a(interface)", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrQualifierNotAllowed) {
		t.Errorf("expected ErrQualifierNotAllowed, got: %v", err)
	}
}

func TestMCPWriteFile_TC06_NonexistentNodeFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := mcpwritefile.MCPWriteFile("SPEC/missing", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrUnreadableFrontmatter) {
		t.Errorf("expected ErrUnreadableFrontmatter, got: %v", err)
	}
}

func TestMCPWriteFile_TC07_NoOutputDeclared(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "# SPEC\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n")

	_, err := mcpwritefile.MCPWriteFile("SPEC/a", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrNoOutput) {
		t.Errorf("expected ErrNoOutput, got: %v", err)
	}
}

func TestMCPWriteFile_TC08_PathNotInOutput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "# SPEC\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "---\noutput: allowed/file.go\n---\n# SPEC/a\n")

	_, err := mcpwritefile.MCPWriteFile("SPEC/a", "other/file.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrPathNotInOutput) {
		t.Errorf("expected ErrPathNotInOutput, got: %v", err)
	}
}

func TestMCPWriteFile_TC09_PathValidation_EmptyPath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "# SPEC\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "---\noutput: out.go\n---\n# SPEC/a\n")

	_, err := mcpwritefile.MCPWriteFile("SPEC/a", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrPathEmpty) {
		t.Errorf("expected ErrPathEmpty, got: %v", err)
	}
}

func TestMCPWriteFile_TC10_PathValidation_Traversal(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "# SPEC\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "---\noutput: out.go\n---\n# SPEC/a\n")

	_, err := mcpwritefile.MCPWriteFile("SPEC/a", "../../etc/passwd", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got: %v", err)
	}
}

func TestMCPWriteFile_TC11_PathValidation_Backslash(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "# SPEC\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "---\noutput: out.go\n---\n# SPEC/a\n")

	_, err := mcpwritefile.MCPWriteFile("SPEC/a", `output\file.go`, "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrPathContainsBackslash) {
		t.Errorf("expected ErrPathContainsBackslash, got: %v", err)
	}
}
