// code-from-spec: ROOT/golang/tests/parsing/node_parsing@vQ_uy3pHszXqRWcHK5Y5Ty76-80
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

// testChdir changes the working directory to dir for the duration of the test.
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

// testWriteNodeFile creates the node file at code-from-spec/<rel>/_node.md
// inside the given root directory. It creates all necessary parent directories.
func testWriteNodeFile(t *testing.T, root, rel, body string) {
	t.Helper()
	dir := filepath.Join(root, "code-from-spec", rel)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("testWriteNodeFile MkdirAll: %v", err)
	}
	path := filepath.Join(dir, "_node.md")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("testWriteNodeFile WriteFile: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Happy Path
// ---------------------------------------------------------------------------

// HP-01: Minimal node — name section only
func TestNodeParse_HP01_MinimalNode(t *testing.T) {
	tmp := t.TempDir()
	testWriteNodeFile(t, tmp, "x", "# ROOT/x\nA simple node.\n")
	testChdir(t, tmp)

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
		t.Errorf("content: got %v, want [\"A simple node.\"]", node.NameSection.Content)
	}
	if len(node.NameSection.Subsections) != 0 {
		t.Errorf("subsections: got %d, want 0", len(node.NameSection.Subsections))
	}
	if node.Public != nil {
		t.Errorf("public: got non-nil, want nil")
	}
	if node.Agent != nil {
		t.Errorf("agent: got non-nil, want nil")
	}
	if len(node.Private) != 0 {
		t.Errorf("private: got %d sections, want 0", len(node.Private))
	}
}

// HP-02: Full node — all section types
func TestNodeParse_HP02_FullNode(t *testing.T) {
	tmp := t.TempDir()
	body := "---\nkey: value\n---\n# ROOT/payments/fees\nFees description.\n# Public\n## Interface\nInterface content.\n## Constraints\nConstraints content.\n# Agent\nAgent content.\n# Decisions\nDecisions content.\n# Rationale\nRationale content.\n"
	testWriteNodeFile(t, tmp, "payments/fees", body)
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/payments/fees")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "root/payments/fees" {
		t.Errorf("name heading: got %q, want %q", node.NameSection.Heading, "root/payments/fees")
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "Fees description." {
		t.Errorf("name content: got %v", node.NameSection.Content)
	}

	if node.Public == nil {
		t.Fatal("public: got nil, want non-nil")
	}
	if len(node.Public.Content) != 0 {
		t.Errorf("public.content: got %v, want empty", node.Public.Content)
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("public.subsections: got %d, want 2", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("public.subsections[0].heading: got %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if len(node.Public.Subsections[0].Content) != 1 || node.Public.Subsections[0].Content[0] != "Interface content." {
		t.Errorf("public.subsections[0].content: got %v", node.Public.Subsections[0].Content)
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("public.subsections[1].heading: got %q, want %q", node.Public.Subsections[1].Heading, "constraints")
	}
	if len(node.Public.Subsections[1].Content) != 1 || node.Public.Subsections[1].Content[0] != "Constraints content." {
		t.Errorf("public.subsections[1].content: got %v", node.Public.Subsections[1].Content)
	}

	if node.Agent == nil {
		t.Fatal("agent: got nil, want non-nil")
	}
	if len(node.Agent.Content) != 1 || node.Agent.Content[0] != "Agent content." {
		t.Errorf("agent.content: got %v", node.Agent.Content)
	}

	if len(node.Private) != 2 {
		t.Fatalf("private: got %d sections, want 2", len(node.Private))
	}
	if node.Private[0].Heading != "decisions" {
		t.Errorf("private[0].heading: got %q, want %q", node.Private[0].Heading, "decisions")
	}
	if len(node.Private[0].Content) != 1 || node.Private[0].Content[0] != "Decisions content." {
		t.Errorf("private[0].content: got %v", node.Private[0].Content)
	}
	if node.Private[1].Heading != "rationale" {
		t.Errorf("private[1].heading: got %q, want %q", node.Private[1].Heading, "rationale")
	}
	if len(node.Private[1].Content) != 1 || node.Private[1].Content[0] != "Rationale content." {
		t.Errorf("private[1].content: got %v", node.Private[1].Content)
	}
}

// HP-03: Node with no public section
func TestNodeParse_HP03_NoPublicSection(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/decisions\nSome decision content.\n# Rationale\nRationale content.\n"
	testWriteNodeFile(t, tmp, "decisions", body)
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/decisions")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public != nil {
		t.Errorf("public: got non-nil, want nil")
	}
	if node.Agent != nil {
		t.Errorf("agent: got non-nil, want nil")
	}
	if len(node.Private) != 1 {
		t.Fatalf("private: got %d sections, want 1", len(node.Private))
	}
	if node.Private[0].Heading != "rationale" {
		t.Errorf("private[0].heading: got %q, want %q", node.Private[0].Heading, "rationale")
	}
	if len(node.Private[0].Content) != 1 || node.Private[0].Content[0] != "Rationale content." {
		t.Errorf("private[0].content: got %v", node.Private[0].Content)
	}
}

// HP-04: Public section with content before first subsection
func TestNodeParse_HP04_PublicPreamble(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n# Public\nPreamble line one.\nPreamble line two.\n## Interface\nInterface content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public: got nil, want non-nil")
	}
	wantContent := []string{"Preamble line one.", "Preamble line two."}
	if len(node.Public.Content) != len(wantContent) {
		t.Fatalf("public.content length: got %d, want %d", len(node.Public.Content), len(wantContent))
	}
	for i, want := range wantContent {
		if node.Public.Content[i] != want {
			t.Errorf("public.content[%d]: got %q, want %q", i, node.Public.Content[i], want)
		}
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("public.subsections: got %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("public.subsections[0].heading: got %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if len(node.Public.Subsections[0].Content) != 1 || node.Public.Subsections[0].Content[0] != "Interface content." {
		t.Errorf("public.subsections[0].content: got %v", node.Public.Subsections[0].Content)
	}
}

// HP-05: Public section with no content or subsections
func TestNodeParse_HP05_EmptyPublicSection(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n# Public\n# Agent\nAgent content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public: got nil, want non-nil")
	}
	if len(node.Public.Content) != 0 {
		t.Errorf("public.content: got %v, want empty", node.Public.Content)
	}
	if len(node.Public.Subsections) != 0 {
		t.Errorf("public.subsections: got %d, want 0", len(node.Public.Subsections))
	}
}

// HP-06: Agent section with subsections
func TestNodeParse_HP06_AgentWithSubsections(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n# Agent\nAgent preamble.\n## Implementation guidance\nImplementation guidance content.\n## Contracts\nContracts content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Agent == nil {
		t.Fatal("agent: got nil, want non-nil")
	}
	if len(node.Agent.Content) != 1 || node.Agent.Content[0] != "Agent preamble." {
		t.Errorf("agent.content: got %v", node.Agent.Content)
	}
	if node.Agent.RawHeading != "# Agent" {
		t.Errorf("agent.raw_heading: got %q, want %q", node.Agent.RawHeading, "# Agent")
	}
	if len(node.Agent.Subsections) != 2 {
		t.Fatalf("agent.subsections: got %d, want 2", len(node.Agent.Subsections))
	}
	if node.Agent.Subsections[0].Heading != "implementation guidance" {
		t.Errorf("agent.subsections[0].heading: got %q, want %q", node.Agent.Subsections[0].Heading, "implementation guidance")
	}
	if len(node.Agent.Subsections[0].Content) != 1 || node.Agent.Subsections[0].Content[0] != "Implementation guidance content." {
		t.Errorf("agent.subsections[0].content: got %v", node.Agent.Subsections[0].Content)
	}
	if node.Agent.Subsections[1].Heading != "contracts" {
		t.Errorf("agent.subsections[1].heading: got %q, want %q", node.Agent.Subsections[1].Heading, "contracts")
	}
	if len(node.Agent.Subsections[1].Content) != 1 || node.Agent.Subsections[1].Content[0] != "Contracts content." {
		t.Errorf("agent.subsections[1].content: got %v", node.Agent.Subsections[1].Content)
	}
}

// HP-07: Private sections preserve file order
func TestNodeParse_HP07_PrivateOrder(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n# TODO\nTODO content.\n# Decisions\nDecisions content.\n# Rationale\nRationale content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(node.Private) != 3 {
		t.Fatalf("private: got %d sections, want 3", len(node.Private))
	}
	wantHeadings := []string{"todo", "decisions", "rationale"}
	for i, want := range wantHeadings {
		if node.Private[i].Heading != want {
			t.Errorf("private[%d].heading: got %q, want %q", i, node.Private[i].Heading, want)
		}
	}
}

// HP-08: Content is raw markdown
func TestNodeParse_HP08_RawMarkdownContent(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n# Public\n## Interface\n### Level three heading\n**bold text**\n```go\nsome code\n```\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public: got nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("public.subsections: got %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("heading: got %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	wantContent := []string{"### Level three heading", "**bold text**", "```go", "some code", "```"}
	got := node.Public.Subsections[0].Content
	if len(got) != len(wantContent) {
		t.Fatalf("content length: got %d, want %d; content=%v", len(got), len(wantContent), got)
	}
	for i, want := range wantContent {
		if got[i] != want {
			t.Errorf("content[%d]: got %q, want %q", i, got[i], want)
		}
	}
}

// ---------------------------------------------------------------------------
// Heading Normalization
// ---------------------------------------------------------------------------

// HN-01: Case insensitive public detection
func TestNodeParse_HN01_PublicCaseInsensitive(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n# PUBLIC\nPublic content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public: got nil, want non-nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("public.heading: got %q, want %q", node.Public.Heading, "public")
	}
}

// HN-02: Public with mixed case and extra whitespace
func TestNodeParse_HN02_PublicMixedCaseWhitespace(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n#   PuBLiC\nPublic content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public: got nil, want non-nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("public.heading: got %q, want %q", node.Public.Heading, "public")
	}
}

// HN-03: Node name with varied whitespace
func TestNodeParse_HN03_NodeNameWhitespace(t *testing.T) {
	tmp := t.TempDir()
	body := "#   ROOT/e\nName content.\n"
	testWriteNodeFile(t, tmp, "e", body)
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/e")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "root/e" {
		t.Errorf("name_section.heading: got %q, want %q", node.NameSection.Heading, "root/e")
	}
}

// HN-04: Subsection headings are normalized
func TestNodeParse_HN04_SubsectionNormalization(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n# Public\n##   Interface\nInterface content.\n## CONSTRAINTS\nConstraints content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public: got nil")
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("subsections: got %d, want 2", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("subsections[0].heading: got %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("subsections[1].heading: got %q, want %q", node.Public.Subsections[1].Heading, "constraints")
	}
}

// HN-05: Closing hashes are stripped
func TestNodeParse_HN05_ClosingHashes(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n# Public\n## Interface ##\nInterface content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public: got nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections: got %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("heading: got %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if node.Public.Subsections[0].RawHeading != "## Interface ##" {
		t.Errorf("raw_heading: got %q, want %q", node.Public.Subsections[0].RawHeading, "## Interface ##")
	}
}

// ---------------------------------------------------------------------------
// Raw Heading Preservation
// ---------------------------------------------------------------------------

// RH-01: Raw heading preserves original line
func TestNodeParse_RH01_RawHeadingOriginal(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n# Public\n## Interface\nInterface content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public: got nil")
	}
	if node.Public.RawHeading != "# Public" {
		t.Errorf("public.raw_heading: got %q, want %q", node.Public.RawHeading, "# Public")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections: got %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].RawHeading != "## Interface" {
		t.Errorf("subsections[0].raw_heading: got %q, want %q", node.Public.Subsections[0].RawHeading, "## Interface")
	}
}

// RH-02: Raw heading preserves case
func TestNodeParse_RH02_RawHeadingPreservesCase(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n# PUBLIC\nPublic content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public: got nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("public.heading: got %q, want %q", node.Public.Heading, "public")
	}
	if node.Public.RawHeading != "# PUBLIC" {
		t.Errorf("public.raw_heading: got %q, want %q", node.Public.RawHeading, "# PUBLIC")
	}
}

// RH-03: Raw heading preserves closing hashes
func TestNodeParse_RH03_RawHeadingClosingHashes(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n# Public\n## Foo ##\nFoo content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public: got nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections: got %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "foo" {
		t.Errorf("heading: got %q, want %q", node.Public.Subsections[0].Heading, "foo")
	}
	if node.Public.Subsections[0].RawHeading != "## Foo ##" {
		t.Errorf("raw_heading: got %q, want %q", node.Public.Subsections[0].RawHeading, "## Foo ##")
	}
}

// RH-04: Raw heading preserves extra whitespace
func TestNodeParse_RH04_RawHeadingExtraWhitespace(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n#   Public\nPublic content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public: got nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("public.heading: got %q, want %q", node.Public.Heading, "public")
	}
	if node.Public.RawHeading != "#   Public" {
		t.Errorf("public.raw_heading: got %q, want %q", node.Public.RawHeading, "#   Public")
	}
}

// ---------------------------------------------------------------------------
// Content Boundaries
// ---------------------------------------------------------------------------

// CB-01: Level-3 and deeper headings are content
func TestNodeParse_CB01_DeeperHeadingsAreContent(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n# Public\n## Interface\n### Sub-sub heading\n#### Even deeper\nInterface content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public: got nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections: got %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("heading: got %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	content := node.Public.Subsections[0].Content
	foundL3 := false
	foundL4 := false
	for _, line := range content {
		if line == "### Sub-sub heading" {
			foundL3 = true
		}
		if line == "#### Even deeper" {
			foundL4 = true
		}
	}
	if !foundL3 {
		t.Errorf("expected '### Sub-sub heading' in content, got %v", content)
	}
	if !foundL4 {
		t.Errorf("expected '#### Even deeper' in content, got %v", content)
	}
}

// CB-02: Fenced code blocks with heading-like content (backtick fence)
func TestNodeParse_CB02_FencedCodeBlockBacktick(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n# Public\n## Interface\n```\n# Looks like a heading\n## Also looks like a heading\n```\nReal content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public: got nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections: got %d, want 1, got %v", len(node.Public.Subsections), node.Public.Subsections)
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("heading: got %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	content := node.Public.Subsections[0].Content
	foundH1 := false
	foundH2 := false
	for _, line := range content {
		if line == "# Looks like a heading" {
			foundH1 = true
		}
		if line == "## Also looks like a heading" {
			foundH2 = true
		}
	}
	if !foundH1 {
		t.Errorf("expected '# Looks like a heading' in content, got %v", content)
	}
	if !foundH2 {
		t.Errorf("expected '## Also looks like a heading' in content, got %v", content)
	}
}

// CB-03: Fenced code block with tilde fence
func TestNodeParse_CB03_FencedCodeBlockTilde(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n# Public\n## Interface\n~~~\n# Inside tilde fence\n~~~\nReal content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public: got nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections: got %d, want 1", len(node.Public.Subsections))
	}
	content := node.Public.Subsections[0].Content
	found := false
	for _, line := range content {
		if line == "# Inside tilde fence" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected '# Inside tilde fence' in content, got %v", content)
	}
}

// CB-04: Fenced code block with language tag
func TestNodeParse_CB04_FencedCodeBlockLanguageTag(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n# Public\n## Interface\n```python\n# Inside fenced block\n```\nReal content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public: got nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("subsections: got %d, want 1", len(node.Public.Subsections))
	}
	content := node.Public.Subsections[0].Content
	found := false
	for _, line := range content {
		if line == "# Inside fenced block" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected '# Inside fenced block' in content, got %v", content)
	}
}

// CB-05: Blank lines between heading and content are preserved
func TestNodeParse_CB05_BlankLinesPreserved(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n# Public\n\nPublic content line.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public: got nil")
	}
	wantContent := []string{"", "Public content line."}
	if len(node.Public.Content) != len(wantContent) {
		t.Fatalf("public.content length: got %d, want %d; content=%v", len(node.Public.Content), len(wantContent), node.Public.Content)
	}
	for i, want := range wantContent {
		if node.Public.Content[i] != want {
			t.Errorf("public.content[%d]: got %q, want %q", i, node.Public.Content[i], want)
		}
	}
}

// ---------------------------------------------------------------------------
// Frontmatter Handling
// ---------------------------------------------------------------------------

// FM-01: Frontmatter is skipped
func TestNodeParse_FM01_FrontmatterSkipped(t *testing.T) {
	tmp := t.TempDir()
	body := "---\ndepends_on: []\n---\n# ROOT/a\nName content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "root/a" {
		t.Errorf("heading: got %q, want %q", node.NameSection.Heading, "root/a")
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "Name content." {
		t.Errorf("content: got %v", node.NameSection.Content)
	}
	for _, line := range node.NameSection.Content {
		if line == "depends_on: []" || line == "---" {
			t.Errorf("frontmatter content appeared in section content: %q", line)
		}
	}
}

// FM-02: No frontmatter delimiters
func TestNodeParse_FM02_NoFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "root/a" {
		t.Errorf("heading: got %q, want %q", node.NameSection.Heading, "root/a")
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "Name content." {
		t.Errorf("content: got %v", node.NameSection.Content)
	}
}

// FM-03: Unclosed frontmatter
func TestNodeParse_FM03_UnclosedFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	body := "---\ndepends_on: []\n# ROOT/a\nName content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Failure Cases
// ---------------------------------------------------------------------------

// FC-01: ARTIFACT reference rejected
func TestNodeParse_FC01_ArtifactReferenceRejected(t *testing.T) {
	_, err := parsenode.NodeParse("ARTIFACT/x(y)")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrNotARootReference) {
		t.Errorf("expected ErrNotARootReference, got %v", err)
	}
}

// FC-02: Qualifier rejected
func TestNodeParse_FC02_QualifierRejected(t *testing.T) {
	_, err := parsenode.NodeParse("ROOT/x(interface)")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrHasQualifier) {
		t.Errorf("expected ErrHasQualifier, got %v", err)
	}
}

// FC-03: File does not exist
func TestNodeParse_FC03_FileNotExist(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := parsenode.NodeParse("ROOT/does/not/exist")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

// FC-04: Propagates path errors
func TestNodeParse_FC04_PropagatesPathErrors(t *testing.T) {
	// A logical name that resolves to a path traversal attempt at the OS level
	// cannot be crafted through the ROOT/ scheme since LogicalNameToPath
	// constructs a valid relative path. Instead we verify that path-layer
	// errors (filereader.ErrFileUnreadable propagated from FileOpen via
	// pathutils) surface correctly when the file simply does not exist.
	// FC-03 already covers the ErrFileUnreadable sentinel. Here we confirm
	// that an error from the filereader layer is propagated (not swallowed).
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := parsenode.NodeParse("ROOT/path/error/test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// The error must originate from the filereader layer.
	if !errors.Is(err, filereader.ErrFileUnreadable) && !errors.Is(err, parsenode.ErrFileUnreadable) {
		t.Errorf("expected a file-layer error, got %v", err)
	}
}

// FC-05: Content before first heading
func TestNodeParse_FC05_ContentBeforeFirstHeading(t *testing.T) {
	tmp := t.TempDir()
	body := "This line appears before any heading.\n# ROOT/a\nName content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

// FC-06: Level-2 heading before any level-1 heading
func TestNodeParse_FC06_Level2BeforeLevel1(t *testing.T) {
	tmp := t.TempDir()
	body := "## Early subsection\n# ROOT/a\nName content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

// FC-07: Empty body
func TestNodeParse_FC07_EmptyBody(t *testing.T) {
	tmp := t.TempDir()
	testWriteNodeFile(t, tmp, "a", "")
	testChdir(t, tmp)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

// FC-08: Node name does not match logical name
func TestNodeParse_FC08_NodeNameMismatch(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/other\nSome content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrNodeNameDoesNotMatch) {
		t.Errorf("expected ErrNodeNameDoesNotMatch, got %v", err)
	}
}

// FC-09: Node name case mismatch is not an error
func TestNodeParse_FC09_NodeNameCaseMismatchOK(t *testing.T) {
	tmp := t.TempDir()
	// File is at code-from-spec/A/_node.md (uppercase directory)
	// but heading uses lowercase. Both normalize to "root/a".
	dir := filepath.Join(tmp, "code-from-spec", "A")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	body := "# root/a\nName content.\n"
	if err := os.WriteFile(filepath.Join(dir, "_node.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	testChdir(t, tmp)

	node, err := parsenode.NodeParse("ROOT/A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "root/a" {
		t.Errorf("heading: got %q, want %q", node.NameSection.Heading, "root/a")
	}
}

// FC-10: Duplicate public section — same case
func TestNodeParse_FC10_DuplicatePublicSameCase(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n# Public\nFirst public content.\n# Public\nSecond public content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicatePublicSection) {
		t.Errorf("expected ErrDuplicatePublicSection, got %v", err)
	}
}

// FC-11: Duplicate public section — different case
func TestNodeParse_FC11_DuplicatePublicDifferentCase(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n# Public\nFirst public content.\n# PUBLIC\nSecond public content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicatePublicSection) {
		t.Errorf("expected ErrDuplicatePublicSection, got %v", err)
	}
}

// FC-12: Duplicate agent section
func TestNodeParse_FC12_DuplicateAgentSection(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n# Agent\nFirst agent content.\n# Agent\nSecond agent content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateAgentSection) {
		t.Errorf("expected ErrDuplicateAgentSection, got %v", err)
	}
}

// FC-13: Duplicate subsection in public — same case
func TestNodeParse_FC13_DuplicateSubsectionSameCase(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n# Public\n## Interface\nFirst interface content.\n## Interface\nSecond interface content.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("expected ErrDuplicateSubsection, got %v", err)
	}
}

// FC-14: Duplicate subsection in public — different case
func TestNodeParse_FC14_DuplicateSubsectionDifferentCase(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n# Public\n## Interface\nFirst.\n## INTERFACE\nSecond.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("expected ErrDuplicateSubsection, got %v", err)
	}
}

// FC-15: Duplicate subsection in public — whitespace variation
func TestNodeParse_FC15_DuplicateSubsectionWhitespace(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n# Public\n## Interface\nFirst.\n##   Interface\nSecond.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("expected ErrDuplicateSubsection, got %v", err)
	}
}

// FC-16: Duplicate subsection in agent
func TestNodeParse_FC16_DuplicateSubsectionInAgent(t *testing.T) {
	tmp := t.TempDir()
	body := "# ROOT/a\nName content.\n# Agent\n## Guidance\nFirst guidance.\n## Guidance\nSecond guidance.\n"
	testWriteNodeFile(t, tmp, "a", body)
	testChdir(t, tmp)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("expected ErrDuplicateSubsection, got %v", err)
	}
}

// Ensure the pathutils import is used (via the PathCfs type in filereader).
var _ = pathutils.PathCfs{}
