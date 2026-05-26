// code-from-spec: ROOT/golang/internal/parsenode/tests@qwIu1xrYxxguHNTW6s-oHlcUq2I

// Package parsenode contains tests for the ParseNode function.
// Tests cover happy-path parsing, heading normalization, content extraction,
// and all validation error paths defined in the specification.
//
// Each test creates a temporary directory (via t.TempDir()), writes a node
// file at the path that logicalnames.PathFromLogicalName would produce, then
// changes the process working directory to that temp dir before calling
// ParseNode, restoring the original working directory afterward.
package parsenode

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// testWriteNode creates the _node.md file for the given logicalName under
// dir, mirroring the path that PathFromLogicalName would produce.
// For logical name ROOT/x/y, the file is written to
// <dir>/code-from-spec/x/y/_node.md.
func testWriteNode(t *testing.T, dir string, logicalName string, content string) {
	t.Helper()

	// Strip the ROOT/ prefix, lower-case everything, then build the path.
	// e.g. ROOT/payments/fees → code-from-spec/payments/fees/_node.md
	rel := strings.TrimPrefix(logicalName, "ROOT/")
	parts := strings.Split(rel, "/")
	segments := append([]string{dir, "code-from-spec"}, parts...)
	segments = append(segments, "_node.md")
	fullPath := filepath.Join(segments...)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("testWriteNode: MkdirAll: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatalf("testWriteNode: WriteFile: %v", err)
	}
}

// testChdir changes the working directory to dir for the duration of the
// test, restoring the original directory when the test ends.
func testChdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChdir: Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChdir: Chdir to %q: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Fatalf("testChdir cleanup: Chdir to %q: %v", orig, err)
		}
	})
}

// ---------------------------------------------------------------------------
// Happy-path tests
// ---------------------------------------------------------------------------

// TestMinimalNode verifies that a node containing only a name section is
// parsed correctly, with nil Public and nil Private.
func TestMinimalNode(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/x"
	const content = `---
---
# ROOT/x

This node has only a name section.
`
	testWriteNode(t, dir, logicalName, content)
	testChdir(t, dir)

	node, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/x" {
		t.Errorf("NameSection.Heading = %q; want %q", node.NameSection.Heading, "root/x")
	}
	if node.NameSection.Content != "This node has only a name section." {
		t.Errorf("NameSection.Content = %q; want %q", node.NameSection.Content, "This node has only a name section.")
	}
	if node.NameSection.Subsections != nil {
		t.Errorf("NameSection.Subsections = %v; want nil", node.NameSection.Subsections)
	}
	if node.Public != nil {
		t.Errorf("Public = %v; want nil", node.Public)
	}
	if node.Agent != nil {
		t.Errorf("Agent = %v; want nil", node.Agent)
	}
	if node.Private != nil {
		t.Errorf("Private = %v; want nil", node.Private)
	}
}

// TestFullNode verifies correct parsing of a node with name, public, and
// multiple private sections including frontmatter with depends_on/outputs.
func TestFullNode(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/payments/fees"
	const content = `---
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
`
	testWriteNode(t, dir, logicalName, content)
	testChdir(t, dir)

	node, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// --- Name section ---
	if node.NameSection.Heading != "root/payments/fees" {
		t.Errorf("NameSection.Heading = %q; want %q", node.NameSection.Heading, "root/payments/fees")
	}
	if node.NameSection.Content != "Calculates transaction fees." {
		t.Errorf("NameSection.Content = %q; want %q", node.NameSection.Content, "Calculates transaction fees.")
	}

	// --- Public section ---
	if node.Public == nil {
		t.Fatal("Public = nil; want non-nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("Public.Heading = %q; want %q", node.Public.Heading, "public")
	}
	if node.Public.Content != "" {
		t.Errorf("Public.Content = %q; want empty string (no content before first ##)", node.Public.Content)
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("len(Public.Subsections) = %d; want 2", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("Public.Subsections[0].Heading = %q; want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if node.Public.Subsections[0].Content != "Fee calculation types and functions." {
		t.Errorf("Public.Subsections[0].Content = %q; want %q", node.Public.Subsections[0].Content, "Fee calculation types and functions.")
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("Public.Subsections[1].Heading = %q; want %q", node.Public.Subsections[1].Heading, "constraints")
	}
	if node.Public.Subsections[1].Content != "Maximum fee is 5%." {
		t.Errorf("Public.Subsections[1].Content = %q; want %q", node.Public.Subsections[1].Content, "Maximum fee is 5%.")
	}

	// --- Private sections ---
	if len(node.Private) != 2 {
		t.Fatalf("len(Private) = %d; want 2", len(node.Private))
	}
	if node.Private[0].Heading != "implementation" {
		t.Errorf("Private[0].Heading = %q; want %q", node.Private[0].Heading, "implementation")
	}
	if node.Private[0].Content != "Step-by-step logic for fee calculation." {
		t.Errorf("Private[0].Content = %q; want %q", node.Private[0].Content, "Step-by-step logic for fee calculation.")
	}
	if node.Private[1].Heading != "decisions" {
		t.Errorf("Private[1].Heading = %q; want %q", node.Private[1].Heading, "decisions")
	}
	if node.Private[1].Content != "Chose percentage-based over flat fees." {
		t.Errorf("Private[1].Content = %q; want %q", node.Private[1].Content, "Chose percentage-based over flat fees.")
	}
}

// TestNoPublicSection verifies a node that has only a name section and a
// private section produces nil Public.
func TestNoPublicSection(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/decisions"
	const content = `---
---
# ROOT/decisions

Architecture decisions.

# Rationale

Why we chose this approach.
`
	testWriteNode(t, dir, logicalName, content)
	testChdir(t, dir)

	node, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public != nil {
		t.Errorf("Public = %v; want nil", node.Public)
	}
	if node.NameSection.Heading != "root/decisions" {
		t.Errorf("NameSection.Heading = %q; want %q", node.NameSection.Heading, "root/decisions")
	}
	if len(node.Private) != 1 {
		t.Fatalf("len(Private) = %d; want 1", len(node.Private))
	}
	if node.Private[0].Heading != "rationale" {
		t.Errorf("Private[0].Heading = %q; want %q", node.Private[0].Heading, "rationale")
	}
}

// TestPublicSectionWithContentBeforeSubsection verifies that text appearing
// between the # Public heading and the first ## subsection is captured in
// Public.Content.
func TestPublicSectionWithContentBeforeSubsection(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/a"
	const content = `---
---
# ROOT/a

Intent.

# Public

This is direct content of the public section.

## Interface

Types and functions.
`
	testWriteNode(t, dir, logicalName, content)
	testChdir(t, dir)

	node, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public = nil; want non-nil")
	}
	if node.Public.Content != "This is direct content of the public section." {
		t.Errorf("Public.Content = %q; want %q", node.Public.Content, "This is direct content of the public section.")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("len(Public.Subsections) = %d; want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("Public.Subsections[0].Heading = %q; want %q", node.Public.Subsections[0].Heading, "interface")
	}
}

// ---------------------------------------------------------------------------
// Heading normalization tests
// ---------------------------------------------------------------------------

// TestCaseInsensitivePublicDetection verifies that # PUBLIC (all caps) is
// recognized as the public section and its heading is stored in lowercase.
func TestCaseInsensitivePublicDetection(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/c"
	const content = `---
---
# ROOT/c

Intent.

# PUBLIC

## Interface

Content.
`
	testWriteNode(t, dir, logicalName, content)
	testChdir(t, dir)

	node, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public = nil; want non-nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("Public.Heading = %q; want %q", node.Public.Heading, "public")
	}
}

// TestPublicWithMixedCaseAndExtraWhitespace verifies that a heading like
// "#   PuBLiC" (mixed case, extra leading spaces) is normalized and
// recognized as the public section.
func TestPublicWithMixedCaseAndExtraWhitespace(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/d"
	const content = `---
---
# ROOT/d

Intent.

#   PuBLiC

## Interface

Content.
`
	testWriteNode(t, dir, logicalName, content)
	testChdir(t, dir)

	node, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public = nil; want non-nil")
	}
	if node.Public.Heading != "public" {
		t.Errorf("Public.Heading = %q; want %q", node.Public.Heading, "public")
	}
}

// TestNodeNameWithVariedWhitespace verifies that extra leading spaces in the
// node-name heading are stripped and the result is lowercased.
func TestNodeNameWithVariedWhitespace(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/e"
	const content = `---
---
#    ROOT/e

Intent.
`
	testWriteNode(t, dir, logicalName, content)
	testChdir(t, dir)

	node, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NameSection.Heading != "root/e" {
		t.Errorf("NameSection.Heading = %q; want %q", node.NameSection.Heading, "root/e")
	}
}

// TestSubsectionHeadingsAreNormalized verifies that subsection headings with
// extra whitespace or mixed case are lowercased and trimmed.
func TestSubsectionHeadingsAreNormalized(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/f"
	const content = `---
---
# ROOT/f

Intent.

# Public

##   Interface

Types.

## CONSTRAINTS

Rules.
`
	testWriteNode(t, dir, logicalName, content)
	testChdir(t, dir)

	node, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public = nil; want non-nil")
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("len(Public.Subsections) = %d; want 2", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("Subsections[0].Heading = %q; want %q", node.Public.Subsections[0].Heading, "interface")
	}
	if node.Public.Subsections[1].Heading != "constraints" {
		t.Errorf("Subsections[1].Heading = %q; want %q", node.Public.Subsections[1].Heading, "constraints")
	}
}

// TestTabCharactersInHeadingWhitespace verifies that tab characters
// surrounding the heading text are treated as whitespace and stripped.
func TestTabCharactersInHeadingWhitespace(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/g"
	// The subsection heading contains a tab before and after "Interface".
	content := "---\n---\n# ROOT/g\n\nIntent.\n\n# Public\n\n## \tInterface\t\n\nContent.\n"
	testWriteNode(t, dir, logicalName, content)
	testChdir(t, dir)

	node, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public = nil; want non-nil")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("len(Public.Subsections) = %d; want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Heading != "interface" {
		t.Errorf("Subsections[0].Heading = %q; want %q", node.Public.Subsections[0].Heading, "interface")
	}
}

// ---------------------------------------------------------------------------
// Content extraction tests
// ---------------------------------------------------------------------------

// TestLevel3AndDeeperHeadingsAreContent verifies that level-3 and deeper
// headings inside a subsection are treated as content, not structural
// delimiters.
func TestLevel3AndDeeperHeadingsAreContent(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/h"
	const content = `---
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
`
	testWriteNode(t, dir, logicalName, content)
	testChdir(t, dir)

	node, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public = nil; want non-nil")
	}
	if len(node.Public.Subsections) != 2 {
		t.Fatalf("len(Public.Subsections) = %d; want 2", len(node.Public.Subsections))
	}

	// The "interface" subsection must contain the ### Types heading and its
	// body plus the #### Nested detail heading and its body — all as raw
	// markdown text.
	interfaceContent := node.Public.Subsections[0].Content
	for _, want := range []string{"### Types", "Type definitions here.", "#### Nested detail", "Even deeper content."} {
		if !strings.Contains(interfaceContent, want) {
			t.Errorf("interface subsection content missing %q; content = %q", want, interfaceContent)
		}
	}

	// The "constraints" subsection must contain the ### Rule one heading and
	// its body.
	constraintsContent := node.Public.Subsections[1].Content
	for _, want := range []string{"### Rule one", "Details."} {
		if !strings.Contains(constraintsContent, want) {
			t.Errorf("constraints subsection content missing %q; content = %q", want, constraintsContent)
		}
	}
}

// TestFencedCodeBlocksWithHeadingLikeContent verifies that # and ## markers
// inside a fenced code block are not treated as headings.
func TestFencedCodeBlocksWithHeadingLikeContent(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/i"
	const content = `---
---
# ROOT/i

Intent.

# Public

## Interface

` + "~~~go\n// # This is not a heading\n// ## Neither is this\nfunc Foo() {}\n~~~" + `

After the code block.
`
	testWriteNode(t, dir, logicalName, content)
	testChdir(t, dir)

	node, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public = nil; want non-nil")
	}
	// The code-block markers inside the fenced block must NOT split subsections.
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("len(Public.Subsections) = %d; want 1 (# inside code block must not create new section)", len(node.Public.Subsections))
	}

	subsContent := node.Public.Subsections[0].Content
	if !strings.Contains(subsContent, "After the code block.") {
		t.Errorf("subsection content missing %q; content = %q", "After the code block.", subsContent)
	}
	if !strings.Contains(subsContent, "func Foo()") {
		t.Errorf("subsection content missing code block body; content = %q", subsContent)
	}
}

// TestContentBetweenSectionsIsTrimmed verifies that leading and trailing blank
// lines in section and subsection content are stripped.
func TestContentBetweenSectionsIsTrimmed(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/j"
	const content = `---
---
# ROOT/j

Intent.

# Public



Content with surrounding blank lines.



## Interface

Also surrounded.

`
	testWriteNode(t, dir, logicalName, content)
	testChdir(t, dir)

	node, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Public == nil {
		t.Fatal("Public = nil; want non-nil")
	}
	if node.Public.Content != "Content with surrounding blank lines." {
		t.Errorf("Public.Content = %q; want %q", node.Public.Content, "Content with surrounding blank lines.")
	}
	if len(node.Public.Subsections) != 1 {
		t.Fatalf("len(Public.Subsections) = %d; want 1", len(node.Public.Subsections))
	}
	if node.Public.Subsections[0].Content != "Also surrounded." {
		t.Errorf("Subsections[0].Content = %q; want %q", node.Public.Subsections[0].Content, "Also surrounded.")
	}
}

// ---------------------------------------------------------------------------
// Validation error tests
// ---------------------------------------------------------------------------

// TestFileDoesNotExist verifies that ParseNode returns ErrRead when the node
// file is absent.
func TestFileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// Do NOT write any file — ParseNode must fail with ErrRead.
	_, err := ParseNode("ROOT/nonexistent")
	if err == nil {
		t.Fatal("expected error; got nil")
	}
	if !errors.Is(err, ErrRead) {
		t.Errorf("errors.Is(err, ErrRead) = false; err = %v", err)
	}
}

// TestNoFrontmatterDelimiters verifies that a file without the opening ---
// delimiter returns ErrFrontmatterMissing.
func TestNoFrontmatterDelimiters(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/m"
	const content = `# ROOT/m

Just text.
`
	testWriteNode(t, dir, logicalName, content)
	testChdir(t, dir)

	_, err := ParseNode(logicalName)
	if err == nil {
		t.Fatal("expected error; got nil")
	}
	if !errors.Is(err, ErrFrontmatterMissing) {
		t.Errorf("errors.Is(err, ErrFrontmatterMissing) = false; err = %v", err)
	}
}

// TestContentBeforeFirstHeading verifies that non-heading content appearing
// before the first level-1 heading triggers ErrUnexpectedContent.
func TestContentBeforeFirstHeading(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/o"
	const content = `---
---
Some text before any heading.

# ROOT/o

Intent.
`
	testWriteNode(t, dir, logicalName, content)
	testChdir(t, dir)

	_, err := ParseNode(logicalName)
	if err == nil {
		t.Fatal("expected error; got nil")
	}
	if !errors.Is(err, ErrUnexpectedContent) {
		t.Errorf("errors.Is(err, ErrUnexpectedContent) = false; err = %v", err)
	}
}

// TestLevel2HeadingBeforeLevel1Heading verifies that a level-2 heading
// appearing before the first level-1 heading triggers ErrUnexpectedContent.
func TestLevel2HeadingBeforeLevel1Heading(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/p"
	const content = `---
---
## Orphan subsection

# ROOT/p

Intent.
`
	testWriteNode(t, dir, logicalName, content)
	testChdir(t, dir)

	_, err := ParseNode(logicalName)
	if err == nil {
		t.Fatal("expected error; got nil")
	}
	if !errors.Is(err, ErrUnexpectedContent) {
		t.Errorf("errors.Is(err, ErrUnexpectedContent) = false; err = %v", err)
	}
}

// TestNodeNameDoesNotMatchLogicalName verifies that a mismatch between the
// first level-1 heading and the expected logical name returns
// ErrInvalidNodeName.
func TestNodeNameDoesNotMatchLogicalName(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/q"
	const content = `---
---
# ROOT/wrong

Intent.
`
	testWriteNode(t, dir, logicalName, content)
	testChdir(t, dir)

	_, err := ParseNode(logicalName)
	if err == nil {
		t.Fatal("expected error; got nil")
	}
	if !errors.Is(err, ErrInvalidNodeName) {
		t.Errorf("errors.Is(err, ErrInvalidNodeName) = false; err = %v", err)
	}
}

// TestNodeNameCaseMismatchIsNotAnError verifies that a heading whose text
// differs only in case from the logical name is accepted (normalization makes
// them equal).
func TestNodeNameCaseMismatchIsNotAnError(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/q"
	// "root/Q" normalizes to "root/q", which equals the normalized logical
	// name "root/q".
	const content = `---
---
# root/Q

Intent.
`
	testWriteNode(t, dir, logicalName, content)
	testChdir(t, dir)

	_, err := ParseNode(logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestDuplicatePublicSectionSameCase verifies that two # Public sections with
// identical casing return ErrDuplicatePublic.
func TestDuplicatePublicSectionSameCase(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/r"
	const content = `---
---
# ROOT/r

Intent.

# Public

First public.

# Public

Second public.
`
	testWriteNode(t, dir, logicalName, content)
	testChdir(t, dir)

	_, err := ParseNode(logicalName)
	if err == nil {
		t.Fatal("expected error; got nil")
	}
	if !errors.Is(err, ErrDuplicatePublic) {
		t.Errorf("errors.Is(err, ErrDuplicatePublic) = false; err = %v", err)
	}
}

// TestDuplicatePublicSectionDifferentCase verifies that two public sections
// differing only in case (# Public vs # PUBLIC) return ErrDuplicatePublic.
func TestDuplicatePublicSectionDifferentCase(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/s"
	const content = `---
---
# ROOT/s

Intent.

# Public

First.

# PUBLIC

Second.
`
	testWriteNode(t, dir, logicalName, content)
	testChdir(t, dir)

	_, err := ParseNode(logicalName)
	if err == nil {
		t.Fatal("expected error; got nil")
	}
	if !errors.Is(err, ErrDuplicatePublic) {
		t.Errorf("errors.Is(err, ErrDuplicatePublic) = false; err = %v", err)
	}
}

// TestDuplicateSubsectionSameCase verifies that two ## Interface headings with
// identical casing under # Public return ErrDuplicateSubsection.
func TestDuplicateSubsectionSameCase(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/t"
	const content = `---
---
# ROOT/t

Intent.

# Public

## Interface

First interface.

## Interface

Second interface.
`
	testWriteNode(t, dir, logicalName, content)
	testChdir(t, dir)

	_, err := ParseNode(logicalName)
	if err == nil {
		t.Fatal("expected error; got nil")
	}
	if !errors.Is(err, ErrDuplicateSubsection) {
		t.Errorf("errors.Is(err, ErrDuplicateSubsection) = false; err = %v", err)
	}
}

// TestDuplicateSubsectionDifferentCase verifies that ## Interface and ##
// INTERFACE under # Public are considered duplicates after normalization.
func TestDuplicateSubsectionDifferentCase(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/u"
	const content = `---
---
# ROOT/u

Intent.

# Public

## Interface

First.

## INTERFACE

Second.
`
	testWriteNode(t, dir, logicalName, content)
	testChdir(t, dir)

	_, err := ParseNode(logicalName)
	if err == nil {
		t.Fatal("expected error; got nil")
	}
	if !errors.Is(err, ErrDuplicateSubsection) {
		t.Errorf("errors.Is(err, ErrDuplicateSubsection) = false; err = %v", err)
	}
}

// TestDuplicateSubsectionWhitespaceVariation verifies that ## Interface and
// ##   Interface (extra spaces) are considered duplicates after normalization.
func TestDuplicateSubsectionWhitespaceVariation(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/v"
	const content = `---
---
# ROOT/v

Intent.

# Public

## Interface

First.

##   Interface

Second.
`
	testWriteNode(t, dir, logicalName, content)
	testChdir(t, dir)

	_, err := ParseNode(logicalName)
	if err == nil {
		t.Fatal("expected error; got nil")
	}
	if !errors.Is(err, ErrDuplicateSubsection) {
		t.Errorf("errors.Is(err, ErrDuplicateSubsection) = false; err = %v", err)
	}
}

// TestFirstElementIsParagraphMissingNodeName verifies that a file whose body
// (after frontmatter) starts with a paragraph — rather than a level-1
// heading — triggers ErrUnexpectedContent.
func TestFirstElementIsParagraphMissingNodeName(t *testing.T) {
	dir := t.TempDir()
	const logicalName = "ROOT/w"
	const content = `---
---
This is a paragraph, not a heading.
`
	testWriteNode(t, dir, logicalName, content)
	testChdir(t, dir)

	_, err := ParseNode(logicalName)
	if err == nil {
		t.Fatal("expected error; got nil")
	}
	if !errors.Is(err, ErrUnexpectedContent) {
		t.Errorf("errors.Is(err, ErrUnexpectedContent) = false; err = %v", err)
	}
}
