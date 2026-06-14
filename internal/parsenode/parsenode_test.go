// code-from-spec: ROOT/golang/tests/parsing/node_parsing@svBa2y9uK793i2dRJGqNEdMZoyM
package parsenode_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
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

func testCreateNodeFile(t *testing.T, logicalName string, content string) {
	t.Helper()
	parts := []string{"code-from-spec"}
	if logicalName == "SPEC" {
		parts = append(parts, "_node.md")
	} else {
		suffix := logicalName[len("SPEC/"):]
		for _, seg := range splitPath(suffix) {
			parts = append(parts, seg)
		}
		parts = append(parts, "_node.md")
	}
	path := filepath.Join(parts...)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testCreateNodeFile MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testCreateNodeFile WriteFile: %v", err)
	}
}

func splitPath(p string) []string {
	var parts []string
	for {
		dir, file := filepath.Split(p)
		if file == "" {
			break
		}
		parts = append([]string{file}, parts...)
		p = filepath.Clean(dir)
		if p == "." {
			break
		}
	}
	return parts
}

func TestNodeParse_HP01_MinimalNode(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testCreateNodeFile(t, "SPEC/x", "# SPEC/x\nA simple node.\n")

	node, err := parsenode.NodeParse("SPEC/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "spec/x" {
		t.Errorf("heading = %q, want %q", node.NameSection.Heading, "spec/x")
	}
	if node.NameSection.RawHeading != "# SPEC/x" {
		t.Errorf("raw_heading = %q, want %q", node.NameSection.RawHeading, "# SPEC/x")
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "A simple node." {
		t.Errorf("content = %v, want [A simple node.]", node.NameSection.Content)
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
	if node.Private != nil {
		t.Errorf("private = %v, want nil", node.Private)
	}
}

func TestNodeParse_HP02_FullNode(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "---\noutput: some/path\n---\n# SPEC/payments/fees\nFee description line.\n# Public\n## Interface\nInterface content line.\n## Constraints\nConstraints content line.\n# Agent\nAgent content line.\n# Private\n## Decisions\nDecisions content line.\n## Rationale\nRationale content line.\n"
	testCreateNodeFile(t, "SPEC/payments/fees", body)

	node, err := parsenode.NodeParse("SPEC/payments/fees")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "spec/payments/fees" {
		t.Errorf("heading = %q, want %q", node.NameSection.Heading, "spec/payments/fees")
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "Fee description line." {
		t.Errorf("name content = %v", node.NameSection.Content)
	}
	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Content) != 0 {
		t.Errorf("public.content = %v, want empty", node.Public.Content)
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("public.subsections len = %d, want 2", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("subsection[0].heading = %q, want interface", node.Public.Subsections[0].Heading)
	}
	if len(node.Public.Subsections[0].Content) != 1 || node.Public.Subsections[0].Content[0] != "Interface content line." {
		t.Errorf("subsection[0].content = %v", node.Public.Subsections[0].Content)
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("subsection[1].heading = %q, want constraints", node.Public.Subsections[1].Heading)
	}
	if len(node.Public.Subsections[1].Content) != 1 || node.Public.Subsections[1].Content[0] != "Constraints content line." {
		t.Errorf("subsection[1].content = %v", node.Public.Subsections[1].Content)
	}
	if node.Agent == nil {
		t.Fatal("agent = nil, want present")
	}
	if len(node.Agent.Content) != 1 || node.Agent.Content[0] != "Agent content line." {
		t.Errorf("agent.content = %v", node.Agent.Content)
	}
	if node.Private == nil {
		t.Fatal("private = nil, want present")
	}
	if len(node.Private.Subsections) != 2 {
		t.Fatalf("private.subsections len = %d, want 2", len(node.Private.Subsections))
	}
	if node.Private.Subsections[0].Heading != "decisions" {
		t.Errorf("private.subsections[0].heading = %q, want decisions", node.Private.Subsections[0].Heading)
	}
	if node.Private.Subsections[1].Heading != "rationale" {
		t.Errorf("private.subsections[1].heading = %q, want rationale", node.Private.Subsections[1].Heading)
	}
}

func TestNodeParse_HP03_NoPublicSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/decisions\nContent line.\n# Private\n## Rationale\nRationale content.\n"
	testCreateNodeFile(t, "SPEC/decisions", body)

	node, err := parsenode.NodeParse("SPEC/decisions")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public != nil {
		t.Errorf("public = %v, want nil", node.Public)
	}
	if node.Agent != nil {
		t.Errorf("agent = %v, want nil", node.Agent)
	}
	if node.Private == nil {
		t.Fatal("private = nil, want present")
	}
	if len(node.Private.Subsections) != 1 {
		t.Fatalf("private.subsections len = %d, want 1", len(node.Private.Subsections))
	}
	if node.Private.Subsections[0].Heading != "rationale" {
		t.Errorf("private.subsections[0].heading = %q, want rationale", node.Private.Subsections[0].Heading)
	}
	if len(node.Private.Subsections[0].Content) != 1 || node.Private.Subsections[0].Content[0] != "Rationale content." {
		t.Errorf("private.subsections[0].content = %v", node.Private.Subsections[0].Content)
	}
}

func TestNodeParse_HP04_PublicContentBeforeSubsection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Public\nPreamble line one.\nPreamble line two.\n## Interface\nInterface content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	node, err := parsenode.NodeParse("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Content) != 2 || node.Public.Content[0] != "Preamble line one." || node.Public.Content[1] != "Preamble line two." {
		t.Errorf("public.content = %v", node.Public.Content)
	}
	if len(node.Public.Subsections) != 1 || node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("public.subsections = %v", node.Public.Subsections)
	}
	if len(node.Public.Subsections[0].Content) != 1 || node.Public.Subsections[0].Content[0] != "Interface content." {
		t.Errorf("public.subsections[0].content = %v", node.Public.Subsections[0].Content)
	}
}

func TestNodeParse_HP05_PublicEmptyNoSubsections(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Public\n# Agent\nAgent content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	node, err := parsenode.NodeParse("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Content) != 0 {
		t.Errorf("public.content = %v, want empty", node.Public.Content)
	}
	if len(node.Public.Subsections) != 0 {
		t.Errorf("public.subsections = %v, want empty", node.Public.Subsections)
	}
	if node.Agent == nil {
		t.Fatal("agent = nil, want present")
	}
	if len(node.Agent.Content) != 1 || node.Agent.Content[0] != "Agent content." {
		t.Errorf("agent.content = %v", node.Agent.Content)
	}
}

func TestNodeParse_HP06_AgentWithSubsections(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Agent\nAgent preamble.\n## Implementation guidance\nImplementation content.\n## Contracts\nContracts content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	node, err := parsenode.NodeParse("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Agent == nil {
		t.Fatal("agent = nil, want present")
	}
	if len(node.Agent.Content) != 1 || node.Agent.Content[0] != "Agent preamble." {
		t.Errorf("agent.content = %v", node.Agent.Content)
	}
	if node.Agent.RawHeading != "# Agent" {
		t.Errorf("agent.raw_heading = %q, want %q", node.Agent.RawHeading, "# Agent")
	}
	if len(node.Agent.Subsections) != 2 {
		t.Fatalf("agent.subsections len = %d, want 2", len(node.Agent.Subsections))
	}
	if node.Agent.Subsections[0].Heading != "implementation guidance" {
		t.Errorf("agent.subsections[0].heading = %q, want implementation guidance", node.Agent.Subsections[0].Heading)
	}
	if len(node.Agent.Subsections[0].Content) != 1 || node.Agent.Subsections[0].Content[0] != "Implementation content." {
		t.Errorf("agent.subsections[0].content = %v", node.Agent.Subsections[0].Content)
	}
	if node.Agent.Subsections[1].Heading != "contracts" {
		t.Errorf("agent.subsections[1].heading = %q, want contracts", node.Agent.Subsections[1].Heading)
	}
	if len(node.Agent.Subsections[1].Content) != 1 || node.Agent.Subsections[1].Content[0] != "Contracts content." {
		t.Errorf("agent.subsections[1].content = %v", node.Agent.Subsections[1].Content)
	}
}

func TestNodeParse_HP07_PrivateWithSubsections(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Private\n## TODO\nTodo content.\n## Decisions\nDecisions content.\n## Rationale\nRationale content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	node, err := parsenode.NodeParse("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Private == nil {
		t.Fatal("private = nil, want present")
	}
	if len(node.Private.Subsections) != 3 {
		t.Fatalf("private.subsections len = %d, want 3", len(node.Private.Subsections))
	}
	if node.Private.Subsections[0].Heading != "todo" {
		t.Errorf("private.subsections[0].heading = %q, want todo", node.Private.Subsections[0].Heading)
	}
	if node.Private.Subsections[1].Heading != "decisions" {
		t.Errorf("private.subsections[1].heading = %q, want decisions", node.Private.Subsections[1].Heading)
	}
	if node.Private.Subsections[2].Heading != "rationale" {
		t.Errorf("private.subsections[2].heading = %q, want rationale", node.Private.Subsections[2].Heading)
	}
}

func TestNodeParse_HP08_ContentIsRawMarkdown(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Public\n## Summary\n### A level-3 heading\n**Bold text here**\n```python\nx = 1\n```\n"
	testCreateNodeFile(t, "SPEC/a", body)

	node, err := parsenode.NodeParse("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Subsections) != 1 || node.Public.Subsections[0].Heading != "summary" {
		t.Fatalf("public.subsections = %v", node.Public.Subsections)
	}
	want := []string{"### A level-3 heading", "**Bold text here**", "```python", "x = 1", "```"}
	got := node.Public.Subsections[0].Content
	if len(got) != len(want) {
		t.Fatalf("subsection[0].content len = %d, want %d: %v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("subsection[0].content[%d] = %q, want %q", i, got[i], w)
		}
	}
}

func TestNodeParse_HN01_CaseInsensitivePublic(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# PUBLIC\nPublic content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	node, err := parsenode.NodeParse("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if node.Public.Heading != "public" {
		t.Errorf("public.heading = %q, want public", node.Public.Heading)
	}
}

func TestNodeParse_HN02_PublicMixedCaseWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n#   PuBLiC\nPublic content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	node, err := parsenode.NodeParse("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if node.Public.Heading != "public" {
		t.Errorf("public.heading = %q, want public", node.Public.Heading)
	}
}

func TestNodeParse_HN03_NodeNameVariedWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "#   SPEC/e\nName content.\n"
	testCreateNodeFile(t, "SPEC/e", body)

	node, err := parsenode.NodeParse("SPEC/e")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "spec/e" {
		t.Errorf("heading = %q, want spec/e", node.NameSection.Heading)
	}
}

func TestNodeParse_HN04_RootXHeadingDoesNotMatchSpecX(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# ROOT/x\nContent.\n"
	testCreateNodeFile(t, "SPEC/x", body)

	_, err := parsenode.NodeParse("SPEC/x")
	if !errors.Is(err, parsenode.ErrNodeNameDoesNotMatch) {
		t.Errorf("error = %v, want ErrNodeNameDoesNotMatch", err)
	}
}

func TestNodeParse_HN05_SubsectionHeadingsNormalized(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Public\n##   Interface\nInterface content.\n## CONSTRAINTS\nConstraints content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	node, err := parsenode.NodeParse("SPEC/a")
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
		t.Errorf("subsection[0].heading = %q, want interface", node.Public.Subsections[0].Heading)
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("subsection[1].heading = %q, want constraints", node.Public.Subsections[1].Heading)
	}
}

func TestNodeParse_HN06_ClosingHashesStripped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Public\n## Interface ##\nInterface content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	node, err := parsenode.NodeParse("SPEC/a")
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
		t.Errorf("subsection.heading = %q, want interface", node.Public.Subsections[0].Heading)
	}
	if node.Public.Subsections[0].RawHeading != "## Interface ##" {
		t.Errorf("subsection.raw_heading = %q, want %q", node.Public.Subsections[0].RawHeading, "## Interface ##")
	}
}

func TestNodeParse_RH01_RawHeadingPreservesOriginalLine(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Public\n## Interface\nInterface content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	node, err := parsenode.NodeParse("SPEC/a")
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
		t.Errorf("subsection.raw_heading = %q, want %q", node.Public.Subsections[0].RawHeading, "## Interface")
	}
}

func TestNodeParse_RH02_RawHeadingPreservesCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# PUBLIC\nPublic content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	node, err := parsenode.NodeParse("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if node.Public.Heading != "public" {
		t.Errorf("public.heading = %q, want public", node.Public.Heading)
	}
	if node.Public.RawHeading != "# PUBLIC" {
		t.Errorf("public.raw_heading = %q, want %q", node.Public.RawHeading, "# PUBLIC")
	}
}

func TestNodeParse_RH03_RawHeadingPreservesClosingHashes(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Public\n## Foo ##\nFoo content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	node, err := parsenode.NodeParse("SPEC/a")
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
		t.Errorf("subsection.heading = %q, want foo", node.Public.Subsections[0].Heading)
	}
	if node.Public.Subsections[0].RawHeading != "## Foo ##" {
		t.Errorf("subsection.raw_heading = %q, want %q", node.Public.Subsections[0].RawHeading, "## Foo ##")
	}
}

func TestNodeParse_RH04_RawHeadingPreservesExtraWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n#   Public\nPublic content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	node, err := parsenode.NodeParse("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if node.Public.Heading != "public" {
		t.Errorf("public.heading = %q, want public", node.Public.Heading)
	}
	if node.Public.RawHeading != "#   Public" {
		t.Errorf("public.raw_heading = %q, want %q", node.Public.RawHeading, "#   Public")
	}
}

func TestNodeParse_CB01_Level3AndDeeperAreContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Public\n## Summary\n### A deeper heading\n#### Even deeper\n"
	testCreateNodeFile(t, "SPEC/a", body)

	node, err := parsenode.NodeParse("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Subsections) != 1 || node.Public.Subsections[0].Heading != "summary" {
		t.Fatalf("public.subsections = %v", node.Public.Subsections)
	}
	want := []string{"### A deeper heading", "#### Even deeper"}
	got := node.Public.Subsections[0].Content
	if len(got) != len(want) {
		t.Fatalf("subsection[0].content = %v, want %v", got, want)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("subsection[0].content[%d] = %q, want %q", i, got[i], w)
		}
	}
}

func TestNodeParse_CB02_FencedCodeBlockBacktick(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Public\n## Interface\n```\n# looks like heading\n## also looks like heading\n```\n"
	testCreateNodeFile(t, "SPEC/a", body)

	node, err := parsenode.NodeParse("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Subsections) != 1 || node.Public.Subsections[0].Heading != "interface" {
		t.Fatalf("public.subsections = %v", node.Public.Subsections)
	}
	content := node.Public.Subsections[0].Content
	found1, found2 := false, false
	for _, line := range content {
		if line == "# looks like heading" {
			found1 = true
		}
		if line == "## also looks like heading" {
			found2 = true
		}
	}
	if !found1 {
		t.Errorf("expected '# looks like heading' in content %v", content)
	}
	if !found2 {
		t.Errorf("expected '## also looks like heading' in content %v", content)
	}
}

func TestNodeParse_CB03_FencedCodeBlockTilde(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Public\n## Interface\n~~~\n# This looks like a heading\n~~~\n"
	testCreateNodeFile(t, "SPEC/a", body)

	node, err := parsenode.NodeParse("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Subsections) != 1 || node.Public.Subsections[0].Heading != "interface" {
		t.Fatalf("public.subsections = %v", node.Public.Subsections)
	}
	content := node.Public.Subsections[0].Content
	found := false
	for _, line := range content {
		if line == "# This looks like a heading" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected '# This looks like a heading' in content %v", content)
	}
}

func TestNodeParse_CB04_FencedCodeBlockWithLanguageTag(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Public\n## Interface\n```python\n# This looks like a heading\n```\n"
	testCreateNodeFile(t, "SPEC/a", body)

	node, err := parsenode.NodeParse("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Subsections) != 1 || node.Public.Subsections[0].Heading != "interface" {
		t.Fatalf("public.subsections = %v", node.Public.Subsections)
	}
	content := node.Public.Subsections[0].Content
	found := false
	for _, line := range content {
		if line == "# This looks like a heading" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected '# This looks like a heading' in content %v", content)
	}
}

func TestNodeParse_CB05_BlankLinesBetweenHeadingAndContentPreserved(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Public\n\nPublic content line.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	node, err := parsenode.NodeParse("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Public == nil {
		t.Fatal("public = nil, want present")
	}
	if len(node.Public.Content) != 2 {
		t.Fatalf("public.content = %v, want [\"\", \"Public content line.\"]", node.Public.Content)
	}
	if node.Public.Content[0] != "" {
		t.Errorf("public.content[0] = %q, want empty string", node.Public.Content[0])
	}
	if node.Public.Content[1] != "Public content line." {
		t.Errorf("public.content[1] = %q, want %q", node.Public.Content[1], "Public content line.")
	}
}

func TestNodeParse_FM01_FrontmatterIsSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "---\noutput: some/path\ndepends_on: []\n---\n# SPEC/a\nName content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	node, err := parsenode.NodeParse("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "spec/a" {
		t.Errorf("heading = %q, want spec/a", node.NameSection.Heading)
	}
	if len(node.NameSection.Content) != 1 || node.NameSection.Content[0] != "Name content." {
		t.Errorf("content = %v", node.NameSection.Content)
	}
}

func TestNodeParse_FM02_NoFrontmatterDelimiters(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	node, err := parsenode.NodeParse("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "spec/a" {
		t.Errorf("heading = %q, want spec/a", node.NameSection.Heading)
	}
}

func TestNodeParse_FM03_UnclosedFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "---\noutput: some/path\n# SPEC/a\nName content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	_, err := parsenode.NodeParse("SPEC/a")
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

func TestNodeParse_FC01_ArtifactRejected(t *testing.T) {
	_, err := parsenode.NodeParse("ARTIFACT/x")
	if !errors.Is(err, parsenode.ErrNotASpecReference) {
		t.Errorf("error = %v, want ErrNotASpecReference", err)
	}
}

func TestNodeParse_FC02_ExternalRejected(t *testing.T) {
	_, err := parsenode.NodeParse("EXTERNAL/x")
	if !errors.Is(err, parsenode.ErrNotASpecReference) {
		t.Errorf("error = %v, want ErrNotASpecReference", err)
	}
}

func TestNodeParse_FC03_QualifierRejected(t *testing.T) {
	_, err := parsenode.NodeParse("SPEC/x(interface)")
	if !errors.Is(err, parsenode.ErrHasQualifier) {
		t.Errorf("error = %v, want ErrHasQualifier", err)
	}
}

func TestNodeParse_FC04_FileDoesNotExist(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := parsenode.NodeParse("SPEC/nonexistent/node")
	if !errors.Is(err, parsenode.ErrFileUnreadable) {
		t.Errorf("error = %v, want ErrFileUnreadable", err)
	}
}

func TestNodeParse_FC05_PropagatesPathErrors(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := parsenode.NodeParse("SPEC/../../escape")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNodeParse_FC06_ContentBeforeFirstHeading(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "This is non-blank content.\n# SPEC/a\nName content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	_, err := parsenode.NodeParse("SPEC/a")
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

func TestNodeParse_FC07_Level2BeforeLevel1(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "## A subsection\n# SPEC/a\nName content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	_, err := parsenode.NodeParse("SPEC/a")
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

func TestNodeParse_FC08_EmptyBody(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testCreateNodeFile(t, "SPEC/a", "")

	_, err := parsenode.NodeParse("SPEC/a")
	if !errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
		t.Errorf("error = %v, want ErrUnexpectedContentBeforeFirstHeading", err)
	}
}

func TestNodeParse_FC09_NodeNameDoesNotMatch(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/other\nName content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	_, err := parsenode.NodeParse("SPEC/a")
	if !errors.Is(err, parsenode.ErrNodeNameDoesNotMatch) {
		t.Errorf("error = %v, want ErrNodeNameDoesNotMatch", err)
	}
}

func TestNodeParse_FC10_NodeNameCaseMismatchIsNotError(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# spec/foo\nName content.\n"
	testCreateNodeFile(t, "SPEC/FOO", body)

	node, err := parsenode.NodeParse("SPEC/FOO")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.NameSection.Heading != "spec/foo" {
		t.Errorf("heading = %q, want spec/foo", node.NameSection.Heading)
	}
}

func TestNodeParse_FC11_DuplicatePublicSameCaseError(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Public\nFirst public content.\n# Public\nSecond public content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	_, err := parsenode.NodeParse("SPEC/a")
	if !errors.Is(err, parsenode.ErrDuplicatePublicSection) {
		t.Errorf("error = %v, want ErrDuplicatePublicSection", err)
	}
}

func TestNodeParse_FC12_DuplicatePublicDifferentCaseError(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Public\nFirst public content.\n# PUBLIC\nSecond public content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	_, err := parsenode.NodeParse("SPEC/a")
	if !errors.Is(err, parsenode.ErrDuplicatePublicSection) {
		t.Errorf("error = %v, want ErrDuplicatePublicSection", err)
	}
}

func TestNodeParse_FC13_DuplicateAgentSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Agent\nFirst agent content.\n# Agent\nSecond agent content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	_, err := parsenode.NodeParse("SPEC/a")
	if !errors.Is(err, parsenode.ErrDuplicateAgentSection) {
		t.Errorf("error = %v, want ErrDuplicateAgentSection", err)
	}
}

func TestNodeParse_FC14_DuplicatePrivateSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Private\nFirst private content.\n# Private\nSecond private content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	_, err := parsenode.NodeParse("SPEC/a")
	if !errors.Is(err, parsenode.ErrDuplicatePrivateSection) {
		t.Errorf("error = %v, want ErrDuplicatePrivateSection", err)
	}
}

func TestNodeParse_FC15_UnrecognizedSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Decisions\nSome content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	_, err := parsenode.NodeParse("SPEC/a")
	if !errors.Is(err, parsenode.ErrUnrecognizedSection) {
		t.Errorf("error = %v, want ErrUnrecognizedSection", err)
	}
}

func TestNodeParse_FC16_UnrecognizedSectionRationale(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Rationale\nRationale content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	_, err := parsenode.NodeParse("SPEC/a")
	if !errors.Is(err, parsenode.ErrUnrecognizedSection) {
		t.Errorf("error = %v, want ErrUnrecognizedSection", err)
	}
}

func TestNodeParse_FC17_UnrecognizedSectionTODO(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# TODO\nTodo content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	_, err := parsenode.NodeParse("SPEC/a")
	if !errors.Is(err, parsenode.ErrUnrecognizedSection) {
		t.Errorf("error = %v, want ErrUnrecognizedSection", err)
	}
}

func TestNodeParse_FC18_DuplicateSubsectionPublicSameCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Public\n## Interface\nFirst interface content.\n## Interface\nSecond interface content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	_, err := parsenode.NodeParse("SPEC/a")
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}

func TestNodeParse_FC19_DuplicateSubsectionPublicDifferentCase(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Public\n## Interface\nFirst interface content.\n## INTERFACE\nSecond interface content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	_, err := parsenode.NodeParse("SPEC/a")
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}

func TestNodeParse_FC20_DuplicateSubsectionPublicWhitespaceVariation(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Public\n## Interface\nFirst interface content.\n##   Interface\nSecond interface content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	_, err := parsenode.NodeParse("SPEC/a")
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}

func TestNodeParse_FC21_DuplicateSubsectionAgent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	body := "# SPEC/a\nName content.\n# Agent\n## Guidance\nFirst guidance content.\n## Guidance\nSecond guidance content.\n"
	testCreateNodeFile(t, "SPEC/a", body)

	_, err := parsenode.NodeParse("SPEC/a")
	if !errors.Is(err, parsenode.ErrDuplicateSubsection) {
		t.Errorf("error = %v, want ErrDuplicateSubsection", err)
	}
}
