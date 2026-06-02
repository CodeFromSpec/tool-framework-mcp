// code-from-spec: ROOT/golang/tests/utils/node_ranking@Q7DDUH4BmdcwppCVukWCb1hIAS8
package noderanking_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/noderanking"
)

func testFindRank(ranked []*noderanking.NodeRankEntry, logicalName string) (int, bool) {
	for _, r := range ranked {
		if r.LogicalName == logicalName {
			return r.Rank, true
		}
	}
	return 0, false
}

func TestNodeRankCompute_RootOnly(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: frontmatter.Frontmatter{}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
	if len(ranked) != 1 {
		t.Fatalf("expected 1 ranked entry, got %d", len(ranked))
	}
	rank, ok := testFindRank(ranked, "ROOT")
	if !ok {
		t.Fatal("ROOT not in ranked")
	}
	if rank != 0 {
		t.Errorf("expected ROOT rank 0, got %d", rank)
	}
}

func TestNodeRankCompute_LinearChain(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a/b", Frontmatter: frontmatter.Frontmatter{}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rootRank, _ := testFindRank(ranked, "ROOT")
	aRank, _ := testFindRank(ranked, "ROOT/a")
	abRank, _ := testFindRank(ranked, "ROOT/a/b")

	if rootRank != 0 {
		t.Errorf("expected ROOT rank 0, got %d", rootRank)
	}
	if aRank != 1 {
		t.Errorf("expected ROOT/a rank 1, got %d", aRank)
	}
	if abRank != 2 {
		t.Errorf("expected ROOT/a/b rank 2, got %d", abRank)
	}
}

func TestNodeRankCompute_IndependentSiblings_EqualRank(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/b", Frontmatter: frontmatter.Frontmatter{}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	aRank, _ := testFindRank(ranked, "ROOT/a")
	bRank, _ := testFindRank(ranked, "ROOT/b")

	if aRank != bRank {
		t.Errorf("expected ROOT/a and ROOT/b to have same rank, got %d and %d", aRank, bRank)
	}
	if aRank != 1 {
		t.Errorf("expected rank 1, got %d", aRank)
	}
}

func TestNodeRankCompute_DependsOn_IncreasesRank(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/b",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/a"}},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	aRank, _ := testFindRank(ranked, "ROOT/a")
	bRank, _ := testFindRank(ranked, "ROOT/b")

	if bRank <= aRank {
		t.Errorf("expected ROOT/b rank > ROOT/a rank, got %d <= %d", bRank, aRank)
	}
}

func TestNodeRankCompute_DependsOn_WithQualifier_QualifierStripped(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/b",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/a(interface)"}},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	aRank, _ := testFindRank(ranked, "ROOT/a")
	bRank, _ := testFindRank(ranked, "ROOT/b")

	if bRank <= aRank {
		t.Errorf("expected ROOT/b rank > ROOT/a rank, got %d <= %d", bRank, aRank)
	}
}

func TestNodeRankCompute_InputArtifact_AddsDependencyEdge(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Output: "out.go"},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: frontmatter.Frontmatter{Input: "ARTIFACT/a"},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	aRank, _ := testFindRank(ranked, "ROOT/a")
	artifactRank, _ := testFindRank(ranked, "ARTIFACT/a")
	bRank, _ := testFindRank(ranked, "ROOT/b")

	if artifactRank <= aRank {
		t.Errorf("expected ARTIFACT/a rank > ROOT/a rank, got %d <= %d", artifactRank, aRank)
	}
	if bRank <= artifactRank {
		t.Errorf("expected ROOT/b rank > ARTIFACT/a rank, got %d <= %d", bRank, artifactRank)
	}
}

func TestNodeRankCompute_ArtifactsGetRankAboveNode(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Output: "foo.go"},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	aRank, aOk := testFindRank(ranked, "ROOT/a")
	artifactRank, artifactOk := testFindRank(ranked, "ARTIFACT/a")

	if !aOk {
		t.Fatal("ROOT/a not in ranked")
	}
	if !artifactOk {
		t.Fatal("ARTIFACT/a not in ranked")
	}
	if artifactRank != aRank+1 {
		t.Errorf("expected ARTIFACT/a rank = ROOT/a rank + 1, got %d and %d", artifactRank, aRank)
	}
}

func TestNodeRankCompute_SingleOutput_ArtifactRanked(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Output: "x.go"},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	aRank, _ := testFindRank(ranked, "ROOT/a")
	artifactRank, artifactOk := testFindRank(ranked, "ARTIFACT/a")

	if !artifactOk {
		t.Fatal("ARTIFACT/a not in ranked")
	}
	if artifactRank != aRank+1 {
		t.Errorf("expected ARTIFACT/a rank = %d, got %d", aRank+1, artifactRank)
	}
}

func TestNodeRankCompute_DependsOnArtifact(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Output: "lib.go"},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ARTIFACT/a"}},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	aRank, _ := testFindRank(ranked, "ROOT/a")
	artifactRank, _ := testFindRank(ranked, "ARTIFACT/a")
	bRank, _ := testFindRank(ranked, "ROOT/b")

	if artifactRank <= aRank {
		t.Errorf("expected ARTIFACT/a rank > ROOT/a rank, got %d <= %d", artifactRank, aRank)
	}
	if bRank <= artifactRank {
		t.Errorf("expected ROOT/b rank > ARTIFACT/a rank, got %d <= %d", bRank, artifactRank)
	}
}

func TestNodeRankCompute_OutputSortedByRankThenLogicalName(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/z", Frontmatter: frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: frontmatter.Frontmatter{}},
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
		t.Errorf("expected first entry ROOT, got %s", ranked[0].LogicalName)
	}
	foundA := -1
	foundZ := -1
	for i, r := range ranked {
		if r.LogicalName == "ROOT/a" {
			foundA = i
		}
		if r.LogicalName == "ROOT/z" {
			foundZ = i
		}
	}
	if foundA == -1 || foundZ == -1 {
		t.Fatal("ROOT/a or ROOT/z not in ranked")
	}
	if foundA >= foundZ {
		t.Errorf("expected ROOT/a before ROOT/z in ranked output, got positions %d and %d", foundA, foundZ)
	}
}

func TestNodeRankCompute_ParallelEntries_EqualRank(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/b", Frontmatter: frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/c", Frontmatter: frontmatter.Frontmatter{}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	aRank, _ := testFindRank(ranked, "ROOT/a")
	bRank, _ := testFindRank(ranked, "ROOT/b")
	cRank, _ := testFindRank(ranked, "ROOT/c")

	if aRank != 1 || bRank != 1 || cRank != 1 {
		t.Errorf("expected all siblings at rank 1, got a=%d b=%d c=%d", aRank, bRank, cRank)
	}
}

func TestNodeRankCompute_DiamondDependency_RankUsesMax(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/c", Frontmatter: frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/c"}},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/c"}},
		},
		{
			LogicalName: "ROOT/d",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/a", "ROOT/b"}},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	cRank, _ := testFindRank(ranked, "ROOT/c")
	aRank, _ := testFindRank(ranked, "ROOT/a")
	bRank, _ := testFindRank(ranked, "ROOT/b")
	dRank, _ := testFindRank(ranked, "ROOT/d")

	if cRank != 1 {
		t.Errorf("expected ROOT/c rank 1, got %d", cRank)
	}
	if aRank != 2 {
		t.Errorf("expected ROOT/a rank 2, got %d", aRank)
	}
	if bRank != 2 {
		t.Errorf("expected ROOT/b rank 2, got %d", bRank)
	}
	if dRank != 3 {
		t.Errorf("expected ROOT/d rank 3 (max not sum), got %d", dRank)
	}
}

func TestNodeRankCompute_DependsOnOutranksParent(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/c"}},
		},
		{LogicalName: "ROOT/c", Frontmatter: frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/c/d", Frontmatter: frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/c/d/e", Frontmatter: frontmatter.Frontmatter{}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	aRank, _ := testFindRank(ranked, "ROOT/a")
	cRank, _ := testFindRank(ranked, "ROOT/c")
	abRank, _ := testFindRank(ranked, "ROOT/a/b")

	expectedRank := 1 + max(aRank, cRank)
	if abRank != expectedRank {
		t.Errorf("expected ROOT/a/b rank %d (1 + max(%d, %d)), got %d", expectedRank, aRank, cRank, abRank)
	}
	if abRank <= aRank {
		t.Errorf("expected ROOT/a/b rank > ROOT/a rank, got %d <= %d", abRank, aRank)
	}
}

func TestNodeRankCompute_MultipleDependsOn_RankFromHighest(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/b",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/a"}},
		},
		{
			LogicalName: "ROOT/c",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/b"}},
		},
		{
			LogicalName: "ROOT/d",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/a", "ROOT/b", "ROOT/c"}},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	aRank, _ := testFindRank(ranked, "ROOT/a")
	bRank, _ := testFindRank(ranked, "ROOT/b")
	cRank, _ := testFindRank(ranked, "ROOT/c")
	dRank, _ := testFindRank(ranked, "ROOT/d")

	if aRank != 1 {
		t.Errorf("expected ROOT/a rank 1, got %d", aRank)
	}
	if bRank != 2 {
		t.Errorf("expected ROOT/b rank 2, got %d", bRank)
	}
	if cRank != 3 {
		t.Errorf("expected ROOT/c rank 3, got %d", cRank)
	}
	if dRank != 4 {
		t.Errorf("expected ROOT/d rank 4, got %d", dRank)
	}
}

func TestNodeRankCompute_BothDependsOnAndInput(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Output: "a.go"},
		},
		{LogicalName: "ROOT/b", Frontmatter: frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/c",
			Frontmatter: frontmatter.Frontmatter{
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

	artifactRank, _ := testFindRank(ranked, "ARTIFACT/a")
	bRank, _ := testFindRank(ranked, "ROOT/b")
	cRank, _ := testFindRank(ranked, "ROOT/c")

	rootRank, _ := testFindRank(ranked, "ROOT")
	expectedCRank := 1 + max(rootRank, bRank, artifactRank)
	if cRank != expectedCRank {
		t.Errorf("expected ROOT/c rank %d, got %d", expectedCRank, cRank)
	}
}

func TestNodeRankCompute_EmptyInputList(t *testing.T) {
	ranked, cycles, err := noderanking.NodeRankCompute([]*noderanking.NodeRankInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ranked) != 0 {
		t.Errorf("expected empty ranked list, got %d", len(ranked))
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
}

func TestNodeRankCompute_SelfReference_CycleDetected(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/a"}},
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

func TestNodeRankCompute_SimpleCycle_TwoNodes(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/b"}},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/a"}},
		},
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
		t.Errorf("expected cycles to contain ROOT/a or ROOT/b, got %v", cycles)
	}
}

func TestNodeRankCompute_CycleThroughArtifacts(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{
				Output:    "a.go",
				DependsOn: []string{"ARTIFACT/b"},
			},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: frontmatter.Frontmatter{
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

func TestNodeRankCompute_CycleDoesNotPreventRankingOfUnrelated(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/b"}},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/a"}},
		},
		{LogicalName: "ROOT/c", Frontmatter: frontmatter.Frontmatter{}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Error("expected cycles to be non-empty")
	}

	rootRank, rootOk := testFindRank(ranked, "ROOT")
	cRank, cOk := testFindRank(ranked, "ROOT/c")

	if !rootOk {
		t.Error("ROOT not in ranked")
	}
	if !cOk {
		t.Error("ROOT/c not in ranked")
	}
	if rootRank != 0 {
		t.Errorf("expected ROOT rank 0, got %d", rootRank)
	}
	if cRank != 1 {
		t.Errorf("expected ROOT/c rank 1, got %d", cRank)
	}

	for _, c := range cycles {
		if c == "ROOT/c" {
			t.Errorf("ROOT/c should not appear in cycles")
		}
	}
}

func TestNodeRankCompute_UnresolvableRootReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/missing"}},
		},
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
		{LogicalName: "ROOT", Frontmatter: frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ARTIFACT/missing"}},
		},
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
		{LogicalName: "ROOT", Frontmatter: frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Input: "ARTIFACT/missing"},
		},
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
