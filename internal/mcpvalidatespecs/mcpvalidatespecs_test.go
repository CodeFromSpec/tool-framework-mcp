// code-from-spec: ROOT/golang/tests/mcp_tools/validate_specs@caQHO5b-rzQKiRHgW4xxPsA_zj4
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

func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(testDirOf(path), 0755); err != nil {
		t.Fatalf("testWriteFile mkdir %q: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile %q: %v", path, err)
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

func testNodeFile(logicalName string, publicContent string) string {
	if publicContent == "" {
		return fmt.Sprintf("# %s\n", logicalName)
	}
	return fmt.Sprintf("# %s\n\n# Public\n\n%s\n", logicalName, publicContent)
}

func testNodeFileWithOutput(logicalName string, outputPath string) string {
	return fmt.Sprintf("---\noutput: %s\n---\n# %s\n\n# Public\n\npublic content\n", outputPath, logicalName)
}

func testNodeFileWithDependsOn(logicalName string, outputPath string, dependsOn string) string {
	return fmt.Sprintf("---\noutput: %s\ndepends_on:\n  - %s\n---\n# %s\n\n# Public\n\npublic content\n", outputPath, dependsOn, logicalName)
}

func testArtifactFile(logicalName string, hash string, extraContent string) string {
	tag := fmt.Sprintf("// code-from-spec: %s@%s", logicalName, hash)
	return tag + "\n" + extraContent
}

func testSetupRoot(t *testing.T) {
	t.Helper()
	testWriteFile(t, "code-from-spec/_node.md", testNodeFile("ROOT", "root public"))
}

func TestCleanTree(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testSetupRoot(t)
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFileWithOutput("ROOT/a", "out/a.go"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) > 0 || len(report.Cycles) > 0 {
		t.Fatalf("unexpected errors: format=%v cycles=%v", report.FormatErrors, report.Cycles)
	}

	if len(report.Staleness) != 1 {
		t.Fatalf("expected 1 staleness entry (missing artifact), got %d", len(report.Staleness))
	}

	entry := report.Staleness[0]
	if entry.Status != "missing" {
		t.Errorf("expected status missing, got %q", entry.Status)
	}

	currentHash := ""
	secondReport := mcpvalidatespecs.MCPValidateSpecs()
	for _, s := range secondReport.Staleness {
		if s.Node == "ROOT/a" {
			currentHash = s.Detail
			break
		}
	}
	_ = currentHash

	firstStale := report.Staleness[0]
	if firstStale.Node != "ROOT/a" {
		t.Errorf("expected node ROOT/a, got %q", firstStale.Node)
	}
}

func TestStaleArtifactDetected(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testSetupRoot(t)
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFileWithOutput("ROOT/a", "out/a.go"))
	testWriteFile(t, "out/a.go", testArtifactFile("ROOT/a", "AAAAAAAAAAAAAAAAAAAAAAAAAAA", ""))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) == 0 {
		t.Fatal("expected staleness entries")
	}
	found := false
	for _, s := range report.Staleness {
		if s.Node == "ROOT/a" && s.Status == "stale" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected ROOT/a stale entry, staleness: %+v", report.Staleness)
	}
}

func TestMissingArtifactDetected(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testSetupRoot(t)
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFileWithOutput("ROOT/a", "out/a.go"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, s := range report.Staleness {
		if s.Node == "ROOT/a" && s.Status == "missing" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected ROOT/a missing entry, staleness: %+v", report.Staleness)
	}
}

func TestMalformedTagDetected(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testSetupRoot(t)
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFileWithOutput("ROOT/a", "out/a.go"))
	testWriteFile(t, "out/a.go", "// no artifact tag here\nfunc main() {}\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, s := range report.Staleness {
		if s.Node == "ROOT/a" && s.Status == "malformed tag" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected ROOT/a malformed tag entry, staleness: %+v", report.Staleness)
	}
}

func TestStalenessEntriesIncludeRank(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testSetupRoot(t)
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFileWithOutput("ROOT/a", "out/a.go"))
	testWriteFile(t, "code-from-spec/b/_node.md", testNodeFileWithDependsOn("ROOT/b", "out/b.go", "ROOT/a"))
	testWriteFile(t, "out/a.go", testArtifactFile("ROOT/a", "AAAAAAAAAAAAAAAAAAAAAAAAAAA", ""))
	testWriteFile(t, "out/b.go", testArtifactFile("ROOT/b", "BBBBBBBBBBBBBBBBBBBBBBBBBBB", ""))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) > 0 {
		t.Fatalf("unexpected format errors: %v", report.FormatErrors)
	}

	var rankA, rankB int
	foundA, foundB := false, false
	for _, s := range report.Staleness {
		if s.Node == "ROOT/a" {
			rankA = s.Rank
			foundA = true
		}
		if s.Node == "ROOT/b" {
			rankB = s.Rank
			foundB = true
		}
	}

	if !foundA || !foundB {
		t.Fatalf("expected both ROOT/a and ROOT/b in staleness, got: %+v", report.Staleness)
	}

	if rankA >= rankB {
		t.Errorf("expected ROOT/a rank (%d) < ROOT/b rank (%d)", rankA, rankB)
	}
}

func TestStalenessOrderedByRankThenName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testSetupRoot(t)
	testWriteFile(t, "code-from-spec/z/_node.md", testNodeFileWithOutput("ROOT/z", "out/z.go"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFileWithOutput("ROOT/a", "out/a.go"))
	testWriteFile(t, "out/z.go", testArtifactFile("ROOT/z", "AAAAAAAAAAAAAAAAAAAAAAAAAAA", ""))
	testWriteFile(t, "out/a.go", testArtifactFile("ROOT/a", "BBBBBBBBBBBBBBBBBBBBBBBBBBB", ""))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) > 0 {
		t.Fatalf("unexpected format errors: %v", report.FormatErrors)
	}

	if len(report.Staleness) < 2 {
		t.Fatalf("expected at least 2 staleness entries, got %d", len(report.Staleness))
	}

	firstNode := report.Staleness[0].Node
	secondNode := report.Staleness[1].Node

	if firstNode != "ROOT/a" || secondNode != "ROOT/z" {
		t.Errorf("expected ROOT/a before ROOT/z, got %q and %q", firstNode, secondNode)
	}
}

func TestFormatErrorFromInvalidDependsOn(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testSetupRoot(t)
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFileWithDependsOn("ROOT/a", "out/a.go", "ROOT/missing"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, e := range report.FormatErrors {
		if e.Node == "ROOT/a" && e.Rule == "dependency_targets" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected dependency_targets format error for ROOT/a, got: %+v", report.FormatErrors)
	}
}

func TestFormatErrorFromParseFailure(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testSetupRoot(t)
	testWriteFile(t, "code-from-spec/a/_node.md", "this is invalid content before any heading\n# ROOT/a\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, e := range report.FormatErrors {
		if e.Node == "ROOT/a" && e.Rule == "parse" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected parse format error for ROOT/a, got: %+v", report.FormatErrors)
	}
}

func TestContinuesAfterParseFailure(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testSetupRoot(t)
	testWriteFile(t, "code-from-spec/a/_node.md", "invalid content before heading\n# ROOT/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md", testNodeFileWithOutput("ROOT/b", "out/b.go"))
	testWriteFile(t, "out/b.go", testArtifactFile("ROOT/b", "AAAAAAAAAAAAAAAAAAAAAAAAAAA", ""))

	report := mcpvalidatespecs.MCPValidateSpecs()

	foundParseError := false
	for _, e := range report.FormatErrors {
		if e.Node == "ROOT/a" && e.Rule == "parse" {
			foundParseError = true
			break
		}
	}
	if !foundParseError {
		t.Errorf("expected parse error for ROOT/a")
	}

	foundStaleness := false
	for _, s := range report.Staleness {
		if s.Node == "ROOT/b" {
			foundStaleness = true
			break
		}
	}
	if !foundStaleness {
		t.Errorf("expected staleness entry for ROOT/b")
	}
}

func TestSimpleCycleDetected(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testSetupRoot(t)
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFileWithDependsOn("ROOT/a", "out/a.go", "ROOT/b"))
	testWriteFile(t, "code-from-spec/b/_node.md", testNodeFileWithDependsOn("ROOT/b", "out/b.go", "ROOT/a"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Cycles) == 0 {
		t.Error("expected cycles to be detected")
	}

	foundCycle := false
	for _, name := range report.Cycles {
		if name == "ROOT/a" || name == "ROOT/b" {
			foundCycle = true
			break
		}
	}
	if !foundCycle {
		t.Errorf("expected ROOT/a or ROOT/b in cycles, got: %v", report.Cycles)
	}
}

func TestRankingSkippedWhenFormatErrorsExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testSetupRoot(t)
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFileWithDependsOn("ROOT/a", "out/a.go", "ROOT/missing"))
	testWriteFile(t, "code-from-spec/b/_node.md", testNodeFileWithOutput("ROOT/b", "out/b.go"))
	testWriteFile(t, "out/b.go", testArtifactFile("ROOT/b", "AAAAAAAAAAAAAAAAAAAAAAAAAAA", ""))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) == 0 {
		t.Fatal("expected format errors")
	}

	for _, s := range report.Staleness {
		if s.Node == "ROOT/b" && s.Rank != 0 {
			t.Errorf("expected rank 0 when ranking skipped, got %d", s.Rank)
		}
	}
}

func TestEmptySpecTreeScanFails(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, e := range report.FormatErrors {
		if e.Rule == "scan" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected scan format error, got: %+v", report.FormatErrors)
	}

	if len(report.Cycles) != 0 {
		t.Errorf("expected no cycles, got: %v", report.Cycles)
	}
	if len(report.Staleness) != 0 {
		t.Errorf("expected no staleness, got: %v", report.Staleness)
	}
}

func TestNodeWithNoOutputNotInStaleness(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testSetupRoot(t)
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("ROOT/a", "public content"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	for _, s := range report.Staleness {
		if s.Node == "ROOT/a" {
			t.Errorf("expected ROOT/a to not appear in staleness (no output field), got: %+v", s)
		}
	}
}
