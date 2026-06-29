// code-from-spec: SPEC/golang/tests/parsing/node_parsing@uQNJncGooIm6rdk9ppqQbr6m6LI
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

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath(path), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func filepath(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return "."
}

func TestParseFrontmatter_Complete(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
depends_on:
  - SPEC/other
  - ARTIFACT/foo
  - EXTERNAL/proto/api.proto
input: "some/input.txt"
output: "some/output.go"
---
# SPEC/a
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter == nil {
		t.Fatal("expected Frontmatter to be non-nil")
	}
	if len(node.Frontmatter.DependsOn) != 3 {
		t.Fatalf("expected 3 depends_on, got %d", len(node.Frontmatter.DependsOn))
	}
	if node.Frontmatter.DependsOn[0] != "SPEC/other" {
		t.Errorf("expected SPEC/other, got %s", node.Frontmatter.DependsOn[0])
	}
	if node.Frontmatter.DependsOn[1] != "ARTIFACT/foo" {
		t.Errorf("expected ARTIFACT/foo, got %s", node.Frontmatter.DependsOn[1])
	}
	if node.Frontmatter.DependsOn[2] != "EXTERNAL/proto/api.proto" {
		t.Errorf("expected EXTERNAL/proto/api.proto, got %s", node.Frontmatter.DependsOn[2])
	}
	if node.Frontmatter.Input == nil {
		t.Fatal("expected Input to be non-nil")
	}
	if *node.Frontmatter.Input != "some/input.txt" {
		t.Errorf("expected input some/input.txt, got %s", *node.Frontmatter.Input)
	}
	if node.Frontmatter.Output == nil {
		t.Fatal("expected Output to be non-nil")
	}
	if *node.Frontmatter.Output != "some/output.go" {
		t.Errorf("expected output some/output.go, got %s", *node.Frontmatter.Output)
	}
}

func TestParseFrontmatter_OnlyOutput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
output: "internal/foo/foo.go"
---
# SPEC/a
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter == nil {
		t.Fatal("expected Frontmatter to be non-nil")
	}
	if node.Frontmatter.DependsOn != nil {
		t.Errorf("expected DependsOn nil, got %v", node.Frontmatter.DependsOn)
	}
	if node.Frontmatter.Input != nil {
		t.Errorf("expected Input nil, got %v", node.Frontmatter.Input)
	}
	if node.Frontmatter.Output == nil {
		t.Fatal("expected Output to be non-nil")
	}
}

func TestParseFrontmatter_OnlyDependsOn(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
depends_on:
  - SPEC/other
  - ARTIFACT/bar
---
# SPEC/a
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter == nil {
		t.Fatal("expected Frontmatter to be non-nil")
	}
	if len(node.Frontmatter.DependsOn) != 2 {
		t.Fatalf("expected 2 depends_on, got %d", len(node.Frontmatter.DependsOn))
	}
	if node.Frontmatter.Input != nil {
		t.Errorf("expected Input nil")
	}
	if node.Frontmatter.Output != nil {
		t.Errorf("expected Output nil")
	}
}

func TestParseFrontmatter_ExternalDependsOn(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
depends_on:
  - EXTERNAL/proto/api.proto
---
# SPEC/a
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter == nil {
		t.Fatal("expected Frontmatter to be non-nil")
	}
	if len(node.Frontmatter.DependsOn) != 1 {
		t.Fatalf("expected 1 depends_on, got %d", len(node.Frontmatter.DependsOn))
	}
	if node.Frontmatter.DependsOn[0] != "EXTERNAL/proto/api.proto" {
		t.Errorf("expected EXTERNAL/proto/api.proto, got %s", node.Frontmatter.DependsOn[0])
	}
}

func TestParseFrontmatter_OnlyInput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
input: "source/material.md"
---
# SPEC/a
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter == nil {
		t.Fatal("expected Frontmatter to be non-nil")
	}
	if node.Frontmatter.Input == nil {
		t.Fatal("expected Input to be non-nil")
	}
	if node.Frontmatter.DependsOn != nil {
		t.Errorf("expected DependsOn nil")
	}
	if node.Frontmatter.Output != nil {
		t.Errorf("expected Output nil")
	}
}

func TestParseFrontmatter_UnknownFieldIgnored(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
output: "internal/foo/foo.go"
custom_field: value
---
# SPEC/a
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter == nil {
		t.Fatal("expected Frontmatter to be non-nil")
	}
	if node.Frontmatter.Output == nil {
		t.Fatal("expected Output to be non-nil")
	}
}

func TestParseFrontmatter_NoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a
Some content.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter != nil {
		t.Errorf("expected Frontmatter nil, got %+v", node.Frontmatter)
	}
}

func TestParseFrontmatter_Empty(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
---
# SPEC/a
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter != nil {
		t.Errorf("expected Frontmatter nil for empty frontmatter, got %+v", node.Frontmatter)
	}
}

func TestParseFrontmatter_OnlyFrontmatterNoBody(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
output: "internal/foo/foo.go"
---
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrUnexpectedContentBeforeFirstHeading) {
		t.Fatalf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

func TestParseFrontmatter_DelimiterWithTrailingWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "---   \noutput: \"internal/foo/foo.go\"\n---\n# SPEC/a\n"
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter != nil {
		t.Errorf("expected Frontmatter nil (delimiter not recognized), got %+v", node.Frontmatter)
	}
}

func TestParseFrontmatter_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
: invalid yaml: [
---
# SPEC/a
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrMalformedYAML) {
		t.Fatalf("expected ErrMalformedYAML, got %v", err)
	}
}

func TestParseFrontmatter_UnclosedFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
output: "internal/foo/foo.go"
# SPEC/a
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrMalformedYAML) {
		t.Fatalf("expected ErrMalformedYAML, got %v", err)
	}
}

func TestParseFrontmatter_UnknownExternalField(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
external: "some/ref"
output: "internal/foo/foo.go"
---
# SPEC/a
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter == nil {
		t.Fatal("expected Frontmatter to be non-nil")
	}
	if node.Frontmatter.Output == nil {
		t.Fatal("expected Output to be non-nil")
	}
}

func TestParseBody_MinimalNode(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/x
A simple node.
`
	writeFile(t, "code-from-spec/x/_node.md", content)

	node, err := parsing.ParseNode("SPEC/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "spec/x" {
		t.Errorf("expected heading spec/x, got %s", node.NameSection.Heading)
	}
	if node.NameSection.RawHeading != "# SPEC/x" {
		t.Errorf("expected raw heading '# SPEC/x', got %s", node.NameSection.RawHeading)
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "A simple node." {
		t.Errorf("expected content ['A simple node.'], got %v", node.NameSection.Content)
	}
	if len(node.NameSection.Subsections) != 0 {
		t.Errorf("expected no subsections, got %v", node.NameSection.Subsections)
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
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
output: "internal/payments/fees/fees.go"
---
# SPEC/payments/fees
Description of fees.

# Public
## Interface
Some interface content.
## Constraints
Some constraints.

# Agent
Agent guidance here.

# Private
## Decisions
Some decisions.
## Rationale
Some rationale.
`
	if err := os.MkdirAll("code-from-spec/payments/fees", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile("code-from-spec/payments/fees/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsing.ParseNode("SPEC/payments/fees")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "spec/payments/fees" {
		t.Errorf("expected spec/payments/fees, got %s", node.NameSection.Heading)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("expected 2 Public subsections, got %d", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("expected interface, got %s", node.Public.Subsections[0].Heading)
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("expected constraints, got %s", node.Public.Subsections[1].Heading)
	}
	if len(node.Public.Content) != 0 {
		t.Errorf("expected empty Public.Content, got %v", node.Public.Content)
	}
	if node.Agent == nil {
		t.Fatal("expected Agent to be non-nil")
	}
	if node.Private == nil {
		t.Fatal("expected Private to be non-nil")
	}
	if len(node.Private.Subsections) != 2 {
		t.Fatalf("expected 2 Private subsections, got %d", len(node.Private.Subsections))
	}
	if node.Private.Subsections[0].Heading != "decisions" {
		t.Errorf("expected decisions, got %s", node.Private.Subsections[0].Heading)
	}
	if node.Private.Subsections[1].Heading != "rationale" {
		t.Errorf("expected rationale, got %s", node.Private.Subsections[1].Heading)
	}
}

func TestParseBody_NoPublicSection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a
Some content.

# Private
## Rationale
Some rationale.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

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
		t.Fatal("expected Private to be non-nil")
	}
}

func TestParseBody_PublicContentBeforeSubsection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a

# Public
Preamble line one.
Preamble line two.
## Interface
Interface content.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if len(node.Public.Content) != 2 {
		t.Fatalf("expected 2 preamble lines, got %d: %v", len(node.Public.Content), node.Public.Content)
	}
	if node.Public.Content[0] != "Preamble line one." {
		t.Errorf("expected 'Preamble line one.', got %s", node.Public.Content[0])
	}
	if node.Public.Content[1] != "Preamble line two." {
		t.Errorf("expected 'Preamble line two.', got %s", node.Public.Content[1])
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
}

func TestParseBody_PublicNoContentNoSubsections(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a

# Public
# Agent
Agent content.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if len(node.Public.Content) != 0 {
		t.Errorf("expected empty Content, got %v", node.Public.Content)
	}
	if len(node.Public.Subsections) != 0 {
		t.Errorf("expected empty Subsections, got %v", node.Public.Subsections)
	}
}

func TestParseBody_AgentWithSubsections(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a

# Agent
Preamble agent line.
## Implementation guidance
Some guidance.
## Contracts
Some contracts.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Agent == nil {
		t.Fatal("expected Agent to be non-nil")
	}
	if len(node.Agent.Content) != 1 {
		t.Fatalf("expected 1 content line, got %d", len(node.Agent.Content))
	}
	if len(node.Agent.Subsections) != 2 {
		t.Fatalf("expected 2 subsections, got %d", len(node.Agent.Subsections))
	}
}

func TestParseBody_PrivateWithSubsections(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a

# Private
## TODO
Todo items.
## Decisions
Some decisions.
## Rationale
Some rationale.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Private == nil {
		t.Fatal("expected Private to be non-nil")
	}
	if len(node.Private.Subsections) != 3 {
		t.Fatalf("expected 3 subsections, got %d", len(node.Private.Subsections))
	}
}

func TestParseBody_ContentIsRawMarkdown(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a

# Public
## Interface
### Level three
**Bold text**
` + "```" + `
code block
` + "```" + `
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	if len(sub.Content) < 4 {
		t.Fatalf("expected at least 4 content lines, got %d: %v", len(sub.Content), sub.Content)
	}
	if sub.Content[0] != "### Level three" {
		t.Errorf("expected '### Level three', got %s", sub.Content[0])
	}
	if sub.Content[1] != "**Bold text**" {
		t.Errorf("expected '**Bold text**', got %s", sub.Content[1])
	}
}

func TestHeading_CaseInsensitivePublic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a

# PUBLIC
Public content.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("expected heading 'public', got %s", node.Public.Heading)
	}
}

func TestHeading_PublicMixedCaseExtraWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "#   PuBLiC\nPublic content.\n"
	writeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n"+content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("expected heading 'public', got %s", node.Public.Heading)
	}
}

func TestHeading_NodeNameVariedWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `#   SPEC/e
Content.
`
	writeFile(t, "code-from-spec/e/_node.md", content)

	node, err := parsing.ParseNode("SPEC/e")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "spec/e" {
		t.Errorf("expected 'spec/e', got %s", node.NameSection.Heading)
	}
}

func TestHeading_RootPrefixDoesNotMatchSpec(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# ROOT/x
Content.
`
	writeFile(t, "code-from-spec/x/_node.md", content)

	_, err := parsing.ParseNode("SPEC/x")
	if !errors.Is(err, parsing.ErrNodeNameDoesNotMatch) {
		t.Fatalf("expected ErrNodeNameDoesNotMatch, got %v", err)
	}
}

func TestHeading_SubsectionsNormalized(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a

# Public
##   Interface
Interface content.
## CONSTRAINTS
Constraints content.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("expected 2 subsections, got %d", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("expected 'interface', got %s", node.Public.Subsections[0].Heading)
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("expected 'constraints', got %s", node.Public.Subsections[1].Heading)
	}
}

func TestHeading_ClosingHashesStripped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a

# Public
## Interface ##
Interface content.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	if sub.Heading != "interface" {
		t.Errorf("expected 'interface', got %s", sub.Heading)
	}
	if sub.RawHeading != "## Interface ##" {
		t.Errorf("expected '## Interface ##', got %s", sub.RawHeading)
	}
}

func TestRawHeading_Preserved(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a

# Public
## Interface
Interface content.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if node.Public.RawHeading != "# Public" {
		t.Errorf("expected '# Public', got %s", node.Public.RawHeading)
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].RawHeading != "## Interface" {
		t.Errorf("expected '## Interface', got %s", node.Public.Subsections[0].RawHeading)
	}
}

func TestRawHeading_PreservesCase(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a

# PUBLIC
Public content.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("expected 'public', got %s", node.Public.Heading)
	}
	if node.Public.RawHeading != "# PUBLIC" {
		t.Errorf("expected '# PUBLIC', got %s", node.Public.RawHeading)
	}
}

func TestRawHeading_PreservesClosingHashes(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a

# Public
## Foo ##
Foo content.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	if sub.Heading != "foo" {
		t.Errorf("expected 'foo', got %s", sub.Heading)
	}
	if sub.RawHeading != "## Foo ##" {
		t.Errorf("expected '## Foo ##', got %s", sub.RawHeading)
	}
}

func TestRawHeading_PreservesExtraWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "#   Public\nPublic content.\n"
	writeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n"+content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("expected 'public', got %s", node.Public.Heading)
	}
	if node.Public.RawHeading != "#   Public" {
		t.Errorf("expected '#   Public', got %s", node.Public.RawHeading)
	}
}

func TestContent_Level3AndDeeperAreContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a

# Public
## Interface
### Level 3
#### Level 4
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	if len(sub.Content) != 2 {
		t.Fatalf("expected 2 content lines, got %d: %v", len(sub.Content), sub.Content)
	}
	if sub.Content[0] != "### Level 3" {
		t.Errorf("expected '### Level 3', got %s", sub.Content[0])
	}
	if sub.Content[1] != "#### Level 4" {
		t.Errorf("expected '#### Level 4', got %s", sub.Content[1])
	}
}

func TestContent_FencedCodeBlockWithHeadingLike(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# SPEC/a\n\n# Public\n## Interface\n```\n# fake heading\n## another fake\n```\n"
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	foundFakeHeading := false
	for _, line := range sub.Content {
		if line == "# fake heading" {
			foundFakeHeading = true
		}
	}
	if !foundFakeHeading {
		t.Errorf("expected '# fake heading' in content, got %v", sub.Content)
	}
}

func TestContent_FencedCodeBlockTildeFence(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# SPEC/a\n\n# Public\n## Interface\n~~~\n# heading inside\n~~~\n"
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	foundHeading := false
	for _, line := range sub.Content {
		if line == "# heading inside" {
			foundHeading = true
		}
	}
	if !foundHeading {
		t.Errorf("expected '# heading inside' in content, got %v", sub.Content)
	}
}

func TestContent_FencedCodeBlockWithLanguageTag(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# SPEC/a\n\n# Public\n## Interface\n```python\n# comment\n```\n"
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	foundComment := false
	for _, line := range sub.Content {
		if line == "# comment" {
			foundComment = true
		}
	}
	if !foundComment {
		t.Errorf("expected '# comment' in content, got %v", sub.Content)
	}
}

func TestContent_BlankLinesPreserved(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a

# Public

Some content.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if len(node.Public.Content) < 2 {
		t.Fatalf("expected at least 2 content lines, got %d: %v", len(node.Public.Content), node.Public.Content)
	}
	if node.Public.Content[0] != "" {
		t.Errorf("expected empty string first, got %q", node.Public.Content[0])
	}
	if node.Public.Content[1] != "Some content." {
		t.Errorf("expected 'Some content.', got %q", node.Public.Content[1])
	}
}

func TestFrontmatterBodyParsing_FrontmatterSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
output: "internal/foo/foo.go"
---
# SPEC/a
Body content.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter == nil {
		t.Fatal("expected Frontmatter non-nil")
	}
	if node.NameSection.Heading != "spec/a" {
		t.Errorf("expected 'spec/a', got %s", node.NameSection.Heading)
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "Body content." {
		t.Errorf("expected ['Body content.'], got %v", node.NameSection.Content)
	}
}

func TestFrontmatterBodyParsing_NoFrontmatterDelimiters(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a
Body content.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter != nil {
		t.Errorf("expected Frontmatter nil")
	}
	if node.NameSection.Heading != "spec/a" {
		t.Errorf("expected 'spec/a', got %s", node.NameSection.Heading)
	}
}

func TestFrontmatterBodyParsing_UnclosedFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
output: "internal/foo/foo.go"
# SPEC/a
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrMalformedYAML) {
		t.Fatalf("expected ErrMalformedYAML, got %v", err)
	}
}

func TestFailure_ArtifactRejected(t *testing.T) {
	_, err := parsing.ParseNode("ARTIFACT/x")
	if !errors.Is(err, parsing.ErrNotASpecReference) {
		t.Fatalf("expected ErrNotASpecReference, got %v", err)
	}
}

func TestFailure_ExternalRejected(t *testing.T) {
	_, err := parsing.ParseNode("EXTERNAL/x")
	if !errors.Is(err, parsing.ErrNotASpecReference) {
		t.Fatalf("expected ErrNotASpecReference, got %v", err)
	}
}

func TestFailure_QualifierRejected(t *testing.T) {
	_, err := parsing.ParseNode("SPEC/x(interface)")
	if !errors.Is(err, parsing.ErrHasQualifier) {
		t.Fatalf("expected ErrHasQualifier, got %v", err)
	}
}

func TestFailure_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := parsing.ParseNode("SPEC/nonexistent/node")
	if !errors.Is(err, oslayer.ErrFileUnreadable) {
		t.Fatalf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestFailure_PropagatesPathErrors(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := parsing.ParseNode("SPEC/tra\\versal")
	if !errors.Is(err, oslayer.ErrPathContainsBackslash) {
		t.Fatalf("expected ErrPathContainsBackslash, got %v", err)
	}
}

func TestFailure_ContentBeforeFirstHeading(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `Some text before heading.
# SPEC/a
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrUnexpectedContentBeforeFirstHeading) {
		t.Fatalf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

func TestFailure_Level2BeforeLevel1(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `## Interface
Some content.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrUnexpectedContentBeforeFirstHeading) {
		t.Fatalf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

func TestFailure_EmptyBody(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	writeFile(t, "code-from-spec/a/_node.md", "")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrUnexpectedContentBeforeFirstHeading) {
		t.Fatalf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

func TestFailure_NodeNameDoesNotMatch(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/b
Content.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrNodeNameDoesNotMatch) {
		t.Fatalf("expected ErrNodeNameDoesNotMatch, got %v", err)
	}
}

func TestFailure_NodeNameCaseMismatchNotError(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# spec/a
Content.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	_, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFailure_DuplicatePublicSameCase(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a

# Public
First public.

# Public
Second public.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicatePublicSection) {
		t.Fatalf("expected ErrDuplicatePublicSection, got %v", err)
	}
}

func TestFailure_DuplicatePublicDifferentCase(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a

# Public
First public.

# PUBLIC
Second public.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicatePublicSection) {
		t.Fatalf("expected ErrDuplicatePublicSection, got %v", err)
	}
}

func TestFailure_DuplicateAgentSection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a

# Agent
First agent.

# Agent
Second agent.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicateAgentSection) {
		t.Fatalf("expected ErrDuplicateAgentSection, got %v", err)
	}
}

func TestFailure_DuplicatePrivateSection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a

# Private
First private.

# Private
Second private.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicatePrivateSection) {
		t.Fatalf("expected ErrDuplicatePrivateSection, got %v", err)
	}
}

func TestFailure_UnrecognizedSectionDecisions(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a

# Decisions
Some decisions.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrUnrecognizedSection) {
		t.Fatalf("expected ErrUnrecognizedSection, got %v", err)
	}
}

func TestFailure_UnrecognizedSectionRationale(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a

# Rationale
Some rationale.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrUnrecognizedSection) {
		t.Fatalf("expected ErrUnrecognizedSection, got %v", err)
	}
}

func TestFailure_UnrecognizedSectionTODO(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a

# TODO
Some todo.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrUnrecognizedSection) {
		t.Fatalf("expected ErrUnrecognizedSection, got %v", err)
	}
}

func TestFailure_DuplicateSubsectionPublicSameCase(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a

# Public
## Interface
First.
## Interface
Second.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicateSubsection) {
		t.Fatalf("expected ErrDuplicateSubsection, got %v", err)
	}
}

func TestFailure_DuplicateSubsectionPublicDifferentCase(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a

# Public
## Interface
First.
## INTERFACE
Second.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicateSubsection) {
		t.Fatalf("expected ErrDuplicateSubsection, got %v", err)
	}
}

func TestFailure_DuplicateSubsectionPublicWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# SPEC/a\n\n# Public\n## Interface\nFirst.\n##   Interface\nSecond.\n"
	writeFile(t, "code-from-spec/a/_node.md", content)

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicateSubsection) {
		t.Fatalf("expected ErrDuplicateSubsection, got %v", err)
	}
}

func TestFailure_DuplicateSubsectionAgent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# SPEC/a

# Agent
## Guidance
First.
## Guidance
Second.
`
	writeFile(t, "code-from-spec/a/_node.md", content)

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicateSubsection) {
		t.Fatalf("expected ErrDuplicateSubsection, got %v", err)
	}
}
