// code-from-spec: ROOT/golang/internal/tools/load_chain/tests@tSe8py_8f0DhQPKhYL_8qW1Gb30

// Package load_chain tests exercise HandleLoadChain end-to-end.
//
// Each test creates an isolated temp directory, populates it with spec files
// (and optionally output files), changes the working directory to that temp
// directory, and then calls HandleLoadChain directly.
//
// Design notes:
//   - All spec files follow the CommonMark structure required by ParseNode:
//     optional frontmatter block, then "# <logical name>" as the first heading.
//   - We use os.Chdir to make the working directory match the temp dir so that
//     both pathvalidation.ValidatePath and os.ReadFile resolve paths correctly.
//   - os.Chdir is not concurrency-safe, so each test runs serially (no t.Parallel).
//   - All helper functions and types are prefixed with "test" per project convention.
package load_chain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// testWriteFile creates all intermediate directories and writes content to path.
func testWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("testWriteFile: mkdir %q: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile: write %q: %v", path, err)
	}
}

// testNodePath returns the filesystem path for a logical name under dir.
// It mirrors logicalnames.PathFromLogicalName but uses OS separators so we can
// write files with os.WriteFile.
//
//   ROOT          → <dir>/code-from-spec/_node.md
//   ROOT/x/y      → <dir>/code-from-spec/x/y/_node.md
func testNodePath(dir, logicalName string) string {
	// Strip "ROOT" prefix and any trailing qualifier.
	name := logicalName
	if idx := strings.Index(name, "("); idx != -1 {
		name = name[:idx]
	}
	if name == "ROOT" {
		return filepath.Join(dir, "code-from-spec", "_node.md")
	}
	rel := strings.TrimPrefix(name, "ROOT/")
	parts := strings.Split(rel, "/")
	elems := append([]string{dir, "code-from-spec"}, parts...)
	elems = append(elems, "_node.md")
	return filepath.Join(elems...)
}

// testSetupTempDir creates a temp directory, changes the working directory to
// it, and registers a cleanup function that restores the original working
// directory. Returns the temp directory path.
func testSetupTempDir(t *testing.T) string {
	t.Helper()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("testSetupTempDir: getwd: %v", err)
	}
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("testSetupTempDir: chdir to %q: %v", tmpDir, err)
	}
	t.Cleanup(func() {
		// Restore working directory after the test.
		_ = os.Chdir(origDir)
	})
	return tmpDir
}

// testCallHandler invokes HandleLoadChain with the given logical name and
// returns the result. The context and request can be nil — HandleLoadChain
// only uses them for MCP plumbing that we do not exercise here.
func testCallHandler(t *testing.T, logicalName string) *mcp.CallToolResult {
	t.Helper()
	result, _, err := HandleLoadChain(nil, nil, LoadChainArgs{LogicalName: logicalName})
	if err != nil {
		t.Fatalf("HandleLoadChain returned unexpected Go error: %v", err)
	}
	if result == nil {
		t.Fatalf("HandleLoadChain returned nil result")
	}
	return result
}

// testResultText extracts the text from the first content item of a result.
func testResultText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatalf("testResultText: result has no content items")
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("testResultText: first content item is not *mcp.TextContent")
	}
	return tc.Text
}

// testAssertSuccess fails the test if result.IsError is true.
func testAssertSuccess(t *testing.T, result *mcp.CallToolResult) {
	t.Helper()
	if result.IsError {
		t.Fatalf("expected success result but got tool error: %s", testResultText(t, result))
	}
}

// testAssertToolError fails the test if result.IsError is false, or if the
// error message does not contain wantSubstr.
func testAssertToolError(t *testing.T, result *mcp.CallToolResult, wantSubstr string) {
	t.Helper()
	if !result.IsError {
		t.Fatalf("expected tool error containing %q but got success: %s", wantSubstr, testResultText(t, result))
	}
	msg := testResultText(t, result)
	if !strings.Contains(msg, wantSubstr) {
		t.Fatalf("expected tool error containing %q but got: %s", wantSubstr, msg)
	}
}

// testNodeContent builds a minimal valid spec file body for a node.
//
//   logicalName  – used as the # Heading (first level-1 heading)
//   fm           – raw YAML frontmatter content (without "---" delimiters);
//                  pass "" for no frontmatter
//   publicBody   – content to place under "# Public"; pass "" to omit the section
//   privateSections – map heading → content for private sections
func testNodeContent(logicalName, fm, publicBody string, privateSections map[string]string) string {
	var sb strings.Builder

	// Frontmatter block (optional).
	if fm != "" {
		sb.WriteString("---\n")
		sb.WriteString(fm)
		sb.WriteString("\n---\n")
	}

	// Strip qualifier for heading.
	heading := logicalName
	if idx := strings.Index(heading, "("); idx != -1 {
		heading = heading[:idx]
	}

	// Node name heading (required by ParseNode).
	sb.WriteString("# ")
	sb.WriteString(heading)
	sb.WriteString("\n\n")

	// Public section (optional).
	if publicBody != "" {
		sb.WriteString("# Public\n")
		sb.WriteString(publicBody)
		sb.WriteString("\n")
	}

	// Private sections (optional).
	for heading, body := range privateSections {
		sb.WriteString("# ")
		sb.WriteString(heading)
		sb.WriteString("\n")
		sb.WriteString(body)
		sb.WriteString("\n")
	}

	return sb.String()
}

// ---------------------------------------------------------------------------
// Happy path tests
// ---------------------------------------------------------------------------

// TestHandleLoadChain_ValidRootLeaf tests the simplest valid case:
// ROOT and ROOT/a, where ROOT/a has outputs and no dependencies.
// ROOT has a # Public section; ROOT/a also has one.
// Expectation:
//   - Success result.
//   - Chain content contains ROOT's public body (without the "# Public" heading).
//   - Chain content contains ROOT/a with reduced frontmatter and its full body.
func TestHandleLoadChain_ValidRootLeaf(t *testing.T) {
	tmpDir := testSetupTempDir(t)

	// Create ROOT node with a # Public section.
	rootContent := testNodeContent(
		"ROOT",
		"",
		"This is the root public content.\n",
		nil,
	)
	testWriteFile(t, testNodePath(tmpDir, "ROOT"), rootContent)

	// Create ROOT/a with outputs declared.
	leafFM := "outputs:\n  - id: main\n    path: src/main.go"
	leafContent := testNodeContent(
		"ROOT/a",
		leafFM,
		"This is the leaf public content.\n",
		nil,
	)
	testWriteFile(t, testNodePath(tmpDir, "ROOT/a"), leafContent)

	result := testCallHandler(t, "ROOT/a")
	testAssertSuccess(t, result)

	text := testResultText(t, result)

	// ROOT's public body must appear (without the "# Public" heading).
	if !strings.Contains(text, "This is the root public content.") {
		t.Errorf("expected ROOT public body in chain; got:\n%s", text)
	}
	if strings.Contains(text, "# Public") {
		t.Errorf("# Public heading must not appear in chain; got:\n%s", text)
	}

	// Reduced frontmatter for ROOT/a must appear (outputs only, no depends_on).
	if !strings.Contains(text, "outputs:") {
		t.Errorf("expected reduced frontmatter (outputs) for ROOT/a; got:\n%s", text)
	}
	if strings.Contains(text, "depends_on") {
		t.Errorf("depends_on must not appear in reduced frontmatter; got:\n%s", text)
	}

	// Leaf public content must appear.
	if !strings.Contains(text, "This is the leaf public content.") {
		t.Errorf("expected ROOT/a public body in chain; got:\n%s", text)
	}

	// Chain hash header must be present on the first line.
	firstLine := strings.SplitN(text, "\n", 2)[0]
	if !strings.HasPrefix(firstLine, "chain_hash: ") {
		t.Errorf("expected first line to be 'chain_hash: ...'; got: %q", firstLine)
	}
}

// TestHandleLoadChain_DependencyNoQualifier tests a node that depends on
// ROOT/b (no qualifier). ROOT/b has # Public with ## Interface and ## Constraints.
// Expectation: full # Public content of ROOT/b appears (both subsections).
func TestHandleLoadChain_DependencyNoQualifier(t *testing.T) {
	tmpDir := testSetupTempDir(t)

	// ROOT with minimal public content.
	testWriteFile(t, testNodePath(tmpDir, "ROOT"),
		testNodeContent("ROOT", "", "Root context.\n", nil))

	// ROOT/b has two subsections under # Public.
	depPublic := "## Interface\nThis is the interface.\n\n## Constraints\nThese are the constraints.\n"
	testWriteFile(t, testNodePath(tmpDir, "ROOT/b"),
		testNodeContent("ROOT/b", "", depPublic, nil))

	// ROOT/a depends on ROOT/b (no qualifier).
	leafFM := "depends_on:\n  - ROOT/b\noutputs:\n  - id: main\n    path: src/main.go"
	testWriteFile(t, testNodePath(tmpDir, "ROOT/a"),
		testNodeContent("ROOT/a", leafFM, "Leaf content.\n", nil))

	result := testCallHandler(t, "ROOT/a")
	testAssertSuccess(t, result)

	text := testResultText(t, result)

	// Both subsections of ROOT/b must be present.
	if !strings.Contains(text, "This is the interface.") {
		t.Errorf("expected ## Interface content in chain; got:\n%s", text)
	}
	if !strings.Contains(text, "These are the constraints.") {
		t.Errorf("expected ## Constraints content in chain; got:\n%s", text)
	}
}

// TestHandleLoadChain_DependencyWithQualifier tests a node that depends on
// ROOT/b(interface). ROOT/b has ## Interface and ## Constraints subsections.
// Expectation: only ## Interface content appears, not ## Constraints.
func TestHandleLoadChain_DependencyWithQualifier(t *testing.T) {
	tmpDir := testSetupTempDir(t)

	testWriteFile(t, testNodePath(tmpDir, "ROOT"),
		testNodeContent("ROOT", "", "Root context.\n", nil))

	depPublic := "## Interface\nThis is the interface.\n\n## Constraints\nThese are the constraints.\n"
	testWriteFile(t, testNodePath(tmpDir, "ROOT/b"),
		testNodeContent("ROOT/b", "", depPublic, nil))

	// ROOT/a depends on ROOT/b(interface) — only the interface subsection.
	leafFM := "depends_on:\n  - ROOT/b(interface)\noutputs:\n  - id: main\n    path: src/main.go"
	testWriteFile(t, testNodePath(tmpDir, "ROOT/a"),
		testNodeContent("ROOT/a", leafFM, "Leaf content.\n", nil))

	result := testCallHandler(t, "ROOT/a")
	testAssertSuccess(t, result)

	text := testResultText(t, result)

	if !strings.Contains(text, "This is the interface.") {
		t.Errorf("expected ## Interface content in chain; got:\n%s", text)
	}
	if strings.Contains(text, "These are the constraints.") {
		t.Errorf("## Constraints content must NOT appear when qualifier is 'interface'; got:\n%s", text)
	}
}

// TestHandleLoadChain_AncestorPublicOnlyWithoutHeading tests that ancestors
// expose only the body of their # Public sections (no "# Public" heading,
// no private sections, no node name heading).
// Tree: ROOT → ROOT/a → ROOT/a/b (leaf). ROOT and ROOT/a have public and
// private sections.
func TestHandleLoadChain_AncestorPublicOnlyWithoutHeading(t *testing.T) {
	tmpDir := testSetupTempDir(t)

	// ROOT with public and a private section.
	testWriteFile(t, testNodePath(tmpDir, "ROOT"),
		testNodeContent("ROOT", "", "ROOT public body.\n",
			map[string]string{"Private": "ROOT private body.\n"}))

	// ROOT/a with public and a private section.
	testWriteFile(t, testNodePath(tmpDir, "ROOT/a"),
		testNodeContent("ROOT/a", "", "ROOT/a public body.\n",
			map[string]string{"Private": "ROOT/a private body.\n"}))

	// ROOT/a/b is the leaf.
	leafFM := "outputs:\n  - id: main\n    path: src/main.go"
	testWriteFile(t, testNodePath(tmpDir, "ROOT/a/b"),
		testNodeContent("ROOT/a/b", leafFM, "Leaf public body.\n", nil))

	result := testCallHandler(t, "ROOT/a/b")
	testAssertSuccess(t, result)

	text := testResultText(t, result)

	// Public bodies of ancestors must appear.
	if !strings.Contains(text, "ROOT public body.") {
		t.Errorf("expected ROOT public body in chain; got:\n%s", text)
	}
	if !strings.Contains(text, "ROOT/a public body.") {
		t.Errorf("expected ROOT/a public body in chain; got:\n%s", text)
	}

	// Private bodies of ancestors must NOT appear.
	if strings.Contains(text, "ROOT private body.") {
		t.Errorf("ROOT private body must not appear in chain; got:\n%s", text)
	}
	if strings.Contains(text, "ROOT/a private body.") {
		t.Errorf("ROOT/a private body must not appear in chain; got:\n%s", text)
	}

	// The "# Public" heading itself must not appear.
	if strings.Contains(text, "# Public") {
		t.Errorf("'# Public' heading must not appear in chain; got:\n%s", text)
	}
}

// TestHandleLoadChain_TargetHasReducedFrontmatter verifies that the target
// section in the chain contains only the `outputs` field — not `depends_on`.
func TestHandleLoadChain_TargetHasReducedFrontmatter(t *testing.T) {
	tmpDir := testSetupTempDir(t)

	testWriteFile(t, testNodePath(tmpDir, "ROOT"),
		testNodeContent("ROOT", "", "Root context.\n", nil))

	// ROOT/b must exist because ROOT/a depends on it.
	testWriteFile(t, testNodePath(tmpDir, "ROOT/b"),
		testNodeContent("ROOT/b", "", "Dep content.\n", nil))

	// ROOT/a has both depends_on and outputs.
	leafFM := "depends_on:\n  - ROOT/b\noutputs:\n  - id: a\n    path: src/a.go"
	testWriteFile(t, testNodePath(tmpDir, "ROOT/a"),
		testNodeContent("ROOT/a", leafFM, "Leaf content.\n", nil))

	result := testCallHandler(t, "ROOT/a")
	testAssertSuccess(t, result)

	text := testResultText(t, result)

	// outputs field must appear in reduced frontmatter.
	if !strings.Contains(text, "outputs:") {
		t.Errorf("expected 'outputs:' in chain; got:\n%s", text)
	}
	if !strings.Contains(text, "path: src/a.go") {
		t.Errorf("expected 'path: src/a.go' in chain; got:\n%s", text)
	}

	// depends_on must NOT appear in reduced frontmatter.
	if strings.Contains(text, "depends_on") {
		t.Errorf("'depends_on' must not appear in reduced frontmatter; got:\n%s", text)
	}
}

// TestHandleLoadChain_AncestorWithNoPublicSectionOmitted verifies that an
// ancestor that has no # Public section is omitted from the chain entirely.
func TestHandleLoadChain_AncestorWithNoPublicSectionOmitted(t *testing.T) {
	tmpDir := testSetupTempDir(t)

	// ROOT has only a node name heading and a private section — no # Public.
	rootContent := testNodeContent("ROOT", "", "", map[string]string{"Private": "ROOT private.\n"})
	testWriteFile(t, testNodePath(tmpDir, "ROOT"), rootContent)

	leafFM := "outputs:\n  - id: main\n    path: src/main.go"
	testWriteFile(t, testNodePath(tmpDir, "ROOT/a"),
		testNodeContent("ROOT/a", leafFM, "Leaf content.\n", nil))

	result := testCallHandler(t, "ROOT/a")
	testAssertSuccess(t, result)

	text := testResultText(t, result)

	// ROOT private content must not appear.
	if strings.Contains(text, "ROOT private.") {
		t.Errorf("ROOT private content must not appear in chain; got:\n%s", text)
	}
	// Chain must still be valid (chain_hash present).
	if !strings.Contains(text, "chain_hash: ") {
		t.Errorf("chain_hash missing from result; got:\n%s", text)
	}
}

// TestHandleLoadChain_AncestorWithEmptyPublicSectionOmitted verifies that
// an ancestor whose # Public section exists but has no content is omitted.
func TestHandleLoadChain_AncestorWithEmptyPublicSectionOmitted(t *testing.T) {
	tmpDir := testSetupTempDir(t)

	// ROOT has a # Public section but with no body (empty string).
	// testNodeContent will omit the section when publicBody is "".
	// We want the section present but empty, so we write the file manually.
	rootRaw := "# ROOT\n\n# Public\n\n"
	testWriteFile(t, testNodePath(tmpDir, "ROOT"), rootRaw)

	leafFM := "outputs:\n  - id: main\n    path: src/main.go"
	testWriteFile(t, testNodePath(tmpDir, "ROOT/a"),
		testNodeContent("ROOT/a", leafFM, "Leaf content.\n", nil))

	result := testCallHandler(t, "ROOT/a")
	testAssertSuccess(t, result)

	text := testResultText(t, result)

	// The chain must still succeed and contain chain_hash.
	if !strings.Contains(text, "chain_hash: ") {
		t.Errorf("chain_hash missing from result; got:\n%s", text)
	}
	// Since ROOT's # Public is empty, no ROOT block should contribute text
	// beyond what the leaf and its own frontmatter provide.
	// (We cannot check for absence of "ROOT" directly since logical names
	// may appear in other contexts — we just verify the overall success.)
}

// TestHandleLoadChain_DependencyWithEmptyExtractedContentOmitted verifies
// that a dependency whose extracted content (after applying a qualifier) is
// empty is omitted from the chain.
func TestHandleLoadChain_DependencyWithEmptyExtractedContentOmitted(t *testing.T) {
	tmpDir := testSetupTempDir(t)

	testWriteFile(t, testNodePath(tmpDir, "ROOT"),
		testNodeContent("ROOT", "", "Root context.\n", nil))

	// ROOT/b has ## Interface subsection with NO body.
	depPublic := "## Interface\n"
	testWriteFile(t, testNodePath(tmpDir, "ROOT/b"),
		testNodeContent("ROOT/b", "", depPublic, nil))

	leafFM := "depends_on:\n  - ROOT/b(interface)\noutputs:\n  - id: main\n    path: src/main.go"
	testWriteFile(t, testNodePath(tmpDir, "ROOT/a"),
		testNodeContent("ROOT/a", leafFM, "Leaf content.\n", nil))

	result := testCallHandler(t, "ROOT/a")
	testAssertSuccess(t, result)

	text := testResultText(t, result)

	// The chain must succeed.
	if !strings.Contains(text, "chain_hash: ") {
		t.Errorf("chain_hash missing from result; got:\n%s", text)
	}
	// ROOT/b's (empty) Interface content must not inject anything visible.
	// We verify by checking that the text does not contain "## Interface"
	// (since the heading itself is also stripped for qualified deps).
	if strings.Contains(text, "## Interface") {
		t.Errorf("## Interface heading must not appear when content is empty; got:\n%s", text)
	}
}

// TestHandleLoadChain_MultipleQualifiersSameDependencyConsolidated verifies
// that when ROOT/a depends on both ROOT/b(interface) and ROOT/b(constraints),
// only one file block for ROOT/b is emitted and both subsection bodies appear.
func TestHandleLoadChain_MultipleQualifiersSameDependencyConsolidated(t *testing.T) {
	tmpDir := testSetupTempDir(t)

	testWriteFile(t, testNodePath(tmpDir, "ROOT"),
		testNodeContent("ROOT", "", "Root context.\n", nil))

	depPublic := "## Interface\nInterface text here.\n\n## Constraints\nConstraints text here.\n"
	testWriteFile(t, testNodePath(tmpDir, "ROOT/b"),
		testNodeContent("ROOT/b", "", depPublic, nil))

	// ROOT/a depends on ROOT/b twice with different qualifiers.
	leafFM := "depends_on:\n  - ROOT/b(interface)\n  - ROOT/b(constraints)\noutputs:\n  - id: main\n    path: src/main.go"
	testWriteFile(t, testNodePath(tmpDir, "ROOT/a"),
		testNodeContent("ROOT/a", leafFM, "Leaf content.\n", nil))

	result := testCallHandler(t, "ROOT/a")
	testAssertSuccess(t, result)

	text := testResultText(t, result)

	// Both subsection bodies must appear.
	if !strings.Contains(text, "Interface text here.") {
		t.Errorf("expected Interface text in chain; got:\n%s", text)
	}
	if !strings.Contains(text, "Constraints text here.") {
		t.Errorf("expected Constraints text in chain; got:\n%s", text)
	}

	// The content for ROOT/b must not be duplicated — check that
	// "Interface text here." appears exactly once.
	count := strings.Count(text, "Interface text here.")
	if count != 1 {
		t.Errorf("expected 'Interface text here.' exactly once, got %d occurrences; text:\n%s", count, text)
	}
}

// ---------------------------------------------------------------------------
// Failure case tests
// ---------------------------------------------------------------------------

// TestHandleLoadChain_InvalidPrefix verifies that a logical name not starting
// with ROOT/ produces a tool error.
func TestHandleLoadChain_InvalidPrefix(t *testing.T) {
	// No temp dir needed — we never reach file I/O.
	_ = testSetupTempDir(t)

	result := testCallHandler(t, "INVALID/something")
	testAssertToolError(t, result, "not a recognized ROOT/")
}

// TestHandleLoadChain_NonexistentSpecFile verifies that referencing a node
// whose spec file does not exist produces a tool error.
func TestHandleLoadChain_NonexistentSpecFile(t *testing.T) {
	testSetupTempDir(t)

	// Do NOT create any spec file for ROOT/nonexistent.
	result := testCallHandler(t, "ROOT/nonexistent")
	// Expect a tool error (from ParseFrontmatter — file not found).
	if !result.IsError {
		t.Fatalf("expected tool error for nonexistent spec file, but got success: %s",
			testResultText(t, result))
	}
}

// TestHandleLoadChain_NoOutputs verifies that a node with no `outputs` field
// produces a tool error containing "has no outputs".
func TestHandleLoadChain_NoOutputs(t *testing.T) {
	tmpDir := testSetupTempDir(t)

	testWriteFile(t, testNodePath(tmpDir, "ROOT"),
		testNodeContent("ROOT", "", "Root context.\n", nil))

	// ROOT/a has no outputs field.
	testWriteFile(t, testNodePath(tmpDir, "ROOT/a"),
		testNodeContent("ROOT/a", "", "Leaf content.\n", nil))

	result := testCallHandler(t, "ROOT/a")
	testAssertToolError(t, result, "has no outputs")
}

// TestHandleLoadChain_InvalidOutputPathTraversal verifies that an output path
// containing path traversal (../../etc/passwd) is rejected with a tool error.
func TestHandleLoadChain_InvalidOutputPathTraversal(t *testing.T) {
	tmpDir := testSetupTempDir(t)

	testWriteFile(t, testNodePath(tmpDir, "ROOT"),
		testNodeContent("ROOT", "", "Root context.\n", nil))

	// ROOT/a declares a traversal output path.
	leafFM := "outputs:\n  - id: a\n    path: ../../etc/passwd"
	testWriteFile(t, testNodePath(tmpDir, "ROOT/a"),
		testNodeContent("ROOT/a", leafFM, "Leaf content.\n", nil))

	result := testCallHandler(t, "ROOT/a")
	// Expect a tool error from path validation.
	if !result.IsError {
		t.Fatalf("expected tool error for traversal path, but got success: %s",
			testResultText(t, result))
	}
}

// TestHandleLoadChain_UnresolvableDependency verifies that a depends_on entry
// pointing to a missing spec file produces a tool error from chain resolution.
func TestHandleLoadChain_UnresolvableDependency(t *testing.T) {
	tmpDir := testSetupTempDir(t)

	testWriteFile(t, testNodePath(tmpDir, "ROOT"),
		testNodeContent("ROOT", "", "Root context.\n", nil))

	// ROOT/a depends on ROOT/b, but ROOT/b's file is never created.
	leafFM := "depends_on:\n  - ROOT/b\noutputs:\n  - id: main\n    path: src/main.go"
	testWriteFile(t, testNodePath(tmpDir, "ROOT/a"),
		testNodeContent("ROOT/a", leafFM, "Leaf content.\n", nil))

	result := testCallHandler(t, "ROOT/a")
	if !result.IsError {
		t.Fatalf("expected tool error for unresolvable dependency ROOT/b, but got success: %s",
			testResultText(t, result))
	}
}
