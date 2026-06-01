// code-from-spec: ROOT/golang/tests/spec_tree/validate@ddW3dSvlnb10bCqnosUIxshB4_E
package spectreevalidate_test

import (
	"crypto/sha1"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectreevalidate"
)

// testChdir changes the working directory to dir for the duration of
// the test, restoring it on cleanup.
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

// testMakeEntry builds a SpecTreeValidateInput with the given logical name
// and minimal node structure. The heading is the path portion after "ROOT/"
// (lowercased), matching what NodeParse would produce.
func testMakeEntry(logicalName string, fm *frontmatter.Frontmatter, node *parsenode.Node) *spectreevalidate.SpecTreeValidateInput {
	return &spectreevalidate.SpecTreeValidateInput{
		LogicalName: logicalName,
		Frontmatter: fm,
		Node:        node,
	}
}

// testEmptyFM returns an empty Frontmatter.
func testEmptyFM() *frontmatter.Frontmatter {
	return &frontmatter.Frontmatter{
		DependsOn: []string{},
		External:  []*frontmatter.FrontmatterExternal{},
		Input:     "",
		Outputs:   []*frontmatter.FrontmatterOutput{},
	}
}

// testMakeNode builds a minimal Node with the given heading (normalized)
// and raw heading.
func testMakeNode(heading, rawHeading string) *parsenode.Node {
	return &parsenode.Node{
		NameSection: &parsenode.NodeSection{
			Heading:     heading,
			RawHeading:  rawHeading,
			Content:     []string{},
			Subsections: []*parsenode.NodeSubsection{},
		},
		Public:  nil,
		Agent:   nil,
		Private: []*parsenode.NodeSection{},
	}
}

// testHasError checks if the slice contains at least one error matching the
// given node and rule.
func testHasError(errs []*spectreevalidate.FormatError, node, rule string) bool {
	for _, e := range errs {
		if e.Node == node && e.Rule == rule {
			return true
		}
	}
	return false
}

// testCountErrors counts errors matching the given node and rule.
func testCountErrors(errs []*spectreevalidate.FormatError, node, rule string) int {
	count := 0
	for _, e := range errs {
		if e.Node == node && e.Rule == rule {
			count++
		}
	}
	return count
}

// testFragmentHash computes the SHA-1 base64url (no padding) hash of
// the given lines, each terminated by '\n'.
func testFragmentHash(lines []string) string {
	h := sha1.New()
	for _, line := range lines {
		h.Write([]byte(line + "\n"))
	}
	digest := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(digest)
}

// testWriteFile creates a file at the given relative path (from cwd),
// creating parent directories as needed.
func testWriteFile(t *testing.T, relPath, content string) {
	t.Helper()
	dir := filepath.Dir(relPath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("testWriteFile mkdir: %v", err)
		}
	}
	if err := os.WriteFile(relPath, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Happy Path
// ---------------------------------------------------------------------------

func TestValidLeafNodePassesAllChecks(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/b"},
			External:  []*frontmatter.FrontmatterExternal{},
			Input:     "",
			Outputs:   []*frontmatter.FrontmatterOutput{{ID: "out", Path: "internal/out.go"}},
		}, testMakeNode("root/a", "# ROOT/a")),
		testMakeEntry("ROOT/b", testEmptyFM(), testMakeNode("root/b", "# ROOT/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
	}
}

func TestValidIntermediateNodePassesAllChecks(t *testing.T) {
	rootNode := testMakeNode("root", "# ROOT")
	rootNode.Public = &parsenode.NodeSection{
		Heading:     "public",
		RawHeading:  "# Public",
		Content:     []string{},
		Subsections: []*parsenode.NodeSubsection{},
	}

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), rootNode),
		testMakeEntry("ROOT/a", testEmptyFM(), testMakeNode("root/a", "# ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
	}
}

func TestLeafWithNoFrontmatterFields(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", testEmptyFM(), testMakeNode("root/a", "# ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
	}
}

// ---------------------------------------------------------------------------
// Rule: name_heading
// ---------------------------------------------------------------------------

func TestNameHeadingMatches(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", testEmptyFM(), testMakeNode("root/a", "# ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	for _, e := range errs {
		if e.Rule == "name_heading" {
			t.Errorf("unexpected name_heading error: %+v", e)
		}
	}
}

func TestNameHeadingDoesNotMatch(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", testEmptyFM(), testMakeNode("root/wrong", "# ROOT/wrong")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "name_heading") {
		t.Errorf("expected name_heading error for ROOT/a, got: %+v", errs)
	}
}

// ---------------------------------------------------------------------------
// Rule: leaf_only_fields
// ---------------------------------------------------------------------------

func TestIntermediateNodeWithDependsOn(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/b"},
			External:  []*frontmatter.FrontmatterExternal{},
			Input:     "",
			Outputs:   []*frontmatter.FrontmatterOutput{},
		}, testMakeNode("root/a", "# ROOT/a")),
		testMakeEntry("ROOT/a/b", testEmptyFM(), testMakeNode("root/a/b", "# ROOT/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected leaf_only_fields error for ROOT/a, got: %+v", errs)
	}
}

func TestIntermediateNodeWithOutputs(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{},
			External:  []*frontmatter.FrontmatterExternal{},
			Input:     "",
			Outputs:   []*frontmatter.FrontmatterOutput{{ID: "x", Path: "x.go"}},
		}, testMakeNode("root/a", "# ROOT/a")),
		testMakeEntry("ROOT/a/b", testEmptyFM(), testMakeNode("root/a/b", "# ROOT/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected leaf_only_fields error for ROOT/a, got: %+v", errs)
	}
}

func TestIntermediateNodeWithInput(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{},
			External:  []*frontmatter.FrontmatterExternal{},
			Input:     "ARTIFACT/c(id)",
			Outputs:   []*frontmatter.FrontmatterOutput{},
		}, testMakeNode("root/a", "# ROOT/a")),
		testMakeEntry("ROOT/a/b", testEmptyFM(), testMakeNode("root/a/b", "# ROOT/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected leaf_only_fields error for ROOT/a, got: %+v", errs)
	}
}

func TestIntermediateNodeWithExternal(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{},
			External: []*frontmatter.FrontmatterExternal{
				{Path: "some/file.txt", Fragments: []*frontmatter.FrontmatterExternalFragment{}},
			},
			Input:   "",
			Outputs: []*frontmatter.FrontmatterOutput{},
		}, testMakeNode("root/a", "# ROOT/a")),
		testMakeEntry("ROOT/a/b", testEmptyFM(), testMakeNode("root/a/b", "# ROOT/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected leaf_only_fields error for ROOT/a, got: %+v", errs)
	}
}

func TestIntermediateNodeWithMultipleRestrictedFields(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/b"},
			External:  []*frontmatter.FrontmatterExternal{},
			Input:     "",
			Outputs:   []*frontmatter.FrontmatterOutput{{ID: "x", Path: "x.go"}},
		}, testMakeNode("root/a", "# ROOT/a")),
		testMakeEntry("ROOT/a/b", testEmptyFM(), testMakeNode("root/a/b", "# ROOT/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	count := testCountErrors(errs, "ROOT/a", "leaf_only_fields")
	if count != 2 {
		t.Errorf("expected exactly 2 leaf_only_fields errors for ROOT/a, got %d: %+v", count, errs)
	}
}

// ---------------------------------------------------------------------------
// Rule: leaf_only_agent
// ---------------------------------------------------------------------------

func TestIntermediateNodeWithAgentSection(t *testing.T) {
	nodeA := testMakeNode("root/a", "# ROOT/a")
	nodeA.Agent = &parsenode.NodeSection{
		Heading:     "agent",
		RawHeading:  "# Agent",
		Content:     []string{"Agent instructions."},
		Subsections: []*parsenode.NodeSubsection{},
	}

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", testEmptyFM(), nodeA),
		testMakeEntry("ROOT/a/b", testEmptyFM(), testMakeNode("root/a/b", "# ROOT/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_agent") {
		t.Errorf("expected leaf_only_agent error for ROOT/a, got: %+v", errs)
	}
}

func TestLeafNodeWithAgentSectionNoError(t *testing.T) {
	nodeA := testMakeNode("root/a", "# ROOT/a")
	nodeA.Agent = &parsenode.NodeSection{
		Heading:     "agent",
		RawHeading:  "# Agent",
		Content:     []string{"Agent instructions."},
		Subsections: []*parsenode.NodeSubsection{},
	}

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", testEmptyFM(), nodeA),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	for _, e := range errs {
		if e.Rule == "leaf_only_agent" {
			t.Errorf("unexpected leaf_only_agent error: %+v", e)
		}
	}
}

// ---------------------------------------------------------------------------
// Rule: dependency_targets
// ---------------------------------------------------------------------------

func TestDependsOnTargetsNonExistentRootNode(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/missing"},
			External:  []*frontmatter.FrontmatterExternal{},
			Input:     "",
			Outputs:   []*frontmatter.FrontmatterOutput{},
		}, testMakeNode("root/a", "# ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a, got: %+v", errs)
	}
}

func TestDependsOnTargetsAncestor(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", testEmptyFM(), testMakeNode("root/a", "# ROOT/a")),
		testMakeEntry("ROOT/a/b", &frontmatter.Frontmatter{
			DependsOn: []string{"ROOT"},
			External:  []*frontmatter.FrontmatterExternal{},
			Input:     "",
			Outputs:   []*frontmatter.FrontmatterOutput{},
		}, testMakeNode("root/a/b", "# ROOT/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a/b", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a/b, got: %+v", errs)
	}
}

func TestDependsOnTargetsDescendant(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/a/b"},
			External:  []*frontmatter.FrontmatterExternal{},
			Input:     "",
			Outputs:   []*frontmatter.FrontmatterOutput{},
		}, testMakeNode("root/a", "# ROOT/a")),
		testMakeEntry("ROOT/a/b", testEmptyFM(), testMakeNode("root/a/b", "# ROOT/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a, got: %+v", errs)
	}
}

func TestDependsOnTargetsSelf(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/a"},
			External:  []*frontmatter.FrontmatterExternal{},
			Input:     "",
			Outputs:   []*frontmatter.FrontmatterOutput{},
		}, testMakeNode("root/a", "# ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a, got: %+v", errs)
	}
}

func TestDependsOnWithValidROOTQualifier(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", testEmptyFM(), testMakeNode("root/a", "# ROOT/a")),
		testMakeEntry("ROOT/b", &frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/a(interface)"},
			External:  []*frontmatter.FrontmatterExternal{},
			Input:     "",
			Outputs:   []*frontmatter.FrontmatterOutput{},
		}, testMakeNode("root/b", "# ROOT/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	for _, e := range errs {
		if e.Rule == "dependency_targets" {
			t.Errorf("unexpected dependency_targets error: %+v", e)
		}
	}
}

func TestDependsOnWithValidARTIFACTReference(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{},
			External:  []*frontmatter.FrontmatterExternal{},
			Input:     "",
			Outputs:   []*frontmatter.FrontmatterOutput{{ID: "lib", Path: "lib.go"}},
		}, testMakeNode("root/a", "# ROOT/a")),
		testMakeEntry("ROOT/b", &frontmatter.Frontmatter{
			DependsOn: []string{"ARTIFACT/a(lib)"},
			External:  []*frontmatter.FrontmatterExternal{},
			Input:     "",
			Outputs:   []*frontmatter.FrontmatterOutput{},
		}, testMakeNode("root/b", "# ROOT/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	for _, e := range errs {
		if e.Rule == "dependency_targets" {
			t.Errorf("unexpected dependency_targets error: %+v", e)
		}
	}
}

func TestDependsOnWithNonExistentARTIFACTReference(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{"ARTIFACT/missing(id)"},
			External:  []*frontmatter.FrontmatterExternal{},
			Input:     "",
			Outputs:   []*frontmatter.FrontmatterOutput{},
		}, testMakeNode("root/a", "# ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a, got: %+v", errs)
	}
}

func TestMultipleInvalidDependsOnOneErrorPerEntry(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/missing", "ROOT/also_missing"},
			External:  []*frontmatter.FrontmatterExternal{},
			Input:     "",
			Outputs:   []*frontmatter.FrontmatterOutput{},
		}, testMakeNode("root/a", "# ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	count := testCountErrors(errs, "ROOT/a", "dependency_targets")
	if count != 2 {
		t.Errorf("expected exactly 2 dependency_targets errors for ROOT/a, got %d: %+v", count, errs)
	}
}

// ---------------------------------------------------------------------------
// Rule: input_target
// ---------------------------------------------------------------------------

func TestValidInputReference(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{},
			External:  []*frontmatter.FrontmatterExternal{},
			Input:     "",
			Outputs:   []*frontmatter.FrontmatterOutput{{ID: "out", Path: "a.go"}},
		}, testMakeNode("root/a", "# ROOT/a")),
		testMakeEntry("ROOT/b", &frontmatter.Frontmatter{
			DependsOn: []string{},
			External:  []*frontmatter.FrontmatterExternal{},
			Input:     "ARTIFACT/a(out)",
			Outputs:   []*frontmatter.FrontmatterOutput{},
		}, testMakeNode("root/b", "# ROOT/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	for _, e := range errs {
		if e.Rule == "input_target" {
			t.Errorf("unexpected input_target error: %+v", e)
		}
	}
}

func TestInputNotStartingWithARTIFACT(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{},
			External:  []*frontmatter.FrontmatterExternal{},
			Input:     "ROOT/something",
			Outputs:   []*frontmatter.FrontmatterOutput{},
		}, testMakeNode("root/a", "# ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "input_target") {
		t.Errorf("expected input_target error for ROOT/a, got: %+v", errs)
	}
}

func TestInputReferencesNonExistentArtifact(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{},
			External:  []*frontmatter.FrontmatterExternal{},
			Input:     "ARTIFACT/missing(id)",
			Outputs:   []*frontmatter.FrontmatterOutput{},
		}, testMakeNode("root/a", "# ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "input_target") {
		t.Errorf("expected input_target error for ROOT/a, got: %+v", errs)
	}
}

// ---------------------------------------------------------------------------
// Rule: external_files
// ---------------------------------------------------------------------------

func TestExternalFileExistsNoFragments(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "some/file.txt", "hello\n")

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{},
			External: []*frontmatter.FrontmatterExternal{
				{Path: "some/file.txt", Fragments: []*frontmatter.FrontmatterExternalFragment{}},
			},
			Input:   "",
			Outputs: []*frontmatter.FrontmatterOutput{},
		}, testMakeNode("root/a", "# ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	for _, e := range errs {
		if e.Rule == "external_files" {
			t.Errorf("unexpected external_files error: %+v", e)
		}
	}
}

func TestExternalFileDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	// Do not create "nonexistent.txt"

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{},
			External: []*frontmatter.FrontmatterExternal{
				{Path: "nonexistent.txt", Fragments: []*frontmatter.FrontmatterExternalFragment{}},
			},
			Input:   "",
			Outputs: []*frontmatter.FrontmatterOutput{},
		}, testMakeNode("root/a", "# ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("expected external_files error for ROOT/a, got: %+v", errs)
	}
}

func TestFragmentWithValidHash(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	fileContent := "alpha\nbravo\ncharlie\ndelta\necho\n"
	testWriteFile(t, "f.txt", fileContent)

	// Compute the correct hash for lines 1-3 (alpha, bravo, charlie).
	correctHash := testFragmentHash([]string{"alpha", "bravo", "charlie"})

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{},
			External: []*frontmatter.FrontmatterExternal{
				{
					Path: "f.txt",
					Fragments: []*frontmatter.FrontmatterExternalFragment{
						{Lines: "1-3", Hash: correctHash},
					},
				},
			},
			Input:   "",
			Outputs: []*frontmatter.FrontmatterOutput{},
		}, testMakeNode("root/a", "# ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	for _, e := range errs {
		if e.Rule == "external_files" {
			t.Errorf("unexpected external_files error: %+v", e)
		}
	}
}

func TestFragmentWithInvalidHash(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	fileContent := "alpha\nbravo\ncharlie\ndelta\necho\n"
	testWriteFile(t, "f.txt", fileContent)

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{},
			External: []*frontmatter.FrontmatterExternal{
				{
					Path: "f.txt",
					Fragments: []*frontmatter.FrontmatterExternalFragment{
						{Lines: "1-3", Hash: "wrong_______________________"},
					},
				},
			},
			Input:   "",
			Outputs: []*frontmatter.FrontmatterOutput{},
		}, testMakeNode("root/a", "# ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("expected external_files error for ROOT/a, got: %+v", errs)
	}
}

func TestFragmentWithInvalidRangeFormat(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "f.txt", "hello\n")

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{},
			External: []*frontmatter.FrontmatterExternal{
				{
					Path: "f.txt",
					Fragments: []*frontmatter.FrontmatterExternalFragment{
						{Lines: "abc", Hash: "x"},
					},
				},
			},
			Input:   "",
			Outputs: []*frontmatter.FrontmatterOutput{},
		}, testMakeNode("root/a", "# ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("expected external_files error for ROOT/a, got: %+v", errs)
	}
}

func TestFragmentWithStartGreaterThanEnd(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	fileContent := "alpha\nbravo\ncharlie\ndelta\necho\n"
	testWriteFile(t, "f.txt", fileContent)

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{},
			External: []*frontmatter.FrontmatterExternal{
				{
					Path: "f.txt",
					Fragments: []*frontmatter.FrontmatterExternalFragment{
						{Lines: "5-3", Hash: "x"},
					},
				},
			},
			Input:   "",
			Outputs: []*frontmatter.FrontmatterOutput{},
		}, testMakeNode("root/a", "# ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("expected external_files error for ROOT/a, got: %+v", errs)
	}
}

func TestFragmentWithStartLessThanOne(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	fileContent := "alpha\nbravo\ncharlie\ndelta\necho\n"
	testWriteFile(t, "f.txt", fileContent)

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{},
			External: []*frontmatter.FrontmatterExternal{
				{
					Path: "f.txt",
					Fragments: []*frontmatter.FrontmatterExternalFragment{
						{Lines: "0-3", Hash: "x"},
					},
				},
			},
			Input:   "",
			Outputs: []*frontmatter.FrontmatterOutput{},
		}, testMakeNode("root/a", "# ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("expected external_files error for ROOT/a, got: %+v", errs)
	}
}

func TestFragmentOutOfRange(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	fileContent := "alpha\nbravo\ncharlie\ndelta\necho\n"
	testWriteFile(t, "f.txt", fileContent)

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{},
			External: []*frontmatter.FrontmatterExternal{
				{
					Path: "f.txt",
					Fragments: []*frontmatter.FrontmatterExternalFragment{
						{Lines: "1-100", Hash: "x"},
					},
				},
			},
			Input:   "",
			Outputs: []*frontmatter.FrontmatterOutput{},
		}, testMakeNode("root/a", "# ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("expected external_files error for ROOT/a, got: %+v", errs)
	}
}

// ---------------------------------------------------------------------------
// Rule: output_paths
// ---------------------------------------------------------------------------

func TestValidOutputPath(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{},
			External:  []*frontmatter.FrontmatterExternal{},
			Input:     "",
			Outputs:   []*frontmatter.FrontmatterOutput{{ID: "x", Path: "internal/x.go"}},
		}, testMakeNode("root/a", "# ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	for _, e := range errs {
		if e.Rule == "output_paths" {
			t.Errorf("unexpected output_paths error: %+v", e)
		}
	}
}

func TestOutputPathWithTraversal(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{},
			External:  []*frontmatter.FrontmatterExternal{},
			Input:     "",
			Outputs:   []*frontmatter.FrontmatterOutput{{ID: "x", Path: "../../etc/passwd"}},
		}, testMakeNode("root/a", "# ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "output_paths") {
		t.Errorf("expected output_paths error for ROOT/a, got: %+v", errs)
	}
}

func TestOutputPathWithBackslash(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{},
			External:  []*frontmatter.FrontmatterExternal{},
			Input:     "",
			Outputs:   []*frontmatter.FrontmatterOutput{{ID: "x", Path: `internal\x.go`}},
		}, testMakeNode("root/a", "# ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "output_paths") {
		t.Errorf("expected output_paths error for ROOT/a, got: %+v", errs)
	}
}

// ---------------------------------------------------------------------------
// Rule: duplicate_subsections
// ---------------------------------------------------------------------------

func TestUniqueSubsectionHeadingsNoError(t *testing.T) {
	nodeA := testMakeNode("root/a", "# ROOT/a")
	nodeA.Public = &parsenode.NodeSection{
		Heading:    "public",
		RawHeading: "# Public",
		Content:    []string{},
		Subsections: []*parsenode.NodeSubsection{
			{Heading: "interface", RawHeading: "## Interface", Content: []string{"Types."}},
			{Heading: "context", RawHeading: "## Context", Content: []string{"Background."}},
		},
	}

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", testEmptyFM(), nodeA),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	for _, e := range errs {
		if e.Rule == "duplicate_subsections" {
			t.Errorf("unexpected duplicate_subsections error: %+v", e)
		}
	}
}

func TestDuplicateSubsectionHeadings(t *testing.T) {
	nodeA := testMakeNode("root/a", "# ROOT/a")
	nodeA.Public = &parsenode.NodeSection{
		Heading:    "public",
		RawHeading: "# Public",
		Content:    []string{},
		Subsections: []*parsenode.NodeSubsection{
			{Heading: "interface", RawHeading: "## Interface", Content: []string{"First."}},
			{Heading: "interface", RawHeading: "## Interface", Content: []string{"Second."}},
		},
	}

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", testEmptyFM(), nodeA),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	count := testCountErrors(errs, "ROOT/a", "duplicate_subsections")
	if count != 1 {
		t.Errorf("expected exactly 1 duplicate_subsections error for ROOT/a, got %d: %+v", count, errs)
	}
}

func TestThreeIdenticalSubsectionHeadings(t *testing.T) {
	nodeA := testMakeNode("root/a", "# ROOT/a")
	nodeA.Public = &parsenode.NodeSection{
		Heading:    "public",
		RawHeading: "# Public",
		Content:    []string{},
		Subsections: []*parsenode.NodeSubsection{
			{Heading: "interface", RawHeading: "## Interface", Content: []string{"First."}},
			{Heading: "interface", RawHeading: "## Interface", Content: []string{"Second."}},
			{Heading: "interface", RawHeading: "## Interface", Content: []string{"Third."}},
		},
	}

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", testEmptyFM(), nodeA),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	count := testCountErrors(errs, "ROOT/a", "duplicate_subsections")
	if count != 2 {
		t.Errorf("expected exactly 2 duplicate_subsections errors for ROOT/a, got %d: %+v", count, errs)
	}
}

func TestNoPublicSectionSkipDuplicateSubsections(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", testEmptyFM(), testMakeNode("root/a", "# ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	for _, e := range errs {
		if e.Rule == "duplicate_subsections" {
			t.Errorf("unexpected duplicate_subsections error: %+v", e)
		}
	}
}

// ---------------------------------------------------------------------------
// Cross-Cutting
// ---------------------------------------------------------------------------

func TestCollectsMultipleErrorsFromDifferentRules(t *testing.T) {
	nodeA := testMakeNode("root/wrong", "# ROOT/wrong")
	nodeA.Public = &parsenode.NodeSection{
		Heading:    "public",
		RawHeading: "# Public",
		Content:    []string{},
		Subsections: []*parsenode.NodeSubsection{
			{Heading: "interface", RawHeading: "## Interface", Content: []string{"First."}},
			{Heading: "interface", RawHeading: "## Interface", Content: []string{"Second."}},
		},
	}

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", testEmptyFM(), testMakeNode("root", "# ROOT")),
		testMakeEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/missing"},
			External:  []*frontmatter.FrontmatterExternal{},
			Input:     "",
			Outputs:   []*frontmatter.FrontmatterOutput{},
		}, nodeA),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if !testHasError(errs, "ROOT/a", "name_heading") {
		t.Errorf("expected name_heading error for ROOT/a, got: %+v", errs)
	}
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a, got: %+v", errs)
	}
	if !testHasError(errs, "ROOT/a", "duplicate_subsections") {
		t.Errorf("expected duplicate_subsections error for ROOT/a, got: %+v", errs)
	}
}

func TestEmptyInputList(t *testing.T) {
	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{})
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty input, got %d: %+v", len(errs), errs)
	}
}
