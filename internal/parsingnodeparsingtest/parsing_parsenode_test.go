// code-from-spec: SPEC/golang/tests/parsing/node_parsing@nXGlhv1BTktepywZLsazhM9QWwY
package parsingnodeparsingtest

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
	dir := relPath[:len(relPath)-len("/"+lastSegment(relPath))]
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	if err := os.WriteFile(relPath, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile %s: %v", relPath, err)
	}
}

func lastSegment(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

func TestParseFrontmatter_Complete(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `---
depends_on:
  - SPEC/b
  - ARTIFACT/c
  - EXTERNAL/d.proto
input: "some/input.txt"
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
		t.Fatalf("expected 3 depends_on, got %d", len(node.Frontmatter.DependsOn))
	}
	if node.Frontmatter.DependsOn[0] != "SPEC/b" {
		t.Errorf("depends_on[0] = %q, want SPEC/b", node.Frontmatter.DependsOn[0])
	}
	if node.Frontmatter.DependsOn[1] != "ARTIFACT/c" {
		t.Errorf("depends_on[1] = %q, want ARTIFACT/c", node.Frontmatter.DependsOn[1])
	}
	if node.Frontmatter.DependsOn[2] != "EXTERNAL/d.proto" {
		t.Errorf("depends_on[2] = %q, want EXTERNAL/d.proto", node.Frontmatter.DependsOn[2])
	}
	if node.Frontmatter.Input == nil {
		t.Fatal("expected Input not nil")
	}
	if *node.Frontmatter.Input != "some/input.txt" {
		t.Errorf("Input = %q, want some/input.txt", *node.Frontmatter.Input)
	}
	if node.Frontmatter.Output == nil {
		t.Fatal("expected Output not nil")
	}
	if *node.Frontmatter.Output != "some/output.go" {
		t.Errorf("Output = %q, want some/output.go", *node.Frontmatter.Output)
	}
}

func TestParseFrontmatter_OnlyOutput(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `---
output: "out.go"
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
	if *node.Frontmatter.Output != "out.go" {
		t.Errorf("Output = %q, want out.go", *node.Frontmatter.Output)
	}
}

func TestParseFrontmatter_OnlyDependsOn(t *testing.T) {
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
		t.Fatalf("expected 2 depends_on, got %d", len(node.Frontmatter.DependsOn))
	}
	if node.Frontmatter.Input != nil {
		t.Errorf("expected Input nil, got %v", node.Frontmatter.Input)
	}
	if node.Frontmatter.Output != nil {
		t.Errorf("expected Output nil, got %v", node.Frontmatter.Output)
	}
}

func TestParseFrontmatter_ExternalInDependsOn(t *testing.T) {
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
	if len(node.Frontmatter.DependsOn) != 1 {
		t.Fatalf("expected 1 depends_on, got %d", len(node.Frontmatter.DependsOn))
	}
	if node.Frontmatter.DependsOn[0] != "EXTERNAL/proto/api.proto" {
		t.Errorf("depends_on[0] = %q, want EXTERNAL/proto/api.proto", node.Frontmatter.DependsOn[0])
	}
}

func TestParseFrontmatter_OnlyInput(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `---
input: "my/input.md"
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
	if *node.Frontmatter.Input != "my/input.md" {
		t.Errorf("Input = %q, want my/input.md", *node.Frontmatter.Input)
	}
	if node.Frontmatter.DependsOn != nil {
		t.Errorf("expected DependsOn nil, got %v", node.Frontmatter.DependsOn)
	}
	if node.Frontmatter.Output != nil {
		t.Errorf("expected Output nil, got %v", node.Frontmatter.Output)
	}
}

func TestParseFrontmatter_UnknownFieldsIgnored(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `---
output: "out.go"
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
	if *node.Frontmatter.Output != "out.go" {
		t.Errorf("Output = %q, want out.go", *node.Frontmatter.Output)
	}
}

func TestParseFrontmatter_NoFrontmatter(t *testing.T) {
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

func TestParseFrontmatter_EmptyFrontmatter(t *testing.T) {
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
		t.Errorf("expected Frontmatter nil, got %+v", node.Frontmatter)
	}
}

func TestParseFrontmatter_OnlyFrontmatterNoBody(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `---
output: "out.go"
---
`)

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

func TestParseFrontmatter_DelimiterWithTrailingWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "---   \n# SPEC/a\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

func TestParseFrontmatter_MalformedYAML(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `---
: invalid yaml [[[
---
# SPEC/a
`)

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got %v", err)
	}
}

func TestParseFrontmatter_UnclosedFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `---
output: "out.go"
# SPEC/a
`)

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got %v", err)
	}
}

func TestParseFrontmatter_UnknownFieldExternal(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `---
external: "some/ref"
output: "out.go"
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
	if *node.Frontmatter.Output != "out.go" {
		t.Errorf("Output = %q, want out.go", *node.Frontmatter.Output)
	}
}

func TestParseBody_MinimalNode(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/x/_node.md", `# SPEC/x
A simple node.
`)

	node, err := parsing.ParseNode("SPEC/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "spec/x" {
		t.Errorf("Heading = %q, want spec/x", node.NameSection.Heading)
	}
	if node.NameSection.RawHeading != "# SPEC/x" {
		t.Errorf("RawHeading = %q, want '# SPEC/x'", node.NameSection.RawHeading)
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "A simple node." {
		t.Errorf("Content = %v, want [A simple node.]", node.NameSection.Content)
	}
	if len(node.NameSection.Subsections) != 0 {
		t.Errorf("expected empty Subsections, got %v", node.NameSection.Subsections)
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

func TestParseBody_FullNode(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/payments/fees/_node.md", `---
output: "out.go"
---
# SPEC/payments/fees
Description of fees.

# Public
## Interface
Some interface.
## Constraints
Some constraints.

# Agent
Agent guidance.

# Private
## Decisions
Some decisions.
## Rationale
Some rationale.
`)

	node, err := parsing.ParseNode("SPEC/payments/fees")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "spec/payments/fees" {
		t.Errorf("Heading = %q, want spec/payments/fees", node.NameSection.Heading)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("expected 2 Public subsections, got %d", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("subsection[0].Heading = %q, want interface", node.Public.Subsections[0].Heading)
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("subsection[1].Heading = %q, want constraints", node.Public.Subsections[1].Heading)
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
		t.Errorf("subsection[0].Heading = %q, want decisions", node.Private.Subsections[0].Heading)
	}
	if node.Private.Subsections[1].Heading != "rationale" {
		t.Errorf("subsection[1].Heading = %q, want rationale", node.Private.Subsections[1].Heading)
	}
}

func TestParseBody_NoPublicSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `# SPEC/a
Some content.

# Private
## Rationale
Rationale content.
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

func TestParseBody_PublicContentBeforeSubsection(t *testing.T) {
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
		t.Fatalf("expected 2 Public.Content lines, got %d", len(node.Public.Content))
	}
	if node.Public.Content[0] != "Preamble line one." {
		t.Errorf("Content[0] = %q, want 'Preamble line one.'", node.Public.Content[0])
	}
	if node.Public.Content[1] != "Preamble line two." {
		t.Errorf("Content[1] = %q, want 'Preamble line two.'", node.Public.Content[1])
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("subsection.Heading = %q, want interface", node.Public.Subsections[0].Heading)
	}
}

func TestParseBody_PublicNoContentNoSubsections(t *testing.T) {
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
		t.Errorf("expected empty Content, got %v", node.Public.Content)
	}
	if len(node.Public.Subsections) != 0 {
		t.Errorf("expected empty Subsections, got %v", node.Public.Subsections)
	}
}

func TestParseBody_AgentWithSubsections(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `# SPEC/a

# Agent
Preamble agent.
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
		t.Fatalf("expected 1 Agent.Content line, got %d", len(node.Agent.Content))
	}
	if node.Agent.Content[0] != "Preamble agent." {
		t.Errorf("Content[0] = %q, want 'Preamble agent.'", node.Agent.Content[0])
	}
	if len(node.Agent.Subsections) != 2 {
		t.Fatalf("expected 2 Agent subsections, got %d", len(node.Agent.Subsections))
	}
}

func TestParseBody_PrivateWithSubsections(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `# SPEC/a

# Private
## TODO
Todo content.
## Decisions
Decision content.
## Rationale
Rationale content.
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
		t.Errorf("subsection[0].Heading = %q, want todo", node.Private.Subsections[0].Heading)
	}
	if node.Private.Subsections[1].Heading != "decisions" {
		t.Errorf("subsection[1].Heading = %q, want decisions", node.Private.Subsections[1].Heading)
	}
	if node.Private.Subsections[2].Heading != "rationale" {
		t.Errorf("subsection[2].Heading = %q, want rationale", node.Private.Subsections[2].Heading)
	}
}

func TestParseBody_ContentIsRawMarkdown(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n## Interface\n### Sub heading\n**bold text**\n```\n# code comment\ncode line\n```\n")

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
	if len(sub.Content) != 5 {
		t.Fatalf("expected 5 content lines, got %d: %v", len(sub.Content), sub.Content)
	}
	if sub.Content[0] != "### Sub heading" {
		t.Errorf("Content[0] = %q, want '### Sub heading'", sub.Content[0])
	}
	if sub.Content[1] != "**bold text**" {
		t.Errorf("Content[1] = %q, want '**bold text**'", sub.Content[1])
	}
	if sub.Content[2] != "```" {
		t.Errorf("Content[2] = %q, want '```'", sub.Content[2])
	}
	if sub.Content[3] != "# code comment" {
		t.Errorf("Content[3] = %q, want '# code comment'", sub.Content[3])
	}
	if sub.Content[4] != "```" {
		t.Errorf("Content[4] = %q, want '```'", sub.Content[4])
	}
}

func TestHeadingNorm_CaseInsensitivePublic(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `# SPEC/a

# PUBLIC
Public content.
`)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("Heading = %q, want public", node.Public.Heading)
	}
}

func TestHeadingNorm_PublicMixedCaseWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n#   PuBLiC\nContent.\n")

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("Heading = %q, want public", node.Public.Heading)
	}
}

func TestHeadingNorm_NodeNameVariedWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/e/_node.md", "#   SPEC/e\nSome content.\n")

	node, err := parsing.ParseNode("SPEC/e")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "spec/e" {
		t.Errorf("Heading = %q, want spec/e", node.NameSection.Heading)
	}
}

func TestHeadingNorm_RootHeadingDoesNotMatchSpec(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/x/_node.md", `# ROOT/x
Content.
`)

	_, err := parsing.ParseNode("SPEC/x")
	if !errors.Is(err, parsing.ErrNodeNameDoesNotMatch) {
		t.Errorf("expected ErrNodeNameDoesNotMatch, got %v", err)
	}
}

func TestHeadingNorm_SubsectionHeadingsNormalized(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n##   Interface\nInterface content.\n## CONSTRAINTS\nConstraints content.\n")

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
		t.Errorf("subsection[0].Heading = %q, want interface", node.Public.Subsections[0].Heading)
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("subsection[1].Heading = %q, want constraints", node.Public.Subsections[1].Heading)
	}
}

func TestHeadingNorm_ClosingHashesStripped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n## Interface ##\nContent.\n")

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
		t.Errorf("Heading = %q, want interface", sub.Heading)
	}
	if sub.RawHeading != "## Interface ##" {
		t.Errorf("RawHeading = %q, want '## Interface ##'", sub.RawHeading)
	}
}

func TestRawHeading_PreservesOriginalLine(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n## Interface\nContent.\n")

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if node.Public.RawHeading != "# Public" {
		t.Errorf("Public.RawHeading = %q, want '# Public'", node.Public.RawHeading)
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].RawHeading != "## Interface" {
		t.Errorf("subsection.RawHeading = %q, want '## Interface'", node.Public.Subsections[0].RawHeading)
	}
}

func TestRawHeading_PreservesCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# PUBLIC\nContent.\n")

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("Heading = %q, want public", node.Public.Heading)
	}
	if node.Public.RawHeading != "# PUBLIC" {
		t.Errorf("RawHeading = %q, want '# PUBLIC'", node.Public.RawHeading)
	}
}

func TestRawHeading_PreservesClosingHashes(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n## Foo ##\nContent.\n")

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
		t.Errorf("Heading = %q, want foo", sub.Heading)
	}
	if sub.RawHeading != "## Foo ##" {
		t.Errorf("RawHeading = %q, want '## Foo ##'", sub.RawHeading)
	}
}

func TestRawHeading_PreservesExtraWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n#   Public\nContent.\n")

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("Heading = %q, want public", node.Public.Heading)
	}
	if node.Public.RawHeading != "#   Public" {
		t.Errorf("RawHeading = %q, want '#   Public'", node.Public.RawHeading)
	}
}

func TestContentBoundaries_Level3DeeperAreContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n## Interface\n### Sub\n#### Deep\nContent.\n")

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
	if len(sub.Content) != 3 {
		t.Fatalf("expected 3 content lines, got %d: %v", len(sub.Content), sub.Content)
	}
	if sub.Content[0] != "### Sub" {
		t.Errorf("Content[0] = %q, want '### Sub'", sub.Content[0])
	}
	if sub.Content[1] != "#### Deep" {
		t.Errorf("Content[1] = %q, want '#### Deep'", sub.Content[1])
	}
	if sub.Content[2] != "Content." {
		t.Errorf("Content[2] = %q, want 'Content.'", sub.Content[2])
	}
}

func TestContentBoundaries_FencedBlockBacktickWithHeadingLike(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n## Interface\n```\n# fake heading\n## also fake\n```\nReal content.\n")

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
	if len(sub.Content) != 4 {
		t.Fatalf("expected 4 content lines, got %d: %v", len(sub.Content), sub.Content)
	}
	if sub.Content[0] != "```" {
		t.Errorf("Content[0] = %q, want '```'", sub.Content[0])
	}
	if sub.Content[1] != "# fake heading" {
		t.Errorf("Content[1] = %q, want '# fake heading'", sub.Content[1])
	}
	if sub.Content[2] != "## also fake" {
		t.Errorf("Content[2] = %q, want '## also fake'", sub.Content[2])
	}
	if sub.Content[3] != "```" {
		t.Errorf("Content[3] = %q, want '```'", sub.Content[3])
	}
}

func TestContentBoundaries_FencedBlockTilde(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n## Interface\n~~~\n# heading inside\n~~~\nReal content.\n")

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
	found := false
	for _, line := range sub.Content {
		if line == "# heading inside" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected '# heading inside' in content, got %v", sub.Content)
	}
}

func TestContentBoundaries_FencedBlockWithLanguageTag(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n## Interface\n```python\n# comment\n```\nReal content.\n")

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
	found := false
	for _, line := range sub.Content {
		if line == "# comment" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected '# comment' in content, got %v", sub.Content)
	}
}

func TestContentBoundaries_BlankLineBetweenHeadingAndContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\nContent line.\n")

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public not nil")
	}
	if len(node.Public.Content) < 1 {
		t.Fatalf("expected at least 1 content line, got %d", len(node.Public.Content))
	}
	if node.Public.Content[0] != "" {
		t.Errorf("Content[0] = %q, want empty string", node.Public.Content[0])
	}
	if len(node.Public.Content) < 2 || node.Public.Content[1] != "Content line." {
		t.Errorf("Content[1] = %q, want 'Content line.'", func() string {
			if len(node.Public.Content) < 2 {
				return "<missing>"
			}
			return node.Public.Content[1]
		}())
	}
}

func TestFrontmatterInBody_SkippedBodyParsedCorrectly(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", `---
output: "out.go"
---
# SPEC/a
Some content.
`)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter == nil {
		t.Fatal("expected Frontmatter not nil")
	}
	if node.NameSection.Heading != "spec/a" {
		t.Errorf("Heading = %q, want spec/a", node.NameSection.Heading)
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "Some content." {
		t.Errorf("Content = %v, want [Some content.]", node.NameSection.Content)
	}
}

func TestFrontmatterInBody_NoDelimiters(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\nSome content.\n")

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter != nil {
		t.Errorf("expected Frontmatter nil, got %+v", node.Frontmatter)
	}
	if node.NameSection.Heading != "spec/a" {
		t.Errorf("Heading = %q, want spec/a", node.NameSection.Heading)
	}
}

func TestFrontmatterInBody_UnclosedFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "---\noutput: \"out.go\"\n# SPEC/a\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got %v", err)
	}
}

func TestFailure_ArtifactReferenceRejected(t *testing.T) {
	_, err := parsing.ParseNode("ARTIFACT/x")
	if !errors.Is(err, parsing.ErrNotASpecReference) {
		t.Errorf("expected ErrNotASpecReference, got %v", err)
	}
}

func TestFailure_ExternalReferenceRejected(t *testing.T) {
	_, err := parsing.ParseNode("EXTERNAL/x")
	if !errors.Is(err, parsing.ErrNotASpecReference) {
		t.Errorf("expected ErrNotASpecReference, got %v", err)
	}
}

func TestFailure_QualifierRejected(t *testing.T) {
	_, err := parsing.ParseNode("SPEC/x(interface)")
	if !errors.Is(err, parsing.ErrHasQualifier) {
		t.Errorf("expected ErrHasQualifier, got %v", err)
	}
}

func TestFailure_FileDoesNotExist(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := parsing.ParseNode("SPEC/nonexistent")
	if !errors.Is(err, oslayer.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestFailure_PropagatesPathErrors(t *testing.T) {
	_, err := parsing.ParseNode("SPEC/tra\\versal")
	if !errors.Is(err, oslayer.ErrPathContainsBackslash) {
		t.Errorf("expected ErrPathContainsBackslash, got %v", err)
	}
}

func TestFailure_ContentBeforeFirstHeading(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "Some text before heading.\n# SPEC/a\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

func TestFailure_Level2BeforeLevel1(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "## Interface\n# SPEC/a\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

func TestFailure_EmptyBody(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

func TestFailure_NodeNameDoesNotMatch(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/b\nContent.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrNodeNameDoesNotMatch) {
		t.Errorf("expected ErrNodeNameDoesNotMatch, got %v", err)
	}
}

func TestFailure_NodeNameCaseMismatchNotError(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# spec/a\nContent.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFailure_DuplicatePublicSameCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\nFirst.\n# Public\nSecond.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicatePublicSection) {
		t.Errorf("expected ErrDuplicatePublicSection, got %v", err)
	}
}

func TestFailure_DuplicatePublicDifferentCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\nFirst.\n# PUBLIC\nSecond.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicatePublicSection) {
		t.Errorf("expected ErrDuplicatePublicSection, got %v", err)
	}
}

func TestFailure_DuplicateAgentSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Agent\nFirst.\n# Agent\nSecond.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicateAgentSection) {
		t.Errorf("expected ErrDuplicateAgentSection, got %v", err)
	}
}

func TestFailure_DuplicatePrivateSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Private\nFirst.\n# Private\nSecond.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicatePrivateSection) {
		t.Errorf("expected ErrDuplicatePrivateSection, got %v", err)
	}
}

func TestFailure_UnrecognizedSectionDecisions(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Decisions\nContent.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrUnrecognizedSection) {
		t.Errorf("expected ErrUnrecognizedSection, got %v", err)
	}
}

func TestFailure_UnrecognizedSectionRationale(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Rationale\nContent.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrUnrecognizedSection) {
		t.Errorf("expected ErrUnrecognizedSection, got %v", err)
	}
}

func TestFailure_UnrecognizedSectionTODO(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# TODO\nContent.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrUnrecognizedSection) {
		t.Errorf("expected ErrUnrecognizedSection, got %v", err)
	}
}

func TestFailure_DuplicateSubsectionPublicSameCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n## Interface\nFirst.\n## Interface\nSecond.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicateSubsection) {
		t.Errorf("expected ErrDuplicateSubsection, got %v", err)
	}
}

func TestFailure_DuplicateSubsectionPublicDifferentCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n## Interface\nFirst.\n## INTERFACE\nSecond.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicateSubsection) {
		t.Errorf("expected ErrDuplicateSubsection, got %v", err)
	}
}

func TestFailure_DuplicateSubsectionPublicWhitespaceVariation(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n## Interface\nFirst.\n##   Interface\nSecond.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicateSubsection) {
		t.Errorf("expected ErrDuplicateSubsection, got %v", err)
	}
}

func TestFailure_DuplicateSubsectionAgent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	writeNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Agent\n## Guidance\nFirst.\n## Guidance\nSecond.\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicateSubsection) {
		t.Errorf("expected ErrDuplicateSubsection, got %v", err)
	}
}
