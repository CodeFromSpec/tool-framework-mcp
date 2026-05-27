// code-from-spec: ROOT/golang/internal/node_ranking/tests@7nlrcu337oQ2mf9Oi_bMY9qP4sk

// Package noderanking — test file.
//
// These tests exercise DetectCycles using real _node.md files written to a
// temporary directory. This is required because DetectCycles parses frontmatter
// internally from FilePath; there is no way to inject pre-parsed frontmatter.
//
// All helper functions and types are prefixed with "test" to avoid name
// collisions with unexported identifiers in the package under test (this file
// is in the same package).
package noderanking

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/nodediscovery"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

// testWriteNodeFile creates a _node.md file at baseDir/<subpath>/_node.md with
// the supplied YAML frontmatter content between "---" delimiters.
//
// frontmatterYAML should be the raw YAML body without the "---" delimiters,
// e.g. "depends_on:\n  - ROOT/a\n".
//
// Returns the absolute path to the created file.
func testWriteNodeFile(t *testing.T, baseDir, subpath, frontmatterYAML string) string {
	t.Helper()

	// Build the directory for this node.
	dir := filepath.Join(baseDir, filepath.FromSlash(subpath))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("testWriteNodeFile: MkdirAll(%q): %v", dir, err)
	}

	filePath := filepath.Join(dir, "_node.md")

	var content string
	if frontmatterYAML == "" {
		// No frontmatter — write an empty file so ParseFrontmatter returns an
		// empty Frontmatter struct (not an error).
		content = ""
	} else {
		content = fmt.Sprintf("---\n%s---\n", frontmatterYAML)
	}

	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("testWriteNodeFile: WriteFile(%q): %v", filePath, err)
	}
	return filePath
}

// testFindRank finds the rank for a given logicalName in the ranked entries.
// Fails the test if the name is not found.
func testFindRank(t *testing.T, entries []RankedEntry, logicalName string) int {
	t.Helper()
	for _, e := range entries {
		if e.LogicalName == logicalName {
			return e.Rank
		}
	}
	t.Fatalf("testFindRank: %q not found in ranked entries %v", logicalName, entries)
	return -1 // unreachable
}

// testHasEntry reports whether a logical name appears in the ranked entries.
func testHasEntry(entries []RankedEntry, logicalName string) bool {
	for _, e := range entries {
		if e.LogicalName == logicalName {
			return true
		}
	}
	return false
}

// ─── happy-path tests ─────────────────────────────────────────────────────────

// TestDetectCycles_LinearChainRanks verifies that a parent-child chain of three
// nodes receives incrementing ranks (0, 1, 2) and no cycles are detected.
//
// Nodes:
//
//	ROOT           (no depends_on, no parent)
//	ROOT/a         (parent = ROOT)
//	ROOT/a/b       (parent = ROOT/a)
func TestDetectCycles_LinearChainRanks(t *testing.T) {
	dir := t.TempDir()

	rootPath := testWriteNodeFile(t, dir, "ROOT", "")
	aPath := testWriteNodeFile(t, dir, "ROOT/a", "")
	abPath := testWriteNodeFile(t, dir, "ROOT/a/b", "")

	nodes := []nodediscovery.DiscoveredNode{
		{LogicalName: "ROOT", FilePath: rootPath},
		{LogicalName: "ROOT/a", FilePath: aPath},
		{LogicalName: "ROOT/a/b", FilePath: abPath},
	}

	entries, cycleNames, err := DetectCycles(nodes)
	if err != nil {
		t.Fatalf("DetectCycles returned unexpected error: %v", err)
	}
	if len(cycleNames) != 0 {
		t.Errorf("expected no cycle participants, got %v", cycleNames)
	}

	rankRoot := testFindRank(t, entries, "ROOT")
	rankA := testFindRank(t, entries, "ROOT/a")
	rankAB := testFindRank(t, entries, "ROOT/a/b")

	if rankRoot != 0 {
		t.Errorf("ROOT: expected rank 0, got %d", rankRoot)
	}
	if rankA != 1 {
		t.Errorf("ROOT/a: expected rank 1, got %d", rankA)
	}
	if rankAB != 2 {
		t.Errorf("ROOT/a/b: expected rank 2, got %d", rankAB)
	}
}

// TestDetectCycles_IndependentSiblingsEqualRank verifies that two sibling nodes
// with the same parent but no cross-dependency receive the same rank, and no
// cycles are detected.
//
// Nodes:
//
//	ROOT      (rank 0)
//	ROOT/a    (rank 1 — depends only on ROOT via parent edge)
//	ROOT/b    (rank 1 — same, no cross-dependency)
func TestDetectCycles_IndependentSiblingsEqualRank(t *testing.T) {
	dir := t.TempDir()

	rootPath := testWriteNodeFile(t, dir, "ROOT", "")
	aPath := testWriteNodeFile(t, dir, "ROOT/a", "")
	bPath := testWriteNodeFile(t, dir, "ROOT/b", "")

	nodes := []nodediscovery.DiscoveredNode{
		{LogicalName: "ROOT", FilePath: rootPath},
		{LogicalName: "ROOT/a", FilePath: aPath},
		{LogicalName: "ROOT/b", FilePath: bPath},
	}

	entries, cycleNames, err := DetectCycles(nodes)
	if err != nil {
		t.Fatalf("DetectCycles returned unexpected error: %v", err)
	}
	if len(cycleNames) != 0 {
		t.Errorf("expected no cycle participants, got %v", cycleNames)
	}

	rankA := testFindRank(t, entries, "ROOT/a")
	rankB := testFindRank(t, entries, "ROOT/b")

	if rankA != rankB {
		t.Errorf("expected ROOT/a and ROOT/b to have the same rank, got %d and %d", rankA, rankB)
	}
}

// TestDetectCycles_DependsOnIncreasesRank verifies that a depends_on reference
// from ROOT/b to ROOT/a forces ROOT/b to have a strictly higher rank than
// ROOT/a, and no cycles are detected.
//
// Nodes:
//
//	ROOT      (rank 0)
//	ROOT/a    (rank 1 — parent = ROOT)
//	ROOT/b    (rank > ROOT/a — depends_on ROOT/a)
func TestDetectCycles_DependsOnIncreasesRank(t *testing.T) {
	dir := t.TempDir()

	rootPath := testWriteNodeFile(t, dir, "ROOT", "")
	aPath := testWriteNodeFile(t, dir, "ROOT/a", "")
	// ROOT/b explicitly depends on ROOT/a via depends_on.
	bPath := testWriteNodeFile(t, dir, "ROOT/b", "depends_on:\n  - ROOT/a\n")

	nodes := []nodediscovery.DiscoveredNode{
		{LogicalName: "ROOT", FilePath: rootPath},
		{LogicalName: "ROOT/a", FilePath: aPath},
		{LogicalName: "ROOT/b", FilePath: bPath},
	}

	entries, cycleNames, err := DetectCycles(nodes)
	if err != nil {
		t.Fatalf("DetectCycles returned unexpected error: %v", err)
	}
	if len(cycleNames) != 0 {
		t.Errorf("expected no cycle participants, got %v", cycleNames)
	}

	rankA := testFindRank(t, entries, "ROOT/a")
	rankB := testFindRank(t, entries, "ROOT/b")

	if rankB <= rankA {
		t.Errorf("ROOT/b (rank %d) should be strictly greater than ROOT/a (rank %d)", rankB, rankA)
	}
}

// TestDetectCycles_QualifiedDependsOnResolvesCorrectly verifies that a
// qualified depends_on reference (e.g. "ROOT/a(interface)") is stripped of
// its qualifier before lookup and resolves cleanly to ROOT/a.
//
// Nodes:
//
//	ROOT           (rank 0)
//	ROOT/a         (contains a # Public / ## Interface section conceptually; rank 1)
//	ROOT/b         (depends_on: ROOT/a(interface); rank > ROOT/a)
func TestDetectCycles_QualifiedDependsOnResolvesCorrectly(t *testing.T) {
	dir := t.TempDir()

	rootPath := testWriteNodeFile(t, dir, "ROOT", "")
	// ROOT/a has no depends_on; its frontmatter is empty.
	aPath := testWriteNodeFile(t, dir, "ROOT/a", "")
	// ROOT/b uses a qualified reference — normalizeRef must strip "(interface)".
	bPath := testWriteNodeFile(t, dir, "ROOT/b", "depends_on:\n  - ROOT/a(interface)\n")

	nodes := []nodediscovery.DiscoveredNode{
		{LogicalName: "ROOT", FilePath: rootPath},
		{LogicalName: "ROOT/a", FilePath: aPath},
		{LogicalName: "ROOT/b", FilePath: bPath},
	}

	entries, cycleNames, err := DetectCycles(nodes)
	if err != nil {
		// A qualified reference should NOT produce ErrUnresolvableRef.
		t.Fatalf("DetectCycles returned unexpected error: %v", err)
	}
	if len(cycleNames) != 0 {
		t.Errorf("expected no cycle participants, got %v", cycleNames)
	}

	rankA := testFindRank(t, entries, "ROOT/a")
	rankB := testFindRank(t, entries, "ROOT/b")

	if rankB <= rankA {
		t.Errorf("ROOT/b (rank %d) should be strictly greater than ROOT/a (rank %d)", rankB, rankA)
	}
}

// TestDetectCycles_ArtifactRankOneAboveNode verifies that an artifact entry
// produced by a node receives rank = node rank + 1.
//
// Nodes:
//
//	ROOT      (rank 0)
//	ROOT/a    (rank 1; declares one output with id "code" and path "out/a.go")
//
// The artifact key built by buildEntryMap for ROOT/a with output id "code" is
// "ARTIFACT/a(code)". The artifact's rank should be 2.
func TestDetectCycles_ArtifactRankOneAboveNode(t *testing.T) {
	dir := t.TempDir()

	rootPath := testWriteNodeFile(t, dir, "ROOT", "")
	// ROOT/a declares one output artifact.
	aFM := "outputs:\n  - id: code\n    path: out/a.go\n"
	aPath := testWriteNodeFile(t, dir, "ROOT/a", aFM)

	nodes := []nodediscovery.DiscoveredNode{
		{LogicalName: "ROOT", FilePath: rootPath},
		{LogicalName: "ROOT/a", FilePath: aPath},
	}

	entries, cycleNames, err := DetectCycles(nodes)
	if err != nil {
		t.Fatalf("DetectCycles returned unexpected error: %v", err)
	}
	if len(cycleNames) != 0 {
		t.Errorf("expected no cycle participants, got %v", cycleNames)
	}

	// The artifact key is constructed as "ARTIFACT/a(code)" by buildEntryMap.
	artifactKey := "ARTIFACT/a(code)"

	if !testHasEntry(entries, artifactKey) {
		// Surface all returned entry names to help diagnose mismatches.
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.LogicalName
		}
		t.Fatalf("artifact key %q not found in entries; got: %v", artifactKey, names)
	}

	rankNode := testFindRank(t, entries, "ROOT/a")
	rankArtifact := testFindRank(t, entries, artifactKey)

	expectedArtifactRank := rankNode + 1
	if rankArtifact != expectedArtifactRank {
		t.Errorf("artifact %q: expected rank %d (node rank %d + 1), got %d",
			artifactKey, expectedArtifactRank, rankNode, rankArtifact)
	}
}

// ─── failure-case tests ───────────────────────────────────────────────────────

// TestDetectCycles_CircularDependencyDetected verifies that a mutual dependency
// between ROOT/a and ROOT/b results in a non-empty cycle-participants list.
//
// Nodes:
//
//	ROOT    (no dependencies)
//	ROOT/a  (depends_on: ROOT/b)
//	ROOT/b  (depends_on: ROOT/a)
//
// The test only asserts that the cycle-participants list is NOT empty; it does
// not check for specific members, because the exact set reported depends on the
// internal Bellman-Ford convergence order.
func TestDetectCycles_CircularDependencyDetected(t *testing.T) {
	dir := t.TempDir()

	rootPath := testWriteNodeFile(t, dir, "ROOT", "")
	aPath := testWriteNodeFile(t, dir, "ROOT/a", "depends_on:\n  - ROOT/b\n")
	bPath := testWriteNodeFile(t, dir, "ROOT/b", "depends_on:\n  - ROOT/a\n")

	nodes := []nodediscovery.DiscoveredNode{
		{LogicalName: "ROOT", FilePath: rootPath},
		{LogicalName: "ROOT/a", FilePath: aPath},
		{LogicalName: "ROOT/b", FilePath: bPath},
	}

	_, cycleNames, err := DetectCycles(nodes)
	if err != nil {
		// A cycle should not produce an error — only a non-empty cycleNames slice.
		t.Fatalf("DetectCycles returned unexpected error: %v", err)
	}

	// The only assertion: at least one participant was identified.
	if len(cycleNames) == 0 {
		t.Errorf("expected at least one cycle participant, got an empty slice")
	}
}

// TestDetectCycles_UnresolvableReference verifies that a depends_on reference
// pointing to a non-existent node causes DetectCycles to return an error
// wrapping ErrUnresolvableRef.
//
// Nodes:
//
//	ROOT      (rank 0)
//	ROOT/a    (depends_on: ROOT/does_not_exist)
func TestDetectCycles_UnresolvableReference(t *testing.T) {
	dir := t.TempDir()

	rootPath := testWriteNodeFile(t, dir, "ROOT", "")
	aPath := testWriteNodeFile(t, dir, "ROOT/a", "depends_on:\n  - ROOT/does_not_exist\n")

	nodes := []nodediscovery.DiscoveredNode{
		{LogicalName: "ROOT", FilePath: rootPath},
		{LogicalName: "ROOT/a", FilePath: aPath},
	}

	_, _, err := DetectCycles(nodes)
	if err == nil {
		t.Fatal("expected an error wrapping ErrUnresolvableRef, got nil")
	}
	if !errors.Is(err, ErrUnresolvableRef) {
		t.Errorf("expected error to wrap ErrUnresolvableRef; got: %v", err)
	}
}
