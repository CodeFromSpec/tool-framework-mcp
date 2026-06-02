// code-from-spec: ROOT/golang/tests/spec_tree/scan@c18XFpPc6Orxw0woMUMAxS1lRyE
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
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

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
		t.Errorf("expected logical name ROOT, got %s", nodes[0].LogicalName)
	}
	if nodes[0].FilePath.Value != "code-from-spec/_node.md" {
		t.Errorf("expected file path code-from-spec/_node.md, got %s", nodes[0].FilePath.Value)
	}
}

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

	expected := []struct {
		logicalName string
		filePath    string
	}{
		{"ROOT", "code-from-spec/_node.md"},
		{"ROOT/a", "code-from-spec/a/_node.md"},
		{"ROOT/a/b", "code-from-spec/a/b/_node.md"},
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
		t.Errorf("expected logical name ROOT, got %s", nodes[0].LogicalName)
	}
}

func TestSpecTreeScan_IgnoresEmptySubdirectories(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md")
	if err := os.MkdirAll("code-from-spec/x/y", 0755); err != nil {
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
		t.Errorf("expected logical name ROOT, got %s", nodes[0].LogicalName)
	}
}

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

	expected := []struct {
		logicalName string
		filePath    string
	}{
		{"ROOT", "code-from-spec/_node.md"},
		{"ROOT/a/b", "code-from-spec/a/b/_node.md"},
		{"ROOT/z", "code-from-spec/z/_node.md"},
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

func TestSpecTreeScan_NoCodeFromSpecDirectory(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, listfiles.ErrDirectoryNotFound) {
		t.Errorf("expected ErrDirectoryNotFound, got %v", err)
	}
}

func TestSpecTreeScan_EmptyCodeFromSpecDirectory(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	if err := os.MkdirAll("code-from-spec", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
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
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/README.md")
	testWriteFile(t, "code-from-spec/x/output.md")

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, spectree.ErrNoNodesFound) {
		t.Errorf("expected ErrNoNodesFound, got %v", err)
	}
}
