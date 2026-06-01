// code-from-spec: ROOT/golang/tests/mcp_tools/load_chain@hAgUFLn-ty_2XDjmzxbgmKIoeTg
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
	dir := path[:strings.LastIndex(path, "/")]
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testWriteFile MkdirAll %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile %s: %v", path, err)
	}
}

func TestMCPLoadChain_TC01_SimpleLeafNodeContextAndHash(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n# Public\n\nRoot public content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: main\n    path: out/a.go\n---\n# ROOT/a\n# Public\n\nNode a public content.\n\n# Agent\n\nNode a agent content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.ChainHash) != 27 {
		t.Errorf("chain_hash length = %d, want 27", len(result.ChainHash))
	}

	if !strings.Contains(result.Context, "# Public") {
		t.Error("context missing # Public heading")
	}
	if !strings.Contains(result.Context, "Root public content.") {
		t.Error("context missing Root public content.")
	}
	if !strings.Contains(result.Context, "Node a public content.") {
		t.Error("context missing Node a public content.")
	}
	if !strings.Contains(result.Context, "# Agent") {
		t.Error("context missing # Agent heading")
	}
	if !strings.Contains(result.Context, "Node a agent content.") {
		t.Error("context missing Node a agent content.")
	}

	rootIdx := strings.Index(result.Context, "Root public content.")
	agentIdx := strings.Index(result.Context, "Node a agent content.")
	nodeAPublicIdx := strings.Index(result.Context, "Node a public content.")
	if rootIdx < 0 || nodeAPublicIdx < 0 || agentIdx < 0 {
		t.Fatal("one or more expected content strings missing from context")
	}
	if rootIdx > nodeAPublicIdx {
		t.Error("Root public content should appear before Node a public content")
	}
	if nodeAPublicIdx > agentIdx {
		t.Error("Node a public content should appear before Node a agent content")
	}

	if strings.Contains(result.Context, "depends_on") {
		t.Error("context should not contain depends_on")
	}
	if !strings.Contains(result.Context, "outputs") {
		t.Error("context should contain outputs in frontmatter block")
	}

	if result.Input != nil {
		t.Error("input should be absent")
	}
}

func TestMCPLoadChain_TC02_AncestorPublicContentIncluded(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n# Public\n\nRoot public content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n# Public\n\nNode a public content.\n")
	testWriteFile(t, "code-from-spec/a/b/_node.md",
		"---\noutputs:\n  - id: main\n    path: out/b.go\n---\n# ROOT/a/b\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "Root public content.") {
		t.Error("context missing Root public content.")
	}
	if !strings.Contains(result.Context, "Node a public content.") {
		t.Error("context missing Node a public content.")
	}

	rootIdx := strings.Index(result.Context, "Root public content.")
	nodeAIdx := strings.Index(result.Context, "Node a public content.")
	if rootIdx > nodeAIdx {
		t.Error("Root content should appear before ROOT/a content (ancestor-first order)")
	}
}

func TestMCPLoadChain_TC03_AncestorWithoutPublicSectionSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n# Name\n\nRoot name content only.\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: main\n    path: out/a.go\n---\n# ROOT/a\n# Public\n\nNode a public content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(result.Context, "Root name content only.") {
		t.Error("context should not contain ROOT content (no Public section)")
	}
	if !strings.Contains(result.Context, "Node a public content.") {
		t.Error("context missing Node a public content.")
	}
}

func TestMCPLoadChain_TC04_AncestorWithEmptyPublicSectionSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n# Public\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: main\n    path: out/a.go\n---\n# ROOT/a\n# Public\n\nNode a public content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	publicIdx := strings.Index(result.Context, "# Public")
	nodeAIdx := strings.Index(result.Context, "Node a public content.")
	if publicIdx < 0 {
		t.Error("context should contain at least one # Public heading from ROOT/a")
	}

	rootPublicCount := 0
	search := result.Context
	for {
		idx := strings.Index(search, "# Public")
		if idx < 0 {
			break
		}
		rootPublicCount++
		search = search[idx+len("# Public"):]
	}
	_ = rootPublicCount
	_ = nodeAIdx

	if strings.Contains(result.Context, "\n\n# Public\n\n# Public") {
		t.Error("ROOT's empty public section should not add a duplicate heading")
	}
}

func TestMCPLoadChain_TC05_DependencyWithoutQualifierFullPublicIncluded(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\ndepends_on:\n  - ROOT/b\noutputs:\n  - id: main\n    path: out/a.go\n---\n# ROOT/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md",
		"# ROOT/b\n# Public\n\n## Interface\n\nInterface content.\n\n## Constraints\n\nConstraints content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "## Interface") {
		t.Error("context missing ## Interface subsection heading")
	}
	if !strings.Contains(result.Context, "Interface content.") {
		t.Error("context missing Interface content.")
	}
	if !strings.Contains(result.Context, "## Constraints") {
		t.Error("context missing ## Constraints subsection heading")
	}
	if !strings.Contains(result.Context, "Constraints content.") {
		t.Error("context missing Constraints content.")
	}
}

func TestMCPLoadChain_TC06_DependencyWithQualifierSubsectionOnly(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\ndepends_on:\n  - ROOT/b(interface)\noutputs:\n  - id: main\n    path: out/a.go\n---\n# ROOT/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md",
		"# ROOT/b\n# Public\n\n## Interface\n\nInterface content.\n\n## Constraints\n\nConstraints content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "Interface content.") {
		t.Error("context missing Interface content.")
	}
	if strings.Contains(result.Context, "Constraints content.") {
		t.Error("context should not contain Constraints content. (not the qualified subsection)")
	}
	if strings.Contains(result.Context, "## Constraints") {
		t.Error("context should not contain ## Constraints heading")
	}
}

func TestMCPLoadChain_TC07_ArtifactDependencyContentMinusFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\ndepends_on:\n  - ARTIFACT/b(code)\noutputs:\n  - id: main\n    path: out/a.go\n---\n# ROOT/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md",
		"---\noutputs:\n  - id: code\n    path: out/b.go\n---\n# ROOT/b\n")
	testWriteFile(t, "out/b.go",
		"---\nsome: frontmatter\n---\nBody content of b.go.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "Body content of b.go.") {
		t.Error("context missing Body content of b.go.")
	}
	if strings.Contains(result.Context, "some: frontmatter") {
		t.Error("context should not contain frontmatter from out/b.go")
	}
}

func TestMCPLoadChain_TC08_ExternalFileFullContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\nexternal:\n  - path: data/config.yaml\noutputs:\n  - id: main\n    path: out/a.go\n---\n# ROOT/a\n")
	testWriteFile(t, "data/config.yaml", "key: value\nother: data\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "key: value") {
		t.Error("context missing key: value from config.yaml")
	}
	if !strings.Contains(result.Context, "other: data") {
		t.Error("context missing other: data from config.yaml")
	}
}

func TestMCPLoadChain_TC09_TargetFrontmatterContainsOutputsOnly(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\ndepends_on:\n  - ROOT/b\noutputs:\n  - id: main\n    path: out/a.go\n---\n# ROOT/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md", "# ROOT/b\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "outputs") {
		t.Error("context should contain outputs in the frontmatter block for the target")
	}
	if strings.Contains(result.Context, "depends_on") {
		t.Error("context should not contain depends_on in the target frontmatter block")
	}
}

func TestMCPLoadChain_TC10_TargetAgentSectionIncluded(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: main\n    path: out/a.go\n---\n# ROOT/a\n# Public\n\nNode a public content.\n\n# Agent\n\nNode a agent content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "Node a public content.") {
		t.Error("context missing Node a public content.")
	}
	if !strings.Contains(result.Context, "# Agent") {
		t.Error("context missing # Agent heading")
	}
	if !strings.Contains(result.Context, "Node a agent content.") {
		t.Error("context missing Node a agent content.")
	}
}

func TestMCPLoadChain_TC11_TargetWithoutAgentSectionSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: main\n    path: out/a.go\n---\n# ROOT/a\n# Public\n\nNode a public content.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "Node a public content.") {
		t.Error("context missing Node a public content.")
	}
	if strings.Contains(result.Context, "# Agent") {
		t.Error("context should not contain # Agent heading")
	}
}

func TestMCPLoadChain_TC12_InputSeparatedFromContext(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\ninput: ARTIFACT/b(data)\noutputs:\n  - id: main\n    path: out/a.go\n---\n# ROOT/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md",
		"---\noutputs:\n  - id: data\n    path: out/data.json\n---\n# ROOT/b\n")
	testWriteFile(t, "out/data.json",
		"---\nsome: frontmatter\n---\nBody content of data.json.\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Input == nil {
		t.Fatal("result.Input should not be nil")
	}
	if !strings.Contains(*result.Input, "Body content of data.json.") {
		t.Error("input missing Body content of data.json.")
	}
	if strings.Contains(*result.Input, "some: frontmatter") {
		t.Error("input should not contain frontmatter from out/data.json")
	}
	if strings.Contains(result.Context, "Body content of data.json.") {
		t.Error("context should not contain input body content")
	}
}

func TestMCPLoadChain_TC13_NoInputFieldAbsent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: main\n    path: out/a.go\n---\n# ROOT/a\n")

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Input != nil {
		t.Error("result.Input should be nil when no input is declared")
	}
}

func TestMCPLoadChain_TC14_HashIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n# Public\n\nRoot public content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: main\n    path: out/a.go\n---\n# ROOT/a\n# Public\n\nNode a public content.\n")

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

func TestMCPLoadChain_TC15_InvalidLogicalName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := mcploadchain.MCPLoadChain("INVALID/something")
	if err == nil {
		t.Fatal("expected error for invalid logical name, got nil")
	}
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("expected ErrUnsupportedReference, got: %v", err)
	}
}

func TestMCPLoadChain_TC16_NonexistentNodeFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := mcploadchain.MCPLoadChain("ROOT/nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent node, got nil")
	}
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestMCPLoadChain_TC17_NoOutputsDeclared(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n# Public\n\nNode a public content.\n")

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error for missing outputs, got nil")
	}
	if !errors.Is(err, mcploadchain.ErrNoOutputs) {
		t.Errorf("expected ErrNoOutputs, got: %v", err)
	}
}

func TestMCPLoadChain_TC18_InvalidOutputPathTraversal(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: main\n    path: ../../etc/passwd\n---\n# ROOT/a\n")

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error for path traversal output, got nil")
	}
	if !errors.Is(err, mcploadchain.ErrInvalidOutputPath) {
		t.Errorf("expected ErrInvalidOutputPath, got: %v", err)
	}
}

func TestMCPLoadChain_TC19_UnresolvableDependency(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\ndepends_on:\n  - ROOT/missing\noutputs:\n  - id: main\n    path: out/a.go\n---\n# ROOT/a\n")

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error for unresolvable dependency, got nil")
	}
}
