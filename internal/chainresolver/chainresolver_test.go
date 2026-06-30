// code-from-spec: SPEC/golang/test/cases/chain/resolver@58sKTQij7HBUNrYiej09vRthXeI
package chainresolver_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"
)

func TestChainResolve_RootAsTarget(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()

	chain, err := chainresolver.ChainResolve("SPEC/root")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Ancestors) != 0 {
		t.Errorf("expected no ancestors, got %d", len(chain.Ancestors))
	}
	if len(chain.Dependencies) != 0 {
		t.Errorf("expected no dependencies, got %d", len(chain.Dependencies))
	}
	if chain.Target.LogicalName != "SPEC/root" {
		t.Errorf("expected target SPEC/root, got %q", chain.Target.LogicalName)
	}
	if chain.Target.Qualifier != nil {
		t.Errorf("expected nil qualifier, got %v", chain.Target.Qualifier)
	}
	if chain.Input != nil {
		t.Errorf("expected no input, got %v", chain.Input)
	}
}

func TestChainResolve_LinearChain(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	testutils.CreateSpecNode(t, "SPEC/root/a").Write()
	testutils.CreateSpecNode(t, "SPEC/root/a/b").Write()

	chain, err := chainresolver.ChainResolve("SPEC/root/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Ancestors) != 2 {
		t.Fatalf("expected 2 ancestors, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].LogicalName != "SPEC/root" {
		t.Errorf("expected ancestor[0] = SPEC/root, got %q", chain.Ancestors[0].LogicalName)
	}
	if chain.Ancestors[1].LogicalName != "SPEC/root/a" {
		t.Errorf("expected ancestor[1] = SPEC/root/a, got %q", chain.Ancestors[1].LogicalName)
	}
	if chain.Target.LogicalName != "SPEC/root/a/b" {
		t.Errorf("expected target SPEC/root/a/b, got %q", chain.Target.LogicalName)
	}
}

func TestChainResolve_SingleParent(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	testutils.CreateSpecNode(t, "SPEC/root/a").Write()

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

func TestChainResolve_EmptyFrontmatter(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	testutils.CreateSpecNode(t, "SPEC/root/a").Write()

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
	if len(chain.Dependencies) != 0 {
		t.Errorf("expected no dependencies, got %d", len(chain.Dependencies))
	}
	if chain.Input != nil {
		t.Errorf("expected no input, got %v", chain.Input)
	}
}

func TestChainResolve_DependencyWithoutQualifier(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	b := testutils.CreateSpecNode(t, "SPEC/root/a")
	b.AddDependsOn("SPEC/root/b")
	b.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()

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
		t.Errorf("expected nil qualifier, got %v", dep.Qualifier)
	}
}

func TestChainResolve_DependencyWithQualifier(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	b := testutils.CreateSpecNode(t, "SPEC/root/a")
	b.AddDependsOn("SPEC/root/b(interface)")
	b.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()

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

func TestChainResolve_DependenciesSortedByLogicalName(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddDependsOn("SPEC/root/z")
	a.AddDependsOn("SPEC/root/m")
	a.AddDependsOn("SPEC/root/b")
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/z").Write()
	testutils.CreateSpecNode(t, "SPEC/root/m").Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()

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
			t.Errorf("dependency[%d]: expected %q, got %q", i, exp, chain.Dependencies[i].LogicalName)
		}
	}
}

func TestChainResolve_ArtifactDependencyResolved(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddDependsOn("ARTIFACT/root/b")
	a.Write()
	bNode := testutils.CreateSpecNode(t, "SPEC/root/b")
	bNode.SetOutput("out/lib.go")
	bNode.Write()

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ARTIFACT/root/b" {
		t.Errorf("expected ARTIFACT/root/b, got %q", dep.LogicalName)
	}
	if dep.Path != "out/lib.go" {
		t.Errorf("expected path out/lib.go, got %q", dep.Path)
	}
}

func TestChainResolve_ArtifactGeneratingNodeNoOutput(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddDependsOn("ARTIFACT/root/b")
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()

	_, err := chainresolver.ChainResolve("SPEC/root/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestChainResolve_ArtifactFileDoesNotExistOnDisk(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddDependsOn("ARTIFACT/root/b")
	a.Write()
	bNode := testutils.CreateSpecNode(t, "SPEC/root/b")
	bNode.SetOutput("out/lib.go")
	bNode.Write()

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].Path != "out/lib.go" {
		t.Errorf("expected path out/lib.go, got %q", chain.Dependencies[0].Path)
	}
}

func TestChainResolve_MixedDependencies(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddDependsOn("SPEC/root/c")
	a.AddDependsOn("ARTIFACT/root/b")
	a.AddDependsOn("EXTERNAL/proto/api.proto")
	a.Write()
	bNode := testutils.CreateSpecNode(t, "SPEC/root/b")
	bNode.SetOutput("out/lib.go")
	bNode.Write()
	testutils.CreateSpecNode(t, "SPEC/root/c").Write()

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
			t.Errorf("dependency[%d]: expected %q, got %q", i, exp, chain.Dependencies[i].LogicalName)
		}
	}
}

func TestChainResolve_ExactDuplicate(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddDependsOn("SPEC/root/b")
	a.AddDependsOn("SPEC/root/b")
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency (deduped), got %d", len(chain.Dependencies))
	}
}

func TestChainResolve_NoQualifierSubsumesQualifier(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddDependsOn("SPEC/root/b")
	a.AddDependsOn("SPEC/root/b(interface)")
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].Qualifier != nil {
		t.Errorf("expected nil qualifier, got %v", chain.Dependencies[0].Qualifier)
	}
}

func TestChainResolve_QualifierBeforeNoQualifier(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddDependsOn("SPEC/root/b(interface)")
	a.AddDependsOn("SPEC/root/b")
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].Qualifier != nil {
		t.Errorf("expected nil qualifier (no-qualifier wins), got %v", chain.Dependencies[0].Qualifier)
	}
}

func TestChainResolve_SameFileDifferentQualifiers(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddDependsOn("SPEC/root/b(interface)")
	a.AddDependsOn("SPEC/root/b(constraints)")
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].Qualifier == nil || *chain.Dependencies[0].Qualifier != "constraints" {
		t.Errorf("expected qualifier 'constraints' at index 0, got %v", chain.Dependencies[0].Qualifier)
	}
	if chain.Dependencies[1].Qualifier == nil || *chain.Dependencies[1].Qualifier != "interface" {
		t.Errorf("expected qualifier 'interface' at index 1, got %v", chain.Dependencies[1].Qualifier)
	}
}

func TestChainResolve_DuplicateArtifact(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddDependsOn("ARTIFACT/root/b")
	a.AddDependsOn("ARTIFACT/root/b")
	a.Write()
	bNode := testutils.CreateSpecNode(t, "SPEC/root/b")
	bNode.SetOutput("out/lib.go")
	bNode.Write()

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 ARTIFACT dependency (deduped), got %d", len(chain.Dependencies))
	}
}

func TestChainResolve_ExternalDependencyResolvedToPath(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddDependsOn("EXTERNAL/docs/api.yaml")
	a.Write()

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "EXTERNAL/docs/api.yaml" {
		t.Errorf("expected EXTERNAL/docs/api.yaml, got %q", dep.LogicalName)
	}
	if dep.Path != "docs/api.yaml" {
		t.Errorf("expected path docs/api.yaml, got %q", dep.Path)
	}
	if dep.Qualifier != nil {
		t.Errorf("expected nil qualifier, got %v", dep.Qualifier)
	}
}

func TestChainResolve_MultipleExternalDependenciesSorted(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddDependsOn("EXTERNAL/proto/v1.proto")
	a.AddDependsOn("EXTERNAL/docs/api.yaml")
	a.Write()

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].LogicalName != "EXTERNAL/docs/api.yaml" {
		t.Errorf("expected EXTERNAL/docs/api.yaml at index 0, got %q", chain.Dependencies[0].LogicalName)
	}
	if chain.Dependencies[1].LogicalName != "EXTERNAL/proto/v1.proto" {
		t.Errorf("expected EXTERNAL/proto/v1.proto at index 1, got %q", chain.Dependencies[1].LogicalName)
	}
}

func TestChainResolve_DuplicateExternal(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddDependsOn("EXTERNAL/x.proto")
	a.AddDependsOn("EXTERNAL/x.proto")
	a.Write()

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 EXTERNAL dependency (deduped), got %d", len(chain.Dependencies))
	}
}

func TestChainResolve_InputArtifactResolved(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetInput("ARTIFACT/root/b")
	a.Write()
	bNode := testutils.CreateSpecNode(t, "SPEC/root/b")
	bNode.SetOutput("out/data.json")
	bNode.Write()

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain.Input == nil {
		t.Fatal("expected input, got nil")
	}
	if chain.Input.LogicalName != "ARTIFACT/root/b" {
		t.Errorf("expected ARTIFACT/root/b, got %q", chain.Input.LogicalName)
	}
	if chain.Input.Path != "out/data.json" {
		t.Errorf("expected path out/data.json, got %q", chain.Input.Path)
	}
}

func TestChainResolve_ExternalInputResolvedToPath(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetInput("EXTERNAL/docs/vendor/spec.yaml")
	a.Write()

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain.Input == nil {
		t.Fatal("expected input, got nil")
	}
	if chain.Input.LogicalName != "EXTERNAL/docs/vendor/spec.yaml" {
		t.Errorf("expected EXTERNAL/docs/vendor/spec.yaml, got %q", chain.Input.LogicalName)
	}
	if chain.Input.Path != "docs/vendor/spec.yaml" {
		t.Errorf("expected path docs/vendor/spec.yaml, got %q", chain.Input.Path)
	}
}

func TestChainResolve_SpecInputResolved(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetInput("SPEC/root/b")
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain.Input == nil {
		t.Fatal("expected input, got nil")
	}
	if chain.Input.LogicalName != "SPEC/root/b" {
		t.Errorf("expected SPEC/root/b, got %q", chain.Input.LogicalName)
	}
	if chain.Input.Path != "code-from-spec/root/b/_node.md" {
		t.Errorf("expected path code-from-spec/root/b/_node.md, got %q", chain.Input.Path)
	}
	if chain.Input.Qualifier != nil {
		t.Errorf("expected nil qualifier, got %v", chain.Input.Qualifier)
	}
}

func TestChainResolve_SpecInputWithQualifier(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetInput("SPEC/root/b(acceptance-tests)")
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain.Input == nil {
		t.Fatal("expected input, got nil")
	}
	if chain.Input.LogicalName != "SPEC/root/b" {
		t.Errorf("expected SPEC/root/b, got %q", chain.Input.LogicalName)
	}
	if chain.Input.Path != "code-from-spec/root/b/_node.md" {
		t.Errorf("expected path code-from-spec/root/b/_node.md, got %q", chain.Input.Path)
	}
	if chain.Input.Qualifier == nil || *chain.Input.Qualifier != "acceptance-tests" {
		t.Errorf("expected qualifier 'acceptance-tests', got %v", chain.Input.Qualifier)
	}
}

func TestChainResolve_NoInput(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	testutils.CreateSpecNode(t, "SPEC/root/a").Write()

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain.Input != nil {
		t.Errorf("expected no input, got %v", chain.Input)
	}
}

func TestChainResolve_UnrecognizedPrefixInDependsOn(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddDependsOn("UNKNOWN/something")
	a.Write()

	_, err := chainresolver.ChainResolve("SPEC/root/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestChainResolve_InvalidTargetLogicalName(t *testing.T) {
	testutils.Chdir(t)

	_, err := chainresolver.ChainResolve("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parsing.ErrUnrecognizedPrefix) {
		t.Errorf("expected ErrUnrecognizedPrefix, got %v", err)
	}
}

func TestChainResolve_InputArtifactGeneratingNodeNotFound(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetInput("ARTIFACT/root/missing")
	a.Write()

	_, err := chainresolver.ChainResolve("SPEC/root/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestChainResolve_UnreadableFrontmatter(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	testutils.WriteRawNode(t, "SPEC/root/a", "---\ninvalid: yaml: [\n---\n# SPEC/root/a\n")

	_, err := chainresolver.ChainResolve("SPEC/root/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnreadableFrontmatter) {
		t.Errorf("expected ErrUnreadableFrontmatter, got %v", err)
	}
}
