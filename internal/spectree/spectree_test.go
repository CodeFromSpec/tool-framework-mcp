// code-from-spec: SPEC/golang/tests/spec_tree/scan@WUtpVjqvMNISZe-hO3Kvy6UNYDg
package spectree_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/listfiles"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/spectree"
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

func testMkFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath(path), 0755); err != nil {
		t.Fatalf("testMkFile MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatalf("testMkFile WriteFile: %v", err)
	}
}

func filepath(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' {
			return p[:i]
		}
	}
	return "."
}

func TestSpecTreeScan_TC01_RootNodeOnly(t *testing.T) {
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
	if nodes[0].LogicalName != "SPEC" {
		t.Errorf("expected logical name SPEC, got %s", nodes[0].LogicalName)
	}
	if nodes[0].FilePath.Value != "code-from-spec/_node.md" {
		t.Errorf("expected file_path code-from-spec/_node.md, got %s", nodes[0].FilePath.Value)
	}
}

func TestSpecTreeScan_TC02_RootAndNestedNodes(t *testing.T) {
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

	expected := []struct {
		logicalName string
		filePath    string
	}{
		{"SPEC", "code-from-spec/_node.md"},
		{"SPEC/a", "code-from-spec/a/_node.md"},
		{"SPEC/a/b", "code-from-spec/a/b/_node.md"},
	}

	for i, e := range expected {
		if nodes[i].LogicalName != e.logicalName {
			t.Errorf("node[%d] logical name: expected %s, got %s", i, e.logicalName, nodes[i].LogicalName)
		}
		if nodes[i].FilePath.Value != e.filePath {
			t.Errorf("node[%d] file path: expected %s, got %s", i, e.filePath, nodes[i].FilePath.Value)
		}
	}
}

func TestSpecTreeScan_TC03_IgnoresNonNodeFiles(t *testing.T) {
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
	if nodes[0].LogicalName != "SPEC" {
		t.Errorf("expected logical name SPEC, got %s", nodes[0].LogicalName)
	}
}

func TestSpecTreeScan_TC04_IgnoresUnderscorePrefixedDirectoriesAtRoot(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkFile(t, "code-from-spec/_node.md")
	testMkFile(t, "code-from-spec/_rules/some/_node.md")
	testMkFile(t, "code-from-spec/_tools/_node.md")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].LogicalName != "SPEC" {
		t.Errorf("expected logical name SPEC, got %s", nodes[0].LogicalName)
	}
}

func TestSpecTreeScan_TC05_UnderscorePrefixedDirsDeeperInTreeNotIgnored(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkFile(t, "code-from-spec/_node.md")
	testMkFile(t, "code-from-spec/a/_node.md")
	testMkFile(t, "code-from-spec/a/_internal/_node.md")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}

	expected := []string{"SPEC", "SPEC/a", "SPEC/a/_internal"}
	for i, e := range expected {
		if nodes[i].LogicalName != e {
			t.Errorf("node[%d]: expected %s, got %s", i, e, nodes[i].LogicalName)
		}
	}
}

func TestSpecTreeScan_TC06_IgnoresDirectoriesWithoutNodeMd(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkFile(t, "code-from-spec/_node.md")
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
	if nodes[0].LogicalName != "SPEC" {
		t.Errorf("expected logical name SPEC, got %s", nodes[0].LogicalName)
	}
}

func TestSpecTreeScan_TC07_ResultSortedAlphabetically(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

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

	expected := []string{"SPEC", "SPEC/a/b", "SPEC/z"}
	for i, e := range expected {
		if nodes[i].LogicalName != e {
			t.Errorf("node[%d]: expected %s, got %s", i, e, nodes[i].LogicalName)
		}
	}
}

func TestSpecTreeScan_TC08_NoCodeFromSpecDirectory(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, listfiles.ErrDirectoryNotFound) {
		t.Errorf("expected ErrDirectoryNotFound, got: %v", err)
	}
}

func TestSpecTreeScan_TC09_EmptyCodeFromSpecDirectory(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.MkdirAll("code-from-spec", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, spectree.ErrNoNodesFound) {
		t.Errorf("expected ErrNoNodesFound, got: %v", err)
	}
}

func TestSpecTreeScan_TC10_OnlyNonNodeFiles(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkFile(t, "code-from-spec/README.md")
	testMkFile(t, "code-from-spec/x/output.md")

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, spectree.ErrNoNodesFound) {
		t.Errorf("expected ErrNoNodesFound, got: %v", err)
	}
}
