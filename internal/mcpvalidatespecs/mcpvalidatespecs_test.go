// code-from-spec: ROOT/golang/tests/mcp_tools/validate_specs@PR6mQfW8wxP6SWi0rECE12avAtc
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

// testChdir changes the working directory to dir for the duration of the test.
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

// testWriteFile creates parent directories and writes content to path
// (relative to the current working directory).
func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("testWriteFile MkdirAll(%q): %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile WriteFile(%q): %v", path, err)
	}
}

// testNodeContent returns a minimal valid _node.md for a node whose logical
// name is logicalName. frontmatterYAML is the raw YAML content between the
// --- delimiters (may be empty).
func testNodeContent(logicalName string, frontmatterYAML string) string {
	return fmt.Sprintf("---\n%s---\n\n# %s\n", frontmatterYAML, logicalName)
}

// testRootNodeContent returns a minimal valid _node.md for the ROOT node.
func testRootNodeContent() string {
	return "---\n---\n\n# ROOT\n\n## Public\n"
}

// testComputeChainHash returns the current chain hash for logicalName.
// It uses chainresolver and chainhash from the internal packages.
func testComputeChainHash(t *testing.T, logicalName string) string {
	t.Helper()
	chain, err := chainresolver.ChainResolve(logicalName)
	if err != nil {
		t.Fatalf("testComputeChainHash: ChainResolve(%q): %v", logicalName, err)
	}
	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("testComputeChainHash: ChainHashCompute(%q): %v", logicalName, err)
	}
	return hash
}

// --------------------------------------------------------------------------
// Happy Path
// --------------------------------------------------------------------------

// TC-HP-01: Clean tree — no errors.
func TestMCPValidateSpecs_CleanTree(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a",
		"outputs:\n  - id: code\n    path: out/a.go\n"))

	// Compute the current chain hash for ROOT/a and write a matching artifact tag.
	hash := testComputeChainHash(t, "ROOT/a")
	testWriteFile(t, "out/a.go", fmt.Sprintf("// code-from-spec: ROOT/a@%s\n", hash))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) != 0 {
		t.Errorf("expected no format errors, got %d: %v", len(report.FormatErrors), report.FormatErrors)
	}
	if len(report.Cycles) != 0 {
		t.Errorf("expected no cycles, got %v", report.Cycles)
	}
	if len(report.Staleness) != 0 {
		t.Errorf("expected no staleness entries, got %d: %v", len(report.Staleness), report.Staleness)
	}
}

// TC-HP-02: Stale artifact detected.
func TestMCPValidateSpecs_StaleArtifact(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a",
		"outputs:\n  - id: code\n    path: out/a.go\n"))

	// Write an artifact tag with a well-formed but non-matching hash.
	testWriteFile(t, "out/a.go", "// code-from-spec: ROOT/a@AAAAAAAAAAAAAAAAAAAAAAAAAAA\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, se := range report.Staleness {
		if se.Node == "ROOT/a" && se.OutputID == "code" {
			if se.Status != "stale" {
				t.Errorf("expected status=stale, got %q", se.Status)
			}
			if se.Rank < 0 {
				t.Errorf("expected non-negative rank, got %d", se.Rank)
			}
			found = true
		}
	}
	if !found {
		t.Errorf("expected a StalenessEntry for ROOT/a output_id=code; got: %v", report.Staleness)
	}
}

// TC-HP-03: Missing artifact detected.
func TestMCPValidateSpecs_MissingArtifact(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a",
		"outputs:\n  - id: code\n    path: out/a.go\n"))

	// Do not create out/a.go.

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, se := range report.Staleness {
		if se.Node == "ROOT/a" && se.OutputID == "code" {
			if se.Status != "missing" {
				t.Errorf("expected status=missing, got %q", se.Status)
			}
			found = true
		}
	}
	if !found {
		t.Errorf("expected a StalenessEntry for ROOT/a output_id=code; got: %v", report.Staleness)
	}
}

// TC-HP-04: Malformed tag detected.
func TestMCPValidateSpecs_MalformedTag(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a",
		"outputs:\n  - id: code\n    path: out/a.go\n"))

	// File exists but contains no artifact tag.
	testWriteFile(t, "out/a.go", "package main\n\n// no artifact tag here\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, se := range report.Staleness {
		if se.Node == "ROOT/a" && se.OutputID == "code" {
			if se.Status != "malformed tag" {
				t.Errorf("expected status=\"malformed tag\", got %q", se.Status)
			}
			found = true
		}
	}
	if !found {
		t.Errorf("expected a StalenessEntry for ROOT/a output_id=code; got: %v", report.Staleness)
	}
}

// TC-HP-05: Multiple outputs — each checked independently.
func TestMCPValidateSpecs_MultipleOutputs(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a",
		"outputs:\n  - id: x\n    path: out/x.go\n  - id: y\n    path: out/y.go\n"))

	// Compute hash for ROOT/a and write a matching tag for output x.
	hash := testComputeChainHash(t, "ROOT/a")
	testWriteFile(t, "out/x.go", fmt.Sprintf("// code-from-spec: ROOT/a@%s\n", hash))
	// Do not create out/y.go.

	report := mcpvalidatespecs.MCPValidateSpecs()

	foundY := false
	for _, se := range report.Staleness {
		if se.Node == "ROOT/a" && se.OutputID == "x" {
			t.Errorf("did not expect a StalenessEntry for output_id=x (hash matches)")
		}
		if se.Node == "ROOT/a" && se.OutputID == "y" {
			if se.Status != "missing" {
				t.Errorf("expected status=missing for output_id=y, got %q", se.Status)
			}
			foundY = true
		}
	}
	if !foundY {
		t.Errorf("expected a StalenessEntry for ROOT/a output_id=y; got: %v", report.Staleness)
	}
}

// TC-HP-06: Staleness entries include rank; ROOT/a rank < ROOT/b rank.
func TestMCPValidateSpecs_StalenessIncludesRank(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a",
		"outputs:\n  - id: code-a\n    path: out/a.go\n"))
	testWriteFile(t, "code-from-spec/b/_node.md", testNodeContent("ROOT/b",
		"depends_on:\n  - ROOT/a\noutputs:\n  - id: code-b\n    path: out/b.go\n"))

	// Both artifacts exist but with outdated hashes.
	testWriteFile(t, "out/a.go", "// code-from-spec: ROOT/a@AAAAAAAAAAAAAAAAAAAAAAAAAAA\n")
	testWriteFile(t, "out/b.go", "// code-from-spec: ROOT/b@AAAAAAAAAAAAAAAAAAAAAAAAAAA\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	var rankA, rankB int
	foundA, foundB := false, false
	for _, se := range report.Staleness {
		switch se.Node {
		case "ROOT/a":
			foundA = true
			rankA = se.Rank
			if se.Rank < 0 {
				t.Errorf("ROOT/a: expected non-negative rank, got %d", se.Rank)
			}
		case "ROOT/b":
			foundB = true
			rankB = se.Rank
			if se.Rank < 0 {
				t.Errorf("ROOT/b: expected non-negative rank, got %d", se.Rank)
			}
		}
	}

	if !foundA {
		t.Errorf("expected StalenessEntry for ROOT/a")
	}
	if !foundB {
		t.Errorf("expected StalenessEntry for ROOT/b")
	}
	if foundA && foundB && rankA >= rankB {
		t.Errorf("expected rank(ROOT/a)=%d < rank(ROOT/b)=%d", rankA, rankB)
	}
}

// TC-HP-07: Staleness ordered by rank then alphabetical node name.
func TestMCPValidateSpecs_StalenessOrder(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())
	testWriteFile(t, "code-from-spec/z/_node.md", testNodeContent("ROOT/z",
		"outputs:\n  - id: code-z\n    path: out/z.go\n"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a",
		"outputs:\n  - id: code-a\n    path: out/a.go\n"))

	// Both stale with outdated hashes.
	testWriteFile(t, "out/z.go", "// code-from-spec: ROOT/z@AAAAAAAAAAAAAAAAAAAAAAAAAAA\n")
	testWriteFile(t, "out/a.go", "// code-from-spec: ROOT/a@AAAAAAAAAAAAAAAAAAAAAAAAAAA\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 2 {
		t.Fatalf("expected 2 staleness entries, got %d: %v", len(report.Staleness), report.Staleness)
	}

	if report.Staleness[0].Node != "ROOT/a" {
		t.Errorf("expected first entry to be ROOT/a (same rank, alphabetical), got %q",
			report.Staleness[0].Node)
	}
	if report.Staleness[1].Node != "ROOT/z" {
		t.Errorf("expected second entry to be ROOT/z, got %q", report.Staleness[1].Node)
	}
}

// --------------------------------------------------------------------------
// Format Errors
// --------------------------------------------------------------------------

// TC-FE-01: Format error from invalid depends_on (target does not exist).
func TestMCPValidateSpecs_FormatError_InvalidDependsOn(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a",
		"depends_on:\n  - ROOT/missing\n"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, fe := range report.FormatErrors {
		if fe.Node == "ROOT/a" && fe.Rule == "dependency_targets" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected FormatError for ROOT/a rule=dependency_targets; got: %v", report.FormatErrors)
	}
}

// TC-FE-02: Format error from parse failure; other nodes still validated.
func TestMCPValidateSpecs_FormatError_ParseFailure(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())

	// ROOT/a has non-blank content before the first heading — unparseable.
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\n---\n\nInvalid text before any heading.\n# ROOT/a\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, fe := range report.FormatErrors {
		if fe.Node == "ROOT/a" && fe.Rule == "parse" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected FormatError for ROOT/a rule=parse; got: %v", report.FormatErrors)
	}
}

// TC-FE-03: Validation continues after a parse failure.
func TestMCPValidateSpecs_FormatError_ContinuesAfterParseFailure(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())

	// ROOT/a is unparseable.
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\n---\n\nInvalid text before any heading.\n# ROOT/a\n")

	// ROOT/b is valid but has a stale artifact.
	testWriteFile(t, "code-from-spec/b/_node.md", testNodeContent("ROOT/b",
		"outputs:\n  - id: code-b\n    path: out/b.go\n"))
	testWriteFile(t, "out/b.go", "// code-from-spec: ROOT/b@AAAAAAAAAAAAAAAAAAAAAAAAAAA\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	foundParseErr := false
	for _, fe := range report.FormatErrors {
		if fe.Node == "ROOT/a" && fe.Rule == "parse" {
			foundParseErr = true
		}
	}
	if !foundParseErr {
		t.Errorf("expected FormatError for ROOT/a rule=parse; got: %v", report.FormatErrors)
	}

	foundStaleness := false
	for _, se := range report.Staleness {
		if se.Node == "ROOT/b" {
			foundStaleness = true
		}
	}
	if !foundStaleness {
		t.Errorf("expected StalenessEntry for ROOT/b; got: %v", report.Staleness)
	}
}

// --------------------------------------------------------------------------
// Cycle Detection
// --------------------------------------------------------------------------

// TC-CY-01: Simple cycle detected.
func TestMCPValidateSpecs_CycleDetected(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a",
		"depends_on:\n  - ROOT/b\n"))
	testWriteFile(t, "code-from-spec/b/_node.md", testNodeContent("ROOT/b",
		"depends_on:\n  - ROOT/a\n"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Cycles) == 0 {
		t.Fatalf("expected non-empty cycles list")
	}

	containsA := false
	containsB := false
	for _, name := range report.Cycles {
		if name == "ROOT/a" {
			containsA = true
		}
		if name == "ROOT/b" {
			containsB = true
		}
	}
	if !containsA && !containsB {
		t.Errorf("expected cycles to contain ROOT/a or ROOT/b; got: %v", report.Cycles)
	}
}

// TC-CY-02: Ranking is skipped when format errors exist; staleness rank defaults to 0.
func TestMCPValidateSpecs_RankingSkippedOnFormatErrors(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())
	// ROOT/a has an invalid depends_on — format error.
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a",
		"depends_on:\n  - ROOT/missing\n"))
	// ROOT/b is valid with a stale artifact.
	testWriteFile(t, "code-from-spec/b/_node.md", testNodeContent("ROOT/b",
		"outputs:\n  - id: code-b\n    path: out/b.go\n"))
	testWriteFile(t, "out/b.go", "// code-from-spec: ROOT/b@AAAAAAAAAAAAAAAAAAAAAAAAAAA\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) == 0 {
		t.Fatalf("expected format errors but got none")
	}

	// When ranking is skipped, any staleness entry for ROOT/b should have rank=0.
	for _, se := range report.Staleness {
		if se.Node == "ROOT/b" && se.Rank != 0 {
			t.Errorf("expected rank=0 for ROOT/b when ranking is skipped, got %d", se.Rank)
		}
	}
}

// --------------------------------------------------------------------------
// Edge Cases
// --------------------------------------------------------------------------

// TC-EC-01: Empty spec tree — scan fails, FormatError with rule="scan".
func TestMCPValidateSpecs_EmptySpecTree(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// No code-from-spec/ directory created.

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, fe := range report.FormatErrors {
		if fe.Rule == "scan" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected FormatError with rule=scan; got: %v", report.FormatErrors)
	}
	if len(report.Cycles) != 0 {
		t.Errorf("expected no cycles, got %v", report.Cycles)
	}
	if len(report.Staleness) != 0 {
		t.Errorf("expected no staleness entries, got %v", report.Staleness)
	}
}

// TC-EC-02: Node with no outputs — not in staleness.
func TestMCPValidateSpecs_NodeWithNoOutputs(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", ""))

	report := mcpvalidatespecs.MCPValidateSpecs()

	for _, se := range report.Staleness {
		if se.Node == "ROOT/a" {
			t.Errorf("did not expect any StalenessEntry for ROOT/a (no outputs); got: %v", se)
		}
	}
}
