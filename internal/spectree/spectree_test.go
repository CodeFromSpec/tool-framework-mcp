// code-from-spec: ROOT/golang/tests/spec_tree/scan@gleLK9Ncw2T7lnJ7Gvs6mBEbvWU
package spectree_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/listfiles"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectree"
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

func testWriteFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("testWriteFile MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatalf("testWriteFile WriteFile: %v", err)
	}
}

// TC-1: Root node only
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
		t.Errorf("expected LogicalName %q, got %q", "ROOT", nodes[0].LogicalName)
	}
	if nodes[0].FilePath.Value != "code-from-spec/_node.md" {
		t.Errorf("expected FilePath %q, got %q", "code-from-spec/_node.md", nodes[0].FilePath.Value)
	}
}

// TC-2: Root and nested nodes
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
	expected := []want{
		{"ROOT", "code-from-spec/_node.md"},
		{"ROOT/a", "code-from-spec/a/_node.md"},
		{"ROOT/a/b", "code-from-spec/a/b/_node.md"},
	}
	for i, w := range expected {
		if nodes[i].LogicalName != w.logicalName {
			t.Errorf("node[%d]: expected LogicalName %q, got %q", i, w.logicalName, nodes[i].LogicalName)
		}
		if nodes[i].FilePath.Value != w.filePath {
			t.Errorf("node[%d]: expected FilePath %q, got %q", i, w.filePath, nodes[i].FilePath.Value)
		}
	}
}

// TC-3: Ignores non-node files
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
		t.Errorf("expected LogicalName %q, got %q", "ROOT", nodes[0].LogicalName)
	}
	if nodes[0].FilePath.Value != "code-from-spec/_node.md" {
		t.Errorf("expected FilePath %q, got %q", "code-from-spec/_node.md", nodes[0].FilePath.Value)
	}
}

// TC-4: Ignores directories without _node.md
func TestSpecTreeScan_IgnoresDirectoriesWithoutNodeMd(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md")
	// Create an empty subdirectory with no files inside.
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
		t.Errorf("expected LogicalName %q, got %q", "ROOT", nodes[0].LogicalName)
	}
	if nodes[0].FilePath.Value != "code-from-spec/_node.md" {
		t.Errorf("expected FilePath %q, got %q", "code-from-spec/_node.md", nodes[0].FilePath.Value)
	}
}

// TC-5: Result is sorted by logical name
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
	expected := []want{
		{"ROOT", "code-from-spec/_node.md"},
		{"ROOT/a/b", "code-from-spec/a/b/_node.md"},
		{"ROOT/z", "code-from-spec/z/_node.md"},
	}
	for i, w := range expected {
		if nodes[i].LogicalName != w.logicalName {
			t.Errorf("node[%d]: expected LogicalName %q, got %q", i, w.logicalName, nodes[i].LogicalName)
		}
		if nodes[i].FilePath.Value != w.filePath {
			t.Errorf("node[%d]: expected FilePath %q, got %q", i, w.filePath, nodes[i].FilePath.Value)
		}
	}
}

// TC-6: No code-from-spec directory
func TestSpecTreeScan_NoCodeFromSpecDirectory(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// Do not create code-from-spec/ at all.

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, listfiles.ErrDirectoryNotFound) {
		t.Errorf("expected ErrDirectoryNotFound in error chain, got: %v", err)
	}
}

// TC-7: Empty code-from-spec directory
func TestSpecTreeScan_EmptyCodeFromSpecDirectory(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.MkdirAll("code-from-spec", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, spectree.ErrNoNodesFound) {
		t.Errorf("expected ErrNoNodesFound in error chain, got: %v", err)
	}
}

// TC-8: Only non-node files in code-from-spec
func TestSpecTreeScan_OnlyNonNodeFiles(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/README.md")
	testWriteFile(t, "code-from-spec/x/output.md")

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, spectree.ErrNoNodesFound) {
		t.Errorf("expected ErrNoNodesFound in error chain, got: %v", err)
	}
}
