// code-from-spec: ROOT/golang/tests/mcp_tools/load_chain@2sdEno29RoZ9F_Jg9M2lFiziGhI
package mcploadchain_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
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

func testWriteFile(t *testing.T, name string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(name, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile write: %v", err)
	}
}

func testWriteNode(t *testing.T, logicalName string, content string) {
	t.Helper()
	path := filepath.Join("code-from-spec", filepath.FromSlash(logicalName[len("ROOT"):]), "_node.md")
	testWriteFile(t, path, content)
}

// --- Happy Path ---

func TestMCPLoadChain_SimpleLeafNode(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "# ROOT\n\n# Public\n\nRoot public content.\n")
	testWriteNode(t, "ROOT/a", "---\noutput: out/a.go\n---\n# ROOT/a\n\n# Public\n\nLeaf public content.\n\n# Agent\n\nLeaf agent content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.ChainHash) != 27 {
		t.Errorf("expected 27-character hash, got %d: %q", len(result.ChainHash), result.ChainHash)
	}
	if !strings.Contains(result.Context, "# Public") {
		t.Errorf("expected context to contain '# Public'")
	}
	if !strings.Contains(result.Context, "Root public content.") {
		t.Errorf("expected context to contain root public content")
	}
	if !strings.Contains(result.Context, "output: out/a.go") {
		t.Errorf("expected context to contain output frontmatter")
	}
	if !strings.Contains(result.Context, "Leaf public content.") {
		t.Errorf("expected context to contain leaf public content")
	}
	if !strings.Contains(result.Context, "# Agent") {
		t.Errorf("expected context to contain '# Agent'")
	}
	if !strings.Contains(result.Context, "Leaf agent content.") {
		t.Errorf("expected context to contain leaf agent content")
	}
	if result.Input != nil {
		t.Errorf("expected no input, got %v", *result.Input)
	}
}

func TestMCPLoadChain_AncestorPublicContentIncluded(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "# ROOT\n\n# Public\n\nRoot public content.\n")
	testWriteNode(t, "ROOT/a", "# ROOT/a\n\n# Public\n\nA public content.\n")
	testWriteNode(t, "ROOT/a/b", "---\noutput: out/b.go\n---\n# ROOT/a/b\n\n# Public\n\nB public content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Context, "Root public content.") {
		t.Errorf("expected context to contain root public content")
	}
	if !strings.Contains(result.Context, "A public content.") {
		t.Errorf("expected context to contain ROOT/a public content")
	}
	rootIdx := strings.Index(result.Context, "Root public content.")
	aIdx := strings.Index(result.Context, "A public content.")
	if rootIdx > aIdx {
		t.Errorf("expected root content before ROOT/a content")
	}
}

func TestMCPLoadChain_AncestorWithoutPublicSectionSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "# ROOT\n")
	testWriteNode(t, "ROOT/a", "---\noutput: out/a.go\n---\n# ROOT/a\n\n# Public\n\nLeaf public content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result.Context, "# ROOT\n") {
		t.Errorf("expected root content to be skipped (no public section)")
	}
	if !strings.Contains(result.Context, "Leaf public content.") {
		t.Errorf("expected context to contain leaf public content")
	}
}

func TestMCPLoadChain_AncestorWithEmptyPublicSectionSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "# ROOT\n\n# Public\n")
	testWriteNode(t, "ROOT/a", "---\noutput: out/a.go\n---\n# ROOT/a\n\n# Public\n\nLeaf public content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Context, "Leaf public content.") {
		t.Errorf("expected context to contain leaf public content")
	}
}

func TestMCPLoadChain_DependencyWithoutQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "# ROOT\n")
	testWriteNode(t, "ROOT/a", "---\noutput: out/a.go\ndepends_on:\n  - ROOT/b\n---\n# ROOT/a\n\n# Public\n\nA content.\n")
	testWriteNode(t, "ROOT/b", "# ROOT/b\n\n# Public\n\n## Interface\n\nInterface content.\n\n## Constraints\n\nConstraints content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Context, "## Interface") {
		t.Errorf("expected context to contain '## Interface'")
	}
	if !strings.Contains(result.Context, "Interface content.") {
		t.Errorf("expected context to contain interface content")
	}
	if !strings.Contains(result.Context, "## Constraints") {
		t.Errorf("expected context to contain '## Constraints'")
	}
	if !strings.Contains(result.Context, "Constraints content.") {
		t.Errorf("expected context to contain constraints content")
	}
}

func TestMCPLoadChain_DependencyWithQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "# ROOT\n")
	testWriteNode(t, "ROOT/a", "---\noutput: out/a.go\ndepends_on:\n  - ROOT/b(interface)\n---\n# ROOT/a\n\n# Public\n\nA content.\n")
	testWriteNode(t, "ROOT/b", "# ROOT/b\n\n# Public\n\n## Interface\n\nInterface content.\n\n## Constraints\n\nConstraints content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Context, "## Interface") {
		t.Errorf("expected context to contain '## Interface'")
	}
	if !strings.Contains(result.Context, "Interface content.") {
		t.Errorf("expected context to contain interface content")
	}
	if strings.Contains(result.Context, "## Constraints") {
		t.Errorf("expected context to NOT contain '## Constraints' (qualifier filters to interface only)")
	}
	if strings.Contains(result.Context, "Constraints content.") {
		t.Errorf("expected context to NOT contain constraints content")
	}
}

func TestMCPLoadChain_ArtifactDependencyContentWithoutFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "# ROOT\n")
	testWriteNode(t, "ROOT/a", "---\noutput: out/a.go\ndepends_on:\n  - ARTIFACT/b\n---\n# ROOT/a\n\n# Public\n\nA content.\n")
	testWriteNode(t, "ROOT/b", "---\noutput: out/b.go\n---\n# ROOT/b\n")
	testWriteFile(t, "out/b.go", "---\nsome: frontmatter\n---\npackage main\n\nfunc B() {}\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Context, "package main") {
		t.Errorf("expected context to contain body content of out/b.go")
	}
	if !strings.Contains(result.Context, "func B()") {
		t.Errorf("expected context to contain func B body")
	}
	if strings.Contains(result.Context, "some: frontmatter") {
		t.Errorf("expected context to NOT contain frontmatter of artifact file")
	}
}

func TestMCPLoadChain_ExternalFileFullContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "# ROOT\n")
	testWriteNode(t, "ROOT/a", "---\noutput: out/a.go\nexternal:\n  - path: data/config.yaml\n---\n# ROOT/a\n\n# Public\n\nA content.\n")
	testWriteFile(t, "data/config.yaml", "key: value\nanother: setting\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Context, "key: value") {
		t.Errorf("expected context to contain external file content")
	}
	if !strings.Contains(result.Context, "another: setting") {
		t.Errorf("expected context to contain full external file content")
	}
}

func TestMCPLoadChain_TargetFrontmatterContainsOnlyOutput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "# ROOT\n")
	testWriteNode(t, "ROOT/a", "---\noutput: out/a.go\ndepends_on:\n  - ROOT/b\n---\n# ROOT/a\n\n# Public\n\nA content.\n")
	testWriteNode(t, "ROOT/b", "# ROOT/b\n\n# Public\n\nB content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Context, "output: out/a.go") {
		t.Errorf("expected context to contain output field in frontmatter")
	}
	if strings.Contains(result.Context, "depends_on:") {
		t.Errorf("expected context to NOT contain depends_on in target frontmatter")
	}
}

func TestMCPLoadChain_TargetAgentSectionIncluded(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "# ROOT\n")
	testWriteNode(t, "ROOT/a", "---\noutput: out/a.go\n---\n# ROOT/a\n\n# Public\n\nPublic content.\n\n# Agent\n\nAgent content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Context, "# Public") {
		t.Errorf("expected context to contain '# Public'")
	}
	if !strings.Contains(result.Context, "Public content.") {
		t.Errorf("expected context to contain public content")
	}
	if !strings.Contains(result.Context, "# Agent") {
		t.Errorf("expected context to contain '# Agent'")
	}
	if !strings.Contains(result.Context, "Agent content.") {
		t.Errorf("expected context to contain agent content")
	}
}

func TestMCPLoadChain_TargetWithoutAgentSection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "# ROOT\n")
	testWriteNode(t, "ROOT/a", "---\noutput: out/a.go\n---\n# ROOT/a\n\n# Public\n\nPublic content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Context, "Public content.") {
		t.Errorf("expected context to contain public content")
	}
	if strings.Contains(result.Context, "# Agent") {
		t.Errorf("expected context to NOT contain '# Agent'")
	}
}

func TestMCPLoadChain_InputSeparatedFromContext(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "# ROOT\n")
	testWriteNode(t, "ROOT/a", "---\noutput: out/a.go\ninput: ARTIFACT/b\n---\n# ROOT/a\n\n# Public\n\nA content.\n")
	testWriteNode(t, "ROOT/b", "---\noutput: out/data.json\n---\n# ROOT/b\n")
	testWriteFile(t, "out/data.json", "---\nsome: meta\n---\n{\"key\": \"value\"}\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Input == nil {
		t.Fatal("expected input, got nil")
	}
	if !strings.Contains(*result.Input, "{\"key\": \"value\"}") {
		t.Errorf("expected input to contain body content, got %q", *result.Input)
	}
	if strings.Contains(*result.Input, "some: meta") {
		t.Errorf("expected input to NOT contain frontmatter")
	}
	if strings.Contains(result.Context, "{\"key\": \"value\"}") {
		t.Errorf("expected input content to NOT appear in context")
	}
}

func TestMCPLoadChain_NoInput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "# ROOT\n")
	testWriteNode(t, "ROOT/a", "---\noutput: out/a.go\n---\n# ROOT/a\n\n# Public\n\nA content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Input != nil {
		t.Errorf("expected no input, got %v", *result.Input)
	}
}

func TestMCPLoadChain_HashIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "# ROOT\n\n# Public\n\nKnown root content.\n")
	testWriteNode(t, "ROOT/a", "---\noutput: out/a.go\n---\n# ROOT/a\n\n# Public\n\nKnown leaf content.\n")

	result1, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("first call unexpected error: %v", err)
	}
	result2, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("second call unexpected error: %v", err)
	}
	if result1.ChainHash != result2.ChainHash {
		t.Errorf("expected deterministic hash, got %q and %q", result1.ChainHash, result2.ChainHash)
	}
}

// --- Error Cases ---

func TestMCPLoadChain_InvalidLogicalName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := mcploadchain.MCPLoadChain("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("expected ErrUnsupportedReference, got %v", err)
	}
}

func TestMCPLoadChain_NonexistentNodeFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := mcploadchain.MCPLoadChain("ROOT/nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrFileUnreadable) {
		t.Errorf("expected frontmatter.ErrFileUnreadable, got %v", err)
	}
}

func TestMCPLoadChain_NoOutputDeclared(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "# ROOT\n")
	testWriteNode(t, "ROOT/a", "# ROOT/a\n\n# Public\n\nA content.\n")

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcploadchain.ErrNoOutput) {
		t.Errorf("expected ErrNoOutput, got %v", err)
	}
}

func TestMCPLoadChain_InvalidOutputPath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "# ROOT\n")
	testWriteNode(t, "ROOT/a", "---\noutput: ../../etc/passwd\n---\n# ROOT/a\n\n# Public\n\nA content.\n")

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcploadchain.ErrInvalidOutputPath) {
		t.Errorf("expected ErrInvalidOutputPath, got %v", err)
	}
}

func TestMCPLoadChain_UnresolvableDependency(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "# ROOT\n")
	testWriteNode(t, "ROOT/a", "---\noutput: out/a.go\ndepends_on:\n  - ROOT/missing\n---\n# ROOT/a\n\n# Public\n\nA content.\n")

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
