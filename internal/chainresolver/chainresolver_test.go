// code-from-spec: ROOT/golang/tests/chain/chain_resolver@ToDJrBvswf6lNrPYN2FwDJtWZwc
package chainresolver_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
)

// testChdir changes the working directory to dir for the duration of the test,
// restoring the original directory in a cleanup function.
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

// testWriteNode writes a _node.md file for the given logical name segment path
// (relative to code-from-spec/) with the given frontmatter content.
// The frontmatter is wrapped in --- delimiters automatically.
func testWriteNode(t *testing.T, logicalPath string, frontmatter string) {
	t.Helper()
	dir := filepath.Join("code-from-spec", filepath.FromSlash(logicalPath))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("testWriteNode: mkdir %s: %v", dir, err)
	}
	var content string
	if frontmatter == "" {
		content = "---\n---\n"
	} else {
		content = "---\n" + frontmatter + "\n---\n"
	}
	path := filepath.Join(dir, "_node.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("testWriteNode: write %s: %v", path, err)
	}
}

// testWriteNodeRaw writes a _node.md file for the given logical name segment
// path with completely raw file content (not wrapped in frontmatter delimiters).
func testWriteNodeRaw(t *testing.T, logicalPath string, rawContent string) {
	t.Helper()
	dir := filepath.Join("code-from-spec", filepath.FromSlash(logicalPath))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("testWriteNodeRaw: mkdir %s: %v", dir, err)
	}
	path := filepath.Join(dir, "_node.md")
	if err := os.WriteFile(path, []byte(rawContent), 0o644); err != nil {
		t.Fatalf("testWriteNodeRaw: write %s: %v", path, err)
	}
}

// testLogicalToSegment converts a ROOT/... logical name to its path segment
// under code-from-spec/ for use with testWriteNode.
// e.g. "ROOT" -> ".", "ROOT/a" -> "a", "ROOT/a/b" -> "a/b"
func testLogicalToSegment(logical string) string {
	if logical == "ROOT" {
		return "."
	}
	// Strip "ROOT/" prefix
	return logical[len("ROOT/"):]
}

// testWriteLogical is a convenience wrapper that accepts a full logical name.
func testWriteLogical(t *testing.T, logicalName string, frontmatter string) {
	t.Helper()
	seg := testLogicalToSegment(logicalName)
	if seg == "." {
		// ROOT node: write to code-from-spec/_node.md
		dir := "code-from-spec"
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("testWriteLogical: mkdir %s: %v", dir, err)
		}
		var content string
		if frontmatter == "" {
			content = "---\n---\n"
		} else {
			content = "---\n" + frontmatter + "\n---\n"
		}
		path := filepath.Join(dir, "_node.md")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("testWriteLogical: write %s: %v", path, err)
		}
	} else {
		testWriteNode(t, seg, frontmatter)
	}
}

// testWriteLogicalRaw writes raw content for a logical name.
func testWriteLogicalRaw(t *testing.T, logicalName string, rawContent string) {
	t.Helper()
	seg := testLogicalToSegment(logicalName)
	if seg == "." {
		dir := "code-from-spec"
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("testWriteLogicalRaw: mkdir %s: %v", dir, err)
		}
		path := filepath.Join(dir, "_node.md")
		if err := os.WriteFile(path, []byte(rawContent), 0o644); err != nil {
			t.Fatalf("testWriteLogicalRaw: write %s: %v", path, err)
		}
	} else {
		testWriteNodeRaw(t, seg, rawContent)
	}
}

// testStringPtr returns a pointer to the given string.
func testStringPtr(s string) *string {
	return &s
}

// --------------------------------------------------------------------------
// Ancestors and target
// --------------------------------------------------------------------------

func TestChainResolve_RootAsTarget(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")

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
		t.Errorf("target logical name: got %q, want %q", chain.Target.LogicalName, "ROOT")
	}
	if chain.Target.Qualifier != nil {
		t.Errorf("target qualifier: expected absent, got %q", *chain.Target.Qualifier)
	}
	if len(chain.Dependencies) != 0 {
		t.Errorf("expected no dependencies, got %d", len(chain.Dependencies))
	}
	if len(chain.External) != 0 {
		t.Errorf("expected no external entries, got %d", len(chain.External))
	}
	if chain.Input != nil {
		t.Errorf("expected no input, got %v", chain.Input)
	}
}

func TestChainResolve_LinearChain_AncestorsInRootFirstOrder(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "")
	testWriteLogical(t, "ROOT/a/b", "")

	chain, err := chainresolver.ChainResolve("ROOT/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Ancestors) != 2 {
		t.Fatalf("expected 2 ancestors, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].LogicalName != "ROOT" {
		t.Errorf("ancestors[0]: got %q, want %q", chain.Ancestors[0].LogicalName, "ROOT")
	}
	if chain.Ancestors[1].LogicalName != "ROOT/a" {
		t.Errorf("ancestors[1]: got %q, want %q", chain.Ancestors[1].LogicalName, "ROOT/a")
	}
	if chain.Target.LogicalName != "ROOT/a/b" {
		t.Errorf("target: got %q, want %q", chain.Target.LogicalName, "ROOT/a/b")
	}
	if chain.Target.Qualifier != nil {
		t.Errorf("target qualifier: expected absent, got %q", *chain.Target.Qualifier)
	}
}

func TestChainResolve_SingleParent(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Ancestors) != 1 {
		t.Fatalf("expected 1 ancestor, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].LogicalName != "ROOT" {
		t.Errorf("ancestors[0]: got %q, want %q", chain.Ancestors[0].LogicalName, "ROOT")
	}
	if chain.Target.LogicalName != "ROOT/a" {
		t.Errorf("target: got %q, want %q", chain.Target.LogicalName, "ROOT/a")
	}
}

func TestChainResolve_TargetWithEmptyFrontmatter(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Ancestors) != 1 {
		t.Errorf("expected 1 ancestor, got %d", len(chain.Ancestors))
	}
	if chain.Target.LogicalName != "ROOT/a" {
		t.Errorf("target: got %q, want %q", chain.Target.LogicalName, "ROOT/a")
	}
	if chain.Target.Qualifier != nil {
		t.Errorf("target qualifier: expected absent, got %q", *chain.Target.Qualifier)
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

// --------------------------------------------------------------------------
// Dependencies — ROOT/ references
// --------------------------------------------------------------------------

func TestChainResolve_DependencyWithoutQualifier(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "depends_on:\n  - ROOT/b")
	testWriteLogical(t, "ROOT/b", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ROOT/b" {
		t.Errorf("dep logical name: got %q, want %q", dep.LogicalName, "ROOT/b")
	}
	if dep.Qualifier != nil {
		t.Errorf("dep qualifier: expected absent, got %q", *dep.Qualifier)
	}
}

func TestChainResolve_DependencyWithQualifier(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "depends_on:\n  - ROOT/b(interface)")
	testWriteLogical(t, "ROOT/b", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ROOT/b" {
		t.Errorf("dep logical name: got %q, want %q", dep.LogicalName, "ROOT/b")
	}
	if dep.Qualifier == nil {
		t.Fatal("dep qualifier: expected present, got nil")
	}
	if *dep.Qualifier != "interface" {
		t.Errorf("dep qualifier: got %q, want %q", *dep.Qualifier, "interface")
	}
}

func TestChainResolve_DependenciesSortedByFilePath(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "depends_on:\n  - ROOT/z\n  - ROOT/m\n  - ROOT/b")
	testWriteLogical(t, "ROOT/z", "")
	testWriteLogical(t, "ROOT/m", "")
	testWriteLogical(t, "ROOT/b", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 3 {
		t.Fatalf("expected 3 dependencies, got %d", len(chain.Dependencies))
	}

	// File paths resolve alphabetically as b < m < z
	wantOrder := []string{"ROOT/b", "ROOT/m", "ROOT/z"}
	for i, want := range wantOrder {
		if chain.Dependencies[i].LogicalName != want {
			t.Errorf("dependencies[%d]: got %q, want %q", i, chain.Dependencies[i].LogicalName, want)
		}
	}
}

// --------------------------------------------------------------------------
// Dependencies — ARTIFACT/ references
// --------------------------------------------------------------------------

func TestChainResolve_ArtifactDependencyResolvedFromGeneratingNode(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "depends_on:\n  - ARTIFACT/b(lib)")
	testWriteLogical(t, "ROOT/b", "outputs:\n  - id: lib\n    path: out/lib.go")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ARTIFACT/b(lib)" {
		t.Errorf("dep logical name: got %q, want %q", dep.LogicalName, "ARTIFACT/b(lib)")
	}
	if dep.FilePath == nil {
		t.Fatal("dep file path: expected present, got nil")
	}
	if dep.FilePath.Value != "out/lib.go" {
		t.Errorf("dep file path: got %q, want %q", dep.FilePath.Value, "out/lib.go")
	}
	if dep.Qualifier == nil {
		t.Fatal("dep qualifier: expected present, got nil")
	}
	if *dep.Qualifier != "lib" {
		t.Errorf("dep qualifier: got %q, want %q", *dep.Qualifier, "lib")
	}
}

func TestChainResolve_ArtifactWithoutQualifier_Error(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "depends_on:\n  - ARTIFACT/b")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestChainResolve_ArtifactGeneratingNodeHasNoOutputs_Error(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "depends_on:\n  - ARTIFACT/b(lib)")
	testWriteLogical(t, "ROOT/b", "")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestChainResolve_ArtifactFileNotOnDisk_NoError(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "depends_on:\n  - ARTIFACT/b(lib)")
	testWriteLogical(t, "ROOT/b", "outputs:\n  - id: lib\n    path: out/lib.go")
	// Intentionally do NOT create out/lib.go on disk.

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.FilePath == nil {
		t.Fatal("dep file path: expected present, got nil")
	}
	if dep.FilePath.Value != "out/lib.go" {
		t.Errorf("dep file path: got %q, want %q", dep.FilePath.Value, "out/lib.go")
	}
}

func TestChainResolve_ArtifactWithNonExistentOutputID_Error(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "depends_on:\n  - ARTIFACT/b(missing)")
	testWriteLogical(t, "ROOT/b", "outputs:\n  - id: lib\n    path: out/lib.go")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestChainResolve_MixedRootAndArtifactDependencies(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "depends_on:\n  - ROOT/c\n  - ARTIFACT/b(lib)")
	testWriteLogical(t, "ROOT/b", "outputs:\n  - id: lib\n    path: out/lib.go")
	testWriteLogical(t, "ROOT/c", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}

	// Find each entry by logical name (order is by file path).
	var foundC, foundArtifact bool
	for _, dep := range chain.Dependencies {
		switch dep.LogicalName {
		case "ROOT/c":
			foundC = true
		case "ARTIFACT/b(lib)":
			foundArtifact = true
		default:
			t.Errorf("unexpected dependency: %q", dep.LogicalName)
		}
	}
	if !foundC {
		t.Error("expected dependency ROOT/c, not found")
	}
	if !foundArtifact {
		t.Error("expected dependency ARTIFACT/b(lib), not found")
	}
}

// --------------------------------------------------------------------------
// Dependencies — dedup
// --------------------------------------------------------------------------

func TestChainResolve_DeduplicateExactDuplicate(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "depends_on:\n  - ROOT/b\n  - ROOT/b")
	testWriteLogical(t, "ROOT/b", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency after dedup, got %d", len(chain.Dependencies))
	}
}

func TestChainResolve_DeduplicateNoQualifierSubsumesQualified(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "depends_on:\n  - ROOT/b\n  - ROOT/b(interface)")
	testWriteLogical(t, "ROOT/b", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency after subsumption, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ROOT/b" {
		t.Errorf("dep logical name: got %q, want %q", dep.LogicalName, "ROOT/b")
	}
	if dep.Qualifier != nil {
		t.Errorf("dep qualifier: expected absent (no-qualifier wins), got %q", *dep.Qualifier)
	}
}

func TestChainResolve_DeduplicateQualifierBeforeNoQualifier_NoQualifierWins(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "depends_on:\n  - ROOT/b(interface)\n  - ROOT/b")
	testWriteLogical(t, "ROOT/b", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency after subsumption, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ROOT/b" {
		t.Errorf("dep logical name: got %q, want %q", dep.LogicalName, "ROOT/b")
	}
	if dep.Qualifier != nil {
		t.Errorf("dep qualifier: expected absent (no-qualifier always wins), got %q", *dep.Qualifier)
	}
}

func TestChainResolve_DeduplicateSameFileDifferentQualifiers_BothKept(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "depends_on:\n  - ROOT/b(interface)\n  - ROOT/b(constraints)")
	testWriteLogical(t, "ROOT/b", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies (both qualifiers kept), got %d", len(chain.Dependencies))
	}

	// Both should be present; sorted alphabetically by qualifier.
	var qualifiers []string
	for _, dep := range chain.Dependencies {
		if dep.Qualifier == nil {
			t.Error("dep qualifier: expected present, got nil")
			continue
		}
		qualifiers = append(qualifiers, *dep.Qualifier)
	}
	if len(qualifiers) == 2 {
		if qualifiers[0] != "constraints" || qualifiers[1] != "interface" {
			t.Errorf("qualifiers sorted order: got %v, want [constraints interface]", qualifiers)
		}
	}
}

func TestChainResolve_DeduplicateArtifactSameLogicalName(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "depends_on:\n  - ARTIFACT/b(lib)\n  - ARTIFACT/b(lib)")
	testWriteLogical(t, "ROOT/b", "outputs:\n  - id: lib\n    path: out/lib.go")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency after dedup, got %d", len(chain.Dependencies))
	}
}

// --------------------------------------------------------------------------
// External
// --------------------------------------------------------------------------

func TestChainResolve_ExternalEntriesCopiedFromFrontmatter(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "external:\n  - path: docs/api.yaml\n  - path: proto/v1.proto")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.External) != 2 {
		t.Fatalf("expected 2 external entries, got %d", len(chain.External))
	}
	if chain.External[0].Path != "docs/api.yaml" {
		t.Errorf("external[0].Path: got %q, want %q", chain.External[0].Path, "docs/api.yaml")
	}
	if chain.External[1].Path != "proto/v1.proto" {
		t.Errorf("external[1].Path: got %q, want %q", chain.External[1].Path, "proto/v1.proto")
	}
}

func TestChainResolve_ExternalWithFragmentsPreserved(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	fm := "external:\n  - path: f.txt\n    fragments:\n      - lines: \"1-10\"\n        hash: abc"
	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", fm)

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.External) != 1 {
		t.Fatalf("expected 1 external entry, got %d", len(chain.External))
	}
	ext := chain.External[0]
	if ext.Path != "f.txt" {
		t.Errorf("external[0].Path: got %q, want %q", ext.Path, "f.txt")
	}
	if len(ext.Fragments) != 1 {
		t.Fatalf("expected 1 fragment, got %d", len(ext.Fragments))
	}
	frag := ext.Fragments[0]
	if frag.Lines != "1-10" {
		t.Errorf("fragment.Lines: got %q, want %q", frag.Lines, "1-10")
	}
	if frag.Hash != "abc" {
		t.Errorf("fragment.Hash: got %q, want %q", frag.Hash, "abc")
	}
}

func TestChainResolve_EmptyExternal(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.External) != 0 {
		t.Errorf("expected no external entries, got %d", len(chain.External))
	}
}

// --------------------------------------------------------------------------
// Input
// --------------------------------------------------------------------------

func TestChainResolve_InputResolvedFromGeneratingNode(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "input: ARTIFACT/b(data)")
	testWriteLogical(t, "ROOT/b", "outputs:\n  - id: data\n    path: out/data.json")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chain.Input == nil {
		t.Fatal("expected input, got nil")
	}
	if chain.Input.LogicalName != "ARTIFACT/b(data)" {
		t.Errorf("input logical name: got %q, want %q", chain.Input.LogicalName, "ARTIFACT/b(data)")
	}
	if chain.Input.FilePath == nil {
		t.Fatal("input file path: expected present, got nil")
	}
	if chain.Input.FilePath.Value != "out/data.json" {
		t.Errorf("input file path: got %q, want %q", chain.Input.FilePath.Value, "out/data.json")
	}
	if chain.Input.Qualifier == nil {
		t.Fatal("input qualifier: expected present, got nil")
	}
	if *chain.Input.Qualifier != "data" {
		t.Errorf("input qualifier: got %q, want %q", *chain.Input.Qualifier, "data")
	}
}

func TestChainResolve_NoInput_Absent(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "")

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chain.Input != nil {
		t.Errorf("expected no input, got %v", chain.Input)
	}
}

func TestChainResolve_InputWithoutQualifier_Error(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "input: ARTIFACT/b")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestChainResolve_InputWithNonExistentOutputID_Error(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "input: ARTIFACT/b(missing)")
	testWriteLogical(t, "ROOT/b", "outputs:\n  - id: data\n    path: out/data.json")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

// --------------------------------------------------------------------------
// Error cases
// --------------------------------------------------------------------------

func TestChainResolve_UnrecognizedPrefixInDependsOn_Error(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	testWriteLogical(t, "ROOT/a", "depends_on:\n  - UNKNOWN/something")

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
	// The exact error is determined by LogicalNameGetParent / LogicalNameToPath;
	// we only verify that an error is returned.
}

func TestChainResolve_UnreadableFrontmatter_Error(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteLogical(t, "ROOT", "")
	// Write invalid YAML between frontmatter delimiters.
	testWriteLogicalRaw(t, "ROOT/a", "---\n: invalid: yaml: [\nunclosed\n---\n")

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnreadableFrontmatter) {
		t.Errorf("expected ErrUnreadableFrontmatter, got %v", err)
	}
}
