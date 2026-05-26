// code-from-spec: ROOT/golang/internal/node_ranking/tests@3NFi7tpxAVTgsL5oIJN9qMGu1_A

// Package noderanking tests the DetectCycles function which ranks discovered
// nodes by their dependency depth and detects circular dependencies.
package noderanking

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/nodediscovery"
)

// ----------------------------------------------------------------------------
// Test helpers
// ----------------------------------------------------------------------------

// testNodeFile creates a _node.md file at <dir>/<relPath>/_node.md with the
// given frontmatter content. It also creates all necessary parent directories.
func testNodeFile(t *testing.T, dir, relPath, frontmatter string) {
	t.Helper()

	nodeDir := filepath.Join(dir, filepath.FromSlash(relPath))
	if err := os.MkdirAll(nodeDir, 0o755); err != nil {
		t.Fatalf("testNodeFile: MkdirAll(%q): %v", nodeDir, err)
	}

	content := "---\n" + frontmatter + "---\n"
	nodeFile := filepath.Join(nodeDir, "_node.md")
	if err := os.WriteFile(nodeFile, []byte(content), 0o644); err != nil {
		t.Fatalf("testNodeFile: WriteFile(%q): %v", nodeFile, err)
	}
}

// testRankOf returns the rank for a given logicalName from a slice of
// RankedEntry values. It fails the test if the name is not found.
func testRankOf(t *testing.T, entries []RankedEntry, logicalName string) int {
	t.Helper()
	for _, e := range entries {
		if e.LogicalName == logicalName {
			return e.Rank
		}
	}
	t.Fatalf("testRankOf: %q not found in ranked entries", logicalName)
	return -1
}

// testContains returns true if the given string slice contains the value.
func testContains(slice []string, value string) bool {
	for _, s := range slice {
		if s == value {
			return true
		}
	}
	return false
}

// testSortedNames returns a sorted copy of the logical names from RankedEntry
// entries, for stable comparisons.
func testSortedNames(entries []RankedEntry) []string {
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.LogicalName
	}
	sort.Strings(names)
	return names
}

// ----------------------------------------------------------------------------
// Happy path tests
// ----------------------------------------------------------------------------

// TestLinearChainHasIncrementingRanks verifies that a simple parent-child-
// grandchild chain produces strictly increasing ranks: ROOT < ROOT/a < ROOT/a/b.
func TestLinearChainHasIncrementingRanks(t *testing.T) {
	dir := t.TempDir()

	// ROOT node — no dependencies, no parent
	testNodeFile(t, dir, "ROOT", "")

	// ROOT/a — child of ROOT (parent relationship, no explicit depends_on)
	testNodeFile(t, dir, "ROOT/a", "")

	// ROOT/a/b — child of ROOT/a
	testNodeFile(t, dir, "ROOT/a/b", "")

	nodes, err := nodediscovery.DiscoverNodes(dir)
	if err != nil {
		t.Fatalf("DiscoverNodes: %v", err)
	}

	entries, cycleParticipants, err := DetectCycles(nodes)
	if err != nil {
		t.Fatalf("DetectCycles: unexpected error: %v", err)
	}

	if len(cycleParticipants) != 0 {
		t.Errorf("expected no cycle participants, got %v", cycleParticipants)
	}

	rankROOT := testRankOf(t, entries, "ROOT")
	rankA := testRankOf(t, entries, "ROOT/a")
	rankAB := testRankOf(t, entries, "ROOT/a/b")

	// Each level must be strictly greater than the one above it.
	if !(rankROOT < rankA) {
		t.Errorf("expected rank(ROOT) < rank(ROOT/a), got %d >= %d", rankROOT, rankA)
	}
	if !(rankA < rankAB) {
		t.Errorf("expected rank(ROOT/a) < rank(ROOT/a/b), got %d >= %d", rankA, rankAB)
	}
}

// TestIndependentSiblingsHaveEqualRank verifies that two sibling nodes with
// the same parent and no cross-dependencies receive the same rank.
func TestIndependentSiblingsHaveEqualRank(t *testing.T) {
	dir := t.TempDir()

	// ROOT — root node
	testNodeFile(t, dir, "ROOT", "")

	// ROOT/a and ROOT/b — independent siblings
	testNodeFile(t, dir, "ROOT/a", "")
	testNodeFile(t, dir, "ROOT/b", "")

	nodes, err := nodediscovery.DiscoverNodes(dir)
	if err != nil {
		t.Fatalf("DiscoverNodes: %v", err)
	}

	entries, cycleParticipants, err := DetectCycles(nodes)
	if err != nil {
		t.Fatalf("DetectCycles: unexpected error: %v", err)
	}

	if len(cycleParticipants) != 0 {
		t.Errorf("expected no cycle participants, got %v", cycleParticipants)
	}

	rankA := testRankOf(t, entries, "ROOT/a")
	rankB := testRankOf(t, entries, "ROOT/b")

	if rankA != rankB {
		t.Errorf("expected ROOT/a and ROOT/b to have equal rank, got %d and %d", rankA, rankB)
	}
}

// TestDependsOnIncreasesRank verifies that when ROOT/b explicitly depends_on
// ROOT/a, ROOT/b receives a strictly higher rank than ROOT/a.
func TestDependsOnIncreasesRank(t *testing.T) {
	dir := t.TempDir()

	// ROOT — root node
	testNodeFile(t, dir, "ROOT", "")

	// ROOT/a — no extra dependencies
	testNodeFile(t, dir, "ROOT/a", "")

	// ROOT/b — explicitly depends on ROOT/a
	testNodeFile(t, dir, "ROOT/b", "depends_on:\n  - ROOT/a\n")

	nodes, err := nodediscovery.DiscoverNodes(dir)
	if err != nil {
		t.Fatalf("DiscoverNodes: %v", err)
	}

	entries, cycleParticipants, err := DetectCycles(nodes)
	if err != nil {
		t.Fatalf("DetectCycles: unexpected error: %v", err)
	}

	if len(cycleParticipants) != 0 {
		t.Errorf("expected no cycle participants, got %v", cycleParticipants)
	}

	rankA := testRankOf(t, entries, "ROOT/a")
	rankB := testRankOf(t, entries, "ROOT/b")

	if !(rankA < rankB) {
		t.Errorf("expected rank(ROOT/a) < rank(ROOT/b), got %d >= %d", rankA, rankB)
	}
}

// TestArtifactGetsRankOneAboveNode verifies that an output artifact entry
// receives a rank exactly one greater than its owning node.
func TestArtifactGetsRankOneAboveNode(t *testing.T) {
	dir := t.TempDir()

	// ROOT/a — node with a declared output artifact
	testNodeFile(t, dir, "ROOT/a",
		"outputs:\n  - id: out1\n    path: internal/foo/foo.go\n")

	nodes, err := nodediscovery.DiscoverNodes(dir)
	if err != nil {
		t.Fatalf("DiscoverNodes: %v", err)
	}

	entries, cycleParticipants, err := DetectCycles(nodes)
	if err != nil {
		t.Fatalf("DetectCycles: unexpected error: %v", err)
	}

	if len(cycleParticipants) != 0 {
		t.Errorf("expected no cycle participants, got %v", cycleParticipants)
	}

	// The artifact logical name is formed as "<node>/<output-id>".
	nodeRank := testRankOf(t, entries, "ROOT/a")
	artifactRank := testRankOf(t, entries, "ROOT/a/out1")

	if artifactRank != nodeRank+1 {
		t.Errorf("expected artifact rank = node rank + 1, got artifact=%d node=%d", artifactRank, nodeRank)
	}
}

// ----------------------------------------------------------------------------
// Failure case tests
// ----------------------------------------------------------------------------

// TestCircularDependencyDetected verifies that a two-node cycle (ROOT/a
// depends_on ROOT/b and ROOT/b depends_on ROOT/a) causes both names to appear
// in the returned cycle participants list.
func TestCircularDependencyDetected(t *testing.T) {
	dir := t.TempDir()

	// ROOT/a depends on ROOT/b
	testNodeFile(t, dir, "ROOT/a", "depends_on:\n  - ROOT/b\n")

	// ROOT/b depends on ROOT/a — creates a cycle
	testNodeFile(t, dir, "ROOT/b", "depends_on:\n  - ROOT/a\n")

	nodes, err := nodediscovery.DiscoverNodes(dir)
	if err != nil {
		t.Fatalf("DiscoverNodes: %v", err)
	}

	// DetectCycles may or may not return a non-nil error for cycles; the key
	// contract is that both names appear in cycleParticipants.
	_, cycleParticipants, _ := DetectCycles(nodes)

	if !testContains(cycleParticipants, "ROOT/a") {
		t.Errorf("expected ROOT/a in cycle participants, got %v", cycleParticipants)
	}
	if !testContains(cycleParticipants, "ROOT/b") {
		t.Errorf("expected ROOT/b in cycle participants, got %v", cycleParticipants)
	}
}

// TestUnresolvableReference verifies that a node with a depends_on target
// that does not exist causes DetectCycles to return ErrUnresolvableRef.
func TestUnresolvableReference(t *testing.T) {
	dir := t.TempDir()

	// ROOT/a depends on a node that will never exist
	testNodeFile(t, dir, "ROOT/a", "depends_on:\n  - ROOT/nonexistent\n")

	nodes, err := nodediscovery.DiscoverNodes(dir)
	if err != nil {
		t.Fatalf("DiscoverNodes: %v", err)
	}

	_, _, err = DetectCycles(nodes)
	if !errors.Is(err, ErrUnresolvableRef) {
		t.Errorf("expected ErrUnresolvableRef, got %v", err)
	}
}
