// code-from-spec: ROOT/golang/tests/mcp_tools/validate_specs@H0V1bRwMjKUm7qOxaFLS3QKTXY0
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
	if err := os.MkdirAll(filepath(path), 0755); err != nil {
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

func filepath(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return "."
}

func testComputeChainHash(t *testing.T, logicalName string) string {
	t.Helper()
	chain, err := chainresolver.ChainResolve(logicalName)
	if err != nil {
		t.Fatalf("ChainResolve(%s): %v", logicalName, err)
	}
	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute(%s): %v", logicalName, err)
	}
	return hash
}

func TestMCPValidateSpecs_TC01_CleanTree(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n\n# Public\n\n## Context\n\nSome context.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# SPEC/a\n")

	hash := testComputeChainHash(t, "SPEC/a")
	testWriteFile(t, "out/a.go", fmt.Sprintf("// code-from-spec: SPEC/a@%s\npackage main\n", hash))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) != 0 {
		t.Errorf("expected no format errors, got %d: %v", len(report.FormatErrors), report.FormatErrors)
	}
	if len(report.Cycles) != 0 {
		t.Errorf("expected no cycles, got %v", report.Cycles)
	}
	if len(report.Staleness) != 0 {
		t.Errorf("expected no staleness entries, got %d", len(report.Staleness))
	}
}

func TestMCPValidateSpecs_TC02_StaleArtifact(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# SPEC/a\n")
	testWriteFile(t, "out/a.go", "// code-from-spec: SPEC/a@AAAAAAAAAAAAAAAAAAAAAAAAAAA\npackage main\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) == 0 {
		t.Fatal("expected a staleness entry, got none")
	}
	found := false
	for _, s := range report.Staleness {
		if s.Node == "SPEC/a" {
			found = true
			if s.Status != "stale" {
				t.Errorf("expected status 'stale', got %q", s.Status)
			}
		}
	}
	if !found {
		t.Error("expected staleness entry for SPEC/a")
	}
}

func TestMCPValidateSpecs_TC03_MissingArtifact(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# SPEC/a\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) == 0 {
		t.Fatal("expected a staleness entry, got none")
	}
	found := false
	for _, s := range report.Staleness {
		if s.Node == "SPEC/a" {
			found = true
			if s.Status != "missing" {
				t.Errorf("expected status 'missing', got %q", s.Status)
			}
		}
	}
	if !found {
		t.Error("expected staleness entry for SPEC/a")
	}
}

func TestMCPValidateSpecs_TC04_MalformedTag(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# SPEC/a\n")
	testWriteFile(t, "out/a.go", "package main\n// no artifact tag here\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) == 0 {
		t.Fatal("expected a staleness entry, got none")
	}
	found := false
	for _, s := range report.Staleness {
		if s.Node == "SPEC/a" {
			found = true
			if s.Status != "malformed tag" {
				t.Errorf("expected status 'malformed tag', got %q", s.Status)
			}
		}
	}
	if !found {
		t.Error("expected staleness entry for SPEC/a")
	}
}

func TestMCPValidateSpecs_TC05_StalenessEntriesIncludeRank(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# SPEC/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md", "---\noutput: out/b.go\ndepends_on:\n  - SPEC/a\n---\n# SPEC/b\n")
	testWriteFile(t, "out/a.go", "// code-from-spec: SPEC/a@AAAAAAAAAAAAAAAAAAAAAAAAAAA\npackage main\n")
	testWriteFile(t, "out/b.go", "// code-from-spec: SPEC/b@AAAAAAAAAAAAAAAAAAAAAAAAAAA\npackage main\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	var rankA, rankB int
	foundA, foundB := false, false
	for _, s := range report.Staleness {
		if s.Node == "SPEC/a" {
			rankA = s.Rank
			foundA = true
		}
		if s.Node == "SPEC/b" {
			rankB = s.Rank
			foundB = true
		}
	}
	if !foundA {
		t.Error("expected staleness entry for SPEC/a")
	}
	if !foundB {
		t.Error("expected staleness entry for SPEC/b")
	}
	if foundA && foundB && rankA >= rankB {
		t.Errorf("expected rank of SPEC/a (%d) < rank of SPEC/b (%d)", rankA, rankB)
	}
}

func TestMCPValidateSpecs_TC06_StalenessOrderedByRankThenName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n")
	testWriteFile(t, "code-from-spec/z/_node.md", "---\noutput: out/z.go\n---\n# SPEC/z\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# SPEC/a\n")
	testWriteFile(t, "out/z.go", "// code-from-spec: SPEC/z@AAAAAAAAAAAAAAAAAAAAAAAAAAA\npackage main\n")
	testWriteFile(t, "out/a.go", "// code-from-spec: SPEC/a@AAAAAAAAAAAAAAAAAAAAAAAAAAA\npackage main\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) < 2 {
		t.Fatalf("expected at least 2 staleness entries, got %d", len(report.Staleness))
	}

	idxA, idxZ := -1, -1
	for i, s := range report.Staleness {
		if s.Node == "SPEC/a" {
			idxA = i
		}
		if s.Node == "SPEC/z" {
			idxZ = i
		}
	}
	if idxA == -1 {
		t.Error("expected staleness entry for SPEC/a")
	}
	if idxZ == -1 {
		t.Error("expected staleness entry for SPEC/z")
	}
	if idxA != -1 && idxZ != -1 && idxA >= idxZ {
		t.Errorf("expected SPEC/a (index %d) to appear before SPEC/z (index %d)", idxA, idxZ)
	}
}

func TestMCPValidateSpecs_TC07_FormatErrorInvalidDependsOn(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\ndepends_on:\n  - SPEC/missing\n---\n# SPEC/a\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) == 0 {
		t.Fatal("expected format errors, got none")
	}
	found := false
	for _, fe := range report.FormatErrors {
		if fe.Node == "SPEC/a" && fe.Rule == "dependency_targets" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected FormatError for SPEC/a with rule 'dependency_targets', got: %v", report.FormatErrors)
	}
}

func TestMCPValidateSpecs_TC08_FormatErrorParseFailure(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "this is text before any heading\n# SPEC/a\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) == 0 {
		t.Fatal("expected format errors, got none")
	}
	found := false
	for _, fe := range report.FormatErrors {
		if fe.Node == "SPEC/a" && fe.Rule == "parse" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected FormatError for SPEC/a with rule 'parse', got: %v", report.FormatErrors)
	}
}

func TestMCPValidateSpecs_TC09_ContinuesAfterParseFailure(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "this is text before any heading\n# SPEC/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md", "---\noutput: out/b.go\n---\n# SPEC/b\n")
	testWriteFile(t, "out/b.go", "// code-from-spec: SPEC/b@AAAAAAAAAAAAAAAAAAAAAAAAAAA\npackage main\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	foundParseErr := false
	for _, fe := range report.FormatErrors {
		if fe.Node == "SPEC/a" {
			foundParseErr = true
		}
	}
	if !foundParseErr {
		t.Error("expected FormatError for SPEC/a")
	}

	foundStaleness := false
	for _, s := range report.Staleness {
		if s.Node == "SPEC/b" {
			foundStaleness = true
		}
	}
	if !foundStaleness {
		t.Error("expected staleness entry for SPEC/b")
	}
}

func TestMCPValidateSpecs_TC10_SubdirWithoutNodeMd(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n")
	if err := os.MkdirAll("code-from-spec/b", 0755); err != nil {
		t.Fatalf("mkdir code-from-spec/b: %v", err)
	}

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, fe := range report.FormatErrors {
		if fe.Rule == "missing_node_md" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected FormatError with rule 'missing_node_md', got: %v", report.FormatErrors)
	}
}

func TestMCPValidateSpecs_TC11_UnderscorePrefixedDirIgnored(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n")
	if err := os.MkdirAll("code-from-spec/_tools", 0755); err != nil {
		t.Fatalf("mkdir code-from-spec/_tools: %v", err)
	}

	report := mcpvalidatespecs.MCPValidateSpecs()

	for _, fe := range report.FormatErrors {
		if fe.Rule == "missing_node_md" {
			t.Errorf("unexpected FormatError referencing _tools: %v", fe)
		}
	}
}

func TestMCPValidateSpecs_TC12_SimpleCycle(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\ndepends_on:\n  - SPEC/b\n---\n# SPEC/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md", "---\ndepends_on:\n  - SPEC/a\n---\n# SPEC/b\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Cycles) == 0 {
		t.Error("expected cycles to be non-empty")
	}
	foundCycleNode := false
	for _, name := range report.Cycles {
		if name == "SPEC/a" || name == "SPEC/b" {
			foundCycleNode = true
		}
	}
	if !foundCycleNode {
		t.Errorf("expected cycles to contain SPEC/a or SPEC/b, got: %v", report.Cycles)
	}
}

func TestMCPValidateSpecs_TC13_RankingSkippedWhenFormatErrors(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\ndepends_on:\n  - SPEC/missing\n---\n# SPEC/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md", "---\noutput: out/b.go\n---\n# SPEC/b\n")
	testWriteFile(t, "out/b.go", "// code-from-spec: SPEC/b@AAAAAAAAAAAAAAAAAAAAAAAAAAA\npackage main\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) == 0 {
		t.Error("expected format errors")
	}

	foundStalenessB := false
	for _, s := range report.Staleness {
		if s.Node == "SPEC/b" {
			foundStalenessB = true
			if s.Rank != 0 {
				t.Errorf("expected rank 0 for SPEC/b when ranking skipped, got %d", s.Rank)
			}
		}
	}
	if !foundStalenessB {
		t.Error("expected staleness entry for SPEC/b")
	}
}

func TestMCPValidateSpecs_TC14_EmptySpecTree(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) == 0 {
		t.Fatal("expected format errors, got none")
	}
	found := false
	for _, fe := range report.FormatErrors {
		if fe.Rule == "scan" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected FormatError with rule 'scan', got: %v", report.FormatErrors)
	}
	if len(report.Cycles) != 0 {
		t.Errorf("expected no cycles, got %v", report.Cycles)
	}
	if len(report.Staleness) != 0 {
		t.Errorf("expected no staleness, got %d entries", len(report.Staleness))
	}
}

func TestMCPValidateSpecs_TC15_NodeWithNoOutput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	for _, s := range report.Staleness {
		if s.Node == "SPEC/a" {
			t.Errorf("expected no staleness entry for SPEC/a (no output declared), got: %v", s)
		}
	}
}
