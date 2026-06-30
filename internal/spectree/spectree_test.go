// code-from-spec: SPEC/golang/test/cases/spec_tree/scan@OPAxC8QsBopMu1PWuHNLFYYTNB8
package spectree_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/spectree"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"
)

func TestSpecTreeScan_SingleRootNode(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/a").Write()

	refs, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(refs) != 1 {
		t.Fatalf("expected 1 reference, got %d", len(refs))
	}
	if refs[0].LogicalName != "SPEC/a" {
		t.Errorf("expected LogicalName SPEC/a, got %s", refs[0].LogicalName)
	}
	if refs[0].Path != "code-from-spec/a/_node.md" {
		t.Errorf("expected Path code-from-spec/a/_node.md, got %s", refs[0].Path)
	}
}

func TestSpecTreeScan_MultipleRootNodes(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/a").Write()
	testutils.CreateSpecNode(t, "SPEC/b").Write()

	refs, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(refs) != 2 {
		t.Fatalf("expected 2 references, got %d", len(refs))
	}
	if refs[0].LogicalName != "SPEC/a" {
		t.Errorf("expected LogicalName SPEC/a, got %s", refs[0].LogicalName)
	}
	if refs[1].LogicalName != "SPEC/b" {
		t.Errorf("expected LogicalName SPEC/b, got %s", refs[1].LogicalName)
	}
}

func TestSpecTreeScan_RootAndNestedNodes(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/a").Write()
	testutils.CreateSpecNode(t, "SPEC/a/b").Write()

	refs, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(refs) != 2 {
		t.Fatalf("expected 2 references, got %d", len(refs))
	}
	if refs[0].LogicalName != "SPEC/a" {
		t.Errorf("expected first LogicalName SPEC/a, got %s", refs[0].LogicalName)
	}
	if refs[0].Path != "code-from-spec/a/_node.md" {
		t.Errorf("expected first Path code-from-spec/a/_node.md, got %s", refs[0].Path)
	}
	if refs[1].LogicalName != "SPEC/a/b" {
		t.Errorf("expected second LogicalName SPEC/a/b, got %s", refs[1].LogicalName)
	}
	if refs[1].Path != "code-from-spec/a/b/_node.md" {
		t.Errorf("expected second Path code-from-spec/a/b/_node.md, got %s", refs[1].Path)
	}
}

func TestSpecTreeScan_IgnoresNonNodeFiles(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/a").Write()

	if err := os.MkdirAll("code-from-spec/x", 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile("code-from-spec/x/output.md", []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	refs, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(refs) != 1 {
		t.Fatalf("expected 1 reference, got %d", len(refs))
	}
	if refs[0].LogicalName != "SPEC/a" {
		t.Errorf("expected LogicalName SPEC/a, got %s", refs[0].LogicalName)
	}
}

func TestSpecTreeScan_IgnoresDotPrefixedDirectories(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/a").Write()

	if err := os.MkdirAll("code-from-spec/.cache/some", 0755); err != nil {
		t.Fatalf("failed to create .cache directory: %v", err)
	}
	if err := os.WriteFile("code-from-spec/.cache/some/_node.md", []byte("# SPEC/.cache/some\n"), 0644); err != nil {
		t.Fatalf("failed to write .cache node: %v", err)
	}

	if err := os.MkdirAll("code-from-spec/.hidden", 0755); err != nil {
		t.Fatalf("failed to create .hidden directory: %v", err)
	}
	if err := os.WriteFile("code-from-spec/.hidden/_node.md", []byte("# SPEC/.hidden\n"), 0644); err != nil {
		t.Fatalf("failed to write .hidden node: %v", err)
	}

	refs, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(refs) != 1 {
		t.Fatalf("expected 1 reference, got %d", len(refs))
	}
	if refs[0].LogicalName != "SPEC/a" {
		t.Errorf("expected LogicalName SPEC/a, got %s", refs[0].LogicalName)
	}
}

func TestSpecTreeScan_IgnoresDotPrefixedDirsDeeper(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/a").Write()

	if err := os.MkdirAll("code-from-spec/a/.internal", 0755); err != nil {
		t.Fatalf("failed to create .internal directory: %v", err)
	}
	if err := os.WriteFile("code-from-spec/a/.internal/_node.md", []byte("# SPEC/a/.internal\n"), 0644); err != nil {
		t.Fatalf("failed to write .internal node: %v", err)
	}

	refs, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(refs) != 1 {
		t.Fatalf("expected 1 reference, got %d", len(refs))
	}
	if refs[0].LogicalName != "SPEC/a" {
		t.Errorf("expected LogicalName SPEC/a, got %s", refs[0].LogicalName)
	}
}

func TestSpecTreeScan_IgnoresRootNodeMd(t *testing.T) {
	testutils.Chdir(t)

	if err := os.MkdirAll("code-from-spec", 0755); err != nil {
		t.Fatalf("failed to create code-from-spec: %v", err)
	}
	if err := os.WriteFile("code-from-spec/_node.md", []byte("# SPEC\n"), 0644); err != nil {
		t.Fatalf("failed to write root _node.md: %v", err)
	}

	testutils.CreateSpecNode(t, "SPEC/a").Write()

	refs, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(refs) != 1 {
		t.Fatalf("expected 1 reference, got %d", len(refs))
	}
	if refs[0].LogicalName != "SPEC/a" {
		t.Errorf("expected LogicalName SPEC/a, got %s", refs[0].LogicalName)
	}
}

func TestSpecTreeScan_IgnoresDirectoriesWithoutNodeMd(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/a").Write()

	if err := os.MkdirAll("code-from-spec/x/y", 0755); err != nil {
		t.Fatalf("failed to create empty subdirectory: %v", err)
	}

	refs, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(refs) != 1 {
		t.Fatalf("expected 1 reference, got %d", len(refs))
	}
	if refs[0].LogicalName != "SPEC/a" {
		t.Errorf("expected LogicalName SPEC/a, got %s", refs[0].LogicalName)
	}
}

func TestSpecTreeScan_ResultIsSortedByLogicalName(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/z").Write()
	testutils.CreateSpecNode(t, "SPEC/a").Write()
	testutils.CreateSpecNode(t, "SPEC/a/b").Write()

	refs, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(refs) != 3 {
		t.Fatalf("expected 3 references, got %d", len(refs))
	}

	expected := []string{"SPEC/a", "SPEC/a/b", "SPEC/z"}
	for i, name := range expected {
		if refs[i].LogicalName != name {
			t.Errorf("index %d: expected %s, got %s", i, name, refs[i].LogicalName)
		}
	}
}

func TestSpecTreeScan_NoCodeFromSpecDirectory(t *testing.T) {
	testutils.Chdir(t)

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, oslayer.ErrDirectoryNotFound) {
		t.Errorf("expected ErrDirectoryNotFound, got %v", err)
	}
}

func TestSpecTreeScan_EmptyCodeFromSpecDirectory(t *testing.T) {
	testutils.Chdir(t)

	if err := os.MkdirAll("code-from-spec", 0755); err != nil {
		t.Fatalf("failed to create code-from-spec: %v", err)
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
	testutils.Chdir(t)

	if err := os.MkdirAll("code-from-spec/x", 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile("code-from-spec/README.md", []byte("readme"), 0644); err != nil {
		t.Fatalf("failed to write README: %v", err)
	}
	if err := os.WriteFile("code-from-spec/x/output.md", []byte("output"), 0644); err != nil {
		t.Fatalf("failed to write output.md: %v", err)
	}

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, spectree.ErrNoNodesFound) {
		t.Errorf("expected ErrNoNodesFound, got %v", err)
	}
}

func TestSpecTreeScan_OnlyRootNodeMdNoSubdirNodes(t *testing.T) {
	testutils.Chdir(t)

	if err := os.MkdirAll("code-from-spec", 0755); err != nil {
		t.Fatalf("failed to create code-from-spec: %v", err)
	}
	if err := os.WriteFile("code-from-spec/_node.md", []byte("# SPEC\n"), 0644); err != nil {
		t.Fatalf("failed to write root _node.md: %v", err)
	}

	_, err := spectree.SpecTreeScan()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, spectree.ErrNoNodesFound) {
		t.Errorf("expected ErrNoNodesFound, got %v", err)
	}
}
