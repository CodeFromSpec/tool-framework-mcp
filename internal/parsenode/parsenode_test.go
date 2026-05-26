// code-from-spec: ROOT/golang/internal/parsenode/tests@zjeB7d_7JEoka04j--Wx-m-oMsg

// Package parsenode contains tests for the ParseNode function and its helpers.
//
// Tests are in the same package (internal test file) to allow access to
// unexported helpers if needed. All test helpers are prefixed with "test"
// to avoid name collisions with the package under test.
//
// Each test that calls ParseNode:
//   1. Creates files under t.TempDir() at the path returned by
//      logicalnames.PathFromLogicalName.
//   2. Changes the working directory to tmpdir before calling ParseNode.
//   3. Restores the working directory after (via t.Cleanup).
package parsenode

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// testWriteNode creates the spec node file for logicalName at the correct
// path relative to dir, creating intermediate directories as needed.
//
// The path is derived by the same rules as logicalnames.PathFromLogicalName:
//   ROOT                 -> <dir>/code-from-spec/_node.md
//   ROOT/x/y             -> <dir>/code-from-spec/x/y/_node.md
//   ROOT/x/y(z)          -> <dir>/code-from-spec/x/y/_node.md
func testWriteNode(t *testing.T, dir, logicalName, content string) {
	t.Helper()

	// Derive the relative path from the logical name.
	relPath := testPathFromLogicalName(t, logicalName)

	absPath := filepath.Join(dir, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		t.Fatalf("testWriteNode: MkdirAll: %v", err)
	}
	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNode: WriteFile: %v", err)
	}
}

// testPathFromLogicalName converts a logical name to a relative file path
// (forward-slash separated) following the same rules as
// logicalnames.PathFromLogicalName but implemented locally so the test helper
// has no external dependency on the logicalnames package.
//
//   ROOT               -> code-from-spec/_node.md
//   ROOT/x/y           -> code-from-spec/x/y/_node.md
//   ROOT/x/y(z)        -> code-from-spec/x/y/_node.md
func testPathFromLogicalName(t *testing.T, logicalName string) string {
	t.Helper()

	name := logicalName

	// Strip qualifier (anything from '(' onward).
	if idx := strings.Index(name, "("); idx >= 0 {
		name = name[:idx]
	}

	if name == "ROOT" {
		return "code-from-spec/_node.md"
	}

	const prefix = "ROOT/"
	if !strings.HasPrefix(name, prefix) {
		t.Fatalf("testPathFromLogicalName: unsupported logical name %q", logicalName)
	}
	rest := name[len(prefix):]
	return "code-from-spec/" + rest + "/_node.md"
}

// testChDir changes the working directory to dir and registers a cleanup
// function to restore the original working directory.
func testChDir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChDir: Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChDir: Chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("testChDir cleanup: Chdir back: %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// Happy path tests
// ---------------------------------------------------------------------------

// TestParseNode_MinimalNode verifies that a node with only a name section is
// parsed correctly: Public and Private are nil/empty, and the NameSection
// heading is normalised to lower-case.
func TestParseNode_MinimalNode(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/x"
	testWriteNode(t, dir, logicalName, `---
---
# ROOT/x

This node has only a name section.
`)
	testChDir(t, dir)

	got, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// NameSection checks.
	if want := "root/x"; got.NameSection.Heading != want {
		t.Errorf("NameSection.Heading = %q, want %q", got.NameSection.Heading, want)
	}
	if want := "This node has only a name section."; got.NameSection.Content != want {
		t.Errorf("NameSection.Content = %q, want %q", got.NameSection.Content, want)
	}
	if got.NameSection.Subsections != nil {
		t.Errorf("NameSection.Subsections = %v, want nil", got.NameSection.Subsections)
	}

	// Public and Private must be nil/empty.
	if got.Public != nil {
		t.Errorf("Public = %v, want nil", got.Public)
	}
	if len(got.Private) != 0 {
		t.Errorf("Private = %v, want empty", got.Private)
	}
}

// TestParseNode_FullNode verifies that a node with name, public, and private
// sections is parsed correctly, including subsections within Public.
func TestParseNode_FullNode(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/payments/fees"
	testWriteNode(t, dir, logicalName, `---
depends_on:
  - ROOT/architecture/backend
outputs:
  - id: fees
    path: internal/fees/fees.go
---
# ROOT/payments/fees

Calculates transaction fees.

# Public

## Interface

Fee calculation types and functions.

## Constraints

Maximum fee is 5%.

# Implementation

Step-by-step logic for fee calculation.

# Decisions

Chose percentage-based over flat fees.
`)
	testChDir(t, dir)

	got, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// NameSection.
	if want := "root/payments/fees"; got.NameSection.Heading != want {
		t.Errorf("NameSection.Heading = %q, want %q", got.NameSection.Heading, want)
	}
	if want := "Calculates transaction fees."; got.NameSection.Content != want {
		t.Errorf("NameSection.Content = %q, want %q", got.NameSection.Content, want)
	}

	// Public section.
	if got.Public == nil {
		t.Fatal("Public = nil, want non-nil")
	}
	if want := "public"; got.Public.Heading != want {
		t.Errorf("Public.Heading = %q, want %q", got.Public.Heading, want)
	}
	// Public.Content must be empty (no content before first ##).
	if got.Public.Content != "" {
		t.Errorf("Public.Content = %q, want %q", got.Public.Content, "")
	}
	if want := 2; len(got.Public.Subsections) != want {
		t.Fatalf("len(Public.Subsections) = %d, want %d", len(got.Public.Subsections), want)
	}

	// First subsection: Interface.
	sub0 := got.Public.Subsections[0]
	if want := "interface"; sub0.Heading != want {
		t.Errorf("Public.Subsections[0].Heading = %q, want %q", sub0.Heading, want)
	}
	if want := "Fee calculation types and functions."; sub0.Content != want {
		t.Errorf("Public.Subsections[0].Content = %q, want %q", sub0.Content, want)
	}

	// Second subsection: Constraints.
	sub1 := got.Public.Subsections[1]
	if want := "constraints"; sub1.Heading != want {
		t.Errorf("Public.Subsections[1].Heading = %q, want %q", sub1.Heading, want)
	}
	if want := "Maximum fee is 5%."; sub1.Content != want {
		t.Errorf("Public.Subsections[1].Content = %q, want %q", sub1.Content, want)
	}

	// Private sections.
	if want := 2; len(got.Private) != want {
		t.Fatalf("len(Private) = %d, want %d", len(got.Private), want)
	}
	if want := "implementation"; got.Private[0].Heading != want {
		t.Errorf("Private[0].Heading = %q, want %q", got.Private[0].Heading, want)
	}
	if want := "Step-by-step logic for fee calculation."; got.Private[0].Content != want {
		t.Errorf("Private[0].Content = %q, want %q", got.Private[0].Content, want)
	}
	if want := "decisions"; got.Private[1].Heading != want {
		t.Errorf("Private[1].Heading = %q, want %q", got.Private[1].Heading, want)
	}
	if want := "Chose percentage-based over flat fees."; got.Private[1].Content != want {
		t.Errorf("Private[1].Content = %q, want %q", got.Private[1].Content, want)
	}
}

// TestParseNode_NoPublicSection verifies that a node without a "# Public"
// section has Public == nil, but private sections are parsed correctly.
func TestParseNode_NoPublicSection(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/decisions"
	testWriteNode(t, dir, logicalName, `---
---
# ROOT/decisions

Architecture decisions.

# Rationale

Why we chose this approach.
`)
	testChDir(t, dir)

	got, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Public != nil {
		t.Errorf("Public = %v, want nil", got.Public)
	}
	if want := "root/decisions"; got.NameSection.Heading != want {
		t.Errorf("NameSection.Heading = %q, want %q", got.NameSection.Heading, want)
	}
	if want := 1; len(got.Private) != want {
		t.Fatalf("len(Private) = %d, want %d", len(got.Private), want)
	}
	if want := "rationale"; got.Private[0].Heading != want {
		t.Errorf("Private[0].Heading = %q, want %q", got.Private[0].Heading, want)
	}
}

// TestParseNode_PublicContentBeforeFirstSubsection verifies that text written
// directly under "# Public" (before any ## heading) is captured in
// Public.Content.
func TestParseNode_PublicContentBeforeFirstSubsection(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/a"
	testWriteNode(t, dir, logicalName, `---
---
# ROOT/a

Intent.

# Public

This is direct content of the public section.

## Interface

Types and functions.
`)
	testChDir(t, dir)

	got, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Public == nil {
		t.Fatal("Public = nil, want non-nil")
	}
	if want := "This is direct content of the public section."; got.Public.Content != want {
		t.Errorf("Public.Content = %q, want %q", got.Public.Content, want)
	}
	if want := 1; len(got.Public.Subsections) != want {
		t.Fatalf("len(Public.Subsections) = %d, want %d", len(got.Public.Subsections), want)
	}
	if want := "interface"; got.Public.Subsections[0].Heading != want {
		t.Errorf("Public.Subsections[0].Heading = %q, want %q", got.Public.Subsections[0].Heading, want)
	}
}

// ---------------------------------------------------------------------------
// Heading normalization tests
// ---------------------------------------------------------------------------

// TestParseNode_CaseInsensitivePublic verifies that "# PUBLIC" (all-caps) is
// recognized as the public section and stored with normalized heading "public".
func TestParseNode_CaseInsensitivePublic(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/c"
	testWriteNode(t, dir, logicalName, `---
---
# ROOT/c

Intent.

# PUBLIC

## Interface

Content.
`)
	testChDir(t, dir)

	got, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Public == nil {
		t.Fatal("Public = nil, want non-nil")
	}
	if want := "public"; got.Public.Heading != want {
		t.Errorf("Public.Heading = %q, want %q", got.Public.Heading, want)
	}
}

// TestParseNode_PublicMixedCaseAndWhitespace verifies that a heading like
// "#   PuBLiC" (extra leading spaces, mixed case) is still recognized as the
// public section.
func TestParseNode_PublicMixedCaseAndWhitespace(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/d"
	// Note: the heading "#   PuBLiC" has extra spaces between # and the word.
	testWriteNode(t, dir, logicalName, "---\n---\n# ROOT/d\n\nIntent.\n\n#   PuBLiC\n\n## Interface\n\nContent.\n")
	testChDir(t, dir)

	got, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Public == nil {
		t.Fatal("Public = nil, want non-nil")
	}
	if want := "public"; got.Public.Heading != want {
		t.Errorf("Public.Heading = %q, want %q", got.Public.Heading, want)
	}
}

// TestParseNode_NodeNameVariedWhitespace verifies that extra whitespace in the
// node name heading (e.g., "#    ROOT/e") does not prevent matching.
func TestParseNode_NodeNameVariedWhitespace(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/e"
	// The node name heading has 4 spaces between '#' and "ROOT/e".
	// In CommonMark, goldmark strips the '#' prefix and trims; however the
	// exact trimming depends on goldmark's behaviour. We write the content as
	// the spec example shows.
	testWriteNode(t, dir, logicalName, "---\n---\n#    ROOT/e\n\nIntent.\n")
	testChDir(t, dir)

	got, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if want := "root/e"; got.NameSection.Heading != want {
		t.Errorf("NameSection.Heading = %q, want %q", got.NameSection.Heading, want)
	}
}

// TestParseNode_SubsectionHeadingsNormalized verifies that subsection headings
// with mixed case and extra whitespace are all stored in normalized form.
func TestParseNode_SubsectionHeadingsNormalized(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/f"
	testWriteNode(t, dir, logicalName, `---
---
# ROOT/f

Intent.

# Public

##   Interface

Types.

## CONSTRAINTS

Rules.
`)
	testChDir(t, dir)

	got, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Public == nil {
		t.Fatal("Public = nil, want non-nil")
	}
	if want := 2; len(got.Public.Subsections) != want {
		t.Fatalf("len(Public.Subsections) = %d, want %d", len(got.Public.Subsections), want)
	}
	if want := "interface"; got.Public.Subsections[0].Heading != want {
		t.Errorf("Public.Subsections[0].Heading = %q, want %q", got.Public.Subsections[0].Heading, want)
	}
	if want := "constraints"; got.Public.Subsections[1].Heading != want {
		t.Errorf("Public.Subsections[1].Heading = %q, want %q", got.Public.Subsections[1].Heading, want)
	}
}

// TestParseNode_TabCharactersInHeadingWhitespace verifies that a subsection
// heading with tab characters around the heading text is normalized correctly.
func TestParseNode_TabCharactersInHeadingWhitespace(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/g"
	// The ## heading contains tabs around "Interface".
	testWriteNode(t, dir, logicalName, "---\n---\n# ROOT/g\n\nIntent.\n\n# Public\n\n## \tInterface\t\n\nContent.\n")
	testChDir(t, dir)

	got, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Public == nil {
		t.Fatal("Public = nil, want non-nil")
	}
	if want := 1; len(got.Public.Subsections) != want {
		t.Fatalf("len(Public.Subsections) = %d, want %d", len(got.Public.Subsections), want)
	}
	if want := "interface"; got.Public.Subsections[0].Heading != want {
		t.Errorf("Public.Subsections[0].Heading = %q, want %q", got.Public.Subsections[0].Heading, want)
	}
}

// ---------------------------------------------------------------------------
// Content extraction tests
// ---------------------------------------------------------------------------

// TestParseNode_Level3AndDeeperAsContent verifies that level-3 and deeper
// headings (### and ####) are treated as content, not as structural headings,
// and are included verbatim in the subsection content.
func TestParseNode_Level3AndDeeperAsContent(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/h"
	testWriteNode(t, dir, logicalName, `---
---
# ROOT/h

Intent.

# Public

## Interface

### Types

Type definitions here.

#### Nested detail

Even deeper content.

## Constraints

### Rule one

Details.
`)
	testChDir(t, dir)

	got, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Public == nil {
		t.Fatal("Public = nil, want non-nil")
	}
	if want := 2; len(got.Public.Subsections) != want {
		t.Fatalf("len(Public.Subsections) = %d, want %d", len(got.Public.Subsections), want)
	}

	// "interface" subsection must contain the ### and #### headings verbatim.
	interfaceContent := got.Public.Subsections[0].Content
	for _, wantFragment := range []string{"### Types", "Type definitions here.", "#### Nested detail", "Even deeper content."} {
		if !strings.Contains(interfaceContent, wantFragment) {
			t.Errorf("Public.Subsections[0].Content missing %q; got:\n%s", wantFragment, interfaceContent)
		}
	}

	// "constraints" subsection must contain the ### heading verbatim.
	constraintsContent := got.Public.Subsections[1].Content
	for _, wantFragment := range []string{"### Rule one", "Details."} {
		if !strings.Contains(constraintsContent, wantFragment) {
			t.Errorf("Public.Subsections[1].Content missing %q; got:\n%s", wantFragment, constraintsContent)
		}
	}
}

// TestParseNode_FencedCodeBlockWithHeadingLikeContent verifies that # and ##
// markers inside a fenced code block are not treated as structural headings.
func TestParseNode_FencedCodeBlockWithHeadingLikeContent(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/i"
	testWriteNode(t, dir, logicalName, "---\n---\n# ROOT/i\n\nIntent.\n\n# Public\n\n## Interface\n\n~~~go\n// # This is not a heading\n// ## Neither is this\nfunc Foo() {}\n~~~\n\nAfter the code block.\n")
	testChDir(t, dir)

	got, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Public == nil {
		t.Fatal("Public = nil, want non-nil")
	}
	// There should be exactly one subsection — the fenced-block content must
	// not have created extra subsections.
	if want := 1; len(got.Public.Subsections) != want {
		t.Fatalf("len(Public.Subsections) = %d, want %d; subsections: %+v",
			len(got.Public.Subsections), want, got.Public.Subsections)
	}
	sub := got.Public.Subsections[0]
	if want := "interface"; sub.Heading != want {
		t.Errorf("subsection heading = %q, want %q", sub.Heading, want)
	}
	// Content must include both the code block and the text after it.
	for _, wantFragment := range []string{
		"// # This is not a heading",
		"// ## Neither is this",
		"func Foo() {}",
		"After the code block.",
	} {
		if !strings.Contains(sub.Content, wantFragment) {
			t.Errorf("subsection content missing %q; got:\n%s", wantFragment, sub.Content)
		}
	}
}

// TestParseNode_ContentBetweenSectionsTrimmed verifies that leading and
// trailing blank lines around content are stripped in both section-level
// content and subsection content.
func TestParseNode_ContentBetweenSectionsTrimmed(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/j"
	testWriteNode(t, dir, logicalName, `---
---
# ROOT/j

Intent.

# Public



Content with surrounding blank lines.



## Interface

Also surrounded.

`)
	testChDir(t, dir)

	got, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Public == nil {
		t.Fatal("Public = nil, want non-nil")
	}
	if want := "Content with surrounding blank lines."; got.Public.Content != want {
		t.Errorf("Public.Content = %q, want %q", got.Public.Content, want)
	}
	if want := 1; len(got.Public.Subsections) != want {
		t.Fatalf("len(Public.Subsections) = %d, want %d", len(got.Public.Subsections), want)
	}
	if want := "Also surrounded."; got.Public.Subsections[0].Content != want {
		t.Errorf("Public.Subsections[0].Content = %q, want %q", got.Public.Subsections[0].Content, want)
	}
}

// ---------------------------------------------------------------------------
// Validation error tests
// ---------------------------------------------------------------------------

// TestParseNode_FileDoesNotExist verifies that calling ParseNode with a logical
// name whose file does not exist returns an error wrapping ErrRead.
func TestParseNode_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	// Do NOT create any file — we want the read to fail.
	testChDir(t, dir)

	_, err := ParseNode("ROOT/nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrRead) {
		t.Errorf("errors.Is(err, ErrRead) = false; got: %v", err)
	}
}

// TestParseNode_NoFrontmatterDelimiters verifies that a file without "---"
// delimiters is still parsed correctly — frontmatter is optional.
func TestParseNode_NoFrontmatterDelimiters(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/m"
	// No "---" delimiters at all.
	testWriteNode(t, dir, logicalName, "# ROOT/m\n\nJust text.\n")
	testChDir(t, dir)

	got, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if want := "root/m"; got.NameSection.Heading != want {
		t.Errorf("NameSection.Heading = %q, want %q", got.NameSection.Heading, want)
	}
	if want := "Just text."; got.NameSection.Content != want {
		t.Errorf("NameSection.Content = %q, want %q", got.NameSection.Content, want)
	}
}

// TestParseNode_ContentBeforeFirstHeading verifies that non-blank content
// appearing before any level-1 heading results in ErrUnexpectedContent.
func TestParseNode_ContentBeforeFirstHeading(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/o"
	testWriteNode(t, dir, logicalName, `---
---
Some text before any heading.

# ROOT/o

Intent.
`)
	testChDir(t, dir)

	_, err := ParseNode(logicalName)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUnexpectedContent) {
		t.Errorf("errors.Is(err, ErrUnexpectedContent) = false; got: %v", err)
	}
}

// TestParseNode_Level2BeforeLevel1 verifies that a level-2 heading appearing
// before any level-1 heading results in ErrUnexpectedContent.
func TestParseNode_Level2BeforeLevel1(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/p"
	testWriteNode(t, dir, logicalName, `---
---
## Orphan subsection

# ROOT/p

Intent.
`)
	testChDir(t, dir)

	_, err := ParseNode(logicalName)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUnexpectedContent) {
		t.Errorf("errors.Is(err, ErrUnexpectedContent) = false; got: %v", err)
	}
}

// TestParseNode_NodeNameDoesNotMatchLogicalName verifies that when the first
// level-1 heading does not match the logical name (after normalization), an
// error wrapping ErrInvalidNodeName is returned.
func TestParseNode_NodeNameDoesNotMatchLogicalName(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/q"
	testWriteNode(t, dir, logicalName, `---
---
# ROOT/wrong

Intent.
`)
	testChDir(t, dir)

	_, err := ParseNode(logicalName)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidNodeName) {
		t.Errorf("errors.Is(err, ErrInvalidNodeName) = false; got: %v", err)
	}
}

// TestParseNode_NodeNameCaseMismatchIsNotError verifies that a heading like
// "# root/Q" (wrong case) is accepted when it normalizes to the same value as
// the logical name "ROOT/q".
func TestParseNode_NodeNameCaseMismatchIsNotError(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/q"
	testWriteNode(t, dir, logicalName, `---
---
# root/Q

Intent.
`)
	testChDir(t, dir)

	got, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// After normalization both "root/Q" and "ROOT/q" become "root/q".
	if want := "root/q"; got.NameSection.Heading != want {
		t.Errorf("NameSection.Heading = %q, want %q", got.NameSection.Heading, want)
	}
}

// TestParseNode_DuplicatePublicSameCase verifies that two "# Public" headings
// (identical case) result in ErrDuplicatePublic.
func TestParseNode_DuplicatePublicSameCase(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/r"
	testWriteNode(t, dir, logicalName, `---
---
# ROOT/r

Intent.

# Public

First public.

# Public

Second public.
`)
	testChDir(t, dir)

	_, err := ParseNode(logicalName)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrDuplicatePublic) {
		t.Errorf("errors.Is(err, ErrDuplicatePublic) = false; got: %v", err)
	}
}

// TestParseNode_DuplicatePublicDifferentCase verifies that two headings that
// normalize to "public" (e.g., "# Public" and "# PUBLIC") result in
// ErrDuplicatePublic.
func TestParseNode_DuplicatePublicDifferentCase(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/s"
	testWriteNode(t, dir, logicalName, `---
---
# ROOT/s

Intent.

# Public

First.

# PUBLIC

Second.
`)
	testChDir(t, dir)

	_, err := ParseNode(logicalName)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrDuplicatePublic) {
		t.Errorf("errors.Is(err, ErrDuplicatePublic) = false; got: %v", err)
	}
}

// TestParseNode_DuplicateSubsectionSameCase verifies that two ## headings with
// the same text (identical case) within "# Public" result in
// ErrDuplicateSubsection.
func TestParseNode_DuplicateSubsectionSameCase(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/t"
	testWriteNode(t, dir, logicalName, `---
---
# ROOT/t

Intent.

# Public

## Interface

First interface.

## Interface

Second interface.
`)
	testChDir(t, dir)

	_, err := ParseNode(logicalName)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrDuplicateSubsection) {
		t.Errorf("errors.Is(err, ErrDuplicateSubsection) = false; got: %v", err)
	}
}

// TestParseNode_DuplicateSubsectionDifferentCase verifies that two ## headings
// that normalize to the same text (e.g., "Interface" and "INTERFACE") within
// "# Public" result in ErrDuplicateSubsection.
func TestParseNode_DuplicateSubsectionDifferentCase(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/u"
	testWriteNode(t, dir, logicalName, `---
---
# ROOT/u

Intent.

# Public

## Interface

First.

## INTERFACE

Second.
`)
	testChDir(t, dir)

	_, err := ParseNode(logicalName)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrDuplicateSubsection) {
		t.Errorf("errors.Is(err, ErrDuplicateSubsection) = false; got: %v", err)
	}
}

// TestParseNode_DuplicateSubsectionWhitespaceVariation verifies that two ##
// headings that normalize to the same text via whitespace trimming within
// "# Public" result in ErrDuplicateSubsection.
func TestParseNode_DuplicateSubsectionWhitespaceVariation(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/v"
	testWriteNode(t, dir, logicalName, `---
---
# ROOT/v

Intent.

# Public

## Interface

First.

##   Interface

Second.
`)
	testChDir(t, dir)

	_, err := ParseNode(logicalName)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrDuplicateSubsection) {
		t.Errorf("errors.Is(err, ErrDuplicateSubsection) = false; got: %v", err)
	}
}

// TestParseNode_ParagraphInsteadOfHeading verifies that a file whose body
// starts with a paragraph (not a heading) results in ErrUnexpectedContent.
func TestParseNode_ParagraphInsteadOfHeading(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/w"
	testWriteNode(t, dir, logicalName, `---
---
This is a paragraph, not a heading.
`)
	testChDir(t, dir)

	_, err := ParseNode(logicalName)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUnexpectedContent) {
		t.Errorf("errors.Is(err, ErrUnexpectedContent) = false; got: %v", err)
	}
}
