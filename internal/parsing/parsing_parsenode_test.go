// code-from-spec: SPEC/golang/tests/parsing/node_parsing@jQwvz0CYHtfc5Rw3oQ6q_B7fcxU
package parsing_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"
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

func writeNodeFile(t *testing.T, relPath string, content string) {
	t.Helper()
	dir := relPath[:len(relPath)-len("_node.md")]
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("MkdirAll %s: %v", dir, err)
	}
	if err := os.WriteFile(relPath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile %s: %v", relPath, err)
	}
}

func TestParseNode_Frontmatter_Complete(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `---
depends_on:
  - SPEC/b
  - ARTIFACT/c
  - EXTERNAL/proto/api.proto
input: "some/input.md"
output: "some/output.go"
---
# SPEC/a
`)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter == nil {
		t.Fatal("expected Frontmatter not nil")
	}
	if len(node.Frontmatter.DependsOn) != 3 {
		t.Fatalf("expected 3 DependsOn, got %d", len(node.Frontmatter.DependsOn))
	}
	wantDeps := []string{"SPEC/b", "ARTIFACT/c", "EXTERNAL/proto/api.proto"}
	for i, dep := range wantDeps {
		if node.Frontmatter.DependsOn[i] != dep {
			t.Errorf("DependsOn[%d] = %q, want %q", i, node.Frontmatter.DependsOn[i], dep)
		}
	}
	if node.Frontmatter.Input == nil {
		t.Fatal("expected Input not nil")
	}
	if *node.Frontmatter.Input != "some/input.md" {
		t.Errorf("Input = %q, want %q", *node.Frontmatter.Input, "some/input.md")
	}
	if node.Frontmatter.Output == nil {
		t.Fatal("expected Output not nil")
	}
	if *node.Frontmatter.Output != "some/output.go" {
		t.Errorf("Output = %q, want %q", *node.Frontmatter.Output, "some/output.go")
	}
}

func TestParseNode_Frontmatter_OnlyOutput(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `---
output: "some/output.go"
---
# SPEC/a
`)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter == nil {
		t.Fatal("expected Frontmatter not nil")
	}
	if node.Frontmatter.DependsOn != nil {
		t.Errorf("expected DependsOn nil, got %v", node.Frontmatter.DependsOn)
	}
	if node.Frontmatter.Input != nil {
		t.Errorf("expected Input nil, got %v", node.Frontmatter.Input)
	}
	if node.Frontmatter.Output == nil {
		t.Fatal("expected Output not nil")
	}
}

func TestParseNode_Frontmatter_OnlyDependsOn(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `---
depends_on:
  - SPEC/b
  - SPEC/c
---
# SPEC/a
`)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter == nil {
		t.Fatal("expected Frontmatter not nil")
	}
	if len(node.Frontmatter.DependsOn) != 2 {
		t.Fatalf("expected 2 DependsOn, got %d", len(node.Frontmatter.DependsOn))
	}
	if node.Frontmatter.Input != nil {
		t.Errorf("expected Input nil")
	}
	if node.Frontmatter.Output != nil {
		t.Errorf("expected Output nil")
	}
}

func TestParseNode_Frontmatter_ExternalInDependsOn(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `---
depends_on:
  - EXTERNAL/proto/api.proto
---
# SPEC/a
`)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter == nil {
		t.Fatal("expected Frontmatter not nil")
	}
	if len(node.Frontmatter.DependsOn) != 1 || node.Frontmatter.DependsOn[0] != "EXTERNAL/proto/api.proto" {
		t.Errorf("unexpected DependsOn: %v", node.Frontmatter.DependsOn)
	}
}

func TestParseNode_Frontmatter_OnlyInput(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `---
input: "some/input.md"
---
# SPEC/a
`)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter == nil {
		t.Fatal("expected Frontmatter not nil")
	}
	if node.Frontmatter.Input == nil {
		t.Fatal("expected Input not nil")
	}
	if node.Frontmatter.DependsOn != nil {
		t.Errorf("expected DependsOn nil")
	}
	if node.Frontmatter.Output != nil {
		t.Errorf("expected Output nil")
	}
}

func TestParseNode_Frontmatter_UnknownFieldsIgnored(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `---
output: "some/output.go"
custom_field: value
---
# SPEC/a
`)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter == nil {
		t.Fatal("expected Frontmatter not nil")
	}
	if node.Frontmatter.Output == nil {
		t.Fatal("expected Output not nil")
	}
	if *node.Frontmatter.Output != "some/output.go" {
		t.Errorf("Output = %q", *node.Frontmatter.Output)
	}
}

func TestParseNode_NoFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `# SPEC/a
Some content.
`)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter != nil {
		t.Errorf("expected Frontmatter nil, got %+v", node.Frontmatter)
	}
}

func TestParseNode_EmptyFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `---
---
# SPEC/a
`)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter != nil {
		t.Errorf("expected Frontmatter nil for empty block, got %+v", node.Frontmatter)
	}
}

func TestParseNode_FrontmatterOnlyNoBody(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `---
output: "some/output.go"
---
`)

	_, err := parsing.ParseNode("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

func TestParseNode_DelimiterWithTrailingWhitespaceNotRecognized(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "---   \n# SPEC/a\n")

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter != nil {
		t.Errorf("expected Frontmatter nil (delimiter not recognized), got %+v", node.Frontmatter)
	}
}

func TestParseNode_MalformedYAML(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `---
: invalid yaml [[[
---
# SPEC/a
`)

	_, err := parsing.ParseNode("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got %v", err)
	}
}

func TestParseNode_UnclosedFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `---
output: "some/output.go"
# SPEC/a
`)

	_, err := parsing.ParseNode("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got %v", err)
	}
}

func TestParseNode_UnknownExternalFieldIgnored(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `---
external: "some/ref"
output: "some/output.go"
---
# SPEC/a
`)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter == nil {
		t.Fatal("expected Frontmatter not nil")
	}
	if node.Frontmatter.Output == nil {
		t.Fatal("expected Output not nil")
	}
	if *node.Frontmatter.Output != "some/output.go" {
		t.Errorf("Output = %q", *node.Frontmatter.Output)
	}
}

func TestParseNode_MinimalNameSectionOnly(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/x/_node.md", "# SPEC/x\nA simple node.\n")

	node, err := parsing.ParseNode("SPEC/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "spec/x" {
		t.Errorf("NameSection.Heading = %q, want %q", node.NameSection.Heading, "spec/x")
	}
	if node.NameSection.RawHeading != "# SPEC/x" {
		t.Errorf("NameSection.RawHeading = %q, want %q", node.NameSection.RawHeading, "# SPEC/x")
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "A simple node." {
		t.Errorf("NameSection.Content = %v", node.NameSection.Content)
	}
	if len(node.NameSection.Subsections) != 0 {
		t.Errorf("expected no subsections, got %d", len(node.NameSection.Subsections))
	}
	if node.Public != nil {
		t.Errorf("expected Public nil")
	}
	if node.Agent != nil {
		t.Errorf("expected Agent nil")
	}
	if node.Private != nil {
		t.Errorf("expected Private nil")
	}
}

func TestParseNode_FullNodeAllSections(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/payments/fees/_node.md", `---
output: "internal/fees/fees.go"
---
# SPEC/payments/fees
Description of the node.
# Public
## Interface
Some interface content.
## Constraints
Some constraints.
# Agent
Agent guidance here.
# Private
## Decisions
Decision text.
## Rationale
Rationale text.
`)

	node, err := parsing.ParseNode("SPEC/payments/fees")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "spec/payments/fees" {
		t.Errorf("NameSection.Heading = %q", node.NameSection.Heading)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("expected 2 Public subsections, got %d", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("Public subsection[0].Heading = %q", node.Public.Subsections[0].Heading)
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("Public subsection[1].Heading = %q", node.Public.Subsections[1].Heading)
	}
	if len(node.Public.Content) != 0 {
		t.Errorf("expected empty Public.Content, got %v", node.Public.Content)
	}
	if node.Agent == nil {
		t.Fatal("expected Agent not nil")
	}
	if node.Private == nil {
		t.Fatal("expected Private not nil")
	}
	if len(node.Private.Subsections) != 2 {
		t.Fatalf("expected 2 Private subsections, got %d", len(node.Private.Subsections))
	}
	if node.Private.Subsections[0].Heading != "decisions" {
		t.Errorf("Private subsection[0].Heading = %q", node.Private.Subsections[0].Heading)
	}
	if node.Private.Subsections[1].Heading != "rationale" {
		t.Errorf("Private subsection[1].Heading = %q", node.Private.Subsections[1].Heading)
	}
}

func TestParseNode_NoPublicSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `# SPEC/a
Content here.
# Private
## Rationale
Some rationale.
`)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public != nil {
		t.Errorf("expected Public nil")
	}
	if node.Agent != nil {
		t.Errorf("expected Agent nil")
	}
	if node.Private == nil {
		t.Fatal("expected Private not nil")
	}
}

func TestParseNode_PublicContentBeforeSubsection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `# SPEC/a
# Public
Preamble line one.
Preamble line two.
## Interface
Interface content.
`)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if len(node.Public.Content) != 2 {
		t.Fatalf("expected 2 Public.Content lines, got %d: %v", len(node.Public.Content), node.Public.Content)
	}
	if node.Public.Content[0] != "Preamble line one." {
		t.Errorf("Public.Content[0] = %q", node.Public.Content[0])
	}
	if node.Public.Content[1] != "Preamble line two." {
		t.Errorf("Public.Content[1] = %q", node.Public.Content[1])
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 Public subsection, got %d", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("subsection heading = %q", node.Public.Subsections[0].Heading)
	}
}

func TestParseNode_PublicSectionEmpty(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `# SPEC/a
# Public
# Agent
Agent content.
`)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if len(node.Public.Content) != 0 {
		t.Errorf("expected empty Public.Content, got %v", node.Public.Content)
	}
	if len(node.Public.Subsections) != 0 {
		t.Errorf("expected no Public subsections, got %d", len(node.Public.Subsections))
	}
}

func TestParseNode_AgentSectionWithSubsections(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `# SPEC/a
# Agent
Preamble line.
## Implementation guidance
Some guidance.
## Contracts
Some contracts.
`)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Agent == nil {
		t.Fatal("expected Agent not nil")
	}
	if len(node.Agent.Content) != 1 {
		t.Fatalf("expected 1 Agent.Content line, got %d: %v", len(node.Agent.Content), node.Agent.Content)
	}
	if node.Agent.Content[0] != "Preamble line." {
		t.Errorf("Agent.Content[0] = %q", node.Agent.Content[0])
	}
	if len(node.Agent.Subsections) != 2 {
		t.Fatalf("expected 2 Agent subsections, got %d", len(node.Agent.Subsections))
	}
}

func TestParseNode_PrivateSectionWithSubsections(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `# SPEC/a
# Private
## TODO
Todo text.
## Decisions
Decision text.
## Rationale
Rationale text.
`)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Private == nil {
		t.Fatal("expected Private not nil")
	}
	if len(node.Private.Subsections) != 3 {
		t.Fatalf("expected 3 Private subsections, got %d", len(node.Private.Subsections))
	}
	if node.Private.Subsections[0].Heading != "todo" {
		t.Errorf("Private subsection[0].Heading = %q", node.Private.Subsections[0].Heading)
	}
	if node.Private.Subsections[1].Heading != "decisions" {
		t.Errorf("Private subsection[1].Heading = %q", node.Private.Subsections[1].Heading)
	}
	if node.Private.Subsections[2].Heading != "rationale" {
		t.Errorf("Private subsection[2].Heading = %q", node.Private.Subsections[2].Heading)
	}
}

func TestParseNode_ContentIsRawMarkdown(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n# Public\n## Interface\n### Level three\n**bold text**\n```\nsome code\n```\n")

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	expectedLines := []string{"### Level three", "**bold text**", "```", "some code", "```"}
	if len(sub.Content) != len(expectedLines) {
		t.Fatalf("expected %d content lines, got %d: %v", len(expectedLines), len(sub.Content), sub.Content)
	}
	for i, line := range expectedLines {
		if sub.Content[i] != line {
			t.Errorf("Content[%d] = %q, want %q", i, sub.Content[i], line)
		}
	}
}

func TestParseNode_CaseInsensitivePublicDetection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n# PUBLIC\nContent.\n")

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("Public.Heading = %q, want %q", node.Public.Heading, "public")
	}
}

func TestParseNode_PublicMixedCaseAndExtraWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "#   SPEC/a\n#   PuBLiC\nContent.\n")

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("Public.Heading = %q, want %q", node.Public.Heading, "public")
	}
}

func TestParseNode_NodeNameWithVariedWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/e/_node.md", "#   SPEC/e\nContent.\n")

	node, err := parsing.ParseNode("SPEC/e")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "spec/e" {
		t.Errorf("NameSection.Heading = %q, want %q", node.NameSection.Heading, "spec/e")
	}
}

func TestParseNode_RootHeadingDoesNotMatchSpec(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/x/_node.md", "# ROOT/x\nContent.\n")

	_, err := parsing.ParseNode("SPEC/x")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrNodeNameDoesNotMatch) {
		t.Errorf("expected ErrNodeNameDoesNotMatch, got %v", err)
	}
}

func TestParseNode_SubsectionHeadingsNormalized(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n# Public\n##   Interface\n## CONSTRAINTS\nContent.\n")

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("expected 2 subsections, got %d", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("subsection[0].Heading = %q", node.Public.Subsections[0].Heading)
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("subsection[1].Heading = %q", node.Public.Subsections[1].Heading)
	}
}

func TestParseNode_ClosingHashesStripped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n# Public\n## Interface ##\nContent.\n")

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	if sub.Heading != "interface" {
		t.Errorf("Heading = %q, want %q", sub.Heading, "interface")
	}
	if sub.RawHeading != "## Interface ##" {
		t.Errorf("RawHeading = %q, want %q", sub.RawHeading, "## Interface ##")
	}
}

func TestParseNode_RawHeadingPreservesOriginalLine(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n# Public\n## Interface\nContent.\n")

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if node.Public.RawHeading != "# Public" {
		t.Errorf("Public.RawHeading = %q, want %q", node.Public.RawHeading, "# Public")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].RawHeading != "## Interface" {
		t.Errorf("subsection RawHeading = %q, want %q", node.Public.Subsections[0].RawHeading, "## Interface")
	}
}

func TestParseNode_RawHeadingPreservesCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n# PUBLIC\nContent.\n")

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("Heading = %q, want %q", node.Public.Heading, "public")
	}
	if node.Public.RawHeading != "# PUBLIC" {
		t.Errorf("RawHeading = %q, want %q", node.Public.RawHeading, "# PUBLIC")
	}
}

func TestParseNode_RawHeadingPreservesClosingHashes(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n# Public\n## Foo ##\nContent.\n")

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	if sub.Heading != "foo" {
		t.Errorf("Heading = %q, want %q", sub.Heading, "foo")
	}
	if sub.RawHeading != "## Foo ##" {
		t.Errorf("RawHeading = %q, want %q", sub.RawHeading, "## Foo ##")
	}
}

func TestParseNode_RawHeadingPreservesExtraWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n#   Public\nContent.\n")

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("Heading = %q, want %q", node.Public.Heading, "public")
	}
	if node.Public.RawHeading != "#   Public" {
		t.Errorf("RawHeading = %q, want %q", node.Public.RawHeading, "#   Public")
	}
}

func TestParseNode_Level3AndDeeperHeadingsAreContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n# Public\n## Interface\n### Subsub\n#### Deep\nContent.\n")

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	expectedLines := []string{"### Subsub", "#### Deep", "Content."}
	if len(sub.Content) != len(expectedLines) {
		t.Fatalf("expected %d content lines, got %d: %v", len(expectedLines), len(sub.Content), sub.Content)
	}
	for i, line := range expectedLines {
		if sub.Content[i] != line {
			t.Errorf("Content[%d] = %q, want %q", i, sub.Content[i], line)
		}
	}
}

func TestParseNode_FencedCodeBlockWithHeadingLikeContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/a\n# Public\n## Interface\n```\n# not a heading\n## also not\n```\nReal content.\n"
	writeNodeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	expectedLines := []string{"```", "# not a heading", "## also not", "```", "Real content."}
	if len(sub.Content) != len(expectedLines) {
		t.Fatalf("expected %d content lines, got %d: %v", len(expectedLines), len(sub.Content), sub.Content)
	}
	for i, line := range expectedLines {
		if sub.Content[i] != line {
			t.Errorf("Content[%d] = %q, want %q", i, sub.Content[i], line)
		}
	}
}

func TestParseNode_TildeFencedCodeBlock(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/a\n# Public\n## Interface\n~~~\n# heading inside\n~~~\nContent.\n"
	writeNodeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	expectedLines := []string{"~~~", "# heading inside", "~~~", "Content."}
	if len(sub.Content) != len(expectedLines) {
		t.Fatalf("expected %d content lines, got %d: %v", len(expectedLines), len(sub.Content), sub.Content)
	}
	for i, line := range expectedLines {
		if sub.Content[i] != line {
			t.Errorf("Content[%d] = %q, want %q", i, sub.Content[i], line)
		}
	}
}

func TestParseNode_FencedCodeBlockWithLanguageTag(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	content := "# SPEC/a\n# Public\n## Interface\n```python\n# comment\n```\nContent.\n"
	writeNodeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	expectedLines := []string{"```python", "# comment", "```", "Content."}
	if len(sub.Content) != len(expectedLines) {
		t.Fatalf("expected %d content lines, got %d: %v", len(expectedLines), len(sub.Content), sub.Content)
	}
	for i, line := range expectedLines {
		if sub.Content[i] != line {
			t.Errorf("Content[%d] = %q, want %q", i, sub.Content[i], line)
		}
	}
}

func TestParseNode_BlankLinesBetweenHeadingAndContentPreserved(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n# Public\n\nContent line.\n")

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if len(node.Public.Content) != 2 {
		t.Fatalf("expected 2 Public.Content lines, got %d: %v", len(node.Public.Content), node.Public.Content)
	}
	if node.Public.Content[0] != "" {
		t.Errorf("Public.Content[0] = %q, want empty string", node.Public.Content[0])
	}
	if node.Public.Content[1] != "Content line." {
		t.Errorf("Public.Content[1] = %q, want %q", node.Public.Content[1], "Content line.")
	}
}

func TestParseNode_FrontmatterSkippedBodyParsedCorrectly(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "---\noutput: \"some/output.go\"\n---\n# SPEC/a\nBody content.\n")

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter == nil {
		t.Fatal("expected Frontmatter not nil")
	}
	if node.NameSection.Heading != "spec/a" {
		t.Errorf("NameSection.Heading = %q", node.NameSection.Heading)
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "Body content." {
		t.Errorf("NameSection.Content = %v", node.NameSection.Content)
	}
}

func TestParseNode_NoFrontmatterDelimiters_BodyParsedCorrectly(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\nBody content.\n")

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter != nil {
		t.Errorf("expected Frontmatter nil")
	}
	if node.NameSection.Heading != "spec/a" {
		t.Errorf("NameSection.Heading = %q", node.NameSection.Heading)
	}
}

func TestParseNode_UnclosedFrontmatterInBodyContext(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "---\noutput: \"some/output.go\"\n# SPEC/a\n")

	_, err := parsing.ParseNode("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got %v", err)
	}
}

func TestParseNode_ArtifactReferenceRejected(t *testing.T) {
	_, err := parsing.ParseNode("ARTIFACT/x")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrNotASpecReference) {
		t.Errorf("expected ErrNotASpecReference, got %v", err)
	}
}

func TestParseNode_ExternalReferenceRejected(t *testing.T) {
	_, err := parsing.ParseNode("EXTERNAL/x")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrNotASpecReference) {
		t.Errorf("expected ErrNotASpecReference, got %v", err)
	}
}

func TestParseNode_QualifierRejected(t *testing.T) {
	_, err := parsing.ParseNode("SPEC/x(interface)")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrHasQualifier) {
		t.Errorf("expected ErrHasQualifier, got %v", err)
	}
}

func TestParseNode_FileDoesNotExist(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := parsing.ParseNode("SPEC/nonexistent/node")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, oslayer.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestParseNode_PropagatesPathErrors(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := parsing.ParseNode("SPEC/tra\\versal")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, oslayer.ErrPathContainsBackslash) {
		t.Errorf("expected ErrPathContainsBackslash, got %v", err)
	}
}

func TestParseNode_ContentBeforeFirstHeading(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "Some text before heading.\n# SPEC/a\n")

	_, err := parsing.ParseNode("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

func TestParseNode_Level2HeadingBeforeLevel1Heading(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "## Interface\n# SPEC/a\n")

	_, err := parsing.ParseNode("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

func TestParseNode_EmptyBody(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "")

	_, err := parsing.ParseNode("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

func TestParseNode_NodeNameDoesNotMatch(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/b\nContent.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrNodeNameDoesNotMatch) {
		t.Errorf("expected ErrNodeNameDoesNotMatch, got %v", err)
	}
}

func TestParseNode_NodeNameCaseMismatchNotAnError(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# spec/a\nContent.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseNode_DuplicatePublicSectionSameCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n# Public\nContent.\n# Public\nMore content.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrDuplicatePublicSection) {
		t.Errorf("expected ErrDuplicatePublicSection, got %v", err)
	}
}

func TestParseNode_DuplicatePublicSectionDifferentCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n# Public\nContent.\n# PUBLIC\nMore content.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrDuplicatePublicSection) {
		t.Errorf("expected ErrDuplicatePublicSection, got %v", err)
	}
}

func TestParseNode_DuplicateAgentSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n# Agent\nContent.\n# Agent\nMore content.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrDuplicateAgentSection) {
		t.Errorf("expected ErrDuplicateAgentSection, got %v", err)
	}
}

func TestParseNode_DuplicatePrivateSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n# Private\nContent.\n# Private\nMore content.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrDuplicatePrivateSection) {
		t.Errorf("expected ErrDuplicatePrivateSection, got %v", err)
	}
}

func TestParseNode_UnrecognizedSectionDecisions(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n# Decisions\nContent.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrUnrecognizedSection) {
		t.Errorf("expected ErrUnrecognizedSection, got %v", err)
	}
}

func TestParseNode_UnrecognizedSectionRationale(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n# Rationale\nContent.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrUnrecognizedSection) {
		t.Errorf("expected ErrUnrecognizedSection, got %v", err)
	}
}

func TestParseNode_UnrecognizedSectionTODO(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n# TODO\nContent.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrUnrecognizedSection) {
		t.Errorf("expected ErrUnrecognizedSection, got %v", err)
	}
}

func TestParseNode_DuplicateSubsectionInPublicSameCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n# Public\n## Interface\nContent.\n## Interface\nMore.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrDuplicateSubsection) {
		t.Errorf("expected ErrDuplicateSubsection, got %v", err)
	}
}

func TestParseNode_DuplicateSubsectionInPublicDifferentCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n# Public\n## Interface\nContent.\n## INTERFACE\nMore.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrDuplicateSubsection) {
		t.Errorf("expected ErrDuplicateSubsection, got %v", err)
	}
}

func TestParseNode_DuplicateSubsectionInPublicWhitespaceVariation(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n# Public\n## Interface\nContent.\n##   Interface\nMore.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrDuplicateSubsection) {
		t.Errorf("expected ErrDuplicateSubsection, got %v", err)
	}
}

func TestParseNode_DuplicateSubsectionInAgent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n# Agent\n## Guidance\nContent.\n## Guidance\nMore.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrDuplicateSubsection) {
		t.Errorf("expected ErrDuplicateSubsection, got %v", err)
	}
}
