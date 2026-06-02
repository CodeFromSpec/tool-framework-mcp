// code-from-spec: ROOT/golang/tests/mcp_tools/load_chain@DcEIr07AGWATNduLC5HKK_Wmjwc
package mcploadchain_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcploadchain"
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

func testWriteFile(t *testing.T, path, content string) {
	t.Helper()
	parts := strings.Split(path, "/")
	if len(parts) > 1 {
		dir := strings.Join(parts[:len(parts)-1], "/")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("testWriteFile MkdirAll: %v", err)
		}
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

func TestSimpleLeafNode(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n\nRoot public content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# ROOT/a\n\n# Public\n\nLeaf public content.\n\n# Agent\n\nLeaf agent content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.ChainHash) != 27 {
		t.Errorf("chain_hash length = %d, want 27", len(result.ChainHash))
	}

	if !strings.Contains(result.Context, "# Public") {
		t.Error("context should contain # Public heading")
	}
	if !strings.Contains(result.Context, "Root public content.") {
		t.Error("context should contain ROOT's public content")
	}
	if !strings.Contains(result.Context, "output: out/a.go") {
		t.Error("context should contain output frontmatter field")
	}
	if !strings.Contains(result.Context, "Leaf public content.") {
		t.Error("context should contain ROOT/a's public content")
	}
	if !strings.Contains(result.Context, "# Agent") {
		t.Error("context should contain # Agent heading")
	}
	if !strings.Contains(result.Context, "Leaf agent content.") {
		t.Error("context should contain ROOT/a's agent content")
	}
	if result.Input != nil {
		t.Error("result.Input should be absent")
	}
}

func TestAncestorPublicContentIncluded(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n\nRoot public content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nA public content.\n")
	testWriteFile(t, "code-from-spec/a/b/_node.md", "---\noutput: out/b.go\n---\n# ROOT/a/b\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rootIdx := strings.Index(result.Context, "Root public content.")
	aIdx := strings.Index(result.Context, "A public content.")

	if rootIdx < 0 {
		t.Error("context should contain ROOT's public content")
	}
	if aIdx < 0 {
		t.Error("context should contain ROOT/a's public content")
	}
	if rootIdx >= 0 && aIdx >= 0 && rootIdx > aIdx {
		t.Error("ROOT's public content should appear before ROOT/a's public content")
	}
}

func TestAncestorWithoutPublicSectionSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# ROOT/a\n\n# Public\n\nLeaf public content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Count(result.Context, "# Public") != 1 {
		t.Error("context should only contain one # Public heading (from ROOT/a, not ROOT)")
	}
}

func TestAncestorWithEmptyPublicSectionSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# ROOT/a\n\n# Public\n\nLeaf public content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Count(result.Context, "# Public") != 1 {
		t.Error("context should only contain one # Public heading (from ROOT/a, not ROOT)")
	}
}

func TestDependencyWithoutQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\ndepends_on:\n  - ROOT/b\n---\n# ROOT/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md", "# ROOT/b\n\n# Public\n\n## Interface\n\nInterface content.\n\n## Constraints\n\nConstraints content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "## Interface") {
		t.Error("context should contain ## Interface heading")
	}
	if !strings.Contains(result.Context, "Interface content.") {
		t.Error("context should contain Interface content")
	}
	if !strings.Contains(result.Context, "## Constraints") {
		t.Error("context should contain ## Constraints heading")
	}
	if !strings.Contains(result.Context, "Constraints content.") {
		t.Error("context should contain Constraints content")
	}
}

func TestDependencyWithQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\ndepends_on:\n  - ROOT/b(interface)\n---\n# ROOT/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md", "# ROOT/b\n\n# Public\n\n## Interface\n\nInterface content.\n\n## Constraints\n\nConstraints content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "## Interface") {
		t.Error("context should contain ## Interface heading")
	}
	if !strings.Contains(result.Context, "Interface content.") {
		t.Error("context should contain Interface content")
	}
	if strings.Contains(result.Context, "## Constraints") {
		t.Error("context should not contain ## Constraints heading")
	}
	if strings.Contains(result.Context, "Constraints content.") {
		t.Error("context should not contain Constraints content")
	}
}

func TestArtifactDependencyContentMinusFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\ndepends_on:\n  - ARTIFACT/b\n---\n# ROOT/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md", "---\noutput: out/b.go\n---\n# ROOT/b\n")
	testWriteFile(t, "out/b.go", "---\nsome: frontmatter\n---\npackage main\n\n// artifact body content\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "artifact body content") {
		t.Error("context should contain body of out/b.go")
	}
	if strings.Contains(result.Context, "some: frontmatter") {
		t.Error("context should not contain frontmatter of out/b.go")
	}
}

func TestExternalFileFullContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\nexternal:\n  - path: data/config.yaml\n---\n# ROOT/a\n")
	testWriteFile(t, "data/config.yaml", "key: value\nother: data\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "key: value") {
		t.Error("context should contain full content of data/config.yaml")
	}
	if !strings.Contains(result.Context, "other: data") {
		t.Error("context should contain full content of data/config.yaml")
	}
}

func TestTargetReducedFrontmatterOutputOnly(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/b/_node.md", "# ROOT/b\n\n# Public\n\nB content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\ndepends_on:\n  - ROOT/b\n---\n# ROOT/a\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "output: out/a.go") {
		t.Error("context should contain output field in frontmatter")
	}
	if strings.Contains(result.Context, "depends_on") {
		t.Error("context frontmatter should not contain depends_on field")
	}
}

func TestTargetAgentSectionIncluded(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# ROOT/a\n\n# Public\n\nPublic content.\n\n# Agent\n\nAgent content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "Public content.") {
		t.Error("context should contain public content")
	}
	if !strings.Contains(result.Context, "# Agent") {
		t.Error("context should contain # Agent heading")
	}
	if !strings.Contains(result.Context, "Agent content.") {
		t.Error("context should contain agent content")
	}
}

func TestTargetWithoutAgentSectionSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# ROOT/a\n\n# Public\n\nPublic content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "Public content.") {
		t.Error("context should contain public content")
	}
	if strings.Contains(result.Context, "# Agent") {
		t.Error("context should not contain # Agent heading")
	}
}

func TestInputSeparatedFromContext(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\ninput: ARTIFACT/b\n---\n# ROOT/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md", "---\noutput: out/data.json\n---\n# ROOT/b\n")
	testWriteFile(t, "out/data.json", "---\nsome: meta\n---\n{\"key\": \"input body value\"}\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Input == nil {
		t.Fatal("result.Input should be present")
	}
	if !strings.Contains(*result.Input, "input body value") {
		t.Error("result.Input should contain body of out/data.json")
	}
	if strings.Contains(*result.Input, "some: meta") {
		t.Error("result.Input should not contain frontmatter of out/data.json")
	}
	if strings.Contains(result.Context, "input body value") {
		t.Error("input content should not appear in result.Context")
	}
}

func TestNoInput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# ROOT/a\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Input != nil {
		t.Error("result.Input should be absent")
	}
}

func TestHashIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n\nKnown root content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# ROOT/a\n\n# Public\n\nKnown leaf content.\n")

	result1, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("first call unexpected error: %v", err)
	}

	result2, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("second call unexpected error: %v", err)
	}

	if result1.ChainHash != result2.ChainHash {
		t.Errorf("hash not deterministic: %q != %q", result1.ChainHash, result2.ChainHash)
	}
}

func TestInvalidLogicalNameNotROOT(t *testing.T) {
	_, err := mcploadchain.MCPLoadChain("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("expected ErrUnsupportedReference, got: %v", err)
	}
}

func TestNonexistentNodeFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := mcploadchain.MCPLoadChain("ROOT/nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestNoOutputDeclared(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n")

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcploadchain.ErrNoOutput) {
		t.Errorf("expected ErrNoOutput, got: %v", err)
	}
}

func TestInvalidOutputPathTraversal(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: ../../etc/passwd\n---\n# ROOT/a\n")

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcploadchain.ErrInvalidOutputPath) {
		t.Errorf("expected ErrInvalidOutputPath, got: %v", err)
	}
}

func TestUnresolvableDependency(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\ndepends_on:\n  - ROOT/missing\n---\n# ROOT/a\n")

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
