// code-from-spec: ROOT/golang/tests/spec_tree/validate@_hFe_ELYpH6FGfpkivuFODZbiyY
package spectreevalidate_test

import (
	"crypto/sha1"
	"encoding/base64"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectreevalidate"
)

// testChdir changes the working directory to dir for the duration of the test.
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

// testMakeRootEntry constructs a minimal valid ROOT node entry.
func testMakeRootEntry() *spectreevalidate.SpecTreeValidateInput {
	return &spectreevalidate.SpecTreeValidateInput{
		LogicalName: "ROOT",
		Frontmatter: &frontmatter.Frontmatter{},
		Node: &parsenode.Node{
			NameSection: &parsenode.NodeSection{
				Heading:    "root",
				RawHeading: "# ROOT",
				Content:    []string{},
			},
			Public: &parsenode.NodeSection{
				Heading:    "public",
				RawHeading: "# Public",
				Content:    []string{},
			},
		},
	}
}

// testMakeEntry constructs a minimal valid leaf entry with the given logical name.
func testMakeEntry(logicalName string) *spectreevalidate.SpecTreeValidateInput {
	return &spectreevalidate.SpecTreeValidateInput{
		LogicalName: logicalName,
		Frontmatter: &frontmatter.Frontmatter{},
		Node: &parsenode.Node{
			NameSection: &parsenode.NodeSection{
				Heading:    logicalName,
				RawHeading: "# " + logicalName,
				Content:    []string{},
			},
		},
	}
}

// testCountErrors counts FormatErrors matching node and rule.
func testCountErrors(errs []*spectreevalidate.FormatError, node, rule string) int {
	count := 0
	for _, e := range errs {
		if e.Node == node && e.Rule == rule {
			count++
		}
	}
	return count
}

// testHasError returns true if any error matches node and rule.
func testHasError(errs []*spectreevalidate.FormatError, node, rule string) bool {
	return testCountErrors(errs, node, rule) > 0
}

// testFragmentHash computes the SHA-1 base64url (no padding) hash of lines joined with \n.
func testFragmentHash(lines ...string) string {
	h := sha1.New()
	for _, line := range lines {
		h.Write([]byte(line + "\n"))
	}
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// ---- Happy path ----

func TestValidLeafNodePassesAllChecks(t *testing.T) {
	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	entry.Frontmatter.DependsOn = []string{"ROOT"}
	entry.Frontmatter.Outputs = []*frontmatter.FrontmatterOutput{
		{ID: "out", Path: "internal/out.go"},
	}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestValidIntermediateNodePassesAllChecks(t *testing.T) {
	root := testMakeRootEntry()

	intermediate := testMakeEntry("ROOT/a")
	intermediate.Node.Public = &parsenode.NodeSection{
		Heading:    "public",
		RawHeading: "# Public",
		Content:    []string{},
	}

	leaf := testMakeEntry("ROOT/a/b")

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, intermediate, leaf})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestLeafWithNoFrontmatterFields(t *testing.T) {
	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

// ---- Rule: name_heading ----

func TestNameHeadingMatches(t *testing.T) {
	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	if testHasError(errs, "ROOT/a", "name_heading") {
		t.Error("expected no name_heading error, but found one")
	}
}

func TestNameHeadingDoesNotMatch(t *testing.T) {
	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	entry.Node.NameSection.Heading = "ROOT/wrong"

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	if !testHasError(errs, "ROOT/a", "name_heading") {
		t.Error("expected name_heading error for ROOT/a, got none")
	}
}

// ---- Rule: leaf_only_fields ----

func TestIntermediateNodeWithDependsOn(t *testing.T) {
	root := testMakeRootEntry()
	intermediate := testMakeEntry("ROOT/a")
	intermediate.Frontmatter.DependsOn = []string{"ROOT/b"}
	leaf := testMakeEntry("ROOT/a/b")

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, intermediate, leaf})
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Error("expected leaf_only_fields error for ROOT/a, got none")
	}
}

func TestIntermediateNodeWithOutputs(t *testing.T) {
	root := testMakeRootEntry()
	intermediate := testMakeEntry("ROOT/a")
	intermediate.Frontmatter.Outputs = []*frontmatter.FrontmatterOutput{
		{ID: "x", Path: "x.go"},
	}
	leaf := testMakeEntry("ROOT/a/b")

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, intermediate, leaf})
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Error("expected leaf_only_fields error for ROOT/a, got none")
	}
}

func TestIntermediateNodeWithInput(t *testing.T) {
	root := testMakeRootEntry()
	intermediate := testMakeEntry("ROOT/a")
	intermediate.Frontmatter.Input = "ARTIFACT/c(id)"
	leaf := testMakeEntry("ROOT/a/b")

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, intermediate, leaf})
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Error("expected leaf_only_fields error for ROOT/a, got none")
	}
}

func TestIntermediateNodeWithExternal(t *testing.T) {
	root := testMakeRootEntry()
	intermediate := testMakeEntry("ROOT/a")
	intermediate.Frontmatter.External = []*frontmatter.FrontmatterExternal{
		{Path: "some/file.txt"},
	}
	leaf := testMakeEntry("ROOT/a/b")

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, intermediate, leaf})
	if !testHasError(errs, "ROOT/a", "leaf_only_fields") {
		t.Error("expected leaf_only_fields error for ROOT/a, got none")
	}
}

func TestIntermediateNodeWithMultipleRestrictedFields(t *testing.T) {
	root := testMakeRootEntry()
	intermediate := testMakeEntry("ROOT/a")
	intermediate.Frontmatter.DependsOn = []string{"ROOT/b"}
	intermediate.Frontmatter.Outputs = []*frontmatter.FrontmatterOutput{
		{ID: "x", Path: "x.go"},
	}
	leaf := testMakeEntry("ROOT/a/b")

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, intermediate, leaf})
	count := testCountErrors(errs, "ROOT/a", "leaf_only_fields")
	if count != 2 {
		t.Errorf("expected exactly 2 leaf_only_fields errors for ROOT/a, got %d", count)
	}
}

// ---- Rule: leaf_only_agent ----

func TestIntermediateNodeWithAgentSection(t *testing.T) {
	root := testMakeRootEntry()
	intermediate := testMakeEntry("ROOT/a")
	intermediate.Node.Agent = &parsenode.NodeSection{
		Heading:    "agent",
		RawHeading: "# Agent",
		Content:    []string{"some content"},
	}
	leaf := testMakeEntry("ROOT/a/b")

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, intermediate, leaf})
	if !testHasError(errs, "ROOT/a", "leaf_only_agent") {
		t.Error("expected leaf_only_agent error for ROOT/a, got none")
	}
}

func TestLeafNodeWithAgentSectionNoError(t *testing.T) {
	root := testMakeRootEntry()
	leaf := testMakeEntry("ROOT/a")
	leaf.Node.Agent = &parsenode.NodeSection{
		Heading:    "agent",
		RawHeading: "# Agent",
		Content:    []string{"agent guidance"},
	}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, leaf})
	if testHasError(errs, "ROOT/a", "leaf_only_agent") {
		t.Error("expected no leaf_only_agent error for leaf ROOT/a, but found one")
	}
}

// ---- Rule: dependency_targets ----

func TestDependsOnTargetsNonExistentRootNode(t *testing.T) {
	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	entry.Frontmatter.DependsOn = []string{"ROOT/missing"}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Error("expected dependency_targets error for ROOT/a, got none")
	}
}

func TestDependsOnTargetsAncestor(t *testing.T) {
	root := testMakeRootEntry()
	parent := testMakeEntry("ROOT/a")
	child := testMakeEntry("ROOT/a/b")
	child.Frontmatter.DependsOn = []string{"ROOT"}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, parent, child})
	if !testHasError(errs, "ROOT/a/b", "dependency_targets") {
		t.Error("expected dependency_targets error for ROOT/a/b (ancestor), got none")
	}
}

func TestDependsOnTargetsDescendant(t *testing.T) {
	root := testMakeRootEntry()
	parent := testMakeEntry("ROOT/a")
	parent.Frontmatter.DependsOn = []string{"ROOT/a/b"}
	child := testMakeEntry("ROOT/a/b")

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, parent, child})
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Error("expected dependency_targets error for ROOT/a (descendant), got none")
	}
}

func TestDependsOnTargetsSelf(t *testing.T) {
	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	entry.Frontmatter.DependsOn = []string{"ROOT/a"}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Error("expected dependency_targets error for ROOT/a (self), got none")
	}
}

func TestDependsOnWithValidRootQualifier(t *testing.T) {
	root := testMakeRootEntry()
	entryA := testMakeEntry("ROOT/a")
	entryB := testMakeEntry("ROOT/b")
	entryB.Frontmatter.DependsOn = []string{"ROOT/a(interface)"}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entryA, entryB})
	if testHasError(errs, "ROOT/b", "dependency_targets") {
		t.Error("expected no dependency_targets error for ROOT/b with valid qualifier, but found one")
	}
}

func TestDependsOnWithValidArtifactReference(t *testing.T) {
	root := testMakeRootEntry()
	entryA := testMakeEntry("ROOT/a")
	entryA.Frontmatter.Outputs = []*frontmatter.FrontmatterOutput{
		{ID: "lib", Path: "lib.go"},
	}
	entryB := testMakeEntry("ROOT/b")
	entryB.Frontmatter.DependsOn = []string{"ARTIFACT/a(lib)"}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entryA, entryB})
	if testHasError(errs, "ROOT/b", "dependency_targets") {
		t.Error("expected no dependency_targets error for ROOT/b with valid ARTIFACT reference, but found one")
	}
}

func TestDependsOnWithNonExistentArtifactReference(t *testing.T) {
	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	entry.Frontmatter.DependsOn = []string{"ARTIFACT/missing(id)"}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Error("expected dependency_targets error for ROOT/a with non-existent ARTIFACT, got none")
	}
}

func TestMultipleInvalidDependsOnOneErrorPerEntry(t *testing.T) {
	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	entry.Frontmatter.DependsOn = []string{"ROOT/missing", "ROOT/also_missing"}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	count := testCountErrors(errs, "ROOT/a", "dependency_targets")
	if count != 2 {
		t.Errorf("expected exactly 2 dependency_targets errors for ROOT/a, got %d", count)
	}
}

// ---- Rule: input_target ----

func TestValidInputReference(t *testing.T) {
	root := testMakeRootEntry()
	entryA := testMakeEntry("ROOT/a")
	entryA.Frontmatter.Outputs = []*frontmatter.FrontmatterOutput{
		{ID: "out", Path: "a.go"},
	}
	entryB := testMakeEntry("ROOT/b")
	entryB.Frontmatter.Input = "ARTIFACT/a(out)"

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entryA, entryB})
	if testHasError(errs, "ROOT/b", "input_target") {
		t.Error("expected no input_target error for ROOT/b with valid input, but found one")
	}
}

func TestInputNotStartingWithArtifact(t *testing.T) {
	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	entry.Frontmatter.Input = "ROOT/something"

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	if !testHasError(errs, "ROOT/a", "input_target") {
		t.Error("expected input_target error for ROOT/a (non-ARTIFACT input), got none")
	}
}

func TestInputReferencesNonExistentArtifact(t *testing.T) {
	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	entry.Frontmatter.Input = "ARTIFACT/missing(id)"

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	if !testHasError(errs, "ROOT/a", "input_target") {
		t.Error("expected input_target error for ROOT/a with non-existent artifact, got none")
	}
}

// ---- Rule: external_files ----

func TestExternalFileExistsNoFragments(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("some", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile("some/file.txt", []byte("hello\nworld\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	entry.Frontmatter.External = []*frontmatter.FrontmatterExternal{
		{Path: "some/file.txt"},
	}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	if testHasError(errs, "ROOT/a", "external_files") {
		t.Error("expected no external_files error, but found one")
	}
}

func TestExternalFileDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	entry.Frontmatter.External = []*frontmatter.FrontmatterExternal{
		{Path: "nonexistent.txt"},
	}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Error("expected external_files error for missing file, got none")
	}
}

func TestFragmentWithValidHash(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "alpha\nbeta\ngamma\ndelta\nepsilon\n"
	if err := os.WriteFile("f.txt", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	correctHash := testFragmentHash("alpha", "beta", "gamma")

	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	entry.Frontmatter.External = []*frontmatter.FrontmatterExternal{
		{
			Path: "f.txt",
			Fragments: []*frontmatter.FrontmatterExternalFragment{
				{Lines: "1-3", Hash: correctHash},
			},
		},
	}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	if testHasError(errs, "ROOT/a", "external_files") {
		t.Errorf("expected no external_files error with valid hash %q, but found one", correctHash)
	}
}

func TestFragmentWithInvalidHash(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "alpha\nbeta\ngamma\ndelta\nepsilon\n"
	if err := os.WriteFile("f.txt", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	entry.Frontmatter.External = []*frontmatter.FrontmatterExternal{
		{
			Path: "f.txt",
			Fragments: []*frontmatter.FrontmatterExternalFragment{
				{Lines: "1-3", Hash: "wrong"},
			},
		},
	}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Error("expected external_files error for invalid hash, got none")
	}
}

func TestFragmentWithInvalidRangeFormat(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("f.txt", []byte("line1\nline2\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	entry.Frontmatter.External = []*frontmatter.FrontmatterExternal{
		{
			Path: "f.txt",
			Fragments: []*frontmatter.FrontmatterExternalFragment{
				{Lines: "abc", Hash: "x"},
			},
		},
	}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Error("expected external_files error for invalid range format, got none")
	}
}

func TestFragmentWithStartGreaterThanEnd(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("f.txt", []byte("line1\nline2\nline3\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	entry.Frontmatter.External = []*frontmatter.FrontmatterExternal{
		{
			Path: "f.txt",
			Fragments: []*frontmatter.FrontmatterExternalFragment{
				{Lines: "5-3", Hash: "x"},
			},
		},
	}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Error("expected external_files error for start > end, got none")
	}
}

func TestFragmentWithStartLessThanOne(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("f.txt", []byte("line1\nline2\nline3\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	entry.Frontmatter.External = []*frontmatter.FrontmatterExternal{
		{
			Path: "f.txt",
			Fragments: []*frontmatter.FrontmatterExternalFragment{
				{Lines: "0-3", Hash: "x"},
			},
		},
	}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Error("expected external_files error for start < 1, got none")
	}
}

func TestFragmentOutOfRange(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "line1\nline2\nline3\nline4\nline5\n"
	if err := os.WriteFile("f.txt", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	entry.Frontmatter.External = []*frontmatter.FrontmatterExternal{
		{
			Path: "f.txt",
			Fragments: []*frontmatter.FrontmatterExternalFragment{
				{Lines: "1-100", Hash: "x"},
			},
		},
	}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	if !testHasError(errs, "ROOT/a", "external_files") {
		t.Error("expected external_files error for out-of-range fragment, got none")
	}
	// Verify detail indicates out of range
	found := false
	for _, e := range errs {
		if e.Node == "ROOT/a" && e.Rule == "external_files" {
			found = true
			if e.Detail == "" {
				t.Error("expected non-empty detail for out-of-range fragment error")
			}
			break
		}
	}
	if !found {
		t.Error("no external_files error found for ROOT/a")
	}
}

// ---- Rule: output_paths ----

func TestValidOutputPath(t *testing.T) {
	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	entry.Frontmatter.Outputs = []*frontmatter.FrontmatterOutput{
		{ID: "x", Path: "internal/x.go"},
	}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	if testHasError(errs, "ROOT/a", "output_paths") {
		t.Error("expected no output_paths error for valid path, but found one")
	}
}

func TestOutputPathWithTraversal(t *testing.T) {
	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	entry.Frontmatter.Outputs = []*frontmatter.FrontmatterOutput{
		{ID: "x", Path: "../../etc/passwd"},
	}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	if !testHasError(errs, "ROOT/a", "output_paths") {
		t.Error("expected output_paths error for traversal path, got none")
	}
}

func TestOutputPathWithBackslash(t *testing.T) {
	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	entry.Frontmatter.Outputs = []*frontmatter.FrontmatterOutput{
		{ID: "x", Path: `internal\x.go`},
	}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	if !testHasError(errs, "ROOT/a", "output_paths") {
		t.Error("expected output_paths error for path with backslash, got none")
	}
}

// ---- Rule: duplicate_subsections ----

func TestUniqueSubsectionHeadingsNoError(t *testing.T) {
	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	entry.Node.Public = &parsenode.NodeSection{
		Heading:    "public",
		RawHeading: "# Public",
		Content:    []string{},
		Subsections: []*parsenode.NodeSubsection{
			{Heading: "interface", RawHeading: "## Interface", Content: []string{}},
			{Heading: "context", RawHeading: "## Context", Content: []string{}},
		},
	}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	if testHasError(errs, "ROOT/a", "duplicate_subsections") {
		t.Error("expected no duplicate_subsections error for unique headings, but found one")
	}
}

func TestDuplicateSubsectionHeadings(t *testing.T) {
	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	entry.Node.Public = &parsenode.NodeSection{
		Heading:    "public",
		RawHeading: "# Public",
		Content:    []string{},
		Subsections: []*parsenode.NodeSubsection{
			{Heading: "interface", RawHeading: "## Interface", Content: []string{}},
			{Heading: "interface", RawHeading: "## Interface", Content: []string{}},
		},
	}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	count := testCountErrors(errs, "ROOT/a", "duplicate_subsections")
	if count != 1 {
		t.Errorf("expected exactly 1 duplicate_subsections error for ROOT/a, got %d", count)
	}
}

func TestThreeIdenticalSubsectionHeadings(t *testing.T) {
	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	entry.Node.Public = &parsenode.NodeSection{
		Heading:    "public",
		RawHeading: "# Public",
		Content:    []string{},
		Subsections: []*parsenode.NodeSubsection{
			{Heading: "interface", RawHeading: "## Interface", Content: []string{}},
			{Heading: "interface", RawHeading: "## Interface", Content: []string{}},
			{Heading: "interface", RawHeading: "## Interface", Content: []string{}},
		},
	}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	count := testCountErrors(errs, "ROOT/a", "duplicate_subsections")
	if count != 2 {
		t.Errorf("expected exactly 2 duplicate_subsections errors for ROOT/a, got %d", count)
	}
}

func TestNoPublicSectionSkipDuplicateCheck(t *testing.T) {
	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	// entry.Node.Public is nil

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})
	if testHasError(errs, "ROOT/a", "duplicate_subsections") {
		t.Error("expected no duplicate_subsections error when public section absent, but found one")
	}
}

// ---- Cross-cutting ----

func TestCollectsMultipleErrorsFromDifferentRules(t *testing.T) {
	root := testMakeRootEntry()
	entry := testMakeEntry("ROOT/a")
	// Triggers name_heading
	entry.Node.NameSection.Heading = "ROOT/wrong"
	// Triggers dependency_targets
	entry.Frontmatter.DependsOn = []string{"ROOT/nonexistent"}
	// Triggers duplicate_subsections
	entry.Node.Public = &parsenode.NodeSection{
		Heading:    "public",
		RawHeading: "# Public",
		Content:    []string{},
		Subsections: []*parsenode.NodeSubsection{
			{Heading: "interface", RawHeading: "## Interface", Content: []string{}},
			{Heading: "interface", RawHeading: "## Interface", Content: []string{}},
		},
	}

	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{root, entry})

	if !testHasError(errs, "ROOT/a", "name_heading") {
		t.Error("expected name_heading error, got none")
	}
	if !testHasError(errs, "ROOT/a", "dependency_targets") {
		t.Error("expected dependency_targets error, got none")
	}
	if !testHasError(errs, "ROOT/a", "duplicate_subsections") {
		t.Error("expected duplicate_subsections error, got none")
	}
}

func TestEmptyInputList(t *testing.T) {
	errs := spectreevalidate.SpecTreeValidate([]*spectreevalidate.SpecTreeValidateInput{})
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty input, got %d", len(errs))
	}
}
