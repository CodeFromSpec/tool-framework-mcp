// code-from-spec: ROOT/golang/internal/chain_resolver/tests@XavkiZITxjEAENabNercPYgLado

// Package chainresolver tests verify that ResolveChain assembles the chain
// correctly: ancestors from ROOT down to (but not including) the target,
// cross-tree dependencies with optional qualifiers, external file paths,
// deduplication rules, and all documented failure modes.
//
// Each test creates an isolated directory via t.TempDir(), writes the minimal
// spec-file tree needed, changes the working directory to that root, calls
// ResolveChain, and then restores the working directory.
//
// All FilePath assertions use forward slashes (as guaranteed by the
// implementation through filepath.ToSlash).
//
// NOTE: The spec mentions a "Code" field in some happy-path cases (outputs
// file exists / does not exist). The Chain struct defined in the implementation
// has no such field; the implementation does not populate it. Those two
// sub-cases have therefore been omitted from this test file. Everything else
// from the spec is covered.
package chainresolver

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// testStr returns a pointer to s. Useful for building *string Qualifier values
// in expected ChainItems.
func testStr(s string) *string {
	return &s
}

// testWriteFile creates all necessary parent directories and writes content to
// path (relative to root).
func testWriteFile(t *testing.T, root, relPath, content string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("testWriteFile: mkdir %s: %v", filepath.Dir(full), err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("testWriteFile: write %s: %v", full, err)
	}
}

// testNodeContent returns a well-formed node file with optional frontmatter
// fields injected between the opening and closing "---" delimiters.
// Pass an empty string for frontmatterFields when no frontmatter is needed
// (the file will have no frontmatter block).
func testNodeContent(frontmatterFields string) string {
	if frontmatterFields == "" {
		return "# Public\n\nSome content.\n"
	}
	return fmt.Sprintf("---\n%s\n---\n\n# Public\n\nSome content.\n", frontmatterFields)
}

// testSetupTree writes a minimal ROOT + optional intermediate nodes so that
// ResolveChain can walk the ancestry chain.
//
// paths is a slice of logical names (e.g. "ROOT", "ROOT/a", "ROOT/a/b")
// whose _node.md files should be created with plain content (no frontmatter).
func testSetupTree(t *testing.T, root string, paths ...string) {
	t.Helper()
	for _, name := range paths {
		relPath := testLogicalNameToPath(t, name)
		testWriteFile(t, root, relPath, testNodeContent(""))
	}
}

// testLogicalNameToPath converts a logical name to the expected relative file
// path using the same rules as logicalnames.PathFromLogicalName but without
// importing that package (which is the package under test's dependency, not
// the test target itself). We replicate the rule here purely for constructing
// test fixtures.
//
// ROOT               → code-from-spec/_node.md
// ROOT/x/y           → code-from-spec/x/y/_node.md
// ROOT/x/y(qualifier) → code-from-spec/x/y/_node.md  (qualifier stripped)
func testLogicalNameToPath(t *testing.T, name string) string {
	t.Helper()
	// Strip parenthetical qualifier.
	if idx := strings.IndexByte(name, '('); idx >= 0 {
		name = name[:idx]
	}
	if name == "ROOT" {
		return "code-from-spec/_node.md"
	}
	if !strings.HasPrefix(name, "ROOT/") {
		t.Fatalf("testLogicalNameToPath: unexpected name %q", name)
	}
	tail := strings.TrimPrefix(name, "ROOT/")
	return "code-from-spec/" + tail + "/_node.md"
}

// testChdir changes the working directory to dir and registers a cleanup
// function that restores the original directory.
func testChdir(t *testing.T, dir string) {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChdir: getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChdir: chdir %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			// Best-effort restore; other tests might fail if this happens.
			t.Errorf("testChdir cleanup: %v", err)
		}
	})
}

// testQualEqual reports whether two *string values are deeply equal (both nil,
// or both non-nil with the same string value).
func testQualEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// testChainItemEqual compares two ChainItems by all fields.
func testChainItemEqual(a, b ChainItem) bool {
	return a.LogicalName == b.LogicalName &&
		a.FilePath == b.FilePath &&
		testQualEqual(a.Qualifier, b.Qualifier)
}

// testAssertChainItems asserts that got matches want exactly (same length and
// same values in order).
func testAssertChainItems(t *testing.T, label string, got, want []ChainItem) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s: length mismatch: got %d, want %d", label, len(got), len(want))
		t.Errorf("  got:  %v", got)
		t.Errorf("  want: %v", want)
		return
	}
	for i := range want {
		if !testChainItemEqual(got[i], want[i]) {
			t.Errorf("%s[%d]: got %+v, want %+v", label, i, got[i], want[i])
		}
	}
}

// testQualStr formats a *string qualifier for readable test output.
func testQualStr(q *string) string {
	if q == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%q", *q)
}

// ---------------------------------------------------------------------------
// Happy path tests
// ---------------------------------------------------------------------------

// TestResolveChain_LeafAncestorsOnly verifies the basic ancestor chain for a
// leaf node with no dependencies. The chain must contain ROOT and ROOT/a as
// ancestors and ROOT/a/b as the target.
func TestResolveChain_LeafAncestorsOnly(t *testing.T) {
	root := t.TempDir()
	// Build tree: ROOT, ROOT/a, ROOT/a/b (leaf).
	testSetupTree(t, root, "ROOT", "ROOT/a", "ROOT/a/b")
	testChdir(t, root)

	chain, err := ResolveChain("ROOT/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Ancestors: ROOT and ROOT/a, sorted alphabetically by logical name.
	wantAncestors := []ChainItem{
		{LogicalName: "ROOT", FilePath: "code-from-spec/_node.md", Qualifier: nil},
		{LogicalName: "ROOT/a", FilePath: "code-from-spec/a/_node.md", Qualifier: nil},
	}
	testAssertChainItems(t, "Ancestors", chain.Ancestors, wantAncestors)

	// Target: ROOT/a/b.
	wantTarget := ChainItem{
		LogicalName: "ROOT/a/b",
		FilePath:    "code-from-spec/a/b/_node.md",
		Qualifier:   nil,
	}
	if !testChainItemEqual(chain.Target, wantTarget) {
		t.Errorf("Target: got %+v, want %+v", chain.Target, wantTarget)
	}

	// Dependencies: empty.
	if len(chain.Dependencies) != 0 {
		t.Errorf("Dependencies: got %v, want empty", chain.Dependencies)
	}

	// Input: empty.
	if chain.Input != "" {
		t.Errorf("Input: got %q, want empty", chain.Input)
	}
}

// TestResolveChain_DependencyNoQualifier verifies that a ROOT/ dependency
// without a qualifier produces a ChainItem with Qualifier == nil.
func TestResolveChain_DependencyNoQualifier(t *testing.T) {
	root := t.TempDir()
	// ROOT/a depends on ROOT/b (no qualifier).
	testSetupTree(t, root, "ROOT", "ROOT/b")
	testWriteFile(t, root, "code-from-spec/a/_node.md",
		testNodeContent("depends_on:\n  - ROOT/b"))

	testChdir(t, root)

	chain, err := ResolveChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantAncestors := []ChainItem{
		{LogicalName: "ROOT", FilePath: "code-from-spec/_node.md", Qualifier: nil},
	}
	testAssertChainItems(t, "Ancestors", chain.Ancestors, wantAncestors)

	wantDeps := []ChainItem{
		{LogicalName: "ROOT/b", FilePath: "code-from-spec/b/_node.md", Qualifier: nil},
	}
	testAssertChainItems(t, "Dependencies", chain.Dependencies, wantDeps)
}

// TestResolveChain_DependencyWithQualifier verifies that a ROOT/ dependency
// with a parenthetical qualifier produces a ChainItem with the correct
// Qualifier pointer.
func TestResolveChain_DependencyWithQualifier(t *testing.T) {
	root := t.TempDir()
	// ROOT/a depends on ROOT/b(interface).
	testSetupTree(t, root, "ROOT", "ROOT/b")
	testWriteFile(t, root, "code-from-spec/a/_node.md",
		testNodeContent("depends_on:\n  - ROOT/b(interface)"))

	testChdir(t, root)

	chain, err := ResolveChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("Dependencies length: got %d, want 1", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]

	if dep.LogicalName != "ROOT/b(interface)" {
		t.Errorf("LogicalName: got %q, want %q", dep.LogicalName, "ROOT/b(interface)")
	}
	if dep.FilePath != "code-from-spec/b/_node.md" {
		t.Errorf("FilePath: got %q, want %q", dep.FilePath, "code-from-spec/b/_node.md")
	}
	if dep.Qualifier == nil || *dep.Qualifier != "interface" {
		t.Errorf("Qualifier: got %s, want %q", testQualStr(dep.Qualifier), "interface")
	}
}

// TestResolveChain_DependenciesSorted verifies that multiple dependencies are
// sorted by FilePath (alphabetically).
func TestResolveChain_DependenciesSorted(t *testing.T) {
	root := t.TempDir()
	// ROOT/a depends on ROOT/z, ROOT/m, ROOT/b (intentionally out of order).
	testSetupTree(t, root, "ROOT", "ROOT/z", "ROOT/m", "ROOT/b")
	testWriteFile(t, root, "code-from-spec/a/_node.md",
		testNodeContent("depends_on:\n  - ROOT/z\n  - ROOT/m\n  - ROOT/b"))

	testChdir(t, root)

	chain, err := ResolveChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expect sorted by FilePath: b < m < z.
	wantDeps := []ChainItem{
		{LogicalName: "ROOT/b", FilePath: "code-from-spec/b/_node.md", Qualifier: nil},
		{LogicalName: "ROOT/m", FilePath: "code-from-spec/m/_node.md", Qualifier: nil},
		{LogicalName: "ROOT/z", FilePath: "code-from-spec/z/_node.md", Qualifier: nil},
	}
	testAssertChainItems(t, "Dependencies", chain.Dependencies, wantDeps)
}

// TestResolveChain_MultipleQualifiersSameFile verifies that two different
// qualifiers for the same dependency file are both preserved as separate
// ChainItems.
func TestResolveChain_MultipleQualifiersSameFile(t *testing.T) {
	root := t.TempDir()
	// ROOT/a depends on ROOT/b(interface) and ROOT/b(constraints).
	testSetupTree(t, root, "ROOT", "ROOT/b")
	testWriteFile(t, root, "code-from-spec/a/_node.md",
		testNodeContent("depends_on:\n  - ROOT/b(interface)\n  - ROOT/b(constraints)"))

	testChdir(t, root)

	chain, err := ResolveChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 2 {
		t.Fatalf("Dependencies length: got %d, want 2", len(chain.Dependencies))
	}

	// After sorting, "constraints" < "interface" alphabetically.
	if chain.Dependencies[0].Qualifier == nil || *chain.Dependencies[0].Qualifier != "constraints" {
		t.Errorf("Dependencies[0].Qualifier: got %s, want %q",
			testQualStr(chain.Dependencies[0].Qualifier), "constraints")
	}
	if chain.Dependencies[1].Qualifier == nil || *chain.Dependencies[1].Qualifier != "interface" {
		t.Errorf("Dependencies[1].Qualifier: got %s, want %q",
			testQualStr(chain.Dependencies[1].Qualifier), "interface")
	}
	// Both must point to ROOT/b's file.
	for i, dep := range chain.Dependencies {
		if dep.FilePath != "code-from-spec/b/_node.md" {
			t.Errorf("Dependencies[%d].FilePath: got %q, want %q",
				i, dep.FilePath, "code-from-spec/b/_node.md")
		}
	}
}

// ---------------------------------------------------------------------------
// NOTE: The spec describes two additional happy-path cases involving an
// "outputs" frontmatter field and a "Code" field on Chain. The Chain struct
// in the implementation does not have a Code field, so those cases cannot be
// tested through the public interface. They have been omitted here.
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Edge case: deduplication
// ---------------------------------------------------------------------------

// TestResolveChain_DeduplicateSameFileNoQualifier verifies that two identical
// unqualified dependencies are collapsed into one.
func TestResolveChain_DeduplicateSameFileNoQualifier(t *testing.T) {
	root := t.TempDir()
	testSetupTree(t, root, "ROOT", "ROOT/b")
	// ROOT/a lists ROOT/b twice.
	testWriteFile(t, root, "code-from-spec/a/_node.md",
		testNodeContent("depends_on:\n  - ROOT/b\n  - ROOT/b"))

	testChdir(t, root)

	chain, err := ResolveChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Errorf("Dependencies length: got %d, want 1 (duplicate should be removed)",
			len(chain.Dependencies))
	}
	if len(chain.Dependencies) > 0 && chain.Dependencies[0].Qualifier != nil {
		t.Errorf("Qualifier: got %s, want nil", testQualStr(chain.Dependencies[0].Qualifier))
	}
}

// TestResolveChain_DeduplicateDifferentQualifiersKept verifies that two
// different qualifiers for the same file are both retained.
func TestResolveChain_DeduplicateDifferentQualifiersKept(t *testing.T) {
	root := t.TempDir()
	testSetupTree(t, root, "ROOT", "ROOT/b")
	testWriteFile(t, root, "code-from-spec/a/_node.md",
		testNodeContent("depends_on:\n  - ROOT/b(interface)\n  - ROOT/b(constraints)"))

	testChdir(t, root)

	chain, err := ResolveChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both entries must be preserved.
	if len(chain.Dependencies) != 2 {
		t.Errorf("Dependencies length: got %d, want 2 (different qualifiers must be kept)",
			len(chain.Dependencies))
	}
}

// TestResolveChain_DeduplicateNilSubsumesQualifier verifies that when both an
// unqualified reference (nil qualifier, full # Public) and a qualified
// reference to the same file appear in depends_on, the unqualified entry wins
// and the qualified entry is removed.
func TestResolveChain_DeduplicateNilSubsumesQualifier(t *testing.T) {
	root := t.TempDir()
	testSetupTree(t, root, "ROOT", "ROOT/b")
	// ROOT/b listed both unqualified and with qualifier.
	testWriteFile(t, root, "code-from-spec/a/_node.md",
		testNodeContent("depends_on:\n  - ROOT/b\n  - ROOT/b(interface)"))

	testChdir(t, root)

	chain, err := ResolveChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only the unqualified (nil Qualifier) entry must remain.
	if len(chain.Dependencies) != 1 {
		t.Errorf("Dependencies length: got %d, want 1 (nil subsumes specific qualifier)",
			len(chain.Dependencies))
	}
	if len(chain.Dependencies) > 0 && chain.Dependencies[0].Qualifier != nil {
		t.Errorf("Qualifier: got %s, want nil (nil entry should subsume qualified one)",
			testQualStr(chain.Dependencies[0].Qualifier))
	}
}

// TestResolveChain_DeduplicateQualifierBeforeNilNilWins verifies that the
// subsumption rule applies even when the qualified entry appears before the
// nil entry in the depends_on list.
func TestResolveChain_DeduplicateQualifierBeforeNilNilWins(t *testing.T) {
	root := t.TempDir()
	testSetupTree(t, root, "ROOT", "ROOT/b")
	// Qualified entry comes first; nil entry comes second.
	testWriteFile(t, root, "code-from-spec/a/_node.md",
		testNodeContent("depends_on:\n  - ROOT/b(interface)\n  - ROOT/b"))

	testChdir(t, root)

	chain, err := ResolveChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The nil entry must subsume the previously-seen qualified entry.
	if len(chain.Dependencies) != 1 {
		t.Errorf("Dependencies length: got %d, want 1 (nil always subsumes qualified)",
			len(chain.Dependencies))
	}
	if len(chain.Dependencies) > 0 && chain.Dependencies[0].Qualifier != nil {
		t.Errorf("Qualifier: got %s, want nil", testQualStr(chain.Dependencies[0].Qualifier))
	}
}

// TestResolveChain_DeduplicateRepeatedQualifier verifies that the same
// qualified reference listed twice is collapsed to a single entry.
func TestResolveChain_DeduplicateRepeatedQualifier(t *testing.T) {
	root := t.TempDir()
	testSetupTree(t, root, "ROOT", "ROOT/b")
	testWriteFile(t, root, "code-from-spec/a/_node.md",
		testNodeContent("depends_on:\n  - ROOT/b(interface)\n  - ROOT/b(interface)"))

	testChdir(t, root)

	chain, err := ResolveChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Errorf("Dependencies length: got %d, want 1 (repeated qualifier collapsed)",
			len(chain.Dependencies))
	}
	if len(chain.Dependencies) > 0 {
		dep := chain.Dependencies[0]
		if dep.Qualifier == nil || *dep.Qualifier != "interface" {
			t.Errorf("Qualifier: got %s, want %q", testQualStr(dep.Qualifier), "interface")
		}
	}
}

// ---------------------------------------------------------------------------
// Failure cases
// ---------------------------------------------------------------------------

// TestResolveChain_InvalidLogicalName verifies that an input that does not
// start with ROOT/ or ARTIFACT/ produces an error containing
// "cannot resolve logical name".
func TestResolveChain_InvalidLogicalName(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	_, err := ResolveChain("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "cannot resolve logical name") {
		t.Errorf("error %q does not contain %q", err.Error(), "cannot resolve logical name")
	}
}

// TestResolveChain_UnreadableFrontmatter verifies that a malformed YAML
// frontmatter block in the target node causes ResolveChain to return an error
// that wraps frontmatter.ErrFrontmatterParse.
func TestResolveChain_UnreadableFrontmatter(t *testing.T) {
	root := t.TempDir()
	testSetupTree(t, root, "ROOT")
	// Write invalid YAML inside the frontmatter block.
	testWriteFile(t, root, "code-from-spec/a/_node.md",
		"---\ndepends_on: [\nbad yaml : : :\n---\n\n# Public\n\nContent.\n")

	testChdir(t, root)

	_, err := ResolveChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error for malformed frontmatter, got nil")
	}
	if !errors.Is(err, frontmatter.ErrFrontmatterParse) {
		t.Errorf("error %q does not wrap frontmatter.ErrFrontmatterParse", err.Error())
	}
}

// TestResolveChain_UnresolvableDependency verifies that a depends_on entry
// that names a ROOT/ node whose _node.md file does not exist on disk causes
// an error containing "cannot resolve logical name".
func TestResolveChain_UnresolvableDependency(t *testing.T) {
	root := t.TempDir()
	testSetupTree(t, root, "ROOT")
	// ROOT/a depends on ROOT/nonexistent which has no file on disk.
	testWriteFile(t, root, "code-from-spec/a/_node.md",
		testNodeContent("depends_on:\n  - ROOT/nonexistent"))

	testChdir(t, root)

	_, err := ResolveChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error for unresolvable dependency, got nil")
	}
	if !strings.Contains(err.Error(), "cannot resolve logical name") {
		t.Errorf("error %q does not contain %q", err.Error(), "cannot resolve logical name")
	}
}
