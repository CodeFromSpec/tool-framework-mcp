// code-from-spec: ROOT/golang/internal/format_validation/tests@JdTD6KQrrcECKfhBpISZc2MAjk8

// Package formatvalidation provides tests for the ValidateFormat function.
// Each test creates an isolated temporary directory, writes _node.md files
// with controlled content, builds a DiscoveredNode slice, and asserts the
// expected FormatError results.
package formatvalidation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/nodediscovery"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// testCase is a table-driven test entry.
type testCase struct {
	name        string
	setup       func(dir string) []nodediscovery.DiscoveredNode
	wantErrLen  int    // expected number of FormatErrors
	wantRule    string // if non-empty, the single expected Rule substring
	wantNoError bool   // if true, expect no FormatError at all (happy path)
}

// testWriteNode writes content into <dir>/<relPath>/_node.md, creating
// intermediate directories as needed.
func testWriteNode(t *testing.T, dir, relPath, content string) string {
	t.Helper()
	fullDir := filepath.Join(dir, filepath.FromSlash(relPath))
	if err := os.MkdirAll(fullDir, 0o755); err != nil {
		t.Fatalf("testWriteNode: MkdirAll %s: %v", fullDir, err)
	}
	p := filepath.Join(fullDir, "_node.md")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("testWriteNode: WriteFile %s: %v", p, err)
	}
	return p
}

// testNode builds a nodediscovery.DiscoveredNode from a logical name and the
// absolute file path returned by testWriteNode.
func testNode(logicalName, filePath string) nodediscovery.DiscoveredNode {
	return nodediscovery.DiscoveredNode{
		LogicalName: logicalName,
		FilePath:    filePath,
	}
}

// testContainsRule returns true when any FormatError in errs has a Rule that
// contains the given substring.
func testContainsRule(errs []FormatError, ruleSubstr string) bool {
	for _, e := range errs {
		if containsString(e.Rule, ruleSubstr) {
			return true
		}
	}
	return false
}

// containsString is a simple substring check without importing strings in
// test helpers (avoids confusion; we use it only here).
func containsString(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}())
}

// ---------------------------------------------------------------------------
// Happy-path tests
// ---------------------------------------------------------------------------

// TestValidLeafNode verifies that a well-formed leaf node produces no errors.
// A leaf node has no children in the discovered set.
func TestValidLeafNode(t *testing.T) {
	dir := t.TempDir()

	// Minimal valid leaf node:
	//   - Heading matches logical name "ROOT/a"
	//   - Valid frontmatter with outputs list
	//   - # Public and # Agent sections
	content := `---
outputs:
  - id: main
    path: internal/a/a.go
---
# ROOT/a

# Public

## Interface

Some interface docs.

# Agent

Agent instructions.
`
	p := testWriteNode(t, dir, "ROOT/a", content)
	nodes := []nodediscovery.DiscoveredNode{testNode("ROOT/a", p)}

	errs, err := ValidateFormat(nodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) != 0 {
		t.Errorf("expected no FormatErrors for valid leaf, got: %+v", errs)
	}
}

// TestValidIntermediateNode verifies that a well-formed intermediate node
// (one that has children) produces no errors. Intermediate nodes must NOT
// have frontmatter fields (outputs, depends_on) or a # Agent section.
func TestValidIntermediateNode(t *testing.T) {
	dir := t.TempDir()

	// Parent: heading only + # Public section; no frontmatter fields, no # Agent.
	parentContent := `# ROOT/parent

# Public

Overview of this subtree.
`
	// Child: valid leaf.
	childContent := `---
outputs:
  - id: main
    path: internal/parent/child/child.go
---
# ROOT/parent/child

# Public

## Interface

Child interface.

# Agent

Child agent instructions.
`
	pp := testWriteNode(t, dir, "ROOT/parent", parentContent)
	cp := testWriteNode(t, dir, "ROOT/parent/child", childContent)

	nodes := []nodediscovery.DiscoveredNode{
		testNode("ROOT/parent", pp),
		testNode("ROOT/parent/child", cp),
	}

	errs, err := ValidateFormat(nodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) != 0 {
		t.Errorf("expected no FormatErrors for valid intermediate node, got: %+v", errs)
	}
}

// ---------------------------------------------------------------------------
// Failure-case tests
// ---------------------------------------------------------------------------

// TestHeadingMismatch verifies that a node whose first heading does not match
// its logical name produces a FormatError about heading/name verification.
func TestHeadingMismatch(t *testing.T) {
	dir := t.TempDir()

	// Heading says "ROOT/wrong" but logical name is "ROOT/a".
	content := `---
outputs:
  - id: main
    path: internal/a/a.go
---
# ROOT/wrong

# Public

# Agent
`
	p := testWriteNode(t, dir, "ROOT/a", content)
	nodes := []nodediscovery.DiscoveredNode{testNode("ROOT/a", p)}

	errs, err := ValidateFormat(nodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) == 0 {
		t.Fatal("expected at least one FormatError for heading mismatch, got none")
	}
	// The error rule should indicate a name/heading verification failure.
	if !testContainsRule(errs, "heading") {
		t.Errorf("expected a FormatError with rule containing 'heading', got: %+v", errs)
	}
}

// TestIntermediateNodeWithOutputs verifies that an intermediate node (one with
// children) that declares outputs in frontmatter produces a FormatError.
func TestIntermediateNodeWithOutputs(t *testing.T) {
	dir := t.TempDir()

	// Parent declares outputs — forbidden for intermediate nodes.
	parentContent := `---
outputs:
  - id: main
    path: internal/parent/parent.go
---
# ROOT/parent

# Public
`
	childContent := `---
outputs:
  - id: main
    path: internal/parent/child/child.go
---
# ROOT/parent/child

# Public

# Agent
`
	pp := testWriteNode(t, dir, "ROOT/parent", parentContent)
	cp := testWriteNode(t, dir, "ROOT/parent/child", childContent)

	nodes := []nodediscovery.DiscoveredNode{
		testNode("ROOT/parent", pp),
		testNode("ROOT/parent/child", cp),
	}

	errs, err := ValidateFormat(nodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) == 0 {
		t.Fatal("expected FormatError for intermediate node with outputs, got none")
	}
	if !testContainsRule(errs, "outputs") {
		t.Errorf("expected a FormatError with rule containing 'outputs', got: %+v", errs)
	}
}

// TestIntermediateNodeWithAgentSection verifies that an intermediate node that
// contains a # Agent section produces a FormatError.
func TestIntermediateNodeWithAgentSection(t *testing.T) {
	dir := t.TempDir()

	// Parent has a # Agent section — forbidden for intermediate nodes.
	parentContent := `# ROOT/parent

# Public

Overview.

# Agent

Should not be here for an intermediate node.
`
	childContent := `---
outputs:
  - id: main
    path: internal/parent/child/child.go
---
# ROOT/parent/child

# Public

# Agent
`
	pp := testWriteNode(t, dir, "ROOT/parent", parentContent)
	cp := testWriteNode(t, dir, "ROOT/parent/child", childContent)

	nodes := []nodediscovery.DiscoveredNode{
		testNode("ROOT/parent", pp),
		testNode("ROOT/parent/child", cp),
	}

	errs, err := ValidateFormat(nodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) == 0 {
		t.Fatal("expected FormatError for intermediate node with Agent section, got none")
	}
	if !testContainsRule(errs, "agent") {
		t.Errorf("expected a FormatError with rule containing 'agent', got: %+v", errs)
	}
}

// TestDependsOnNonExistentNode verifies that a depends_on entry pointing to a
// logical name not in the discovered set produces a FormatError.
func TestDependsOnNonExistentNode(t *testing.T) {
	dir := t.TempDir()

	content := `---
outputs:
  - id: main
    path: internal/a/a.go
depends_on:
  - ROOT/does-not-exist
---
# ROOT/a

# Public

# Agent
`
	p := testWriteNode(t, dir, "ROOT/a", content)
	nodes := []nodediscovery.DiscoveredNode{testNode("ROOT/a", p)}

	errs, err := ValidateFormat(nodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) == 0 {
		t.Fatal("expected FormatError for depends_on targeting non-existent node, got none")
	}
	if !testContainsRule(errs, "depends_on") {
		t.Errorf("expected a FormatError with rule containing 'depends_on', got: %+v", errs)
	}
}

// TestDependsOnTargetsAncestor verifies that a depends_on entry pointing to an
// ancestor node (which is already an implicit dependency) produces a
// FormatError for a redundant ancestor dependency.
//
// Setup: ROOT, ROOT/a, ROOT/a/b — ROOT/a/b depends_on ROOT.
func TestDependsOnTargetsAncestor(t *testing.T) {
	dir := t.TempDir()

	rootContent := `# ROOT

# Public

Root overview.
`
	aContent := `# ROOT/a

# Public

A overview.
`
	bContent := `---
outputs:
  - id: main
    path: internal/a/b/b.go
depends_on:
  - ROOT
---
# ROOT/a/b

# Public

# Agent
`
	rp := testWriteNode(t, dir, "ROOT", rootContent)
	ap := testWriteNode(t, dir, "ROOT/a", aContent)
	bp := testWriteNode(t, dir, "ROOT/a/b", bContent)

	nodes := []nodediscovery.DiscoveredNode{
		testNode("ROOT", rp),
		testNode("ROOT/a", ap),
		testNode("ROOT/a/b", bp),
	}

	errs, err := ValidateFormat(nodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) == 0 {
		t.Fatal("expected FormatError for depends_on targeting ancestor, got none")
	}
	if !testContainsRule(errs, "ancestor") {
		t.Errorf("expected a FormatError with rule containing 'ancestor', got: %+v", errs)
	}
}

// TestDependsOnTargetsDescendant verifies that a depends_on entry pointing to
// a descendant node produces a FormatError for a circular descendant
// dependency.
//
// Setup: ROOT/a depends_on ROOT/a/b, and ROOT/a/b exists.
func TestDependsOnTargetsDescendant(t *testing.T) {
	dir := t.TempDir()

	aContent := `---
depends_on:
  - ROOT/a/b
---
# ROOT/a

# Public
`
	bContent := `---
outputs:
  - id: main
    path: internal/a/b/b.go
---
# ROOT/a/b

# Public

# Agent
`
	ap := testWriteNode(t, dir, "ROOT/a", aContent)
	bp := testWriteNode(t, dir, "ROOT/a/b", bContent)

	nodes := []nodediscovery.DiscoveredNode{
		testNode("ROOT/a", ap),
		testNode("ROOT/a/b", bp),
	}

	errs, err := ValidateFormat(nodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) == 0 {
		t.Fatal("expected FormatError for depends_on targeting descendant, got none")
	}
	if !testContainsRule(errs, "descendant") {
		t.Errorf("expected a FormatError with rule containing 'descendant', got: %+v", errs)
	}
}

// TestOutputPathWithTraversal verifies that an output path containing ".."
// (directory traversal) produces a FormatError.
func TestOutputPathWithTraversal(t *testing.T) {
	dir := t.TempDir()

	content := `---
outputs:
  - id: main
    path: ../../../etc/passwd
---
# ROOT/a

# Public

# Agent
`
	p := testWriteNode(t, dir, "ROOT/a", content)
	nodes := []nodediscovery.DiscoveredNode{testNode("ROOT/a", p)}

	errs, err := ValidateFormat(nodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) == 0 {
		t.Fatal("expected FormatError for output path with traversal, got none")
	}
	if !testContainsRule(errs, "path") {
		t.Errorf("expected a FormatError with rule containing 'path', got: %+v", errs)
	}
}

// TestDuplicatePublicSubsections verifies that a node with two subsections of
// the same heading under # Public produces a FormatError.
func TestDuplicatePublicSubsections(t *testing.T) {
	dir := t.TempDir()

	// Two ## Interface headings under # Public — a duplicate.
	content := `---
outputs:
  - id: main
    path: internal/a/a.go
---
# ROOT/a

# Public

## Interface

First interface block.

## Interface

Duplicate interface block — should be rejected.

# Agent
`
	p := testWriteNode(t, dir, "ROOT/a", content)
	nodes := []nodediscovery.DiscoveredNode{testNode("ROOT/a", p)}

	errs, err := ValidateFormat(nodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) == 0 {
		t.Fatal("expected FormatError for duplicate Public subsections, got none")
	}
	if !testContainsRule(errs, "duplicate") {
		t.Errorf("expected a FormatError with rule containing 'duplicate', got: %+v", errs)
	}
}

// TestCollectsMultipleErrors verifies that ValidateFormat reports all
// violations on a single node rather than short-circuiting after the first.
func TestCollectsMultipleErrors(t *testing.T) {
	dir := t.TempDir()

	// This node has multiple violations at once:
	//   1. Heading mismatch (says ROOT/wrong instead of ROOT/a)
	//   2. Output path with traversal (../../../etc/passwd)
	//   3. depends_on targeting non-existent node
	content := `---
outputs:
  - id: main
    path: ../../../etc/passwd
depends_on:
  - ROOT/does-not-exist
---
# ROOT/wrong

# Public

# Agent
`
	p := testWriteNode(t, dir, "ROOT/a", content)
	nodes := []nodediscovery.DiscoveredNode{testNode("ROOT/a", p)}

	errs, err := ValidateFormat(nodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// We expect at least 2 distinct errors to confirm all violations are collected.
	if len(errs) < 2 {
		t.Errorf("expected multiple FormatErrors (at least 2), got %d: %+v", len(errs), errs)
	}
}

// ---------------------------------------------------------------------------
// Error-path test: unreadable node
// ---------------------------------------------------------------------------

// TestUnreadableNode verifies that ValidateFormat returns ErrUnreadableNode
// when a node's file cannot be read (e.g., path does not exist).
func TestUnreadableNode(t *testing.T) {
	// Provide a file path that does not exist on disk.
	nodes := []nodediscovery.DiscoveredNode{
		testNode("ROOT/a", "/this/path/does/not/exist/_node.md"),
	}

	_, err := ValidateFormat(nodes)
	if err == nil {
		t.Fatal("expected an error for unreadable node, got nil")
	}
	// The returned error must wrap or equal ErrUnreadableNode.
	if !isUnreadableNode(err) {
		t.Errorf("expected ErrUnreadableNode (or wrapping it), got: %v", err)
	}
}

// isUnreadableNode checks whether err is or wraps ErrUnreadableNode.
func isUnreadableNode(err error) bool {
	// Use errors.Is to handle wrapped errors correctly.
	// Import is omitted at the top to keep the helper self-contained;
	// we call errors.Is via the standard library here.
	return checkErrorIs(err, ErrUnreadableNode)
}

// checkErrorIs wraps errors.Is to avoid importing "errors" at package level
// just for this helper. In Go the standard errors package is always available.
func checkErrorIs(err, target error) bool {
	// Walk the error chain manually so we don't need an extra import.
	for err != nil {
		if err == target {
			return true
		}
		// Try to unwrap.
		type unwrapper interface{ Unwrap() error }
		if u, ok := err.(unwrapper); ok {
			err = u.Unwrap()
		} else {
			break
		}
	}
	return false
}
