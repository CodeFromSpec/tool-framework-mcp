// code-from-spec: ROOT/golang/tests/utils/node_ranking@GuO95nhFSHrOLGmiIplaG3TOFtg
package noderanking_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/noderanking"
)

// testMakeInput builds a NodeRankInput with the given logical name and frontmatter.
func testMakeInput(logicalName string, fm *frontmatter.Frontmatter) *noderanking.NodeRankInput {
	return &noderanking.NodeRankInput{
		LogicalName: logicalName,
		Frontmatter: fm,
	}
}

// testEmptyFM returns an empty Frontmatter.
func testEmptyFM() *frontmatter.Frontmatter {
	return &frontmatter.Frontmatter{}
}

// testFMWithDependsOn returns a Frontmatter with the given depends_on entries.
func testFMWithDependsOn(deps []string) *frontmatter.Frontmatter {
	return &frontmatter.Frontmatter{
		DependsOn: deps,
	}
}

// testFMWithOutputs returns a Frontmatter with the given outputs.
func testFMWithOutputs(outputs []*frontmatter.FrontmatterOutput) *frontmatter.Frontmatter {
	return &frontmatter.Frontmatter{
		Outputs: outputs,
	}
}

// testFMWithInput returns a Frontmatter with the given input artifact.
func testFMWithInput(input string) *frontmatter.Frontmatter {
	return &frontmatter.Frontmatter{
		Input: input,
	}
}

// testFMWithDepsAndInput returns a Frontmatter with both depends_on and input set.
func testFMWithDepsAndInput(deps []string, input string) *frontmatter.Frontmatter {
	return &frontmatter.Frontmatter{
		DependsOn: deps,
		Input:     input,
	}
}

// testFMWithOutputsAndDeps returns a Frontmatter with outputs and depends_on set.
func testFMWithOutputsAndDeps(outputs []*frontmatter.FrontmatterOutput, deps []string) *frontmatter.Frontmatter {
	return &frontmatter.Frontmatter{
		Outputs:   outputs,
		DependsOn: deps,
	}
}

// testRankOf finds the rank for a given logical name in the ranked list.
// Returns -1 if not found.
func testRankOf(ranked []*noderanking.NodeRankEntry, logicalName string) int {
	for _, e := range ranked {
		if e.LogicalName == logicalName {
			return e.Rank
		}
	}
	return -1
}

// testContains checks if a string slice contains a value.
func testContains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

// TC-01: Root only
func TestNodeRankCompute_RootOnly(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
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
	if ranked[0].LogicalName != "ROOT" {
		t.Errorf("expected ROOT, got %s", ranked[0].LogicalName)
	}
	if ranked[0].Rank != 0 {
		t.Errorf("expected rank 0, got %d", ranked[0].Rank)
	}
}

// TC-02: Linear chain — incrementing ranks
func TestNodeRankCompute_LinearChain(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testEmptyFM()),
		testMakeInput("ROOT/a/b", testEmptyFM()),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	if r := testRankOf(ranked, "ROOT"); r != 0 {
		t.Errorf("ROOT: expected rank 0, got %d", r)
	}
	if r := testRankOf(ranked, "ROOT/a"); r != 1 {
		t.Errorf("ROOT/a: expected rank 1, got %d", r)
	}
	if r := testRankOf(ranked, "ROOT/a/b"); r != 2 {
		t.Errorf("ROOT/a/b: expected rank 2, got %d", r)
	}
}

// TC-03: Independent siblings — equal rank
func TestNodeRankCompute_IndependentSiblings(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testEmptyFM()),
		testMakeInput("ROOT/b", testEmptyFM()),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rankA := testRankOf(ranked, "ROOT/a")
	rankB := testRankOf(ranked, "ROOT/b")
	if rankA != rankB {
		t.Errorf("expected ROOT/a (rank %d) == ROOT/b (rank %d)", rankA, rankB)
	}
	if rankA != 1 {
		t.Errorf("expected rank 1, got %d", rankA)
	}
}

// TC-04: depends_on increases rank
func TestNodeRankCompute_DependsOnIncreasesRank(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testEmptyFM()),
		testMakeInput("ROOT/b", testFMWithDependsOn([]string{"ROOT/a"})),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rankA := testRankOf(ranked, "ROOT/a")
	rankB := testRankOf(ranked, "ROOT/b")
	if rankB <= rankA {
		t.Errorf("expected rank of ROOT/b (%d) > rank of ROOT/a (%d)", rankB, rankA)
	}
}

// TC-05: depends_on with qualifier — qualifier stripped
func TestNodeRankCompute_DependsOnQualifierStripped(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testEmptyFM()),
		testMakeInput("ROOT/b", testFMWithDependsOn([]string{"ROOT/a(interface)"})),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rankA := testRankOf(ranked, "ROOT/a")
	rankB := testRankOf(ranked, "ROOT/b")
	if rankB <= rankA {
		t.Errorf("expected rank of ROOT/b (%d) > rank of ROOT/a (%d)", rankB, rankA)
	}
}

// TC-06: input artifact adds dependency edge
func TestNodeRankCompute_InputArtifactAddsEdge(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testFMWithOutputs([]*frontmatter.FrontmatterOutput{
			{ID: "code", Path: "out.go"},
		})),
		testMakeInput("ROOT/b", testFMWithInput("ARTIFACT/a(code)")),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rankNodeA := testRankOf(ranked, "ROOT/a")
	rankArtifact := testRankOf(ranked, "ARTIFACT/a(code)")
	rankNodeB := testRankOf(ranked, "ROOT/b")

	if rankArtifact < 0 {
		t.Fatal("expected ARTIFACT/a(code) in ranked list")
	}
	if rankArtifact <= rankNodeA {
		t.Errorf("expected rank of ARTIFACT/a(code) (%d) > rank of ROOT/a (%d)", rankArtifact, rankNodeA)
	}
	if rankNodeB <= rankArtifact {
		t.Errorf("expected rank of ROOT/b (%d) > rank of ARTIFACT/a(code) (%d)", rankNodeB, rankArtifact)
	}
}

// TC-07: Artifacts get rank one above their node
func TestNodeRankCompute_ArtifactRankOneAboveNode(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testFMWithOutputs([]*frontmatter.FrontmatterOutput{
			{ID: "foo", Path: "foo.go"},
		})),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rankNodeA := testRankOf(ranked, "ROOT/a")
	rankArtifact := testRankOf(ranked, "ARTIFACT/a(foo)")

	if rankArtifact < 0 {
		t.Fatal("expected ARTIFACT/a(foo) in ranked list")
	}
	if rankArtifact != rankNodeA+1 {
		t.Errorf("expected rank of ARTIFACT/a(foo) = %d, got %d", rankNodeA+1, rankArtifact)
	}
}

// TC-08: Multiple outputs — each artifact ranked
func TestNodeRankCompute_MultipleOutputsRanked(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testFMWithOutputs([]*frontmatter.FrontmatterOutput{
			{ID: "x", Path: "x.go"},
			{ID: "y", Path: "y.go"},
		})),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rankNodeA := testRankOf(ranked, "ROOT/a")
	rankX := testRankOf(ranked, "ARTIFACT/a(x)")
	rankY := testRankOf(ranked, "ARTIFACT/a(y)")

	if rankX < 0 {
		t.Fatal("expected ARTIFACT/a(x) in ranked list")
	}
	if rankY < 0 {
		t.Fatal("expected ARTIFACT/a(y) in ranked list")
	}
	if rankX != rankNodeA+1 {
		t.Errorf("expected rank of ARTIFACT/a(x) = %d, got %d", rankNodeA+1, rankX)
	}
	if rankY != rankNodeA+1 {
		t.Errorf("expected rank of ARTIFACT/a(y) = %d, got %d", rankNodeA+1, rankY)
	}
}

// TC-09: depends_on ARTIFACT reference — used as-is
func TestNodeRankCompute_DependsOnArtifactReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testFMWithOutputs([]*frontmatter.FrontmatterOutput{
			{ID: "lib", Path: "lib.go"},
		})),
		testMakeInput("ROOT/b", testFMWithDependsOn([]string{"ARTIFACT/a(lib)"})),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rankNodeA := testRankOf(ranked, "ROOT/a")
	rankArtifact := testRankOf(ranked, "ARTIFACT/a(lib)")
	rankNodeB := testRankOf(ranked, "ROOT/b")

	if rankArtifact <= rankNodeA {
		t.Errorf("expected rank of ARTIFACT/a(lib) (%d) > rank of ROOT/a (%d)", rankArtifact, rankNodeA)
	}
	if rankNodeB <= rankArtifact {
		t.Errorf("expected rank of ROOT/b (%d) > rank of ARTIFACT/a(lib) (%d)", rankNodeB, rankArtifact)
	}
}

// TC-10: Output sorted by rank then logical name
func TestNodeRankCompute_SortedByRankThenName(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/z", testEmptyFM()),
		testMakeInput("ROOT/a", testEmptyFM()),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	if len(ranked) == 0 {
		t.Fatal("ranked is empty")
	}
	if ranked[0].LogicalName != "ROOT" {
		t.Errorf("expected ROOT first, got %s", ranked[0].LogicalName)
	}

	// Find positions of ROOT/a and ROOT/z
	posA, posZ := -1, -1
	for i, e := range ranked {
		if e.LogicalName == "ROOT/a" {
			posA = i
		}
		if e.LogicalName == "ROOT/z" {
			posZ = i
		}
	}
	if posA < 0 {
		t.Fatal("ROOT/a not found in ranked")
	}
	if posZ < 0 {
		t.Fatal("ROOT/z not found in ranked")
	}
	if posA >= posZ {
		t.Errorf("expected ROOT/a (pos %d) before ROOT/z (pos %d)", posA, posZ)
	}
}

// TC-11: Parallel entries — equal rank means no dependency
func TestNodeRankCompute_ParallelEntriesEqualRank(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testEmptyFM()),
		testMakeInput("ROOT/b", testEmptyFM()),
		testMakeInput("ROOT/c", testEmptyFM()),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	for _, name := range []string{"ROOT/a", "ROOT/b", "ROOT/c"} {
		if r := testRankOf(ranked, name); r != 1 {
			t.Errorf("%s: expected rank 1, got %d", name, r)
		}
	}
}

// TC-12: Diamond dependency — rank uses max not sum
func TestNodeRankCompute_DiamondDependency(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/c", testEmptyFM()),
		testMakeInput("ROOT/a", testFMWithDependsOn([]string{"ROOT/c"})),
		testMakeInput("ROOT/b", testFMWithDependsOn([]string{"ROOT/c"})),
		testMakeInput("ROOT/d", testFMWithDependsOn([]string{"ROOT/a", "ROOT/b"})),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	tests := []struct {
		name string
		want int
	}{
		{"ROOT/c", 1},
		{"ROOT/a", 2},
		{"ROOT/b", 2},
		{"ROOT/d", 3},
	}
	for _, tt := range tests {
		if r := testRankOf(ranked, tt.name); r != tt.want {
			t.Errorf("%s: expected rank %d, got %d", tt.name, tt.want, r)
		}
	}
}

// TC-13: depends_on outranks parent
func TestNodeRankCompute_DependsOnOutranksParent(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testEmptyFM()),
		testMakeInput("ROOT/a/b", testFMWithDependsOn([]string{"ROOT/c"})),
		testMakeInput("ROOT/c", testEmptyFM()),
		testMakeInput("ROOT/c/d", testEmptyFM()),
		testMakeInput("ROOT/c/d/e", testEmptyFM()),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rankA := testRankOf(ranked, "ROOT/a")
	rankAB := testRankOf(ranked, "ROOT/a/b")
	rankC := testRankOf(ranked, "ROOT/c")

	if rankAB <= rankA {
		t.Errorf("expected rank of ROOT/a/b (%d) > rank of ROOT/a (%d)", rankAB, rankA)
	}

	// rank of ROOT/a/b should be 1 + max(rank of ROOT/a, rank of ROOT/c)
	maxDep := rankA
	if rankC > maxDep {
		maxDep = rankC
	}
	expectedRankAB := 1 + maxDep
	if rankAB != expectedRankAB {
		t.Errorf("ROOT/a/b: expected rank %d, got %d", expectedRankAB, rankAB)
	}
}

// TC-14: Multiple depends_on — rank from highest
func TestNodeRankCompute_MultiDependsOnRankFromHighest(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testEmptyFM()),
		testMakeInput("ROOT/b", testFMWithDependsOn([]string{"ROOT/a"})),
		testMakeInput("ROOT/c", testFMWithDependsOn([]string{"ROOT/b"})),
		testMakeInput("ROOT/d", testFMWithDependsOn([]string{"ROOT/a", "ROOT/b", "ROOT/c"})),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	tests := []struct {
		name string
		want int
	}{
		{"ROOT/a", 1},
		{"ROOT/b", 2},
		{"ROOT/c", 3},
		{"ROOT/d", 4},
	}
	for _, tt := range tests {
		if r := testRankOf(ranked, tt.name); r != tt.want {
			t.Errorf("%s: expected rank %d, got %d", tt.name, tt.want, r)
		}
	}
}

// TC-15: Node with both depends_on and input
func TestNodeRankCompute_NodeWithDepsAndInput(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testFMWithOutputs([]*frontmatter.FrontmatterOutput{
			{ID: "out", Path: "a.go"},
		})),
		testMakeInput("ROOT/b", testEmptyFM()),
		testMakeInput("ROOT/c", testFMWithDepsAndInput([]string{"ROOT/b"}, "ARTIFACT/a(out)")),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rankRoot := testRankOf(ranked, "ROOT")
	rankB := testRankOf(ranked, "ROOT/b")
	rankArtifact := testRankOf(ranked, "ARTIFACT/a(out)")
	rankC := testRankOf(ranked, "ROOT/c")

	// rank of ROOT/c = 1 + max(rank of ROOT (parent), rank of ROOT/b, rank of ARTIFACT/a(out))
	maxDep := rankRoot
	if rankB > maxDep {
		maxDep = rankB
	}
	if rankArtifact > maxDep {
		maxDep = rankArtifact
	}
	expectedRankC := 1 + maxDep
	if rankC != expectedRankC {
		t.Errorf("ROOT/c: expected rank %d, got %d", expectedRankC, rankC)
	}
}

// TC-16: Empty input list
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

// TC-17: Self-reference
func TestNodeRankCompute_SelfReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testFMWithDependsOn([]string{"ROOT/a"})),
	}

	_, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Error("expected cycles to be non-empty for self-reference")
	}
}

// TC-18: Simple cycle — two nodes
func TestNodeRankCompute_SimpleCycle(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testFMWithDependsOn([]string{"ROOT/b"})),
		testMakeInput("ROOT/b", testFMWithDependsOn([]string{"ROOT/a"})),
	}

	_, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Error("expected cycles to be non-empty")
	}
	if !testContains(cycles, "ROOT/a") && !testContains(cycles, "ROOT/b") {
		t.Errorf("expected cycles to contain ROOT/a or ROOT/b, got %v", cycles)
	}
}

// TC-19: Cycle through artifacts
func TestNodeRankCompute_CycleThroughArtifacts(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testFMWithOutputsAndDeps(
			[]*frontmatter.FrontmatterOutput{{ID: "out", Path: "a.go"}},
			[]string{"ARTIFACT/b(out)"},
		)),
		testMakeInput("ROOT/b", testFMWithOutputsAndDeps(
			[]*frontmatter.FrontmatterOutput{{ID: "out", Path: "b.go"}},
			[]string{"ARTIFACT/a(out)"},
		)),
	}

	_, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Error("expected cycles to be non-empty for artifact cycle")
	}
}

// TC-20: Cycle does not prevent ranking of unrelated nodes
func TestNodeRankCompute_CycleDoesNotPreventOtherRanking(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testFMWithDependsOn([]string{"ROOT/b"})),
		testMakeInput("ROOT/b", testFMWithDependsOn([]string{"ROOT/a"})),
		testMakeInput("ROOT/c", testEmptyFM()),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Error("expected cycles to be non-empty")
	}

	rankRoot := testRankOf(ranked, "ROOT")
	if rankRoot != 0 {
		t.Errorf("ROOT: expected rank 0, got %d", rankRoot)
	}

	rankC := testRankOf(ranked, "ROOT/c")
	if rankC < 0 {
		t.Error("expected ROOT/c to have a valid rank")
	}
	if rankC != 1 {
		t.Errorf("ROOT/c: expected rank 1, got %d", rankC)
	}

	if testContains(cycles, "ROOT/c") {
		t.Error("ROOT/c should not be in cycles")
	}

	if !testContains(cycles, "ROOT/a") && !testContains(cycles, "ROOT/b") {
		t.Errorf("expected cycles to contain ROOT/a or ROOT/b, got %v", cycles)
	}
}

// TC-21: Unresolvable ROOT reference
func TestNodeRankCompute_UnresolvableRootReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testFMWithDependsOn([]string{"ROOT/missing"})),
	}

	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got %v", err)
	}
}

// TC-22: Unresolvable ARTIFACT reference
func TestNodeRankCompute_UnresolvableArtifactReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testFMWithDependsOn([]string{"ARTIFACT/missing(id)"})),
	}

	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got %v", err)
	}
}

// TC-23: Unresolvable input reference
func TestNodeRankCompute_UnresolvableInputReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testFMWithInput("ARTIFACT/missing(id)")),
	}

	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got %v", err)
	}
}
