// code-from-spec: ROOT/golang/tests/chain/resolver@B5oGnUJIgLEeYs59mJALC5xdoHQ
package chainresolver_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

// testWriteNode writes a _node.md file at the given path segments under the
// working directory. The frontmatter is the raw string between the --- delimiters.
// Pass an empty string for an empty frontmatter block.
func testWriteNode(t *testing.T, frontmatter string, pathSegments ...string) {
	t.Helper()
	dir := filepath.Join(pathSegments[:len(pathSegments)-1]...)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("testWriteNode MkdirAll: %v", err)
	}
	content := fmt.Sprintf("---\n%s---\n", frontmatter)
	full := filepath.Join(pathSegments...)
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("testWriteNode WriteFile: %v", err)
	}
}

// testSetupRoot creates the code-from-spec root directory structure inside
// the given temp dir and changes to that directory.
func testSetupRoot(t *testing.T) {
	t.Helper()
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	if err := os.MkdirAll("code-from-spec", 0o755); err != nil {
		t.Fatalf("testSetupRoot MkdirAll: %v", err)
	}
}

// testNodePath returns the path to a _node.md file for a given set of path
// segments under code-from-spec/.
func testNodePath(segments ...string) []string {
	parts := append([]string{"code-from-spec"}, segments...)
	return append(parts, "_node.md")
}

// --- TC-01: Root as target ---

func TestChainResolve_TC01_RootAsTarget(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)

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
		t.Fatal("expected non-nil target")
	}
	if chain.Target.LogicalName != "ROOT" {
		t.Errorf("expected target logical name ROOT, got %q", chain.Target.LogicalName)
	}
	if chain.Target.Qualifier != nil {
		t.Errorf("expected nil qualifier, got %q", *chain.Target.Qualifier)
	}
	if chain.Input != nil {
		t.Errorf("expected nil input, got %+v", chain.Input)
	}
}

// --- TC-02: Linear chain — ancestors in root-first order ---

func TestChainResolve_TC02_LinearChainAncestors(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "", testNodePath("a")...)
	testWriteNode(t, "", testNodePath("a", "b")...)

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
	if chain.Target == nil {
		t.Fatal("expected non-nil target")
	}
	if chain.Target.LogicalName != "ROOT/a/b" {
		t.Errorf("expected target ROOT/a/b, got %q", chain.Target.LogicalName)
	}
}

// --- TC-03: Single parent ---

func TestChainResolve_TC03_SingleParent(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "", testNodePath("a")...)

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
	if chain.Target.LogicalName != "ROOT/a" {
		t.Errorf("expected target ROOT/a, got %q", chain.Target.LogicalName)
	}
}

// --- TC-04: Target with empty frontmatter ---

func TestChainResolve_TC04_EmptyFrontmatter(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "", testNodePath("a")...)

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
		t.Errorf("expected nil input, got %+v", chain.Input)
	}
	if chain.Target.LogicalName != "ROOT/a" {
		t.Errorf("expected target ROOT/a, got %q", chain.Target.LogicalName)
	}
}

// --- TC-05: Dependency without qualifier ---

func TestChainResolve_TC05_DependencyWithoutQualifier(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "depends_on:\n  - ROOT/b\n", testNodePath("a")...)
	testWriteNode(t, "", testNodePath("b")...)

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
		t.Errorf("expected nil qualifier, got %q", *dep.Qualifier)
	}
}

// --- TC-06: Dependency with qualifier ---

func TestChainResolve_TC06_DependencyWithQualifier(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "depends_on:\n  - ROOT/b(interface)\n", testNodePath("a")...)
	testWriteNode(t, "", testNodePath("b")...)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ROOT/b" {
		t.Errorf("expected dependency logical name ROOT/b, got %q", dep.LogicalName)
	}
	if dep.Qualifier == nil {
		t.Fatal("expected non-nil qualifier")
	}
	if *dep.Qualifier != "interface" {
		t.Errorf("expected qualifier 'interface', got %q", *dep.Qualifier)
	}
}

// --- TC-07: Dependencies sorted by file path then qualifier ---

func TestChainResolve_TC07_DependenciesSortedByFilePath(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "depends_on:\n  - ROOT/z\n  - ROOT/m\n  - ROOT/b\n", testNodePath("a")...)
	testWriteNode(t, "", testNodePath("b")...)
	testWriteNode(t, "", testNodePath("m")...)
	testWriteNode(t, "", testNodePath("z")...)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 3 {
		t.Fatalf("expected 3 dependencies, got %d", len(chain.Dependencies))
	}
	// Expect alphabetical order by file path: b, m, z
	wantOrder := []string{"ROOT/b", "ROOT/m", "ROOT/z"}
	for i, dep := range chain.Dependencies {
		if dep.LogicalName != wantOrder[i] {
			t.Errorf("dependencies[%d]: expected %q, got %q", i, wantOrder[i], dep.LogicalName)
		}
	}
}

// --- TC-08: ARTIFACT dependency resolved from generating node ---

func TestChainResolve_TC08_ArtifactDependencyResolved(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "depends_on:\n  - ARTIFACT/b(lib)\n", testNodePath("a")...)
	testWriteNode(t, "outputs:\n  - id: lib\n    path: out/lib.go\n", testNodePath("b")...)

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
	if dep.FilePath == nil {
		t.Fatal("expected non-nil file path")
	}
	if dep.FilePath.Value != "out/lib.go" {
		t.Errorf("expected file path out/lib.go, got %q", dep.FilePath.Value)
	}
	if dep.Qualifier == nil {
		t.Fatal("expected non-nil qualifier")
	}
	if *dep.Qualifier != "lib" {
		t.Errorf("expected qualifier 'lib', got %q", *dep.Qualifier)
	}
}

// --- TC-09: ARTIFACT without qualifier — error ---

func TestChainResolve_TC09_ArtifactWithoutQualifier_Error(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "depends_on:\n  - ARTIFACT/b\n", testNodePath("a")...)

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// --- TC-10: ARTIFACT — generating node has no outputs ---

func TestChainResolve_TC10_ArtifactGeneratingNodeNoOutputs_Error(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "depends_on:\n  - ARTIFACT/b(lib)\n", testNodePath("a")...)
	testWriteNode(t, "", testNodePath("b")...)

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// --- TC-11: ARTIFACT — artifact file does not exist on disk ---

func TestChainResolve_TC11_ArtifactFileNotOnDisk_NoError(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "depends_on:\n  - ARTIFACT/b(lib)\n", testNodePath("a")...)
	testWriteNode(t, "outputs:\n  - id: lib\n    path: out/lib.go\n", testNodePath("b")...)
	// out/lib.go is intentionally NOT created

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.FilePath == nil {
		t.Fatal("expected non-nil file path")
	}
	if dep.FilePath.Value != "out/lib.go" {
		t.Errorf("expected file path out/lib.go, got %q", dep.FilePath.Value)
	}
}

// --- TC-12: ARTIFACT with non-existent output id — error ---

func TestChainResolve_TC12_ArtifactNonExistentOutputID_Error(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "depends_on:\n  - ARTIFACT/b(missing)\n", testNodePath("a")...)
	testWriteNode(t, "outputs:\n  - id: lib\n    path: out/lib.go\n", testNodePath("b")...)

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
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "depends_on:\n  - ROOT/c\n  - ARTIFACT/b(lib)\n", testNodePath("a")...)
	testWriteNode(t, "outputs:\n  - id: lib\n    path: out/lib.go\n", testNodePath("b")...)
	testWriteNode(t, "", testNodePath("c")...)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}

	// Find entries by checking logical names
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
		t.Error("expected dependency ROOT/c not found")
	}
	if !foundArtifactB {
		t.Error("expected dependency ARTIFACT/b(lib) not found")
	}
}

// --- TC-14: Exact duplicate — same file, same qualifier ---

func TestChainResolve_TC14_ExactDuplicateDependency(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "depends_on:\n  - ROOT/b\n  - ROOT/b\n", testNodePath("a")...)
	testWriteNode(t, "", testNodePath("b")...)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency after dedup, got %d", len(chain.Dependencies))
	}
}

// --- TC-15: No qualifier subsumes qualifier ---

func TestChainResolve_TC15_NoQualifierSubsumesQualifier(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "depends_on:\n  - ROOT/b\n  - ROOT/b(interface)\n", testNodePath("a")...)
	testWriteNode(t, "", testNodePath("b")...)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency after dedup, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ROOT/b" {
		t.Errorf("expected logical name ROOT/b, got %q", dep.LogicalName)
	}
	if dep.Qualifier != nil {
		t.Errorf("expected nil qualifier (no-qualifier wins), got %q", *dep.Qualifier)
	}
}

// --- TC-16: Qualifier before no-qualifier — no-qualifier wins ---

func TestChainResolve_TC16_QualifierBeforeNoQualifier_NoQualifierWins(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "depends_on:\n  - ROOT/b(interface)\n  - ROOT/b\n", testNodePath("a")...)
	testWriteNode(t, "", testNodePath("b")...)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency after dedup, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.Qualifier != nil {
		t.Errorf("expected nil qualifier (no-qualifier wins), got %q", *dep.Qualifier)
	}
}

// --- TC-17: Same file, different qualifiers — both kept ---

func TestChainResolve_TC17_SameFileDifferentQualifiers_BothKept(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "depends_on:\n  - ROOT/b(interface)\n  - ROOT/b(constraints)\n", testNodePath("a")...)
	testWriteNode(t, "", testNodePath("b")...)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}

	foundConstraints := false
	foundInterface := false
	for _, dep := range chain.Dependencies {
		if dep.Qualifier != nil && *dep.Qualifier == "constraints" {
			foundConstraints = true
		}
		if dep.Qualifier != nil && *dep.Qualifier == "interface" {
			foundInterface = true
		}
	}
	if !foundConstraints {
		t.Error("expected dependency with qualifier 'constraints' not found")
	}
	if !foundInterface {
		t.Error("expected dependency with qualifier 'interface' not found")
	}
}

// --- TC-18: Duplicate ARTIFACT — same logical name ---

func TestChainResolve_TC18_DuplicateArtifact_Dedup(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "depends_on:\n  - ARTIFACT/b(lib)\n  - ARTIFACT/b(lib)\n", testNodePath("a")...)
	testWriteNode(t, "outputs:\n  - id: lib\n    path: out/lib.go\n", testNodePath("b")...)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency after dedup, got %d", len(chain.Dependencies))
	}
}

// --- TC-19: External entries copied from frontmatter ---

func TestChainResolve_TC19_ExternalEntriesCopied(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "external:\n  - path: docs/api.yaml\n  - path: proto/v1.proto\n", testNodePath("a")...)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.External) != 2 {
		t.Fatalf("expected 2 external entries, got %d", len(chain.External))
	}
	// Sorted alphabetically: docs/api.yaml < proto/v1.proto
	if chain.External[0].Path != "docs/api.yaml" {
		t.Errorf("expected first external path docs/api.yaml, got %q", chain.External[0].Path)
	}
	if chain.External[1].Path != "proto/v1.proto" {
		t.Errorf("expected second external path proto/v1.proto, got %q", chain.External[1].Path)
	}
}

// --- TC-20: External with fragments preserved ---

func TestChainResolve_TC20_ExternalFragmentsPreserved(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t,
		"external:\n  - path: f.txt\n    fragments:\n      - lines: \"1-10\"\n        hash: abc\n",
		testNodePath("a")...,
	)

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
		t.Errorf("expected lines '1-10', got %q", frag.Lines)
	}
	if frag.Hash != "abc" {
		t.Errorf("expected hash 'abc', got %q", frag.Hash)
	}
}

// --- TC-21: Empty external — no entries ---

func TestChainResolve_TC21_EmptyExternal(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "", testNodePath("a")...)

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
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "input: \"ARTIFACT/b(data)\"\n", testNodePath("a")...)
	testWriteNode(t, "outputs:\n  - id: data\n    path: out/data.json\n", testNodePath("b")...)

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
	if chain.Input.FilePath == nil {
		t.Fatal("expected non-nil input file path")
	}
	if chain.Input.FilePath.Value != "out/data.json" {
		t.Errorf("expected input file path out/data.json, got %q", chain.Input.FilePath.Value)
	}
	if chain.Input.Qualifier == nil {
		t.Fatal("expected non-nil input qualifier")
	}
	if *chain.Input.Qualifier != "data" {
		t.Errorf("expected input qualifier 'data', got %q", *chain.Input.Qualifier)
	}
}

// --- TC-23: No input — absent ---

func TestChainResolve_TC23_NoInput_Absent(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "", testNodePath("a")...)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chain.Input != nil {
		t.Errorf("expected nil input, got %+v", chain.Input)
	}
}

// --- TC-24: Input without qualifier — error ---

func TestChainResolve_TC24_InputWithoutQualifier_Error(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "input: \"ARTIFACT/b\"\n", testNodePath("a")...)

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// --- TC-25: Input with non-existent output id — error ---

func TestChainResolve_TC25_InputNonExistentOutputID_Error(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "input: \"ARTIFACT/b(missing)\"\n", testNodePath("a")...)
	testWriteNode(t, "outputs:\n  - id: data\n    path: out/data.json\n", testNodePath("b")...)

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// --- TC-26: Unrecognized prefix in depends_on ---

func TestChainResolve_TC26_UnrecognizedPrefix_Error(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)
	testWriteNode(t, "depends_on:\n  - UNKNOWN/something\n", testNodePath("a")...)

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// --- TC-27: Invalid target logical name ---

func TestChainResolve_TC27_InvalidTargetLogicalName_Error(t *testing.T) {
	testSetupRoot(t)

	_, err := chainresolver.ChainResolve("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- TC-28: Unreadable frontmatter ---

func TestChainResolve_TC28_UnreadableFrontmatter_Error(t *testing.T) {
	testSetupRoot(t)
	testWriteNode(t, "", testNodePath()...)

	// Write invalid YAML in frontmatter
	badContent := "---\nkey: [unclosed bracket\n---\n"
	if err := os.WriteFile(filepath.Join("code-from-spec", "a", "_node.md"), []byte(badContent), 0o644); err != nil {
		// File may not exist yet — create dirs first
		if mkErr := os.MkdirAll(filepath.Join("code-from-spec", "a"), 0o755); mkErr != nil {
			t.Fatalf("MkdirAll: %v", mkErr)
		}
		if err2 := os.WriteFile(filepath.Join("code-from-spec", "a", "_node.md"), []byte(badContent), 0o644); err2 != nil {
			t.Fatalf("WriteFile: %v", err2)
		}
	}

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
