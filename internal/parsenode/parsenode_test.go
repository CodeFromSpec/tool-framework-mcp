// code-from-spec: ROOT/golang/tests/parsing/node_parsing@uyAYjszyhCnfvO89lbl9qP53Qpk

package parsenode_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
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

// testWriteNodeFile creates the directory structure and writes content to the
// _node.md file corresponding to a logical name under the given base dir.
// The logical name must start with "ROOT/".
func testWriteNodeFile(t *testing.T, logicalName string, content string) {
	t.Helper()
	// Strip "ROOT/" prefix and map to code-from-spec/<rest>/_node.md
	rest := logicalName[len("ROOT/"):]
	dir := filepath.Join("code-from-spec", filepath.FromSlash(rest))
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testWriteNodeFile: mkdir %q: %v", dir, err)
	}
	path := filepath.Join(dir, "_node.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNodeFile: write %q: %v", path, err)
	}
}

// ---------------------------------------------------------------------------
// Happy Path
// ---------------------------------------------------------------------------

func TestNodeParse_MinimalNodeNameSectionOnly(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/x", "# ROOT/x\n\nA simple node.\n")

	node, err := parsenode.NodeParse("ROOT/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "root/x" {
		t.Errorf("name_section.heading = %q; want %q", node.NameSection.Heading, "root/x")
	}
	if node.NameSection.Content != "A simple node." {
		t.Errorf("name_section.content = %q; want %q", node.NameSection.Content, "A simple node.")
	}
	if len(node.NameSection.Subsections) != 0 {
		t.Errorf("name_section.subsections length = %d; want 0", len(node.NameSection.Subsections))
	}
	if node.Public != nil {
		t.Errorf("public should be absent")
	}
	if node.Agent != nil {
		t.Errorf("agent should be absent")
	}
	if len(node.Private) != 0 {
		t.Errorf("private length = %d; want 0", len(node.Private))
	}
}

func TestNodeParse_FullNode(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `---
outputs:
  - id: foo
    path: some/path.go
---

# ROOT/payments/fees

Some name content.

# Public

## Interface

Interface content.

## Constraints

Constraints content.

# Agent

Agent content.

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
		t.Errorf("name_section.heading = %q; want %q", node.NameSection.Heading, "root/payments/fees")
	}
	if node.Public == nil {
		t.Fatalf("public should be present")
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("public.subsections length = %d; want 2", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("public.subsections[0].heading = %q; want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("public.subsections[1].heading = %q; want %q", node.Public.Subsections[1].Heading, "constraints")
	}
	if node.Agent == nil {
		t.Fatalf("agent should be present")
	}
	if node.Agent.Content == "" {
		t.Errorf("agent.content should not be empty")
	}
	if len(node.Private) != 2 {
		t.Fatalf("private length = %d; want 2", len(node.Private))
	}
	if node.Private[0].Heading != "decisions" {
		t.Errorf("private[0].heading = %q; want %q", node.Private[0].Heading, "decisions")
	}
	if node.Private[1].Heading != "rationale" {
		t.Errorf("private[1].heading = %q; want %q", node.Private[1].Heading, "rationale")
	}
}

func TestNodeParse_NodeWithNoPublicSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/decisions\n\nSome content.\n\n# Rationale\n\nRationale content.\n"
	testWriteNodeFile(t, "ROOT/decisions", content)

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
		t.Fatalf("private length = %d; want 1", len(node.Private))
	}
	if node.Private[0].Heading != "rationale" {
		t.Errorf("private[0].heading = %q; want %q", node.Private[0].Heading, "rationale")
	}
}

func TestNodeParse_PublicSectionContentBeforeFirstSubsection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/a\n\n# Public\n\nSome introductory text.\n\n## Interface\n\nInterface content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public should be present")
	}
	if node.Public.Content != "Some introductory text." {
		t.Errorf("public.content = %q; want %q", node.Public.Content, "Some introductory text.")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("public.subsections length = %d; want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("public.subsections[0].heading = %q; want %q", node.Public.Subsections[0].Heading, "interface")
	}
}

func TestNodeParse_PublicSectionEmpty(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/b\n\n# Public\n\n# Agent\n\nAgent content.\n"
	testWriteNodeFile(t, "ROOT/b", content)

	node, err := parsenode.NodeParse("ROOT/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public should be present")
	}
	if node.Public.Content != "" {
		t.Errorf("public.content = %q; want empty", node.Public.Content)
	}
	if len(node.Public.Subsections) != 0 {
		t.Errorf("public.subsections length = %d; want 0", len(node.Public.Subsections))
	}
}

func TestNodeParse_AgentSectionWithSubsections(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `# ROOT/c

# Agent

Agent preamble.

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
		t.Fatalf("agent should be present")
	}
	if node.Agent.Content != "Agent preamble." {
		t.Errorf("agent.content = %q; want %q", node.Agent.Content, "Agent preamble.")
	}
	if len(node.Agent.Subsections) != 2 {
		t.Fatalf("agent.subsections length = %d; want 2", len(node.Agent.Subsections))
	}
	if node.Agent.Subsections[0].Heading != "implementation guidance" {
		t.Errorf("agent.subsections[0].heading = %q; want %q", node.Agent.Subsections[0].Heading, "implementation guidance")
	}
	if node.Agent.Subsections[0].Content == "" {
		t.Errorf("agent.subsections[0].content should not be empty")
	}
	if node.Agent.Subsections[1].Heading != "contracts" {
		t.Errorf("agent.subsections[1].heading = %q; want %q", node.Agent.Subsections[1].Heading, "contracts")
	}
	if node.Agent.Subsections[1].Content == "" {
		t.Errorf("agent.subsections[1].content should not be empty")
	}
}

func TestNodeParse_PrivateSectionsPreserveFileOrder(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `# ROOT/d

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
		t.Fatalf("private length = %d; want 3", len(node.Private))
	}
	if node.Private[0].Heading != "todo" {
		t.Errorf("private[0].heading = %q; want %q", node.Private[0].Heading, "todo")
	}
	if node.Private[1].Heading != "decisions" {
		t.Errorf("private[1].heading = %q; want %q", node.Private[1].Heading, "decisions")
	}
	if node.Private[2].Heading != "rationale" {
		t.Errorf("private[2].heading = %q; want %q", node.Private[2].Heading, "rationale")
	}
}

func TestNodeParse_ContentIsRawMarkdown(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	subsectionContent := "### Details\n\nSome detail text.\n\n**important**\n\n```go\nfunc main() {}\n```"
	content := "# ROOT/f\n\n# Public\n\n## Interface\n\n" + subsectionContent + "\n"
	testWriteNodeFile(t, "ROOT/f", content)

	node, err := parsenode.NodeParse("ROOT/f")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public should be present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("public.subsections length = %d; want 1", len(node.Public.Subsections))
	}
	got := node.Public.Subsections[0].Content
	if got == "" {
		t.Fatalf("subsection content should not be empty")
	}
	// The raw markdown elements must appear in the content.
	for _, want := range []string{"### Details", "**important**", "```go"} {
		found := false
		for i := 0; i <= len(got)-len(want); i++ {
			if got[i:i+len(want)] == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subsection content does not contain %q; got:\n%s", want, got)
		}
	}
}

// ---------------------------------------------------------------------------
// Heading Normalization
// ---------------------------------------------------------------------------

func TestNodeParse_CaseInsensitivePublicDetection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/g\n\n# PUBLIC\n\nPublic content.\n"
	testWriteNodeFile(t, "ROOT/g", content)

	node, err := parsenode.NodeParse("ROOT/g")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public should be present")
	}
	if node.Public.Heading != "public" {
		t.Errorf("public.heading = %q; want %q", node.Public.Heading, "public")
	}
}

func TestNodeParse_PublicWithMixedCaseAndExtraWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/h\n\n#   PuBLiC\n\nPublic content.\n"
	testWriteNodeFile(t, "ROOT/h", content)

	node, err := parsenode.NodeParse("ROOT/h")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public should be present")
	}
	if node.Public.Heading != "public" {
		t.Errorf("public.heading = %q; want %q", node.Public.Heading, "public")
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
		t.Errorf("name_section.heading = %q; want %q", node.NameSection.Heading, "root/e")
	}
}

func TestNodeParse_SubsectionHeadingsAreNormalized(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/i\n\n# Public\n\n##   Interface\n\nInterface content.\n\n## CONSTRAINTS\n\nConstraints content.\n"
	testWriteNodeFile(t, "ROOT/i", content)

	node, err := parsenode.NodeParse("ROOT/i")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public should be present")
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("public.subsections length = %d; want 2", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("public.subsections[0].heading = %q; want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("public.subsections[1].heading = %q; want %q", node.Public.Subsections[1].Heading, "constraints")
	}
}

func TestNodeParse_ClosingHashesAreStripped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/j\n\n# Public\n\n## Interface ##\n\nInterface content.\n"
	testWriteNodeFile(t, "ROOT/j", content)

	node, err := parsenode.NodeParse("ROOT/j")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public should be present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("public.subsections length = %d; want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("subsection heading = %q; want %q", node.Public.Subsections[0].Heading, "interface")
	}
}

// ---------------------------------------------------------------------------
// Content Boundaries
// ---------------------------------------------------------------------------

func TestNodeParse_Level3AndDeeperHeadingsAreContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := `# ROOT/k

# Public

## Interface

### Details

Detail text.

#### Sub-details

Sub-detail text.
`
	testWriteNodeFile(t, "ROOT/k", content)

	node, err := parsenode.NodeParse("ROOT/k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public should be present")
	}
	// Only one subsection — level 3+ do not create new subsections.
	if len(node.Public.Subsections) != 1 {
		t.Errorf("public.subsections length = %d; want 1", len(node.Public.Subsections))
	}
	got := node.Public.Subsections[0].Content
	for _, want := range []string{"### Details", "#### Sub-details"} {
		found := false
		for i := 0; i <= len(got)-len(want); i++ {
			if got[i:i+len(want)] == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("content does not contain %q; got:\n%s", want, got)
		}
	}
}

func TestNodeParse_FencedCodeBlockWithBacktickFenceIgnoresHeadingLikeContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/l\n\n# Public\n\n## Interface\n\nSome text.\n\n```\n# Heading inside fence\n## Subheading inside fence\n```\n\nMore text.\n"
	testWriteNodeFile(t, "ROOT/l", content)

	node, err := parsenode.NodeParse("ROOT/l")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public should be present")
	}
	// Still only one subsection.
	if len(node.Public.Subsections) != 1 {
		t.Errorf("public.subsections length = %d; want 1", len(node.Public.Subsections))
	}
}

func TestNodeParse_FencedCodeBlockWithTildeFence(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/m\n\n# Public\n\n## Interface\n\nSome text.\n\n~~~\n# Heading inside tilde fence\n~~~\n\nMore text.\n"
	testWriteNodeFile(t, "ROOT/m", content)

	node, err := parsenode.NodeParse("ROOT/m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public should be present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Errorf("public.subsections length = %d; want 1", len(node.Public.Subsections))
	}
}

func TestNodeParse_FencedCodeBlockWithLanguageTag(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/n\n\n# Public\n\n## Interface\n\nSome text.\n\n```yaml\n# Heading\nkey: value\n```\n\nMore text.\n"
	testWriteNodeFile(t, "ROOT/n", content)

	node, err := parsenode.NodeParse("ROOT/n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatalf("public should be present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Errorf("public.subsections length = %d; want 1", len(node.Public.Subsections))
	}
}

func TestNodeParse_LeadingAndTrailingBlankLinesTrimmed(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/o\n\n\n\nName content.\n\n\n\n# Public\n\n\n\nPublic content.\n\n\n\n## Interface\n\n\n\nInterface content.\n\n\n\n"
	testWriteNodeFile(t, "ROOT/o", content)

	node, err := parsenode.NodeParse("ROOT/o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Content != "Name content." {
		t.Errorf("name_section.content = %q; want %q", node.NameSection.Content, "Name content.")
	}
	if node.Public == nil {
		t.Fatalf("public should be present")
	}
	if node.Public.Content != "Public content." {
		t.Errorf("public.content = %q; want %q", node.Public.Content, "Public content.")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("public.subsections length = %d; want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Content != "Interface content." {
		t.Errorf("subsection content = %q; want %q", node.Public.Subsections[0].Content, "Interface content.")
	}
}

// ---------------------------------------------------------------------------
// Frontmatter Handling
// ---------------------------------------------------------------------------

func TestNodeParse_FrontmatterIsSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "---\noutputs:\n  - id: foo\n    path: some/path.go\n---\n\n# ROOT/p\n\nBody content.\n"
	testWriteNodeFile(t, "ROOT/p", content)

	node, err := parsenode.NodeParse("ROOT/p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "root/p" {
		t.Errorf("name_section.heading = %q; want %q", node.NameSection.Heading, "root/p")
	}
	if node.NameSection.Content != "Body content." {
		t.Errorf("name_section.content = %q; want %q", node.NameSection.Content, "Body content.")
	}
}

func TestNodeParse_NoFrontmatterDelimiters(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/q\n\nBody content.\n"
	testWriteNodeFile(t, "ROOT/q", content)

	node, err := parsenode.NodeParse("ROOT/q")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "root/q" {
		t.Errorf("name_section.heading = %q; want %q", node.NameSection.Heading, "root/q")
	}
	if node.NameSection.Content != "Body content." {
		t.Errorf("name_section.content = %q; want %q", node.NameSection.Content, "Body content.")
	}
}

func TestNodeParse_UnclosedFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "---\nkey: value\n# ROOT/r\n\nBody content.\n"
	testWriteNodeFile(t, "ROOT/r", content)

	_, err := parsenode.NodeParse("ROOT/r")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v; want %v", err, parsenode.ErrUnexpectedContentBeforeFirstHeading)
	}
}

// ---------------------------------------------------------------------------
// Failure Cases
// ---------------------------------------------------------------------------

func TestNodeParse_ArtifactReferenceRejected(t *testing.T) {
	_, err := parsenode.NodeParse("ARTIFACT/x(y)")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !errors.Is(err, parsenode.ErrNotRootReference) {
		t.Errorf("error = %v; want %v", err, parsenode.ErrNotRootReference)
	}
}

func TestNodeParse_QualifierRejected(t *testing.T) {
	_, err := parsenode.NodeParse("ROOT/x(interface)")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !errors.Is(err, parsenode.ErrHasQualifier) {
		t.Errorf("error = %v; want %v", err, parsenode.ErrHasQualifier)
	}
}

func TestNodeParse_FileDoesNotExist(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := parsenode.NodeParse("ROOT/nonexistent/node")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !errors.Is(err, parsenode.ErrFileUnreadable) {
		t.Errorf("error = %v; want %v", err, parsenode.ErrFileUnreadable)
	}
}

func TestNodeParse_ContentBeforeFirstHeading(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "Some non-blank text before any heading.\n\n# ROOT/s\n\nContent.\n"
	testWriteNodeFile(t, "ROOT/s", content)

	_, err := parsenode.NodeParse("ROOT/s")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v; want %v", err, parsenode.ErrUnexpectedContentBeforeFirstHeading)
	}
}

func TestNodeParse_Level2HeadingBeforeLevel1Heading(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "## Subheading before any section\n\n# ROOT/t\n\nContent.\n"
	testWriteNodeFile(t, "ROOT/t", content)

	_, err := parsenode.NodeParse("ROOT/t")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v; want %v", err, parsenode.ErrUnexpectedContentBeforeFirstHeading)
	}
}

func TestNodeParse_EmptyBody(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "ROOT/u", "")

	_, err := parsenode.NodeParse("ROOT/u")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v; want %v", err, parsenode.ErrUnexpectedContentBeforeFirstHeading)
	}
}

func TestNodeParse_EmptyBodyWithFrontmatterOnly(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "---\nkey: value\n---\n"
	testWriteNodeFile(t, "ROOT/v", content)

	_, err := parsenode.NodeParse("ROOT/v")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v; want %v", err, parsenode.ErrUnexpectedContentBeforeFirstHeading)
	}
}

func TestNodeParse_NodeNameDoesNotMatch(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/other\n\nContent.\n"
	testWriteNodeFile(t, "ROOT/x", content)

	_, err := parsenode.NodeParse("ROOT/x")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !errors.Is(err, parsenode.ErrNodeNameDoesNotMatch) {
		t.Errorf("error = %v; want %v", err, parsenode.ErrNodeNameDoesNotMatch)
	}
}

func TestNodeParse_NodeNameCaseMismatchIsNotError(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# root/x\n\nContent.\n"
	testWriteNodeFile(t, "ROOT/x", content)

	node, err := parsenode.NodeParse("ROOT/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "root/x" {
		t.Errorf("name_section.heading = %q; want %q", node.NameSection.Heading, "root/x")
	}
}

func TestNodeParse_DuplicatePublicSectionSameCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/w\n\n# Public\n\nFirst public.\n\n# Public\n\nSecond public.\n"
	testWriteNodeFile(t, "ROOT/w", content)

	_, err := parsenode.NodeParse("ROOT/w")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicatePublicSection) {
		t.Errorf("error = %v; want %v", err, parsenode.ErrDuplicatePublicSection)
	}
}

func TestNodeParse_DuplicatePublicSectionDifferentCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/w2\n\n# Public\n\nFirst public.\n\n# PUBLIC\n\nSecond public.\n"
	testWriteNodeFile(t, "ROOT/w2", content)

	_, err := parsenode.NodeParse("ROOT/w2")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicatePublicSection) {
		t.Errorf("error = %v; want %v", err, parsenode.ErrDuplicatePublicSection)
	}
}

func TestNodeParse_DuplicateAgentSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/w3\n\n# Agent\n\nFirst agent.\n\n# Agent\n\nSecond agent.\n"
	testWriteNodeFile(t, "ROOT/w3", content)

	_, err := parsenode.NodeParse("ROOT/w3")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateAgentSection) {
		t.Errorf("error = %v; want %v", err, parsenode.ErrDuplicateAgentSection)
	}
}

func TestNodeParse_DuplicateSubsectionInPublicSameCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/w4\n\n# Public\n\n## Interface\n\nFirst.\n\n## Interface\n\nSecond.\n"
	testWriteNodeFile(t, "ROOT/w4", content)

	_, err := parsenode.NodeParse("ROOT/w4")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v; want %v", err, parsenode.ErrDuplicateSubsection)
	}
}

func TestNodeParse_DuplicateSubsectionInPublicDifferentCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/w5\n\n# Public\n\n## Interface\n\nFirst.\n\n## INTERFACE\n\nSecond.\n"
	testWriteNodeFile(t, "ROOT/w5", content)

	_, err := parsenode.NodeParse("ROOT/w5")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v; want %v", err, parsenode.ErrDuplicateSubsection)
	}
}

func TestNodeParse_DuplicateSubsectionInPublicWhitespaceVariation(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/w6\n\n# Public\n\n## Interface\n\nFirst.\n\n##   Interface\n\nSecond.\n"
	testWriteNodeFile(t, "ROOT/w6", content)

	_, err := parsenode.NodeParse("ROOT/w6")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v; want %v", err, parsenode.ErrDuplicateSubsection)
	}
}

func TestNodeParse_DuplicateSubsectionInAgent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# ROOT/w7\n\n# Agent\n\n## Details\n\nFirst.\n\n## Details\n\nSecond.\n"
	testWriteNodeFile(t, "ROOT/w7", content)

	_, err := parsenode.NodeParse("ROOT/w7")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v; want %v", err, parsenode.ErrDuplicateSubsection)
	}
}
