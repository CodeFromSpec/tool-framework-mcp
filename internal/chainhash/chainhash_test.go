// code-from-spec: ROOT/golang/tests/chain/hash@z74vv_Nw9KlE7-9NbMrdKrrYPcI

package chainhash_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// testChdir changes the working directory to dir and registers a cleanup
// to restore the original directory.
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

// testWriteFile writes content to path, creating parent directories as needed.
func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("testWriteFile MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile WriteFile: %v", err)
	}
}

// testNodeFile returns the minimal node file content for a given logical name.
// logicalName should be like "ROOT/a" which results in heading "root/a".
func testNodeFile(heading string) string {
	return "# " + heading + "\n"
}

// testNodeFileWithPublic returns a node file with a # Public section.
func testNodeFileWithPublic(heading, publicContent string) string {
	return "# " + heading + "\n\n# Public\n\n" + publicContent + "\n"
}

// testNodeFileWithPublicAndAgent returns a node file with # Public and # Agent sections.
func testNodeFileWithPublicAndAgent(heading, publicContent, agentContent string) string {
	return "# " + heading + "\n\n# Public\n\n" + publicContent + "\n\n# Agent\n\n" + agentContent + "\n"
}

// testNodeFileWithAgent returns a node file with a # Agent section.
func testNodeFileWithAgent(heading, agentContent string) string {
	return "# " + heading + "\n\n# Agent\n\n" + agentContent + "\n"
}

// testNodeFileWithPublicSubsection returns a node file with a # Public section
// containing a ## subsection.
func testNodeFileWithPublicSubsection(heading, subsectionName, subsectionContent string) string {
	return "# " + heading + "\n\n# Public\n\n## " + subsectionName + "\n\n" + subsectionContent + "\n"
}

// testArtifactFile returns an artifact file with frontmatter and body.
func testArtifactFile(body string) string {
	return "---\n---\n" + body
}

// testChainItem builds a ChainItem for a ROOT spec node.
func testChainItem(logicalName string, cfsPath string) *chainresolver.ChainItem {
	return &chainresolver.ChainItem{
		LogicalName: logicalName,
		FilePath:    &pathutils.PathCfs{Value: cfsPath},
	}
}

// testChainItemWithQualifier builds a ChainItem with a qualifier.
func testChainItemWithQualifier(logicalName string, cfsPath string, qualifier string) *chainresolver.ChainItem {
	q := qualifier
	return &chainresolver.ChainItem{
		LogicalName: logicalName,
		FilePath:    &pathutils.PathCfs{Value: cfsPath},
		Qualifier:   &q,
	}
}

// testChain builds a minimal Chain with the given target.
func testChain(target *chainresolver.ChainItem) *chainresolver.Chain {
	return &chainresolver.Chain{
		Target: target,
	}
}

// --------------------------------------------------------------------------
// Properties
// --------------------------------------------------------------------------

func TestHashIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFileWithPublic("root/a", "some content"))

	chain := testChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md"))

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if hash1 != hash2 {
		t.Errorf("expected same hash; got %q and %q", hash1, hash2)
	}
}

func TestHashIs27Characters(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))

	chain := testChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md"))

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27 characters; got %d (%q)", len(hash), hash)
	}
}

func TestHashChangesWhenAncestorContentChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testNodeFileWithPublic("root", "original ancestor"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md"),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "code-from-spec/_node.md", testNodeFileWithPublic("root", "modified ancestor"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected different hashes after ancestor content change")
	}
}

func TestHashChangesWhenDependencyContentChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testNodeFile("root"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))
	testWriteFile(t, "code-from-spec/b/_node.md", testNodeFileWithPublic("root/b", "original dependency"))

	chain := &chainresolver.Chain{
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ROOT/b", "code-from-spec/b/_node.md"),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "code-from-spec/b/_node.md", testNodeFileWithPublic("root/b", "modified dependency"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected different hashes after dependency content change")
	}
}

func TestHashChangesWhenTargetPublicChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testNodeFile("root"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFileWithPublic("root/a", "original public"))

	chain := testChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md"))

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFileWithPublic("root/a", "modified public"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected different hashes after target public content change")
	}
}

func TestHashChangesWhenTargetAgentChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testNodeFile("root"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFileWithAgent("root/a", "original agent"))

	chain := testChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md"))

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFileWithAgent("root/a", "modified agent"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected different hashes after target agent content change")
	}
}

// --------------------------------------------------------------------------
// Ancestors
// --------------------------------------------------------------------------

func TestAncestorWithPublicSectionContributesHash(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testNodeFileWithPublic("root", "ancestor public content"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md"),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27 characters; got %d (%q)", len(hash), hash)
	}
	if hash == "" {
		t.Error("expected non-empty hash")
	}
}

func TestAncestorWithoutPublicSectionSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// Chain A: ROOT has no public section
	testWriteFile(t, "code-from-spec/_node.md", testNodeFile("root"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))

	chainA := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md"),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hashA, err := chainhash.ChainHashCompute(chainA)
	if err != nil {
		t.Fatalf("chain A: %v", err)
	}

	// Chain B: ROOT has a public section
	testWriteFile(t, "code-from-spec/_node.md", testNodeFileWithPublic("root", "some public content"))

	chainB := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md"),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hashB, err := chainhash.ChainHashCompute(chainB)
	if err != nil {
		t.Fatalf("chain B: %v", err)
	}

	if hashA == hashB {
		t.Error("expected different hashes when ancestor has vs. lacks public section")
	}
}

func TestMultipleAncestorsOrderMatters(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testNodeFileWithPublic("root", "root public"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFileWithPublic("root/a", "a public"))
	testWriteFile(t, "code-from-spec/a/b/_node.md", testNodeFile("root/a/b"))

	rootItem := testChainItem("ROOT", "code-from-spec/_node.md")
	aItem := testChainItem("ROOT/a", "code-from-spec/a/_node.md")
	abItem := testChainItem("ROOT/a/b", "code-from-spec/a/b/_node.md")

	// Chain A: root-first order
	chainA := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{rootItem, aItem},
		Target:    abItem,
	}

	// Chain B: reversed order
	chainB := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{aItem, rootItem},
		Target:    abItem,
	}

	hashA, err := chainhash.ChainHashCompute(chainA)
	if err != nil {
		t.Fatalf("chain A: %v", err)
	}
	hashB, err := chainhash.ChainHashCompute(chainB)
	if err != nil {
		t.Fatalf("chain B: %v", err)
	}

	if hashA == hashB {
		t.Error("expected different hashes for different ancestor orderings")
	}
}

// --------------------------------------------------------------------------
// Dependencies
// --------------------------------------------------------------------------

func TestRootDependencyWithoutQualifierHashesPublic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/b/_node.md", testNodeFileWithPublic("root/b", "b public original"))

	chain := &chainresolver.Chain{
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ROOT/b", "code-from-spec/b/_node.md"),
		},
		Target: testChainItem("ROOT/b", "code-from-spec/b/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "code-from-spec/b/_node.md", testNodeFileWithPublic("root/b", "b public modified"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected different hashes after dependency public content change")
	}
}

func TestRootDependencyWithQualifierHashesSubsection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/b/_node.md",
		testNodeFileWithPublicSubsection("root/b", "Interface", "interface original"))

	chain := &chainresolver.Chain{
		Dependencies: []*chainresolver.ChainItem{
			testChainItemWithQualifier("ROOT/b", "code-from-spec/b/_node.md", "interface"),
		},
		Target: testChainItem("ROOT/b", "code-from-spec/b/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "code-from-spec/b/_node.md",
		testNodeFileWithPublicSubsection("root/b", "Interface", "interface modified"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected different hashes after subsection content change")
	}
}

func TestQualifierCaseNormalization(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/b/_node.md",
		testNodeFileWithPublicSubsection("root/b", "Interface", "some content"))

	chain := &chainresolver.Chain{
		Dependencies: []*chainresolver.ChainItem{
			testChainItemWithQualifier("ROOT/b", "code-from-spec/b/_node.md", "INTERFACE"),
		},
		Target: testChainItem("ROOT/b", "code-from-spec/b/_node.md"),
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27 characters; got %d (%q)", len(hash), hash)
	}
}

func TestArtifactDependencyHashesFileMinusFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/b/_node.md", testNodeFile("root/b"))
	testWriteFile(t, "artifacts/dep.md", testArtifactFile("original body\n"))

	chain := &chainresolver.Chain{
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ARTIFACT/dep", "artifacts/dep.md"),
		},
		Target: testChainItem("ROOT/b", "code-from-spec/b/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "artifacts/dep.md", testArtifactFile("modified body\n"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected different hashes after artifact body change")
	}
}

func TestArtifactDependencyFrontmatterChangeIgnored(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/b/_node.md", testNodeFile("root/b"))
	testWriteFile(t, "artifacts/dep.md", "---\n---\nstable body\n")

	chain := &chainresolver.Chain{
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ARTIFACT/dep", "artifacts/dep.md"),
		},
		Target: testChainItem("ROOT/b", "code-from-spec/b/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Change only the frontmatter (add a field), body stays the same
	testWriteFile(t, "artifacts/dep.md", "---\nextra: value\n---\nstable body\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 != hash2 {
		t.Error("expected same hash after frontmatter-only change in artifact")
	}
}

// --------------------------------------------------------------------------
// External Files
// --------------------------------------------------------------------------

func TestExternalWholeFileHashesAllContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))
	testWriteFile(t, "external/file.txt", "external original\n")

	chain := &chainresolver.Chain{
		External: []*frontmatter.FrontmatterExternal{
			{Path: "external/file.txt", Fragments: nil},
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "external/file.txt", "external modified\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected different hashes after external file content change")
	}
}

func TestExternalWithFragmentsHashesDeclaredRanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))
	testWriteFile(t, "external/file.txt",
		"line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10\n")

	chain := &chainresolver.Chain{
		External: []*frontmatter.FrontmatterExternal{
			{
				Path: "external/file.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "3-5"},
				},
			},
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Modify line 4 (within range 3-5)
	testWriteFile(t, "external/file.txt",
		"line1\nline2\nline3\nMODIFIED_LINE4\nline5\nline6\nline7\nline8\nline9\nline10\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected different hashes after modifying line within fragment range")
	}
}

func TestExternalWithFragmentsChangeOutsideRangeIgnored(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))
	testWriteFile(t, "external/file.txt",
		"line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10\n")

	chain := &chainresolver.Chain{
		External: []*frontmatter.FrontmatterExternal{
			{
				Path: "external/file.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "3-5"},
				},
			},
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Modify line 8 (outside range 3-5)
	testWriteFile(t, "external/file.txt",
		"line1\nline2\nline3\nline4\nline5\nline6\nline7\nMODIFIED_LINE8\nline9\nline10\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 != hash2 {
		t.Error("expected same hash when change is outside declared fragment range")
	}
}

func TestExternalWithMultipleFragmentsDeclarationOrderMatters(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))
	testWriteFile(t, "external/file.txt",
		"line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10\n")

	// Chain A: fragments in order [6-8, 1-3]
	chainA := &chainresolver.Chain{
		External: []*frontmatter.FrontmatterExternal{
			{
				Path: "external/file.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "6-8"},
					{Lines: "1-3"},
				},
			},
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	// Chain B: fragments in reversed order [1-3, 6-8]
	chainB := &chainresolver.Chain{
		External: []*frontmatter.FrontmatterExternal{
			{
				Path: "external/file.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "1-3"},
					{Lines: "6-8"},
				},
			},
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hashA, err := chainhash.ChainHashCompute(chainA)
	if err != nil {
		t.Fatalf("chain A: %v", err)
	}
	hashB, err := chainhash.ChainHashCompute(chainB)
	if err != nil {
		t.Fatalf("chain B: %v", err)
	}

	if hashA == hashB {
		t.Error("expected different hashes for different fragment orderings")
	}
}

// --------------------------------------------------------------------------
// Target
// --------------------------------------------------------------------------

func TestTargetPublicAndAgentBothContribute(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md",
		testNodeFileWithPublicAndAgent("root/a", "target public", "target agent"))

	chain := testChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md"))

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Remove agent section
	testWriteFile(t, "code-from-spec/a/_node.md",
		testNodeFileWithPublic("root/a", "target public"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected different hashes after removing agent section")
	}
}

func TestTargetWithoutAgentAgentSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md",
		testNodeFileWithPublic("root/a", "only public, no agent"))

	chain := testChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md"))

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27 characters; got %d (%q)", len(hash), hash)
	}
}

// --------------------------------------------------------------------------
// Input
// --------------------------------------------------------------------------

func TestInputHashesFileMinusFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))
	testWriteFile(t, "artifacts/input.md", testArtifactFile("input body original\n"))

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Input:  testChainItem("ARTIFACT/input", "artifacts/input.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "artifacts/input.md", testArtifactFile("input body modified\n"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected different hashes after input body change")
	}
}

func TestNoInputSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Input:  nil,
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27 characters; got %d (%q)", len(hash), hash)
	}
}

// --------------------------------------------------------------------------
// Error Cases
// --------------------------------------------------------------------------

func TestUnreadableSpecNodeFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// Do NOT create the spec node file — it must not exist
	chain := testChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md"))

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected an error for unreadable spec node file, got nil")
	}
}

func TestUnreadableArtifactFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))

	// Do NOT create the artifact file
	chain := &chainresolver.Chain{
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ARTIFACT/missing", "artifacts/missing.md"),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected an error for unreadable artifact file, got nil")
	}
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable; got %v", err)
	}
}

func TestUnreadableExternalFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))

	// Do NOT create the external file
	chain := &chainresolver.Chain{
		External: []*frontmatter.FrontmatterExternal{
			{Path: "external/nonexistent.txt", Fragments: nil},
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected an error for unreadable external file, got nil")
	}
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable; got %v", err)
	}
}
