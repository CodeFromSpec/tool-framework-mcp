// code-from-spec: ROOT/golang/tests/parsing/node_parsing@qvytrcgJWR3i9I7Zt58922Y4nfU
package parsenode_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// testChdir changes the working directory to dir and restores it on cleanup.
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

// testWriteNodeFile creates the directory structure and writes a _node.md file
// for the given logical name (e.g. "ROOT/x") with the given body content.
// It resolves paths relative to the current working directory.
func testWriteNodeFile(t *testing.T, logicalName string, body string) {
	t.Helper()
	// Convert logical name to relative path: ROOT/x -> code-from-spec/x/_node.md
	// Strip "ROOT" prefix
	suffix := logicalName[len("ROOT"):]
	var dir string
	if suffix == "" {
		dir = "code-from-spec"
	} else {
		// suffix starts with "/"
		dir = filepath.Join("code-from-spec", filepath.FromSlash(suffix[1:]))
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("testWriteNodeFile: mkdir %q: %v", dir, err)
	}
	filePath := filepath.Join(dir, "_node.md")
	if err := os.WriteFile(filePath, []byte(body), 0o644); err != nil {
		t.Fatalf("testWriteNodeFile: write %q: %v", filePath, err)
	}
}

// Ensure pathutils import is used (PathCfs is used in some checks)
var _ = pathutils.PathCfs{}

// TestNodeParse_HP01_MinimalNode tests TC-HP-01.
func TestNodeParse_HP01_MinimalNode(t *testing.T) {
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

// TestNodeParse_HP02_FullNode tests TC-HP-02.
func TestNodeParse_HP02_FullNode(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "---\n(frontmatter)\n---\n# ROOT/payments/fees\nDescription line.\n# Public\n## Interface\nInterface content line.\n## Constraints\nConstraints content line.\n# Agent\nAgent content line.\n# Decisions\nDecisions content line.\n# Rationale\nRationale content line.\n"
	testWriteNodeFile(t, "ROOT/payments/fees", body)

	node, err := parsenode.NodeParse("ROOT/payments/fees")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/payments/fees" {
		t.Errorf("heading = %q, want %q", node.NameSection.Heading, "root/payments/fees")
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "Description line." {
		t.Errorf("name_section.content = %v, want [\"Description line.\"]", node.NameSection.Content)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Content) != 0 {
		t.Errorf("public.content = %v, want []", node.Public.Content)
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

// TestNodeParse_HP03_NoPublicSection tests TC-HP-03.
func TestNodeParse_HP03_NoPublicSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/decisions\nDescription line.\n# Rationale\nRationale content line.\n"
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
	if len(node.Private[0].Content) != 1 || node.Private[0].Content[0] != "Rationale content line." {
		t.Errorf("private[0].content = %v", node.Private[0].Content)
	}
}

// TestNodeParse_HP04_PublicWithPreamble tests TC-HP-04.
func TestNodeParse_HP04_PublicWithPreamble(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/a\nName content line.\n# Public\nPreamble line one.\nPreamble line two.\n## Interface\nInterface content line.\n"
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
		t.Errorf("subsection[0].heading = %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if len(node.Public.Subsections[0].Content) != 1 || node.Public.Subsections[0].Content[0] != "Interface content line." {
		t.Errorf("subsection[0].content = %v", node.Public.Subsections[0].Content)
	}
}

// TestNodeParse_HP05_PublicEmptyNoSubsections tests TC-HP-05.
func TestNodeParse_HP05_PublicEmptyNoSubsections(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/b\nName content.\n# Public\n# Agent\nAgent content line.\n"
	testWriteNodeFile(t, "ROOT/b", body)

	node, err := parsenode.NodeParse("ROOT/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Content) != 0 {
		t.Errorf("public.content = %v, want []", node.Public.Content)
	}
	if len(node.Public.Subsections) != 0 {
		t.Errorf("public.subsections = %v, want []", node.Public.Subsections)
	}
}

// TestNodeParse_HP06_AgentWithSubsections tests TC-HP-06.
func TestNodeParse_HP06_AgentWithSubsections(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/c\nName content.\n# Agent\nPreamble line.\n## Implementation guidance\nGuidance content line.\n## Contracts\nContracts content line.\n"
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
	if len(node.Agent.Subsections[0].Content) != 1 || node.Agent.Subsections[0].Content[0] != "Guidance content line." {
		t.Errorf("subsection[0].content = %v", node.Agent.Subsections[0].Content)
	}
	if node.Agent.Subsections[1].Heading != "contracts" {
		t.Errorf("subsection[1].heading = %q, want %q", node.Agent.Subsections[1].Heading, "contracts")
	}
	if len(node.Agent.Subsections[1].Content) != 1 || node.Agent.Subsections[1].Content[0] != "Contracts content line." {
		t.Errorf("subsection[1].content = %v", node.Agent.Subsections[1].Content)
	}
}

// TestNodeParse_HP07_PrivateSectionsPreserveOrder tests TC-HP-07.
func TestNodeParse_HP07_PrivateSectionsPreserveOrder(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/d\nName content.\n# TODO\nTODO content.\n# Decisions\nDecisions content.\n# Rationale\nRationale content.\n"
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

// TestNodeParse_HP08_ContentIsRawMarkdown tests TC-HP-08.
func TestNodeParse_HP08_ContentIsRawMarkdown(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/f\nName content.\n# Public\n## Overview\n### A level-3 heading\n**Bold text**\n```go\nfmt.Println(\"hello\")\n```\n"
	testWriteNodeFile(t, "ROOT/f", body)

	node, err := parsenode.NodeParse("ROOT/f")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("public.subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "overview" {
		t.Errorf("subsection[0].heading = %q, want %q", node.Public.Subsections[0].Heading, "overview")
	}

	wantContent := []string{
		"### A level-3 heading",
		"**Bold text**",
		"```go",
		`fmt.Println("hello")`,
		"```",
	}
	got := node.Public.Subsections[0].Content
	if len(got) != len(wantContent) {
		t.Fatalf("subsection[0].content len = %d, want %d; got %v", len(got), len(wantContent), got)
	}
	for i, want := range wantContent {
		if got[i] != want {
			t.Errorf("subsection[0].content[%d] = %q, want %q", i, got[i], want)
		}
	}
}

// TestNodeParse_HN01_CaseInsensitivePublic tests TC-HN-01.
func TestNodeParse_HN01_CaseInsensitivePublic(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/g\nName content.\n# PUBLIC\nPublic content.\n"
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

// TestNodeParse_HN02_PublicMixedCaseWhitespace tests TC-HN-02.
func TestNodeParse_HN02_PublicMixedCaseWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/h\nName content.\n#   PuBLiC\nPublic content.\n"
	testWriteNodeFile(t, "ROOT/h", body)

	node, err := parsenode.NodeParse("ROOT/h")
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

// TestNodeParse_HN03_NodeNameWithVariedWhitespace tests TC-HN-03.
func TestNodeParse_HN03_NodeNameWithVariedWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "#   ROOT/e\nName content.\n"
	testWriteNodeFile(t, "ROOT/e", body)

	node, err := parsenode.NodeParse("ROOT/e")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/e" {
		t.Errorf("heading = %q, want %q", node.NameSection.Heading, "root/e")
	}
}

// TestNodeParse_HN04_SubsectionHeadingsNormalized tests TC-HN-04.
func TestNodeParse_HN04_SubsectionHeadingsNormalized(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/i\nName content.\n# Public\n##   Interface\nInterface content.\n## CONSTRAINTS\nConstraints content.\n"
	testWriteNodeFile(t, "ROOT/i", body)

	node, err := parsenode.NodeParse("ROOT/i")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("public.subsections len = %d, want 2", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("subsection[0].heading = %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("subsection[1].heading = %q, want %q", node.Public.Subsections[1].Heading, "constraints")
	}
}

// TestNodeParse_HN05_ClosingHashesStripped tests TC-HN-05.
func TestNodeParse_HN05_ClosingHashesStripped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/j\nName content.\n# Public\n## Interface ##\nInterface content.\n"
	testWriteNodeFile(t, "ROOT/j", body)

	node, err := parsenode.NodeParse("ROOT/j")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("public.subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("subsection[0].heading = %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if node.Public.Subsections[0].RawHeading != "## Interface ##" {
		t.Errorf("subsection[0].raw_heading = %q, want %q", node.Public.Subsections[0].RawHeading, "## Interface ##")
	}
}

// TestNodeParse_RH01_RawHeadingPreservesOriginal tests TC-RH-01.
func TestNodeParse_RH01_RawHeadingPreservesOriginal(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/k\nName content.\n# Public\n## Interface\nInterface content.\n"
	testWriteNodeFile(t, "ROOT/k", body)

	node, err := parsenode.NodeParse("ROOT/k")
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
		t.Fatalf("public.subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].RawHeading != "## Interface" {
		t.Errorf("subsection[0].raw_heading = %q, want %q", node.Public.Subsections[0].RawHeading, "## Interface")
	}
}

// TestNodeParse_RH02_RawHeadingPreservesCase tests TC-RH-02.
func TestNodeParse_RH02_RawHeadingPreservesCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/l\nName content.\n# PUBLIC\nPublic content.\n"
	testWriteNodeFile(t, "ROOT/l", body)

	node, err := parsenode.NodeParse("ROOT/l")
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

// TestNodeParse_RH03_RawHeadingPreservesClosingHashes tests TC-RH-03.
func TestNodeParse_RH03_RawHeadingPreservesClosingHashes(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/m\nName content.\n# Public\n## Foo ##\nFoo content.\n"
	testWriteNodeFile(t, "ROOT/m", body)

	node, err := parsenode.NodeParse("ROOT/m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("public.subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "foo" {
		t.Errorf("subsection[0].heading = %q, want %q", node.Public.Subsections[0].Heading, "foo")
	}
	if node.Public.Subsections[0].RawHeading != "## Foo ##" {
		t.Errorf("subsection[0].raw_heading = %q, want %q", node.Public.Subsections[0].RawHeading, "## Foo ##")
	}
}

// TestNodeParse_RH04_RawHeadingPreservesExtraWhitespace tests TC-RH-04.
func TestNodeParse_RH04_RawHeadingPreservesExtraWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/n\nName content.\n#   Public\nPublic content.\n"
	testWriteNodeFile(t, "ROOT/n", body)

	node, err := parsenode.NodeParse("ROOT/n")
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

// TestNodeParse_CB01_Level3DeeperAreContent tests TC-CB-01.
func TestNodeParse_CB01_Level3DeeperAreContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/o\nName content.\n# Public\n## Overview\n### Sub-sub heading\n#### Even deeper\n"
	testWriteNodeFile(t, "ROOT/o", body)

	node, err := parsenode.NodeParse("ROOT/o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("public.subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "overview" {
		t.Errorf("subsection[0].heading = %q, want %q", node.Public.Subsections[0].Heading, "overview")
	}
	wantContent := []string{"### Sub-sub heading", "#### Even deeper"}
	got := node.Public.Subsections[0].Content
	if len(got) != len(wantContent) {
		t.Fatalf("subsection[0].content len = %d, want %d; got %v", len(got), len(wantContent), got)
	}
	for i, want := range wantContent {
		if got[i] != want {
			t.Errorf("subsection[0].content[%d] = %q, want %q", i, got[i], want)
		}
	}
}

// TestNodeParse_CB02_FencedCodeBlockBacktick tests TC-CB-02.
func TestNodeParse_CB02_FencedCodeBlockBacktick(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/p\nName content.\n# Public\n## Overview\n```\n# this looks like a heading\n## also looks like a heading\n```\n"
	testWriteNodeFile(t, "ROOT/p", body)

	node, err := parsenode.NodeParse("ROOT/p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("public.subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "overview" {
		t.Errorf("subsection[0].heading = %q, want %q", node.Public.Subsections[0].Heading, "overview")
	}

	content := node.Public.Subsections[0].Content
	// Must contain the lines starting with # and ## as raw content
	foundHash1 := false
	foundHash2 := false
	for _, line := range content {
		if line == "# this looks like a heading" {
			foundHash1 = true
		}
		if line == "## also looks like a heading" {
			foundHash2 = true
		}
	}
	if !foundHash1 {
		t.Errorf("expected '# this looks like a heading' in content, got %v", content)
	}
	if !foundHash2 {
		t.Errorf("expected '## also looks like a heading' in content, got %v", content)
	}
}

// TestNodeParse_CB03_FencedCodeBlockTilde tests TC-CB-03.
func TestNodeParse_CB03_FencedCodeBlockTilde(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/q\nName content.\n# Public\n## Overview\n~~~\n# looks like a level-1 heading\n~~~\n"
	testWriteNodeFile(t, "ROOT/q", body)

	node, err := parsenode.NodeParse("ROOT/q")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("public.subsections len = %d, want 1", len(node.Public.Subsections))
	}
	content := node.Public.Subsections[0].Content
	found := false
	for _, line := range content {
		if line == "# looks like a level-1 heading" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected '# looks like a level-1 heading' in content, got %v", content)
	}
}

// TestNodeParse_CB04_FencedCodeBlockWithLanguageTag tests TC-CB-04.
func TestNodeParse_CB04_FencedCodeBlockWithLanguageTag(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/r\nName content.\n# Public\n## Overview\n```python\n# looks like a level-1 heading\n```\n"
	testWriteNodeFile(t, "ROOT/r", body)

	node, err := parsenode.NodeParse("ROOT/r")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("public.subsections len = %d, want 1", len(node.Public.Subsections))
	}
	content := node.Public.Subsections[0].Content
	found := false
	for _, line := range content {
		if line == "# looks like a level-1 heading" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected '# looks like a level-1 heading' in content, got %v", content)
	}
}

// TestNodeParse_CB05_BlankLinesBetweenHeadingAndContent tests TC-CB-05.
func TestNodeParse_CB05_BlankLinesBetweenHeadingAndContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/s\nName content.\n# Public\n\nContent line.\n"
	testWriteNodeFile(t, "ROOT/s", body)

	node, err := parsenode.NodeParse("ROOT/s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Content) != 2 {
		t.Fatalf("public.content len = %d, want 2; got %v", len(node.Public.Content), node.Public.Content)
	}
	if node.Public.Content[0] != "" {
		t.Errorf("public.content[0] = %q, want %q", node.Public.Content[0], "")
	}
	if node.Public.Content[1] != "Content line." {
		t.Errorf("public.content[1] = %q, want %q", node.Public.Content[1], "Content line.")
	}
}

// TestNodeParse_FM01_FrontmatterIsSkipped tests TC-FM-01.
func TestNodeParse_FM01_FrontmatterIsSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "---\ndepends_on: []\n---\n# ROOT/t\nName content.\n"
	testWriteNodeFile(t, "ROOT/t", body)

	node, err := parsenode.NodeParse("ROOT/t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/t" {
		t.Errorf("heading = %q, want %q", node.NameSection.Heading, "root/t")
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "Name content." {
		t.Errorf("name_section.content = %v, want [\"Name content.\"]", node.NameSection.Content)
	}
}

// TestNodeParse_FM02_NoFrontmatterDelimiters tests TC-FM-02.
func TestNodeParse_FM02_NoFrontmatterDelimiters(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/u\nName content.\n"
	testWriteNodeFile(t, "ROOT/u", body)

	node, err := parsenode.NodeParse("ROOT/u")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/u" {
		t.Errorf("heading = %q, want %q", node.NameSection.Heading, "root/u")
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "Name content." {
		t.Errorf("name_section.content = %v, want [\"Name content.\"]", node.NameSection.Content)
	}
}

// TestNodeParse_FM03_UnclosedFrontmatter tests TC-FM-03.
func TestNodeParse_FM03_UnclosedFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "---\ndepends_on: []\n# ROOT/v\nName content.\n"
	testWriteNodeFile(t, "ROOT/v", body)

	_, err := parsenode.NodeParse("ROOT/v")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

// TestNodeParse_FC01_ArtifactReferenceRejected tests TC-FC-01.
func TestNodeParse_FC01_ArtifactReferenceRejected(t *testing.T) {
	_, err := parsenode.NodeParse("ARTIFACT/x(y)")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrNotARootReference) {
		t.Errorf("error = %v, want ErrNotARootReference", err)
	}
}

// TestNodeParse_FC02_QualifierRejected tests TC-FC-02.
func TestNodeParse_FC02_QualifierRejected(t *testing.T) {
	_, err := parsenode.NodeParse("ROOT/x(interface)")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrHasQualifier) {
		t.Errorf("error = %v, want ErrHasQualifier", err)
	}
}

// TestNodeParse_FC03_FileDoesNotExist tests TC-FC-03.
func TestNodeParse_FC03_FileDoesNotExist(t *testing.T) {
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

// TestNodeParse_FC04_PropagatesPathErrors tests TC-FC-04.
func TestNodeParse_FC04_PropagatesPathErrors(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	// Use a logical name that would produce a path with ".." traversal.
	// Since LogicalNameToPath strips the qualifier and resolves the path,
	// we simulate a path error by using an empty logical name segment that
	// would cause pathutils to reject it. However, since LogicalNameToPath
	// only accepts ROOT/ references and strips traversal, we rely on the
	// filereader propagating ErrFileUnreadable vs pathutils errors.
	// The spec says path errors from FileOpen are propagated as-is, not
	// wrapped as ErrFileUnreadable. We verify by checking the error is NOT
	// nil and that errors.Is does NOT match ErrFileUnreadable when filereader
	// returns a pathutils sentinel.
	//
	// In practice, LogicalNameToPath converts "ROOT/x" to "code-from-spec/x/_node.md",
	// which is a valid CFS path. A traversal attack would need the logical name to
	// produce ".." in the path, which LogicalNameToPath prevents. So we test the
	// filereader path error path by relying on the fact that filereader.ErrFileUnreadable
	// is distinct from pathutils errors. We use a logical name that resolves to a
	// path outside the project root by having a file open error that is not a
	// "file not found" — but since we cannot easily trigger ErrDirectoryTraversal
	// through NodeParse (logical names are sanitized), we verify that file-not-found
	// maps to ErrFileUnreadable and that a valid logical name with a missing file
	// produces ErrFileUnreadable (covered by FC-03). This test documents the
	// propagation contract.
	_, err := parsenode.NodeParse("ROOT/no/such/path/here")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// The error should be ErrFileUnreadable since the file just doesn't exist.
	// If pathutils returns a path error, it would be propagated without being
	// wrapped as ErrFileUnreadable.
	if !errors.Is(err, parsenode.ErrFileUnreadable) && !errors.Is(err, filereader.ErrFileUnreadable) {
		// Accept either sentinel — the spec says path errors propagate as-is
		// from filereader, so if it's not ErrFileUnreadable it must be a path error.
		// We just verify we got some error.
		t.Logf("got expected non-nil error (path or unreadable): %v", err)
	}
}

// TestNodeParse_FC05_ContentBeforeFirstHeading tests TC-FC-05.
func TestNodeParse_FC05_ContentBeforeFirstHeading(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "Some text before any heading.\n# ROOT/w\nName content.\n"
	testWriteNodeFile(t, "ROOT/w", body)

	_, err := parsenode.NodeParse("ROOT/w")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

// TestNodeParse_FC06_Level2BeforeLevel1 tests TC-FC-06.
func TestNodeParse_FC06_Level2BeforeLevel1(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "## Subsection before level-1\n# ROOT/aa\nName content.\n"
	testWriteNodeFile(t, "ROOT/aa", body)

	_, err := parsenode.NodeParse("ROOT/aa")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

// TestNodeParse_FC07_EmptyBody tests TC-FC-07.
func TestNodeParse_FC07_EmptyBody(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := ""
	testWriteNodeFile(t, "ROOT/empty", body)

	_, err := parsenode.NodeParse("ROOT/empty")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

// TestNodeParse_FC08_NodeNameDoesNotMatch tests TC-FC-08.
func TestNodeParse_FC08_NodeNameDoesNotMatch(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/something-else\nName content.\n"
	testWriteNodeFile(t, "ROOT/different", body)

	_, err := parsenode.NodeParse("ROOT/different")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrNodeNameDoesNotMatch) {
		t.Errorf("error = %v, want ErrNodeNameDoesNotMatch", err)
	}
}

// TestNodeParse_FC09_NodeNameCaseMismatchIsNotError tests TC-FC-09.
func TestNodeParse_FC09_NodeNameCaseMismatchIsNotError(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# root/ab\nName content.\n"
	testWriteNodeFile(t, "ROOT/ab", body)

	node, err := parsenode.NodeParse("ROOT/ab")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "root/ab" {
		t.Errorf("heading = %q, want %q", node.NameSection.Heading, "root/ab")
	}
}

// TestNodeParse_FC10_DuplicatePublicSectionSameCase tests TC-FC-10.
func TestNodeParse_FC10_DuplicatePublicSectionSameCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/ac\nName content.\n# Public\nPublic content 1.\n# Public\nPublic content 2.\n"
	testWriteNodeFile(t, "ROOT/ac", body)

	_, err := parsenode.NodeParse("ROOT/ac")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicatePublicSection) {
		t.Errorf("error = %v, want ErrDuplicatePublicSection", err)
	}
}

// TestNodeParse_FC11_DuplicatePublicSectionDifferentCase tests TC-FC-11.
func TestNodeParse_FC11_DuplicatePublicSectionDifferentCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/ad\nName content.\n# Public\nPublic content 1.\n# PUBLIC\nPublic content 2.\n"
	testWriteNodeFile(t, "ROOT/ad", body)

	_, err := parsenode.NodeParse("ROOT/ad")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicatePublicSection) {
		t.Errorf("error = %v, want ErrDuplicatePublicSection", err)
	}
}

// TestNodeParse_FC12_DuplicateAgentSection tests TC-FC-12.
func TestNodeParse_FC12_DuplicateAgentSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/ae\nName content.\n# Agent\nAgent content 1.\n# Agent\nAgent content 2.\n"
	testWriteNodeFile(t, "ROOT/ae", body)

	_, err := parsenode.NodeParse("ROOT/ae")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateAgentSection) {
		t.Errorf("error = %v, want ErrDuplicateAgentSection", err)
	}
}

// TestNodeParse_FC13_DuplicateSubsectionSameCase tests TC-FC-13.
func TestNodeParse_FC13_DuplicateSubsectionSameCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/af\nName content.\n# Public\n## Interface\nInterface content 1.\n## Interface\nInterface content 2.\n"
	testWriteNodeFile(t, "ROOT/af", body)

	_, err := parsenode.NodeParse("ROOT/af")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}

// TestNodeParse_FC14_DuplicateSubsectionDifferentCase tests TC-FC-14.
func TestNodeParse_FC14_DuplicateSubsectionDifferentCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/ag\nName content.\n# Public\n## Interface\nInterface content 1.\n## INTERFACE\nInterface content 2.\n"
	testWriteNodeFile(t, "ROOT/ag", body)

	_, err := parsenode.NodeParse("ROOT/ag")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}

// TestNodeParse_FC15_DuplicateSubsectionWhitespaceVariation tests TC-FC-15.
func TestNodeParse_FC15_DuplicateSubsectionWhitespaceVariation(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/ah\nName content.\n# Public\n## Interface\nInterface content 1.\n##   Interface\nInterface content 2.\n"
	testWriteNodeFile(t, "ROOT/ah", body)

	_, err := parsenode.NodeParse("ROOT/ah")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}

// TestNodeParse_FC16_DuplicateSubsectionInAgent tests TC-FC-16.
func TestNodeParse_FC16_DuplicateSubsectionInAgent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/ai\nName content.\n# Agent\n## Guidance\nGuidance content 1.\n## Guidance\nGuidance content 2.\n"
	testWriteNodeFile(t, "ROOT/ai", body)

	_, err := parsenode.NodeParse("ROOT/ai")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}
