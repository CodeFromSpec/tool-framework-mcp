// code-from-spec: ROOT/golang/tests/spec_tree/scan@5L1ByY6r7Ba8uW2injqmi51q5bA

package spectree_test

import (
	"errors"
	"os"
	"path/filepath"
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

// testMakeFile creates a file at the given relative path (under the current
// working directory), creating any necessary parent directories.
func testMakeFile(t *testing.T, relPath string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(relPath), 0755); err != nil {
		t.Fatalf("testMakeFile mkdir: %v", err)
	}
	if err := os.WriteFile(relPath, []byte(""), 0644); err != nil {
		t.Fatalf("testMakeFile write: %v", err)
	}
}

// testMakeDir creates an empty directory at the given relative path.
func testMakeDir(t *testing.T, relPath string) {
	t.Helper()
	if err := os.MkdirAll(relPath, 0755); err != nil {
		t.Fatalf("testMakeDir: %v", err)
	}
}

// Test 1: Root node only — a single _node.md at the top of code-from-spec/.
func TestSpecTreeScan_RootOnly(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMakeFile(t, "code-from-spec/_node.md")

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

// Test 2: Root and nested nodes.
func TestSpecTreeScan_RootAndNestedNodes(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMakeFile(t, "code-from-spec/_node.md")
	testMakeFile(t, "code-from-spec/a/_node.md")
	testMakeFile(t, "code-from-spec/a/b/_node.md")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}

	type expected struct {
		logicalName string
		filePath    string
	}
	want := []expected{
		{"ROOT", "code-from-spec/_node.md"},
		{"ROOT/a", "code-from-spec/a/_node.md"},
		{"ROOT/a/b", "code-from-spec/a/b/_node.md"},
	}

	for i, w := range want {
		if nodes[i].LogicalName != w.logicalName {
			t.Errorf("node[%d]: expected LogicalName %q, got %q", i, w.logicalName, nodes[i].LogicalName)
		}
		if nodes[i].FilePath.Value != w.filePath {
			t.Errorf("node[%d]: expected FilePath %q, got %q", i, w.filePath, nodes[i].FilePath.Value)
		}
	}
}

// Test 3: Ignores non-node files.
func TestSpecTreeScan_IgnoresNonNodeFiles(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMakeFile(t, "code-from-spec/_node.md")
	testMakeFile(t, "code-from-spec/x/output.md")

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

// Test 4: Ignores directories without _node.md.
func TestSpecTreeScan_IgnoresEmptySubdirectories(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMakeFile(t, "code-from-spec/_node.md")
	testMakeDir(t, "code-from-spec/x/y")

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

// Test 5: Result is sorted by logical name.
func TestSpecTreeScan_SortedByLogicalName(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMakeFile(t, "code-from-spec/z/_node.md")
	testMakeFile(t, "code-from-spec/_node.md")
	testMakeFile(t, "code-from-spec/a/b/_node.md")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}

	type expected struct {
		logicalName string
		filePath    string
	}
	want := []expected{
		{"ROOT", "code-from-spec/_node.md"},
		{"ROOT/a/b", "code-from-spec/a/b/_node.md"},
		{"ROOT/z", "code-from-spec/z/_node.md"},
	}

	for i, w := range want {
		if nodes[i].LogicalName != w.logicalName {
			t.Errorf("node[%d]: expected LogicalName %q, got %q", i, w.logicalName, nodes[i].LogicalName)
		}
		if nodes[i].FilePath.Value != w.filePath {
			t.Errorf("node[%d]: expected FilePath %q, got %q", i, w.filePath, nodes[i].FilePath.Value)
		}
	}
}

// Test 6: No code-from-spec directory — error propagated from ListFiles.
func TestSpecTreeScan_NoCodeFromSpecDir(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	// Do not create code-from-spec/ at all.

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, listfiles.ErrDirectoryNotFound) {
		t.Errorf("expected ErrDirectoryNotFound, got: %v", err)
	}
}

// Test 7: Empty code-from-spec directory — ErrNoNodesFound.
func TestSpecTreeScan_EmptyCodeFromSpecDir(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMakeDir(t, "code-from-spec")

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, spectree.ErrNoNodesFound) {
		t.Errorf("expected ErrNoNodesFound, got: %v", err)
	}
}

// Test 8: Only non-node files — ErrNoNodesFound.
func TestSpecTreeScan_OnlyNonNodeFiles(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMakeFile(t, "code-from-spec/README.md")
	testMakeFile(t, "code-from-spec/x/output.md")

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, spectree.ErrNoNodesFound) {
		t.Errorf("expected ErrNoNodesFound, got: %v", err)
	}
}
