// code-from-spec: SPEC/golang/tests/spec_tree/ranking@lzXowH88_BZoXGRzD4SpwaDrGKA
package noderanking_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/noderanking"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"
)

func strPtr(s string) *string {
	return &s
}

func makeNode(logicalName string, parentName *string, dependsOn []string, input *string, output *string) parsing.Node {
	ref := parsing.CfsReference{
		NodeType:    parsing.CfsNodeTypeSpec,
		LogicalName: logicalName,
		ParentName:  parentName,
	}
	var fm *parsing.NodeFrontmatter
	if dependsOn != nil || input != nil || output != nil {
		fm = &parsing.NodeFrontmatter{
			DependsOn: dependsOn,
			Input:     input,
			Output:    output,
		}
	}
	return parsing.Node{
		Reference:   ref,
		Frontmatter: fm,
	}
}

func testRankOf(ranked []noderanking.NodeRankEntry, logicalName string) (int, bool) {
	for _, e := range ranked {
		if e.Reference.LogicalName == logicalName {
			return e.Rank, true
		}
	}
	return 0, false
}

func testContains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func TestNodeRankCompute_RootOnly(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
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
	if ranked[0].Reference.LogicalName != "SPEC/root" || ranked[0].Rank != 0 {
		t.Errorf("expected SPEC/root rank 0, got %v", ranked[0])
	}
}

func TestNodeRankCompute_LinearChain(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), nil, nil, nil),
		makeNode("SPEC/root/a/b", strPtr("SPEC/root/a"), nil, nil, nil),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rRoot, ok := testRankOf(ranked, "SPEC/root")
	if !ok {
		t.Fatal("SPEC/root not found")
	}
	rA, ok := testRankOf(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	rAB, ok := testRankOf(ranked, "SPEC/root/a/b")
	if !ok {
		t.Fatal("SPEC/root/a/b not found")
	}

	if rRoot != 0 {
		t.Errorf("expected SPEC/root rank 0, got %d", rRoot)
	}
	if rA != 1 {
		t.Errorf("expected SPEC/root/a rank 1, got %d", rA)
	}
	if rAB != 2 {
		t.Errorf("expected SPEC/root/a/b rank 2, got %d", rAB)
	}
}

func TestNodeRankCompute_IndependentSiblings(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), nil, nil, nil),
		makeNode("SPEC/root/b", strPtr("SPEC/root"), nil, nil, nil),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rA, ok := testRankOf(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	rB, ok := testRankOf(ranked, "SPEC/root/b")
	if !ok {
		t.Fatal("SPEC/root/b not found")
	}
	if rA != rB {
		t.Errorf("expected equal ranks for siblings, got SPEC/root/a=%d SPEC/root/b=%d", rA, rB)
	}
	if rA != 1 {
		t.Errorf("expected rank 1 for siblings, got %d", rA)
	}
}

func TestNodeRankCompute_MultipleIndependentRoots(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/alpha", nil, nil, nil, nil),
		makeNode("SPEC/beta", nil, nil, nil, nil),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rAlpha, ok := testRankOf(ranked, "SPEC/alpha")
	if !ok {
		t.Fatal("SPEC/alpha not found")
	}
	rBeta, ok := testRankOf(ranked, "SPEC/beta")
	if !ok {
		t.Fatal("SPEC/beta not found")
	}
	if rAlpha != 0 {
		t.Errorf("expected SPEC/alpha rank 0, got %d", rAlpha)
	}
	if rBeta != 0 {
		t.Errorf("expected SPEC/beta rank 0, got %d", rBeta)
	}
}

func TestNodeRankCompute_DependsOnIncreasesRank(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), nil, nil, nil),
		makeNode("SPEC/root/b", strPtr("SPEC/root"), []string{"SPEC/root/a"}, nil, nil),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rA, ok := testRankOf(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	rB, ok := testRankOf(ranked, "SPEC/root/b")
	if !ok {
		t.Fatal("SPEC/root/b not found")
	}
	if rB <= rA {
		t.Errorf("expected rank of SPEC/root/b (%d) > rank of SPEC/root/a (%d)", rB, rA)
	}
}

func TestNodeRankCompute_DependsOnWithQualifier(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), nil, nil, nil),
		makeNode("SPEC/root/b", strPtr("SPEC/root"), []string{"SPEC/root/a(interface)"}, nil, nil),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rA, ok := testRankOf(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	rB, ok := testRankOf(ranked, "SPEC/root/b")
	if !ok {
		t.Fatal("SPEC/root/b not found")
	}
	if rB <= rA {
		t.Errorf("expected rank of SPEC/root/b (%d) > rank of SPEC/root/a (%d)", rB, rA)
	}
}

func TestNodeRankCompute_ExternalDependsOnSkipped(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), []string{"EXTERNAL/proto/api.proto"}, nil, nil),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rA, ok := testRankOf(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	if rA != 1 {
		t.Errorf("expected SPEC/root/a rank 1, got %d", rA)
	}
}

func TestNodeRankCompute_InputArtifactAddsDependencyEdge(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), nil, nil, strPtr("out.go")),
		makeNode("SPEC/root/b", strPtr("SPEC/root"), nil, strPtr("ARTIFACT/root/a"), nil),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rA, ok := testRankOf(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	rArtA, ok := testRankOf(ranked, "ARTIFACT/root/a")
	if !ok {
		t.Fatal("ARTIFACT/root/a not found")
	}
	rB, ok := testRankOf(ranked, "SPEC/root/b")
	if !ok {
		t.Fatal("SPEC/root/b not found")
	}

	if rArtA <= rA {
		t.Errorf("expected rank of ARTIFACT/root/a (%d) > rank of SPEC/root/a (%d)", rArtA, rA)
	}
	if rB <= rArtA {
		t.Errorf("expected rank of SPEC/root/b (%d) > rank of ARTIFACT/root/a (%d)", rB, rArtA)
	}
}

func TestNodeRankCompute_SpecInputAddsDependencyEdge(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), nil, nil, nil),
		makeNode("SPEC/root/b", strPtr("SPEC/root"), nil, strPtr("SPEC/root/a"), nil),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rA, ok := testRankOf(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	rB, ok := testRankOf(ranked, "SPEC/root/b")
	if !ok {
		t.Fatal("SPEC/root/b not found")
	}
	if rB <= rA {
		t.Errorf("expected rank of SPEC/root/b (%d) > rank of SPEC/root/a (%d)", rB, rA)
	}
}

func TestNodeRankCompute_ExternalInputSkipped(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), nil, strPtr("EXTERNAL/docs/spec.yaml"), nil),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rA, ok := testRankOf(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	if rA != 1 {
		t.Errorf("expected SPEC/root/a rank 1, got %d", rA)
	}
}

func TestNodeRankCompute_ArtifactsRankedOneAboveNode(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), nil, nil, strPtr("foo.go")),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rA, ok := testRankOf(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	rArt, ok := testRankOf(ranked, "ARTIFACT/root/a")
	if !ok {
		t.Fatal("ARTIFACT/root/a not found")
	}
	if rArt != rA+1 {
		t.Errorf("expected ARTIFACT/root/a rank %d, got %d", rA+1, rArt)
	}
}

func TestNodeRankCompute_SingleOutputArtifactRanked(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), nil, nil, strPtr("x.go")),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rA, ok := testRankOf(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	rArt, ok := testRankOf(ranked, "ARTIFACT/root/a")
	if !ok {
		t.Fatal("ARTIFACT/root/a not in ranked")
	}
	if rArt != rA+1 {
		t.Errorf("expected ARTIFACT/root/a rank %d, got %d", rA+1, rArt)
	}
}

func TestNodeRankCompute_DependsOnArtifactReference(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), nil, nil, strPtr("lib.go")),
		makeNode("SPEC/root/b", strPtr("SPEC/root"), []string{"ARTIFACT/root/a"}, nil, nil),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rA, ok := testRankOf(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	rArtA, ok := testRankOf(ranked, "ARTIFACT/root/a")
	if !ok {
		t.Fatal("ARTIFACT/root/a not found")
	}
	rB, ok := testRankOf(ranked, "SPEC/root/b")
	if !ok {
		t.Fatal("SPEC/root/b not found")
	}

	if rArtA <= rA {
		t.Errorf("expected ARTIFACT/root/a (%d) > SPEC/root/a (%d)", rArtA, rA)
	}
	if rB <= rArtA {
		t.Errorf("expected SPEC/root/b (%d) > ARTIFACT/root/a (%d)", rB, rArtA)
	}
}

func TestNodeRankCompute_OutputSortedByRankThenName(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/z", strPtr("SPEC/root"), nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), nil, nil, nil),
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

	if ranked[0].Reference.LogicalName != "SPEC/root" || ranked[0].Rank != 0 {
		t.Errorf("ranked[0] expected SPEC/root rank 0, got %v", ranked[0])
	}
	if ranked[1].Reference.LogicalName != "SPEC/root/a" || ranked[1].Rank != 1 {
		t.Errorf("ranked[1] expected SPEC/root/a rank 1, got %v", ranked[1])
	}
	if ranked[2].Reference.LogicalName != "SPEC/root/z" || ranked[2].Rank != 1 {
		t.Errorf("ranked[2] expected SPEC/root/z rank 1, got %v", ranked[2])
	}
}

func TestNodeRankCompute_ParallelEntriesEqualRank(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), nil, nil, nil),
		makeNode("SPEC/root/b", strPtr("SPEC/root"), nil, nil, nil),
		makeNode("SPEC/root/c", strPtr("SPEC/root"), nil, nil, nil),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	for _, name := range []string{"SPEC/root/a", "SPEC/root/b", "SPEC/root/c"} {
		r, ok := testRankOf(ranked, name)
		if !ok {
			t.Fatalf("%s not found", name)
		}
		if r != 1 {
			t.Errorf("expected %s rank 1, got %d", name, r)
		}
	}
}

func TestNodeRankCompute_DiamondDependency(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/c", strPtr("SPEC/root"), nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), []string{"SPEC/root/c"}, nil, nil),
		makeNode("SPEC/root/b", strPtr("SPEC/root"), []string{"SPEC/root/c"}, nil, nil),
		makeNode("SPEC/root/d", strPtr("SPEC/root"), []string{"SPEC/root/a", "SPEC/root/b"}, nil, nil),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	expected := map[string]int{
		"SPEC/root/c": 1,
		"SPEC/root/a": 2,
		"SPEC/root/b": 2,
		"SPEC/root/d": 3,
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

func TestNodeRankCompute_DependsOnOutranksParent(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), nil, nil, nil),
		makeNode("SPEC/root/a/b", strPtr("SPEC/root/a"), []string{"SPEC/root/c"}, nil, nil),
		makeNode("SPEC/root/c", strPtr("SPEC/root"), nil, nil, nil),
		makeNode("SPEC/root/c/d", strPtr("SPEC/root/c"), nil, nil, nil),
		makeNode("SPEC/root/c/d/e", strPtr("SPEC/root/c/d"), nil, nil, nil),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rA, ok := testRankOf(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	rAB, ok := testRankOf(ranked, "SPEC/root/a/b")
	if !ok {
		t.Fatal("SPEC/root/a/b not found")
	}
	rC, ok := testRankOf(ranked, "SPEC/root/c")
	if !ok {
		t.Fatal("SPEC/root/c not found")
	}

	if rAB <= rA {
		t.Errorf("expected rank of SPEC/root/a/b (%d) > rank of SPEC/root/a (%d)", rAB, rA)
	}

	maxParentDep := rA
	if rC > maxParentDep {
		maxParentDep = rC
	}
	if rAB != 1+maxParentDep {
		t.Errorf("expected SPEC/root/a/b rank %d (1+max(%d,%d)), got %d", 1+maxParentDep, rA, rC, rAB)
	}
}

func TestNodeRankCompute_MultipleDependsOnRankFromHighest(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), nil, nil, nil),
		makeNode("SPEC/root/b", strPtr("SPEC/root"), []string{"SPEC/root/a"}, nil, nil),
		makeNode("SPEC/root/c", strPtr("SPEC/root"), []string{"SPEC/root/b"}, nil, nil),
		makeNode("SPEC/root/d", strPtr("SPEC/root"), []string{"SPEC/root/a", "SPEC/root/b", "SPEC/root/c"}, nil, nil),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	expected := map[string]int{
		"SPEC/root/a": 1,
		"SPEC/root/b": 2,
		"SPEC/root/c": 3,
		"SPEC/root/d": 4,
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

func TestNodeRankCompute_NodeWithBothDependsOnAndInput(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), nil, nil, strPtr("a.go")),
		makeNode("SPEC/root/b", strPtr("SPEC/root"), nil, nil, nil),
		makeNode("SPEC/root/c", strPtr("SPEC/root"), []string{"SPEC/root/b"}, strPtr("ARTIFACT/root/a"), nil),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rRoot, ok := testRankOf(ranked, "SPEC/root")
	if !ok {
		t.Fatal("SPEC/root not found")
	}
	rB, ok := testRankOf(ranked, "SPEC/root/b")
	if !ok {
		t.Fatal("SPEC/root/b not found")
	}
	rArtA, ok := testRankOf(ranked, "ARTIFACT/root/a")
	if !ok {
		t.Fatal("ARTIFACT/root/a not found")
	}
	rC, ok := testRankOf(ranked, "SPEC/root/c")
	if !ok {
		t.Fatal("SPEC/root/c not found")
	}

	maxDep := rRoot
	if rB > maxDep {
		maxDep = rB
	}
	if rArtA > maxDep {
		maxDep = rArtA
	}
	if rC != 1+maxDep {
		t.Errorf("expected SPEC/root/c rank %d, got %d", 1+maxDep, rC)
	}
}

func TestNodeRankCompute_EmptyInputList(t *testing.T) {
	ranked, cycles, err := noderanking.NodeRankCompute([]parsing.Node{})
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
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), []string{"SPEC/root/a"}, nil, nil),
	}

	_, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Error("expected cycles to be non-empty")
	}
}

func TestNodeRankCompute_SimpleCycleTwoNodes(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), []string{"SPEC/root/b"}, nil, nil),
		makeNode("SPEC/root/b", strPtr("SPEC/root"), []string{"SPEC/root/a"}, nil, nil),
	}

	_, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Error("expected cycles to be non-empty")
	}

	hasA := testContains(cycles, "SPEC/root/a")
	hasB := testContains(cycles, "SPEC/root/b")
	if !hasA && !hasB {
		t.Errorf("expected cycles to contain SPEC/root/a or SPEC/root/b, got %v", cycles)
	}
}

func TestNodeRankCompute_CycleThroughArtifacts(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), []string{"ARTIFACT/root/b"}, nil, strPtr("a.go")),
		makeNode("SPEC/root/b", strPtr("SPEC/root"), []string{"ARTIFACT/root/a"}, nil, strPtr("b.go")),
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
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), []string{"SPEC/root/b"}, nil, nil),
		makeNode("SPEC/root/b", strPtr("SPEC/root"), []string{"SPEC/root/a"}, nil, nil),
		makeNode("SPEC/root/c", strPtr("SPEC/root"), nil, nil, nil),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Error("expected cycles to be non-empty")
	}

	rRoot, ok := testRankOf(ranked, "SPEC/root")
	if !ok {
		t.Fatal("SPEC/root not found")
	}
	if rRoot != 0 {
		t.Errorf("expected SPEC/root rank 0, got %d", rRoot)
	}

	rC, ok := testRankOf(ranked, "SPEC/root/c")
	if !ok {
		t.Fatal("SPEC/root/c not found")
	}
	if rC != 1 {
		t.Errorf("expected SPEC/root/c rank 1, got %d", rC)
	}

	if testContains(cycles, "SPEC/root/c") {
		t.Errorf("expected SPEC/root/c not in cycles, but got %v", cycles)
	}

	hasA := testContains(cycles, "SPEC/root/a")
	hasB := testContains(cycles, "SPEC/root/b")
	if !hasA && !hasB {
		t.Errorf("expected cycles to contain SPEC/root/a and/or SPEC/root/b, got %v", cycles)
	}
}

func TestNodeRankCompute_UnresolvableSpecReference(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), []string{"SPEC/root/missing"}, nil, nil),
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
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), []string{"ARTIFACT/root/missing"}, nil, nil),
	}

	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got %v", err)
	}
}

func TestNodeRankCompute_UnresolvableArtifactInputReference(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), nil, strPtr("ARTIFACT/root/missing"), nil),
	}

	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got %v", err)
	}
}

func TestNodeRankCompute_UnresolvableSpecInputReference(t *testing.T) {
	entries := []parsing.Node{
		makeNode("SPEC/root", nil, nil, nil, nil),
		makeNode("SPEC/root/a", strPtr("SPEC/root"), nil, strPtr("SPEC/root/missing"), nil),
	}

	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got %v", err)
	}
}
