// code-from-spec: ROOT/golang/tests/spec_tree/validate@guOx8688NdMII9Ny_U8x81yLfkM
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

func testPublicSection(subsections []*parsenode.NodeSubsection) *parsenode.NodeSection {
	return &parsenode.NodeSection{
		Heading:     "public",
		RawHeading:  "# Public",
		Content:     []string{},
		Subsections: subsections,
	}
}

func testAgentSection(content []string) *parsenode.NodeSection {
	return &parsenode.NodeSection{
		Heading:    "agent",
		RawHeading: "# Agent",
		Content:    content,
	}
}

func testNodeWith(nameHeading string) parsenode.Node {
	return parsenode.Node{
		NameSection: testNameSection(nameHeading),
	}
}

func testCountErrors(errs []*spectreevalidate.FormatError, node, rule string) int {
	count := 0
	for _, e := range errs {
		if (node == "" || e.Node == node) && (rule == "" || e.Rule == rule) {
			count++
		}
	}
	return count
}

func TestValidLeafNodePassesAllChecks(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/b"},
				Output:    "internal/out.go",
			},
			Node: parsenode.Node{
				NameSection: testNameSection("ROOT/a"),
				Public:      testPublicSection(nil),
			},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT/b"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
	}
}

func TestValidIntermediateNodePassesAllChecks(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node: parsenode.Node{
				NameSection: testNameSection("ROOT"),
				Public:      testPublicSection(nil),
			},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT/a"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
	}
}

func TestLeafWithNoFrontmatterFields(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT/a"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
	}
}

func TestNameHeadingMatchesLogicalName(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT/a"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "", "name_heading") != 0 {
		t.Errorf("expected no name_heading errors, got some: %+v", errs)
	}
}

func TestNameHeadingDoesNotMatchLogicalName(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT/wrong"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "ROOT/a", "name_heading") != 1 {
		t.Errorf("expected one name_heading error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "name_heading"))
	}
}

func TestIntermediateNodeWithDependsOn(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/b"}},
			Node:        testNodeWith("ROOT/a"),
		},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT/a/b"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "ROOT/a", "leaf_only_fields") != 1 {
		t.Errorf("expected one leaf_only_fields error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "leaf_only_fields"))
	}
}

func TestIntermediateNodeWithOutput(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Output: "x.go"},
			Node:        testNodeWith("ROOT/a"),
		},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT/a/b"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "ROOT/a", "leaf_only_fields") != 1 {
		t.Errorf("expected one leaf_only_fields error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "leaf_only_fields"))
	}
}

func TestIntermediateNodeWithInput(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Input: "ARTIFACT/c"},
			Node:        testNodeWith("ROOT/a"),
		},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT/a/b"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "ROOT/a", "leaf_only_fields") != 1 {
		t.Errorf("expected one leaf_only_fields error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "leaf_only_fields"))
	}
}

func TestIntermediateNodeWithExternal(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{
				External: []*frontmatter.FrontmatterExternal{{Path: "some/file.txt"}},
			},
			Node: testNodeWith("ROOT/a"),
		},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT/a/b"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "ROOT/a", "leaf_only_fields") != 1 {
		t.Errorf("expected one leaf_only_fields error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "leaf_only_fields"))
	}
}

func TestIntermediateNodeWithMultipleRestrictedFields(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/b"},
				Output:    "x.go",
			},
			Node: testNodeWith("ROOT/a"),
		},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT/a/b"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "ROOT/a", "leaf_only_fields") != 2 {
		t.Errorf("expected two leaf_only_fields errors for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "leaf_only_fields"))
	}
}

func TestIntermediateNodeWithAgentSection(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node: parsenode.Node{
				NameSection: testNameSection("ROOT/a"),
				Agent:       testAgentSection([]string{"Agent instructions."}),
			},
		},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT/a/b"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "ROOT/a", "leaf_only_agent") != 1 {
		t.Errorf("expected one leaf_only_agent error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "leaf_only_agent"))
	}
}

func TestLeafNodeWithAgentSectionNoError(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node: parsenode.Node{
				NameSection: testNameSection("ROOT/a"),
				Agent:       testAgentSection([]string{"Agent instructions."}),
			},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "", "leaf_only_agent") != 0 {
		t.Errorf("expected no leaf_only_agent errors, got some: %+v", errs)
	}
}

func TestDependsOnTargetsNonExistentRootNode(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/missing"}},
			Node:        testNodeWith("ROOT/a"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "ROOT/a", "dependency_targets") != 1 {
		t.Errorf("expected one dependency_targets error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "dependency_targets"))
	}
}

func TestDependsOnTargetsAncestor(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT/a"),
		},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT"}},
			Node:        testNodeWith("ROOT/a/b"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "ROOT/a/b", "dependency_targets") != 1 {
		t.Errorf("expected one dependency_targets error for ROOT/a/b, got %d", testCountErrors(errs, "ROOT/a/b", "dependency_targets"))
	}
}

func TestDependsOnTargetsDescendant(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/a/b"}},
			Node:        testNodeWith("ROOT/a"),
		},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT/a/b"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "ROOT/a", "dependency_targets") != 1 {
		t.Errorf("expected one dependency_targets error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "dependency_targets"))
	}
}

func TestDependsOnTargetsSelf(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/a"}},
			Node:        testNodeWith("ROOT/a"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "ROOT/a", "dependency_targets") != 1 {
		t.Errorf("expected one dependency_targets error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "dependency_targets"))
	}
}

func TestDependsOnWithValidRootQualifier(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT/a"),
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/a(interface)"}},
			Node:        testNodeWith("ROOT/b"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "", "dependency_targets") != 0 {
		t.Errorf("expected no dependency_targets errors, got some: %+v", errs)
	}
}

func TestDependsOnWithValidArtifactReference(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Output: "lib.go"},
			Node:        testNodeWith("ROOT/a"),
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ARTIFACT/a"}},
			Node:        testNodeWith("ROOT/b"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "", "dependency_targets") != 0 {
		t.Errorf("expected no dependency_targets errors, got some: %+v", errs)
	}
}

func TestDependsOnWithNonExistentArtifactReference(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ARTIFACT/missing"}},
			Node:        testNodeWith("ROOT/a"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "ROOT/a", "dependency_targets") != 1 {
		t.Errorf("expected one dependency_targets error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "dependency_targets"))
	}
}

func TestMultipleInvalidDependsOnOneErrorPerEntry(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/missing", "ROOT/also_missing"}},
			Node:        testNodeWith("ROOT/a"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "ROOT/a", "dependency_targets") != 2 {
		t.Errorf("expected two dependency_targets errors for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "dependency_targets"))
	}
}

func TestValidInputReference(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Output: "a.go"},
			Node:        testNodeWith("ROOT/a"),
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: frontmatter.Frontmatter{Input: "ARTIFACT/a"},
			Node:        testNodeWith("ROOT/b"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "", "input_target") != 0 {
		t.Errorf("expected no input_target errors, got some: %+v", errs)
	}
}

func TestInputNotStartingWithArtifact(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Input: "ROOT/something"},
			Node:        testNodeWith("ROOT/a"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "ROOT/a", "input_target") != 1 {
		t.Errorf("expected one input_target error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "input_target"))
	}
}

func TestInputReferencesNonExistentArtifact(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Input: "ARTIFACT/missing"},
			Node:        testNodeWith("ROOT/a"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "ROOT/a", "input_target") != 1 {
		t.Errorf("expected one input_target error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "input_target"))
	}
}

func TestExternalFileExists(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("some", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("some/file.txt", []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}

	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{
				External: []*frontmatter.FrontmatterExternal{{Path: "some/file.txt"}},
			},
			Node: testNodeWith("ROOT/a"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "", "external_files") != 0 {
		t.Errorf("expected no external_files errors, got some: %+v", errs)
	}
}

func TestExternalFileDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{
				External: []*frontmatter.FrontmatterExternal{{Path: "nonexistent.txt"}},
			},
			Node: testNodeWith("ROOT/a"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "ROOT/a", "external_files") != 1 {
		t.Errorf("expected one external_files error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "external_files"))
	}
}

func TestValidOutputPath(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Output: "internal/x.go"},
			Node:        testNodeWith("ROOT/a"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "", "output_paths") != 0 {
		t.Errorf("expected no output_paths errors, got some: %+v", errs)
	}
}

func TestOutputPathWithTraversal(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Output: "../../etc/passwd"},
			Node:        testNodeWith("ROOT/a"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "ROOT/a", "output_paths") != 1 {
		t.Errorf("expected one output_paths error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "output_paths"))
	}
}

func TestOutputPathWithBackslash(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Output: `internal\x.go`},
			Node:        testNodeWith("ROOT/a"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "ROOT/a", "output_paths") != 1 {
		t.Errorf("expected one output_paths error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "output_paths"))
	}
}

func TestUniqueSubsectionHeadingsNoError(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node: parsenode.Node{
				NameSection: testNameSection("ROOT/a"),
				Public: testPublicSection([]*parsenode.NodeSubsection{
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"Types."}},
					{Heading: "context", RawHeading: "## Context", Content: []string{"Background."}},
				}),
			},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "", "duplicate_subsections") != 0 {
		t.Errorf("expected no duplicate_subsections errors, got some: %+v", errs)
	}
}

func TestDuplicateSubsectionHeadings(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node: parsenode.Node{
				NameSection: testNameSection("ROOT/a"),
				Public: testPublicSection([]*parsenode.NodeSubsection{
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"First."}},
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"Second."}},
				}),
			},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "ROOT/a", "duplicate_subsections") != 1 {
		t.Errorf("expected one duplicate_subsections error for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "duplicate_subsections"))
	}
}

func TestThreeIdenticalSubsectionHeadings(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node: parsenode.Node{
				NameSection: testNameSection("ROOT/a"),
				Public: testPublicSection([]*parsenode.NodeSubsection{
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"First."}},
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"Second."}},
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"Third."}},
				}),
			},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "ROOT/a", "duplicate_subsections") != 2 {
		t.Errorf("expected two duplicate_subsections errors for ROOT/a, got %d", testCountErrors(errs, "ROOT/a", "duplicate_subsections"))
	}
}

func TestNoPublicSectionSkipDuplicateSubsections(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT/a"),
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "", "duplicate_subsections") != 0 {
		t.Errorf("expected no duplicate_subsections errors, got some: %+v", errs)
	}
}

func TestCollectsMultipleErrorsFromDifferentRules(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        testNodeWith("ROOT"),
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/missing"},
			},
			Node: parsenode.Node{
				NameSection: testNameSection("ROOT/wrong"),
				Public: testPublicSection([]*parsenode.NodeSubsection{
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"First."}},
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"Second."}},
				}),
			},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "ROOT/a", "name_heading") < 1 {
		t.Errorf("expected at least one name_heading error for ROOT/a")
	}
	if testCountErrors(errs, "ROOT/a", "dependency_targets") < 1 {
		t.Errorf("expected at least one dependency_targets error for ROOT/a")
	}
	if testCountErrors(errs, "ROOT/a", "duplicate_subsections") < 1 {
		t.Errorf("expected at least one duplicate_subsections error for ROOT/a")
	}
}

func TestEmptyInputList(t *testing.T) {
	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{})
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty input, got %d: %+v", len(errs), errs)
	}
}
