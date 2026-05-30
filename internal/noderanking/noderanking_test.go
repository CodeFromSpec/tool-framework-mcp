// code-from-spec: ROOT/golang/tests/utils/node_ranking@jup7fASjv-3spPYNIOHwVC7iowU
package noderanking_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/noderanking"
)

// testMakeInput constructs a NodeRankInput with the given logical name and frontmatter.
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

// testFMWithDependsOn returns a Frontmatter with the given depends_on list.
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

// testFMWithDependsOnAndInput returns a Frontmatter with depends_on and input.
func testFMWithDependsOnAndInput(deps []string, input string) *frontmatter.Frontmatter {
	return &frontmatter.Frontmatter{
		DependsOn: deps,
		Input:     input,
	}
}

// testFMWithInput returns a Frontmatter with only an input field.
func testFMWithInput(input string) *frontmatter.Frontmatter {
	return &frontmatter.Frontmatter{
		Input: input,
	}
}

// testFMWithOutputsAndDependsOn returns a Frontmatter with outputs and depends_on.
func testFMWithOutputsAndDependsOn(outputs []*frontmatter.FrontmatterOutput, deps []string) *frontmatter.Frontmatter {
	return &frontmatter.Frontmatter{
		Outputs:   outputs,
		DependsOn: deps,
	}
}

// testRankOf finds the rank for the given logical name in the ranked list.
// Returns -1 if not found.
func testRankOf(ranked []*noderanking.NodeRankEntry, logicalName string) int {
	for _, e := range ranked {
		if e.LogicalName == logicalName {
			return e.Rank
		}
	}
	return -1
}

// testContains checks whether a string slice contains the given value.
func testContains(slice []string, value string) bool {
	for _, s := range slice {
		if s == value {
			return true
		}
	}
	return false
}

// testOutput returns a FrontmatterOutput.
func testOutput(id, path string) *frontmatter.FrontmatterOutput {
	return &frontmatter.FrontmatterOutput{ID: id, Path: path}
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
		t.Fatalf("expected 1 ranked entry, got %d", len(ranked))
	}
	if ranked[0].LogicalName != "ROOT" {
		t.Errorf("expected logical name ROOT, got %s", ranked[0].LogicalName)
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

	if testRankOf(ranked, "ROOT") != 0 {
		t.Errorf("expected ROOT rank 0, got %d", testRankOf(ranked, "ROOT"))
	}
	if testRankOf(ranked, "ROOT/a") != 1 {
		t.Errorf("expected ROOT/a rank 1, got %d", testRankOf(ranked, "ROOT/a"))
	}
	if testRankOf(ranked, "ROOT/a/b") != 2 {
		t.Errorf("expected ROOT/a/b rank 2, got %d", testRankOf(ranked, "ROOT/a/b"))
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
		t.Errorf("expected ROOT/a and ROOT/b to have equal rank, got %d and %d", rankA, rankB)
	}
	if rankA != 1 {
		t.Errorf("expected siblings rank 1, got %d", rankA)
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
func TestNodeRankCompute_InputArtifactAddsDependencyEdge(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testFMWithOutputs([]*frontmatter.FrontmatterOutput{
			testOutput("code", "out.go"),
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

	rankA := testRankOf(ranked, "ROOT/a")
	rankArtifact := testRankOf(ranked, "ARTIFACT/a(code)")
	rankB := testRankOf(ranked, "ROOT/b")

	if rankArtifact == -1 {
		t.Fatalf("expected ranked to contain ARTIFACT/a(code)")
	}
	if rankArtifact <= rankA {
		t.Errorf("expected rank of ARTIFACT/a(code) (%d) > rank of ROOT/a (%d)", rankArtifact, rankA)
	}
	if rankB <= rankArtifact {
		t.Errorf("expected rank of ROOT/b (%d) > rank of ARTIFACT/a(code) (%d)", rankB, rankArtifact)
	}
}

// TC-07: Artifacts get rank one above their node
func TestNodeRankCompute_ArtifactRankOneAboveNode(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testFMWithOutputs([]*frontmatter.FrontmatterOutput{
			testOutput("foo", "foo.go"),
		})),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rankA := testRankOf(ranked, "ROOT/a")
	rankArtifact := testRankOf(ranked, "ARTIFACT/a(foo)")

	if rankArtifact == -1 {
		t.Fatalf("expected ranked to contain ARTIFACT/a(foo)")
	}
	if rankArtifact != rankA+1 {
		t.Errorf("expected ARTIFACT/a(foo) rank = ROOT/a rank + 1 (%d), got %d", rankA+1, rankArtifact)
	}
}

// TC-08: Multiple outputs — each artifact ranked
func TestNodeRankCompute_MultipleOutputsEachArtifactRanked(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testFMWithOutputs([]*frontmatter.FrontmatterOutput{
			testOutput("x", "x.go"),
			testOutput("y", "y.go"),
		})),
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	rankA := testRankOf(ranked, "ROOT/a")
	rankX := testRankOf(ranked, "ARTIFACT/a(x)")
	rankY := testRankOf(ranked, "ARTIFACT/a(y)")

	if rankX == -1 {
		t.Fatalf("expected ranked to contain ARTIFACT/a(x)")
	}
	if rankY == -1 {
		t.Fatalf("expected ranked to contain ARTIFACT/a(y)")
	}
	if rankX != rankA+1 {
		t.Errorf("expected ARTIFACT/a(x) rank = ROOT/a rank + 1 (%d), got %d", rankA+1, rankX)
	}
	if rankY != rankA+1 {
		t.Errorf("expected ARTIFACT/a(y) rank = ROOT/a rank + 1 (%d), got %d", rankA+1, rankY)
	}
}

// TC-09: depends_on ARTIFACT reference — used as-is
func TestNodeRankCompute_DependsOnArtifactReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testFMWithOutputs([]*frontmatter.FrontmatterOutput{
			testOutput("lib", "lib.go"),
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

	rankA := testRankOf(ranked, "ROOT/a")
	rankArtifact := testRankOf(ranked, "ARTIFACT/a(lib)")
	rankB := testRankOf(ranked, "ROOT/b")

	if rankArtifact == -1 {
		t.Fatalf("expected ranked to contain ARTIFACT/a(lib)")
	}
	if rankArtifact <= rankA {
		t.Errorf("expected rank of ARTIFACT/a(lib) (%d) > rank of ROOT/a (%d)", rankArtifact, rankA)
	}
	if rankB <= rankArtifact {
		t.Errorf("expected rank of ROOT/b (%d) > rank of ARTIFACT/a(lib) (%d)", rankB, rankArtifact)
	}
}

// TC-10: Output sorted by rank then logical name
func TestNodeRankCompute_OutputSortedByRankThenName(t *testing.T) {
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

	// Find positions in the ranked slice.
	posRoot := -1
	posA := -1
	posZ := -1
	for i, e := range ranked {
		switch e.LogicalName {
		case "ROOT":
			posRoot = i
		case "ROOT/a":
			posA = i
		case "ROOT/z":
			posZ = i
		}
	}

	if posRoot == -1 || posA == -1 || posZ == -1 {
		t.Fatalf("missing entries in ranked output: ROOT=%d, ROOT/a=%d, ROOT/z=%d", posRoot, posA, posZ)
	}
	if posRoot >= posA {
		t.Errorf("expected ROOT (pos %d) before ROOT/a (pos %d)", posRoot, posA)
	}
	if posRoot >= posZ {
		t.Errorf("expected ROOT (pos %d) before ROOT/z (pos %d)", posRoot, posZ)
	}
	if posA >= posZ {
		t.Errorf("expected ROOT/a (pos %d) before ROOT/z (pos %d) alphabetically", posA, posZ)
	}
}

// TC-11: Parallel entries — equal rank means no dependency
func TestNodeRankCompute_ParallelEqualRank(t *testing.T) {
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

	rankA := testRankOf(ranked, "ROOT/a")
	rankB := testRankOf(ranked, "ROOT/b")
	rankC := testRankOf(ranked, "ROOT/c")

	if rankA != 1 || rankB != 1 || rankC != 1 {
		t.Errorf("expected ROOT/a, ROOT/b, ROOT/c all rank 1, got %d, %d, %d", rankA, rankB, rankC)
	}
}

// TC-12: Diamond dependency — rank uses max not sum
func TestNodeRankCompute_DiamondDependencyUsesMax(t *testing.T) {
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

	rankC := testRankOf(ranked, "ROOT/c")
	rankA := testRankOf(ranked, "ROOT/a")
	rankB := testRankOf(ranked, "ROOT/b")
	rankD := testRankOf(ranked, "ROOT/d")

	if rankC != 1 {
		t.Errorf("expected ROOT/c rank 1, got %d", rankC)
	}
	if rankA != 2 {
		t.Errorf("expected ROOT/a rank 2, got %d", rankA)
	}
	if rankB != 2 {
		t.Errorf("expected ROOT/b rank 2, got %d", rankB)
	}
	if rankD != 3 {
		t.Errorf("expected ROOT/d rank 3 (max not sum), got %d", rankD)
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
	rankC := testRankOf(ranked, "ROOT/c")
	rankAB := testRankOf(ranked, "ROOT/a/b")

	if rankAB <= rankA {
		t.Errorf("expected rank of ROOT/a/b (%d) > rank of ROOT/a (%d)", rankAB, rankA)
	}

	expectedRankAB := 1 + max(rankA, rankC)
	if rankAB != expectedRankAB {
		t.Errorf("expected ROOT/a/b rank = 1 + max(ROOT/a=%d, ROOT/c=%d) = %d, got %d",
			rankA, rankC, expectedRankAB, rankAB)
	}
}

// TC-14: Multiple depends_on — rank from highest
func TestNodeRankCompute_MultipleDependsOnRankFromHighest(t *testing.T) {
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

	rankA := testRankOf(ranked, "ROOT/a")
	rankB := testRankOf(ranked, "ROOT/b")
	rankC := testRankOf(ranked, "ROOT/c")
	rankD := testRankOf(ranked, "ROOT/d")

	if rankA != 1 {
		t.Errorf("expected ROOT/a rank 1, got %d", rankA)
	}
	if rankB != 2 {
		t.Errorf("expected ROOT/b rank 2, got %d", rankB)
	}
	if rankC != 3 {
		t.Errorf("expected ROOT/c rank 3, got %d", rankC)
	}
	if rankD != 4 {
		t.Errorf("expected ROOT/d rank 4 (1 + max(1,2,3)), got %d", rankD)
	}
}

// TC-15: Node with both depends_on and input
func TestNodeRankCompute_BothDependsOnAndInput(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		testMakeInput("ROOT", testEmptyFM()),
		testMakeInput("ROOT/a", testFMWithOutputs([]*frontmatter.FrontmatterOutput{
			testOutput("out", "a.go"),
		})),
		testMakeInput("ROOT/b", testEmptyFM()),
		testMakeInput("ROOT/c", testFMWithDependsOnAndInput([]string{"ROOT/b"}, "ARTIFACT/a(out)")),
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

	if rankArtifact == -1 {
		t.Fatalf("expected ranked to contain ARTIFACT/a(out)")
	}

	expectedRankC := 1 + max3(rankRoot, rankB, rankArtifact)
	if rankC != expectedRankC {
		t.Errorf("expected ROOT/c rank = 1 + max(ROOT=%d, ROOT/b=%d, ARTIFACT/a(out)=%d) = %d, got %d",
			rankRoot, rankB, rankArtifact, expectedRankC, rankC)
	}
}

// max3 returns the maximum of three integers.
func max3(a, b, c int) int {
	if a >= b && a >= c {
		return a
	}
	if b >= c {
		return b
	}
	return c
}

// TC-16: Empty input list
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
		t.Errorf("expected empty ranked list, got %v", ranked)
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
func TestNodeRankCompute_SimpleCycleTwoNodes(t *testing.T) {
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
		testMakeInput("ROOT/a", testFMWithOutputsAndDependsOn(
			[]*frontmatter.FrontmatterOutput{testOutput("out", "a.go")},
			[]string{"ARTIFACT/b(out)"},
		)),
		testMakeInput("ROOT/b", testFMWithOutputsAndDependsOn(
			[]*frontmatter.FrontmatterOutput{testOutput("out", "b.go")},
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
func TestNodeRankCompute_CycleDoesNotPreventUnrelatedRanking(t *testing.T) {
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
	rankC := testRankOf(ranked, "ROOT/c")

	if rankRoot == -1 {
		t.Error("expected ROOT to have a valid rank")
	}
	if rankC == -1 {
		t.Error("expected ROOT/c to have a valid rank")
	}
	if rankRoot != 0 {
		t.Errorf("expected ROOT rank 0, got %d", rankRoot)
	}
	if rankC != 1 {
		t.Errorf("expected ROOT/c rank 1, got %d", rankC)
	}

	if testContains(cycles, "ROOT/c") {
		t.Error("expected ROOT/c not to be in cycles")
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
		t.Fatal("expected error for unresolvable reference, got nil")
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
		t.Fatal("expected error for unresolvable artifact reference, got nil")
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
		t.Fatal("expected error for unresolvable input reference, got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got %v", err)
	}
}
