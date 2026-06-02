// code-from-spec: ROOT/golang/tests/chain/resolver@cPOoVkPtMFk3Z14o4K0zi-vSesA
package chainresolver_test

import (
	"errors"
	"os"
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

func testWriteNode(t *testing.T, relPath string, frontmatter string) {
	t.Helper()
	if err := os.MkdirAll(relPath[:len(relPath)-len("_node.md")], 0755); err != nil {
		t.Fatalf("testWriteNode mkdir: %v", err)
	}
	content := "---\n" + frontmatter + "---\n"
	if err := os.WriteFile(relPath, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNode write: %v", err)
	}
}

func testWriteEmptyNode(t *testing.T, relPath string) {
	t.Helper()
	testWriteNode(t, relPath, "")
}

func TestChainResolve_RootAsTarget(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "code-from-spec/_node.md")

	chain, err := chainresolver.ChainResolve("ROOT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Ancestors) != 0 {
		t.Errorf("expected empty ancestors, got %d", len(chain.Ancestors))
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
		t.Errorf("expected empty dependencies, got %d", len(chain.Dependencies))
	}
	if len(chain.External) != 0 {
		t.Errorf("expected empty external, got %d", len(chain.External))
	}
	if chain.Input != nil {
		t.Errorf("expected no input, got %v", chain.Input)
	}
}

func TestChainResolve_LinearChain(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "code-from-spec/_node.md")
	testWriteEmptyNode(t, "code-from-spec/a/_node.md")
	testWriteEmptyNode(t, "code-from-spec/a/b/_node.md")

	chain, err := chainresolver.ChainResolve("ROOT/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Ancestors) != 2 {
		t.Fatalf("expected 2 ancestors, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].LogicalName != "ROOT" {
		t.Errorf("expected ancestor[0] = ROOT, got %q", chain.Ancestors[0].LogicalName)
	}
	if chain.Ancestors[1].LogicalName != "ROOT/a" {
		t.Errorf("expected ancestor[1] = ROOT/a, got %q", chain.Ancestors[1].LogicalName)
	}
	if chain.Target == nil || chain.Target.LogicalName != "ROOT/a/b" {
		t.Errorf("expected target = ROOT/a/b")
	}
}

func TestChainResolve_SingleParent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "code-from-spec/_node.md")
	testWriteEmptyNode(t, "code-from-spec/a/_node.md")

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
	if chain.Target == nil || chain.Target.LogicalName != "ROOT/a" {
		t.Errorf("expected target ROOT/a")
	}
}

func TestChainResolve_TargetEmptyFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "code-from-spec/_node.md")
	testWriteEmptyNode(t, "code-from-spec/a/_node.md")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Ancestors) != 1 {
		t.Errorf("expected 1 ancestor, got %d", len(chain.Ancestors))
	}
	if chain.Target == nil || chain.Target.LogicalName != "ROOT/a" {
		t.Errorf("expected target ROOT/a")
	}
	if len(chain.Dependencies) != 0 {
		t.Errorf("expected empty dependencies, got %d", len(chain.Dependencies))
	}
	if len(chain.External) != 0 {
		t.Errorf("expected empty external, got %d", len(chain.External))
	}
	if chain.Input != nil {
		t.Errorf("expected no input")
	}
}

func TestChainResolve_DependencyWithoutQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "code-from-spec/_node.md")
	testWriteNode(t, "code-from-spec/a/_node.md", "depends_on:\n  - ROOT/b\n")
	testWriteEmptyNode(t, "code-from-spec/b/_node.md")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ROOT/b" {
		t.Errorf("expected logical name ROOT/b, got %q", dep.LogicalName)
	}
	if dep.Qualifier != "" {
		t.Errorf("expected no qualifier, got %q", dep.Qualifier)
	}
}

func TestChainResolve_DependencyWithQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "code-from-spec/_node.md")
	testWriteNode(t, "code-from-spec/a/_node.md", "depends_on:\n  - ROOT/b(interface)\n")
	testWriteEmptyNode(t, "code-from-spec/b/_node.md")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ROOT/b" {
		t.Errorf("expected logical name ROOT/b, got %q", dep.LogicalName)
	}
	if dep.Qualifier != "interface" {
		t.Errorf("expected qualifier interface, got %q", dep.Qualifier)
	}
}

func TestChainResolve_DependenciesSortedByFilePath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "code-from-spec/_node.md")
	testWriteNode(t, "code-from-spec/a/_node.md", "depends_on:\n  - ROOT/z\n  - ROOT/m\n  - ROOT/b\n")
	testWriteEmptyNode(t, "code-from-spec/z/_node.md")
	testWriteEmptyNode(t, "code-from-spec/m/_node.md")
	testWriteEmptyNode(t, "code-from-spec/b/_node.md")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 3 {
		t.Fatalf("expected 3 dependencies, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].LogicalName != "ROOT/b" {
		t.Errorf("expected first dep ROOT/b, got %q", chain.Dependencies[0].LogicalName)
	}
	if chain.Dependencies[1].LogicalName != "ROOT/m" {
		t.Errorf("expected second dep ROOT/m, got %q", chain.Dependencies[1].LogicalName)
	}
	if chain.Dependencies[2].LogicalName != "ROOT/z" {
		t.Errorf("expected third dep ROOT/z, got %q", chain.Dependencies[2].LogicalName)
	}
}

func TestChainResolve_ArtifactDependencyResolved(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "code-from-spec/_node.md")
	testWriteNode(t, "code-from-spec/a/_node.md", "depends_on:\n  - ARTIFACT/b\n")
	testWriteNode(t, "code-from-spec/b/_node.md", "output: out/lib.go\n")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ARTIFACT/b" {
		t.Errorf("expected logical name ARTIFACT/b, got %q", dep.LogicalName)
	}
	if dep.FilePath.Value != "out/lib.go" {
		t.Errorf("expected file path out/lib.go, got %q", dep.FilePath.Value)
	}
}

func TestChainResolve_ArtifactNoOutput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "code-from-spec/_node.md")
	testWriteNode(t, "code-from-spec/a/_node.md", "depends_on:\n  - ARTIFACT/b\n")
	testWriteEmptyNode(t, "code-from-spec/b/_node.md")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestChainResolve_ArtifactFileNotOnDisk(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "code-from-spec/_node.md")
	testWriteNode(t, "code-from-spec/a/_node.md", "depends_on:\n  - ARTIFACT/b\n")
	testWriteNode(t, "code-from-spec/b/_node.md", "output: out/lib.go\n")

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

func TestChainResolve_MixedDependencies(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "code-from-spec/_node.md")
	testWriteNode(t, "code-from-spec/a/_node.md", "depends_on:\n  - ROOT/c\n  - ARTIFACT/b\n")
	testWriteNode(t, "code-from-spec/b/_node.md", "output: out/lib.go\n")
	testWriteEmptyNode(t, "code-from-spec/c/_node.md")

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

func TestChainResolve_ExactDuplicateDep(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "code-from-spec/_node.md")
	testWriteNode(t, "code-from-spec/a/_node.md", "depends_on:\n  - ROOT/b\n  - ROOT/b\n")
	testWriteEmptyNode(t, "code-from-spec/b/_node.md")

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

	testWriteEmptyNode(t, "code-from-spec/_node.md")
	testWriteNode(t, "code-from-spec/a/_node.md", "depends_on:\n  - ROOT/b\n  - ROOT/b(interface)\n")
	testWriteEmptyNode(t, "code-from-spec/b/_node.md")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].Qualifier != "" {
		t.Errorf("expected no qualifier, got %q", chain.Dependencies[0].Qualifier)
	}
}

func TestChainResolve_QualifierBeforeNoQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "code-from-spec/_node.md")
	testWriteNode(t, "code-from-spec/a/_node.md", "depends_on:\n  - ROOT/b(interface)\n  - ROOT/b\n")
	testWriteEmptyNode(t, "code-from-spec/b/_node.md")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].Qualifier != "" {
		t.Errorf("expected no qualifier (no-qualifier wins), got %q", chain.Dependencies[0].Qualifier)
	}
}

func TestChainResolve_SameFileDifferentQualifiers(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "code-from-spec/_node.md")
	testWriteNode(t, "code-from-spec/a/_node.md", "depends_on:\n  - ROOT/b(interface)\n  - ROOT/b(constraints)\n")
	testWriteEmptyNode(t, "code-from-spec/b/_node.md")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}
	q0 := chain.Dependencies[0].Qualifier
	q1 := chain.Dependencies[1].Qualifier
	if q0 >= q1 {
		t.Errorf("expected qualifiers sorted alphabetically, got %q then %q", q0, q1)
	}
	if q0 != "constraints" || q1 != "interface" {
		t.Errorf("expected constraints and interface, got %q and %q", q0, q1)
	}
}

func TestChainResolve_DuplicateArtifact(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "code-from-spec/_node.md")
	testWriteNode(t, "code-from-spec/a/_node.md", "depends_on:\n  - ARTIFACT/b\n  - ARTIFACT/b\n")
	testWriteNode(t, "code-from-spec/b/_node.md", "output: out/lib.go\n")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 ARTIFACT dependency (deduped), got %d", len(chain.Dependencies))
	}
}

func TestChainResolve_ExternalSorted(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "code-from-spec/_node.md")
	testWriteNode(t, "code-from-spec/a/_node.md", "external:\n  - path: docs/api.yaml\n  - path: proto/v1.proto\n")

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

	testWriteEmptyNode(t, "code-from-spec/_node.md")
	testWriteEmptyNode(t, "code-from-spec/a/_node.md")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.External) != 0 {
		t.Errorf("expected empty external, got %d", len(chain.External))
	}
}

func TestChainResolve_InputResolved(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "code-from-spec/_node.md")
	testWriteNode(t, "code-from-spec/a/_node.md", "input: ARTIFACT/b\n")
	testWriteNode(t, "code-from-spec/b/_node.md", "output: out/data.json\n")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain.Input == nil {
		t.Fatal("expected input, got nil")
	}
	if chain.Input.LogicalName != "ARTIFACT/b" {
		t.Errorf("expected input logical name ARTIFACT/b, got %q", chain.Input.LogicalName)
	}
	if chain.Input.FilePath.Value != "out/data.json" {
		t.Errorf("expected input file path out/data.json, got %q", chain.Input.FilePath.Value)
	}
}

func TestChainResolve_NoInput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "code-from-spec/_node.md")
	testWriteEmptyNode(t, "code-from-spec/a/_node.md")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain.Input != nil {
		t.Errorf("expected no input, got %v", chain.Input)
	}
}

func TestChainResolve_UnrecognizedPrefixInDependsOn(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "code-from-spec/_node.md")
	testWriteNode(t, "code-from-spec/a/_node.md", "depends_on:\n  - UNKNOWN/something\n")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestChainResolve_InvalidTargetLogicalName(t *testing.T) {
	_, err := chainresolver.ChainResolve("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestChainResolve_UnreadableFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "code-from-spec/_node.md")
	if err := os.MkdirAll("code-from-spec/a", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/a/_node.md", []byte("---\n: invalid: yaml: {\n---\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnreadableFrontmatter) {
		t.Errorf("expected ErrUnreadableFrontmatter, got %v", err)
	}
}
