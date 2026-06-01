// code-from-spec: ROOT/golang/tests/spec_tree/scan@wvHDZLCWttkiStsCBx8pDoZA-Lw
package spectree_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/listfiles"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectree"
)

// testChdir changes the working directory to dir for the duration of the test.
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

// testMkFile creates a file (and any needed parent directories) at the given
// path relative to the current working directory.
func testMkFile(t *testing.T, relPath string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(relPath), 0755); err != nil {
		t.Fatalf("testMkFile: mkdir %s: %v", filepath.Dir(relPath), err)
	}
	if err := os.WriteFile(relPath, []byte(""), 0644); err != nil {
		t.Fatalf("testMkFile: write %s: %v", relPath, err)
	}
}

// testMkDir creates a directory (and any needed parents) at the given path
// relative to the current working directory.
func testMkDir(t *testing.T, relPath string) {
	t.Helper()
	if err := os.MkdirAll(relPath, 0755); err != nil {
		t.Fatalf("testMkDir: %v", err)
	}
}

// TC-01: Root node only.
func TestSpecTreeScan_RootNodeOnly(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testMkFile(t, filepath.Join("code-from-spec", "_node.md"))

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}

	if nodes[0].LogicalName != "ROOT" {
		t.Errorf("expected logical_name %q, got %q", "ROOT", nodes[0].LogicalName)
	}
	if nodes[0].FilePath.Value != "code-from-spec/_node.md" {
		t.Errorf("expected file_path %q, got %q", "code-from-spec/_node.md", nodes[0].FilePath.Value)
	}
}

// TC-02: Root and nested nodes.
func TestSpecTreeScan_RootAndNestedNodes(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testMkFile(t, filepath.Join("code-from-spec", "_node.md"))
	testMkFile(t, filepath.Join("code-from-spec", "a", "_node.md"))
	testMkFile(t, filepath.Join("code-from-spec", "a", "b", "_node.md"))

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}

	type want struct {
		logicalName string
		filePath    string
	}
	expected := []want{
		{"ROOT", "code-from-spec/_node.md"},
		{"ROOT/a", "code-from-spec/a/_node.md"},
		{"ROOT/a/b", "code-from-spec/a/b/_node.md"},
	}

	for i, w := range expected {
		if nodes[i].LogicalName != w.logicalName {
			t.Errorf("node[%d]: expected logical_name %q, got %q", i, w.logicalName, nodes[i].LogicalName)
		}
		if nodes[i].FilePath.Value != w.filePath {
			t.Errorf("node[%d]: expected file_path %q, got %q", i, w.filePath, nodes[i].FilePath.Value)
		}
	}
}

// TC-03: Ignores non-node files.
func TestSpecTreeScan_IgnoresNonNodeFiles(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testMkFile(t, filepath.Join("code-from-spec", "_node.md"))
	testMkFile(t, filepath.Join("code-from-spec", "x", "output.md"))

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}

	if nodes[0].LogicalName != "ROOT" {
		t.Errorf("expected logical_name %q, got %q", "ROOT", nodes[0].LogicalName)
	}
	if nodes[0].FilePath.Value != "code-from-spec/_node.md" {
		t.Errorf("expected file_path %q, got %q", "code-from-spec/_node.md", nodes[0].FilePath.Value)
	}
}

// TC-04: Ignores directories without _node.md.
func TestSpecTreeScan_IgnoresEmptySubdirectories(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testMkFile(t, filepath.Join("code-from-spec", "_node.md"))
	testMkDir(t, filepath.Join("code-from-spec", "x", "y"))

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}

	if nodes[0].LogicalName != "ROOT" {
		t.Errorf("expected logical_name %q, got %q", "ROOT", nodes[0].LogicalName)
	}
	if nodes[0].FilePath.Value != "code-from-spec/_node.md" {
		t.Errorf("expected file_path %q, got %q", "code-from-spec/_node.md", nodes[0].FilePath.Value)
	}
}

// TC-05: Result is sorted by logical name.
func TestSpecTreeScan_SortedByLogicalName(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testMkFile(t, filepath.Join("code-from-spec", "z", "_node.md"))
	testMkFile(t, filepath.Join("code-from-spec", "_node.md"))
	testMkFile(t, filepath.Join("code-from-spec", "a", "b", "_node.md"))

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}

	expectedNames := []string{"ROOT", "ROOT/a/b", "ROOT/z"}
	for i, name := range expectedNames {
		if nodes[i].LogicalName != name {
			t.Errorf("node[%d]: expected logical_name %q, got %q", i, name, nodes[i].LogicalName)
		}
	}
}

// TC-06: No code-from-spec directory.
func TestSpecTreeScan_NoCodeFromSpecDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Do not create code-from-spec/ directory.
	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if !errors.Is(err, listfiles.ErrDirectoryNotFound) {
		t.Errorf("expected error wrapping listfiles.ErrDirectoryNotFound, got: %v", err)
	}
}

// TC-07: Empty code-from-spec directory.
func TestSpecTreeScan_EmptyCodeFromSpecDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testMkDir(t, "code-from-spec")

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if !errors.Is(err, spectree.ErrNoNodesFound) {
		t.Errorf("expected spectree.ErrNoNodesFound, got: %v", err)
	}
}

// TC-08: Only non-node files in code-from-spec.
func TestSpecTreeScan_OnlyNonNodeFiles(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testMkFile(t, filepath.Join("code-from-spec", "README.md"))
	testMkFile(t, filepath.Join("code-from-spec", "x", "output.md"))

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if !errors.Is(err, spectree.ErrNoNodesFound) {
		t.Errorf("expected spectree.ErrNoNodesFound, got: %v", err)
	}
}
