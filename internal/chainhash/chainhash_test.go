// code-from-spec: ROOT/golang/tests/chain/hash@5WXCXiJDidW-bBUpF4O0bp-HCfg
package chainhash_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// testChdir changes the working directory to dir and registers a cleanup
// to restore the original directory when the test ends.
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

// testWriteFile writes content to path, creating intermediate directories as needed.
func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("testWriteFile MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile WriteFile: %v", err)
	}
}

// testNodeFile returns the minimal content for a _node.md file with the given
// logical name (e.g. "root/a"). The heading must be the normalized form.
func testNodeFile(normalizedName string) string {
	return fmt.Sprintf("# %s\n", normalizedName)
}

// testNodeFileWithPublic returns a _node.md with a # Public section.
func testNodeFileWithPublic(normalizedName string, publicContent string) string {
	return fmt.Sprintf("# %s\n\n# Public\n\n%s\n", normalizedName, publicContent)
}

// testNodeFileWithPublicAndAgent returns a _node.md with # Public and # Agent sections.
func testNodeFileWithPublicAndAgent(normalizedName string, publicContent string, agentContent string) string {
	return fmt.Sprintf("# %s\n\n# Public\n\n%s\n\n# Agent\n\n%s\n", normalizedName, publicContent, agentContent)
}

// testNodeFileWithPublicSubsection returns a _node.md with a # Public section
// containing a ## <subsection> subsection.
func testNodeFileWithPublicSubsection(normalizedName string, subsectionName string, subsectionContent string) string {
	return fmt.Sprintf("# %s\n\n# Public\n\n## %s\n\n%s\n", normalizedName, subsectionName, subsectionContent)
}

// testChainItem creates a ChainItem for a ROOT/ spec node.
func testChainItem(logicalName string, filePath string) *chainresolver.ChainItem {
	return &chainresolver.ChainItem{
		LogicalName: logicalName,
		FilePath:    &pathutils.PathCfs{Value: filePath},
	}
}

// testChainItemWithQualifier creates a ChainItem for a ROOT/ spec node with a qualifier.
func testChainItemWithQualifier(logicalName string, filePath string, qualifier string) *chainresolver.ChainItem {
	return &chainresolver.ChainItem{
		LogicalName: logicalName,
		FilePath:    &pathutils.PathCfs{Value: filePath},
		Qualifier:   &qualifier,
	}
}

// testArtifactChainItem creates a ChainItem for an ARTIFACT/ dependency.
func testArtifactChainItem(logicalName string, filePath string) *chainresolver.ChainItem {
	return &chainresolver.ChainItem{
		LogicalName: logicalName,
		FilePath:    &pathutils.PathCfs{Value: filePath},
	}
}

// testMinimalChain builds a Chain with only a target node.
func testMinimalChain(target *chainresolver.ChainItem) *chainresolver.Chain {
	return &chainresolver.Chain{
		Target: target,
	}
}

// testArtifactContent returns file content with YAML frontmatter followed by a body.
func testArtifactContent(body string) string {
	return fmt.Sprintf("---\ndepends_on: []\n---\n%s", body)
}

// testExternalLines builds a string of n distinct lines.
func testExternalLines(n int) string {
	var sb strings.Builder
	for i := 1; i <= n; i++ {
		sb.WriteString(fmt.Sprintf("line %d content here\n", i))
	}
	return sb.String()
}

// testReplaceLineN replaces the nth line (1-indexed) of content with newLine.
func testReplaceLineN(content string, n int, newLine string) string {
	lines := strings.Split(content, "\n")
	if n > 0 && n <= len(lines) {
		lines[n-1] = newLine
	}
	return strings.Join(lines, "\n")
}

// ---------------------------------------------------------------------------
// Properties
// ---------------------------------------------------------------------------

func TestChainHashCompute_Deterministic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFileWithPublic("root/a", "some content"))

	chain := testMinimalChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md"))

	result1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (first): %v", err)
	}
	result2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (second): %v", err)
	}

	if result1 != result2 {
		t.Errorf("expected deterministic hash, got %q and %q", result1, result2)
	}
}

func TestChainHashCompute_Length27(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFileWithPublic("root/a", "some content"))

	chain := testMinimalChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md"))

	result, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute: %v", err)
	}

	if len(result) != 27 {
		t.Errorf("expected hash length 27, got %d (%q)", len(result), result)
	}
}

func TestChainHashCompute_ChangesWhenAncestorContentChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testNodeFileWithPublic("root", "original content"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md"),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (before): %v", err)
	}

	testWriteFile(t, "code-from-spec/_node.md", testNodeFileWithPublic("root", "modified content"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (after): %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change after ancestor content change, but got same hash %q", hash1)
	}
}

func TestChainHashCompute_ChangesWhenDependencyContentChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testNodeFile("root"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))
	testWriteFile(t, "code-from-spec/b/_node.md", testNodeFileWithPublic("root/b", "original dep content"))

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ROOT/b", "code-from-spec/b/_node.md"),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (before): %v", err)
	}

	testWriteFile(t, "code-from-spec/b/_node.md", testNodeFileWithPublic("root/b", "modified dep content"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (after): %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change after dependency content change, but got same hash %q", hash1)
	}
}

func TestChainHashCompute_ChangesWhenTargetPublicChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testNodeFile("root"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFileWithPublic("root/a", "original public"))

	chain := testMinimalChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md"))

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (before): %v", err)
	}

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFileWithPublic("root/a", "modified public"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (after): %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change after target Public change, but got same hash %q", hash1)
	}
}

func TestChainHashCompute_ChangesWhenTargetAgentChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testNodeFile("root"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFileWithPublicAndAgent("root/a", "public content", "original agent"))

	chain := testMinimalChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md"))

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (before): %v", err)
	}

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFileWithPublicAndAgent("root/a", "public content", "modified agent"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (after): %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change after target Agent change, but got same hash %q", hash1)
	}
}

// ---------------------------------------------------------------------------
// Ancestors
// ---------------------------------------------------------------------------

func TestChainHashCompute_AncestorWithPublicContributesHash(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testNodeFileWithPublic("root", "root public section"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md"),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	result, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute: %v", err)
	}

	if len(result) != 27 {
		t.Errorf("expected hash length 27, got %d (%q)", len(result), result)
	}
}

func TestChainHashCompute_AncestorWithoutPublicDiffersFromWithPublic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))

	// Chain with ancestor that has no public section
	testWriteFile(t, "code-from-spec/_node.md", testNodeFile("root"))

	chainNoPublic := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md"),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hashNoPublic, err := chainhash.ChainHashCompute(chainNoPublic)
	if err != nil {
		t.Fatalf("ChainHashCompute (no public): %v", err)
	}

	// Chain with ancestor that has a public section
	testWriteFile(t, "code-from-spec/_node.md", testNodeFileWithPublic("root", "some content in public"))

	chainWithPublic := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md"),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hashWithPublic, err := chainhash.ChainHashCompute(chainWithPublic)
	if err != nil {
		t.Fatalf("ChainHashCompute (with public): %v", err)
	}

	if hashNoPublic == hashWithPublic {
		t.Errorf("expected different hashes for ancestor with and without Public section, got same hash %q", hashNoPublic)
	}
}

func TestChainHashCompute_MultipleAncestorsOrderMatters(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testNodeFileWithPublic("root", "root public content"))
	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFileWithPublic("root/a", "a public content"))
	testWriteFile(t, "code-from-spec/a/b/_node.md", testNodeFile("root/a/b"))

	chainForward := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md"),
			testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		},
		Target: testChainItem("ROOT/a/b", "code-from-spec/a/b/_node.md"),
	}

	chainReversed := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
			testChainItem("ROOT", "code-from-spec/_node.md"),
		},
		Target: testChainItem("ROOT/a/b", "code-from-spec/a/b/_node.md"),
	}

	hashForward, err := chainhash.ChainHashCompute(chainForward)
	if err != nil {
		t.Fatalf("ChainHashCompute (forward): %v", err)
	}

	hashReversed, err := chainhash.ChainHashCompute(chainReversed)
	if err != nil {
		t.Fatalf("ChainHashCompute (reversed): %v", err)
	}

	if hashForward == hashReversed {
		t.Errorf("expected different hashes for different ancestor orders, got same hash %q", hashForward)
	}
}

// ---------------------------------------------------------------------------
// Dependencies
// ---------------------------------------------------------------------------

func TestChainHashCompute_RootDepNoQualifierHashesPublic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))
	testWriteFile(t, "code-from-spec/b/_node.md", testNodeFileWithPublic("root/b", "original public dep"))

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ROOT/b", "code-from-spec/b/_node.md"),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (before): %v", err)
	}

	testWriteFile(t, "code-from-spec/b/_node.md", testNodeFileWithPublic("root/b", "modified public dep"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (after): %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change after ROOT dep Public change, got same hash %q", hash1)
	}
}

func TestChainHashCompute_RootDepWithQualifierHashesSubsection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))
	testWriteFile(t, "code-from-spec/b/_node.md",
		testNodeFileWithPublicSubsection("root/b", "Interface", "original interface content"))

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItemWithQualifier("ROOT/b", "code-from-spec/b/_node.md", "interface"),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (before): %v", err)
	}

	testWriteFile(t, "code-from-spec/b/_node.md",
		testNodeFileWithPublicSubsection("root/b", "Interface", "modified interface content"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (after): %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change after qualified dep subsection change, got same hash %q", hash1)
	}
}

func TestChainHashCompute_QualifierCaseNormalization(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))
	testWriteFile(t, "code-from-spec/b/_node.md",
		testNodeFileWithPublicSubsection("root/b", "Interface", "interface content"))

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItemWithQualifier("ROOT/b", "code-from-spec/b/_node.md", "INTERFACE"),
		},
	}

	result, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute: %v", err)
	}

	if len(result) != 27 {
		t.Errorf("expected hash length 27, got %d (%q)", len(result), result)
	}
}

func TestChainHashCompute_ArtifactDepHashesBodyNotFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))
	testWriteFile(t, "generated/artifact.go", testArtifactContent("original body content\n"))

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testArtifactChainItem("ARTIFACT/a", "generated/artifact.go"),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (before): %v", err)
	}

	testWriteFile(t, "generated/artifact.go", testArtifactContent("modified body content\n"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (after): %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change after artifact body change, got same hash %q", hash1)
	}
}

func TestChainHashCompute_ArtifactDepFrontmatterChangeIgnored(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))
	testWriteFile(t, "generated/artifact.go", "---\ndepends_on: []\n---\nstable body content\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testArtifactChainItem("ARTIFACT/a", "generated/artifact.go"),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (before): %v", err)
	}

	// Change only frontmatter, keep body the same
	testWriteFile(t, "generated/artifact.go", "---\ndepends_on: [x]\n---\nstable body content\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (after): %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("expected hash to remain stable after artifact frontmatter-only change, got %q and %q", hash1, hash2)
	}
}

// ---------------------------------------------------------------------------
// External files
// ---------------------------------------------------------------------------

func TestChainHashCompute_ExternalWholeFileHashesAllContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))
	testWriteFile(t, "external/file.txt", "original external content\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		External: []*frontmatter.FrontmatterExternal{
			{Path: "external/file.txt"},
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (before): %v", err)
	}

	testWriteFile(t, "external/file.txt", "modified external content\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (after): %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change after external file content change, got same hash %q", hash1)
	}
}

func TestChainHashCompute_ExternalWithFragmentsHashesDeclaredRange(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))
	original := testExternalLines(10)
	testWriteFile(t, "external/file.txt", original)

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		External: []*frontmatter.FrontmatterExternal{
			{
				Path: "external/file.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "3-5"},
				},
			},
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (before): %v", err)
	}

	// Modify line 4 (inside range 3-5)
	modified := testReplaceLineN(original, 4, "MODIFIED line 4 content here")
	testWriteFile(t, "external/file.txt", modified)

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (after): %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change after modifying line inside fragment range, got same hash %q", hash1)
	}
}

func TestChainHashCompute_ExternalWithFragmentsChangeOutsideRangeIgnored(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))
	original := testExternalLines(10)
	testWriteFile(t, "external/file.txt", original)

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		External: []*frontmatter.FrontmatterExternal{
			{
				Path: "external/file.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "3-5"},
				},
			},
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (before): %v", err)
	}

	// Modify line 8 (outside range 3-5)
	modified := testReplaceLineN(original, 8, "MODIFIED line 8 content here")
	testWriteFile(t, "external/file.txt", modified)

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (after): %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("expected hash to stay stable after modifying line outside fragment range, got %q and %q", hash1, hash2)
	}
}

func TestChainHashCompute_ExternalMultipleFragmentsDeclarationOrderMatters(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))
	content := testExternalLines(10)
	testWriteFile(t, "external/file.txt", content)

	chainOrderA := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		External: []*frontmatter.FrontmatterExternal{
			{
				Path: "external/file.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "6-8"},
					{Lines: "1-3"},
				},
			},
		},
	}

	chainOrderB := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		External: []*frontmatter.FrontmatterExternal{
			{
				Path: "external/file.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "1-3"},
					{Lines: "6-8"},
				},
			},
		},
	}

	hashA, err := chainhash.ChainHashCompute(chainOrderA)
	if err != nil {
		t.Fatalf("ChainHashCompute (order A): %v", err)
	}

	hashB, err := chainhash.ChainHashCompute(chainOrderB)
	if err != nil {
		t.Fatalf("ChainHashCompute (order B): %v", err)
	}

	if hashA == hashB {
		t.Errorf("expected different hashes for different fragment declaration orders, got same hash %q", hashA)
	}
}

// ---------------------------------------------------------------------------
// Target
// ---------------------------------------------------------------------------

func TestChainHashCompute_TargetPublicAndAgentBothContribute(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md",
		testNodeFileWithPublicAndAgent("root/a", "public section content", "agent section content"))

	chain := testMinimalChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md"))

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (with agent): %v", err)
	}

	// Remove the Agent section
	testWriteFile(t, "code-from-spec/a/_node.md",
		testNodeFileWithPublic("root/a", "public section content"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (without agent): %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected different hashes when Agent section is present vs absent, got same hash %q", hash1)
	}
}

func TestChainHashCompute_TargetWithoutAgentSkipsAgent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md",
		testNodeFileWithPublic("root/a", "public section only"))

	chain := testMinimalChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md"))

	result, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute: %v", err)
	}

	if len(result) != 27 {
		t.Errorf("expected hash length 27, got %d (%q)", len(result), result)
	}
}

// ---------------------------------------------------------------------------
// Input
// ---------------------------------------------------------------------------

func TestChainHashCompute_InputHashesBodyNotFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))
	testWriteFile(t, "input/artifact.md", testArtifactContent("original input body\n"))

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Input:  testArtifactChainItem("ARTIFACT/input", "input/artifact.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (before): %v", err)
	}

	testWriteFile(t, "input/artifact.md", testArtifactContent("modified input body\n"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute (after): %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change after input body change, got same hash %q", hash1)
	}
}

func TestChainHashCompute_NoInputSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFileWithPublic("root/a", "some content"))

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Input:  nil,
	}

	result, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute: %v", err)
	}

	if len(result) != 27 {
		t.Errorf("expected hash length 27, got %d (%q)", len(result), result)
	}
}

// ---------------------------------------------------------------------------
// Error cases
// ---------------------------------------------------------------------------

func TestChainHashCompute_UnreadableSpecNodeFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// Do NOT create code-from-spec/a/_node.md — it should not exist
	chain := testMinimalChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md"))

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected an error for missing spec node file, got nil")
	}

	if !errors.Is(err, chainhash.ErrParseFailure) {
		t.Errorf("expected ErrParseFailure, got: %v", err)
	}
}

func TestChainHashCompute_UnreadableArtifactFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))

	// Artifact file does not exist
	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testArtifactChainItem("ARTIFACT/x", "generated/nonexistent.go"),
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected an error for missing artifact file, got nil")
	}

	if !errors.Is(err, chainhash.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestChainHashCompute_UnreadableExternalFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", testNodeFile("root/a"))

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		External: []*frontmatter.FrontmatterExternal{
			{Path: "external/nonexistent.txt"},
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected an error for missing external file, got nil")
	}

	if !errors.Is(err, chainhash.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}
