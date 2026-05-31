// code-from-spec: ROOT/golang/tests/mcp_tools/load_chain@jADIHFMwHb63-ia3a_9Ytu5pamQ
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

// testWriteFile creates a file at path (relative to cwd), creating parent
// directories as needed, and writes content to it.
func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	dir := path[:strings.LastIndex(path, "/")]
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("testWriteFile mkdir %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("testWriteFile %s: %v", path, err)
	}
}

// TestMCPLoadChain_TC01_SimpleLeafNode verifies a leaf node with one ancestor.
func TestMCPLoadChain_TC01_SimpleLeafNode(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md",
		"---\nname: ROOT\n---\n# ROOT\n\n# Public\nRoot public content line.\n")

	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: out\n    path: out/a.txt\n---\n# ROOT/a\n\n# Public\nNode A public content.\n\n# Agent\nNode A agent content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.ChainHash) != 27 {
		t.Errorf("chain_hash length = %d, want 27", len(result.ChainHash))
	}

	if !strings.Contains(result.Context, "Root public content line.") {
		t.Error("context missing ROOT public content")
	}
	if strings.Contains(result.Context, "# Public") {
		t.Error("context must not contain '# Public' heading")
	}
	if !strings.Contains(result.Context, "outputs:") {
		t.Error("context missing outputs frontmatter field for ROOT/a")
	}
	if strings.Contains(result.Context, "depends_on") {
		t.Error("context must not contain depends_on in frontmatter")
	}
	if !strings.Contains(result.Context, "Node A public content.") {
		t.Error("context missing Node A public content")
	}
	if !strings.Contains(result.Context, "Node A agent content.") {
		t.Error("context missing Node A agent content")
	}
	if result.Input != nil {
		t.Error("input should be absent")
	}
}

// TestMCPLoadChain_TC02_AncestorPublicContentIncluded verifies multi-level ancestors.
func TestMCPLoadChain_TC02_AncestorPublicContentIncluded(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md",
		"---\nname: ROOT\n---\n# ROOT\n\n# Public\nRoot public content.\n")

	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\nname: A\n---\n# ROOT/a\n\n# Public\nNode A public content.\n")

	testWriteFile(t, "code-from-spec/a/b/_node.md",
		"---\noutputs:\n  - id: out\n    path: out/b.txt\n---\n# ROOT/a/b\n\n# Public\nNode B public content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rootIdx := strings.Index(result.Context, "Root public content.")
	aIdx := strings.Index(result.Context, "Node A public content.")
	if rootIdx < 0 {
		t.Error("context missing ROOT public content")
	}
	if aIdx < 0 {
		t.Error("context missing Node A public content")
	}
	if rootIdx > aIdx {
		t.Error("ROOT content must appear before Node A content")
	}
	if strings.Contains(result.Context, "# Public") {
		t.Error("context must not contain '# Public' headings")
	}
}

// TestMCPLoadChain_TC03_AncestorWithoutPublicSectionSkipped verifies that ancestors
// without a public section do not contribute content.
func TestMCPLoadChain_TC03_AncestorWithoutPublicSectionSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md",
		"---\nname: ROOT\n---\n# ROOT\nRoot name section only — no public section.\n")

	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: out\n    path: out/a.txt\n---\n# ROOT/a\n\n# Public\nNode A public content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(result.Context, "Root name section only") {
		t.Error("context must not contain ROOT name-section body")
	}
	if !strings.Contains(result.Context, "Node A public content.") {
		t.Error("context missing Node A public content")
	}
}

// TestMCPLoadChain_TC04_AncestorWithEmptyPublicSectionSkipped verifies that an
// ancestor with an empty public section is skipped.
func TestMCPLoadChain_TC04_AncestorWithEmptyPublicSectionSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md",
		"---\nname: ROOT\n---\n# ROOT\n\n# Public\n")

	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: out\n    path: out/a.txt\n---\n# ROOT/a\n\n# Public\nNode A public content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "Node A public content.") {
		t.Error("context missing Node A public content")
	}
}

// TestMCPLoadChain_TC05_DependencyWithoutQualifier verifies dependency public content
// is included in full when no qualifier is present.
func TestMCPLoadChain_TC05_DependencyWithoutQualifier(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md",
		"---\nname: ROOT\n---\n# ROOT\n")

	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\ndepends_on:\n  - ROOT/b\noutputs:\n  - id: out\n    path: out/a.txt\n---\n# ROOT/a\n")

	testWriteFile(t, "code-from-spec/b/_node.md",
		"---\nname: B\n---\n# ROOT/b\n\n# Public\n## Interface\nB interface content.\n\n## Constraints\nB constraints content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "## Interface") {
		t.Error("context missing ## Interface heading from ROOT/b")
	}
	if !strings.Contains(result.Context, "B interface content.") {
		t.Error("context missing B interface content")
	}
	if !strings.Contains(result.Context, "## Constraints") {
		t.Error("context missing ## Constraints heading from ROOT/b")
	}
	if !strings.Contains(result.Context, "B constraints content.") {
		t.Error("context missing B constraints content")
	}
}

// TestMCPLoadChain_TC06_DependencyWithQualifier verifies that only the named
// subsection is included when a qualifier is specified.
func TestMCPLoadChain_TC06_DependencyWithQualifier(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md",
		"---\nname: ROOT\n---\n# ROOT\n")

	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\ndepends_on:\n  - ROOT/b(interface)\noutputs:\n  - id: out\n    path: out/a.txt\n---\n# ROOT/a\n")

	testWriteFile(t, "code-from-spec/b/_node.md",
		"---\nname: B\n---\n# ROOT/b\n\n# Public\n## Interface\nB interface content.\n\n## Constraints\nB constraints content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "B interface content.") {
		t.Error("context missing B interface content")
	}
	if strings.Contains(result.Context, "B constraints content.") {
		t.Error("context must not contain B constraints content (wrong qualifier)")
	}
}

// TestMCPLoadChain_TC07_ArtifactDependency verifies that an ARTIFACT dependency
// contributes its file content minus frontmatter.
func TestMCPLoadChain_TC07_ArtifactDependency(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md",
		"---\nname: ROOT\n---\n# ROOT\n")

	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\ndepends_on:\n  - ARTIFACT/b(code)\noutputs:\n  - id: out\n    path: out/a.txt\n---\n# ROOT/a\n")

	testWriteFile(t, "code-from-spec/b/_node.md",
		"---\noutputs:\n  - id: code\n    path: out/b.go\n---\n# ROOT/b\n")

	testWriteFile(t, "out/b.go",
		"// code-from-spec: ROOT/b@somehash\npackage main\n\nfunc Hello() {}\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "package main") {
		t.Error("context missing artifact body content")
	}
	if !strings.Contains(result.Context, "func Hello()") {
		t.Error("context missing artifact body function")
	}
}

// TestMCPLoadChain_TC08_ExternalFileFullContent verifies that an external file
// without fragments is included in full.
func TestMCPLoadChain_TC08_ExternalFileFullContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md",
		"---\nname: ROOT\n---\n# ROOT\n")

	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: out\n    path: out/a.txt\nexternal:\n  - path: data/config.yaml\n---\n# ROOT/a\n")

	testWriteFile(t, "data/config.yaml",
		"key: value\nother: 123\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "key: value") {
		t.Error("context missing external file content 'key: value'")
	}
	if !strings.Contains(result.Context, "other: 123") {
		t.Error("context missing external file content 'other: 123'")
	}
}

// TestMCPLoadChain_TC09_ExternalFileWithFragments verifies that only the specified
// line range from an external file is included.
func TestMCPLoadChain_TC09_ExternalFileWithFragments(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md",
		"---\nname: ROOT\n---\n# ROOT\n")

	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: out\n    path: out/a.txt\nexternal:\n  - path: data/big.txt\n    fragments:\n      - lines: \"2-4\"\n        hash: \"ignored\"\n---\n# ROOT/a\n")

	testWriteFile(t, "data/big.txt",
		"line 1\nline 2\nline 3\nline 4\nline 5\nline 6\nline 7\nline 8\nline 9\nline 10\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "line 2") {
		t.Error("context missing 'line 2'")
	}
	if !strings.Contains(result.Context, "line 3") {
		t.Error("context missing 'line 3'")
	}
	if !strings.Contains(result.Context, "line 4") {
		t.Error("context missing 'line 4'")
	}
	if strings.Contains(result.Context, "line 1\n") || strings.HasPrefix(result.Context, "line 1") {
		t.Error("context must not contain 'line 1' (outside fragment range)")
	}
	if strings.Contains(result.Context, "line 5") {
		t.Error("context must not contain 'line 5' (outside fragment range)")
	}
}

// TestMCPLoadChain_TC10_TargetFrontmatterReducedToOutputs verifies the target node
// frontmatter in context contains only outputs, not depends_on.
func TestMCPLoadChain_TC10_TargetFrontmatterReducedToOutputs(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md",
		"---\nname: ROOT\n---\n# ROOT\n")

	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\ndepends_on:\n  - ROOT/b\noutputs:\n  - id: out\n    path: out/a.txt\n---\n# ROOT/a\n")

	testWriteFile(t, "code-from-spec/b/_node.md",
		"---\nname: B\n---\n# ROOT/b\n\n# Public\nB content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "outputs:") {
		t.Error("context missing outputs field in frontmatter")
	}
	if strings.Contains(result.Context, "depends_on:") {
		t.Error("context must not contain depends_on in target frontmatter")
	}

	// Verify the outputs block is inside a --- delimited frontmatter block.
	parts := strings.Split(result.Context, "---")
	hasFrontmatter := false
	for _, part := range parts {
		if strings.Contains(part, "outputs:") {
			hasFrontmatter = true
			break
		}
	}
	if !hasFrontmatter {
		t.Error("context must contain outputs inside a --- delimited frontmatter block")
	}
}

// TestMCPLoadChain_TC11_TargetAgentSectionIncluded verifies both public and agent
// sections of the target are included.
func TestMCPLoadChain_TC11_TargetAgentSectionIncluded(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md",
		"---\nname: ROOT\n---\n# ROOT\n")

	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: out\n    path: out/a.txt\n---\n# ROOT/a\n\n# Public\nNode A public content.\n\n# Agent\nNode A agent guidance.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "Node A public content.") {
		t.Error("context missing Node A public content")
	}
	if !strings.Contains(result.Context, "Node A agent guidance.") {
		t.Error("context missing Node A agent guidance")
	}
	if strings.Contains(result.Context, "# Public") {
		t.Error("context must not contain '# Public' heading")
	}
	if strings.Contains(result.Context, "# Agent") {
		t.Error("context must not contain '# Agent' heading")
	}
}

// TestMCPLoadChain_TC12_TargetWithoutAgentSectionSkipped verifies that absence of
// an agent section causes no error and no placeholder in context.
func TestMCPLoadChain_TC12_TargetWithoutAgentSectionSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md",
		"---\nname: ROOT\n---\n# ROOT\n")

	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: out\n    path: out/a.txt\n---\n# ROOT/a\n\n# Public\nNode A public content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "Node A public content.") {
		t.Error("context missing Node A public content")
	}
	if strings.Contains(result.Context, "# Agent") {
		t.Error("context must not contain '# Agent' heading when section is absent")
	}
}

// TestMCPLoadChain_TC13_InputSeparatedFromContext verifies that the input artifact
// is returned separately and is not in context.
func TestMCPLoadChain_TC13_InputSeparatedFromContext(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md",
		"---\nname: ROOT\n---\n# ROOT\n")

	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: out\n    path: out/a.txt\ninput: ARTIFACT/b(data)\n---\n# ROOT/a\n")

	testWriteFile(t, "code-from-spec/b/_node.md",
		"---\noutputs:\n  - id: data\n    path: out/data.json\n---\n# ROOT/b\n")

	testWriteFile(t, "out/data.json",
		"---\nfrontmatter: present\n---\n{\"key\": \"value\"}\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Input == nil {
		t.Fatal("result.Input must not be nil")
	}
	if !strings.Contains(*result.Input, `{"key": "value"}`) {
		t.Errorf("result.Input = %q, want to contain JSON body", *result.Input)
	}
	if strings.Contains(*result.Input, "frontmatter: present") {
		t.Error("result.Input must not contain frontmatter")
	}
	if strings.Contains(result.Context, `{"key": "value"}`) {
		t.Error("context must not contain input content")
	}
}

// TestMCPLoadChain_TC14_NoInputFieldAbsent verifies that result.Input is nil when
// no input field is declared.
func TestMCPLoadChain_TC14_NoInputFieldAbsent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md",
		"---\nname: ROOT\n---\n# ROOT\n")

	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: out\n    path: out/a.txt\n---\n# ROOT/a\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Input != nil {
		t.Errorf("result.Input = %q, want nil", *result.Input)
	}
}

// TestMCPLoadChain_TC15_HashIsDeterministic verifies that calling MCPLoadChain
// twice for the same tree produces the same hash.
func TestMCPLoadChain_TC15_HashIsDeterministic(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md",
		"---\nname: ROOT\n---\n# ROOT\n\n# Public\nDeterministic content.\n")

	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: out\n    path: out/a.txt\n---\n# ROOT/a\n\n# Public\nLeaf content.\n")

	result1, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}

	result2, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}

	if result1.ChainHash != result2.ChainHash {
		t.Errorf("hashes differ: %q vs %q", result1.ChainHash, result2.ChainHash)
	}
}

// TestMCPLoadChain_TCE01_InvalidLogicalName verifies that a non-ROOT/ logical name
// returns ErrUnsupportedReference.
func TestMCPLoadChain_TCE01_InvalidLogicalName(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := mcploadchain.MCPLoadChain("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("error = %v, want errors.Is ErrUnsupportedReference", err)
	}
}

// TestMCPLoadChain_TCE02_NonexistentNodeFile verifies that a missing node file
// returns ErrFileUnreadable.
func TestMCPLoadChain_TCE02_NonexistentNodeFile(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := mcploadchain.MCPLoadChain("ROOT/nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("error = %v, want errors.Is ErrFileUnreadable", err)
	}
}

// TestMCPLoadChain_TCE03_NoOutputsDeclared verifies that ErrNoOutputs is returned
// when a node has no outputs field.
func TestMCPLoadChain_TCE03_NoOutputsDeclared(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md",
		"---\nname: ROOT\n---\n# ROOT\n")

	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\nname: A\n---\n# ROOT/a\n")

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcploadchain.ErrNoOutputs) {
		t.Errorf("error = %v, want errors.Is ErrNoOutputs", err)
	}
}

// TestMCPLoadChain_TCE04_InvalidOutputPathTraversal verifies that ErrInvalidOutputPath
// is returned when an output path attempts directory traversal.
func TestMCPLoadChain_TCE04_InvalidOutputPathTraversal(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md",
		"---\nname: ROOT\n---\n# ROOT\n")

	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: out\n    path: ../../etc/passwd\n---\n# ROOT/a\n")

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcploadchain.ErrInvalidOutputPath) {
		t.Errorf("error = %v, want errors.Is ErrInvalidOutputPath", err)
	}
}

// TestMCPLoadChain_TCE05_UnresolvableDependency verifies that a reference to a
// missing node returns an error during chain processing.
func TestMCPLoadChain_TCE05_UnresolvableDependency(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md",
		"---\nname: ROOT\n---\n# ROOT\n")

	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\ndepends_on:\n  - ROOT/missing\noutputs:\n  - id: out\n    path: out/a.txt\n---\n# ROOT/a\n")

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error for missing dependency, got nil")
	}
}
