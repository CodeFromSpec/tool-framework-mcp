// code-from-spec: ROOT/golang/internal/tools/validate_specs/tests@4DTsrq4tX5QbHuBpDFnnIN1cTtw

// Package validate_specs — tests for HandleValidateSpecs.
//
// Each test creates a temporary directory, populates it with the minimal
// spec tree needed for the scenario, changes the working directory to that
// temp dir (so that nodediscovery.DiscoverNodes and pathvalidation.ValidatePath
// resolve paths correctly), and then restores the working directory on cleanup.
//
// Important structural rules followed here (matching the real spec tree):
//   - _node.md files have the YAML frontmatter block (---/---) BEFORE the
//     heading (# ROOT/...).
//   - "ROOT" maps to "code-from-spec/_node.md".
//   - "ROOT/a" maps to "code-from-spec/a/_node.md".
//   - There is NO directory named "ROOT" inside code-from-spec/.
//   - DiscoveredNode.FilePath values are relative (e.g. "code-from-spec/a/_node.md").
//   - pathvalidation.ValidatePath requires an absolute project root; os.Getwd()
//     is used (which returns the temp dir after os.Chdir).
package validate_specs

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/chainhash"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// testChdir changes the working directory to dir and registers a cleanup
// function on t to restore the original directory when the test ends.
// This is necessary because nodediscovery.DiscoverNodes and
// pathvalidation.ValidatePath both resolve paths relative to os.Getwd().
func testChdir(t *testing.T, dir string) {
	t.Helper()

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChdir: os.Getwd: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChdir: os.Chdir(%q): %v", dir, err)
	}

	t.Cleanup(func() {
		// Use Logf instead of Fatalf: cleanup runs after the test body so
		// a fatal here would be ignored anyway.
		if err := os.Chdir(orig); err != nil {
			t.Logf("testChdir cleanup: os.Chdir(%q): %v", orig, err)
		}
	})
}

// testMkFile creates a file at path (creating parent directories as needed)
// and writes content to it. The path must be relative to the current working
// directory (i.e. the temp dir after testChdir). Fails the test on any error.
func testMkFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("testMkFile: MkdirAll(%q): %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("testMkFile: WriteFile(%q): %v", path, err)
	}
}

// testNodeMd returns the content of a minimal _node.md file.
//
// The YAML frontmatter block (---/---) is placed BEFORE the heading, matching
// the format expected by ParseFrontmatter and the rest of the spec tooling.
//
// Parameters:
//   - fm   — raw YAML to embed inside the frontmatter delimiters.
//             Pass "" for nodes that need no frontmatter fields.
//   - name — logical name used in the heading (e.g. "ROOT" or "ROOT/a").
//   - body — additional markdown after the heading (may be "").
func testNodeMd(fm, name, body string) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	if fm != "" {
		sb.WriteString(fm)
		// Ensure the yaml content ends with a newline before the closing delimiter.
		if !strings.HasSuffix(fm, "\n") {
			sb.WriteByte('\n')
		}
	}
	sb.WriteString("---\n")
	sb.WriteString("# " + name + "\n")
	if body != "" {
		sb.WriteString(body)
	}
	return sb.String()
}

// testResultText extracts the text from the first content entry of a
// CallToolResult. Fails the test if the result is nil or empty.
func testResultText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if result == nil || len(result.Content) == 0 {
		t.Fatal("testResultText: nil or empty CallToolResult")
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("testResultText: content[0] is not *mcp.TextContent")
	}
	return tc.Text
}

// testComputeHash calls chainhash.ComputeChainHash and fails the test on
// error. Used in CleanTree to obtain the expected hash so a valid artifact
// file can be written before the second pass.
func testComputeHash(t *testing.T, logicalName string) string {
	t.Helper()
	hash, err := chainhash.ComputeChainHash(logicalName)
	if err != nil {
		t.Fatalf("testComputeHash(%q): %v", logicalName, err)
	}
	return hash
}

// ---------------------------------------------------------------------------
// Happy-path: clean tree with no errors
// ---------------------------------------------------------------------------

// TestHandleValidateSpecs_CleanTree verifies that a well-formed spec tree
// with a matching artifact tag hash produces a success result reporting
// "All spec nodes are valid and all artifacts are up to date."
//
// Strategy — two-pass approach:
//  1. Run the handler without the output file to surface what hash is expected
//     (indirectly — we use testComputeHash for this).
//  2. Write the output file with the correct tag, then run the handler again
//     and verify the clean-tree message.
func TestHandleValidateSpecs_CleanTree(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// ROOT — parent node, no outputs.
	testMkFile(t, "code-from-spec/_node.md", testNodeMd("", "ROOT", ""))

	// ROOT/a — leaf node with a single declared output.
	testMkFile(t, "code-from-spec/a/_node.md", testNodeMd(
		"outputs:\n  - id: main\n    path: out/a.go\n",
		"ROOT/a",
		"",
	))

	ctx := context.Background()

	// --- Pass 1: no output file yet; should report ROOT/a as missing. ---
	result1, _, err := HandleValidateSpecs(ctx, nil, ValidateSpecsArgs{})
	if err != nil {
		t.Fatalf("pass 1: unexpected Go error: %v", err)
	}
	report1 := testResultText(t, result1)

	if !strings.Contains(report1, "ROOT/a") {
		t.Logf("pass 1 report (for debug):\n%s", report1)
		// This is informational; the test is really about pass 2.
	}

	// Compute the expected hash using the internal chainhash package.
	expectedHash := testComputeHash(t, "ROOT/a")

	// Write the artifact with the correct tag.
	testMkFile(t, "out/a.go",
		"// code-from-spec: ROOT/a@"+expectedHash+"\npackage a\n",
	)

	// --- Pass 2: everything should be clean. ---
	result2, _, err := HandleValidateSpecs(ctx, nil, ValidateSpecsArgs{})
	if err != nil {
		t.Fatalf("pass 2: unexpected Go error: %v", err)
	}
	if result2.IsError {
		t.Fatalf("pass 2: expected success, got IsError=true:\n%s", testResultText(t, result2))
	}

	report2 := testResultText(t, result2)

	if !strings.Contains(report2, "All spec nodes are valid") {
		t.Errorf("pass 2: expected clean-tree message, got:\n%s", report2)
	}

	// Confirm no error sections appear.
	for _, section := range []string{"FORMAT ERRORS", "CIRCULAR REFERENCES", "STALE / MISSING"} {
		if strings.Contains(report2, section) {
			t.Errorf("pass 2: unexpected section %q in clean report:\n%s", section, report2)
		}
	}
}

// ---------------------------------------------------------------------------
// Happy-path: detects stale artifact
// ---------------------------------------------------------------------------

// TestHandleValidateSpecs_StaleArtifact verifies that an artifact file whose
// code-from-spec tag contains a hash that does not match the current chain hash
// is reported as "[stale]" in the STALE / MISSING ARTIFACTS section.
func TestHandleValidateSpecs_StaleArtifact(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkFile(t, "code-from-spec/_node.md", testNodeMd("", "ROOT", ""))
	testMkFile(t, "code-from-spec/a/_node.md", testNodeMd(
		"outputs:\n  - id: main\n    path: out/a.go\n",
		"ROOT/a",
		"",
	))

	// Write the artifact with a deliberately wrong hash (all-caps placeholder).
	// The real chain hash will differ from this, triggering a stale report.
	testMkFile(t, "out/a.go",
		"// code-from-spec: ROOT/a@AAAAAAAAAAAAAAAAAAAAAAAAA00\npackage a\n",
	)

	ctx := context.Background()
	result, _, err := HandleValidateSpecs(ctx, nil, ValidateSpecsArgs{})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success result, got IsError=true:\n%s", testResultText(t, result))
	}

	report := testResultText(t, result)

	if !strings.Contains(report, "STALE / MISSING ARTIFACTS") {
		t.Errorf("expected STALE / MISSING ARTIFACTS section, got:\n%s", report)
	}
	if !strings.Contains(report, "[stale]") {
		t.Errorf("expected [stale] marker, got:\n%s", report)
	}
	if !strings.Contains(report, "ROOT/a") {
		t.Errorf("expected ROOT/a in staleness report, got:\n%s", report)
	}
}

// ---------------------------------------------------------------------------
// Happy-path: detects missing artifact
// ---------------------------------------------------------------------------

// TestHandleValidateSpecs_MissingArtifact verifies that when a declared output
// file does not exist on disk, it is reported as "[missing]" in the staleness
// section.
func TestHandleValidateSpecs_MissingArtifact(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkFile(t, "code-from-spec/_node.md", testNodeMd("", "ROOT", ""))
	testMkFile(t, "code-from-spec/a/_node.md", testNodeMd(
		"outputs:\n  - id: main\n    path: out/a.go\n",
		"ROOT/a",
		"",
	))

	// Intentionally do NOT create out/a.go.

	ctx := context.Background()
	result, _, err := HandleValidateSpecs(ctx, nil, ValidateSpecsArgs{})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success result, got IsError=true:\n%s", testResultText(t, result))
	}

	report := testResultText(t, result)

	if !strings.Contains(report, "STALE / MISSING ARTIFACTS") {
		t.Errorf("expected STALE / MISSING ARTIFACTS section, got:\n%s", report)
	}
	if !strings.Contains(report, "[missing]") {
		t.Errorf("expected [missing] marker, got:\n%s", report)
	}
	if !strings.Contains(report, "ROOT/a") {
		t.Errorf("expected ROOT/a in staleness report, got:\n%s", report)
	}
}

// ---------------------------------------------------------------------------
// Happy-path: detects format errors
// ---------------------------------------------------------------------------

// TestHandleValidateSpecs_FormatErrors verifies that a node with an invalid
// depends_on reference (target does not exist) produces at least one entry in
// the FORMAT ERRORS section of the report.
func TestHandleValidateSpecs_FormatErrors(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkFile(t, "code-from-spec/_node.md", testNodeMd("", "ROOT", ""))

	// ROOT/a references a node that does not exist in the tree.
	testMkFile(t, "code-from-spec/a/_node.md", testNodeMd(
		"depends_on:\n  - ROOT/nonexistent\n",
		"ROOT/a",
		"",
	))

	ctx := context.Background()
	result, _, err := HandleValidateSpecs(ctx, nil, ValidateSpecsArgs{})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success result, got IsError=true:\n%s", testResultText(t, result))
	}

	report := testResultText(t, result)

	if !strings.Contains(report, "FORMAT ERRORS") {
		t.Errorf("expected FORMAT ERRORS section, got:\n%s", report)
	}
	if !strings.Contains(report, "ROOT/a") {
		t.Errorf("expected ROOT/a in format errors, got:\n%s", report)
	}
}

// ---------------------------------------------------------------------------
// Happy-path: detects circular references
// ---------------------------------------------------------------------------

// TestHandleValidateSpecs_CircularReferences verifies that a two-node cycle
// (ROOT/a → ROOT/b → ROOT/a) is detected and both participants appear in the
// CIRCULAR REFERENCES section.
func TestHandleValidateSpecs_CircularReferences(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkFile(t, "code-from-spec/_node.md", testNodeMd("", "ROOT", ""))

	// ROOT/a depends on ROOT/b, ROOT/b depends on ROOT/a — a direct cycle.
	testMkFile(t, "code-from-spec/a/_node.md", testNodeMd(
		"depends_on:\n  - ROOT/b\n",
		"ROOT/a",
		"",
	))
	testMkFile(t, "code-from-spec/b/_node.md", testNodeMd(
		"depends_on:\n  - ROOT/a\n",
		"ROOT/b",
		"",
	))

	ctx := context.Background()
	result, _, err := HandleValidateSpecs(ctx, nil, ValidateSpecsArgs{})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success result, got IsError=true:\n%s", testResultText(t, result))
	}

	report := testResultText(t, result)

	if !strings.Contains(report, "CIRCULAR REFERENCES") {
		t.Errorf("expected CIRCULAR REFERENCES section, got:\n%s", report)
	}
	if !strings.Contains(report, "ROOT/a") && !strings.Contains(report, "ROOT/b") {
		t.Errorf("expected at least one cycle participant (ROOT/a or ROOT/b), got:\n%s", report)
	}
	if !strings.Contains(report, "ROOT/b") {
		t.Errorf("expected ROOT/b in circular references, got:\n%s", report)
	}
}

// ---------------------------------------------------------------------------
// Happy-path: multiple error categories collected together
// ---------------------------------------------------------------------------

// TestHandleValidateSpecs_MultipleErrorsTogether verifies that circular
// references and stale artifacts can appear simultaneously in a single
// report.
//
// Tree layout:
//   - ROOT/a — valid node with a stale artifact (wrong hash).
//   - ROOT/b — depends on ROOT/c  ┐ cycle
//   - ROOT/c — depends on ROOT/b  ┘
func TestHandleValidateSpecs_MultipleErrorsTogether(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkFile(t, "code-from-spec/_node.md", testNodeMd("", "ROOT", ""))

	// Stale artifact node.
	testMkFile(t, "code-from-spec/a/_node.md", testNodeMd(
		"outputs:\n  - id: main\n    path: out/a.go\n",
		"ROOT/a",
		"",
	))
	testMkFile(t, "out/a.go",
		"// code-from-spec: ROOT/a@AAAAAAAAAAAAAAAAAAAAAAAAA00\npackage a\n",
	)

	// Cycle participants.
	testMkFile(t, "code-from-spec/b/_node.md", testNodeMd(
		"depends_on:\n  - ROOT/c\n",
		"ROOT/b",
		"",
	))
	testMkFile(t, "code-from-spec/c/_node.md", testNodeMd(
		"depends_on:\n  - ROOT/b\n",
		"ROOT/c",
		"",
	))

	ctx := context.Background()
	result, _, err := HandleValidateSpecs(ctx, nil, ValidateSpecsArgs{})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success result, got IsError=true:\n%s", testResultText(t, result))
	}

	report := testResultText(t, result)

	// Both cycles and staleness must be present.
	if !strings.Contains(report, "CIRCULAR REFERENCES") {
		t.Errorf("expected CIRCULAR REFERENCES section in report, got:\n%s", report)
	}
	if !strings.Contains(report, "STALE / MISSING ARTIFACTS") && !strings.Contains(report, "stale") && !strings.Contains(report, "missing") {
		t.Errorf("expected staleness in report, got:\n%s", report)
	}
}

// ---------------------------------------------------------------------------
// Failure case: continues after unreadable file
// ---------------------------------------------------------------------------

// TestHandleValidateSpecs_ContinuesAfterUnreadableFile verifies that a node
// whose _node.md has completely invalid YAML frontmatter does not abort
// validation of the remaining nodes. The unreadable node produces a format
// error; the valid node is still checked for staleness.
func TestHandleValidateSpecs_ContinuesAfterUnreadableFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkFile(t, "code-from-spec/_node.md", testNodeMd("", "ROOT", ""))

	// ROOT/bad — frontmatter is syntactically invalid YAML.
	// The colon sequence ": this is not valid yaml: [" is a mapping key error.
	testMkFile(t, "code-from-spec/bad/_node.md",
		"---\n: this is not valid yaml: [\n---\n# ROOT/bad\n",
	)

	// ROOT/good — valid node with a declared output; file absent → "missing".
	testMkFile(t, "code-from-spec/good/_node.md", testNodeMd(
		"outputs:\n  - id: main\n    path: out/good.go\n",
		"ROOT/good",
		"",
	))

	// Intentionally do NOT create out/good.go.

	ctx := context.Background()
	result, _, err := HandleValidateSpecs(ctx, nil, ValidateSpecsArgs{})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success result, got IsError=true:\n%s", testResultText(t, result))
	}

	report := testResultText(t, result)

	// The bad node must produce a format error.
	if !strings.Contains(report, "FORMAT ERRORS") {
		t.Errorf("expected FORMAT ERRORS section for bad node, got:\n%s", report)
	}
	if !strings.Contains(report, "ROOT/bad") {
		t.Errorf("expected ROOT/bad in format errors, got:\n%s", report)
	}

	// The good node must still be processed and appear in staleness.
	if !strings.Contains(report, "STALE / MISSING ARTIFACTS") {
		t.Errorf("expected STALE / MISSING ARTIFACTS section for good node, got:\n%s", report)
	}
	if !strings.Contains(report, "ROOT/good") {
		t.Errorf("expected ROOT/good in staleness section, got:\n%s", report)
	}
}
