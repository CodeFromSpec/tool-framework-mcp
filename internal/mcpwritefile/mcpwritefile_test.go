// code-from-spec: ROOT/golang/tests/mcp_tools/write_file@sFdQkjbcVY_M24QKTbDsm15XlmA
package mcpwritefile_test

import (
	"errors"
	"os"
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

func testCreateNodeFile(t *testing.T, logicalName string, frontmatterContent string) {
	t.Helper()
	nodePath, err := logicalnames.LogicalNameToPath(logicalName)
	if err != nil {
		t.Fatalf("testCreateNodeFile: LogicalNameToPath: %v", err)
	}
	osPath, err := pathutils.PathCfsToOs(nodePath)
	if err != nil {
		t.Fatalf("testCreateNodeFile: PathCfsToOs: %v", err)
	}
	dir := osPath.Value[:len(osPath.Value)-len("_node.md")]
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testCreateNodeFile: MkdirAll: %v", err)
	}
	content := "---\n" + frontmatterContent + "\n---\n"
	if err := os.WriteFile(osPath.Value, []byte(content), 0644); err != nil {
		t.Fatalf("testCreateNodeFile: WriteFile: %v", err)
	}
}

func TestMCPWriteFile_WritesFileSuccessfully(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testCreateNodeFile(t, "ROOT/a", "output: output/file.go")

	result, err := mcpwritefile.MCPWriteFile("ROOT/a", "output/file.go", "package main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "wrote output/file.go" {
		t.Errorf("expected 'wrote output/file.go', got %q", result)
	}

	data, err := os.ReadFile("output/file.go")
	if err != nil {
		t.Fatalf("file not found: %v", err)
	}
	if string(data) != "package main" {
		t.Errorf("expected 'package main', got %q", string(data))
	}
}

func TestMCPWriteFile_CreatesIntermediateDirectories(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testCreateNodeFile(t, "ROOT/a", "output: deep/nested/dir/file.go")

	result, err := mcpwritefile.MCPWriteFile("ROOT/a", "deep/nested/dir/file.go", "package main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "wrote deep/nested/dir/file.go" {
		t.Errorf("expected 'wrote deep/nested/dir/file.go', got %q", result)
	}

	if _, err := os.Stat("deep/nested/dir/file.go"); err != nil {
		t.Errorf("file does not exist: %v", err)
	}
}

func TestMCPWriteFile_OverwritesExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testCreateNodeFile(t, "ROOT/a", "output: output/file.go")
	if err := os.MkdirAll("output", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile("output/file.go", []byte("old"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "output/file.go", "new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile("output/file.go")
	if err != nil {
		t.Fatalf("file not found: %v", err)
	}
	if string(data) != "new" {
		t.Errorf("expected 'new', got %q", string(data))
	}
}

func TestMCPWriteFile_InvalidLogicalName_ArtifactReference(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := mcpwritefile.MCPWriteFile("ARTIFACT/x", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("expected ErrUnsupportedReference, got %v", err)
	}
}

func TestMCPWriteFile_InvalidLogicalName_WithQualifier(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := mcpwritefile.MCPWriteFile("ROOT/a(interface)", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrUnreadableFrontmatter) {
		t.Errorf("expected ErrUnreadableFrontmatter, got %v", err)
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

	testCreateNodeFile(t, "ROOT/a", "")

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

	testCreateNodeFile(t, "ROOT/a", "output: allowed/file.go")

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

	testCreateNodeFile(t, "ROOT/a", "output: out.go")

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

	testCreateNodeFile(t, "ROOT/a", "output: out.go")

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

	testCreateNodeFile(t, "ROOT/a", "output: out.go")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "output\\file.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrPathContainsBackslash) {
		t.Errorf("expected ErrPathContainsBackslash, got %v", err)
	}
}
