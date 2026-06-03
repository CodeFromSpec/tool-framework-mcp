// code-from-spec: ROOT/golang/tests/utils/node_ranking@_OtD0Wra-8_LSWgsnW3TiMCgJtA
package noderanking_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/noderanking"
)

func testFindEntry(ranked []*noderanking.NodeRankEntry, name string) *noderanking.NodeRankEntry {
	for _, e := range ranked {
		if e.LogicalName == name {
			return e
		}
	}
	return nil
}

func testContains(list []string, name string) bool {
	for _, s := range list {
		if s == name {
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
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	e := testFindEntry(ranked, "ROOT")
	if e == nil {
		t.Fatal("expected ROOT in ranked")
	}
	if e.Rank != 0 {
		t.Fatalf("expected rank 0, got %d", e.Rank)
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
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	root := testFindEntry(ranked, "ROOT")
	a := testFindEntry(ranked, "ROOT/a")
	ab := testFindEntry(ranked, "ROOT/a/b")
	if root == nil || a == nil || ab == nil {
		t.Fatal("missing entries")
	}
	if root.Rank != 0 {
		t.Errorf("ROOT rank: got %d, want 0", root.Rank)
	}
	if a.Rank != 1 {
		t.Errorf("ROOT/a rank: got %d, want 1", a.Rank)
	}
	if ab.Rank != 2 {
		t.Errorf("ROOT/a/b rank: got %d, want 2", ab.Rank)
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
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	a := testFindEntry(ranked, "ROOT/a")
	b := testFindEntry(ranked, "ROOT/b")
	if a == nil || b == nil {
		t.Fatal("missing entries")
	}
	if a.Rank != 1 {
		t.Errorf("ROOT/a rank: got %d, want 1", a.Rank)
	}
	if b.Rank != 1 {
		t.Errorf("ROOT/b rank: got %d, want 1", b.Rank)
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
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	a := testFindEntry(ranked, "ROOT/a")
	b := testFindEntry(ranked, "ROOT/b")
	if a == nil || b == nil {
		t.Fatal("missing entries")
	}
	if b.Rank <= a.Rank {
		t.Errorf("expected rank of ROOT/b (%d) > ROOT/a (%d)", b.Rank, a.Rank)
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
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	a := testFindEntry(ranked, "ROOT/a")
	b := testFindEntry(ranked, "ROOT/b")
	if a == nil || b == nil {
		t.Fatal("missing entries")
	}
	if b.Rank <= a.Rank {
		t.Errorf("expected rank of ROOT/b (%d) > ROOT/a (%d)", b.Rank, a.Rank)
	}
}

func TestNodeRankCompute_InputArtifactAddsDependencyEdge(t *testing.T) {
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
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	artifact := testFindEntry(ranked, "ARTIFACT/a")
	nodeA := testFindEntry(ranked, "ROOT/a")
	nodeB := testFindEntry(ranked, "ROOT/b")
	if artifact == nil {
		t.Fatal("expected ARTIFACT/a in ranked")
	}
	if nodeA == nil || nodeB == nil {
		t.Fatal("missing entries")
	}
	if nodeB.Rank <= artifact.Rank {
		t.Errorf("expected rank of ROOT/b (%d) > ARTIFACT/a (%d)", nodeB.Rank, artifact.Rank)
	}
	if artifact.Rank <= nodeA.Rank {
		t.Errorf("expected rank of ARTIFACT/a (%d) > ROOT/a (%d)", artifact.Rank, nodeA.Rank)
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
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	artifact := testFindEntry(ranked, "ARTIFACT/a")
	nodeA := testFindEntry(ranked, "ROOT/a")
	if artifact == nil {
		t.Fatal("expected ARTIFACT/a in ranked")
	}
	if nodeA == nil {
		t.Fatal("expected ROOT/a in ranked")
	}
	if artifact.Rank != nodeA.Rank+1 {
		t.Errorf("expected ARTIFACT/a rank = ROOT/a rank + 1, got %d and %d", artifact.Rank, nodeA.Rank)
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
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	count := 0
	for _, e := range ranked {
		if e.LogicalName == "ARTIFACT/a" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly one ARTIFACT/a entry, got %d", count)
	}
	artifact := testFindEntry(ranked, "ARTIFACT/a")
	nodeA := testFindEntry(ranked, "ROOT/a")
	if artifact == nil || nodeA == nil {
		t.Fatal("missing entries")
	}
	if artifact.Rank != nodeA.Rank+1 {
		t.Errorf("expected ARTIFACT/a rank = ROOT/a rank + 1, got %d and %d", artifact.Rank, nodeA.Rank)
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
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	artifact := testFindEntry(ranked, "ARTIFACT/a")
	nodeA := testFindEntry(ranked, "ROOT/a")
	nodeB := testFindEntry(ranked, "ROOT/b")
	if artifact == nil || nodeA == nil || nodeB == nil {
		t.Fatal("missing entries")
	}
	if nodeB.Rank <= artifact.Rank {
		t.Errorf("expected rank of ROOT/b (%d) > ARTIFACT/a (%d)", nodeB.Rank, artifact.Rank)
	}
	if artifact.Rank <= nodeA.Rank {
		t.Errorf("expected rank of ARTIFACT/a (%d) > ROOT/a (%d)", artifact.Rank, nodeA.Rank)
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
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	if len(ranked) < 3 {
		t.Fatalf("expected at least 3 entries, got %d", len(ranked))
	}
	if ranked[0].LogicalName != "ROOT" {
		t.Errorf("expected ROOT first, got %s", ranked[0].LogicalName)
	}
	indexA := -1
	indexZ := -1
	for i, e := range ranked {
		if e.LogicalName == "ROOT/a" {
			indexA = i
		}
		if e.LogicalName == "ROOT/z" {
			indexZ = i
		}
	}
	if indexA == -1 || indexZ == -1 {
		t.Fatal("missing ROOT/a or ROOT/z")
	}
	if indexA >= indexZ {
		t.Errorf("expected ROOT/a before ROOT/z in ranked order")
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
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	for _, name := range []string{"ROOT/a", "ROOT/b", "ROOT/c"} {
		e := testFindEntry(ranked, name)
		if e == nil {
			t.Fatalf("missing %s", name)
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
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	c := testFindEntry(ranked, "ROOT/c")
	a := testFindEntry(ranked, "ROOT/a")
	b := testFindEntry(ranked, "ROOT/b")
	d := testFindEntry(ranked, "ROOT/d")
	if c == nil || a == nil || b == nil || d == nil {
		t.Fatal("missing entries")
	}
	if c.Rank != 1 {
		t.Errorf("ROOT/c rank: got %d, want 1", c.Rank)
	}
	if a.Rank != 2 {
		t.Errorf("ROOT/a rank: got %d, want 2", a.Rank)
	}
	if b.Rank != 2 {
		t.Errorf("ROOT/b rank: got %d, want 2", b.Rank)
	}
	if d.Rank != 3 {
		t.Errorf("ROOT/d rank: got %d, want 3", d.Rank)
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
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	a := testFindEntry(ranked, "ROOT/a")
	ab := testFindEntry(ranked, "ROOT/a/b")
	c := testFindEntry(ranked, "ROOT/c")
	if a == nil || ab == nil || c == nil {
		t.Fatal("missing entries")
	}
	if ab.Rank <= a.Rank {
		t.Errorf("expected ROOT/a/b rank (%d) > ROOT/a rank (%d)", ab.Rank, a.Rank)
	}
	maxParentDep := a.Rank
	if c.Rank > maxParentDep {
		maxParentDep = c.Rank
	}
	if ab.Rank != 1+maxParentDep {
		t.Errorf("expected ROOT/a/b rank = 1 + max(ROOT/a, ROOT/c) = %d, got %d", 1+maxParentDep, ab.Rank)
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
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	a := testFindEntry(ranked, "ROOT/a")
	b := testFindEntry(ranked, "ROOT/b")
	c := testFindEntry(ranked, "ROOT/c")
	d := testFindEntry(ranked, "ROOT/d")
	if a == nil || b == nil || c == nil || d == nil {
		t.Fatal("missing entries")
	}
	if a.Rank != 1 {
		t.Errorf("ROOT/a rank: got %d, want 1", a.Rank)
	}
	if b.Rank != 2 {
		t.Errorf("ROOT/b rank: got %d, want 2", b.Rank)
	}
	if c.Rank != 3 {
		t.Errorf("ROOT/c rank: got %d, want 3", c.Rank)
	}
	if d.Rank != 4 {
		t.Errorf("ROOT/d rank: got %d, want 4", d.Rank)
	}
}

func TestNodeRankCompute_NodeWithBothDependsOnAndInput(t *testing.T) {
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
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	root := testFindEntry(ranked, "ROOT")
	b := testFindEntry(ranked, "ROOT/b")
	artifact := testFindEntry(ranked, "ARTIFACT/a")
	c := testFindEntry(ranked, "ROOT/c")
	if root == nil || b == nil || artifact == nil || c == nil {
		t.Fatal("missing entries")
	}
	maxDep := root.Rank
	if b.Rank > maxDep {
		maxDep = b.Rank
	}
	if artifact.Rank > maxDep {
		maxDep = artifact.Rank
	}
	if c.Rank != 1+maxDep {
		t.Errorf("expected ROOT/c rank = 1 + max(...) = %d, got %d", 1+maxDep, c.Rank)
	}
}

func TestNodeRankCompute_EmptyInput(t *testing.T) {
	ranked, cycles, err := noderanking.NodeRankCompute([]*noderanking.NodeRankInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	if len(ranked) != 0 {
		t.Fatalf("expected empty ranked, got %v", ranked)
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
		t.Fatal("expected cycles, got none")
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
		t.Fatal("expected cycles, got none")
	}
	if !testContains(cycles, "ROOT/a") && !testContains(cycles, "ROOT/b") {
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
		t.Fatal("expected cycles, got none")
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
		t.Fatal("expected cycles, got none")
	}
	root := testFindEntry(ranked, "ROOT")
	c := testFindEntry(ranked, "ROOT/c")
	if root == nil {
		t.Fatal("expected ROOT in ranked")
	}
	if c == nil {
		t.Fatal("expected ROOT/c in ranked")
	}
	if root.Rank != 0 {
		t.Errorf("ROOT rank: got %d, want 0", root.Rank)
	}
	if c.Rank != 1 {
		t.Errorf("ROOT/c rank: got %d, want 1", c.Rank)
	}
	if testContains(cycles, "ROOT/c") {
		t.Errorf("ROOT/c should not be in cycles")
	}
}

func TestNodeRankCompute_UnresolvableROOTReference(t *testing.T) {
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

func TestNodeRankCompute_UnresolvableARTIFACTReference(t *testing.T) {
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
