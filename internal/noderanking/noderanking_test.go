// code-from-spec: ROOT/golang/tests/utils/node_ranking@FaAiplys0jIo2PKX4Fp4S2mvdTQ
package noderanking_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/noderanking"
)

// testEmptyFrontmatter returns a Frontmatter with all zero/empty fields.
func testEmptyFrontmatter() *frontmatter.Frontmatter {
	return &frontmatter.Frontmatter{
		DependsOn: []string{},
		External:  []*frontmatter.FrontmatterExternal{},
		Input:     "",
		Outputs:   []*frontmatter.FrontmatterOutput{},
	}
}

// testFindEntry finds a NodeRankEntry by logical name in the ranked slice.
// Returns nil if not found.
func testFindEntry(ranked []*noderanking.NodeRankEntry, logicalName string) *noderanking.NodeRankEntry {
	for _, e := range ranked {
		if e.LogicalName == logicalName {
			return e
		}
	}
	return nil
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

// TC-01: Root only — single entry with no dependencies should receive rank 0.
func TestNodeRankCompute_TC01_RootOnly(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: testEmptyFrontmatter()},
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
	entry := testFindEntry(ranked, "ROOT")
	if entry == nil {
		t.Fatal("expected entry for ROOT")
	}
	if entry.Rank != 0 {
		t.Errorf("expected rank 0 for ROOT, got %d", entry.Rank)
	}
}

// TC-02: Linear chain — parent-child chain should produce incrementing ranks.
func TestNodeRankCompute_TC02_LinearChain(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: testEmptyFrontmatter()},
		{LogicalName: "ROOT/a", Frontmatter: testEmptyFrontmatter()},
		{LogicalName: "ROOT/a/b", Frontmatter: testEmptyFrontmatter()},
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
		t.Fatal("missing expected entries in ranked output")
	}
	if root.Rank != 0 {
		t.Errorf("expected ROOT rank 0, got %d", root.Rank)
	}
	if a.Rank != 1 {
		t.Errorf("expected ROOT/a rank 1, got %d", a.Rank)
	}
	if ab.Rank != 2 {
		t.Errorf("expected ROOT/a/b rank 2, got %d", ab.Rank)
	}
}

// TC-03: Independent siblings — sibling nodes should receive the same rank.
func TestNodeRankCompute_TC03_IndependentSiblings(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: testEmptyFrontmatter()},
		{LogicalName: "ROOT/a", Frontmatter: testEmptyFrontmatter()},
		{LogicalName: "ROOT/b", Frontmatter: testEmptyFrontmatter()},
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
	b := testFindEntry(ranked, "ROOT/b")

	if root == nil || a == nil || b == nil {
		t.Fatal("missing expected entries")
	}
	if root.Rank != 0 {
		t.Errorf("expected ROOT rank 0, got %d", root.Rank)
	}
	if a.Rank != b.Rank {
		t.Errorf("expected ROOT/a and ROOT/b to have equal rank, got %d and %d", a.Rank, b.Rank)
	}
}

// TC-04: depends_on increases rank — node depending on sibling must rank higher.
func TestNodeRankCompute_TC04_DependsOnIncreasesRank(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: testEmptyFrontmatter()},
		{LogicalName: "ROOT/a", Frontmatter: testEmptyFrontmatter()},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/a"},
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

	a := testFindEntry(ranked, "ROOT/a")
	b := testFindEntry(ranked, "ROOT/b")

	if a == nil || b == nil {
		t.Fatal("missing expected entries")
	}
	if b.Rank <= a.Rank {
		t.Errorf("expected ROOT/b rank (%d) > ROOT/a rank (%d)", b.Rank, a.Rank)
	}
}

// TC-05: depends_on with qualifier — qualifier stripped before resolution.
func TestNodeRankCompute_TC05_DependsOnQualifierStripped(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: testEmptyFrontmatter()},
		{LogicalName: "ROOT/a", Frontmatter: testEmptyFrontmatter()},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/a(interface)"},
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

	a := testFindEntry(ranked, "ROOT/a")
	b := testFindEntry(ranked, "ROOT/b")

	if a == nil || b == nil {
		t.Fatal("missing expected entries")
	}
	if b.Rank <= a.Rank {
		t.Errorf("expected ROOT/b rank (%d) > ROOT/a rank (%d)", b.Rank, a.Rank)
	}
}

// TC-06: input artifact adds dependency edge.
func TestNodeRankCompute_TC06_InputArtifactAddsEdge(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: testEmptyFrontmatter()},
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
		t.Errorf("expected no cycles, got %v", cycles)
	}

	a := testFindEntry(ranked, "ROOT/a")
	art := testFindEntry(ranked, "ARTIFACT/a(code)")
	b := testFindEntry(ranked, "ROOT/b")

	if a == nil || art == nil || b == nil {
		t.Fatalf("missing expected entries; got ranked: %v", ranked)
	}
	if art.Rank <= a.Rank {
		t.Errorf("expected ARTIFACT/a(code) rank (%d) > ROOT/a rank (%d)", art.Rank, a.Rank)
	}
	if b.Rank <= art.Rank {
		t.Errorf("expected ROOT/b rank (%d) > ARTIFACT/a(code) rank (%d)", b.Rank, art.Rank)
	}
}

// TC-07: Artifacts get rank one above their node.
func TestNodeRankCompute_TC07_ArtifactRankOneAboveNode(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: testEmptyFrontmatter()},
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
		t.Errorf("expected no cycles, got %v", cycles)
	}

	a := testFindEntry(ranked, "ROOT/a")
	art := testFindEntry(ranked, "ARTIFACT/a(foo)")

	if a == nil || art == nil {
		t.Fatalf("missing expected entries; got: %v", ranked)
	}
	if art.Rank != a.Rank+1 {
		t.Errorf("expected ARTIFACT/a(foo) rank = ROOT/a rank + 1, got artifact=%d node=%d", art.Rank, a.Rank)
	}
}

// TC-08: Multiple outputs — each artifact ranked.
func TestNodeRankCompute_TC08_MultipleOutputsEachArtifactRanked(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: testEmptyFrontmatter()},
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
		t.Errorf("expected no cycles, got %v", cycles)
	}

	a := testFindEntry(ranked, "ROOT/a")
	artX := testFindEntry(ranked, "ARTIFACT/a(x)")
	artY := testFindEntry(ranked, "ARTIFACT/a(y)")

	if a == nil || artX == nil || artY == nil {
		t.Fatalf("missing expected entries; got: %v", ranked)
	}
	if artX.Rank != a.Rank+1 {
		t.Errorf("expected ARTIFACT/a(x) rank = ROOT/a rank + 1, got artifact=%d node=%d", artX.Rank, a.Rank)
	}
	if artY.Rank != a.Rank+1 {
		t.Errorf("expected ARTIFACT/a(y) rank = ROOT/a rank + 1, got artifact=%d node=%d", artY.Rank, a.Rank)
	}
}

// TC-09: depends_on ARTIFACT reference — used as-is.
func TestNodeRankCompute_TC09_DependsOnArtifactReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: testEmptyFrontmatter()},
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
				DependsOn: []string{"ARTIFACT/a(lib)"},
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

	a := testFindEntry(ranked, "ROOT/a")
	art := testFindEntry(ranked, "ARTIFACT/a(lib)")
	b := testFindEntry(ranked, "ROOT/b")

	if a == nil || art == nil || b == nil {
		t.Fatalf("missing expected entries; got: %v", ranked)
	}
	if art.Rank <= a.Rank {
		t.Errorf("expected ARTIFACT/a(lib) rank (%d) > ROOT/a rank (%d)", art.Rank, a.Rank)
	}
	if b.Rank <= art.Rank {
		t.Errorf("expected ROOT/b rank (%d) > ARTIFACT/a(lib) rank (%d)", b.Rank, art.Rank)
	}
}

// TC-10: Output sorted by rank then logical name.
func TestNodeRankCompute_TC10_OutputSortedByRankThenName(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: testEmptyFrontmatter()},
		{LogicalName: "ROOT/z", Frontmatter: testEmptyFrontmatter()},
		{LogicalName: "ROOT/a", Frontmatter: testEmptyFrontmatter()},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}

	if len(ranked) == 0 {
		t.Fatal("expected non-empty ranked list")
	}
	if ranked[0].LogicalName != "ROOT" {
		t.Errorf("expected first entry to be ROOT, got %s", ranked[0].LogicalName)
	}

	// Find positions of ROOT/a and ROOT/z
	posA := -1
	posZ := -1
	for i, e := range ranked {
		if e.LogicalName == "ROOT/a" {
			posA = i
		}
		if e.LogicalName == "ROOT/z" {
			posZ = i
		}
	}
	if posA == -1 || posZ == -1 {
		t.Fatal("ROOT/a or ROOT/z missing from ranked")
	}
	if posA >= posZ {
		t.Errorf("expected ROOT/a (pos %d) before ROOT/z (pos %d) alphabetically", posA, posZ)
	}
}

// TC-11: Parallel entries — multiple siblings all share the same rank.
func TestNodeRankCompute_TC11_ParallelEntriesEqualRank(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: testEmptyFrontmatter()},
		{LogicalName: "ROOT/a", Frontmatter: testEmptyFrontmatter()},
		{LogicalName: "ROOT/b", Frontmatter: testEmptyFrontmatter()},
		{LogicalName: "ROOT/c", Frontmatter: testEmptyFrontmatter()},
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
		t.Fatal("missing expected entries")
	}
	if a.Rank != b.Rank || b.Rank != c.Rank {
		t.Errorf("expected ROOT/a, ROOT/b, ROOT/c to have equal rank, got %d, %d, %d", a.Rank, b.Rank, c.Rank)
	}
}

// TC-12: Diamond dependency — rank uses max not sum.
func TestNodeRankCompute_TC12_DiamondDependencyMaxNotSum(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: testEmptyFrontmatter()},
		{LogicalName: "ROOT/c", Frontmatter: testEmptyFrontmatter()},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/c"},
			},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/c"},
			},
		},
		{
			LogicalName: "ROOT/d",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/a", "ROOT/b"},
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

	c := testFindEntry(ranked, "ROOT/c")
	a := testFindEntry(ranked, "ROOT/a")
	b := testFindEntry(ranked, "ROOT/b")
	d := testFindEntry(ranked, "ROOT/d")

	if c == nil || a == nil || b == nil || d == nil {
		t.Fatal("missing expected entries")
	}
	if c.Rank != 1 {
		t.Errorf("expected ROOT/c rank 1, got %d", c.Rank)
	}
	if a.Rank != 2 {
		t.Errorf("expected ROOT/a rank 2, got %d", a.Rank)
	}
	if b.Rank != 2 {
		t.Errorf("expected ROOT/b rank 2, got %d", b.Rank)
	}
	if d.Rank != 3 {
		t.Errorf("expected ROOT/d rank 3, got %d", d.Rank)
	}
}

// TC-13: depends_on outranks parent.
func TestNodeRankCompute_TC13_DependsOnOutranksParent(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: testEmptyFrontmatter()},
		{LogicalName: "ROOT/a", Frontmatter: testEmptyFrontmatter()},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/c"},
			},
		},
		{LogicalName: "ROOT/c", Frontmatter: testEmptyFrontmatter()},
		{LogicalName: "ROOT/c/d", Frontmatter: testEmptyFrontmatter()},
		{LogicalName: "ROOT/c/d/e", Frontmatter: testEmptyFrontmatter()},
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
		t.Fatal("missing expected entries")
	}
	if ab.Rank <= a.Rank {
		t.Errorf("expected ROOT/a/b rank (%d) > ROOT/a rank (%d)", ab.Rank, a.Rank)
	}

	expectedRank := 1 + max(a.Rank, c.Rank)
	if ab.Rank != expectedRank {
		t.Errorf("expected ROOT/a/b rank = 1 + max(rank(ROOT/a), rank(ROOT/c)) = %d, got %d", expectedRank, ab.Rank)
	}
}

// TC-14: Multiple depends_on — rank from highest.
func TestNodeRankCompute_TC14_MultipleDependsOnRankFromHighest(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: testEmptyFrontmatter()},
		{LogicalName: "ROOT/a", Frontmatter: testEmptyFrontmatter()},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/a"},
			},
		},
		{
			LogicalName: "ROOT/c",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/b"},
			},
		},
		{
			LogicalName: "ROOT/d",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/a", "ROOT/b", "ROOT/c"},
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

	a := testFindEntry(ranked, "ROOT/a")
	b := testFindEntry(ranked, "ROOT/b")
	c := testFindEntry(ranked, "ROOT/c")
	d := testFindEntry(ranked, "ROOT/d")

	if a == nil || b == nil || c == nil || d == nil {
		t.Fatal("missing expected entries")
	}
	if a.Rank != 1 {
		t.Errorf("expected ROOT/a rank 1, got %d", a.Rank)
	}
	if b.Rank != 2 {
		t.Errorf("expected ROOT/b rank 2, got %d", b.Rank)
	}
	if c.Rank != 3 {
		t.Errorf("expected ROOT/c rank 3, got %d", c.Rank)
	}
	if d.Rank != 4 {
		t.Errorf("expected ROOT/d rank 4, got %d", d.Rank)
	}
}

// TC-15: Node with both depends_on and input.
func TestNodeRankCompute_TC15_BothDependsOnAndInput(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: testEmptyFrontmatter()},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				Outputs: []*frontmatter.FrontmatterOutput{
					{ID: "out", Path: "a.go"},
				},
			},
		},
		{LogicalName: "ROOT/b", Frontmatter: testEmptyFrontmatter()},
		{
			LogicalName: "ROOT/c",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/b"},
				Input:     "ARTIFACT/a(out)",
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
	art := testFindEntry(ranked, "ARTIFACT/a(out)")
	c := testFindEntry(ranked, "ROOT/c")

	if root == nil || b == nil || art == nil || c == nil {
		t.Fatalf("missing expected entries; got: %v", ranked)
	}

	expectedRank := 1 + max(root.Rank, b.Rank, art.Rank)
	if c.Rank != expectedRank {
		t.Errorf("expected ROOT/c rank = 1 + max(rank(ROOT), rank(ROOT/b), rank(ARTIFACT/a(out))) = %d, got %d", expectedRank, c.Rank)
	}
}

// TC-16: Empty input list — should return empty results with no error.
func TestNodeRankCompute_TC16_EmptyInputList(t *testing.T) {
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

// TC-17: Self-reference — node that lists itself in depends_on forms a cycle.
func TestNodeRankCompute_TC17_SelfReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: testEmptyFrontmatter()},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/a"},
			},
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

// TC-18: Simple cycle — two nodes that depend on each other.
func TestNodeRankCompute_TC18_SimpleCycleTwoNodes(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: testEmptyFrontmatter()},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/b"},
			},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/a"},
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
	if !testContains(cycles, "ROOT/a") && !testContains(cycles, "ROOT/b") {
		t.Errorf("expected cycles to contain ROOT/a or ROOT/b, got %v", cycles)
	}
}

// TC-19: Cycle through artifacts.
func TestNodeRankCompute_TC19_CycleThroughArtifacts(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: testEmptyFrontmatter()},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				Outputs:   []*frontmatter.FrontmatterOutput{{ID: "out", Path: "a.go"}},
				DependsOn: []string{"ARTIFACT/b(out)"},
			},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{
				Outputs:   []*frontmatter.FrontmatterOutput{{ID: "out", Path: "b.go"}},
				DependsOn: []string{"ARTIFACT/a(out)"},
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
func TestNodeRankCompute_TC20_CycleDoesNotPreventUnrelatedRanking(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: testEmptyFrontmatter()},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/b"},
			},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/a"},
			},
		},
		{LogicalName: "ROOT/c", Frontmatter: testEmptyFrontmatter()},
	}

	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Error("expected non-empty cycles")
	}

	root := testFindEntry(ranked, "ROOT")
	c := testFindEntry(ranked, "ROOT/c")

	if root == nil || c == nil {
		t.Fatal("expected ROOT and ROOT/c in ranked output")
	}
	if root.Rank != 0 {
		t.Errorf("expected ROOT rank 0, got %d", root.Rank)
	}
	if c.Rank != 1 {
		t.Errorf("expected ROOT/c rank 1, got %d", c.Rank)
	}
	if testContains(cycles, "ROOT/c") {
		t.Error("ROOT/c should not be in cycles")
	}
}

// TC-21: Unresolvable ROOT reference.
func TestNodeRankCompute_TC21_UnresolvableRootReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: testEmptyFrontmatter()},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/missing"},
			},
		},
	}

	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got %v", err)
	}
}

// TC-22: Unresolvable ARTIFACT reference.
func TestNodeRankCompute_TC22_UnresolvableArtifactReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: testEmptyFrontmatter()},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ARTIFACT/missing(id)"},
			},
		},
	}

	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got %v", err)
	}
}

// TC-23: Unresolvable input reference.
func TestNodeRankCompute_TC23_UnresolvableInputReference(t *testing.T) {
	entries := []*noderanking.NodeRankInput{
		{LogicalName: "ROOT", Frontmatter: testEmptyFrontmatter()},
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				Input: "ARTIFACT/missing(id)",
			},
		},
	}

	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Errorf("expected ErrUnresolvableReference, got %v", err)
	}
}

// max returns the maximum of two or more integers.
func max(a int, rest ...int) int {
	m := a
	for _, v := range rest {
		if v > m {
			m = v
		}
	}
	return m
}
