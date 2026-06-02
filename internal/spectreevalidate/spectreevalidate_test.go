// code-from-spec: ROOT/golang/tests/spec_tree/validate@g2Tstes8YmLCc2No2750eOu5w8M
package spectreevalidate_test

import (
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectreevalidate"
)

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

func testMakeEntry(logicalName string, fm frontmatter.Frontmatter, node parsenode.Node) *spectreevalidate.SpecTreeValidateInput {
	return &spectreevalidate.SpecTreeValidateInput{
		LogicalName: logicalName,
		Frontmatter: fm,
		Node:        node,
	}
}

func testMakeNode(heading string) parsenode.Node {
	return parsenode.Node{
		NameSection: &parsenode.NodeSection{
			Heading:    heading,
			RawHeading: "# " + heading,
			Content:    []string{},
		},
	}
}

func testMakeNodeWithPublic(heading string, subsections []*parsenode.NodeSubsection) parsenode.Node {
	node := testMakeNode(heading)
	node.Public = &parsenode.NodeSection{
		Heading:     "public",
		RawHeading:  "# Public",
		Content:     []string{},
		Subsections: subsections,
	}
	return node
}

func testMakeNodeWithAgent(heading string, content []string) parsenode.Node {
	node := testMakeNode(heading)
	node.Agent = &parsenode.NodeSection{
		Heading:    "agent",
		RawHeading: "# Agent",
		Content:    content,
	}
	return node
}

func testHasError(errs []*spectreevalidate.FormatError, node, rule string) bool {
	for _, e := range errs {
		if e.Node == node && e.Rule == rule {
			return true
		}
	}
	return false
}

func testCountErrors(errs []*spectreevalidate.FormatError, node, rule string) int {
	count := 0
	for _, e := range errs {
		if e.Node == node && e.Rule == rule {
			count++
		}
	}
	return count
}

func TestValidLeafNodePassesAllChecks(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNodeWithPublic("ROOT", nil)),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{DependsOn: []string{"ROOT/b"}, Output: "internal/out.go"}, testMakeNode("ROOT/a")),
		testMakeEntry("ROOT/b", frontmatter.Frontmatter{}, testMakeNode("ROOT/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
	}
}

func TestValidIntermediateNodePassesAllChecks(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNodeWithPublic("ROOT", nil)),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{}, testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
	}
}

func TestLeafWithNoFrontmatterFields(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{}, testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
	}
}

func TestHeadingMatchesLogicalName(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{}, testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testHasError(errs, "ROOT/a", "name_heading") {
		t.Errorf("expected no name_heading error for ROOT/a, got one")
	}
}

func TestHeadingDoesNotMatchLogicalName(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{}, testMakeNode("ROOT/wrong")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testCountErrors(errs, "ROOT/a", "name_heading") != 1 {
		t.Errorf("expected exactly one name_heading error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "name_heading"))
	}
}

func TestIntermediateNodeWithDependsOn(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{DependsOn: []string{"ROOT/b"}}, testMakeNode("ROOT/a")),
		testMakeEntry("ROOT/a/b", frontmatter.Frontmatter{}, testMakeNode("ROOT/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testCountErrors(errs, "ROOT/a", "leaf_only_fields") != 1 {
		t.Errorf("expected exactly one leaf_only_fields error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "leaf_only_fields"))
	}
}

func TestIntermediateNodeWithOutput(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{Output: "x.go"}, testMakeNode("ROOT/a")),
		testMakeEntry("ROOT/a/b", frontmatter.Frontmatter{}, testMakeNode("ROOT/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testCountErrors(errs, "ROOT/a", "leaf_only_fields") != 1 {
		t.Errorf("expected exactly one leaf_only_fields error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "leaf_only_fields"))
	}
}

func TestIntermediateNodeWithInput(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{Input: "ARTIFACT/c"}, testMakeNode("ROOT/a")),
		testMakeEntry("ROOT/a/b", frontmatter.Frontmatter{}, testMakeNode("ROOT/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testCountErrors(errs, "ROOT/a", "leaf_only_fields") != 1 {
		t.Errorf("expected exactly one leaf_only_fields error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "leaf_only_fields"))
	}
}

func TestIntermediateNodeWithExternal(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{External: []*frontmatter.FrontmatterExternal{{Path: "some/file.txt"}}}, testMakeNode("ROOT/a")),
		testMakeEntry("ROOT/a/b", frontmatter.Frontmatter{}, testMakeNode("ROOT/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testCountErrors(errs, "ROOT/a", "leaf_only_fields") != 1 {
		t.Errorf("expected exactly one leaf_only_fields error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "leaf_only_fields"))
	}
}

func TestIntermediateNodeWithMultipleRestrictedFields(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{DependsOn: []string{"ROOT/b"}, Output: "x.go"}, testMakeNode("ROOT/a")),
		testMakeEntry("ROOT/a/b", frontmatter.Frontmatter{}, testMakeNode("ROOT/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testCountErrors(errs, "ROOT/a", "leaf_only_fields") != 2 {
		t.Errorf("expected exactly two leaf_only_fields errors for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "leaf_only_fields"))
	}
}

func TestIntermediateNodeWithAgentSection(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{}, testMakeNodeWithAgent("ROOT/a", []string{"Agent instructions."})),
		testMakeEntry("ROOT/a/b", frontmatter.Frontmatter{}, testMakeNode("ROOT/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testCountErrors(errs, "ROOT/a", "leaf_only_agent") != 1 {
		t.Errorf("expected exactly one leaf_only_agent error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "leaf_only_agent"))
	}
}

func TestLeafNodeWithAgentSectionNoError(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{}, testMakeNodeWithAgent("ROOT/a", []string{"Agent instructions."})),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testHasError(errs, "ROOT/a", "leaf_only_agent") {
		t.Errorf("expected no leaf_only_agent error for leaf ROOT/a, got one")
	}
}

func TestDependsOnTargetsNonExistentROOTNode(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{DependsOn: []string{"ROOT/missing"}}, testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testCountErrors(errs, "ROOT/a", "dependency_targets") != 1 {
		t.Errorf("expected exactly one dependency_targets error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "dependency_targets"))
	}
}

func TestDependsOnTargetsAncestor(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{}, testMakeNode("ROOT/a")),
		testMakeEntry("ROOT/a/b", frontmatter.Frontmatter{DependsOn: []string{"ROOT"}}, testMakeNode("ROOT/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testCountErrors(errs, "ROOT/a/b", "dependency_targets") != 1 {
		t.Errorf("expected exactly one dependency_targets error for ROOT/a/b, got %d", testCountErrors(errs, "ROOT/a/b", "dependency_targets"))
	}
}

func TestDependsOnTargetsDescendant(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{DependsOn: []string{"ROOT/a/b"}}, testMakeNode("ROOT/a")),
		testMakeEntry("ROOT/a/b", frontmatter.Frontmatter{}, testMakeNode("ROOT/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testCountErrors(errs, "ROOT/a", "dependency_targets") != 1 {
		t.Errorf("expected exactly one dependency_targets error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "dependency_targets"))
	}
}

func TestDependsOnTargetsSelf(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{DependsOn: []string{"ROOT/a"}}, testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testCountErrors(errs, "ROOT/a", "dependency_targets") != 1 {
		t.Errorf("expected exactly one dependency_targets error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "dependency_targets"))
	}
}

func TestDependsOnWithValidROOTQualifier(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{}, testMakeNode("ROOT/a")),
		testMakeEntry("ROOT/b", frontmatter.Frontmatter{DependsOn: []string{"ROOT/a(interface)"}}, testMakeNode("ROOT/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testHasError(errs, "ROOT/b", "dependency_targets") {
		t.Errorf("expected no dependency_targets error for ROOT/b with qualified ROOT/a, got one")
	}
}

func TestDependsOnWithValidARTIFACTReference(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{Output: "lib.go"}, testMakeNode("ROOT/a")),
		testMakeEntry("ROOT/b", frontmatter.Frontmatter{DependsOn: []string{"ARTIFACT/a"}}, testMakeNode("ROOT/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testHasError(errs, "ROOT/b", "dependency_targets") {
		t.Errorf("expected no dependency_targets error for ROOT/b with ARTIFACT/a, got one")
	}
}

func TestDependsOnWithNonExistentARTIFACTReference(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{DependsOn: []string{"ARTIFACT/missing"}}, testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testCountErrors(errs, "ROOT/a", "dependency_targets") != 1 {
		t.Errorf("expected exactly one dependency_targets error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "dependency_targets"))
	}
}

func TestMultipleInvalidDependsOn(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{DependsOn: []string{"ROOT/missing", "ROOT/also_missing"}}, testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testCountErrors(errs, "ROOT/a", "dependency_targets") != 2 {
		t.Errorf("expected exactly two dependency_targets errors for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "dependency_targets"))
	}
}

func TestValidInputReference(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{Output: "a.go"}, testMakeNode("ROOT/a")),
		testMakeEntry("ROOT/b", frontmatter.Frontmatter{Input: "ARTIFACT/a"}, testMakeNode("ROOT/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testHasError(errs, "ROOT/b", "input_target") {
		t.Errorf("expected no input_target error for ROOT/b with valid ARTIFACT/a, got one")
	}
}

func TestInputNotStartingWithARTIFACT(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{Input: "ROOT/something"}, testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testCountErrors(errs, "ROOT/a", "input_target") != 1 {
		t.Errorf("expected exactly one input_target error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "input_target"))
	}
}

func TestInputReferencesNonExistentArtifact(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{Input: "ARTIFACT/missing"}, testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testCountErrors(errs, "ROOT/a", "input_target") != 1 {
		t.Errorf("expected exactly one input_target error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "input_target"))
	}
}

func TestExternalFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.MkdirAll("some", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile("some/file.txt", []byte("hello\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{External: []*frontmatter.FrontmatterExternal{{Path: "some/file.txt"}}}, testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("expected no external_files error for ROOT/a with existing file, got one")
	}
}

func TestExternalFileDoesNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{External: []*frontmatter.FrontmatterExternal{{Path: "nonexistent.txt"}}}, testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testCountErrors(errs, "ROOT/a", "external_files") != 1 {
		t.Errorf("expected exactly one external_files error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "external_files"))
	}
}

func TestValidOutputPath(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{Output: "internal/x.go"}, testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testHasError(errs, "ROOT/a", "output_paths") {
		t.Errorf("expected no output_paths error for ROOT/a with valid path, got one")
	}
}

func TestOutputPathWithTraversal(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{Output: "../../etc/passwd"}, testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testCountErrors(errs, "ROOT/a", "output_paths") != 1 {
		t.Errorf("expected exactly one output_paths error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "output_paths"))
	}
}

func TestOutputPathWithBackslash(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{Output: `internal\x.go`}, testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testCountErrors(errs, "ROOT/a", "output_paths") != 1 {
		t.Errorf("expected exactly one output_paths error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "output_paths"))
	}
}

func TestUniqueSubsectionHeadingsNoError(t *testing.T) {
	subsections := []*parsenode.NodeSubsection{
		{Heading: "interface", RawHeading: "## Interface", Content: []string{"Types."}},
		{Heading: "context", RawHeading: "## Context", Content: []string{"Background."}},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{}, testMakeNodeWithPublic("ROOT/a", subsections)),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testHasError(errs, "ROOT/a", "duplicate_subsections") {
		t.Errorf("expected no duplicate_subsections error for ROOT/a with unique headings, got one")
	}
}

func TestDuplicateSubsectionHeadings(t *testing.T) {
	subsections := []*parsenode.NodeSubsection{
		{Heading: "interface", RawHeading: "## Interface", Content: []string{"First."}},
		{Heading: "interface", RawHeading: "## Interface", Content: []string{"Second."}},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{}, testMakeNodeWithPublic("ROOT/a", subsections)),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testCountErrors(errs, "ROOT/a", "duplicate_subsections") != 1 {
		t.Errorf("expected exactly one duplicate_subsections error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "duplicate_subsections"))
	}
}

func TestThreeIdenticalSubsectionHeadings(t *testing.T) {
	subsections := []*parsenode.NodeSubsection{
		{Heading: "interface", RawHeading: "## Interface", Content: []string{"First."}},
		{Heading: "interface", RawHeading: "## Interface", Content: []string{"Second."}},
		{Heading: "interface", RawHeading: "## Interface", Content: []string{"Third."}},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{}, testMakeNodeWithPublic("ROOT/a", subsections)),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testCountErrors(errs, "ROOT/a", "duplicate_subsections") != 2 {
		t.Errorf("expected exactly two duplicate_subsections errors for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "duplicate_subsections"))
	}
}

func TestNoPublicSectionSkipDuplicateCheck(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{}, testMakeNode("ROOT/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if testHasError(errs, "ROOT/a", "duplicate_subsections") {
		t.Errorf("expected no duplicate_subsections error for ROOT/a with no public section, got one")
	}
}

func TestCollectsMultipleErrorsFromDifferentRules(t *testing.T) {
	subsections := []*parsenode.NodeSubsection{
		{Heading: "interface", RawHeading: "## Interface", Content: []string{"First."}},
		{Heading: "interface", RawHeading: "## Interface", Content: []string{"Second."}},
	}
	node := testMakeNodeWithPublic("ROOT/wrong", subsections)

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", frontmatter.Frontmatter{}, testMakeNode("ROOT")),
		testMakeEntry("ROOT/a", frontmatter.Frontmatter{DependsOn: []string{"ROOT/missing"}}, node),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	nameHeadingCount := testCountErrors(errs, "ROOT/a", "name_heading")
	depTargetsCount := testCountErrors(errs, "ROOT/a", "dependency_targets")
	dupSubsCount := testCountErrors(errs, "ROOT/a", "duplicate_subsections")

	if nameHeadingCount < 1 {
		t.Errorf("expected at least one name_heading error for ROOT/a, got %d", nameHeadingCount)
	}
	if depTargetsCount < 1 {
		t.Errorf("expected at least one dependency_targets error for ROOT/a, got %d", depTargetsCount)
	}
	if dupSubsCount < 1 {
		t.Errorf("expected at least one duplicate_subsections error for ROOT/a, got %d", dupSubsCount)
	}
	if len(errs) < 3 {
		t.Errorf("expected at least three total errors, got %d", len(errs))
	}
}

func TestEmptyInputList(t *testing.T) {
	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{})

	if len(errs) != 0 {
		t.Errorf("expected no errors for empty input, got %d: %+v", len(errs), errs)
	}
}
