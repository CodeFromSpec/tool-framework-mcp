// code-from-spec: ROOT/golang/tests/spec_tree/validate@YifZvtXJ_7rEwvFZlsPsKuyuGG4
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

func testMakeNode(heading string) *parsenode.Node {
	return &parsenode.Node{
		NameSection: &parsenode.NodeSection{
			Heading:    heading,
			RawHeading: "# " + heading,
			Content:    []string{},
		},
	}
}

func testMakeNodeWithPublic(heading string, publicContent []string, subsections []*parsenode.NodeSubsection) *parsenode.Node {
	node := testMakeNode(heading)
	node.Public = &parsenode.NodeSection{
		Heading:     "public",
		RawHeading:  "# Public",
		Content:     publicContent,
		Subsections: subsections,
	}
	return node
}

func testMakeNodeWithAgent(heading string, agentContent []string) *parsenode.Node {
	node := testMakeNode(heading)
	node.Agent = &parsenode.NodeSection{
		Heading:    "agent",
		RawHeading: "# Agent",
		Content:    agentContent,
	}
	return node
}

func testCountErrors(errs []*spectreevalidate.FormatError, rule string, node string) int {
	count := 0
	for _, e := range errs {
		if e.Rule == rule && e.Node == node {
			count++
		}
	}
	return count
}

func testHasError(errs []*spectreevalidate.FormatError, rule string, node string) bool {
	return testCountErrors(errs, rule, node) > 0
}

func TestHappyPath(t *testing.T) {
	t.Run("TC-HP-1_valid_leaf_node", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{
					DependsOn: []string{"SPEC/b"},
					Output:    "internal/out.go",
				},
				Node: testMakeNode("spec/a"),
			},
			{
				LogicalName: "SPEC/b",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec/b"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/b"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		if len(errs) != 0 {
			t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
		}
	})

	t.Run("TC-HP-2_valid_intermediate_node", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node: testMakeNodeWithPublic("spec", []string{}, []*parsenode.NodeSubsection{}),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec/a"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		if len(errs) != 0 {
			t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
		}
	})

	t.Run("TC-HP-3_leaf_with_no_frontmatter_fields", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec/a"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		if len(errs) != 0 {
			t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
		}
	})
}

func TestNameHeading(t *testing.T) {
	t.Run("TC-NH-1_heading_matches_logical_name", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec/a"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		if testHasError(errs, "name_heading", "SPEC/a") {
			t.Errorf("expected no name_heading error for SPEC/a")
		}
	})

	t.Run("TC-NH-2_heading_does_not_match_logical_name", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec/wrong"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		count := testCountErrors(errs, "name_heading", "SPEC/a")
		if count != 1 {
			t.Errorf("expected 1 name_heading error for SPEC/a, got %d", count)
		}
	})
}

func TestLeafOnlyFields(t *testing.T) {
	t.Run("TC-LOF-1_intermediate_with_depends_on", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/b"}},
				Node:        testMakeNode("spec/a"),
			},
			{
				LogicalName: "SPEC/a/b",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec/a/b"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/a/b"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		count := testCountErrors(errs, "leaf_only_fields", "SPEC/a")
		if count != 1 {
			t.Errorf("expected 1 leaf_only_fields error for SPEC/a, got %d", count)
		}
	})

	t.Run("TC-LOF-2_intermediate_with_output", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{Output: "x.go"},
				Node:        testMakeNode("spec/a"),
			},
			{
				LogicalName: "SPEC/a/b",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec/a/b"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/a/b"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		count := testCountErrors(errs, "leaf_only_fields", "SPEC/a")
		if count != 1 {
			t.Errorf("expected 1 leaf_only_fields error for SPEC/a, got %d", count)
		}
	})

	t.Run("TC-LOF-3_intermediate_with_input", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{Input: "ARTIFACT/c"},
				Node:        testMakeNode("spec/a"),
			},
			{
				LogicalName: "SPEC/a/b",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec/a/b"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/a/b"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		count := testCountErrors(errs, "leaf_only_fields", "SPEC/a")
		if count != 1 {
			t.Errorf("expected 1 leaf_only_fields error for SPEC/a, got %d", count)
		}
	})

	t.Run("TC-LOF-4_intermediate_with_multiple_restricted_fields", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{
					DependsOn: []string{"SPEC/b"},
					Output:    "x.go",
				},
				Node: testMakeNode("spec/a"),
			},
			{
				LogicalName: "SPEC/a/b",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec/a/b"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/a/b"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		count := testCountErrors(errs, "leaf_only_fields", "SPEC/a")
		if count != 2 {
			t.Errorf("expected 2 leaf_only_fields errors for SPEC/a, got %d", count)
		}
	})
}

func TestLeafOnlyAgent(t *testing.T) {
	t.Run("TC-LOA-1_intermediate_with_agent_section", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNodeWithAgent("spec/a", []string{"Agent instructions."}),
			},
			{
				LogicalName: "SPEC/a/b",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec/a/b"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/a/b"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		count := testCountErrors(errs, "leaf_only_agent", "SPEC/a")
		if count != 1 {
			t.Errorf("expected 1 leaf_only_agent error for SPEC/a, got %d", count)
		}
	})

	t.Run("TC-LOA-2_leaf_with_agent_section_no_error", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNodeWithAgent("spec/a", []string{"Agent instructions."}),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		if testHasError(errs, "leaf_only_agent", "SPEC/a") {
			t.Errorf("expected no leaf_only_agent error for leaf SPEC/a")
		}
	})
}

func TestDependencyTargets(t *testing.T) {
	t.Run("TC-DT-1_nonexistent_spec_node", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/missing"}},
				Node:        testMakeNode("spec/a"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		count := testCountErrors(errs, "dependency_targets", "SPEC/a")
		if count != 1 {
			t.Errorf("expected 1 dependency_targets error for SPEC/a, got %d", count)
		}
	})

	t.Run("TC-DT-2_depends_on_ancestor", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec/a"),
			},
			{
				LogicalName: "SPEC/a/b",
				Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC"}},
				Node:        testMakeNode("spec/a/b"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/a/b"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		count := testCountErrors(errs, "dependency_targets", "SPEC/a/b")
		if count != 1 {
			t.Errorf("expected 1 dependency_targets error for SPEC/a/b, got %d", count)
		}
	})

	t.Run("TC-DT-3_depends_on_descendant", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/a/b"}},
				Node:        testMakeNode("spec/a"),
			},
			{
				LogicalName: "SPEC/a/b",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec/a/b"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/a/b"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		count := testCountErrors(errs, "dependency_targets", "SPEC/a")
		if count != 1 {
			t.Errorf("expected 1 dependency_targets error for SPEC/a, got %d", count)
		}
	})

	t.Run("TC-DT-4_depends_on_self", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/a"}},
				Node:        testMakeNode("spec/a"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		count := testCountErrors(errs, "dependency_targets", "SPEC/a")
		if count != 1 {
			t.Errorf("expected 1 dependency_targets error for SPEC/a, got %d", count)
		}
	})

	t.Run("TC-DT-5_valid_spec_qualifier", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec/a"),
			},
			{
				LogicalName: "SPEC/b",
				Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/a(interface)"}},
				Node:        testMakeNode("spec/b"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/b"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		if testHasError(errs, "dependency_targets", "SPEC/b") {
			t.Errorf("expected no dependency_targets error for SPEC/b with qualified SPEC/a")
		}
	})

	t.Run("TC-DT-6_valid_artifact_reference", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{Output: "lib.go"},
				Node:        testMakeNode("spec/a"),
			},
			{
				LogicalName: "SPEC/b",
				Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ARTIFACT/a"}},
				Node:        testMakeNode("spec/b"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/b"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		if testHasError(errs, "dependency_targets", "SPEC/b") {
			t.Errorf("expected no dependency_targets error for SPEC/b with ARTIFACT/a")
		}
	})

	t.Run("TC-DT-7_nonexistent_artifact_reference", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"ARTIFACT/missing"}},
				Node:        testMakeNode("spec/a"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		count := testCountErrors(errs, "dependency_targets", "SPEC/a")
		if count != 1 {
			t.Errorf("expected 1 dependency_targets error for SPEC/a, got %d", count)
		}
	})

	t.Run("TC-DT-8_valid_external_reference", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		if err := os.MkdirAll("proto", 0755); err != nil {
			t.Fatalf("failed to create proto dir: %v", err)
		}
		if err := os.WriteFile("proto/api.proto", []byte("syntax = \"proto3\";"), 0644); err != nil {
			t.Fatalf("failed to create proto file: %v", err)
		}

		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"EXTERNAL/proto/api.proto"}},
				Node:        testMakeNode("spec/a"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		if testHasError(errs, "dependency_targets", "SPEC/a") {
			t.Errorf("expected no dependency_targets error for SPEC/a with valid EXTERNAL reference")
		}
	})

	t.Run("TC-DT-9_nonexistent_external_file", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"EXTERNAL/nonexistent.txt"}},
				Node:        testMakeNode("spec/a"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		count := testCountErrors(errs, "dependency_targets", "SPEC/a")
		if count != 1 {
			t.Errorf("expected 1 dependency_targets error for SPEC/a, got %d", count)
		}
	})

	t.Run("TC-DT-10_unrecognized_prefix", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"UNKNOWN/something"}},
				Node:        testMakeNode("spec/a"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		count := testCountErrors(errs, "dependency_targets", "SPEC/a")
		if count != 1 {
			t.Errorf("expected 1 dependency_targets error for SPEC/a, got %d", count)
		}
	})

	t.Run("TC-DT-11_multiple_invalid_depends_on", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/missing", "SPEC/also_missing"}},
				Node:        testMakeNode("spec/a"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		count := testCountErrors(errs, "dependency_targets", "SPEC/a")
		if count != 2 {
			t.Errorf("expected 2 dependency_targets errors for SPEC/a, got %d", count)
		}
	})
}

func TestInputTarget(t *testing.T) {
	t.Run("TC-IT-1_valid_artifact_input", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{Output: "a.go"},
				Node:        testMakeNode("spec/a"),
			},
			{
				LogicalName: "SPEC/b",
				Frontmatter: &frontmatter.Frontmatter{Input: "ARTIFACT/a"},
				Node:        testMakeNode("spec/b"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/b"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		if testHasError(errs, "input_target", "SPEC/b") {
			t.Errorf("expected no input_target error for SPEC/b with valid ARTIFACT input")
		}
	})

	t.Run("TC-IT-2_valid_external_input", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		if err := os.MkdirAll("docs", 0755); err != nil {
			t.Fatalf("failed to create docs dir: %v", err)
		}
		if err := os.WriteFile("docs/spec.yaml", []byte("openapi: 3.0"), 0644); err != nil {
			t.Fatalf("failed to create spec file: %v", err)
		}

		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{Input: "EXTERNAL/docs/spec.yaml"},
				Node:        testMakeNode("spec/a"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		if testHasError(errs, "input_target", "SPEC/a") {
			t.Errorf("expected no input_target error for SPEC/a with valid EXTERNAL input")
		}
	})

	t.Run("TC-IT-3_unsupported_prefix", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{Input: "SPEC/something"},
				Node:        testMakeNode("spec/a"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		count := testCountErrors(errs, "input_target", "SPEC/a")
		if count != 1 {
			t.Errorf("expected 1 input_target error for SPEC/a, got %d", count)
		}
	})

	t.Run("TC-IT-4_nonexistent_artifact", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{Input: "ARTIFACT/missing"},
				Node:        testMakeNode("spec/a"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		count := testCountErrors(errs, "input_target", "SPEC/a")
		if count != 1 {
			t.Errorf("expected 1 input_target error for SPEC/a, got %d", count)
		}
	})

	t.Run("TC-IT-5_nonexistent_external_file", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{Input: "EXTERNAL/nonexistent.txt"},
				Node:        testMakeNode("spec/a"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		count := testCountErrors(errs, "input_target", "SPEC/a")
		if count != 1 {
			t.Errorf("expected 1 input_target error for SPEC/a, got %d", count)
		}
	})
}

func TestMissingNodeMd(t *testing.T) {
	t.Run("TC-MN-1_subdir_without_node_md", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec/a"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/b"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		count := testCountErrors(errs, "missing_node_md", "code-from-spec/b")
		if count != 1 {
			t.Errorf("expected 1 missing_node_md error for code-from-spec/b, got %d", count)
		}
	})

	t.Run("TC-MN-2_underscore_prefixed_dir_no_error", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/_rules"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		if testHasError(errs, "missing_node_md", "code-from-spec/_rules") {
			t.Errorf("expected no missing_node_md error for underscore-prefixed dir")
		}
	})

	t.Run("TC-MN-3_all_subdirs_have_node_md", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec/a"),
			},
			{
				LogicalName: "SPEC/b",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec/b"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/b"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		for _, e := range errs {
			if e.Rule == "missing_node_md" {
				t.Errorf("expected no missing_node_md errors, got one for node %s", e.Node)
			}
		}
	})
}

func TestOutputPaths(t *testing.T) {
	t.Run("TC-OP-1_valid_output_path", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{Output: "internal/x.go"},
				Node:        testMakeNode("spec/a"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		if testHasError(errs, "output_paths", "SPEC/a") {
			t.Errorf("expected no output_paths error for SPEC/a with valid output path")
		}
	})

	t.Run("TC-OP-2_output_path_with_traversal", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{Output: "../../etc/passwd"},
				Node:        testMakeNode("spec/a"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		count := testCountErrors(errs, "output_paths", "SPEC/a")
		if count != 1 {
			t.Errorf("expected 1 output_paths error for SPEC/a with traversal, got %d", count)
		}
	})

	t.Run("TC-OP-3_output_path_with_backslash", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{Output: "internal\\x.go"},
				Node:        testMakeNode("spec/a"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		count := testCountErrors(errs, "output_paths", "SPEC/a")
		if count != 1 {
			t.Errorf("expected 1 output_paths error for SPEC/a with backslash, got %d", count)
		}
	})
}

func TestPublicSubsectionRequired(t *testing.T) {
	t.Run("TC-PSR-1_public_with_content_before_subsection", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{},
				Node: testMakeNodeWithPublic("spec/a", []string{"Some loose content."}, []*parsenode.NodeSubsection{
					{
						Heading:    "interface",
						RawHeading: "## Interface",
						Content:    []string{"Types."},
					},
				}),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		found := false
		for _, e := range errs {
			if e.Rule == "public_subsection_required" && e.Node == "SPEC/a" && e.Detail == "content in # Public must be under a ## subsection" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected public_subsection_required error with specific detail for SPEC/a")
		}
	})

	t.Run("TC-PSR-2_public_with_only_blank_lines_before_subsection", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{},
				Node: testMakeNodeWithPublic("spec/a", []string{"", "  ", ""}, []*parsenode.NodeSubsection{
					{
						Heading:    "interface",
						RawHeading: "## Interface",
						Content:    []string{"Types."},
					},
				}),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		if testHasError(errs, "public_subsection_required", "SPEC/a") {
			t.Errorf("expected no public_subsection_required error for SPEC/a with only blank lines")
		}
	})

	t.Run("TC-PSR-3_public_with_content_and_no_subsections", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{},
				Node: testMakeNodeWithPublic("spec/a", []string{"Some content."}, []*parsenode.NodeSubsection{}),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		count := testCountErrors(errs, "public_subsection_required", "SPEC/a")
		if count != 1 {
			t.Errorf("expected 1 public_subsection_required error for SPEC/a, got %d", count)
		}
	})

	t.Run("TC-PSR-4_public_with_only_subsections", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{},
				Node: testMakeNodeWithPublic("spec/a", []string{}, []*parsenode.NodeSubsection{
					{
						Heading:    "interface",
						RawHeading: "## Interface",
						Content:    []string{"Types."},
					},
				}),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		if testHasError(errs, "public_subsection_required", "SPEC/a") {
			t.Errorf("expected no public_subsection_required error for SPEC/a with only subsections")
		}
	})

	t.Run("TC-PSR-5_no_public_section_skip", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec/a"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		if testHasError(errs, "public_subsection_required", "SPEC/a") {
			t.Errorf("expected no public_subsection_required error when no public section")
		}
	})
}

func TestDuplicateSubsections(t *testing.T) {
	t.Run("TC-DS-1_unique_subsection_headings", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{},
				Node: testMakeNodeWithPublic("spec/a", []string{}, []*parsenode.NodeSubsection{
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"Types."}},
					{Heading: "context", RawHeading: "## Context", Content: []string{"Background."}},
				}),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		if testHasError(errs, "duplicate_subsections", "SPEC/a") {
			t.Errorf("expected no duplicate_subsections error for unique headings")
		}
	})

	t.Run("TC-DS-2_duplicate_subsection_headings", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{},
				Node: testMakeNodeWithPublic("spec/a", []string{}, []*parsenode.NodeSubsection{
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"First."}},
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"Second."}},
				}),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		count := testCountErrors(errs, "duplicate_subsections", "SPEC/a")
		if count != 1 {
			t.Errorf("expected 1 duplicate_subsections error for SPEC/a, got %d", count)
		}
	})

	t.Run("TC-DS-3_three_identical_subsection_headings", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{},
				Node: testMakeNodeWithPublic("spec/a", []string{}, []*parsenode.NodeSubsection{
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"First."}},
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"Second."}},
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"Third."}},
				}),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		count := testCountErrors(errs, "duplicate_subsections", "SPEC/a")
		if count != 2 {
			t.Errorf("expected 2 duplicate_subsections errors for SPEC/a, got %d", count)
		}
	})

	t.Run("TC-DS-4_no_public_section_skip", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec/a"),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		if testHasError(errs, "duplicate_subsections", "SPEC/a") {
			t.Errorf("expected no duplicate_subsections error when no public section")
		}
	})
}

func TestCrossCutting(t *testing.T) {
	t.Run("TC-CC-1_multiple_errors_from_different_rules", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{
			{
				LogicalName: "SPEC",
				Frontmatter: &frontmatter.Frontmatter{},
				Node:        testMakeNode("spec"),
			},
			{
				LogicalName: "SPEC/a",
				Frontmatter: &frontmatter.Frontmatter{
					DependsOn: []string{"SPEC/missing"},
				},
				Node: testMakeNodeWithPublic("spec/wrong", []string{}, []*parsenode.NodeSubsection{
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"First."}},
					{Heading: "interface", RawHeading: "## Interface", Content: []string{"Second."}},
				}),
			},
		}
		allDirs := []string{"code-from-spec", "code-from-spec/a"}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)

		if !testHasError(errs, "name_heading", "SPEC/a") {
			t.Errorf("expected name_heading error for SPEC/a")
		}
		if !testHasError(errs, "dependency_targets", "SPEC/a") {
			t.Errorf("expected dependency_targets error for SPEC/a")
		}
		if !testHasError(errs, "duplicate_subsections", "SPEC/a") {
			t.Errorf("expected duplicate_subsections error for SPEC/a")
		}
	})

	t.Run("TC-CC-2_empty_input_list", func(t *testing.T) {
		entries := []*spectreevalidate.SpecTreeValidateInput{}
		allDirs := []string{}

		errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
		if len(errs) != 0 {
			t.Errorf("expected no errors for empty input, got %d: %+v", len(errs), errs)
		}
	})
}
