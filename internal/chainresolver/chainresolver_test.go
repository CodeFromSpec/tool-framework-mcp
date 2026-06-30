// code-from-spec: SPEC/golang/tests/chain/resolver@F-Vgi6JUUtd7t-JJ2Jn9EQkuZ54
package chainresolver_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/chainresolver"
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

func testWriteNodeFile(t *testing.T, relPath string, content string) {
	t.Helper()
	dir := relPath[:len(relPath)-len("_node.md")]
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testWriteNodeFile mkdir: %v", err)
	}
	if err := os.WriteFile(relPath, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNodeFile write: %v", err)
	}
}

func testEmptyNode(logicalName string) string {
	return "# " + logicalName + "\n"
}

func testNodeWithFrontmatter(logicalName string, fm string) string {
	return "---\n" + fm + "---\n\n# " + logicalName + "\n"
}

func qualifierPtr(s string) *string {
	return &s
}

func TestChainResolve_TC1_RootAsTarget(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))

	chain, err := chainresolver.ChainResolve("SPEC/root")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Ancestors) != 0 {
		t.Errorf("expected 0 ancestors, got %d", len(chain.Ancestors))
	}
	if len(chain.Dependencies) != 0 {
		t.Errorf("expected 0 dependencies, got %d", len(chain.Dependencies))
	}
	if chain.Target.LogicalName != "SPEC/root" {
		t.Errorf("expected target SPEC/root, got %q", chain.Target.LogicalName)
	}
	if chain.Target.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", chain.Target.Qualifier)
	}
	if chain.Input != nil {
		t.Errorf("expected input absent, got %v", chain.Input)
	}
}

func TestChainResolve_TC2_LinearChainAncestorsRootFirst(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testEmptyNode("SPEC/root/a"))
	testWriteNodeFile(t, "code-from-spec/root/a/b/_node.md", testEmptyNode("SPEC/root/a/b"))

	chain, err := chainresolver.ChainResolve("SPEC/root/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Ancestors) != 2 {
		t.Fatalf("expected 2 ancestors, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].LogicalName != "SPEC/root" {
		t.Errorf("expected first ancestor SPEC/root, got %q", chain.Ancestors[0].LogicalName)
	}
	if chain.Ancestors[1].LogicalName != "SPEC/root/a" {
		t.Errorf("expected second ancestor SPEC/root/a, got %q", chain.Ancestors[1].LogicalName)
	}
	if chain.Target.LogicalName != "SPEC/root/a/b" {
		t.Errorf("expected target SPEC/root/a/b, got %q", chain.Target.LogicalName)
	}
	if chain.Target.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", chain.Target.Qualifier)
	}
}

func TestChainResolve_TC3_SingleParent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testEmptyNode("SPEC/root/a"))

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Ancestors) != 1 {
		t.Fatalf("expected 1 ancestor, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].LogicalName != "SPEC/root" {
		t.Errorf("expected ancestor SPEC/root, got %q", chain.Ancestors[0].LogicalName)
	}
	if chain.Target.LogicalName != "SPEC/root/a" {
		t.Errorf("expected target SPEC/root/a, got %q", chain.Target.LogicalName)
	}
}

func TestChainResolve_TC4_TargetWithEmptyFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testEmptyNode("SPEC/root/a"))

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Ancestors) != 1 {
		t.Fatalf("expected 1 ancestor, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].LogicalName != "SPEC/root" {
		t.Errorf("expected ancestor SPEC/root, got %q", chain.Ancestors[0].LogicalName)
	}
	if chain.Target.LogicalName != "SPEC/root/a" {
		t.Errorf("expected target SPEC/root/a, got %q", chain.Target.LogicalName)
	}
	if chain.Target.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", chain.Target.Qualifier)
	}
	if len(chain.Dependencies) != 0 {
		t.Errorf("expected 0 dependencies, got %d", len(chain.Dependencies))
	}
	if chain.Input != nil {
		t.Errorf("expected input absent, got %v", chain.Input)
	}
}

func TestChainResolve_TC5_DependencyWithoutQualifier(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testNodeWithFrontmatter("SPEC/root/a", "depends_on:\n  - SPEC/root/b\n"))
	testWriteNodeFile(t, "code-from-spec/root/b/_node.md", testEmptyNode("SPEC/root/b"))

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "SPEC/root/b" {
		t.Errorf("expected dependency SPEC/root/b, got %q", dep.LogicalName)
	}
	if dep.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", dep.Qualifier)
	}
}

func TestChainResolve_TC6_DependencyWithQualifier(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testNodeWithFrontmatter("SPEC/root/a", "depends_on:\n  - SPEC/root/b(interface)\n"))
	testWriteNodeFile(t, "code-from-spec/root/b/_node.md", testEmptyNode("SPEC/root/b"))

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "SPEC/root/b" {
		t.Errorf("expected dependency SPEC/root/b, got %q", dep.LogicalName)
	}
	if dep.Qualifier == nil || *dep.Qualifier != "interface" {
		t.Errorf("expected qualifier 'interface', got %v", dep.Qualifier)
	}
}

func TestChainResolve_TC7_DependenciesSortedByLogicalName(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testNodeWithFrontmatter("SPEC/root/a", "depends_on:\n  - SPEC/root/z\n  - SPEC/root/m\n  - SPEC/root/b\n"))
	testWriteNodeFile(t, "code-from-spec/root/z/_node.md", testEmptyNode("SPEC/root/z"))
	testWriteNodeFile(t, "code-from-spec/root/m/_node.md", testEmptyNode("SPEC/root/m"))
	testWriteNodeFile(t, "code-from-spec/root/b/_node.md", testEmptyNode("SPEC/root/b"))

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 3 {
		t.Fatalf("expected 3 dependencies, got %d", len(chain.Dependencies))
	}
	expected := []string{"SPEC/root/b", "SPEC/root/m", "SPEC/root/z"}
	for i, exp := range expected {
		if chain.Dependencies[i].LogicalName != exp {
			t.Errorf("dependencies[%d]: expected %q, got %q", i, exp, chain.Dependencies[i].LogicalName)
		}
	}
}

func TestChainResolve_TC8_ArtifactDependencyResolvedFromGeneratingNode(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testNodeWithFrontmatter("SPEC/root/a", "depends_on:\n  - ARTIFACT/root/b\n"))
	testWriteNodeFile(t, "code-from-spec/root/b/_node.md", testNodeWithFrontmatter("SPEC/root/b", "output: out/lib.go\n"))

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ARTIFACT/root/b" {
		t.Errorf("expected dependency ARTIFACT/root/b, got %q", dep.LogicalName)
	}
	if dep.Path != "out/lib.go" {
		t.Errorf("expected path 'out/lib.go', got %q", dep.Path)
	}
	if dep.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", dep.Qualifier)
	}
}

func TestChainResolve_TC9_ArtifactGeneratingNodeHasNoOutput(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testNodeWithFrontmatter("SPEC/root/a", "depends_on:\n  - ARTIFACT/root/b\n"))
	testWriteNodeFile(t, "code-from-spec/root/b/_node.md", testEmptyNode("SPEC/root/b"))

	_, err := chainresolver.ChainResolve("SPEC/root/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestChainResolve_TC10_ArtifactFileDoesNotExistOnDisk(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testNodeWithFrontmatter("SPEC/root/a", "depends_on:\n  - ARTIFACT/root/b\n"))
	testWriteNodeFile(t, "code-from-spec/root/b/_node.md", testNodeWithFrontmatter("SPEC/root/b", "output: out/lib.go\n"))

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ARTIFACT/root/b" {
		t.Errorf("expected dependency ARTIFACT/root/b, got %q", dep.LogicalName)
	}
	if dep.Path != "out/lib.go" {
		t.Errorf("expected path 'out/lib.go', got %q", dep.Path)
	}
}

func TestChainResolve_TC11_MixedDependenciesSorted(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testNodeWithFrontmatter("SPEC/root/a", "depends_on:\n  - SPEC/root/c\n  - ARTIFACT/root/b\n  - EXTERNAL/proto/api.proto\n"))
	testWriteNodeFile(t, "code-from-spec/root/b/_node.md", testNodeWithFrontmatter("SPEC/root/b", "output: out/lib.go\n"))
	testWriteNodeFile(t, "code-from-spec/root/c/_node.md", testEmptyNode("SPEC/root/c"))

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 3 {
		t.Fatalf("expected 3 dependencies, got %d", len(chain.Dependencies))
	}
	expected := []string{"ARTIFACT/root/b", "EXTERNAL/proto/api.proto", "SPEC/root/c"}
	for i, exp := range expected {
		if chain.Dependencies[i].LogicalName != exp {
			t.Errorf("dependencies[%d]: expected %q, got %q", i, exp, chain.Dependencies[i].LogicalName)
		}
	}
}

func TestChainResolve_TC12_ExactDuplicateDependency(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testNodeWithFrontmatter("SPEC/root/a", "depends_on:\n  - SPEC/root/b\n  - SPEC/root/b\n"))
	testWriteNodeFile(t, "code-from-spec/root/b/_node.md", testEmptyNode("SPEC/root/b"))

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
}

func TestChainResolve_TC13_NoQualifierSubsumesQualifier(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testNodeWithFrontmatter("SPEC/root/a", "depends_on:\n  - SPEC/root/b\n  - SPEC/root/b(interface)\n"))
	testWriteNodeFile(t, "code-from-spec/root/b/_node.md", testEmptyNode("SPEC/root/b"))

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "SPEC/root/b" {
		t.Errorf("expected dependency SPEC/root/b, got %q", dep.LogicalName)
	}
	if dep.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", dep.Qualifier)
	}
}

func TestChainResolve_TC14_QualifierBeforeNoQualifier_NoQualifierWins(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testNodeWithFrontmatter("SPEC/root/a", "depends_on:\n  - SPEC/root/b(interface)\n  - SPEC/root/b\n"))
	testWriteNodeFile(t, "code-from-spec/root/b/_node.md", testEmptyNode("SPEC/root/b"))

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "SPEC/root/b" {
		t.Errorf("expected dependency SPEC/root/b, got %q", dep.LogicalName)
	}
	if dep.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", dep.Qualifier)
	}
}

func TestChainResolve_TC15_SameFileDifferentQualifiers_BothKept(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testNodeWithFrontmatter("SPEC/root/a", "depends_on:\n  - SPEC/root/b(interface)\n  - SPEC/root/b(constraints)\n"))
	testWriteNodeFile(t, "code-from-spec/root/b/_node.md", testEmptyNode("SPEC/root/b"))

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].Qualifier == nil || *chain.Dependencies[0].Qualifier != "constraints" {
		t.Errorf("expected first qualifier 'constraints', got %v", chain.Dependencies[0].Qualifier)
	}
	if chain.Dependencies[1].Qualifier == nil || *chain.Dependencies[1].Qualifier != "interface" {
		t.Errorf("expected second qualifier 'interface', got %v", chain.Dependencies[1].Qualifier)
	}
}

func TestChainResolve_TC16_DuplicateArtifactDependency(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testNodeWithFrontmatter("SPEC/root/a", "depends_on:\n  - ARTIFACT/root/b\n  - ARTIFACT/root/b\n"))
	testWriteNodeFile(t, "code-from-spec/root/b/_node.md", testNodeWithFrontmatter("SPEC/root/b", "output: out/lib.go\n"))

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
}

func TestChainResolve_TC17_ExternalDependencyResolvedToPath(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testNodeWithFrontmatter("SPEC/root/a", "depends_on:\n  - EXTERNAL/docs/api.yaml\n"))

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "EXTERNAL/docs/api.yaml" {
		t.Errorf("expected dependency EXTERNAL/docs/api.yaml, got %q", dep.LogicalName)
	}
	if dep.Path != "docs/api.yaml" {
		t.Errorf("expected path 'docs/api.yaml', got %q", dep.Path)
	}
	if dep.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", dep.Qualifier)
	}
}

func TestChainResolve_TC18_MultipleExternalDependenciesSorted(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testNodeWithFrontmatter("SPEC/root/a", "depends_on:\n  - EXTERNAL/proto/v1.proto\n  - EXTERNAL/docs/api.yaml\n"))

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].LogicalName != "EXTERNAL/docs/api.yaml" {
		t.Errorf("expected first dependency EXTERNAL/docs/api.yaml, got %q", chain.Dependencies[0].LogicalName)
	}
	if chain.Dependencies[1].LogicalName != "EXTERNAL/proto/v1.proto" {
		t.Errorf("expected second dependency EXTERNAL/proto/v1.proto, got %q", chain.Dependencies[1].LogicalName)
	}
}

func TestChainResolve_TC19_DuplicateExternalDependency(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testNodeWithFrontmatter("SPEC/root/a", "depends_on:\n  - EXTERNAL/x.proto\n  - EXTERNAL/x.proto\n"))

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
}

func TestChainResolve_TC20_InputResolvedFromGeneratingNode(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testNodeWithFrontmatter("SPEC/root/a", "input: ARTIFACT/root/b\n"))
	testWriteNodeFile(t, "code-from-spec/root/b/_node.md", testNodeWithFrontmatter("SPEC/root/b", "output: out/data.json\n"))

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chain.Input == nil {
		t.Fatal("expected input, got nil")
	}
	if chain.Input.LogicalName != "ARTIFACT/root/b" {
		t.Errorf("expected input ARTIFACT/root/b, got %q", chain.Input.LogicalName)
	}
	if chain.Input.Path != "out/data.json" {
		t.Errorf("expected input path 'out/data.json', got %q", chain.Input.Path)
	}
	if chain.Input.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", chain.Input.Qualifier)
	}
}

func TestChainResolve_TC21_ExternalInputResolvedToPath(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testNodeWithFrontmatter("SPEC/root/a", "input: EXTERNAL/docs/vendor/spec.yaml\n"))

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chain.Input == nil {
		t.Fatal("expected input, got nil")
	}
	if chain.Input.LogicalName != "EXTERNAL/docs/vendor/spec.yaml" {
		t.Errorf("expected input EXTERNAL/docs/vendor/spec.yaml, got %q", chain.Input.LogicalName)
	}
	if chain.Input.Path != "docs/vendor/spec.yaml" {
		t.Errorf("expected input path 'docs/vendor/spec.yaml', got %q", chain.Input.Path)
	}
	if chain.Input.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", chain.Input.Qualifier)
	}
}

func TestChainResolve_TC22_SpecInputResolved(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testNodeWithFrontmatter("SPEC/root/a", "input: SPEC/root/b\n"))
	testWriteNodeFile(t, "code-from-spec/root/b/_node.md", testEmptyNode("SPEC/root/b"))

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chain.Input == nil {
		t.Fatal("expected input, got nil")
	}
	if chain.Input.LogicalName != "SPEC/root/b" {
		t.Errorf("expected input SPEC/root/b, got %q", chain.Input.LogicalName)
	}
	if chain.Input.Path != "code-from-spec/root/b/_node.md" {
		t.Errorf("expected input path 'code-from-spec/root/b/_node.md', got %q", chain.Input.Path)
	}
	if chain.Input.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", chain.Input.Qualifier)
	}
}

func TestChainResolve_TC23_SpecInputWithQualifier(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testNodeWithFrontmatter("SPEC/root/a", "input: SPEC/root/b(acceptance-tests)\n"))
	testWriteNodeFile(t, "code-from-spec/root/b/_node.md", testEmptyNode("SPEC/root/b"))

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chain.Input == nil {
		t.Fatal("expected input, got nil")
	}
	if chain.Input.LogicalName != "SPEC/root/b" {
		t.Errorf("expected input SPEC/root/b, got %q", chain.Input.LogicalName)
	}
	if chain.Input.Path != "code-from-spec/root/b/_node.md" {
		t.Errorf("expected input path 'code-from-spec/root/b/_node.md', got %q", chain.Input.Path)
	}
	if chain.Input.Qualifier == nil || *chain.Input.Qualifier != "acceptance-tests" {
		t.Errorf("expected qualifier 'acceptance-tests', got %v", chain.Input.Qualifier)
	}
}

func TestChainResolve_TC24_NoInput_Absent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testEmptyNode("SPEC/root/a"))

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chain.Input != nil {
		t.Errorf("expected input absent, got %v", chain.Input)
	}
}

func TestChainResolve_TC25_UnrecognizedPrefixInDependsOn(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testNodeWithFrontmatter("SPEC/root/a", "depends_on:\n  - UNKNOWN/something\n"))

	_, err := chainresolver.ChainResolve("SPEC/root/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestChainResolve_TC26_InvalidTargetLogicalName(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := chainresolver.ChainResolve("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestChainResolve_TC27_InputArtifactGeneratingNodeNotFound(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", testNodeWithFrontmatter("SPEC/root/a", "input: ARTIFACT/root/missing\n"))

	_, err := chainresolver.ChainResolve("SPEC/root/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestChainResolve_TC28_UnreadableFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/root/_node.md", testEmptyNode("SPEC/root"))
	testWriteNodeFile(t, "code-from-spec/root/a/_node.md", "---\ninvalid: yaml: content: [\n---\n\n# SPEC/root/a\n")

	_, err := chainresolver.ChainResolve("SPEC/root/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnreadableFrontmatter) {
		t.Errorf("expected ErrUnreadableFrontmatter, got %v", err)
	}
}

var _ = qualifierPtr
