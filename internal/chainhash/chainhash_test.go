// code-from-spec: ROOT/golang/tests/chain/hash@sWctBJKHiUtvxEn4pwIRWt_2Dy8
package chainhash_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// testChdir changes the working directory for the duration of the test.
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

// testMkdirAll creates the parent directory for the given path.
func testMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("testMkdirAll: %v", err)
	}
}

// testWriteFile writes content to a relative path, creating directories as needed.
func testWriteFile(t *testing.T, path, content string) {
	t.Helper()
	testMkdirAll(t, path)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile %s: %v", path, err)
	}
}

// testChainItem builds a ChainItem from a CFS path string, with no qualifier.
func testChainItem(path string) *chainresolver.ChainItem {
	return &chainresolver.ChainItem{
		FilePath: &pathutils.PathCfs{Value: path},
	}
}

// testChainItemWithQualifier builds a ChainItem from a CFS path string, with a qualifier.
func testChainItemWithQualifier(path, qualifier string) *chainresolver.ChainItem {
	q := qualifier
	return &chainresolver.ChainItem{
		FilePath:  &pathutils.PathCfs{Value: path},
		Qualifier: &q,
	}
}

// testChainItemArtifact builds a ChainItem that represents an ARTIFACT reference.
// The LogicalName starts with "ARTIFACT/" so ChainHashCompute can distinguish it.
func testChainItemArtifact(logicalName, path string) *chainresolver.ChainItem {
	return &chainresolver.ChainItem{
		LogicalName: logicalName,
		FilePath:    &pathutils.PathCfs{Value: path},
	}
}

// testRootNodeContent returns a minimal ROOT node file content.
func testRootNodeContent() string {
	return "# ROOT\n"
}

// testNodeContent returns a minimal node file content for the given logical name.
func testNodeContent(logicalName string) string {
	return fmt.Sprintf("# %s\n", logicalName)
}

// testNodeContentWithPublic returns a node file with a # Public section.
func testNodeContentWithPublic(logicalName, publicContent string) string {
	return fmt.Sprintf("# %s\n\n# Public\n\n%s\n", logicalName, publicContent)
}

// testNodeContentWithAgent returns a node file with a # Agent section.
func testNodeContentWithAgent(logicalName, agentContent string) string {
	return fmt.Sprintf("# %s\n\n# Agent\n\n%s\n", logicalName, agentContent)
}

// testNodeContentWithPublicAndAgent returns a node file with both # Public and # Agent sections.
func testNodeContentWithPublicAndAgent(logicalName, publicContent, agentContent string) string {
	return fmt.Sprintf("# %s\n\n# Public\n\n%s\n\n# Agent\n\n%s\n", logicalName, publicContent, agentContent)
}

// testNodeContentWithPublicSubsection returns a node file with # Public containing a ## subsection.
func testNodeContentWithPublicSubsection(logicalName, subsectionName, subsectionContent string) string {
	return fmt.Sprintf("# %s\n\n# Public\n\n## %s\n\n%s\n", logicalName, subsectionName, subsectionContent)
}

// testArtifactContent returns a simple artifact file with frontmatter and body.
func testArtifactContent(frontmatterYAML, body string) string {
	if frontmatterYAML == "" {
		return fmt.Sprintf("---\n---\n\n%s\n", body)
	}
	return fmt.Sprintf("---\n%s\n---\n\n%s\n", frontmatterYAML, body)
}

// testExternalEntry builds a FrontmatterExternal with no fragments.
func testExternalEntry(path string) *frontmatter.FrontmatterExternal {
	return &frontmatter.FrontmatterExternal{
		Path: path,
	}
}

// testExternalEntryWithFragments builds a FrontmatterExternal with the given line ranges.
// Each range is a string like "3-5".
func testExternalEntryWithFragments(path string, lineRanges []string) *frontmatter.FrontmatterExternal {
	frags := make([]*frontmatter.FrontmatterExternalFragment, len(lineRanges))
	for i, r := range lineRanges {
		frags[i] = &frontmatter.FrontmatterExternalFragment{
			Lines: r,
		}
	}
	return &frontmatter.FrontmatterExternal{
		Path:      path,
		Fragments: frags,
	}
}

// testMinimalChain returns a Chain with only the target set.
func testMinimalChain(target *chainresolver.ChainItem) *chainresolver.Chain {
	return &chainresolver.Chain{
		Target: target,
	}
}

// --------------------------------------------------------------------------
// Properties
// --------------------------------------------------------------------------

func TestHashIsDeterministic(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", testNodeContentWithPublic("ROOT", "some content"))

	chain := testMinimalChain(testChainItem("code-from-spec/_node.md"))

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("hashes differ: %q != %q", hash1, hash2)
	}
}

func TestHashIs27Characters(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", testNodeContent("ROOT"))

	chain := testMinimalChain(testChainItem("code-from-spec/_node.md"))

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hash) != 27 {
		t.Errorf("expected 27 chars, got %d: %q", len(hash), hash)
	}
}

func TestHashChangesWhenAncestorContentChanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", testNodeContentWithPublic("ROOT", "original ancestor"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a"))

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("code-from-spec/_node.md"),
		},
		Target: testChainItem("code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "code-from-spec/_node.md", testNodeContentWithPublic("ROOT", "modified ancestor"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hashes to differ after ancestor change, but both are %q", hash1)
	}
}

func TestHashChangesWhenDependencyContentChanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", testNodeContent("ROOT"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a"))
	testWriteFile(t, "code-from-spec/b/_node.md", testNodeContentWithPublic("ROOT/b", "original dependency"))

	chain := &chainresolver.Chain{
		Target: testChainItem("code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("code-from-spec/b/_node.md"),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "code-from-spec/b/_node.md", testNodeContentWithPublic("ROOT/b", "modified dependency"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hashes to differ after dependency change, but both are %q", hash1)
	}
}

func TestHashChangesWhenTargetPublicChanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", testNodeContent("ROOT"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContentWithPublic("ROOT/a", "original public"))

	chain := testMinimalChain(testChainItem("code-from-spec/a/_node.md"))

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContentWithPublic("ROOT/a", "modified public"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hashes to differ after target Public change, but both are %q", hash1)
	}
}

func TestHashChangesWhenTargetAgentChanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", testNodeContent("ROOT"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContentWithAgent("ROOT/a", "original agent"))

	chain := testMinimalChain(testChainItem("code-from-spec/a/_node.md"))

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContentWithAgent("ROOT/a", "modified agent"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hashes to differ after target Agent change, but both are %q", hash1)
	}
}

// --------------------------------------------------------------------------
// Ancestors
// --------------------------------------------------------------------------

func TestAncestorWithPublicSectionContributesHash(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", testNodeContentWithPublic("ROOT", "ancestor public content"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a"))

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("code-from-spec/_node.md"),
		},
		Target: testChainItem("code-from-spec/a/_node.md"),
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hash) != 27 || hash == "" {
		t.Errorf("expected 27-char non-empty hash, got %q", hash)
	}
}

func TestAncestorWithoutPublicSectionSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	// Chain A: ancestor ROOT has no Public section
	testWriteFile(t, "code-from-spec/_node.md", testNodeContent("ROOT"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a"))

	chainA := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("code-from-spec/_node.md"),
		},
		Target: testChainItem("code-from-spec/a/_node.md"),
	}

	hashA, err := chainhash.ChainHashCompute(chainA)
	if err != nil {
		t.Fatalf("chain A: %v", err)
	}

	// Chain B: ancestor ROOT has a Public section
	testWriteFile(t, "code-from-spec/_node.md", testNodeContentWithPublic("ROOT", "some public content"))

	chainB := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("code-from-spec/_node.md"),
		},
		Target: testChainItem("code-from-spec/a/_node.md"),
	}

	hashB, err := chainhash.ChainHashCompute(chainB)
	if err != nil {
		t.Fatalf("chain B: %v", err)
	}

	if hashA == hashB {
		t.Errorf("expected hashes to differ when ancestor has/lacks Public section, but both are %q", hashA)
	}
}

func TestMultipleAncestorsOrderMatters(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", testNodeContentWithPublic("ROOT", "root public"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContentWithPublic("ROOT/a", "a public"))
	testWriteFile(t, "code-from-spec/a/b/_node.md", testNodeContent("ROOT/a/b"))

	// Chain A: root-first order
	chainA := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("code-from-spec/_node.md"),
			testChainItem("code-from-spec/a/_node.md"),
		},
		Target: testChainItem("code-from-spec/a/b/_node.md"),
	}

	hashA, err := chainhash.ChainHashCompute(chainA)
	if err != nil {
		t.Fatalf("chain A: %v", err)
	}

	// Chain B: reversed order
	chainB := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("code-from-spec/a/_node.md"),
			testChainItem("code-from-spec/_node.md"),
		},
		Target: testChainItem("code-from-spec/a/b/_node.md"),
	}

	hashB, err := chainhash.ChainHashCompute(chainB)
	if err != nil {
		t.Fatalf("chain B: %v", err)
	}

	if hashA == hashB {
		t.Errorf("expected hashes to differ based on ancestor order, but both are %q", hashA)
	}
}

// --------------------------------------------------------------------------
// Dependencies
// --------------------------------------------------------------------------

func TestRootDependencyWithoutQualifierHashesPublic(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/b/_node.md", testNodeContentWithPublic("ROOT/b", "b public original"))

	chain := &chainresolver.Chain{
		Target: testChainItem("code-from-spec/b/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("code-from-spec/b/_node.md"),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "code-from-spec/b/_node.md", testNodeContentWithPublic("ROOT/b", "b public modified"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hashes to differ after dependency Public change, but both are %q", hash1)
	}
}

func TestRootDependencyWithQualifierHashesSubsection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/b/_node.md", testNodeContentWithPublicSubsection("ROOT/b", "Interface", "interface original"))

	chain := &chainresolver.Chain{
		Target: testChainItem("code-from-spec/b/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItemWithQualifier("code-from-spec/b/_node.md", "interface"),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "code-from-spec/b/_node.md", testNodeContentWithPublicSubsection("ROOT/b", "Interface", "interface modified"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hashes to differ after subsection change, but both are %q", hash1)
	}
}

func TestQualifierCaseNormalization(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/b/_node.md", testNodeContentWithPublicSubsection("ROOT/b", "Interface", "some content"))

	chain := &chainresolver.Chain{
		Target: testChainItem("code-from-spec/b/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItemWithQualifier("code-from-spec/b/_node.md", "INTERFACE"),
		},
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hash) != 27 {
		t.Errorf("expected 27-char hash, got %d: %q", len(hash), hash)
	}
}

func TestArtifactDependencyHashesFileMinusFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	artifactPath := "artifacts/myartifact.md"
	testWriteFile(t, artifactPath, testArtifactContent("", "original body"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a"))

	chain := &chainresolver.Chain{
		Target: testChainItem("code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItemArtifact("ARTIFACT/a", artifactPath),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, artifactPath, testArtifactContent("", "modified body"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hashes to differ after artifact body change, but both are %q", hash1)
	}
}

func TestArtifactDependencyFrontmatterChangeIgnored(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	artifactPath := "artifacts/myartifact.md"
	testWriteFile(t, artifactPath, testArtifactContent("key: value1", "stable body"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a"))

	chain := &chainresolver.Chain{
		Target: testChainItem("code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItemArtifact("ARTIFACT/a", artifactPath),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Modify only the frontmatter, body stays "stable body"
	testWriteFile(t, artifactPath, testArtifactContent("key: value2", "stable body"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("expected hashes to be equal after frontmatter-only change, got %q and %q", hash1, hash2)
	}
}

// --------------------------------------------------------------------------
// External Files
// --------------------------------------------------------------------------

func TestExternalWholeFileHashesAllContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	extPath := "external/myfile.txt"
	testWriteFile(t, extPath, "external original")
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a"))

	chain := &chainresolver.Chain{
		Target:   testChainItem("code-from-spec/a/_node.md"),
		External: []*frontmatter.FrontmatterExternal{testExternalEntry(extPath)},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, extPath, "external modified")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hashes to differ after external file change, but both are %q", hash1)
	}
}

func TestExternalWithFragmentsHashesDeclaredRange(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	extPath := "external/tenlines.txt"
	lines := "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10\n"
	testWriteFile(t, extPath, lines)
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a"))

	chain := &chainresolver.Chain{
		Target:   testChainItem("code-from-spec/a/_node.md"),
		External: []*frontmatter.FrontmatterExternal{testExternalEntryWithFragments(extPath, []string{"3-5"})},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Modify line 4 (within range 3-5)
	modifiedLines := "line1\nline2\nline3\nMODIFIED_LINE4\nline5\nline6\nline7\nline8\nline9\nline10\n"
	testWriteFile(t, extPath, modifiedLines)

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hashes to differ after modifying line within fragment range, but both are %q", hash1)
	}
}

func TestExternalWithFragmentsChangeOutsideRangeIgnored(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	extPath := "external/tenlines.txt"
	lines := "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10\n"
	testWriteFile(t, extPath, lines)
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a"))

	chain := &chainresolver.Chain{
		Target:   testChainItem("code-from-spec/a/_node.md"),
		External: []*frontmatter.FrontmatterExternal{testExternalEntryWithFragments(extPath, []string{"3-5"})},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Modify line 8 (outside range 3-5)
	modifiedLines := "line1\nline2\nline3\nline4\nline5\nline6\nline7\nMODIFIED_LINE8\nline9\nline10\n"
	testWriteFile(t, extPath, modifiedLines)

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("expected hashes to be equal after modifying line outside fragment range, got %q and %q", hash1, hash2)
	}
}

func TestExternalMultipleFragmentsDeclarationOrderMatters(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	extPath := "external/tenlines.txt"
	lines := "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10\n"
	testWriteFile(t, extPath, lines)
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a"))

	// Chain A: fragments = [6-8, 1-3]
	chainA := &chainresolver.Chain{
		Target:   testChainItem("code-from-spec/a/_node.md"),
		External: []*frontmatter.FrontmatterExternal{testExternalEntryWithFragments(extPath, []string{"6-8", "1-3"})},
	}

	hashA, err := chainhash.ChainHashCompute(chainA)
	if err != nil {
		t.Fatalf("chain A: %v", err)
	}

	// Chain B: fragments = [1-3, 6-8] (reversed)
	chainB := &chainresolver.Chain{
		Target:   testChainItem("code-from-spec/a/_node.md"),
		External: []*frontmatter.FrontmatterExternal{testExternalEntryWithFragments(extPath, []string{"1-3", "6-8"})},
	}

	hashB, err := chainhash.ChainHashCompute(chainB)
	if err != nil {
		t.Fatalf("chain B: %v", err)
	}

	if hashA == hashB {
		t.Errorf("expected hashes to differ based on fragment order, but both are %q", hashA)
	}
}

// --------------------------------------------------------------------------
// Target
// --------------------------------------------------------------------------

func TestTargetPublicAndAgentBothContribute(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContentWithPublicAndAgent("ROOT/a", "target public", "target agent"))

	chain := testMinimalChain(testChainItem("code-from-spec/a/_node.md"))

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Remove the # Agent section — Public only remains
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContentWithPublic("ROOT/a", "target public"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hashes to differ after removing Agent section, but both are %q", hash1)
	}
}

func TestTargetWithoutAgentAgentSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContentWithPublic("ROOT/a", "some public content"))

	chain := testMinimalChain(testChainItem("code-from-spec/a/_node.md"))

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hash) != 27 {
		t.Errorf("expected 27-char hash, got %d: %q", len(hash), hash)
	}
}

// --------------------------------------------------------------------------
// Input
// --------------------------------------------------------------------------

func TestInputHashesFileMinusFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	artifactPath := "artifacts/input.md"
	testWriteFile(t, artifactPath, testArtifactContent("", "input body original"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a"))

	chain := &chainresolver.Chain{
		Target: testChainItem("code-from-spec/a/_node.md"),
		Input:  testChainItemArtifact("ARTIFACT/a", artifactPath),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, artifactPath, testArtifactContent("", "input body modified"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hashes to differ after input body change, but both are %q", hash1)
	}
}

func TestNoInputSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a"))

	chain := &chainresolver.Chain{
		Target: testChainItem("code-from-spec/a/_node.md"),
		Input:  nil,
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hash) != 27 {
		t.Errorf("expected 27-char hash, got %d: %q", len(hash), hash)
	}
}

// --------------------------------------------------------------------------
// Error Cases
// --------------------------------------------------------------------------

func TestUnreadableSpecNodeFileReturnsError(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	// Reference a file that does not exist
	chain := testMinimalChain(testChainItem("code-from-spec/nonexistent/_node.md"))

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for unreadable spec node file, got nil")
	}
}

func TestUnreadableArtifactFileReturnsError(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a"))

	chain := &chainresolver.Chain{
		Target: testChainItem("code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItemArtifact("ARTIFACT/missing", "artifacts/nonexistent.md"),
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for unreadable artifact file, got nil")
	}
}

func TestUnreadableExternalFileReturnsError(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a"))

	chain := &chainresolver.Chain{
		Target:   testChainItem("code-from-spec/a/_node.md"),
		External: []*frontmatter.FrontmatterExternal{testExternalEntry("external/nonexistent.txt")},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for unreadable external file, got nil")
	}
}
