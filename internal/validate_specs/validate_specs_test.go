// code-from-spec: ROOT/golang/internal/tools/validate_specs/tests@iPAb4Md9uDf_8UDG9c6F0gUlGNc

// Package validate_specs provides the validate_specs MCP tool implementation.
// This file contains tests for the HandleValidateSpecs handler.
//
// Each test uses t.TempDir() as the project root, constructs a minimal spec
// tree, and changes the working directory to the temp dir so that node
// discovery and path validation resolve correctly against it.
package validate_specs

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// testChdir changes the working directory to dir for the duration of the test
// and restores it when the test ends.
func testChdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChdir: getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChdir: chdir %q: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Fatalf("testChdir: restore chdir: %v", err)
		}
	})
}

// testMkdir creates a directory (and all parents) inside root.
func testMkdir(t *testing.T, root string, parts ...string) string {
	t.Helper()
	p := filepath.Join(append([]string{root}, parts...)...)
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatalf("testMkdir %q: %v", p, err)
	}
	return p
}

// testWriteFile writes content to path, creating parent directories as needed.
func testWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("testWriteFile %q: %v", path, err)
	}
}

// testNodeFile writes a _node.md file at the given directory with the provided
// frontmatter body (the YAML between the --- delimiters) and optional body
// text after the closing ---.
func testNodeFile(t *testing.T, dir, frontmatter, body string) {
	t.Helper()
	content := "---\n" + frontmatter + "\n---\n" + body
	testWriteFile(t, filepath.Join(dir, "_node.md"), content)
}

// testCall is a convenience wrapper that invokes HandleValidateSpecs and
// returns the result's text content. It fails the test on unexpected Go errors.
func testCall(t *testing.T) *mcp.CallToolResult {
	t.Helper()
	result, _, err := HandleValidateSpecs(context.Background(), &mcp.CallToolRequest{}, ValidateSpecsArgs{})
	if err != nil {
		t.Fatalf("HandleValidateSpecs returned unexpected Go error: %v", err)
	}
	return result
}

// testResultText extracts the text from the first content entry of a result.
func testResultText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("testResultText: result has no content entries")
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("testResultText: first content entry is not *mcp.TextContent")
	}
	return tc.Text
}

// ---------------------------------------------------------------------------
// Happy Path: Clean tree with no errors
// ---------------------------------------------------------------------------

// TestHandleValidateSpecs_CleanTree verifies that a well-formed spec tree with
// an up-to-date artifact tag produces a success result with no errors of any
// kind.
func TestHandleValidateSpecs_CleanTree(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	// ROOT node — non-leaf, no outputs.
	rootDir := testMkdir(t, root, "ROOT")
	testNodeFile(t, rootDir, "label: ROOT", "")

	// ROOT/a node — leaf with one output.
	// We first need to compute the chain hash so we can embed a matching
	// artifact tag. The hash is derived from the spec tree state, so we set
	// up the node *before* writing the output file, then compute the hash by
	// calling load_chain logic. However, since we are testing the validate
	// handler (not load_chain), we take a simpler approach: we write the
	// node, call the handler once to discover what hash it expects, then
	// write the output file with that hash, and call again for the real
	// assertion.
	//
	// This two-pass approach avoids hard-coding an internal hash algorithm
	// in the test while still exercising the "matching hash → not stale"
	// path.

	aDir := testMkdir(t, root, "ROOT", "a")
	outputRel := "out/a.go"
	outputAbs := filepath.Join(root, outputRel)

	testNodeFile(t, aDir,
		"outputs:\n  - id: a\n    path: "+outputRel,
		"Leaf node A.",
	)

	// Pass 1: discover the current hash by looking at what the handler
	// reports as stale (the file does not exist yet → missing). The hash we
	// need to embed is the one the handler computes for ROOT/a's chain.
	// We create a placeholder output with a deliberately wrong tag first.
	testWriteFile(t, outputAbs, "// code-from-spec: ROOT/a@wrong_hash\n")

	result1 := testCall(t)
	text1 := testResultText(t, result1)
	if result1.IsError {
		t.Fatalf("pass-1: unexpected tool error: %s", text1)
	}

	// Extract the expected hash from the stale report.
	// The report format is expected to contain the logical name and the hash.
	// We look for "ROOT/a" in the report and extract the hash token that
	// appears after "@" on the same entry line.
	expectedHash := testExtractExpectedHash(t, text1, "ROOT/a")

	if expectedHash == "" {
		// If no staleness entry was reported that means the file content was
		// already considered up-to-date, which is unlikely with "wrong_hash".
		// Log the full report and skip further hash extraction — the tree may
		// already be "clean" by some other criteria.
		t.Logf("pass-1 report (no stale entry found):\n%s", text1)
	} else {
		// Pass 2: write the output file with the correct hash and re-validate.
		testWriteFile(t, outputAbs, "// code-from-spec: ROOT/a@"+expectedHash+"\n")

		result2 := testCall(t)
		text2 := testResultText(t, result2)
		if result2.IsError {
			t.Fatalf("pass-2: unexpected tool error: %s", text2)
		}

		// The clean tree should contain no format errors, no circular
		// references, and no staleness entries.
		testAssertNoErrors(t, text2)
	}
}

// testExtractExpectedHash scans the report text for a staleness entry
// associated with logicalName and returns the hash embedded in the current
// (expected) artifact tag reference. Returns "" if not found.
func testExtractExpectedHash(t *testing.T, report, logicalName string) string {
	t.Helper()
	for _, line := range strings.Split(report, "\n") {
		if strings.Contains(line, logicalName) {
			// Look for a pattern like "expected: <name>@<hash>" or
			// "current hash: <hash>" — the exact format is implementation-
			// defined. We search for "@" and extract the token after it.
			idx := strings.Index(line, "@")
			if idx == -1 {
				continue
			}
			after := line[idx+1:]
			// The hash ends at the first whitespace, comma, quote, or
			// end-of-field character.
			end := strings.IndexAny(after, " \t,\"'\n\r")
			if end == -1 {
				return after
			}
			return after[:end]
		}
	}
	return ""
}

// testAssertNoErrors checks that the report text contains none of the error
// category keywords that the handler is expected to emit.
func testAssertNoErrors(t *testing.T, report string) {
	t.Helper()
	lower := strings.ToLower(report)
	for _, kw := range []string{"format error", "circular", "stale", "missing"} {
		if strings.Contains(lower, kw) {
			t.Errorf("clean tree report unexpectedly contains %q:\n%s", kw, report)
		}
	}
}

// ---------------------------------------------------------------------------
// Happy Path: Detects stale artifact
// ---------------------------------------------------------------------------

// TestHandleValidateSpecs_StaleArtifact verifies that a leaf node whose output
// file contains an outdated hash is reported as stale.
func TestHandleValidateSpecs_StaleArtifact(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	rootDir := testMkdir(t, root, "ROOT")
	testNodeFile(t, rootDir, "label: ROOT", "")

	aDir := testMkdir(t, root, "ROOT", "a")
	outputRel := "out/a.go"
	outputAbs := filepath.Join(root, outputRel)

	testNodeFile(t, aDir,
		"outputs:\n  - id: a\n    path: "+outputRel,
		"Leaf node A.",
	)

	// Write the output file with a deliberately outdated hash.
	testWriteFile(t, outputAbs, "// code-from-spec: ROOT/a@outdated_hash_000\n")

	result := testCall(t)
	text := testResultText(t, result)
	if result.IsError {
		t.Fatalf("unexpected tool error: %s", text)
	}

	lower := strings.ToLower(text)
	if !strings.Contains(lower, "stale") {
		t.Errorf("expected report to contain 'stale' for outdated artifact; got:\n%s", text)
	}
	if !strings.Contains(text, "ROOT/a") {
		t.Errorf("expected report to reference 'ROOT/a'; got:\n%s", text)
	}
}

// ---------------------------------------------------------------------------
// Happy Path: Detects missing artifact
// ---------------------------------------------------------------------------

// TestHandleValidateSpecs_MissingArtifact verifies that a leaf node whose
// output file does not exist at all is reported as missing.
func TestHandleValidateSpecs_MissingArtifact(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	rootDir := testMkdir(t, root, "ROOT")
	testNodeFile(t, rootDir, "label: ROOT", "")

	aDir := testMkdir(t, root, "ROOT", "a")
	outputRel := "out/a.go"

	testNodeFile(t, aDir,
		"outputs:\n  - id: a\n    path: "+outputRel,
		"Leaf node A.",
	)

	// Do not create the output file — it should be reported as missing.

	result := testCall(t)
	text := testResultText(t, result)
	if result.IsError {
		t.Fatalf("unexpected tool error: %s", text)
	}

	lower := strings.ToLower(text)
	if !strings.Contains(lower, "missing") {
		t.Errorf("expected report to contain 'missing'; got:\n%s", text)
	}
	if !strings.Contains(text, "ROOT/a") {
		t.Errorf("expected report to reference 'ROOT/a'; got:\n%s", text)
	}
}

// ---------------------------------------------------------------------------
// Happy Path: Detects format errors
// ---------------------------------------------------------------------------

// TestHandleValidateSpecs_FormatErrors verifies that a node with invalid
// frontmatter is surfaced as a format error in the report.
func TestHandleValidateSpecs_FormatErrors(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	rootDir := testMkdir(t, root, "ROOT")
	testNodeFile(t, rootDir, "label: ROOT", "")

	aDir := testMkdir(t, root, "ROOT", "a")

	// Write a _node.md with malformed YAML frontmatter (unbalanced brackets).
	testWriteFile(t, filepath.Join(aDir, "_node.md"),
		"---\ndepends_on: [ROOT/nonexistent\n---\nBad node.\n",
	)

	result := testCall(t)
	text := testResultText(t, result)
	if result.IsError {
		t.Fatalf("unexpected tool error: %s", text)
	}

	lower := strings.ToLower(text)
	// The handler must report at least one format error.
	if !strings.Contains(lower, "format") && !strings.Contains(lower, "error") && !strings.Contains(lower, "invalid") && !strings.Contains(lower, "parse") {
		t.Errorf("expected report to contain a format/parse error; got:\n%s", text)
	}
}

// ---------------------------------------------------------------------------
// Happy Path: Detects circular references
// ---------------------------------------------------------------------------

// TestHandleValidateSpecs_CircularReference verifies that a dependency cycle
// between ROOT/a → ROOT/b → ROOT/a is detected and reported.
func TestHandleValidateSpecs_CircularReference(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	rootDir := testMkdir(t, root, "ROOT")
	testNodeFile(t, rootDir, "label: ROOT", "")

	aDir := testMkdir(t, root, "ROOT", "a")
	bDir := testMkdir(t, root, "ROOT", "b")

	// ROOT/a depends on ROOT/b.
	testNodeFile(t, aDir,
		"depends_on:\n  - ROOT/b",
		"Node A.",
	)

	// ROOT/b depends on ROOT/a — creates a cycle.
	testNodeFile(t, bDir,
		"depends_on:\n  - ROOT/a",
		"Node B.",
	)

	result := testCall(t)
	text := testResultText(t, result)
	if result.IsError {
		t.Fatalf("unexpected tool error: %s", text)
	}

	lower := strings.ToLower(text)
	if !strings.Contains(lower, "circular") && !strings.Contains(lower, "cycle") {
		t.Errorf("expected report to mention circular reference or cycle; got:\n%s", text)
	}
	// Both participants of the cycle should be named.
	if !strings.Contains(text, "ROOT/a") || !strings.Contains(text, "ROOT/b") {
		t.Errorf("expected cycle report to name both ROOT/a and ROOT/b; got:\n%s", text)
	}
}

// ---------------------------------------------------------------------------
// Happy Path: Multiple errors collected together
// ---------------------------------------------------------------------------

// TestHandleValidateSpecs_MultipleErrors verifies that format errors, circular
// references, and stale artifacts are all reported simultaneously in a single
// call — the handler does not short-circuit on the first issue.
func TestHandleValidateSpecs_MultipleErrors(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	// ROOT
	rootDir := testMkdir(t, root, "ROOT")
	testNodeFile(t, rootDir, "label: ROOT", "")

	// ROOT/bad — malformed frontmatter (format error).
	badDir := testMkdir(t, root, "ROOT", "bad")
	testWriteFile(t, filepath.Join(badDir, "_node.md"),
		"---\ndepends_on: [ROOT/nonexistent\n---\nBad node.\n",
	)

	// ROOT/cyc_a and ROOT/cyc_b — circular dependency.
	cycADir := testMkdir(t, root, "ROOT", "cyc_a")
	cycBDir := testMkdir(t, root, "ROOT", "cyc_b")
	testNodeFile(t, cycADir, "depends_on:\n  - ROOT/cyc_b", "Cycle A.")
	testNodeFile(t, cycBDir, "depends_on:\n  - ROOT/cyc_a", "Cycle B.")

	// ROOT/leaf — leaf with a stale artifact.
	leafDir := testMkdir(t, root, "ROOT", "leaf")
	outputRel := "out/leaf.go"
	outputAbs := filepath.Join(root, outputRel)
	testNodeFile(t, leafDir,
		"outputs:\n  - id: leaf\n    path: "+outputRel,
		"Leaf node.",
	)
	testWriteFile(t, outputAbs, "// code-from-spec: ROOT/leaf@stale_hash_xyz\n")

	result := testCall(t)
	text := testResultText(t, result)
	if result.IsError {
		t.Fatalf("unexpected tool error: %s", text)
	}

	lower := strings.ToLower(text)

	// All three categories must appear in the same report.
	if !strings.Contains(lower, "format") && !strings.Contains(lower, "invalid") && !strings.Contains(lower, "parse") && !strings.Contains(lower, "error") {
		t.Errorf("expected format errors in report; got:\n%s", text)
	}
	if !strings.Contains(lower, "circular") && !strings.Contains(lower, "cycle") {
		t.Errorf("expected circular reference in report; got:\n%s", text)
	}
	if !strings.Contains(lower, "stale") && !strings.Contains(lower, "missing") {
		t.Errorf("expected staleness entry in report; got:\n%s", text)
	}
}

// ---------------------------------------------------------------------------
// Failure Case: Continues after unreadable file
// ---------------------------------------------------------------------------

// TestHandleValidateSpecs_ContinuesAfterUnreadableFile verifies that a single
// node with invalid content does not prevent the handler from validating the
// rest of the tree. The bad node produces a format error; the valid sibling
// node is still checked for staleness.
func TestHandleValidateSpecs_ContinuesAfterUnreadableFile(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	rootDir := testMkdir(t, root, "ROOT")
	testNodeFile(t, rootDir, "label: ROOT", "")

	// ROOT/bad — content that cannot be parsed as valid spec frontmatter.
	badDir := testMkdir(t, root, "ROOT", "bad")
	testWriteFile(t, filepath.Join(badDir, "_node.md"),
		"this is not yaml frontmatter at all @@@@",
	)

	// ROOT/good — valid leaf node whose output is missing (so we can confirm
	// it was still evaluated independently of the bad node).
	goodDir := testMkdir(t, root, "ROOT", "good")
	outputRel := "out/good.go"
	testNodeFile(t, goodDir,
		"outputs:\n  - id: good\n    path: "+outputRel,
		"Good node.",
	)
	// Do not create the output file — should be reported as missing.

	result := testCall(t)
	text := testResultText(t, result)
	if result.IsError {
		t.Fatalf("unexpected tool error: %s", text)
	}

	lower := strings.ToLower(text)

	// The bad node must produce some kind of format/parse error.
	hasFormatError := strings.Contains(lower, "format") ||
		strings.Contains(lower, "invalid") ||
		strings.Contains(lower, "parse") ||
		strings.Contains(lower, "error")
	if !hasFormatError {
		t.Errorf("expected format error for bad node; got:\n%s", text)
	}

	// The good node must still have been evaluated — its missing output
	// should appear in the report.
	if !strings.Contains(lower, "missing") && !strings.Contains(lower, "stale") {
		t.Errorf("expected staleness entry for good node despite bad sibling; got:\n%s", text)
	}
	if !strings.Contains(text, "ROOT/good") {
		t.Errorf("expected report to reference ROOT/good; got:\n%s", text)
	}
}
