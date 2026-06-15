// code-from-spec: SPEC/golang/tests/parsing/node_parsing@jrJ9NukLb3qsfowAoBzWUkqKvjw
package parsenode_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/parsenode"
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

func testWriteNodeFile(t *testing.T, cfsPath string, content string) {
	t.Helper()
	dir := filepath.Dir(cfsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testWriteNodeFile: mkdir %v: %v", dir, err)
	}
	if err := os.WriteFile(cfsPath, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNodeFile: write %v: %v", cfsPath, err)
	}
}

func TestNodeParse_MinimalNameSectionOnly(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/x/_node.md", "# SPEC/x\nA simple node.\n")

	node, err := parsenode.NodeParse("SPEC/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection == nil {
		t.Fatal("NameSection is nil")
	}
	if node.NameSection.Heading != "spec/x" {
		t.Errorf("Heading = %q, want %q", node.NameSection.Heading, "spec/x")
	}
	if node.NameSection.RawHeading != "# SPEC/x" {
		t.Errorf("RawHeading = %q, want %q", node.NameSection.RawHeading, "# SPEC/x")
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "A simple node." {
		t.Errorf("Content = %v, want [\"A simple node.\"]", node.NameSection.Content)
	}
	if len(node.NameSection.Subsections) != 0 {
		t.Errorf("Subsections = %v, want empty", node.NameSection.Subsections)
	}
	if node.Public != nil {
		t.Error("Public should be nil")
	}
	if node.Agent != nil {
		t.Error("Agent should be nil")
	}
	if node.Private != nil {
		t.Error("Private should be nil")
	}
}

func TestNodeParse_FullNodeAllSections(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `---
depends_on: []
---
# SPEC/payments/fees
Fees node description.
# Public
## Interface
Interface line.
## Constraints
Constraints line.
# Agent
Agent guidance line.
# Private
## Decisions
Decisions line.
## Rationale
Rationale line.
`
	testWriteNodeFile(t, "code-from-spec/payments/fees/_node.md", content)

	node, err := parsenode.NodeParse("SPEC/payments/fees")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "spec/payments/fees" {
		t.Errorf("NameSection.Heading = %q, want %q", node.NameSection.Heading, "spec/payments/fees")
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "Fees node description." {
		t.Errorf("NameSection.Content = %v", node.NameSection.Content)
	}

	if node.Public == nil {
		t.Fatal("Public is nil")
	}
	if len(node.Public.Content) != 0 {
		t.Errorf("Public.Content = %v, want empty", node.Public.Content)
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("Public.Subsections len = %d, want 2", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("Public.Subsections[0].Heading = %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if len(node.Public.Subsections[0].Content) != 1 || node.Public.Subsections[0].Content[0] != "Interface line." {
		t.Errorf("Public.Subsections[0].Content = %v", node.Public.Subsections[0].Content)
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("Public.Subsections[1].Heading = %q, want %q", node.Public.Subsections[1].Heading, "constraints")
	}
	if len(node.Public.Subsections[1].Content) != 1 || node.Public.Subsections[1].Content[0] != "Constraints line." {
		t.Errorf("Public.Subsections[1].Content = %v", node.Public.Subsections[1].Content)
	}

	if node.Agent == nil {
		t.Fatal("Agent is nil")
	}
	if len(node.Agent.Content) != 1 || node.Agent.Content[0] != "Agent guidance line." {
		t.Errorf("Agent.Content = %v", node.Agent.Content)
	}

	if node.Private == nil {
		t.Fatal("Private is nil")
	}
	if len(node.Private.Subsections) != 2 {
		t.Fatalf("Private.Subsections len = %d, want 2", len(node.Private.Subsections))
	}
	if node.Private.Subsections[0].Heading != "decisions" {
		t.Errorf("Private.Subsections[0].Heading = %q, want %q", node.Private.Subsections[0].Heading, "decisions")
	}
	if len(node.Private.Subsections[0].Content) != 1 || node.Private.Subsections[0].Content[0] != "Decisions line." {
		t.Errorf("Private.Subsections[0].Content = %v", node.Private.Subsections[0].Content)
	}
	if node.Private.Subsections[1].Heading != "rationale" {
		t.Errorf("Private.Subsections[1].Heading = %q, want %q", node.Private.Subsections[1].Heading, "rationale")
	}
	if len(node.Private.Subsections[1].Content) != 1 || node.Private.Subsections[1].Content[0] != "Rationale line." {
		t.Errorf("Private.Subsections[1].Content = %v", node.Private.Subsections[1].Content)
	}
}

func TestNodeParse_NoPublicSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/decisions\nDecision description.\n# Private\n## Rationale\nRationale content.\n"
	testWriteNodeFile(t, "code-from-spec/decisions/_node.md", content)

	node, err := parsenode.NodeParse("SPEC/decisions")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public != nil {
		t.Error("Public should be nil")
	}
	if node.Agent != nil {
		t.Error("Agent should be nil")
	}
	if node.Private == nil {
		t.Fatal("Private is nil")
	}
	if len(node.Private.Subsections) != 1 {
		t.Fatalf("Private.Subsections len = %d, want 1", len(node.Private.Subsections))
	}
	if node.Private.Subsections[0].Heading != "rationale" {
		t.Errorf("Private.Subsections[0].Heading = %q, want %q", node.Private.Subsections[0].Heading, "rationale")
	}
	if len(node.Private.Subsections[0].Content) != 1 || node.Private.Subsections[0].Content[0] != "Rationale content." {
		t.Errorf("Private.Subsections[0].Content = %v", node.Private.Subsections[0].Content)
	}
}

func TestNodeParse_PublicContentBeforeFirstSubsection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/a\nNode content.\n# Public\nPreamble line one.\nPreamble line two.\n## Interface\nInterface line.\n"
	testWriteNodeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsenode.NodeParse("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public is nil")
	}
	if len(node.Public.Content) != 2 {
		t.Fatalf("Public.Content len = %d, want 2", len(node.Public.Content))
	}
	if node.Public.Content[0] != "Preamble line one." {
		t.Errorf("Public.Content[0] = %q", node.Public.Content[0])
	}
	if node.Public.Content[1] != "Preamble line two." {
		t.Errorf("Public.Content[1] = %q", node.Public.Content[1])
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Public.Subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("Public.Subsections[0].Heading = %q", node.Public.Subsections[0].Heading)
	}
	if len(node.Public.Subsections[0].Content) != 1 || node.Public.Subsections[0].Content[0] != "Interface line." {
		t.Errorf("Public.Subsections[0].Content = %v", node.Public.Subsections[0].Content)
	}
}

func TestNodeParse_PublicNoContentOrSubsections(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/b\nNode content.\n# Public\n# Agent\nAgent line.\n"
	testWriteNodeFile(t, "code-from-spec/b/_node.md", content)

	node, err := parsenode.NodeParse("SPEC/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public is nil")
	}
	if len(node.Public.Content) != 0 {
		t.Errorf("Public.Content = %v, want empty", node.Public.Content)
	}
	if len(node.Public.Subsections) != 0 {
		t.Errorf("Public.Subsections = %v, want empty", node.Public.Subsections)
	}
}

func TestNodeParse_AgentSectionWithSubsections(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/c\nNode content.\n# Agent\nPreamble line.\n## Implementation guidance\nGuidance content.\n## Contracts\nContracts content.\n"
	testWriteNodeFile(t, "code-from-spec/c/_node.md", content)

	node, err := parsenode.NodeParse("SPEC/c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Agent == nil {
		t.Fatal("Agent is nil")
	}
	if len(node.Agent.Content) != 1 || node.Agent.Content[0] != "Preamble line." {
		t.Errorf("Agent.Content = %v", node.Agent.Content)
	}
	if node.Agent.RawHeading != "# Agent" {
		t.Errorf("Agent.RawHeading = %q, want %q", node.Agent.RawHeading, "# Agent")
	}
	if len(node.Agent.Subsections) != 2 {
		t.Fatalf("Agent.Subsections len = %d, want 2", len(node.Agent.Subsections))
	}
	if node.Agent.Subsections[0].Heading != "implementation guidance" {
		t.Errorf("Agent.Subsections[0].Heading = %q", node.Agent.Subsections[0].Heading)
	}
	if len(node.Agent.Subsections[0].Content) != 1 || node.Agent.Subsections[0].Content[0] != "Guidance content." {
		t.Errorf("Agent.Subsections[0].Content = %v", node.Agent.Subsections[0].Content)
	}
	if node.Agent.Subsections[1].Heading != "contracts" {
		t.Errorf("Agent.Subsections[1].Heading = %q", node.Agent.Subsections[1].Heading)
	}
	if len(node.Agent.Subsections[1].Content) != 1 || node.Agent.Subsections[1].Content[0] != "Contracts content." {
		t.Errorf("Agent.Subsections[1].Content = %v", node.Agent.Subsections[1].Content)
	}
}

func TestNodeParse_PrivateSectionWithSubsections(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/d\nNode content.\n# Private\n## TODO\nTodo content.\n## Decisions\nDecisions content.\n## Rationale\nRationale content.\n"
	testWriteNodeFile(t, "code-from-spec/d/_node.md", content)

	node, err := parsenode.NodeParse("SPEC/d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Private == nil {
		t.Fatal("Private is nil")
	}
	if len(node.Private.Subsections) != 3 {
		t.Fatalf("Private.Subsections len = %d, want 3", len(node.Private.Subsections))
	}
	if node.Private.Subsections[0].Heading != "todo" {
		t.Errorf("Private.Subsections[0].Heading = %q", node.Private.Subsections[0].Heading)
	}
	if node.Private.Subsections[1].Heading != "decisions" {
		t.Errorf("Private.Subsections[1].Heading = %q", node.Private.Subsections[1].Heading)
	}
	if node.Private.Subsections[2].Heading != "rationale" {
		t.Errorf("Private.Subsections[2].Heading = %q", node.Private.Subsections[2].Heading)
	}
}

func TestNodeParse_ContentIsRawMarkdown(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/f\nNode content.\n# Public\n## Interface\n### A level-3 heading\n**bold text**\n```\ncode here\n```\n"
	testWriteNodeFile(t, "code-from-spec/f/_node.md", content)

	node, err := parsenode.NodeParse("SPEC/f")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public is nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Public.Subsections len = %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	expected := []string{"### A level-3 heading", "**bold text**", "```", "code here", "```"}
	if len(sub.Content) != len(expected) {
		t.Fatalf("Content len = %d, want %d; content = %v", len(sub.Content), len(expected), sub.Content)
	}
	for i, want := range expected {
		if sub.Content[i] != want {
			t.Errorf("Content[%d] = %q, want %q", i, sub.Content[i], want)
		}
	}
}

func TestNodeParse_CaseInsensitivePublicDetection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/g\nNode content.\n# PUBLIC\n"
	testWriteNodeFile(t, "code-from-spec/g/_node.md", content)

	node, err := parsenode.NodeParse("SPEC/g")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public is nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("Public.Heading = %q, want %q", node.Public.Heading, "public")
	}
}

func TestNodeParse_PublicMixedCaseExtraWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/h\nNode content.\n#   PuBLiC\n"
	testWriteNodeFile(t, "code-from-spec/h/_node.md", content)

	node, err := parsenode.NodeParse("SPEC/h")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public is nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("Public.Heading = %q, want %q", node.Public.Heading, "public")
	}
}

func TestNodeParse_NodeNameWithVariedWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "#   SPEC/e\nNode content.\n"
	testWriteNodeFile(t, "code-from-spec/e/_node.md", content)

	node, err := parsenode.NodeParse("SPEC/e")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "spec/e" {
		t.Errorf("NameSection.Heading = %q, want %q", node.NameSection.Heading, "spec/e")
	}
}

func TestNodeParse_NodeNameROOTPrefixDoesNotMatchSPEC(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/x\nNode content.\n"
	testWriteNodeFile(t, "code-from-spec/x/_node.md", content)

	_, err := parsenode.NodeParse("SPEC/x")
	if !errors.Is(err, parsenode.ErrNodeNameDoesNotMatch) {
		t.Errorf("error = %v, want ErrNodeNameDoesNotMatch", err)
	}
}

func TestNodeParse_SubsectionHeadingsNormalized(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/i\nNode content.\n# Public\n##   Interface\nInterface content.\n## CONSTRAINTS\nConstraints content.\n"
	testWriteNodeFile(t, "code-from-spec/i/_node.md", content)

	node, err := parsenode.NodeParse("SPEC/i")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public is nil")
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("Public.Subsections len = %d, want 2", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("Subsections[0].Heading = %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("Subsections[1].Heading = %q, want %q", node.Public.Subsections[1].Heading, "constraints")
	}
}

func TestNodeParse_ClosingHashesStripped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/j\nNode content.\n# Public\n## Interface ##\nInterface content.\n"
	testWriteNodeFile(t, "code-from-spec/j/_node.md", content)

	node, err := parsenode.NodeParse("SPEC/j")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public is nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Public.Subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("Subsections[0].Heading = %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if node.Public.Subsections[0].RawHeading != "## Interface ##" {
		t.Errorf("Subsections[0].RawHeading = %q, want %q", node.Public.Subsections[0].RawHeading, "## Interface ##")
	}
}

func TestNodeParse_RawHeadingPreservesOriginalLine(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/k\nNode content.\n# Public\n## Interface\nInterface content.\n"
	testWriteNodeFile(t, "code-from-spec/k/_node.md", content)

	node, err := parsenode.NodeParse("SPEC/k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public is nil")
	}
	if node.Public.RawHeading != "# Public" {
		t.Errorf("Public.RawHeading = %q, want %q", node.Public.RawHeading, "# Public")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Public.Subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].RawHeading != "## Interface" {
		t.Errorf("Subsections[0].RawHeading = %q, want %q", node.Public.Subsections[0].RawHeading, "## Interface")
	}
}

func TestNodeParse_RawHeadingPreservesCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/l\nNode content.\n# PUBLIC\n"
	testWriteNodeFile(t, "code-from-spec/l/_node.md", content)

	node, err := parsenode.NodeParse("SPEC/l")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public is nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("Public.Heading = %q, want %q", node.Public.Heading, "public")
	}
	if node.Public.RawHeading != "# PUBLIC" {
		t.Errorf("Public.RawHeading = %q, want %q", node.Public.RawHeading, "# PUBLIC")
	}
}

func TestNodeParse_RawHeadingPreservesClosingHashes(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/m\nNode content.\n# Public\n## Foo ##\nFoo content.\n"
	testWriteNodeFile(t, "code-from-spec/m/_node.md", content)

	node, err := parsenode.NodeParse("SPEC/m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public is nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Public.Subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "foo" {
		t.Errorf("Subsections[0].Heading = %q, want %q", node.Public.Subsections[0].Heading, "foo")
	}
	if node.Public.Subsections[0].RawHeading != "## Foo ##" {
		t.Errorf("Subsections[0].RawHeading = %q, want %q", node.Public.Subsections[0].RawHeading, "## Foo ##")
	}
}

func TestNodeParse_RawHeadingPreservesExtraWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/n\nNode content.\n#   Public\n"
	testWriteNodeFile(t, "code-from-spec/n/_node.md", content)

	node, err := parsenode.NodeParse("SPEC/n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public is nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("Public.Heading = %q, want %q", node.Public.Heading, "public")
	}
	if node.Public.RawHeading != "#   Public" {
		t.Errorf("Public.RawHeading = %q, want %q", node.Public.RawHeading, "#   Public")
	}
}

func TestNodeParse_Level3AndDeeperAreContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/o\nNode content.\n# Public\n## Interface\n### A subsub heading\n#### Even deeper\nInterface content.\n"
	testWriteNodeFile(t, "code-from-spec/o/_node.md", content)

	node, err := parsenode.NodeParse("SPEC/o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public is nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Public.Subsections len = %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	found3 := false
	found4 := false
	for _, line := range sub.Content {
		if line == "### A subsub heading" {
			found3 = true
		}
		if line == "#### Even deeper" {
			found4 = true
		}
	}
	if !found3 {
		t.Errorf("Content missing '### A subsub heading'; content = %v", sub.Content)
	}
	if !found4 {
		t.Errorf("Content missing '#### Even deeper'; content = %v", sub.Content)
	}
}

func TestNodeParse_FencedCodeBlockBacktickHeadingLike(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/p\nNode content.\n# Public\n## Interface\n```\n# looks like heading\n## also heading-like\n```\nNormal content.\n"
	testWriteNodeFile(t, "code-from-spec/p/_node.md", content)

	node, err := parsenode.NodeParse("SPEC/p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public is nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Public.Subsections len = %d, want 1; subsections = %v", len(node.Public.Subsections), node.Public.Subsections)
	}
	sub := node.Public.Subsections[0]
	foundH1 := false
	foundH2 := false
	for _, line := range sub.Content {
		if line == "# looks like heading" {
			foundH1 = true
		}
		if line == "## also heading-like" {
			foundH2 = true
		}
	}
	if !foundH1 {
		t.Errorf("Content missing '# looks like heading'; content = %v", sub.Content)
	}
	if !foundH2 {
		t.Errorf("Content missing '## also heading-like'; content = %v", sub.Content)
	}
}

func TestNodeParse_FencedCodeBlockTildeFence(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/ptilde\nNode content.\n# Public\n## Interface\n~~~\n# looks like heading\n~~~\n"
	testWriteNodeFile(t, "code-from-spec/ptilde/_node.md", content)

	node, err := parsenode.NodeParse("SPEC/ptilde")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public is nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Public.Subsections len = %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	found := false
	for _, line := range sub.Content {
		if line == "# looks like heading" {
			found = true
		}
	}
	if !found {
		t.Errorf("Content missing '# looks like heading' inside tilde fence; content = %v", sub.Content)
	}
}

func TestNodeParse_FencedCodeBlockWithLanguageTag(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/plang\nNode content.\n# Public\n## Interface\n```python\n# python comment that looks like heading\n```\n"
	testWriteNodeFile(t, "code-from-spec/plang/_node.md", content)

	node, err := parsenode.NodeParse("SPEC/plang")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public is nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Public.Subsections len = %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	found := false
	for _, line := range sub.Content {
		if line == "# python comment that looks like heading" {
			found = true
		}
	}
	if !found {
		t.Errorf("Content missing python comment; content = %v", sub.Content)
	}
}

func TestNodeParse_BlankLinesBetweenHeadingAndContentPreserved(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/q\nNode content.\n# Public\n\nPublic content.\n"
	testWriteNodeFile(t, "code-from-spec/q/_node.md", content)

	node, err := parsenode.NodeParse("SPEC/q")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public is nil")
	}
	if len(node.Public.Content) != 2 {
		t.Fatalf("Public.Content len = %d, want 2; content = %v", len(node.Public.Content), node.Public.Content)
	}
	if node.Public.Content[0] != "" {
		t.Errorf("Public.Content[0] = %q, want empty string", node.Public.Content[0])
	}
	if node.Public.Content[1] != "Public content." {
		t.Errorf("Public.Content[1] = %q, want %q", node.Public.Content[1], "Public content.")
	}
}

func TestNodeParse_FrontmatterIsSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "---\ndepends_on: []\n---\n# SPEC/r\nBody content.\n"
	testWriteNodeFile(t, "code-from-spec/r/_node.md", content)

	node, err := parsenode.NodeParse("SPEC/r")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "spec/r" {
		t.Errorf("NameSection.Heading = %q", node.NameSection.Heading)
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "Body content." {
		t.Errorf("NameSection.Content = %v", node.NameSection.Content)
	}
}

func TestNodeParse_NoFrontmatterDelimiters(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/s\nBody content.\n"
	testWriteNodeFile(t, "code-from-spec/s/_node.md", content)

	node, err := parsenode.NodeParse("SPEC/s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "Body content." {
		t.Errorf("NameSection.Content = %v", node.NameSection.Content)
	}
}

func TestNodeParse_UnclosedFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "---\ndepends_on: []\n"
	testWriteNodeFile(t, "code-from-spec/s2/_node.md", content)

	_, err := parsenode.NodeParse("SPEC/s2")
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

func TestNodeParse_ARTIFACTReferenceRejected(t *testing.T) {
	_, err := parsenode.NodeParse("ARTIFACT/x")
	if !errors.Is(err, parsenode.ErrNotASpecReference) {
		t.Errorf("error = %v, want ErrNotASpecReference", err)
	}
}

func TestNodeParse_EXTERNALReferenceRejected(t *testing.T) {
	_, err := parsenode.NodeParse("EXTERNAL/x")
	if !errors.Is(err, parsenode.ErrNotASpecReference) {
		t.Errorf("error = %v, want ErrNotASpecReference", err)
	}
}

func TestNodeParse_QualifierRejected(t *testing.T) {
	_, err := parsenode.NodeParse("SPEC/x(interface)")
	if !errors.Is(err, parsenode.ErrHasQualifier) {
		t.Errorf("error = %v, want ErrHasQualifier", err)
	}
}

func TestNodeParse_FileDoesNotExist(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := parsenode.NodeParse("SPEC/nonexistent/path")
	if !errors.Is(err, parsenode.ErrFileUnreadable) {
		t.Errorf("error = %v, want ErrFileUnreadable", err)
	}
}

func TestNodeParse_ContentBeforeFirstHeading(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "Some text before heading.\n# SPEC/t\nContent.\n"
	testWriteNodeFile(t, "code-from-spec/t/_node.md", content)

	_, err := parsenode.NodeParse("SPEC/t")
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

func TestNodeParse_Level2HeadingBeforeLevel1(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "## Interface\nContent.\n"
	testWriteNodeFile(t, "code-from-spec/u/_node.md", content)

	_, err := parsenode.NodeParse("SPEC/u")
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

func TestNodeParse_EmptyBody(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/v/_node.md", "")

	_, err := parsenode.NodeParse("SPEC/v")
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

func TestNodeParse_NodeNameDoesNotMatchLogicalName(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/wrong-name\nContent.\n"
	testWriteNodeFile(t, "code-from-spec/correct-name/_node.md", content)

	_, err := parsenode.NodeParse("SPEC/correct-name")
	if !errors.Is(err, parsenode.ErrNodeNameDoesNotMatch) {
		t.Errorf("error = %v, want ErrNodeNameDoesNotMatch", err)
	}
}

func TestNodeParse_NodeNameCaseMismatchIsNotError(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# spec/mynode\nContent.\n"
	testWriteNodeFile(t, "code-from-spec/MYNODE/_node.md", content)

	_, err := parsenode.NodeParse("SPEC/MYNODE")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNodeParse_DuplicatePublicSectionSameCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/w\nContent.\n# Public\nPublic content.\n# Public\nMore public content.\n"
	testWriteNodeFile(t, "code-from-spec/w/_node.md", content)

	_, err := parsenode.NodeParse("SPEC/w")
	if !errors.Is(err, parsenode.ErrDuplicatePublicSection) {
		t.Errorf("error = %v, want ErrDuplicatePublicSection", err)
	}
}

func TestNodeParse_DuplicatePublicSectionDifferentCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/ww\nContent.\n# Public\nPublic content.\n# PUBLIC\nMore public content.\n"
	testWriteNodeFile(t, "code-from-spec/ww/_node.md", content)

	_, err := parsenode.NodeParse("SPEC/ww")
	if !errors.Is(err, parsenode.ErrDuplicatePublicSection) {
		t.Errorf("error = %v, want ErrDuplicatePublicSection", err)
	}
}

func TestNodeParse_DuplicateAgentSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/dupagent\nContent.\n# Agent\nAgent content.\n# Agent\nMore agent content.\n"
	testWriteNodeFile(t, "code-from-spec/dupagent/_node.md", content)

	_, err := parsenode.NodeParse("SPEC/dupagent")
	if !errors.Is(err, parsenode.ErrDuplicateAgentSection) {
		t.Errorf("error = %v, want ErrDuplicateAgentSection", err)
	}
}

func TestNodeParse_DuplicatePrivateSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/duppriv\nContent.\n# Private\nPrivate content.\n# Private\nMore private content.\n"
	testWriteNodeFile(t, "code-from-spec/duppriv/_node.md", content)

	_, err := parsenode.NodeParse("SPEC/duppriv")
	if !errors.Is(err, parsenode.ErrDuplicatePrivateSection) {
		t.Errorf("error = %v, want ErrDuplicatePrivateSection", err)
	}
}

func TestNodeParse_UnrecognizedSectionDecisions(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/y\nContent.\n# Decisions\nDecisions content.\n"
	testWriteNodeFile(t, "code-from-spec/y/_node.md", content)

	_, err := parsenode.NodeParse("SPEC/y")
	if !errors.Is(err, parsenode.ErrUnrecognizedSection) {
		t.Errorf("error = %v, want ErrUnrecognizedSection", err)
	}
}

func TestNodeParse_UnrecognizedSectionRationale(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/rationale\nContent.\n# Rationale\nRationale content.\n"
	testWriteNodeFile(t, "code-from-spec/rationale/_node.md", content)

	_, err := parsenode.NodeParse("SPEC/rationale")
	if !errors.Is(err, parsenode.ErrUnrecognizedSection) {
		t.Errorf("error = %v, want ErrUnrecognizedSection", err)
	}
}

func TestNodeParse_UnrecognizedSectionTODO(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/todo\nContent.\n# TODO\nTodo content.\n"
	testWriteNodeFile(t, "code-from-spec/todo/_node.md", content)

	_, err := parsenode.NodeParse("SPEC/todo")
	if !errors.Is(err, parsenode.ErrUnrecognizedSection) {
		t.Errorf("error = %v, want ErrUnrecognizedSection", err)
	}
}

func TestNodeParse_DuplicateSubsectionPublicSameCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/z\nContent.\n# Public\n## Interface\nContent A.\n## Interface\nContent B.\n"
	testWriteNodeFile(t, "code-from-spec/z/_node.md", content)

	_, err := parsenode.NodeParse("SPEC/z")
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}

func TestNodeParse_DuplicateSubsectionPublicDifferentCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/z2\nContent.\n# Public\n## Interface\nContent A.\n## INTERFACE\nContent B.\n"
	testWriteNodeFile(t, "code-from-spec/z2/_node.md", content)

	_, err := parsenode.NodeParse("SPEC/z2")
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}

func TestNodeParse_DuplicateSubsectionPublicWhitespaceVariation(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/z3\nContent.\n# Public\n## Interface\nContent A.\n##   Interface\nContent B.\n"
	testWriteNodeFile(t, "code-from-spec/z3/_node.md", content)

	_, err := parsenode.NodeParse("SPEC/z3")
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}

func TestNodeParse_DuplicateSubsectionAgent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/z4\nContent.\n# Agent\n## Guidance\nGuidance content.\n## Guidance\nDuplicate content.\n"
	testWriteNodeFile(t, "code-from-spec/z4/_node.md", content)

	_, err := parsenode.NodeParse("SPEC/z4")
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}
