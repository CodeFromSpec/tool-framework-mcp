// code-from-spec: ROOT/golang/tests/utils/spec_tree@Iztj7ZAkGv8ovVz3AdqrjektSu4

package spectree_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/listfiles"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/spectree"
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

// testMkFile creates the file at path (relative to cwd) with empty content,
// creating parent directories as needed.
func testMkFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(dirOf(path), 0o755); err != nil {
		t.Fatalf("testMkFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatalf("testMkFile write: %v", err)
	}
}

// dirOf returns the directory component of a slash-separated path.
func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return "."
}

// ---------------------------------------------------------------------------
// Happy path
// ---------------------------------------------------------------------------

func TestSpecTreeScan_RootNodeOnly(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkFile(t, "code-from-spec/_node.md")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].LogicalName != "ROOT" {
		t.Errorf("logical name: got %q, want %q", nodes[0].LogicalName, "ROOT")
	}
	if nodes[0].FilePath.Value != "code-from-spec/_node.md" {
		t.Errorf("file path: got %q, want %q", nodes[0].FilePath.Value, "code-from-spec/_node.md")
	}
}

func TestSpecTreeScan_RootAndNestedNodes(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkFile(t, "code-from-spec/_node.md")
	testMkFile(t, "code-from-spec/a/_node.md")
	testMkFile(t, "code-from-spec/a/b/_node.md")

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
			t.Errorf("node[%d] logical name: got %q, want %q", i, nodes[i].LogicalName, w.logicalName)
		}
		if nodes[i].FilePath.Value != w.filePath {
			t.Errorf("node[%d] file path: got %q, want %q", i, nodes[i].FilePath.Value, w.filePath)
		}
	}
}

func TestSpecTreeScan_IgnoresNonNodeFiles(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkFile(t, "code-from-spec/_node.md")
	testMkFile(t, "code-from-spec/x/output.md")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].LogicalName != "ROOT" {
		t.Errorf("logical name: got %q, want %q", nodes[0].LogicalName, "ROOT")
	}
	if nodes[0].FilePath.Value != "code-from-spec/_node.md" {
		t.Errorf("file path: got %q, want %q", nodes[0].FilePath.Value, "code-from-spec/_node.md")
	}
}

func TestSpecTreeScan_IgnoresDirectoriesWithoutNodeMd(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkFile(t, "code-from-spec/_node.md")
	// Create an empty subdirectory (no files inside).
	if err := os.MkdirAll("code-from-spec/x/y", 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].LogicalName != "ROOT" {
		t.Errorf("logical name: got %q, want %q", nodes[0].LogicalName, "ROOT")
	}
}

func TestSpecTreeScan_ResultIsSortedByLogicalName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// Create files in non-alphabetical order on disk.
	testMkFile(t, "code-from-spec/z/_node.md")
	testMkFile(t, "code-from-spec/_node.md")
	testMkFile(t, "code-from-spec/a/b/_node.md")

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
			t.Errorf("node[%d] logical name: got %q, want %q", i, nodes[i].LogicalName, w.logicalName)
		}
		if nodes[i].FilePath.Value != w.filePath {
			t.Errorf("node[%d] file path: got %q, want %q", i, nodes[i].FilePath.Value, w.filePath)
		}
	}
}

// ---------------------------------------------------------------------------
// Failure cases
// ---------------------------------------------------------------------------

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

func TestSpecTreeScan_EmptyCodeFromSpecDirectory(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.MkdirAll("code-from-spec", 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, spectree.ErrNoNodesFound) {
		t.Errorf("expected ErrNoNodesFound, got: %v", err)
	}
}

func TestSpecTreeScan_OnlyNonNodeFiles(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkFile(t, "code-from-spec/README.md")
	testMkFile(t, "code-from-spec/x/output.md")

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, spectree.ErrNoNodesFound) {
		t.Errorf("expected ErrNoNodesFound, got: %v", err)
	}
}
