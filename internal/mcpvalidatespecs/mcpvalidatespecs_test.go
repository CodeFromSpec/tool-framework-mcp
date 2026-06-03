// code-from-spec: ROOT/golang/tests/mcp_tools/validate_specs@cjsjcfpMHNZiF3Vw43M9oGJXB4c
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
	return "---\n---\n# ROOT\n\n# Public\n"
}

func testLeafNode(logicalName, output string) string {
	if output != "" {
		return fmt.Sprintf("---\noutput: %s\n---\n# %s\n", output, logicalName)
	}
	return fmt.Sprintf("---\n---\n# %s\n", logicalName)
}

func testLeafNodeWithDeps(logicalName, output string, deps []string) string {
	depsYAML := ""
	if len(deps) > 0 {
		depsYAML = "depends_on:\n"
		for _, d := range deps {
			depsYAML += fmt.Sprintf("  - %s\n", d)
		}
	}
	if output != "" {
		return fmt.Sprintf("---\n%soutput: %s\n---\n# %s\n", depsYAML, output, logicalName)
	}
	return fmt.Sprintf("---\n%s---\n# %s\n", depsYAML, logicalName)
}

func testComputeHash(t *testing.T, logicalName string) string {
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
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNode("ROOT/a", "out/a.go"))

	hash := testComputeHash(t, "ROOT/a")
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
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNode("ROOT/a", "out/a.go"))

	testWriteFile(t, "out/a.go", "// code-from-spec: ROOT/a@aaaaaaaaaaaaaaaaaaaaaaaaaaa\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 1 {
		t.Fatalf("expected 1 staleness entry, got %d", len(report.Staleness))
	}
	entry := report.Staleness[0]
	if entry.Node != "ROOT/a" {
		t.Errorf("expected node ROOT/a, got %q", entry.Node)
	}
	if entry.Status != "stale" {
		t.Errorf("expected status 'stale', got %q", entry.Status)
	}
	if entry.Rank == 0 {
		t.Errorf("expected non-zero rank")
	}
}

func TestMCPValidateSpecs_MissingArtifact(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNode("ROOT/a", "out/a.go"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 1 {
		t.Fatalf("expected 1 staleness entry, got %d", len(report.Staleness))
	}
	entry := report.Staleness[0]
	if entry.Node != "ROOT/a" {
		t.Errorf("expected node ROOT/a, got %q", entry.Node)
	}
	if entry.Status != "missing" {
		t.Errorf("expected status 'missing', got %q", entry.Status)
	}
}

func TestMCPValidateSpecs_MalformedTag(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNode("ROOT/a", "out/a.go"))

	testWriteFile(t, "out/a.go", "package main\n\nfunc main() {}\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 1 {
		t.Fatalf("expected 1 staleness entry, got %d", len(report.Staleness))
	}
	entry := report.Staleness[0]
	if entry.Node != "ROOT/a" {
		t.Errorf("expected node ROOT/a, got %q", entry.Node)
	}
	if entry.Status != "malformed tag" {
		t.Errorf("expected status 'malformed tag', got %q", entry.Status)
	}
}

func TestMCPValidateSpecs_StalenessIncludesRank(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNode("ROOT/a", "out/a.go"))
	testWriteFile(t, "code-from-spec/b/_node.md", testLeafNodeWithDeps("ROOT/b", "out/b.go", []string{"ROOT/a"}))

	testWriteFile(t, "out/a.go", "// code-from-spec: ROOT/a@aaaaaaaaaaaaaaaaaaaaaaaaaaa\n")
	testWriteFile(t, "out/b.go", "// code-from-spec: ROOT/b@aaaaaaaaaaaaaaaaaaaaaaaaaaa\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 2 {
		t.Fatalf("expected 2 staleness entries, got %d", len(report.Staleness))
	}

	rankA := 0
	rankB := 0
	for _, e := range report.Staleness {
		switch e.Node {
		case "ROOT/a":
			rankA = e.Rank
		case "ROOT/b":
			rankB = e.Rank
		}
	}

	if rankA == 0 {
		t.Errorf("expected non-zero rank for ROOT/a")
	}
	if rankB == 0 {
		t.Errorf("expected non-zero rank for ROOT/b")
	}
	if rankA >= rankB {
		t.Errorf("expected ROOT/a rank (%d) < ROOT/b rank (%d)", rankA, rankB)
	}
}

func TestMCPValidateSpecs_StalenessOrderedByRankThenName(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/z/_node.md", testLeafNode("ROOT/z", "out/z.go"))
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNode("ROOT/a", "out/a.go"))

	testWriteFile(t, "out/z.go", "// code-from-spec: ROOT/z@aaaaaaaaaaaaaaaaaaaaaaaaaaa\n")
	testWriteFile(t, "out/a.go", "// code-from-spec: ROOT/a@aaaaaaaaaaaaaaaaaaaaaaaaaaa\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 2 {
		t.Fatalf("expected 2 staleness entries, got %d", len(report.Staleness))
	}

	if report.Staleness[0].Node != "ROOT/a" {
		t.Errorf("expected ROOT/a first, got %q", report.Staleness[0].Node)
	}
	if report.Staleness[1].Node != "ROOT/z" {
		t.Errorf("expected ROOT/z second, got %q", report.Staleness[1].Node)
	}
}

func TestMCPValidateSpecs_FormatErrorInvalidDependsOn(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNodeWithDeps("ROOT/a", "", []string{"ROOT/missing"}))

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, e := range report.FormatErrors {
		if e.Node == "ROOT/a" && e.Rule == "dependency_targets" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected FormatError for ROOT/a with rule 'dependency_targets', got: %+v", report.FormatErrors)
	}
}

func TestMCPValidateSpecs_FormatErrorParseFailure(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", "this is invalid content before any heading\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, e := range report.FormatErrors {
		if e.Node == "ROOT/a" && e.Rule == "parse" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected FormatError for ROOT/a with rule 'parse', got: %+v", report.FormatErrors)
	}
}

func TestMCPValidateSpecs_ContinuesAfterParseFailure(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", "this is invalid content before any heading\n")
	testWriteFile(t, "code-from-spec/b/_node.md", testLeafNode("ROOT/b", "out/b.go"))

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
		t.Errorf("expected FormatError for ROOT/a with rule 'parse', got: %+v", report.FormatErrors)
	}

	foundStaleness := false
	for _, s := range report.Staleness {
		if s.Node == "ROOT/b" {
			foundStaleness = true
			break
		}
	}
	if !foundStaleness {
		t.Errorf("expected StalenessEntry for ROOT/b, got: %+v", report.Staleness)
	}
}

func TestMCPValidateSpecs_SimpleCycle(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNodeWithDeps("ROOT/a", "", []string{"ROOT/b"}))
	testWriteFile(t, "code-from-spec/b/_node.md", testLeafNodeWithDeps("ROOT/b", "", []string{"ROOT/a"}))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Cycles) == 0 {
		t.Errorf("expected non-empty cycles, got none")
	}

	foundCycle := false
	for _, c := range report.Cycles {
		if c == "ROOT/a" || c == "ROOT/b" {
			foundCycle = true
			break
		}
	}
	if !foundCycle {
		t.Errorf("expected cycle to contain ROOT/a or ROOT/b, got: %v", report.Cycles)
	}
}

func TestMCPValidateSpecs_RankingSkippedWhenFormatErrors(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNodeWithDeps("ROOT/a", "", []string{"ROOT/missing"}))
	testWriteFile(t, "code-from-spec/b/_node.md", testLeafNode("ROOT/b", "out/b.go"))

	testWriteFile(t, "out/b.go", "// code-from-spec: ROOT/b@aaaaaaaaaaaaaaaaaaaaaaaaaaa\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) == 0 {
		t.Errorf("expected format errors, got none")
	}

	for _, s := range report.Staleness {
		if s.Node == "ROOT/b" && s.Rank != 0 {
			t.Errorf("expected rank 0 for ROOT/b when format errors exist, got %d", s.Rank)
		}
	}
}

func TestMCPValidateSpecs_EmptySpecTree(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, e := range report.FormatErrors {
		if e.Rule == "scan" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected FormatError with rule 'scan', got: %+v", report.FormatErrors)
	}

	if len(report.Cycles) != 0 {
		t.Errorf("expected no cycles, got %v", report.Cycles)
	}
	if len(report.Staleness) != 0 {
		t.Errorf("expected no staleness entries, got %+v", report.Staleness)
	}
}

func TestMCPValidateSpecs_NodeWithNoOutput(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNode("ROOT/a", ""))

	report := mcpvalidatespecs.MCPValidateSpecs()

	for _, s := range report.Staleness {
		if s.Node == "ROOT/a" {
			t.Errorf("expected no staleness entry for ROOT/a (no output), got one: %+v", s)
		}
	}
}
