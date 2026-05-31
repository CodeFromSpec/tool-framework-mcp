// code-from-spec: ROOT/golang/tests/chain/resolver@78r_iByjS1IKhy85TzLrkam5McA
package chainresolver_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
)

// testChdir changes the working directory to dir for the duration of the test.
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

// testWriteNode creates a _node.md file at the given relative path inside
// code-from-spec/, with the provided frontmatter content.
// If frontmatter is empty, the file will have no frontmatter block.
func testWriteNode(t *testing.T, relPath string, frontmatter string) {
	t.Helper()
	fullPath := "code-from-spec/" + relPath + "/_node.md"
	if err := os.MkdirAll("code-from-spec/"+relPath, 0755); err != nil {
		t.Fatalf("testWriteNode MkdirAll: %v", err)
	}
	var content string
	if frontmatter != "" {
		content = "---\n" + frontmatter + "\n---\n"
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNode WriteFile: %v", err)
	}
}

// testWriteRootNode creates a _node.md file for the ROOT node.
func testWriteRootNode(t *testing.T) {
	t.Helper()
	if err := os.MkdirAll("code-from-spec", 0755); err != nil {
		t.Fatalf("testWriteRootNode MkdirAll: %v", err)
	}
	if err := os.WriteFile("code-from-spec/_node.md", []byte(""), 0644); err != nil {
		t.Fatalf("testWriteRootNode WriteFile: %v", err)
	}
}

// TC-01: Root as target
func TestChainResolve_RootAsTarget(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)

	chain, err := chainresolver.ChainResolve("ROOT")
	if err != nil {
		t.Fatalf("ChainResolve: %v", err)
	}

	if len(chain.Ancestors) != 0 {
		t.Errorf("expected 0 ancestors, got %d", len(chain.Ancestors))
	}
	if chain.Target == nil {
		t.Fatal("expected target, got nil")
	}
	if chain.Target.LogicalName != "ROOT" {
		t.Errorf("expected target logical name ROOT, got %q", chain.Target.LogicalName)
	}
	if chain.Target.Qualifier != nil {
		t.Errorf("expected target qualifier absent, got %q", *chain.Target.Qualifier)
	}
	if len(chain.Dependencies) != 0 {
		t.Errorf("expected 0 dependencies, got %d", len(chain.Dependencies))
	}
	if len(chain.External) != 0 {
		t.Errorf("expected 0 external, got %d", len(chain.External))
	}
	if chain.Input != nil {
		t.Errorf("expected input absent, got %v", chain.Input)
	}
}

// TC-02: Linear chain — ancestors in root-first order
func TestChainResolve_LinearChain_AncestorsInOrder(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "")
	testWriteNode(t, "a/b", "")

	chain, err := chainresolver.ChainResolve("ROOT/a/b")
	if err != nil {
		t.Fatalf("ChainResolve: %v", err)
	}

	if len(chain.Ancestors) != 2 {
		t.Fatalf("expected 2 ancestors, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].LogicalName != "ROOT" {
		t.Errorf("expected ancestors[0] = ROOT, got %q", chain.Ancestors[0].LogicalName)
	}
	if chain.Ancestors[1].LogicalName != "ROOT/a" {
		t.Errorf("expected ancestors[1] = ROOT/a, got %q", chain.Ancestors[1].LogicalName)
	}
	if chain.Target == nil || chain.Target.LogicalName != "ROOT/a/b" {
		t.Errorf("expected target ROOT/a/b, got %v", chain.Target)
	}
}

// TC-03: Single parent
func TestChainResolve_SingleParent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("ChainResolve: %v", err)
	}

	if len(chain.Ancestors) != 1 {
		t.Fatalf("expected 1 ancestor, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].LogicalName != "ROOT" {
		t.Errorf("expected ancestors[0] = ROOT, got %q", chain.Ancestors[0].LogicalName)
	}
	if chain.Target == nil || chain.Target.LogicalName != "ROOT/a" {
		t.Errorf("expected target ROOT/a, got %v", chain.Target)
	}
}

// TC-04: Target with empty frontmatter
func TestChainResolve_TargetWithEmptyFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("ChainResolve: %v", err)
	}

	if len(chain.Ancestors) != 1 {
		t.Errorf("expected 1 ancestor, got %d", len(chain.Ancestors))
	}
	if chain.Target == nil || chain.Target.LogicalName != "ROOT/a" {
		t.Errorf("expected target ROOT/a, got %v", chain.Target)
	}
	if len(chain.Dependencies) != 0 {
		t.Errorf("expected 0 dependencies, got %d", len(chain.Dependencies))
	}
	if len(chain.External) != 0 {
		t.Errorf("expected 0 external, got %d", len(chain.External))
	}
	if chain.Input != nil {
		t.Errorf("expected input absent, got %v", chain.Input)
	}
}

// TC-05: Dependency without qualifier
func TestChainResolve_DependencyWithoutQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "depends_on:\n  - ROOT/b")
	testWriteNode(t, "b", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("ChainResolve: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ROOT/b" {
		t.Errorf("expected dependency logical name ROOT/b, got %q", dep.LogicalName)
	}
	if dep.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %q", *dep.Qualifier)
	}
}

// TC-06: Dependency with qualifier
func TestChainResolve_DependencyWithQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "depends_on:\n  - ROOT/b(interface)")
	testWriteNode(t, "b", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("ChainResolve: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ROOT/b" {
		t.Errorf("expected dependency logical name ROOT/b, got %q", dep.LogicalName)
	}
	if dep.Qualifier == nil {
		t.Fatal("expected qualifier present, got nil")
	}
	if *dep.Qualifier != "interface" {
		t.Errorf("expected qualifier interface, got %q", *dep.Qualifier)
	}
}

// TC-07: Dependencies sorted by file path then qualifier
func TestChainResolve_DependenciesSortedByFilePath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "depends_on:\n  - ROOT/z\n  - ROOT/m\n  - ROOT/b")
	testWriteNode(t, "z", "")
	testWriteNode(t, "m", "")
	testWriteNode(t, "b", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("ChainResolve: %v", err)
	}

	if len(chain.Dependencies) != 3 {
		t.Fatalf("expected 3 dependencies, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].LogicalName != "ROOT/b" {
		t.Errorf("expected dependencies[0] = ROOT/b, got %q", chain.Dependencies[0].LogicalName)
	}
	if chain.Dependencies[1].LogicalName != "ROOT/m" {
		t.Errorf("expected dependencies[1] = ROOT/m, got %q", chain.Dependencies[1].LogicalName)
	}
	if chain.Dependencies[2].LogicalName != "ROOT/z" {
		t.Errorf("expected dependencies[2] = ROOT/z, got %q", chain.Dependencies[2].LogicalName)
	}
}

// TC-08: ARTIFACT dependency resolved from generating node
func TestChainResolve_ArtifactDependencyResolved(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "depends_on:\n  - ARTIFACT/b(lib)")
	testWriteNode(t, "b", "outputs:\n  - id: lib\n    path: out/lib.go")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("ChainResolve: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ARTIFACT/b(lib)" {
		t.Errorf("expected dependency logical name ARTIFACT/b(lib), got %q", dep.LogicalName)
	}
	if dep.FilePath == nil || dep.FilePath.Value != "out/lib.go" {
		t.Errorf("expected file_path out/lib.go, got %v", dep.FilePath)
	}
	if dep.Qualifier == nil {
		t.Fatal("expected qualifier present, got nil")
	}
	if *dep.Qualifier != "lib" {
		t.Errorf("expected qualifier lib, got %q", *dep.Qualifier)
	}
}

// TC-09: ARTIFACT without qualifier — error
func TestChainResolve_ArtifactWithoutQualifier_Error(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "depends_on:\n  - ARTIFACT/b")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// TC-10: ARTIFACT — generating node has no outputs
func TestChainResolve_ArtifactGeneratingNodeNoOutputs_Error(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "depends_on:\n  - ARTIFACT/b(lib)")
	testWriteNode(t, "b", "")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// TC-11: ARTIFACT — artifact file does not exist on disk
func TestChainResolve_ArtifactFileNotOnDisk_NoError(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "depends_on:\n  - ARTIFACT/b(lib)")
	testWriteNode(t, "b", "outputs:\n  - id: lib\n    path: out/lib.go")
	// Intentionally do NOT create out/lib.go

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.FilePath == nil || dep.FilePath.Value != "out/lib.go" {
		t.Errorf("expected file_path out/lib.go, got %v", dep.FilePath)
	}
}

// TC-12: ARTIFACT with non-existent output id — error
func TestChainResolve_ArtifactNonExistentOutputID_Error(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "depends_on:\n  - ARTIFACT/b(missing)")
	testWriteNode(t, "b", "outputs:\n  - id: lib\n    path: out/lib.go")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// TC-13: Mixed ROOT/ and ARTIFACT/ dependencies
func TestChainResolve_MixedDependencies_SortedByFilePath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "depends_on:\n  - ROOT/c\n  - ARTIFACT/b(lib)")
	testWriteNode(t, "b", "outputs:\n  - id: lib\n    path: out/lib.go")
	testWriteNode(t, "c", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("ChainResolve: %v", err)
	}

	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}
	// ROOT/c maps to code-from-spec/c/_node.md
	// ARTIFACT/b(lib) maps to out/lib.go
	// Sorted by file path: "code-from-spec/c/_node.md" vs "out/lib.go"
	// "code-from-spec/..." < "out/..." alphabetically
	if chain.Dependencies[0].LogicalName != "ROOT/c" {
		t.Errorf("expected dependencies[0] logical name ROOT/c, got %q", chain.Dependencies[0].LogicalName)
	}
	if chain.Dependencies[1].LogicalName != "ARTIFACT/b(lib)" {
		t.Errorf("expected dependencies[1] logical name ARTIFACT/b(lib), got %q", chain.Dependencies[1].LogicalName)
	}
}

// TC-14: Exact duplicate — same file, same qualifier
func TestChainResolve_ExactDuplicateDependency_Deduped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "depends_on:\n  - ROOT/b\n  - ROOT/b")
	testWriteNode(t, "b", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("ChainResolve: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency after dedup, got %d", len(chain.Dependencies))
	}
}

// TC-15: No qualifier subsumes qualifier
func TestChainResolve_UnqualifiedSubsumesQualified(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "depends_on:\n  - ROOT/b\n  - ROOT/b(interface)")
	testWriteNode(t, "b", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("ChainResolve: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency after dedup, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ROOT/b" {
		t.Errorf("expected dependency logical name ROOT/b, got %q", dep.LogicalName)
	}
	if dep.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %q", *dep.Qualifier)
	}
}

// TC-16: Qualifier before no-qualifier — no-qualifier wins
func TestChainResolve_QualifiedBeforeUnqualified_UnqualifiedWins(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "depends_on:\n  - ROOT/b(interface)\n  - ROOT/b")
	testWriteNode(t, "b", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("ChainResolve: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency after dedup, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ROOT/b" {
		t.Errorf("expected dependency logical name ROOT/b, got %q", dep.LogicalName)
	}
	if dep.Qualifier != nil {
		t.Errorf("expected qualifier absent, got %q", *dep.Qualifier)
	}
}

// TC-17: Same file, different qualifiers — both kept
func TestChainResolve_DifferentQualifiers_BothKept(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "depends_on:\n  - ROOT/b(interface)\n  - ROOT/b(constraints)")
	testWriteNode(t, "b", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("ChainResolve: %v", err)
	}

	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}
	// Sorted by qualifier: "constraints" before "interface"
	if chain.Dependencies[0].Qualifier == nil || *chain.Dependencies[0].Qualifier != "constraints" {
		t.Errorf("expected dependencies[0] qualifier constraints, got %v", chain.Dependencies[0].Qualifier)
	}
	if chain.Dependencies[1].Qualifier == nil || *chain.Dependencies[1].Qualifier != "interface" {
		t.Errorf("expected dependencies[1] qualifier interface, got %v", chain.Dependencies[1].Qualifier)
	}
}

// TC-18: Duplicate ARTIFACT — same logical name
func TestChainResolve_DuplicateArtifactDependency_Deduped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "depends_on:\n  - ARTIFACT/b(lib)\n  - ARTIFACT/b(lib)")
	testWriteNode(t, "b", "outputs:\n  - id: lib\n    path: out/lib.go")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("ChainResolve: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency after dedup, got %d", len(chain.Dependencies))
	}
}

// TC-19: External entries copied from frontmatter
func TestChainResolve_ExternalEntriesSorted(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "external:\n  - path: docs/api.yaml\n  - path: proto/v1.proto")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("ChainResolve: %v", err)
	}

	if len(chain.External) != 2 {
		t.Fatalf("expected 2 external entries, got %d", len(chain.External))
	}
	if chain.External[0].Path != "docs/api.yaml" {
		t.Errorf("expected external[0].Path = docs/api.yaml, got %q", chain.External[0].Path)
	}
	if chain.External[1].Path != "proto/v1.proto" {
		t.Errorf("expected external[1].Path = proto/v1.proto, got %q", chain.External[1].Path)
	}
}

// TC-20: External with fragments preserved
func TestChainResolve_ExternalFragmentsPreserved(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "external:\n  - path: f.txt\n    fragments:\n      - lines: \"1-10\"\n        hash: abc")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("ChainResolve: %v", err)
	}

	if len(chain.External) != 1 {
		t.Fatalf("expected 1 external entry, got %d", len(chain.External))
	}
	ext := chain.External[0]
	if ext.Path != "f.txt" {
		t.Errorf("expected path f.txt, got %q", ext.Path)
	}
	if len(ext.Fragments) != 1 {
		t.Fatalf("expected 1 fragment, got %d", len(ext.Fragments))
	}
	frag := ext.Fragments[0]
	if frag.Lines != "1-10" {
		t.Errorf("expected fragment lines 1-10, got %q", frag.Lines)
	}
	if frag.Hash != "abc" {
		t.Errorf("expected fragment hash abc, got %q", frag.Hash)
	}
}

// TC-21: Empty external — no entries
func TestChainResolve_EmptyExternal(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("ChainResolve: %v", err)
	}

	if len(chain.External) != 0 {
		t.Errorf("expected 0 external entries, got %d", len(chain.External))
	}
}

// TC-22: Input resolved from generating node
func TestChainResolve_InputResolved(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "input: ARTIFACT/b(data)")
	testWriteNode(t, "b", "outputs:\n  - id: data\n    path: out/data.json")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("ChainResolve: %v", err)
	}

	if chain.Input == nil {
		t.Fatal("expected input present, got nil")
	}
	if chain.Input.LogicalName != "ARTIFACT/b(data)" {
		t.Errorf("expected input logical name ARTIFACT/b(data), got %q", chain.Input.LogicalName)
	}
	if chain.Input.FilePath == nil || chain.Input.FilePath.Value != "out/data.json" {
		t.Errorf("expected input file_path out/data.json, got %v", chain.Input.FilePath)
	}
	if chain.Input.Qualifier == nil {
		t.Fatal("expected input qualifier present, got nil")
	}
	if *chain.Input.Qualifier != "data" {
		t.Errorf("expected input qualifier data, got %q", *chain.Input.Qualifier)
	}
}

// TC-23: No input — absent
func TestChainResolve_NoInput_Absent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("ChainResolve: %v", err)
	}

	if chain.Input != nil {
		t.Errorf("expected input absent, got %v", chain.Input)
	}
}

// TC-24: Input without qualifier — error
func TestChainResolve_InputWithoutQualifier_Error(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "input: ARTIFACT/b")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// TC-25: Input with non-existent output id — error
func TestChainResolve_InputNonExistentOutputID_Error(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "input: ARTIFACT/b(missing)")
	testWriteNode(t, "b", "outputs:\n  - id: data\n    path: out/data.json")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// TC-26: Unrecognized prefix in depends_on
func TestChainResolve_UnrecognizedPrefixInDependsOn_Error(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	testWriteNode(t, "a", "depends_on:\n  - UNKNOWN/something")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// TC-27: Invalid target logical name
func TestChainResolve_InvalidTargetLogicalName_Error(t *testing.T) {
	_, err := chainresolver.ChainResolve("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TC-28: Unreadable frontmatter
func TestChainResolve_UnreadableFrontmatter_Error(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteRootNode(t)
	// Write a _node.md with invalid YAML between --- delimiters
	if err := os.MkdirAll("code-from-spec/a", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	content := "---\n: invalid: yaml: [\n---\n"
	if err := os.WriteFile("code-from-spec/a/_node.md", []byte(content), 0644); err != nil {
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
