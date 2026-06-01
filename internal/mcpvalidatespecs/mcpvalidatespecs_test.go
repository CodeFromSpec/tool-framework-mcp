// code-from-spec: ROOT/golang/tests/mcp_tools/validate_specs@Va3ERTH-pUV5f9_d0GmeP_OSaKA
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

// testChdir changes the working directory to dir for the duration of
// the test, restoring the original directory on cleanup.
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

// testWriteFile writes data to path (relative to cwd), creating parent
// directories as needed.
func testWriteFile(t *testing.T, path string, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("testWriteFile mkdir %q: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("testWriteFile %q: %v", path, err)
	}
}

// testComputeChainHash computes the chain hash for the given logical
// name using chainresolver and chainhash. Must be called after
// testChdir has set the working directory to the temp spec tree root.
func testComputeChainHash(t *testing.T, logicalName string) string {
	t.Helper()
	chain, err := chainresolver.ChainResolve(logicalName)
	if err != nil {
		t.Fatalf("testComputeChainHash ChainResolve(%s): %v", logicalName, err)
	}
	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("testComputeChainHash ChainHashCompute(%s): %v", logicalName, err)
	}
	return hash
}

// testRootNodeContent returns minimal valid content for a ROOT _node.md.
func testRootNodeContent() string {
	return "---\n---\n# ROOT\n\n## Public\n\nRoot public section.\n"
}

// testLeafNodeContent returns minimal valid content for a leaf _node.md
// with the given logical name and no frontmatter fields.
func testLeafNodeContent(logicalName string) string {
	return fmt.Sprintf("---\n---\n# %s\n", logicalName)
}

// testLeafWithOutputs returns valid _node.md content for a leaf that
// declares outputs. outputsYAML is a YAML snippet like:
//
//	"outputs:\n  - id: code\n    path: out/a.go"
func testLeafWithOutputs(logicalName, outputsYAML string) string {
	return fmt.Sprintf("---\n%s\n---\n# %s\n", outputsYAML, logicalName)
}

// testLeafWithDependsOn returns valid _node.md content for a leaf that
// declares depends_on. dependsOnYAML is a YAML snippet like:
//
//	"depends_on:\n  - ROOT/a"
func testLeafWithDependsOn(logicalName, dependsOnYAML string) string {
	return fmt.Sprintf("---\n%s\n---\n# %s\n", dependsOnYAML, logicalName)
}

// testLeafWithOutputsAndDependsOn returns valid _node.md content for a
// leaf that declares both outputs and depends_on.
func testLeafWithOutputsAndDependsOn(logicalName, outputsYAML, dependsOnYAML string) string {
	return fmt.Sprintf("---\n%s\n%s\n---\n# %s\n", outputsYAML, dependsOnYAML, logicalName)
}

// testArtifactContent returns Go file content with a valid artifact tag
// for the given logical name and hash.
func testArtifactContent(logicalName, hash string) string {
	return fmt.Sprintf("// code-from-spec: %s@%s\npackage example\n", logicalName, hash)
}

// testStaleArtifactContent returns Go file content with an artifact tag
// that has an intentionally wrong (stale) hash.
func testStaleArtifactContent(logicalName string) string {
	// The hash is exactly 27 characters — use an arbitrary but well-formed value.
	return fmt.Sprintf("// code-from-spec: %s@AAAAAAAAAAAAAAAAAAAAAAAAAA_\npackage example\n", logicalName)
}

// testNoTagContent returns Go file content with no artifact tag.
func testNoTagContent() string {
	return "package example\n// no code-from-spec tag present\n"
}

// --- Happy Path ---

// TestCleanTree_NoErrors verifies that MCPValidateSpecs returns an
// empty ValidationReport when the spec tree is valid and all artifacts
// are up to date.
func TestCleanTree_NoErrors(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())
	testWriteFile(t, "code-from-spec/a/_node.md",
		testLeafWithOutputs("ROOT/a", "outputs:\n  - id: code\n    path: out/a.go"))

	// Compute the current chain hash for ROOT/a.
	hash := testComputeChainHash(t, "ROOT/a")

	// Write the artifact with the correct hash.
	testWriteFile(t, "out/a.go", testArtifactContent("ROOT/a", hash))

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

// TestStaleArtifactDetected verifies that MCPValidateSpecs reports a
// stale entry when the artifact file exists but has a non-matching hash.
func TestStaleArtifactDetected(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())
	testWriteFile(t, "code-from-spec/a/_node.md",
		testLeafWithOutputs("ROOT/a", "outputs:\n  - id: code\n    path: out/a.go"))

	// Write a stale artifact (hash does not match).
	testWriteFile(t, "out/a.go", testStaleArtifactContent("ROOT/a"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 1 {
		t.Fatalf("expected 1 staleness entry, got %d", len(report.Staleness))
	}
	se := report.Staleness[0]
	if se.Node != "ROOT/a" {
		t.Errorf("expected node ROOT/a, got %s", se.Node)
	}
	if se.OutputID != "code" {
		t.Errorf("expected output_id code, got %s", se.OutputID)
	}
	if se.Status != "stale" {
		t.Errorf("expected status stale, got %s", se.Status)
	}
	// Rank is present; its exact value is not verified here.
	_ = se.Rank
}

// TestMissingArtifactDetected verifies that MCPValidateSpecs reports a
// missing entry when the output file does not exist.
func TestMissingArtifactDetected(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())
	testWriteFile(t, "code-from-spec/a/_node.md",
		testLeafWithOutputs("ROOT/a", "outputs:\n  - id: code\n    path: out/a.go"))

	// Do not create "out/a.go".

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 1 {
		t.Fatalf("expected 1 staleness entry, got %d", len(report.Staleness))
	}
	se := report.Staleness[0]
	if se.Node != "ROOT/a" {
		t.Errorf("expected node ROOT/a, got %s", se.Node)
	}
	if se.OutputID != "code" {
		t.Errorf("expected output_id code, got %s", se.OutputID)
	}
	if se.Status != "missing" {
		t.Errorf("expected status missing, got %s", se.Status)
	}
}

// TestMalformedTagDetected verifies that MCPValidateSpecs reports a
// malformed-tag entry when the artifact file contains no artifact tag.
func TestMalformedTagDetected(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())
	testWriteFile(t, "code-from-spec/a/_node.md",
		testLeafWithOutputs("ROOT/a", "outputs:\n  - id: code\n    path: out/a.go"))

	// Write a file with no artifact tag.
	testWriteFile(t, "out/a.go", testNoTagContent())

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 1 {
		t.Fatalf("expected 1 staleness entry, got %d", len(report.Staleness))
	}
	se := report.Staleness[0]
	if se.Node != "ROOT/a" {
		t.Errorf("expected node ROOT/a, got %s", se.Node)
	}
	if se.OutputID != "code" {
		t.Errorf("expected output_id code, got %s", se.OutputID)
	}
	if se.Status != "malformed tag" {
		t.Errorf("expected status 'malformed tag', got %s", se.Status)
	}
}

// TestMultipleOutputsCheckedIndependently verifies that when a node
// declares multiple outputs, each is checked independently. A matching
// artifact is not reported; a missing one is.
func TestMultipleOutputsCheckedIndependently(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())
	testWriteFile(t, "code-from-spec/a/_node.md",
		testLeafWithOutputs("ROOT/a",
			"outputs:\n  - id: x\n    path: out/x.go\n  - id: y\n    path: out/y.go"))

	// Compute the chain hash and write a valid artifact for output x.
	hash := testComputeChainHash(t, "ROOT/a")
	testWriteFile(t, "out/x.go", testArtifactContent("ROOT/a", hash))

	// Do not create "out/y.go".

	report := mcpvalidatespecs.MCPValidateSpecs()

	// No staleness entry for output_id "x".
	for _, se := range report.Staleness {
		if se.Node == "ROOT/a" && se.OutputID == "x" {
			t.Errorf("unexpected staleness entry for output_id x (hash matches)")
		}
	}

	// Exactly one staleness entry for output_id "y" with status "missing".
	var foundY bool
	for _, se := range report.Staleness {
		if se.Node == "ROOT/a" && se.OutputID == "y" {
			foundY = true
			if se.Status != "missing" {
				t.Errorf("expected status missing for output_id y, got %s", se.Status)
			}
		}
	}
	if !foundY {
		t.Errorf("expected staleness entry for output_id y, found none")
	}
}

// TestStalenessEntriesIncludeRank verifies that staleness entries carry
// a rank, and that ROOT/a (no dependencies) ranks lower than ROOT/b
// (which depends on ROOT/a).
func TestStalenessEntriesIncludeRank(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())
	testWriteFile(t, "code-from-spec/a/_node.md",
		testLeafWithOutputs("ROOT/a", "outputs:\n  - id: code\n    path: out/a.go"))
	testWriteFile(t, "code-from-spec/b/_node.md",
		testLeafWithOutputsAndDependsOn("ROOT/b",
			"outputs:\n  - id: code\n    path: out/b.go",
			"depends_on:\n  - ROOT/a"))

	// Write stale artifacts for both.
	testWriteFile(t, "out/a.go", testStaleArtifactContent("ROOT/a"))
	testWriteFile(t, "out/b.go", testStaleArtifactContent("ROOT/b"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	var rankA, rankB int
	var foundA, foundB bool
	for _, se := range report.Staleness {
		switch se.Node {
		case "ROOT/a":
			rankA = se.Rank
			foundA = true
		case "ROOT/b":
			rankB = se.Rank
			foundB = true
		}
	}

	if !foundA {
		t.Fatal("expected staleness entry for ROOT/a")
	}
	if !foundB {
		t.Fatal("expected staleness entry for ROOT/b")
	}
	if rankA >= rankB {
		t.Errorf("expected rank of ROOT/a (%d) < rank of ROOT/b (%d)", rankA, rankB)
	}
}

// TestStalenessEntriesOrderedByRankThenName verifies that entries with
// the same rank appear in alphabetical order by node name.
func TestStalenessEntriesOrderedByRankThenName(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())
	testWriteFile(t, "code-from-spec/z/_node.md",
		testLeafWithOutputs("ROOT/z", "outputs:\n  - id: code\n    path: out/z.go"))
	testWriteFile(t, "code-from-spec/a/_node.md",
		testLeafWithOutputs("ROOT/a", "outputs:\n  - id: code\n    path: out/a.go"))

	// Write stale artifacts for both.
	testWriteFile(t, "out/z.go", testStaleArtifactContent("ROOT/z"))
	testWriteFile(t, "out/a.go", testStaleArtifactContent("ROOT/a"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) < 2 {
		t.Fatalf("expected at least 2 staleness entries, got %d", len(report.Staleness))
	}

	idxA, idxZ := -1, -1
	for i, se := range report.Staleness {
		switch se.Node {
		case "ROOT/a":
			idxA = i
		case "ROOT/z":
			idxZ = i
		}
	}

	if idxA == -1 {
		t.Fatal("expected staleness entry for ROOT/a")
	}
	if idxZ == -1 {
		t.Fatal("expected staleness entry for ROOT/z")
	}

	rankA := report.Staleness[idxA].Rank
	rankZ := report.Staleness[idxZ].Rank
	if rankA != rankZ {
		t.Errorf("expected ROOT/a rank (%d) == ROOT/z rank (%d)", rankA, rankZ)
	}
	if idxA >= idxZ {
		t.Errorf("expected ROOT/a (idx %d) before ROOT/z (idx %d) in staleness list", idxA, idxZ)
	}
}

// --- Format Errors ---

// TestFormatError_InvalidDependsOn verifies that a format error is
// reported when depends_on references a logical name not in the tree.
func TestFormatError_InvalidDependsOn(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())
	testWriteFile(t, "code-from-spec/a/_node.md",
		testLeafWithDependsOn("ROOT/a", "depends_on:\n  - ROOT/missing"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	var found bool
	for _, fe := range report.FormatErrors {
		if fe.Node == "ROOT/a" && fe.Rule == "dependency_targets" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected FormatError for ROOT/a with rule dependency_targets, got: %v", report.FormatErrors)
	}
}

// TestFormatError_ParseFailure verifies that a format error with rule
// "parse" is reported when a node file contains content before the
// first heading (making it unparseable).
func TestFormatError_ParseFailure(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())

	// ROOT/a has non-blank content before the first heading.
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\n---\nThis text before any heading makes the file unparseable.\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	var found bool
	for _, fe := range report.FormatErrors {
		if fe.Node == "ROOT/a" && fe.Rule == "parse" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected FormatError for ROOT/a with rule parse, got: %v", report.FormatErrors)
	}
}

// TestContinuesAfterParseFailure verifies that after a parse failure
// for ROOT/a, MCPValidateSpecs continues to process ROOT/b and reports
// its staleness.
func TestContinuesAfterParseFailure(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())

	// ROOT/a is unparseable.
	testWriteFile(t, "code-from-spec/a/_node.md",
		"---\n---\nInvalid content before any heading.\n")

	// ROOT/b is valid with a stale output.
	testWriteFile(t, "code-from-spec/b/_node.md",
		testLeafWithOutputs("ROOT/b", "outputs:\n  - id: code\n    path: out/b.go"))
	testWriteFile(t, "out/b.go", testStaleArtifactContent("ROOT/b"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	// Parse error for ROOT/a.
	var foundParseError bool
	for _, fe := range report.FormatErrors {
		if fe.Node == "ROOT/a" && fe.Rule == "parse" {
			foundParseError = true
			break
		}
	}
	if !foundParseError {
		t.Errorf("expected parse FormatError for ROOT/a, got: %v", report.FormatErrors)
	}

	// Staleness entry for ROOT/b.
	var foundStaleness bool
	for _, se := range report.Staleness {
		if se.Node == "ROOT/b" && se.Status == "stale" {
			foundStaleness = true
			break
		}
	}
	if !foundStaleness {
		t.Errorf("expected staleness entry for ROOT/b with status stale, got: %v", report.Staleness)
	}
}

// --- Cycle Detection ---

// TestSimpleCycleDetected verifies that MCPValidateSpecs detects a
// cycle when ROOT/a depends on ROOT/b and ROOT/b depends on ROOT/a.
func TestSimpleCycleDetected(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())
	testWriteFile(t, "code-from-spec/a/_node.md",
		testLeafWithDependsOn("ROOT/a", "depends_on:\n  - ROOT/b"))
	testWriteFile(t, "code-from-spec/b/_node.md",
		testLeafWithDependsOn("ROOT/b", "depends_on:\n  - ROOT/a"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Cycles) == 0 {
		t.Fatal("expected non-empty cycles list")
	}

	var found bool
	for _, name := range report.Cycles {
		if name == "ROOT/a" || name == "ROOT/b" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected ROOT/a or ROOT/b in cycles, got: %v", report.Cycles)
	}
}

// TestRankingSkippedWhenFormatErrorsExist verifies that when format
// errors are present, staleness entries receive rank 0 (ranking is
// skipped).
func TestRankingSkippedWhenFormatErrorsExist(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())
	// ROOT/a has an invalid depends_on target — causes a format error.
	testWriteFile(t, "code-from-spec/a/_node.md",
		testLeafWithDependsOn("ROOT/a", "depends_on:\n  - ROOT/missing"))
	// ROOT/b is valid with a stale output.
	testWriteFile(t, "code-from-spec/b/_node.md",
		testLeafWithOutputs("ROOT/b", "outputs:\n  - id: code\n    path: out/b.go"))
	testWriteFile(t, "out/b.go", testStaleArtifactContent("ROOT/b"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) == 0 {
		t.Fatal("expected format errors")
	}

	var foundB bool
	for _, se := range report.Staleness {
		if se.Node == "ROOT/b" {
			foundB = true
			if se.Rank != 0 {
				t.Errorf("expected rank 0 when ranking is skipped, got %d", se.Rank)
			}
			break
		}
	}
	if !foundB {
		t.Errorf("expected staleness entry for ROOT/b, got: %v", report.Staleness)
	}
}

// --- Edge Cases ---

// TestEmptySpecTree_ScanFails verifies that when the code-from-spec/
// directory does not exist, MCPValidateSpecs returns a format error
// with rule "scan", empty cycles, and empty staleness.
func TestEmptySpecTree_ScanFails(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Do NOT create the code-from-spec/ directory.

	report := mcpvalidatespecs.MCPValidateSpecs()

	var found bool
	for _, fe := range report.FormatErrors {
		if fe.Rule == "scan" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected FormatError with rule scan, got: %v", report.FormatErrors)
	}
	if len(report.Cycles) != 0 {
		t.Errorf("expected no cycles, got %v", report.Cycles)
	}
	if len(report.Staleness) != 0 {
		t.Errorf("expected no staleness entries, got %v", report.Staleness)
	}
}

// TestNodeWithNoOutputs_NotInStaleness verifies that nodes with no
// declared outputs do not appear in the staleness list.
func TestNodeWithNoOutputs_NotInStaleness(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNodeContent())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNodeContent("ROOT/a"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	for _, se := range report.Staleness {
		if se.Node == "ROOT/a" {
			t.Errorf("unexpected staleness entry for ROOT/a (no outputs declared)")
		}
	}
}
