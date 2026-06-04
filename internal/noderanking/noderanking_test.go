// code-from-spec: ROOT/golang/tests/utils/node_ranking@WcOrg7Qe6zR07tcKw9D7QQ_AXzM
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
	if len(ranked) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(ranked))
	}
	e := testFindEntry(ranked, "ROOT")
	if e == nil {
		t.Fatal("expected ROOT in ranked")
	}
	if e.Rank != 0 {
		t.Errorf("expected rank 0, got %d", e.Rank)
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
		t.Errorf("ROOT: expected rank 0, got %d", root.Rank)
	}
	if a.Rank != 1 {
		t.Errorf("ROOT/a: expected rank 1, got %d", a.Rank)
	}
	if ab.Rank != 2 {
		t.Errorf("ROOT/a/b: expected rank 2, got %d", ab.Rank)
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
		t.Errorf("ROOT/a: expected rank 1, got %d", a.Rank)
	}
	if b.Rank != 1 {
		t.Errorf("ROOT/b: expected rank 1, got %d", b.Rank)
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
		t.Errorf("expected rank of ROOT/b (%d) > rank of ROOT/a (%d)", b.Rank, a.Rank)
	}
}

func TestNodeRankCompute_DependsOnWithQualifier(t *testing.T) {
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
		t.Errorf("expected rank of ROOT/b (%d) > rank of ROOT/a (%d)", b.Rank, a.Rank)
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
		t.Errorf("expected no cycles, got %v", cycles)
	}

	a := testFindEntry(ranked, "ROOT/a")
	artifact := testFindEntry(ranked, "ARTIFACT/a")
	b := testFindEntry(ranked, "ROOT/b")

	if a == nil {
		t.Fatal("missing ROOT/a")
	}
	if artifact == nil {
		t.Fatal("missing ARTIFACT/a in ranked")
	}
	if b == nil {
		t.Fatal("missing ROOT/b")
	}
	if artifact.Rank <= a.Rank {
		t.Errorf("expected rank of ARTIFACT/a (%d) > rank of ROOT/a (%d)", artifact.Rank, a.Rank)
	}
	if b.Rank <= artifact.Rank {
		t.Errorf("expected rank of ROOT/b (%d) > rank of ARTIFACT/a (%d)", b.Rank, artifact.Rank)
	}
}

func TestNodeRankCompute_ArtifactsRankOneAboveNode(t *testing.T) {
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

	if a == nil {
		t.Fatal("missing ROOT/a")
	}
	if artifact == nil {
		t.Fatal("missing ARTIFACT/a in ranked")
	}
	if artifact.Rank != a.Rank+1 {
		t.Errorf("expected ARTIFACT/a rank = %d, got %d", a.Rank+1, artifact.Rank)
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

	count := 0
	for _, e := range ranked {
		if e.LogicalName == "ARTIFACT/a" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly one ARTIFACT/a entry, got %d", count)
	}

	a := testFindEntry(ranked, "ROOT/a")
	artifact := testFindEntry(ranked, "ARTIFACT/a")

	if a == nil || artifact == nil {
		t.Fatal("missing expected entries")
	}
	if artifact.Rank != a.Rank+1 {
		t.Errorf("expected ARTIFACT/a rank = %d, got %d", a.Rank+1, artifact.Rank)
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
		t.Errorf("expected rank of ARTIFACT/a (%d) > rank of ROOT/a (%d)", artifact.Rank, a.Rank)
	}
	if b.Rank <= artifact.Rank {
		t.Errorf("expected rank of ROOT/b (%d) > rank of ARTIFACT/a (%d)", b.Rank, artifact.Rank)
	}
}

func TestNodeRankCompute_OutputSortedByRankThenLogicalName(t *testing.T) {
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

	if len(ranked) == 0 {
		t.Fatal("empty ranked")
	}
	if ranked[0].LogicalName != "ROOT" {
		t.Errorf("expected ROOT first, got %s", ranked[0].LogicalName)
	}

	aIdx := -1
	zIdx := -1
	for i, e := range ranked {
		if e.LogicalName == "ROOT/a" {
			aIdx = i
		}
		if e.LogicalName == "ROOT/z" {
			zIdx = i
		}
	}
	if aIdx == -1 || zIdx == -1 {
		t.Fatal("missing ROOT/a or ROOT/z")
	}
	if aIdx >= zIdx {
		t.Errorf("expected ROOT/a before ROOT/z in ranked output")
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
			t.Fatalf("missing %s", name)
		}
		if e.Rank != 1 {
			t.Errorf("%s: expected rank 1, got %d", name, e.Rank)
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
		t.Errorf("ROOT/c: expected rank 1, got %d", c.Rank)
	}
	if a.Rank != 2 {
		t.Errorf("ROOT/a: expected rank 2, got %d", a.Rank)
	}
	if b.Rank != 2 {
		t.Errorf("ROOT/b: expected rank 2, got %d", b.Rank)
	}
	if d.Rank != 3 {
		t.Errorf("ROOT/d: expected rank 3, got %d", d.Rank)
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
		t.Errorf("expected rank of ROOT/a/b (%d) > rank of ROOT/a (%d)", ab.Rank, a.Rank)
	}
	expectedRank := 1 + max(a.Rank, c.Rank)
	if ab.Rank != expectedRank {
		t.Errorf("ROOT/a/b: expected rank %d (1 + max(%d, %d)), got %d", expectedRank, a.Rank, c.Rank, ab.Rank)
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
		t.Errorf("ROOT/a: expected rank 1, got %d", a.Rank)
	}
	if b.Rank != 2 {
		t.Errorf("ROOT/b: expected rank 2, got %d", b.Rank)
	}
	if c.Rank != 3 {
		t.Errorf("ROOT/c: expected rank 3, got %d", c.Rank)
	}
	if d.Rank != 4 {
		t.Errorf("ROOT/d: expected rank 4, got %d", d.Rank)
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
		t.Errorf("expected no cycles, got %v", cycles)
	}

	root := testFindEntry(ranked, "ROOT")
	b := testFindEntry(ranked, "ROOT/b")
	artifactA := testFindEntry(ranked, "ARTIFACT/a")
	c := testFindEntry(ranked, "ROOT/c")

	if root == nil || b == nil || artifactA == nil || c == nil {
		t.Fatal("missing expected entries")
	}

	expectedRank := 1 + max(root.Rank, max(b.Rank, artifactA.Rank))
	if c.Rank != expectedRank {
		t.Errorf("ROOT/c: expected rank %d, got %d", expectedRank, c.Rank)
	}
}

func TestNodeRankCompute_EmptyInputList(t *testing.T) {
	entries := []*noderanking.NodeRankInput{}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
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

	hasA := false
	hasB := false
	for _, c := range cycles {
		if c == "ROOT/a" {
			hasA = true
		}
		if c == "ROOT/b" {
			hasB = true
		}
	}
	if !hasA && !hasB {
		t.Errorf("cycles should contain ROOT/a or ROOT/b, got %v", cycles)
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

func TestNodeRankCompute_CycleDoesNotPreventRankingUnrelatedNodes(t *testing.T) {
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

	if root == nil {
		t.Fatal("missing ROOT in ranked")
	}
	if c == nil {
		t.Fatal("missing ROOT/c in ranked")
	}
	if root.Rank != 0 {
		t.Errorf("ROOT: expected rank 0, got %d", root.Rank)
	}
	if c.Rank != 1 {
		t.Errorf("ROOT/c: expected rank 1, got %d", c.Rank)
	}

	for _, cycle := range cycles {
		if cycle == "ROOT/c" {
			t.Errorf("ROOT/c should not be in cycles, got %v", cycles)
		}
	}
}

func TestNodeRankCompute_UnresolvableROOTReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/missing"}}},
	}

	_, _, err := noderanking.NodeRankCompute(entries)
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
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got %v", err)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
