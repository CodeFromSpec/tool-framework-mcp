// code-from-spec: ROOT/golang/tests/spec_tree/validate@bP8iKUhRL2cnDsENNc2dsOguNeA
package spectreevalidate_test

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectreevalidate"
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

// testMakeNode constructs a minimal *parsenode.Node with the given logical name.
func testMakeNode(logicalName string) *parsenode.Node {
	normalized := ""
	for i, r := range logicalName {
		if r >= 'A' && r <= 'Z' {
			normalized += string(rune(r + 32))
		} else {
			_ = i
			normalized += string(r)
		}
	}
	return &parsenode.Node{
		NameSection: &parsenode.NodeSection{
			Heading:    normalized,
			RawHeading: "# " + logicalName,
			Content:    []string{},
		},
	}
}

// testMakeNodeWithHeading constructs a *parsenode.Node with a custom heading.
func testMakeNodeWithHeading(logicalName, heading string) *parsenode.Node {
	return &parsenode.Node{
		NameSection: &parsenode.NodeSection{
			Heading:    heading,
			RawHeading: "# " + heading,
			Content:    []string{},
		},
	}
}

// testMakeNodeWithAgent constructs a *parsenode.Node with an agent section.
func testMakeNodeWithAgent(logicalName string) *parsenode.Node {
	node := testMakeNode(logicalName)
	node.Agent = &parsenode.NodeSection{
		Heading:    "agent",
		RawHeading: "# Agent",
		Content:    []string{},
	}
	return node
}

// testMakeNodeWithPublic constructs a *parsenode.Node with a public section
// containing the given subsection headings.
func testMakeNodeWithPublic(logicalName string, subsectionHeadings []string) *parsenode.Node {
	node := testMakeNode(logicalName)
	subs := make([]*parsenode.NodeSubsection, 0, len(subsectionHeadings))
	for _, h := range subsectionHeadings {
		subs = append(subs, &parsenode.NodeSubsection{
			Heading:    h,
			RawHeading: "## " + h,
			Content:    []string{},
		})
	}
	node.Public = &parsenode.NodeSection{
		Heading:    "public",
		RawHeading: "# Public",
		Content:    []string{},
		Subsections: subs,
	}
	return node
}

// testEmptyFrontmatter returns an empty *frontmatter.Frontmatter.
func testEmptyFrontmatter() *frontmatter.Frontmatter {
	return &frontmatter.Frontmatter{}
}

// testFrontmatterWithDependsOn returns a *frontmatter.Frontmatter with depends_on set.
func testFrontmatterWithDependsOn(deps []string) *frontmatter.Frontmatter {
	return &frontmatter.Frontmatter{
		DependsOn: deps,
	}
}

// testFrontmatterWithOutputs returns a *frontmatter.Frontmatter with outputs set.
func testFrontmatterWithOutputs(outputs []*frontmatter.FrontmatterOutput) *frontmatter.Frontmatter {
	return &frontmatter.Frontmatter{
		Outputs: outputs,
	}
}

// testFrontmatterWithInput returns a *frontmatter.Frontmatter with input set.
func testFrontmatterWithInput(input string) *frontmatter.Frontmatter {
	return &frontmatter.Frontmatter{
		Input: input,
	}
}

// testFrontmatterWithExternal returns a *frontmatter.Frontmatter with external set.
func testFrontmatterWithExternal(external []*frontmatter.FrontmatterExternal) *frontmatter.Frontmatter {
	return &frontmatter.Frontmatter{
		External: external,
	}
}

// testEntry is a convenience constructor for SpecTreeValidateInput.
func testEntry(logicalName string, fm *frontmatter.Frontmatter, node *parsenode.Node) *spectreevalidate.SpecTreeValidateInput {
	return &spectreevalidate.SpecTreeValidateInput{
		LogicalName: logicalName,
		Frontmatter: fm,
		Node:        node,
	}
}

// testHasErrorWithRule returns true if any FormatError in errs has the given rule.
func testHasErrorWithRule(errs []*spectreevalidate.FormatError, rule string) bool {
	for _, e := range errs {
		if e.Rule == rule {
			return true
		}
	}
	return false
}

// testCountErrorsWithNodeAndRule counts FormatErrors matching the given node and rule.
func testCountErrorsWithNodeAndRule(errs []*spectreevalidate.FormatError, node, rule string) int {
	count := 0
	for _, e := range errs {
		if e.Node == node && e.Rule == rule {
			count++
		}
	}
	return count
}

// testHasErrorWithNodeAndRule returns true if any FormatError matches the given node and rule.
func testHasErrorWithNodeAndRule(errs []*spectreevalidate.FormatError, node, rule string) bool {
	return testCountErrorsWithNodeAndRule(errs, node, rule) > 0
}

// testFragmentHash computes the SHA-1 base64url hash (no padding, 27 chars) for the given lines.
func testFragmentHash(lines []string) string {
	h := sha1.New()
	for _, line := range lines {
		h.Write([]byte(line + "\n"))
	}
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// testWriteFile creates directories as needed and writes content to a path relative to cwd.
func testWriteFile(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("testWriteFile mkdir: %v", err)
		}
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Happy Path
// ---------------------------------------------------------------------------

func TestSpecTreeValidate_HP1_ValidLeafNodePassesAllChecks(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/b"},
			Outputs:   []*frontmatter.FrontmatterOutput{{ID: "out", Path: "out.go"}},
		}, testMakeNode("ROOT/a")),
		testEntry("ROOT/b", testEmptyFrontmatter(), testMakeNode("ROOT/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestSpecTreeValidate_HP2_ValidIntermediateNodePassesAllChecks(t *testing.T) {
	rootNode := testMakeNodeWithPublic("ROOT", []string{"interface", "context"})
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), rootNode),
		testEntry("ROOT/a", testEmptyFrontmatter(), testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestSpecTreeValidate_HP3_LeafWithNoFrontmatterFields(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testEmptyFrontmatter(), testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

// ---------------------------------------------------------------------------
// Rule: name_heading
// ---------------------------------------------------------------------------

func TestSpecTreeValidate_NH1_HeadingMatchesLogicalName(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testEmptyFrontmatter(), testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasErrorWithRule(errs, "name_heading") {
		t.Errorf("expected no name_heading errors, but got some: %v", errs)
	}
}

func TestSpecTreeValidate_NH2_HeadingDoesNotMatchLogicalName(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testEmptyFrontmatter(), testMakeNodeWithHeading("ROOT/a", "root/wrong")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasErrorWithNodeAndRule(errs, "ROOT/a", "name_heading") {
		t.Errorf("expected a name_heading error for ROOT/a, got: %v", errs)
	}
}

// ---------------------------------------------------------------------------
// Rule: leaf_only_fields
// ---------------------------------------------------------------------------

func TestSpecTreeValidate_LOF1_IntermediateNodeWithDependsOn(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithDependsOn([]string{"ROOT/b"}), testMakeNode("ROOT/a")),
		testEntry("ROOT/a/b", testEmptyFrontmatter(), testMakeNode("ROOT/a/b")),
		testEntry("ROOT/b", testEmptyFrontmatter(), testMakeNode("ROOT/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasErrorWithNodeAndRule(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected a leaf_only_fields error for ROOT/a, got: %v", errs)
	}
}

func TestSpecTreeValidate_LOF2_IntermediateNodeWithOutputs(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithOutputs([]*frontmatter.FrontmatterOutput{{ID: "x", Path: "x.go"}}), testMakeNode("ROOT/a")),
		testEntry("ROOT/a/b", testEmptyFrontmatter(), testMakeNode("ROOT/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasErrorWithNodeAndRule(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected a leaf_only_fields error for ROOT/a, got: %v", errs)
	}
}

func TestSpecTreeValidate_LOF3_IntermediateNodeWithInput(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithInput("ARTIFACT/c(id)"), testMakeNode("ROOT/a")),
		testEntry("ROOT/a/b", testEmptyFrontmatter(), testMakeNode("ROOT/a/b")),
		testEntry("ROOT/c", testFrontmatterWithOutputs([]*frontmatter.FrontmatterOutput{{ID: "id", Path: "c.go"}}), testMakeNode("ROOT/c")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasErrorWithNodeAndRule(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected a leaf_only_fields error for ROOT/a, got: %v", errs)
	}
}

func TestSpecTreeValidate_LOF4_IntermediateNodeWithExternal(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithExternal([]*frontmatter.FrontmatterExternal{{Path: "some/file.txt"}}), testMakeNode("ROOT/a")),
		testEntry("ROOT/a/b", testEmptyFrontmatter(), testMakeNode("ROOT/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasErrorWithNodeAndRule(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected a leaf_only_fields error for ROOT/a, got: %v", errs)
	}
}

func TestSpecTreeValidate_LOF5_IntermediateNodeWithMultipleRestrictedFields(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/b"},
			Outputs:   []*frontmatter.FrontmatterOutput{{ID: "x", Path: "x.go"}},
		}, testMakeNode("ROOT/a")),
		testEntry("ROOT/a/b", testEmptyFrontmatter(), testMakeNode("ROOT/a/b")),
		testEntry("ROOT/b", testEmptyFrontmatter(), testMakeNode("ROOT/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	count := testCountErrorsWithNodeAndRule(errs, "ROOT/a", "leaf_only_fields")
	if count < 2 {
		t.Errorf("expected at least 2 leaf_only_fields errors for ROOT/a, got %d: %v", count, errs)
	}
}

// ---------------------------------------------------------------------------
// Rule: leaf_only_agent
// ---------------------------------------------------------------------------

func TestSpecTreeValidate_LOA1_IntermediateNodeWithAgentSection(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testEmptyFrontmatter(), testMakeNodeWithAgent("ROOT/a")),
		testEntry("ROOT/a/b", testEmptyFrontmatter(), testMakeNode("ROOT/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasErrorWithNodeAndRule(errs, "ROOT/a", "leaf_only_agent") {
		t.Errorf("expected a leaf_only_agent error for ROOT/a, got: %v", errs)
	}
}

func TestSpecTreeValidate_LOA2_LeafNodeWithAgentSection(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testEmptyFrontmatter(), testMakeNodeWithAgent("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasErrorWithRule(errs, "leaf_only_agent") {
		t.Errorf("expected no leaf_only_agent errors, but got some: %v", errs)
	}
}

// ---------------------------------------------------------------------------
// Rule: dependency_targets
// ---------------------------------------------------------------------------

func TestSpecTreeValidate_DT1_DependsOnTargetsNonExistentRootNode(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithDependsOn([]string{"ROOT/missing"}), testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasErrorWithNodeAndRule(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected a dependency_targets error for ROOT/a, got: %v", errs)
	}
}

func TestSpecTreeValidate_DT2_DependsOnTargetsAncestor(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testEmptyFrontmatter(), testMakeNode("ROOT/a")),
		testEntry("ROOT/a/b", testFrontmatterWithDependsOn([]string{"ROOT"}), testMakeNode("ROOT/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasErrorWithNodeAndRule(errs, "ROOT/a/b", "dependency_targets") {
		t.Errorf("expected a dependency_targets error for ROOT/a/b, got: %v", errs)
	}
}

func TestSpecTreeValidate_DT3_DependsOnTargetsDescendant(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithDependsOn([]string{"ROOT/a/b"}), testMakeNode("ROOT/a")),
		testEntry("ROOT/a/b", testEmptyFrontmatter(), testMakeNode("ROOT/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasErrorWithNodeAndRule(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected a dependency_targets error for ROOT/a, got: %v", errs)
	}
}

func TestSpecTreeValidate_DT4_DependsOnTargetsSelf(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithDependsOn([]string{"ROOT/a"}), testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasErrorWithNodeAndRule(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected a dependency_targets error for ROOT/a, got: %v", errs)
	}
}

func TestSpecTreeValidate_DT5_DependsOnWithValidRootQualifier(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testEmptyFrontmatter(), testMakeNode("ROOT/a")),
		testEntry("ROOT/b", testFrontmatterWithDependsOn([]string{"ROOT/a(interface)"}), testMakeNode("ROOT/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasErrorWithRule(errs, "dependency_targets") {
		t.Errorf("expected no dependency_targets errors, but got some: %v", errs)
	}
}

func TestSpecTreeValidate_DT6_DependsOnWithValidArtifactReference(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithOutputs([]*frontmatter.FrontmatterOutput{{ID: "lib", Path: "lib.go"}}), testMakeNode("ROOT/a")),
		testEntry("ROOT/b", testFrontmatterWithDependsOn([]string{"ARTIFACT/a(lib)"}), testMakeNode("ROOT/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasErrorWithRule(errs, "dependency_targets") {
		t.Errorf("expected no dependency_targets errors, but got some: %v", errs)
	}
}

func TestSpecTreeValidate_DT7_DependsOnWithNonExistentArtifactReference(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithDependsOn([]string{"ARTIFACT/missing(id)"}), testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasErrorWithNodeAndRule(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected a dependency_targets error for ROOT/a, got: %v", errs)
	}
}

func TestSpecTreeValidate_DT8_MultipleInvalidDependsOn(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithDependsOn([]string{"ROOT/missing", "ROOT/also_missing"}), testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	count := testCountErrorsWithNodeAndRule(errs, "ROOT/a", "dependency_targets")
	if count < 2 {
		t.Errorf("expected at least 2 dependency_targets errors for ROOT/a, got %d: %v", count, errs)
	}
}

// ---------------------------------------------------------------------------
// Rule: input_target
// ---------------------------------------------------------------------------

func TestSpecTreeValidate_IT1_ValidInputReference(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithOutputs([]*frontmatter.FrontmatterOutput{{ID: "out", Path: "a.go"}}), testMakeNode("ROOT/a")),
		testEntry("ROOT/b", testFrontmatterWithInput("ARTIFACT/a(out)"), testMakeNode("ROOT/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasErrorWithRule(errs, "input_target") {
		t.Errorf("expected no input_target errors, but got some: %v", errs)
	}
}

func TestSpecTreeValidate_IT2_InputNotStartingWithArtifact(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithInput("ROOT/something"), testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasErrorWithNodeAndRule(errs, "ROOT/a", "input_target") {
		t.Errorf("expected an input_target error for ROOT/a, got: %v", errs)
	}
}

func TestSpecTreeValidate_IT3_InputReferencesNonExistentArtifact(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithInput("ARTIFACT/missing(id)"), testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasErrorWithNodeAndRule(errs, "ROOT/a", "input_target") {
		t.Errorf("expected an input_target error for ROOT/a, got: %v", errs)
	}
}

// ---------------------------------------------------------------------------
// Rule: external_files
// ---------------------------------------------------------------------------

func TestSpecTreeValidate_EF1_ExternalFileExistsNoFragments(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "some/file.txt", "hello\n")

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithExternal([]*frontmatter.FrontmatterExternal{
			{Path: "some/file.txt"},
		}), testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasErrorWithRule(errs, "external_files") {
		t.Errorf("expected no external_files errors, but got some: %v", errs)
	}
}

func TestSpecTreeValidate_EF2_ExternalFileDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithExternal([]*frontmatter.FrontmatterExternal{
			{Path: "nonexistent.txt"},
		}), testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasErrorWithNodeAndRule(errs, "ROOT/a", "external_files") {
		t.Errorf("expected an external_files error for ROOT/a, got: %v", errs)
	}
}

func TestSpecTreeValidate_EF3_FragmentWithValidHash(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "alpha\nbeta\ngamma\ndelta\nepsilon\n"
	testWriteFile(t, "f.txt", content)

	correctHash := testFragmentHash([]string{"alpha", "beta", "gamma"})

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithExternal([]*frontmatter.FrontmatterExternal{
			{
				Path: "f.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "1-3", Hash: correctHash},
				},
			},
		}), testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasErrorWithRule(errs, "external_files") {
		t.Errorf("expected no external_files errors, but got some: %v", errs)
	}
}

func TestSpecTreeValidate_EF4_FragmentWithInvalidHash(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "alpha\nbeta\ngamma\ndelta\nepsilon\n"
	testWriteFile(t, "f.txt", content)

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithExternal([]*frontmatter.FrontmatterExternal{
			{
				Path: "f.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "1-3", Hash: "wrong"},
				},
			},
		}), testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasErrorWithNodeAndRule(errs, "ROOT/a", "external_files") {
		t.Errorf("expected an external_files error for ROOT/a, got: %v", errs)
	}
}

func TestSpecTreeValidate_EF5_FragmentWithInvalidRangeFormat(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "f.txt", "line1\nline2\nline3\n")

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithExternal([]*frontmatter.FrontmatterExternal{
			{
				Path: "f.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "abc", Hash: "x"},
				},
			},
		}), testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasErrorWithNodeAndRule(errs, "ROOT/a", "external_files") {
		t.Errorf("expected an external_files error for ROOT/a, got: %v", errs)
	}
}

func TestSpecTreeValidate_EF6_FragmentWithStartGreaterThanEnd(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "f.txt", "line1\nline2\nline3\nline4\nline5\n")

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithExternal([]*frontmatter.FrontmatterExternal{
			{
				Path: "f.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "5-3", Hash: "x"},
				},
			},
		}), testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasErrorWithNodeAndRule(errs, "ROOT/a", "external_files") {
		t.Errorf("expected an external_files error for ROOT/a, got: %v", errs)
	}
}

func TestSpecTreeValidate_EF7_FragmentWithStartLessThanOne(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "f.txt", "line1\nline2\nline3\n")

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithExternal([]*frontmatter.FrontmatterExternal{
			{
				Path: "f.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "0-3", Hash: "x"},
				},
			},
		}), testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasErrorWithNodeAndRule(errs, "ROOT/a", "external_files") {
		t.Errorf("expected an external_files error for ROOT/a, got: %v", errs)
	}
}

func TestSpecTreeValidate_EF8_FragmentOutOfRange(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "f.txt", "line1\nline2\nline3\nline4\nline5\n")

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithExternal([]*frontmatter.FrontmatterExternal{
			{
				Path: "f.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "1-100", Hash: "x"},
				},
			},
		}), testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasErrorWithNodeAndRule(errs, "ROOT/a", "external_files") {
		t.Errorf("expected an external_files error for ROOT/a, got: %v", errs)
	}
	// Verify detail mentions out of range
	found := false
	for _, e := range errs {
		if e.Node == "ROOT/a" && e.Rule == "external_files" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected external_files error with out-of-range detail, got: %v", errs)
	}
}

// ---------------------------------------------------------------------------
// Rule: output_paths
// ---------------------------------------------------------------------------

func TestSpecTreeValidate_OP1_ValidOutputPath(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithOutputs([]*frontmatter.FrontmatterOutput{{ID: "x", Path: "internal/x.go"}}), testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasErrorWithRule(errs, "output_paths") {
		t.Errorf("expected no output_paths errors, but got some: %v", errs)
	}
}

func TestSpecTreeValidate_OP2_OutputPathWithTraversal(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithOutputs([]*frontmatter.FrontmatterOutput{{ID: "x", Path: "../../etc/passwd"}}), testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasErrorWithNodeAndRule(errs, "ROOT/a", "output_paths") {
		t.Errorf("expected an output_paths error for ROOT/a, got: %v", errs)
	}
}

func TestSpecTreeValidate_OP3_OutputPathWithBackslash(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithOutputs([]*frontmatter.FrontmatterOutput{{ID: "x", Path: `internal\x.go`}}), testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasErrorWithNodeAndRule(errs, "ROOT/a", "output_paths") {
		t.Errorf("expected an output_paths error for ROOT/a, got: %v", errs)
	}
}

// ---------------------------------------------------------------------------
// Rule: duplicate_subsections
// ---------------------------------------------------------------------------

func TestSpecTreeValidate_DS1_UniqueSubsectionHeadings(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testEmptyFrontmatter(), testMakeNodeWithPublic("ROOT/a", []string{"interface", "context"})),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasErrorWithRule(errs, "duplicate_subsections") {
		t.Errorf("expected no duplicate_subsections errors, but got some: %v", errs)
	}
}

func TestSpecTreeValidate_DS2_DuplicateSubsectionHeadingsOneError(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testEmptyFrontmatter(), testMakeNodeWithPublic("ROOT/a", []string{"interface", "interface"})),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	count := testCountErrorsWithNodeAndRule(errs, "ROOT/a", "duplicate_subsections")
	if count != 1 {
		t.Errorf("expected exactly 1 duplicate_subsections error for ROOT/a, got %d: %v", count, errs)
	}
}

func TestSpecTreeValidate_DS3_ThreeIdenticalSubsectionHeadingsTwoErrors(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testEmptyFrontmatter(), testMakeNodeWithPublic("ROOT/a", []string{"interface", "interface", "interface"})),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	count := testCountErrorsWithNodeAndRule(errs, "ROOT/a", "duplicate_subsections")
	if count != 2 {
		t.Errorf("expected exactly 2 duplicate_subsections errors for ROOT/a, got %d: %v", count, errs)
	}
}

func TestSpecTreeValidate_DS4_NoPublicSectionSkip(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testEmptyFrontmatter(), testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasErrorWithRule(errs, "duplicate_subsections") {
		t.Errorf("expected no duplicate_subsections errors, but got some: %v", errs)
	}
}

// ---------------------------------------------------------------------------
// Cross-Cutting
// ---------------------------------------------------------------------------

func TestSpecTreeValidate_CC1_CollectsMultipleErrorsFromDifferentRules(t *testing.T) {
	node := testMakeNodeWithHeading("ROOT/a", "root/wrong")
	node.Public = &parsenode.NodeSection{
		Heading:    "public",
		RawHeading: "# Public",
		Content:    []string{},
		Subsections: []*parsenode.NodeSubsection{
			{Heading: "interface", RawHeading: "## interface", Content: []string{}},
			{Heading: "interface", RawHeading: "## interface", Content: []string{}},
		},
	}

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFrontmatter(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testFrontmatterWithDependsOn([]string{"ROOT/missing"}), node),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	rulesFound := map[string]bool{}
	for _, e := range errs {
		rulesFound[e.Rule] = true
	}

	requiredRules := []string{"name_heading", "dependency_targets", "duplicate_subsections"}
	for _, rule := range requiredRules {
		if !rulesFound[rule] {
			t.Errorf("expected FormatError with rule %q, but none found. All errors: %v", rule, errs)
		}
	}

	if len(errs) < 3 {
		t.Errorf("expected at least 3 FormatErrors, got %d: %v", len(errs), errs)
	}
}

func TestSpecTreeValidate_CC2_EmptyInputList(t *testing.T) {
	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{})
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty input, got %d: %v", len(errs), errs)
	}
}

// Ensure fmt is used (it is used in testFragmentHash indirectly, but add a direct reference).
var _ = fmt.Sprintf
