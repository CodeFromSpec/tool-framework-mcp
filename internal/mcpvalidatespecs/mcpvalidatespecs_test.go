// code-from-spec: ROOT/golang/tests/mcp_tools/validate_specs@pge45X_Yh3sgpqR6YI5LZILFFFE
package mcpvalidatespecs_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

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

// testWriteFile creates the file at path (relative to cwd), creating parent
// directories as needed, and writes content.
func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("testWriteFile mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile write %s: %v", path, err)
	}
}

// testRootNode returns a minimal valid _node.md for the ROOT node.
func testRootNode() string {
	return "---\n---\n# root\n\nRoot node.\n"
}

// testLeafNode returns a minimal valid _node.md for a leaf node with the given
// heading and frontmatter block.
func testLeafNode(heading string, frontmatter string) string {
	return fmt.Sprintf("---\n%s---\n# %s\n\nLeaf node.\n", frontmatter, heading)
}

// testFindStalenessEntry returns the StalenessEntry with the given node name
// and output id, or nil if not found.
func testFindStalenessEntry(entries []*mcpvalidatespecs.StalenessEntry, node, outputID string) *mcpvalidatespecs.StalenessEntry {
	for _, e := range entries {
		if e.Node == node && e.OutputID == outputID {
			return e
		}
	}
	return nil
}

// testFindFormatError returns the first FormatError with the given node and rule,
// or nil if not found.
func testFindFormatError(errs []*mcpvalidatespecs.FormatError, node, rule string) *mcpvalidatespecs.FormatError {
	for _, e := range errs {
		if e.Node == node && e.Rule == rule {
			return e
		}
	}
	return nil
}

// TC-HP-01: Clean tree — no errors.
func TestMCPValidateSpecs_HP01_CleanTree(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNode("root/a", "outputs:\n  - id: code\n    path: out/a.go\n"))

	// First pass to discover the current chain hash for ROOT/a.
	report := mcpvalidatespecs.MCPValidateSpecs()
	var currentHash string
	for _, se := range report.Staleness {
		if se.Node == "ROOT/a" && se.OutputID == "code" {
			// The hash embedded in the detail or we need another approach.
			// We rely on the fact that if we write a tag with an arbitrary hash, it
			// will show as stale. We need to figure out the real hash by running a
			// second pass after writing a matching tag — but we don't have the hash
			// yet. Instead, run once to get a stale entry (which tells us what the
			// real hash is), then write the correct tag and verify the tree is clean.
			//
			// The StalenessEntry does not expose the expected hash directly.
			// Per the spec guidance, we call MCPValidateSpecs once to discover the
			// chain hash. Since it doesn't return the hash, we instead write a
			// placeholder, note the entry is stale, then use a different strategy:
			// write the tag with the hash extracted from the detail field.
			_ = se
		}
	}
	// Strategy: write an arbitrary tag first so the file exists, then inspect
	// the detail field of the stale entry to find the expected hash, then rewrite.
	testWriteFile(t, "out/a.go", "// code-from-spec: ROOT/a@AAAAAAAAAAAAAAAAAAAAAAAAAA_\n")

	report = mcpvalidatespecs.MCPValidateSpecs()
	se := testFindStalenessEntry(report.Staleness, "ROOT/a", "code")
	if se == nil {
		t.Fatal("expected a stale entry for ROOT/a after writing dummy tag, got none")
	}
	// The detail field contains the expected hash (format: "expected <hash>, got <hash>").
	// Extract it by parsing the detail string.
	var expectedHash string
	fmt.Sscanf(se.Detail, "expected %s", &expectedHash)
	// Remove trailing comma or extra chars.
	if len(expectedHash) > 27 {
		expectedHash = expectedHash[:27]
	}
	if len(expectedHash) != 27 {
		t.Fatalf("could not extract expected hash from detail %q", se.Detail)
	}
	currentHash = expectedHash

	// Write a correctly tagged file.
	testWriteFile(t, "out/a.go", fmt.Sprintf("// code-from-spec: ROOT/a@%s\n", currentHash))

	report = mcpvalidatespecs.MCPValidateSpecs()

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

// TC-HP-02: Stale artifact detected.
func TestMCPValidateSpecs_HP02_StaleArtifact(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNode("root/a", "outputs:\n  - id: code\n    path: out/a.go\n"))
	// Outdated hash (wrong length would be malformed, so use a valid 27-char but wrong hash).
	testWriteFile(t, "out/a.go", "// code-from-spec: ROOT/a@AAAAAAAAAAAAAAAAAAAAAAAAAA_\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	se := testFindStalenessEntry(report.Staleness, "ROOT/a", "code")
	if se == nil {
		t.Fatalf("expected a StalenessEntry for ROOT/a/code, got none; staleness=%v", report.Staleness)
	}
	if se.Status != "stale" {
		t.Errorf("expected status=stale, got %q", se.Status)
	}
	if se.Rank < 0 {
		t.Errorf("expected non-negative rank, got %d", se.Rank)
	}
}

// TC-HP-03: Missing artifact detected.
func TestMCPValidateSpecs_HP03_MissingArtifact(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNode("root/a", "outputs:\n  - id: code\n    path: out/a.go\n"))
	// Do NOT create out/a.go.

	report := mcpvalidatespecs.MCPValidateSpecs()

	se := testFindStalenessEntry(report.Staleness, "ROOT/a", "code")
	if se == nil {
		t.Fatalf("expected a StalenessEntry for ROOT/a/code, got none; staleness=%v", report.Staleness)
	}
	if se.Status != "missing" {
		t.Errorf("expected status=missing, got %q", se.Status)
	}
}

// TC-HP-04: Malformed tag detected.
func TestMCPValidateSpecs_HP04_MalformedTag(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNode("root/a", "outputs:\n  - id: code\n    path: out/a.go\n"))
	// File has no parseable artifact tag.
	testWriteFile(t, "out/a.go", "package main\n// no tag here\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	se := testFindStalenessEntry(report.Staleness, "ROOT/a", "code")
	if se == nil {
		t.Fatalf("expected a StalenessEntry for ROOT/a/code, got none; staleness=%v", report.Staleness)
	}
	if se.Status != "malformed tag" {
		t.Errorf("expected status=\"malformed tag\", got %q", se.Status)
	}
}

// TC-HP-05: Multiple outputs — each checked independently.
func TestMCPValidateSpecs_HP05_MultipleOutputs(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNode("root/a",
		"outputs:\n  - id: x\n    path: out/x.go\n  - id: y\n    path: out/y.go\n"))

	// Write x.go with a dummy stale tag first, then discover the real hash.
	testWriteFile(t, "out/x.go", "// code-from-spec: ROOT/a@AAAAAAAAAAAAAAAAAAAAAAAAAA_\n")

	report := mcpvalidatespecs.MCPValidateSpecs()
	seX := testFindStalenessEntry(report.Staleness, "ROOT/a", "x")
	if seX == nil {
		t.Fatal("expected stale entry for x in first pass")
	}
	var expectedHash string
	fmt.Sscanf(seX.Detail, "expected %s", &expectedHash)
	if len(expectedHash) > 27 {
		expectedHash = expectedHash[:27]
	}
	if len(expectedHash) != 27 {
		t.Fatalf("could not extract expected hash from detail %q", seX.Detail)
	}

	// Write x.go with the correct hash; do NOT create y.go.
	testWriteFile(t, "out/x.go", fmt.Sprintf("// code-from-spec: ROOT/a@%s\n", expectedHash))

	report = mcpvalidatespecs.MCPValidateSpecs()

	// x should not appear in staleness.
	if testFindStalenessEntry(report.Staleness, "ROOT/a", "x") != nil {
		t.Error("expected no StalenessEntry for x (hash matches), but one was found")
	}

	// y should appear as missing.
	seY := testFindStalenessEntry(report.Staleness, "ROOT/a", "y")
	if seY == nil {
		t.Fatalf("expected a StalenessEntry for ROOT/a/y, got none; staleness=%v", report.Staleness)
	}
	if seY.Status != "missing" {
		t.Errorf("expected status=missing for y, got %q", seY.Status)
	}
}

// TC-HP-06: Staleness entries include rank.
func TestMCPValidateSpecs_HP06_StalenessRank(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNode("root/a",
		"outputs:\n  - id: code-a\n    path: out/a.go\n"))
	testWriteFile(t, "code-from-spec/b/_node.md", testLeafNode("root/b",
		"depends_on:\n  - ROOT/a\noutputs:\n  - id: code-b\n    path: out/b.go\n"))

	// Outdated hashes.
	testWriteFile(t, "out/a.go", "// code-from-spec: ROOT/a@AAAAAAAAAAAAAAAAAAAAAAAAAA_\n")
	testWriteFile(t, "out/b.go", "// code-from-spec: ROOT/b@AAAAAAAAAAAAAAAAAAAAAAAAAA_\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	seA := testFindStalenessEntry(report.Staleness, "ROOT/a", "code-a")
	seB := testFindStalenessEntry(report.Staleness, "ROOT/b", "code-b")

	if seA == nil {
		t.Fatal("expected StalenessEntry for ROOT/a, got none")
	}
	if seB == nil {
		t.Fatal("expected StalenessEntry for ROOT/b, got none")
	}

	if seA.Rank < 0 {
		t.Errorf("expected non-negative rank for ROOT/a, got %d", seA.Rank)
	}
	if seB.Rank < 0 {
		t.Errorf("expected non-negative rank for ROOT/b, got %d", seB.Rank)
	}
	if seA.Rank >= seB.Rank {
		t.Errorf("expected rank(ROOT/a) < rank(ROOT/b), got %d >= %d", seA.Rank, seB.Rank)
	}
}

// TC-HP-07: Staleness ordered by rank then name.
func TestMCPValidateSpecs_HP07_StalenessOrder(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/z/_node.md", testLeafNode("root/z",
		"outputs:\n  - id: code-z\n    path: out/z.go\n"))
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNode("root/a",
		"outputs:\n  - id: code-a\n    path: out/a.go\n"))

	// Both stale (outdated hashes).
	testWriteFile(t, "out/z.go", "// code-from-spec: ROOT/z@AAAAAAAAAAAAAAAAAAAAAAAAAA_\n")
	testWriteFile(t, "out/a.go", "// code-from-spec: ROOT/a@AAAAAAAAAAAAAAAAAAAAAAAAAA_\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) < 2 {
		t.Fatalf("expected at least 2 staleness entries, got %d", len(report.Staleness))
	}

	var idxA, idxZ int = -1, -1
	for i, se := range report.Staleness {
		if se.Node == "ROOT/a" {
			idxA = i
		}
		if se.Node == "ROOT/z" {
			idxZ = i
		}
	}

	if idxA == -1 {
		t.Fatal("expected StalenessEntry for ROOT/a")
	}
	if idxZ == -1 {
		t.Fatal("expected StalenessEntry for ROOT/z")
	}
	if idxA >= idxZ {
		t.Errorf("expected ROOT/a (index %d) before ROOT/z (index %d) in staleness list", idxA, idxZ)
	}
}

// TC-FE-01: Format error from invalid depends_on.
func TestMCPValidateSpecs_FE01_InvalidDependsOn(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNode("root/a",
		"depends_on:\n  - ROOT/missing\n"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	fe := testFindFormatError(report.FormatErrors, "ROOT/a", "dependency_targets")
	if fe == nil {
		t.Errorf("expected FormatError for ROOT/a with rule=dependency_targets, got format_errors=%v", report.FormatErrors)
	}
}

// TC-FE-02: Format error from parse failure.
func TestMCPValidateSpecs_FE02_ParseFailure(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	// Invalid: body text before any heading.
	testWriteFile(t, "code-from-spec/a/_node.md", "---\n---\nThis text appears before any heading\n# root/a\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	fe := testFindFormatError(report.FormatErrors, "ROOT/a", "parse")
	if fe == nil {
		t.Errorf("expected FormatError for ROOT/a with rule=parse, got format_errors=%v", report.FormatErrors)
	}
}

// TC-FE-03: Continues after parse failure.
func TestMCPValidateSpecs_FE03_ContinuesAfterParseFailure(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	// ROOT/a is unparseable.
	testWriteFile(t, "code-from-spec/a/_node.md", "---\n---\nThis text appears before any heading\n# root/a\n")
	// ROOT/b is valid with a stale output.
	testWriteFile(t, "code-from-spec/b/_node.md", testLeafNode("root/b",
		"outputs:\n  - id: code-b\n    path: out/b.go\n"))
	testWriteFile(t, "out/b.go", "// code-from-spec: ROOT/b@AAAAAAAAAAAAAAAAAAAAAAAAAA_\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	feA := testFindFormatError(report.FormatErrors, "ROOT/a", "parse")
	if feA == nil {
		t.Errorf("expected FormatError for ROOT/a with rule=parse, got format_errors=%v", report.FormatErrors)
	}

	seB := testFindStalenessEntry(report.Staleness, "ROOT/b", "code-b")
	if seB == nil {
		t.Errorf("expected StalenessEntry for ROOT/b, got staleness=%v", report.Staleness)
	}
}

// TC-CY-01: Simple cycle detected.
func TestMCPValidateSpecs_CY01_SimpleCycle(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNode("root/a",
		"depends_on:\n  - ROOT/b\n"))
	testWriteFile(t, "code-from-spec/b/_node.md", testLeafNode("root/b",
		"depends_on:\n  - ROOT/a\n"))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Cycles) == 0 {
		t.Fatal("expected cycles to be non-empty")
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

// TC-CY-02: Ranking skipped when format errors exist.
func TestMCPValidateSpecs_CY02_RankingSkippedOnFormatErrors(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	// ROOT/a has an invalid dependency (format error).
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNode("root/a",
		"depends_on:\n  - ROOT/missing\n"))
	// ROOT/b is valid with a stale output.
	testWriteFile(t, "code-from-spec/b/_node.md", testLeafNode("root/b",
		"outputs:\n  - id: code-b\n    path: out/b.go\n"))
	testWriteFile(t, "out/b.go", "// code-from-spec: ROOT/b@AAAAAAAAAAAAAAAAAAAAAAAAAA_\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) == 0 {
		t.Fatal("expected format_errors to be non-empty")
	}

	seB := testFindStalenessEntry(report.Staleness, "ROOT/b", "code-b")
	if seB == nil {
		// Ranking is skipped; staleness may still be reported but with rank=0.
		// If no entry exists the output was not checked — that is also acceptable
		// since ranking is skipped. But per the spec guidance, when format errors
		// exist, any StalenessEntry has rank=0. We only assert rank=0 if present.
		return
	}
	if seB.Rank != 0 {
		t.Errorf("expected rank=0 for ROOT/b when ranking is skipped (format errors present), got %d", seB.Rank)
	}
}

// TC-EC-01: Empty spec tree — scan fails.
func TestMCPValidateSpecs_EC01_EmptySpecTree(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	// Do NOT create code-from-spec/ directory.

	report := mcpvalidatespecs.MCPValidateSpecs()

	fe := testFindFormatError(report.FormatErrors, "", "scan")
	if fe == nil {
		// The node field may be non-empty (implementation-defined). Accept any
		// FormatError with rule="scan".
		found := false
		for _, e := range report.FormatErrors {
			if e.Rule == "scan" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected a FormatError with rule=scan, got format_errors=%v", report.FormatErrors)
		}
	}

	if len(report.Cycles) != 0 {
		t.Errorf("expected no cycles, got %v", report.Cycles)
	}
	if len(report.Staleness) != 0 {
		t.Errorf("expected no staleness entries, got %d", len(report.Staleness))
	}
}

// TC-EC-02: Node with no outputs — not in staleness.
func TestMCPValidateSpecs_EC02_NodeWithNoOutputs(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
	testWriteFile(t, "code-from-spec/a/_node.md", testLeafNode("root/a", ""))

	report := mcpvalidatespecs.MCPValidateSpecs()

	if testFindStalenessEntry(report.Staleness, "ROOT/a", "") != nil {
		t.Error("expected no StalenessEntry for ROOT/a (no outputs declared)")
	}
	for _, se := range report.Staleness {
		if se.Node == "ROOT/a" {
			t.Errorf("unexpected StalenessEntry for ROOT/a: %+v", se)
		}
	}
}
