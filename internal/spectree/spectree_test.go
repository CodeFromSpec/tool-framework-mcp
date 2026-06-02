// code-from-spec: ROOT/golang/tests/spec_tree/scan@7UOLI9wruaZK4ij7Fb6_ZPPlgLY
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
		t.Fatalf("MkdirAll %q: %v", path, err)
	}
}

func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile %q: %v", path, err)
	}
}

func TestSpecTreeScan_RootNodeOnly(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkdirAll(t, "code-from-spec")
	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("len = %d, want 1", len(nodes))
	}
	if nodes[0].LogicalName != "ROOT" {
		t.Errorf("LogicalName = %q, want %q", nodes[0].LogicalName, "ROOT")
	}
	if nodes[0].FilePath.Value != "code-from-spec/_node.md" {
		t.Errorf("FilePath = %q, want %q", nodes[0].FilePath.Value, "code-from-spec/_node.md")
	}
}

func TestSpecTreeScan_RootAndNestedNodes(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkdirAll(t, "code-from-spec/a/b")
	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n")
	testWriteFile(t, "code-from-spec/a/b/_node.md", "# ROOT/a/b\n")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 3 {
		t.Fatalf("len = %d, want 3", len(nodes))
	}

	cases := []struct {
		logicalName string
		filePath    string
	}{
		{"ROOT", "code-from-spec/_node.md"},
		{"ROOT/a", "code-from-spec/a/_node.md"},
		{"ROOT/a/b", "code-from-spec/a/b/_node.md"},
	}
	for i, c := range cases {
		if nodes[i].LogicalName != c.logicalName {
			t.Errorf("nodes[%d].LogicalName = %q, want %q", i, nodes[i].LogicalName, c.logicalName)
		}
		if nodes[i].FilePath.Value != c.filePath {
			t.Errorf("nodes[%d].FilePath = %q, want %q", i, nodes[i].FilePath.Value, c.filePath)
		}
	}
}

func TestSpecTreeScan_IgnoresNonNodeFiles(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkdirAll(t, "code-from-spec/x")
	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/x/output.md", "some output\n")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("len = %d, want 1", len(nodes))
	}
	if nodes[0].LogicalName != "ROOT" {
		t.Errorf("LogicalName = %q, want %q", nodes[0].LogicalName, "ROOT")
	}
}

func TestSpecTreeScan_IgnoresDirectoriesWithoutNodeMd(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkdirAll(t, "code-from-spec/x/y")
	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("len = %d, want 1", len(nodes))
	}
	if nodes[0].LogicalName != "ROOT" {
		t.Errorf("LogicalName = %q, want %q", nodes[0].LogicalName, "ROOT")
	}
}

func TestSpecTreeScan_ResultIsSortedByLogicalName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkdirAll(t, "code-from-spec/z")
	testMkdirAll(t, "code-from-spec/a/b")
	testWriteFile(t, "code-from-spec/z/_node.md", "# ROOT/z\n")
	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/b/_node.md", "# ROOT/a/b\n")

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 3 {
		t.Fatalf("len = %d, want 3", len(nodes))
	}

	cases := []struct {
		logicalName string
		filePath    string
	}{
		{"ROOT", "code-from-spec/_node.md"},
		{"ROOT/a/b", "code-from-spec/a/b/_node.md"},
		{"ROOT/z", "code-from-spec/z/_node.md"},
	}
	for i, c := range cases {
		if nodes[i].LogicalName != c.logicalName {
			t.Errorf("nodes[%d].LogicalName = %q, want %q", i, nodes[i].LogicalName, c.logicalName)
		}
		if nodes[i].FilePath.Value != c.filePath {
			t.Errorf("nodes[%d].FilePath = %q, want %q", i, nodes[i].FilePath.Value, c.filePath)
		}
	}
}

func TestSpecTreeScan_NoCodeFromSpecDirectory(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, listfiles.ErrDirectoryNotFound) {
		t.Errorf("error = %v, want ErrDirectoryNotFound", err)
	}
}

func TestSpecTreeScan_EmptyCodeFromSpecDirectory(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkdirAll(t, "code-from-spec")

	_, err := spectree.SpecTreeScan()
	if !errors.Is(err, spectree.ErrNoNodesFound) {
		t.Errorf("error = %v, want ErrNoNodesFound", err)
	}
}

func TestSpecTreeScan_OnlyNonNodeFilesInCodeFromSpec(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkdirAll(t, "code-from-spec/x")
	testWriteFile(t, "code-from-spec/README.md", "readme\n")
	testWriteFile(t, "code-from-spec/x/output.md", "output\n")

	_, err := spectree.SpecTreeScan()
	if !errors.Is(err, spectree.ErrNoNodesFound) {
		t.Errorf("error = %v, want ErrNoNodesFound", err)
	}
}
