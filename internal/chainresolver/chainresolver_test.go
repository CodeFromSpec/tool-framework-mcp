// code-from-spec: ROOT/golang/tests/chain/resolver@COStax6fd_XO8UhJ8VqwlaLKwkA
package chainresolver_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
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
	path := filepath.Join("code-from-spec", filepath.FromSlash(logicalName[len("ROOT"):]), "_node.md")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("testWriteNode mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNode write: %v", err)
	}
}

func testNodeContent(logicalName string) string {
	return "# " + logicalName + "\n"
}

func testNodeWithFrontmatter(logicalName string, frontmatter string) string {
	return "---\n" + frontmatter + "---\n# " + logicalName + "\n"
}

// --- Ancestors and target ---

func TestChainResolve_RootAsTarget(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", testNodeContent("ROOT"))

	chain, err := chainresolver.ChainResolve("ROOT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Ancestors) != 0 {
		t.Errorf("expected no ancestors, got %d", len(chain.Ancestors))
	}
	if chain.Target == nil {
		t.Fatal("expected target, got nil")
	}
	if chain.Target.LogicalName != "ROOT" {
		t.Errorf("expected target logical name ROOT, got %q", chain.Target.LogicalName)
	}
	if chain.Target.Qualifier != "" {
		t.Errorf("expected no qualifier, got %q", chain.Target.Qualifier)
	}
	if len(chain.Dependencies) != 0 {
		t.Errorf("expected no dependencies, got %d", len(chain.Dependencies))
	}
	if len(chain.External) != 0 {
		t.Errorf("expected no external, got %d", len(chain.External))
	}
	if chain.Input != nil {
		t.Errorf("expected no input, got %v", chain.Input)
	}
}

func TestChainResolve_LinearChain(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", testNodeContent("ROOT"))
	testWriteNode(t, "ROOT/a", testNodeContent("ROOT/a"))
	testWriteNode(t, "ROOT/a/b", testNodeContent("ROOT/a/b"))

	chain, err := chainresolver.ChainResolve("ROOT/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Ancestors) != 2 {
		t.Fatalf("expected 2 ancestors, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].LogicalName != "ROOT" {
		t.Errorf("expected first ancestor ROOT, got %q", chain.Ancestors[0].LogicalName)
	}
	if chain.Ancestors[1].LogicalName != "ROOT/a" {
		t.Errorf("expected second ancestor ROOT/a, got %q", chain.Ancestors[1].LogicalName)
	}
	if chain.Target.LogicalName != "ROOT/a/b" {
		t.Errorf("expected target ROOT/a/b, got %q", chain.Target.LogicalName)
	}
}

func TestChainResolve_SingleParent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", testNodeContent("ROOT"))
	testWriteNode(t, "ROOT/a", testNodeContent("ROOT/a"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Ancestors) != 1 {
		t.Fatalf("expected 1 ancestor, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].LogicalName != "ROOT" {
		t.Errorf("expected ancestor ROOT, got %q", chain.Ancestors[0].LogicalName)
	}
	if chain.Target.LogicalName != "ROOT/a" {
		t.Errorf("expected target ROOT/a, got %q", chain.Target.LogicalName)
	}
}

func TestChainResolve_TargetWithEmptyFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", testNodeContent("ROOT"))
	testWriteNode(t, "ROOT/a", testNodeContent("ROOT/a"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Ancestors) != 1 {
		t.Fatalf("expected 1 ancestor, got %d", len(chain.Ancestors))
	}
	if chain.Target.LogicalName != "ROOT/a" {
		t.Errorf("expected target ROOT/a, got %q", chain.Target.LogicalName)
	}
	if len(chain.Dependencies) != 0 {
		t.Errorf("expected no dependencies, got %d", len(chain.Dependencies))
	}
	if len(chain.External) != 0 {
		t.Errorf("expected no external, got %d", len(chain.External))
	}
	if chain.Input != nil {
		t.Errorf("expected no input, got %v", chain.Input)
	}
}

// --- Dependencies — ROOT/ references ---

func TestChainResolve_DependencyWithoutQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", testNodeContent("ROOT"))
	testWriteNode(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ROOT/b\n"))
	testWriteNode(t, "ROOT/b", testNodeContent("ROOT/b"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].LogicalName != "ROOT/b" {
		t.Errorf("expected dependency ROOT/b, got %q", chain.Dependencies[0].LogicalName)
	}
	if chain.Dependencies[0].Qualifier != "" {
		t.Errorf("expected no qualifier, got %q", chain.Dependencies[0].Qualifier)
	}
}

func TestChainResolve_DependencyWithQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", testNodeContent("ROOT"))
	testWriteNode(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ROOT/b(interface)\n"))
	testWriteNode(t, "ROOT/b", testNodeContent("ROOT/b"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].LogicalName != "ROOT/b" {
		t.Errorf("expected dependency ROOT/b, got %q", chain.Dependencies[0].LogicalName)
	}
	if chain.Dependencies[0].Qualifier != "interface" {
		t.Errorf("expected qualifier 'interface', got %q", chain.Dependencies[0].Qualifier)
	}
}

func TestChainResolve_DependenciesSortedByFilePath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", testNodeContent("ROOT"))
	testWriteNode(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ROOT/z\n  - ROOT/m\n  - ROOT/b\n"))
	testWriteNode(t, "ROOT/z", testNodeContent("ROOT/z"))
	testWriteNode(t, "ROOT/m", testNodeContent("ROOT/m"))
	testWriteNode(t, "ROOT/b", testNodeContent("ROOT/b"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 3 {
		t.Fatalf("expected 3 dependencies, got %d", len(chain.Dependencies))
	}
	names := []string{
		chain.Dependencies[0].LogicalName,
		chain.Dependencies[1].LogicalName,
		chain.Dependencies[2].LogicalName,
	}
	expected := []string{"ROOT/b", "ROOT/m", "ROOT/z"}
	for i, n := range expected {
		if names[i] != n {
			t.Errorf("expected dependency[%d] = %q, got %q", i, n, names[i])
		}
	}
}

// --- Dependencies — ARTIFACT/ references ---

func TestChainResolve_ArtifactDependencyResolved(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", testNodeContent("ROOT"))
	testWriteNode(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ARTIFACT/b\n"))
	testWriteNode(t, "ROOT/b", testNodeWithFrontmatter("ROOT/b", "output: out/lib.go\n"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].LogicalName != "ARTIFACT/b" {
		t.Errorf("expected dependency ARTIFACT/b, got %q", chain.Dependencies[0].LogicalName)
	}
	if chain.Dependencies[0].FilePath.Value != "out/lib.go" {
		t.Errorf("expected file path out/lib.go, got %q", chain.Dependencies[0].FilePath.Value)
	}
}

func TestChainResolve_ArtifactGeneratingNodeHasNoOutput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", testNodeContent("ROOT"))
	testWriteNode(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ARTIFACT/b\n"))
	testWriteNode(t, "ROOT/b", testNodeContent("ROOT/b"))

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestChainResolve_ArtifactFileDoesNotExistOnDisk(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", testNodeContent("ROOT"))
	testWriteNode(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ARTIFACT/b\n"))
	testWriteNode(t, "ROOT/b", testNodeWithFrontmatter("ROOT/b", "output: out/lib.go\n"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].FilePath.Value != "out/lib.go" {
		t.Errorf("expected file path out/lib.go, got %q", chain.Dependencies[0].FilePath.Value)
	}
}

func TestChainResolve_MixedRootAndArtifactDependencies(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", testNodeContent("ROOT"))
	testWriteNode(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ROOT/c\n  - ARTIFACT/b\n"))
	testWriteNode(t, "ROOT/b", testNodeWithFrontmatter("ROOT/b", "output: out/lib.go\n"))
	testWriteNode(t, "ROOT/c", testNodeContent("ROOT/c"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}
}

// --- Dependencies — dedup ---

func TestChainResolve_ExactDuplicate(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", testNodeContent("ROOT"))
	testWriteNode(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ROOT/b\n  - ROOT/b\n"))
	testWriteNode(t, "ROOT/b", testNodeContent("ROOT/b"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency (deduped), got %d", len(chain.Dependencies))
	}
}

func TestChainResolve_NoQualifierSubsumesQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", testNodeContent("ROOT"))
	testWriteNode(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ROOT/b\n  - ROOT/b(interface)\n"))
	testWriteNode(t, "ROOT/b", testNodeContent("ROOT/b"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].Qualifier != "" {
		t.Errorf("expected no qualifier (unqualified wins), got %q", chain.Dependencies[0].Qualifier)
	}
}

func TestChainResolve_QualifierBeforeNoQualifier_NoQualifierWins(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", testNodeContent("ROOT"))
	testWriteNode(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ROOT/b(interface)\n  - ROOT/b\n"))
	testWriteNode(t, "ROOT/b", testNodeContent("ROOT/b"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].Qualifier != "" {
		t.Errorf("expected no qualifier (unqualified wins), got %q", chain.Dependencies[0].Qualifier)
	}
}

func TestChainResolve_SameFileDifferentQualifiers(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", testNodeContent("ROOT"))
	testWriteNode(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ROOT/b(interface)\n  - ROOT/b(constraints)\n"))
	testWriteNode(t, "ROOT/b", testNodeContent("ROOT/b"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}
	qualifiers := []string{chain.Dependencies[0].Qualifier, chain.Dependencies[1].Qualifier}
	if qualifiers[0] != "constraints" || qualifiers[1] != "interface" {
		t.Errorf("expected qualifiers [constraints interface], got %v", qualifiers)
	}
}

func TestChainResolve_DuplicateArtifact(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", testNodeContent("ROOT"))
	testWriteNode(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ARTIFACT/b\n  - ARTIFACT/b\n"))
	testWriteNode(t, "ROOT/b", testNodeWithFrontmatter("ROOT/b", "output: out/lib.go\n"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency (deduped), got %d", len(chain.Dependencies))
	}
}

// --- External ---

func TestChainResolve_ExternalEntriesCopied(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", testNodeContent("ROOT"))
	testWriteNode(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "external:\n  - path: docs/api.yaml\n  - path: proto/v1.proto\n"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.External) != 2 {
		t.Fatalf("expected 2 external entries, got %d", len(chain.External))
	}
	if chain.External[0].Path != "docs/api.yaml" {
		t.Errorf("expected first external docs/api.yaml, got %q", chain.External[0].Path)
	}
	if chain.External[1].Path != "proto/v1.proto" {
		t.Errorf("expected second external proto/v1.proto, got %q", chain.External[1].Path)
	}
}

func TestChainResolve_EmptyExternal(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", testNodeContent("ROOT"))
	testWriteNode(t, "ROOT/a", testNodeContent("ROOT/a"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.External) != 0 {
		t.Errorf("expected no external entries, got %d", len(chain.External))
	}
}

// --- Input ---

func TestChainResolve_InputResolvedFromGeneratingNode(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", testNodeContent("ROOT"))
	testWriteNode(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "input: ARTIFACT/b\n"))
	testWriteNode(t, "ROOT/b", testNodeWithFrontmatter("ROOT/b", "output: out/data.json\n"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain.Input == nil {
		t.Fatal("expected input, got nil")
	}
	if chain.Input.LogicalName != "ARTIFACT/b" {
		t.Errorf("expected input ARTIFACT/b, got %q", chain.Input.LogicalName)
	}
	if chain.Input.FilePath.Value != "out/data.json" {
		t.Errorf("expected input file path out/data.json, got %q", chain.Input.FilePath.Value)
	}
}

func TestChainResolve_NoInput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", testNodeContent("ROOT"))
	testWriteNode(t, "ROOT/a", testNodeContent("ROOT/a"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain.Input != nil {
		t.Errorf("expected no input, got %v", chain.Input)
	}
}

// --- Error cases ---

func TestChainResolve_UnrecognizedPrefixInDependsOn(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", testNodeContent("ROOT"))
	testWriteNode(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - UNKNOWN/something\n"))

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestChainResolve_InvalidTargetLogicalName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := chainresolver.ChainResolve("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestChainResolve_UnreadableFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", testNodeContent("ROOT"))
	testWriteNode(t, "ROOT/a", "---\nkey: [unclosed bracket\n---\n# ROOT/a\n")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnreadableFrontmatter) {
		t.Errorf("expected ErrUnreadableFrontmatter, got %v", err)
	}
}
