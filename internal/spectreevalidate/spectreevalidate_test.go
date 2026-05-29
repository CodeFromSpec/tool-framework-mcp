// code-from-spec: ROOT/golang/tests/spec_tree/validate@O0FwMdD2ljEi-fz2W4lgKda7xKY
package spectreevalidate_test

import (
	"crypto/sha1"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectreevalidate"
)

// testChdir changes the working directory to dir and restores it on cleanup.
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

// testMakeNode builds a minimal *parsenode.Node with the given heading.
func testMakeNode(heading string) *parsenode.Node {
	return &parsenode.Node{
		NameSection: &parsenode.NodeSection{
			Heading:    heading,
			RawHeading: "# " + heading,
			Content:    nil,
		},
	}
}

// testMakeNodeWithPublic builds a *parsenode.Node that also has a Public section
// containing the given subsections.
func testMakeNodeWithPublic(heading string, subsections []*parsenode.NodeSubsection) *parsenode.Node {
	n := testMakeNode(heading)
	n.Public = &parsenode.NodeSection{
		Heading:     "public",
		RawHeading:  "# Public",
		Subsections: subsections,
	}
	return n
}

// testMakeNodeWithAgent builds a *parsenode.Node that also has an Agent section.
func testMakeNodeWithAgent(heading string) *parsenode.Node {
	n := testMakeNode(heading)
	n.Agent = &parsenode.NodeSection{
		Heading:    "agent",
		RawHeading: "# Agent",
	}
	return n
}

// testMakeSubsection creates a NodeSubsection with the given heading.
func testMakeSubsection(heading string) *parsenode.NodeSubsection {
	return &parsenode.NodeSubsection{
		Heading:    heading,
		RawHeading: "## " + heading,
	}
}

// testMakeFM builds a *frontmatter.Frontmatter with all fields empty.
func testMakeFM() *frontmatter.Frontmatter {
	return &frontmatter.Frontmatter{}
}

// testEntry creates a *spectreevalidate.SpecTreeValidateInput.
func testEntry(logicalName string, fm *frontmatter.Frontmatter, node *parsenode.Node) *spectreevalidate.SpecTreeValidateInput {
	return &spectreevalidate.SpecTreeValidateInput{
		LogicalName: logicalName,
		Frontmatter: fm,
		Node:        node,
	}
}

// testCountErrors counts errors matching the given node and rule (empty string matches any).
func testCountErrors(errs []*spectreevalidate.FormatError, node, rule string) int {
	count := 0
	for _, e := range errs {
		nodeMatch := node == "" || e.Node == node
		ruleMatch := rule == "" || e.Rule == rule
		if nodeMatch && ruleMatch {
			count++
		}
	}
	return count
}

// testHasError reports whether at least one error matches node and rule.
func testHasError(errs []*spectreevalidate.FormatError, node, rule string) bool {
	return testCountErrors(errs, node, rule) > 0
}

// testSha1Base64URL computes SHA-1 of data and returns base64url (no padding).
func testSha1Base64URL(data []byte) string {
	sum := sha1.Sum(data)
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// testWriteLines writes lines to path (relative), creating directories as needed.
func testWriteLines(t *testing.T, path string, lines []string) {
	t.Helper()
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("testWriteLines MkdirAll: %v", err)
		}
	}
	var content []byte
	for _, l := range lines {
		content = append(content, []byte(l+"\n")...)
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("testWriteLines WriteFile: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Happy Path
// ---------------------------------------------------------------------------

func TestSpecTreeValidate_HP1_ValidLeafNodePassesAllChecks(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/b"},
			Outputs:   []*frontmatter.FrontmatterOutput{{ID: "out", Path: "out.go"}},
		}, testMakeNode("ROOT/a")),
		testEntry("ROOT/b", testMakeFM(), testMakeNode("ROOT/b")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
	}
}

func TestSpecTreeValidate_HP2_ValidIntermediateNodePassesAllChecks(t *testing.T) {
	rootNode := testMakeNodeWithPublic("ROOT", []*parsenode.NodeSubsection{
		testMakeSubsection("Overview"),
		testMakeSubsection("Details"),
	})
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), rootNode),
		testEntry("ROOT/a", testMakeFM(), testMakeNode("ROOT/a")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
	}
}

func TestSpecTreeValidate_HP3_LeafWithNoFrontmatterFields(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testMakeFM(), testMakeNode("ROOT/a")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
	}
}

// ---------------------------------------------------------------------------
// Rule: name_heading
// ---------------------------------------------------------------------------

func TestSpecTreeValidate_NH1_HeadingMatchesLogicalName_NoError(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testMakeFM(), testMakeNode("ROOT/a")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "", "name_heading") {
		t.Errorf("expected no name_heading errors, got: %+v", errs)
	}
}

func TestSpecTreeValidate_NH2_HeadingDoesNotMatchLogicalName(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testMakeFM(), testMakeNode("ROOT/wrong")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "name_heading") {
		t.Errorf("expected name_heading error for ROOT/a, got: %+v", errs)
	}
}

// ---------------------------------------------------------------------------
// Rule: leaf_only_fields
// ---------------------------------------------------------------------------

func TestSpecTreeValidate_LOF1_IntermediateNodeWithDependsOn(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{DependsOn: []string{"ROOT/b"}}, testMakeNode("ROOT/a")),
		testEntry("ROOT/a/b", testMakeFM(), testMakeNode("ROOT/a/b")),
		testEntry("ROOT/b", testMakeFM(), testMakeNode("ROOT/b")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected leaf_only_fields error for ROOT/a, got: %+v", errs)
	}
}

func TestSpecTreeValidate_LOF2_IntermediateNodeWithOutputs(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			Outputs: []*frontmatter.FrontmatterOutput{{ID: "x", Path: "x.go"}},
		}, testMakeNode("ROOT/a")),
		testEntry("ROOT/a/b", testMakeFM(), testMakeNode("ROOT/a/b")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected leaf_only_fields error for ROOT/a, got: %+v", errs)
	}
}

func TestSpecTreeValidate_LOF3_IntermediateNodeWithInput(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{Input: "ARTIFACT/c(id)"}, testMakeNode("ROOT/a")),
		testEntry("ROOT/a/b", testMakeFM(), testMakeNode("ROOT/a/b")),
		testEntry("ROOT/c", &frontmatter.Frontmatter{
			Outputs: []*frontmatter.FrontmatterOutput{{ID: "id", Path: "c.go"}},
		}, testMakeNode("ROOT/c")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected leaf_only_fields error for ROOT/a, got: %+v", errs)
	}
}

func TestSpecTreeValidate_LOF4_IntermediateNodeWithExternal(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteLines(t, "some/file.txt", []string{"content"})

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			External: []*frontmatter.FrontmatterExternal{{Path: "some/file.txt"}},
		}, testMakeNode("ROOT/a")),
		testEntry("ROOT/a/b", testMakeFM(), testMakeNode("ROOT/a/b")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected leaf_only_fields error for ROOT/a, got: %+v", errs)
	}
}

func TestSpecTreeValidate_LOF5_IntermediateNodeWithMultipleRestrictedFields_OneErrorPerField(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/b"},
			Outputs:   []*frontmatter.FrontmatterOutput{{ID: "x", Path: "x.go"}},
		}, testMakeNode("ROOT/a")),
		testEntry("ROOT/a/b", testMakeFM(), testMakeNode("ROOT/a/b")),
		testEntry("ROOT/b", testMakeFM(), testMakeNode("ROOT/b")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	count := testCountErrors(errs, "ROOT/a", "leaf_only_fields")
	if count < 2 {
		t.Errorf("expected at least 2 leaf_only_fields errors for ROOT/a, got %d: %+v", count, errs)
	}
}

// ---------------------------------------------------------------------------
// Rule: leaf_only_agent
// ---------------------------------------------------------------------------

func TestSpecTreeValidate_LOA1_IntermediateNodeWithAgentSection(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testMakeFM(), testMakeNodeWithAgent("ROOT/a")),
		testEntry("ROOT/a/b", testMakeFM(), testMakeNode("ROOT/a/b")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "leaf_only_agent") {
		t.Errorf("expected leaf_only_agent error for ROOT/a, got: %+v", errs)
	}
}

func TestSpecTreeValidate_LOA2_LeafNodeWithAgentSection_NoError(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testMakeFM(), testMakeNodeWithAgent("ROOT/a")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "", "leaf_only_agent") {
		t.Errorf("expected no leaf_only_agent errors, got: %+v", errs)
	}
}

// ---------------------------------------------------------------------------
// Rule: dependency_targets
// ---------------------------------------------------------------------------

func TestSpecTreeValidate_DT1_DependsOnTargetsNonExistentROOTNode(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{DependsOn: []string{"ROOT/missing"}}, testMakeNode("ROOT/a")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a, got: %+v", errs)
	}
}

func TestSpecTreeValidate_DT2_DependsOnTargetsAncestor(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testMakeFM(), testMakeNode("ROOT/a")),
		testEntry("ROOT/a/b", &frontmatter.Frontmatter{DependsOn: []string{"ROOT"}}, testMakeNode("ROOT/a/b")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a/b", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a/b, got: %+v", errs)
	}
}

func TestSpecTreeValidate_DT3_DependsOnTargetsDescendant(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{DependsOn: []string{"ROOT/a/b"}}, testMakeNode("ROOT/a")),
		testEntry("ROOT/a/b", testMakeFM(), testMakeNode("ROOT/a/b")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a, got: %+v", errs)
	}
}

func TestSpecTreeValidate_DT4_DependsOnTargetsSelf(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{DependsOn: []string{"ROOT/a"}}, testMakeNode("ROOT/a")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a, got: %+v", errs)
	}
}

func TestSpecTreeValidate_DT5_DependsOnWithValidROOTQualifier_NoError(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testMakeFM(), testMakeNode("ROOT/a")),
		testEntry("ROOT/b", &frontmatter.Frontmatter{DependsOn: []string{"ROOT/a(interface)"}}, testMakeNode("ROOT/b")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "", "dependency_targets") {
		t.Errorf("expected no dependency_targets errors, got: %+v", errs)
	}
}

func TestSpecTreeValidate_DT6_DependsOnWithValidARTIFACTReference_NoError(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			Outputs: []*frontmatter.FrontmatterOutput{{ID: "lib", Path: "lib.go"}},
		}, testMakeNode("ROOT/a")),
		testEntry("ROOT/b", &frontmatter.Frontmatter{DependsOn: []string{"ARTIFACT/a(lib)"}}, testMakeNode("ROOT/b")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "", "dependency_targets") {
		t.Errorf("expected no dependency_targets errors, got: %+v", errs)
	}
}

func TestSpecTreeValidate_DT7_DependsOnWithNonExistentARTIFACTReference(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{DependsOn: []string{"ARTIFACT/missing(id)"}}, testMakeNode("ROOT/a")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a, got: %+v", errs)
	}
}

func TestSpecTreeValidate_DT8_MultipleInvalidDependsOn_OneErrorPerEntry(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/missing", "ROOT/also_missing"},
		}, testMakeNode("ROOT/a")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	count := testCountErrors(errs, "ROOT/a", "dependency_targets")
	if count < 2 {
		t.Errorf("expected at least 2 dependency_targets errors for ROOT/a, got %d: %+v", count, errs)
	}
}

// ---------------------------------------------------------------------------
// Rule: input_target
// ---------------------------------------------------------------------------

func TestSpecTreeValidate_IT1_ValidInputReference_NoError(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			Outputs: []*frontmatter.FrontmatterOutput{{ID: "out", Path: "a.go"}},
		}, testMakeNode("ROOT/a")),
		testEntry("ROOT/b", &frontmatter.Frontmatter{Input: "ARTIFACT/a(out)"}, testMakeNode("ROOT/b")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "", "input_target") {
		t.Errorf("expected no input_target errors, got: %+v", errs)
	}
}

func TestSpecTreeValidate_IT2_InputNotStartingWithARTIFACT(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{Input: "ROOT/something"}, testMakeNode("ROOT/a")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "input_target") {
		t.Errorf("expected input_target error for ROOT/a, got: %+v", errs)
	}
}

func TestSpecTreeValidate_IT3_InputReferencesNonExistentArtifact(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{Input: "ARTIFACT/missing(id)"}, testMakeNode("ROOT/a")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "input_target") {
		t.Errorf("expected input_target error for ROOT/a, got: %+v", errs)
	}
}

// ---------------------------------------------------------------------------
// Rule: external_files
// ---------------------------------------------------------------------------

func TestSpecTreeValidate_EF1_ExternalFileExists_NoFragments(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteLines(t, "some/file.txt", []string{"hello"})

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			External: []*frontmatter.FrontmatterExternal{{Path: "some/file.txt"}},
		}, testMakeNode("ROOT/a")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "", "external_files") {
		t.Errorf("expected no external_files errors, got: %+v", errs)
	}
}

func TestSpecTreeValidate_EF2_ExternalFileDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	// Do NOT create "nonexistent.txt"

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			External: []*frontmatter.FrontmatterExternal{{Path: "nonexistent.txt"}},
		}, testMakeNode("ROOT/a")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("expected external_files error for ROOT/a, got: %+v", errs)
	}
}

func TestSpecTreeValidate_EF3_FragmentWithValidHash(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteLines(t, "f.txt", []string{"alpha", "beta", "gamma", "delta", "epsilon"})

	// Compute correct hash for lines 1-3: "alpha\nbeta\ngamma\n"
	correctHash := testSha1Base64URL([]byte("alpha\nbeta\ngamma\n"))

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			External: []*frontmatter.FrontmatterExternal{
				{
					Path: "f.txt",
					Fragments: []*frontmatter.FrontmatterExternalFragment{
						{Lines: "1-3", Hash: correctHash},
					},
				},
			},
		}, testMakeNode("ROOT/a")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "", "external_files") {
		t.Errorf("expected no external_files errors (correct hash), got: %+v", errs)
	}
}

func TestSpecTreeValidate_EF4_FragmentWithInvalidHash(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteLines(t, "f.txt", []string{"alpha", "beta", "gamma", "delta", "epsilon"})

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			External: []*frontmatter.FrontmatterExternal{
				{
					Path: "f.txt",
					Fragments: []*frontmatter.FrontmatterExternalFragment{
						{Lines: "1-3", Hash: "wrong"},
					},
				},
			},
		}, testMakeNode("ROOT/a")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("expected external_files error for ROOT/a (wrong hash), got: %+v", errs)
	}
}

func TestSpecTreeValidate_EF5_FragmentWithInvalidRangeFormat(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteLines(t, "f.txt", []string{"line1"})

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			External: []*frontmatter.FrontmatterExternal{
				{
					Path: "f.txt",
					Fragments: []*frontmatter.FrontmatterExternalFragment{
						{Lines: "abc", Hash: "x"},
					},
				},
			},
		}, testMakeNode("ROOT/a")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("expected external_files error for ROOT/a (invalid range format), got: %+v", errs)
	}
}

func TestSpecTreeValidate_EF6_FragmentWithStartGreaterThanEnd(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteLines(t, "f.txt", []string{"a", "b", "c", "d", "e"})

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			External: []*frontmatter.FrontmatterExternal{
				{
					Path: "f.txt",
					Fragments: []*frontmatter.FrontmatterExternalFragment{
						{Lines: "5-3", Hash: "x"},
					},
				},
			},
		}, testMakeNode("ROOT/a")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("expected external_files error for ROOT/a (start > end), got: %+v", errs)
	}
}

func TestSpecTreeValidate_EF7_FragmentWithStartLessThan1(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteLines(t, "f.txt", []string{"line1"})

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			External: []*frontmatter.FrontmatterExternal{
				{
					Path: "f.txt",
					Fragments: []*frontmatter.FrontmatterExternalFragment{
						{Lines: "0-3", Hash: "x"},
					},
				},
			},
		}, testMakeNode("ROOT/a")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("expected external_files error for ROOT/a (start < 1), got: %+v", errs)
	}
}

func TestSpecTreeValidate_EF8_FragmentOutOfRange(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteLines(t, "f.txt", []string{"a", "b", "c", "d", "e"})

	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			External: []*frontmatter.FrontmatterExternal{
				{
					Path: "f.txt",
					Fragments: []*frontmatter.FrontmatterExternalFragment{
						{Lines: "1-100", Hash: "x"},
					},
				},
			},
		}, testMakeNode("ROOT/a")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("expected external_files error for ROOT/a (out of range), got: %+v", errs)
	}
	// Verify detail mentions out of range
	found := false
	for _, e := range errs {
		if e.Node == "ROOT/a" && e.Rule == "external_files" {
			found = true
			if e.Detail == "" {
				t.Errorf("expected non-empty detail for out-of-range fragment error")
			}
			break
		}
	}
	if !found {
		t.Errorf("expected external_files error for ROOT/a but none found")
	}
}

// ---------------------------------------------------------------------------
// Rule: output_paths
// ---------------------------------------------------------------------------

func TestSpecTreeValidate_OP1_ValidOutputPath_NoError(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			Outputs: []*frontmatter.FrontmatterOutput{{ID: "x", Path: "internal/x.go"}},
		}, testMakeNode("ROOT/a")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "", "output_paths") {
		t.Errorf("expected no output_paths errors, got: %+v", errs)
	}
}

func TestSpecTreeValidate_OP2_OutputPathWithTraversal(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			Outputs: []*frontmatter.FrontmatterOutput{{ID: "x", Path: "../../etc/passwd"}},
		}, testMakeNode("ROOT/a")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "output_paths") {
		t.Errorf("expected output_paths error for ROOT/a (traversal), got: %+v", errs)
	}
}

func TestSpecTreeValidate_OP3_OutputPathWithBackslash(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			Outputs: []*frontmatter.FrontmatterOutput{{ID: "x", Path: `internal\x.go`}},
		}, testMakeNode("ROOT/a")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if !testHasError(errs, "ROOT/a", "output_paths") {
		t.Errorf("expected output_paths error for ROOT/a (backslash), got: %+v", errs)
	}
}

// ---------------------------------------------------------------------------
// Rule: duplicate_subsections
// ---------------------------------------------------------------------------

func TestSpecTreeValidate_DS1_UniqueSubsectionHeadings_NoError(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testMakeFM(), testMakeNodeWithPublic("ROOT/a", []*parsenode.NodeSubsection{
			testMakeSubsection("Interface"),
			testMakeSubsection("Context"),
		})),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "", "duplicate_subsections") {
		t.Errorf("expected no duplicate_subsections errors, got: %+v", errs)
	}
}

func TestSpecTreeValidate_DS2_DuplicateSubsectionHeadings_OneError(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testMakeFM(), testMakeNodeWithPublic("ROOT/a", []*parsenode.NodeSubsection{
			testMakeSubsection("Interface"),
			testMakeSubsection("Interface"),
		})),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	count := testCountErrors(errs, "ROOT/a", "duplicate_subsections")
	if count != 1 {
		t.Errorf("expected exactly 1 duplicate_subsections error for ROOT/a, got %d: %+v", count, errs)
	}
}

func TestSpecTreeValidate_DS3_ThreeIdenticalSubsectionHeadings_TwoErrors(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testMakeFM(), testMakeNodeWithPublic("ROOT/a", []*parsenode.NodeSubsection{
			testMakeSubsection("Interface"),
			testMakeSubsection("Interface"),
			testMakeSubsection("Interface"),
		})),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	count := testCountErrors(errs, "ROOT/a", "duplicate_subsections")
	if count != 2 {
		t.Errorf("expected exactly 2 duplicate_subsections errors for ROOT/a, got %d: %+v", count, errs)
	}
}

func TestSpecTreeValidate_DS4_NoPublicSection_Skip(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", testMakeFM(), testMakeNode("ROOT/a")),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)
	if testHasError(errs, "", "duplicate_subsections") {
		t.Errorf("expected no duplicate_subsections errors, got: %+v", errs)
	}
}

// ---------------------------------------------------------------------------
// Cross-Cutting
// ---------------------------------------------------------------------------

func TestSpecTreeValidate_CC1_CollectsMultipleErrorsFromDifferentRules(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testEntry("ROOT", testMakeFM(), testMakeNode("ROOT")),
		testEntry("ROOT/a", &frontmatter.Frontmatter{
			DependsOn: []string{"ROOT/missing"},
		}, testMakeNodeWithPublic("ROOT/wrong", []*parsenode.NodeSubsection{
			testMakeSubsection("Interface"),
			testMakeSubsection("Interface"),
		})),
	}
	errs := spectreevalidate.SpecTreeValidate(entries)

	rules := make(map[string]bool)
	for _, e := range errs {
		if e.Node == "ROOT/a" || e.Rule == "name_heading" {
			rules[e.Rule] = true
		}
	}
	// name_heading fires on the node whose heading is "ROOT/wrong" but logical name is "ROOT/a"
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

func TestSpecTreeValidate_CC2_EmptyInputList(t *testing.T) {
	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{})
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty input, got %d: %+v", len(errs), errs)
	}
}
