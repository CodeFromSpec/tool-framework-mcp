// code-from-spec: ROOT/golang/tests/mcp_tools/validate_specs@mNo8L8ckRf1XbpGRCLFXBkkiRwM
package mcpvalidatespecs_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpvalidatespecs"
)

func testChdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChdir: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("testChdir cleanup: %v", err)
		}
	})
}

func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(testDirOf(path), 0o755); err != nil {
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

func testDirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return "."
}

func testRootNode() string {
	return "# ROOT\n\n# Public\n\n## Context\n\nPublic content.\n"
}

func testLeafNode(logicalName string) string {
	return fmt.Sprintf("# %s\n\nLeaf content.\n", logicalName)
}

func testLeafNodeWithFrontmatter(logicalName string, frontmatter string) string {
	return fmt.Sprintf("---\n%s---\n\n# %s\n\nLeaf content.\n", frontmatter, logicalName)
}

func testComputeChainHash(t *testing.T, logicalName string) string {
	t.Helper()
	chain, err := chainresolver.ChainResolve(logicalName)
	if err != nil {
		t.Fatalf("ChainResolve(%q): %v", logicalName, err)
	}
	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute: %v", err)
	}
	return hash
}

func TestMCPValidateSpecs_CleanTree(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNodeWithFrontmatter("ROOT/a", "output: out/a.go\n"))

	hash := testComputeChainHash(t, "ROOT/a")
	testWriteFile(t, "out/a.go", fmt.Sprintf("// code-from-spec: ROOT/a@%s\n", hash))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) != 0 {
		t.Errorf("expected no format errors, got %d: %+v", len(report.FormatErrors), report.FormatErrors)
	}
	if len(report.Cycles) != 0 {
		t.Errorf("expected no cycles, got %v", report.Cycles)
	}
	if len(report.Staleness) != 0 {
		t.Errorf("expected no staleness entries, got %d: %+v", len(report.Staleness), report.Staleness)
	}
}

func TestMCPValidateSpecs_StaleArtifact(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNodeWithFrontmatter("ROOT/a", "output: out/a.go\n"))
	testWriteFile(t, "out/a.go", "// code-from-spec: ROOT/a@aaaaaaaaaaaaaaaaaaaaaaaaaaa\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 1 {
		t.Fatalf("expected 1 staleness entry, got %d", len(report.Staleness))
	}
	entry := report.Staleness[0]
	if entry.Node != "ROOT/a" {
		t.Errorf("expected Node=ROOT/a, got %q", entry.Node)
	}
	if entry.Status != "stale" {
		t.Errorf("expected Status=stale, got %q", entry.Status)
	}
	_ = entry.Rank
}

func TestMCPValidateSpecs_MissingArtifact(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNodeWithFrontmatter("ROOT/a", "output: out/a.go\n"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 1 {
		t.Fatalf("expected 1 staleness entry, got %d", len(report.Staleness))
	}
	entry := report.Staleness[0]
	if entry.Node != "ROOT/a" {
		t.Errorf("expected Node=ROOT/a, got %q", entry.Node)
	}
	if entry.Status != "missing" {
		t.Errorf("expected Status=missing, got %q", entry.Status)
	}
}

func TestMCPValidateSpecs_MalformedTag(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNodeWithFrontmatter("ROOT/a", "output: out/a.go\n"))
	testWriteFile(t, "out/a.go", "no artifact tag here\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 1 {
		t.Fatalf("expected 1 staleness entry, got %d", len(report.Staleness))
	}
	entry := report.Staleness[0]
	if entry.Node != "ROOT/a" {
		t.Errorf("expected Node=ROOT/a, got %q", entry.Node)
	}
	if entry.Status != "malformed tag" {
		t.Errorf("expected Status=malformed tag, got %q", entry.Status)
	}
}

func TestMCPValidateSpecs_StalenessIncludesRank(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNodeWithFrontmatter("ROOT/a", "output: out/a.go\n"))
	testWriteFile(t, "code-from-spec/b/_node.md", testLeafNodeWithFrontmatter("ROOT/b", "output: out/b.go\ndepends_on:\n  - ROOT/a\n"))
	testWriteFile(t, "out/a.go", "// code-from-spec: ROOT/a@aaaaaaaaaaaaaaaaaaaaaaaaaaa\n")
	testWriteFile(t, "out/b.go", "// code-from-spec: ROOT/b@aaaaaaaaaaaaaaaaaaaaaaaaaaa\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 2 {
		t.Fatalf("expected 2 staleness entries, got %d", len(report.Staleness))
	}

	var rankA, rankB int
	for _, entry := range report.Staleness {
		switch entry.Node {
		case "ROOT/a":
			rankA = entry.Rank
		case "ROOT/b":
			rankB = entry.Rank
		}
	}

	if rankA >= rankB {
		t.Errorf("expected ROOT/a rank (%d) < ROOT/b rank (%d)", rankA, rankB)
	}
}

func TestMCPValidateSpecs_StalenessOrderedByRankThenName(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/z/_node.md", testLeafNodeWithFrontmatter("ROOT/z", "output: out/z.go\n"))
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNodeWithFrontmatter("ROOT/a", "output: out/a.go\n"))
	testWriteFile(t, "out/z.go", "// code-from-spec: ROOT/z@aaaaaaaaaaaaaaaaaaaaaaaaaaa\n")
	testWriteFile(t, "out/a.go", "// code-from-spec: ROOT/a@aaaaaaaaaaaaaaaaaaaaaaaaaaa\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 2 {
		t.Fatalf("expected 2 staleness entries, got %d", len(report.Staleness))
	}

	if report.Staleness[0].Node != "ROOT/a" {
		t.Errorf("expected first entry to be ROOT/a, got %q", report.Staleness[0].Node)
	}
	if report.Staleness[1].Node != "ROOT/z" {
		t.Errorf("expected second entry to be ROOT/z, got %q", report.Staleness[1].Node)
	}
}

func TestMCPValidateSpecs_FormatErrorInvalidDependsOn(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ROOT/missing\n"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) == 0 {
		t.Fatal("expected format errors, got none")
	}

	found := false
	for _, e := range report.FormatErrors {
		if e.Node == "ROOT/a" && e.Rule == "dependency_targets" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected FormatError for ROOT/a with rule=dependency_targets, got %+v", report.FormatErrors)
	}
}

func TestMCPValidateSpecs_FormatErrorFromParseFailure(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", "this is text before any heading\n# ROOT/a\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) == 0 {
		t.Fatal("expected format errors, got none")
	}

	found := false
	for _, e := range report.FormatErrors {
		if e.Node == "ROOT/a" && e.Rule == "parse" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected FormatError for ROOT/a with rule=parse, got %+v", report.FormatErrors)
	}
}

func TestMCPValidateSpecs_ContinuesAfterParseFailure(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", "this is text before any heading\n# ROOT/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md", testLeafNodeWithFrontmatter("ROOT/b", "output: out/b.go\n"))
	testWriteFile(t, "out/b.go", "// code-from-spec: ROOT/b@aaaaaaaaaaaaaaaaaaaaaaaaaaa\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	foundParseError := false
	for _, e := range report.FormatErrors {
		if e.Node == "ROOT/a" && e.Rule == "parse" {
			foundParseError = true
			break
		}
	}
	if !foundParseError {
		t.Errorf("expected parse FormatError for ROOT/a, got %+v", report.FormatErrors)
	}

	foundStaleness := false
	for _, s := range report.Staleness {
		if s.Node == "ROOT/b" {
			foundStaleness = true
			break
		}
	}
	if !foundStaleness {
		t.Errorf("expected StalenessEntry for ROOT/b, got %+v", report.Staleness)
	}
}

func TestMCPValidateSpecs_SimpleCycleDetected(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ROOT/b\n"))
	testWriteFile(t, "code-from-spec/b/_node.md", testLeafNodeWithFrontmatter("ROOT/b", "depends_on:\n  - ROOT/a\n"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Cycles) == 0 {
		t.Fatal("expected cycles to be detected, got none")
	}

	foundAorB := false
	for _, name := range report.Cycles {
		if name == "ROOT/a" || name == "ROOT/b" {
			foundAorB = true
			break
		}
	}
	if !foundAorB {
		t.Errorf("expected cycles to contain ROOT/a or ROOT/b, got %v", report.Cycles)
	}
}

func TestMCPValidateSpecs_RankingSkippedWhenFormatErrors(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ROOT/missing\n"))
	testWriteFile(t, "code-from-spec/b/_node.md", testLeafNodeWithFrontmatter("ROOT/b", "output: out/b.go\n"))
	testWriteFile(t, "out/b.go", "// code-from-spec: ROOT/b@aaaaaaaaaaaaaaaaaaaaaaaaaaa\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) == 0 {
		t.Fatal("expected format errors, got none")
	}

	for _, s := range report.Staleness {
		if s.Node == "ROOT/b" && s.Rank != 0 {
			t.Errorf("expected ROOT/b rank=0 when ranking skipped, got %d", s.Rank)
		}
	}
}

func TestMCPValidateSpecs_EmptySpecTree(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) == 0 {
		t.Fatal("expected format errors for empty spec tree, got none")
	}

	found := false
	for _, e := range report.FormatErrors {
		if e.Rule == "scan" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected FormatError with rule=scan, got %+v", report.FormatErrors)
	}

	if len(report.Cycles) != 0 {
		t.Errorf("expected no cycles, got %v", report.Cycles)
	}
	if len(report.Staleness) != 0 {
		t.Errorf("expected no staleness entries, got %+v", report.Staleness)
	}
}

func TestMCPValidateSpecs_NodeWithNoOutputNotInStaleness(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNode("ROOT/a"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	for _, s := range report.Staleness {
		if s.Node == "ROOT/a" {
			t.Errorf("expected ROOT/a not in staleness (no output), but found entry: %+v", s)
		}
	}
}
