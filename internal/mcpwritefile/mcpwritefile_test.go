// code-from-spec: ROOT/golang/tests/mcp_tools/write_file@zI_PXcCsIBu5LcPd5DFDmxjIWv8

package mcpwritefile_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpwritefile"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// testChdir changes the working directory to dir for the duration of the test,
// restoring it on cleanup.
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

// testCreateNodeFile creates a _node.md file for a node at a given logical name
// (e.g., "ROOT/a") inside the temp dir. The frontmatter is provided as a raw
// YAML string (between --- delimiters).
func testCreateNodeFile(t *testing.T, tempDir string, logicalName string, frontmatterYAML string) {
	t.Helper()
	nodePath, err := logicalnames.LogicalNameToPath(logicalName)
	if err != nil {
		t.Fatalf("testCreateNodeFile: LogicalNameToPath(%q): %v", logicalName, err)
	}
	fullPath := filepath.Join(tempDir, filepath.FromSlash(nodePath.Value))
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testCreateNodeFile: MkdirAll(%q): %v", dir, err)
	}
	content := fmt.Sprintf("---\n%s\n---\n", frontmatterYAML)
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("testCreateNodeFile: WriteFile(%q): %v", fullPath, err)
	}
}

// TestMCPWriteFile_HappyPath_WritesFileSuccessfully covers TC-01.
func TestMCPWriteFile_HappyPath_WritesFileSuccessfully(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testCreateNodeFile(t, tempDir, "ROOT/a", "outputs:\n  - id: \"code\"\n    path: \"output/file.go\"")

	result, err := mcpwritefile.MCPWriteFile("ROOT/a", "output/file.go", "package main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "wrote output/file.go" {
		t.Errorf("expected %q, got %q", "wrote output/file.go", result)
	}

	writtenPath := filepath.Join(tempDir, "output", "file.go")
	data, err := os.ReadFile(writtenPath)
	if err != nil {
		t.Fatalf("file not found on disk: %v", err)
	}
	if string(data) != "package main" {
		t.Errorf("expected file content %q, got %q", "package main", string(data))
	}
}

// TestMCPWriteFile_HappyPath_CreatesIntermediateDirectories covers TC-02.
func TestMCPWriteFile_HappyPath_CreatesIntermediateDirectories(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testCreateNodeFile(t, tempDir, "ROOT/a", "outputs:\n  - id: \"code\"\n    path: \"deep/nested/dir/file.go\"")

	result, err := mcpwritefile.MCPWriteFile("ROOT/a", "deep/nested/dir/file.go", "package main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "wrote deep/nested/dir/file.go" {
		t.Errorf("expected %q, got %q", "wrote deep/nested/dir/file.go", result)
	}

	writtenPath := filepath.Join(tempDir, "deep", "nested", "dir", "file.go")
	data, err := os.ReadFile(writtenPath)
	if err != nil {
		t.Fatalf("file not found on disk: %v", err)
	}
	if string(data) != "package main" {
		t.Errorf("expected file content %q, got %q", "package main", string(data))
	}
}

// TestMCPWriteFile_HappyPath_OverwritesExistingFile covers TC-03.
func TestMCPWriteFile_HappyPath_OverwritesExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testCreateNodeFile(t, tempDir, "ROOT/a", "outputs:\n  - id: \"code\"\n    path: \"output/file.go\"")

	// Create the file with old content.
	outDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "file.go"), []byte("old"), 0644); err != nil {
		t.Fatalf("WriteFile (old content): %v", err)
	}

	result, err := mcpwritefile.MCPWriteFile("ROOT/a", "output/file.go", "new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "wrote output/file.go" {
		t.Errorf("expected %q, got %q", "wrote output/file.go", result)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "file.go"))
	if err != nil {
		t.Fatalf("file not found on disk: %v", err)
	}
	if string(data) != "new" {
		t.Errorf("expected file content %q, got %q", "new", string(data))
	}
}

// TestMCPWriteFile_Error_InvalidLogicalName_ArtifactReference covers TC-04.
func TestMCPWriteFile_Error_InvalidLogicalName_ArtifactReference(t *testing.T) {
	_, err := mcpwritefile.MCPWriteFile("ARTIFACT/x(y)", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("expected ErrUnsupportedReference, got: %v", err)
	}
}

// TestMCPWriteFile_Error_InvalidLogicalName_WithQualifier covers TC-05.
func TestMCPWriteFile_Error_InvalidLogicalName_WithQualifier(t *testing.T) {
	_, err := mcpwritefile.MCPWriteFile("ROOT/a(interface)", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("expected ErrUnsupportedReference, got: %v", err)
	}
}

// TestMCPWriteFile_Error_NonexistentNodeFile covers TC-06.
func TestMCPWriteFile_Error_NonexistentNodeFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// No _node.md file created for ROOT/missing.
	_, err := mcpwritefile.MCPWriteFile("ROOT/missing", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrUnreadableFrontmatter) {
		t.Errorf("expected ErrUnreadableFrontmatter, got: %v", err)
	}
}

// TestMCPWriteFile_Error_NoOutputsDeclared covers TC-07.
func TestMCPWriteFile_Error_NoOutputsDeclared(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create a node file with empty frontmatter (no outputs field).
	testCreateNodeFile(t, tempDir, "ROOT/a", "")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrNoOutputs) {
		t.Errorf("expected ErrNoOutputs, got: %v", err)
	}
}

// TestMCPWriteFile_Error_PathNotInOutputs covers TC-08.
func TestMCPWriteFile_Error_PathNotInOutputs(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testCreateNodeFile(t, tempDir, "ROOT/a", "outputs:\n  - id: \"code\"\n    path: \"allowed/file.go\"")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "other/file.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrPathNotInOutputs) {
		t.Errorf("expected ErrPathNotInOutputs, got: %v", err)
	}
}

// TestMCPWriteFile_Error_EmptyPath covers TC-09.
func TestMCPWriteFile_Error_EmptyPath(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testCreateNodeFile(t, tempDir, "ROOT/a", "outputs:\n  - id: \"code\"\n    path: \"out.go\"")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrPathEmpty) {
		t.Errorf("expected ErrPathEmpty, got: %v", err)
	}
}

// TestMCPWriteFile_Error_DirectoryTraversal covers TC-10.
func TestMCPWriteFile_Error_DirectoryTraversal(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testCreateNodeFile(t, tempDir, "ROOT/a", "outputs:\n  - id: \"code\"\n    path: \"out.go\"")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "../../etc/passwd", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got: %v", err)
	}
}

// TestMCPWriteFile_Error_BackslashInPath covers TC-11.
func TestMCPWriteFile_Error_BackslashInPath(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testCreateNodeFile(t, tempDir, "ROOT/a", "outputs:\n  - id: \"code\"\n    path: \"out.go\"")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "output\\file.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrPathContainsBackslash) {
		t.Errorf("expected ErrPathContainsBackslash, got: %v", err)
	}
}
