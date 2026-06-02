// code-from-spec: ROOT/golang/tests/parsing/node_parsing@V92JLY1NQpEDP43EaUJcDSjbqNU
package parsenode_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
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

func testWriteNode(t *testing.T, logicalName string, content string) {
	t.Helper()
	path, err := nodeFilePath(logicalName)
	if err != nil {
		t.Fatalf("testWriteNode path: %v", err)
	}
	if err := os.MkdirAll(filepath_dir(path), 0755); err != nil {
		t.Fatalf("testWriteNode mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNode write: %v", err)
	}
}

func nodeFilePath(logicalName string) (string, error) {
	if logicalName == "ROOT" {
		return "code-from-spec/_node.md", nil
	}
	if len(logicalName) > 5 && logicalName[:5] == "ROOT/" {
		rest := logicalName[5:]
		return "code-from-spec/" + rest + "/_node.md", nil
	}
	return "", errors.New("not a ROOT reference")
}

func filepath_dir(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' {
			return p[:i]
		}
	}
	return "."
}

func testStrSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestNodeParse_MinimalNameSectionOnly(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.MkdirAll("code-from-spec/x", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/x/_node.md", []byte("# ROOT/x\nA simple node.\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/x" {
		t.Errorf("Heading = %q, want %q", node.NameSection.Heading, "root/x")
	}
	if node.NameSection.RawHeading != "# ROOT/x" {
		t.Errorf("RawHeading = %q, want %q", node.NameSection.RawHeading, "# ROOT/x")
	}
	if !testStrSliceEqual(node.NameSection.Content, []string{"A simple node."}) {
		t.Errorf("Content = %v, want [A simple node.]", node.NameSection.Content)
	}
	if len(node.NameSection.Subsections) != 0 {
		t.Errorf("Subsections = %v, want empty", node.NameSection.Subsections)
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

func TestNodeParse_FullNodeAllSectionTypes(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "---\noutput: some/output.md\n---\n# ROOT/payments/fees\nDescription of this node.\n# Public\n## Interface\nInterface content line.\n## Constraints\nConstraints content line.\n# Agent\nAgent content line.\n# Decisions\nDecisions content line.\n# Rationale\nRationale content line.\n"
	if err := os.MkdirAll("code-from-spec/payments/fees", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/payments/fees/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/payments/fees")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/payments/fees" {
		t.Errorf("Heading = %q, want %q", node.NameSection.Heading, "root/payments/fees")
	}
	if !testStrSliceEqual(node.NameSection.Content, []string{"Description of this node."}) {
		t.Errorf("Content = %v", node.NameSection.Content)
	}

	if node.Public == nil {
		t.Fatalf("Public = nil, want non-nil")
	}
	if len(node.Public.Content) != 0 {
		t.Errorf("Public.Content = %v, want empty", node.Public.Content)
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("Public.Subsections len = %d, want 2", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("Subsection[0].Heading = %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if !testStrSliceEqual(node.Public.Subsections[0].Content, []string{"Interface content line."}) {
		t.Errorf("Subsection[0].Content = %v", node.Public.Subsections[0].Content)
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("Subsection[1].Heading = %q, want %q", node.Public.Subsections[1].Heading, "constraints")
	}
	if !testStrSliceEqual(node.Public.Subsections[1].Content, []string{"Constraints content line."}) {
		t.Errorf("Subsection[1].Content = %v", node.Public.Subsections[1].Content)
	}

	if node.Agent == nil {
		t.Fatalf("Agent = nil, want non-nil")
	}
	if !testStrSliceEqual(node.Agent.Content, []string{"Agent content line."}) {
		t.Errorf("Agent.Content = %v", node.Agent.Content)
	}

	if len(node.Private) != 2 {
		t.Fatalf("Private len = %d, want 2", len(node.Private))
	}
	if node.Private[0].Heading != "decisions" {
		t.Errorf("Private[0].Heading = %q, want %q", node.Private[0].Heading, "decisions")
	}
	if !testStrSliceEqual(node.Private[0].Content, []string{"Decisions content line."}) {
		t.Errorf("Private[0].Content = %v", node.Private[0].Content)
	}
	if node.Private[1].Heading != "rationale" {
		t.Errorf("Private[1].Heading = %q, want %q", node.Private[1].Heading, "rationale")
	}
	if !testStrSliceEqual(node.Private[1].Content, []string{"Rationale content line."}) {
		t.Errorf("Private[1].Content = %v", node.Private[1].Content)
	}
}

func TestNodeParse_NoPublicSection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/decisions\nDescription.\n# Rationale\nRationale content.\n"
	if err := os.MkdirAll("code-from-spec/decisions", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/decisions/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

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
	if !testStrSliceEqual(node.Private[0].Content, []string{"Rationale content."}) {
		t.Errorf("Private[0].Content = %v", node.Private[0].Content)
	}
}

func TestNodeParse_PublicWithContentBeforeFirstSubsection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/a\nNode description.\n# Public\nPreamble line one.\nPreamble line two.\n## Interface\nInterface content.\n"
	if err := os.MkdirAll("code-from-spec/a", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/a/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatalf("Public = nil")
	}
	if !testStrSliceEqual(node.Public.Content, []string{"Preamble line one.", "Preamble line two."}) {
		t.Errorf("Public.Content = %v", node.Public.Content)
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Public.Subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("Subsection[0].Heading = %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if !testStrSliceEqual(node.Public.Subsections[0].Content, []string{"Interface content."}) {
		t.Errorf("Subsection[0].Content = %v", node.Public.Subsections[0].Content)
	}
}

func TestNodeParse_PublicWithNoContentOrSubsections(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/b\nNode description.\n# Public\n# Agent\nAgent content.\n"
	if err := os.MkdirAll("code-from-spec/b", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/b/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatalf("Public = nil")
	}
	if len(node.Public.Content) != 0 {
		t.Errorf("Public.Content = %v, want empty", node.Public.Content)
	}
	if len(node.Public.Subsections) != 0 {
		t.Errorf("Public.Subsections = %v, want empty", node.Public.Subsections)
	}

	if node.Agent == nil {
		t.Fatalf("Agent = nil")
	}
	if !testStrSliceEqual(node.Agent.Content, []string{"Agent content."}) {
		t.Errorf("Agent.Content = %v", node.Agent.Content)
	}
}

func TestNodeParse_AgentSectionWithSubsections(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/c\nNode description.\n# Agent\nAgent preamble line.\n## Implementation guidance\nImplementation content.\n## Contracts\nContracts content.\n"
	if err := os.MkdirAll("code-from-spec/c", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/c/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Agent == nil {
		t.Fatalf("Agent = nil")
	}
	if !testStrSliceEqual(node.Agent.Content, []string{"Agent preamble line."}) {
		t.Errorf("Agent.Content = %v", node.Agent.Content)
	}
	if node.Agent.RawHeading != "# Agent" {
		t.Errorf("Agent.RawHeading = %q, want %q", node.Agent.RawHeading, "# Agent")
	}
	if len(node.Agent.Subsections) != 2 {
		t.Fatalf("Agent.Subsections len = %d, want 2", len(node.Agent.Subsections))
	}
	if node.Agent.Subsections[0].Heading != "implementation guidance" {
		t.Errorf("Subsection[0].Heading = %q, want %q", node.Agent.Subsections[0].Heading, "implementation guidance")
	}
	if !testStrSliceEqual(node.Agent.Subsections[0].Content, []string{"Implementation content."}) {
		t.Errorf("Subsection[0].Content = %v", node.Agent.Subsections[0].Content)
	}
	if node.Agent.Subsections[1].Heading != "contracts" {
		t.Errorf("Subsection[1].Heading = %q, want %q", node.Agent.Subsections[1].Heading, "contracts")
	}
	if !testStrSliceEqual(node.Agent.Subsections[1].Content, []string{"Contracts content."}) {
		t.Errorf("Subsection[1].Content = %v", node.Agent.Subsections[1].Content)
	}
}

func TestNodeParse_PrivateSectionsPreserveOrder(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/d\nNode description.\n# TODO\nTodo content.\n# Decisions\nDecisions content.\n# Rationale\nRationale content.\n"
	if err := os.MkdirAll("code-from-spec/d", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/d/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/d")
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

func TestNodeParse_ContentIsRawMarkdown(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/e\nNode description.\n# Public\n## Interface\n### Sub-heading\n**bold text**\n```go\nsome code\n```\n"
	if err := os.MkdirAll("code-from-spec/e", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/e/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/e")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatalf("Public = nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Subsections len = %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	want := []string{"### Sub-heading", "**bold text**", "```go", "some code", "```"}
	if !testStrSliceEqual(sub.Content, want) {
		t.Errorf("Content = %v, want %v", sub.Content, want)
	}
}

func TestNodeParse_CaseInsensitivePublicDetection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/f\nDescription.\n# PUBLIC\nPublic content.\n"
	if err := os.MkdirAll("code-from-spec/f", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/f/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/f")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatalf("Public = nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("Public.Heading = %q, want %q", node.Public.Heading, "public")
	}
}

func TestNodeParse_PublicWithMixedCaseAndExtraWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/g\nDescription.\n#   PuBLiC\nPublic content.\n"
	if err := os.MkdirAll("code-from-spec/g", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/g/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/g")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatalf("Public = nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("Public.Heading = %q, want %q", node.Public.Heading, "public")
	}
}

func TestNodeParse_NodeNameWithVariedWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "#   ROOT/e\nDescription.\n"
	if err := os.MkdirAll("code-from-spec/e", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/e/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/e")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/e" {
		t.Errorf("Heading = %q, want %q", node.NameSection.Heading, "root/e")
	}
}

func TestNodeParse_SubsectionHeadingsAreNormalized(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/h\nDescription.\n# Public\n##   Interface\nInterface content.\n## CONSTRAINTS\nConstraints content.\n"
	if err := os.MkdirAll("code-from-spec/h", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/h/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/h")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatalf("Public = nil")
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("Subsections len = %d, want 2", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("Subsections[0].Heading = %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("Subsections[1].Heading = %q, want %q", node.Public.Subsections[1].Heading, "constraints")
	}
}

func TestNodeParse_ClosingHashesAreStripped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/i\nDescription.\n# Public\n## Interface ##\nInterface content.\n"
	if err := os.MkdirAll("code-from-spec/i", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/i/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/i")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatalf("Public = nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("Heading = %q, want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if node.Public.Subsections[0].RawHeading != "## Interface ##" {
		t.Errorf("RawHeading = %q, want %q", node.Public.Subsections[0].RawHeading, "## Interface ##")
	}
}

func TestNodeParse_RawHeadingPreservesOriginalLine(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/j\nDescription.\n# Public\n## Interface\nInterface content.\n"
	if err := os.MkdirAll("code-from-spec/j", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/j/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/j")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatalf("Public = nil")
	}
	if node.Public.RawHeading != "# Public" {
		t.Errorf("Public.RawHeading = %q, want %q", node.Public.RawHeading, "# Public")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].RawHeading != "## Interface" {
		t.Errorf("Subsection.RawHeading = %q, want %q", node.Public.Subsections[0].RawHeading, "## Interface")
	}
}

func TestNodeParse_RawHeadingPreservesCase(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/k\nDescription.\n# PUBLIC\nPublic content.\n"
	if err := os.MkdirAll("code-from-spec/k", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/k/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatalf("Public = nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("Heading = %q, want %q", node.Public.Heading, "public")
	}
	if node.Public.RawHeading != "# PUBLIC" {
		t.Errorf("RawHeading = %q, want %q", node.Public.RawHeading, "# PUBLIC")
	}
}

func TestNodeParse_RawHeadingPreservesClosingHashes(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/l\nDescription.\n# Public\n## Foo ##\nFoo content.\n"
	if err := os.MkdirAll("code-from-spec/l", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/l/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/l")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatalf("Public = nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Subsections len = %d, want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "foo" {
		t.Errorf("Heading = %q, want %q", node.Public.Subsections[0].Heading, "foo")
	}
	if node.Public.Subsections[0].RawHeading != "## Foo ##" {
		t.Errorf("RawHeading = %q, want %q", node.Public.Subsections[0].RawHeading, "## Foo ##")
	}
}

func TestNodeParse_RawHeadingPreservesExtraWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/m\nDescription.\n#   Public\nPublic content.\n"
	if err := os.MkdirAll("code-from-spec/m", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/m/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatalf("Public = nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("Heading = %q, want %q", node.Public.Heading, "public")
	}
	if node.Public.RawHeading != "#   Public" {
		t.Errorf("RawHeading = %q, want %q", node.Public.RawHeading, "#   Public")
	}
}

func TestNodeParse_Level3AndDeeperHeadingsAreContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/n\nDescription.\n# Public\n## Interface\n### Sub-heading\n#### Deep heading\nContent line.\n"
	if err := os.MkdirAll("code-from-spec/n", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/n/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatalf("Public = nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Subsections len = %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	want := []string{"### Sub-heading", "#### Deep heading", "Content line."}
	if !testStrSliceEqual(sub.Content, want) {
		t.Errorf("Content = %v, want %v", sub.Content, want)
	}
}

func TestNodeParse_FencedCodeBlockWithHeadingLikeContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/o\nDescription.\n# Public\n## Interface\n```\n# looks like heading\n## also looks like heading\n```\n"
	if err := os.MkdirAll("code-from-spec/o", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/o/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatalf("Public = nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Subsections len = %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	want := []string{"```", "# looks like heading", "## also looks like heading", "```"}
	if !testStrSliceEqual(sub.Content, want) {
		t.Errorf("Content = %v, want %v", sub.Content, want)
	}
}

func TestNodeParse_FencedCodeBlockWithTildeFence(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/p\nDescription.\n# Public\n## Interface\n~~~\n# looks like level-1 heading\n~~~\n"
	if err := os.MkdirAll("code-from-spec/p", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/p/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatalf("Public = nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Subsections len = %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	want := []string{"~~~", "# looks like level-1 heading", "~~~"}
	if !testStrSliceEqual(sub.Content, want) {
		t.Errorf("Content = %v, want %v", sub.Content, want)
	}
}

func TestNodeParse_FencedCodeBlockWithLanguageTag(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/q\nDescription.\n# Public\n## Interface\n```go\n# looks like level-1 heading\n```\n"
	if err := os.MkdirAll("code-from-spec/q", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/q/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/q")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatalf("Public = nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("Subsections len = %d, want 1", len(node.Public.Subsections))
	}
	sub := node.Public.Subsections[0]
	want := []string{"```go", "# looks like level-1 heading", "```"}
	if !testStrSliceEqual(sub.Content, want) {
		t.Errorf("Content = %v, want %v", sub.Content, want)
	}
}

func TestNodeParse_BlankLinesBetweenHeadingAndContentPreserved(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/r\nDescription.\n# Public\n\nContent line.\n"
	if err := os.MkdirAll("code-from-spec/r", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/r/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/r")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatalf("Public = nil")
	}
	if !testStrSliceEqual(node.Public.Content, []string{"", "Content line."}) {
		t.Errorf("Public.Content = %v, want ['', 'Content line.']", node.Public.Content)
	}
}

func TestNodeParse_FrontmatterIsSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "---\noutput: some/path.md\n---\n# ROOT/s\nDescription.\n"
	if err := os.MkdirAll("code-from-spec/s", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/s/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/s" {
		t.Errorf("Heading = %q, want %q", node.NameSection.Heading, "root/s")
	}
}

func TestNodeParse_NoFrontmatterDelimiters(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/t\nDescription.\n"
	if err := os.MkdirAll("code-from-spec/t", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/t/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	node, err := parsenode.NodeParse("ROOT/t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/t" {
		t.Errorf("Heading = %q, want %q", node.NameSection.Heading, "root/t")
	}
}

func TestNodeParse_UnclosedFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "---\noutput: some/path.md\n# ROOT/u\nDescription.\n"
	if err := os.MkdirAll("code-from-spec/u", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/u/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := parsenode.NodeParse("ROOT/u")
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

func TestNodeParse_ARTIFACTReferenceRejected(t *testing.T) {
	_, err := parsenode.NodeParse("ARTIFACT/x")
	if !errors.Is(err, parsenode.ErrNotARootReference) {
		t.Errorf("error = %v, want ErrNotARootReference", err)
	}
}

func TestNodeParse_QualifierRejected(t *testing.T) {
	_, err := parsenode.NodeParse("ROOT/x(interface)")
	if !errors.Is(err, parsenode.ErrHasQualifier) {
		t.Errorf("error = %v, want ErrHasQualifier", err)
	}
}

func TestNodeParse_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := parsenode.NodeParse("ROOT/nonexistent/node")
	if !errors.Is(err, parsenode.ErrFileUnreadable) {
		t.Errorf("error = %v, want ErrFileUnreadable", err)
	}
}

func TestNodeParse_PropagatesPathErrors(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := parsenode.NodeParse("ROOT/../../outside")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) && !errors.Is(err, parsenode.ErrFileUnreadable) {
		t.Errorf("error = %v, want path traversal or file unreadable error", err)
	}
}

func TestNodeParse_ContentBeforeFirstHeading(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "This line appears before any heading.\n# ROOT/v\nDescription.\n"
	if err := os.MkdirAll("code-from-spec/v", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/v/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := parsenode.NodeParse("ROOT/v")
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

func TestNodeParse_Level2HeadingBeforeAnyLevel1Heading(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "## Some subsection\nDescription.\n"
	if err := os.MkdirAll("code-from-spec/w", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/w/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := parsenode.NodeParse("ROOT/w")
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

func TestNodeParse_EmptyBody(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.MkdirAll("code-from-spec/empty", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/empty/_node.md", []byte(""), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := parsenode.NodeParse("ROOT/empty")
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

func TestNodeParse_NodeNameDoesNotMatch(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/actual/name\nDescription.\n"
	if err := os.MkdirAll("code-from-spec/different/name", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/different/name/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := parsenode.NodeParse("ROOT/different/name")
	if !errors.Is(err, parsenode.ErrNodeNameDoesNotMatch) {
		t.Errorf("error = %v, want ErrNodeNameDoesNotMatch", err)
	}
}

func TestNodeParse_NodeNameCaseMismatchIsNotError(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# root/x\nDescription.\n"
	if err := os.MkdirAll("code-from-spec/X", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/X/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := parsenode.NodeParse("ROOT/X")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNodeParse_DuplicatePublicSectionSameCase(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/dup\nDescription.\n# Public\nFirst public content.\n# Public\nSecond public content.\n"
	if err := os.MkdirAll("code-from-spec/dup", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/dup/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := parsenode.NodeParse("ROOT/dup")
	if !errors.Is(err, parsenode.ErrDuplicatePublicSection) {
		t.Errorf("error = %v, want ErrDuplicatePublicSection", err)
	}
}

func TestNodeParse_DuplicatePublicSectionDifferentCase(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/dup2\nDescription.\n# Public\nFirst public content.\n# PUBLIC\nSecond public content.\n"
	if err := os.MkdirAll("code-from-spec/dup2", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/dup2/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := parsenode.NodeParse("ROOT/dup2")
	if !errors.Is(err, parsenode.ErrDuplicatePublicSection) {
		t.Errorf("error = %v, want ErrDuplicatePublicSection", err)
	}
}

func TestNodeParse_DuplicateAgentSection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/dup3\nDescription.\n# Agent\nFirst agent content.\n# Agent\nSecond agent content.\n"
	if err := os.MkdirAll("code-from-spec/dup3", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/dup3/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := parsenode.NodeParse("ROOT/dup3")
	if !errors.Is(err, parsenode.ErrDuplicateAgentSection) {
		t.Errorf("error = %v, want ErrDuplicateAgentSection", err)
	}
}

func TestNodeParse_DuplicateSubsectionInPublicSameCase(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/dup4\nDescription.\n# Public\n## Interface\nContent.\n## Interface\nMore content.\n"
	if err := os.MkdirAll("code-from-spec/dup4", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/dup4/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := parsenode.NodeParse("ROOT/dup4")
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}

func TestNodeParse_DuplicateSubsectionInPublicDifferentCase(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/dup5\nDescription.\n# Public\n## Interface\nContent.\n## INTERFACE\nMore content.\n"
	if err := os.MkdirAll("code-from-spec/dup5", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/dup5/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := parsenode.NodeParse("ROOT/dup5")
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}

func TestNodeParse_DuplicateSubsectionInPublicWhitespaceVariation(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/dup6\nDescription.\n# Public\n## Interface\nContent.\n##   Interface\nMore content.\n"
	if err := os.MkdirAll("code-from-spec/dup6", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/dup6/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := parsenode.NodeParse("ROOT/dup6")
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}

func TestNodeParse_DuplicateSubsectionInAgent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# ROOT/dup7\nDescription.\n# Agent\n## Rules\nContent.\n## Rules\nMore content.\n"
	if err := os.MkdirAll("code-from-spec/dup7", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/dup7/_node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := parsenode.NodeParse("ROOT/dup7")
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}
