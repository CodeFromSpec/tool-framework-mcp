// code-from-spec: ROOT/golang/tests/chain/resolver@hcaf_7X6ZllcR9zQTcANyG43dyQ
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

func testWriteNode(t *testing.T, dir string, logicalPath string, content string) {
	t.Helper()
	fullPath := filepath.Join(dir, filepath.FromSlash(logicalPath))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("testWriteNode mkdir: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNode write: %v", err)
	}
}

func testSetupRoot(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	testWriteNode(t, dir, "code-from-spec/_node.md", "")
	return dir
}

func TestChainResolve_RootAsTarget(t *testing.T) {
	dir := testSetupRoot(t)
	testChdir(t, dir)

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
		t.Errorf("expected target logical name ROOT, got %s", chain.Target.LogicalName)
	}
	if chain.Target.Qualifier != "" {
		t.Errorf("expected no qualifier, got %s", chain.Target.Qualifier)
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
	dir := testSetupRoot(t)
	testWriteNode(t, dir, "code-from-spec/a/_node.md", "")
	testWriteNode(t, dir, "code-from-spec/a/b/_node.md", "")
	testChdir(t, dir)

	chain, err := chainresolver.ChainResolve("ROOT/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Ancestors) != 2 {
		t.Fatalf("expected 2 ancestors, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].LogicalName != "ROOT" {
		t.Errorf("expected first ancestor ROOT, got %s", chain.Ancestors[0].LogicalName)
	}
	if chain.Ancestors[1].LogicalName != "ROOT/a" {
		t.Errorf("expected second ancestor ROOT/a, got %s", chain.Ancestors[1].LogicalName)
	}
	if chain.Target == nil || chain.Target.LogicalName != "ROOT/a/b" {
		t.Errorf("expected target ROOT/a/b, got %v", chain.Target)
	}
}

func TestChainResolve_SingleParent(t *testing.T) {
	dir := testSetupRoot(t)
	testWriteNode(t, dir, "code-from-spec/a/_node.md", "")
	testChdir(t, dir)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Ancestors) != 1 {
		t.Fatalf("expected 1 ancestor, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].LogicalName != "ROOT" {
		t.Errorf("expected ancestor ROOT, got %s", chain.Ancestors[0].LogicalName)
	}
	if chain.Target == nil || chain.Target.LogicalName != "ROOT/a" {
		t.Errorf("expected target ROOT/a, got %v", chain.Target)
	}
}

func TestChainResolve_TargetWithEmptyFrontmatter(t *testing.T) {
	dir := testSetupRoot(t)
	testWriteNode(t, dir, "code-from-spec/a/_node.md", "---\n---\n")
	testChdir(t, dir)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Ancestors) != 1 {
		t.Errorf("expected 1 ancestor, got %d", len(chain.Ancestors))
	}
	if chain.Target == nil || chain.Target.LogicalName != "ROOT/a" {
		t.Errorf("expected target ROOT/a, got %v", chain.Target)
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

func TestChainResolve_DependencyWithoutQualifier(t *testing.T) {
	dir := testSetupRoot(t)
	testWriteNode(t, dir, "code-from-spec/a/_node.md", "---\ndepends_on:\n  - ROOT/b\n---\n")
	testWriteNode(t, dir, "code-from-spec/b/_node.md", "")
	testChdir(t, dir)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ROOT/b" {
		t.Errorf("expected logical name ROOT/b, got %s", dep.LogicalName)
	}
	if dep.Qualifier != "" {
		t.Errorf("expected no qualifier, got %s", dep.Qualifier)
	}
}

func TestChainResolve_DependencyWithQualifier(t *testing.T) {
	dir := testSetupRoot(t)
	testWriteNode(t, dir, "code-from-spec/a/_node.md", "---\ndepends_on:\n  - ROOT/b(interface)\n---\n")
	testWriteNode(t, dir, "code-from-spec/b/_node.md", "")
	testChdir(t, dir)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ROOT/b" {
		t.Errorf("expected logical name ROOT/b, got %s", dep.LogicalName)
	}
	if dep.Qualifier != "interface" {
		t.Errorf("expected qualifier interface, got %s", dep.Qualifier)
	}
}

func TestChainResolve_DependenciesSortedByFilePath(t *testing.T) {
	dir := testSetupRoot(t)
	testWriteNode(t, dir, "code-from-spec/a/_node.md", "---\ndepends_on:\n  - ROOT/z\n  - ROOT/m\n  - ROOT/b\n---\n")
	testWriteNode(t, dir, "code-from-spec/z/_node.md", "")
	testWriteNode(t, dir, "code-from-spec/m/_node.md", "")
	testWriteNode(t, dir, "code-from-spec/b/_node.md", "")
	testChdir(t, dir)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 3 {
		t.Fatalf("expected 3 dependencies, got %d", len(chain.Dependencies))
	}
	names := []string{chain.Dependencies[0].LogicalName, chain.Dependencies[1].LogicalName, chain.Dependencies[2].LogicalName}
	if names[0] != "ROOT/b" || names[1] != "ROOT/m" || names[2] != "ROOT/z" {
		t.Errorf("expected order ROOT/b, ROOT/m, ROOT/z but got %v", names)
	}
}

func TestChainResolve_ArtifactDependencyResolved(t *testing.T) {
	dir := testSetupRoot(t)
	testWriteNode(t, dir, "code-from-spec/a/_node.md", "---\ndepends_on:\n  - ARTIFACT/b\n---\n")
	testWriteNode(t, dir, "code-from-spec/b/_node.md", "---\noutput: out/lib.go\n---\n")
	testChdir(t, dir)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ARTIFACT/b" {
		t.Errorf("expected logical name ARTIFACT/b, got %s", dep.LogicalName)
	}
	if dep.FilePath.Value != "out/lib.go" {
		t.Errorf("expected file path out/lib.go, got %s", dep.FilePath.Value)
	}
}

func TestChainResolve_ArtifactGeneratingNodeHasNoOutput(t *testing.T) {
	dir := testSetupRoot(t)
	testWriteNode(t, dir, "code-from-spec/a/_node.md", "---\ndepends_on:\n  - ARTIFACT/b\n---\n")
	testWriteNode(t, dir, "code-from-spec/b/_node.md", "---\n---\n")
	testChdir(t, dir)

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestChainResolve_ArtifactFileDoesNotExistOnDisk(t *testing.T) {
	dir := testSetupRoot(t)
	testWriteNode(t, dir, "code-from-spec/a/_node.md", "---\ndepends_on:\n  - ARTIFACT/b\n---\n")
	testWriteNode(t, dir, "code-from-spec/b/_node.md", "---\noutput: out/lib.go\n---\n")
	testChdir(t, dir)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].FilePath.Value != "out/lib.go" {
		t.Errorf("expected file path out/lib.go, got %s", chain.Dependencies[0].FilePath.Value)
	}
}

func TestChainResolve_MixedRootAndArtifactDependencies(t *testing.T) {
	dir := testSetupRoot(t)
	testWriteNode(t, dir, "code-from-spec/a/_node.md", "---\ndepends_on:\n  - ROOT/c\n  - ARTIFACT/b\n---\n")
	testWriteNode(t, dir, "code-from-spec/b/_node.md", "---\noutput: out/lib.go\n---\n")
	testWriteNode(t, dir, "code-from-spec/c/_node.md", "")
	testChdir(t, dir)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}
	paths := []string{chain.Dependencies[0].FilePath.Value, chain.Dependencies[1].FilePath.Value}
	if paths[0] >= paths[1] {
		t.Errorf("expected dependencies sorted by file path, got %v", paths)
	}
}

func TestChainResolve_DedupExactDuplicate(t *testing.T) {
	dir := testSetupRoot(t)
	testWriteNode(t, dir, "code-from-spec/a/_node.md", "---\ndepends_on:\n  - ROOT/b\n  - ROOT/b\n---\n")
	testWriteNode(t, dir, "code-from-spec/b/_node.md", "")
	testChdir(t, dir)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
}

func TestChainResolve_DedupNoQualifierSubsumesQualifier(t *testing.T) {
	dir := testSetupRoot(t)
	testWriteNode(t, dir, "code-from-spec/a/_node.md", "---\ndepends_on:\n  - ROOT/b\n  - ROOT/b(interface)\n---\n")
	testWriteNode(t, dir, "code-from-spec/b/_node.md", "")
	testChdir(t, dir)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].Qualifier != "" {
		t.Errorf("expected no qualifier (no-qualifier wins), got %s", chain.Dependencies[0].Qualifier)
	}
}

func TestChainResolve_DedupQualifierBeforeNoQualifier(t *testing.T) {
	dir := testSetupRoot(t)
	testWriteNode(t, dir, "code-from-spec/a/_node.md", "---\ndepends_on:\n  - ROOT/b(interface)\n  - ROOT/b\n---\n")
	testWriteNode(t, dir, "code-from-spec/b/_node.md", "")
	testChdir(t, dir)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].Qualifier != "" {
		t.Errorf("expected no qualifier (no-qualifier wins), got %s", chain.Dependencies[0].Qualifier)
	}
}

func TestChainResolve_DedupSameFileDifferentQualifiers(t *testing.T) {
	dir := testSetupRoot(t)
	testWriteNode(t, dir, "code-from-spec/a/_node.md", "---\ndepends_on:\n  - ROOT/b(interface)\n  - ROOT/b(constraints)\n---\n")
	testWriteNode(t, dir, "code-from-spec/b/_node.md", "")
	testChdir(t, dir)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}
	qualifiers := []string{chain.Dependencies[0].Qualifier, chain.Dependencies[1].Qualifier}
	if qualifiers[0] != "constraints" || qualifiers[1] != "interface" {
		t.Errorf("expected qualifiers [constraints, interface], got %v", qualifiers)
	}
}

func TestChainResolve_DedupDuplicateArtifact(t *testing.T) {
	dir := testSetupRoot(t)
	testWriteNode(t, dir, "code-from-spec/a/_node.md", "---\ndepends_on:\n  - ARTIFACT/b\n  - ARTIFACT/b\n---\n")
	testWriteNode(t, dir, "code-from-spec/b/_node.md", "---\noutput: out/lib.go\n---\n")
	testChdir(t, dir)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
}

func TestChainResolve_ExternalEntriesCopiedSorted(t *testing.T) {
	dir := testSetupRoot(t)
	testWriteNode(t, dir, "code-from-spec/a/_node.md", "---\nexternal:\n  - path: proto/v1.proto\n  - path: docs/api.yaml\n---\n")
	testChdir(t, dir)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.External) != 2 {
		t.Fatalf("expected 2 external entries, got %d", len(chain.External))
	}
	if chain.External[0].Path != "docs/api.yaml" {
		t.Errorf("expected first external docs/api.yaml, got %s", chain.External[0].Path)
	}
	if chain.External[1].Path != "proto/v1.proto" {
		t.Errorf("expected second external proto/v1.proto, got %s", chain.External[1].Path)
	}
}

func TestChainResolve_ExternalEmpty(t *testing.T) {
	dir := testSetupRoot(t)
	testWriteNode(t, dir, "code-from-spec/a/_node.md", "---\n---\n")
	testChdir(t, dir)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.External) != 0 {
		t.Errorf("expected no external entries, got %d", len(chain.External))
	}
}

func TestChainResolve_InputResolved(t *testing.T) {
	dir := testSetupRoot(t)
	testWriteNode(t, dir, "code-from-spec/a/_node.md", "---\ninput: ARTIFACT/b\n---\n")
	testWriteNode(t, dir, "code-from-spec/b/_node.md", "---\noutput: out/data.json\n---\n")
	testChdir(t, dir)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chain.Input == nil {
		t.Fatal("expected input, got nil")
	}
	if chain.Input.LogicalName != "ARTIFACT/b" {
		t.Errorf("expected input logical name ARTIFACT/b, got %s", chain.Input.LogicalName)
	}
	if chain.Input.FilePath.Value != "out/data.json" {
		t.Errorf("expected input file path out/data.json, got %s", chain.Input.FilePath.Value)
	}
}

func TestChainResolve_InputAbsent(t *testing.T) {
	dir := testSetupRoot(t)
	testWriteNode(t, dir, "code-from-spec/a/_node.md", "---\n---\n")
	testChdir(t, dir)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chain.Input != nil {
		t.Errorf("expected no input, got %v", chain.Input)
	}
}

func TestChainResolve_UnrecognizedPrefixInDependsOn(t *testing.T) {
	dir := testSetupRoot(t)
	testWriteNode(t, dir, "code-from-spec/a/_node.md", "---\ndepends_on:\n  - UNKNOWN/something\n---\n")
	testChdir(t, dir)

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
	dir := testSetupRoot(t)
	testWriteNode(t, dir, "code-from-spec/a/_node.md", "---\n: invalid: yaml: [\n---\n")
	testChdir(t, dir)

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnreadableFrontmatter) {
		t.Errorf("expected ErrUnreadableFrontmatter, got %v", err)
	}
}
