// code-from-spec: ROOT/golang/tests/spec_tree/validate@uBT5CPgLfNIjODdAKPy9PMOi7aA
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

func testNameSection(heading string) *parsenode.NodeSection {
	return &parsenode.NodeSection{
		Heading:    heading,
		RawHeading: "# " + heading,
		Content:    []string{},
	}
}

func testEntry(logicalName string, fm *frontmatter.Frontmatter, node *parsenode.Node) *spectreevalidate.SpecTreeValidateInput {
	return &spectreevalidate.SpecTreeValidateInput{
		LogicalName: logicalName,
		Frontmatter: fm,
		Node:        node,
	}
}

func testEmptyFM() *frontmatter.Frontmatter {
	return &frontmatter.Frontmatter{}
}

func testNode(heading string) *parsenode.Node {
	return &parsenode.Node{
		NameSection: testNameSection(heading),
	}
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

func testHasError(errs []*spectreevalidate.FormatError, node, rule string) bool {
	return testCountErrors(errs, node, rule) > 0
}

func TestHappyPath_ValidLeafNodePassesAllChecks(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/b"},
			Outputs:   []*frontmatter.FrontmatterOutput{{ID: "out", Path: "internal/out.go"}},
		}, testNode("root/a")),
		testEntry("ROOT/b", testEmptyFM(), testNode("root/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestHappyPath_ValidIntermediateNodePassesAllChecks(t *testing.T) {
	rootNode := testNode("root")
	rootNode.Public = &parsenode.NodeSection{
		Heading:     "public",
		RawHeading:  "# Public",
		Content:     []string{},
		Subsections: []*parsenode.NodeSubsection{},
	}

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), rootNode),
		testEntry("ROOT/a", testEmptyFM(), testNode("root/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestHappyPath_LeafWithNoFrontmatterFields(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", testEmptyFM(), testNode("root/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestNameHeading_HeadingMatchesLogicalName(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", testEmptyFM(), testNode("root/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "ROOT/a", "name_heading") {
		t.Errorf("expected no name_heading error for ROOT/a")
	}
}

func TestNameHeading_HeadingDoesNotMatchLogicalName(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", testEmptyFM(), testNode("root/wrong")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "name_heading") {
		t.Errorf("expected name_heading error for ROOT/a")
	}
}

func TestLeafOnlyFields_IntermediateNodeWithDependsOn(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{DependsOn: []string{"ROOT/b"}}, testNode("root/a")),
		testEntry("ROOT/a/b", testEmptyFM(), testNode("root/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected leaf_only_fields error for ROOT/a")
	}
}

func TestLeafOnlyFields_IntermediateNodeWithOutputs(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			Outputs: []*frontmatter.FrontmatterOutput{{ID: "x", Path: "x.go"}},
		}, testNode("root/a")),
		testEntry("ROOT/a/b", testEmptyFM(), testNode("root/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected leaf_only_fields error for ROOT/a")
	}
}

func TestLeafOnlyFields_IntermediateNodeWithInput(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{Input: "ARTIFACT/c(id)"}, testNode("root/a")),
		testEntry("ROOT/a/b", testEmptyFM(), testNode("root/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected leaf_only_fields error for ROOT/a")
	}
}

func TestLeafOnlyFields_IntermediateNodeWithExternal(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			External: []*frontmatter.FrontmatterExternal{{Path: "some/file.txt"}},
		}, testNode("root/a")),
		testEntry("ROOT/a/b", testEmptyFM(), testNode("root/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected leaf_only_fields error for ROOT/a")
	}
}

func TestLeafOnlyFields_IntermediateNodeWithMultipleRestrictedFields(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/b"},
			Outputs:   []*frontmatter.FrontmatterOutput{{ID: "x", Path: "x.go"}},
		}, testNode("root/a")),
		testEntry("ROOT/a/b", testEmptyFM(), testNode("root/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	count := testCountErrors(errs, "ROOT/a", "leaf_only_fields")
	if count != 2 {
		t.Errorf("expected exactly 2 leaf_only_fields errors for ROOT/a, got %d", count)
	}
}

func TestLeafOnlyAgent_IntermediateNodeWithAgentSection(t *testing.T) {
	nodeA := testNode("root/a")
	nodeA.Agent = &parsenode.NodeSection{
		Heading:     "agent",
		RawHeading:  "# Agent",
		Content:     []string{"Agent instructions."},
		Subsections: []*parsenode.NodeSubsection{},
	}

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", testEmptyFM(), nodeA),
		testEntry("ROOT/a/b", testEmptyFM(), testNode("root/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_agent") {
		t.Errorf("expected leaf_only_agent error for ROOT/a")
	}
}

func TestLeafOnlyAgent_LeafNodeWithAgentSection_NoError(t *testing.T) {
	nodeA := testNode("root/a")
	nodeA.Agent = &parsenode.NodeSection{
		Heading:     "agent",
		RawHeading:  "# Agent",
		Content:     []string{"Agent instructions."},
		Subsections: []*parsenode.NodeSubsection{},
	}

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", testEmptyFM(), nodeA),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "ROOT/a", "leaf_only_agent") {
		t.Errorf("expected no leaf_only_agent error for ROOT/a")
	}
}

func TestDependencyTargets_DependsOnTargetsNonExistentRootNode(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{DependsOn: []string{"ROOT/missing"}}, testNode("root/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a")
	}
}

func TestDependencyTargets_DependsOnTargetsAncestor(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", testEmptyFM(), testNode("root/a")),
		testEntry("ROOT/a/b", &frontmatter.Frontmatter{DependsOn: []string{"ROOT"}}, testNode("root/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a/b", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a/b")
	}
}

func TestDependencyTargets_DependsOnTargetsDescendant(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{DependsOn: []string{"ROOT/a/b"}}, testNode("root/a")),
		testEntry("ROOT/a/b", testEmptyFM(), testNode("root/a/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a")
	}
}

func TestDependencyTargets_DependsOnTargetsSelf(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{DependsOn: []string{"ROOT/a"}}, testNode("root/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a")
	}
}

func TestDependencyTargets_DependsOnWithValidRootQualifier(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", testEmptyFM(), testNode("root/a")),
		testEntry("ROOT/b", &frontmatter.Frontmatter{DependsOn: []string{"ROOT/a(interface)"}}, testNode("root/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "ROOT/b", "dependency_targets") {
		t.Errorf("expected no dependency_targets error for ROOT/b")
	}
}

func TestDependencyTargets_DependsOnWithValidArtifactReference(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			Outputs: []*frontmatter.FrontmatterOutput{{ID: "lib", Path: "lib.go"}},
		}, testNode("root/a")),
		testEntry("ROOT/b", &frontmatter.Frontmatter{DependsOn: []string{"ARTIFACT/a(lib)"}}, testNode("root/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "ROOT/b", "dependency_targets") {
		t.Errorf("expected no dependency_targets error for ROOT/b")
	}
}

func TestDependencyTargets_DependsOnWithNonExistentArtifactReference(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{DependsOn: []string{"ARTIFACT/missing(id)"}}, testNode("root/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a")
	}
}

func TestDependencyTargets_MultipleInvalidDependsOn_OneErrorPerEntry(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/missing", "ROOT/also_missing"},
		}, testNode("root/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	count := testCountErrors(errs, "ROOT/a", "dependency_targets")
	if count != 2 {
		t.Errorf("expected exactly 2 dependency_targets errors for ROOT/a, got %d", count)
	}
}

func TestInputTarget_ValidInputReference(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			Outputs: []*frontmatter.FrontmatterOutput{{ID: "out", Path: "a.go"}},
		}, testNode("root/a")),
		testEntry("ROOT/b", &frontmatter.Frontmatter{Input: "ARTIFACT/a(out)"}, testNode("root/b")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "ROOT/b", "input_target") {
		t.Errorf("expected no input_target error for ROOT/b")
	}
}

func TestInputTarget_InputNotStartingWithArtifact(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{Input: "ROOT/something"}, testNode("root/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "input_target") {
		t.Errorf("expected input_target error for ROOT/a")
	}
}

func TestInputTarget_InputReferencesNonExistentArtifact(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{Input: "ARTIFACT/missing(id)"}, testNode("root/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "input_target") {
		t.Errorf("expected input_target error for ROOT/a")
	}
}

func TestExternalFiles_ExternalFileExists(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("some", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile("some/file.txt", []byte("hello\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			External: []*frontmatter.FrontmatterExternal{{Path: "some/file.txt"}},
		}, testNode("root/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("expected no external_files error for ROOT/a")
	}
}

func TestExternalFiles_ExternalFileDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			External: []*frontmatter.FrontmatterExternal{{Path: "nonexistent.txt"}},
		}, testNode("root/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("expected external_files error for ROOT/a")
	}
}

func TestOutputPaths_ValidOutputPath(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			Outputs: []*frontmatter.FrontmatterOutput{{ID: "x", Path: "internal/x.go"}},
		}, testNode("root/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "ROOT/a", "output_paths") {
		t.Errorf("expected no output_paths error for ROOT/a")
	}
}

func TestOutputPaths_OutputPathWithTraversal(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			Outputs: []*frontmatter.FrontmatterOutput{{ID: "x", Path: "../../etc/passwd"}},
		}, testNode("root/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "output_paths") {
		t.Errorf("expected output_paths error for ROOT/a")
	}
}

func TestOutputPaths_OutputPathWithBackslash(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			Outputs: []*frontmatter.FrontmatterOutput{{ID: "x", Path: `internal\x.go`}},
		}, testNode("root/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "output_paths") {
		t.Errorf("expected output_paths error for ROOT/a")
	}
}

func TestDuplicateSubsections_UniqueSubsectionHeadings_NoError(t *testing.T) {
	nodeA := testNode("root/a")
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
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", testEmptyFM(), nodeA),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "ROOT/a", "duplicate_subsections") {
		t.Errorf("expected no duplicate_subsections error for ROOT/a")
	}
}

func TestDuplicateSubsections_DuplicateSubsectionHeadings(t *testing.T) {
	nodeA := testNode("root/a")
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
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", testEmptyFM(), nodeA),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	count := testCountErrors(errs, "ROOT/a", "duplicate_subsections")
	if count != 1 {
		t.Errorf("expected exactly 1 duplicate_subsections error for ROOT/a, got %d", count)
	}
}

func TestDuplicateSubsections_ThreeIdenticalSubsectionHeadings(t *testing.T) {
	nodeA := testNode("root/a")
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
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", testEmptyFM(), nodeA),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	count := testCountErrors(errs, "ROOT/a", "duplicate_subsections")
	if count != 2 {
		t.Errorf("expected exactly 2 duplicate_subsections errors for ROOT/a, got %d", count)
	}
}

func TestDuplicateSubsections_NoPublicSection_Skip(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", testEmptyFM(), testNode("root/a")),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "ROOT/a", "duplicate_subsections") {
		t.Errorf("expected no duplicate_subsections error for ROOT/a")
	}
}

func TestCrossCutting_CollectsMultipleErrorsFromDifferentRules(t *testing.T) {
	nodeA := testNode("root/wrong")
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
		testEntry("ROOT", testEmptyFM(), testNode("root")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{DependsOn: []string{"ROOT/missing"}}, nodeA),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)

	if !testHasError(errs, "ROOT/a", "name_heading") {
		t.Errorf("expected name_heading error for ROOT/a")
	}
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a")
	}
	if !testHasError(errs, "ROOT/a", "duplicate_subsections") {
		t.Errorf("expected duplicate_subsections error for ROOT/a")
	}
}

func TestCrossCutting_EmptyInputList(t *testing.T) {
	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{})
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty input, got %d", len(errs))
	}
}
