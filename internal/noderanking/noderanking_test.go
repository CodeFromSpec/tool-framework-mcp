// code-from-spec: ROOT/golang/tests/utils/node_ranking@OoclGPKVFlT_PmEcBXvzRGyzPIQ

package noderanking_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/noderanking"
)

// testStr returns a pointer to the given string, for convenience when
// building DependsOn slices.
func testStr(s string) *string {
	return &s
}

// testFindEntry returns the NodeRankEntry with the given logical name,
// or nil if not found.
func testFindEntry(ranked []*noderanking.NodeRankEntry, logicalName string) *noderanking.NodeRankEntry {
	for _, e := range ranked {
		if e.LogicalName == logicalName {
			return e
		}
	}
	return nil
}

// testRankOf returns the rank of the entry with the given logical name.
// It calls t.Fatalf if the entry is not found.
func testRankOf(t *testing.T, ranked []*noderanking.NodeRankEntry, logicalName string) int {
	t.Helper()
	e := testFindEntry(ranked, logicalName)
	if e == nil {
		t.Fatalf("entry not found in ranked list: %s", logicalName)
	}
	return e.Rank
}

// TC-01: Root only
func TestNodeRankCompute_RootOnly(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{
			LogicalName: "ROOT",
			Frontmatter: &frontmatter.Frontmatter{},
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
	rank := testRankOf(t, ranked, "ROOT")
	if rank != 0 {
		t.Errorf("expected ROOT rank=0, got %d", rank)
	}
}

// TC-02: Linear chain — incrementing ranks
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
		t.Errorf("expected no cycles, got: %v", cycles)
	}

	if testRankOf(t, ranked, "ROOT") != 0 {
		t.Errorf("expected ROOT rank=0")
	}
	if testRankOf(t, ranked, "ROOT/a") != 1 {
		t.Errorf("expected ROOT/a rank=1, got %d", testRankOf(t, ranked, "ROOT/a"))
	}
	if testRankOf(t, ranked, "ROOT/a/b") != 2 {
		t.Errorf("expected ROOT/a/b rank=2, got %d", testRankOf(t, ranked, "ROOT/a/b"))
	}
}

// TC-03: Independent siblings — equal rank
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
		t.Errorf("expected no cycles, got: %v", cycles)
	}

	rankA := testRankOf(t, ranked, "ROOT/a")
	rankB := testRankOf(t, ranked, "ROOT/b")
	if rankA != rankB {
		t.Errorf("expected ROOT/a and ROOT/b to have equal rank, got %d and %d", rankA, rankB)
	}
	if rankA != 1 {
		t.Errorf("expected sibling rank=1, got %d", rankA)
	}
}

// TC-04: depends_on increases rank
func TestNodeRankCompute_DependsOnIncreasesRank(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []*string{testStr("ROOT/a")},
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
	rankB := testRankOf(t, ranked, "ROOT/b")
	if rankB <= rankA {
		t.Errorf("expected ROOT/b rank (%d) > ROOT/a rank (%d)", rankB, rankA)
	}
}

// TC-05: depends_on with qualifier — qualifier stripped
func TestNodeRankCompute_DependsOnQualifierStripped(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []*string{testStr("ROOT/a(interface)")},
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
	rankB := testRankOf(t, ranked, "ROOT/b")
	if rankB <= rankA {
		t.Errorf("expected ROOT/b rank (%d) > ROOT/a rank (%d) after qualifier stripping", rankB, rankA)
	}
}

// TC-06: input artifact adds dependency edge
func TestNodeRankCompute_InputArtifactEdge(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				Outputs: []*frontmatter.FrontmatterOutput{
					{ID: "code", Path: "out.go"},
				},
			},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{
				Input: "ARTIFACT/a(code)",
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

	artifactEntry := testFindEntry(ranked, "ARTIFACT/a(code)")
	if artifactEntry == nil {
		t.Fatalf("expected ARTIFACT/a(code) in ranked list")
	}

	rankA := testRankOf(t, ranked, "ROOT/a")
	rankArtifact := artifactEntry.Rank
	rankB := testRankOf(t, ranked, "ROOT/b")

	if rankArtifact <= rankA {
		t.Errorf("expected ARTIFACT/a(code) rank (%d) > ROOT/a rank (%d)", rankArtifact, rankA)
	}
	if rankB <= rankArtifact {
		t.Errorf("expected ROOT/b rank (%d) > ARTIFACT/a(code) rank (%d)", rankB, rankArtifact)
	}
}

// TC-07: Artifacts get rank one above their node
func TestNodeRankCompute_ArtifactRankOneAboveNode(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				Outputs: []*frontmatter.FrontmatterOutput{
					{ID: "foo", Path: "foo.go"},
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
	rankArtifact := testRankOf(t, ranked, "ARTIFACT/a(foo)")
	if rankArtifact != rankA+1 {
		t.Errorf("expected ARTIFACT/a(foo) rank=%d (ROOT/a+1), got %d", rankA+1, rankArtifact)
	}
}

// TC-08: Multiple outputs — each artifact ranked
func TestNodeRankCompute_MultipleOutputsEachArtifactRanked(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
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
	rankX := testRankOf(t, ranked, "ARTIFACT/a(x)")
	rankY := testRankOf(t, ranked, "ARTIFACT/a(y)")

	if rankX != rankA+1 {
		t.Errorf("expected ARTIFACT/a(x) rank=%d, got %d", rankA+1, rankX)
	}
	if rankY != rankA+1 {
		t.Errorf("expected ARTIFACT/a(y) rank=%d, got %d", rankA+1, rankY)
	}
}

// TC-09: depends_on ARTIFACT reference — used as-is
func TestNodeRankCompute_DependsOnArtifactReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				Outputs: []*frontmatter.FrontmatterOutput{
					{ID: "lib", Path: "lib.go"},
				},
			},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []*string{testStr("ARTIFACT/a(lib)")},
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
	rankArtifact := testRankOf(t, ranked, "ARTIFACT/a(lib)")
	rankB := testRankOf(t, ranked, "ROOT/b")

	if rankArtifact <= rankA {
		t.Errorf("expected ARTIFACT/a(lib) rank (%d) > ROOT/a rank (%d)", rankArtifact, rankA)
	}
	if rankB <= rankArtifact {
		t.Errorf("expected ROOT/b rank (%d) > ARTIFACT/a(lib) rank (%d)", rankB, rankArtifact)
	}
}

// TC-10: Output sorted by rank then logical name
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
		t.Errorf("expected no cycles, got: %v", cycles)
	}

	if len(ranked) < 3 {
		t.Fatalf("expected at least 3 ranked entries, got %d", len(ranked))
	}

	if ranked[0].LogicalName != "ROOT" {
		t.Errorf("expected ROOT first, got %s", ranked[0].LogicalName)
	}

	// Find positions of ROOT/a and ROOT/z
	posA, posZ := -1, -1
	for i, e := range ranked {
		switch e.LogicalName {
		case "ROOT/a":
			posA = i
		case "ROOT/z":
			posZ = i
		}
	}
	if posA == -1 || posZ == -1 {
		t.Fatalf("ROOT/a or ROOT/z not found in ranked list")
	}
	if posA >= posZ {
		t.Errorf("expected ROOT/a (pos %d) before ROOT/z (pos %d) in sorted output", posA, posZ)
	}
}

// TC-11: Parallel entries — equal rank means no dependency
func TestNodeRankCompute_ParallelEqualRank(t *testing.T) {
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
		t.Errorf("expected no cycles, got: %v", cycles)
	}

	rankA := testRankOf(t, ranked, "ROOT/a")
	rankB := testRankOf(t, ranked, "ROOT/b")
	rankC := testRankOf(t, ranked, "ROOT/c")

	if rankA != 1 || rankB != 1 || rankC != 1 {
		t.Errorf("expected ROOT/a, ROOT/b, ROOT/c all rank=1, got %d, %d, %d", rankA, rankB, rankC)
	}
}

// TC-12: Diamond dependency — rank uses max not sum
func TestNodeRankCompute_DiamondDependency(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/c", Frontmatter: &frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []*string{testStr("ROOT/c")},
			},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []*string{testStr("ROOT/c")},
			},
		},
		{
			LogicalName: "ROOT/d",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []*string{testStr("ROOT/a"), testStr("ROOT/b")},
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

	if testRankOf(t, ranked, "ROOT/c") != 1 {
		t.Errorf("expected ROOT/c rank=1, got %d", testRankOf(t, ranked, "ROOT/c"))
	}
	if testRankOf(t, ranked, "ROOT/a") != 2 {
		t.Errorf("expected ROOT/a rank=2, got %d", testRankOf(t, ranked, "ROOT/a"))
	}
	if testRankOf(t, ranked, "ROOT/b") != 2 {
		t.Errorf("expected ROOT/b rank=2, got %d", testRankOf(t, ranked, "ROOT/b"))
	}
	if testRankOf(t, ranked, "ROOT/d") != 3 {
		t.Errorf("expected ROOT/d rank=3, got %d", testRankOf(t, ranked, "ROOT/d"))
	}
}

// TC-13: depends_on outranks parent
func TestNodeRankCompute_DependsOnOutranksParent(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []*string{testStr("ROOT/c")},
			},
		},
		{LogicalName: "ROOT/c", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/c/d", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/c/d/e", Frontmatter: &frontmatter.Frontmatter{}},
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
	if rankAB <= rankA {
		t.Errorf("expected ROOT/a/b rank (%d) > ROOT/a rank (%d)", rankAB, rankA)
	}
}

// TC-14: Multiple depends_on — rank from highest
func TestNodeRankCompute_MultipleDependsOnRankFromHighest(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{LogicalName: "ROOT/a", Frontmatter: &frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []*string{testStr("ROOT/a")},
			},
		},
		{
			LogicalName: "ROOT/c",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []*string{testStr("ROOT/b")},
			},
		},
		{
			LogicalName: "ROOT/d",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []*string{testStr("ROOT/a"), testStr("ROOT/b"), testStr("ROOT/c")},
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

	if testRankOf(t, ranked, "ROOT/a") != 1 {
		t.Errorf("expected ROOT/a rank=1, got %d", testRankOf(t, ranked, "ROOT/a"))
	}
	if testRankOf(t, ranked, "ROOT/b") != 2 {
		t.Errorf("expected ROOT/b rank=2, got %d", testRankOf(t, ranked, "ROOT/b"))
	}
	if testRankOf(t, ranked, "ROOT/c") != 3 {
		t.Errorf("expected ROOT/c rank=3, got %d", testRankOf(t, ranked, "ROOT/c"))
	}
	if testRankOf(t, ranked, "ROOT/d") != 4 {
		t.Errorf("expected ROOT/d rank=4, got %d", testRankOf(t, ranked, "ROOT/d"))
	}
}

// TC-15: Node with both depends_on and input
func TestNodeRankCompute_DependsOnAndInput(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				Outputs: []*frontmatter.FrontmatterOutput{
					{ID: "out", Path: "a.go"},
				},
			},
		},
		{LogicalName: "ROOT/b", Frontmatter: &frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/c",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []*string{testStr("ROOT/b")},
				Input:     "ARTIFACT/a(out)",
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

	rankArtifact := testRankOf(t, ranked, "ARTIFACT/a(out)")
	rankB := testRankOf(t, ranked, "ROOT/b")
	rankC := testRankOf(t, ranked, "ROOT/c")

	// ROOT/c depends on ROOT/b and ARTIFACT/a(out), so its rank must be > both
	if rankC <= rankB {
		t.Errorf("expected ROOT/c rank (%d) > ROOT/b rank (%d)", rankC, rankB)
	}
	if rankC <= rankArtifact {
		t.Errorf("expected ROOT/c rank (%d) > ARTIFACT/a(out) rank (%d)", rankC, rankArtifact)
	}
}

// TC-16: Empty input list
func TestNodeRankCompute_EmptyInput(t *testing.T) {
	ranked, cycles, err := noderanking.NodeRankCompute([]*noderanking.NodeRankInput{})
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

// TC-17: Self-reference
func TestNodeRankCompute_SelfReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []*string{testStr("ROOT/a")},
			},
		},
	}

	_, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Errorf("expected cycles to be non-empty for self-reference")
	}
}

// TC-18: Simple cycle — two nodes
func TestNodeRankCompute_SimpleCycleTwoNodes(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []*string{testStr("ROOT/b")},
			},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []*string{testStr("ROOT/a")},
			},
		},
	}

	_, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Errorf("expected cycles to be non-empty")
	}

	foundAorB := false
	for _, c := range cycles {
		if c == "ROOT/a" || c == "ROOT/b" {
			foundAorB = true
			break
		}
	}
	if !foundAorB {
		t.Errorf("expected cycles to contain ROOT/a or ROOT/b, got: %v", cycles)
	}
}

// TC-19: Cycle through artifacts
func TestNodeRankCompute_CycleThroughArtifacts(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				Outputs:   []*frontmatter.FrontmatterOutput{{ID: "out", Path: "a.go"}},
				DependsOn: []*string{testStr("ARTIFACT/b(out)")},
			},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{
				Outputs:   []*frontmatter.FrontmatterOutput{{ID: "out", Path: "b.go"}},
				DependsOn: []*string{testStr("ARTIFACT/a(out)")},
			},
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

// TC-20: Cycle does not prevent ranking of unrelated nodes
func TestNodeRankCompute_CycleDoesNotBlockUnrelated(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []*string{testStr("ROOT/b")},
			},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []*string{testStr("ROOT/a")},
			},
		},
		{LogicalName: "ROOT/c", Frontmatter: &frontmatter.Frontmatter{}},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Errorf("expected cycles to be non-empty")
	}

	// ROOT and ROOT/c should have valid ranks
	if testRankOf(t, ranked, "ROOT") != 0 {
		t.Errorf("expected ROOT rank=0")
	}
	if testRankOf(t, ranked, "ROOT/c") != 1 {
		t.Errorf("expected ROOT/c rank=1, got %d", testRankOf(t, ranked, "ROOT/c"))
	}

	// cycles should contain ROOT/a or ROOT/b, but not ROOT/c
	for _, c := range cycles {
		if c == "ROOT/c" {
			t.Errorf("ROOT/c should not appear in cycles list")
		}
	}

	foundCycled := false
	for _, c := range cycles {
		if c == "ROOT/a" || c == "ROOT/b" {
			foundCycled = true
			break
		}
	}
	if !foundCycled {
		t.Errorf("expected cycles to reference ROOT/a or ROOT/b, got: %v", cycles)
	}
}

// TC-21: Unresolvable ROOT reference
func TestNodeRankCompute_UnresolvableRootReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []*string{testStr("ROOT/missing")},
			},
		},
	}

	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error for unresolvable ROOT reference, got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got: %v", err)
	}
}

// TC-22: Unresolvable ARTIFACT reference
func TestNodeRankCompute_UnresolvableArtifactReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []*string{testStr("ARTIFACT/missing(id)")},
			},
		},
	}

	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error for unresolvable ARTIFACT reference, got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got: %v", err)
	}
}

// TC-23: Unresolvable input reference
func TestNodeRankCompute_UnresolvableInputReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: &frontmatter.Frontmatter{}},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				Input: "ARTIFACT/missing(id)",
			},
		},
	}

	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error for unresolvable input reference, got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got: %v", err)
	}
}
