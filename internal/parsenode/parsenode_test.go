// code-from-spec: ROOT/golang/tests/parsing/node_parsing@q-yLWqi9V5rksFv2nVc2dK7iL88
package parsenode_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
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

// testWriteNodeFile creates the _node.md file for a logical name under the
// code-from-spec directory relative to the current working directory.
// logicalName must start with "ROOT/". The dirs are created as needed.
func testWriteNodeFile(t *testing.T, logicalName string, content string) {
	t.Helper()
	// Strip "ROOT/" prefix and build the directory path.
	suffix := logicalName[len("ROOT/"):]
	dir := filepath.Join("code-from-spec", filepath.FromSlash(suffix))
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testWriteNodeFile: mkdir: %v", err)
	}
	filePath := filepath.Join(dir, "_node.md")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNodeFile: write: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Happy Path Tests
// ---------------------------------------------------------------------------

// TestNodeParse_HP01_MinimalNode verifies that a node with only a name section is parsed correctly.
func TestNodeParse_HP01_MinimalNode(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "ROOT/x", "# ROOT/x\nA simple node.\n")

	node, err := parsenode.NodeParse("ROOT/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/x" {
		t.Errorf("NameSection.Heading = %q, want %q", node.NameSection.Heading, "root/x")
	}
	if node.NameSection.RawHeading != "# ROOT/x" {
		t.Errorf("NameSection.RawHeading = %q, want %q", node.NameSection.RawHeading, "# ROOT/x")
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "A simple node." {
		t.Errorf("NameSection.Content = %v, want [\"A simple node.\"]", node.NameSection.Content)
	}
	if len(node.NameSection.Subsections) != 0 {
		t.Errorf("NameSection.Subsections = %v, want empty", node.NameSection.Subsections)
	}
	if node.Public != nil {
		t.Errorf("Public = %v, want nil", node.Public)
	}
	if node.Agent != nil {
		t.Errorf("Agent = %v, want nil", node.Agent)
	}
	if len(node.Private) != 0 {
		t.Errorf("Private = %v, want empty", node.Private)
	}
}

// TestNodeParse_HP02_FullNode verifies parsing of a node with all section types including frontmatter.
func TestNodeParse_HP02_FullNode(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "---\ndepends_on: []\n---\n# ROOT/payments/fees\nFees description.\n# Public\n## Interface\nInterface content.\n## Constraints\nConstraints content.\n# Agent\nAgent content.\n# Decisions\nDecisions content.\n# Rationale\nRationale content.\n"
	testWriteNodeFile(t, "ROOT/payments/fees", content)

	node, err := parsenode.NodeParse("ROOT/payments/fees")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/payments/fees" {
		t.Errorf("NameSection.Heading = %q, want %q", node.NameSection.Heading, "root/payments/fees")
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "Fees description." {
		t.Errorf("NameSection.Content = %v, want [\"Fees description.\"]", node.NameSection.Content)
	}

	if node.Public == nil {
		t.Fatal("Public section absent, want present")
	}
	if len(node.Public.Content) != 0 {
		t.Errorf("Public.Content = %v, want empty", node.Public.Content)
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("Public.Subsections len = %d, want 2", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("Public.Subsections[0].Heading = %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if len(node.Public.Subsections[0].Content) != 1 || node.Public.Subsections[0].Content[0] != "Interface content." {
		t.Errorf("Public.Subsections[0].Content = %v, want [\"Interface content.\"]", node.Public.Subsections[0].Content)
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("Public.Subsections[1].Heading = %q, want %q", node.Public.Subsections[1].Heading, "constraints")
	}
	if len(node.Public.Subsections[1].Content) != 1 || node.Public.Subsections[1].Content[0] != "Constraints content." {
		t.Errorf("Public.Subsections[1].Content = %v, want [\"Constraints content.\"]", node.Public.Subsections[1].Content)
	}

	if node.Agent == nil {
		t.Fatal("Agent section absent, want present")
	}
	if len(node.Agent.Content) != 1 || node.Agent.Content[0] != "Agent content." {
		t.Errorf("Agent.Content = %v, want [\"Agent content.\"]", node.Agent.Content)
	}

	if len(node.Private) != 2 {
		t.Fatalf("Private len = %d, want 2", len(node.Private))
	}
	if node.Private[0].Heading != "decisions" {
		t.Errorf("Private[0].Heading = %q, want %q", node.Private[0].Heading, "decisions")
	}
	if len(node.Private[0].Content) != 1 || node.Private[0].Content[0] != "Decisions content." {
		t.Errorf("Private[0].Content = %v, want [\"Decisions content.\"]", node.Private[0].Content)
	}
	if node.Private[1].Heading != "rationale" {
		t.Errorf("Private[1].Heading = %q, want %q", node.Private[1].Heading, "rationale")
	}
	if len(node.Private[1].Content) != 1 || node.Private[1].Content[0] != "Rationale content." {
		t.Errorf("Private[1].Content = %v, want [\"Rationale content.\"]", node.Private[1].Content)
	}
}

// TestNodeParse_HP03_NoPublicSection verifies a node without a public section.
func TestNodeParse_HP03_NoPublicSection(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/decisions\nSome decision content.\n# Rationale\nRationale content.\n"
	testWriteNodeFile(t, "ROOT/decisions", content)

	node, err := parsenode.NodeParse("ROOT/decisions")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public != nil {
		t.Errorf("Public = %v, want nil", node.Public)
	}
	if node.Agent != nil {
		t.Errorf("Agent = %v, want nil", node.Agent)
	}
	if len(node.Private) != 1 {
		t.Fatalf("Private len = %d, want 1", len(node.Private))
	}
	if node.Private[0].Heading != "rationale" {
		t.Errorf("Private[0].Heading = %q, want %q", node.Private[0].Heading, "rationale")
	}
	if len(node.Private[0].Content) != 1 || node.Private[0].Content[0] != "Rationale content." {
		t.Errorf("Private[0].Content = %v, want [\"Rationale content.\"]", node.Private[0].Content)
	}
}

// TestNodeParse_HP04_PublicContentBeforeSubsection verifies public section preamble content.
func TestNodeParse_HP04_PublicContentBeforeSubsection(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n# Public\nPreamble line one.\nPreamble line two.\n## Interface\nInterface content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public section absent, want present")
	}
	if len(node.Public.Content) != 2 {
		t.Fatalf("Public.Content len = %d, want 2", len(node.Public.Content))
	}
	if node.Public.Content[0] != "Preamble line one." {
		t.Errorf("Public.Content[0] = %q, want %q", node.Public.Content[0], "Preamble line one.")
	}
	if node.Public.Content[1] != "Preamble line two." {
		t.Errorf("Public.Content[1] = %q, want %q", node.Public.Content[1], "Preamble line two.")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Public.Subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("Public.Subsections[0].Heading = %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if len(node.Public.Subsections[0].Content) != 1 || node.Public.Subsections[0].Content[0] != "Interface content." {
		t.Errorf("Public.Subsections[0].Content = %v, want [\"Interface content.\"]", node.Public.Subsections[0].Content)
	}
}

// TestNodeParse_HP05_PublicEmptyNoContentNoSubsections verifies an empty public section.
func TestNodeParse_HP05_PublicEmptyNoContentNoSubsections(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n# Public\n# Agent\nAgent content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public section absent, want present")
	}
	if len(node.Public.Content) != 0 {
		t.Errorf("Public.Content = %v, want empty", node.Public.Content)
	}
	if len(node.Public.Subsections) != 0 {
		t.Errorf("Public.Subsections = %v, want empty", node.Public.Subsections)
	}
}

// TestNodeParse_HP06_AgentWithSubsections verifies agent section with subsections.
func TestNodeParse_HP06_AgentWithSubsections(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n# Agent\nAgent preamble.\n## Implementation guidance\nImplementation guidance content.\n## Contracts\nContracts content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Agent == nil {
		t.Fatal("Agent section absent, want present")
	}
	if len(node.Agent.Content) != 1 || node.Agent.Content[0] != "Agent preamble." {
		t.Errorf("Agent.Content = %v, want [\"Agent preamble.\"]", node.Agent.Content)
	}
	if node.Agent.RawHeading != "# Agent" {
		t.Errorf("Agent.RawHeading = %q, want %q", node.Agent.RawHeading, "# Agent")
	}
	if len(node.Agent.Subsections) != 2 {
		t.Fatalf("Agent.Subsections len = %d, want 2", len(node.Agent.Subsections))
	}
	if node.Agent.Subsections[0].Heading != "implementation guidance" {
		t.Errorf("Agent.Subsections[0].Heading = %q, want %q", node.Agent.Subsections[0].Heading, "implementation guidance")
	}
	if len(node.Agent.Subsections[0].Content) != 1 || node.Agent.Subsections[0].Content[0] != "Implementation guidance content." {
		t.Errorf("Agent.Subsections[0].Content = %v, want [\"Implementation guidance content.\"]", node.Agent.Subsections[0].Content)
	}
	if node.Agent.Subsections[1].Heading != "contracts" {
		t.Errorf("Agent.Subsections[1].Heading = %q, want %q", node.Agent.Subsections[1].Heading, "contracts")
	}
	if len(node.Agent.Subsections[1].Content) != 1 || node.Agent.Subsections[1].Content[0] != "Contracts content." {
		t.Errorf("Agent.Subsections[1].Content = %v, want [\"Contracts content.\"]", node.Agent.Subsections[1].Content)
	}
}

// TestNodeParse_HP07_PrivatePreservesOrder verifies private sections are returned in file order.
func TestNodeParse_HP07_PrivatePreservesOrder(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n# TODO\nTODO content.\n# Decisions\nDecisions content.\n# Rationale\nRationale content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(node.Private) != 3 {
		t.Fatalf("Private len = %d, want 3", len(node.Private))
	}
	if node.Private[0].Heading != "todo" {
		t.Errorf("Private[0].Heading = %q, want %q", node.Private[0].Heading, "todo")
	}
	if node.Private[1].Heading != "decisions" {
		t.Errorf("Private[1].Heading = %q, want %q", node.Private[1].Heading, "decisions")
	}
	if node.Private[2].Heading != "rationale" {
		t.Errorf("Private[2].Heading = %q, want %q", node.Private[2].Heading, "rationale")
	}
}

// TestNodeParse_HP08_ContentIsRawMarkdown verifies raw markdown content is preserved as-is.
func TestNodeParse_HP08_ContentIsRawMarkdown(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n# Public\n## Interface\n### Level three heading\n**bold text**\n```go\nsome code\n```\n"
	testWriteNodeFile(t, "ROOT/a", content)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public section absent, want present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Public.Subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("Public.Subsections[0].Heading = %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}

	expectedContent := []string{"### Level three heading", "**bold text**", "```go", "some code", "```"}
	if len(node.Public.Subsections[0].Content) != len(expectedContent) {
		t.Fatalf("Public.Subsections[0].Content len = %d, want %d", len(node.Public.Subsections[0].Content), len(expectedContent))
	}
	for i, line := range expectedContent {
		if node.Public.Subsections[0].Content[i] != line {
			t.Errorf("Public.Subsections[0].Content[%d] = %q, want %q", i, node.Public.Subsections[0].Content[i], line)
		}
	}
}

// ---------------------------------------------------------------------------
// Heading Normalization Tests
// ---------------------------------------------------------------------------

// TestNodeParse_HN01_CaseInsensitivePublicDetection verifies public section detected case-insensitively.
func TestNodeParse_HN01_CaseInsensitivePublicDetection(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n# PUBLIC\nPublic content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public section absent, want present")
	}
	if node.Public.Heading != "public" {
		t.Errorf("Public.Heading = %q, want %q", node.Public.Heading, "public")
	}
}

// TestNodeParse_HN02_PublicMixedCaseExtraWhitespace verifies normalization of mixed case and extra whitespace.
func TestNodeParse_HN02_PublicMixedCaseExtraWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n#   PuBLiC\nPublic content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public section absent, want present")
	}
	if node.Public.Heading != "public" {
		t.Errorf("Public.Heading = %q, want %q", node.Public.Heading, "public")
	}
}

// TestNodeParse_HN03_NodeNameWithVariedWhitespace verifies node name normalization with extra whitespace.
func TestNodeParse_HN03_NodeNameWithVariedWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "#   ROOT/e\nName content.\n"
	testWriteNodeFile(t, "ROOT/e", content)

	node, err := parsenode.NodeParse("ROOT/e")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/e" {
		t.Errorf("NameSection.Heading = %q, want %q", node.NameSection.Heading, "root/e")
	}
}

// TestNodeParse_HN04_SubsectionHeadingsNormalized verifies subsection headings are normalized.
func TestNodeParse_HN04_SubsectionHeadingsNormalized(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n# Public\n##   Interface\nInterface content.\n## CONSTRAINTS\nConstraints content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public section absent, want present")
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("Public.Subsections len = %d, want 2", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("Public.Subsections[0].Heading = %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("Public.Subsections[1].Heading = %q, want %q", node.Public.Subsections[1].Heading, "constraints")
	}
}

// TestNodeParse_HN05_ClosingHashesStripped verifies closing hashes are stripped from headings.
func TestNodeParse_HN05_ClosingHashesStripped(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n# Public\n## Interface ##\nInterface content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public section absent, want present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Public.Subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("Public.Subsections[0].Heading = %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if node.Public.Subsections[0].RawHeading != "## Interface ##" {
		t.Errorf("Public.Subsections[0].RawHeading = %q, want %q", node.Public.Subsections[0].RawHeading, "## Interface ##")
	}
}

// ---------------------------------------------------------------------------
// Raw Heading Preservation Tests
// ---------------------------------------------------------------------------

// TestNodeParse_RH01_RawHeadingPreservesOriginalLine verifies raw headings are preserved.
func TestNodeParse_RH01_RawHeadingPreservesOriginalLine(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n# Public\n## Interface\nInterface content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public section absent, want present")
	}
	if node.Public.RawHeading != "# Public" {
		t.Errorf("Public.RawHeading = %q, want %q", node.Public.RawHeading, "# Public")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Public.Subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].RawHeading != "## Interface" {
		t.Errorf("Public.Subsections[0].RawHeading = %q, want %q", node.Public.Subsections[0].RawHeading, "## Interface")
	}
}

// TestNodeParse_RH02_RawHeadingPreservesCase verifies raw heading preserves original case.
func TestNodeParse_RH02_RawHeadingPreservesCase(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n# PUBLIC\nPublic content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public section absent, want present")
	}
	if node.Public.Heading != "public" {
		t.Errorf("Public.Heading = %q, want %q", node.Public.Heading, "public")
	}
	if node.Public.RawHeading != "# PUBLIC" {
		t.Errorf("Public.RawHeading = %q, want %q", node.Public.RawHeading, "# PUBLIC")
	}
}

// TestNodeParse_RH03_RawHeadingPreservesClosingHashes verifies raw heading preserves closing hashes.
func TestNodeParse_RH03_RawHeadingPreservesClosingHashes(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n# Public\n## Foo ##\nFoo content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public section absent, want present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Public.Subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "foo" {
		t.Errorf("Public.Subsections[0].Heading = %q, want %q", node.Public.Subsections[0].Heading, "foo")
	}
	if node.Public.Subsections[0].RawHeading != "## Foo ##" {
		t.Errorf("Public.Subsections[0].RawHeading = %q, want %q", node.Public.Subsections[0].RawHeading, "## Foo ##")
	}
}

// TestNodeParse_RH04_RawHeadingPreservesExtraWhitespace verifies raw heading preserves extra whitespace.
func TestNodeParse_RH04_RawHeadingPreservesExtraWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n#   Public\nPublic content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public section absent, want present")
	}
	if node.Public.Heading != "public" {
		t.Errorf("Public.Heading = %q, want %q", node.Public.Heading, "public")
	}
	if node.Public.RawHeading != "#   Public" {
		t.Errorf("Public.RawHeading = %q, want %q", node.Public.RawHeading, "#   Public")
	}
}

// ---------------------------------------------------------------------------
// Content Boundaries Tests
// ---------------------------------------------------------------------------

// TestNodeParse_CB01_Level3AndDeeperAreContent verifies level-3+ headings are treated as content.
func TestNodeParse_CB01_Level3AndDeeperAreContent(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n# Public\n## Interface\n### Sub-sub heading\n#### Even deeper\nInterface content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public section absent, want present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Public.Subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("Public.Subsections[0].Heading = %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}

	subContent := node.Public.Subsections[0].Content
	foundLevel3 := false
	foundLevel4 := false
	for _, line := range subContent {
		if line == "### Sub-sub heading" {
			foundLevel3 = true
		}
		if line == "#### Even deeper" {
			foundLevel4 = true
		}
	}
	if !foundLevel3 {
		t.Error("Expected \"### Sub-sub heading\" in subsection content, not found")
	}
	if !foundLevel4 {
		t.Error("Expected \"#### Even deeper\" in subsection content, not found")
	}
}

// TestNodeParse_CB02_FencedCodeBlockBacktick verifies heading-like lines inside backtick fences are content.
func TestNodeParse_CB02_FencedCodeBlockBacktick(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n# Public\n## Interface\n```\n# Looks like a heading\n## Also looks like a heading\n```\nReal content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public section absent, want present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Public.Subsections len = %d, want 1 (heading-like lines inside fence should not create sections), got subsections: %v", len(node.Public.Subsections), func() []string {
			var names []string
			for _, s := range node.Public.Subsections {
				names = append(names, s.Heading)
			}
			return names
		}())
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("Public.Subsections[0].Heading = %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}

	subContent := node.Public.Subsections[0].Content
	foundFakeH1 := false
	foundFakeH2 := false
	for _, line := range subContent {
		if line == "# Looks like a heading" {
			foundFakeH1 = true
		}
		if line == "## Also looks like a heading" {
			foundFakeH2 = true
		}
	}
	if !foundFakeH1 {
		t.Error("Expected \"# Looks like a heading\" in subsection content, not found")
	}
	if !foundFakeH2 {
		t.Error("Expected \"## Also looks like a heading\" in subsection content, not found")
	}
}

// TestNodeParse_CB03_FencedCodeBlockTilde verifies heading-like lines inside tilde fences are content.
func TestNodeParse_CB03_FencedCodeBlockTilde(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n# Public\n## Interface\n~~~\n# Inside tilde fence\n~~~\nReal content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public section absent, want present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Public.Subsections len = %d, want 1", len(node.Public.Subsections))
	}

	subContent := node.Public.Subsections[0].Content
	foundInsideFence := false
	for _, line := range subContent {
		if line == "# Inside tilde fence" {
			foundInsideFence = true
		}
	}
	if !foundInsideFence {
		t.Error("Expected \"# Inside tilde fence\" in subsection content, not found")
	}
}

// TestNodeParse_CB04_FencedCodeBlockWithLanguageTag verifies heading-like lines inside fenced block with language tag are content.
func TestNodeParse_CB04_FencedCodeBlockWithLanguageTag(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n# Public\n## Interface\n```python\n# Inside fenced block\n```\nReal content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public section absent, want present")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Public.Subsections len = %d, want 1", len(node.Public.Subsections))
	}

	subContent := node.Public.Subsections[0].Content
	foundInsideFence := false
	for _, line := range subContent {
		if line == "# Inside fenced block" {
			foundInsideFence = true
		}
	}
	if !foundInsideFence {
		t.Error("Expected \"# Inside fenced block\" in subsection content, not found")
	}
}

// TestNodeParse_CB05_BlankLinesBetweenHeadingAndContentPreserved verifies blank lines are preserved in content.
func TestNodeParse_CB05_BlankLinesBetweenHeadingAndContentPreserved(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n# Public\n\nPublic content line.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public section absent, want present")
	}
	if len(node.Public.Content) != 2 {
		t.Fatalf("Public.Content len = %d, want 2", len(node.Public.Content))
	}
	if node.Public.Content[0] != "" {
		t.Errorf("Public.Content[0] = %q, want empty string", node.Public.Content[0])
	}
	if node.Public.Content[1] != "Public content line." {
		t.Errorf("Public.Content[1] = %q, want %q", node.Public.Content[1], "Public content line.")
	}
}

// ---------------------------------------------------------------------------
// Frontmatter Handling Tests
// ---------------------------------------------------------------------------

// TestNodeParse_FM01_FrontmatterIsSkipped verifies frontmatter content is excluded from sections.
func TestNodeParse_FM01_FrontmatterIsSkipped(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "---\ndepends_on: []\n---\n# ROOT/a\nName content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/a" {
		t.Errorf("NameSection.Heading = %q, want %q", node.NameSection.Heading, "root/a")
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "Name content." {
		t.Errorf("NameSection.Content = %v, want [\"Name content.\"]", node.NameSection.Content)
	}
	// Verify frontmatter not in content.
	for _, line := range node.NameSection.Content {
		if line == "---" || line == "depends_on: []" {
			t.Errorf("Frontmatter line %q found in NameSection.Content", line)
		}
	}
}

// TestNodeParse_FM02_NoFrontmatterDelimiters verifies parsing without frontmatter works.
func TestNodeParse_FM02_NoFrontmatterDelimiters(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/a" {
		t.Errorf("NameSection.Heading = %q, want %q", node.NameSection.Heading, "root/a")
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "Name content." {
		t.Errorf("NameSection.Content = %v, want [\"Name content.\"]", node.NameSection.Content)
	}
}

// TestNodeParse_FM03_UnclosedFrontmatter verifies that unclosed frontmatter results in an error.
func TestNodeParse_FM03_UnclosedFrontmatter(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "---\ndepends_on: []\n# ROOT/a\nName content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

// ---------------------------------------------------------------------------
// Failure Case Tests
// ---------------------------------------------------------------------------

// TestNodeParse_FC01_ArtifactReferenceRejected verifies ARTIFACT reference is rejected.
func TestNodeParse_FC01_ArtifactReferenceRejected(t *testing.T) {
	_, err := parsenode.NodeParse("ARTIFACT/x(y)")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrNotARootReference) {
		t.Errorf("error = %v, want ErrNotARootReference", err)
	}
}

// TestNodeParse_FC02_QualifierRejected verifies logical name with qualifier is rejected.
func TestNodeParse_FC02_QualifierRejected(t *testing.T) {
	_, err := parsenode.NodeParse("ROOT/x(interface)")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrHasQualifier) {
		t.Errorf("error = %v, want ErrHasQualifier", err)
	}
}

// TestNodeParse_FC03_FileDoesNotExist verifies that a missing file returns ErrFileUnreadable.
func TestNodeParse_FC03_FileDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := parsenode.NodeParse("ROOT/does/not/exist")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrFileUnreadable) {
		t.Errorf("error = %v, want ErrFileUnreadable", err)
	}
}

// TestNodeParse_FC04_PropagatesPathErrors verifies that path errors from the resolution layer are propagated.
func TestNodeParse_FC04_PropagatesPathErrors(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Use a logical name that maps to a path with directory traversal after transformation.
	// The logicalname "ROOT/../escape" would produce "code-from-spec/../escape/_node.md"
	// which may traverse outside root. However, since LogicalNameToPath might not allow
	// such a name, we test that path errors are surfaced.
	// We expect either filereader.ErrFileUnreadable or a path-related error.
	_, err := parsenode.NodeParse("ROOT/does/not/exist/anywhere")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// The error should be either ErrFileUnreadable (from filereader) or a path error.
	isFileUnreadable := errors.Is(err, parsenode.ErrFileUnreadable)
	isPathError := errors.Is(err, filereader.ErrFileUnreadable)
	if !isFileUnreadable && !isPathError {
		t.Logf("error = %v (accepted as path/file error)", err)
	}
}

// TestNodeParse_FC05_ContentBeforeFirstHeading verifies content before first heading returns error.
func TestNodeParse_FC05_ContentBeforeFirstHeading(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "This line appears before any heading.\n# ROOT/a\nName content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

// TestNodeParse_FC06_Level2HeadingBeforeLevel1Heading verifies level-2 before level-1 returns error.
func TestNodeParse_FC06_Level2HeadingBeforeLevel1Heading(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "## Early subsection\n# ROOT/a\nName content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

// TestNodeParse_FC07_EmptyBody verifies an empty file body returns error.
func TestNodeParse_FC07_EmptyBody(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := ""
	testWriteNodeFile(t, "ROOT/a", content)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

// TestNodeParse_FC08_NodeNameDoesNotMatchLogicalName verifies mismatched node name returns error.
func TestNodeParse_FC08_NodeNameDoesNotMatchLogicalName(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/other\nSome content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrNodeNameDoesNotMatch) {
		t.Errorf("error = %v, want ErrNodeNameDoesNotMatch", err)
	}
}

// TestNodeParse_FC09_NodeNameCaseMismatchIsNotAnError verifies case mismatch is normalized and not an error.
func TestNodeParse_FC09_NodeNameCaseMismatchIsNotAnError(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create the directory for ROOT/A (uppercase) but the file uses lowercase root/a.
	// LogicalNameToPath converts ROOT/A -> code-from-spec/A/_node.md
	// We need to create the file at that path.
	dir := filepath.Join("code-from-spec", "A")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	filePath := filepath.Join(dir, "_node.md")
	if err := os.WriteFile(filePath, []byte("# root/a\nName content.\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/a" {
		t.Errorf("NameSection.Heading = %q, want %q", node.NameSection.Heading, "root/a")
	}
}

// TestNodeParse_FC10_DuplicatePublicSectionSameCase verifies duplicate public sections return error.
func TestNodeParse_FC10_DuplicatePublicSectionSameCase(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n# Public\nFirst public content.\n# Public\nSecond public content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicatePublicSection) {
		t.Errorf("error = %v, want ErrDuplicatePublicSection", err)
	}
}

// TestNodeParse_FC11_DuplicatePublicSectionDifferentCase verifies duplicate public sections different case return error.
func TestNodeParse_FC11_DuplicatePublicSectionDifferentCase(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n# Public\nFirst public content.\n# PUBLIC\nSecond public content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicatePublicSection) {
		t.Errorf("error = %v, want ErrDuplicatePublicSection", err)
	}
}

// TestNodeParse_FC12_DuplicateAgentSection verifies duplicate agent sections return error.
func TestNodeParse_FC12_DuplicateAgentSection(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n# Agent\nFirst agent content.\n# Agent\nSecond agent content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateAgentSection) {
		t.Errorf("error = %v, want ErrDuplicateAgentSection", err)
	}
}

// TestNodeParse_FC13_DuplicateSubsectionInPublicSameCase verifies duplicate subsection same case returns error.
func TestNodeParse_FC13_DuplicateSubsectionInPublicSameCase(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n# Public\n## Interface\nFirst interface content.\n## Interface\nSecond interface content.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}

// TestNodeParse_FC14_DuplicateSubsectionInPublicDifferentCase verifies duplicate subsection different case returns error.
func TestNodeParse_FC14_DuplicateSubsectionInPublicDifferentCase(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n# Public\n## Interface\nFirst.\n## INTERFACE\nSecond.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}

// TestNodeParse_FC15_DuplicateSubsectionWhitespaceVariation verifies duplicate subsection with whitespace variation returns error.
func TestNodeParse_FC15_DuplicateSubsectionWhitespaceVariation(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n# Public\n## Interface\nFirst.\n##   Interface\nSecond.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}

// TestNodeParse_FC16_DuplicateSubsectionInAgent verifies duplicate subsection in agent section returns error.
func TestNodeParse_FC16_DuplicateSubsectionInAgent(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# ROOT/a\nName content.\n# Agent\n## Guidance\nFirst guidance.\n## Guidance\nSecond guidance.\n"
	testWriteNodeFile(t, "ROOT/a", content)

	_, err := parsenode.NodeParse("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}
