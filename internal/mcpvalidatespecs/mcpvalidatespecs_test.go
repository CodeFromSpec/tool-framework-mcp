// code-from-spec: ROOT/golang/tests/mcp_tools/validate_specs@pS7XLfyG0BH-szTsEawNsLhMi8I
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

func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

func testWriteNodeFile(t *testing.T, path string, frontmatter string, logicalName string) {
	t.Helper()
	var content string
	if frontmatter != "" {
		content = frontmatter + "\n# " + logicalName + "\n"
	} else {
		content = "# " + logicalName + "\n"
	}
	testWriteFile(t, path, content)
}

func testComputeHash(t *testing.T, logicalName string) string {
	t.Helper()
	chain, err := chainresolver.ChainResolve(logicalName)
	if err != nil {
		t.Fatalf("testComputeHash ChainResolve(%s): %v", logicalName, err)
	}
	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("testComputeHash ChainHashCompute(%s): %v", logicalName, err)
	}
	return hash
}

func TestCleanTree_NoErrors(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t,
		"code-from-spec/_node.md",
		"",
		"ROOT",
	)
	testWriteFile(t, "code-from-spec/_node.md",
		"# ROOT\n\n# Public\n\nRoot public section.\n",
	)

	testWriteNodeFile(t,
		"code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: code\n    path: out/a.go\n---",
		"ROOT/a",
	)

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

func TestStaleArtifactDetected(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "", "ROOT")
	testWriteNodeFile(t,
		"code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: code\n    path: out/a.go\n---",
		"ROOT/a",
	)

	testWriteFile(t, "out/a.go", "// code-from-spec: ROOT/a@aaaaaaaaaaaaaaaaaaaaaaaaaaa\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 1 {
		t.Fatalf("expected 1 staleness entry, got %d: %+v", len(report.Staleness), report.Staleness)
	}
	entry := report.Staleness[0]
	if entry.Node != "ROOT/a" {
		t.Errorf("expected node ROOT/a, got %s", entry.Node)
	}
	if entry.OutputID != "code" {
		t.Errorf("expected output_id code, got %s", entry.OutputID)
	}
	if entry.Status != "stale" {
		t.Errorf("expected status stale, got %s", entry.Status)
	}
	_ = entry.Rank
}

func TestMissingArtifactDetected(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "", "ROOT")
	testWriteNodeFile(t,
		"code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: code\n    path: out/a.go\n---",
		"ROOT/a",
	)

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 1 {
		t.Fatalf("expected 1 staleness entry, got %d: %+v", len(report.Staleness), report.Staleness)
	}
	entry := report.Staleness[0]
	if entry.Node != "ROOT/a" {
		t.Errorf("expected node ROOT/a, got %s", entry.Node)
	}
	if entry.OutputID != "code" {
		t.Errorf("expected output_id code, got %s", entry.OutputID)
	}
	if entry.Status != "missing" {
		t.Errorf("expected status missing, got %s", entry.Status)
	}
}

func TestMalformedTagDetected(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "", "ROOT")
	testWriteNodeFile(t,
		"code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: code\n    path: out/a.go\n---",
		"ROOT/a",
	)

	testWriteFile(t, "out/a.go", "// no artifact tag here\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 1 {
		t.Fatalf("expected 1 staleness entry, got %d: %+v", len(report.Staleness), report.Staleness)
	}
	entry := report.Staleness[0]
	if entry.Node != "ROOT/a" {
		t.Errorf("expected node ROOT/a, got %s", entry.Node)
	}
	if entry.OutputID != "code" {
		t.Errorf("expected output_id code, got %s", entry.OutputID)
	}
	if entry.Status != "malformed tag" {
		t.Errorf("expected status 'malformed tag', got %s", entry.Status)
	}
}

func TestMultipleOutputs_EachCheckedIndependently(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "", "ROOT")
	testWriteNodeFile(t,
		"code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: x\n    path: out/x.go\n  - id: y\n    path: out/y.go\n---",
		"ROOT/a",
	)

	hash := testComputeHash(t, "ROOT/a")
	testWriteFile(t, "out/x.go", fmt.Sprintf("// code-from-spec: ROOT/a@%s\n", hash))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 1 {
		t.Fatalf("expected 1 staleness entry, got %d: %+v", len(report.Staleness), report.Staleness)
	}
	entry := report.Staleness[0]
	if entry.OutputID != "y" {
		t.Errorf("expected output_id y, got %s", entry.OutputID)
	}
	if entry.Status != "missing" {
		t.Errorf("expected status missing, got %s", entry.Status)
	}
}

func TestStalenessEntriesIncludeRank(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "", "ROOT")
	testWriteNodeFile(t,
		"code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: code\n    path: out/a.go\n---",
		"ROOT/a",
	)
	testWriteNodeFile(t,
		"code-from-spec/b/_node.md",
		"---\ndepends_on:\n  - ROOT/a\noutputs:\n  - id: code\n    path: out/b.go\n---",
		"ROOT/b",
	)

	testWriteFile(t, "out/a.go", "// code-from-spec: ROOT/a@aaaaaaaaaaaaaaaaaaaaaaaaaaa\n")
	testWriteFile(t, "out/b.go", "// code-from-spec: ROOT/b@aaaaaaaaaaaaaaaaaaaaaaaaaaa\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 2 {
		t.Fatalf("expected 2 staleness entries, got %d: %+v", len(report.Staleness), report.Staleness)
	}

	var rankA, rankB int
	for _, e := range report.Staleness {
		switch e.Node {
		case "ROOT/a":
			rankA = e.Rank
		case "ROOT/b":
			rankB = e.Rank
		}
	}
	if rankA >= rankB {
		t.Errorf("expected ROOT/a rank (%d) < ROOT/b rank (%d)", rankA, rankB)
	}
}

func TestStalenessEntriesOrderedByRankThenName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "", "ROOT")
	testWriteNodeFile(t,
		"code-from-spec/z/_node.md",
		"---\noutputs:\n  - id: code\n    path: out/z.go\n---",
		"ROOT/z",
	)
	testWriteNodeFile(t,
		"code-from-spec/a/_node.md",
		"---\noutputs:\n  - id: code\n    path: out/a.go\n---",
		"ROOT/a",
	)

	testWriteFile(t, "out/z.go", "// code-from-spec: ROOT/z@aaaaaaaaaaaaaaaaaaaaaaaaaaa\n")
	testWriteFile(t, "out/a.go", "// code-from-spec: ROOT/a@aaaaaaaaaaaaaaaaaaaaaaaaaaa\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 2 {
		t.Fatalf("expected 2 staleness entries, got %d: %+v", len(report.Staleness), report.Staleness)
	}
	if report.Staleness[0].Node != "ROOT/a" {
		t.Errorf("expected first entry ROOT/a, got %s", report.Staleness[0].Node)
	}
	if report.Staleness[1].Node != "ROOT/z" {
		t.Errorf("expected second entry ROOT/z, got %s", report.Staleness[1].Node)
	}
}

func TestFormatError_InvalidDependsOn(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "", "ROOT")
	testWriteNodeFile(t,
		"code-from-spec/a/_node.md",
		"---\ndepends_on:\n  - ROOT/missing\n---",
		"ROOT/a",
	)

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, fe := range report.FormatErrors {
		if fe.Node == "ROOT/a" && fe.Rule == "dependency_targets" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected FormatError for ROOT/a with rule=dependency_targets, got: %+v", report.FormatErrors)
	}
}

func TestFormatError_ParseFailure(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "", "ROOT")

	if err := os.MkdirAll("code-from-spec/a", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	testWriteFile(t, "code-from-spec/a/_node.md", "this is text before any heading\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, fe := range report.FormatErrors {
		if fe.Rule == "parse" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected FormatError with rule=parse, got: %+v", report.FormatErrors)
	}
}

func TestContinuesAfterParseFailure(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "", "ROOT")

	if err := os.MkdirAll("code-from-spec/a", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	testWriteFile(t, "code-from-spec/a/_node.md", "this is text before any heading\n")

	testWriteNodeFile(t,
		"code-from-spec/b/_node.md",
		"---\noutputs:\n  - id: code\n    path: out/b.go\n---",
		"ROOT/b",
	)
	testWriteFile(t, "out/b.go", "// code-from-spec: ROOT/b@aaaaaaaaaaaaaaaaaaaaaaaaaaa\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	foundParseError := false
	for _, fe := range report.FormatErrors {
		if fe.Rule == "parse" {
			foundParseError = true
			break
		}
	}
	if !foundParseError {
		t.Errorf("expected FormatError with rule=parse, got: %+v", report.FormatErrors)
	}

	foundBStale := false
	for _, se := range report.Staleness {
		if se.Node == "ROOT/b" && se.Status == "stale" {
			foundBStale = true
			break
		}
	}
	if !foundBStale {
		t.Errorf("expected StalenessEntry for ROOT/b with status=stale, got: %+v", report.Staleness)
	}
}

func TestSimpleCycleDetected(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "", "ROOT")
	testWriteNodeFile(t,
		"code-from-spec/a/_node.md",
		"---\ndepends_on:\n  - ROOT/b\n---",
		"ROOT/a",
	)
	testWriteNodeFile(t,
		"code-from-spec/b/_node.md",
		"---\ndepends_on:\n  - ROOT/a\n---",
		"ROOT/b",
	)

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Cycles) == 0 {
		t.Error("expected at least one cycle entry, got none")
	}

	foundCycleMember := false
	for _, name := range report.Cycles {
		if name == "ROOT/a" || name == "ROOT/b" {
			foundCycleMember = true
			break
		}
	}
	if !foundCycleMember {
		t.Errorf("expected cycles to contain ROOT/a or ROOT/b, got: %v", report.Cycles)
	}
}

func TestRankingSkippedWhenFormatErrorsExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "", "ROOT")
	testWriteNodeFile(t,
		"code-from-spec/a/_node.md",
		"---\ndepends_on:\n  - ROOT/missing\n---",
		"ROOT/a",
	)
	testWriteNodeFile(t,
		"code-from-spec/b/_node.md",
		"---\noutputs:\n  - id: code\n    path: out/b.go\n---",
		"ROOT/b",
	)
	testWriteFile(t, "out/b.go", "// code-from-spec: ROOT/b@aaaaaaaaaaaaaaaaaaaaaaaaaaa\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) == 0 {
		t.Error("expected format errors, got none")
	}

	foundBStale := false
	for _, se := range report.Staleness {
		if se.Node == "ROOT/b" {
			foundBStale = true
			if se.Rank != 0 {
				t.Errorf("expected rank 0 when ranking skipped, got %d", se.Rank)
			}
			break
		}
	}
	if !foundBStale {
		t.Errorf("expected StalenessEntry for ROOT/b, got: %+v", report.Staleness)
	}
}

func TestEmptySpecTree_ScanFails(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	report := mcpvalidatespecs.MCPValidateSpecs()

	foundScanError := false
	for _, fe := range report.FormatErrors {
		if fe.Rule == "scan" {
			foundScanError = true
			break
		}
	}
	if !foundScanError {
		t.Errorf("expected FormatError with rule=scan, got: %+v", report.FormatErrors)
	}
	if len(report.Cycles) != 0 {
		t.Errorf("expected no cycles, got %v", report.Cycles)
	}
	if len(report.Staleness) != 0 {
		t.Errorf("expected no staleness entries, got %+v", report.Staleness)
	}
}

func TestNodeWithNoOutputs_NotInStaleness(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "", "ROOT")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "", "ROOT/a")

	report := mcpvalidatespecs.MCPValidateSpecs()

	for _, se := range report.Staleness {
		if se.Node == "ROOT/a" {
			t.Errorf("expected no staleness entry for ROOT/a (no outputs declared), got: %+v", se)
		}
	}
}
