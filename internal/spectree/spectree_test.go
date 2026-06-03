// code-from-spec: ROOT/golang/tests/spec_tree/scan@DqZw5cidIDsrhn5q00PRMWLBgdM
package spectree_test

import (
	"errors"
	"os"
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

func testMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("testMkdirAll: %v", err)
	}
}

func testWriteFile(t *testing.T, path string) {
	t.Helper()
	dir := path[:len(path)-len("/"+lastSegment(path))]
	if dir != path {
		testMkdirAll(t, dir)
	}
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

func lastSegment(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

func TestSpecTreeScan_RootNodeOnly(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].LogicalName != "ROOT" {
		t.Errorf("expected logical_name=ROOT, got %q", nodes[0].LogicalName)
	}
	if nodes[0].FilePath.Value != "code-from-spec/_node.md" {
		t.Errorf("expected file_path=code-from-spec/_node.md, got %q", nodes[0].FilePath.Value)
	}
}

func TestSpecTreeScan_RootAndNestedNodes(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

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

	for i, exp := range expected {
		if nodes[i].LogicalName != exp.logicalName {
			t.Errorf("nodes[%d].LogicalName: expected %q, got %q", i, exp.logicalName, nodes[i].LogicalName)
		}
		if nodes[i].FilePath.Value != exp.filePath {
			t.Errorf("nodes[%d].FilePath.Value: expected %q, got %q", i, exp.filePath, nodes[i].FilePath.Value)
		}
	}
}

func TestSpecTreeScan_IgnoresNonNodeFiles(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

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
		t.Errorf("expected logical_name=ROOT, got %q", nodes[0].LogicalName)
	}
}

func TestSpecTreeScan_IgnoresDirectoriesWithoutNodeMd(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md")
	testMkdirAll(t, "code-from-spec/x/y")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].LogicalName != "ROOT" {
		t.Errorf("expected logical_name=ROOT, got %q", nodes[0].LogicalName)
	}
}

func TestSpecTreeScan_ResultSortedByLogicalName(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

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

	for i, exp := range expected {
		if nodes[i].LogicalName != exp.logicalName {
			t.Errorf("nodes[%d].LogicalName: expected %q, got %q", i, exp.logicalName, nodes[i].LogicalName)
		}
		if nodes[i].FilePath.Value != exp.filePath {
			t.Errorf("nodes[%d].FilePath.Value: expected %q, got %q", i, exp.filePath, nodes[i].FilePath.Value)
		}
	}
}

func TestSpecTreeScan_NoCodeFromSpecDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, listfiles.ErrDirectoryNotFound) {
		t.Errorf("expected ErrDirectoryNotFound, got %v", err)
	}
}

func TestSpecTreeScan_EmptyCodeFromSpecDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testMkdirAll(t, "code-from-spec")

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, spectree.ErrNoNodesFound) {
		t.Errorf("expected ErrNoNodesFound, got %v", err)
	}
}

func TestSpecTreeScan_OnlyNonNodeFiles(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

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
