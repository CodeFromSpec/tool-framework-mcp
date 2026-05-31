// code-from-spec: ROOT/golang/tests/utils/node_ranking@qFk91yd20zVu-9sh-BBYZzjqWWI
package noderanking_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/noderanking"
)

// testRankOf returns the rank for a given logical name from the ranked list.
// Fails the test if the entry is not found.
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

// testContains returns true if the slice contains the given string.
func testContains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// testHasEntry returns true if ranked contains an entry with the given logical name.
func testHasEntry(ranked []*noderanking.NodeRankEntry, logicalName string) bool {
	for _, e := range ranked {
		if e.LogicalName == logicalName {
			return true
		}
	}
	return false
}

// TC-01: Root only — single entry with no dependencies yields rank 0.
func TestNodeRankCompute_TC01_RootOnly(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{},
				Outputs:   []*frontmatter.FrontmatterOutput{},
			},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got: %v", cycles)
	}
	if len(ranked) != 1 {
		t.Fatalf("expected 1 ranked entry, got %d", len(ranked))
	}
	if ranked[0].LogicalName != "ROOT" {
		t.Errorf("expected ROOT, got %q", ranked[0].LogicalName)
	}
	if ranked[0].Rank != 0 {
		t.Errorf("expected rank 0, got %d", ranked[0].Rank)
	}
}

// TC-02: Linear chain — incrementing ranks.
func TestNodeRankCompute_TC02_LinearChain(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got: %v", cycles)
	}

	rankROOT := testRankOf(t, ranked, "ROOT")
	rankA := testRankOf(t, ranked, "ROOT/a")
	rankAB := testRankOf(t, ranked, "ROOT/a/b")

	if rankROOT != 0 {
		t.Errorf("ROOT: expected rank 0, got %d", rankROOT)
	}
	if rankA != 1 {
		t.Errorf("ROOT/a: expected rank 1, got %d", rankA)
	}
	if rankAB != 2 {
		t.Errorf("ROOT/a/b: expected rank 2, got %d", rankAB)
	}
}

// TC-03: Independent siblings — equal rank.
func TestNodeRankCompute_TC03_IndependentSiblings(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got: %v", cycles)
	}

	rankA := testRankOf(t, ranked, "ROOT/a")
	rankB := testRankOf(t, ranked, "ROOT/b")

	if rankA != rankB {
		t.Errorf("expected ROOT/a and ROOT/b to have the same rank, got %d and %d", rankA, rankB)
	}
	if rankA != 1 {
		t.Errorf("expected rank 1, got %d", rankA)
	}
}

// TC-04: depends_on increases rank.
func TestNodeRankCompute_TC04_DependsOnIncreasesRank(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/a"}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got: %v", cycles)
	}

	rankA := testRankOf(t, ranked, "ROOT/a")
	rankB := testRankOf(t, ranked, "ROOT/b")

	if rankB <= rankA {
		t.Errorf("expected rank of ROOT/b (%d) > rank of ROOT/a (%d)", rankB, rankA)
	}
}

// TC-05: depends_on with qualifier — qualifier stripped.
func TestNodeRankCompute_TC05_DependsOnWithQualifier(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/a(interface)"}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got: %v", cycles)
	}

	rankA := testRankOf(t, ranked, "ROOT/a")
	rankB := testRankOf(t, ranked, "ROOT/b")

	if rankB <= rankA {
		t.Errorf("expected rank of ROOT/b (%d) > rank of ROOT/a (%d)", rankB, rankA)
	}
}

// TC-06: input artifact adds dependency edge.
func TestNodeRankCompute_TC06_InputArtifactAddsEdge(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{},
				Outputs:   []*frontmatter.FrontmatterOutput{{ID: "code", Path: "out.go"}},
			},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{
				Input:     "ARTIFACT/a(code)",
				DependsOn: []string{},
				Outputs:   []*frontmatter.FrontmatterOutput{},
			},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got: %v", cycles)
	}

	rankA := testRankOf(t, ranked, "ROOT/a")
	rankArtifact := testRankOf(t, ranked, "ARTIFACT/a(code)")
	rankB := testRankOf(t, ranked, "ROOT/b")

	if rankArtifact <= rankA {
		t.Errorf("expected rank of ARTIFACT/a(code) (%d) > rank of ROOT/a (%d)", rankArtifact, rankA)
	}
	if rankB <= rankArtifact {
		t.Errorf("expected rank of ROOT/b (%d) > rank of ARTIFACT/a(code) (%d)", rankB, rankArtifact)
	}
}

// TC-07: Artifacts get rank one above their node.
func TestNodeRankCompute_TC07_ArtifactRankOneAboveNode(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{},
				Outputs:   []*frontmatter.FrontmatterOutput{{ID: "foo", Path: "foo.go"}},
			},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got: %v", cycles)
	}

	if !testHasEntry(ranked, "ARTIFACT/a(foo)") {
		t.Fatal("expected ARTIFACT/a(foo) in ranked list")
	}

	rankA := testRankOf(t, ranked, "ROOT/a")
	rankFoo := testRankOf(t, ranked, "ARTIFACT/a(foo)")

	if rankFoo != rankA+1 {
		t.Errorf("expected rank of ARTIFACT/a(foo) = %d (ROOT/a rank + 1), got %d", rankA+1, rankFoo)
	}
}

// TC-08: Multiple outputs — each artifact ranked.
func TestNodeRankCompute_TC08_MultipleOutputsEachArtifactRanked(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{},
				Outputs: []*frontmatter.FrontmatterOutput{
					{ID: "x", Path: "x.go"},
					{ID: "y", Path: "y.go"},
				},
			},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got: %v", cycles)
	}

	rankA := testRankOf(t, ranked, "ROOT/a")

	for _, artifactName := range []string{"ARTIFACT/a(x)", "ARTIFACT/a(y)"} {
		if !testHasEntry(ranked, artifactName) {
			t.Fatalf("expected %q in ranked list", artifactName)
		}
		rank := testRankOf(t, ranked, artifactName)
		if rank != rankA+1 {
			t.Errorf("expected rank of %q = %d (ROOT/a rank + 1), got %d", artifactName, rankA+1, rank)
		}
	}
}

// TC-09: depends_on ARTIFACT reference — used as-is.
func TestNodeRankCompute_TC09_DependsOnArtifactReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{},
				Outputs:   []*frontmatter.FrontmatterOutput{{ID: "lib", Path: "lib.go"}},
			},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ARTIFACT/a(lib)"},
				Outputs:   []*frontmatter.FrontmatterOutput{},
			},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got: %v", cycles)
	}

	rankA := testRankOf(t, ranked, "ROOT/a")
	rankLib := testRankOf(t, ranked, "ARTIFACT/a(lib)")
	rankB := testRankOf(t, ranked, "ROOT/b")

	if rankLib <= rankA {
		t.Errorf("expected rank of ARTIFACT/a(lib) (%d) > rank of ROOT/a (%d)", rankLib, rankA)
	}
	if rankB <= rankLib {
		t.Errorf("expected rank of ROOT/b (%d) > rank of ARTIFACT/a(lib) (%d)", rankB, rankLib)
	}
}

// TC-10: Output sorted by rank then logical name.
func TestNodeRankCompute_TC10_OutputSortedByRankThenName(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/z",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got: %v", cycles)
	}
	if len(ranked) != 3 {
		t.Fatalf("expected 3 ranked entries, got %d", len(ranked))
	}

	if ranked[0].LogicalName != "ROOT" {
		t.Errorf("expected first entry ROOT, got %q", ranked[0].LogicalName)
	}
	if ranked[1].LogicalName != "ROOT/a" {
		t.Errorf("expected second entry ROOT/a, got %q", ranked[1].LogicalName)
	}
	if ranked[2].LogicalName != "ROOT/z" {
		t.Errorf("expected third entry ROOT/z, got %q", ranked[2].LogicalName)
	}
}

// TC-11: Parallel entries — equal rank means no dependency.
func TestNodeRankCompute_TC11_ParallelEqualRank(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/c",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got: %v", cycles)
	}

	rankA := testRankOf(t, ranked, "ROOT/a")
	rankB := testRankOf(t, ranked, "ROOT/b")
	rankC := testRankOf(t, ranked, "ROOT/c")

	if rankA != 1 || rankB != 1 || rankC != 1 {
		t.Errorf("expected ROOT/a, ROOT/b, ROOT/c all rank 1, got %d, %d, %d", rankA, rankB, rankC)
	}
}

// TC-12: Diamond dependency — rank uses max not sum.
func TestNodeRankCompute_TC12_DiamondDependency(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/c",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/c"}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/c"}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/d",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/a", "ROOT/b"}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got: %v", cycles)
	}

	rankC := testRankOf(t, ranked, "ROOT/c")
	rankA := testRankOf(t, ranked, "ROOT/a")
	rankB := testRankOf(t, ranked, "ROOT/b")
	rankD := testRankOf(t, ranked, "ROOT/d")

	if rankC != 1 {
		t.Errorf("ROOT/c: expected rank 1, got %d", rankC)
	}
	if rankA != 2 {
		t.Errorf("ROOT/a: expected rank 2, got %d", rankA)
	}
	if rankB != 2 {
		t.Errorf("ROOT/b: expected rank 2, got %d", rankB)
	}
	if rankD != 3 {
		t.Errorf("ROOT/d: expected rank 3 (not 5), got %d", rankD)
	}
}

// TC-13: depends_on outranks parent.
func TestNodeRankCompute_TC13_DependsOnOutranksParent(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/c"}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/c",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/c/d",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/c/d/e",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got: %v", cycles)
	}

	rankA := testRankOf(t, ranked, "ROOT/a")
	rankAB := testRankOf(t, ranked, "ROOT/a/b")
	rankC := testRankOf(t, ranked, "ROOT/c")

	if rankAB <= rankA {
		t.Errorf("expected rank of ROOT/a/b (%d) > rank of ROOT/a (%d)", rankAB, rankA)
	}

	// rank of ROOT/a/b = 1 + max(rank of ROOT/a, rank of ROOT/c)
	maxDep := rankA
	if rankC > maxDep {
		maxDep = rankC
	}
	expectedAB := 1 + maxDep
	if rankAB != expectedAB {
		t.Errorf("expected rank of ROOT/a/b = %d (1 + max(%d, %d)), got %d", expectedAB, rankA, rankC, rankAB)
	}
}

// TC-14: Multiple depends_on — rank from highest.
func TestNodeRankCompute_TC14_MultipleDependsOnRankFromHighest(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/a"}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/c",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/b"}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/d",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/a", "ROOT/b", "ROOT/c"}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got: %v", cycles)
	}

	rankA := testRankOf(t, ranked, "ROOT/a")
	rankB := testRankOf(t, ranked, "ROOT/b")
	rankC := testRankOf(t, ranked, "ROOT/c")
	rankD := testRankOf(t, ranked, "ROOT/d")

	if rankA != 1 {
		t.Errorf("ROOT/a: expected rank 1, got %d", rankA)
	}
	if rankB != 2 {
		t.Errorf("ROOT/b: expected rank 2, got %d", rankB)
	}
	if rankC != 3 {
		t.Errorf("ROOT/c: expected rank 3, got %d", rankC)
	}
	if rankD != 4 {
		t.Errorf("ROOT/d: expected rank 4, got %d", rankD)
	}
}

// TC-15: Node with both depends_on and input.
func TestNodeRankCompute_TC15_BothDependsOnAndInput(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{},
				Outputs:   []*frontmatter.FrontmatterOutput{{ID: "out", Path: "a.go"}},
			},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/c",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/b"},
				Input:     "ARTIFACT/a(out)",
				Outputs:   []*frontmatter.FrontmatterOutput{},
			},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got: %v", cycles)
	}

	rankROOT := testRankOf(t, ranked, "ROOT")
	rankB := testRankOf(t, ranked, "ROOT/b")
	rankArtifactOut := testRankOf(t, ranked, "ARTIFACT/a(out)")
	rankC := testRankOf(t, ranked, "ROOT/c")

	// rank of ROOT/c = 1 + max(rank of ROOT (parent), rank of ROOT/b, rank of ARTIFACT/a(out))
	maxDep := rankROOT
	if rankB > maxDep {
		maxDep = rankB
	}
	if rankArtifactOut > maxDep {
		maxDep = rankArtifactOut
	}
	expectedC := 1 + maxDep
	if rankC != expectedC {
		t.Errorf("expected rank of ROOT/c = %d, got %d", expectedC, rankC)
	}
}

// TC-16: Empty input list.
func TestNodeRankCompute_TC16_EmptyInput(t *testing.T) {
	entries := []*noderanking.NodeRankInput{}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got: %v", cycles)
	}
	if len(ranked) != 0 {
		t.Errorf("expected empty ranked list, got %d entries", len(ranked))
	}
}

// TC-17: Self-reference cycle.
func TestNodeRankCompute_TC17_SelfReferenceCycle(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/a"}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
	}

	_, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Error("expected cycles to be non-empty for self-reference")
	}
}

// TC-18: Simple cycle — two nodes.
func TestNodeRankCompute_TC18_SimpleCycleTwoNodes(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/b"}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/a"}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
	}

	_, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Error("expected cycles to be non-empty")
	}
	if !testContains(cycles, "ROOT/a") && !testContains(cycles, "ROOT/b") {
		t.Errorf("expected cycles to contain ROOT/a or ROOT/b, got: %v", cycles)
	}
}

// TC-19: Cycle through artifacts.
func TestNodeRankCompute_TC19_CycleThroughArtifacts(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ARTIFACT/b(out)"},
				Outputs:   []*frontmatter.FrontmatterOutput{{ID: "out", Path: "a.go"}},
			},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ARTIFACT/a(out)"},
				Outputs:   []*frontmatter.FrontmatterOutput{{ID: "out", Path: "b.go"}},
			},
		},
	}

	_, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Error("expected cycles to be non-empty for artifact cycle")
	}
}

// TC-20: Cycle does not prevent ranking of unrelated nodes.
func TestNodeRankCompute_TC20_CycleDoesNotBlockUnrelated(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/b"}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/a"}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/c",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Error("expected cycles to be non-empty")
	}

	rankROOT := testRankOf(t, ranked, "ROOT")
	rankC := testRankOf(t, ranked, "ROOT/c")

	if rankROOT != 0 {
		t.Errorf("ROOT: expected rank 0, got %d", rankROOT)
	}
	if rankC != 1 {
		t.Errorf("ROOT/c: expected rank 1, got %d", rankC)
	}
	if testContains(cycles, "ROOT/c") {
		t.Errorf("ROOT/c should not be in cycles, got: %v", cycles)
	}
}

// TC-21: Unresolvable ROOT reference.
func TestNodeRankCompute_TC21_UnresolvableRootReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ROOT/missing"}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
	}

	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got: %v", err)
	}
}

// TC-22: Unresolvable ARTIFACT reference.
func TestNodeRankCompute_TC22_UnresolvableArtifactReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ARTIFACT/missing(id)"}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
	}

	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got: %v", err)
	}
}

// TC-23: Unresolvable input reference.
func TestNodeRankCompute_TC23_UnresolvableInputReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{}, Outputs: []*frontmatter.FrontmatterOutput{}},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				Input:     "ARTIFACT/missing(id)",
				DependsOn: []string{},
				Outputs:   []*frontmatter.FrontmatterOutput{},
			},
		},
	}

	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got: %v", err)
	}
}
