// code-from-spec: ROOT/golang/tests/mcp_tools/load_chain@eaRsfX1w0hJIL2705xpGtcNcjXM
package mcploadchain_test

import (
	"errors"
	"os"
	"path/filepath"
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
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

func TestMCPLoadChain_SimpleLeafNodeContextAndHash(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n\n## Context\n\nRoot context content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# ROOT/a\n\n# Public\n\n## Interface\n\nLeaf interface content.\n\n# Agent\n\nLeaf agent content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(result, "chain_hash: ") {
		t.Errorf("result does not start with 'chain_hash: '")
	}
	hashLine := strings.SplitN(result, "\n", 2)[0]
	hash := strings.TrimPrefix(hashLine, "chain_hash: ")
	if len(hash) != 27 {
		t.Errorf("hash length = %d, want 27", len(hash))
	}

	if !strings.Contains(result, "--- context ---") {
		t.Errorf("result missing '--- context ---'")
	}
	if !strings.Contains(result, "## Context") {
		t.Errorf("result missing '## Context' heading from root public section")
	}
	if !strings.Contains(result, "Root context content.") {
		t.Errorf("result missing root context content")
	}
	if !strings.Contains(result, "output: out/a.go") {
		t.Errorf("result missing reduced frontmatter with output")
	}
	if !strings.Contains(result, "## Interface") {
		t.Errorf("result missing '## Interface' heading from leaf public section")
	}
	if !strings.Contains(result, "Leaf interface content.") {
		t.Errorf("result missing leaf interface content")
	}
	if !strings.Contains(result, "# Agent") {
		t.Errorf("result missing '# Agent' heading")
	}
	if !strings.Contains(result, "Leaf agent content.") {
		t.Errorf("result missing leaf agent content")
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

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n\n## RootSub\n\nRoot public content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\n## ASub\n\nA public content.\n")
	testWriteFile(t, "code-from-spec/a/b/_node.md", "---\noutput: out/b.go\n---\n# ROOT/a/b\n\n# Public\n\n## BSub\n\nB public content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rootIdx := strings.Index(result, "Root public content.")
	aIdx := strings.Index(result, "A public content.")
	if rootIdx == -1 {
		t.Errorf("result missing root public content")
	}
	if aIdx == -1 {
		t.Errorf("result missing a public content")
	}
	if rootIdx > aIdx {
		t.Errorf("root public content should appear before a public content")
	}
	if strings.Contains(result, "# Public") {
		t.Errorf("result should not contain '# Public' heading for ancestors")
	}
}

func TestMCPLoadChain_AncestorWithoutPublicSectionSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n# Name\n\nRoot name content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# ROOT/a\n\n# Public\n\n## ASub\n\nA public content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(result, "Root name content.") {
		t.Errorf("result should not contain root content (no Public section)")
	}
}

func TestMCPLoadChain_AncestorWithEmptyPublicSectionSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# ROOT/a\n\n# Public\n\n## ASub\n\nA public content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	contextStart := strings.Index(result, "--- context ---")
	if contextStart == -1 {
		t.Fatal("result missing '--- context ---'")
	}
	contextPart := result[contextStart:]
	if inputStart := strings.Index(contextPart, "--- input ---"); inputStart != -1 {
		contextPart = contextPart[:inputStart]
	}
	if artifactStart := strings.Index(contextPart, "--- existing artifact ---"); artifactStart != -1 {
		contextPart = contextPart[:artifactStart]
	}

	if strings.Contains(contextPart, "# ROOT") {
		t.Errorf("root empty public section should be skipped — no root heading expected in context")
	}
}

func TestMCPLoadChain_DependencyWithoutQualifierPublicIncluded(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\ndepends_on:\n  - ROOT/b\n---\n# ROOT/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md", "# ROOT/b\n\n# Public\n\n## Interface\n\nInterface content.\n\n## Constraints\n\nConstraints content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "## Interface") {
		t.Errorf("result missing '## Interface' heading from dependency")
	}
	if !strings.Contains(result, "Interface content.") {
		t.Errorf("result missing interface content from dependency")
	}
	if !strings.Contains(result, "## Constraints") {
		t.Errorf("result missing '## Constraints' heading from dependency")
	}
	if !strings.Contains(result, "Constraints content.") {
		t.Errorf("result missing constraints content from dependency")
	}
}

func TestMCPLoadChain_DependencyWithQualifierSubsectionOnly(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\ndepends_on:\n  - ROOT/b(interface)\n---\n# ROOT/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md", "# ROOT/b\n\n# Public\n\n## Interface\n\nInterface content.\n\n## Constraints\n\nConstraints content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "## Interface") {
		t.Errorf("result missing '## Interface' heading from qualified dependency")
	}
	if !strings.Contains(result, "Interface content.") {
		t.Errorf("result missing interface content from qualified dependency")
	}
	if strings.Contains(result, "## Constraints") {
		t.Errorf("result should not contain '## Constraints' from qualified dependency")
	}
	if strings.Contains(result, "Constraints content.") {
		t.Errorf("result should not contain constraints content from qualified dependency")
	}
}

func TestMCPLoadChain_ArtifactDependencyArtifactTagLineRemoved(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\ndepends_on:\n  - ARTIFACT/b\n---\n# ROOT/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md", "---\noutput: out/b.go\n---\n# ROOT/b\n")
	testWriteFile(t, "out/b.go", "// code-from-spec: ROOT/b@aaaaaaaaaaaaaaaaaaaaaaaaaaa\nBody content of b.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "Body content of b.") {
		t.Errorf("result missing body content of artifact dependency")
	}
	if strings.Contains(result, "code-from-spec: ROOT/b@aaaaaaaaaaaaaaaaaaaaaaaaaaa") {
		t.Errorf("result should not contain the artifact tag line of artifact dependency")
	}
}

func TestMCPLoadChain_ExternalFileFullContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\nexternal:\n  - path: data/config.yaml\n---\n# ROOT/a\n")
	testWriteFile(t, "data/config.yaml", "key: value\nfoo: bar\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "key: value") {
		t.Errorf("result missing external file content")
	}
	if !strings.Contains(result, "foo: bar") {
		t.Errorf("result missing external file content")
	}
}

func TestMCPLoadChain_TargetHasReducedFrontmatterWithOutputOnly(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\ndepends_on:\n  - ROOT/b\n---\n# ROOT/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md", "# ROOT/b\n\n# Public\n\nB public content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "output: out/a.go") {
		t.Errorf("result missing output field in reduced frontmatter")
	}
	if strings.Contains(result, "depends_on") {
		t.Errorf("result should not contain depends_on in reduced frontmatter")
	}
}

func TestMCPLoadChain_TargetAgentSectionIncluded(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# ROOT/a\n\n# Public\n\n## Sub\n\nPublic content.\n\n# Agent\n\nAgent content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "## Sub") {
		t.Errorf("result missing '## Sub' heading from target public section")
	}
	if !strings.Contains(result, "Public content.") {
		t.Errorf("result missing public content")
	}
	if !strings.Contains(result, "# Agent") {
		t.Errorf("result missing Agent heading")
	}
	if !strings.Contains(result, "Agent content.") {
		t.Errorf("result missing agent content")
	}
}

func TestMCPLoadChain_TargetWithoutAgentSectionSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# ROOT/a\n\n# Public\n\n## Sub\n\nPublic content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "## Sub") {
		t.Errorf("result missing '## Sub' heading")
	}
	if !strings.Contains(result, "Public content.") {
		t.Errorf("result missing public content")
	}
	if strings.Contains(result, "# Agent") {
		t.Errorf("result should not contain '# Agent' heading")
	}
}

func TestMCPLoadChain_InputPresentInSeparateSection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\ninput: ARTIFACT/b\n---\n# ROOT/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md", "---\noutput: out/data.json\n---\n# ROOT/b\n")
	testWriteFile(t, "out/data.json", "// code-from-spec: ROOT/b@aaaaaaaaaaaaaaaaaaaaaaaaaaa\nJSON body content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "--- input ---") {
		t.Errorf("result missing '--- input ---' section")
	}

	inputIdx := strings.Index(result, "--- input ---")
	afterInput := result[inputIdx:]

	if !strings.Contains(afterInput, "JSON body content.") {
		t.Errorf("result missing input body content after '--- input ---'")
	}
	if strings.Contains(afterInput, "code-from-spec: ROOT/b@aaaaaaaaaaaaaaaaaaaaaaaaaaa") {
		t.Errorf("result should not contain the artifact tag line of input file after '--- input ---'")
	}

	contextSection := result[:inputIdx]
	if strings.Contains(contextSection, "JSON body content.") {
		t.Errorf("input content should not appear in context section")
	}
}

func TestMCPLoadChain_NoInputSectionAbsent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# ROOT/a\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(result, "--- input ---") {
		t.Errorf("result should not contain '--- input ---'")
	}
}

func TestMCPLoadChain_ExistingArtifactPresentInSeparateSection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# ROOT/a\n")
	testWriteFile(t, "out/a.go", "package main\n\nfunc main() {}\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "--- existing artifact ---") {
		t.Errorf("result missing '--- existing artifact ---' section")
	}

	artifactIdx := strings.Index(result, "--- existing artifact ---")
	afterArtifact := result[artifactIdx:]

	if !strings.Contains(afterArtifact, "package main") {
		t.Errorf("result missing existing artifact content")
	}
	if !strings.Contains(afterArtifact, "func main() {}") {
		t.Errorf("result missing existing artifact content")
	}
}

func TestMCPLoadChain_ExistingArtifactAbsentSectionOmitted(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# ROOT/a\n")

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

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n\n## Fixed\n\nFixed root content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# ROOT/a\n\n# Public\n\n## Fixed\n\nFixed a content.\n")

	result1, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("first call unexpected error: %v", err)
	}

	result2, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("second call unexpected error: %v", err)
	}

	hash1 := strings.SplitN(result1, "\n", 2)[0]
	hash2 := strings.SplitN(result2, "\n", 2)[0]
	if hash1 != hash2 {
		t.Errorf("hashes differ: %q vs %q", hash1, hash2)
	}
}

func TestMCPLoadChain_InvalidLogicalNameNotRoot(t *testing.T) {
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
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestMCPLoadChain_NoOutputDeclared(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nSome content.\n")

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcploadchain.ErrNoOutput) {
		t.Errorf("expected ErrNoOutput, got %v", err)
	}
}

func TestMCPLoadChain_InvalidOutputPathTraversal(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: ../../etc/passwd\n---\n# ROOT/a\n")

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

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\ndepends_on:\n  - ROOT/missing\n---\n# ROOT/a\n")

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
