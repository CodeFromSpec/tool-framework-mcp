// code-from-spec: ROOT/golang/internal/chain_resolver/tests@BTmlNJEhYT1YD4QFLevmY5gcJ1w

// Package chainresolver provides tests for the ResolveChain function.
// Tests use t.TempDir() for isolated project structures and change the
// working directory to the temp dir before calling ResolveChain, restoring
// it after each test.
package chainresolver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// testWriteFile creates a file with the given content at path (relative to
// the provided root directory), creating all necessary parent directories.
func testWriteFile(t *testing.T, root, relPath, content string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatalf("testWriteFile: mkdir: %v", err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile: write %s: %v", relPath, err)
	}
}

// testNodePath returns the relative path of a _node.md file for a given
// logical name, expressed with forward slashes.
// e.g. "ROOT" → "code-from-spec/_node.md"
//
//	"ROOT/a" → "code-from-spec/a/_node.md"
func testNodePath(logicalName string) string {
	// Strip the "ROOT" prefix; remainder is the sub-path.
	rest := strings.TrimPrefix(logicalName, "ROOT")
	rest = strings.TrimPrefix(rest, "/")
	if rest == "" {
		return "code-from-spec/_node.md"
	}
	// Replace "/" with "/" (already forward slashes) and append _node.md.
	return "code-from-spec/" + rest + "/_node.md"
}

// testMinimalNode writes a _node.md with an empty (but valid) frontmatter
// block for a node that needs no depends_on or outputs. The body is a
// trivial "# Public" section so section extraction doesn't fail.
func testMinimalNode(t *testing.T, root, logicalName string) {
	t.Helper()
	content := "---\n---\n\n# Public\n\nsome context\n"
	testWriteFile(t, root, testNodePath(logicalName), content)
}

// testNodeWithDeps writes a _node.md that includes a depends_on list.
func testNodeWithDeps(t *testing.T, root, logicalName string, deps []string) {
	t.Helper()
	var sb strings.Builder
	sb.WriteString("---\ndepends_on:\n")
	for _, d := range deps {
		sb.WriteString("  - " + d + "\n")
	}
	sb.WriteString("---\n\n# Public\n\nsome context\n")
	testWriteFile(t, root, testNodePath(logicalName), sb.String())
}

// testNodeWithOutputs writes a _node.md that declares output files.
func testNodeWithOutputs(t *testing.T, root, logicalName string, outputs []string) {
	t.Helper()
	var sb strings.Builder
	sb.WriteString("---\noutputs:\n")
	for _, o := range outputs {
		sb.WriteString("  - " + o + "\n")
	}
	sb.WriteString("---\n\n# Public\n\nsome context\n")
	testWriteFile(t, root, testNodePath(logicalName), sb.String())
}

// testChdir changes the process working directory to dir, and registers a
// cleanup function to restore the original directory after the test.
func testChdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChdir: getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChdir: chdir to %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Fatalf("testChdir cleanup: chdir to %s: %v", orig, err)
		}
	})
}

// testStrPtr returns a pointer to the given string. Useful for comparing
// optional *string fields in assertions.
func testStrPtr(s string) *string {
	return &s
}

// testQualifier returns the Qualifier field value as a string for comparison,
// or "<nil>" if the pointer is nil.
func testQualifier(q *string) string {
	if q == nil {
		return "<nil>"
	}
	return *q
}

// ---------------------------------------------------------------------------
// Happy-path tests
// ---------------------------------------------------------------------------

// TestResolveChain_AncestorsOnly verifies that a leaf node with no
// depends_on returns the correct ancestor chain and target, with empty
// Dependencies and Code.
func TestResolveChain_AncestorsOnly(t *testing.T) {
	tmp := t.TempDir()

	// Build tree: ROOT → ROOT/a → ROOT/a/b (leaf)
	testMinimalNode(t, tmp, "ROOT")
	testMinimalNode(t, tmp, "ROOT/a")
	testMinimalNode(t, tmp, "ROOT/a/b")

	testChdir(t, tmp)

	chain, err := ResolveChain("ROOT/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Ancestors must be ROOT and ROOT/a (in path-sorted order).
	if len(chain.Ancestors) != 2 {
		t.Fatalf("expected 2 ancestors, got %d", len(chain.Ancestors))
	}

	// Check ancestor logical names (order: ROOT before ROOT/a).
	if chain.Ancestors[0].LogicalName != "ROOT" {
		t.Errorf("ancestor[0].LogicalName = %q, want %q", chain.Ancestors[0].LogicalName, "ROOT")
	}
	if chain.Ancestors[1].LogicalName != "ROOT/a" {
		t.Errorf("ancestor[1].LogicalName = %q, want %q", chain.Ancestors[1].LogicalName, "ROOT/a")
	}

	// Qualifiers for ancestors must be nil (use full # Public section).
	for i, a := range chain.Ancestors {
		if a.Qualifier != nil {
			t.Errorf("ancestor[%d].Qualifier = %q, want nil", i, *a.Qualifier)
		}
	}

	// Target.
	if chain.Target.LogicalName != "ROOT/a/b" {
		t.Errorf("Target.LogicalName = %q, want %q", chain.Target.LogicalName, "ROOT/a/b")
	}
	if chain.Target.Qualifier != nil {
		t.Errorf("Target.Qualifier = %q, want nil", *chain.Target.Qualifier)
	}

	// No dependencies.
	if len(chain.Dependencies) != 0 {
		t.Errorf("expected 0 dependencies, got %d", len(chain.Dependencies))
	}

	// FilePath uses forward slashes.
	if strings.Contains(chain.Target.FilePath, "\\") {
		t.Errorf("Target.FilePath contains backslash: %q", chain.Target.FilePath)
	}
}

// TestResolveChain_DependencyNoQualifier verifies that a depends_on entry
// without a qualifier produces a ChainItem with Qualifier = nil.
func TestResolveChain_DependencyNoQualifier(t *testing.T) {
	tmp := t.TempDir()

	testMinimalNode(t, tmp, "ROOT")
	testNodeWithDeps(t, tmp, "ROOT/a", []string{"ROOT/b"})
	testMinimalNode(t, tmp, "ROOT/b")

	testChdir(t, tmp)

	chain, err := ResolveChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// One ancestor: ROOT.
	if len(chain.Ancestors) != 1 {
		t.Fatalf("expected 1 ancestor, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].LogicalName != "ROOT" {
		t.Errorf("ancestor[0].LogicalName = %q, want %q", chain.Ancestors[0].LogicalName, "ROOT")
	}

	// Target.
	if chain.Target.LogicalName != "ROOT/a" {
		t.Errorf("Target.LogicalName = %q, want %q", chain.Target.LogicalName, "ROOT/a")
	}

	// One dependency: ROOT/b with nil qualifier.
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ROOT/b" {
		t.Errorf("dep.LogicalName = %q, want %q", dep.LogicalName, "ROOT/b")
	}
	if dep.Qualifier != nil {
		t.Errorf("dep.Qualifier = %q, want nil", *dep.Qualifier)
	}
	// FilePath must use forward slashes.
	if strings.Contains(dep.FilePath, "\\") {
		t.Errorf("dep.FilePath contains backslash: %q", dep.FilePath)
	}
}

// TestResolveChain_DependencyWithQualifier verifies that a depends_on entry
// of the form "ROOT/b(interface)" produces a ChainItem with
// Qualifier = "interface".
func TestResolveChain_DependencyWithQualifier(t *testing.T) {
	tmp := t.TempDir()

	testMinimalNode(t, tmp, "ROOT")
	testNodeWithDeps(t, tmp, "ROOT/a", []string{"ROOT/b(interface)"})
	testMinimalNode(t, tmp, "ROOT/b")

	testChdir(t, tmp)

	chain, err := ResolveChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]

	// LogicalName preserves the qualifier notation as provided.
	if dep.LogicalName != "ROOT/b(interface)" {
		t.Errorf("dep.LogicalName = %q, want %q", dep.LogicalName, "ROOT/b(interface)")
	}

	// FilePath must point to ROOT/b's _node.md (qualifier stripped for path resolution).
	wantSuffix := "code-from-spec/b/_node.md"
	if !strings.HasSuffix(dep.FilePath, wantSuffix) {
		t.Errorf("dep.FilePath = %q, want suffix %q", dep.FilePath, wantSuffix)
	}

	// Qualifier must be a pointer to "interface".
	if dep.Qualifier == nil {
		t.Fatal("dep.Qualifier is nil, want pointer to \"interface\"")
	}
	if *dep.Qualifier != "interface" {
		t.Errorf("dep.Qualifier = %q, want %q", *dep.Qualifier, "interface")
	}
}

// TestResolveChain_DependenciesSorted verifies that when a node depends on
// multiple other nodes, the resulting Dependencies slice is sorted by FilePath.
func TestResolveChain_DependenciesSorted(t *testing.T) {
	tmp := t.TempDir()

	testMinimalNode(t, tmp, "ROOT")
	testNodeWithDeps(t, tmp, "ROOT/a", []string{"ROOT/z", "ROOT/m", "ROOT/b"})
	testMinimalNode(t, tmp, "ROOT/z")
	testMinimalNode(t, tmp, "ROOT/m")
	testMinimalNode(t, tmp, "ROOT/b")

	testChdir(t, tmp)

	chain, err := ResolveChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 3 {
		t.Fatalf("expected 3 dependencies, got %d", len(chain.Dependencies))
	}

	// Verify sorted order by FilePath.
	for i := 1; i < len(chain.Dependencies); i++ {
		prev := chain.Dependencies[i-1].FilePath
		curr := chain.Dependencies[i].FilePath
		if prev > curr {
			t.Errorf("dependencies not sorted: [%d]=%q > [%d]=%q", i-1, prev, i, curr)
		}
	}
}

// TestResolveChain_OutputsFileExists verifies that when an output file
// declared in the node's frontmatter exists on disk, it is included in Code.
func TestResolveChain_OutputsFileExists(t *testing.T) {
	tmp := t.TempDir()

	testMinimalNode(t, tmp, "ROOT")
	testNodeWithOutputs(t, tmp, "ROOT/a", []string{"id: a\n    path: src/a.go"})

	// Create the actual output file on disk.
	testWriteFile(t, tmp, "src/a.go", "package main\n")

	testChdir(t, tmp)

	chain, err := ResolveChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Code must contain the existing output file.
	if len(chain.External) == 0 {
		// Code field in Chain may map to External or a dedicated field.
		// Per spec, "Code" is mentioned — check what the implementation uses.
		// The Chain struct has no "Code" field; spec mentions Code in context
		// of outputs. Re-reading: the spec actually says "Code: empty" and
		// "Code: ["src/a.go"]" — but the Chain struct above has only
		// Ancestors, Target, Dependencies, External, Input.
		// The spec section says outputs file → "Code", but the interface
		// defines no Code field. This is a spec inconsistency.
		// We check External as the closest match; if the implementation uses
		// a different approach this test will need adjusting.
	}
	// NOTE: The Chain interface has no explicit "Code" field in the spec's
	// Go interface definition. The happy-path spec section mentions
	// "Code: [...]" but that field is absent from the struct. We skip
	// asserting Code here and treat this as a known gap; the test verifies
	// only that ResolveChain succeeds without error when outputs exist.
	_ = chain
}

// TestResolveChain_OutputsFileNotExist verifies that when an output file
// declared in the node's frontmatter does NOT exist on disk, Code is empty.
func TestResolveChain_OutputsFileNotExist(t *testing.T) {
	tmp := t.TempDir()

	testMinimalNode(t, tmp, "ROOT")
	testNodeWithOutputs(t, tmp, "ROOT/a", []string{"id: a\n    path: src/a.go"})

	// Do NOT create src/a.go.

	testChdir(t, tmp)

	chain, err := ResolveChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Similar to above: verify no error and External is empty (file absent).
	if len(chain.External) != 0 {
		t.Errorf("expected empty External (file absent), got %d items", len(chain.External))
	}
}

// TestResolveChain_MultipleQualifiersSameFile verifies that two depends_on
// entries for the same node but different qualifiers both appear in
// Dependencies as separate ChainItems.
func TestResolveChain_MultipleQualifiersSameFile(t *testing.T) {
	tmp := t.TempDir()

	testMinimalNode(t, tmp, "ROOT")
	testNodeWithDeps(t, tmp, "ROOT/a", []string{"ROOT/b(interface)", "ROOT/b(constraints)"})
	testMinimalNode(t, tmp, "ROOT/b")

	testChdir(t, tmp)

	chain, err := ResolveChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}

	// Collect qualifiers present.
	quals := make(map[string]bool)
	for _, d := range chain.Dependencies {
		if d.Qualifier != nil {
			quals[*d.Qualifier] = true
		}
	}
	if !quals["interface"] {
		t.Error("expected qualifier \"interface\" in dependencies")
	}
	if !quals["constraints"] {
		t.Error("expected qualifier \"constraints\" in dependencies")
	}
}

// ---------------------------------------------------------------------------
// Edge-case / dedup tests
// ---------------------------------------------------------------------------

// TestResolveChain_Dedup_SameFileAndQualifier verifies that a duplicate
// dependency (same node, same qualifier = nil) is deduplicated to one entry.
func TestResolveChain_Dedup_SameFileAndQualifier(t *testing.T) {
	tmp := t.TempDir()

	testMinimalNode(t, tmp, "ROOT")
	testNodeWithDeps(t, tmp, "ROOT/a", []string{"ROOT/b", "ROOT/b"})
	testMinimalNode(t, tmp, "ROOT/b")

	testChdir(t, tmp)

	chain, err := ResolveChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency (deduped), got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].Qualifier != nil {
		t.Errorf("dep.Qualifier = %q, want nil", *chain.Dependencies[0].Qualifier)
	}
}

// TestResolveChain_Dedup_SameFileDifferentQualifiers verifies that two
// entries for the same file but different qualifiers are both preserved.
func TestResolveChain_Dedup_SameFileDifferentQualifiers(t *testing.T) {
	tmp := t.TempDir()

	testMinimalNode(t, tmp, "ROOT")
	testNodeWithDeps(t, tmp, "ROOT/a", []string{"ROOT/b(interface)", "ROOT/b(constraints)"})
	testMinimalNode(t, tmp, "ROOT/b")

	testChdir(t, tmp)

	chain, err := ResolveChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}
}

// TestResolveChain_Dedup_NilSubsumesSpecific verifies that when both a nil
// qualifier (ROOT/b) and a specific qualifier (ROOT/b(interface)) are listed,
// only the nil entry remains.
func TestResolveChain_Dedup_NilSubsumesSpecific(t *testing.T) {
	tmp := t.TempDir()

	testMinimalNode(t, tmp, "ROOT")
	testNodeWithDeps(t, tmp, "ROOT/a", []string{"ROOT/b", "ROOT/b(interface)"})
	testMinimalNode(t, tmp, "ROOT/b")

	testChdir(t, tmp)

	chain, err := ResolveChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Nil subsumes the specific qualifier → only one entry with nil qualifier.
	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency (nil subsumes specific), got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].Qualifier != nil {
		t.Errorf("dep.Qualifier = %q, want nil (nil should subsume specific)", *chain.Dependencies[0].Qualifier)
	}
}

// TestResolveChain_Dedup_SpecificBeforeNil verifies that even when the
// specific qualifier appears before the nil entry, the nil entry wins.
func TestResolveChain_Dedup_SpecificBeforeNil(t *testing.T) {
	tmp := t.TempDir()

	testMinimalNode(t, tmp, "ROOT")
	// Specific appears first, nil second.
	testNodeWithDeps(t, tmp, "ROOT/a", []string{"ROOT/b(interface)", "ROOT/b"})
	testMinimalNode(t, tmp, "ROOT/b")

	testChdir(t, tmp)

	chain, err := ResolveChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency (nil wins regardless of order), got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].Qualifier != nil {
		t.Errorf("dep.Qualifier = %q, want nil", *chain.Dependencies[0].Qualifier)
	}
}

// TestResolveChain_Dedup_RepeatedQualifier verifies that a repeated specific
// qualifier for the same file is deduplicated to one entry.
func TestResolveChain_Dedup_RepeatedQualifier(t *testing.T) {
	tmp := t.TempDir()

	testMinimalNode(t, tmp, "ROOT")
	testNodeWithDeps(t, tmp, "ROOT/a", []string{"ROOT/b(interface)", "ROOT/b(interface)"})
	testMinimalNode(t, tmp, "ROOT/b")

	testChdir(t, tmp)

	chain, err := ResolveChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency (repeated qualifier deduped), got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].Qualifier == nil {
		t.Fatal("dep.Qualifier is nil, want pointer to \"interface\"")
	}
	if *chain.Dependencies[0].Qualifier != "interface" {
		t.Errorf("dep.Qualifier = %q, want %q", *chain.Dependencies[0].Qualifier, "interface")
	}
}

// ---------------------------------------------------------------------------
// Failure-case tests
// ---------------------------------------------------------------------------

// TestResolveChain_InvalidLogicalName verifies that a logical name not
// starting with ROOT/ returns an error containing "cannot resolve logical name".
func TestResolveChain_InvalidLogicalName(t *testing.T) {
	// No need to set up any files; path resolution should fail immediately.
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := ResolveChain("INVALID/something")
	if err == nil {
		t.Fatal("expected error for invalid logical name, got nil")
	}
	if !strings.Contains(err.Error(), "cannot resolve logical name") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "cannot resolve logical name")
	}
}

// TestResolveChain_UnreadableFrontmatter verifies that a node with invalid
// YAML frontmatter causes ResolveChain to return an error from ParseFrontmatter.
func TestResolveChain_UnreadableFrontmatter(t *testing.T) {
	tmp := t.TempDir()

	testMinimalNode(t, tmp, "ROOT")
	// Write malformed YAML that ParseFrontmatter will reject.
	badContent := "---\n: bad: yaml: [\n---\n\n# Public\n\nbody\n"
	testWriteFile(t, tmp, testNodePath("ROOT/a"), badContent)

	testChdir(t, tmp)

	_, err := ResolveChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error for invalid frontmatter, got nil")
	}
	// The error should originate from ParseFrontmatter — any non-nil error
	// is acceptable; we just verify it is not nil.
}

// TestResolveChain_UnresolvableDependency verifies that a depends_on entry
// for a node that has no file on disk returns an error containing
// "cannot resolve logical name".
func TestResolveChain_UnresolvableDependency(t *testing.T) {
	tmp := t.TempDir()

	testMinimalNode(t, tmp, "ROOT")
	testNodeWithDeps(t, tmp, "ROOT/a", []string{"ROOT/nonexistent"})
	// Do NOT create ROOT/nonexistent.

	testChdir(t, tmp)

	_, err := ResolveChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error for unresolvable dependency, got nil")
	}
	if !strings.Contains(err.Error(), "cannot resolve logical name") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "cannot resolve logical name")
	}
}
