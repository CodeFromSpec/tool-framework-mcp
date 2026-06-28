// code-from-spec: SPEC/golang/tests/mcp_tools/validate_specs@cvsnBWjEkY5ZsbdQEVh1qoB2NK8
package mcpvalidatespecs_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcpvalidatespecs"
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

func testMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("testMkdirAll %q: %v", path, err)
	}
}

func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("testWriteFile %q: %v", path, err)
	}
}

func testComputeChainHash(t *testing.T, logicalName string) string {
	t.Helper()
	chain, err := chainresolver.ChainResolve(logicalName)
	if err != nil {
		t.Fatalf("ChainResolve(%q): %v", logicalName, err)
	}
	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute(%q): %v", logicalName, err)
	}
	return hash
}

func testRootNode() string {
	return "# SPEC\n\n# Public\n\n## Context\n\nRoot context.\n"
}

func TestMCPValidateSpecs_CleanTree(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMkdirAll(t, "code-from-spec/a")
	testMkdirAll(t, "out")
	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutput: \"out/a.go\"\n---\n# SPEC/a\n")

	hash := testComputeChainHash(t, "SPEC/a")
	testWriteFile(t, "out/a.go",
		fmt.Sprintf("// code-from-spec: SPEC/a@%s\npackage main\n", hash))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) != 0 {
		t.Errorf("expected no format errors, got %d: %+v", len(report.FormatErrors), report.FormatErrors)
	}
	if len(report.Cycles) != 0 {
		t.Errorf("expected no cycles, got %v", report.Cycles)
	}
	if len(report.Staleness) != 0 {
		t.Errorf("expected no staleness entries, got %+v", report.Staleness)
	}
}

func TestMCPValidateSpecs_StaleArtifact(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMkdirAll(t, "code-from-spec/a")
	testMkdirAll(t, "out")
	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutput: \"out/a.go\"\n---\n# SPEC/a\n")

	testWriteFile(t, "out/a.go",
		"// code-from-spec: SPEC/a@AAAAAAAAAAAAAAAAAAAAAAAAAAA\npackage main\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 1 {
		t.Fatalf("expected 1 staleness entry, got %d: %+v", len(report.Staleness), report.Staleness)
	}
	entry := report.Staleness[0]
	if entry.Node != "SPEC/a" {
		t.Errorf("expected node SPEC/a, got %q", entry.Node)
	}
	if entry.Status != "stale" {
		t.Errorf("expected status stale, got %q", entry.Status)
	}
	_ = entry.Rank
}

func TestMCPValidateSpecs_MissingArtifact(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMkdirAll(t, "code-from-spec/a")
	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutput: \"out/a.go\"\n---\n# SPEC/a\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 1 {
		t.Fatalf("expected 1 staleness entry, got %d: %+v", len(report.Staleness), report.Staleness)
	}
	entry := report.Staleness[0]
	if entry.Node != "SPEC/a" {
		t.Errorf("expected node SPEC/a, got %q", entry.Node)
	}
	if entry.Status != "missing" {
		t.Errorf("expected status missing, got %q", entry.Status)
	}
}

func TestMCPValidateSpecs_MalformedTag(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMkdirAll(t, "code-from-spec/a")
	testMkdirAll(t, "out")
	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutput: \"out/a.go\"\n---\n# SPEC/a\n")

	testWriteFile(t, "out/a.go", "package main\n// no artifact tag here\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 1 {
		t.Fatalf("expected 1 staleness entry, got %d: %+v", len(report.Staleness), report.Staleness)
	}
	entry := report.Staleness[0]
	if entry.Node != "SPEC/a" {
		t.Errorf("expected node SPEC/a, got %q", entry.Node)
	}
	if entry.Status != "malformed tag" {
		t.Errorf("expected status 'malformed tag', got %q", entry.Status)
	}
}

func TestMCPValidateSpecs_StalenessEntriesIncludeRank(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMkdirAll(t, "code-from-spec/a")
	testMkdirAll(t, "code-from-spec/b")
	testMkdirAll(t, "out")
	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutput: \"out/a.go\"\n---\n# SPEC/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md",
		"---\noutput: \"out/b.go\"\ndepends_on:\n  - SPEC/a\n---\n# SPEC/b\n")

	testWriteFile(t, "out/a.go",
		"// code-from-spec: SPEC/a@AAAAAAAAAAAAAAAAAAAAAAAAAAA\npackage main\n")
	testWriteFile(t, "out/b.go",
		"// code-from-spec: SPEC/b@AAAAAAAAAAAAAAAAAAAAAAAAAAA\npackage main\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 2 {
		t.Fatalf("expected 2 staleness entries, got %d: %+v", len(report.Staleness), report.Staleness)
	}

	var rankA, rankB int
	foundA, foundB := false, false
	for _, entry := range report.Staleness {
		switch entry.Node {
		case "SPEC/a":
			rankA = entry.Rank
			foundA = true
		case "SPEC/b":
			rankB = entry.Rank
			foundB = true
		}
	}

	if !foundA {
		t.Error("missing staleness entry for SPEC/a")
	}
	if !foundB {
		t.Error("missing staleness entry for SPEC/b")
	}
	if foundA && foundB && rankA >= rankB {
		t.Errorf("expected rank(SPEC/a) < rank(SPEC/b), got %d >= %d", rankA, rankB)
	}
}

func TestMCPValidateSpecs_StalenessOrderedByRankThenName(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMkdirAll(t, "code-from-spec/z")
	testMkdirAll(t, "code-from-spec/a")
	testMkdirAll(t, "out")
	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/z/_node.md",
		"---\noutput: \"out/z.go\"\n---\n# SPEC/z\n")
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\noutput: \"out/a.go\"\n---\n# SPEC/a\n")

	testWriteFile(t, "out/z.go",
		"// code-from-spec: SPEC/z@AAAAAAAAAAAAAAAAAAAAAAAAAAA\npackage main\n")
	testWriteFile(t, "out/a.go",
		"// code-from-spec: SPEC/a@AAAAAAAAAAAAAAAAAAAAAAAAAAA\npackage main\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 2 {
		t.Fatalf("expected 2 staleness entries, got %d: %+v", len(report.Staleness), report.Staleness)
	}

	if report.Staleness[0].Node != "SPEC/a" {
		t.Errorf("expected first entry to be SPEC/a, got %q", report.Staleness[0].Node)
	}
	if report.Staleness[1].Node != "SPEC/z" {
		t.Errorf("expected second entry to be SPEC/z, got %q", report.Staleness[1].Node)
	}
}

func TestMCPValidateSpecs_FormatErrorDependencyTargets(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMkdirAll(t, "code-from-spec/a")
	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\ndepends_on:\n  - SPEC/missing\n---\n# SPEC/a\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, e := range report.FormatErrors {
		if e.Node == "SPEC/a" && e.Rule == "dependency_targets" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected format error for SPEC/a with rule=dependency_targets, got: %+v", report.FormatErrors)
	}
}

func TestMCPValidateSpecs_FormatErrorParse(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMkdirAll(t, "code-from-spec/a")
	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md",
		"this is plain text before any heading\nno level-1 heading here\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, e := range report.FormatErrors {
		if e.Node == "SPEC/a" && e.Rule == "parse" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected format error for SPEC/a with rule=parse, got: %+v", report.FormatErrors)
	}
}

func TestMCPValidateSpecs_ContinuesAfterParseFailure(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMkdirAll(t, "code-from-spec/a")
	testMkdirAll(t, "code-from-spec/b")
	testMkdirAll(t, "out")
	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md",
		"this is plain text before any heading\nno level-1 heading here\n")
	testWriteFile(t, "code-from-spec/b/_node.md",
		"---\noutput: \"out/b.go\"\n---\n# SPEC/b\n")

	testWriteFile(t, "out/b.go",
		"// code-from-spec: SPEC/b@AAAAAAAAAAAAAAAAAAAAAAAAAAA\npackage main\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	foundParseErr := false
	for _, e := range report.FormatErrors {
		if e.Node == "SPEC/a" && e.Rule == "parse" {
			foundParseErr = true
			break
		}
	}
	if !foundParseErr {
		t.Errorf("expected parse format error for SPEC/a, got: %+v", report.FormatErrors)
	}

	foundStale := false
	for _, s := range report.Staleness {
		if s.Node == "SPEC/b" && s.Status == "stale" {
			foundStale = true
			break
		}
	}
	if !foundStale {
		t.Errorf("expected stale entry for SPEC/b, got: %+v", report.Staleness)
	}
}

func TestMCPValidateSpecs_SubdirWithoutNodeMd(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMkdirAll(t, "code-from-spec/a")
	testMkdirAll(t, "code-from-spec/b")
	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md",
		"# SPEC/a\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, e := range report.FormatErrors {
		if e.Rule == "missing_node_md" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected format error with rule=missing_node_md for code-from-spec/b/, got: %+v", report.FormatErrors)
	}
}

func TestMCPValidateSpecs_UnderscorePrefixedDirIgnored(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMkdirAll(t, "code-from-spec/_tools")
	testWriteFile(t, "code-from-spec/_node.md", testRootNode())

	report := mcpvalidatespecs.MCPValidateSpecs()

	for _, e := range report.FormatErrors {
		if e.Node == "code-from-spec/_tools/" || e.Node == "code-from-spec/_tools" {
			t.Errorf("expected _tools dir to be ignored, but got format error: %+v", e)
		}
	}
}

func TestMCPValidateSpecs_SimpleCycleDetected(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMkdirAll(t, "code-from-spec/a")
	testMkdirAll(t, "code-from-spec/b")
	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\ndepends_on:\n  - SPEC/b\n---\n# SPEC/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md",
		"---\ndepends_on:\n  - SPEC/a\n---\n# SPEC/b\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Cycles) == 0 {
		t.Error("expected cycles to be non-empty")
	}

	foundCycleNode := false
	for _, name := range report.Cycles {
		if name == "SPEC/a" || name == "SPEC/b" {
			foundCycleNode = true
			break
		}
	}
	if !foundCycleNode {
		t.Errorf("expected cycles to contain SPEC/a or SPEC/b, got: %v", report.Cycles)
	}
}

func TestMCPValidateSpecs_RankingSkippedWhenFormatErrors(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMkdirAll(t, "code-from-spec/a")
	testMkdirAll(t, "code-from-spec/b")
	testMkdirAll(t, "out")
	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\ndepends_on:\n  - SPEC/missing\n---\n# SPEC/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md",
		"---\noutput: \"out/b.go\"\n---\n# SPEC/b\n")

	testWriteFile(t, "out/b.go",
		"// code-from-spec: SPEC/b@AAAAAAAAAAAAAAAAAAAAAAAAAAA\npackage main\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) == 0 {
		t.Error("expected format errors to be non-empty")
	}

	for _, s := range report.Staleness {
		if s.Node == "SPEC/b" && s.Rank != 0 {
			t.Errorf("expected rank=0 for SPEC/b when ranking skipped, got %d", s.Rank)
		}
	}
}

func TestMCPValidateSpecs_EmptySpecTree(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, e := range report.FormatErrors {
		if e.Rule == "scan" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected format error with rule=scan for empty spec tree, got: %+v", report.FormatErrors)
	}
	if len(report.Cycles) != 0 {
		t.Errorf("expected no cycles, got %v", report.Cycles)
	}
	if len(report.Staleness) != 0 {
		t.Errorf("expected no staleness, got %+v", report.Staleness)
	}
}

func TestMCPValidateSpecs_NodeWithNoOutputNotInStaleness(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMkdirAll(t, "code-from-spec/a")
	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md",
		"# SPEC/a\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	for _, s := range report.Staleness {
		if s.Node == "SPEC/a" {
			t.Errorf("expected no staleness entry for SPEC/a (no output declared), got: %+v", s)
		}
	}
}
