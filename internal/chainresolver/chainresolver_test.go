// code-from-spec: ROOT/golang/tests/chain/resolver@6tF4f5RzngnC9oiSOQQygWKaEVA
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

func testWriteNodeFile(t *testing.T, logicalName string, content string) {
	t.Helper()
	var relPath string
	if logicalName == "SPEC" {
		relPath = "code-from-spec/_node.md"
	} else {
		suffix := logicalName[len("SPEC/"):]
		relPath = "code-from-spec/" + suffix + "/_node.md"
	}
	dir := filepath.Dir(relPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testWriteNodeFile: mkdir %s: %v", dir, err)
	}
	if err := os.WriteFile(relPath, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNodeFile: write %s: %v", relPath, err)
	}
}

func testNodeContent(logicalName string) string {
	return "# " + logicalName + "\n"
}

func testNodeContentWithFrontmatter(logicalName string, fm string) string {
	return "---\n" + fm + "---\n# " + logicalName + "\n"
}

func TestChainResolve_RootAsTarget(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))

	chain, err := chainresolver.ChainResolve("SPEC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Ancestors) != 0 {
		t.Errorf("expected empty ancestors, got %d", len(chain.Ancestors))
	}
	if chain.Target == nil {
		t.Fatal("expected target, got nil")
	}
	if chain.Target.UnqualifiedLogicalName != "SPEC" {
		t.Errorf("expected target SPEC, got %s", chain.Target.UnqualifiedLogicalName)
	}
	if chain.Target.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", chain.Target.Qualifier)
	}
	if len(chain.Dependencies) != 0 {
		t.Errorf("expected empty dependencies, got %d", len(chain.Dependencies))
	}
	if chain.Input != nil {
		t.Errorf("expected input absent, got %v", chain.Input)
	}
}

func TestChainResolve_LinearChain(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))
	testWriteNodeFile(t, "SPEC/a", testNodeContent("SPEC/a"))
	testWriteNodeFile(t, "SPEC/a/b", testNodeContent("SPEC/a/b"))

	chain, err := chainresolver.ChainResolve("SPEC/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Ancestors) != 2 {
		t.Fatalf("expected 2 ancestors, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].UnqualifiedLogicalName != "SPEC" {
		t.Errorf("expected ancestor[0] SPEC, got %s", chain.Ancestors[0].UnqualifiedLogicalName)
	}
	if chain.Ancestors[1].UnqualifiedLogicalName != "SPEC/a" {
		t.Errorf("expected ancestor[1] SPEC/a, got %s", chain.Ancestors[1].UnqualifiedLogicalName)
	}
	if chain.Target == nil {
		t.Fatal("expected target, got nil")
	}
	if chain.Target.UnqualifiedLogicalName != "SPEC/a/b" {
		t.Errorf("expected target SPEC/a/b, got %s", chain.Target.UnqualifiedLogicalName)
	}
	if chain.Target.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", chain.Target.Qualifier)
	}
	if len(chain.Dependencies) != 0 {
		t.Errorf("expected empty dependencies, got %d", len(chain.Dependencies))
	}
	if chain.Input != nil {
		t.Errorf("expected input absent, got %v", chain.Input)
	}
}

func TestChainResolve_SingleParent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))
	testWriteNodeFile(t, "SPEC/a", testNodeContent("SPEC/a"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Ancestors) != 1 {
		t.Fatalf("expected 1 ancestor, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].UnqualifiedLogicalName != "SPEC" {
		t.Errorf("expected ancestor SPEC, got %s", chain.Ancestors[0].UnqualifiedLogicalName)
	}
	if chain.Target == nil {
		t.Fatal("expected target, got nil")
	}
	if chain.Target.UnqualifiedLogicalName != "SPEC/a" {
		t.Errorf("expected target SPEC/a, got %s", chain.Target.UnqualifiedLogicalName)
	}
	if chain.Target.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", chain.Target.Qualifier)
	}
	if len(chain.Dependencies) != 0 {
		t.Errorf("expected empty dependencies, got %d", len(chain.Dependencies))
	}
	if chain.Input != nil {
		t.Errorf("expected input absent, got %v", chain.Input)
	}
}

func TestChainResolve_TargetWithEmptyFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))
	testWriteNodeFile(t, "SPEC/a", testNodeContent("SPEC/a"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Ancestors) != 1 {
		t.Fatalf("expected 1 ancestor, got %d", len(chain.Ancestors))
	}
	if chain.Target == nil {
		t.Fatal("expected target, got nil")
	}
	if chain.Target.UnqualifiedLogicalName != "SPEC/a" {
		t.Errorf("expected target SPEC/a, got %s", chain.Target.UnqualifiedLogicalName)
	}
	if len(chain.Dependencies) != 0 {
		t.Errorf("expected empty dependencies, got %d", len(chain.Dependencies))
	}
	if chain.Input != nil {
		t.Errorf("expected input absent, got %v", chain.Input)
	}
}

func TestChainResolve_DependencyWithoutQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))
	testWriteNodeFile(t, "SPEC/a", testNodeContentWithFrontmatter("SPEC/a", "depends_on:\n  - SPEC/b\n"))
	testWriteNodeFile(t, "SPEC/b", testNodeContent("SPEC/b"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.UnqualifiedLogicalName != "SPEC/b" {
		t.Errorf("expected SPEC/b, got %s", dep.UnqualifiedLogicalName)
	}
	if dep.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", dep.Qualifier)
	}
}

func TestChainResolve_DependencyWithQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))
	testWriteNodeFile(t, "SPEC/a", testNodeContentWithFrontmatter("SPEC/a", "depends_on:\n  - SPEC/b(interface)\n"))
	testWriteNodeFile(t, "SPEC/b", testNodeContent("SPEC/b"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.UnqualifiedLogicalName != "SPEC/b" {
		t.Errorf("expected SPEC/b, got %s", dep.UnqualifiedLogicalName)
	}
	if dep.Qualifier == nil {
		t.Fatal("expected qualifier present, got nil")
	}
	if *dep.Qualifier != "interface" {
		t.Errorf("expected qualifier 'interface', got %s", *dep.Qualifier)
	}
}

func TestChainResolve_DependenciesSorted(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))
	testWriteNodeFile(t, "SPEC/a", testNodeContentWithFrontmatter("SPEC/a", "depends_on:\n  - SPEC/z\n  - SPEC/m\n  - SPEC/b\n"))
	testWriteNodeFile(t, "SPEC/z", testNodeContent("SPEC/z"))
	testWriteNodeFile(t, "SPEC/m", testNodeContent("SPEC/m"))
	testWriteNodeFile(t, "SPEC/b", testNodeContent("SPEC/b"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 3 {
		t.Fatalf("expected 3 dependencies, got %d", len(chain.Dependencies))
	}
	expected := []string{"SPEC/b", "SPEC/m", "SPEC/z"}
	for i, name := range expected {
		if chain.Dependencies[i].UnqualifiedLogicalName != name {
			t.Errorf("dependencies[%d]: expected %s, got %s", i, name, chain.Dependencies[i].UnqualifiedLogicalName)
		}
	}
}

func TestChainResolve_ArtifactDependencyResolved(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))
	testWriteNodeFile(t, "SPEC/a", testNodeContentWithFrontmatter("SPEC/a", "depends_on:\n  - ARTIFACT/b\n"))
	testWriteNodeFile(t, "SPEC/b", testNodeContentWithFrontmatter("SPEC/b", "output: out/lib.go\n"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.UnqualifiedLogicalName != "ARTIFACT/b" {
		t.Errorf("expected ARTIFACT/b, got %s", dep.UnqualifiedLogicalName)
	}
	if dep.FilePath.Value != "out/lib.go" {
		t.Errorf("expected file_path out/lib.go, got %s", dep.FilePath.Value)
	}
	if dep.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", dep.Qualifier)
	}
}

func TestChainResolve_ArtifactGeneratingNodeHasNoOutput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))
	testWriteNodeFile(t, "SPEC/a", testNodeContentWithFrontmatter("SPEC/a", "depends_on:\n  - ARTIFACT/b\n"))
	testWriteNodeFile(t, "SPEC/b", testNodeContent("SPEC/b"))

	_, err := chainresolver.ChainResolve("SPEC/a")
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

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))
	testWriteNodeFile(t, "SPEC/a", testNodeContentWithFrontmatter("SPEC/a", "depends_on:\n  - ARTIFACT/b\n"))
	testWriteNodeFile(t, "SPEC/b", testNodeContentWithFrontmatter("SPEC/b", "output: out/lib.go\n"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].FilePath.Value != "out/lib.go" {
		t.Errorf("expected file_path out/lib.go, got %s", chain.Dependencies[0].FilePath.Value)
	}
}

func TestChainResolve_MixedDependencies(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))
	testWriteNodeFile(t, "SPEC/a", testNodeContentWithFrontmatter("SPEC/a", "depends_on:\n  - SPEC/c\n  - ARTIFACT/b\n  - EXTERNAL/proto/api.proto\n"))
	testWriteNodeFile(t, "SPEC/b", testNodeContentWithFrontmatter("SPEC/b", "output: out/lib.go\n"))
	testWriteNodeFile(t, "SPEC/c", testNodeContent("SPEC/c"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 3 {
		t.Fatalf("expected 3 dependencies, got %d", len(chain.Dependencies))
	}

	dep0 := chain.Dependencies[0]
	if dep0.UnqualifiedLogicalName != "ARTIFACT/b" {
		t.Errorf("expected ARTIFACT/b, got %s", dep0.UnqualifiedLogicalName)
	}
	if dep0.FilePath.Value != "out/lib.go" {
		t.Errorf("expected file_path out/lib.go, got %s", dep0.FilePath.Value)
	}
	if dep0.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", dep0.Qualifier)
	}

	dep1 := chain.Dependencies[1]
	if dep1.UnqualifiedLogicalName != "EXTERNAL/proto/api.proto" {
		t.Errorf("expected EXTERNAL/proto/api.proto, got %s", dep1.UnqualifiedLogicalName)
	}
	if dep1.FilePath.Value != "proto/api.proto" {
		t.Errorf("expected file_path proto/api.proto, got %s", dep1.FilePath.Value)
	}
	if dep1.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", dep1.Qualifier)
	}

	dep2 := chain.Dependencies[2]
	if dep2.UnqualifiedLogicalName != "SPEC/c" {
		t.Errorf("expected SPEC/c, got %s", dep2.UnqualifiedLogicalName)
	}
	if dep2.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", dep2.Qualifier)
	}
}

func TestChainResolve_DedupExactDuplicate(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))
	testWriteNodeFile(t, "SPEC/a", testNodeContentWithFrontmatter("SPEC/a", "depends_on:\n  - SPEC/b\n  - SPEC/b\n"))
	testWriteNodeFile(t, "SPEC/b", testNodeContent("SPEC/b"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
}

func TestChainResolve_DedupNoQualifierSubsumesQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))
	testWriteNodeFile(t, "SPEC/a", testNodeContentWithFrontmatter("SPEC/a", "depends_on:\n  - SPEC/b\n  - SPEC/b(interface)\n"))
	testWriteNodeFile(t, "SPEC/b", testNodeContent("SPEC/b"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.UnqualifiedLogicalName != "SPEC/b" {
		t.Errorf("expected SPEC/b, got %s", dep.UnqualifiedLogicalName)
	}
	if dep.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", dep.Qualifier)
	}
}

func TestChainResolve_DedupQualifierBeforeNoQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))
	testWriteNodeFile(t, "SPEC/a", testNodeContentWithFrontmatter("SPEC/a", "depends_on:\n  - SPEC/b(interface)\n  - SPEC/b\n"))
	testWriteNodeFile(t, "SPEC/b", testNodeContent("SPEC/b"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.UnqualifiedLogicalName != "SPEC/b" {
		t.Errorf("expected SPEC/b, got %s", dep.UnqualifiedLogicalName)
	}
	if dep.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", dep.Qualifier)
	}
}

func TestChainResolve_DedupSameFileDifferentQualifiers(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))
	testWriteNodeFile(t, "SPEC/a", testNodeContentWithFrontmatter("SPEC/a", "depends_on:\n  - SPEC/b(interface)\n  - SPEC/b(constraints)\n"))
	testWriteNodeFile(t, "SPEC/b", testNodeContent("SPEC/b"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}
	dep0 := chain.Dependencies[0]
	dep1 := chain.Dependencies[1]
	if dep0.UnqualifiedLogicalName != "SPEC/b" || dep1.UnqualifiedLogicalName != "SPEC/b" {
		t.Errorf("expected both SPEC/b, got %s and %s", dep0.UnqualifiedLogicalName, dep1.UnqualifiedLogicalName)
	}
	if dep0.Qualifier == nil || dep1.Qualifier == nil {
		t.Fatal("expected qualifiers present")
	}
	if *dep0.Qualifier != "constraints" {
		t.Errorf("expected dep0 qualifier 'constraints', got %s", *dep0.Qualifier)
	}
	if *dep1.Qualifier != "interface" {
		t.Errorf("expected dep1 qualifier 'interface', got %s", *dep1.Qualifier)
	}
}

func TestChainResolve_DedupArtifactDuplicate(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))
	testWriteNodeFile(t, "SPEC/a", testNodeContentWithFrontmatter("SPEC/a", "depends_on:\n  - ARTIFACT/b\n  - ARTIFACT/b\n"))
	testWriteNodeFile(t, "SPEC/b", testNodeContentWithFrontmatter("SPEC/b", "output: out/lib.go\n"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
}

func TestChainResolve_ExternalDependencyResolved(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))
	testWriteNodeFile(t, "SPEC/a", testNodeContentWithFrontmatter("SPEC/a", "depends_on:\n  - EXTERNAL/docs/api.yaml\n"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.UnqualifiedLogicalName != "EXTERNAL/docs/api.yaml" {
		t.Errorf("expected EXTERNAL/docs/api.yaml, got %s", dep.UnqualifiedLogicalName)
	}
	if dep.FilePath.Value != "docs/api.yaml" {
		t.Errorf("expected file_path docs/api.yaml, got %s", dep.FilePath.Value)
	}
	if dep.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", dep.Qualifier)
	}
}

func TestChainResolve_MultipleExternalDependenciesSorted(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))
	testWriteNodeFile(t, "SPEC/a", testNodeContentWithFrontmatter("SPEC/a", "depends_on:\n  - EXTERNAL/proto/v1.proto\n  - EXTERNAL/docs/api.yaml\n"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].UnqualifiedLogicalName != "EXTERNAL/docs/api.yaml" {
		t.Errorf("expected EXTERNAL/docs/api.yaml first, got %s", chain.Dependencies[0].UnqualifiedLogicalName)
	}
	if chain.Dependencies[0].FilePath.Value != "docs/api.yaml" {
		t.Errorf("expected file_path docs/api.yaml, got %s", chain.Dependencies[0].FilePath.Value)
	}
	if chain.Dependencies[1].UnqualifiedLogicalName != "EXTERNAL/proto/v1.proto" {
		t.Errorf("expected EXTERNAL/proto/v1.proto second, got %s", chain.Dependencies[1].UnqualifiedLogicalName)
	}
	if chain.Dependencies[1].FilePath.Value != "proto/v1.proto" {
		t.Errorf("expected file_path proto/v1.proto, got %s", chain.Dependencies[1].FilePath.Value)
	}
}

func TestChainResolve_DedupExternalDuplicate(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))
	testWriteNodeFile(t, "SPEC/a", testNodeContentWithFrontmatter("SPEC/a", "depends_on:\n  - EXTERNAL/x.proto\n  - EXTERNAL/x.proto\n"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
}

func TestChainResolve_InputArtifactResolved(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))
	testWriteNodeFile(t, "SPEC/a", testNodeContentWithFrontmatter("SPEC/a", "input: ARTIFACT/b\n"))
	testWriteNodeFile(t, "SPEC/b", testNodeContentWithFrontmatter("SPEC/b", "output: out/data.json\n"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain.Input == nil {
		t.Fatal("expected input, got nil")
	}
	if chain.Input.UnqualifiedLogicalName != "ARTIFACT/b" {
		t.Errorf("expected ARTIFACT/b, got %s", chain.Input.UnqualifiedLogicalName)
	}
	if chain.Input.FilePath.Value != "out/data.json" {
		t.Errorf("expected file_path out/data.json, got %s", chain.Input.FilePath.Value)
	}
	if chain.Input.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", chain.Input.Qualifier)
	}
}

func TestChainResolve_ExternalInputResolved(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))
	testWriteNodeFile(t, "SPEC/a", testNodeContentWithFrontmatter("SPEC/a", "input: EXTERNAL/docs/vendor/spec.yaml\n"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain.Input == nil {
		t.Fatal("expected input, got nil")
	}
	if chain.Input.UnqualifiedLogicalName != "EXTERNAL/docs/vendor/spec.yaml" {
		t.Errorf("expected EXTERNAL/docs/vendor/spec.yaml, got %s", chain.Input.UnqualifiedLogicalName)
	}
	if chain.Input.FilePath.Value != "docs/vendor/spec.yaml" {
		t.Errorf("expected file_path docs/vendor/spec.yaml, got %s", chain.Input.FilePath.Value)
	}
	if chain.Input.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %v", chain.Input.Qualifier)
	}
}

func TestChainResolve_NoInput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))
	testWriteNodeFile(t, "SPEC/a", testNodeContent("SPEC/a"))

	chain, err := chainresolver.ChainResolve("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain.Input != nil {
		t.Errorf("expected input absent, got %v", chain.Input)
	}
}

func TestChainResolve_UnrecognizedPrefixInDependsOn(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))
	testWriteNodeFile(t, "SPEC/a", testNodeContentWithFrontmatter("SPEC/a", "depends_on:\n  - UNKNOWN/something\n"))

	_, err := chainresolver.ChainResolve("SPEC/a")
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

	testWriteNodeFile(t, "SPEC", testNodeContent("SPEC"))

	if err := os.MkdirAll("code-from-spec/a", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/a/_node.md", []byte("---\ninvalid: yaml: content: [\n---\n# SPEC/a\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, err := chainresolver.ChainResolve("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnreadableFrontmatter) {
		t.Errorf("expected ErrUnreadableFrontmatter, got %v", err)
	}
}
