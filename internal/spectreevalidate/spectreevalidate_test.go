package spectreevalidate_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/spectreevalidate"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"
)

func makeNode(logicalName string, parentName *string) parsing.Node {
	normalized := strings.ToLower(logicalName)
	return parsing.Node{
		Reference: parsing.CfsReference{
			LogicalName: logicalName,
			ParentName:  parentName,
		},
		NameSection: parsing.NodeSection{
			Heading:    normalized,
			RawHeading: "# " + logicalName,
			Content:    []string{},
		},
	}
}

func makeNodeWithHeading(logicalName string, parentName *string, heading string) parsing.Node {
	n := makeNode(logicalName, parentName)
	n.NameSection.Heading = heading
	return n
}

func makeNodeWithFrontmatter(logicalName string, parentName *string, fm *parsing.NodeFrontmatter) parsing.Node {
	n := makeNode(logicalName, parentName)
	n.Frontmatter = fm
	return n
}

func withPublic(n parsing.Node, content []string, subsections []*parsing.NodeSubsection) parsing.Node {
	n.Public = &parsing.NodeSection{
		Heading:     "public",
		RawHeading:  "# Public",
		Content:     content,
		Subsections: subsections,
	}
	return n
}

func withAgent(n parsing.Node, content []string) parsing.Node {
	n.Agent = &parsing.NodeSection{
		Heading:    "agent",
		RawHeading: "# Agent",
		Content:    content,
	}
	return n
}

func findErrors(errs []spectreevalidate.FormatError, node, rule string) []spectreevalidate.FormatError {
	var found []spectreevalidate.FormatError
	for _, e := range errs {
		if e.Node == node && e.Rule == rule {
			found = append(found, e)
		}
	}
	return found
}

func hasError(errs []spectreevalidate.FormatError, node, rule string) bool {
	return len(findErrors(errs, node, rule)) > 0
}

func TestHappyPath_ValidLeafNode(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		DependsOn: []string{"SPEC/root/b"},
		Output:    testutils.Ptr("internal/out.go"),
	})
	nodeB := makeNode("SPEC/root/b", testutils.Ptr("SPEC/root"))

	entries := []parsing.Node{rootNode, nodeA, nodeB}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
		"code-from-spec/root/b",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestHappyPath_ValidIntermediateNode(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	rootNode = withPublic(rootNode, []string{}, nil)

	nodeA := makeNode("SPEC/root/a", testutils.Ptr("SPEC/root"))

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestHappyPath_LeafWithNoFrontmatterFields(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{})

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestNameHeading_Matches(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNode("SPEC/root/a", testutils.Ptr("SPEC/root"))

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if hasError(errs, "SPEC/root/a", "name_heading") {
		t.Errorf("expected no name_heading error for matching heading")
	}
}

func TestNameHeading_DoesNotMatch(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithHeading("SPEC/root/a", testutils.Ptr("SPEC/root"), "spec/wrong")

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if !hasError(errs, "SPEC/root/a", "name_heading") {
		t.Errorf("expected name_heading error, got %v", errs)
	}
}

func TestLeafOnlyFields_IntermediateWithDependsOn(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		DependsOn: []string{"SPEC/root/b"},
	})
	nodeAB := makeNode("SPEC/root/a/b", testutils.Ptr("SPEC/root/a"))

	entries := []parsing.Node{rootNode, nodeA, nodeAB}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
		"code-from-spec/root/a/b",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if !hasError(errs, "SPEC/root/a", "leaf_only_fields") {
		t.Errorf("expected leaf_only_fields error, got %v", errs)
	}
}

func TestLeafOnlyFields_IntermediateWithOutput(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		Output: testutils.Ptr("x.go"),
	})
	nodeAB := makeNode("SPEC/root/a/b", testutils.Ptr("SPEC/root/a"))

	entries := []parsing.Node{rootNode, nodeA, nodeAB}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
		"code-from-spec/root/a/b",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if !hasError(errs, "SPEC/root/a", "leaf_only_fields") {
		t.Errorf("expected leaf_only_fields error, got %v", errs)
	}
}

func TestLeafOnlyFields_IntermediateWithInput(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		Input: testutils.Ptr("ARTIFACT/root/c"),
	})
	nodeAB := makeNode("SPEC/root/a/b", testutils.Ptr("SPEC/root/a"))

	entries := []parsing.Node{rootNode, nodeA, nodeAB}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
		"code-from-spec/root/a/b",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if !hasError(errs, "SPEC/root/a", "leaf_only_fields") {
		t.Errorf("expected leaf_only_fields error, got %v", errs)
	}
}

func TestLeafOnlyFields_IntermediateWithMultipleFields(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		DependsOn: []string{"SPEC/root/b"},
		Output:    testutils.Ptr("x.go"),
	})
	nodeAB := makeNode("SPEC/root/a/b", testutils.Ptr("SPEC/root/a"))

	entries := []parsing.Node{rootNode, nodeA, nodeAB}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
		"code-from-spec/root/a/b",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	found := findErrors(errs, "SPEC/root/a", "leaf_only_fields")
	if len(found) != 2 {
		t.Errorf("expected 2 leaf_only_fields errors, got %d: %v", len(found), errs)
	}
}

func TestLeafOnlyAgent_IntermediateWithAgent(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := withAgent(makeNode("SPEC/root/a", testutils.Ptr("SPEC/root")), []string{"some content"})
	nodeAB := makeNode("SPEC/root/a/b", testutils.Ptr("SPEC/root/a"))

	entries := []parsing.Node{rootNode, nodeA, nodeAB}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
		"code-from-spec/root/a/b",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if !hasError(errs, "SPEC/root/a", "leaf_only_agent") {
		t.Errorf("expected leaf_only_agent error, got %v", errs)
	}
}

func TestLeafOnlyAgent_LeafWithAgent_NoError(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := withAgent(makeNode("SPEC/root/a", testutils.Ptr("SPEC/root")), []string{"some content"})

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if hasError(errs, "SPEC/root/a", "leaf_only_agent") {
		t.Errorf("expected no leaf_only_agent error for leaf node")
	}
}

func TestDependencyTargets_NonExistentSPEC(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		DependsOn: []string{"SPEC/root/missing"},
	})

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if !hasError(errs, "SPEC/root/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error, got %v", errs)
	}
}

func TestDependencyTargets_TargetsAncestor(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		DependsOn: []string{"SPEC/root"},
	})

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if !hasError(errs, "SPEC/root/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ancestor, got %v", errs)
	}
}

func TestDependencyTargets_TargetsDescendant(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		DependsOn: []string{"SPEC/root/a/b"},
	})
	nodeAB := makeNode("SPEC/root/a/b", testutils.Ptr("SPEC/root/a"))

	entries := []parsing.Node{rootNode, nodeA, nodeAB}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
		"code-from-spec/root/a/b",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if !hasError(errs, "SPEC/root/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for descendant, got %v", errs)
	}
}

func TestDependencyTargets_TargetsSelf(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		DependsOn: []string{"SPEC/root/a"},
	})

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if !hasError(errs, "SPEC/root/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for self, got %v", errs)
	}
}

func TestDependencyTargets_ValidSPECWithQualifier(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNode("SPEC/root/a", testutils.Ptr("SPEC/root"))
	nodeB := makeNodeWithFrontmatter("SPEC/root/b", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		DependsOn: []string{"SPEC/root/a(interface)"},
	})

	entries := []parsing.Node{rootNode, nodeA, nodeB}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
		"code-from-spec/root/b",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if hasError(errs, "SPEC/root/b", "dependency_targets") {
		t.Errorf("expected no dependency_targets error for valid qualified SPEC reference")
	}
}

func TestDependencyTargets_ValidARTIFACT(t *testing.T) {
	testutils.Chdir(t)

	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		Output: testutils.Ptr("lib.go"),
	})
	nodeB := makeNodeWithFrontmatter("SPEC/root/b", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		DependsOn: []string{"ARTIFACT/root/a"},
	})

	b := testutils.CreateSpecNode(t, "SPEC/root/a")
	b.SetOutput("lib.go")
	b.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()
	testutils.CreateSpecNode(t, "SPEC/root").Write()

	entries := []parsing.Node{rootNode, nodeA, nodeB}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
		"code-from-spec/root/b",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if hasError(errs, "SPEC/root/b", "dependency_targets") {
		t.Errorf("expected no dependency_targets error for valid ARTIFACT reference, got %v", errs)
	}
}

func TestDependencyTargets_NonExistentARTIFACT(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	testutils.CreateSpecNode(t, "SPEC/root/a").Write()

	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		DependsOn: []string{"ARTIFACT/root/missing"},
	})

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if !hasError(errs, "SPEC/root/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for non-existent ARTIFACT, got %v", errs)
	}
}

func TestDependencyTargets_ValidEXTERNAL(t *testing.T) {
	testutils.Chdir(t)

	if err := os.MkdirAll("proto", 0755); err != nil {
		t.Fatalf("mkdir proto: %v", err)
	}
	if err := os.WriteFile(filepath.Join("proto", "api.proto"), []byte("syntax = \"proto3\";"), 0644); err != nil {
		t.Fatalf("write api.proto: %v", err)
	}

	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		DependsOn: []string{"EXTERNAL/proto/api.proto"},
	})

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if hasError(errs, "SPEC/root/a", "dependency_targets") {
		t.Errorf("expected no dependency_targets error for valid EXTERNAL reference, got %v", errs)
	}
}

func TestDependencyTargets_NonExistentEXTERNAL(t *testing.T) {
	testutils.Chdir(t)

	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		DependsOn: []string{"EXTERNAL/nonexistent.txt"},
	})

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if !hasError(errs, "SPEC/root/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for non-existent EXTERNAL, got %v", errs)
	}
}

func TestDependencyTargets_UnrecognizedPrefix(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		DependsOn: []string{"UNKNOWN/something"},
	})

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if !hasError(errs, "SPEC/root/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for unrecognized prefix, got %v", errs)
	}
}

func TestDependencyTargets_MultipleInvalidEntries(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		DependsOn: []string{"SPEC/root/missing", "SPEC/root/also_missing"},
	})

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	found := findErrors(errs, "SPEC/root/a", "dependency_targets")
	if len(found) != 2 {
		t.Errorf("expected 2 dependency_targets errors, got %d: %v", len(found), errs)
	}
}

func TestInputTarget_ValidARTIFACT(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/root/a")
	b.SetOutput("a.go")
	b.Write()
	testutils.CreateSpecNode(t, "SPEC/root/b").Write()
	testutils.CreateSpecNode(t, "SPEC/root").Write()

	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		Output: testutils.Ptr("a.go"),
	})
	nodeB := makeNodeWithFrontmatter("SPEC/root/b", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		Input: testutils.Ptr("ARTIFACT/root/a"),
	})

	entries := []parsing.Node{rootNode, nodeA, nodeB}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
		"code-from-spec/root/b",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if hasError(errs, "SPEC/root/b", "input_target") {
		t.Errorf("expected no input_target error for valid ARTIFACT, got %v", errs)
	}
}

func TestInputTarget_ValidEXTERNAL(t *testing.T) {
	testutils.Chdir(t)

	if err := os.MkdirAll("docs", 0755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}
	if err := os.WriteFile(filepath.Join("docs", "spec.yaml"), []byte("---"), 0644); err != nil {
		t.Fatalf("write spec.yaml: %v", err)
	}

	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		Input: testutils.Ptr("EXTERNAL/docs/spec.yaml"),
	})

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if hasError(errs, "SPEC/root/a", "input_target") {
		t.Errorf("expected no input_target error for valid EXTERNAL, got %v", errs)
	}
}

func TestInputTarget_ValidSPEC(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNode("SPEC/root/a", testutils.Ptr("SPEC/root"))
	nodeB := makeNodeWithFrontmatter("SPEC/root/b", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		Input: testutils.Ptr("SPEC/root/a"),
	})

	entries := []parsing.Node{rootNode, nodeA, nodeB}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
		"code-from-spec/root/b",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if hasError(errs, "SPEC/root/b", "input_target") {
		t.Errorf("expected no input_target error for valid SPEC, got %v", errs)
	}
}

func TestInputTarget_ValidSPECWithQualifier(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNode("SPEC/root/a", testutils.Ptr("SPEC/root"))
	nodeB := makeNodeWithFrontmatter("SPEC/root/b", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		Input: testutils.Ptr("SPEC/root/a(acceptance-tests)"),
	})

	entries := []parsing.Node{rootNode, nodeA, nodeB}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
		"code-from-spec/root/b",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if hasError(errs, "SPEC/root/b", "input_target") {
		t.Errorf("expected no input_target error for valid SPEC with qualifier, got %v", errs)
	}
}

func TestInputTarget_NonExistentSPEC(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		Input: testutils.Ptr("SPEC/root/missing"),
	})

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if !hasError(errs, "SPEC/root/a", "input_target") {
		t.Errorf("expected input_target error for non-existent SPEC, got %v", errs)
	}
}

func TestInputTarget_UnsupportedPrefix(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		Input: testutils.Ptr("UNKNOWN/something"),
	})

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if !hasError(errs, "SPEC/root/a", "input_target") {
		t.Errorf("expected input_target error for unsupported prefix, got %v", errs)
	}
}

func TestInputTarget_NonExistentARTIFACT(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()
	testutils.CreateSpecNode(t, "SPEC/root/a").Write()

	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		Input: testutils.Ptr("ARTIFACT/root/missing"),
	})

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if !hasError(errs, "SPEC/root/a", "input_target") {
		t.Errorf("expected input_target error for non-existent ARTIFACT, got %v", errs)
	}
}

func TestInputTarget_NonExistentEXTERNAL(t *testing.T) {
	testutils.Chdir(t)

	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		Input: testutils.Ptr("EXTERNAL/nonexistent.txt"),
	})

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if !hasError(errs, "SPEC/root/a", "input_target") {
		t.Errorf("expected input_target error for non-existent EXTERNAL, got %v", errs)
	}
}

func TestMissingNodeMd_SubdirWithoutNode(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNode("SPEC/root/a", testutils.Ptr("SPEC/root"))

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
		"code-from-spec/root/b",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if !hasError(errs, "code-from-spec/root/b", "missing_node_md") {
		t.Errorf("expected missing_node_md error for b, got %v", errs)
	}
}

func TestMissingNodeMd_DotPrefixedDirectly_NoError(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)

	entries := []parsing.Node{rootNode}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/.cache",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if hasError(errs, "code-from-spec/.cache", "missing_node_md") {
		t.Errorf("expected no missing_node_md for .-prefixed dir")
	}
}

func TestMissingNodeMd_DotPrefixedDeeper_NoError(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNode("SPEC/root/a", testutils.Ptr("SPEC/root"))

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
		"code-from-spec/root/a/.internal",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if hasError(errs, "code-from-spec/root/a/.internal", "missing_node_md") {
		t.Errorf("expected no missing_node_md for .-prefixed deeper dir")
	}
}

func TestMissingNodeMd_AllHaveNodes_NoError(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNode("SPEC/root/a", testutils.Ptr("SPEC/root"))
	nodeB := makeNode("SPEC/root/b", testutils.Ptr("SPEC/root"))

	entries := []parsing.Node{rootNode, nodeA, nodeB}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
		"code-from-spec/root/b",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if hasError(errs, "", "missing_node_md") {
		for _, e := range errs {
			if e.Rule == "missing_node_md" {
				t.Errorf("unexpected missing_node_md error: %v", e)
			}
		}
	}
}

func TestOutputPaths_ValidPath(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		Output: testutils.Ptr("internal/x.go"),
	})

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if hasError(errs, "SPEC/root/a", "output_paths") {
		t.Errorf("expected no output_paths error for valid path")
	}
}

func TestOutputPaths_TraversalPath(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		Output: testutils.Ptr("../../etc/passwd"),
	})

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if !hasError(errs, "SPEC/root/a", "output_paths") {
		t.Errorf("expected output_paths error for traversal path, got %v", errs)
	}
}

func TestOutputPaths_BackslashPath(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNodeWithFrontmatter("SPEC/root/a", testutils.Ptr("SPEC/root"), &parsing.NodeFrontmatter{
		Output: testutils.Ptr(`internal\x.go`),
	})

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if !hasError(errs, "SPEC/root/a", "output_paths") {
		t.Errorf("expected output_paths error for backslash path, got %v", errs)
	}
}

func TestPublicSubsectionRequired_ContentBeforeSubsection(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := withPublic(
		makeNode("SPEC/root/a", testutils.Ptr("SPEC/root")),
		[]string{"Some loose content."},
		[]*parsing.NodeSubsection{
			{
				Heading:    "interface",
				RawHeading: "## Interface",
				Content:    []string{"Types."},
			},
		},
	)

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	found := findErrors(errs, "SPEC/root/a", "public_subsection_required")
	if len(found) == 0 {
		t.Errorf("expected public_subsection_required error, got %v", errs)
	}
	if found[0].Detail == "" {
		t.Errorf("expected non-empty Detail in public_subsection_required error")
	}
}

func TestPublicSubsectionRequired_BlankLinesBeforeSubsection_NoError(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := withPublic(
		makeNode("SPEC/root/a", testutils.Ptr("SPEC/root")),
		[]string{"", "  ", ""},
		[]*parsing.NodeSubsection{
			{
				Heading:    "interface",
				RawHeading: "## Interface",
				Content:    []string{"Types."},
			},
		},
	)

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if hasError(errs, "SPEC/root/a", "public_subsection_required") {
		t.Errorf("expected no public_subsection_required error for blank lines before subsection")
	}
}

func TestPublicSubsectionRequired_ContentNoSubsections(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := withPublic(
		makeNode("SPEC/root/a", testutils.Ptr("SPEC/root")),
		[]string{"Some content."},
		[]*parsing.NodeSubsection{},
	)

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if !hasError(errs, "SPEC/root/a", "public_subsection_required") {
		t.Errorf("expected public_subsection_required error for content with no subsections, got %v", errs)
	}
}

func TestPublicSubsectionRequired_OnlySubsections_NoError(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := withPublic(
		makeNode("SPEC/root/a", testutils.Ptr("SPEC/root")),
		[]string{},
		[]*parsing.NodeSubsection{
			{
				Heading:    "interface",
				RawHeading: "## Interface",
				Content:    []string{"Types."},
			},
		},
	)

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if hasError(errs, "SPEC/root/a", "public_subsection_required") {
		t.Errorf("expected no public_subsection_required error for only-subsections public section")
	}
}

func TestPublicSubsectionRequired_NoPublicSection_Skip(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNode("SPEC/root/a", testutils.Ptr("SPEC/root"))

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if hasError(errs, "SPEC/root/a", "public_subsection_required") {
		t.Errorf("expected no public_subsection_required error when no public section")
	}
}

func TestDuplicateSubsections_Unique_NoError(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := withPublic(
		makeNode("SPEC/root/a", testutils.Ptr("SPEC/root")),
		[]string{},
		[]*parsing.NodeSubsection{
			{Heading: "interface", RawHeading: "## Interface", Content: []string{}},
			{Heading: "context", RawHeading: "## Context", Content: []string{}},
		},
	)

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if hasError(errs, "SPEC/root/a", "duplicate_subsections") {
		t.Errorf("expected no duplicate_subsections error for unique headings")
	}
}

func TestDuplicateSubsections_Duplicate(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := withPublic(
		makeNode("SPEC/root/a", testutils.Ptr("SPEC/root")),
		[]string{},
		[]*parsing.NodeSubsection{
			{Heading: "interface", RawHeading: "## Interface", Content: []string{"First."}},
			{Heading: "interface", RawHeading: "## Interface", Content: []string{"Second."}},
		},
	)

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	found := findErrors(errs, "SPEC/root/a", "duplicate_subsections")
	if len(found) != 1 {
		t.Errorf("expected 1 duplicate_subsections error, got %d: %v", len(found), errs)
	}
}

func TestDuplicateSubsections_ThreeIdentical(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := withPublic(
		makeNode("SPEC/root/a", testutils.Ptr("SPEC/root")),
		[]string{},
		[]*parsing.NodeSubsection{
			{Heading: "interface", RawHeading: "## Interface", Content: []string{}},
			{Heading: "interface", RawHeading: "## Interface", Content: []string{}},
			{Heading: "interface", RawHeading: "## Interface", Content: []string{}},
		},
	)

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	found := findErrors(errs, "SPEC/root/a", "duplicate_subsections")
	if len(found) != 2 {
		t.Errorf("expected 2 duplicate_subsections errors, got %d: %v", len(found), errs)
	}
}

func TestDuplicateSubsections_NoPublicSection_Skip(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := makeNode("SPEC/root/a", testutils.Ptr("SPEC/root"))

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)
	if hasError(errs, "SPEC/root/a", "duplicate_subsections") {
		t.Errorf("expected no duplicate_subsections error when no public section")
	}
}

func TestCrossCutting_MultipleErrorsFromDifferentRules(t *testing.T) {
	rootNode := makeNode("SPEC/root", nil)
	nodeA := withPublic(
		makeNodeWithHeading("SPEC/root/a", testutils.Ptr("SPEC/root"), "spec/wrong"),
		[]string{},
		[]*parsing.NodeSubsection{
			{Heading: "interface", RawHeading: "## Interface", Content: []string{"First."}},
			{Heading: "interface", RawHeading: "## Interface", Content: []string{"Second."}},
		},
	)
	nodeA.Frontmatter = &parsing.NodeFrontmatter{
		DependsOn: []string{"SPEC/root/missing"},
	}

	entries := []parsing.Node{rootNode, nodeA}
	allDirs := []string{
		"code-from-spec",
		"code-from-spec/root",
		"code-from-spec/root/a",
	}

	errs := spectreevalidate.SpecTreeValidate(entries, allDirs)

	if !hasError(errs, "SPEC/root/a", "name_heading") {
		t.Errorf("expected name_heading error")
	}
	if !hasError(errs, "SPEC/root/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error")
	}
	if !hasError(errs, "SPEC/root/a", "duplicate_subsections") {
		t.Errorf("expected duplicate_subsections error")
	}
}

func TestCrossCutting_EmptyInput(t *testing.T) {
	errs := spectreevalidate.SpecTreeValidate([]parsing.Node{}, []string{})
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty input, got %v", errs)
	}
}
