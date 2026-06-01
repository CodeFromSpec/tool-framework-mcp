// code-from-spec: ROOT/golang/tests/chain/resolver@v1njTFnekRwoocHfIKQtDnP-qjI
package chainresolver_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
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
	path, err := logicalnames.LogicalNameToPath(logicalName)
	if err != nil {
		t.Fatalf("testWriteNode: LogicalNameToPath(%q): %v", logicalName, err)
	}
	dir := filepath.Dir(path.Value)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testWriteNode: MkdirAll(%q): %v", dir, err)
	}
	content := "---\n" + frontmatter + "---\n"
	if err := os.WriteFile(path.Value, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNode: WriteFile(%q): %v", path.Value, err)
	}
}

func testWriteEmptyNode(t *testing.T, logicalName string) {
	t.Helper()
	testWriteNode(t, logicalName, "")
}

func TestChainResolve_RootAsTarget(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")

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
	if chain.Target.Qualifier != nil {
		t.Errorf("expected no qualifier, got %q", *chain.Target.Qualifier)
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

func TestChainResolve_LinearChain_AncestorsRootFirst(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteEmptyNode(t, "ROOT/a")
	testWriteEmptyNode(t, "ROOT/a/b")

	chain, err := chainresolver.ChainResolve("ROOT/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Ancestors) != 2 {
		t.Fatalf("expected 2 ancestors, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].LogicalName != "ROOT" {
		t.Errorf("expected ancestors[0]=ROOT, got %q", chain.Ancestors[0].LogicalName)
	}
	if chain.Ancestors[1].LogicalName != "ROOT/a" {
		t.Errorf("expected ancestors[1]=ROOT/a, got %q", chain.Ancestors[1].LogicalName)
	}
	if chain.Target == nil || chain.Target.LogicalName != "ROOT/a/b" {
		t.Errorf("expected target ROOT/a/b, got %v", chain.Target)
	}
}

func TestChainResolve_SingleParent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteEmptyNode(t, "ROOT/a")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Ancestors) != 1 {
		t.Fatalf("expected 1 ancestor, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].LogicalName != "ROOT" {
		t.Errorf("expected ancestors[0]=ROOT, got %q", chain.Ancestors[0].LogicalName)
	}
	if chain.Target == nil || chain.Target.LogicalName != "ROOT/a" {
		t.Errorf("expected target ROOT/a, got %v", chain.Target)
	}
}

func TestChainResolve_TargetWithEmptyFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteEmptyNode(t, "ROOT/a")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Ancestors) != 1 {
		t.Fatalf("expected 1 ancestor, got %d", len(chain.Ancestors))
	}
	if chain.Target == nil || chain.Target.LogicalName != "ROOT/a" {
		t.Errorf("expected target ROOT/a, got %v", chain.Target)
	}
	if chain.Target.Qualifier != nil {
		t.Errorf("expected no qualifier, got %q", *chain.Target.Qualifier)
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
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ROOT/b\n")
	testWriteEmptyNode(t, "ROOT/b")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ROOT/b" {
		t.Errorf("expected dependency ROOT/b, got %q", dep.LogicalName)
	}
	if dep.Qualifier != nil {
		t.Errorf("expected no qualifier, got %q", *dep.Qualifier)
	}
}

func TestChainResolve_DependencyWithQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ROOT/b(interface)\n")
	testWriteEmptyNode(t, "ROOT/b")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ROOT/b" {
		t.Errorf("expected dependency ROOT/b, got %q", dep.LogicalName)
	}
	if dep.Qualifier == nil || *dep.Qualifier != "interface" {
		t.Errorf("expected qualifier=interface, got %v", dep.Qualifier)
	}
}

func TestChainResolve_DependenciesSortedByFilePath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ROOT/z\n  - ROOT/m\n  - ROOT/b\n")
	testWriteEmptyNode(t, "ROOT/z")
	testWriteEmptyNode(t, "ROOT/m")
	testWriteEmptyNode(t, "ROOT/b")

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
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("dependencies[%d]: expected %q, got %q", i, expected[i], name)
		}
	}
}

func TestChainResolve_ArtifactDependencyResolved(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ARTIFACT/b(lib)\n")
	testWriteNode(t, "ROOT/b", "outputs:\n  - id: lib\n    path: out/lib.go\n")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ARTIFACT/b(lib)" {
		t.Errorf("expected logical name ARTIFACT/b(lib), got %q", dep.LogicalName)
	}
	if dep.FilePath.Value != "out/lib.go" {
		t.Errorf("expected file_path out/lib.go, got %q", dep.FilePath.Value)
	}
	if dep.Qualifier == nil || *dep.Qualifier != "lib" {
		t.Errorf("expected qualifier=lib, got %v", dep.Qualifier)
	}
}

func TestChainResolve_ArtifactWithoutQualifier_Error(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ARTIFACT/b\n")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestChainResolve_ArtifactGeneratingNodeNoOutputs_Error(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ARTIFACT/b(lib)\n")
	testWriteEmptyNode(t, "ROOT/b")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestChainResolve_ArtifactFileNotOnDisk_NoError(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ARTIFACT/b(lib)\n")
	testWriteNode(t, "ROOT/b", "outputs:\n  - id: lib\n    path: out/lib.go\n")

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

func TestChainResolve_ArtifactNonExistentOutputId_Error(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ARTIFACT/b(missing)\n")
	testWriteNode(t, "ROOT/b", "outputs:\n  - id: lib\n    path: out/lib.go\n")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestChainResolve_MixedRootAndArtifactDependencies(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ROOT/c\n  - ARTIFACT/b(lib)\n")
	testWriteNode(t, "ROOT/b", "outputs:\n  - id: lib\n    path: out/lib.go\n")
	testWriteEmptyNode(t, "ROOT/c")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}
}

func TestChainResolve_DedupExactDuplicate(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ROOT/b\n  - ROOT/b\n")
	testWriteEmptyNode(t, "ROOT/b")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency after dedup, got %d", len(chain.Dependencies))
	}
}

func TestChainResolve_DedupNoQualifierSubsumesQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ROOT/b\n  - ROOT/b(interface)\n")
	testWriteEmptyNode(t, "ROOT/b")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency after dedup, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].Qualifier != nil {
		t.Errorf("expected no qualifier (no-qualifier wins), got %q", *chain.Dependencies[0].Qualifier)
	}
}

func TestChainResolve_DedupQualifierBeforeNoQualifier_NoQualifierWins(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ROOT/b(interface)\n  - ROOT/b\n")
	testWriteEmptyNode(t, "ROOT/b")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency after dedup, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].Qualifier != nil {
		t.Errorf("expected no qualifier (no-qualifier wins), got %q", *chain.Dependencies[0].Qualifier)
	}
}

func TestChainResolve_DedupDifferentQualifiersBothKept(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ROOT/b(interface)\n  - ROOT/b(constraints)\n")
	testWriteEmptyNode(t, "ROOT/b")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}
	q0 := chain.Dependencies[0].Qualifier
	q1 := chain.Dependencies[1].Qualifier
	if q0 == nil || q1 == nil {
		t.Fatalf("expected both dependencies to have qualifiers")
	}
	if *q0 != "constraints" || *q1 != "interface" {
		t.Errorf("expected qualifiers sorted: constraints, interface; got %q, %q", *q0, *q1)
	}
}

func TestChainResolve_DedupDuplicateArtifact(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ARTIFACT/b(lib)\n  - ARTIFACT/b(lib)\n")
	testWriteNode(t, "ROOT/b", "outputs:\n  - id: lib\n    path: out/lib.go\n")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency after dedup, got %d", len(chain.Dependencies))
	}
}

func TestChainResolve_ExternalEntriesSortedAlphabetically(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteNode(t, "ROOT/a", "external:\n  - path: proto/v1.proto\n  - path: docs/api.yaml\n")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.External) != 2 {
		t.Fatalf("expected 2 external entries, got %d", len(chain.External))
	}
	if chain.External[0].Path != "docs/api.yaml" {
		t.Errorf("expected external[0]=docs/api.yaml, got %q", chain.External[0].Path)
	}
	if chain.External[1].Path != "proto/v1.proto" {
		t.Errorf("expected external[1]=proto/v1.proto, got %q", chain.External[1].Path)
	}
}

func TestChainResolve_EmptyExternal(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteEmptyNode(t, "ROOT/a")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain.External) != 0 {
		t.Errorf("expected no external, got %d", len(chain.External))
	}
}

func TestChainResolve_InputResolvedFromGeneratingNode(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteNode(t, "ROOT/a", "input: ARTIFACT/b(data)\n")
	testWriteNode(t, "ROOT/b", "outputs:\n  - id: data\n    path: out/data.json\n")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain.Input == nil {
		t.Fatal("expected input, got nil")
	}
	if chain.Input.LogicalName != "ARTIFACT/b(data)" {
		t.Errorf("expected input logical name ARTIFACT/b(data), got %q", chain.Input.LogicalName)
	}
	if chain.Input.FilePath.Value != "out/data.json" {
		t.Errorf("expected input file_path out/data.json, got %q", chain.Input.FilePath.Value)
	}
	if chain.Input.Qualifier == nil || *chain.Input.Qualifier != "data" {
		t.Errorf("expected input qualifier=data, got %v", chain.Input.Qualifier)
	}
}

func TestChainResolve_NoInput_Absent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteEmptyNode(t, "ROOT/a")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain.Input != nil {
		t.Errorf("expected no input, got %v", chain.Input)
	}
}

func TestChainResolve_InputWithoutQualifier_Error(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteNode(t, "ROOT/a", "input: ARTIFACT/b\n")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestChainResolve_InputNonExistentOutputId_Error(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteNode(t, "ROOT/a", "input: ARTIFACT/b(missing)\n")
	testWriteNode(t, "ROOT/b", "outputs:\n  - id: data\n    path: out/data.json\n")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestChainResolve_UnrecognizedPrefixInDependsOn_Error(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")
	testWriteNode(t, "ROOT/a", "depends_on:\n  - UNKNOWN/something\n")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestChainResolve_InvalidTargetLogicalName_Error(t *testing.T) {
	_, err := chainresolver.ChainResolve("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestChainResolve_UnreadableFrontmatter_Error(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteEmptyNode(t, "ROOT")

	path, err := logicalnames.LogicalNameToPath("ROOT/a")
	if err != nil {
		t.Fatalf("LogicalNameToPath: %v", err)
	}
	nodeDir := filepath.Dir(path.Value)
	if err := os.MkdirAll(nodeDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	content := "---\n: invalid: yaml: [\n---\n"
	if err := os.WriteFile(path.Value, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err = chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnreadableFrontmatter) {
		t.Errorf("expected ErrUnreadableFrontmatter, got %v", err)
	}
}
