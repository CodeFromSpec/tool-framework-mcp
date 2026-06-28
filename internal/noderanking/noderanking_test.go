// code-from-spec: SPEC/golang/tests/utils/node_ranking@QThGVn3QZABylwrphseD5oKYZcs
package noderanking_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/noderanking"
)

func testFindEntry(ranked []*noderanking.NodeRankEntry, logicalName string) *noderanking.NodeRankEntry {
	for _, e := range ranked {
		if e.LogicalName == logicalName {
			return e
		}
	}
	return nil
}

func testRankOf(ranked []*noderanking.NodeRankEntry, logicalName string) (int, bool) {
	e := testFindEntry(ranked, logicalName)
	if e == nil {
		return 0, false
	}
	return e.Rank, true
}

func testContains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
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
		t.Errorf("expected no cycles, got %v", cycles)
	}
	if len(ranked) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(ranked))
	}
	if ranked[0].LogicalName != "SPEC" || ranked[0].Rank != 0 {
		t.Errorf("expected SPEC rank 0, got %v", ranked[0])
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
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rSpec, ok := testRankOf(ranked, "SPEC")
	if !ok {
		t.Fatal("SPEC not found")
	}
	rA, ok := testRankOf(ranked, "SPEC/a")
	if !ok {
		t.Fatal("SPEC/a not found")
	}
	rAB, ok := testRankOf(ranked, "SPEC/a/b")
	if !ok {
		t.Fatal("SPEC/a/b not found")
	}

	if rSpec != 0 {
		t.Errorf("expected SPEC rank 0, got %d", rSpec)
	}
	if rA != 1 {
		t.Errorf("expected SPEC/a rank 1, got %d", rA)
	}
	if rAB != 2 {
		t.Errorf("expected SPEC/a/b rank 2, got %d", rAB)
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
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rA, ok := testRankOf(ranked, "SPEC/a")
	if !ok {
		t.Fatal("SPEC/a not found")
	}
	rB, ok := testRankOf(ranked, "SPEC/b")
	if !ok {
		t.Fatal("SPEC/b not found")
	}
	if rA != rB {
		t.Errorf("expected equal ranks for siblings, got SPEC/a=%d SPEC/b=%d", rA, rB)
	}
	if rA != 1 {
		t.Errorf("expected rank 1 for siblings, got %d", rA)
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
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rA, ok := testRankOf(ranked, "SPEC/a")
	if !ok {
		t.Fatal("SPEC/a not found")
	}
	rB, ok := testRankOf(ranked, "SPEC/b")
	if !ok {
		t.Fatal("SPEC/b not found")
	}
	if rB <= rA {
		t.Errorf("expected rank of SPEC/b (%d) > rank of SPEC/a (%d)", rB, rA)
	}
}

func TestNodeRankCompute_TC05_DependsOnWithQualifier(t *testing.T) {
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
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rA, ok := testRankOf(ranked, "SPEC/a")
	if !ok {
		t.Fatal("SPEC/a not found")
	}
	rB, ok := testRankOf(ranked, "SPEC/b")
	if !ok {
		t.Fatal("SPEC/b not found")
	}
	if rB <= rA {
		t.Errorf("expected rank of SPEC/b (%d) > rank of SPEC/a (%d)", rB, rA)
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
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rA, ok := testRankOf(ranked, "SPEC/a")
	if !ok {
		t.Fatal("SPEC/a not found")
	}
	if rA != 1 {
		t.Errorf("expected SPEC/a rank 1, got %d", rA)
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
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rA, ok := testRankOf(ranked, "SPEC/a")
	if !ok {
		t.Fatal("SPEC/a not found")
	}
	rArtA, ok := testRankOf(ranked, "ARTIFACT/a")
	if !ok {
		t.Fatal("ARTIFACT/a not found")
	}
	rB, ok := testRankOf(ranked, "SPEC/b")
	if !ok {
		t.Fatal("SPEC/b not found")
	}

	if rArtA <= rA {
		t.Errorf("expected rank of ARTIFACT/a (%d) > rank of SPEC/a (%d)", rArtA, rA)
	}
	if rB <= rArtA {
		t.Errorf("expected rank of SPEC/b (%d) > rank of ARTIFACT/a (%d)", rB, rArtA)
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
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rA, ok := testRankOf(ranked, "SPEC/a")
	if !ok {
		t.Fatal("SPEC/a not found")
	}
	if rA != 1 {
		t.Errorf("expected SPEC/a rank 1, got %d", rA)
	}
}

func TestNodeRankCompute_TC09_ArtifactsRankedOneAboveNode(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{Output: "foo.go"}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rA, ok := testRankOf(ranked, "SPEC/a")
	if !ok {
		t.Fatal("SPEC/a not found")
	}
	rArt, ok := testRankOf(ranked, "ARTIFACT/a")
	if !ok {
		t.Fatal("ARTIFACT/a not found")
	}
	if rArt != rA+1 {
		t.Errorf("expected ARTIFACT/a rank %d, got %d", rA+1, rArt)
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
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rA, ok := testRankOf(ranked, "SPEC/a")
	if !ok {
		t.Fatal("SPEC/a not found")
	}
	rArt, ok := testRankOf(ranked, "ARTIFACT/a")
	if !ok {
		t.Fatal("ARTIFACT/a not in ranked")
	}
	if rArt != rA+1 {
		t.Errorf("expected ARTIFACT/a rank %d, got %d", rA+1, rArt)
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
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rA, ok := testRankOf(ranked, "SPEC/a")
	if !ok {
		t.Fatal("SPEC/a not found")
	}
	rArtA, ok := testRankOf(ranked, "ARTIFACT/a")
	if !ok {
		t.Fatal("ARTIFACT/a not found")
	}
	rB, ok := testRankOf(ranked, "SPEC/b")
	if !ok {
		t.Fatal("SPEC/b not found")
	}

	if rArtA <= rA {
		t.Errorf("expected ARTIFACT/a (%d) > SPEC/a (%d)", rArtA, rA)
	}
	if rB <= rArtA {
		t.Errorf("expected SPEC/b (%d) > ARTIFACT/a (%d)", rB, rArtA)
	}
}

func TestNodeRankCompute_TC12_OutputSortedByRankThenName(t *testing.T) {
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
		t.Errorf("expected no cycles, got %v", cycles)
	}
	if len(ranked) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(ranked))
	}

	if ranked[0].LogicalName != "SPEC" || ranked[0].Rank != 0 {
		t.Errorf("ranked[0] expected SPEC rank 0, got %v", ranked[0])
	}
	if ranked[1].LogicalName != "SPEC/a" || ranked[1].Rank != 1 {
		t.Errorf("ranked[1] expected SPEC/a rank 1, got %v", ranked[1])
	}
	if ranked[2].LogicalName != "SPEC/z" || ranked[2].Rank != 1 {
		t.Errorf("ranked[2] expected SPEC/z rank 1, got %v", ranked[2])
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
		t.Errorf("expected no cycles, got %v", cycles)
	}

	for _, name := range []string{"SPEC/a", "SPEC/b", "SPEC/c"} {
		r, ok := testRankOf(ranked, name)
		if !ok {
			t.Fatalf("%s not found", name)
		}
		if r != 1 {
			t.Errorf("expected %s rank 1, got %d", name, r)
		}
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
		t.Errorf("expected no cycles, got %v", cycles)
	}

	expected := map[string]int{
		"SPEC/c": 1,
		"SPEC/a": 2,
		"SPEC/b": 2,
		"SPEC/d": 3,
	}
	for name, want := range expected {
		got, ok := testRankOf(ranked, name)
		if !ok {
			t.Fatalf("%s not found", name)
		}
		if got != want {
			t.Errorf("expected %s rank %d, got %d", name, want, got)
		}
	}
}

func TestNodeRankCompute_TC15_DependsOnOutranksParent(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "SPEC", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/a/b", Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/c"}}},
		{LogicalName: "SPEC/c", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/c/d", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "SPEC/c/d/e", Frontmatter: &frontmatter.Frontmatter{}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rA, ok := testRankOf(ranked, "SPEC/a")
	if !ok {
		t.Fatal("SPEC/a not found")
	}
	rAB, ok := testRankOf(ranked, "SPEC/a/b")
	if !ok {
		t.Fatal("SPEC/a/b not found")
	}
	rC, ok := testRankOf(ranked, "SPEC/c")
	if !ok {
		t.Fatal("SPEC/c not found")
	}

	if rAB <= rA {
		t.Errorf("expected rank of SPEC/a/b (%d) > rank of SPEC/a (%d)", rAB, rA)
	}

	maxParentDep := rA
	if rC > maxParentDep {
		maxParentDep = rC
	}
	if rAB != 1+maxParentDep {
		t.Errorf("expected SPEC/a/b rank %d (1+max(%d,%d)), got %d", 1+maxParentDep, rA, rC, rAB)
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
		t.Errorf("expected no cycles, got %v", cycles)
	}

	expected := map[string]int{
		"SPEC/a": 1,
		"SPEC/b": 2,
		"SPEC/c": 3,
		"SPEC/d": 4,
	}
	for name, want := range expected {
		got, ok := testRankOf(ranked, name)
		if !ok {
			t.Fatalf("%s not found", name)
		}
		if got != want {
			t.Errorf("expected %s rank %d, got %d", name, want, got)
		}
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
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rSpec, ok := testRankOf(ranked, "SPEC")
	if !ok {
		t.Fatal("SPEC not found")
	}
	rB, ok := testRankOf(ranked, "SPEC/b")
	if !ok {
		t.Fatal("SPEC/b not found")
	}
	rArtA, ok := testRankOf(ranked, "ARTIFACT/a")
	if !ok {
		t.Fatal("ARTIFACT/a not found")
	}
	rC, ok := testRankOf(ranked, "SPEC/c")
	if !ok {
		t.Fatal("SPEC/c not found")
	}

	maxDep := rSpec
	if rB > maxDep {
		maxDep = rB
	}
	if rArtA > maxDep {
		maxDep = rArtA
	}
	if rC != 1+maxDep {
		t.Errorf("expected SPEC/c rank %d, got %d", 1+maxDep, rC)
	}
}

func TestNodeRankCompute_TC18_EmptyInputList(t *testing.T) {
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

	hasA := testContains(cycles, "SPEC/a")
	hasB := testContains(cycles, "SPEC/b")
	if !hasA && !hasB {
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

	rSpec, ok := testRankOf(ranked, "SPEC")
	if !ok {
		t.Fatal("SPEC not found")
	}
	if rSpec != 0 {
		t.Errorf("expected SPEC rank 0, got %d", rSpec)
	}

	rC, ok := testRankOf(ranked, "SPEC/c")
	if !ok {
		t.Fatal("SPEC/c not found")
	}
	if rC != 1 {
		t.Errorf("expected SPEC/c rank 1, got %d", rC)
	}

	if testContains(cycles, "SPEC/c") {
		t.Errorf("expected SPEC/c not in cycles, but got %v", cycles)
	}

	hasA := testContains(cycles, "SPEC/a")
	hasB := testContains(cycles, "SPEC/b")
	if !hasA && !hasB {
		t.Errorf("expected cycles to contain SPEC/a and/or SPEC/b, got %v", cycles)
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
