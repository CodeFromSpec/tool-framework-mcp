// code-from-spec: ROOT/golang/tests/spec_tree/validate@60DsOoOwZTF95zOfmwDt5915se0
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

func testEntry(logicalName string, fm frontmatter.Frontmatter, node parsenode.Node) *spectreevalidate.SpecTreeValidateInput {
	return &spectreevalidate.SpecTreeValidateInput{
		LogicalName: logicalName,
		Frontmatter: fm,
		Node:        node,
	}
}

func testCountErrors(errs []*spectreevalidate.FormatError, rule string) int {
	count := 0
	for _, e := range errs {
		if e.Rule == rule {
			count++
		}
	}
	return count
}

func testHasError(errs []*spectreevalidate.FormatError, node, rule string) bool {
	for _, e := range errs {
		if e.Node == node && e.Rule == rule {
			return true
		}
	}
	return false
}

func TestValidLeafNodePassesAllChecks(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node: parsenode.Node{
				NameSection: testNameSection("root"),
				Public:      testPublicSection(nil),
			},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/b"},
				Output:    "internal/out.go",
			},
			Node: parsenode.Node{
				NameSection: testNameSection("root/a"),
			},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: frontmatter.Frontmatter{},
			Node: parsenode.Node{
				NameSection: testNameSection("root/b"),
			},
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
				NameSection: testNameSection("root"),
				Public:      testPublicSection(nil),
			},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node: parsenode.Node{
				NameSection: testNameSection("root/a"),
			},
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
			Node: parsenode.Node{
				NameSection: testNameSection("root"),
			},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node: parsenode.Node{
				NameSection: testNameSection("root/a"),
			},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
	}
}

func TestHeadingMatchesLogicalName(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node: parsenode.Node{
				NameSection: testNameSection("root"),
			},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node: parsenode.Node{
				NameSection: testNameSection("root/a"),
			},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "name_heading") != 0 {
		t.Errorf("expected no name_heading errors, got some: %+v", errs)
	}
}

func TestHeadingDoesNotMatchLogicalName(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node: parsenode.Node{
				NameSection: testNameSection("root"),
			},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node: parsenode.Node{
				NameSection: testNameSection("root/wrong"),
			},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "name_heading") {
		t.Errorf("expected name_heading error for ROOT/a, got: %+v", errs)
	}
	if testCountErrors(errs, "name_heading") != 1 {
		t.Errorf("expected exactly 1 name_heading error, got %d", testCountErrors(errs, "name_heading"))
	}
}

func TestIntermediateNodeWithDependsOn(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/b"}},
			Node:        parsenode.Node{NameSection: testNameSection("root/a")},
		},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root/a/b")},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected leaf_only_fields error for ROOT/a, got: %+v", errs)
	}
	if testCountErrors(errs, "leaf_only_fields") != 1 {
		t.Errorf("expected exactly 1 leaf_only_fields error, got %d", testCountErrors(errs, "leaf_only_fields"))
	}
}

func TestIntermediateNodeWithOutput(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Output: "x.go"},
			Node:        parsenode.Node{NameSection: testNameSection("root/a")},
		},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root/a/b")},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected leaf_only_fields error for ROOT/a, got: %+v", errs)
	}
	if testCountErrors(errs, "leaf_only_fields") != 1 {
		t.Errorf("expected exactly 1 leaf_only_fields error, got %d", testCountErrors(errs, "leaf_only_fields"))
	}
}

func TestIntermediateNodeWithInput(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Input: "ARTIFACT/c"},
			Node:        parsenode.Node{NameSection: testNameSection("root/a")},
		},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root/a/b")},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected leaf_only_fields error for ROOT/a, got: %+v", errs)
	}
	if testCountErrors(errs, "leaf_only_fields") != 1 {
		t.Errorf("expected exactly 1 leaf_only_fields error, got %d", testCountErrors(errs, "leaf_only_fields"))
	}
}

func TestIntermediateNodeWithExternal(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{
				External: []*frontmatter.FrontmatterExternal{{Path: "some/file.txt"}},
			},
			Node: parsenode.Node{NameSection: testNameSection("root/a")},
		},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root/a/b")},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected leaf_only_fields error for ROOT/a, got: %+v", errs)
	}
	if testCountErrors(errs, "leaf_only_fields") != 1 {
		t.Errorf("expected exactly 1 leaf_only_fields error, got %d", testCountErrors(errs, "leaf_only_fields"))
	}
}

func TestIntermediateNodeWithMultipleRestrictedFields(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/b"},
				Output:    "x.go",
			},
			Node: parsenode.Node{NameSection: testNameSection("root/a")},
		},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root/a/b")},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	count := testCountErrors(errs, "leaf_only_fields")
	if count != 2 {
		t.Errorf("expected exactly 2 leaf_only_fields errors, got %d: %+v", count, errs)
	}
}

func TestIntermediateNodeWithAgentSection(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node: parsenode.Node{
				NameSection: testNameSection("root/a"),
				Agent:       testAgentSection([]string{"Agent instructions."}),
			},
		},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root/a/b")},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_agent") {
		t.Errorf("expected leaf_only_agent error for ROOT/a, got: %+v", errs)
	}
	if testCountErrors(errs, "leaf_only_agent") != 1 {
		t.Errorf("expected exactly 1 leaf_only_agent error, got %d", testCountErrors(errs, "leaf_only_agent"))
	}
}

func TestLeafNodeWithAgentSectionNoError(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node: parsenode.Node{
				NameSection: testNameSection("root/a"),
				Agent:       testAgentSection([]string{"Agent instructions."}),
			},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "leaf_only_agent") != 0 {
		t.Errorf("expected no leaf_only_agent errors, got some: %+v", errs)
	}
}

func TestDependsOnTargetsNonExistentRootNode(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/missing"}},
			Node:        parsenode.Node{NameSection: testNameSection("root/a")},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a, got: %+v", errs)
	}
	if testCountErrors(errs, "dependency_targets") != 1 {
		t.Errorf("expected exactly 1 dependency_targets error, got %d", testCountErrors(errs, "dependency_targets"))
	}
}

func TestDependsOnTargetsAncestor(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root/a")},
		},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT"}},
			Node:        parsenode.Node{NameSection: testNameSection("root/a/b")},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a/b", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a/b, got: %+v", errs)
	}
	if testCountErrors(errs, "dependency_targets") != 1 {
		t.Errorf("expected exactly 1 dependency_targets error, got %d", testCountErrors(errs, "dependency_targets"))
	}
}

func TestDependsOnTargetsDescendant(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/a/b"}},
			Node:        parsenode.Node{NameSection: testNameSection("root/a")},
		},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root/a/b")},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a, got: %+v", errs)
	}
	if testCountErrors(errs, "dependency_targets") != 1 {
		t.Errorf("expected exactly 1 dependency_targets error, got %d", testCountErrors(errs, "dependency_targets"))
	}
}

func TestDependsOnTargetsSelf(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/a"}},
			Node:        parsenode.Node{NameSection: testNameSection("root/a")},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a, got: %+v", errs)
	}
	if testCountErrors(errs, "dependency_targets") != 1 {
		t.Errorf("expected exactly 1 dependency_targets error, got %d", testCountErrors(errs, "dependency_targets"))
	}
}

func TestDependsOnWithValidRootQualifier(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root/a")},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/a(interface)"}},
			Node:        parsenode.Node{NameSection: testNameSection("root/b")},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "dependency_targets") != 0 {
		t.Errorf("expected no dependency_targets errors, got some: %+v", errs)
	}
}

func TestDependsOnWithValidArtifactReference(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Output: "lib.go"},
			Node:        parsenode.Node{NameSection: testNameSection("root/a")},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ARTIFACT/a"}},
			Node:        parsenode.Node{NameSection: testNameSection("root/b")},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "dependency_targets") != 0 {
		t.Errorf("expected no dependency_targets errors, got some: %+v", errs)
	}
}

func TestDependsOnWithNonExistentArtifactReference(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ARTIFACT/missing"}},
			Node:        parsenode.Node{NameSection: testNameSection("root/a")},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a, got: %+v", errs)
	}
	if testCountErrors(errs, "dependency_targets") != 1 {
		t.Errorf("expected exactly 1 dependency_targets error, got %d", testCountErrors(errs, "dependency_targets"))
	}
}

func TestMultipleInvalidDependsOnOneErrorPerEntry(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/missing", "ROOT/also_missing"}},
			Node:        parsenode.Node{NameSection: testNameSection("root/a")},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	count := testCountErrors(errs, "dependency_targets")
	if count != 2 {
		t.Errorf("expected exactly 2 dependency_targets errors, got %d: %+v", count, errs)
	}
}

func TestValidInputReference(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Output: "a.go"},
			Node:        parsenode.Node{NameSection: testNameSection("root/a")},
		},
		{
			LogicalName: "ROOT/b",
			Frontmatter: frontmatter.Frontmatter{Input: "ARTIFACT/a"},
			Node:        parsenode.Node{NameSection: testNameSection("root/b")},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "input_target") != 0 {
		t.Errorf("expected no input_target errors, got some: %+v", errs)
	}
}

func TestInputNotStartingWithArtifact(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Input: "ROOT/something"},
			Node:        parsenode.Node{NameSection: testNameSection("root/a")},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "input_target") {
		t.Errorf("expected input_target error for ROOT/a, got: %+v", errs)
	}
	if testCountErrors(errs, "input_target") != 1 {
		t.Errorf("expected exactly 1 input_target error, got %d", testCountErrors(errs, "input_target"))
	}
}

func TestInputReferencesNonExistentArtifact(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Input: "ARTIFACT/missing"},
			Node:        parsenode.Node{NameSection: testNameSection("root/a")},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "input_target") {
		t.Errorf("expected input_target error for ROOT/a, got: %+v", errs)
	}
	if testCountErrors(errs, "input_target") != 1 {
		t.Errorf("expected exactly 1 input_target error, got %d", testCountErrors(errs, "input_target"))
	}
}

func TestExternalFileExists(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("some", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile("some/file.txt", []byte("hello\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{
				External: []*frontmatter.FrontmatterExternal{{Path: "some/file.txt"}},
			},
			Node: parsenode.Node{NameSection: testNameSection("root/a")},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "external_files") != 0 {
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
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{
				External: []*frontmatter.FrontmatterExternal{{Path: "nonexistent.txt"}},
			},
			Node: parsenode.Node{NameSection: testNameSection("root/a")},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("expected external_files error for ROOT/a, got: %+v", errs)
	}
	if testCountErrors(errs, "external_files") != 1 {
		t.Errorf("expected exactly 1 external_files error, got %d", testCountErrors(errs, "external_files"))
	}
}

func TestValidOutputPath(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Output: "internal/x.go"},
			Node:        parsenode.Node{NameSection: testNameSection("root/a")},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "output_paths") != 0 {
		t.Errorf("expected no output_paths errors, got some: %+v", errs)
	}
}

func TestOutputPathWithTraversal(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Output: "../../etc/passwd"},
			Node:        parsenode.Node{NameSection: testNameSection("root/a")},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "output_paths") {
		t.Errorf("expected output_paths error for ROOT/a, got: %+v", errs)
	}
	if testCountErrors(errs, "output_paths") != 1 {
		t.Errorf("expected exactly 1 output_paths error, got %d", testCountErrors(errs, "output_paths"))
	}
}

func TestOutputPathWithBackslash(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{Output: "internal\\x.go"},
			Node:        parsenode.Node{NameSection: testNameSection("root/a")},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "output_paths") {
		t.Errorf("expected output_paths error for ROOT/a, got: %+v", errs)
	}
	if testCountErrors(errs, "output_paths") != 1 {
		t.Errorf("expected exactly 1 output_paths error, got %d", testCountErrors(errs, "output_paths"))
	}
}

func TestUniqueSubsectionHeadingsNoError(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node: parsenode.Node{
				NameSection: testNameSection("root/a"),
				Public: testPublicSection([]*parsenode.NodeSubsection{
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"Types."}},
					{Heading: "context", RawHeading: "## Context", Content: []string{"Background."}},
				}),
			},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "duplicate_subsections") != 0 {
		t.Errorf("expected no duplicate_subsections errors, got some: %+v", errs)
	}
}

func TestDuplicateSubsectionHeadings(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node: parsenode.Node{
				NameSection: testNameSection("root/a"),
				Public: testPublicSection([]*parsenode.NodeSubsection{
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"First."}},
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"Second."}},
				}),
			},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "duplicate_subsections") {
		t.Errorf("expected duplicate_subsections error for ROOT/a, got: %+v", errs)
	}
	if testCountErrors(errs, "duplicate_subsections") != 1 {
		t.Errorf("expected exactly 1 duplicate_subsections error, got %d", testCountErrors(errs, "duplicate_subsections"))
	}
}

func TestThreeIdenticalSubsectionHeadings(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node: parsenode.Node{
				NameSection: testNameSection("root/a"),
				Public: testPublicSection([]*parsenode.NodeSubsection{
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"First."}},
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"Second."}},
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"Third."}},
				}),
			},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	count := testCountErrors(errs, "duplicate_subsections")
	if count != 2 {
		t.Errorf("expected exactly 2 duplicate_subsections errors, got %d: %+v", count, errs)
	}
}

func TestNoPublicSectionSkipDuplicateCheck(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node: parsenode.Node{
				NameSection: testNameSection("root/a"),
			},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testCountErrors(errs, "duplicate_subsections") != 0 {
		t.Errorf("expected no duplicate_subsections errors, got some: %+v", errs)
	}
}

func TestCollectsMultipleErrorsFromDifferentRules(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{NameSection: testNameSection("root")},
		},
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{DependsOn: []string{"ROOT/missing"}},
			Node: parsenode.Node{
				NameSection: testNameSection("root/wrong"),
				Public: testPublicSection([]*parsenode.NodeSubsection{
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"First."}},
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"Second."}},
				}),
			},
		},
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
	if len(errs) < 3 {
		t.Errorf("expected at least 3 errors, got %d: %+v", len(errs), errs)
	}
}

func TestEmptyInputList(t *testing.T) {
	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{})
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty input, got %d: %+v", len(errs), errs)
	}
}
