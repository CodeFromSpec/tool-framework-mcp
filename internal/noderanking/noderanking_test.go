// code-from-spec: SPEC/golang/test/cases/spec_tree/ranking@7XQNNcFxJ7HQr6_rDOIIsXoe-0s
package noderanking_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/noderanking"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"
)

func specRef(logicalName string, parentName *string) parsing.CfsReference {
	return parsing.CfsReference{
		NodeType:    parsing.CfsNodeTypeSpec,
		LogicalName: logicalName,
		ParentName:  parentName,
	}
}

func artifactRef(logicalName string, parentName *string, path string) parsing.CfsReference {
	return parsing.CfsReference{
		NodeType:    parsing.CfsNodeTypeArtifact,
		LogicalName: logicalName,
		ParentName:  parentName,
		Path:        path,
	}
}

func specNode(logicalName string, parentName *string, frontmatter *parsing.NodeFrontmatter) parsing.Node {
	return parsing.Node{
		Reference:   specRef(logicalName, parentName),
		Frontmatter: frontmatter,
	}
}

func findRank(ranked []noderanking.NodeRankEntry, logicalName string) (int, bool) {
	for _, e := range ranked {
		if e.Reference.LogicalName == logicalName {
			return e.Rank, true
		}
	}
	return 0, false
}

func containsCycle(cycles []string, logicalName string) bool {
	for _, c := range cycles {
		if c == logicalName {
			return true
		}
	}
	return false
}

func TestNodeRankCompute_RootOnly(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	rank, ok := findRank(ranked, "SPEC/root")
	if !ok {
		t.Fatal("SPEC/root not found in ranked")
	}
	if rank != 0 {
		t.Fatalf("expected rank 0, got %d", rank)
	}
}

func TestNodeRankCompute_LinearChain(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), nil),
		specNode("SPEC/root/a/b", testutils.Ptr("SPEC/root/a"), nil),
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	checks := map[string]int{
		"SPEC/root":     0,
		"SPEC/root/a":   1,
		"SPEC/root/a/b": 2,
	}
	for name, want := range checks {
		got, ok := findRank(ranked, name)
		if !ok {
			t.Fatalf("%s not found in ranked", name)
		}
		if got != want {
			t.Fatalf("%s: expected rank %d, got %d", name, want, got)
		}
	}
}

func TestNodeRankCompute_IndependentSiblings(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), nil),
		specNode("SPEC/root/b", testutils.Ptr("SPEC/root"), nil),
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	ra, ok := findRank(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	rb, ok := findRank(ranked, "SPEC/root/b")
	if !ok {
		t.Fatal("SPEC/root/b not found")
	}
	if ra != rb {
		t.Fatalf("expected equal ranks, got a=%d b=%d", ra, rb)
	}
}

func TestNodeRankCompute_MultipleIndependentRoots(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/alpha", nil, nil),
		specNode("SPEC/beta", nil, nil),
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	ra, ok := findRank(ranked, "SPEC/alpha")
	if !ok {
		t.Fatal("SPEC/alpha not found")
	}
	rb, ok := findRank(ranked, "SPEC/beta")
	if !ok {
		t.Fatal("SPEC/beta not found")
	}
	if ra != 0 {
		t.Fatalf("SPEC/alpha: expected rank 0, got %d", ra)
	}
	if rb != 0 {
		t.Fatalf("SPEC/beta: expected rank 0, got %d", rb)
	}
}

func TestNodeRankCompute_DependsOnIncreasesRank(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), nil),
		specNode("SPEC/root/b", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			DependsOn: []string{"SPEC/root/a"},
		}),
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	ra, ok := findRank(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	rb, ok := findRank(ranked, "SPEC/root/b")
	if !ok {
		t.Fatal("SPEC/root/b not found")
	}
	if rb <= ra {
		t.Fatalf("expected rank of SPEC/root/b > rank of SPEC/root/a, got b=%d a=%d", rb, ra)
	}
}

func TestNodeRankCompute_DependsOnWithQualifier(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), nil),
		specNode("SPEC/root/b", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			DependsOn: []string{"SPEC/root/a(interface)"},
		}),
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	ra, ok := findRank(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	rb, ok := findRank(ranked, "SPEC/root/b")
	if !ok {
		t.Fatal("SPEC/root/b not found")
	}
	if rb <= ra {
		t.Fatalf("expected rank of SPEC/root/b > rank of SPEC/root/a, got b=%d a=%d", rb, ra)
	}
}

func TestNodeRankCompute_ExternalDependsOnSkipped(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			DependsOn: []string{"EXTERNAL/proto/api.proto"},
		}),
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	ra, ok := findRank(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	if ra != 1 {
		t.Fatalf("expected rank 1, got %d", ra)
	}
}

func TestNodeRankCompute_InputArtifactAddsDependencyEdge(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			Output: testutils.Ptr("out.go"),
		}),
		specNode("SPEC/root/b", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			Input: testutils.Ptr("ARTIFACT/root/a"),
		}),
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	ra, ok := findRank(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	rart, ok := findRank(ranked, "ARTIFACT/root/a")
	if !ok {
		t.Fatal("ARTIFACT/root/a not found")
	}
	rb, ok := findRank(ranked, "SPEC/root/b")
	if !ok {
		t.Fatal("SPEC/root/b not found")
	}
	if rart <= ra {
		t.Fatalf("expected rank of ARTIFACT/root/a > rank of SPEC/root/a, got art=%d a=%d", rart, ra)
	}
	if rb <= rart {
		t.Fatalf("expected rank of SPEC/root/b > rank of ARTIFACT/root/a, got b=%d art=%d", rb, rart)
	}
}

func TestNodeRankCompute_SpecInputAddsDependencyEdge(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), nil),
		specNode("SPEC/root/b", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			Input: testutils.Ptr("SPEC/root/a"),
		}),
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	ra, ok := findRank(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	rb, ok := findRank(ranked, "SPEC/root/b")
	if !ok {
		t.Fatal("SPEC/root/b not found")
	}
	if rb <= ra {
		t.Fatalf("expected rank of SPEC/root/b > rank of SPEC/root/a, got b=%d a=%d", rb, ra)
	}
}

func TestNodeRankCompute_ExternalInputSkipped(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			Input: testutils.Ptr("EXTERNAL/docs/spec.yaml"),
		}),
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	ra, ok := findRank(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	if ra != 1 {
		t.Fatalf("expected rank 1, got %d", ra)
	}
}

func TestNodeRankCompute_ArtifactsGetRankOneAboveNode(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			Output: testutils.Ptr("foo.go"),
		}),
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	ra, ok := findRank(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	rart, ok := findRank(ranked, "ARTIFACT/root/a")
	if !ok {
		t.Fatal("ARTIFACT/root/a not found")
	}
	if rart != ra+1 {
		t.Fatalf("expected ARTIFACT/root/a rank = SPEC/root/a rank + 1, got art=%d a=%d", rart, ra)
	}
}

func TestNodeRankCompute_SingleOutputArtifactRanked(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			Output: testutils.Ptr("x.go"),
		}),
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	ra, ok := findRank(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	rart, ok := findRank(ranked, "ARTIFACT/root/a")
	if !ok {
		t.Fatal("ARTIFACT/root/a not found in ranked")
	}
	if rart != ra+1 {
		t.Fatalf("expected ARTIFACT/root/a rank = SPEC/root/a rank + 1, got art=%d a=%d", rart, ra)
	}
}

func TestNodeRankCompute_DependsOnArtifactReference(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			Output: testutils.Ptr("lib.go"),
		}),
		specNode("SPEC/root/b", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			DependsOn: []string{"ARTIFACT/root/a"},
		}),
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	ra, ok := findRank(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	rart, ok := findRank(ranked, "ARTIFACT/root/a")
	if !ok {
		t.Fatal("ARTIFACT/root/a not found")
	}
	rb, ok := findRank(ranked, "SPEC/root/b")
	if !ok {
		t.Fatal("SPEC/root/b not found")
	}
	if rart <= ra {
		t.Fatalf("expected rank of ARTIFACT/root/a > rank of SPEC/root/a, got art=%d a=%d", rart, ra)
	}
	if rb <= rart {
		t.Fatalf("expected rank of SPEC/root/b > rank of ARTIFACT/root/a, got b=%d art=%d", rb, rart)
	}
}

func TestNodeRankCompute_OutputSortedByRankThenLogicalName(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/z", testutils.Ptr("SPEC/root"), nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), nil),
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	if len(ranked) < 3 {
		t.Fatalf("expected at least 3 entries, got %d", len(ranked))
	}
	if ranked[0].Reference.LogicalName != "SPEC/root" {
		t.Fatalf("expected ranked[0] = SPEC/root, got %s", ranked[0].Reference.LogicalName)
	}
	ra, ok := findRank(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	rz, ok := findRank(ranked, "SPEC/root/z")
	if !ok {
		t.Fatal("SPEC/root/z not found")
	}
	if ra != rz {
		t.Fatalf("expected equal ranks, got a=%d z=%d", ra, rz)
	}
	var idxA, idxZ int
	for i, e := range ranked {
		if e.Reference.LogicalName == "SPEC/root/a" {
			idxA = i
		}
		if e.Reference.LogicalName == "SPEC/root/z" {
			idxZ = i
		}
	}
	if idxA >= idxZ {
		t.Fatalf("expected SPEC/root/a before SPEC/root/z in ranked, got idxA=%d idxZ=%d", idxA, idxZ)
	}
}

func TestNodeRankCompute_ParallelEntries(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), nil),
		specNode("SPEC/root/b", testutils.Ptr("SPEC/root"), nil),
		specNode("SPEC/root/c", testutils.Ptr("SPEC/root"), nil),
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	for _, name := range []string{"SPEC/root/a", "SPEC/root/b", "SPEC/root/c"} {
		r, ok := findRank(ranked, name)
		if !ok {
			t.Fatalf("%s not found", name)
		}
		if r != 1 {
			t.Fatalf("%s: expected rank 1, got %d", name, r)
		}
	}
}

func TestNodeRankCompute_DiamondDependency(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/c", testutils.Ptr("SPEC/root"), nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			DependsOn: []string{"SPEC/root/c"},
		}),
		specNode("SPEC/root/b", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			DependsOn: []string{"SPEC/root/c"},
		}),
		specNode("SPEC/root/d", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			DependsOn: []string{"SPEC/root/a", "SPEC/root/b"},
		}),
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	checks := map[string]int{
		"SPEC/root/c": 1,
		"SPEC/root/a": 2,
		"SPEC/root/b": 2,
		"SPEC/root/d": 3,
	}
	for name, want := range checks {
		got, ok := findRank(ranked, name)
		if !ok {
			t.Fatalf("%s not found", name)
		}
		if got != want {
			t.Fatalf("%s: expected rank %d, got %d", name, want, got)
		}
	}
}

func TestNodeRankCompute_DependsOnOutranksParent(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), nil),
		specNode("SPEC/root/c", testutils.Ptr("SPEC/root"), nil),
		specNode("SPEC/root/c/d", testutils.Ptr("SPEC/root/c"), nil),
		specNode("SPEC/root/c/d/e", testutils.Ptr("SPEC/root/c/d"), nil),
		specNode("SPEC/root/a/b", testutils.Ptr("SPEC/root/a"), &parsing.NodeFrontmatter{
			DependsOn: []string{"SPEC/root/c"},
		}),
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	ra, ok := findRank(ranked, "SPEC/root/a")
	if !ok {
		t.Fatal("SPEC/root/a not found")
	}
	rab, ok := findRank(ranked, "SPEC/root/a/b")
	if !ok {
		t.Fatal("SPEC/root/a/b not found")
	}
	if rab <= ra {
		t.Fatalf("expected rank of SPEC/root/a/b > rank of SPEC/root/a, got ab=%d a=%d", rab, ra)
	}
	rc, ok := findRank(ranked, "SPEC/root/c")
	if !ok {
		t.Fatal("SPEC/root/c not found")
	}
	wantRab := 1 + max(ra, rc)
	if rab != wantRab {
		t.Fatalf("expected SPEC/root/a/b rank = 1 + max(%d, %d) = %d, got %d", ra, rc, wantRab, rab)
	}
}

func TestNodeRankCompute_MultipleDependsOnRankFromHighest(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), nil),
		specNode("SPEC/root/b", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			DependsOn: []string{"SPEC/root/a"},
		}),
		specNode("SPEC/root/c", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			DependsOn: []string{"SPEC/root/b"},
		}),
		specNode("SPEC/root/d", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			DependsOn: []string{"SPEC/root/a", "SPEC/root/b", "SPEC/root/c"},
		}),
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	checks := map[string]int{
		"SPEC/root/a": 1,
		"SPEC/root/b": 2,
		"SPEC/root/c": 3,
		"SPEC/root/d": 4,
	}
	for name, want := range checks {
		got, ok := findRank(ranked, name)
		if !ok {
			t.Fatalf("%s not found", name)
		}
		if got != want {
			t.Fatalf("%s: expected rank %d, got %d", name, want, got)
		}
	}
}

func TestNodeRankCompute_BothDependsOnAndInput(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			Output: testutils.Ptr("a.go"),
		}),
		specNode("SPEC/root/b", testutils.Ptr("SPEC/root"), nil),
		specNode("SPEC/root/c", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			DependsOn: []string{"SPEC/root/b"},
			Input:     testutils.Ptr("ARTIFACT/root/a"),
		}),
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	rroot, ok := findRank(ranked, "SPEC/root")
	if !ok {
		t.Fatal("SPEC/root not found")
	}
	rb, ok := findRank(ranked, "SPEC/root/b")
	if !ok {
		t.Fatal("SPEC/root/b not found")
	}
	rart, ok := findRank(ranked, "ARTIFACT/root/a")
	if !ok {
		t.Fatal("ARTIFACT/root/a not found")
	}
	rc, ok := findRank(ranked, "SPEC/root/c")
	if !ok {
		t.Fatal("SPEC/root/c not found")
	}
	want := 1 + max(rroot, rb, rart)
	if rc != want {
		t.Fatalf("expected SPEC/root/c rank = %d, got %d", want, rc)
	}
}

func TestNodeRankCompute_EmptyInput(t *testing.T) {
	ranked, cycles, err := noderanking.NodeRankCompute([]parsing.Node{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	if len(ranked) != 0 {
		t.Fatalf("expected empty ranked, got %v", ranked)
	}
}

func TestNodeRankCompute_SelfReference(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			DependsOn: []string{"SPEC/root/a"},
		}),
	}
	_, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Fatal("expected cycles, got none")
	}
}

func TestNodeRankCompute_SimpleCycleTwoNodes(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			DependsOn: []string{"SPEC/root/b"},
		}),
		specNode("SPEC/root/b", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			DependsOn: []string{"SPEC/root/a"},
		}),
	}
	_, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Fatal("expected cycles, got none")
	}
	hasA := containsCycle(cycles, "SPEC/root/a")
	hasB := containsCycle(cycles, "SPEC/root/b")
	if !hasA && !hasB {
		t.Fatalf("expected cycles to contain SPEC/root/a or SPEC/root/b, got %v", cycles)
	}
}

func TestNodeRankCompute_CycleThroughArtifacts(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			Output:    testutils.Ptr("a.go"),
			DependsOn: []string{"ARTIFACT/root/b"},
		}),
		specNode("SPEC/root/b", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			Output:    testutils.Ptr("b.go"),
			DependsOn: []string{"ARTIFACT/root/a"},
		}),
	}
	_, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Fatal("expected cycles, got none")
	}
}

func TestNodeRankCompute_CycleDoesNotPreventRankingUnrelatedNodes(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			DependsOn: []string{"SPEC/root/b"},
		}),
		specNode("SPEC/root/b", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			DependsOn: []string{"SPEC/root/a"},
		}),
		specNode("SPEC/root/c", testutils.Ptr("SPEC/root"), nil),
	}
	ranked, cycles, err := noderanking.NodeRankCompute(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) == 0 {
		t.Fatal("expected cycles, got none")
	}
	rroot, ok := findRank(ranked, "SPEC/root")
	if !ok {
		t.Fatal("SPEC/root not found")
	}
	if rroot != 0 {
		t.Fatalf("expected SPEC/root rank 0, got %d", rroot)
	}
	rc, ok := findRank(ranked, "SPEC/root/c")
	if !ok {
		t.Fatal("SPEC/root/c not found")
	}
	if rc != 1 {
		t.Fatalf("expected SPEC/root/c rank 1, got %d", rc)
	}
	hasA := containsCycle(cycles, "SPEC/root/a")
	hasB := containsCycle(cycles, "SPEC/root/b")
	if !hasA && !hasB {
		t.Fatalf("expected cycles to contain SPEC/root/a or SPEC/root/b, got %v", cycles)
	}
	if containsCycle(cycles, "SPEC/root/c") {
		t.Fatal("SPEC/root/c should not be in cycles")
	}
}

func TestNodeRankCompute_UnresolvableSpecReference(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			DependsOn: []string{"SPEC/root/missing"},
		}),
	}
	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Fatalf("expected ErrUnresolvableReference, got %v", err)
	}
}

func TestNodeRankCompute_UnresolvableArtifactReference(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			DependsOn: []string{"ARTIFACT/root/missing"},
		}),
	}
	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Fatalf("expected ErrUnresolvableReference, got %v", err)
	}
}

func TestNodeRankCompute_UnresolvableArtifactInputReference(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			Input: testutils.Ptr("ARTIFACT/root/missing"),
		}),
	}
	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Fatalf("expected ErrUnresolvableReference, got %v", err)
	}
}

func TestNodeRankCompute_UnresolvableSpecInputReference(t *testing.T) {
	entries := []parsing.Node{
		specNode("SPEC/root", nil, nil),
		specNode("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
			Input: testutils.Ptr("SPEC/root/missing"),
		}),
	}
	_, _, err := noderanking.NodeRankCompute(entries)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, noderanking.ErrUnresolvableReference) {
		t.Fatalf("expected ErrUnresolvableReference, got %v", err)
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

var _ = artifactRef
