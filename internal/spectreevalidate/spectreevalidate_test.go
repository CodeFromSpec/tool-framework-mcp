// code-from-spec: ROOT/golang/tests/spec_tree/validate@Y3vIJquzh7uwBvP8FYk7PtMrucA

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

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

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

// testStrPtr returns a pointer to the given string.
func testStrPtr(s string) *string {
	return &s
}

// testMakeEntry builds a SpecTreeValidateInput with the given logical name
// and a NodeSection whose heading equals the given heading string.
func testMakeEntry(logicalName, heading string, fm *frontmatter.Frontmatter, agent *parsenode.NodeSection, public *parsenode.NodeSection) *spectreevalidate.SpecTreeValidateInput {
	return &spectreevalidate.SpecTreeValidateInput{
		LogicalName: logicalName,
		Frontmatter: fm,
		Node: &parsenode.Node{
			NameSection: &parsenode.NodeSection{Heading: heading},
			Agent:       agent,
			Public:      public,
		},
	}
}

// testEmptyFM returns an empty Frontmatter.
func testEmptyFM() *frontmatter.Frontmatter {
	return &frontmatter.Frontmatter{}
}

// testCountErrors counts FormatErrors matching the given node and rule.
// Pass empty string to match any value for that field.
func testCountErrors(errs []*spectreevalidate.FormatError, node, rule string) int {
	count := 0
	for _, fe := range errs {
		nodeMatch := node == "" || fe.Node == node
		ruleMatch := rule == "" || fe.Rule == rule
		if nodeMatch && ruleMatch {
			count++
		}
	}
	return count
}

// testHasError returns true if any FormatError matches both node and rule.
func testHasError(errs []*spectreevalidate.FormatError, node, rule string) bool {
	return testCountErrors(errs, node, rule) > 0
}

// testComputeHash computes the base64url SHA-1 hash of the given lines joined by "\n".
func testComputeHash(lines []string) string {
	joined := ""
	for i, l := range lines {
		if i > 0 {
			joined += "\n"
		}
		joined += l
	}
	sum := sha1.Sum([]byte(joined))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// testMakeSubsection builds a NodeSubsection with the given heading.
func testMakeSubsection(heading string) *parsenode.NodeSubsection {
	return &parsenode.NodeSubsection{Heading: heading}
}

// testMakePublicWithSubs builds a NodeSection named "public" with the given subsections.
func testMakePublicWithSubs(subs ...*parsenode.NodeSubsection) *parsenode.NodeSection {
	return &parsenode.NodeSection{
		Heading:     "public",
		Subsections: subs,
	}
}

// testAgentSection returns a non-nil Agent NodeSection.
func testAgentSection() *parsenode.NodeSection {
	return &parsenode.NodeSection{Heading: "agent"}
}

// ---------------------------------------------------------------------------
// TC-HP: Happy Path
// ---------------------------------------------------------------------------

func TestHappyPath_ValidLeafNode(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", testEmptyFM(), nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
	}
}

func TestHappyPath_ValidIntermediateNode(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", testEmptyFM(), nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
	}
}

func TestHappyPath_LeafWithNoFrontmatterFields(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", testEmptyFM(), nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %+v", len(errs), errs)
	}
}

// ---------------------------------------------------------------------------
// TC-NH: Rule name_heading
// ---------------------------------------------------------------------------

func TestNameHeading_Matches(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", testEmptyFM(), nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if testHasError(errs, "ROOT/a", "name_heading") {
		t.Errorf("unexpected name_heading error for ROOT/a")
	}
}

func TestNameHeading_DoesNotMatch(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/wrong", testEmptyFM(), nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !testHasError(errs, "ROOT/a", "name_heading") {
		t.Errorf("expected name_heading error for ROOT/a, got: %+v", errs)
	}
}

// ---------------------------------------------------------------------------
// TC-LOF: Rule leaf_only_fields
// ---------------------------------------------------------------------------

func TestLeafOnlyFields_IntermediateWithDependsOn(t *testing.T) {
	fm := &frontmatter.Frontmatter{
		DependsOn: []*string{testStrPtr("ROOT/b")},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fm, nil, nil),
		testMakeEntry("ROOT/a/b", "ROOT/a/b", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/b", "ROOT/b", testEmptyFM(), nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected leaf_only_fields error for ROOT/a, got: %+v", errs)
	}
}

func TestLeafOnlyFields_IntermediateWithOutputs(t *testing.T) {
	fm := &frontmatter.Frontmatter{
		Outputs: []*frontmatter.FrontmatterOutput{
			{ID: "x", Path: "x.go"},
		},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fm, nil, nil),
		testMakeEntry("ROOT/a/b", "ROOT/a/b", testEmptyFM(), nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected leaf_only_fields error for ROOT/a, got: %+v", errs)
	}
}

func TestLeafOnlyFields_IntermediateWithInput(t *testing.T) {
	fm := &frontmatter.Frontmatter{
		Input: "ARTIFACT/c(id)",
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fm, nil, nil),
		testMakeEntry("ROOT/a/b", "ROOT/a/b", testEmptyFM(), nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected leaf_only_fields error for ROOT/a, got: %+v", errs)
	}
}

func TestLeafOnlyFields_IntermediateWithExternal(t *testing.T) {
	fm := &frontmatter.Frontmatter{
		External: []*frontmatter.FrontmatterExternal{
			{Path: "some/file.txt"},
		},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fm, nil, nil),
		testMakeEntry("ROOT/a/b", "ROOT/a/b", testEmptyFM(), nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Errorf("expected leaf_only_fields error for ROOT/a, got: %+v", errs)
	}
}

func TestLeafOnlyFields_IntermediateWithMultipleFields(t *testing.T) {
	fm := &frontmatter.Frontmatter{
		DependsOn: []*string{testStrPtr("ROOT/b")},
		Outputs: []*frontmatter.FrontmatterOutput{
			{ID: "x", Path: "x.go"},
		},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fm, nil, nil),
		testMakeEntry("ROOT/a/b", "ROOT/a/b", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/b", "ROOT/b", testEmptyFM(), nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	count := testCountErrors(errs, "ROOT/a", "leaf_only_fields")
	if count != 2 {
		t.Errorf("expected 2 leaf_only_fields errors for ROOT/a, got %d: %+v", count, errs)
	}
}

// ---------------------------------------------------------------------------
// TC-LOA: Rule leaf_only_agent
// ---------------------------------------------------------------------------

func TestLeafOnlyAgent_IntermediateWithAgent(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", testEmptyFM(), testAgentSection(), nil),
		testMakeEntry("ROOT/a/b", "ROOT/a/b", testEmptyFM(), nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !testHasError(errs, "ROOT/a", "leaf_only_agent") {
		t.Errorf("expected leaf_only_agent error for ROOT/a, got: %+v", errs)
	}
}

func TestLeafOnlyAgent_LeafWithAgent(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", testEmptyFM(), testAgentSection(), nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if testHasError(errs, "ROOT/a", "leaf_only_agent") {
		t.Errorf("unexpected leaf_only_agent error for leaf ROOT/a")
	}
}

// ---------------------------------------------------------------------------
// TC-DT: Rule dependency_targets
// ---------------------------------------------------------------------------

func TestDependencyTargets_NonExistentROOT(t *testing.T) {
	fm := &frontmatter.Frontmatter{
		DependsOn: []*string{testStrPtr("ROOT/missing")},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fm, nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a, got: %+v", errs)
	}
}

func TestDependencyTargets_TargetsAncestor(t *testing.T) {
	fm := &frontmatter.Frontmatter{
		DependsOn: []*string{testStrPtr("ROOT")},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a/b", "ROOT/a/b", fm, nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !testHasError(errs, "ROOT/a/b", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a/b, got: %+v", errs)
	}
}

func TestDependencyTargets_TargetsDescendant(t *testing.T) {
	fm := &frontmatter.Frontmatter{
		DependsOn: []*string{testStrPtr("ROOT/a/b")},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fm, nil, nil),
		testMakeEntry("ROOT/a/b", "ROOT/a/b", testEmptyFM(), nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a, got: %+v", errs)
	}
}

func TestDependencyTargets_TargetsSelf(t *testing.T) {
	fm := &frontmatter.Frontmatter{
		DependsOn: []*string{testStrPtr("ROOT/a")},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fm, nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a, got: %+v", errs)
	}
}

func TestDependencyTargets_ValidROOTWithQualifier(t *testing.T) {
	fm := &frontmatter.Frontmatter{
		DependsOn: []*string{testStrPtr("ROOT/a(interface)")},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/b", "ROOT/b", fm, nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if testHasError(errs, "ROOT/b", "dependency_targets") {
		t.Errorf("unexpected dependency_targets error for ROOT/b")
	}
}

func TestDependencyTargets_ValidARTIFACTReference(t *testing.T) {
	fmA := &frontmatter.Frontmatter{
		Outputs: []*frontmatter.FrontmatterOutput{
			{ID: "lib", Path: "lib.go"},
		},
	}
	fmB := &frontmatter.Frontmatter{
		DependsOn: []*string{testStrPtr("ARTIFACT/a(lib)")},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fmA, nil, nil),
		testMakeEntry("ROOT/b", "ROOT/b", fmB, nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if testHasError(errs, "ROOT/b", "dependency_targets") {
		t.Errorf("unexpected dependency_targets error for ROOT/b")
	}
}

func TestDependencyTargets_NonExistentARTIFACT(t *testing.T) {
	fm := &frontmatter.Frontmatter{
		DependsOn: []*string{testStrPtr("ARTIFACT/missing(id)")},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fm, nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Errorf("expected dependency_targets error for ROOT/a, got: %+v", errs)
	}
}

func TestDependencyTargets_MultipleInvalidEntries(t *testing.T) {
	fm := &frontmatter.Frontmatter{
		DependsOn: []*string{
			testStrPtr("ROOT/missing"),
			testStrPtr("ROOT/also_missing"),
		},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fm, nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	count := testCountErrors(errs, "ROOT/a", "dependency_targets")
	if count != 2 {
		t.Errorf("expected 2 dependency_targets errors for ROOT/a, got %d: %+v", count, errs)
	}
}

// ---------------------------------------------------------------------------
// TC-IT: Rule input_target
// ---------------------------------------------------------------------------

func TestInputTarget_ValidReference(t *testing.T) {
	fmA := &frontmatter.Frontmatter{
		Outputs: []*frontmatter.FrontmatterOutput{
			{ID: "out", Path: "a.go"},
		},
	}
	fmB := &frontmatter.Frontmatter{
		Input: "ARTIFACT/a(out)",
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fmA, nil, nil),
		testMakeEntry("ROOT/b", "ROOT/b", fmB, nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if testHasError(errs, "ROOT/b", "input_target") {
		t.Errorf("unexpected input_target error for ROOT/b")
	}
}

func TestInputTarget_NotStartingWithARTIFACT(t *testing.T) {
	fm := &frontmatter.Frontmatter{
		Input: "ROOT/something",
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fm, nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !testHasError(errs, "ROOT/a", "input_target") {
		t.Errorf("expected input_target error for ROOT/a, got: %+v", errs)
	}
}

func TestInputTarget_NonExistentArtifact(t *testing.T) {
	fm := &frontmatter.Frontmatter{
		Input: "ARTIFACT/missing(id)",
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fm, nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !testHasError(errs, "ROOT/a", "input_target") {
		t.Errorf("expected input_target error for ROOT/a, got: %+v", errs)
	}
}

// ---------------------------------------------------------------------------
// TC-EF: Rule external_files
// ---------------------------------------------------------------------------

func TestExternalFiles_ExistsNoFragments(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("some", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("some/file.txt", []byte("hello\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	fm := &frontmatter.Frontmatter{
		External: []*frontmatter.FrontmatterExternal{
			{Path: "some/file.txt"},
		},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fm, nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("unexpected external_files error for ROOT/a")
	}
}

func TestExternalFiles_FileDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	fm := &frontmatter.Frontmatter{
		External: []*frontmatter.FrontmatterExternal{
			{Path: "nonexistent.txt"},
		},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fm, nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("expected external_files error for ROOT/a, got: %+v", errs)
	}
}

func TestExternalFiles_FragmentValidHash(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	fileContent := "line one\nline two\nline three\nline four\nline five\n"
	if err := os.WriteFile("f.txt", []byte(fileContent), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Compute expected hash for lines 1-3: "line one\nline two\nline three"
	correctHash := testComputeHash([]string{"line one", "line two", "line three"})

	fm := &frontmatter.Frontmatter{
		External: []*frontmatter.FrontmatterExternal{
			{
				Path: "f.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "1-3", Hash: correctHash},
				},
			},
		},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fm, nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("unexpected external_files error for ROOT/a (hash=%s)", correctHash)
	}
}

func TestExternalFiles_FragmentInvalidHash(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	fileContent := "line one\nline two\nline three\n"
	if err := os.WriteFile("f.txt", []byte(fileContent), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	fm := &frontmatter.Frontmatter{
		External: []*frontmatter.FrontmatterExternal{
			{
				Path: "f.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "1-3", Hash: "wrong"},
				},
			},
		},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fm, nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("expected external_files error for ROOT/a, got: %+v", errs)
	}
}

func TestExternalFiles_FragmentInvalidRangeFormat(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("f.txt", []byte("hello\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	fm := &frontmatter.Frontmatter{
		External: []*frontmatter.FrontmatterExternal{
			{
				Path: "f.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "abc", Hash: "x"},
				},
			},
		},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fm, nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("expected external_files error for ROOT/a, got: %+v", errs)
	}
}

func TestExternalFiles_FragmentStartGreaterThanEnd(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("f.txt", []byte("hello\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	fm := &frontmatter.Frontmatter{
		External: []*frontmatter.FrontmatterExternal{
			{
				Path: "f.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "5-3", Hash: "x"},
				},
			},
		},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fm, nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("expected external_files error for ROOT/a, got: %+v", errs)
	}
}

func TestExternalFiles_FragmentStartLessThanOne(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("f.txt", []byte("hello\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	fm := &frontmatter.Frontmatter{
		External: []*frontmatter.FrontmatterExternal{
			{
				Path: "f.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "0-3", Hash: "x"},
				},
			},
		},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fm, nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("expected external_files error for ROOT/a, got: %+v", errs)
	}
}

func TestExternalFiles_FragmentOutOfRange(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// File with exactly 5 lines.
	fileContent := "line one\nline two\nline three\nline four\nline five\n"
	if err := os.WriteFile("f.txt", []byte(fileContent), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	fm := &frontmatter.Frontmatter{
		External: []*frontmatter.FrontmatterExternal{
			{
				Path: "f.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "1-100", Hash: "x"},
				},
			},
		},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fm, nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("expected external_files error for ROOT/a, got: %+v", errs)
	}
	// Verify detail mentions out of range.
	found := false
	for _, fe := range errs {
		if fe.Node == "ROOT/a" && fe.Rule == "external_files" {
			if fe.Detail != "" {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("expected external_files error with non-empty detail for ROOT/a, got: %+v", errs)
	}
}

// ---------------------------------------------------------------------------
// TC-OP: Rule output_paths
// ---------------------------------------------------------------------------

func TestOutputPaths_ValidPath(t *testing.T) {
	fm := &frontmatter.Frontmatter{
		Outputs: []*frontmatter.FrontmatterOutput{
			{ID: "x", Path: "internal/x.go"},
		},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fm, nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if testHasError(errs, "ROOT/a", "output_paths") {
		t.Errorf("unexpected output_paths error for ROOT/a")
	}
}

func TestOutputPaths_PathWithTraversal(t *testing.T) {
	fm := &frontmatter.Frontmatter{
		Outputs: []*frontmatter.FrontmatterOutput{
			{ID: "x", Path: "../../etc/passwd"},
		},
	}
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fm, nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !testHasError(errs, "ROOT/a", "output_paths") {
		t.Errorf("expected output_paths error for ROOT/a, got: %+v", errs)
	}
}

func TestOutputPaths_PathWithBackslash(t *testing.T) {
	fm := &frontmatter.Frontmatter{
		Outputs: []*frontmatter.FrontmatterOutput{
			{ID: "x", Path: filepath.FromSlash("internal\\x.go")},
		},
	}
	// Use the literal backslash string as provided in the spec.
	fm.Outputs[0].Path = `internal\x.go`
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", fm, nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !testHasError(errs, "ROOT/a", "output_paths") {
		t.Errorf("expected output_paths error for ROOT/a, got: %+v", errs)
	}
}

// ---------------------------------------------------------------------------
// TC-DS: Rule duplicate_subsections
// ---------------------------------------------------------------------------

func TestDuplicateSubsections_UniqueHeadings(t *testing.T) {
	public := testMakePublicWithSubs(
		testMakeSubsection("interface"),
		testMakeSubsection("context"),
	)
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", testEmptyFM(), nil, public),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if testHasError(errs, "ROOT/a", "duplicate_subsections") {
		t.Errorf("unexpected duplicate_subsections error for ROOT/a")
	}
}

func TestDuplicateSubsections_TwoDuplicates(t *testing.T) {
	public := testMakePublicWithSubs(
		testMakeSubsection("interface"),
		testMakeSubsection("interface"),
	)
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", testEmptyFM(), nil, public),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	count := testCountErrors(errs, "ROOT/a", "duplicate_subsections")
	if count != 1 {
		t.Errorf("expected 1 duplicate_subsections error for ROOT/a, got %d: %+v", count, errs)
	}
}

func TestDuplicateSubsections_ThreeDuplicates(t *testing.T) {
	public := testMakePublicWithSubs(
		testMakeSubsection("interface"),
		testMakeSubsection("interface"),
		testMakeSubsection("interface"),
	)
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", testEmptyFM(), nil, public),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	count := testCountErrors(errs, "ROOT/a", "duplicate_subsections")
	if count != 2 {
		t.Errorf("expected 2 duplicate_subsections errors for ROOT/a, got %d: %+v", count, errs)
	}
}

func TestDuplicateSubsections_NoPublicSection(t *testing.T) {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		testMakeEntry("ROOT/a", "ROOT/a", testEmptyFM(), nil, nil),
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if testHasError(errs, "ROOT/a", "duplicate_subsections") {
		t.Errorf("unexpected duplicate_subsections error for ROOT/a with no public section")
	}
}

// ---------------------------------------------------------------------------
// TC-CC: Cross-Cutting
// ---------------------------------------------------------------------------

func TestCrossCutting_MultipleRulesViolated(t *testing.T) {
	fm := &frontmatter.Frontmatter{
		DependsOn: []*string{testStrPtr("ROOT/missing")},
	}
	public := testMakePublicWithSubs(
		testMakeSubsection("interface"),
		testMakeSubsection("interface"),
	)
	entries := []*spectreevalidate.SpecTreeValidateInput{
		testMakeEntry("ROOT", "ROOT", testEmptyFM(), nil, nil),
		{
			LogicalName: "ROOT/a",
			Frontmatter: fm,
			Node: &parsenode.Node{
				NameSection: &parsenode.NodeSection{Heading: "ROOT/wrong"},
				Public:      public,
			},
		},
	}
	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) < 3 {
		t.Errorf("expected at least 3 errors, got %d: %+v", len(errs), errs)
	}
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
	errs, err := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty input, got %d: %+v", len(errs), errs)
	}
}
