// code-from-spec: ROOT/golang/tests/spec_tree/validate@LYrB8ciroy20r3W8mroulapdgGw
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

func TestSpecTreeValidate_HappyPath_ValidLeafNode(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/b"},
			Output:    "internal/out.go",
		}, parsenode.Node{
			NameSection: testNameSection("root/a"),
			Public:      testPublicSection(nil),
		}),
		testEntry("ROOT/b", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root/b"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestSpecTreeValidate_HappyPath_ValidIntermediateNode(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
			Public:      testPublicSection(nil),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestSpecTreeValidate_HappyPath_LeafNoFrontmatter(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestSpecTreeValidate_NameHeading_Matches(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "ROOT/a", "name_heading") {
		t.Error("expected no name_heading error for ROOT/a")
	}
}

func TestSpecTreeValidate_NameHeading_DoesNotMatch(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root/wrong"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "name_heading") {
		t.Error("expected name_heading error for ROOT/a")
	}
}

func TestSpecTreeValidate_LeafOnlyFields_IntermediateWithDependsOn(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/b"},
		}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
		testEntry("ROOT/a/b", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root/a/b"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Error("expected leaf_only_fields error for ROOT/a")
	}
}

func TestSpecTreeValidate_LeafOnlyFields_IntermediateWithOutput(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{
			Output: "x.go",
		}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
		testEntry("ROOT/a/b", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root/a/b"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Error("expected leaf_only_fields error for ROOT/a")
	}
}

func TestSpecTreeValidate_LeafOnlyFields_IntermediateWithInput(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{
			Input: "ARTIFACT/c",
		}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
		testEntry("ROOT/a/b", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root/a/b"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Error("expected leaf_only_fields error for ROOT/a")
	}
}

func TestSpecTreeValidate_LeafOnlyFields_IntermediateWithExternal(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{
			External: []*frontmatter.FrontmatterExternal{{Path: "some/file.txt"}},
		}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
		testEntry("ROOT/a/b", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root/a/b"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Error("expected leaf_only_fields error for ROOT/a")
	}
}

func TestSpecTreeValidate_LeafOnlyFields_IntermediateWithMultipleRestrictedFields(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/b"},
			Output:    "x.go",
		}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
		testEntry("ROOT/a/b", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root/a/b"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	count := testCountErrors(errs, "ROOT/a", "leaf_only_fields")
	if count != 2 {
		t.Errorf("expected 2 leaf_only_fields errors for ROOT/a, got %d", count)
	}
}

func TestSpecTreeValidate_LeafOnlyAgent_IntermediateWithAgent(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root/a"),
			Agent:       testAgentSection([]string{"Agent instructions."}),
		}),
		testEntry("ROOT/a/b", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root/a/b"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_agent") {
		t.Error("expected leaf_only_agent error for ROOT/a")
	}
}

func TestSpecTreeValidate_LeafOnlyAgent_LeafWithAgentNoError(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root/a"),
			Agent:       testAgentSection([]string{"Agent instructions."}),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "ROOT/a", "leaf_only_agent") {
		t.Error("expected no leaf_only_agent error for leaf ROOT/a")
	}
}

func TestSpecTreeValidate_DependencyTargets_NonExistent(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/missing"},
		}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Error("expected dependency_targets error for ROOT/a")
	}
}

func TestSpecTreeValidate_DependencyTargets_Ancestor(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
		testEntry("ROOT/a/b", frontmatter.Frontmatter{
			DependsOn: []string{"ROOT"},
		}, parsenode.Node{
			NameSection: testNameSection("root/a/b"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a/b", "dependency_targets") {
		t.Error("expected dependency_targets error for ROOT/a/b (ancestor dep)")
	}
}

func TestSpecTreeValidate_DependencyTargets_Descendant(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/a/b"},
		}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
		testEntry("ROOT/a/b", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root/a/b"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Error("expected dependency_targets error for ROOT/a (descendant dep)")
	}
}

func TestSpecTreeValidate_DependencyTargets_Self(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/a"},
		}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Error("expected dependency_targets error for ROOT/a (self dep)")
	}
}

func TestSpecTreeValidate_DependencyTargets_ValidQualifier(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
		testEntry("ROOT/b", frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/a(interface)"},
		}, parsenode.Node{
			NameSection: testNameSection("root/b"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "ROOT/b", "dependency_targets") {
		t.Error("expected no dependency_targets error for ROOT/b (qualifier stripped)")
	}
}

func TestSpecTreeValidate_DependencyTargets_ValidArtifact(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{
			Output: "lib.go",
		}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
		testEntry("ROOT/b", frontmatter.Frontmatter{
			DependsOn: []string{"ARTIFACT/a"},
		}, parsenode.Node{
			NameSection: testNameSection("root/b"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "ROOT/b", "dependency_targets") {
		t.Error("expected no dependency_targets error for ROOT/b (valid artifact)")
	}
}

func TestSpecTreeValidate_DependencyTargets_NonExistentArtifact(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{
			DependsOn: []string{"ARTIFACT/missing"},
		}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Error("expected dependency_targets error for ROOT/a (non-existent artifact)")
	}
}

func TestSpecTreeValidate_DependencyTargets_MultipleInvalid(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/missing", "ROOT/also_missing"},
		}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	count := testCountErrors(errs, "ROOT/a", "dependency_targets")
	if count != 2 {
		t.Errorf("expected 2 dependency_targets errors for ROOT/a, got %d", count)
	}
}

func TestSpecTreeValidate_InputTarget_Valid(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{
			Output: "a.go",
		}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
		testEntry("ROOT/b", frontmatter.Frontmatter{
			Input: "ARTIFACT/a",
		}, parsenode.Node{
			NameSection: testNameSection("root/b"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "ROOT/b", "input_target") {
		t.Error("expected no input_target error for ROOT/b")
	}
}

func TestSpecTreeValidate_InputTarget_NotArtifactPrefix(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{
			Input: "ROOT/something",
		}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "input_target") {
		t.Error("expected input_target error for ROOT/a (not ARTIFACT/ prefix)")
	}
}

func TestSpecTreeValidate_InputTarget_NonExistentArtifact(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{
			Input: "ARTIFACT/missing",
		}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "input_target") {
		t.Error("expected input_target error for ROOT/a (non-existent artifact)")
	}
}

func TestSpecTreeValidate_ExternalFiles_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.MkdirAll("some", 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("some/file.txt", []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{
			External: []*frontmatter.FrontmatterExternal{{Path: "some/file.txt"}},
		}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "ROOT/a", "external_files") {
		t.Error("expected no external_files error when file exists")
	}
}

func TestSpecTreeValidate_ExternalFiles_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{
			External: []*frontmatter.FrontmatterExternal{{Path: "nonexistent.txt"}},
		}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Error("expected external_files error when file does not exist")
	}
}

func TestSpecTreeValidate_OutputPaths_Valid(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{
			Output: "internal/x.go",
		}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "ROOT/a", "output_paths") {
		t.Error("expected no output_paths error for valid path")
	}
}

func TestSpecTreeValidate_OutputPaths_Traversal(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{
			Output: "../../etc/passwd",
		}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "output_paths") {
		t.Error("expected output_paths error for path with traversal")
	}
}

func TestSpecTreeValidate_OutputPaths_Backslash(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{
			Output: `internal\x.go`,
		}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "output_paths") {
		t.Error("expected output_paths error for path with backslash")
	}
}

func TestSpecTreeValidate_DuplicateSubsections_UniqueHeadings(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root/a"),
			Public: testPublicSection([]*parsenode.NodeSubsection{
				{Heading: "interface", RawHeading: "## Interface", Content: []string{"Types."}},
				{Heading: "context", RawHeading: "## Context", Content: []string{"Background."}},
			}),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "ROOT/a", "duplicate_subsections") {
		t.Error("expected no duplicate_subsections error for unique headings")
	}
}

func TestSpecTreeValidate_DuplicateSubsections_Duplicate(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root/a"),
			Public: testPublicSection([]*parsenode.NodeSubsection{
				{Heading: "interface", RawHeading: "## Interface", Content: []string{"First."}},
				{Heading: "interface", RawHeading: "## Interface", Content: []string{"Second."}},
			}),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "duplicate_subsections") {
		t.Error("expected duplicate_subsections error for ROOT/a")
	}
}

func TestSpecTreeValidate_DuplicateSubsections_ThreeIdentical(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root/a"),
			Public: testPublicSection([]*parsenode.NodeSubsection{
				{Heading: "interface", RawHeading: "## Interface", Content: []string{"First."}},
				{Heading: "interface", RawHeading: "## Interface", Content: []string{"Second."}},
				{Heading: "interface", RawHeading: "## Interface", Content: []string{"Third."}},
			}),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	count := testCountErrors(errs, "ROOT/a", "duplicate_subsections")
	if count != 2 {
		t.Errorf("expected 2 duplicate_subsections errors for ROOT/a, got %d", count)
	}
}

func TestSpecTreeValidate_DuplicateSubsections_NoPublicSection(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root/a"),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "ROOT/a", "duplicate_subsections") {
		t.Error("expected no duplicate_subsections error when public section absent")
	}
}

func TestSpecTreeValidate_CrossCutting_MultipleErrors(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", frontmatter.Frontmatter{}, parsenode.Node{
			NameSection: testNameSection("root"),
		}),
		testEntry("ROOT/a", frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/missing"},
		}, parsenode.Node{
			NameSection: testNameSection("root/wrong"),
			Public: testPublicSection([]*parsenode.NodeSubsection{
				{Heading: "interface", RawHeading: "## Interface", Content: []string{"First."}},
				{Heading: "interface", RawHeading: "## Interface", Content: []string{"Second."}},
			}),
		}),
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "name_heading") {
		t.Error("expected name_heading error for ROOT/a")
	}
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Error("expected dependency_targets error for ROOT/a")
	}
	if !testHasError(errs, "ROOT/a", "duplicate_subsections") {
		t.Error("expected duplicate_subsections error for ROOT/a")
	}
}

func TestSpecTreeValidate_EmptyInput(t *testing.T) {
	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{})
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty input, got %d", len(errs))
	}
}
