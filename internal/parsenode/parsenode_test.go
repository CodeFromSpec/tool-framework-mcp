// code-from-spec: ROOT/golang/tests/parsing/node_parsing@aNcw_3u9OzqLLIWDpWCYlSIXcOk
package parsenode_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
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

func testWriteNodeFile(t *testing.T, logicalName string, content string) {
	t.Helper()
	dir := filepath.Join("code-from-spec", filepath.FromSlash(logicalName[len("ROOT/"):]))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("testWriteNodeFile mkdir: %v", err)
	}
	path := filepath.Join(dir, "_node.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("testWriteNodeFile write: %v", err)
	}
}

func TestNodeParse_MinimalNode(t *testing.T) {
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
		t.Errorf("content = %v, want [A simple node.]", node.NameSection.Content)
	}
	if len(node.NameSection.Subsections) != 0 {
		t.Errorf("subsections = %v, want empty", node.NameSection.Subsections)
	}
	if node.Public != nil {
		t.Errorf("public should be absent")
	}
	if node.Agent != nil {
		t.Errorf("agent should be absent")
	}
	if len(node.Private) != 0 {
		t.Errorf("private = %v, want empty", node.Private)
	}
}

func TestNodeParse_FullNode(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "---\nkey: value\n---\n# ROOT/payments/fees\nDescription line.\n# Public\n## Interface\nInterface content line.\n## Constraints\nConstraints content line.\n# Agent\nAgent content line.\n# Decisions\nDecisions content line.\n# Rationale\nRationale content line.\n"
	if err := os.MkdirAll(filepath.Join("code-from-spec", "payments", "fees"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join("code-from-spec", "payments", "fees", "_node.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/payments/fees")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "root/payments/fees" {
		t.Errorf("heading = %q, want %q", node.NameSection.Heading, "root/payments/fees")
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "Description line." {
		t.Errorf("name content = %v, want [Description line.]", node.NameSection.Content)
	}
	if node.Public == nil {
		t.Fatal("public should be present")
	}
	if len(node.Public.Content) != 0 {
		t.Errorf("public content = %v, want empty", node.Public.Content)
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("public subsections len = %d, want 2", len(node.Public.Subsections))
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
		t.Fatal("agent should be present")
	}
	if len(node.Agent.Content) != 1 || node.Agent.Content[0] != "Agent content line." {
		t.Errorf("agent content = %v, want [Agent content line.]", node.Agent.Content)
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

	testWriteNodeFile(t, "ROOT/decisions", "# ROOT/decisions\nDescription line.\n# Rationale\nRationale content.\n")

	node, err := parsenode.NodeParse("ROOT/decisions")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public != nil {
		t.Errorf("public should be absent")
	}
	if node.Agent != nil {
		t.Errorf("agent should be absent")
	}
	if len(node.Private) != 1 {
		t.Fatalf("private len = %d, want 1", len(node.Private))
	}
	if node.Private[0].Heading != "rationale" {
		t.Errorf("private[0].heading = %q, want %q", node.Private[0].Heading, "rationale")
	}
}

func TestNodeParse_PublicContentBeforeSubsection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/a", "# ROOT/a\nName content.\n# Public\nPreamble line one.\nPreamble line two.\n## Interface\nInterface content.\n")

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public should be present")
	}
	if len(node.Public.Content) != 2 || node.Public.Content[0] != "Preamble line one." || node.Public.Content[1] != "Preamble line two." {
		t.Errorf("public content = %v", node.Public.Content)
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("subsection heading = %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if len(node.Public.Subsections[0].Content) != 1 || node.Public.Subsections[0].Content[0] != "Interface content." {
		t.Errorf("subsection content = %v", node.Public.Subsections[0].Content)
	}
}

func TestNodeParse_PublicNoContentNoSubsections(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/b", "# ROOT/b\nName content.\n# Public\n# Agent\nAgent content.\n")

	node, err := parsenode.NodeParse("ROOT/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public should be present")
	}
	if len(node.Public.Content) != 0 {
		t.Errorf("public content = %v, want empty", node.Public.Content)
	}
	if len(node.Public.Subsections) != 0 {
		t.Errorf("public subsections = %v, want empty", node.Public.Subsections)
	}
}

func TestNodeParse_AgentWithSubsections(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/c", "# ROOT/c\nName content.\n# Agent\nPreamble line.\n## Implementation guidance\nGuidance content.\n## Contracts\nContracts content.\n")

	node, err := parsenode.NodeParse("ROOT/c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Agent == nil {
		t.Fatal("agent should be present")
	}
	if len(node.Agent.Content) != 1 || node.Agent.Content[0] != "Preamble line." {
		t.Errorf("agent content = %v, want [Preamble line.]", node.Agent.Content)
	}
	if node.Agent.RawHeading != "# Agent" {
		t.Errorf("agent raw_heading = %q, want %q", node.Agent.RawHeading, "# Agent")
	}
	if len(node.Agent.Subsections) != 2 {
		t.Fatalf("agent subsections len = %d, want 2", len(node.Agent.Subsections))
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

func TestNodeParse_PrivateSectionsPreserveOrder(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/d", "# ROOT/d\nName content.\n# TODO\nTodo content.\n# Decisions\nDecisions content.\n# Rationale\nRationale content.\n")

	node, err := parsenode.NodeParse("ROOT/d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(node.Private) != 3 {
		t.Fatalf("private len = %d, want 3", len(node.Private))
	}
	wantHeadings := []string{"todo", "decisions", "rationale"}
	for i, want := range wantHeadings {
		if node.Private[i].Heading != want {
			t.Errorf("private[%d].heading = %q, want %q", i, node.Private[i].Heading, want)
		}
	}
}

func TestNodeParse_ContentIsRawMarkdown(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/e\nName content.\n# Public\n## Details\n### Sub-heading\n**Bold text**\n```go\ncode content\n```\n"
	testWriteNodeFile(t, "ROOT/e", content)

	node, err := parsenode.NodeParse("ROOT/e")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public should be present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections len = %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	wantContent := []string{"### Sub-heading", "**Bold text**", "```go", "code content", "```"}
	if len(sub.Content) != len(wantContent) {
		t.Fatalf("subsection content len = %d, want %d: %v", len(sub.Content), len(wantContent), sub.Content)
	}
	for i, want := range wantContent {
		if sub.Content[i] != want {
			t.Errorf("content[%d] = %q, want %q", i, sub.Content[i], want)
		}
	}
}

func TestNodeParse_CaseInsensitivePublicDetection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/f", "# ROOT/f\nName content.\n# PUBLIC\nContent.\n")

	node, err := parsenode.NodeParse("ROOT/f")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public should be present")
	}
	if node.Public.Heading != "public" {
		t.Errorf("public.heading = %q, want %q", node.Public.Heading, "public")
	}
}

func TestNodeParse_PublicMixedCaseExtraWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/g", "# ROOT/g\nName content.\n#   PuBLiC\nContent.\n")

	node, err := parsenode.NodeParse("ROOT/g")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public should be present")
	}
	if node.Public.Heading != "public" {
		t.Errorf("public.heading = %q, want %q", node.Public.Heading, "public")
	}
}

func TestNodeParse_NodeNameVariedWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/e", "#   ROOT/e\nName content.\n")

	node, err := parsenode.NodeParse("ROOT/e")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "root/e" {
		t.Errorf("heading = %q, want %q", node.NameSection.Heading, "root/e")
	}
}

func TestNodeParse_SubsectionHeadingsNormalized(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/h", "# ROOT/h\nName content.\n# Public\n##   Interface\nInterface content.\n## CONSTRAINTS\nConstraints content.\n")

	node, err := parsenode.NodeParse("ROOT/h")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public should be present")
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

func TestNodeParse_ClosingHashesStripped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/i", "# ROOT/i\nName content.\n# Public\n## Interface ##\nContent.\n")

	node, err := parsenode.NodeParse("ROOT/i")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public should be present")
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

	testWriteNodeFile(t, "ROOT/j", "# ROOT/j\nName content.\n# Public\n## Interface\nContent.\n")

	node, err := parsenode.NodeParse("ROOT/j")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public should be present")
	}
	if node.Public.RawHeading != "# Public" {
		t.Errorf("public.raw_heading = %q, want %q", node.Public.RawHeading, "# Public")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].RawHeading != "## Interface" {
		t.Errorf("subsection.raw_heading = %q, want %q", node.Public.Subsections[0].RawHeading, "## Interface")
	}
}

func TestNodeParse_RawHeadingPreservesCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/k", "# ROOT/k\nName content.\n# PUBLIC\nContent.\n")

	node, err := parsenode.NodeParse("ROOT/k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public should be present")
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

	testWriteNodeFile(t, "ROOT/l", "# ROOT/l\nName content.\n# Public\n## Foo ##\nContent.\n")

	node, err := parsenode.NodeParse("ROOT/l")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public should be present")
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

	testWriteNodeFile(t, "ROOT/m", "# ROOT/m\nName content.\n#   Public\nContent.\n")

	node, err := parsenode.NodeParse("ROOT/m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public should be present")
	}
	if node.Public.Heading != "public" {
		t.Errorf("public.heading = %q, want %q", node.Public.Heading, "public")
	}
	if node.Public.RawHeading != "#   Public" {
		t.Errorf("public.raw_heading = %q, want %q", node.Public.RawHeading, "#   Public")
	}
}

func TestNodeParse_Level3AndDeeperAreContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/n", "# ROOT/n\nName content.\n# Public\n## Details\n### Third level heading\n#### Fourth level heading\n")

	node, err := parsenode.NodeParse("ROOT/n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public should be present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections len = %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
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

func TestNodeParse_FencedCodeBlockBacktick(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/o\nName content.\n# Public\n## Details\n```\n# heading inside\n## subsection inside\n```\n"
	testWriteNodeFile(t, "ROOT/o", content)

	node, err := parsenode.NodeParse("ROOT/o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public should be present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections len = %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	wantContent := []string{"```", "# heading inside", "## subsection inside", "```"}
	if len(sub.Content) != len(wantContent) {
		t.Fatalf("content len = %d, want %d: %v", len(sub.Content), len(wantContent), sub.Content)
	}
	for i, want := range wantContent {
		if sub.Content[i] != want {
			t.Errorf("content[%d] = %q, want %q", i, sub.Content[i], want)
		}
	}
}

func TestNodeParse_FencedCodeBlockTilde(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/p\nName content.\n# Public\n## Details\n~~~\n# heading inside\n~~~\n"
	testWriteNodeFile(t, "ROOT/p", content)

	node, err := parsenode.NodeParse("ROOT/p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public should be present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections len = %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	wantContent := []string{"~~~", "# heading inside", "~~~"}
	if len(sub.Content) != len(wantContent) {
		t.Fatalf("content len = %d, want %d: %v", len(sub.Content), len(wantContent), sub.Content)
	}
	for i, want := range wantContent {
		if sub.Content[i] != want {
			t.Errorf("content[%d] = %q, want %q", i, sub.Content[i], want)
		}
	}
}

func TestNodeParse_FencedCodeBlockWithLanguageTag(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/q\nName content.\n# Public\n## Details\n```go\n# heading inside\n```\n"
	testWriteNodeFile(t, "ROOT/q", content)

	node, err := parsenode.NodeParse("ROOT/q")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public should be present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections len = %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	wantContent := []string{"```go", "# heading inside", "```"}
	if len(sub.Content) != len(wantContent) {
		t.Fatalf("content len = %d, want %d: %v", len(sub.Content), len(wantContent), sub.Content)
	}
	for i, want := range wantContent {
		if sub.Content[i] != want {
			t.Errorf("content[%d] = %q, want %q", i, sub.Content[i], want)
		}
	}
}

func TestNodeParse_BlankLinesPreserved(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/r", "# ROOT/r\nName content.\n# Public\n\nFirst content line.\n")

	node, err := parsenode.NodeParse("ROOT/r")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public should be present")
	}
	if len(node.Public.Content) != 2 {
		t.Fatalf("public content len = %d, want 2: %v", len(node.Public.Content), node.Public.Content)
	}
	if node.Public.Content[0] != "" {
		t.Errorf("public.content[0] = %q, want empty string", node.Public.Content[0])
	}
	if node.Public.Content[1] != "First content line." {
		t.Errorf("public.content[1] = %q, want %q", node.Public.Content[1], "First content line.")
	}
}

func TestNodeParse_FrontmatterIsSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "---\nkey: value\nauthor: test\n---\n# ROOT/s\nBody content.\n"
	testWriteNodeFile(t, "ROOT/s", content)

	node, err := parsenode.NodeParse("ROOT/s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "root/s" {
		t.Errorf("heading = %q, want %q", node.NameSection.Heading, "root/s")
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "Body content." {
		t.Errorf("name content = %v", node.NameSection.Content)
	}
}

func TestNodeParse_NoFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/t", "# ROOT/t\nBody content.\n")

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

	testWriteNodeFile(t, "ROOT/u", "---\nkey: value\n# ROOT/u\nBody content.\n")

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

func TestNodeParse_PropagatesPathErrors(t *testing.T) {
	_, err := parsenode.NodeParse("ROOT/../etc/passwd")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filereader.ErrFileUnreadable) && !errors.Is(err, parsenode.ErrFileUnreadable) {
		t.Errorf("error = %v, want a path/file error", err)
	}
}

func TestNodeParse_ContentBeforeFirstHeading(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/v", "This is content before any heading.\n# ROOT/v\n")

	_, err := parsenode.NodeParse("ROOT/v")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

func TestNodeParse_Level2HeadingBeforeLevel1(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/w", "## Some subsection\n# ROOT/w\n")

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

func TestNodeParse_NodeNameDoesNotMatch(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/other", "# ROOT/actual\nContent.\n")

	_, err := parsenode.NodeParse("ROOT/other")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrNodeNameDoesNotMatch) {
		t.Errorf("error = %v, want ErrNodeNameDoesNotMatch", err)
	}
}

func TestNodeParse_NodeNameCaseMismatchNotError(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	if err := os.MkdirAll(filepath.Join("code-from-spec", "CASEMATCH"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join("code-from-spec", "CASEMATCH", "_node.md"), []byte("# root/casematch\nContent.\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/CASEMATCH")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "root/casematch" {
		t.Errorf("heading = %q, want %q", node.NameSection.Heading, "root/casematch")
	}
}

func TestNodeParse_DuplicatePublicSameCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/y", "# ROOT/y\nName content.\n# Public\nFirst public.\n# Public\nSecond public.\n")

	_, err := parsenode.NodeParse("ROOT/y")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicatePublicSection) {
		t.Errorf("error = %v, want ErrDuplicatePublicSection", err)
	}
}

func TestNodeParse_DuplicatePublicDifferentCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/z", "# ROOT/z\nName content.\n# Public\nFirst public.\n# PUBLIC\nSecond public.\n")

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

	testWriteNodeFile(t, "ROOT/aa", "# ROOT/aa\nName content.\n# Agent\nFirst agent.\n# Agent\nSecond agent.\n")

	_, err := parsenode.NodeParse("ROOT/aa")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateAgentSection) {
		t.Errorf("error = %v, want ErrDuplicateAgentSection", err)
	}
}

func TestNodeParse_DuplicateSubsectionPublicSameCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/bb", "# ROOT/bb\nName content.\n# Public\n## Interface\nFirst.\n## Interface\nSecond.\n")

	_, err := parsenode.NodeParse("ROOT/bb")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}

func TestNodeParse_DuplicateSubsectionPublicDifferentCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/cc", "# ROOT/cc\nName content.\n# Public\n## Interface\nFirst.\n## INTERFACE\nSecond.\n")

	_, err := parsenode.NodeParse("ROOT/cc")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}

func TestNodeParse_DuplicateSubsectionPublicWhitespaceVariation(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/dd", "# ROOT/dd\nName content.\n# Public\n## Interface\nFirst.\n##   Interface\nSecond.\n")

	_, err := parsenode.NodeParse("ROOT/dd")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}

func TestNodeParse_DuplicateSubsectionAgent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/ee", "# ROOT/ee\nName content.\n# Agent\n## Guidance\nFirst.\n## Guidance\nSecond.\n")

	_, err := parsenode.NodeParse("ROOT/ee")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}
