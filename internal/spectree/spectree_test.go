// code-from-spec: ROOT/golang/tests/spec_tree/scan@wwQUONMzUfUOr25pEUu_FieYSPI
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

// testWriteFile creates a file (and any necessary parent directories) at
// path relative to the current working directory with empty content.
func testWriteFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
		t.Fatalf("testWriteFile write: %v", err)
	}
}

// testMkdir creates an empty directory (and any necessary parents) relative
// to the current working directory.
func testMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("testMkdir: %v", err)
	}
}

// TC-01: Root node only — a single _node.md at the top level.
func TestSpecTreeScan_RootNodeOnly(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

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

// TC-02: Root and nested nodes.
func TestSpecTreeScan_RootAndNestedNodes(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

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
	wants := []want{
		{"ROOT", "code-from-spec/_node.md"},
		{"ROOT/a", "code-from-spec/a/_node.md"},
		{"ROOT/a/b", "code-from-spec/a/b/_node.md"},
	}
	for i, w := range wants {
		if nodes[i].LogicalName != w.logicalName {
			t.Errorf("node[%d]: expected logical name %q, got %q", i, w.logicalName, nodes[i].LogicalName)
		}
		if nodes[i].FilePath.Value != w.filePath {
			t.Errorf("node[%d]: expected file path %q, got %q", i, w.filePath, nodes[i].FilePath.Value)
		}
	}
}

// TC-03: Non-_node.md files are ignored.
func TestSpecTreeScan_IgnoresNonNodeFiles(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

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

// TC-04: Directories without _node.md are ignored.
func TestSpecTreeScan_IgnoresDirectoriesWithoutNodeFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md")
	testMkdir(t, "code-from-spec/x/y")

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
}

// TC-05: Results are sorted alphabetically by logical name.
func TestSpecTreeScan_SortedByLogicalName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

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
	wants := []want{
		{"ROOT", "code-from-spec/_node.md"},
		{"ROOT/a/b", "code-from-spec/a/b/_node.md"},
		{"ROOT/z", "code-from-spec/z/_node.md"},
	}
	for i, w := range wants {
		if nodes[i].LogicalName != w.logicalName {
			t.Errorf("node[%d]: expected logical name %q, got %q", i, w.logicalName, nodes[i].LogicalName)
		}
		if nodes[i].FilePath.Value != w.filePath {
			t.Errorf("node[%d]: expected file path %q, got %q", i, w.filePath, nodes[i].FilePath.Value)
		}
	}
}

// TC-06: No code-from-spec directory — error propagated from ListFiles.
func TestSpecTreeScan_NoCodeFromSpecDirectory(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// Do not create code-from-spec/ at all.
	nodes, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatalf("expected an error, got nodes: %v", nodes)
	}
	if !errors.Is(err, listfiles.ErrDirectoryNotFound) {
		t.Errorf("expected errors.Is(err, listfiles.ErrDirectoryNotFound), got: %v", err)
	}
}

// TC-07: Empty code-from-spec directory — no nodes found error.
func TestSpecTreeScan_EmptyCodeFromSpecDirectory(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkdir(t, "code-from-spec")

	nodes, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatalf("expected an error, got nodes: %v", nodes)
	}
	if err.Error() != "no nodes found" {
		t.Errorf("expected error %q, got %q", "no nodes found", err.Error())
	}
}

// TC-08: Only non-_node.md files in code-from-spec — no nodes found error.
func TestSpecTreeScan_OnlyNonNodeFiles(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/README.md")
	testWriteFile(t, "code-from-spec/x/output.md")

	nodes, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatalf("expected an error, got nodes: %v", nodes)
	}
	if err.Error() != "no nodes found" {
		t.Errorf("expected error %q, got %q", "no nodes found", err.Error())
	}
}
