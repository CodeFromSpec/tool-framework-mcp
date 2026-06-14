// code-from-spec: ROOT/golang/tests/utils/node_ranking@eQKNd1CzOHN2iB-APNcuyeaN06Q
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

func testRankOf(t *testing.T, ranked []*noderanking.NodeRankEntry, logicalName string) int {
	t.Helper()
	e := testFindEntry(ranked, logicalName)
	if e == nil {
		t.Fatalf("entry %q not found in ranked", logicalName)
	}
	return e.Rank
}

func testContainsCycle(cycles []string, logicalName string) bool {
	for _, c := range cycles {
		if c == logicalName {
			return true
		}
	}
	return false
}

func TestNodeRankCompute_TC01_RootOnly(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	if len(ranked) != 1 {
		t.Fatalf("expected 1 ranked entry, got %d", len(ranked))
	}
	if ranked[0].LogicalName != "SPEC" || ranked[0].Rank != 0 {
		t.Errorf("expected SPEC rank 0, got %+v", ranked[0])
	}
}

func TestNodeRankCompute_TC02_LinearChain(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a/b", Frontmatter: &frontmatter.Frontmatter{}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}

	if testRankOf(t, ranked, "SPEC") != 0 {
		t.Errorf("expected SPEC rank 0")
	}
	if testRankOf(t, ranked, "SPEC/a") != 1 {
		t.Errorf("expected SPEC/a rank 1")
	}
	if testRankOf(t, ranked, "SPEC/a/b") != 2 {
		t.Errorf("expected SPEC/a/b rank 2")
	}
}

func TestNodeRankCompute_TC03_IndependentSiblings(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/b", Frontmatter: &frontmatter.Frontmatter{}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}

	rankA := testRankOf(t, ranked, "SPEC/a")
	rankB := testRankOf(t, ranked, "SPEC/b")
	if rankA != rankB {
		t.Errorf("expected SPEC/a and SPEC/b to have equal rank, got %d and %d", rankA, rankB)
	}
	if rankA != 1 {
		t.Errorf("expected rank 1, got %d", rankA)
	}
}

func TestNodeRankCompute_TC04_DependsOnIncreasesRank(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/b", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/a"}}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}

	rankA := testRankOf(t, ranked, "SPEC/a")
	rankB := testRankOf(t, ranked, "SPEC/b")
	if rankB <= rankA {
		t.Errorf("expected rank of SPEC/b (%d) > rank of SPEC/a (%d)", rankB, rankA)
	}
}

func TestNodeRankCompute_TC05_DependsOnQualifierStripped(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/b", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/a(interface)"}}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}

	rankA := testRankOf(t, ranked, "SPEC/a")
	rankB := testRankOf(t, ranked, "SPEC/b")
	if rankB <= rankA {
		t.Errorf("expected rank of SPEC/b (%d) > rank of SPEC/a (%d)", rankB, rankA)
	}
}

func TestNodeRankCompute_TC06_ExternalDependsOnSkipped(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"EXTERNAL/proto/api.proto"}}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}

	rankA := testRankOf(t, ranked, "SPEC/a")
	if rankA != 1 {
		t.Errorf("expected SPEC/a rank 1, got %d", rankA)
	}
}

func TestNodeRankCompute_TC07_InputArtifactAddsDependencyEdge(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{Output: "out.go"}},
		{LogicalName: "SPEC/b", Frontmatter: &frontmatter.Frontmatter{Input: "ARTIFACT/a"}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}

	rankA := testRankOf(t, ranked, "SPEC/a")
	rankArtA := testRankOf(t, ranked, "ARTIFACT/a")
	rankB := testRankOf(t, ranked, "SPEC/b")

	if rankArtA <= rankA {
		t.Errorf("expected rank of ARTIFACT/a (%d) > rank of SPEC/a (%d)", rankArtA, rankA)
	}
	if rankB <= rankArtA {
		t.Errorf("expected rank of SPEC/b (%d) > rank of ARTIFACT/a (%d)", rankB, rankArtA)
	}
}

func TestNodeRankCompute_TC08_ExternalInputSkipped(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{Input: "EXTERNAL/docs/spec.yaml"}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}

	rankA := testRankOf(t, ranked, "SPEC/a")
	if rankA != 1 {
		t.Errorf("expected SPEC/a rank 1, got %d", rankA)
	}
}

func TestNodeRankCompute_TC09_ArtifactRankOneAboveNode(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{Output: "foo.go"}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}

	rankA := testRankOf(t, ranked, "SPEC/a")
	rankArtA := testRankOf(t, ranked, "ARTIFACT/a")
	if rankArtA != rankA+1 {
		t.Errorf("expected ARTIFACT/a rank = SPEC/a rank + 1, got SPEC/a=%d ARTIFACT/a=%d", rankA, rankArtA)
	}
}

func TestNodeRankCompute_TC10_SingleOutputArtifactRanked(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{Output: "x.go"}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}

	var artifactEntries []*noderanking.NodeRankEntry
	for _, e := range ranked {
		if len(e.LogicalName) > 9 && e.LogicalName[:9] == "ARTIFACT/" {
			artifactEntries = append(artifactEntries, e)
		}
	}
	if len(artifactEntries) != 1 {
		t.Fatalf("expected exactly 1 artifact entry, got %d", len(artifactEntries))
	}
	if artifactEntries[0].LogicalName != "ARTIFACT/a" {
		t.Errorf("expected ARTIFACT/a, got %s", artifactEntries[0].LogicalName)
	}

	rankA := testRankOf(t, ranked, "SPEC/a")
	if artifactEntries[0].Rank != rankA+1 {
		t.Errorf("expected ARTIFACT/a rank = SPEC/a rank + 1")
	}
}

func TestNodeRankCompute_TC11_DependsOnArtifactReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{Output: "lib.go"}},
		{LogicalName: "SPEC/b", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ARTIFACT/a"}}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}

	rankA := testRankOf(t, ranked, "SPEC/a")
	rankArtA := testRankOf(t, ranked, "ARTIFACT/a")
	rankB := testRankOf(t, ranked, "SPEC/b")

	if rankArtA <= rankA {
		t.Errorf("expected ARTIFACT/a rank (%d) > SPEC/a rank (%d)", rankArtA, rankA)
	}
	if rankB <= rankArtA {
		t.Errorf("expected SPEC/b rank (%d) > ARTIFACT/a rank (%d)", rankB, rankArtA)
	}
}

func TestNodeRankCompute_TC12_SortedByRankThenLogicalName(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/z", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}

	if ranked[0].LogicalName != "SPEC" || ranked[0].Rank != 0 {
		t.Errorf("expected first entry to be SPEC rank 0, got %+v", ranked[0])
	}

	idxA := -1
	idxZ := -1
	for i, e := range ranked {
		if e.LogicalName == "SPEC/a" {
			idxA = i
		}
		if e.LogicalName == "SPEC/z" {
			idxZ = i
		}
	}
	if idxA == -1 || idxZ == -1 {
		t.Fatal("expected both SPEC/a and SPEC/z in ranked")
	}
	if idxA >= idxZ {
		t.Errorf("expected SPEC/a (idx %d) to appear before SPEC/z (idx %d)", idxA, idxZ)
	}
}

func TestNodeRankCompute_TC13_ParallelEntriesEqualRank(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/b", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/c", Frontmatter: &frontmatter.Frontmatter{}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}

	rankA := testRankOf(t, ranked, "SPEC/a")
	rankB := testRankOf(t, ranked, "SPEC/b")
	rankC := testRankOf(t, ranked, "SPEC/c")

	if rankA != 1 || rankB != 1 || rankC != 1 {
		t.Errorf("expected all to have rank 1, got a=%d b=%d c=%d", rankA, rankB, rankC)
	}
}

func TestNodeRankCompute_TC14_DiamondDependency(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/c", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/c"}}},
		{LogicalName: "SPEC/b", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/c"}}},
		{LogicalName: "SPEC/d", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/a", "SPEC/b"}}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}

	rankC := testRankOf(t, ranked, "SPEC/c")
	rankA := testRankOf(t, ranked, "SPEC/a")
	rankB := testRankOf(t, ranked, "SPEC/b")
	rankD := testRankOf(t, ranked, "SPEC/d")

	if rankC != 1 {
		t.Errorf("expected SPEC/c rank 1, got %d", rankC)
	}
	if rankA != 2 {
		t.Errorf("expected SPEC/a rank 2, got %d", rankA)
	}
	if rankB != 2 {
		t.Errorf("expected SPEC/b rank 2, got %d", rankB)
	}
	if rankD != 3 {
		t.Errorf("expected SPEC/d rank 3, got %d", rankD)
	}
}

func TestNodeRankCompute_TC15_DependsOnOutranksParent(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/c", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/c/d", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/c/d/e", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a/b", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/c"}}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}

	rankA := testRankOf(t, ranked, "SPEC/a")
	rankAB := testRankOf(t, ranked, "SPEC/a/b")
	rankC := testRankOf(t, ranked, "SPEC/c")

	if rankAB <= rankA {
		t.Errorf("expected rank of SPEC/a/b (%d) > rank of SPEC/a (%d)", rankAB, rankA)
	}

	expectedRank := 1 + max(rankA, rankC)
	if rankAB != expectedRank {
		t.Errorf("expected SPEC/a/b rank = 1 + max(%d, %d) = %d, got %d", rankA, rankC, expectedRank, rankAB)
	}
}

func TestNodeRankCompute_TC16_MultipleDependsOnRankFromHighest(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/b", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/a"}}},
		{LogicalName: "SPEC/c", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/b"}}},
		{LogicalName: "SPEC/d", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/a", "SPEC/b", "SPEC/c"}}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}

	if testRankOf(t, ranked, "SPEC/a") != 1 {
		t.Errorf("expected SPEC/a rank 1")
	}
	if testRankOf(t, ranked, "SPEC/b") != 2 {
		t.Errorf("expected SPEC/b rank 2")
	}
	if testRankOf(t, ranked, "SPEC/c") != 3 {
		t.Errorf("expected SPEC/c rank 3")
	}
	if testRankOf(t, ranked, "SPEC/d") != 4 {
		t.Errorf("expected SPEC/d rank 4")
	}
}

func TestNodeRankCompute_TC17_NodeWithBothDependsOnAndInput(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{Output: "a.go"}},
		{LogicalName: "SPEC/b", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/c", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/b"}, Input: "ARTIFACT/a"}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}

	rankSpec := testRankOf(t, ranked, "SPEC")
	rankB := testRankOf(t, ranked, "SPEC/b")
	rankArtA := testRankOf(t, ranked, "ARTIFACT/a")
	rankC := testRankOf(t, ranked, "SPEC/c")

	expectedRank := 1 + max(rankSpec, max(rankB, rankArtA))
	if rankC != expectedRank {
		t.Errorf("expected SPEC/c rank = 1 + max(rank SPEC=%d, rank SPEC/b=%d, rank ARTIFACT/a=%d) = %d, got %d",
			rankSpec, rankB, rankArtA, expectedRank, rankC)
	}
}

func TestNodeRankCompute_TC18_EmptyInputList(t *testing.T) {
	entries := []*noderanking.NodeRankInput{}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	if len(ranked) != 0 {
		t.Errorf("expected empty ranked, got %v", ranked)
	}
}

func TestNodeRankCompute_TC19_SelfReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/a"}}},
	}

	_, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Error("expected cycles to be non-empty")
	}
}

func TestNodeRankCompute_TC20_SimpleCycleTwoNodes(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/b"}}},
		{LogicalName: "SPEC/b", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/a"}}},
	}

	_, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Error("expected cycles to be non-empty")
	}
	if !testContainsCycle(cycles, "SPEC/a") && !testContainsCycle(cycles, "SPEC/b") {
		t.Errorf("expected cycles to contain SPEC/a or SPEC/b, got %v", cycles)
	}
}

func TestNodeRankCompute_TC21_CycleThroughArtifacts(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{Output: "a.go", DependsOn: []string{"ARTIFACT/b"}}},
		{LogicalName: "SPEC/b", Frontmatter: &frontmatter.Frontmatter{Output: "b.go", DependsOn: []string{"ARTIFACT/a"}}},
	}

	_, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Error("expected cycles to be non-empty")
	}
}

func TestNodeRankCompute_TC22_CycleDoesNotPreventRankingUnrelated(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/b"}}},
		{LogicalName: "SPEC/b", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/a"}}},
		{LogicalName: "SPEC/c", Frontmatter: &frontmatter.Frontmatter{}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Error("expected cycles to be non-empty")
	}

	rankSpec := testRankOf(t, ranked, "SPEC")
	if rankSpec != 0 {
		t.Errorf("expected SPEC rank 0, got %d", rankSpec)
	}

	rankC := testRankOf(t, ranked, "SPEC/c")
	if rankC != 1 {
		t.Errorf("expected SPEC/c rank 1, got %d", rankC)
	}

	if testContainsCycle(cycles, "SPEC/c") {
		t.Error("SPEC/c should not be in cycles")
	}

	if !testContainsCycle(cycles, "SPEC/a") && !testContainsCycle(cycles, "SPEC/b") {
		t.Errorf("expected cycles to relate to SPEC/a and/or SPEC/b, got %v", cycles)
	}
}

func TestNodeRankCompute_TC23_UnresolvableSpecReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/missing"}}},
	}

	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got %v", err)
	}
}

func TestNodeRankCompute_TC24_UnresolvableArtifactReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ARTIFACT/missing"}}},
	}

	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got %v", err)
	}
}

func TestNodeRankCompute_TC25_UnresolvableInputReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{Input: "ARTIFACT/missing"}},
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
