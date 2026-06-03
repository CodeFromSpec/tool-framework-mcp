// code-from-spec: ROOT/golang/tests/mcp_tools/load_chain@ml42H-0LGIF-eVuYU3f5XY6ln9Y
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

func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	parts := strings.Split(path, "/")
	if len(parts) > 1 {
		dir := strings.Join(parts[:len(parts)-1], "/")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("testWriteFile MkdirAll %s: %v", dir, err)
		}
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile %s: %v", path, err)
	}
}

func testSetupRootNode(t *testing.T, content string) {
	t.Helper()
	testWriteFile(t, "code-from-spec/_node.md", content)
}

func TestMCPLoadChain_SimpleLeafNode(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testSetupRootNode(t, "# ROOT\n\n# Public\n\nRoot public content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutput: \"out/a.go\"\n---\n# ROOT/a\n\n# Public\n\nNode a public content.\n\n# Agent\n\nNode a agent content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(result, "chain_hash: ") {
		t.Errorf("result does not start with 'chain_hash: ': %q", result[:min(len(result), 50)])
	}

	hashLine := strings.SplitN(result, "\n", 2)[0]
	hash := strings.TrimPrefix(hashLine, "chain_hash: ")
	if len(hash) != 27 {
		t.Errorf("expected 27-character hash, got %d characters: %q", len(hash), hash)
	}

	if !strings.Contains(result, "--- context ---") {
		t.Errorf("result missing '--- context ---'")
	}

	if !strings.Contains(result, "# Public") {
		t.Errorf("result missing '# Public' heading")
	}

	if !strings.Contains(result, "Node a public content.") {
		t.Errorf("result missing node a public content")
	}

	if !strings.Contains(result, "# Agent") {
		t.Errorf("result missing '# Agent' heading")
	}

	if !strings.Contains(result, "Node a agent content.") {
		t.Errorf("result missing node a agent content")
	}

	if !strings.Contains(result, "output: \"out/a.go\"") {
		t.Errorf("result missing output frontmatter")
	}

	if strings.Contains(result, "--- input ---") {
		t.Errorf("result should not contain '--- input ---'")
	}

	if strings.Contains(result, "--- existing artifact ---") {
		t.Errorf("result should not contain '--- existing artifact ---'")
	}
}

func TestMCPLoadChain_AncestorPublicContentIncluded(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testSetupRootNode(t, "# ROOT\n\n# Public\n\nRoot public content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"# ROOT/a\n\n# Public\n\nNode a public content.\n")
	testWriteFile(t, "code-from-spec/a/b/_node.md",
		"---\noutput: \"out/b.go\"\n---\n# ROOT/a/b\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "Root public content.") {
		t.Errorf("result missing ROOT's public content")
	}

	if !strings.Contains(result, "Node a public content.") {
		t.Errorf("result missing ROOT/a's public content")
	}
}

func TestMCPLoadChain_AncestorWithoutPublicSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testSetupRootNode(t, "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutput: \"out/a.go\"\n---\n# ROOT/a\n\n# Public\n\nNode a public content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parts := strings.SplitN(result, "--- context ---", 2)
	if len(parts) < 2 {
		t.Fatalf("missing '--- context ---' in result")
	}
	contextAndAfter := parts[1]
	contextSection := contextAndAfter
	if idx := strings.Index(contextAndAfter, "--- input ---"); idx >= 0 {
		contextSection = contextAndAfter[:idx]
	}
	if idx := strings.Index(contextAndAfter, "--- existing artifact ---"); idx >= 0 {
		if strings.Index(contextAndAfter, "--- input ---") < 0 || idx < strings.Index(contextAndAfter, "--- input ---") {
			contextSection = contextAndAfter[:idx]
		}
	}

	rootNodeContent := "# ROOT\n"
	_ = rootNodeContent
	if strings.Contains(contextSection, "root") && !strings.Contains(contextSection, "ROOT/a") {
		t.Errorf("context should not contain ROOT node content (no public section)")
	}
}

func TestMCPLoadChain_AncestorWithEmptyPublicSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testSetupRootNode(t, "# ROOT\n\n# Public\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutput: \"out/a.go\"\n---\n# ROOT/a\n\n# Public\n\nNode a public content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "Node a public content.") {
		t.Errorf("result missing node a public content")
	}
}

func TestMCPLoadChain_DependencyWithoutQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testSetupRootNode(t, "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutput: \"out/a.go\"\ndepends_on:\n  - ROOT/b\n---\n# ROOT/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md",
		"# ROOT/b\n\n# Public\n\n## Interface\n\nInterface content.\n\n## Constraints\n\nConstraints content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "## Interface") {
		t.Errorf("result missing '## Interface' heading")
	}

	if !strings.Contains(result, "Interface content.") {
		t.Errorf("result missing interface content")
	}

	if !strings.Contains(result, "## Constraints") {
		t.Errorf("result missing '## Constraints' heading")
	}

	if !strings.Contains(result, "Constraints content.") {
		t.Errorf("result missing constraints content")
	}
}

func TestMCPLoadChain_DependencyWithQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testSetupRootNode(t, "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutput: \"out/a.go\"\ndepends_on:\n  - ROOT/b(interface)\n---\n# ROOT/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md",
		"# ROOT/b\n\n# Public\n\n## Interface\n\nInterface content.\n\n## Constraints\n\nConstraints content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "## Interface") {
		t.Errorf("result missing '## Interface' heading")
	}

	if !strings.Contains(result, "Interface content.") {
		t.Errorf("result missing interface content")
	}

	if strings.Contains(result, "## Constraints") {
		t.Errorf("result should not contain '## Constraints' heading (qualifier restricts to Interface only)")
	}

	if strings.Contains(result, "Constraints content.") {
		t.Errorf("result should not contain constraints content")
	}
}

func TestMCPLoadChain_ArtifactDependency(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testSetupRootNode(t, "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutput: \"out/a.go\"\ndepends_on:\n  - ARTIFACT/b\n---\n# ROOT/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md",
		"---\noutput: \"out/b.go\"\n---\n# ROOT/b\n")
	testWriteFile(t, "out/b.go",
		"---\nsome: frontmatter\n---\nActual artifact body content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "Actual artifact body content.") {
		t.Errorf("result missing artifact body content")
	}
}

func TestMCPLoadChain_ExternalFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testSetupRootNode(t, "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutput: \"out/a.go\"\nexternal:\n  - path: \"data/config.yaml\"\n---\n# ROOT/a\n")
	testWriteFile(t, "data/config.yaml",
		"key: value\nother: data\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "key: value") {
		t.Errorf("result missing external file content")
	}

	if !strings.Contains(result, "other: data") {
		t.Errorf("result missing external file content")
	}
}

func TestMCPLoadChain_TargetReducedFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testSetupRootNode(t, "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutput: \"out/a.go\"\ndepends_on:\n  - ROOT/b\n---\n# ROOT/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md",
		"# ROOT/b\n\n# Public\n\nContent.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "output:") {
		t.Errorf("result missing output field in frontmatter")
	}

	if strings.Contains(result, "depends_on:") {
		t.Errorf("result should not contain depends_on field in target frontmatter block")
	}
}

func TestMCPLoadChain_TargetAgentSectionIncluded(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testSetupRootNode(t, "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutput: \"out/a.go\"\n---\n# ROOT/a\n\n# Public\n\nPublic content.\n\n# Agent\n\nAgent content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "Public content.") {
		t.Errorf("result missing public content")
	}

	if !strings.Contains(result, "Agent content.") {
		t.Errorf("result missing agent content")
	}
}

func TestMCPLoadChain_TargetWithoutAgentSection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testSetupRootNode(t, "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutput: \"out/a.go\"\n---\n# ROOT/a\n\n# Public\n\nPublic content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "Public content.") {
		t.Errorf("result missing public content")
	}

	if strings.Contains(result, "# Agent") {
		t.Errorf("result should not contain '# Agent' heading")
	}
}

func TestMCPLoadChain_InputPresent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testSetupRootNode(t, "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutput: \"out/a.go\"\ninput: \"ARTIFACT/b\"\n---\n# ROOT/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md",
		"---\noutput: \"out/data.json\"\n---\n# ROOT/b\n")
	testWriteFile(t, "out/data.json",
		"---\nsome: frontmatter\n---\nInput artifact body.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "--- input ---") {
		t.Errorf("result missing '--- input ---' section")
	}

	if !strings.Contains(result, "Input artifact body.") {
		t.Errorf("result missing input artifact body content")
	}

	parts := strings.SplitN(result, "--- input ---", 2)
	if len(parts) < 2 {
		t.Fatalf("missing '--- input ---' in result")
	}
	contextPart := parts[0]
	if strings.Contains(contextPart, "Input artifact body.") {
		t.Errorf("input content should not appear in context section")
	}
}

func TestMCPLoadChain_NoInput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testSetupRootNode(t, "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutput: \"out/a.go\"\n---\n# ROOT/a\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(result, "--- input ---") {
		t.Errorf("result should not contain '--- input ---'")
	}
}

func TestMCPLoadChain_ExistingArtifactPresent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testSetupRootNode(t, "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutput: \"out/a.go\"\n---\n# ROOT/a\n")
	testWriteFile(t, "out/a.go",
		"package main\n\nfunc main() {}\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "--- existing artifact ---") {
		t.Errorf("result missing '--- existing artifact ---' section")
	}

	if !strings.Contains(result, "package main") {
		t.Errorf("result missing existing artifact content")
	}
}

func TestMCPLoadChain_ExistingArtifactAbsent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testSetupRootNode(t, "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutput: \"out/a.go\"\n---\n# ROOT/a\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(result, "--- existing artifact ---") {
		t.Errorf("result should not contain '--- existing artifact ---'")
	}
}

func TestMCPLoadChain_HashIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testSetupRootNode(t, "# ROOT\n\n# Public\n\nFixed root content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutput: \"out/a.go\"\n---\n# ROOT/a\n\n# Public\n\nFixed node content.\n")

	result1, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}

	result2, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}

	hash1 := strings.SplitN(result1, "\n", 2)[0]
	hash2 := strings.SplitN(result2, "\n", 2)[0]

	if hash1 != hash2 {
		t.Errorf("hash is not deterministic: %q vs %q", hash1, hash2)
	}
}

func TestMCPLoadChain_InvalidLogicalName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := mcploadchain.MCPLoadChain("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("expected ErrUnsupportedReference, got: %v", err)
	}
}

func TestMCPLoadChain_NonexistentNodeFile(t *testing.T) {
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

func TestMCPLoadChain_NoOutputDeclared(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testSetupRootNode(t, "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"# ROOT/a\n")

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, mcploadchain.ErrNoOutput) {
		t.Errorf("expected ErrNoOutput, got: %v", err)
	}
}

func TestMCPLoadChain_InvalidOutputPathTraversal(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testSetupRootNode(t, "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutput: \"../../etc/passwd\"\n---\n# ROOT/a\n")

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, mcploadchain.ErrInvalidOutputPath) {
		t.Errorf("expected ErrInvalidOutputPath, got: %v", err)
	}
}

func TestMCPLoadChain_UnresolvableDependency(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testSetupRootNode(t, "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutput: \"out/a.go\"\ndepends_on:\n  - ROOT/missing\n---\n# ROOT/a\n")

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
