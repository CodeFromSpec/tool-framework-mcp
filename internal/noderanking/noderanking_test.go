// code-from-spec: ROOT/golang/tests/utils/node_ranking@WbNaIxXveJj_uLhlS-P3g1qTthk
package noderanking_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/noderanking"
)

func testFindEntry(ranked []*noderanking.NodeRankEntry, logicalName string) *noderanking.NodeRankEntry {
	for _, e := range ranked {
		if e.LogicalName == logicalName {
			return e
		}
	}
	return nil
}

func testInCycles(cycles []string, logicalName string) bool {
	for _, c := range cycles {
		if c == logicalName {
			return true
		}
	}
	return false
}

func TestNodeRankCompute_RootOnly(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
	entry := testFindEntry(ranked, "ROOT")
	if entry == nil {
		t.Fatal("expected entry for ROOT")
	}
	if entry.Rank != 0 {
		t.Errorf("expected rank 0 for ROOT, got %d", entry.Rank)
	}
}

func TestNodeRankCompute_LinearChain(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a/b", Frontmatter: &frontmatter.Frontmatter{}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	root := testFindEntry(ranked, "ROOT")
	a := testFindEntry(ranked, "ROOT/a")
	ab := testFindEntry(ranked, "ROOT/a/b")

	if root == nil || a == nil || ab == nil {
		t.Fatal("missing expected entries")
	}
	if root.Rank != 0 {
		t.Errorf("expected ROOT rank 0, got %d", root.Rank)
	}
	if a.Rank != 1 {
		t.Errorf("expected ROOT/a rank 1, got %d", a.Rank)
	}
	if ab.Rank != 2 {
		t.Errorf("expected ROOT/a/b rank 2, got %d", ab.Rank)
	}
}

func TestNodeRankCompute_IndependentSiblings(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/b", Frontmatter: &frontmatter.Frontmatter{}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	a := testFindEntry(ranked, "ROOT/a")
	b := testFindEntry(ranked, "ROOT/b")
	if a == nil || b == nil {
		t.Fatal("missing expected entries")
	}
	if a.Rank != 1 {
		t.Errorf("expected ROOT/a rank 1, got %d", a.Rank)
	}
	if b.Rank != 1 {
		t.Errorf("expected ROOT/b rank 1, got %d", b.Rank)
	}
}

func TestNodeRankCompute_DependsOnIncreasesRank(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/b", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/a"}}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	a := testFindEntry(ranked, "ROOT/a")
	b := testFindEntry(ranked, "ROOT/b")
	if a == nil || b == nil {
		t.Fatal("missing expected entries")
	}
	if b.Rank <= a.Rank {
		t.Errorf("expected ROOT/b rank > ROOT/a rank, got %d <= %d", b.Rank, a.Rank)
	}
}

func TestNodeRankCompute_DependsOnQualifierStripped(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/b", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/a(interface)"}}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	a := testFindEntry(ranked, "ROOT/a")
	b := testFindEntry(ranked, "ROOT/b")
	if a == nil || b == nil {
		t.Fatal("missing expected entries")
	}
	if b.Rank <= a.Rank {
		t.Errorf("expected ROOT/b rank > ROOT/a rank, got %d <= %d", b.Rank, a.Rank)
	}
}

func TestNodeRankCompute_InputArtifactAddsDependency(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{Output: "out.go"}},
		{LogicalName: "ROOT/b", Frontmatter: &frontmatter.Frontmatter{Input: "ARTIFACT/a"}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	a := testFindEntry(ranked, "ROOT/a")
	artifact := testFindEntry(ranked, "ARTIFACT/a")
	b := testFindEntry(ranked, "ROOT/b")
	if a == nil || artifact == nil || b == nil {
		t.Fatal("missing expected entries")
	}
	if artifact.Rank <= a.Rank {
		t.Errorf("expected ARTIFACT/a rank > ROOT/a rank, got %d <= %d", artifact.Rank, a.Rank)
	}
	if b.Rank <= artifact.Rank {
		t.Errorf("expected ROOT/b rank > ARTIFACT/a rank, got %d <= %d", b.Rank, artifact.Rank)
	}
}

func TestNodeRankCompute_ArtifactRankOneAboveNode(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{Output: "foo.go"}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	a := testFindEntry(ranked, "ROOT/a")
	artifact := testFindEntry(ranked, "ARTIFACT/a")
	if a == nil || artifact == nil {
		t.Fatal("missing expected entries")
	}
	if artifact.Rank != a.Rank+1 {
		t.Errorf("expected ARTIFACT/a rank = ROOT/a rank + 1, got %d vs %d", artifact.Rank, a.Rank)
	}
}

func TestNodeRankCompute_SingleOutputArtifactRanked(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{Output: "x.go"}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	artifactCount := 0
	for _, e := range ranked {
		if e.LogicalName == "ARTIFACT/a" {
			artifactCount++
		}
	}
	if artifactCount != 1 {
		t.Errorf("expected exactly one ARTIFACT/a entry, got %d", artifactCount)
	}

	a := testFindEntry(ranked, "ROOT/a")
	artifact := testFindEntry(ranked, "ARTIFACT/a")
	if a == nil || artifact == nil {
		t.Fatal("missing expected entries")
	}
	if artifact.Rank != a.Rank+1 {
		t.Errorf("expected ARTIFACT/a rank = ROOT/a rank + 1, got %d vs %d", artifact.Rank, a.Rank)
	}
}

func TestNodeRankCompute_DependsOnArtifactReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{Output: "lib.go"}},
		{LogicalName: "ROOT/b", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ARTIFACT/a"}}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	a := testFindEntry(ranked, "ROOT/a")
	artifact := testFindEntry(ranked, "ARTIFACT/a")
	b := testFindEntry(ranked, "ROOT/b")
	if a == nil || artifact == nil || b == nil {
		t.Fatal("missing expected entries")
	}
	if artifact.Rank <= a.Rank {
		t.Errorf("expected ARTIFACT/a rank > ROOT/a rank, got %d <= %d", artifact.Rank, a.Rank)
	}
	if b.Rank <= artifact.Rank {
		t.Errorf("expected ROOT/b rank > ARTIFACT/a rank, got %d <= %d", b.Rank, artifact.Rank)
	}
}

func TestNodeRankCompute_OutputSortedByRankThenName(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/z", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	if len(ranked) < 3 {
		t.Fatalf("expected at least 3 entries, got %d", len(ranked))
	}
	if ranked[0].LogicalName != "ROOT" {
		t.Errorf("expected ROOT first, got %s", ranked[0].LogicalName)
	}

	var aIdx, zIdx int
	for i, e := range ranked {
		if e.LogicalName == "ROOT/a" {
			aIdx = i
		}
		if e.LogicalName == "ROOT/z" {
			zIdx = i
		}
	}
	if aIdx >= zIdx {
		t.Errorf("expected ROOT/a before ROOT/z in output, got indices %d and %d", aIdx, zIdx)
	}
}

func TestNodeRankCompute_ParallelEntries(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/b", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/c", Frontmatter: &frontmatter.Frontmatter{}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	for _, name := range []string{"ROOT/a", "ROOT/b", "ROOT/c"} {
		e := testFindEntry(ranked, name)
		if e == nil {
			t.Fatalf("missing entry for %s", name)
		}
		if e.Rank != 1 {
			t.Errorf("expected %s rank 1, got %d", name, e.Rank)
		}
	}
}

func TestNodeRankCompute_DiamondDependency(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/c", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/c"}}},
		{LogicalName: "ROOT/b", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/c"}}},
		{LogicalName: "ROOT/d", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/a", "ROOT/b"}}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	c := testFindEntry(ranked, "ROOT/c")
	a := testFindEntry(ranked, "ROOT/a")
	b := testFindEntry(ranked, "ROOT/b")
	d := testFindEntry(ranked, "ROOT/d")
	if c == nil || a == nil || b == nil || d == nil {
		t.Fatal("missing expected entries")
	}
	if c.Rank != 1 {
		t.Errorf("expected ROOT/c rank 1, got %d", c.Rank)
	}
	if a.Rank != 2 {
		t.Errorf("expected ROOT/a rank 2, got %d", a.Rank)
	}
	if b.Rank != 2 {
		t.Errorf("expected ROOT/b rank 2, got %d", b.Rank)
	}
	if d.Rank != 3 {
		t.Errorf("expected ROOT/d rank 3, got %d", d.Rank)
	}
}

func TestNodeRankCompute_DependsOnOutranksParent(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a/b", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/c"}}},
		{LogicalName: "ROOT/c", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/c/d", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/c/d/e", Frontmatter: &frontmatter.Frontmatter{}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	a := testFindEntry(ranked, "ROOT/a")
	ab := testFindEntry(ranked, "ROOT/a/b")
	c := testFindEntry(ranked, "ROOT/c")
	if a == nil || ab == nil || c == nil {
		t.Fatal("missing expected entries")
	}
	if ab.Rank <= a.Rank {
		t.Errorf("expected ROOT/a/b rank > ROOT/a rank, got %d <= %d", ab.Rank, a.Rank)
	}
	expectedRank := 1 + max(a.Rank, c.Rank)
	if ab.Rank != expectedRank {
		t.Errorf("expected ROOT/a/b rank = 1 + max(ROOT/a, ROOT/c) = %d, got %d", expectedRank, ab.Rank)
	}
}

func TestNodeRankCompute_MultipleDependsOnRankFromHighest(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/b", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/a"}}},
		{LogicalName: "ROOT/c", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/b"}}},
		{LogicalName: "ROOT/d", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/a", "ROOT/b", "ROOT/c"}}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	a := testFindEntry(ranked, "ROOT/a")
	b := testFindEntry(ranked, "ROOT/b")
	c := testFindEntry(ranked, "ROOT/c")
	d := testFindEntry(ranked, "ROOT/d")
	if a == nil || b == nil || c == nil || d == nil {
		t.Fatal("missing expected entries")
	}
	if a.Rank != 1 {
		t.Errorf("expected ROOT/a rank 1, got %d", a.Rank)
	}
	if b.Rank != 2 {
		t.Errorf("expected ROOT/b rank 2, got %d", b.Rank)
	}
	if c.Rank != 3 {
		t.Errorf("expected ROOT/c rank 3, got %d", c.Rank)
	}
	if d.Rank != 4 {
		t.Errorf("expected ROOT/d rank 4, got %d", d.Rank)
	}
}

func TestNodeRankCompute_BothDependsOnAndInput(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{Output: "a.go"}},
		{LogicalName: "ROOT/b", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/c", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/b"}, Input: "ARTIFACT/a"}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	root := testFindEntry(ranked, "ROOT")
	b := testFindEntry(ranked, "ROOT/b")
	artifactA := testFindEntry(ranked, "ARTIFACT/a")
	c := testFindEntry(ranked, "ROOT/c")
	if root == nil || b == nil || artifactA == nil || c == nil {
		t.Fatal("missing expected entries")
	}
	expectedRank := 1 + max(root.Rank, b.Rank, artifactA.Rank)
	if c.Rank != expectedRank {
		t.Errorf("expected ROOT/c rank = %d, got %d", expectedRank, c.Rank)
	}
}

func TestNodeRankCompute_EmptyInput(t *testing.T) {
	ranked, cycles, err := noderanking.NodeRankCompute([]*noderanking.NodeRankInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
	if len(ranked) != 0 {
		t.Errorf("expected empty ranked, got %v", ranked)
	}
}

func TestNodeRankCompute_SelfReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/a"}}},
	}

	_, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Error("expected cycles to be non-empty for self-reference")
	}
}

func TestNodeRankCompute_SimpleCycleTwoNodes(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/b"}}},
		{LogicalName: "ROOT/b", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/a"}}},
	}

	_, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Error("expected cycles to be non-empty")
	}
	if !testInCycles(cycles, "ROOT/a") && !testInCycles(cycles, "ROOT/b") {
		t.Errorf("expected cycles to contain ROOT/a or ROOT/b, got %v", cycles)
	}
}

func TestNodeRankCompute_CycleThroughArtifacts(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{Output: "a.go", DependsOn: []string{"ARTIFACT/b"}}},
		{LogicalName: "ROOT/b", Frontmatter: &frontmatter.Frontmatter{Output: "b.go", DependsOn: []string{"ARTIFACT/a"}}},
	}

	_, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Error("expected cycles to be non-empty")
	}
}

func TestNodeRankCompute_CycleDoesNotPreventUnrelatedRanking(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/b"}}},
		{LogicalName: "ROOT/b", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/a"}}},
		{LogicalName: "ROOT/c", Frontmatter: &frontmatter.Frontmatter{}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Error("expected cycles to be non-empty")
	}

	root := testFindEntry(ranked, "ROOT")
	c := testFindEntry(ranked, "ROOT/c")
	if root == nil || c == nil {
		t.Fatal("missing expected entries for ROOT or ROOT/c")
	}
	if root.Rank != 0 {
		t.Errorf("expected ROOT rank 0, got %d", root.Rank)
	}
	if c.Rank != 1 {
		t.Errorf("expected ROOT/c rank 1, got %d", c.Rank)
	}
	if testInCycles(cycles, "ROOT/c") {
		t.Error("ROOT/c should not be in cycles")
	}
	if !testInCycles(cycles, "ROOT/a") && !testInCycles(cycles, "ROOT/b") {
		t.Errorf("expected cycles to contain ROOT/a or ROOT/b, got %v", cycles)
	}
}

func TestNodeRankCompute_UnresolvableRootReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/missing"}}},
	}

	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got %v", err)
	}
}

func TestNodeRankCompute_UnresolvableArtifactReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ARTIFACT/missing"}}},
	}

	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got %v", err)
	}
}

func TestNodeRankCompute_UnresolvableInputReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{Input: "ARTIFACT/missing"}},
	}

	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got %v", err)
	}
}

func max(vals ...int) int {
	m := vals[0]
	for _, v := range vals[1:] {
		if v > m {
			m = v
		}
	}
	return m
}
