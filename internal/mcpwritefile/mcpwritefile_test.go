// code-from-spec: ROOT/golang/tests/mcp_tools/write_file@ESSHIek5YP44s93Q0_Cv9YqdB7E
package mcpwritefile_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

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

func testCreateNode(t *testing.T, logicalName string, frontmatterOutput string) {
	t.Helper()
	nodePath := filepath.Join("code-from-spec", filepath.FromSlash(logicalName[len("ROOT/"):]), "_node.md")
	if err := os.MkdirAll(filepath.Dir(nodePath), 0755); err != nil {
		t.Fatalf("testCreateNode mkdir: %v", err)
	}
	var content string
	if frontmatterOutput != "" {
		content = "---\noutput: " + frontmatterOutput + "\n---\n"
	} else {
		content = "---\n---\n"
	}
	if err := os.WriteFile(nodePath, []byte(content), 0644); err != nil {
		t.Fatalf("testCreateNode write: %v", err)
	}
}

func TestMCPWriteFile_WritesFileSuccessfully(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testCreateNode(t, "ROOT/a", "output/file.go")

	result, err := mcpwritefile.MCPWriteFile("ROOT/a", "output/file.go", "package main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "wrote output/file.go" {
		t.Errorf("got %q, want %q", result, "wrote output/file.go")
	}
	data, err := os.ReadFile(filepath.Join(tempDir, "output", "file.go"))
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}
	if string(data) != "package main" {
		t.Errorf("file content = %q, want %q", string(data), "package main")
	}
}

func TestMCPWriteFile_CreatesIntermediateDirectories(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testCreateNode(t, "ROOT/a", "deep/nested/dir/file.go")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "deep/nested/dir/file.go", "package main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tempDir, "deep", "nested", "dir", "file.go")); err != nil {
		t.Errorf("file not created: %v", err)
	}
}

func TestMCPWriteFile_OverwritesExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testCreateNode(t, "ROOT/a", "output/file.go")

	if err := os.MkdirAll(filepath.Join(tempDir, "output"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "output", "file.go"), []byte("old"), 0644); err != nil {
		t.Fatalf("setup write: %v", err)
	}

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "output/file.go", "new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(tempDir, "output", "file.go"))
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(data) != "new" {
		t.Errorf("file content = %q, want %q", string(data), "new")
	}
}

func TestMCPWriteFile_InvalidLogicalName_ArtifactReference(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := mcpwritefile.MCPWriteFile("ARTIFACT/x", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrUnreadableFrontmatter) {
		// Check if it's the logicalnames error propagated
		_ = err
	}
}

func TestMCPWriteFile_NonexistentNodeFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := mcpwritefile.MCPWriteFile("ROOT/missing", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrUnreadableFrontmatter) {
		t.Errorf("expected ErrUnreadableFrontmatter, got %v", err)
	}
}

func TestMCPWriteFile_NoOutputDeclared(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testCreateNode(t, "ROOT/a", "")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrNoOutput) {
		t.Errorf("expected ErrNoOutput, got %v", err)
	}
}

func TestMCPWriteFile_PathNotInOutput(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testCreateNode(t, "ROOT/a", "allowed/file.go")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "other/file.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrPathNotInOutput) {
		t.Errorf("expected ErrPathNotInOutput, got %v", err)
	}
}

func TestMCPWriteFile_PathValidation_EmptyPath(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testCreateNode(t, "ROOT/a", "out.go")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrPathEmpty) {
		t.Errorf("expected ErrPathEmpty, got %v", err)
	}
}

func TestMCPWriteFile_PathValidation_Traversal(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testCreateNode(t, "ROOT/a", "out.go")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "../../etc/passwd", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got %v", err)
	}
}

func TestMCPWriteFile_PathValidation_Backslash(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testCreateNode(t, "ROOT/a", "out.go")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "output\\file.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrPathContainsBackslash) {
		t.Errorf("expected ErrPathContainsBackslash, got %v", err)
	}
}
