package chainresolver_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/parsing"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/testutils"
)

func TestChainResolve_RootAsTarget(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()

	knownSpecNodes := []string{"SPEC/root"}

	chain, err := chainresolver.ChainResolve("SPEC/root", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Ancestors) != 0 {
		t.Errorf("expected no ancestors, got %d", len(chain.Ancestors))
	}
	if len(chain.Imports) != 0 {
		t.Errorf("expected no imports, got %d", len(chain.Imports))
	}
	if chain.Target.LogicalName != "SPEC/root" {
		t.Errorf("expected target SPEC/root, got %q", chain.Target.LogicalName)
	}
	if chain.Target.Qualifier != nil {
		t.Errorf("expected nil qualifier, got %v", chain.Target.Qualifier)
	}
	if len(chain.Input) != 0 {
		t.Errorf("expected empty input, got %d", len(chain.Input))
	}
}

func TestChainResolve_LinearChain(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	testutils.CreateSpecNode(t, "SPEC/root/a").Write()
	testutils.CreateSpecNode(t, "SPEC/root/a/b").Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/a/b"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a/b", knownSpecNodes)
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

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
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

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
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
	if len(chain.Imports) != 0 {
		t.Errorf("expected no imports, got %d", len(chain.Imports))
	}
	if len(chain.Input) != 0 {
		t.Errorf("expected empty input, got %d", len(chain.Input))
	}
}

func TestChainResolve_DependencyWithoutQualifier(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	b := testutils.CreateSpecNode(t, "SPEC/root/a")
	b.AddImport("SPEC/root/b")
	b.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/b"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(chain.Imports))
	}
	dep := chain.Imports[0]
	if dep.LogicalName != "SPEC/root/b" {
		t.Errorf("expected import SPEC/root/b, got %q", dep.LogicalName)
	}
	if dep.Qualifier != nil {
		t.Errorf("expected nil qualifier, got %v", dep.Qualifier)
	}
}

func TestChainResolve_DependencyWithQualifier(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	b := testutils.CreateSpecNode(t, "SPEC/root/a")
	b.AddImport("SPEC/root/b(interface)")
	b.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/b"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(chain.Imports))
	}
	dep := chain.Imports[0]
	if dep.LogicalName != "SPEC/root/b" {
		t.Errorf("expected import SPEC/root/b, got %q", dep.LogicalName)
	}
	if dep.Qualifier == nil || *dep.Qualifier != "interface" {
		t.Errorf("expected qualifier 'interface', got %v", dep.Qualifier)
	}
}

func TestChainResolve_ImportsSortedByLogicalName(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddImport("SPEC/root/z")
	a.AddImport("SPEC/root/m")
	a.AddImport("SPEC/root/b")
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/z").Write()
	testutils.CreateSpecNode(t, "SPEC/root/m").Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/b", "SPEC/root/m", "SPEC/root/z"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Imports) != 3 {
		t.Fatalf("expected 3 imports, got %d", len(chain.Imports))
	}
	expected := []string{"SPEC/root/b", "SPEC/root/m", "SPEC/root/z"}
	for i, exp := range expected {
		if chain.Imports[i].LogicalName != exp {
			t.Errorf("import[%d]: expected %q, got %q", i, exp, chain.Imports[i].LogicalName)
		}
	}
}

func TestChainResolve_ArtifactDependencyResolved(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddImport("ARTIFACT/root/b")
	a.Write()
	bNode := testutils.CreateSpecNode(t, "SPEC/root/b")
	bNode.SetType("artifact")
	bNode.SetOutput("out/lib.go")
	bNode.Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/b"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(chain.Imports))
	}
	dep := chain.Imports[0]
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
	a.AddImport("ARTIFACT/root/b")
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/b"}

	_, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
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
	a.AddImport("ARTIFACT/root/b")
	a.Write()
	bNode := testutils.CreateSpecNode(t, "SPEC/root/b")
	bNode.SetType("artifact")
	bNode.SetOutput("out/lib.go")
	bNode.Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/b"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(chain.Imports))
	}
	if chain.Imports[0].Path != "out/lib.go" {
		t.Errorf("expected path out/lib.go, got %q", chain.Imports[0].Path)
	}
}

func TestChainResolve_MixedImports(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddImport("SPEC/root/c")
	a.AddImport("ARTIFACT/root/b")
	a.AddImport("EXTERNAL/proto/api.proto")
	a.Write()
	bNode := testutils.CreateSpecNode(t, "SPEC/root/b")
	bNode.SetType("artifact")
	bNode.SetOutput("out/lib.go")
	bNode.Write()
	testutils.CreateSpecNode(t, "SPEC/root/c").Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/b", "SPEC/root/c"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Imports) != 3 {
		t.Fatalf("expected 3 imports, got %d", len(chain.Imports))
	}
	expected := []string{"ARTIFACT/root/b", "EXTERNAL/proto/api.proto", "SPEC/root/c"}
	for i, exp := range expected {
		if chain.Imports[i].LogicalName != exp {
			t.Errorf("import[%d]: expected %q, got %q", i, exp, chain.Imports[i].LogicalName)
		}
	}
}

func TestChainResolve_ExactDuplicate(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddImport("SPEC/root/b")
	a.AddImport("SPEC/root/b")
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/b"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Imports) != 1 {
		t.Errorf("expected 1 import (deduped), got %d", len(chain.Imports))
	}
}

func TestChainResolve_NoQualifierSubsumesQualifier(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddImport("SPEC/root/b")
	a.AddImport("SPEC/root/b(interface)")
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/b"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(chain.Imports))
	}
	if chain.Imports[0].Qualifier != nil {
		t.Errorf("expected nil qualifier, got %v", chain.Imports[0].Qualifier)
	}
}

func TestChainResolve_QualifierBeforeNoQualifier(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddImport("SPEC/root/b(interface)")
	a.AddImport("SPEC/root/b")
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/b"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(chain.Imports))
	}
	if chain.Imports[0].Qualifier != nil {
		t.Errorf("expected nil qualifier (no-qualifier wins), got %v", chain.Imports[0].Qualifier)
	}
}

func TestChainResolve_SameFileDifferentQualifiers(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddImport("SPEC/root/b(interface)")
	a.AddImport("SPEC/root/b(constraints)")
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/b"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Imports) != 2 {
		t.Fatalf("expected 2 imports, got %d", len(chain.Imports))
	}
	if chain.Imports[0].Qualifier == nil || *chain.Imports[0].Qualifier != "constraints" {
		t.Errorf("expected qualifier 'constraints' at index 0, got %v", chain.Imports[0].Qualifier)
	}
	if chain.Imports[1].Qualifier == nil || *chain.Imports[1].Qualifier != "interface" {
		t.Errorf("expected qualifier 'interface' at index 1, got %v", chain.Imports[1].Qualifier)
	}
}

func TestChainResolve_DuplicateArtifact(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddImport("ARTIFACT/root/b")
	a.AddImport("ARTIFACT/root/b")
	a.Write()
	bNode := testutils.CreateSpecNode(t, "SPEC/root/b")
	bNode.SetType("artifact")
	bNode.SetOutput("out/lib.go")
	bNode.Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/b"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Imports) != 1 {
		t.Errorf("expected 1 ARTIFACT import (deduped), got %d", len(chain.Imports))
	}
}

func TestChainResolve_ExternalDependencyResolvedToPath(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddImport("EXTERNAL/docs/api.yaml")
	a.Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(chain.Imports))
	}
	dep := chain.Imports[0]
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

func TestChainResolve_MultipleExternalImportsSorted(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddImport("EXTERNAL/proto/v1.proto")
	a.AddImport("EXTERNAL/docs/api.yaml")
	a.Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Imports) != 2 {
		t.Fatalf("expected 2 imports, got %d", len(chain.Imports))
	}
	if chain.Imports[0].LogicalName != "EXTERNAL/docs/api.yaml" {
		t.Errorf("expected EXTERNAL/docs/api.yaml at index 0, got %q", chain.Imports[0].LogicalName)
	}
	if chain.Imports[1].LogicalName != "EXTERNAL/proto/v1.proto" {
		t.Errorf("expected EXTERNAL/proto/v1.proto at index 1, got %q", chain.Imports[1].LogicalName)
	}
}

func TestChainResolve_DuplicateExternal(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddImport("EXTERNAL/x.proto")
	a.AddImport("EXTERNAL/x.proto")
	a.Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Imports) != 1 {
		t.Errorf("expected 1 EXTERNAL import (deduped), got %d", len(chain.Imports))
	}
}

func TestChainResolve_InputArtifactResolved(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetInputScalar("ARTIFACT/root/b")
	a.Write()
	bNode := testutils.CreateSpecNode(t, "SPEC/root/b")
	bNode.SetType("artifact")
	bNode.SetOutput("out/data.json")
	bNode.Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/b"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Input) != 1 {
		t.Fatalf("expected 1 input, got %d", len(chain.Input))
	}
	if chain.Input[0].LogicalName != "ARTIFACT/root/b" {
		t.Errorf("expected ARTIFACT/root/b, got %q", chain.Input[0].LogicalName)
	}
	if chain.Input[0].Path != "out/data.json" {
		t.Errorf("expected path out/data.json, got %q", chain.Input[0].Path)
	}
}

func TestChainResolve_ExternalInputResolvedToPath(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetInputScalar("EXTERNAL/docs/vendor/spec.yaml")
	a.Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Input) != 1 {
		t.Fatalf("expected 1 input, got %d", len(chain.Input))
	}
	if chain.Input[0].LogicalName != "EXTERNAL/docs/vendor/spec.yaml" {
		t.Errorf("expected EXTERNAL/docs/vendor/spec.yaml, got %q", chain.Input[0].LogicalName)
	}
	if chain.Input[0].Path != "docs/vendor/spec.yaml" {
		t.Errorf("expected path docs/vendor/spec.yaml, got %q", chain.Input[0].Path)
	}
}

func TestChainResolve_SpecInputResolved(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetInputScalar("SPEC/root/b")
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/b"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Input) != 1 {
		t.Fatalf("expected 1 input, got %d", len(chain.Input))
	}
	if chain.Input[0].LogicalName != "SPEC/root/b" {
		t.Errorf("expected SPEC/root/b, got %q", chain.Input[0].LogicalName)
	}
	if chain.Input[0].Path != "code-from-spec/root/b/_node.md" {
		t.Errorf("expected path code-from-spec/root/b/_node.md, got %q", chain.Input[0].Path)
	}
	if chain.Input[0].Qualifier != nil {
		t.Errorf("expected nil qualifier, got %v", chain.Input[0].Qualifier)
	}
}

func TestChainResolve_SpecInputWithQualifier(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetInputScalar("SPEC/root/b(acceptance-tests)")
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/b"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Input) != 1 {
		t.Fatalf("expected 1 input, got %d", len(chain.Input))
	}
	if chain.Input[0].LogicalName != "SPEC/root/b" {
		t.Errorf("expected SPEC/root/b, got %q", chain.Input[0].LogicalName)
	}
	if chain.Input[0].Path != "code-from-spec/root/b/_node.md" {
		t.Errorf("expected path code-from-spec/root/b/_node.md, got %q", chain.Input[0].Path)
	}
	if chain.Input[0].Qualifier == nil || *chain.Input[0].Qualifier != "acceptance-tests" {
		t.Errorf("expected qualifier 'acceptance-tests', got %v", chain.Input[0].Qualifier)
	}
}

func TestChainResolve_MultipleInputsSorted(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetInputList([]string{"SPEC/root/z", "SPEC/root/b"})
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/z").Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/b", "SPEC/root/z"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Input) != 2 {
		t.Fatalf("expected 2 inputs, got %d", len(chain.Input))
	}
	if chain.Input[0].LogicalName != "SPEC/root/b" {
		t.Errorf("expected input[0] = SPEC/root/b, got %q", chain.Input[0].LogicalName)
	}
	if chain.Input[1].LogicalName != "SPEC/root/z" {
		t.Errorf("expected input[1] = SPEC/root/z, got %q", chain.Input[1].LogicalName)
	}
}

func TestChainResolve_DuplicateInput(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetInputList([]string{"SPEC/root/b", "SPEC/root/b"})
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/b"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Input) != 1 {
		t.Errorf("expected 1 input (deduped), got %d", len(chain.Input))
	}
}

func TestChainResolve_NoInput(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	testutils.CreateSpecNode(t, "SPEC/root/a").Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Input) != 0 {
		t.Errorf("expected empty input, got %d", len(chain.Input))
	}
}

func TestChainResolve_UnrecognizedPrefixInImports(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddImport("UNKNOWN/something")
	a.Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a"}

	_, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestChainResolve_InvalidTargetLogicalName(t *testing.T) {
	testutils.Chdir(t)

	_, err := chainresolver.ChainResolve("INVALID/something", nil)
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
	a.SetInputScalar("ARTIFACT/root/missing")
	a.Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a"}

	_, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestChainResolve_UnreadableFrontmatter(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	testutils.WriteRawNode(t, "SPEC/root/a", "---\ninvalid: yaml: [\n---\n# SPEC/root/a\n")

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a"}

	_, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnreadableFrontmatter) {
		t.Errorf("expected ErrUnreadableFrontmatter, got %v", err)
	}
}

func TestChainResolve_SpecGlobExpandsToDescendants(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddImport("SPEC/root/b/*")
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()
	testutils.CreateSpecNode(t, "SPEC/root/b/x").Write()
	testutils.CreateSpecNode(t, "SPEC/root/b/y").Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/b", "SPEC/root/b/x", "SPEC/root/b/y"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Imports) != 2 {
		t.Fatalf("expected 2 imports, got %d", len(chain.Imports))
	}
	if chain.Imports[0].LogicalName != "SPEC/root/b/x" {
		t.Errorf("expected import[0] = SPEC/root/b/x, got %q", chain.Imports[0].LogicalName)
	}
	if chain.Imports[1].LogicalName != "SPEC/root/b/y" {
		t.Errorf("expected import[1] = SPEC/root/b/y, got %q", chain.Imports[1].LogicalName)
	}
}

func TestChainResolve_SpecGlobExpandsToDeepDescendants(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddImport("SPEC/root/b/*")
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()
	testutils.CreateSpecNode(t, "SPEC/root/b/x").Write()
	testutils.CreateSpecNode(t, "SPEC/root/b/x/deep").Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/b", "SPEC/root/b/x", "SPEC/root/b/x/deep"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Imports) != 2 {
		t.Fatalf("expected 2 imports, got %d", len(chain.Imports))
	}
	if chain.Imports[0].LogicalName != "SPEC/root/b/x" {
		t.Errorf("expected import[0] = SPEC/root/b/x, got %q", chain.Imports[0].LogicalName)
	}
	if chain.Imports[1].LogicalName != "SPEC/root/b/x/deep" {
		t.Errorf("expected import[1] = SPEC/root/b/x/deep, got %q", chain.Imports[1].LogicalName)
	}
}

func TestChainResolve_ArtifactGlobExpandsWithPrefixConversion(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddImport("ARTIFACT/root/b/*")
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()
	xNode := testutils.CreateSpecNode(t, "SPEC/root/b/x")
	xNode.SetType("artifact")
	xNode.SetOutput("out/x.go")
	xNode.Write()
	yNode := testutils.CreateSpecNode(t, "SPEC/root/b/y")
	yNode.SetType("artifact")
	yNode.SetOutput("out/y.go")
	yNode.Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/b", "SPEC/root/b/x", "SPEC/root/b/y"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Imports) != 2 {
		t.Fatalf("expected 2 imports, got %d", len(chain.Imports))
	}
	if chain.Imports[0].LogicalName != "ARTIFACT/root/b/x" {
		t.Errorf("expected import[0] = ARTIFACT/root/b/x, got %q", chain.Imports[0].LogicalName)
	}
	if chain.Imports[0].Path != "out/x.go" {
		t.Errorf("expected import[0].Path = out/x.go, got %q", chain.Imports[0].Path)
	}
	if chain.Imports[1].LogicalName != "ARTIFACT/root/b/y" {
		t.Errorf("expected import[1] = ARTIFACT/root/b/y, got %q", chain.Imports[1].LogicalName)
	}
	if chain.Imports[1].Path != "out/y.go" {
		t.Errorf("expected import[1].Path = out/y.go, got %q", chain.Imports[1].Path)
	}
}

func TestChainResolve_GlobExcludesDeclaringNodeAndAncestors(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	testutils.CreateSpecNode(t, "SPEC/root/a").Write()
	ab := testutils.CreateSpecNode(t, "SPEC/root/a/b")
	ab.AddImport("SPEC/root/*")
	ab.Write()
	testutils.CreateSpecNode(t, "SPEC/root/c").Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/a/b", "SPEC/root/c"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a/b", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(chain.Imports))
	}
	if chain.Imports[0].LogicalName != "SPEC/root/c" {
		t.Errorf("expected import = SPEC/root/c, got %q", chain.Imports[0].LogicalName)
	}
}

func TestChainResolve_GlobWithEmptyMatch(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddImport("SPEC/root/empty/*")
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/empty").Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/empty"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Imports) != 0 {
		t.Errorf("expected empty imports, got %d", len(chain.Imports))
	}
}

func TestChainResolve_GlobDeduplicatesWithExplicitEntries(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.AddImport("SPEC/root/b/x")
	a.AddImport("SPEC/root/b/*")
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()
	testutils.CreateSpecNode(t, "SPEC/root/b/x").Write()
	testutils.CreateSpecNode(t, "SPEC/root/b/y").Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/b", "SPEC/root/b/x", "SPEC/root/b/y"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Imports) != 2 {
		t.Fatalf("expected 2 imports (deduped), got %d", len(chain.Imports))
	}
	if chain.Imports[0].LogicalName != "SPEC/root/b/x" {
		t.Errorf("expected import[0] = SPEC/root/b/x, got %q", chain.Imports[0].LogicalName)
	}
	if chain.Imports[1].LogicalName != "SPEC/root/b/y" {
		t.Errorf("expected import[1] = SPEC/root/b/y, got %q", chain.Imports[1].LogicalName)
	}
}

func TestChainResolve_SpecGlobInInput(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetInputList([]string{"SPEC/root/b/*"})
	a.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()
	testutils.CreateSpecNode(t, "SPEC/root/b/x").Write()
	testutils.CreateSpecNode(t, "SPEC/root/b/y").Write()

	knownSpecNodes := []string{"SPEC/root", "SPEC/root/a", "SPEC/root/b", "SPEC/root/b/x", "SPEC/root/b/y"}

	chain, err := chainresolver.ChainResolve("SPEC/root/a", knownSpecNodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Input) != 2 {
		t.Fatalf("expected 2 inputs, got %d", len(chain.Input))
	}
	if chain.Input[0].LogicalName != "SPEC/root/b/x" {
		t.Errorf("expected input[0] = SPEC/root/b/x, got %q", chain.Input[0].LogicalName)
	}
	if chain.Input[1].LogicalName != "SPEC/root/b/y" {
		t.Errorf("expected input[1] = SPEC/root/b/y, got %q", chain.Input[1].LogicalName)
	}
}
