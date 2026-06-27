// code-from-spec: SPEC/golang/tests/spec_tree/validate@sKz-DRdlGiMqPlYKyp4ekw_QCrw
package spectreevalidate_test

import (
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/spectreevalidate"
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

func testNodeSection(heading string) *parsenode.NodeSection {
	return &parsenode.NodeSection{
		Heading:    heading,
		RawHeading: "# " + heading,
		Content:    []string{},
	}
}

func testPublicSection(content []string, subsections []*parsenode.NodeSubsection) *parsenode.NodeSection {
	return &parsenode.NodeSection{
		Heading:     "public",
		RawHeading:  "# Public",
		Content:     content,
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

func testSubsection(heading, rawHeading string, content []string) *parsenode.NodeSubsection {
	return &parsenode.NodeSubsection{
		Heading:    heading,
		RawHeading: rawHeading,
		Content:    content,
	}
}

func testEntry(logicalName string, fm *frontmatter.Frontmatter, nameHeading string, public *parsenode.NodeSection, agent *parsenode.NodeSection) *spectreevalidate.SpecTreeValidateInput {
	return &spectreevalidate.SpecTreeValidateInput{
		LogicalName: logicalName,
		Frontmatter: fm,
		Node: &parsenode.Node{
			NameSection: testNodeSection(nameHeading),
			Public:      public,
			Agent:       agent,
		},
	}
}

func testEmptyFM() *frontmatter.Frontmatter {
	return &frontmatter.Frontmatter{}
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

func TestHappyPath_ValidLeafNodePassesAllChecks(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", &frontmatter.Frontmatter{DependsOn: []string{"SPEC/b"}, Output: "internal/out.go"}, "spec/a", nil, nil),
		testEntry("SPEC/b", testEmptyFM(), "spec/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
	}
}

func TestHappyPath_ValidIntermediateNodePassesAllChecks(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", testPublicSection([]string{}, nil), nil),
		testEntry("SPEC/a", testEmptyFM(), "spec/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
	}
}

func TestHappyPath_LeafWithNoFrontmatterFields(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", testEmptyFM(), "spec/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
	}
}

func TestNameHeading_HeadingMatchesLogicalName(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", testEmptyFM(), "spec/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if testCountErrors(errs, "", "name_heading") != 0 {
		t.Errorf("expected no name_heading errors, got some: %+v", errs)
	}
}

func TestNameHeading_HeadingDoesNotMatchLogicalName(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", testEmptyFM(), "spec/wrong", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/a", "name_heading")
	if count != 1 {
		t.Errorf("expected exactly 1 name_heading error for SPEC/a, got %d: %+v", count, errs)
	}
}

func TestLeafOnlyFields_IntermediateNodeWithDependsOn(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", &frontmatter.Frontmatter{DependsOn: []string{"SPEC/b"}}, "spec/a", nil, nil),
		testEntry("SPEC/a/b", testEmptyFM(), "spec/a/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/a/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/a", "leaf_only_fields")
	if count != 1 {
		t.Errorf("expected exactly 1 leaf_only_fields error for SPEC/a, got %d: %+v", count, errs)
	}
}

func TestLeafOnlyFields_IntermediateNodeWithOutput(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", &frontmatter.Frontmatter{Output: "x.go"}, "spec/a", nil, nil),
		testEntry("SPEC/a/b", testEmptyFM(), "spec/a/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/a/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/a", "leaf_only_fields")
	if count != 1 {
		t.Errorf("expected exactly 1 leaf_only_fields error for SPEC/a, got %d: %+v", count, errs)
	}
}

func TestLeafOnlyFields_IntermediateNodeWithInput(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", &frontmatter.Frontmatter{Input: "ARTIFACT/c"}, "spec/a", nil, nil),
		testEntry("SPEC/a/b", testEmptyFM(), "spec/a/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/a/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/a", "leaf_only_fields")
	if count != 1 {
		t.Errorf("expected exactly 1 leaf_only_fields error for SPEC/a, got %d: %+v", count, errs)
	}
}

func TestLeafOnlyFields_IntermediateNodeWithMultipleRestrictedFields(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", &frontmatter.Frontmatter{DependsOn: []string{"SPEC/b"}, Output: "x.go"}, "spec/a", nil, nil),
		testEntry("SPEC/a/b", testEmptyFM(), "spec/a/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/a/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/a", "leaf_only_fields")
	if count != 2 {
		t.Errorf("expected exactly 2 leaf_only_fields errors for SPEC/a, got %d: %+v", count, errs)
	}
}

func TestLeafOnlyAgent_IntermediateNodeWithAgentSection(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", testEmptyFM(), "spec/a", nil, testAgentSection([]string{"Agent instructions."})),
		testEntry("SPEC/a/b", testEmptyFM(), "spec/a/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/a/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/a", "leaf_only_agent")
	if count != 1 {
		t.Errorf("expected exactly 1 leaf_only_agent error for SPEC/a, got %d: %+v", count, errs)
	}
}

func TestLeafOnlyAgent_LeafNodeWithAgentSection_NoError(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", testEmptyFM(), "spec/a", nil, testAgentSection([]string{"Agent instructions."})),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "leaf_only_agent")
	if count != 0 {
		t.Errorf("expected no leaf_only_agent errors, got %d: %+v", count, errs)
	}
}

func TestDependencyTargets_NonExistentSpecNode(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", &frontmatter.Frontmatter{DependsOn: []string{"SPEC/missing"}}, "spec/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/a", "dependency_targets")
	if count != 1 {
		t.Errorf("expected exactly 1 dependency_targets error for SPEC/a, got %d: %+v", count, errs)
	}
}

func TestDependencyTargets_TargetsAncestor(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", testEmptyFM(), "spec/a", nil, nil),
		testEntry("SPEC/a/b", &frontmatter.Frontmatter{DependsOn: []string{"SPEC"}}, "spec/a/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/a/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/a/b", "dependency_targets")
	if count != 1 {
		t.Errorf("expected exactly 1 dependency_targets error for SPEC/a/b, got %d: %+v", count, errs)
	}
}

func TestDependencyTargets_TargetsDescendant(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", &frontmatter.Frontmatter{DependsOn: []string{"SPEC/a/b"}}, "spec/a", nil, nil),
		testEntry("SPEC/a/b", testEmptyFM(), "spec/a/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/a/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/a", "dependency_targets")
	if count != 1 {
		t.Errorf("expected exactly 1 dependency_targets error for SPEC/a, got %d: %+v", count, errs)
	}
}

func TestDependencyTargets_TargetsSelf(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", &frontmatter.Frontmatter{DependsOn: []string{"SPEC/a"}}, "spec/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/a", "dependency_targets")
	if count != 1 {
		t.Errorf("expected exactly 1 dependency_targets error for SPEC/a, got %d: %+v", count, errs)
	}
}

func TestDependencyTargets_ValidSpecQualifier(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", testEmptyFM(), "spec/a", nil, nil),
		testEntry("SPEC/b", &frontmatter.Frontmatter{DependsOn: []string{"SPEC/a(interface)"}}, "spec/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "dependency_targets")
	if count != 0 {
		t.Errorf("expected no dependency_targets errors, got %d: %+v", count, errs)
	}
}

func TestDependencyTargets_ValidArtifactReference(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", &frontmatter.Frontmatter{Output: "lib.go"}, "spec/a", nil, nil),
		testEntry("SPEC/b", &frontmatter.Frontmatter{DependsOn: []string{"ARTIFACT/a"}}, "spec/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "dependency_targets")
	if count != 0 {
		t.Errorf("expected no dependency_targets errors, got %d: %+v", count, errs)
	}
}

func TestDependencyTargets_NonExistentArtifactReference(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", &frontmatter.Frontmatter{DependsOn: []string{"ARTIFACT/missing"}}, "spec/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/a", "dependency_targets")
	if count != 1 {
		t.Errorf("expected exactly 1 dependency_targets error for SPEC/a, got %d: %+v", count, errs)
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

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", &frontmatter.Frontmatter{DependsOn: []string{"EXTERNAL/proto/api.proto"}}, "spec/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "dependency_targets")
	if count != 0 {
		t.Errorf("expected no dependency_targets errors, got %d: %+v", count, errs)
	}
}

func TestDependencyTargets_NonExistentExternalFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", &frontmatter.Frontmatter{DependsOn: []string{"EXTERNAL/nonexistent.txt"}}, "spec/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/a", "dependency_targets")
	if count != 1 {
		t.Errorf("expected exactly 1 dependency_targets error for SPEC/a, got %d: %+v", count, errs)
	}
}

func TestDependencyTargets_UnrecognizedPrefix(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", &frontmatter.Frontmatter{DependsOn: []string{"UNKNOWN/something"}}, "spec/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/a", "dependency_targets")
	if count != 1 {
		t.Errorf("expected exactly 1 dependency_targets error for SPEC/a, got %d: %+v", count, errs)
	}
}

func TestDependencyTargets_MultipleInvalidEntries(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", &frontmatter.Frontmatter{DependsOn: []string{"SPEC/missing", "SPEC/also_missing"}}, "spec/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/a", "dependency_targets")
	if count != 2 {
		t.Errorf("expected exactly 2 dependency_targets errors for SPEC/a, got %d: %+v", count, errs)
	}
}

func TestInputTarget_ValidArtifactReference(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", &frontmatter.Frontmatter{Output: "a.go"}, "spec/a", nil, nil),
		testEntry("SPEC/b", &frontmatter.Frontmatter{Input: "ARTIFACT/a"}, "spec/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/b"}

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

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", &frontmatter.Frontmatter{Input: "EXTERNAL/docs/spec.yaml"}, "spec/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "input_target")
	if count != 0 {
		t.Errorf("expected no input_target errors, got %d: %+v", count, errs)
	}
}

func TestInputTarget_UnsupportedPrefix(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", &frontmatter.Frontmatter{Input: "SPEC/something"}, "spec/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/a", "input_target")
	if count != 1 {
		t.Errorf("expected exactly 1 input_target error for SPEC/a, got %d: %+v", count, errs)
	}
}

func TestInputTarget_NonExistentArtifact(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", &frontmatter.Frontmatter{Input: "ARTIFACT/missing"}, "spec/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/a", "input_target")
	if count != 1 {
		t.Errorf("expected exactly 1 input_target error for SPEC/a, got %d: %+v", count, errs)
	}
}

func TestInputTarget_NonExistentExternalFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", &frontmatter.Frontmatter{Input: "EXTERNAL/nonexistent.txt"}, "spec/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/a", "input_target")
	if count != 1 {
		t.Errorf("expected exactly 1 input_target error for SPEC/a, got %d: %+v", count, errs)
	}
}

func TestMissingNodeMd_SubdirectoryWithoutNodeMd(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", testEmptyFM(), "spec/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "code-from-spec/b", "missing_node_md")
	if count != 1 {
		t.Errorf("expected exactly 1 missing_node_md error for code-from-spec/b, got %d: %+v", count, errs)
	}
}

func TestMissingNodeMd_UnderscorePrefixedDir_NoError(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/_rules"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "missing_node_md")
	if count != 0 {
		t.Errorf("expected no missing_node_md errors, got %d: %+v", count, errs)
	}
}

func TestMissingNodeMd_AllSubdirsHaveNodeMd_NoError(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", testEmptyFM(), "spec/a", nil, nil),
		testEntry("SPEC/b", testEmptyFM(), "spec/b", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a", "code-from-spec/b"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "missing_node_md")
	if count != 0 {
		t.Errorf("expected no missing_node_md errors, got %d: %+v", count, errs)
	}
}

func TestOutputPaths_ValidOutputPath(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", &frontmatter.Frontmatter{Output: "internal/x.go"}, "spec/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "output_paths")
	if count != 0 {
		t.Errorf("expected no output_paths errors, got %d: %+v", count, errs)
	}
}

func TestOutputPaths_TraversalInOutputPath(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", &frontmatter.Frontmatter{Output: "../../etc/passwd"}, "spec/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/a", "output_paths")
	if count != 1 {
		t.Errorf("expected exactly 1 output_paths error for SPEC/a, got %d: %+v", count, errs)
	}
}

func TestOutputPaths_BackslashInOutputPath(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", &frontmatter.Frontmatter{Output: "internal\\x.go"}, "spec/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/a", "output_paths")
	if count != 1 {
		t.Errorf("expected exactly 1 output_paths error for SPEC/a, got %d: %+v", count, errs)
	}
}

func TestPublicSubsectionRequired_ContentBeforeFirstSubsection(t *testing.T) {
	subs := []*parsenode.NodeSubsection{
		testSubsection("interface", "## Interface", []string{"Types."}),
	}
	pub := testPublicSection([]string{"Some loose content."}, subs)

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", testEmptyFM(), "spec/a", pub, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/a", "public_subsection_required")
	if count != 1 {
		t.Errorf("expected exactly 1 public_subsection_required error for SPEC/a, got %d: %+v", count, errs)
	}
	for _, e := range errs {
		if e.Node == "SPEC/a" && e.Rule == "public_subsection_required" {
			if e.Detail != "content in # Public must be under a ## subsection" {
				t.Errorf("unexpected detail: %q", e.Detail)
			}
		}
	}
}

func TestPublicSubsectionRequired_OnlyBlankLinesBeforeSubsection_NoError(t *testing.T) {
	subs := []*parsenode.NodeSubsection{
		testSubsection("interface", "## Interface", []string{"Types."}),
	}
	pub := testPublicSection([]string{"", "  ", ""}, subs)

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", testEmptyFM(), "spec/a", pub, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "public_subsection_required")
	if count != 0 {
		t.Errorf("expected no public_subsection_required errors, got %d: %+v", count, errs)
	}
}

func TestPublicSubsectionRequired_ContentWithNoSubsections(t *testing.T) {
	pub := testPublicSection([]string{"Some content."}, nil)

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", testEmptyFM(), "spec/a", pub, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/a", "public_subsection_required")
	if count != 1 {
		t.Errorf("expected exactly 1 public_subsection_required error for SPEC/a, got %d: %+v", count, errs)
	}
}

func TestPublicSubsectionRequired_OnlySubsections_NoError(t *testing.T) {
	subs := []*parsenode.NodeSubsection{
		testSubsection("interface", "## Interface", []string{"Types."}),
	}
	pub := testPublicSection([]string{}, subs)

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", testEmptyFM(), "spec/a", pub, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "public_subsection_required")
	if count != 0 {
		t.Errorf("expected no public_subsection_required errors, got %d: %+v", count, errs)
	}
}

func TestPublicSubsectionRequired_NoPublicSection_Skip(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", testEmptyFM(), "spec/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "public_subsection_required")
	if count != 0 {
		t.Errorf("expected no public_subsection_required errors, got %d: %+v", count, errs)
	}
}

func TestDuplicateSubsections_UniqueHeadings_NoError(t *testing.T) {
	subs := []*parsenode.NodeSubsection{
		testSubsection("interface", "## Interface", []string{"Types."}),
		testSubsection("context", "## Context", []string{"Background."}),
	}
	pub := testPublicSection([]string{}, subs)

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", testEmptyFM(), "spec/a", pub, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "duplicate_subsections")
	if count != 0 {
		t.Errorf("expected no duplicate_subsections errors, got %d: %+v", count, errs)
	}
}

func TestDuplicateSubsections_DuplicateHeadings(t *testing.T) {
	subs := []*parsenode.NodeSubsection{
		testSubsection("interface", "## Interface", []string{"First."}),
		testSubsection("interface", "## Interface", []string{"Second."}),
	}
	pub := testPublicSection([]string{}, subs)

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", testEmptyFM(), "spec/a", pub, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/a", "duplicate_subsections")
	if count != 1 {
		t.Errorf("expected exactly 1 duplicate_subsections error for SPEC/a, got %d: %+v", count, errs)
	}
}

func TestDuplicateSubsections_ThreeIdenticalHeadings(t *testing.T) {
	subs := []*parsenode.NodeSubsection{
		testSubsection("interface", "## Interface", []string{"First."}),
		testSubsection("interface", "## Interface", []string{"Second."}),
		testSubsection("interface", "## Interface", []string{"Third."}),
	}
	pub := testPublicSection([]string{}, subs)

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", testEmptyFM(), "spec/a", pub, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "SPEC/a", "duplicate_subsections")
	if count != 2 {
		t.Errorf("expected exactly 2 duplicate_subsections errors for SPEC/a, got %d: %+v", count, errs)
	}
}

func TestDuplicateSubsections_NoPublicSection_Skip(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		testEntry("SPEC/a", testEmptyFM(), "spec/a", nil, nil),
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	count := testCountErrors(errs, "", "duplicate_subsections")
	if count != 0 {
		t.Errorf("expected no duplicate_subsections errors, got %d: %+v", count, errs)
	}
}

func TestCrossCutting_CollectsMultipleErrorsFromDifferentRules(t *testing.T) {
	subs := []*parsenode.NodeSubsection{
		testSubsection("interface", "## Interface", []string{"First."}),
		testSubsection("interface", "## Interface", []string{"Second."}),
	}
	pub := testPublicSection([]string{}, subs)

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("SPEC", testEmptyFM(), "spec", nil, nil),
		{
			LogicalName: "SPEC/a",
			Frontmatter: &frontmatter.Frontmatter{DependsOn: []string{"SPEC/missing"}},
			Node: &parsenode.Node{
				NameSection: testNodeSection("spec/wrong"),
				Public:      pub,
				Agent:       nil,
			},
		},
	}
	allDirs := []string{"code-from-spec", "code-from-spec/a"}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)

	if testCountErrors(errs, "SPEC/a", "name_heading") < 1 {
		t.Errorf("expected at least 1 name_heading error for SPEC/a: %+v", errs)
	}
	if testCountErrors(errs, "SPEC/a", "dependency_targets") < 1 {
		t.Errorf("expected at least 1 dependency_targets error for SPEC/a: %+v", errs)
	}
	if testCountErrors(errs, "SPEC/a", "duplicate_subsections") < 1 {
		t.Errorf("expected at least 1 duplicate_subsections error for SPEC/a: %+v", errs)
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
