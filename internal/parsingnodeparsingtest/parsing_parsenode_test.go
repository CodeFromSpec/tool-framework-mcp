// code-from-spec: SPEC/golang/test/cases/parsing/node_parsing@G5d-ss71Fa7sDMATcp7nXDu-AuI
package parsingnodeparsingtest

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"
)

func TestParsesCompleteFrontmatter(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/a")
	b.AddDependsOn("SPEC/other")
	b.AddDependsOn("ARTIFACT/thing")
	b.AddDependsOn("EXTERNAL/proto/api.proto")
	b.SetInput("some/input.md")
	b.SetOutput("internal/a/a.go")
	b.Write()

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter == nil {
		t.Fatal("expected Frontmatter to be non-nil")
	}
	if len(node.Frontmatter.DependsOn) != 3 {
		t.Fatalf("expected 3 DependsOn entries, got %d", len(node.Frontmatter.DependsOn))
	}
	if node.Frontmatter.Input == nil {
		t.Fatal("expected Input to be non-nil")
	}
	if *node.Frontmatter.Input != "some/input.md" {
		t.Errorf("expected Input = %q, got %q", "some/input.md", *node.Frontmatter.Input)
	}
	if node.Frontmatter.Output == nil {
		t.Fatal("expected Output to be non-nil")
	}
	if *node.Frontmatter.Output != "internal/a/a.go" {
		t.Errorf("expected Output = %q, got %q", "internal/a/a.go", *node.Frontmatter.Output)
	}
}

func TestParsesFrontmatterWithOnlyOutput(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/a")
	b.SetOutput("internal/a/a.go")
	b.Write()

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter == nil {
		t.Fatal("expected Frontmatter to be non-nil")
	}
	if node.Frontmatter.DependsOn != nil {
		t.Errorf("expected DependsOn to be nil")
	}
	if node.Frontmatter.Input != nil {
		t.Errorf("expected Input to be nil")
	}
	if node.Frontmatter.Output == nil {
		t.Fatal("expected Output to be non-nil")
	}
}

func TestParsesFrontmatterWithOnlyDependsOn(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/a")
	b.AddDependsOn("SPEC/other")
	b.AddDependsOn("SPEC/another")
	b.Write()

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter == nil {
		t.Fatal("expected Frontmatter to be non-nil")
	}
	if len(node.Frontmatter.DependsOn) != 2 {
		t.Fatalf("expected 2 DependsOn entries, got %d", len(node.Frontmatter.DependsOn))
	}
	if node.Frontmatter.Input != nil {
		t.Errorf("expected Input to be nil")
	}
	if node.Frontmatter.Output != nil {
		t.Errorf("expected Output to be nil")
	}
}

func TestParsesFrontmatterWithExternalDependsOn(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/a")
	b.AddDependsOn("EXTERNAL/proto/api.proto")
	b.Write()

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter == nil {
		t.Fatal("expected Frontmatter to be non-nil")
	}
	if len(node.Frontmatter.DependsOn) != 1 {
		t.Fatalf("expected 1 DependsOn entry, got %d", len(node.Frontmatter.DependsOn))
	}
	if node.Frontmatter.DependsOn[0] != "EXTERNAL/proto/api.proto" {
		t.Errorf("unexpected DependsOn value: %q", node.Frontmatter.DependsOn[0])
	}
}

func TestParsesFrontmatterWithOnlyInput(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/a")
	b.SetInput("some/input.md")
	b.Write()

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
		t.Errorf("expected DependsOn to be nil")
	}
	if node.Frontmatter.Output != nil {
		t.Errorf("expected Output to be nil")
	}
}

func TestIgnoresUnknownFrontmatterFields(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/a", "---\noutput: internal/a/a.go\ncustom_field: value\n---\n# SPEC/a\n")

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
	if *node.Frontmatter.Output != "internal/a/a.go" {
		t.Errorf("unexpected Output value: %q", *node.Frontmatter.Output)
	}
}

func TestNoFrontmatterIsNil(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/a", "# SPEC/a\nsome content\n")

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter != nil {
		t.Errorf("expected Frontmatter to be nil")
	}
}

func TestEmptyFrontmatter(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/a", "---\n---\n# SPEC/a\nsome content\n")

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter != nil {
		t.Errorf("expected Frontmatter to be nil for empty frontmatter")
	}
}

func TestFrontmatterOnlyNoBody(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/a", "---\noutput: internal/a/a.go\n---\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

func TestDelimiterWithTrailingWhitespaceNotRecognized(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/a", "---   \noutput: internal/a/a.go\n---\n# SPEC/a\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

func TestMalformedYAML(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/a", "---\n: invalid: yaml: content\n---\n# SPEC/a\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got %v", err)
	}
}

func TestUnclosedFrontmatterBlock(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/a", "---\noutput: internal/a/a.go\n# SPEC/a\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got %v", err)
	}
}

func TestUnknownFieldExternalIgnored(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/a", "---\nexternal: \"some/ref\"\noutput: internal/a/a.go\n---\n# SPEC/a\n")

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
	if *node.Frontmatter.Output != "internal/a/a.go" {
		t.Errorf("unexpected Output: %q", *node.Frontmatter.Output)
	}
}

func TestMinimalNodeNameSectionOnly(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/x", "# SPEC/x\nA simple node.\n")

	node, err := parsing.ParseNode("SPEC/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "spec/x" {
		t.Errorf("expected Heading = %q, got %q", "spec/x", node.NameSection.Heading)
	}
	if node.NameSection.RawHeading != "# SPEC/x" {
		t.Errorf("expected RawHeading = %q, got %q", "# SPEC/x", node.NameSection.RawHeading)
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "A simple node." {
		t.Errorf("unexpected Content: %v", node.NameSection.Content)
	}
	if len(node.NameSection.Subsections) != 0 {
		t.Errorf("expected no Subsections, got %d", len(node.NameSection.Subsections))
	}
	if node.Public != nil {
		t.Error("expected Public to be nil")
	}
	if node.Agent != nil {
		t.Error("expected Agent to be nil")
	}
	if node.Private != nil {
		t.Error("expected Private to be nil")
	}
}

func TestFullNodeAllSectionTypes(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/payments/fees")
	b.SetOutput("internal/payments/fees.go")
	b.SetPublic("## Interface\nsome interface content\n## Constraints\nsome constraints\n")
	b.SetAgent("agent guidance\n")
	b.SetPrivate("## Decisions\ndecision content\n## Rationale\nrationale content\n")
	b.Write()

	node, err := parsing.ParseNode("SPEC/payments/fees")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "spec/payments/fees" {
		t.Errorf("unexpected NameSection.Heading: %q", node.NameSection.Heading)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("expected 2 Public subsections, got %d", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("unexpected first subsection heading: %q", node.Public.Subsections[0].Heading)
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("unexpected second subsection heading: %q", node.Public.Subsections[1].Heading)
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
		t.Errorf("unexpected first private subsection heading: %q", node.Private.Subsections[0].Heading)
	}
	if node.Private.Subsections[1].Heading != "rationale" {
		t.Errorf("unexpected second private subsection heading: %q", node.Private.Subsections[1].Heading)
	}
}

func TestNodeWithNoPublicSection(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/x")
	b.SetPrivate("## Rationale\nsome rationale\n")
	b.Write()

	node, err := parsing.ParseNode("SPEC/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public != nil {
		t.Error("expected Public to be nil")
	}
	if node.Agent != nil {
		t.Error("expected Agent to be nil")
	}
	if node.Private == nil {
		t.Fatal("expected Private to be non-nil")
	}
}

func TestPublicSectionWithContentBeforeFirstSubsection(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/x")
	b.SetPublic("preamble line one\npreamble line two\n## Interface\ninterface content\n")
	b.Write()

	node, err := parsing.ParseNode("SPEC/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if len(node.Public.Content) != 2 {
		t.Errorf("expected 2 Public.Content lines, got %d: %v", len(node.Public.Content), node.Public.Content)
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("unexpected subsection heading: %q", node.Public.Subsections[0].Heading)
	}
}

func TestPublicSectionWithNoContentOrSubsections(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/x")
	b.SetPublic("")
	b.SetAgent("agent content\n")
	b.Write()

	node, err := parsing.ParseNode("SPEC/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if len(node.Public.Content) != 0 {
		t.Errorf("expected empty Public.Content, got %v", node.Public.Content)
	}
	if len(node.Public.Subsections) != 0 {
		t.Errorf("expected empty Public.Subsections, got %d", len(node.Public.Subsections))
	}
}

func TestAgentSectionWithSubsections(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/x")
	b.SetAgent("preamble agent line\n## Implementation guidance\nimpl content\n## Contracts\ncontracts content\n")
	b.Write()

	node, err := parsing.ParseNode("SPEC/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Agent == nil {
		t.Fatal("expected Agent to be non-nil")
	}
	if len(node.Agent.Content) != 1 {
		t.Errorf("expected 1 Agent.Content line, got %d: %v", len(node.Agent.Content), node.Agent.Content)
	}
	if len(node.Agent.Subsections) != 2 {
		t.Fatalf("expected 2 Agent.Subsections, got %d", len(node.Agent.Subsections))
	}
}

func TestPrivateSectionWithSubsections(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/x")
	b.SetPrivate("## TODO\ntodo content\n## Decisions\ndecision content\n## Rationale\nrationale content\n")
	b.Write()

	node, err := parsing.ParseNode("SPEC/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Private == nil {
		t.Fatal("expected Private to be non-nil")
	}
	if len(node.Private.Subsections) != 3 {
		t.Fatalf("expected 3 Private subsections, got %d", len(node.Private.Subsections))
	}
	headings := []string{"todo", "decisions", "rationale"}
	for i, expected := range headings {
		if node.Private.Subsections[i].Heading != expected {
			t.Errorf("subsection[%d].Heading = %q, want %q", i, node.Private.Subsections[i].Heading, expected)
		}
	}
}

func TestContentIsRawMarkdown(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/x")
	b.SetPublic("## Interface\n### Sub heading\n**bold text**\n```go\ncode line\n```\n")
	b.Write()

	node, err := parsing.ParseNode("SPEC/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	content := node.Public.Subsections[0].Content
	if len(content) != 5 {
		t.Errorf("expected 5 content lines, got %d: %v", len(content), content)
	}
}

func TestCaseInsensitivePublicDetection(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/x", "# SPEC/x\n# PUBLIC\npublic content\n")

	node, err := parsing.ParseNode("SPEC/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("unexpected Heading: %q", node.Public.Heading)
	}
}

func TestPublicWithMixedCaseAndExtraWhitespace(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/x", "# SPEC/x\n#   PuBLiC\npublic content\n")

	node, err := parsing.ParseNode("SPEC/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("unexpected Heading: %q", node.Public.Heading)
	}
}

func TestNodeNameWithVariedWhitespace(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/e", "#   SPEC/e\nsome content\n")

	node, err := parsing.ParseNode("SPEC/e")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "spec/e" {
		t.Errorf("unexpected Heading: %q", node.NameSection.Heading)
	}
}

func TestNodeNameROOTDoesNotMatchSPEC(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/x", "# ROOT/x\nsome content\n")

	_, err := parsing.ParseNode("SPEC/x")
	if !errors.Is(err, parsing.ErrNodeNameDoesNotMatch) {
		t.Errorf("expected ErrNodeNameDoesNotMatch, got %v", err)
	}
}

func TestSubsectionHeadingsNormalized(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/x")
	b.SetPublic("##   Interface\ncontent\n## CONSTRAINTS\ncontent\n")
	b.Write()

	node, err := parsing.ParseNode("SPEC/x")
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
		t.Errorf("unexpected heading[0]: %q", node.Public.Subsections[0].Heading)
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("unexpected heading[1]: %q", node.Public.Subsections[1].Heading)
	}
}

func TestClosingHashesStripped(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/x")
	b.SetPublic("## Interface ##\nsome content\n")
	b.Write()

	node, err := parsing.ParseNode("SPEC/x")
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
		t.Errorf("unexpected Heading: %q", sub.Heading)
	}
	if sub.RawHeading != "## Interface ##" {
		t.Errorf("unexpected RawHeading: %q", sub.RawHeading)
	}
}

func TestRawHeadingPreservesOriginalLine(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/x", "# SPEC/x\n# Public\n## Interface\nsome content\n")

	node, err := parsing.ParseNode("SPEC/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if node.Public.RawHeading != "# Public" {
		t.Errorf("unexpected Public.RawHeading: %q", node.Public.RawHeading)
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].RawHeading != "## Interface" {
		t.Errorf("unexpected subsection RawHeading: %q", node.Public.Subsections[0].RawHeading)
	}
}

func TestRawHeadingPreservesCase(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/x", "# SPEC/x\n# PUBLIC\ncontent\n")

	node, err := parsing.ParseNode("SPEC/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("unexpected Heading: %q", node.Public.Heading)
	}
	if node.Public.RawHeading != "# PUBLIC" {
		t.Errorf("unexpected RawHeading: %q", node.Public.RawHeading)
	}
}

func TestRawHeadingPreservesClosingHashes(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/x")
	b.SetPublic("## Foo ##\ncontent\n")
	b.Write()

	node, err := parsing.ParseNode("SPEC/x")
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
		t.Errorf("unexpected Heading: %q", sub.Heading)
	}
	if sub.RawHeading != "## Foo ##" {
		t.Errorf("unexpected RawHeading: %q", sub.RawHeading)
	}
}

func TestRawHeadingPreservesExtraWhitespace(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/x", "# SPEC/x\n#   Public\ncontent\n")

	node, err := parsing.ParseNode("SPEC/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("unexpected Heading: %q", node.Public.Heading)
	}
	if node.Public.RawHeading != "#   Public" {
		t.Errorf("unexpected RawHeading: %q", node.Public.RawHeading)
	}
}

func TestLevel3AndDeeperHeadingsAreContent(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/x")
	b.SetPublic("## Interface\n### Sub\n#### Deep\ncontent\n")
	b.Write()

	node, err := parsing.ParseNode("SPEC/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	content := node.Public.Subsections[0].Content
	if len(content) != 3 {
		t.Errorf("expected 3 content lines, got %d: %v", len(content), content)
	}
	if content[0] != "### Sub" {
		t.Errorf("expected content[0] = %q, got %q", "### Sub", content[0])
	}
	if content[1] != "#### Deep" {
		t.Errorf("expected content[1] = %q, got %q", "#### Deep", content[1])
	}
}

func TestFencedCodeBlockWithHeadingLikeContent(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/x")
	b.SetPublic("## Interface\n```\n# heading inside fence\n## another\n```\n")
	b.Write()

	node, err := parsing.ParseNode("SPEC/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	content := node.Public.Subsections[0].Content
	if len(content) != 4 {
		t.Errorf("expected 4 content lines, got %d: %v", len(content), content)
	}
}

func TestFencedCodeBlockWithTildeFence(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/x")
	b.SetPublic("## Interface\n~~~\n# heading\n~~~\n")
	b.Write()

	node, err := parsing.ParseNode("SPEC/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	content := node.Public.Subsections[0].Content
	if len(content) != 3 {
		t.Errorf("expected 3 content lines, got %d: %v", len(content), content)
	}
}

func TestFencedCodeBlockWithLanguageTag(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/x")
	b.SetPublic("## Interface\n```python\n# comment\n```\n")
	b.Write()

	node, err := parsing.ParseNode("SPEC/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("expected 1 subsection, got %d", len(node.Public.Subsections))
	}
	content := node.Public.Subsections[0].Content
	if len(content) != 3 {
		t.Errorf("expected 3 content lines, got %d: %v", len(content), content)
	}
}

func TestBlankLinesBetweenHeadingAndContentPreserved(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/x", "# SPEC/x\n# Public\n\ncontent line\n")

	node, err := parsing.ParseNode("SPEC/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("expected Public to be non-nil")
	}
	if len(node.Public.Content) != 2 {
		t.Errorf("expected 2 Public.Content lines, got %d: %v", len(node.Public.Content), node.Public.Content)
	}
	if node.Public.Content[0] != "" {
		t.Errorf("expected Content[0] = empty string, got %q", node.Public.Content[0])
	}
	if node.Public.Content[1] != "content line" {
		t.Errorf("expected Content[1] = %q, got %q", "content line", node.Public.Content[1])
	}
}

func TestFrontmatterSkippedBodyParsedCorrectly(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/a")
	b.SetOutput("internal/a/a.go")
	b.Write()

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter == nil {
		t.Fatal("expected Frontmatter to be non-nil")
	}
	if node.NameSection.Heading != "spec/a" {
		t.Errorf("unexpected NameSection.Heading: %q", node.NameSection.Heading)
	}
}

func TestNoFrontmatterDelimitersBodyParsedCorrectly(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/a", "# SPEC/a\nbody content\n")

	node, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Frontmatter != nil {
		t.Error("expected Frontmatter to be nil")
	}
	if node.NameSection.Heading != "spec/a" {
		t.Errorf("unexpected NameSection.Heading: %q", node.NameSection.Heading)
	}
}

func TestUnclosedFrontmatterInBodyContext(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/a", "---\noutput: internal/a/a.go\n# SPEC/a\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got %v", err)
	}
}

func TestArtifactReferenceRejected(t *testing.T) {
	_, err := parsing.ParseNode("ARTIFACT/x")
	if !errors.Is(err, parsing.ErrNotASpecReference) {
		t.Errorf("expected ErrNotASpecReference, got %v", err)
	}
}

func TestExternalReferenceRejected(t *testing.T) {
	_, err := parsing.ParseNode("EXTERNAL/x")
	if !errors.Is(err, parsing.ErrNotASpecReference) {
		t.Errorf("expected ErrNotASpecReference, got %v", err)
	}
}

func TestQualifierRejected(t *testing.T) {
	_, err := parsing.ParseNode("SPEC/x(interface)")
	if !errors.Is(err, parsing.ErrHasQualifier) {
		t.Errorf("expected ErrHasQualifier, got %v", err)
	}
}

func TestFileDoesNotExist(t *testing.T) {
	testutils.Chdir(t)

	_, err := parsing.ParseNode("SPEC/nonexistent")
	if !errors.Is(err, oslayer.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestPropagatesPathErrors(t *testing.T) {
	testutils.Chdir(t)

	_, err := parsing.ParseNode("SPEC/tra\\versal")
	if !errors.Is(err, oslayer.ErrPathContainsBackslash) {
		t.Errorf("expected ErrPathContainsBackslash, got %v", err)
	}
}

func TestContentBeforeFirstHeading(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/a", "some text before heading\n# SPEC/a\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

func TestLevel2HeadingBeforeLevel1Heading(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/a", "## Interface\n# SPEC/a\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

func TestEmptyBody(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/a", "")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("expected ErrUnexpectedContentBeforeFirstHeading, got %v", err)
	}
}

func TestNodeNameDoesNotMatchLogicalName(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/a", "# SPEC/wrongname\ncontent\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrNodeNameDoesNotMatch) {
		t.Errorf("expected ErrNodeNameDoesNotMatch, got %v", err)
	}
}

func TestNodeNameCaseMismatchIsNotError(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/a", "# spec/a\ncontent\n")

	_, err := parsing.ParseNode("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDuplicatePublicSectionSameCase(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/a", "# SPEC/a\n# Public\ncontent\n# Public\nmore content\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicatePublicSection) {
		t.Errorf("expected ErrDuplicatePublicSection, got %v", err)
	}
}

func TestDuplicatePublicSectionDifferentCase(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/a", "# SPEC/a\n# Public\ncontent\n# PUBLIC\nmore content\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicatePublicSection) {
		t.Errorf("expected ErrDuplicatePublicSection, got %v", err)
	}
}

func TestDuplicateAgentSection(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/a", "# SPEC/a\n# Agent\ncontent\n# Agent\nmore content\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicateAgentSection) {
		t.Errorf("expected ErrDuplicateAgentSection, got %v", err)
	}
}

func TestDuplicatePrivateSection(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/a", "# SPEC/a\n# Private\ncontent\n# Private\nmore content\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicatePrivateSection) {
		t.Errorf("expected ErrDuplicatePrivateSection, got %v", err)
	}
}

func TestUnrecognizedSectionHeading(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/a", "# SPEC/a\n# Decisions\ncontent\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrUnrecognizedSection) {
		t.Errorf("expected ErrUnrecognizedSection, got %v", err)
	}
}

func TestUnrecognizedSectionRationale(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/a", "# SPEC/a\n# Rationale\ncontent\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrUnrecognizedSection) {
		t.Errorf("expected ErrUnrecognizedSection, got %v", err)
	}
}

func TestUnrecognizedSectionTODO(t *testing.T) {
	testutils.Chdir(t)

	testutils.WriteRawNode(t, "SPEC/a", "# SPEC/a\n# TODO\ncontent\n")

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrUnrecognizedSection) {
		t.Errorf("expected ErrUnrecognizedSection, got %v", err)
	}
}

func TestDuplicateSubsectionInPublicSameCase(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/a")
	b.SetPublic("## Interface\ncontent\n## Interface\nmore content\n")
	b.Write()

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicateSubsection) {
		t.Errorf("expected ErrDuplicateSubsection, got %v", err)
	}
}

func TestDuplicateSubsectionInPublicDifferentCase(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/a")
	b.SetPublic("## Interface\ncontent\n## INTERFACE\nmore content\n")
	b.Write()

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicateSubsection) {
		t.Errorf("expected ErrDuplicateSubsection, got %v", err)
	}
}

func TestDuplicateSubsectionInPublicWhitespaceVariation(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/a")
	b.SetPublic("## Interface\ncontent\n##   Interface\nmore content\n")
	b.Write()

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicateSubsection) {
		t.Errorf("expected ErrDuplicateSubsection, got %v", err)
	}
}

func TestDuplicateSubsectionInAgent(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/a")
	b.SetAgent("## Guidance\ncontent\n## Guidance\nmore content\n")
	b.Write()

	_, err := parsing.ParseNode("SPEC/a")
	if !errors.Is(err, parsing.ErrDuplicateSubsection) {
		t.Errorf("expected ErrDuplicateSubsection, got %v", err)
	}
}
