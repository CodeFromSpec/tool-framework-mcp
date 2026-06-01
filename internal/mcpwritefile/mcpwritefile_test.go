// code-from-spec: ROOT/golang/tests/mcp_tools/write_file@zI_PXcCsIBu5LcPd5DFDmxjIWv8
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

// testChdir changes the working directory to dir for the duration of
// the test. Registers a cleanup to restore the original directory.
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

// testWriteNodeFile creates the _node.md file for the given CFS path
// (relative to the working directory) and writes frontmatterContent
// between YAML front matter delimiters.
func testWriteNodeFile(t *testing.T, cfsDirPath string, frontmatterContent string) {
	t.Helper()
	if err := os.MkdirAll(cfsDirPath, 0755); err != nil {
		t.Fatalf("testWriteNodeFile MkdirAll: %v", err)
	}
	nodePath := filepath.Join(cfsDirPath, "_node.md")
	body := "---\n" + frontmatterContent + "---\n"
	if err := os.WriteFile(nodePath, []byte(body), 0644); err != nil {
		t.Fatalf("testWriteNodeFile WriteFile: %v", err)
	}
}

// TC-01: Writes file successfully.
func TestMCPWriteFile_WritesFileSuccessfully(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/a", "outputs:\n  - id: \"code\"\n    path: \"output/file.go\"\n")

	result, err := mcpwritefile.MCPWriteFile("ROOT/a", "output/file.go", "package main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "wrote output/file.go" {
		t.Errorf("got result %q, want %q", result, "wrote output/file.go")
	}

	got, err := os.ReadFile(filepath.Join(tempDir, "output", "file.go"))
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}
	if string(got) != "package main" {
		t.Errorf("file content = %q, want %q", string(got), "package main")
	}
}

// TC-02: Creates intermediate directories.
func TestMCPWriteFile_CreatesIntermediateDirectories(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/a", "outputs:\n  - id: \"code\"\n    path: \"deep/nested/dir/file.go\"\n")

	result, err := mcpwritefile.MCPWriteFile("ROOT/a", "deep/nested/dir/file.go", "package main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "wrote deep/nested/dir/file.go" {
		t.Errorf("got result %q, want %q", result, "wrote deep/nested/dir/file.go")
	}

	got, err := os.ReadFile(filepath.Join(tempDir, "deep", "nested", "dir", "file.go"))
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}
	if string(got) != "package main" {
		t.Errorf("file content = %q, want %q", string(got), "package main")
	}
}

// TC-03: Overwrites existing file.
func TestMCPWriteFile_OverwritesExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/a", "outputs:\n  - id: \"code\"\n    path: \"output/file.go\"\n")

	// Pre-create the file with old content.
	if err := os.MkdirAll(filepath.Join(tempDir, "output"), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "output", "file.go"), []byte("old"), 0644); err != nil {
		t.Fatalf("WriteFile pre-existing: %v", err)
	}

	result, err := mcpwritefile.MCPWriteFile("ROOT/a", "output/file.go", "new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "wrote output/file.go" {
		t.Errorf("got result %q, want %q", result, "wrote output/file.go")
	}

	got, err := os.ReadFile(filepath.Join(tempDir, "output", "file.go"))
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}
	if string(got) != "new" {
		t.Errorf("file content = %q, want %q", string(got), "new")
	}
}

// TC-04: Invalid logical name — ARTIFACT reference.
func TestMCPWriteFile_ArtifactReference_ReturnsUnsupportedReference(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := mcpwritefile.MCPWriteFile("ARTIFACT/x(y)", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("expected ErrUnsupportedReference, got: %v", err)
	}
}

// TC-05: Invalid logical name — with qualifier, node file does not exist.
// LogicalNameToPath strips the qualifier before resolving, so ROOT/a(interface)
// resolves to code-from-spec/a/_node.md. Since that file does not exist,
// the error is ErrUnreadableFrontmatter (from FrontmatterParse).
func TestMCPWriteFile_QualifiedLogicalName_NodeMissing_ReturnsUnreadableFrontmatter(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// No spec tree created — code-from-spec/a/_node.md does not exist.
	_, err := mcpwritefile.MCPWriteFile("ROOT/a(interface)", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrUnreadableFrontmatter) {
		t.Errorf("expected ErrUnreadableFrontmatter, got: %v", err)
	}
}

// TC-06: Nonexistent node file.
func TestMCPWriteFile_NonexistentNodeFile_ReturnsUnreadableFrontmatter(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// No spec tree created — code-from-spec/missing/_node.md does not exist.
	_, err := mcpwritefile.MCPWriteFile("ROOT/missing", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrUnreadableFrontmatter) {
		t.Errorf("expected ErrUnreadableFrontmatter, got: %v", err)
	}
}

// TC-07: No outputs declared.
func TestMCPWriteFile_NoOutputsDeclared_ReturnsNoOutputs(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Empty frontmatter — no outputs field.
	testWriteNodeFile(t, "code-from-spec/a", "")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "out.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrNoOutputs) {
		t.Errorf("expected ErrNoOutputs, got: %v", err)
	}
}

// TC-08: Path not in outputs.
func TestMCPWriteFile_PathNotInOutputs_ReturnsPathNotInOutputs(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/a", "outputs:\n  - id: \"code\"\n    path: \"allowed/file.go\"\n")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "other/file.go", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpwritefile.ErrPathNotInOutputs) {
		t.Errorf("expected ErrPathNotInOutputs, got: %v", err)
	}
}

// TC-09: Path validation — empty path.
func TestMCPWriteFile_EmptyPath_ReturnsPathEmpty(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/a", "outputs:\n  - id: \"code\"\n    path: \"out.go\"\n")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrPathEmpty) {
		t.Errorf("expected ErrPathEmpty, got: %v", err)
	}
}

// TC-10: Path validation — directory traversal.
func TestMCPWriteFile_DirectoryTraversal_ReturnsDirectoryTraversal(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/a", "outputs:\n  - id: \"code\"\n    path: \"out.go\"\n")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", "../../etc/passwd", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got: %v", err)
	}
}

// TC-11: Path validation — backslash.
func TestMCPWriteFile_BackslashPath_ReturnsPathContainsBackslash(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/a", "outputs:\n  - id: \"code\"\n    path: \"out.go\"\n")

	_, err := mcpwritefile.MCPWriteFile("ROOT/a", `output\file.go`, "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrPathContainsBackslash) {
		t.Errorf("expected ErrPathContainsBackslash, got: %v", err)
	}
}

// Ensure imports used for error sentinel checks are referenced.
var _ = frontmatter.ErrFileUnreadable
