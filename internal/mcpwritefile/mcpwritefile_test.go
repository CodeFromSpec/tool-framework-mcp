// code-from-spec: ROOT/golang/tests/mcp_tools/write_file@gmdwGb9hzK1X2PNqF_BhQX0dMp8
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

func testWriteNodeFile(t *testing.T, logicalName string, frontmatter string) {
	t.Helper()
	nodePath, err := logicalnames.LogicalNameToPath(logicalName)
	if err != nil {
		t.Fatalf("testWriteNodeFile: LogicalNameToPath(%q): %v", logicalName, err)
	}
	dir := filepath.Dir(nodePath.Value)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testWriteNodeFile: MkdirAll(%q): %v", dir, err)
	}
	content := "---\n" + frontmatter + "---\n"
	if err := os.WriteFile(nodePath.Value, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNodeFile: WriteFile(%q): %v", nodePath.Value, err)
	}
}

func TestMCPWriteFile_TC01_WritesFileSuccessfully(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/a", "outputs:\n  - id: \"code\"\n    path: \"output/file.go\"\n")

	result, err := mcpwritefile.MCPWriteFile("ROOT/a", "output/file.go", "package main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "wrote output/file.go" {
		t.Errorf("result = %q, want %q", result, "wrote output/file.go")
	}

	root, err := pathutils.PathGetProjectRoot()
	if err != nil {
		t.Fatalf("PathGetProjectRoot: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root.Value, "output", "file.go"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "package main" {
		t.Errorf("file content = %q, want %q", string(data), "package main")
	}
}

func TestMCPWriteFile_TC02_CreatesIntermediateDirectories(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/a", "outputs:\n  - id: \"code\"\n    path: \"deep/nested/dir/file.go\"\n")

	result, err := mcpwritefile.MCPWriteFile("ROOT/a", "deep/nested/dir/file.go", "package main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "wrote deep/nested/dir/file.go" {
		t.Errorf("result = %q, want %q", result, "wrote deep/nested/dir/file.go")
	}

	root, err := pathutils.PathGetProjectRoot()
	if err != nil {
		t.Fatalf("PathGetProjectRoot: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root.Value, "deep", "nested", "dir", "file.go"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "package main" {
		t.Errorf("file content = %q, want %q", string(data), "package main")
	}
}

func TestMCPWriteFile_TC03_OverwritesExistingFile(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/a", "outputs:\n  - id: \"code\"\n    path: \"output/file.go\"\n")

	if err := os.MkdirAll("output", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile("output/file.go", []byte("old"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	result, err := mcpwritefile.MCPWriteFile("ROOT/a", "output/file.go", "new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "wrote output/file.go" {
		t.Errorf("result = %q, want %q", result, "wrote output/file.go")
	}

	data, err := os.ReadFile("output/file.go")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "new" {
		t.Errorf("file content = %q, want %q", string(data), "new")
	}
}

func TestMCPWriteFile_TC04_InvalidLogicalName_ArtifactReference(t *testing.T) {
	_, err := mcpwritefile.MCPWriteFile("ARTIFACT/x(y)", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("error = %v, want ErrUnsupportedReference", err)
	}
}

func TestMCPWriteFile_TC05_InvalidLogicalName_WithQualifier(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := mcpwritefile.MCPWriteFile("ROOT/a(interface)", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrUnreadableFrontmatter) {
		t.Errorf("error = %v, want ErrUnreadableFrontmatter", err)
	}
}

func TestMCPWriteFile_TC06_NonexistentNodeFile(t *testing.T) {
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

func TestMCPWriteFile_TC07_NoOutputsDeclared(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/a", "")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrNoOutputs) {
		t.Errorf("error = %v, want ErrNoOutputs", err)
	}
}

func TestMCPWriteFile_TC08_PathNotInOutputs(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/a", "outputs:\n  - id: \"code\"\n    path: \"allowed/file.go\"\n")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "other/file.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrPathNotInOutputs) {
		t.Errorf("error = %v, want ErrPathNotInOutputs", err)
	}
}

func TestMCPWriteFile_TC09_PathValidation_EmptyPath(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/a", "outputs:\n  - id: \"code\"\n    path: \"out.go\"\n")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrPathEmpty) {
		t.Errorf("error = %v, want ErrPathEmpty", err)
	}
}

func TestMCPWriteFile_TC10_PathValidation_DirectoryTraversal(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/a", "outputs:\n  - id: \"code\"\n    path: \"out.go\"\n")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "../../etc/passwd", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("error = %v, want ErrDirectoryTraversal", err)
	}
}

func TestMCPWriteFile_TC11_PathValidation_Backslash(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/a", "outputs:\n  - id: \"code\"\n    path: \"out.go\"\n")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "output\\file.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrPathContainsBackslash) {
		t.Errorf("error = %v, want ErrPathContainsBackslash", err)
	}
}
