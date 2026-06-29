// code-from-spec: SPEC/golang/tests/spec_tree/validate@9OPld9SR6OU4pnxQzuBvm3F5t5g
package spectreevalidate_test

import (
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/spectreevalidate"
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

func strPtr(s string) *string {
	return &s
}

func testNode(logicalName string, fm *parsing.NodeFrontmatter, nameHeading string, public *parsing.NodeSection, agent *parsing.NodeSection) parsing.Node {
	return parsing.Node{
		Reference: parsing.CfsReference{
			LogicalName: logicalName,
		},
		Frontmatter: fm,
		NameSection: parsing.NodeSection{
			Heading:    nameHeading,
			RawHeading: "# " + logicalName,
			Content:    []string{},
		},
		Public: public,
		Agent:  agent,
	}
}

func testPublicSection(content []string, subsections []*parsing.NodeSubsection) *parsing.NodeSection {
	return &parsing.NodeSection{
		Heading:     "public",
		RawHeading:  "# Public",
		Content:     content,
		Subsections: subsections,
	}
}

func testAgentSection(content []string) *parsing.NodeSection {
	return &parsing.NodeSection{
		Heading:    "agent",
		RawHeading: "# Agent",
		Content:    content,
	}
}

func testSubsection(heading, rawHeading string, content []string) *parsing.NodeSubsection {
	return &parsing.NodeSubsection{
		Heading:    heading,
		RawHeading: rawHeading,
		Content:    content,
	}
}

func testCountErrors(errs []spectreevalidate.FormatError, node, rule string) int {
	count := 0
	for _, e := range errs {
		if (node == "" || e.Node == node) && (rule == "" || e.Rule == rule) {
			count++
		}
	}
	return count
}

func TestHappyPath_ValidLeafNodePassesAllChecks(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{DependsOn: []string{"SPEC/root/b"}, Output: strPtr("internal/out.go")}, "spec/root/a", nil, nil),
		testNode("SPEC/root/b", nil, "spec/root/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a", "code-from-spec/root/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
	}
}

func TestHappyPath_ValidIntermediateNodePassesAllChecks(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", testPublicSection([]string{}, nil), nil),
		testNode("SPEC/root/a", nil, "spec/root/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
	}
}

func TestHappyPath_LeafWithNoFrontmatterFields(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{}, "spec/root/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
	}
}

func TestNameHeading_HeadingMatchesLogicalName(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", nil, "spec/root/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if testCountErrors(errs, "", "name_heading") != 0 {
		t.Errorf("expected no name_heading errors, got some: %+v", errs)
	}
}

func TestNameHeading_HeadingDoesNotMatchLogicalName(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", nil, "spec/wrong", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a", "name_heading")
	if count != 1 {
		t.Errorf("expected exactly 1 name_heading error for SPEC/root/a, got %d: %+v", count, errs)
	}
}

func TestLeafOnlyFields_IntermediateNodeWithDependsOn(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{DependsOn: []string{"SPEC/root/b"}}, "spec/root/a", nil, nil),
		testNode("SPEC/root/a/b", nil, "spec/root/a/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a", "code-from-spec/root/a/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a", "leaf_only_fields")
	if count != 1 {
		t.Errorf("expected exactly 1 leaf_only_fields error for SPEC/root/a, got %d: %+v", count, errs)
	}
}

func TestLeafOnlyFields_IntermediateNodeWithOutput(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{Output: strPtr("x.go")}, "spec/root/a", nil, nil),
		testNode("SPEC/root/a/b", nil, "spec/root/a/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a", "code-from-spec/root/a/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a", "leaf_only_fields")
	if count != 1 {
		t.Errorf("expected exactly 1 leaf_only_fields error for SPEC/root/a, got %d: %+v", count, errs)
	}
}

func TestLeafOnlyFields_IntermediateNodeWithInput(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{Input: strPtr("ARTIFACT/root/c")}, "spec/root/a", nil, nil),
		testNode("SPEC/root/a/b", nil, "spec/root/a/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a", "code-from-spec/root/a/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a", "leaf_only_fields")
	if count != 1 {
		t.Errorf("expected exactly 1 leaf_only_fields error for SPEC/root/a, got %d: %+v", count, errs)
	}
}

func TestLeafOnlyFields_IntermediateNodeWithMultipleRestrictedFields(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{DependsOn: []string{"SPEC/root/b"}, Output: strPtr("x.go")}, "spec/root/a", nil, nil),
		testNode("SPEC/root/a/b", nil, "spec/root/a/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a", "code-from-spec/root/a/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a", "leaf_only_fields")
	if count != 2 {
		t.Errorf("expected exactly 2 leaf_only_fields errors for SPEC/root/a, got %d: %+v", count, errs)
	}
}

func TestLeafOnlyAgent_IntermediateNodeWithAgentSection(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", nil, "spec/root/a", nil, testAgentSection([]string{"Agent instructions."})),
		testNode("SPEC/root/a/b", nil, "spec/root/a/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a", "code-from-spec/root/a/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a", "leaf_only_agent")
	if count != 1 {
		t.Errorf("expected exactly 1 leaf_only_agent error for SPEC/root/a, got %d: %+v", count, errs)
	}
}

func TestLeafOnlyAgent_LeafNodeWithAgentSection_NoError(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", nil, "spec/root/a", nil, testAgentSection([]string{"Agent instructions."})),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "leaf_only_agent")
	if count != 0 {
		t.Errorf("expected no leaf_only_agent errors, got %d: %+v", count, errs)
	}
}

func TestDependencyTargets_NonExistentSpecNode(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{DependsOn: []string{"SPEC/root/missing"}}, "spec/root/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a", "dependency_targets")
	if count != 1 {
		t.Errorf("expected exactly 1 dependency_targets error for SPEC/root/a, got %d: %+v", count, errs)
	}
}

func TestDependencyTargets_TargetsAncestor(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", nil, "spec/root/a", nil, nil),
		testNode("SPEC/root/a/b", &parsing.NodeFrontmatter{DependsOn: []string{"SPEC/root"}}, "spec/root/a/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a", "code-from-spec/root/a/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a/b", "dependency_targets")
	if count != 1 {
		t.Errorf("expected exactly 1 dependency_targets error for SPEC/root/a/b, got %d: %+v", count, errs)
	}
}

func TestDependencyTargets_TargetsDescendant(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{DependsOn: []string{"SPEC/root/a/b"}}, "spec/root/a", nil, nil),
		testNode("SPEC/root/a/b", nil, "spec/root/a/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a", "code-from-spec/root/a/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a", "dependency_targets")
	if count != 1 {
		t.Errorf("expected exactly 1 dependency_targets error for SPEC/root/a, got %d: %+v", count, errs)
	}
}

func TestDependencyTargets_TargetsSelf(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{DependsOn: []string{"SPEC/root/a"}}, "spec/root/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a", "dependency_targets")
	if count != 1 {
		t.Errorf("expected exactly 1 dependency_targets error for SPEC/root/a, got %d: %+v", count, errs)
	}
}

func TestDependencyTargets_ValidSpecQualifier(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", nil, "spec/root/a", nil, nil),
		testNode("SPEC/root/b", &parsing.NodeFrontmatter{DependsOn: []string{"SPEC/root/a(interface)"}}, "spec/root/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a", "code-from-spec/root/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "dependency_targets")
	if count != 0 {
		t.Errorf("expected no dependency_targets errors, got %d: %+v", count, errs)
	}
}

func TestDependencyTargets_ValidArtifactReference(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{Output: strPtr("lib.go")}, "spec/root/a", nil, nil),
		testNode("SPEC/root/b", &parsing.NodeFrontmatter{DependsOn: []string{"ARTIFACT/root/a"}}, "spec/root/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a", "code-from-spec/root/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "dependency_targets")
	if count != 0 {
		t.Errorf("expected no dependency_targets errors, got %d: %+v", count, errs)
	}
}

func TestDependencyTargets_NonExistentArtifactReference(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{DependsOn: []string{"ARTIFACT/root/missing"}}, "spec/root/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a", "dependency_targets")
	if count != 1 {
		t.Errorf("expected exactly 1 dependency_targets error for SPEC/root/a, got %d: %+v", count, errs)
	}
}

func TestDependencyTargets_ValidExternalReference(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("proto", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("proto/api.proto", []byte("syntax = \"proto3\";"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{DependsOn: []string{"EXTERNAL/proto/api.proto"}}, "spec/root/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "dependency_targets")
	if count != 0 {
		t.Errorf("expected no dependency_targets errors, got %d: %+v", count, errs)
	}
}

func TestDependencyTargets_NonExistentExternalFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{DependsOn: []string{"EXTERNAL/nonexistent.txt"}}, "spec/root/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a", "dependency_targets")
	if count != 1 {
		t.Errorf("expected exactly 1 dependency_targets error for SPEC/root/a, got %d: %+v", count, errs)
	}
}

func TestDependencyTargets_UnrecognizedPrefix(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{DependsOn: []string{"UNKNOWN/something"}}, "spec/root/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a", "dependency_targets")
	if count != 1 {
		t.Errorf("expected exactly 1 dependency_targets error for SPEC/root/a, got %d: %+v", count, errs)
	}
}

func TestDependencyTargets_MultipleInvalidEntries(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{DependsOn: []string{"SPEC/root/missing", "SPEC/root/also_missing"}}, "spec/root/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a", "dependency_targets")
	if count != 2 {
		t.Errorf("expected exactly 2 dependency_targets errors for SPEC/root/a, got %d: %+v", count, errs)
	}
}

func TestInputTarget_ValidArtifactReference(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{Output: strPtr("a.go")}, "spec/root/a", nil, nil),
		testNode("SPEC/root/b", &parsing.NodeFrontmatter{Input: strPtr("ARTIFACT/root/a")}, "spec/root/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a", "code-from-spec/root/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "input_target")
	if count != 0 {
		t.Errorf("expected no input_target errors, got %d: %+v", count, errs)
	}
}

func TestInputTarget_ValidExternalReference(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("docs", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("docs/spec.yaml", []byte("version: 1"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{Input: strPtr("EXTERNAL/docs/spec.yaml")}, "spec/root/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "input_target")
	if count != 0 {
		t.Errorf("expected no input_target errors, got %d: %+v", count, errs)
	}
}

func TestInputTarget_ValidSpecReference(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", nil, "spec/root/a", nil, nil),
		testNode("SPEC/root/b", &parsing.NodeFrontmatter{Input: strPtr("SPEC/root/a")}, "spec/root/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a", "code-from-spec/root/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "input_target")
	if count != 0 {
		t.Errorf("expected no input_target errors, got %d: %+v", count, errs)
	}
}

func TestInputTarget_ValidSpecReferenceWithQualifier(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", nil, "spec/root/a", nil, nil),
		testNode("SPEC/root/b", &parsing.NodeFrontmatter{Input: strPtr("SPEC/root/a(acceptance-tests)")}, "spec/root/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a", "code-from-spec/root/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "input_target")
	if count != 0 {
		t.Errorf("expected no input_target errors, got %d: %+v", count, errs)
	}
}

func TestInputTarget_NonExistentSpecReference(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{Input: strPtr("SPEC/root/missing")}, "spec/root/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a", "input_target")
	if count != 1 {
		t.Errorf("expected exactly 1 input_target error for SPEC/root/a, got %d: %+v", count, errs)
	}
}

func TestInputTarget_UnsupportedPrefix(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{Input: strPtr("UNKNOWN/something")}, "spec/root/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a", "input_target")
	if count != 1 {
		t.Errorf("expected exactly 1 input_target error for SPEC/root/a, got %d: %+v", count, errs)
	}
}

func TestInputTarget_NonExistentArtifact(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{Input: strPtr("ARTIFACT/root/missing")}, "spec/root/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a", "input_target")
	if count != 1 {
		t.Errorf("expected exactly 1 input_target error for SPEC/root/a, got %d: %+v", count, errs)
	}
}

func TestInputTarget_NonExistentExternalFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{Input: strPtr("EXTERNAL/nonexistent.txt")}, "spec/root/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a", "input_target")
	if count != 1 {
		t.Errorf("expected exactly 1 input_target error for SPEC/root/a, got %d: %+v", count, errs)
	}
}

func TestMissingNodeMd_SubdirectoryWithoutNodeMd(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", nil, "spec/root/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a", "code-from-spec/root/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "code-from-spec/root/b", "missing_node_md")
	if count != 1 {
		t.Errorf("expected exactly 1 missing_node_md error for code-from-spec/root/b, got %d: %+v", count, errs)
	}
}

func TestMissingNodeMd_DotPrefixedDirUnderCodeFromSpec_NoError(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/.cache"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "missing_node_md")
	if count != 0 {
		t.Errorf("expected no missing_node_md errors, got %d: %+v", count, errs)
	}
}

func TestMissingNodeMd_DotPrefixedDirDeeperInTree_NoError(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", nil, "spec/root/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a", "code-from-spec/root/a/.internal"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "missing_node_md")
	if count != 0 {
		t.Errorf("expected no missing_node_md errors, got %d: %+v", count, errs)
	}
}

func TestMissingNodeMd_AllSubdirsHaveNodeMd_NoError(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", nil, "spec/root/a", nil, nil),
		testNode("SPEC/root/b", nil, "spec/root/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a", "code-from-spec/root/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "missing_node_md")
	if count != 0 {
		t.Errorf("expected no missing_node_md errors, got %d: %+v", count, errs)
	}
}

func TestOutputPaths_ValidOutputPath(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{Output: strPtr("internal/x.go")}, "spec/root/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "output_paths")
	if count != 0 {
		t.Errorf("expected no output_paths errors, got %d: %+v", count, errs)
	}
}

func TestOutputPaths_TraversalInOutputPath(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{Output: strPtr("../../etc/passwd")}, "spec/root/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a", "output_paths")
	if count != 1 {
		t.Errorf("expected exactly 1 output_paths error for SPEC/root/a, got %d: %+v", count, errs)
	}
}

func TestOutputPaths_BackslashInOutputPath(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", &parsing.NodeFrontmatter{Output: strPtr("internal\\x.go")}, "spec/root/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a", "output_paths")
	if count != 1 {
		t.Errorf("expected exactly 1 output_paths error for SPEC/root/a, got %d: %+v", count, errs)
	}
}

func TestPublicSubsectionRequired_ContentBeforeFirstSubsection(t *testing.T) {
	subs := []*parsing.NodeSubsection{
		testSubsection("interface", "## Interface", []string{"Types."}),
	}
	pub := testPublicSection([]string{"Some loose content."}, subs)

	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", nil, "spec/root/a", pub, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a", "public_subsection_required")
	if count != 1 {
		t.Errorf("expected exactly 1 public_subsection_required error for SPEC/root/a, got %d: %+v", count, errs)
	}
	for _, e := range errs {
		if e.Node == "SPEC/root/a" && e.Rule == "public_subsection_required" {
			if e.Detail != "content in # Public must be under a ## subsection" {
				t.Errorf("unexpected detail: %q", e.Detail)
			}
		}
	}
}

func TestPublicSubsectionRequired_OnlyBlankLinesBeforeSubsection_NoError(t *testing.T) {
	subs := []*parsing.NodeSubsection{
		testSubsection("interface", "## Interface", []string{"Types."}),
	}
	pub := testPublicSection([]string{"", "  ", ""}, subs)

	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", nil, "spec/root/a", pub, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "public_subsection_required")
	if count != 0 {
		t.Errorf("expected no public_subsection_required errors, got %d: %+v", count, errs)
	}
}

func TestPublicSubsectionRequired_ContentWithNoSubsections(t *testing.T) {
	pub := testPublicSection([]string{"Some content."}, nil)

	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", nil, "spec/root/a", pub, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a", "public_subsection_required")
	if count != 1 {
		t.Errorf("expected exactly 1 public_subsection_required error for SPEC/root/a, got %d: %+v", count, errs)
	}
}

func TestPublicSubsectionRequired_OnlySubsections_NoError(t *testing.T) {
	subs := []*parsing.NodeSubsection{
		testSubsection("interface", "## Interface", []string{"Types."}),
	}
	pub := testPublicSection([]string{}, subs)

	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", nil, "spec/root/a", pub, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "public_subsection_required")
	if count != 0 {
		t.Errorf("expected no public_subsection_required errors, got %d: %+v", count, errs)
	}
}

func TestPublicSubsectionRequired_NoPublicSection_Skip(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", nil, "spec/root/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "public_subsection_required")
	if count != 0 {
		t.Errorf("expected no public_subsection_required errors, got %d: %+v", count, errs)
	}
}

func TestDuplicateSubsections_UniqueHeadings_NoError(t *testing.T) {
	subs := []*parsing.NodeSubsection{
		testSubsection("interface", "## Interface", []string{"Types."}),
		testSubsection("context", "## Context", []string{"Background."}),
	}
	pub := testPublicSection([]string{}, subs)

	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", nil, "spec/root/a", pub, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "duplicate_subsections")
	if count != 0 {
		t.Errorf("expected no duplicate_subsections errors, got %d: %+v", count, errs)
	}
}

func TestDuplicateSubsections_DuplicateHeadings(t *testing.T) {
	subs := []*parsing.NodeSubsection{
		testSubsection("interface", "## Interface", []string{"First."}),
		testSubsection("interface", "## Interface", []string{"Second."}),
	}
	pub := testPublicSection([]string{}, subs)

	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", nil, "spec/root/a", pub, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a", "duplicate_subsections")
	if count != 1 {
		t.Errorf("expected exactly 1 duplicate_subsections error for SPEC/root/a, got %d: %+v", count, errs)
	}
}

func TestDuplicateSubsections_ThreeIdenticalHeadings(t *testing.T) {
	subs := []*parsing.NodeSubsection{
		testSubsection("interface", "## Interface", []string{"First."}),
		testSubsection("interface", "## Interface", []string{"Second."}),
		testSubsection("interface", "## Interface", []string{"Third."}),
	}
	pub := testPublicSection([]string{}, subs)

	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", nil, "spec/root/a", pub, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/root/a", "duplicate_subsections")
	if count != 2 {
		t.Errorf("expected exactly 2 duplicate_subsections errors for SPEC/root/a, got %d: %+v", count, errs)
	}
}

func TestDuplicateSubsections_NoPublicSection_Skip(t *testing.T) {
	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		testNode("SPEC/root/a", nil, "spec/root/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "duplicate_subsections")
	if count != 0 {
		t.Errorf("expected no duplicate_subsections errors, got %d: %+v", count, errs)
	}
}

func TestCrossCutting_CollectsMultipleErrorsFromDifferentRules(t *testing.T) {
	subs := []*parsing.NodeSubsection{
		testSubsection("interface", "## Interface", []string{"First."}),
		testSubsection("interface", "## Interface", []string{"Second."}),
	}
	pub := testPublicSection([]string{}, subs)

	entries := []parsing.Node{
		testNode("SPEC/root", nil, "spec/root", nil, nil),
		{
			Reference: parsing.CfsReference{
				LogicalName: "SPEC/root/a",
			},
			Frontmatter: &parsing.NodeFrontmatter{DependsOn: []string{"SPEC/root/missing"}},
			NameSection: parsing.NodeSection{
				Heading:    "spec/wrong",
				RawHeading: "# SPEC/root/a",
				Content:    []string{},
			},
			Public: pub,
		},
	}
	allDirs := []string{"code-from-spec", "code-from-spec/root", "code-from-spec/root/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)

	if testCountErrors(errs, "SPEC/root/a", "name_heading") < 1 {
		t.Errorf("expected at least 1 name_heading error for SPEC/root/a: %+v", errs)
	}
	if testCountErrors(errs, "SPEC/root/a", "dependency_targets") < 1 {
		t.Errorf("expected at least 1 dependency_targets error for SPEC/root/a: %+v", errs)
	}
	if testCountErrors(errs, "SPEC/root/a", "duplicate_subsections") < 1 {
		t.Errorf("expected at least 1 duplicate_subsections error for SPEC/root/a: %+v", errs)
	}
	if len(errs) < 3 {
		t.Errorf("expected at least 3 total errors, got %d: %+v", len(errs), errs)
	}
}

func TestCrossCutting_EmptyInputList(t *testing.T) {
	errs := spectreevalidate.SpecTreeValidate(nil, nil)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
	}
}
