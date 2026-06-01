// code-from-spec: ROOT/golang/tests/utils/node_ranking@d1AQZR2eiA1XFE13OG-AjHGfejc
package noderanking_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/noderanking"
)

func testRankOf(t *testing.T, ranked []*noderanking.NodeRankEntry, logicalName string) int {
	t.Helper()
	for _, e := range ranked {
		if e.LogicalName == logicalName {
			return e.Rank
		}
	}
	t.Fatalf("testRankOf: %q not found in ranked list", logicalName)
	return -1
}

func testContains(t *testing.T, ranked []*noderanking.NodeRankEntry, logicalName string) bool {
	t.Helper()
	for _, e := range ranked {
		if e.LogicalName == logicalName {
			return true
		}
	}
	return false
}

func testCyclesContain(cycles []string, logicalName string) bool {
	for _, c := range cycles {
		if c == logicalName {
			return true
		}
	}
	return false
}

func testFM(dependsOn []string, inputs string, outputs []*frontmatter.FrontmatterOutput) *frontmatter.Frontmatter {
	return &frontmatter.Frontmatter{
		DependsOn: dependsOn,
		Input:     inputs,
		Outputs:   outputs,
	}
}

func testOut(id, path string) *frontmatter.FrontmatterOutput {
	return &frontmatter.FrontmatterOutput{ID: id, Path: path}
}

func TestNodeRankCompute_TC01_RootOnly(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ranked) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(ranked))
	}
	if ranked[0].LogicalName != "ROOT" || ranked[0].Rank != 0 {
		t.Errorf("expected ROOT rank 0, got %+v", ranked[0])
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
}

func TestNodeRankCompute_TC02_LinearChain(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a/b", Frontmatter: &frontmatter.Frontmatter{}},
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if testRankOf(t, ranked, "ROOT") != 0 {
		t.Errorf("expected ROOT rank 0")
	}
	if testRankOf(t, ranked, "ROOT/a") != 1 {
		t.Errorf("expected ROOT/a rank 1")
	}
	if testRankOf(t, ranked, "ROOT/a/b") != 2 {
		t.Errorf("expected ROOT/a/b rank 2")
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
}

func TestNodeRankCompute_TC03_IndependentSiblings(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/b", Frontmatter: &frontmatter.Frontmatter{}},
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if testRankOf(t, ranked, "ROOT") != 0 {
		t.Errorf("expected ROOT rank 0")
	}
	if testRankOf(t, ranked, "ROOT/a") != 1 {
		t.Errorf("expected ROOT/a rank 1")
	}
	if testRankOf(t, ranked, "ROOT/b") != 1 {
		t.Errorf("expected ROOT/b rank 1")
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
}

func TestNodeRankCompute_TC04_DependsOnIncreasesRank(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/b", Frontmatter: testFM([]string{"ROOT/a"}, "", nil)},
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rankA := testRankOf(t, ranked, "ROOT/a")
	rankB := testRankOf(t, ranked, "ROOT/b")
	if rankB <= rankA {
		t.Errorf("expected ROOT/b rank (%d) > ROOT/a rank (%d)", rankB, rankA)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
}

func TestNodeRankCompute_TC05_DependsOnQualifierStripped(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/b", Frontmatter: testFM([]string{"ROOT/a(interface)"}, "", nil)},
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rankA := testRankOf(t, ranked, "ROOT/a")
	rankB := testRankOf(t, ranked, "ROOT/b")
	if rankB <= rankA {
		t.Errorf("expected ROOT/b rank (%d) > ROOT/a rank (%d)", rankB, rankA)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
}

func TestNodeRankCompute_TC06_InputArtifactAddsEdge(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: testFM(nil, "", []*frontmatter.FrontmatterOutput{testOut("code", "out.go")})},
		{LogicalName: "ROOT/b", Frontmatter: testFM(nil, "ARTIFACT/a(code)", nil)},
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rankA := testRankOf(t, ranked, "ROOT/a")
	rankArtifact := testRankOf(t, ranked, "ARTIFACT/a(code)")
	rankB := testRankOf(t, ranked, "ROOT/b")
	if rankArtifact <= rankA {
		t.Errorf("expected ARTIFACT/a(code) rank (%d) > ROOT/a rank (%d)", rankArtifact, rankA)
	}
	if rankB <= rankArtifact {
		t.Errorf("expected ROOT/b rank (%d) > ARTIFACT/a(code) rank (%d)", rankB, rankArtifact)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
}

func TestNodeRankCompute_TC07_ArtifactRankOneAboveNode(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: testFM(nil, "", []*frontmatter.FrontmatterOutput{testOut("foo", "foo.go")})},
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rankA := testRankOf(t, ranked, "ROOT/a")
	rankArtifact := testRankOf(t, ranked, "ARTIFACT/a(foo)")
	if rankArtifact != rankA+1 {
		t.Errorf("expected ARTIFACT/a(foo) rank = ROOT/a rank + 1, got %d and %d", rankArtifact, rankA)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
}

func TestNodeRankCompute_TC08_MultipleOutputsEachRanked(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: testFM(nil, "", []*frontmatter.FrontmatterOutput{
				testOut("x", "x.go"),
				testOut("y", "y.go"),
			}),
		},
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !testContains(t, ranked, "ARTIFACT/a(x)") {
		t.Errorf("expected ARTIFACT/a(x) in ranked list")
	}
	if !testContains(t, ranked, "ARTIFACT/a(y)") {
		t.Errorf("expected ARTIFACT/a(y) in ranked list")
	}
	rankA := testRankOf(t, ranked, "ROOT/a")
	rankX := testRankOf(t, ranked, "ARTIFACT/a(x)")
	rankY := testRankOf(t, ranked, "ARTIFACT/a(y)")
	if rankX != rankA+1 {
		t.Errorf("expected ARTIFACT/a(x) rank = ROOT/a rank + 1")
	}
	if rankY != rankA+1 {
		t.Errorf("expected ARTIFACT/a(y) rank = ROOT/a rank + 1")
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
}

func TestNodeRankCompute_TC09_DependsOnArtifactReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: testFM(nil, "", []*frontmatter.FrontmatterOutput{testOut("lib", "lib.go")})},
		{LogicalName: "ROOT/b", Frontmatter: testFM([]string{"ARTIFACT/a(lib)"}, "", nil)},
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rankA := testRankOf(t, ranked, "ROOT/a")
	rankArtifact := testRankOf(t, ranked, "ARTIFACT/a(lib)")
	rankB := testRankOf(t, ranked, "ROOT/b")
	if rankArtifact <= rankA {
		t.Errorf("expected ARTIFACT/a(lib) rank (%d) > ROOT/a rank (%d)", rankArtifact, rankA)
	}
	if rankB <= rankArtifact {
		t.Errorf("expected ROOT/b rank (%d) > ARTIFACT/a(lib) rank (%d)", rankB, rankArtifact)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
}

func TestNodeRankCompute_TC10_OutputSortedByRankThenName(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/z", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{}},
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ranked[0].LogicalName != "ROOT" {
		t.Errorf("expected first entry to be ROOT, got %q", ranked[0].LogicalName)
	}
	var idxA, idxZ int
	for i, e := range ranked {
		if e.LogicalName == "ROOT/a" {
			idxA = i
		}
		if e.LogicalName == "ROOT/z" {
			idxZ = i
		}
	}
	if idxA >= idxZ {
		t.Errorf("expected ROOT/a before ROOT/z in ranked list")
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
}

func TestNodeRankCompute_TC11_ParallelSiblingsEqualRank(t *testing.T) {
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
	rankA := testRankOf(t, ranked, "ROOT/a")
	rankB := testRankOf(t, ranked, "ROOT/b")
	rankC := testRankOf(t, ranked, "ROOT/c")
	if rankA != 1 || rankB != 1 || rankC != 1 {
		t.Errorf("expected all siblings rank 1, got a=%d b=%d c=%d", rankA, rankB, rankC)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
}

func TestNodeRankCompute_TC12_DiamondDependency(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/c", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: testFM([]string{"ROOT/c"}, "", nil)},
		{LogicalName: "ROOT/b", Frontmatter: testFM([]string{"ROOT/c"}, "", nil)},
		{LogicalName: "ROOT/d", Frontmatter: testFM([]string{"ROOT/a", "ROOT/b"}, "", nil)},
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if testRankOf(t, ranked, "ROOT/c") != 1 {
		t.Errorf("expected ROOT/c rank 1")
	}
	if testRankOf(t, ranked, "ROOT/a") != 2 {
		t.Errorf("expected ROOT/a rank 2")
	}
	if testRankOf(t, ranked, "ROOT/b") != 2 {
		t.Errorf("expected ROOT/b rank 2")
	}
	if testRankOf(t, ranked, "ROOT/d") != 3 {
		t.Errorf("expected ROOT/d rank 3")
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
}

func TestNodeRankCompute_TC13_DependsOnOutranksParent(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a/b", Frontmatter: testFM([]string{"ROOT/c"}, "", nil)},
		{LogicalName: "ROOT/c", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/c/d", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/c/d/e", Frontmatter: &frontmatter.Frontmatter{}},
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rankA := testRankOf(t, ranked, "ROOT/a")
	rankAB := testRankOf(t, ranked, "ROOT/a/b")
	rankC := testRankOf(t, ranked, "ROOT/c")
	if rankAB <= rankA {
		t.Errorf("expected ROOT/a/b rank (%d) > ROOT/a rank (%d)", rankAB, rankA)
	}
	expectedRankAB := 1 + max(rankA, rankC)
	if rankAB != expectedRankAB {
		t.Errorf("expected ROOT/a/b rank = 1 + max(rank(ROOT/a), rank(ROOT/c)) = %d, got %d", expectedRankAB, rankAB)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
}

func TestNodeRankCompute_TC14_MultipleDependsOnRankFromHighest(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/b", Frontmatter: testFM([]string{"ROOT/a"}, "", nil)},
		{LogicalName: "ROOT/c", Frontmatter: testFM([]string{"ROOT/b"}, "", nil)},
		{LogicalName: "ROOT/d", Frontmatter: testFM([]string{"ROOT/a", "ROOT/b", "ROOT/c"}, "", nil)},
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if testRankOf(t, ranked, "ROOT/a") != 1 {
		t.Errorf("expected ROOT/a rank 1")
	}
	if testRankOf(t, ranked, "ROOT/b") != 2 {
		t.Errorf("expected ROOT/b rank 2")
	}
	if testRankOf(t, ranked, "ROOT/c") != 3 {
		t.Errorf("expected ROOT/c rank 3")
	}
	if testRankOf(t, ranked, "ROOT/d") != 4 {
		t.Errorf("expected ROOT/d rank 4")
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
}

func TestNodeRankCompute_TC15_BothDependsOnAndInput(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: testFM(nil, "", []*frontmatter.FrontmatterOutput{testOut("out", "a.go")})},
		{LogicalName: "ROOT/b", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/c", Frontmatter: testFM([]string{"ROOT/b"}, "ARTIFACT/a(out)", nil)},
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rankROOT := testRankOf(t, ranked, "ROOT")
	rankB := testRankOf(t, ranked, "ROOT/b")
	rankArtifact := testRankOf(t, ranked, "ARTIFACT/a(out)")
	rankC := testRankOf(t, ranked, "ROOT/c")
	expectedRankC := 1 + max(rankROOT, rankB, rankArtifact)
	if rankC != expectedRankC {
		t.Errorf("expected ROOT/c rank = 1 + max(rank(ROOT), rank(ROOT/b), rank(ARTIFACT/a(out))) = %d, got %d", expectedRankC, rankC)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
}

func TestNodeRankCompute_TC16_EmptyInputList(t *testing.T) {
	entries := []*noderanking.NodeRankInput{}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ranked) != 0 {
		t.Errorf("expected empty ranked list, got %d entries", len(ranked))
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
}

func TestNodeRankCompute_TC17_SelfReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: testFM([]string{"ROOT/a"}, "", nil)},
	}
	_, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Errorf("expected cycles to be non-empty for self-reference")
	}
}

func TestNodeRankCompute_TC18_SimpleCycleTwoNodes(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: testFM([]string{"ROOT/b"}, "", nil)},
		{LogicalName: "ROOT/b", Frontmatter: testFM([]string{"ROOT/a"}, "", nil)},
	}
	_, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Errorf("expected cycles to be non-empty")
	}
	if !testCyclesContain(cycles, "ROOT/a") && !testCyclesContain(cycles, "ROOT/b") {
		t.Errorf("expected cycles to contain ROOT/a or ROOT/b, got %v", cycles)
	}
}

func TestNodeRankCompute_TC19_CycleThroughArtifacts(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: testFM([]string{"ARTIFACT/b(out)"}, "", []*frontmatter.FrontmatterOutput{testOut("out", "a.go")}),
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: testFM([]string{"ARTIFACT/a(out)"}, "", []*frontmatter.FrontmatterOutput{testOut("out", "b.go")}),
		},
	}
	_, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Errorf("expected cycles to be non-empty for artifact cycle")
	}
}

func TestNodeRankCompute_TC20_CycleDoesNotPreventUnrelatedRanking(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: testFM([]string{"ROOT/b"}, "", nil)},
		{LogicalName: "ROOT/b", Frontmatter: testFM([]string{"ROOT/a"}, "", nil)},
		{LogicalName: "ROOT/c", Frontmatter: &frontmatter.Frontmatter{}},
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if testRankOf(t, ranked, "ROOT") != 0 {
		t.Errorf("expected ROOT rank 0")
	}
	if testRankOf(t, ranked, "ROOT/c") != 1 {
		t.Errorf("expected ROOT/c rank 1")
	}
	if len(cycles) == 0 {
		t.Errorf("expected cycles to be non-empty")
	}
	if testCyclesContain(cycles, "ROOT/c") {
		t.Errorf("expected ROOT/c not to be in cycles, got %v", cycles)
	}
}

func TestNodeRankCompute_TC21_UnresolvableRootReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: testFM([]string{"ROOT/missing"}, "", nil)},
	}
	_, _, err := noderanking.NodeRankCompute(entries)
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got %v", err)
	}
}

func TestNodeRankCompute_TC22_UnresolvableArtifactReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: testFM([]string{"ARTIFACT/missing(id)"}, "", nil)},
	}
	_, _, err := noderanking.NodeRankCompute(entries)
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got %v", err)
	}
}

func TestNodeRankCompute_TC23_UnresolvableInputReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: testFM(nil, "ARTIFACT/missing(id)", nil)},
	}
	_, _, err := noderanking.NodeRankCompute(entries)
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got %v", err)
	}
}

func max(a int, rest ...int) int {
	m := a
	for _, v := range rest {
		if v > m {
			m = v
		}
	}
	return m
}
