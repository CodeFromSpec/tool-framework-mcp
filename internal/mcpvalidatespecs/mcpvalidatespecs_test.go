// code-from-spec: ROOT/golang/tests/mcp_tools/validate_specs@jkbgLenEV9ATOosrtSasU1v5WFw
package mcpvalidatespecs_test

import (
	"fmt"
	"os"
	"path/filepath"
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

func testWriteNode(t *testing.T, logicalName string, frontmatter string) {
	t.Helper()
	parts := logicalName[len("ROOT/"):]
	dir := filepath.Join("code-from-spec", filepath.FromSlash(parts))
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testWriteNode MkdirAll: %v", err)
	}
	body := fmt.Sprintf("---\n%s---\n# %s\n", frontmatter, logicalName)
	if err := os.WriteFile(filepath.Join(dir, "_node.md"), []byte(body), 0644); err != nil {
		t.Fatalf("testWriteNode WriteFile: %v", err)
	}
}

func testWriteRootNode(t *testing.T) {
	t.Helper()
	dir := "code-from-spec"
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testWriteRootNode MkdirAll: %v", err)
	}
	body := "---\n---\n# ROOT\n\n# Public\n"
	if err := os.WriteFile(filepath.Join(dir, "_node.md"), []byte(body), 0644); err != nil {
		t.Fatalf("testWriteRootNode WriteFile: %v", err)
	}
}

func testComputeHash(t *testing.T, logicalName string) string {
	t.Helper()
	chain, err := chainresolver.ChainResolve(logicalName)
	if err != nil {
		t.Fatalf("testComputeHash ChainResolve: %v", err)
	}
	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("testComputeHash ChainHashCompute: %v", err)
	}
	return hash
}

func testWriteArtifact(t *testing.T, path string, logicalName string, hash string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testWriteArtifact MkdirAll: %v", err)
	}
	content := fmt.Sprintf("// code-from-spec: %s@%s\n", logicalName, hash)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteArtifact WriteFile: %v", err)
	}
}

func TestMCPValidateSpecs_CleanTree(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteRootNode(t)
	testWriteNode(t, "ROOT/a", "output: out/a.go\n")

	hash := testComputeHash(t, "ROOT/a")
	testWriteArtifact(t, "out/a.go", "ROOT/a", hash)

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) != 0 {
		t.Errorf("expected no format errors, got %d", len(report.FormatErrors))
	}
	if len(report.Cycles) != 0 {
		t.Errorf("expected no cycles, got %v", report.Cycles)
	}
	if len(report.Staleness) != 0 {
		t.Errorf("expected no staleness, got %d entries", len(report.Staleness))
	}
}

func TestMCPValidateSpecs_StaleArtifact(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteRootNode(t)
	testWriteNode(t, "ROOT/a", "output: out/a.go\n")

	testWriteArtifact(t, "out/a.go", "ROOT/a", "AAAAAAAAAAAAAAAAAAAAAAAAAA_")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 1 {
		t.Fatalf("expected 1 staleness entry, got %d", len(report.Staleness))
	}
	entry := report.Staleness[0]
	if entry.Node != "ROOT/a" {
		t.Errorf("expected node ROOT/a, got %s", entry.Node)
	}
	if entry.Status != "stale" {
		t.Errorf("expected status stale, got %s", entry.Status)
	}
}

func TestMCPValidateSpecs_MissingArtifact(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteRootNode(t)
	testWriteNode(t, "ROOT/a", "output: out/a.go\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 1 {
		t.Fatalf("expected 1 staleness entry, got %d", len(report.Staleness))
	}
	entry := report.Staleness[0]
	if entry.Node != "ROOT/a" {
		t.Errorf("expected node ROOT/a, got %s", entry.Node)
	}
	if entry.Status != "missing" {
		t.Errorf("expected status missing, got %s", entry.Status)
	}
}

func TestMCPValidateSpecs_MalformedTag(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteRootNode(t)
	testWriteNode(t, "ROOT/a", "output: out/a.go\n")

	if err := os.MkdirAll("out", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile("out/a.go", []byte("package main\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 1 {
		t.Fatalf("expected 1 staleness entry, got %d", len(report.Staleness))
	}
	entry := report.Staleness[0]
	if entry.Node != "ROOT/a" {
		t.Errorf("expected node ROOT/a, got %s", entry.Node)
	}
	if entry.Status != "malformed tag" {
		t.Errorf("expected status 'malformed tag', got %s", entry.Status)
	}
}

func TestMCPValidateSpecs_StalenessEntriesIncludeRank(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteRootNode(t)
	testWriteNode(t, "ROOT/a", "output: out/a.go\n")
	testWriteNode(t, "ROOT/b", "depends_on:\n  - ROOT/a\noutput: out/b.go\n")

	testWriteArtifact(t, "out/a.go", "ROOT/a", "AAAAAAAAAAAAAAAAAAAAAAAAAA_")
	testWriteArtifact(t, "out/b.go", "ROOT/b", "AAAAAAAAAAAAAAAAAAAAAAAAAA_")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 2 {
		t.Fatalf("expected 2 staleness entries, got %d", len(report.Staleness))
	}

	var rankA, rankB int
	for _, entry := range report.Staleness {
		if entry.Node == "ROOT/a" {
			rankA = entry.Rank
		} else if entry.Node == "ROOT/b" {
			rankB = entry.Rank
		}
	}

	if rankA >= rankB {
		t.Errorf("expected ROOT/a rank (%d) < ROOT/b rank (%d)", rankA, rankB)
	}
}

func TestMCPValidateSpecs_StalenessOrderedByRankThenName(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteRootNode(t)
	testWriteNode(t, "ROOT/a", "output: out/a.go\n")
	testWriteNode(t, "ROOT/z", "output: out/z.go\n")

	testWriteArtifact(t, "out/a.go", "ROOT/a", "AAAAAAAAAAAAAAAAAAAAAAAAAA_")
	testWriteArtifact(t, "out/z.go", "ROOT/z", "AAAAAAAAAAAAAAAAAAAAAAAAAA_")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 2 {
		t.Fatalf("expected 2 staleness entries, got %d", len(report.Staleness))
	}

	if report.Staleness[0].Node != "ROOT/a" {
		t.Errorf("expected ROOT/a first, got %s", report.Staleness[0].Node)
	}
	if report.Staleness[1].Node != "ROOT/z" {
		t.Errorf("expected ROOT/z second, got %s", report.Staleness[1].Node)
	}
}

func TestMCPValidateSpecs_FormatErrorFromInvalidDependsOn(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteRootNode(t)
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ROOT/missing\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, e := range report.FormatErrors {
		if e.Node == "ROOT/a" && e.Rule == "dependency_targets" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected format error for ROOT/a with rule dependency_targets, got %v", report.FormatErrors)
	}
}

func TestMCPValidateSpecs_FormatErrorFromParseFailure(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteRootNode(t)

	if err := os.MkdirAll(filepath.Join("code-from-spec", "a"), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join("code-from-spec", "a", "_node.md"), []byte("invalid content before heading\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, e := range report.FormatErrors {
		if e.Node == "ROOT/a" && e.Rule == "parse" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected format error for ROOT/a with rule parse, got %v", report.FormatErrors)
	}
}

func TestMCPValidateSpecs_ContinuesAfterParseFailure(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteRootNode(t)

	if err := os.MkdirAll(filepath.Join("code-from-spec", "a"), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join("code-from-spec", "a", "_node.md"), []byte("invalid content before heading\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	testWriteNode(t, "ROOT/b", "output: out/b.go\n")
	testWriteArtifact(t, "out/b.go", "ROOT/b", "AAAAAAAAAAAAAAAAAAAAAAAAAA_")

	report := mcpvalidatespecs.MCPValidateSpecs()

	foundParseErr := false
	for _, e := range report.FormatErrors {
		if e.Node == "ROOT/a" && e.Rule == "parse" {
			foundParseErr = true
			break
		}
	}
	if !foundParseErr {
		t.Errorf("expected parse error for ROOT/a, got %v", report.FormatErrors)
	}

	foundStaleness := false
	for _, s := range report.Staleness {
		if s.Node == "ROOT/b" {
			foundStaleness = true
			break
		}
	}
	if !foundStaleness {
		t.Errorf("expected staleness entry for ROOT/b, got %v", report.Staleness)
	}
}

func TestMCPValidateSpecs_SimpleCycleDetected(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteRootNode(t)
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ROOT/b\n")
	testWriteNode(t, "ROOT/b", "depends_on:\n  - ROOT/a\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Cycles) == 0 {
		t.Error("expected cycles to be non-empty")
	}

	foundA := false
	foundB := false
	for _, name := range report.Cycles {
		if name == "ROOT/a" {
			foundA = true
		}
		if name == "ROOT/b" {
			foundB = true
		}
	}
	if !foundA && !foundB {
		t.Errorf("expected cycles to contain ROOT/a or ROOT/b, got %v", report.Cycles)
	}
}

func TestMCPValidateSpecs_RankingSkippedWhenFormatErrors(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteRootNode(t)
	testWriteNode(t, "ROOT/a", "depends_on:\n  - ROOT/missing\n")
	testWriteNode(t, "ROOT/b", "output: out/b.go\n")

	testWriteArtifact(t, "out/b.go", "ROOT/b", "AAAAAAAAAAAAAAAAAAAAAAAAAA_")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) == 0 {
		t.Error("expected format errors, got none")
	}

	for _, s := range report.Staleness {
		if s.Node == "ROOT/b" && s.Rank != 0 {
			t.Errorf("expected rank 0 for ROOT/b when ranking is skipped, got %d", s.Rank)
		}
	}
}

func TestMCPValidateSpecs_EmptySpecTreeScanFails(t *testing.T) {
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
		t.Errorf("expected format error with rule 'scan', got %v", report.FormatErrors)
	}

	if len(report.Cycles) != 0 {
		t.Errorf("expected no cycles, got %v", report.Cycles)
	}
	if len(report.Staleness) != 0 {
		t.Errorf("expected no staleness, got %d entries", len(report.Staleness))
	}
}

func TestMCPValidateSpecs_NodeWithNoOutputNotInStaleness(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteRootNode(t)
	testWriteNode(t, "ROOT/a", "")

	report := mcpvalidatespecs.MCPValidateSpecs()

	for _, s := range report.Staleness {
		if s.Node == "ROOT/a" {
			t.Errorf("expected ROOT/a not in staleness, but found entry: %v", s)
		}
	}
}
