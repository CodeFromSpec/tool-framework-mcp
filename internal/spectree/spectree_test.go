// code-from-spec: ROOT/golang/tests/spec_tree/scan@Abtqwiy21MchVnD789TJ_54SITY
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

func testMkdirAndWrite(t *testing.T, path string, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testMkdirAndWrite MkdirAll %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testMkdirAndWrite WriteFile %s: %v", path, err)
	}
}

func TestSpecTreeScan_RootNodeOnly(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkdirAndWrite(t, "code-from-spec/_node.md", "# ROOT\n")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].LogicalName != "ROOT" {
		t.Errorf("expected LogicalName ROOT, got %s", nodes[0].LogicalName)
	}
	if nodes[0].FilePath.Value != "code-from-spec/_node.md" {
		t.Errorf("expected FilePath code-from-spec/_node.md, got %s", nodes[0].FilePath.Value)
	}
}

func TestSpecTreeScan_RootAndNestedNodes(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkdirAndWrite(t, "code-from-spec/_node.md", "# ROOT\n")
	testMkdirAndWrite(t, "code-from-spec/a/_node.md", "# ROOT/a\n")
	testMkdirAndWrite(t, "code-from-spec/a/b/_node.md", "# ROOT/a/b\n")

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
		{"ROOT", "code-from-spec/_node.md"},
		{"ROOT/a", "code-from-spec/a/_node.md"},
		{"ROOT/a/b", "code-from-spec/a/b/_node.md"},
	}
	for i, exp := range expected {
		if nodes[i].LogicalName != exp.logicalName {
			t.Errorf("node[%d] LogicalName: expected %s, got %s", i, exp.logicalName, nodes[i].LogicalName)
		}
		if nodes[i].FilePath.Value != exp.filePath {
			t.Errorf("node[%d] FilePath: expected %s, got %s", i, exp.filePath, nodes[i].FilePath.Value)
		}
	}
}

func TestSpecTreeScan_IgnoresNonNodeFiles(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkdirAndWrite(t, "code-from-spec/_node.md", "# ROOT\n")
	testMkdirAndWrite(t, "code-from-spec/x/output.md", "content\n")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].LogicalName != "ROOT" {
		t.Errorf("expected LogicalName ROOT, got %s", nodes[0].LogicalName)
	}
}

func TestSpecTreeScan_IgnoresDirectoriesWithoutNodeMd(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkdirAndWrite(t, "code-from-spec/_node.md", "# ROOT\n")
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
		t.Errorf("expected LogicalName ROOT, got %s", nodes[0].LogicalName)
	}
}

func TestSpecTreeScan_SortedByLogicalName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkdirAndWrite(t, "code-from-spec/z/_node.md", "# ROOT/z\n")
	testMkdirAndWrite(t, "code-from-spec/_node.md", "# ROOT\n")
	testMkdirAndWrite(t, "code-from-spec/a/b/_node.md", "# ROOT/a/b\n")

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
		{"ROOT", "code-from-spec/_node.md"},
		{"ROOT/a/b", "code-from-spec/a/b/_node.md"},
		{"ROOT/z", "code-from-spec/z/_node.md"},
	}
	for i, exp := range expected {
		if nodes[i].LogicalName != exp.logicalName {
			t.Errorf("node[%d] LogicalName: expected %s, got %s", i, exp.logicalName, nodes[i].LogicalName)
		}
		if nodes[i].FilePath.Value != exp.filePath {
			t.Errorf("node[%d] FilePath: expected %s, got %s", i, exp.filePath, nodes[i].FilePath.Value)
		}
	}
}

func TestSpecTreeScan_NoCodeFromSpecDirectory(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, listfiles.ErrDirectoryNotFound) {
		t.Errorf("expected ErrDirectoryNotFound, got %v", err)
	}
}

func TestSpecTreeScan_EmptyCodeFromSpecDirectory(t *testing.T) {
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
		t.Errorf("expected ErrNoNodesFound, got %v", err)
	}
}

func TestSpecTreeScan_OnlyNonNodeFiles(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkdirAndWrite(t, "code-from-spec/README.md", "readme\n")
	testMkdirAndWrite(t, "code-from-spec/x/output.md", "output\n")

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, spectree.ErrNoNodesFound) {
		t.Errorf("expected ErrNoNodesFound, got %v", err)
	}
}
