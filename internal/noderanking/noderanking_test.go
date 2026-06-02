// code-from-spec: ROOT/golang/tests/utils/node_ranking@rDJNK9m7vA9xOwmUnyEmldP10Zo
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

func testContainsAny(list []string, names ...string) bool {
	for _, name := range names {
		if testContains(list, name) {
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
		t.Fatal("ROOT not found in ranked")
	}
	if entry.Rank != 0 {
		t.Errorf("expected ROOT rank=0, got %d", entry.Rank)
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
		t.Fatal("missing entries in ranked")
	}
	if root.Rank != 0 {
		t.Errorf("ROOT rank: want 0, got %d", root.Rank)
	}
	if a.Rank != 1 {
		t.Errorf("ROOT/a rank: want 1, got %d", a.Rank)
	}
	if ab.Rank != 2 {
		t.Errorf("ROOT/a/b rank: want 2, got %d", ab.Rank)
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
		t.Fatal("missing entries in ranked")
	}
	if a.Rank != b.Rank {
		t.Errorf("ROOT/a and ROOT/b should have equal rank, got %d and %d", a.Rank, b.Rank)
	}
	if a.Rank != 1 {
		t.Errorf("expected rank=1, got %d", a.Rank)
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
		t.Fatal("missing entries in ranked")
	}
	if b.Rank <= a.Rank {
		t.Errorf("ROOT/b rank (%d) should be > ROOT/a rank (%d)", b.Rank, a.Rank)
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
		t.Fatal("missing entries in ranked")
	}
	if b.Rank <= a.Rank {
		t.Errorf("ROOT/b rank (%d) should be > ROOT/a rank (%d)", b.Rank, a.Rank)
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
		t.Fatalf("missing entries: ROOT/a=%v, ARTIFACT/a=%v, ROOT/b=%v", a, artifact, b)
	}
	if artifact.Rank <= a.Rank {
		t.Errorf("ARTIFACT/a rank (%d) should be > ROOT/a rank (%d)", artifact.Rank, a.Rank)
	}
	if b.Rank <= artifact.Rank {
		t.Errorf("ROOT/b rank (%d) should be > ARTIFACT/a rank (%d)", b.Rank, artifact.Rank)
	}
}

func TestNodeRankCompute_ArtifactRankedAboveNode(t *testing.T) {
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
		t.Fatalf("missing entries: ROOT/a=%v, ARTIFACT/a=%v", a, artifact)
	}
	if artifact.Rank != a.Rank+1 {
		t.Errorf("ARTIFACT/a rank (%d) should be ROOT/a rank (%d) + 1", artifact.Rank, a.Rank)
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

	a := testFindEntry(ranked, "ROOT/a")
	artifact := testFindEntry(ranked, "ARTIFACT/a")

	if a == nil {
		t.Fatal("ROOT/a not found in ranked")
	}
	if artifact == nil {
		t.Fatal("ARTIFACT/a not found in ranked")
	}
	if artifact.Rank != a.Rank+1 {
		t.Errorf("ARTIFACT/a rank (%d) should equal ROOT/a rank (%d) + 1", artifact.Rank, a.Rank)
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
		t.Fatalf("missing entries: ROOT/a=%v, ARTIFACT/a=%v, ROOT/b=%v", a, artifact, b)
	}
	if artifact.Rank <= a.Rank {
		t.Errorf("ARTIFACT/a rank (%d) should be > ROOT/a rank (%d)", artifact.Rank, a.Rank)
	}
	if b.Rank <= artifact.Rank {
		t.Errorf("ROOT/b rank (%d) should be > ARTIFACT/a rank (%d)", b.Rank, artifact.Rank)
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
		t.Errorf("first entry should be ROOT, got %s", ranked[0].LogicalName)
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
		t.Errorf("ROOT/a (idx %d) should come before ROOT/z (idx %d)", aIdx, zIdx)
	}
}

func TestNodeRankCompute_ParallelEntriesEqualRank(t *testing.T) {
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

	a := testFindEntry(ranked, "ROOT/a")
	b := testFindEntry(ranked, "ROOT/b")
	c := testFindEntry(ranked, "ROOT/c")

	if a == nil || b == nil || c == nil {
		t.Fatal("missing entries in ranked")
	}
	if a.Rank != 1 || b.Rank != 1 || c.Rank != 1 {
		t.Errorf("ROOT/a, ROOT/b, ROOT/c should all have rank=1, got %d, %d, %d", a.Rank, b.Rank, c.Rank)
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
		t.Fatal("missing entries in ranked")
	}
	if c.Rank != 1 {
		t.Errorf("ROOT/c rank: want 1, got %d", c.Rank)
	}
	if a.Rank != 2 {
		t.Errorf("ROOT/a rank: want 2, got %d", a.Rank)
	}
	if b.Rank != 2 {
		t.Errorf("ROOT/b rank: want 2, got %d", b.Rank)
	}
	if d.Rank != 3 {
		t.Errorf("ROOT/d rank: want 3, got %d", d.Rank)
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
		t.Fatal("missing entries in ranked")
	}

	expectedRank := 1 + max(a.Rank, c.Rank)
	if ab.Rank != expectedRank {
		t.Errorf("ROOT/a/b rank: want %d, got %d", expectedRank, ab.Rank)
	}
	if ab.Rank <= a.Rank {
		t.Errorf("ROOT/a/b rank (%d) should be > ROOT/a rank (%d)", ab.Rank, a.Rank)
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
		t.Fatal("missing entries in ranked")
	}
	if a.Rank != 1 {
		t.Errorf("ROOT/a rank: want 1, got %d", a.Rank)
	}
	if b.Rank != 2 {
		t.Errorf("ROOT/b rank: want 2, got %d", b.Rank)
	}
	if c.Rank != 3 {
		t.Errorf("ROOT/c rank: want 3, got %d", c.Rank)
	}
	if d.Rank != 4 {
		t.Errorf("ROOT/d rank: want 4, got %d", d.Rank)
	}
}

func TestNodeRankCompute_NodeWithBothDependsOnAndInput(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{Output: "a.go"}},
		{LogicalName: "ROOT/b", Frontmatter: &frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/c",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/b"},
				Input:     "ARTIFACT/a",
			},
		},
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
	artifact := testFindEntry(ranked, "ARTIFACT/a")
	c := testFindEntry(ranked, "ROOT/c")

	if root == nil || b == nil || artifact == nil || c == nil {
		t.Fatalf("missing entries: ROOT=%v, ROOT/b=%v, ARTIFACT/a=%v, ROOT/c=%v", root, b, artifact, c)
	}

	expectedMin := 1 + max(root.Rank, max(b.Rank, artifact.Rank))
	if c.Rank < expectedMin {
		t.Errorf("ROOT/c rank (%d) should be >= %d", c.Rank, expectedMin)
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
		t.Errorf("expected empty ranked list, got %v", ranked)
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
	if !testContainsAny(cycles, "ROOT/a", "ROOT/b") {
		t.Errorf("cycles should contain ROOT/a or ROOT/b, got %v", cycles)
	}
}

func TestNodeRankCompute_CycleThroughArtifacts(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				Output:    "a.go",
				DependsOn: []string{"ARTIFACT/b"},
			},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{
				Output:    "b.go",
				DependsOn: []string{"ARTIFACT/a"},
			},
		},
	}

	_, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Error("expected cycles to be non-empty")
	}
}

func TestNodeRankCompute_CycleDoesNotPreventRankingUnrelated(t *testing.T) {
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
		t.Error("ROOT should have a valid rank")
	} else if root.Rank != 0 {
		t.Errorf("ROOT rank: want 0, got %d", root.Rank)
	}

	if c == nil {
		t.Error("ROOT/c should have a valid rank")
	} else if c.Rank != 1 {
		t.Errorf("ROOT/c rank: want 1, got %d", c.Rank)
	}

	if testContains(cycles, "ROOT/c") {
		t.Errorf("ROOT/c should not be in cycles, got %v", cycles)
	}
	if !testContainsAny(cycles, "ROOT/a", "ROOT/b") {
		t.Errorf("cycles should contain ROOT/a or ROOT/b, got %v", cycles)
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
