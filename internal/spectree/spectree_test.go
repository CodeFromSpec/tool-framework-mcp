// code-from-spec: ROOT/golang/tests/spec_tree/scan@emMvwqvQH1RZQJ_Nn-pWi_YzkTw
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

func testMkFile(t *testing.T, path string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testMkFile MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatalf("testMkFile WriteFile: %v", err)
	}
}

func testMkDir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("testMkDir: %v", err)
	}
}

func TestSpecTreeScan_TC01_RootNodeOnly(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMkFile(t, "code-from-spec/_node.md")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].LogicalName != "SPEC" {
		t.Errorf("expected LogicalName SPEC, got %s", nodes[0].LogicalName)
	}
	if nodes[0].FilePath.Value != "code-from-spec/_node.md" {
		t.Errorf("expected FilePath code-from-spec/_node.md, got %s", nodes[0].FilePath.Value)
	}
}

func TestSpecTreeScan_TC02_RootAndNestedNodes(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

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
			t.Errorf("node[%d]: expected LogicalName %s, got %s", i, e.logicalName, nodes[i].LogicalName)
		}
		if nodes[i].FilePath.Value != e.filePath {
			t.Errorf("node[%d]: expected FilePath %s, got %s", i, e.filePath, nodes[i].FilePath.Value)
		}
	}
}

func TestSpecTreeScan_TC03_IgnoresNonNodeFiles(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

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
		t.Errorf("expected LogicalName SPEC, got %s", nodes[0].LogicalName)
	}
	if nodes[0].FilePath.Value != "code-from-spec/_node.md" {
		t.Errorf("expected FilePath code-from-spec/_node.md, got %s", nodes[0].FilePath.Value)
	}
}

func TestSpecTreeScan_TC04_IgnoresUnderscorePrefixedDirectlyUnderRoot(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

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
		t.Errorf("expected LogicalName SPEC, got %s", nodes[0].LogicalName)
	}
}

func TestSpecTreeScan_TC05_UnderscorePrefixedDeeperNotIgnored(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

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

	expected := []struct {
		logicalName string
		filePath    string
	}{
		{"SPEC", "code-from-spec/_node.md"},
		{"SPEC/a", "code-from-spec/a/_node.md"},
		{"SPEC/a/_internal", "code-from-spec/a/_internal/_node.md"},
	}
	for i, e := range expected {
		if nodes[i].LogicalName != e.logicalName {
			t.Errorf("node[%d]: expected LogicalName %s, got %s", i, e.logicalName, nodes[i].LogicalName)
		}
		if nodes[i].FilePath.Value != e.filePath {
			t.Errorf("node[%d]: expected FilePath %s, got %s", i, e.filePath, nodes[i].FilePath.Value)
		}
	}
}

func TestSpecTreeScan_TC06_IgnoresDirectoriesWithoutNodeMd(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMkFile(t, "code-from-spec/_node.md")
	testMkDir(t, "code-from-spec/x/y")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].LogicalName != "SPEC" {
		t.Errorf("expected LogicalName SPEC, got %s", nodes[0].LogicalName)
	}
}

func TestSpecTreeScan_TC07_ResultSortedByLogicalName(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

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

	expected := []struct {
		logicalName string
		filePath    string
	}{
		{"SPEC", "code-from-spec/_node.md"},
		{"SPEC/a/b", "code-from-spec/a/b/_node.md"},
		{"SPEC/z", "code-from-spec/z/_node.md"},
	}
	for i, e := range expected {
		if nodes[i].LogicalName != e.logicalName {
			t.Errorf("node[%d]: expected LogicalName %s, got %s", i, e.logicalName, nodes[i].LogicalName)
		}
		if nodes[i].FilePath.Value != e.filePath {
			t.Errorf("node[%d]: expected FilePath %s, got %s", i, e.filePath, nodes[i].FilePath.Value)
		}
	}
}

func TestSpecTreeScan_TC08_NoCodeFromSpecDirectory(t *testing.T) {
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

func TestSpecTreeScan_TC09_EmptyCodeFromSpecDirectory(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMkDir(t, "code-from-spec")

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, spectree.ErrNoNodesFound) {
		t.Errorf("expected ErrNoNodesFound, got %v", err)
	}
}

func TestSpecTreeScan_TC10_OnlyNonNodeFiles(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMkFile(t, "code-from-spec/README.md")
	testMkFile(t, "code-from-spec/x/output.md")

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, spectree.ErrNoNodesFound) {
		t.Errorf("expected ErrNoNodesFound, got %v", err)
	}
}
