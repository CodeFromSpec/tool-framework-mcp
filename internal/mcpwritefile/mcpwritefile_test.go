// code-from-spec: ROOT/golang/tests/mcp_tools/write_file@z8q241i_rxNflq1ECT0Pe6oSH9Y
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

// testChdir changes the working directory to dir and restores it on cleanup.
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

// testWriteNodeMd creates the _node.md file for the given node logical name
// (must start with ROOT/) inside the temp dir, writing the provided frontmatter
// content. Intermediate directories are created as needed.
func testWriteNodeMd(t *testing.T, logicalName string, frontmatter string) {
	t.Helper()
	cfsPath, err := logicalnames.LogicalNameToPath(logicalName)
	if err != nil {
		t.Fatalf("testWriteNodeMd LogicalNameToPath: %v", err)
	}
	osPath, err := pathutils.PathCfsToOs(cfsPath)
	if err != nil {
		t.Fatalf("testWriteNodeMd PathCfsToOs: %v", err)
	}
	dir := filepath.Dir(osPath.Value)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("testWriteNodeMd MkdirAll: %v", err)
	}
	content := "---\n" + frontmatter + "---\n"
	if err := os.WriteFile(osPath.Value, []byte(content), 0o644); err != nil {
		t.Fatalf("testWriteNodeMd WriteFile: %v", err)
	}
}

// TC-01: Writes file successfully.
func TestMCPWriteFile_WritesFileSuccessfully(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeMd(t, "ROOT/a", "outputs:\n  - id: \"code\"\n    path: \"output/file.go\"\n")

	result, err := mcpwritefile.MCPWriteFile("ROOT/a", "output/file.go", "package main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "wrote output/file.go" {
		t.Errorf("result = %q, want %q", result, "wrote output/file.go")
	}

	osPath, err := pathutils.PathCfsToOs(&pathutils.PathCfs{Value: "output/file.go"})
	if err != nil {
		t.Fatalf("PathCfsToOs: %v", err)
	}
	data, err := os.ReadFile(osPath.Value)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "package main" {
		t.Errorf("file content = %q, want %q", string(data), "package main")
	}
}

// TC-02: Creates intermediate directories.
func TestMCPWriteFile_CreatesIntermediateDirectories(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeMd(t, "ROOT/a", "outputs:\n  - id: \"code\"\n    path: \"deep/nested/dir/file.go\"\n")

	result, err := mcpwritefile.MCPWriteFile("ROOT/a", "deep/nested/dir/file.go", "package main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "wrote deep/nested/dir/file.go" {
		t.Errorf("result = %q, want %q", result, "wrote deep/nested/dir/file.go")
	}

	osPath, err := pathutils.PathCfsToOs(&pathutils.PathCfs{Value: "deep/nested/dir/file.go"})
	if err != nil {
		t.Fatalf("PathCfsToOs: %v", err)
	}
	data, err := os.ReadFile(osPath.Value)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "package main" {
		t.Errorf("file content = %q, want %q", string(data), "package main")
	}
}

// TC-03: Overwrites existing file.
func TestMCPWriteFile_OverwritesExistingFile(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeMd(t, "ROOT/a", "outputs:\n  - id: \"code\"\n    path: \"output/file.go\"\n")

	// Pre-create the file with old content.
	if err := os.MkdirAll(filepath.Join(tmp, "output"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "output", "file.go"), []byte("old"), 0o644); err != nil {
		t.Fatalf("WriteFile (setup): %v", err)
	}

	result, err := mcpwritefile.MCPWriteFile("ROOT/a", "output/file.go", "new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "wrote output/file.go" {
		t.Errorf("result = %q, want %q", result, "wrote output/file.go")
	}

	data, err := os.ReadFile(filepath.Join(tmp, "output", "file.go"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "new" {
		t.Errorf("file content = %q, want %q", string(data), "new")
	}
}

// TC-04: Invalid logical name — ARTIFACT reference.
func TestMCPWriteFile_ArtifactReference(t *testing.T) {
	_, err := mcpwritefile.MCPWriteFile("ARTIFACT/x(y)", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("error = %v, want ErrUnsupportedReference", err)
	}
}

// TC-05: Invalid logical name — with qualifier.
func TestMCPWriteFile_QualifiedLogicalName(t *testing.T) {
	_, err := mcpwritefile.MCPWriteFile("ROOT/a(interface)", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Qualifiers are stripped before path resolution; the node file won't
	// exist, so frontmatter parsing fails with ErrUnreadableFrontmatter.
	if !errors.Is(err, mcpwritefile.ErrUnreadableFrontmatter) {
		t.Errorf("error = %v, want ErrUnreadableFrontmatter", err)
	}
}

// TC-06: Nonexistent node file.
func TestMCPWriteFile_NonexistentNodeFile(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := mcpwritefile.MCPWriteFile("ROOT/missing", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrUnreadableFrontmatter) {
		t.Errorf("error = %v, want ErrUnreadableFrontmatter", err)
	}
}

// TC-07: No outputs declared.
func TestMCPWriteFile_NoOutputsDeclared(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeMd(t, "ROOT/a", "")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrNoOutputs) {
		t.Errorf("error = %v, want ErrNoOutputs", err)
	}
}

// TC-08: Path not in outputs.
func TestMCPWriteFile_PathNotInOutputs(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeMd(t, "ROOT/a", "outputs:\n  - id: \"code\"\n    path: \"allowed/file.go\"\n")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "other/file.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrPathNotInOutputs) {
		t.Errorf("error = %v, want ErrPathNotInOutputs", err)
	}
}

// TC-09: Path validation — empty path.
func TestMCPWriteFile_EmptyPath(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeMd(t, "ROOT/a", "outputs:\n  - id: \"code\"\n    path: \"out.go\"\n")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrPathEmpty) {
		t.Errorf("error = %v, want ErrPathEmpty", err)
	}
}

// TC-10: Path validation — directory traversal.
func TestMCPWriteFile_DirectoryTraversal(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeMd(t, "ROOT/a", "outputs:\n  - id: \"code\"\n    path: \"out.go\"\n")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "../../etc/passwd", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("error = %v, want ErrDirectoryTraversal", err)
	}
}

// TC-11: Path validation — backslash.
func TestMCPWriteFile_Backslash(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeMd(t, "ROOT/a", "outputs:\n  - id: \"code\"\n    path: \"out.go\"\n")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", `output\file.go`, "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrPathContainsBackslash) {
		t.Errorf("error = %v, want ErrPathContainsBackslash", err)
	}
}
