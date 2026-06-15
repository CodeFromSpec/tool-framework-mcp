// code-from-spec: SPEC/golang/tests/chain/resolver@ze1al_f55_KXDgwbvf2krh3WTTE
package chainresolver_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainresolver"
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

func testNodeWithFrontmatter(logicalName string, frontmatter string) string {
	return "---\n" + frontmatter + "---\n\n# " + logicalName + "\n"
}

func TestChainResolve_TC1_RootAsTarget(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))

	chain, err := chainresolver.ChainResolve("SPEC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Ancestors) != 0 {
		t.Errorf("expected 0 ancestors, got %d", len(chain.Ancestors))
	}
	if len(chain.Dependencies) != 0 {
		t.Errorf("expected 0 dependencies, got %d", len(chain.Dependencies))
	}
	if chain.Target == nil {
		t.Fatal("expected target, got nil")
	}
	if chain.Target.UnqualifiedLogicalName != "SPEC" {
		t.Errorf("expected target SPEC, got %q", chain.Target.UnqualifiedLogicalName)
	}
	if chain.Target.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %q", *chain.Target.Qualifier)
	}
	if chain.Input != nil {
		t.Errorf("expected input absent, got %v", chain.Input)
	}
}

func TestChainResolve_TC2_LinearChainAncestorsRootFirst(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testEmptyNode("SPEC/a"))
	testWriteNodeFile(t, "code-from-spec/a/b/_node.md", testEmptyNode("SPEC/a/b"))

	chain, err := chainresolver.ChainResolve("SPEC/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Ancestors) != 2 {
		t.Fatalf("expected 2 ancestors, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].UnqualifiedLogicalName != "SPEC" {
		t.Errorf("expected first ancestor SPEC, got %q", chain.Ancestors[0].UnqualifiedLogicalName)
	}
	if chain.Ancestors[1].UnqualifiedLogicalName != "SPEC/a" {
		t.Errorf("expected second ancestor SPEC/a, got %q", chain.Ancestors[1].UnqualifiedLogicalName)
	}
	if chain.Target == nil {
		t.Fatal("expected target, got nil")
	}
	if chain.Target.UnqualifiedLogicalName != "SPEC/a/b" {
		t.Errorf("expected target SPEC/a/b, got %q", chain.Target.UnqualifiedLogicalName)
	}
	if chain.Target.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %q", *chain.Target.Qualifier)
	}
}

func TestChainResolve_TC3_SingleParent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testEmptyNode("SPEC/a"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Ancestors) != 1 {
		t.Fatalf("expected 1 ancestor, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].UnqualifiedLogicalName != "SPEC" {
		t.Errorf("expected ancestor SPEC, got %q", chain.Ancestors[0].UnqualifiedLogicalName)
	}
	if chain.Target == nil {
		t.Fatal("expected target, got nil")
	}
	if chain.Target.UnqualifiedLogicalName != "SPEC/a" {
		t.Errorf("expected target SPEC/a, got %q", chain.Target.UnqualifiedLogicalName)
	}
}

func TestChainResolve_TC4_TargetWithEmptyFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testEmptyNode("SPEC/a"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Ancestors) != 1 {
		t.Fatalf("expected 1 ancestor, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].UnqualifiedLogicalName != "SPEC" {
		t.Errorf("expected ancestor SPEC, got %q", chain.Ancestors[0].UnqualifiedLogicalName)
	}
	if chain.Target == nil {
		t.Fatal("expected target, got nil")
	}
	if chain.Target.UnqualifiedLogicalName != "SPEC/a" {
		t.Errorf("expected target SPEC/a, got %q", chain.Target.UnqualifiedLogicalName)
	}
	if chain.Target.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %q", *chain.Target.Qualifier)
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

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeWithFrontmatter("SPEC/a", "depends_on:\n  - SPEC/b\n"))
	testWriteNodeFile(t, "code-from-spec/b/_node.md", testEmptyNode("SPEC/b"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.UnqualifiedLogicalName != "SPEC/b" {
		t.Errorf("expected dependency SPEC/b, got %q", dep.UnqualifiedLogicalName)
	}
	if dep.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %q", *dep.Qualifier)
	}
}

func TestChainResolve_TC6_DependencyWithQualifier(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeWithFrontmatter("SPEC/a", "depends_on:\n  - SPEC/b(interface)\n"))
	testWriteNodeFile(t, "code-from-spec/b/_node.md", testEmptyNode("SPEC/b"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.UnqualifiedLogicalName != "SPEC/b" {
		t.Errorf("expected dependency SPEC/b, got %q", dep.UnqualifiedLogicalName)
	}
	if dep.Qualifier == nil {
		t.Fatal("expected qualifier, got nil")
	}
	if *dep.Qualifier != "interface" {
		t.Errorf("expected qualifier 'interface', got %q", *dep.Qualifier)
	}
}

func TestChainResolve_TC7_DependenciesSortedByLogicalName(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeWithFrontmatter("SPEC/a", "depends_on:\n  - SPEC/z\n  - SPEC/m\n  - SPEC/b\n"))
	testWriteNodeFile(t, "code-from-spec/z/_node.md", testEmptyNode("SPEC/z"))
	testWriteNodeFile(t, "code-from-spec/m/_node.md", testEmptyNode("SPEC/m"))
	testWriteNodeFile(t, "code-from-spec/b/_node.md", testEmptyNode("SPEC/b"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 3 {
		t.Fatalf("expected 3 dependencies, got %d", len(chain.Dependencies))
	}
	expected := []string{"SPEC/b", "SPEC/m", "SPEC/z"}
	for i, exp := range expected {
		if chain.Dependencies[i].UnqualifiedLogicalName != exp {
			t.Errorf("dependencies[%d]: expected %q, got %q", i, exp, chain.Dependencies[i].UnqualifiedLogicalName)
		}
	}
}

func TestChainResolve_TC8_ArtifactDependencyResolvedFromGeneratingNode(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeWithFrontmatter("SPEC/a", "depends_on:\n  - ARTIFACT/b\n"))
	testWriteNodeFile(t, "code-from-spec/b/_node.md", testNodeWithFrontmatter("SPEC/b", "output: out/lib.go\n"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.UnqualifiedLogicalName != "ARTIFACT/b" {
		t.Errorf("expected dependency ARTIFACT/b, got %q", dep.UnqualifiedLogicalName)
	}
	if dep.FilePath.Value != "out/lib.go" {
		t.Errorf("expected file_path 'out/lib.go', got %q", dep.FilePath.Value)
	}
	if dep.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %q", *dep.Qualifier)
	}
}

func TestChainResolve_TC9_ArtifactGeneratingNodeHasNoOutput(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeWithFrontmatter("SPEC/a", "depends_on:\n  - ARTIFACT/b\n"))
	testWriteNodeFile(t, "code-from-spec/b/_node.md", testEmptyNode("SPEC/b"))

	_, err := chainresolver.ChainResolve("SPEC/a")
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

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeWithFrontmatter("SPEC/a", "depends_on:\n  - ARTIFACT/b\n"))
	testWriteNodeFile(t, "code-from-spec/b/_node.md", testNodeWithFrontmatter("SPEC/b", "output: out/lib.go\n"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.UnqualifiedLogicalName != "ARTIFACT/b" {
		t.Errorf("expected dependency ARTIFACT/b, got %q", dep.UnqualifiedLogicalName)
	}
	if dep.FilePath.Value != "out/lib.go" {
		t.Errorf("expected file_path 'out/lib.go', got %q", dep.FilePath.Value)
	}
}

func TestChainResolve_TC11_MixedDependenciesSorted(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeWithFrontmatter("SPEC/a", "depends_on:\n  - SPEC/c\n  - ARTIFACT/b\n  - EXTERNAL/proto/api.proto\n"))
	testWriteNodeFile(t, "code-from-spec/b/_node.md", testNodeWithFrontmatter("SPEC/b", "output: out/lib.go\n"))
	testWriteNodeFile(t, "code-from-spec/c/_node.md", testEmptyNode("SPEC/c"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 3 {
		t.Fatalf("expected 3 dependencies, got %d", len(chain.Dependencies))
	}
	expected := []string{"ARTIFACT/b", "EXTERNAL/proto/api.proto", "SPEC/c"}
	for i, exp := range expected {
		if chain.Dependencies[i].UnqualifiedLogicalName != exp {
			t.Errorf("dependencies[%d]: expected %q, got %q", i, exp, chain.Dependencies[i].UnqualifiedLogicalName)
		}
	}
}

func TestChainResolve_TC12_ExactDuplicateDependency(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeWithFrontmatter("SPEC/a", "depends_on:\n  - SPEC/b\n  - SPEC/b\n"))
	testWriteNodeFile(t, "code-from-spec/b/_node.md", testEmptyNode("SPEC/b"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
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

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeWithFrontmatter("SPEC/a", "depends_on:\n  - SPEC/b\n  - SPEC/b(interface)\n"))
	testWriteNodeFile(t, "code-from-spec/b/_node.md", testEmptyNode("SPEC/b"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.UnqualifiedLogicalName != "SPEC/b" {
		t.Errorf("expected dependency SPEC/b, got %q", dep.UnqualifiedLogicalName)
	}
	if dep.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %q", *dep.Qualifier)
	}
}

func TestChainResolve_TC14_QualifierBeforeNoQualifier_NoQualifierWins(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeWithFrontmatter("SPEC/a", "depends_on:\n  - SPEC/b(interface)\n  - SPEC/b\n"))
	testWriteNodeFile(t, "code-from-spec/b/_node.md", testEmptyNode("SPEC/b"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.UnqualifiedLogicalName != "SPEC/b" {
		t.Errorf("expected dependency SPEC/b, got %q", dep.UnqualifiedLogicalName)
	}
	if dep.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %q", *dep.Qualifier)
	}
}

func TestChainResolve_TC15_SameFileDifferentQualifiers_BothKept(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeWithFrontmatter("SPEC/a", "depends_on:\n  - SPEC/b(interface)\n  - SPEC/b(constraints)\n"))
	testWriteNodeFile(t, "code-from-spec/b/_node.md", testEmptyNode("SPEC/b"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
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

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeWithFrontmatter("SPEC/a", "depends_on:\n  - ARTIFACT/b\n  - ARTIFACT/b\n"))
	testWriteNodeFile(t, "code-from-spec/b/_node.md", testNodeWithFrontmatter("SPEC/b", "output: out/lib.go\n"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
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

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeWithFrontmatter("SPEC/a", "depends_on:\n  - EXTERNAL/docs/api.yaml\n"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.UnqualifiedLogicalName != "EXTERNAL/docs/api.yaml" {
		t.Errorf("expected dependency EXTERNAL/docs/api.yaml, got %q", dep.UnqualifiedLogicalName)
	}
	if dep.FilePath.Value != "docs/api.yaml" {
		t.Errorf("expected file_path 'docs/api.yaml', got %q", dep.FilePath.Value)
	}
	if dep.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %q", *dep.Qualifier)
	}
}

func TestChainResolve_TC18_MultipleExternalDependenciesSorted(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeWithFrontmatter("SPEC/a", "depends_on:\n  - EXTERNAL/proto/v1.proto\n  - EXTERNAL/docs/api.yaml\n"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].UnqualifiedLogicalName != "EXTERNAL/docs/api.yaml" {
		t.Errorf("expected first dependency EXTERNAL/docs/api.yaml, got %q", chain.Dependencies[0].UnqualifiedLogicalName)
	}
	if chain.Dependencies[1].UnqualifiedLogicalName != "EXTERNAL/proto/v1.proto" {
		t.Errorf("expected second dependency EXTERNAL/proto/v1.proto, got %q", chain.Dependencies[1].UnqualifiedLogicalName)
	}
}

func TestChainResolve_TC19_DuplicateExternalDependency(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeWithFrontmatter("SPEC/a", "depends_on:\n  - EXTERNAL/x.proto\n  - EXTERNAL/x.proto\n"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
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

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeWithFrontmatter("SPEC/a", "input: ARTIFACT/b\n"))
	testWriteNodeFile(t, "code-from-spec/b/_node.md", testNodeWithFrontmatter("SPEC/b", "output: out/data.json\n"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chain.Input == nil {
		t.Fatal("expected input, got nil")
	}
	if chain.Input.UnqualifiedLogicalName != "ARTIFACT/b" {
		t.Errorf("expected input ARTIFACT/b, got %q", chain.Input.UnqualifiedLogicalName)
	}
	if chain.Input.FilePath.Value != "out/data.json" {
		t.Errorf("expected input file_path 'out/data.json', got %q", chain.Input.FilePath.Value)
	}
	if chain.Input.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %q", *chain.Input.Qualifier)
	}
}

func TestChainResolve_TC21_ExternalInputResolvedToPath(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeWithFrontmatter("SPEC/a", "input: EXTERNAL/docs/vendor/spec.yaml\n"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chain.Input == nil {
		t.Fatal("expected input, got nil")
	}
	if chain.Input.UnqualifiedLogicalName != "EXTERNAL/docs/vendor/spec.yaml" {
		t.Errorf("expected input EXTERNAL/docs/vendor/spec.yaml, got %q", chain.Input.UnqualifiedLogicalName)
	}
	if chain.Input.FilePath.Value != "docs/vendor/spec.yaml" {
		t.Errorf("expected input file_path 'docs/vendor/spec.yaml', got %q", chain.Input.FilePath.Value)
	}
	if chain.Input.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %q", *chain.Input.Qualifier)
	}
}

func TestChainResolve_TC22_NoInput_Absent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testEmptyNode("SPEC/a"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chain.Input != nil {
		t.Errorf("expected input absent, got %v", chain.Input)
	}
}

func TestChainResolve_TC23_UnrecognizedPrefixInDependsOn(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeWithFrontmatter("SPEC/a", "depends_on:\n  - UNKNOWN/something\n"))

	_, err := chainresolver.ChainResolve("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestChainResolve_TC24_InvalidTargetLogicalName(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := chainresolver.ChainResolve("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestChainResolve_TC25_UnreadableFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", testEmptyNode("SPEC"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "---\ninvalid: yaml: content: [\n---\n\n# SPEC/a\n")

	_, err := chainresolver.ChainResolve("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnreadableFrontmatter) {
		t.Errorf("expected ErrUnreadableFrontmatter, got %v", err)
	}
}
