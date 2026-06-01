// code-from-spec: ROOT/golang/tests/chain/resolver@Nr_CPN5-ecBBXE9XjAZsk0uRVR0

package chainresolver_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
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

// testWriteNode creates a _node.md file at the given CFS path (relative to cwd)
// with the provided frontmatter content.
func testWriteNode(t *testing.T, cfsPath string, frontmatter string) {
	t.Helper()
	dir := filepath.Dir(cfsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testWriteNode mkdir %s: %v", dir, err)
	}
	content := "---\n" + frontmatter + "---\n"
	if err := os.WriteFile(cfsPath, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNode write %s: %v", cfsPath, err)
	}
}

// testEmptyNode creates a _node.md file with empty frontmatter.
func testEmptyNode(t *testing.T, cfsPath string) {
	t.Helper()
	testWriteNode(t, cfsPath, "")
}

// testLogicalNameToFilePath returns the file path that LogicalNameToPath would produce,
// in OS-native form relative to the temp dir. Used for constructing spec tree files.
func testNodeFilePath(logicalName string) string {
	// ROOT      -> code-from-spec/_node.md
	// ROOT/a    -> code-from-spec/a/_node.md
	// ROOT/a/b  -> code-from-spec/a/b/_node.md
	stripped := logicalnames.LogicalNameStripQualifier(logicalName)
	if stripped == "ROOT" {
		return filepath.Join("code-from-spec", "_node.md")
	}
	// Remove "ROOT/" prefix.
	rel := stripped[len("ROOT/"):]
	parts := filepath.FromSlash(rel)
	return filepath.Join("code-from-spec", parts, "_node.md")
}

// ========================
// TC-AT: Ancestors and Target
// ========================

// TC-AT-01: Root as target
func TestChainResolve_AT01_RootAsTarget(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))

	chain, err := chainresolver.ChainResolve("ROOT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Ancestors) != 0 {
		t.Errorf("expected 0 ancestors, got %d", len(chain.Ancestors))
	}
	if chain.Target == nil {
		t.Fatal("expected non-nil target")
	}
	if chain.Target.LogicalName != "ROOT" {
		t.Errorf("expected target logical name ROOT, got %q", chain.Target.LogicalName)
	}
	if chain.Target.Qualifier != "" {
		t.Errorf("expected no qualifier, got %q", chain.Target.Qualifier)
	}
	if len(chain.Dependencies) != 0 {
		t.Errorf("expected 0 dependencies, got %d", len(chain.Dependencies))
	}
	if len(chain.External) != 0 {
		t.Errorf("expected 0 external, got %d", len(chain.External))
	}
	if chain.Input != nil {
		t.Errorf("expected nil input, got %+v", chain.Input)
	}
}

// TC-AT-02: Linear chain — ancestors in root-first order
func TestChainResolve_AT02_LinearChainAncestors(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testEmptyNode(t, testNodeFilePath("ROOT/a"))
	testEmptyNode(t, testNodeFilePath("ROOT/a/b"))

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
	if chain.Target == nil {
		t.Fatal("expected non-nil target")
	}
	if chain.Target.LogicalName != "ROOT/a/b" {
		t.Errorf("expected target = ROOT/a/b, got %q", chain.Target.LogicalName)
	}
}

// TC-AT-03: Single parent
func TestChainResolve_AT03_SingleParent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testEmptyNode(t, testNodeFilePath("ROOT/a"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Ancestors) != 1 {
		t.Fatalf("expected 1 ancestor, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].LogicalName != "ROOT" {
		t.Errorf("expected ancestor[0] = ROOT, got %q", chain.Ancestors[0].LogicalName)
	}
	if chain.Target.LogicalName != "ROOT/a" {
		t.Errorf("expected target = ROOT/a, got %q", chain.Target.LogicalName)
	}
}

// TC-AT-04: Target with empty frontmatter
func TestChainResolve_AT04_EmptyFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testEmptyNode(t, testNodeFilePath("ROOT/a"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 0 {
		t.Errorf("expected 0 dependencies, got %d", len(chain.Dependencies))
	}
	if len(chain.External) != 0 {
		t.Errorf("expected 0 external, got %d", len(chain.External))
	}
	if chain.Input != nil {
		t.Errorf("expected nil input")
	}
}

// ========================
// TC-DEP: Dependencies — ROOT/ References
// ========================

// TC-DEP-01: Dependency without qualifier
func TestChainResolve_DEP01_DependencyNoQualifier(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testWriteNode(t, testNodeFilePath("ROOT/a"), "depends_on:\n  - ROOT/b\n")
	testEmptyNode(t, testNodeFilePath("ROOT/b"))

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

// TC-DEP-02: Dependency with qualifier
func TestChainResolve_DEP02_DependencyWithQualifier(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testWriteNode(t, testNodeFilePath("ROOT/a"), "depends_on:\n  - ROOT/b(interface)\n")
	testEmptyNode(t, testNodeFilePath("ROOT/b"))

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
		t.Errorf("expected qualifier 'interface', got %q", dep.Qualifier)
	}
}

// TC-DEP-03: Dependencies sorted by file path then qualifier
func TestChainResolve_DEP03_DependenciesSortedByFilePath(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testWriteNode(t, testNodeFilePath("ROOT/a"), "depends_on:\n  - ROOT/z\n  - ROOT/m\n  - ROOT/b\n")
	testEmptyNode(t, testNodeFilePath("ROOT/z"))
	testEmptyNode(t, testNodeFilePath("ROOT/m"))
	testEmptyNode(t, testNodeFilePath("ROOT/b"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

// ========================
// TC-ART: Dependencies — ARTIFACT/ References
// ========================

// TC-ART-01: ARTIFACT dependency resolved from generating node
func TestChainResolve_ART01_ArtifactDependency(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testWriteNode(t, testNodeFilePath("ROOT/a"), "depends_on:\n  - ARTIFACT/b(lib)\n")
	testWriteNode(t, testNodeFilePath("ROOT/b"), "outputs:\n  - id: lib\n    path: out/lib.go\n")

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
	if dep.Qualifier != "lib" {
		t.Errorf("expected qualifier 'lib', got %q", dep.Qualifier)
	}
}

// TC-ART-02: ARTIFACT without qualifier — error
func TestChainResolve_ART02_ArtifactNoQualifierError(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testWriteNode(t, testNodeFilePath("ROOT/a"), "depends_on:\n  - ARTIFACT/b\n")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// TC-ART-03: ARTIFACT — generating node has no outputs
func TestChainResolve_ART03_ArtifactNoOutputs(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testWriteNode(t, testNodeFilePath("ROOT/a"), "depends_on:\n  - ARTIFACT/b(lib)\n")
	testEmptyNode(t, testNodeFilePath("ROOT/b"))

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// TC-ART-04: ARTIFACT — artifact file does not exist on disk (no error expected)
func TestChainResolve_ART04_ArtifactFileMissing(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testWriteNode(t, testNodeFilePath("ROOT/a"), "depends_on:\n  - ARTIFACT/b(lib)\n")
	testWriteNode(t, testNodeFilePath("ROOT/b"), "outputs:\n  - id: lib\n    path: out/lib.go\n")
	// Do NOT create out/lib.go on disk.

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.FilePath.Value != "out/lib.go" {
		t.Errorf("expected file_path out/lib.go, got %q", dep.FilePath.Value)
	}
}

// TC-ART-05: ARTIFACT with non-existent output id — error
func TestChainResolve_ART05_ArtifactMissingOutputID(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testWriteNode(t, testNodeFilePath("ROOT/a"), "depends_on:\n  - ARTIFACT/b(missing)\n")
	testWriteNode(t, testNodeFilePath("ROOT/b"), "outputs:\n  - id: lib\n    path: out/lib.go\n")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// TC-ART-06: Mixed ROOT/ and ARTIFACT/ dependencies
func TestChainResolve_ART06_MixedDependencies(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testWriteNode(t, testNodeFilePath("ROOT/a"), "depends_on:\n  - ROOT/c\n  - ARTIFACT/b(lib)\n")
	testWriteNode(t, testNodeFilePath("ROOT/b"), "outputs:\n  - id: lib\n    path: out/lib.go\n")
	testEmptyNode(t, testNodeFilePath("ROOT/c"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}

	// Verify both are present (exact order depends on file path sorting).
	foundRootC := false
	foundArtifactB := false
	for _, dep := range chain.Dependencies {
		if dep.LogicalName == "ROOT/c" {
			foundRootC = true
		}
		if dep.LogicalName == "ARTIFACT/b(lib)" {
			foundArtifactB = true
		}
	}
	if !foundRootC {
		t.Error("expected ROOT/c in dependencies")
	}
	if !foundArtifactB {
		t.Error("expected ARTIFACT/b(lib) in dependencies")
	}
}

// ========================
// TC-DEDUP: Dependencies — Dedup
// ========================

// TC-DEDUP-01: Exact duplicate — same file, same qualifier
func TestChainResolve_DEDUP01_ExactDuplicate(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testWriteNode(t, testNodeFilePath("ROOT/a"), "depends_on:\n  - ROOT/b\n  - ROOT/b\n")
	testEmptyNode(t, testNodeFilePath("ROOT/b"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency (deduped), got %d", len(chain.Dependencies))
	}
}

// TC-DEDUP-02: No qualifier subsumes qualifier
func TestChainResolve_DEDUP02_NoQualifierSubsumesQualifier(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testWriteNode(t, testNodeFilePath("ROOT/a"), "depends_on:\n  - ROOT/b\n  - ROOT/b(interface)\n")
	testEmptyNode(t, testNodeFilePath("ROOT/b"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency (unqualified subsumes qualified), got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].Qualifier != "" {
		t.Errorf("expected no qualifier, got %q", chain.Dependencies[0].Qualifier)
	}
}

// TC-DEDUP-03: Qualifier before no-qualifier — no-qualifier wins
func TestChainResolve_DEDUP03_QualifierBeforeNoQualifier(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testWriteNode(t, testNodeFilePath("ROOT/a"), "depends_on:\n  - ROOT/b(interface)\n  - ROOT/b\n")
	testEmptyNode(t, testNodeFilePath("ROOT/b"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency (unqualified wins), got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].Qualifier != "" {
		t.Errorf("expected no qualifier, got %q", chain.Dependencies[0].Qualifier)
	}
}

// TC-DEDUP-04: Same file, different qualifiers — both kept
func TestChainResolve_DEDUP04_DifferentQualifiersKept(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testWriteNode(t, testNodeFilePath("ROOT/a"), "depends_on:\n  - ROOT/b(interface)\n  - ROOT/b(constraints)\n")
	testEmptyNode(t, testNodeFilePath("ROOT/b"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies (both kept), got %d", len(chain.Dependencies))
	}
	// Sorted by qualifier alphabetically.
	if chain.Dependencies[0].Qualifier != "constraints" {
		t.Errorf("expected dependencies[0].qualifier = constraints, got %q", chain.Dependencies[0].Qualifier)
	}
	if chain.Dependencies[1].Qualifier != "interface" {
		t.Errorf("expected dependencies[1].qualifier = interface, got %q", chain.Dependencies[1].Qualifier)
	}
}

// TC-DEDUP-05: Duplicate ARTIFACT — same logical name
func TestChainResolve_DEDUP05_DuplicateArtifact(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testWriteNode(t, testNodeFilePath("ROOT/a"), "depends_on:\n  - ARTIFACT/b(lib)\n  - ARTIFACT/b(lib)\n")
	testWriteNode(t, testNodeFilePath("ROOT/b"), "outputs:\n  - id: lib\n    path: out/lib.go\n")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency (deduped), got %d", len(chain.Dependencies))
	}
}

// ========================
// TC-EXT: External
// ========================

// TC-EXT-01: External entries copied from frontmatter
func TestChainResolve_EXT01_ExternalEntries(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testWriteNode(t, testNodeFilePath("ROOT/a"),
		"external:\n  - path: docs/api.yaml\n  - path: proto/v1.proto\n")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.External) != 2 {
		t.Fatalf("expected 2 external entries, got %d", len(chain.External))
	}
	// Sorted alphabetically.
	if chain.External[0].Path != "docs/api.yaml" {
		t.Errorf("expected external[0].path = docs/api.yaml, got %q", chain.External[0].Path)
	}
	if chain.External[1].Path != "proto/v1.proto" {
		t.Errorf("expected external[1].path = proto/v1.proto, got %q", chain.External[1].Path)
	}
}

// TC-EXT-02: External with fragments preserved
func TestChainResolve_EXT02_ExternalFragmentsPreserved(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testWriteNode(t, testNodeFilePath("ROOT/a"),
		"external:\n  - path: f.txt\n    fragments:\n      - lines: \"1-10\"\n        hash: abc\n")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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
		t.Errorf("expected fragment lines '1-10', got %q", frag.Lines)
	}
	if frag.Hash != "abc" {
		t.Errorf("expected fragment hash 'abc', got %q", frag.Hash)
	}
}

// TC-EXT-03: Empty external — no entries
func TestChainResolve_EXT03_NoExternal(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testEmptyNode(t, testNodeFilePath("ROOT/a"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.External) != 0 {
		t.Errorf("expected 0 external entries, got %d", len(chain.External))
	}
}

// ========================
// TC-INP: Input
// ========================

// TC-INP-01: Input resolved from generating node
func TestChainResolve_INP01_InputResolved(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testWriteNode(t, testNodeFilePath("ROOT/a"), "input: \"ARTIFACT/b(data)\"\n")
	testWriteNode(t, testNodeFilePath("ROOT/b"), "outputs:\n  - id: data\n    path: out/data.json\n")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chain.Input == nil {
		t.Fatal("expected non-nil input")
	}
	if chain.Input.LogicalName != "ARTIFACT/b(data)" {
		t.Errorf("expected input logical name ARTIFACT/b(data), got %q", chain.Input.LogicalName)
	}
	if chain.Input.FilePath.Value != "out/data.json" {
		t.Errorf("expected input file_path out/data.json, got %q", chain.Input.FilePath.Value)
	}
	if chain.Input.Qualifier != "data" {
		t.Errorf("expected input qualifier 'data', got %q", chain.Input.Qualifier)
	}
}

// TC-INP-02: No input — absent
func TestChainResolve_INP02_NoInput(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testEmptyNode(t, testNodeFilePath("ROOT/a"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chain.Input != nil {
		t.Errorf("expected nil input, got %+v", chain.Input)
	}
}

// TC-INP-03: Input without qualifier — error
func TestChainResolve_INP03_InputNoQualifierError(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testWriteNode(t, testNodeFilePath("ROOT/a"), "input: \"ARTIFACT/b\"\n")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// TC-INP-04: Input with non-existent output id — error
func TestChainResolve_INP04_InputMissingOutputID(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testWriteNode(t, testNodeFilePath("ROOT/a"), "input: \"ARTIFACT/b(missing)\"\n")
	testWriteNode(t, testNodeFilePath("ROOT/b"), "outputs:\n  - id: data\n    path: out/data.json\n")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// ========================
// TC-ERR: Error Cases
// ========================

// TC-ERR-01: Unrecognized prefix in depends_on
func TestChainResolve_ERR01_UnrecognizedPrefix(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))
	testWriteNode(t, testNodeFilePath("ROOT/a"), "depends_on:\n  - UNKNOWN/something\n")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// TC-ERR-02: Invalid target logical name
func TestChainResolve_ERR02_InvalidTargetLogicalName(t *testing.T) {
	_, err := chainresolver.ChainResolve("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Should propagate from LogicalNameGetParent or LogicalNameToPath.
	// We just verify an error is returned.
}

// TC-ERR-03: Unreadable frontmatter
func TestChainResolve_ERR03_UnreadableFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testEmptyNode(t, testNodeFilePath("ROOT"))

	// Write a _node.md with malformed YAML in frontmatter.
	nodePath := testNodeFilePath("ROOT/a")
	dir := filepath.Dir(nodePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	malformedContent := "---\ndepends_on: [unclosed\n---\n"
	if err := os.WriteFile(nodePath, []byte(malformedContent), 0644); err != nil {
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
