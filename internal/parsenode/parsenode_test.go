// code-from-spec: ROOT/golang/tests/parsing/node_parsing@NkXxWQoGdiYsbz1m2KWL576SPtk
package parsenode_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
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

func testWriteNodeFile(t *testing.T, logicalName string, body string) {
	t.Helper()
	parts := []string{"code-from-spec"}
	segments := splitLogicalName(logicalName)
	parts = append(parts, segments...)
	parts = append(parts, "_node.md")
	path := filepath.Join(parts...)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("testWriteNodeFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("testWriteNodeFile write: %v", err)
	}
}

func splitLogicalName(logicalName string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(logicalName); i++ {
		if logicalName[i] == '/' {
			parts = append(parts, logicalName[start:i])
			start = i + 1
		}
	}
	parts = append(parts, logicalName[start:])
	return parts
}

func TestNodeParse_MinimalNameSectionOnly(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/x", "# ROOT/x\nA simple node.\n")

	node, err := parsenode.NodeParse("ROOT/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/x" {
		t.Errorf("heading = %q, want %q", node.NameSection.Heading, "root/x")
	}
	if node.NameSection.RawHeading != "# ROOT/x" {
		t.Errorf("raw_heading = %q, want %q", node.NameSection.RawHeading, "# ROOT/x")
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "A simple node." {
		t.Errorf("content = %v, want [\"A simple node.\"]", node.NameSection.Content)
	}
	if len(node.NameSection.Subsections) != 0 {
		t.Errorf("subsections = %v, want empty", node.NameSection.Subsections)
	}
	if node.Public != nil {
		t.Errorf("public = %v, want nil", node.Public)
	}
	if node.Agent != nil {
		t.Errorf("agent = %v, want nil", node.Agent)
	}
	if len(node.Private) != 0 {
		t.Errorf("private = %v, want empty", node.Private)
	}
}

func TestNodeParse_FullNodeAllSectionTypes(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/payments/fees\nDescription line.\n# Public\n## Interface\nInterface content line.\n## Constraints\nConstraints content line.\n# Agent\nAgent content line.\n# Decisions\nDecisions content line.\n# Rationale\nRationale content line.\n"
	testWriteNodeFile(t, "ROOT/payments/fees", body)

	node, err := parsenode.NodeParse("ROOT/payments/fees")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/payments/fees" {
		t.Errorf("name heading = %q, want %q", node.NameSection.Heading, "root/payments/fees")
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "Description line." {
		t.Errorf("name content = %v, want [\"Description line.\"]", node.NameSection.Content)
	}
	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Content) != 0 {
		t.Errorf("public.content = %v, want empty", node.Public.Content)
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("public.subsections len = %d, want 2", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("subsection[0].heading = %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if len(node.Public.Subsections[0].Content) != 1 || node.Public.Subsections[0].Content[0] != "Interface content line." {
		t.Errorf("subsection[0].content = %v", node.Public.Subsections[0].Content)
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("subsection[1].heading = %q, want %q", node.Public.Subsections[1].Heading, "constraints")
	}
	if len(node.Public.Subsections[1].Content) != 1 || node.Public.Subsections[1].Content[0] != "Constraints content line." {
		t.Errorf("subsection[1].content = %v", node.Public.Subsections[1].Content)
	}
	if node.Agent == nil {
		t.Fatal("agent = nil, want present")
	}
	if len(node.Agent.Content) != 1 || node.Agent.Content[0] != "Agent content line." {
		t.Errorf("agent.content = %v, want [\"Agent content line.\"]", node.Agent.Content)
	}
	if len(node.Private) != 2 {
		t.Fatalf("private len = %d, want 2", len(node.Private))
	}
	if node.Private[0].Heading != "decisions" {
		t.Errorf("private[0].heading = %q, want %q", node.Private[0].Heading, "decisions")
	}
	if len(node.Private[0].Content) != 1 || node.Private[0].Content[0] != "Decisions content line." {
		t.Errorf("private[0].content = %v", node.Private[0].Content)
	}
	if node.Private[1].Heading != "rationale" {
		t.Errorf("private[1].heading = %q, want %q", node.Private[1].Heading, "rationale")
	}
	if len(node.Private[1].Content) != 1 || node.Private[1].Content[0] != "Rationale content line." {
		t.Errorf("private[1].content = %v", node.Private[1].Content)
	}
}

func TestNodeParse_NoPublicSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/decisions\nDescription line.\n# Rationale\nRationale content.\n"
	testWriteNodeFile(t, "ROOT/decisions", body)

	node, err := parsenode.NodeParse("ROOT/decisions")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public != nil {
		t.Errorf("public = %v, want nil", node.Public)
	}
	if node.Agent != nil {
		t.Errorf("agent = %v, want nil", node.Agent)
	}
	if len(node.Private) != 1 {
		t.Fatalf("private len = %d, want 1", len(node.Private))
	}
	if node.Private[0].Heading != "rationale" {
		t.Errorf("private[0].heading = %q, want %q", node.Private[0].Heading, "rationale")
	}
}

func TestNodeParse_PublicSectionContentBeforeFirstSubsection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/a\nName content.\n# Public\nPreamble line one.\nPreamble line two.\n## Interface\nInterface content.\n"
	testWriteNodeFile(t, "ROOT/a", body)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Content) != 2 {
		t.Fatalf("public.content len = %d, want 2", len(node.Public.Content))
	}
	if node.Public.Content[0] != "Preamble line one." {
		t.Errorf("public.content[0] = %q, want %q", node.Public.Content[0], "Preamble line one.")
	}
	if node.Public.Content[1] != "Preamble line two." {
		t.Errorf("public.content[1] = %q, want %q", node.Public.Content[1], "Preamble line two.")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("public.subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("subsection.heading = %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if len(node.Public.Subsections[0].Content) != 1 || node.Public.Subsections[0].Content[0] != "Interface content." {
		t.Errorf("subsection.content = %v", node.Public.Subsections[0].Content)
	}
}

func TestNodeParse_PublicSectionNoContentOrSubsections(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/b\nName content.\n# Public\n# Agent\nAgent content.\n"
	testWriteNodeFile(t, "ROOT/b", body)

	node, err := parsenode.NodeParse("ROOT/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Content) != 0 {
		t.Errorf("public.content = %v, want empty", node.Public.Content)
	}
	if len(node.Public.Subsections) != 0 {
		t.Errorf("public.subsections = %v, want empty", node.Public.Subsections)
	}
}

func TestNodeParse_AgentSectionWithSubsections(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/c\nName content.\n# Agent\nPreamble line.\n## Implementation guidance\nGuidance content.\n## Contracts\nContracts content.\n"
	testWriteNodeFile(t, "ROOT/c", body)

	node, err := parsenode.NodeParse("ROOT/c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Agent == nil {
		t.Fatal("agent = nil, want present")
	}
	if len(node.Agent.Content) != 1 || node.Agent.Content[0] != "Preamble line." {
		t.Errorf("agent.content = %v, want [\"Preamble line.\"]", node.Agent.Content)
	}
	if node.Agent.RawHeading != "# Agent" {
		t.Errorf("agent.raw_heading = %q, want %q", node.Agent.RawHeading, "# Agent")
	}
	if len(node.Agent.Subsections) != 2 {
		t.Fatalf("agent.subsections len = %d, want 2", len(node.Agent.Subsections))
	}
	if node.Agent.Subsections[0].Heading != "implementation guidance" {
		t.Errorf("subsection[0].heading = %q, want %q", node.Agent.Subsections[0].Heading, "implementation guidance")
	}
	if len(node.Agent.Subsections[0].Content) != 1 || node.Agent.Subsections[0].Content[0] != "Guidance content." {
		t.Errorf("subsection[0].content = %v", node.Agent.Subsections[0].Content)
	}
	if node.Agent.Subsections[1].Heading != "contracts" {
		t.Errorf("subsection[1].heading = %q, want %q", node.Agent.Subsections[1].Heading, "contracts")
	}
	if len(node.Agent.Subsections[1].Content) != 1 || node.Agent.Subsections[1].Content[0] != "Contracts content." {
		t.Errorf("subsection[1].content = %v", node.Agent.Subsections[1].Content)
	}
}

func TestNodeParse_PrivateSectionsPreserveFileOrder(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/d\nName content.\n# TODO\nTodo content.\n# Decisions\nDecisions content.\n# Rationale\nRationale content.\n"
	testWriteNodeFile(t, "ROOT/d", body)

	node, err := parsenode.NodeParse("ROOT/d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(node.Private) != 3 {
		t.Fatalf("private len = %d, want 3", len(node.Private))
	}
	if node.Private[0].Heading != "todo" {
		t.Errorf("private[0].heading = %q, want %q", node.Private[0].Heading, "todo")
	}
	if node.Private[1].Heading != "decisions" {
		t.Errorf("private[1].heading = %q, want %q", node.Private[1].Heading, "decisions")
	}
	if node.Private[2].Heading != "rationale" {
		t.Errorf("private[2].heading = %q, want %q", node.Private[2].Heading, "rationale")
	}
}

func TestNodeParse_ContentIsRawMarkdown(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/e\nName content.\n# Public\n## Details\n### Sub-heading\n**Bold text**\n```go\ncode content\n```\n"
	testWriteNodeFile(t, "ROOT/e", body)

	node, err := parsenode.NodeParse("ROOT/e")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections len = %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	if sub.Heading != "details" {
		t.Errorf("subsection.heading = %q, want %q", sub.Heading, "details")
	}
	expected := []string{"### Sub-heading", "**Bold text**", "```go", "code content", "```"}
	if len(sub.Content) != len(expected) {
		t.Fatalf("subsection.content len = %d, want %d: %v", len(sub.Content), len(expected), sub.Content)
	}
	for i, line := range expected {
		if sub.Content[i] != line {
			t.Errorf("content[%d] = %q, want %q", i, sub.Content[i], line)
		}
	}
}

func TestNodeParse_CaseInsensitivePublicDetection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/f\nName content.\n# PUBLIC\nPublic content.\n"
	testWriteNodeFile(t, "ROOT/f", body)

	node, err := parsenode.NodeParse("ROOT/f")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if node.Public.Heading != "public" {
		t.Errorf("public.heading = %q, want %q", node.Public.Heading, "public")
	}
}

func TestNodeParse_PublicWithMixedCaseAndExtraWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/g\nName content.\n#   PuBLiC\nPublic content.\n"
	testWriteNodeFile(t, "ROOT/g", body)

	node, err := parsenode.NodeParse("ROOT/g")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if node.Public.Heading != "public" {
		t.Errorf("public.heading = %q, want %q", node.Public.Heading, "public")
	}
}

func TestNodeParse_NodeNameWithVariedWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "#   ROOT/e\nName content.\n"
	testWriteNodeFile(t, "ROOT/e", body)

	node, err := parsenode.NodeParse("ROOT/e")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/e" {
		t.Errorf("name_section.heading = %q, want %q", node.NameSection.Heading, "root/e")
	}
}

func TestNodeParse_SubsectionHeadingsAreNormalized(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/h\nName content.\n# Public\n##   Interface\nInterface content.\n## CONSTRAINTS\nConstraints content.\n"
	testWriteNodeFile(t, "ROOT/h", body)

	node, err := parsenode.NodeParse("ROOT/h")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("subsections len = %d, want 2", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("subsection[0].heading = %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("subsection[1].heading = %q, want %q", node.Public.Subsections[1].Heading, "constraints")
	}
}

func TestNodeParse_ClosingHashesAreStripped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/i\nName content.\n# Public\n## Interface ##\nInterface content.\n"
	testWriteNodeFile(t, "ROOT/i", body)

	node, err := parsenode.NodeParse("ROOT/i")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections len = %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	if sub.Heading != "interface" {
		t.Errorf("heading = %q, want %q", sub.Heading, "interface")
	}
	if sub.RawHeading != "## Interface ##" {
		t.Errorf("raw_heading = %q, want %q", sub.RawHeading, "## Interface ##")
	}
}

func TestNodeParse_RawHeadingPreservesOriginalLine(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/j\nName content.\n# Public\n## Interface\nInterface content.\n"
	testWriteNodeFile(t, "ROOT/j", body)

	node, err := parsenode.NodeParse("ROOT/j")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if node.Public.RawHeading != "# Public" {
		t.Errorf("public.raw_heading = %q, want %q", node.Public.RawHeading, "# Public")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].RawHeading != "## Interface" {
		t.Errorf("interface.raw_heading = %q, want %q", node.Public.Subsections[0].RawHeading, "## Interface")
	}
}

func TestNodeParse_RawHeadingPreservesCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/k\nName content.\n# PUBLIC\nPublic content.\n"
	testWriteNodeFile(t, "ROOT/k", body)

	node, err := parsenode.NodeParse("ROOT/k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if node.Public.Heading != "public" {
		t.Errorf("public.heading = %q, want %q", node.Public.Heading, "public")
	}
	if node.Public.RawHeading != "# PUBLIC" {
		t.Errorf("public.raw_heading = %q, want %q", node.Public.RawHeading, "# PUBLIC")
	}
}

func TestNodeParse_RawHeadingPreservesClosingHashes(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/l\nName content.\n# Public\n## Foo ##\nFoo content.\n"
	testWriteNodeFile(t, "ROOT/l", body)

	node, err := parsenode.NodeParse("ROOT/l")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections len = %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	if sub.Heading != "foo" {
		t.Errorf("heading = %q, want %q", sub.Heading, "foo")
	}
	if sub.RawHeading != "## Foo ##" {
		t.Errorf("raw_heading = %q, want %q", sub.RawHeading, "## Foo ##")
	}
}

func TestNodeParse_RawHeadingPreservesExtraWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/m\nName content.\n#   Public\nPublic content.\n"
	testWriteNodeFile(t, "ROOT/m", body)

	node, err := parsenode.NodeParse("ROOT/m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if node.Public.Heading != "public" {
		t.Errorf("public.heading = %q, want %q", node.Public.Heading, "public")
	}
	if node.Public.RawHeading != "#   Public" {
		t.Errorf("public.raw_heading = %q, want %q", node.Public.RawHeading, "#   Public")
	}
}

func TestNodeParse_Level3AndDeeperHeadingsAreContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/n\nName content.\n# Public\n## Details\n### Third level heading\n#### Fourth level heading\n"
	testWriteNodeFile(t, "ROOT/n", body)

	node, err := parsenode.NodeParse("ROOT/n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections len = %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	if sub.Heading != "details" {
		t.Errorf("heading = %q, want %q", sub.Heading, "details")
	}
	if len(sub.Content) != 2 {
		t.Fatalf("content len = %d, want 2: %v", len(sub.Content), sub.Content)
	}
	if sub.Content[0] != "### Third level heading" {
		t.Errorf("content[0] = %q, want %q", sub.Content[0], "### Third level heading")
	}
	if sub.Content[1] != "#### Fourth level heading" {
		t.Errorf("content[1] = %q, want %q", sub.Content[1], "#### Fourth level heading")
	}
}

func TestNodeParse_FencedCodeBlockBacktickHeadingLike(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/o\nName content.\n# Public\n## Details\n```\n# Not a heading\n## Also not a heading\n```\n"
	testWriteNodeFile(t, "ROOT/o", body)

	node, err := parsenode.NodeParse("ROOT/o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections len = %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	if sub.Heading != "details" {
		t.Errorf("heading = %q, want %q", sub.Heading, "details")
	}
	found1, found2 := false, false
	for _, line := range sub.Content {
		if line == "# Not a heading" {
			found1 = true
		}
		if line == "## Also not a heading" {
			found2 = true
		}
	}
	if !found1 {
		t.Errorf("expected '# Not a heading' in content, got %v", sub.Content)
	}
	if !found2 {
		t.Errorf("expected '## Also not a heading' in content, got %v", sub.Content)
	}
}

func TestNodeParse_FencedCodeBlockTildeFence(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/p\nName content.\n# Public\n## Details\n~~~\n# Not a heading\n~~~\n"
	testWriteNodeFile(t, "ROOT/p", body)

	node, err := parsenode.NodeParse("ROOT/p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections len = %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	found := false
	for _, line := range sub.Content {
		if line == "# Not a heading" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected '# Not a heading' in content, got %v", sub.Content)
	}
}

func TestNodeParse_FencedCodeBlockWithLanguageTag(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/q\nName content.\n# Public\n## Details\n```go\n# Not a heading\n```\n"
	testWriteNodeFile(t, "ROOT/q", body)

	node, err := parsenode.NodeParse("ROOT/q")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections len = %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	found := false
	for _, line := range sub.Content {
		if line == "# Not a heading" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected '# Not a heading' in content, got %v", sub.Content)
	}
}

func TestNodeParse_BlankLinesBetweenHeadingAndContentArePreserved(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/r\nName content.\n# Public\n\nFirst content line.\n"
	testWriteNodeFile(t, "ROOT/r", body)

	node, err := parsenode.NodeParse("ROOT/r")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Content) != 2 {
		t.Fatalf("public.content len = %d, want 2: %v", len(node.Public.Content), node.Public.Content)
	}
	if node.Public.Content[0] != "" {
		t.Errorf("public.content[0] = %q, want %q", node.Public.Content[0], "")
	}
	if node.Public.Content[1] != "First content line." {
		t.Errorf("public.content[1] = %q, want %q", node.Public.Content[1], "First content line.")
	}
}

func TestNodeParse_FrontmatterIsSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "---\nkey: value\n---\n# ROOT/s\nBody content.\n"
	testWriteNodeFile(t, "ROOT/s", body)

	node, err := parsenode.NodeParse("ROOT/s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/s" {
		t.Errorf("heading = %q, want %q", node.NameSection.Heading, "root/s")
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "Body content." {
		t.Errorf("content = %v", node.NameSection.Content)
	}
}

func TestNodeParse_NoFrontmatterDelimiters(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/t\nBody content.\n"
	testWriteNodeFile(t, "ROOT/t", body)

	node, err := parsenode.NodeParse("ROOT/t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/t" {
		t.Errorf("heading = %q, want %q", node.NameSection.Heading, "root/t")
	}
}

func TestNodeParse_UnclosedFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "---\nkey: value\n# ROOT/u\nBody content.\n"
	testWriteNodeFile(t, "ROOT/u", body)

	_, err := parsenode.NodeParse("ROOT/u")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

func TestNodeParse_ArtifactReferenceRejected(t *testing.T) {
	_, err := parsenode.NodeParse("ARTIFACT/x")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrNotARootReference) {
		t.Errorf("error = %v, want ErrNotARootReference", err)
	}
}

func TestNodeParse_QualifierRejected(t *testing.T) {
	_, err := parsenode.NodeParse("ROOT/x(interface)")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrHasQualifier) {
		t.Errorf("error = %v, want ErrHasQualifier", err)
	}
}

func TestNodeParse_FileDoesNotExist(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := parsenode.NodeParse("ROOT/nonexistent/node")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrFileUnreadable) {
		t.Errorf("error = %v, want ErrFileUnreadable", err)
	}
}

func TestNodeParse_ContentBeforeFirstHeading(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "This is content before any heading.\n# ROOT/v\n"
	testWriteNodeFile(t, "ROOT/v", body)

	_, err := parsenode.NodeParse("ROOT/v")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

func TestNodeParse_Level2HeadingBeforeLevel1Heading(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "## Some subsection\n# ROOT/w\n"
	testWriteNodeFile(t, "ROOT/w", body)

	_, err := parsenode.NodeParse("ROOT/w")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

func TestNodeParse_EmptyBody(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/x2", "")

	_, err := parsenode.NodeParse("ROOT/x2")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

func TestNodeParse_NodeNameDoesNotMatchLogicalName(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/actual\nSome content.\n"
	testWriteNodeFile(t, "ROOT/other", body)

	_, err := parsenode.NodeParse("ROOT/other")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrNodeNameDoesNotMatch) {
		t.Errorf("error = %v, want ErrNodeNameDoesNotMatch", err)
	}
}

func TestNodeParse_NodeNameCaseMismatchIsNotAnError(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# root/casematch\nSome content.\n"
	testWriteNodeFile(t, "ROOT/CASEMATCH", body)

	node, err := parsenode.NodeParse("ROOT/CASEMATCH")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/casematch" {
		t.Errorf("heading = %q, want %q", node.NameSection.Heading, "root/casematch")
	}
}

func TestNodeParse_DuplicatePublicSectionSameCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/y\nName content.\n# Public\nFirst public.\n# Public\nSecond public.\n"
	testWriteNodeFile(t, "ROOT/y", body)

	_, err := parsenode.NodeParse("ROOT/y")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicatePublicSection) {
		t.Errorf("error = %v, want ErrDuplicatePublicSection", err)
	}
}

func TestNodeParse_DuplicatePublicSectionDifferentCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/z\nName content.\n# Public\nFirst public.\n# PUBLIC\nSecond public.\n"
	testWriteNodeFile(t, "ROOT/z", body)

	_, err := parsenode.NodeParse("ROOT/z")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicatePublicSection) {
		t.Errorf("error = %v, want ErrDuplicatePublicSection", err)
	}
}

func TestNodeParse_DuplicateAgentSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/aa\nName content.\n# Agent\nFirst agent.\n# Agent\nSecond agent.\n"
	testWriteNodeFile(t, "ROOT/aa", body)

	_, err := parsenode.NodeParse("ROOT/aa")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateAgentSection) {
		t.Errorf("error = %v, want ErrDuplicateAgentSection", err)
	}
}

func TestNodeParse_DuplicateSubsectionInPublicSameCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/bb\nName content.\n# Public\n## Interface\nFirst.\n## Interface\nSecond.\n"
	testWriteNodeFile(t, "ROOT/bb", body)

	_, err := parsenode.NodeParse("ROOT/bb")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}

func TestNodeParse_DuplicateSubsectionInPublicDifferentCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/cc\nName content.\n# Public\n## Interface\nFirst.\n## INTERFACE\nSecond.\n"
	testWriteNodeFile(t, "ROOT/cc", body)

	_, err := parsenode.NodeParse("ROOT/cc")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}

func TestNodeParse_DuplicateSubsectionInPublicWhitespaceVariation(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/dd\nName content.\n# Public\n## Interface\nFirst.\n##   Interface\nSecond.\n"
	testWriteNodeFile(t, "ROOT/dd", body)

	_, err := parsenode.NodeParse("ROOT/dd")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}

func TestNodeParse_DuplicateSubsectionInAgent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/ee\nName content.\n# Agent\n## Guidance\nFirst.\n## Guidance\nSecond.\n"
	testWriteNodeFile(t, "ROOT/ee", body)

	_, err := parsenode.NodeParse("ROOT/ee")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}
