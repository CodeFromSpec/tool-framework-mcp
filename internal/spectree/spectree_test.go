// code-from-spec: ROOT/golang/tests/spec_tree/scan@mYNEWGUDLx2aX7ug1nMNNntQA44
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

// testWriteFile creates a file at the given relative path (creating parent
// directories as needed) with empty content.
func testWriteFile(t *testing.T, relPath string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(relPath), 0755); err != nil {
		t.Fatalf("testWriteFile MkdirAll: %v", err)
	}
	if err := os.WriteFile(relPath, []byte(""), 0644); err != nil {
		t.Fatalf("testWriteFile WriteFile: %v", err)
	}
}

// TestSpecTreeScan_RootNodeOnly covers TC-01: a single root _node.md.
func TestSpecTreeScan_RootNodeOnly(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].LogicalName != "ROOT" {
		t.Errorf("expected logical name %q, got %q", "ROOT", nodes[0].LogicalName)
	}
	if nodes[0].FilePath.Value != "code-from-spec/_node.md" {
		t.Errorf("expected file path %q, got %q", "code-from-spec/_node.md", nodes[0].FilePath.Value)
	}
}

// TestSpecTreeScan_RootAndNestedNodes covers TC-02: root and nested nodes.
func TestSpecTreeScan_RootAndNestedNodes(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md")
	testWriteFile(t, "code-from-spec/a/_node.md")
	testWriteFile(t, "code-from-spec/a/b/_node.md")

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
			t.Errorf("node[%d]: expected logical name %q, got %q", i, w.logicalName, nodes[i].LogicalName)
		}
		if nodes[i].FilePath.Value != w.filePath {
			t.Errorf("node[%d]: expected file path %q, got %q", i, w.filePath, nodes[i].FilePath.Value)
		}
	}
}

// TestSpecTreeScan_IgnoresNonNodeFiles covers TC-03: non-_node.md files are ignored.
func TestSpecTreeScan_IgnoresNonNodeFiles(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md")
	testWriteFile(t, "code-from-spec/x/output.md")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].LogicalName != "ROOT" {
		t.Errorf("expected logical name %q, got %q", "ROOT", nodes[0].LogicalName)
	}
	if nodes[0].FilePath.Value != "code-from-spec/_node.md" {
		t.Errorf("expected file path %q, got %q", "code-from-spec/_node.md", nodes[0].FilePath.Value)
	}
}

// TestSpecTreeScan_IgnoresDirectoriesWithoutNodeMd covers TC-04: directories
// without _node.md are not included.
func TestSpecTreeScan_IgnoresDirectoriesWithoutNodeMd(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md")
	// Create an empty subdirectory with no files.
	if err := os.MkdirAll("code-from-spec/x/y", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].LogicalName != "ROOT" {
		t.Errorf("expected logical name %q, got %q", "ROOT", nodes[0].LogicalName)
	}
	if nodes[0].FilePath.Value != "code-from-spec/_node.md" {
		t.Errorf("expected file path %q, got %q", "code-from-spec/_node.md", nodes[0].FilePath.Value)
	}
}

// TestSpecTreeScan_SortedByLogicalName covers TC-05: results are sorted
// alphabetically by logical name.
func TestSpecTreeScan_SortedByLogicalName(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/z/_node.md")
	testWriteFile(t, "code-from-spec/_node.md")
	testWriteFile(t, "code-from-spec/a/b/_node.md")

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
		{"ROOT/a/b", "code-from-spec/a/b/_node.md"},
		{"ROOT/z", "code-from-spec/z/_node.md"},
	}
	for i, w := range expected {
		if nodes[i].LogicalName != w.logicalName {
			t.Errorf("node[%d]: expected logical name %q, got %q", i, w.logicalName, nodes[i].LogicalName)
		}
		if nodes[i].FilePath.Value != w.filePath {
			t.Errorf("node[%d]: expected file path %q, got %q", i, w.filePath, nodes[i].FilePath.Value)
		}
	}
}

// TestSpecTreeScan_NoCodeFromSpecDirectory covers TC-06: no code-from-spec/
// directory exists.
func TestSpecTreeScan_NoCodeFromSpecDirectory(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	// Do not create code-from-spec/ at all.
	nodes, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatalf("expected an error, got %d nodes", len(nodes))
	}
	if !errors.Is(err, listfiles.ErrDirectoryNotFound) {
		t.Errorf("expected ErrDirectoryNotFound, got: %v", err)
	}
}

// TestSpecTreeScan_EmptyCodeFromSpecDirectory covers TC-07: code-from-spec/
// exists but contains no files.
func TestSpecTreeScan_EmptyCodeFromSpecDirectory(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	if err := os.MkdirAll("code-from-spec", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	nodes, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatalf("expected an error, got %d nodes", len(nodes))
	}
	if err.Error() != "no nodes found" {
		t.Errorf("expected error %q, got %q", "no nodes found", err.Error())
	}
}

// TestSpecTreeScan_OnlyNonNodeFiles covers TC-08: only non-_node.md files
// exist under code-from-spec/.
func TestSpecTreeScan_OnlyNonNodeFiles(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/README.md")
	testWriteFile(t, "code-from-spec/x/output.md")

	nodes, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatalf("expected an error, got %d nodes", len(nodes))
	}
	if err.Error() != "no nodes found" {
		t.Errorf("expected error %q, got %q", "no nodes found", err.Error())
	}
}
