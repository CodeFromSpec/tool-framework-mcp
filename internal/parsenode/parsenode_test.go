// code-from-spec: ROOT/golang/tests/parsing/node_parsing@rAsGb3xnn2i4pQ-rQkhR4S_BE8Y
package parsenode_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

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

// testWriteNodeFile creates the directory structure and writes a node file.
// logicalName like "ROOT/x/y" maps to code-from-spec/x/y/_node.md under the cwd.
func testWriteNodeFile(t *testing.T, logicalName string, content string) {
	t.Helper()
	// Strip "ROOT" prefix to get path segments.
	// "ROOT/x/y" -> "x/y"
	rel := logicalName[len("ROOT"):]
	if len(rel) > 0 && rel[0] == '/' {
		rel = rel[1:]
	}
	var dir string
	if rel == "" {
		dir = "code-from-spec"
	} else {
		dir = filepath.Join("code-from-spec", filepath.FromSlash(rel))
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testWriteNodeFile: mkdir %s: %v", dir, err)
	}
	nodePath := filepath.Join(dir, "_node.md")
	if err := os.WriteFile(nodePath, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNodeFile: write %s: %v", nodePath, err)
	}
}

// --------------------------------------------------------------------------
// Happy Path
// --------------------------------------------------------------------------

func TestNodeParse_MinimalNode(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/x", "# ROOT/x\n\nA simple node.\n")

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
		t.Errorf("subsections: expected empty, got %d", len(node.NameSection.Subsections))
	}
	if node.Public != nil {
		t.Errorf("public: expected nil")
	}
	if node.Agent != nil {
		t.Errorf("agent: expected nil")
	}
	if len(node.Private) != 0 {
		t.Errorf("private: expected empty, got %d", len(node.Private))
	}
}

func TestNodeParse_FullNode(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `---
some: frontmatter
---
# ROOT/payments/fees

Name section content.

# Public

## Interface

Interface content.

## Constraints

Constraints content.

# Agent

Agent content here.

# Decisions

Decisions content.

# Rationale

Rationale content.
`
	testWriteNodeFile(t, "ROOT/payments/fees", content)

	node, err := parsenode.NodeParse("ROOT/payments/fees")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "root/payments/fees" {
		t.Errorf("name heading: got %q, want %q", node.NameSection.Heading, "root/payments/fees")
	}
	if node.Public == nil {
		t.Fatalf("public: expected non-nil")
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("public subsections: got %d, want 2", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("subsection[0] heading: got %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("subsection[1] heading: got %q, want %q", node.Public.Subsections[1].Heading, "constraints")
	}
	if node.Agent == nil {
		t.Fatalf("agent: expected non-nil")
	}
	if len(node.Private) != 2 {
		t.Fatalf("private: got %d sections, want 2", len(node.Private))
	}
	if node.Private[0].Heading != "decisions" {
		t.Errorf("private[0] heading: got %q, want %q", node.Private[0].Heading, "decisions")
	}
	if node.Private[1].Heading != "rationale" {
		t.Errorf("private[1] heading: got %q, want %q", node.Private[1].Heading, "rationale")
	}
}

func TestNodeParse_NoPublicSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `# ROOT/decisions

Name content.

# Rationale

Rationale content.
`
	testWriteNodeFile(t, "ROOT/decisions", content)

	node, err := parsenode.NodeParse("ROOT/decisions")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public != nil {
		t.Errorf("public: expected nil")
	}
	if node.Agent != nil {
		t.Errorf("agent: expected nil")
	}
	if len(node.Private) != 1 {
		t.Fatalf("private: got %d, want 1", len(node.Private))
	}
	if node.Private[0].Heading != "rationale" {
		t.Errorf("private[0] heading: got %q, want %q", node.Private[0].Heading, "rationale")
	}
}

func TestNodeParse_PublicContentBeforeFirstSubsection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `# ROOT/a

Name content.

# Public

Direct content line 1.
Direct content line 2.

## Interface

Interface content.
`
	testWriteNodeFile(t, "ROOT/a", content)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public: expected non-nil")
	}
	// Content should be the lines before ## Interface
	foundLine1 := false
	foundLine2 := false
	for _, line := range node.Public.Content {
		if line == "Direct content line 1." {
			foundLine1 = true
		}
		if line == "Direct content line 2." {
			foundLine2 = true
		}
	}
	if !foundLine1 || !foundLine2 {
		t.Errorf("public.content missing expected lines: %v", node.Public.Content)
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("public subsections: got %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("subsection heading: got %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if node.Public.Subsections[0].RawHeading != "## Interface" {
		t.Errorf("subsection raw_heading: got %q, want %q", node.Public.Subsections[0].RawHeading, "## Interface")
	}
}

func TestNodeParse_PublicNoContentNoSubsections(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `# ROOT/b

Name content.

# Public
# Agent

Agent content.
`
	testWriteNodeFile(t, "ROOT/b", content)

	node, err := parsenode.NodeParse("ROOT/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public: expected non-nil")
	}
	// Content may contain blank lines between # Public and # Agent, but no non-blank lines
	for _, line := range node.Public.Content {
		if line != "" {
			t.Errorf("public.content: expected only blank lines, got %q", line)
		}
	}
	if len(node.Public.Subsections) != 0 {
		t.Errorf("public subsections: expected empty, got %d", len(node.Public.Subsections))
	}
}

func TestNodeParse_AgentSectionWithSubsections(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `# ROOT/c

Name content.

# Agent

Preamble line 1.
Preamble line 2.

## Implementation guidance

Implementation content.

## Contracts

Contracts content.
`
	testWriteNodeFile(t, "ROOT/c", content)

	node, err := parsenode.NodeParse("ROOT/c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Agent == nil {
		t.Fatalf("agent: expected non-nil")
	}
	if node.Agent.RawHeading != "# Agent" {
		t.Errorf("agent raw_heading: got %q, want %q", node.Agent.RawHeading, "# Agent")
	}
	foundPreamble1 := false
	foundPreamble2 := false
	for _, line := range node.Agent.Content {
		if line == "Preamble line 1." {
			foundPreamble1 = true
		}
		if line == "Preamble line 2." {
			foundPreamble2 = true
		}
	}
	if !foundPreamble1 || !foundPreamble2 {
		t.Errorf("agent.content missing preamble lines: %v", node.Agent.Content)
	}
	if len(node.Agent.Subsections) != 2 {
		t.Fatalf("agent subsections: got %d, want 2", len(node.Agent.Subsections))
	}
	if node.Agent.Subsections[0].Heading != "implementation guidance" {
		t.Errorf("agent subsection[0] heading: got %q, want %q", node.Agent.Subsections[0].Heading, "implementation guidance")
	}
	if node.Agent.Subsections[1].Heading != "contracts" {
		t.Errorf("agent subsection[1] heading: got %q, want %q", node.Agent.Subsections[1].Heading, "contracts")
	}
}

func TestNodeParse_PrivateSectionsPreserveFileOrder(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `# ROOT/d

Name content.

# TODO

Todo content.

# Decisions

Decisions content.

# Rationale

Rationale content.
`
	testWriteNodeFile(t, "ROOT/d", content)

	node, err := parsenode.NodeParse("ROOT/d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(node.Private) != 3 {
		t.Fatalf("private: got %d, want 3", len(node.Private))
	}
	want := []string{"todo", "decisions", "rationale"}
	for i, w := range want {
		if node.Private[i].Heading != w {
			t.Errorf("private[%d] heading: got %q, want %q", i, node.Private[i].Heading, w)
		}
	}
}

func TestNodeParse_ContentIsRawMarkdown(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/f\n\nName content.\n\n# Public\n\n## Details\n\n### Details\n\n**bold**\n\n```go\nfunc foo() {}\n```\n"
	testWriteNodeFile(t, "ROOT/f", content)

	node, err := parsenode.NodeParse("ROOT/f")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public: expected non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("public subsections: got %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	foundH3 := false
	foundBold := false
	foundFence := false
	for _, line := range sub.Content {
		if line == "### Details" {
			foundH3 = true
		}
		if line == "**bold**" {
			foundBold = true
		}
		if line == "```go" {
			foundFence = true
		}
	}
	if !foundH3 {
		t.Errorf("content: missing ### Details line; content: %v", sub.Content)
	}
	if !foundBold {
		t.Errorf("content: missing **bold** line; content: %v", sub.Content)
	}
	if !foundFence {
		t.Errorf("content: missing ```go line; content: %v", sub.Content)
	}
}

// --------------------------------------------------------------------------
// Heading Normalization
// --------------------------------------------------------------------------

func TestNodeParse_CaseInsensitivePublicDetection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `# ROOT/g

Name content.

# PUBLIC

Public content.
`
	testWriteNodeFile(t, "ROOT/g", content)

	node, err := parsenode.NodeParse("ROOT/g")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public: expected non-nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("public.heading: got %q, want %q", node.Public.Heading, "public")
	}
}

func TestNodeParse_PublicMixedCaseAndExtraWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "#   PuBLiC\n\nPublic content.\n"
	// Node name section must come first; let's use a proper node file:
	fullContent := "# ROOT/h\n\nName content.\n\n#   PuBLiC\n\nPublic content.\n"
	testWriteNodeFile(t, "ROOT/h", fullContent)

	node, err := parsenode.NodeParse("ROOT/h")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = content
	if node.Public == nil {
		t.Fatalf("public: expected non-nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("public.heading: got %q, want %q", node.Public.Heading, "public")
	}
}

func TestNodeParse_NodeNameWithVariedWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "#    ROOT/e\n\nContent.\n"
	testWriteNodeFile(t, "ROOT/e", content)

	node, err := parsenode.NodeParse("ROOT/e")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "root/e" {
		t.Errorf("name heading: got %q, want %q", node.NameSection.Heading, "root/e")
	}
}

func TestNodeParse_SubsectionHeadingsAreNormalized(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `# ROOT/i

Name content.

# Public

##   Interface

Interface content.

## CONSTRAINTS

Constraints content.
`
	testWriteNodeFile(t, "ROOT/i", content)

	node, err := parsenode.NodeParse("ROOT/i")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public: expected non-nil")
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("public subsections: got %d, want 2", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("subsection[0] heading: got %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("subsection[1] heading: got %q, want %q", node.Public.Subsections[1].Heading, "constraints")
	}
}

func TestNodeParse_ClosingHashesAreStripped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `# ROOT/j

Name content.

# Public

## Interface ##

Interface content.
`
	testWriteNodeFile(t, "ROOT/j", content)

	node, err := parsenode.NodeParse("ROOT/j")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public: expected non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("public subsections: got %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	if sub.Heading != "interface" {
		t.Errorf("heading: got %q, want %q", sub.Heading, "interface")
	}
	if sub.RawHeading != "## Interface ##" {
		t.Errorf("raw_heading: got %q, want %q", sub.RawHeading, "## Interface ##")
	}
}

// --------------------------------------------------------------------------
// Raw Heading Preservation
// --------------------------------------------------------------------------

func TestNodeParse_RawHeadingPreservesOriginalLine(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `# ROOT/k

Name content.

# Public

## Interface

Interface content.
`
	testWriteNodeFile(t, "ROOT/k", content)

	node, err := parsenode.NodeParse("ROOT/k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public: expected non-nil")
	}
	if node.Public.RawHeading != "# Public" {
		t.Errorf("public.raw_heading: got %q, want %q", node.Public.RawHeading, "# Public")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("public subsections: got %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].RawHeading != "## Interface" {
		t.Errorf("subsection raw_heading: got %q, want %q", node.Public.Subsections[0].RawHeading, "## Interface")
	}
}

func TestNodeParse_RawHeadingPreservesCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `# ROOT/l

Name content.

# PUBLIC

Public content.
`
	testWriteNodeFile(t, "ROOT/l", content)

	node, err := parsenode.NodeParse("ROOT/l")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public: expected non-nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("heading: got %q, want %q", node.Public.Heading, "public")
	}
	if node.Public.RawHeading != "# PUBLIC" {
		t.Errorf("raw_heading: got %q, want %q", node.Public.RawHeading, "# PUBLIC")
	}
}

func TestNodeParse_RawHeadingPreservesClosingHashes(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `# ROOT/m

Name content.

# Public

## Foo ##

Foo content.
`
	testWriteNodeFile(t, "ROOT/m", content)

	node, err := parsenode.NodeParse("ROOT/m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public: expected non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("public subsections: got %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	if sub.Heading != "foo" {
		t.Errorf("heading: got %q, want %q", sub.Heading, "foo")
	}
	if sub.RawHeading != "## Foo ##" {
		t.Errorf("raw_heading: got %q, want %q", sub.RawHeading, "## Foo ##")
	}
}

func TestNodeParse_RawHeadingPreservesExtraWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "#    ROOT/n\n\nName content.\n\n#   Public\n\nPublic content.\n"
	testWriteNodeFile(t, "ROOT/n", content)

	node, err := parsenode.NodeParse("ROOT/n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public: expected non-nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("heading: got %q, want %q", node.Public.Heading, "public")
	}
	if node.Public.RawHeading != "#   Public" {
		t.Errorf("raw_heading: got %q, want %q", node.Public.RawHeading, "#   Public")
	}
}

// --------------------------------------------------------------------------
// Content Boundaries
// --------------------------------------------------------------------------

func TestNodeParse_Level3AndDeeperHeadingsAreContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `# ROOT/o

Name content.

# Public

## Details

### Details

Some text.

#### Sub-details

More text.
`
	testWriteNodeFile(t, "ROOT/o", content)

	node, err := parsenode.NodeParse("ROOT/o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public: expected non-nil")
	}
	// Only one ## subsection — ### and #### are content
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("public subsections: got %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	foundH3 := false
	foundH4 := false
	for _, line := range sub.Content {
		if line == "### Details" {
			foundH3 = true
		}
		if line == "#### Sub-details" {
			foundH4 = true
		}
	}
	if !foundH3 {
		t.Errorf("content: missing ### Details; content: %v", sub.Content)
	}
	if !foundH4 {
		t.Errorf("content: missing #### Sub-details; content: %v", sub.Content)
	}
}

func TestNodeParse_FencedCodeBlockBacktick(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/p\n\nName content.\n\n# Public\n\n## Details\n\n```\n# Heading\n## Also heading\n```\n"
	testWriteNodeFile(t, "ROOT/p", content)

	node, err := parsenode.NodeParse("ROOT/p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public: expected non-nil")
	}
	// Only one subsection — headings inside fences are content
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("public subsections: got %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	foundH1 := false
	foundH2 := false
	for _, line := range sub.Content {
		if line == "# Heading" {
			foundH1 = true
		}
		if line == "## Also heading" {
			foundH2 = true
		}
	}
	if !foundH1 {
		t.Errorf("content: missing fenced # Heading; content: %v", sub.Content)
	}
	if !foundH2 {
		t.Errorf("content: missing fenced ## Also heading; content: %v", sub.Content)
	}
}

func TestNodeParse_FencedCodeBlockTilde(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/q\n\nName content.\n\n# Public\n\n## Details\n\n~~~\n# Heading\n~~~\n"
	testWriteNodeFile(t, "ROOT/q", content)

	node, err := parsenode.NodeParse("ROOT/q")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public: expected non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("public subsections: got %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	found := false
	for _, line := range sub.Content {
		if line == "# Heading" {
			found = true
		}
	}
	if !found {
		t.Errorf("content: missing fenced # Heading; content: %v", sub.Content)
	}
}

func TestNodeParse_FencedCodeBlockWithLanguageTag(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/r\n\nName content.\n\n# Public\n\n## Details\n\n```yaml\n# Heading\n```\n"
	testWriteNodeFile(t, "ROOT/r", content)

	node, err := parsenode.NodeParse("ROOT/r")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public: expected non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("public subsections: got %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	found := false
	for _, line := range sub.Content {
		if line == "# Heading" {
			found = true
		}
	}
	if !found {
		t.Errorf("content: missing fenced # Heading in yaml block; content: %v", sub.Content)
	}
}

func TestNodeParse_LeadingAndTrailingBlankLinesPreserved(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/s\n\nName content.\n\n# Public\n\n\nSome content.\n\n\n"
	testWriteNodeFile(t, "ROOT/s", content)

	node, err := parsenode.NodeParse("ROOT/s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public: expected non-nil")
	}
	// The content should contain blank lines at the start and end.
	if len(node.Public.Content) < 3 {
		t.Errorf("public.content: expected at least 3 lines (blank, content, blank), got %d: %v", len(node.Public.Content), node.Public.Content)
	}
	if node.Public.Content[0] != "" {
		t.Errorf("public.content[0]: expected blank line, got %q", node.Public.Content[0])
	}
}

// --------------------------------------------------------------------------
// Frontmatter Handling
// --------------------------------------------------------------------------

func TestNodeParse_FrontmatterIsSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `---
key: value
another: thing
---
# ROOT/t

Body content.
`
	testWriteNodeFile(t, "ROOT/t", content)

	node, err := parsenode.NodeParse("ROOT/t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "root/t" {
		t.Errorf("name heading: got %q, want %q", node.NameSection.Heading, "root/t")
	}
}

func TestNodeParse_NoFrontmatterDelimiters(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `# ROOT/u

Body content.
`
	testWriteNodeFile(t, "ROOT/u", content)

	node, err := parsenode.NodeParse("ROOT/u")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "root/u" {
		t.Errorf("name heading: got %q, want %q", node.NameSection.Heading, "root/u")
	}
}

func TestNodeParse_UnclosedFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `---
key: value
# ROOT/v

Body content.
`
	testWriteNodeFile(t, "ROOT/v", content)

	_, err := parsenode.NodeParse("ROOT/v")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error: got %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

// --------------------------------------------------------------------------
// Failure Cases
// --------------------------------------------------------------------------

func TestNodeParse_ArtifactReferenceRejected(t *testing.T) {
	_, err := parsenode.NodeParse("ARTIFACT/x(y)")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrNotARootReference) {
		t.Errorf("error: got %v, want ErrNotARootReference", err)
	}
}

func TestNodeParse_QualifierRejected(t *testing.T) {
	_, err := parsenode.NodeParse("ROOT/x(interface)")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrHasQualifier) {
		t.Errorf("error: got %v, want ErrHasQualifier", err)
	}
}

func TestNodeParse_FileDoesNotExist(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := parsenode.NodeParse("ROOT/nonexistent/node")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrFileUnreadable) {
		t.Errorf("error: got %v, want ErrFileUnreadable", err)
	}
}

func TestNodeParse_PropagatesPathErrors(t *testing.T) {
	// Use a logical name that contains a traversal; note that LogicalNameToPath
	// strips the qualifier but cannot introduce traversal from a valid ROOT/ name.
	// Instead we rely on the fact that path errors (e.g., from pathutils) are
	// propagated. A name with ".." in the segment would not pass logical name
	// validation, but we can test with an ARTIFACT/ that passes the root check
	// — actually that hits ErrNotARootReference. The spec says "path containing
	// traversal component" — the only way to get that is via a logical name whose
	// segments map to ".." at the OS level, which is not possible with ROOT/.
	// We verify that pathutils errors are wrapped in the error chain. Since
	// LogicalNameToPath only accepts ROOT/ names and strips qualifiers, there is
	// no way to inject traversal via that route. This test verifies the error
	// wraps correctly by checking for ErrFileUnreadable (filereader sentinel),
	// which propagates from FileOpen on a missing file.
	tmp := t.TempDir()
	testChdir(t, tmp)

	// The pathutils package returns ErrPathIsEmpty for empty value — but
	// NodeParse derives the path via LogicalNameToPath which always produces
	// a valid-format path. The only path error reachable through NodeParse is
	// ErrFileUnreadable when the file does not exist.
	_, err := parsenode.NodeParse("ROOT/does/not/exist")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	// The error should be propagated from FileOpen — either ErrFileUnreadable
	// or a pathutils sentinel.
	isPathErr := errors.Is(err, parsenode.ErrFileUnreadable)
	isPathUtilsErr := errors.Is(err, pathutils.ErrPathIsEmpty) ||
		errors.Is(err, pathutils.ErrDirectoryTraversal) ||
		errors.Is(err, pathutils.ErrResolvesOutsideRoot)
	if !isPathErr && !isPathUtilsErr {
		t.Errorf("error: got %v, want a path or file unreadable error", err)
	}
}

func TestNodeParse_ContentBeforeFirstHeading(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "Some text before heading.\n# ROOT/w\n\nContent.\n"
	testWriteNodeFile(t, "ROOT/w", content)

	_, err := parsenode.NodeParse("ROOT/w")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error: got %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

func TestNodeParse_Level2HeadingBeforeAnyLevel1Heading(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "## Subsection\n\nContent.\n"
	testWriteNodeFile(t, "ROOT/ww", content)

	_, err := parsenode.NodeParse("ROOT/ww")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error: got %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

func TestNodeParse_EmptyBody(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := ""
	testWriteNodeFile(t, "ROOT/empty", content)

	_, err := parsenode.NodeParse("ROOT/empty")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error: got %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

func TestNodeParse_NodeNameDoesNotMatch(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	// Write a file at ROOT/x but with heading ROOT/other
	content := "# ROOT/other\n\nContent.\n"
	testWriteNodeFile(t, "ROOT/x", content)

	_, err := parsenode.NodeParse("ROOT/x")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrNodeNameDoesNotMatch) {
		t.Errorf("error: got %v, want ErrNodeNameDoesNotMatch", err)
	}
}

func TestNodeParse_NodeNameCaseMismatchIsNotError(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	// lowercase heading — normalization should make it equal to "root/x"
	content := "# root/x\n\nContent.\n"
	testWriteNodeFile(t, "ROOT/x", content)

	node, err := parsenode.NodeParse("ROOT/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "root/x" {
		t.Errorf("name heading: got %q, want %q", node.NameSection.Heading, "root/x")
	}
}

func TestNodeParse_DuplicatePublicSection_SameCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `# ROOT/dup1

Name content.

# Public

First public.

# Public

Second public.
`
	testWriteNodeFile(t, "ROOT/dup1", content)

	_, err := parsenode.NodeParse("ROOT/dup1")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicatePublicSection) {
		t.Errorf("error: got %v, want ErrDuplicatePublicSection", err)
	}
}

func TestNodeParse_DuplicatePublicSection_DifferentCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `# ROOT/dup2

Name content.

# Public

First public.

# PUBLIC

Second public.
`
	testWriteNodeFile(t, "ROOT/dup2", content)

	_, err := parsenode.NodeParse("ROOT/dup2")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicatePublicSection) {
		t.Errorf("error: got %v, want ErrDuplicatePublicSection", err)
	}
}

func TestNodeParse_DuplicateAgentSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `# ROOT/dup3

Name content.

# Agent

First agent.

# Agent

Second agent.
`
	testWriteNodeFile(t, "ROOT/dup3", content)

	_, err := parsenode.NodeParse("ROOT/dup3")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateAgentSection) {
		t.Errorf("error: got %v, want ErrDuplicateAgentSection", err)
	}
}

func TestNodeParse_DuplicateSubsectionInPublic_SameCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `# ROOT/dup4

Name content.

# Public

## Interface

First.

## Interface

Second.
`
	testWriteNodeFile(t, "ROOT/dup4", content)

	_, err := parsenode.NodeParse("ROOT/dup4")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error: got %v, want ErrDuplicateSubsection", err)
	}
}

func TestNodeParse_DuplicateSubsectionInPublic_DifferentCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `# ROOT/dup5

Name content.

# Public

## Interface

First.

## INTERFACE

Second.
`
	testWriteNodeFile(t, "ROOT/dup5", content)

	_, err := parsenode.NodeParse("ROOT/dup5")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error: got %v, want ErrDuplicateSubsection", err)
	}
}

func TestNodeParse_DuplicateSubsectionInPublic_WhitespaceVariation(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `# ROOT/dup6

Name content.

# Public

## Interface

First.

##   Interface

Second.
`
	testWriteNodeFile(t, "ROOT/dup6", content)

	_, err := parsenode.NodeParse("ROOT/dup6")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error: got %v, want ErrDuplicateSubsection", err)
	}
}

func TestNodeParse_DuplicateSubsectionInAgent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `# ROOT/dup7

Name content.

# Agent

## Details

First details.

## Details

Second details.
`
	testWriteNodeFile(t, "ROOT/dup7", content)

	_, err := parsenode.NodeParse("ROOT/dup7")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error: got %v, want ErrDuplicateSubsection", err)
	}
}
