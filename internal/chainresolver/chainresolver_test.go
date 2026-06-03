// code-from-spec: ROOT/golang/tests/chain/resolver@VE95K3QxfTdh9L2EtgAHRK5sMbs
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

func testWriteNode(t *testing.T, logicalName string, frontmatter string) {
	t.Helper()
	parts := []string{"code-from-spec"}
	segments := filepath.SplitList(logicalName)
	_ = segments

	path := logicalNameToRelPath(logicalName)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testWriteNode MkdirAll: %v", err)
	}

	var content string
	if frontmatter != "" {
		content = "---\n" + frontmatter + "\n---\n\n# " + logicalName + "\n"
	} else {
		content = "# " + logicalName + "\n"
	}
	_ = parts

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNode WriteFile: %v", err)
	}
}

func logicalNameToRelPath(logicalName string) string {
	if logicalName == "ROOT" {
		return "code-from-spec/_node.md"
	}
	rest := logicalName[len("ROOT/"):]
	return "code-from-spec/" + rest + "/_node.md"
}

func TestChainResolveRootAsTarget(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")

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

func TestChainResolveLinearChainAncestorsOrder(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "")
	testWriteNode(t, "ROOT/a/b", "")

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
	if chain.Target == nil || chain.Target.LogicalName != "ROOT/a/b" {
		t.Errorf("expected target ROOT/a/b")
	}
}

func TestChainResolveSingleParent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "")

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

func TestChainResolveTargetWithEmptyFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "")

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
		t.Errorf("expected no dependencies, got %d", len(chain.Dependencies))
	}
	if len(chain.External) != 0 {
		t.Errorf("expected no external, got %d", len(chain.External))
	}
	if chain.Input != nil {
		t.Errorf("expected no input")
	}
}

func TestChainResolveDependencyWithoutQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ROOT/b")
	testWriteNode(t, "ROOT/b", "")

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

func TestChainResolveDependencyWithQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ROOT/b(interface)")
	testWriteNode(t, "ROOT/b", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].LogicalName != "ROOT/b" {
		t.Errorf("expected dependency logical name ROOT/b, got %q", chain.Dependencies[0].LogicalName)
	}
	if chain.Dependencies[0].Qualifier != "interface" {
		t.Errorf("expected qualifier 'interface', got %q", chain.Dependencies[0].Qualifier)
	}
}

func TestChainResolveDependenciesSortedByFilePath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ROOT/z\n  - ROOT/m\n  - ROOT/b")
	testWriteNode(t, "ROOT/z", "")
	testWriteNode(t, "ROOT/m", "")
	testWriteNode(t, "ROOT/b", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 3 {
		t.Fatalf("expected 3 dependencies, got %d", len(chain.Dependencies))
	}

	for i := 1; i < len(chain.Dependencies); i++ {
		if chain.Dependencies[i].FilePath.Value < chain.Dependencies[i-1].FilePath.Value {
			t.Errorf("dependencies not sorted by file path: %q > %q",
				chain.Dependencies[i-1].FilePath.Value,
				chain.Dependencies[i].FilePath.Value)
		}
	}

	if chain.Dependencies[0].LogicalName != "ROOT/b" {
		t.Errorf("expected first dependency ROOT/b, got %q", chain.Dependencies[0].LogicalName)
	}
	if chain.Dependencies[1].LogicalName != "ROOT/m" {
		t.Errorf("expected second dependency ROOT/m, got %q", chain.Dependencies[1].LogicalName)
	}
	if chain.Dependencies[2].LogicalName != "ROOT/z" {
		t.Errorf("expected third dependency ROOT/z, got %q", chain.Dependencies[2].LogicalName)
	}
}

func TestChainResolveArtifactDependencyResolved(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ARTIFACT/b")
	testWriteNode(t, "ROOT/b", "output: out/lib.go")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].LogicalName != "ARTIFACT/b" {
		t.Errorf("expected logical name ARTIFACT/b, got %q", chain.Dependencies[0].LogicalName)
	}
	if chain.Dependencies[0].FilePath.Value != "out/lib.go" {
		t.Errorf("expected file_path out/lib.go, got %q", chain.Dependencies[0].FilePath.Value)
	}
}

func TestChainResolveArtifactNoOutput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ARTIFACT/b")
	testWriteNode(t, "ROOT/b", "")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestChainResolveArtifactFileNotOnDisk(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ARTIFACT/b")
	testWriteNode(t, "ROOT/b", "output: out/lib.go")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].FilePath.Value != "out/lib.go" {
		t.Errorf("expected file_path out/lib.go, got %q", chain.Dependencies[0].FilePath.Value)
	}
}

func TestChainResolveMixedDependencies(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ROOT/c\n  - ARTIFACT/b")
	testWriteNode(t, "ROOT/b", "output: out/lib.go")
	testWriteNode(t, "ROOT/c", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}

	for i := 1; i < len(chain.Dependencies); i++ {
		if chain.Dependencies[i].FilePath.Value < chain.Dependencies[i-1].FilePath.Value {
			t.Errorf("dependencies not sorted by file path: %q > %q",
				chain.Dependencies[i-1].FilePath.Value,
				chain.Dependencies[i].FilePath.Value)
		}
	}
}

func TestChainResolveDedupExactDuplicate(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ROOT/b\n  - ROOT/b")
	testWriteNode(t, "ROOT/b", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency after dedup, got %d", len(chain.Dependencies))
	}
}

func TestChainResolveDedupNoQualifierSubsumesQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ROOT/b\n  - ROOT/b(interface)")
	testWriteNode(t, "ROOT/b", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency after dedup, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].Qualifier != "" {
		t.Errorf("expected no qualifier (no-qualifier wins), got %q", chain.Dependencies[0].Qualifier)
	}
}

func TestChainResolveDedupQualifierBeforeNoQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ROOT/b(interface)\n  - ROOT/b")
	testWriteNode(t, "ROOT/b", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency after dedup, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].Qualifier != "" {
		t.Errorf("expected no qualifier (no-qualifier wins), got %q", chain.Dependencies[0].Qualifier)
	}
}

func TestChainResolveDedupSameFileDifferentQualifiers(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ROOT/b(interface)\n  - ROOT/b(constraints)")
	testWriteNode(t, "ROOT/b", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies for different qualifiers, got %d", len(chain.Dependencies))
	}

	qualifiers := map[string]bool{}
	for _, dep := range chain.Dependencies {
		qualifiers[dep.Qualifier] = true
	}
	if !qualifiers["interface"] {
		t.Error("expected qualifier 'interface'")
	}
	if !qualifiers["constraints"] {
		t.Error("expected qualifier 'constraints'")
	}
}

func TestChainResolveDedupArtifactDuplicate(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ARTIFACT/b\n  - ARTIFACT/b")
	testWriteNode(t, "ROOT/b", "output: out/lib.go")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency after dedup, got %d", len(chain.Dependencies))
	}
}

func TestChainResolveExternalEntriesCopied(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "external:\n  - path: docs/api.yaml\n  - path: proto/v1.proto")

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

func TestChainResolveExternalEmpty(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.External) != 0 {
		t.Errorf("expected no external entries, got %d", len(chain.External))
	}
}

func TestChainResolveInputResolved(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "input: ARTIFACT/b")
	testWriteNode(t, "ROOT/b", "output: out/data.json")

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
		t.Errorf("expected input file_path out/data.json, got %q", chain.Input.FilePath.Value)
	}
}

func TestChainResolveNoInput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chain.Input != nil {
		t.Errorf("expected no input, got %v", chain.Input)
	}
}

func TestChainResolveUnrecognizedPrefixInDependsOn(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - UNKNOWN/something")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestChainResolveInvalidTargetLogicalName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := chainresolver.ChainResolve("INVALID/something")
	if err == nil {
		t.Fatal("expected error for invalid logical name, got nil")
	}
}

func TestChainResolveUnreadableFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "")

	nodeDir := "code-from-spec/a"
	if err := os.MkdirAll(nodeDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	invalidYAML := "---\n: invalid: yaml: content: [\n---\n\n# ROOT/a\n"
	if err := os.WriteFile(nodeDir+"/_node.md", []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnreadableFrontmatter) {
		t.Errorf("expected ErrUnreadableFrontmatter, got %v", err)
	}
}
