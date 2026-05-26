// code-from-spec: ROOT/golang/internal/node_ranking/tests@6sQQEQ23Vu31ebNHl_I7s71tPFg

// Package noderanking contains tests for the DetectCycles ranking algorithm.
//
// Test strategy:
//   - Each test creates real _node.md files under t.TempDir() so that
//     frontmatter.ParseFrontmatter can read them without touching the actual
//     code-from-spec/ tree.
//   - nodediscovery.DiscoveredNode slices are built manually (DetectCycles
//     accepts them directly, so no DiscoverNodes call is needed).
//   - All helper types and functions are prefixed with "test" per convention.
package noderanking

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/nodediscovery"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// testNodeFile writes a _node.md file at the given absolute path with the
// supplied frontmatter content. If frontmatter is empty string, the file is
// written with no frontmatter delimiters (so ParseFrontmatter returns an empty
// Frontmatter, which is fine — no deps, no outputs).
func testNodeFile(t *testing.T, path string, fmYAML string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("testNodeFile: MkdirAll: %v", err)
	}

	var content string
	if fmYAML != "" {
		content = "---\n" + fmYAML + "\n---\n"
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("testNodeFile: WriteFile: %v", err)
	}
}

// testMakeNode returns a DiscoveredNode with the given logical name and a
// _node.md file created under dir.
//
// The file path mirrors the logical name hierarchy so that the tests remain
// easy to read: ROOT → root, ROOT/a → root/a, etc.
func testMakeNode(t *testing.T, dir string, logicalName string, fmYAML string) nodediscovery.DiscoveredNode {
	t.Helper()

	// Derive a sub-path from the logical name for file placement.
	// e.g. "ROOT/a/b" → "<dir>/ROOT/a/b/_node.md"
	filePath := filepath.Join(dir, filepath.FromSlash(logicalName), "_node.md")
	testNodeFile(t, filePath, fmYAML)

	return nodediscovery.DiscoveredNode{
		LogicalName: logicalName,
		FilePath:    filePath,
	}
}

// testFindRank returns the rank of the entry with the given logical name from
// a slice of RankedEntry. It fails the test if the name is not found.
func testFindRank(t *testing.T, entries []RankedEntry, logicalName string) int {
	t.Helper()
	for _, e := range entries {
		if e.LogicalName == logicalName {
			return e.Rank
		}
	}
	t.Fatalf("testFindRank: entry %q not found in ranked entries", logicalName)
	return -1
}

// testContains returns true when slice contains target.
func testContains(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Happy-path tests
// ---------------------------------------------------------------------------

// TestLinearChainHasIncrementingRanks verifies that a simple parent→child→
// grandchild hierarchy gets ranks 0, 1, 2 respectively.
//
// Spec: "Create three nodes: ROOT, ROOT/a, ROOT/a/b (parent chain).
// Expect ranks 0, 1, 2 respectively. No cycle participants."
func TestLinearChainHasIncrementingRanks(t *testing.T) {
	dir := t.TempDir()

	// ROOT has no parent — expect rank 0.
	// ROOT/a's parent is ROOT — expect rank 1.
	// ROOT/a/b's parent is ROOT/a — expect rank 2.
	nodes := []nodediscovery.DiscoveredNode{
		testMakeNode(t, dir, "ROOT", ""),
		testMakeNode(t, dir, "ROOT/a", ""),
		testMakeNode(t, dir, "ROOT/a/b", ""),
	}

	ranked, cycles, err := DetectCycles(nodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycle participants, got %v", cycles)
	}

	// Verify each rank individually.
	tests := []struct {
		name         string
		expectedRank int
	}{
		{"ROOT", 0},
		{"ROOT/a", 1},
		{"ROOT/a/b", 2},
	}
	for _, tc := range tests {
		got := testFindRank(t, ranked, tc.name)
		if got != tc.expectedRank {
			t.Errorf("node %q: want rank %d, got %d", tc.name, tc.expectedRank, got)
		}
	}
}

// TestIndependentSiblingsHaveEqualRank verifies that two children of ROOT
// with no cross-dependencies receive identical ranks.
//
// Spec: "Create ROOT and two children ROOT/a and ROOT/b with no
// cross-dependencies. Expect ROOT/a and ROOT/b have the same rank. No cycle
// participants."
func TestIndependentSiblingsHaveEqualRank(t *testing.T) {
	dir := t.TempDir()

	nodes := []nodediscovery.DiscoveredNode{
		testMakeNode(t, dir, "ROOT", ""),
		testMakeNode(t, dir, "ROOT/a", ""),
		testMakeNode(t, dir, "ROOT/b", ""),
	}

	ranked, cycles, err := DetectCycles(nodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycle participants, got %v", cycles)
	}

	rankA := testFindRank(t, ranked, "ROOT/a")
	rankB := testFindRank(t, ranked, "ROOT/b")
	if rankA != rankB {
		t.Errorf("sibling ranks differ: ROOT/a=%d, ROOT/b=%d", rankA, rankB)
	}
}

// TestDependsOnIncreasesRank verifies that an explicit depends_on edge pushes
// the dependent node's rank above its dependency.
//
// Spec: "Create ROOT, ROOT/a, ROOT/b where ROOT/b depends_on ROOT/a.
// Expect ROOT/b has higher rank than ROOT/a. No cycle participants."
func TestDependsOnIncreasesRank(t *testing.T) {
	dir := t.TempDir()

	// ROOT/b explicitly depends on ROOT/a via the depends_on frontmatter key.
	nodes := []nodediscovery.DiscoveredNode{
		testMakeNode(t, dir, "ROOT", ""),
		testMakeNode(t, dir, "ROOT/a", ""),
		testMakeNode(t, dir, "ROOT/b", "depends_on:\n  - ROOT/a\n"),
	}

	ranked, cycles, err := DetectCycles(nodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycle participants, got %v", cycles)
	}

	rankA := testFindRank(t, ranked, "ROOT/a")
	rankB := testFindRank(t, ranked, "ROOT/b")
	if rankB <= rankA {
		t.Errorf("expected ROOT/b rank (%d) > ROOT/a rank (%d)", rankB, rankA)
	}
}

// TestArtifactGetsRankOneAboveNode verifies that an artifact output is
// assigned a rank exactly one above its owning node.
//
// Spec: "Create ROOT/a with an output artifact. Expect the artifact entry has
// rank = rank of ROOT/a + 1."
func TestArtifactGetsRankOneAboveNode(t *testing.T) {
	dir := t.TempDir()

	// ROOT/a declares one output artifact with id "impl".
	// DetectCycles creates an artifact entry for it automatically.
	nodes := []nodediscovery.DiscoveredNode{
		testMakeNode(t, dir, "ROOT", ""),
		testMakeNode(t, dir, "ROOT/a", "outputs:\n  - id: impl\n    path: some/path.go\n"),
	}

	ranked, cycles, err := DetectCycles(nodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycle participants, got %v", cycles)
	}

	// The artifact logical name is constructed as "ARTIFACT/a(impl)" because
	// the algorithm strips the "ROOT/" prefix from the node name.
	// See implementation: artifactKey = "ARTIFACT/" + nodePathWithoutRoot + "(" + out.ID + ")"
	nodeRank := testFindRank(t, ranked, "ROOT/a")
	artifactKey := "ARTIFACT/a(impl)"
	artifactRank := testFindRank(t, ranked, artifactKey)

	if artifactRank != nodeRank+1 {
		t.Errorf("artifact %q: want rank %d (node rank %d + 1), got %d",
			artifactKey, nodeRank+1, nodeRank, artifactRank)
	}
}

// ---------------------------------------------------------------------------
// Failure-case tests
// ---------------------------------------------------------------------------

// TestCircularDependencyDetected verifies that a mutual dependency between two
// nodes is reported as a cycle.
//
// Spec: "Create ROOT/a depends_on ROOT/b and ROOT/b depends_on ROOT/a.
// Expect both logical names appear in the cycle participants list."
func TestCircularDependencyDetected(t *testing.T) {
	dir := t.TempDir()

	// Both ROOT/a and ROOT/b depend on each other — classic two-node cycle.
	nodes := []nodediscovery.DiscoveredNode{
		testMakeNode(t, dir, "ROOT", ""),
		testMakeNode(t, dir, "ROOT/a", "depends_on:\n  - ROOT/b\n"),
		testMakeNode(t, dir, "ROOT/b", "depends_on:\n  - ROOT/a\n"),
	}

	_, cycles, err := DetectCycles(nodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both participants must appear in the cycle slice.
	if !testContains(cycles, "ROOT/a") {
		t.Errorf("expected ROOT/a in cycle participants, got %v", cycles)
	}
	if !testContains(cycles, "ROOT/b") {
		t.Errorf("expected ROOT/b in cycle participants, got %v", cycles)
	}
}

// TestUnresolvableReference verifies that a depends_on pointing to an unknown
// node returns ErrUnresolvableRef.
//
// Spec: "Create a node with depends_on pointing to a non-existent node.
// Expect errors.Is(err, ErrUnresolvableRef)."
func TestUnresolvableReference(t *testing.T) {
	dir := t.TempDir()

	// ROOT/a references "ROOT/nonexistent" which is not in the node list.
	nodes := []nodediscovery.DiscoveredNode{
		testMakeNode(t, dir, "ROOT", ""),
		testMakeNode(t, dir, "ROOT/a", "depends_on:\n  - ROOT/nonexistent\n"),
	}

	_, _, err := DetectCycles(nodes)
	if err == nil {
		t.Fatal("expected an error for unresolvable reference, got nil")
	}
	if !errors.Is(err, ErrUnresolvableRef) {
		t.Errorf("expected errors.Is(err, ErrUnresolvableRef), got: %v", err)
	}
}
