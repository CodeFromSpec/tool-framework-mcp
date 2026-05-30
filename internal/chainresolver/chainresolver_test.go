// code-from-spec: ROOT/golang/tests/chain/resolver@SyFJ2OtnhjDJL998NnQtItBHZTw
package chainresolver_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
)

// testChdir changes the working directory to dir and restores it on cleanup.
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

// testWriteNode creates a _node.md file at a path relative to the current
// working directory. It creates parent directories as needed.
func testWriteNode(t *testing.T, relPath string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(relPath), 0755); err != nil {
		t.Fatalf("testWriteNode mkdir: %v", err)
	}
	if err := os.WriteFile(relPath, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNode write: %v", err)
	}
}

// testEmptyFrontmatter is a _node.md with a valid but empty frontmatter block.
const testEmptyFrontmatter = "---\n---\n"

// testFrontmatter wraps YAML content in frontmatter delimiters.
func testFrontmatter(yaml string) string {
	return fmt.Sprintf("---\n%s---\n", yaml)
}

// testSetupCfsRoot creates the "code-from-spec" directory structure needed by
// logicalnames under the current working directory.
func testNodePath(logicalName string) string {
	// "ROOT" → "code-from-spec/_node.md"
	// "ROOT/a" → "code-from-spec/a/_node.md"
	if logicalName == "ROOT" {
		return "code-from-spec/_node.md"
	}
	// Strip "ROOT/" prefix
	rest := logicalName[len("ROOT/"):]
	return filepath.Join("code-from-spec", filepath.FromSlash(rest), "_node.md")
}

// --- TC-01: Root as target ---

func TestChainResolve_TC01_RootAsTarget(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)

	chain, err := chainresolver.ChainResolve("ROOT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Ancestors) != 0 {
		t.Errorf("expected 0 ancestors, got %d", len(chain.Ancestors))
	}
	if len(chain.Dependencies) != 0 {
		t.Errorf("expected 0 dependencies, got %d", len(chain.Dependencies))
	}
	if len(chain.External) != 0 {
		t.Errorf("expected 0 external, got %d", len(chain.External))
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
	if chain.Input != nil {
		t.Errorf("expected no input, got %+v", chain.Input)
	}
}

// --- TC-02: Linear chain — ancestors in root-first order ---

func TestChainResolve_TC02_LinearChainAncestors(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a/b"), testEmptyFrontmatter)

	chain, err := chainresolver.ChainResolve("ROOT/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

// --- TC-03: Single parent ---

func TestChainResolve_TC03_SingleParent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testEmptyFrontmatter)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

// --- TC-04: Target with empty frontmatter ---

func TestChainResolve_TC04_TargetEmptyFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testEmptyFrontmatter)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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
		t.Errorf("expected no input, got %+v", chain.Input)
	}
}

// --- TC-05: Dependency without qualifier ---

func TestChainResolve_TC05_DependencyNoQualifier(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testFrontmatter("depends_on:\n  - ROOT/b\n"))
	testWriteNode(t, testNodePath("ROOT/b"), testEmptyFrontmatter)

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
	if dep.Qualifier != nil {
		t.Errorf("expected no qualifier, got %q", *dep.Qualifier)
	}
}

// --- TC-06: Dependency with qualifier ---

func TestChainResolve_TC06_DependencyWithQualifier(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testFrontmatter("depends_on:\n  - ROOT/b(interface)\n"))
	testWriteNode(t, testNodePath("ROOT/b"), testEmptyFrontmatter)

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
	if dep.Qualifier == nil {
		t.Fatal("expected qualifier, got nil")
	}
	if *dep.Qualifier != "interface" {
		t.Errorf("expected qualifier interface, got %q", *dep.Qualifier)
	}
}

// --- TC-07: Dependencies sorted by file path then qualifier ---

func TestChainResolve_TC07_DependenciesSortedByFilePath(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testFrontmatter("depends_on:\n  - ROOT/z\n  - ROOT/m\n  - ROOT/b\n"))
	testWriteNode(t, testNodePath("ROOT/b"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/m"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/z"), testEmptyFrontmatter)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 3 {
		t.Fatalf("expected 3 dependencies, got %d", len(chain.Dependencies))
	}

	// Sorted by file path alphabetically: ROOT/b < ROOT/m < ROOT/z
	wantOrder := []string{"ROOT/b", "ROOT/m", "ROOT/z"}
	for i, want := range wantOrder {
		if chain.Dependencies[i].LogicalName != want {
			t.Errorf("dependencies[%d]: expected %q, got %q", i, want, chain.Dependencies[i].LogicalName)
		}
	}
}

// --- TC-08: ARTIFACT dependency resolved from generating node ---

func TestChainResolve_TC08_ArtifactDependencyResolved(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testFrontmatter("depends_on:\n  - ARTIFACT/b(lib)\n"))
	testWriteNode(t, testNodePath("ROOT/b"), testFrontmatter("outputs:\n  - id: lib\n    path: out/lib.go\n"))

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
	if dep.FilePath == nil || dep.FilePath.Value != "out/lib.go" {
		t.Errorf("expected file path out/lib.go, got %v", dep.FilePath)
	}
	if dep.Qualifier == nil || *dep.Qualifier != "lib" {
		t.Errorf("expected qualifier lib, got %v", dep.Qualifier)
	}
}

// --- TC-09: ARTIFACT without qualifier — error ---

func TestChainResolve_TC09_ArtifactNoQualifierError(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testFrontmatter("depends_on:\n  - ARTIFACT/b\n"))

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// --- TC-10: ARTIFACT — generating node has no outputs ---

func TestChainResolve_TC10_ArtifactGeneratorNoOutputs(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testFrontmatter("depends_on:\n  - ARTIFACT/b(lib)\n"))
	testWriteNode(t, testNodePath("ROOT/b"), testEmptyFrontmatter)

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// --- TC-11: ARTIFACT — artifact file does not exist on disk ---

func TestChainResolve_TC11_ArtifactFileNotOnDisk(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testFrontmatter("depends_on:\n  - ARTIFACT/b(lib)\n"))
	testWriteNode(t, testNodePath("ROOT/b"), testFrontmatter("outputs:\n  - id: lib\n    path: out/lib.go\n"))
	// Intentionally do NOT create out/lib.go

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.FilePath == nil || dep.FilePath.Value != "out/lib.go" {
		t.Errorf("expected file path out/lib.go, got %v", dep.FilePath)
	}
}

// --- TC-12: ARTIFACT with non-existent output id — error ---

func TestChainResolve_TC12_ArtifactNonExistentOutputId(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testFrontmatter("depends_on:\n  - ARTIFACT/b(missing)\n"))
	testWriteNode(t, testNodePath("ROOT/b"), testFrontmatter("outputs:\n  - id: lib\n    path: out/lib.go\n"))

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// --- TC-13: Mixed ROOT/ and ARTIFACT/ dependencies ---

func TestChainResolve_TC13_MixedDependencies(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testFrontmatter("depends_on:\n  - ROOT/c\n  - ARTIFACT/b(lib)\n"))
	testWriteNode(t, testNodePath("ROOT/b"), testFrontmatter("outputs:\n  - id: lib\n    path: out/lib.go\n"))
	testWriteNode(t, testNodePath("ROOT/c"), testEmptyFrontmatter)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}

	// Both must be present; check by logical name
	found := map[string]bool{}
	for _, dep := range chain.Dependencies {
		found[dep.LogicalName] = true
	}
	if !found["ROOT/c"] {
		t.Error("expected ROOT/c in dependencies")
	}
	if !found["ARTIFACT/b(lib)"] {
		t.Error("expected ARTIFACT/b(lib) in dependencies")
	}

	// Verify sorted by file path value
	// ROOT/c path: code-from-spec/c/_node.md
	// ARTIFACT/b(lib) path: out/lib.go
	// "code-from-spec/..." < "out/..." alphabetically
	if chain.Dependencies[0].LogicalName != "ROOT/c" {
		t.Errorf("expected first dependency ROOT/c (by file path), got %q", chain.Dependencies[0].LogicalName)
	}
	if chain.Dependencies[1].LogicalName != "ARTIFACT/b(lib)" {
		t.Errorf("expected second dependency ARTIFACT/b(lib), got %q", chain.Dependencies[1].LogicalName)
	}
}

// --- TC-14: Exact duplicate — same file, same qualifier ---

func TestChainResolve_TC14_ExactDuplicate(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testFrontmatter("depends_on:\n  - ROOT/b\n  - ROOT/b\n"))
	testWriteNode(t, testNodePath("ROOT/b"), testEmptyFrontmatter)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency (deduplicated), got %d", len(chain.Dependencies))
	}
}

// --- TC-15: No qualifier subsumes qualifier ---

func TestChainResolve_TC15_NoQualifierSubsumesQualifier(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testFrontmatter("depends_on:\n  - ROOT/b\n  - ROOT/b(interface)\n"))
	testWriteNode(t, testNodePath("ROOT/b"), testEmptyFrontmatter)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.Qualifier != nil {
		t.Errorf("expected qualifier to be absent (no-qualifier wins), got %q", *dep.Qualifier)
	}
}

// --- TC-16: Qualifier before no-qualifier — no-qualifier wins ---

func TestChainResolve_TC16_QualifierBeforeNoQualifierNoQualifierWins(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testFrontmatter("depends_on:\n  - ROOT/b(interface)\n  - ROOT/b\n"))
	testWriteNode(t, testNodePath("ROOT/b"), testEmptyFrontmatter)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.Qualifier != nil {
		t.Errorf("expected qualifier to be absent (no-qualifier wins), got %q", *dep.Qualifier)
	}
}

// --- TC-17: Same file, different qualifiers — both kept ---

func TestChainResolve_TC17_DifferentQualifiersBothKept(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testFrontmatter("depends_on:\n  - ROOT/b(interface)\n  - ROOT/b(constraints)\n"))
	testWriteNode(t, testNodePath("ROOT/b"), testEmptyFrontmatter)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}

	qualifiers := map[string]bool{}
	for _, dep := range chain.Dependencies {
		if dep.Qualifier != nil {
			qualifiers[*dep.Qualifier] = true
		}
	}
	if !qualifiers["interface"] {
		t.Error("expected qualifier 'interface' to be present")
	}
	if !qualifiers["constraints"] {
		t.Error("expected qualifier 'constraints' to be present")
	}

	// Sorted by qualifier: "constraints" < "interface"
	if chain.Dependencies[0].Qualifier == nil || *chain.Dependencies[0].Qualifier != "constraints" {
		t.Errorf("expected first dep qualifier=constraints, got %v", chain.Dependencies[0].Qualifier)
	}
	if chain.Dependencies[1].Qualifier == nil || *chain.Dependencies[1].Qualifier != "interface" {
		t.Errorf("expected second dep qualifier=interface, got %v", chain.Dependencies[1].Qualifier)
	}
}

// --- TC-18: Duplicate ARTIFACT — same logical name ---

func TestChainResolve_TC18_DuplicateArtifact(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testFrontmatter("depends_on:\n  - ARTIFACT/b(lib)\n  - ARTIFACT/b(lib)\n"))
	testWriteNode(t, testNodePath("ROOT/b"), testFrontmatter("outputs:\n  - id: lib\n    path: out/lib.go\n"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency (deduplicated), got %d", len(chain.Dependencies))
	}
}

// --- TC-19: External entries copied from frontmatter ---

func TestChainResolve_TC19_ExternalEntriesCopied(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testFrontmatter("external:\n  - path: docs/api.yaml\n  - path: proto/v1.proto\n"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.External) != 2 {
		t.Fatalf("expected 2 external entries, got %d", len(chain.External))
	}

	// Sorted alphabetically: "docs/api.yaml" < "proto/v1.proto"
	if chain.External[0].Path != "docs/api.yaml" {
		t.Errorf("expected external[0].Path = docs/api.yaml, got %q", chain.External[0].Path)
	}
	if chain.External[1].Path != "proto/v1.proto" {
		t.Errorf("expected external[1].Path = proto/v1.proto, got %q", chain.External[1].Path)
	}
}

// --- TC-20: External with fragments preserved ---

func TestChainResolve_TC20_ExternalFragmentsPreserved(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testFrontmatter(
		"external:\n  - path: f.txt\n    fragments:\n      - lines: \"1-10\"\n        hash: abc\n",
	))

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
		t.Errorf("expected fragment lines 1-10, got %q", frag.Lines)
	}
	if frag.Hash != "abc" {
		t.Errorf("expected fragment hash abc, got %q", frag.Hash)
	}
}

// --- TC-21: Empty external — no entries ---

func TestChainResolve_TC21_EmptyExternal(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testEmptyFrontmatter)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.External) != 0 {
		t.Errorf("expected 0 external entries, got %d", len(chain.External))
	}
}

// --- TC-22: Input resolved from generating node ---

func TestChainResolve_TC22_InputResolved(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testFrontmatter("input: \"ARTIFACT/b(data)\"\n"))
	testWriteNode(t, testNodePath("ROOT/b"), testFrontmatter("outputs:\n  - id: data\n    path: out/data.json\n"))

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
	if chain.Input.FilePath == nil || chain.Input.FilePath.Value != "out/data.json" {
		t.Errorf("expected input file path out/data.json, got %v", chain.Input.FilePath)
	}
	if chain.Input.Qualifier == nil || *chain.Input.Qualifier != "data" {
		t.Errorf("expected input qualifier data, got %v", chain.Input.Qualifier)
	}
}

// --- TC-23: No input — absent ---

func TestChainResolve_TC23_NoInput(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testEmptyFrontmatter)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chain.Input != nil {
		t.Errorf("expected no input, got %+v", chain.Input)
	}
}

// --- TC-24: Input without qualifier — error ---

func TestChainResolve_TC24_InputNoQualifierError(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testFrontmatter("input: \"ARTIFACT/b\"\n"))

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// --- TC-25: Input with non-existent output id — error ---

func TestChainResolve_TC25_InputNonExistentOutputId(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testFrontmatter("input: \"ARTIFACT/b(missing)\"\n"))
	testWriteNode(t, testNodePath("ROOT/b"), testFrontmatter("outputs:\n  - id: data\n    path: out/data.json\n"))

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// --- TC-26: Unrecognized prefix in depends_on ---

func TestChainResolve_TC26_UnrecognizedPrefixInDependsOn(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	testWriteNode(t, testNodePath("ROOT/a"), testFrontmatter("depends_on:\n  - UNKNOWN/something\n"))

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// --- TC-27: Invalid target logical name ---

func TestChainResolve_TC27_InvalidTargetLogicalName(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := chainresolver.ChainResolve("INVALID/something")
	if err == nil {
		t.Fatal("expected error for invalid logical name, got nil")
	}
}

// --- TC-28: Unreadable frontmatter ---

func TestChainResolve_TC28_UnreadableFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, testNodePath("ROOT"), testEmptyFrontmatter)
	// Write a file with invalid YAML between frontmatter delimiters
	testWriteNode(t, testNodePath("ROOT/a"), "---\n: : invalid yaml : :\n---\n")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error for malformed YAML, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML (or wrapping it), got %v", err)
	}
}
