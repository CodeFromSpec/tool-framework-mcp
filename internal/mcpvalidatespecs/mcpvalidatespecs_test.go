// code-from-spec: ROOT/golang/tests/mcp_tools/validate_specs@vDtKjEuZtgzzdu6_a5vRYSqJkYc
package mcpvalidatespecs_test

import (
	"fmt"
	"os"
	"testing"

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

func testWriteNodeFile(t *testing.T, relPath string, content string) {
	t.Helper()
	if err := os.MkdirAll(relPath[:len(relPath)-len("_node.md")], 0755); err != nil {
		t.Fatalf("testWriteNodeFile mkdir: %v", err)
	}
	if err := os.WriteFile(relPath, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNodeFile write: %v", err)
	}
}

func testNodeContent(logicalName string, extra string) string {
	return fmt.Sprintf("---\n%s---\n# %s\n", extra, logicalName)
}

func testRootNode() string {
	return "---\n---\n# ROOT\n\n# Public\n\nsome content\n"
}

func TestCleanTree(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "output: out/a.go\n"))

	if err := os.MkdirAll("out", 0755); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}

	initialReport := mcpvalidatespecs.MCPValidateSpecs()
	var matchingHash string
	for _, s := range initialReport.Staleness {
		if s.Node == "ROOT/a" {
			matchingHash = s.Detail
			_ = matchingHash
		}
	}

	report1 := mcpvalidatespecs.MCPValidateSpecs()
	var chainHash string
	for _, s := range report1.Staleness {
		if s.Node == "ROOT/a" {
			chainHash = s.Detail
		}
	}

	artifactContent := fmt.Sprintf("// code-from-spec: ROOT/a@%s\npackage a\n", chainHash)
	if err := os.WriteFile("out/a.go", []byte(artifactContent), 0644); err != nil {
		t.Fatalf("write out/a.go: %v", err)
	}

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) != 0 {
		t.Errorf("expected no format errors, got %d", len(report.FormatErrors))
	}
	if len(report.Cycles) != 0 {
		t.Errorf("expected no cycles, got %d", len(report.Cycles))
	}
	if len(report.Staleness) != 0 {
		t.Errorf("expected no staleness entries, got %d", len(report.Staleness))
	}
}

func TestStaleArtifact(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "output: out/a.go\n"))

	if err := os.MkdirAll("out", 0755); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	if err := os.WriteFile("out/a.go", []byte("// code-from-spec: ROOT/a@outdatedhashvalue000000000\npackage a\n"), 0644); err != nil {
		t.Fatalf("write out/a.go: %v", err)
	}

	report := mcpvalidatespecs.MCPValidateSpecs()

	var found *mcpvalidatespecs.StalenessEntry
	for _, s := range report.Staleness {
		if s.Node == "ROOT/a" {
			found = s
			break
		}
	}
	if found == nil {
		t.Fatal("expected StalenessEntry for ROOT/a, got none")
	}
	if found.Status != "stale" {
		t.Errorf("expected status 'stale', got %q", found.Status)
	}
}

func TestMissingArtifact(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "output: out/a.go\n"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	var found *mcpvalidatespecs.StalenessEntry
	for _, s := range report.Staleness {
		if s.Node == "ROOT/a" {
			found = s
			break
		}
	}
	if found == nil {
		t.Fatal("expected StalenessEntry for ROOT/a, got none")
	}
	if found.Status != "missing" {
		t.Errorf("expected status 'missing', got %q", found.Status)
	}
}

func TestMalformedTag(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "output: out/a.go\n"))

	if err := os.MkdirAll("out", 0755); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	if err := os.WriteFile("out/a.go", []byte("package a\n// no artifact tag here\n"), 0644); err != nil {
		t.Fatalf("write out/a.go: %v", err)
	}

	report := mcpvalidatespecs.MCPValidateSpecs()

	var found *mcpvalidatespecs.StalenessEntry
	for _, s := range report.Staleness {
		if s.Node == "ROOT/a" {
			found = s
			break
		}
	}
	if found == nil {
		t.Fatal("expected StalenessEntry for ROOT/a, got none")
	}
	if found.Status != "malformed tag" {
		t.Errorf("expected status 'malformed tag', got %q", found.Status)
	}
}

func TestStalenessEntriesIncludeRank(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "output: out/a.go\n"))
	testWriteNodeFile(t, "code-from-spec/b/_node.md", testNodeContent("ROOT/b", "output: out/b.go\ndepends_on:\n  - ROOT/a\n"))

	if err := os.MkdirAll("out", 0755); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	if err := os.WriteFile("out/a.go", []byte("// code-from-spec: ROOT/a@outdatedhashvalue000000000\npackage a\n"), 0644); err != nil {
		t.Fatalf("write out/a.go: %v", err)
	}
	if err := os.WriteFile("out/b.go", []byte("// code-from-spec: ROOT/b@outdatedhashvalue000000000\npackage b\n"), 0644); err != nil {
		t.Fatalf("write out/b.go: %v", err)
	}

	report := mcpvalidatespecs.MCPValidateSpecs()

	var entryA, entryB *mcpvalidatespecs.StalenessEntry
	for _, s := range report.Staleness {
		if s.Node == "ROOT/a" {
			entryA = s
		}
		if s.Node == "ROOT/b" {
			entryB = s
		}
	}
	if entryA == nil {
		t.Fatal("expected StalenessEntry for ROOT/a, got none")
	}
	if entryB == nil {
		t.Fatal("expected StalenessEntry for ROOT/b, got none")
	}
	if entryA.Rank >= entryB.Rank {
		t.Errorf("expected ROOT/a rank (%d) < ROOT/b rank (%d)", entryA.Rank, entryB.Rank)
	}
}

func TestStalenessOrderedByRankThenName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteNodeFile(t, "code-from-spec/z/_node.md", testNodeContent("ROOT/z", "output: out/z.go\n"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "output: out/a.go\n"))

	if err := os.MkdirAll("out", 0755); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	if err := os.WriteFile("out/z.go", []byte("// code-from-spec: ROOT/z@outdatedhashvalue000000000\npackage z\n"), 0644); err != nil {
		t.Fatalf("write out/z.go: %v", err)
	}
	if err := os.WriteFile("out/a.go", []byte("// code-from-spec: ROOT/a@outdatedhashvalue000000000\npackage a\n"), 0644); err != nil {
		t.Fatalf("write out/a.go: %v", err)
	}

	report := mcpvalidatespecs.MCPValidateSpecs()

	var idxA, idxZ int = -1, -1
	for i, s := range report.Staleness {
		if s.Node == "ROOT/a" {
			idxA = i
		}
		if s.Node == "ROOT/z" {
			idxZ = i
		}
	}
	if idxA == -1 {
		t.Fatal("expected StalenessEntry for ROOT/a, got none")
	}
	if idxZ == -1 {
		t.Fatal("expected StalenessEntry for ROOT/z, got none")
	}
	if idxA >= idxZ {
		t.Errorf("expected ROOT/a (idx %d) before ROOT/z (idx %d)", idxA, idxZ)
	}
}

func TestFormatErrorFromInvalidDependsOn(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "depends_on:\n  - ROOT/missing\n"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	var found *mcpvalidatespecs.FormatError
	for _, e := range report.FormatErrors {
		if e.Node == "ROOT/a" && e.Rule == "dependency_targets" {
			found = e
			break
		}
	}
	if found == nil {
		t.Errorf("expected FormatError for ROOT/a with rule='dependency_targets', got errors: %v", report.FormatErrors)
	}
}

func TestFormatErrorFromParseFailure(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "this text appears before any heading\n# ROOT/a\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	var found *mcpvalidatespecs.FormatError
	for _, e := range report.FormatErrors {
		if e.Node == "ROOT/a" && e.Rule == "parse" {
			found = e
			break
		}
	}
	if found == nil {
		t.Errorf("expected FormatError for ROOT/a with rule='parse', got errors: %v", report.FormatErrors)
	}
}

func TestContinuesAfterParseFailure(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "this text appears before any heading\n# ROOT/a\n")
	testWriteNodeFile(t, "code-from-spec/b/_node.md", testNodeContent("ROOT/b", "output: out/b.go\n"))

	if err := os.MkdirAll("out", 0755); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	if err := os.WriteFile("out/b.go", []byte("// code-from-spec: ROOT/b@outdatedhashvalue000000000\npackage b\n"), 0644); err != nil {
		t.Fatalf("write out/b.go: %v", err)
	}

	report := mcpvalidatespecs.MCPValidateSpecs()

	var parseErr *mcpvalidatespecs.FormatError
	for _, e := range report.FormatErrors {
		if e.Node == "ROOT/a" && e.Rule == "parse" {
			parseErr = e
			break
		}
	}
	if parseErr == nil {
		t.Errorf("expected FormatError for ROOT/a with rule='parse'")
	}

	var stalenessB *mcpvalidatespecs.StalenessEntry
	for _, s := range report.Staleness {
		if s.Node == "ROOT/b" {
			stalenessB = s
			break
		}
	}
	if stalenessB == nil {
		t.Errorf("expected StalenessEntry for ROOT/b")
	}
}

func TestSimpleCycleDetected(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "depends_on:\n  - ROOT/b\n"))
	testWriteNodeFile(t, "code-from-spec/b/_node.md", testNodeContent("ROOT/b", "depends_on:\n  - ROOT/a\n"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Cycles) == 0 {
		t.Fatal("expected non-empty cycles, got none")
	}

	foundCycle := false
	for _, name := range report.Cycles {
		if name == "ROOT/a" || name == "ROOT/b" {
			foundCycle = true
			break
		}
	}
	if !foundCycle {
		t.Errorf("expected cycles to contain ROOT/a or ROOT/b, got %v", report.Cycles)
	}
}

func TestRankingSkippedWhenFormatErrorsExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "depends_on:\n  - ROOT/missing\n"))
	testWriteNodeFile(t, "code-from-spec/b/_node.md", testNodeContent("ROOT/b", "output: out/b.go\n"))

	if err := os.MkdirAll("out", 0755); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	if err := os.WriteFile("out/b.go", []byte("// code-from-spec: ROOT/b@outdatedhashvalue000000000\npackage b\n"), 0644); err != nil {
		t.Fatalf("write out/b.go: %v", err)
	}

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) == 0 {
		t.Error("expected format errors, got none")
	}

	for _, s := range report.Staleness {
		if s.Node == "ROOT/b" {
			if s.Rank != 0 {
				t.Errorf("expected rank=0 for ROOT/b when ranking skipped, got %d", s.Rank)
			}
			break
		}
	}
}

func TestEmptySpecTreeScanFails(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	report := mcpvalidatespecs.MCPValidateSpecs()

	var found *mcpvalidatespecs.FormatError
	for _, e := range report.FormatErrors {
		if e.Rule == "scan" {
			found = e
			break
		}
	}
	if found == nil {
		t.Errorf("expected FormatError with rule='scan', got errors: %v", report.FormatErrors)
	}
	if len(report.Cycles) != 0 {
		t.Errorf("expected no cycles, got %v", report.Cycles)
	}
	if len(report.Staleness) != 0 {
		t.Errorf("expected no staleness, got %v", report.Staleness)
	}
}

func TestNodeWithNoOutputNotInStaleness(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", ""))

	report := mcpvalidatespecs.MCPValidateSpecs()

	for _, s := range report.Staleness {
		if s.Node == "ROOT/a" {
			t.Errorf("expected no StalenessEntry for ROOT/a (no output), but found one with status=%q", s.Status)
		}
	}
}
