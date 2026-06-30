// code-from-spec: SPEC/golang/tests/spec_tree/scan@zpIbpVkywKhyUrLj75L2Vr2xAas
package spectree_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/spectree"
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

func testMkNodeFile(t *testing.T, cfsPath string, logicalName string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(cfsPath), 0755); err != nil {
		t.Fatalf("testMkNodeFile MkdirAll: %v", err)
	}
	content := "# " + logicalName + "\n"
	if err := os.WriteFile(cfsPath, []byte(content), 0644); err != nil {
		t.Fatalf("testMkNodeFile WriteFile: %v", err)
	}
}

func testMkFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("testMkFile MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatalf("testMkFile WriteFile: %v", err)
	}
}

func TestSpecTreeScan_SingleRootNode(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkNodeFile(t, "code-from-spec/a/_node.md", "SPEC/a")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].LogicalName != "SPEC/a" {
		t.Errorf("expected logical name SPEC/a, got %s", nodes[0].LogicalName)
	}
	if nodes[0].Path != "code-from-spec/a/_node.md" {
		t.Errorf("expected path code-from-spec/a/_node.md, got %s", nodes[0].Path)
	}
}

func TestSpecTreeScan_MultipleRootNodes(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkNodeFile(t, "code-from-spec/a/_node.md", "SPEC/a")
	testMkNodeFile(t, "code-from-spec/b/_node.md", "SPEC/b")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}

	expected := []struct {
		logicalName string
		path        string
	}{
		{"SPEC/a", "code-from-spec/a/_node.md"},
		{"SPEC/b", "code-from-spec/b/_node.md"},
	}
	for i, e := range expected {
		if nodes[i].LogicalName != e.logicalName {
			t.Errorf("node[%d] logical name: expected %s, got %s", i, e.logicalName, nodes[i].LogicalName)
		}
		if nodes[i].Path != e.path {
			t.Errorf("node[%d] path: expected %s, got %s", i, e.path, nodes[i].Path)
		}
	}
}

func TestSpecTreeScan_RootAndNestedNodes(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkNodeFile(t, "code-from-spec/a/_node.md", "SPEC/a")
	testMkNodeFile(t, "code-from-spec/a/b/_node.md", "SPEC/a/b")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}

	expected := []struct {
		logicalName string
		path        string
	}{
		{"SPEC/a", "code-from-spec/a/_node.md"},
		{"SPEC/a/b", "code-from-spec/a/b/_node.md"},
	}
	for i, e := range expected {
		if nodes[i].LogicalName != e.logicalName {
			t.Errorf("node[%d] logical name: expected %s, got %s", i, e.logicalName, nodes[i].LogicalName)
		}
		if nodes[i].Path != e.path {
			t.Errorf("node[%d] path: expected %s, got %s", i, e.path, nodes[i].Path)
		}
	}
}

func TestSpecTreeScan_IgnoresNonNodeFiles(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkNodeFile(t, "code-from-spec/a/_node.md", "SPEC/a")
	testMkFile(t, "code-from-spec/x/output.md")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].LogicalName != "SPEC/a" {
		t.Errorf("expected logical name SPEC/a, got %s", nodes[0].LogicalName)
	}
}

func TestSpecTreeScan_IgnoresDotPrefixedDirectoriesUnderCodeFromSpec(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkNodeFile(t, "code-from-spec/a/_node.md", "SPEC/a")
	testMkNodeFile(t, "code-from-spec/.cache/some/_node.md", "SPEC/.cache/some")
	testMkNodeFile(t, "code-from-spec/.hidden/_node.md", "SPEC/.hidden")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].LogicalName != "SPEC/a" {
		t.Errorf("expected logical name SPEC/a, got %s", nodes[0].LogicalName)
	}
}

func TestSpecTreeScan_DotPrefixedDirsDeeperInTreeIgnored(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkNodeFile(t, "code-from-spec/a/_node.md", "SPEC/a")
	testMkNodeFile(t, "code-from-spec/a/.internal/_node.md", "SPEC/a/.internal")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].LogicalName != "SPEC/a" {
		t.Errorf("expected logical name SPEC/a, got %s", nodes[0].LogicalName)
	}
}

func TestSpecTreeScan_IgnoresNodeMdDirectlyInCodeFromSpec(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkFile(t, "code-from-spec/_node.md")
	testMkNodeFile(t, "code-from-spec/a/_node.md", "SPEC/a")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].LogicalName != "SPEC/a" {
		t.Errorf("expected logical name SPEC/a, got %s", nodes[0].LogicalName)
	}
}

func TestSpecTreeScan_IgnoresDirectoriesWithoutNodeMd(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkNodeFile(t, "code-from-spec/a/_node.md", "SPEC/a")
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
	if nodes[0].LogicalName != "SPEC/a" {
		t.Errorf("expected logical name SPEC/a, got %s", nodes[0].LogicalName)
	}
}

func TestSpecTreeScan_ResultSortedByLogicalName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkNodeFile(t, "code-from-spec/z/_node.md", "SPEC/z")
	testMkNodeFile(t, "code-from-spec/a/_node.md", "SPEC/a")
	testMkNodeFile(t, "code-from-spec/a/b/_node.md", "SPEC/a/b")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}

	expected := []string{"SPEC/a", "SPEC/a/b", "SPEC/z"}
	for i, e := range expected {
		if nodes[i].LogicalName != e {
			t.Errorf("node[%d]: expected %s, got %s", i, e, nodes[i].LogicalName)
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
	if !errors.Is(err, oslayer.ErrDirectoryNotFound) {
		t.Errorf("expected ErrDirectoryNotFound, got: %v", err)
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
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, spectree.ErrNoNodesFound) {
		t.Errorf("expected ErrNoNodesFound, got: %v", err)
	}
}

func TestSpecTreeScan_OnlyRootNodeMdNoSubdirectoryNodes(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkFile(t, "code-from-spec/_node.md")

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, spectree.ErrNoNodesFound) {
		t.Errorf("expected ErrNoNodesFound, got: %v", err)
	}
}
