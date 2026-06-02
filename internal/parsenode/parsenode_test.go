// code-from-spec: ROOT/golang/tests/parsing/node_parsing@5QY2hcqBIZPxsd3bTkLvlJYoboY
package parsenode_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
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

func TestNodeParse_MinimalNode(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("code-from-spec/x", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/x/_node.md", []byte("# ROOT/x\nA simple node.\n"), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/x" {
		t.Errorf("heading: got %q, want %q", node.NameSection.Heading, "root/x")
	}
	if node.NameSection.RawHeading != "# ROOT/x" {
		t.Errorf("raw_heading: got %q, want %q", node.NameSection.RawHeading, "# ROOT/x")
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "A simple node." {
		t.Errorf("content: got %v, want [A simple node.]", node.NameSection.Content)
	}
	if len(node.NameSection.Subsections) != 0 {
		t.Errorf("subsections: got %d, want 0", len(node.NameSection.Subsections))
	}
	if node.Public != nil {
		t.Error("public: expected nil")
	}
	if node.Agent != nil {
		t.Error("agent: expected nil")
	}
	if len(node.Private) != 0 {
		t.Errorf("private: got %d, want 0", len(node.Private))
	}
}

func TestNodeParse_FullNode(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "---\noutput: some/output.md\n---\n# ROOT/payments/fees\nDescription of this node.\n# Public\n## Interface\nInterface content line.\n## Constraints\nConstraints content line.\n# Agent\nAgent content line.\n# Decisions\nDecisions content line.\n# Rationale\nRationale content line.\n"
	if err := os.MkdirAll("code-from-spec/payments/fees", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/payments/fees/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/payments/fees")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/payments/fees" {
		t.Errorf("name heading: got %q", node.NameSection.Heading)
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "Description of this node." {
		t.Errorf("name content: got %v", node.NameSection.Content)
	}

	if node.Public == nil {
		t.Fatal("public: expected non-nil")
	}
	if len(node.Public.Content) != 0 {
		t.Errorf("public content: got %v, want empty", node.Public.Content)
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("public subsections: got %d, want 2", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("sub[0] heading: got %q", node.Public.Subsections[0].Heading)
	}
	if len(node.Public.Subsections[0].Content) != 1 || node.Public.Subsections[0].Content[0] != "Interface content line." {
		t.Errorf("sub[0] content: got %v", node.Public.Subsections[0].Content)
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("sub[1] heading: got %q", node.Public.Subsections[1].Heading)
	}
	if len(node.Public.Subsections[1].Content) != 1 || node.Public.Subsections[1].Content[0] != "Constraints content line." {
		t.Errorf("sub[1] content: got %v", node.Public.Subsections[1].Content)
	}

	if node.Agent == nil {
		t.Fatal("agent: expected non-nil")
	}
	if len(node.Agent.Content) != 1 || node.Agent.Content[0] != "Agent content line." {
		t.Errorf("agent content: got %v", node.Agent.Content)
	}

	if len(node.Private) != 2 {
		t.Fatalf("private: got %d, want 2", len(node.Private))
	}
	if node.Private[0].Heading != "decisions" {
		t.Errorf("private[0]: got %q", node.Private[0].Heading)
	}
	if node.Private[1].Heading != "rationale" {
		t.Errorf("private[1]: got %q", node.Private[1].Heading)
	}
}

func TestNodeParse_NoPublicSection(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/decisions\nDescription.\n# Rationale\nRationale content.\n"
	if err := os.MkdirAll("code-from-spec/decisions", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/decisions/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/decisions")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public != nil {
		t.Error("public: expected nil")
	}
	if node.Agent != nil {
		t.Error("agent: expected nil")
	}
	if len(node.Private) != 1 {
		t.Fatalf("private: got %d, want 1", len(node.Private))
	}
	if node.Private[0].Heading != "rationale" {
		t.Errorf("private[0]: got %q", node.Private[0].Heading)
	}
	if len(node.Private[0].Content) != 1 || node.Private[0].Content[0] != "Rationale content." {
		t.Errorf("private[0] content: got %v", node.Private[0].Content)
	}
}

func TestNodeParse_PublicContentBeforeFirstSubsection(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nNode description.\n# Public\nPreamble line one.\nPreamble line two.\n## Interface\nInterface content.\n"
	if err := os.MkdirAll("code-from-spec/a", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/a/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public: expected non-nil")
	}
	if len(node.Public.Content) != 2 {
		t.Fatalf("public content: got %v, want 2 lines", node.Public.Content)
	}
	if node.Public.Content[0] != "Preamble line one." {
		t.Errorf("public content[0]: got %q", node.Public.Content[0])
	}
	if node.Public.Content[1] != "Preamble line two." {
		t.Errorf("public content[1]: got %q", node.Public.Content[1])
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections: got %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("sub[0] heading: got %q", node.Public.Subsections[0].Heading)
	}
	if len(node.Public.Subsections[0].Content) != 1 || node.Public.Subsections[0].Content[0] != "Interface content." {
		t.Errorf("sub[0] content: got %v", node.Public.Subsections[0].Content)
	}
}

func TestNodeParse_PublicSectionNoContentOrSubsections(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/b\nNode description.\n# Public\n# Agent\nAgent content.\n"
	if err := os.MkdirAll("code-from-spec/b", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/b/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public: expected non-nil")
	}
	if len(node.Public.Content) != 0 {
		t.Errorf("public content: got %v, want empty", node.Public.Content)
	}
	if len(node.Public.Subsections) != 0 {
		t.Errorf("public subsections: got %d, want 0", len(node.Public.Subsections))
	}
	if node.Agent == nil {
		t.Fatal("agent: expected non-nil")
	}
	if len(node.Agent.Content) != 1 || node.Agent.Content[0] != "Agent content." {
		t.Errorf("agent content: got %v", node.Agent.Content)
	}
}

func TestNodeParse_AgentSectionWithSubsections(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/c\nNode description.\n# Agent\nAgent preamble line.\n## Implementation guidance\nImplementation content.\n## Contracts\nContracts content.\n"
	if err := os.MkdirAll("code-from-spec/c", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/c/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Agent == nil {
		t.Fatal("agent: expected non-nil")
	}
	if len(node.Agent.Content) != 1 || node.Agent.Content[0] != "Agent preamble line." {
		t.Errorf("agent content: got %v", node.Agent.Content)
	}
	if node.Agent.RawHeading != "# Agent" {
		t.Errorf("agent raw_heading: got %q", node.Agent.RawHeading)
	}
	if len(node.Agent.Subsections) != 2 {
		t.Fatalf("agent subsections: got %d, want 2", len(node.Agent.Subsections))
	}
	if node.Agent.Subsections[0].Heading != "implementation guidance" {
		t.Errorf("sub[0] heading: got %q", node.Agent.Subsections[0].Heading)
	}
	if len(node.Agent.Subsections[0].Content) != 1 || node.Agent.Subsections[0].Content[0] != "Implementation content." {
		t.Errorf("sub[0] content: got %v", node.Agent.Subsections[0].Content)
	}
	if node.Agent.Subsections[1].Heading != "contracts" {
		t.Errorf("sub[1] heading: got %q", node.Agent.Subsections[1].Heading)
	}
}

func TestNodeParse_PrivateSectionsPreserveOrder(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/d\nNode description.\n# TODO\nTodo content.\n# Decisions\nDecisions content.\n# Rationale\nRationale content.\n"
	if err := os.MkdirAll("code-from-spec/d", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/d/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(node.Private) != 3 {
		t.Fatalf("private: got %d, want 3", len(node.Private))
	}
	if node.Private[0].Heading != "todo" {
		t.Errorf("private[0]: got %q", node.Private[0].Heading)
	}
	if node.Private[1].Heading != "decisions" {
		t.Errorf("private[1]: got %q", node.Private[1].Heading)
	}
	if node.Private[2].Heading != "rationale" {
		t.Errorf("private[2]: got %q", node.Private[2].Heading)
	}
}

func TestNodeParse_ContentIsRawMarkdown(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/e\nNode description.\n# Public\n## Interface\n### Sub-heading\n**bold text**\n```go\nsome code\n```\n"
	if err := os.MkdirAll("code-from-spec/e", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/e/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/e")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public: expected non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections: got %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	expected := []string{"### Sub-heading", "**bold text**", "```go", "some code", "```"}
	if len(sub.Content) != len(expected) {
		t.Fatalf("interface content: got %v, want %v", sub.Content, expected)
	}
	for i, line := range expected {
		if sub.Content[i] != line {
			t.Errorf("content[%d]: got %q, want %q", i, sub.Content[i], line)
		}
	}
}

func TestNodeParse_CaseInsensitivePublicDetection(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/f\nDescription.\n# PUBLIC\nPublic content.\n"
	if err := os.MkdirAll("code-from-spec/f", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/f/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/f")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public: expected non-nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("public heading: got %q", node.Public.Heading)
	}
}

func TestNodeParse_PublicMixedCaseAndWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/g\nDescription.\n#   PuBLiC\nPublic content.\n"
	if err := os.MkdirAll("code-from-spec/g", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/g/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/g")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public: expected non-nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("public heading: got %q", node.Public.Heading)
	}
}

func TestNodeParse_NodeNameWithVariedWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "#   ROOT/e\nDescription.\n"
	if err := os.MkdirAll("code-from-spec/e", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/e/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/e")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/e" {
		t.Errorf("heading: got %q", node.NameSection.Heading)
	}
}

func TestNodeParse_SubsectionHeadingsNormalized(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/h\nDescription.\n# Public\n##   Interface\nInterface content.\n## CONSTRAINTS\nConstraints content.\n"
	if err := os.MkdirAll("code-from-spec/h", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/h/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/h")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public: expected non-nil")
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("subsections: got %d, want 2", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("sub[0]: got %q", node.Public.Subsections[0].Heading)
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("sub[1]: got %q", node.Public.Subsections[1].Heading)
	}
}

func TestNodeParse_ClosingHashesStripped(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/i\nDescription.\n# Public\n## Interface ##\nInterface content.\n"
	if err := os.MkdirAll("code-from-spec/i", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/i/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/i")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public: expected non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections: got %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("heading: got %q", node.Public.Subsections[0].Heading)
	}
	if node.Public.Subsections[0].RawHeading != "## Interface ##" {
		t.Errorf("raw_heading: got %q", node.Public.Subsections[0].RawHeading)
	}
}

func TestNodeParse_RawHeadingPreservesOriginalLine(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/j\nDescription.\n# Public\n## Interface\nInterface content.\n"
	if err := os.MkdirAll("code-from-spec/j", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/j/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/j")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public: expected non-nil")
	}
	if node.Public.RawHeading != "# Public" {
		t.Errorf("public raw_heading: got %q", node.Public.RawHeading)
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections: got %d", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].RawHeading != "## Interface" {
		t.Errorf("sub raw_heading: got %q", node.Public.Subsections[0].RawHeading)
	}
}

func TestNodeParse_RawHeadingPreservesCase(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/k\nDescription.\n# PUBLIC\nPublic content.\n"
	if err := os.MkdirAll("code-from-spec/k", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/k/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public: expected non-nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("heading: got %q", node.Public.Heading)
	}
	if node.Public.RawHeading != "# PUBLIC" {
		t.Errorf("raw_heading: got %q", node.Public.RawHeading)
	}
}

func TestNodeParse_RawHeadingPreservesClosingHashes(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/l\nDescription.\n# Public\n## Foo ##\nFoo content.\n"
	if err := os.MkdirAll("code-from-spec/l", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/l/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/l")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public: expected non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections: got %d", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "foo" {
		t.Errorf("heading: got %q", node.Public.Subsections[0].Heading)
	}
	if node.Public.Subsections[0].RawHeading != "## Foo ##" {
		t.Errorf("raw_heading: got %q", node.Public.Subsections[0].RawHeading)
	}
}

func TestNodeParse_RawHeadingPreservesExtraWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/m\nDescription.\n#   Public\nPublic content.\n"
	if err := os.MkdirAll("code-from-spec/m", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/m/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public: expected non-nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("heading: got %q", node.Public.Heading)
	}
	if node.Public.RawHeading != "#   Public" {
		t.Errorf("raw_heading: got %q", node.Public.RawHeading)
	}
}

func TestNodeParse_Level3AndDeeperAreContent(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/n\nDescription.\n# Public\n## Interface\n### Sub-heading\n#### Deep heading\nContent line.\n"
	if err := os.MkdirAll("code-from-spec/n", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/n/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public: expected non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections: got %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	expected := []string{"### Sub-heading", "#### Deep heading", "Content line."}
	if len(sub.Content) != len(expected) {
		t.Fatalf("content: got %v, want %v", sub.Content, expected)
	}
	for i, line := range expected {
		if sub.Content[i] != line {
			t.Errorf("content[%d]: got %q, want %q", i, sub.Content[i], line)
		}
	}
}

func TestNodeParse_FencedCodeBlockHeadingLike(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/o\nDescription.\n# Public\n## Interface\n```\n# looks like heading\n## also looks like heading\n```\n"
	if err := os.MkdirAll("code-from-spec/o", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/o/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public: expected non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections: got %d", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	expected := []string{"```", "# looks like heading", "## also looks like heading", "```"}
	if len(sub.Content) != len(expected) {
		t.Fatalf("content: got %v, want %v", sub.Content, expected)
	}
	for i, line := range expected {
		if sub.Content[i] != line {
			t.Errorf("content[%d]: got %q, want %q", i, sub.Content[i], line)
		}
	}
}

func TestNodeParse_FencedCodeBlockTilde(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/p\nDescription.\n# Public\n## Interface\n~~~\n# looks like level-1 heading\n~~~\n"
	if err := os.MkdirAll("code-from-spec/p", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/p/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public: expected non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections: got %d", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	expected := []string{"~~~", "# looks like level-1 heading", "~~~"}
	if len(sub.Content) != len(expected) {
		t.Fatalf("content: got %v, want %v", sub.Content, expected)
	}
	for i, line := range expected {
		if sub.Content[i] != line {
			t.Errorf("content[%d]: got %q, want %q", i, sub.Content[i], line)
		}
	}
}

func TestNodeParse_FencedCodeBlockWithLanguageTag(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/q\nDescription.\n# Public\n## Interface\n```go\n# looks like level-1 heading\n```\n"
	if err := os.MkdirAll("code-from-spec/q", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/q/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/q")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public: expected non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections: got %d", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	expected := []string{"```go", "# looks like level-1 heading", "```"}
	if len(sub.Content) != len(expected) {
		t.Fatalf("content: got %v, want %v", sub.Content, expected)
	}
	for i, line := range expected {
		if sub.Content[i] != line {
			t.Errorf("content[%d]: got %q, want %q", i, sub.Content[i], line)
		}
	}
}

func TestNodeParse_BlankLinesPreserved(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/r\nDescription.\n# Public\n\nContent line.\n"
	if err := os.MkdirAll("code-from-spec/r", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/r/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/r")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public: expected non-nil")
	}
	if len(node.Public.Content) != 2 {
		t.Fatalf("public content: got %v, want 2 lines", node.Public.Content)
	}
	if node.Public.Content[0] != "" {
		t.Errorf("content[0]: got %q, want empty string", node.Public.Content[0])
	}
	if node.Public.Content[1] != "Content line." {
		t.Errorf("content[1]: got %q", node.Public.Content[1])
	}
}

func TestNodeParse_FrontmatterSkipped(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "---\noutput: some/path.md\n---\n# ROOT/s\nDescription.\n"
	if err := os.MkdirAll("code-from-spec/s", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/s/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/s" {
		t.Errorf("heading: got %q", node.NameSection.Heading)
	}
}

func TestNodeParse_NoFrontmatter(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/t\nDescription.\n"
	if err := os.MkdirAll("code-from-spec/t", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/t/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	node, err := parsenode.NodeParse("ROOT/t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/t" {
		t.Errorf("heading: got %q", node.NameSection.Heading)
	}
}

func TestNodeParse_UnclosedFrontmatter(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "---\noutput: some/path.md\n# ROOT/u\nDescription.\n"
	if err := os.MkdirAll("code-from-spec/u", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/u/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := parsenode.NodeParse("ROOT/u")
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

func TestNodeParse_ArtifactReferenceRejected(t *testing.T) {
	_, err := parsenode.NodeParse("ARTIFACT/x")
	if !errors.Is(err, parsenode.ErrNotARootReference) {
		t.Errorf("expected ErrNotARootReference, got %v", err)
	}
}

func TestNodeParse_QualifierRejected(t *testing.T) {
	_, err := parsenode.NodeParse("ROOT/x(interface)")
	if !errors.Is(err, parsenode.ErrHasQualifier) {
		t.Errorf("expected ErrHasQualifier, got %v", err)
	}
}

func TestNodeParse_FileDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := parsenode.NodeParse("ROOT/nonexistent/file")
	if !errors.Is(err, parsenode.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestNodeParse_PropagatesPathErrors(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := parsenode.NodeParse("ROOT/../traversal")
	if err == nil {
		t.Error("expected error for directory traversal, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) && !errors.Is(err, parsenode.ErrFileUnreadable) {
		t.Errorf("expected path or file error, got %v", err)
	}
}

func TestNodeParse_ContentBeforeFirstHeading(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "This line appears before any heading.\n# ROOT/v\nDescription.\n"
	if err := os.MkdirAll("code-from-spec/v", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/v/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := parsenode.NodeParse("ROOT/v")
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

func TestNodeParse_Level2BeforeLevel1(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "## Some subsection\nDescription.\n"
	if err := os.MkdirAll("code-from-spec/w", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/w/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := parsenode.NodeParse("ROOT/w")
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

func TestNodeParse_EmptyBody(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("code-from-spec/empty", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/empty/_node.md", []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := parsenode.NodeParse("ROOT/empty")
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

func TestNodeParse_NodeNameDoesNotMatch(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/actual/name\nDescription.\n"
	if err := os.MkdirAll("code-from-spec/different/name", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/different/name/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := parsenode.NodeParse("ROOT/different/name")
	if !errors.Is(err, parsenode.ErrNodeNameDoesNotMatch) {
		t.Errorf("expected ErrNodeNameDoesNotMatch, got %v", err)
	}
}

func TestNodeParse_NodeNameCaseMismatchNotAnError(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# root/x\nDescription.\n"
	if err := os.MkdirAll("code-from-spec/X", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/X/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := parsenode.NodeParse("ROOT/X")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNodeParse_DuplicatePublicSectionSameCase(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/dup\nDescription.\n# Public\nFirst public content.\n# Public\nSecond public content.\n"
	if err := os.MkdirAll("code-from-spec/dup", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/dup/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := parsenode.NodeParse("ROOT/dup")
	if !errors.Is(err, parsenode.ErrDuplicatePublicSection) {
		t.Errorf("expected ErrDuplicatePublicSection, got %v", err)
	}
}

func TestNodeParse_DuplicatePublicSectionDifferentCase(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/dup2\nDescription.\n# Public\nFirst public content.\n# PUBLIC\nSecond public content.\n"
	if err := os.MkdirAll("code-from-spec/dup2", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/dup2/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := parsenode.NodeParse("ROOT/dup2")
	if !errors.Is(err, parsenode.ErrDuplicatePublicSection) {
		t.Errorf("expected ErrDuplicatePublicSection, got %v", err)
	}
}

func TestNodeParse_DuplicateAgentSection(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/dup3\nDescription.\n# Agent\nFirst agent content.\n# Agent\nSecond agent content.\n"
	if err := os.MkdirAll("code-from-spec/dup3", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/dup3/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := parsenode.NodeParse("ROOT/dup3")
	if !errors.Is(err, parsenode.ErrDuplicateAgentSection) {
		t.Errorf("expected ErrDuplicateAgentSection, got %v", err)
	}
}

func TestNodeParse_DuplicateSubsectionInPublicSameCase(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/dup4\nDescription.\n# Public\n## Interface\nContent.\n## Interface\nMore content.\n"
	if err := os.MkdirAll("code-from-spec/dup4", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/dup4/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := parsenode.NodeParse("ROOT/dup4")
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("expected ErrDuplicateSubsection, got %v", err)
	}
}

func TestNodeParse_DuplicateSubsectionInPublicDifferentCase(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/dup5\nDescription.\n# Public\n## Interface\nContent.\n## INTERFACE\nMore content.\n"
	if err := os.MkdirAll("code-from-spec/dup5", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/dup5/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := parsenode.NodeParse("ROOT/dup5")
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("expected ErrDuplicateSubsection, got %v", err)
	}
}

func TestNodeParse_DuplicateSubsectionInPublicWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/dup6\nDescription.\n# Public\n## Interface\nContent.\n##   Interface\nMore content.\n"
	if err := os.MkdirAll("code-from-spec/dup6", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/dup6/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := parsenode.NodeParse("ROOT/dup6")
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("expected ErrDuplicateSubsection, got %v", err)
	}
}

func TestNodeParse_DuplicateSubsectionInAgent(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/dup7\nDescription.\n# Agent\n## Rules\nContent.\n## Rules\nMore content.\n"
	if err := os.MkdirAll("code-from-spec/dup7", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("code-from-spec/dup7/_node.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := parsenode.NodeParse("ROOT/dup7")
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("expected ErrDuplicateSubsection, got %v", err)
	}
}
