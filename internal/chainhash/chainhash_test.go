// code-from-spec: ROOT/golang/tests/chain/hash@qY7LdHa9HCyd275zKRHucvzdQDY

package chainhash_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// testChdir changes the working directory to dir for the duration of
// the test, restoring the original directory on cleanup.
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

// testMakeNodeFile writes a minimal _node.md file for the given logical
// name segment under the code-from-spec directory. The heading is the
// last path component. If publicContent is non-empty a "# Public" section
// is appended. If agentContent is non-empty a "# Agent" section is appended.
func testMakeNodeFile(t *testing.T, relDir string, name string, publicContent string, agentContent string) string {
	t.Helper()
	dir := filepath.Join("code-from-spec", relDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testMakeNodeFile mkdir: %v", err)
	}
	path := filepath.Join(dir, "_node.md")
	content := fmt.Sprintf("# %s\n\nsome content\n", name)
	if publicContent != "" {
		content += fmt.Sprintf("\n# Public\n\n%s\n", publicContent)
	}
	if agentContent != "" {
		content += fmt.Sprintf("\n# Agent\n\n%s\n", agentContent)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testMakeNodeFile write: %v", err)
	}
	// Return the CFS path (forward slashes).
	return filepath.ToSlash(filepath.Join("code-from-spec", relDir, "_node.md"))
}

// testMakeNodeFileWithSubsection writes a _node.md with a Public section
// that contains a named subsection.
func testMakeNodeFileWithSubsection(t *testing.T, relDir string, name string, subsectionName string, subsectionContent string) string {
	t.Helper()
	dir := filepath.Join("code-from-spec", relDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testMakeNodeFileWithSubsection mkdir: %v", err)
	}
	path := filepath.Join(dir, "_node.md")
	content := fmt.Sprintf("# %s\n\nsome content\n\n# Public\n\n## %s\n\n%s\n", name, subsectionName, subsectionContent)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testMakeNodeFileWithSubsection write: %v", err)
	}
	return filepath.ToSlash(filepath.Join("code-from-spec", relDir, "_node.md"))
}

// testMakeArtifactFile writes a file with YAML frontmatter and a body.
func testMakeArtifactFile(t *testing.T, relPath string, frontmatterContent string, body string) string {
	t.Helper()
	dir := filepath.Dir(relPath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("testMakeArtifactFile mkdir: %v", err)
		}
	}
	var content string
	if frontmatterContent != "" {
		content = fmt.Sprintf("---\n%s\n---\n%s", frontmatterContent, body)
	} else {
		content = body
	}
	if err := os.WriteFile(relPath, []byte(content), 0644); err != nil {
		t.Fatalf("testMakeArtifactFile write: %v", err)
	}
	return filepath.ToSlash(relPath)
}

// testMakeExternalFile writes a plain file at relPath with the given lines.
func testMakeExternalFile(t *testing.T, relPath string, lines []string) string {
	t.Helper()
	dir := filepath.Dir(relPath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("testMakeExternalFile mkdir: %v", err)
		}
	}
	content := ""
	for _, l := range lines {
		content += l + "\n"
	}
	if err := os.WriteFile(relPath, []byte(content), 0644); err != nil {
		t.Fatalf("testMakeExternalFile write: %v", err)
	}
	return filepath.ToSlash(relPath)
}

// testChainItem builds a *chainresolver.ChainItem from a CFS path string.
func testChainItem(t *testing.T, cfsPath string, qualifier *string) *chainresolver.ChainItem {
	t.Helper()
	return &chainresolver.ChainItem{
		LogicalName: cfsPath,
		FilePath:    &pathutils.PathCfs{Value: cfsPath},
		Qualifier:   qualifier,
	}
}

// testStrPtr returns a pointer to s.
func testStrPtr(s string) *string {
	return &s
}

// ---------------------------------------------------------------------------
// Properties
// ---------------------------------------------------------------------------

func TestHashIsDeterministic(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	cfsPath := testMakeNodeFile(t, "", "ROOT", "", "")
	item := testChainItem(t, cfsPath, nil)
	chain := &chainresolver.Chain{
		Target: item,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if hash1 != hash2 {
		t.Errorf("expected deterministic hash, got %q and %q", hash1, hash2)
	}
}

func TestHashIs27Characters(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	cfsPath := testMakeNodeFile(t, "", "ROOT", "some public content", "")
	item := testChainItem(t, cfsPath, nil)
	chain := &chainresolver.Chain{
		Target: item,
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27-character hash, got %d characters: %q", len(hash), hash)
	}
}

func TestHashChangesWhenAncestorContentChanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	rootCfsPath := testMakeNodeFile(t, "", "ROOT", "original content", "")
	aCfsPath := testMakeNodeFile(t, "a", "a", "", "")

	rootItem := testChainItem(t, rootCfsPath, nil)
	aItem := testChainItem(t, aCfsPath, nil)
	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{rootItem},
		Target:    aItem,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Overwrite ROOT with modified content.
	testMakeNodeFile(t, "", "ROOT", "modified content", "")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hash to change when ancestor content changes, but it did not")
	}
}

func TestHashChangesWhenDependencyContentChanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMakeNodeFile(t, "", "ROOT", "", "")
	aCfsPath := testMakeNodeFile(t, "a", "a", "", "")
	bCfsPath := testMakeNodeFile(t, "b", "b", "original content", "")

	aItem := testChainItem(t, aCfsPath, nil)
	bItem := testChainItem(t, bCfsPath, nil)
	chain := &chainresolver.Chain{
		Target:       aItem,
		Dependencies: []*chainresolver.ChainItem{bItem},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testMakeNodeFile(t, "b", "b", "modified content", "")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hash to change when dependency content changes, but it did not")
	}
}

func TestHashChangesWhenTargetPublicChanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMakeNodeFile(t, "", "ROOT", "", "")
	aCfsPath := testMakeNodeFile(t, "a", "a", "original content", "")

	aItem := testChainItem(t, aCfsPath, nil)
	chain := &chainresolver.Chain{
		Target: aItem,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testMakeNodeFile(t, "a", "a", "modified content", "")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hash to change when target Public changes, but it did not")
	}
}

func TestHashChangesWhenTargetAgentChanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMakeNodeFile(t, "", "ROOT", "", "")
	aCfsPath := testMakeNodeFile(t, "a", "a", "", "original agent instructions")

	aItem := testChainItem(t, aCfsPath, nil)
	chain := &chainresolver.Chain{
		Target: aItem,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testMakeNodeFile(t, "a", "a", "", "modified agent instructions")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hash to change when target Agent changes, but it did not")
	}
}

// ---------------------------------------------------------------------------
// Ancestors
// ---------------------------------------------------------------------------

func TestAncestorWithPublicSectionContributesHash(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	rootCfsPath := testMakeNodeFile(t, "", "ROOT", "some ancestor public content", "")
	aCfsPath := testMakeNodeFile(t, "a", "a", "", "")

	rootItem := testChainItem(t, rootCfsPath, nil)
	aItem := testChainItem(t, aCfsPath, nil)
	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{rootItem},
		Target:    aItem,
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27-character hash, got %d characters: %q", len(hash), hash)
	}
	if hash == "" {
		t.Error("expected non-empty hash")
	}
}

func TestAncestorWithoutPublicSectionSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	// Ancestor A: no Public section.
	aCfsPath := testMakeNodeFile(t, "ancestor_a", "ancestor_a", "", "")
	// Ancestor B: with Public section.
	bCfsPath := testMakeNodeFile(t, "ancestor_b", "ancestor_b", "some content", "")
	// Shared target.
	targetCfsPath := testMakeNodeFile(t, "target", "target", "", "")

	targetItem := testChainItem(t, targetCfsPath, nil)
	aItem := testChainItem(t, aCfsPath, nil)
	bItem := testChainItem(t, bCfsPath, nil)

	chain1 := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{aItem},
		Target:    targetItem,
	}
	chain2 := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{bItem},
		Target:    targetItem,
	}

	hash1, err := chainhash.ChainHashCompute(chain1)
	if err != nil {
		t.Fatalf("chain1: %v", err)
	}
	hash2, err := chainhash.ChainHashCompute(chain2)
	if err != nil {
		t.Fatalf("chain2: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ when one ancestor has Public and the other does not")
	}
}

func TestMultipleAncestorsOrderMatters(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	rootCfsPath := testMakeNodeFile(t, "", "ROOT", "root public", "")
	aCfsPath := testMakeNodeFile(t, "a", "a", "a public", "")
	bCfsPath := testMakeNodeFile(t, "a/b", "b", "", "")

	rootItem := testChainItem(t, rootCfsPath, nil)
	aItem := testChainItem(t, aCfsPath, nil)
	bItem := testChainItem(t, bCfsPath, nil)

	// Chain1: root-first order.
	chain1 := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{rootItem, aItem},
		Target:    bItem,
	}
	// Chain2: reversed ancestor order.
	chain2 := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{aItem, rootItem},
		Target:    bItem,
	}

	hash1, err := chainhash.ChainHashCompute(chain1)
	if err != nil {
		t.Fatalf("chain1: %v", err)
	}
	hash2, err := chainhash.ChainHashCompute(chain2)
	if err != nil {
		t.Fatalf("chain2: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ when ancestor order is reversed")
	}
}

// ---------------------------------------------------------------------------
// Dependencies
// ---------------------------------------------------------------------------

func TestRootDependencyWithoutQualifierHashesPublic(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	aCfsPath := testMakeNodeFile(t, "a", "a", "", "")
	bCfsPath := testMakeNodeFile(t, "b", "b", "original public", "")

	aItem := testChainItem(t, aCfsPath, nil)
	bItem := testChainItem(t, bCfsPath, nil)
	chain := &chainresolver.Chain{
		Target:       aItem,
		Dependencies: []*chainresolver.ChainItem{bItem},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testMakeNodeFile(t, "b", "b", "modified public", "")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hash to change when dependency Public changes")
	}
}

func TestRootDependencyWithQualifierHashesSubsection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	aCfsPath := testMakeNodeFile(t, "a", "a", "", "")
	bCfsPath := testMakeNodeFileWithSubsection(t, "b", "b", "Interface", "original interface content")

	aItem := testChainItem(t, aCfsPath, nil)
	bItem := testChainItem(t, bCfsPath, testStrPtr("interface"))
	chain := &chainresolver.Chain{
		Target:       aItem,
		Dependencies: []*chainresolver.ChainItem{bItem},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testMakeNodeFileWithSubsection(t, "b", "b", "Interface", "modified interface content")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hash to change when qualified subsection content changes")
	}
}

func TestQualifierCaseNormalization(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	aCfsPath := testMakeNodeFile(t, "a", "a", "", "")
	bCfsPath := testMakeNodeFileWithSubsection(t, "b", "b", "Interface", "some interface content")

	aItem := testChainItem(t, aCfsPath, nil)
	// Uppercase qualifier.
	bItem := testChainItem(t, bCfsPath, testStrPtr("INTERFACE"))
	chain := &chainresolver.Chain{
		Target:       aItem,
		Dependencies: []*chainresolver.ChainItem{bItem},
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27-character hash, got %d: %q", len(hash), hash)
	}
}

func TestArtifactDependencyHashesBodyNotFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	targetCfsPath := testMakeNodeFile(t, "target", "target", "", "")
	artifactPath := testMakeArtifactFile(t, "artifacts/dep.md", "key: value", "original body\n")

	targetItem := testChainItem(t, targetCfsPath, nil)
	artifactItem := &chainresolver.ChainItem{
		LogicalName: "ARTIFACT/dep",
		FilePath:    &pathutils.PathCfs{Value: artifactPath},
		Qualifier:   nil,
	}
	chain := &chainresolver.Chain{
		Target:       targetItem,
		Dependencies: []*chainresolver.ChainItem{artifactItem},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Modify only the body.
	testMakeArtifactFile(t, "artifacts/dep.md", "key: value", "modified body\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hash to change when artifact body changes")
	}
}

func TestArtifactDependencyFrontmatterChangeIgnored(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	targetCfsPath := testMakeNodeFile(t, "target", "target", "", "")
	artifactPath := testMakeArtifactFile(t, "artifacts/dep2.md", "key: value", "stable body\n")

	targetItem := testChainItem(t, targetCfsPath, nil)
	artifactItem := &chainresolver.ChainItem{
		LogicalName: "ARTIFACT/dep2",
		FilePath:    &pathutils.PathCfs{Value: artifactPath},
		Qualifier:   nil,
	}
	chain := &chainresolver.Chain{
		Target:       targetItem,
		Dependencies: []*chainresolver.ChainItem{artifactItem},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Modify only the frontmatter.
	testMakeArtifactFile(t, "artifacts/dep2.md", "key: changed_value", "stable body\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 != hash2 {
		t.Error("expected hash to remain the same when only artifact frontmatter changes")
	}
}

// ---------------------------------------------------------------------------
// External files
// ---------------------------------------------------------------------------

func TestExternalWholeFileHashesAllContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	targetCfsPath := testMakeNodeFile(t, "target", "target", "", "")
	extPath := testMakeExternalFile(t, "external/file.txt", []string{"line1", "line2", "line3"})

	targetItem := testChainItem(t, targetCfsPath, nil)
	chain := &chainresolver.Chain{
		Target: targetItem,
		External: []*frontmatter.FrontmatterExternal{
			{
				Path:      extPath,
				Fragments: nil,
			},
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testMakeExternalFile(t, "external/file.txt", []string{"line1", "CHANGED", "line3"})

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hash to change when external file content changes")
	}
}

func TestExternalWithFragmentsHashesDeclaredRanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	lines := []string{"l1", "l2", "l3", "l4", "l5", "l6", "l7", "l8", "l9", "l10"}
	targetCfsPath := testMakeNodeFile(t, "target", "target", "", "")
	extPath := testMakeExternalFile(t, "external/frag.txt", lines)

	targetItem := testChainItem(t, targetCfsPath, nil)
	chain := &chainresolver.Chain{
		Target: targetItem,
		External: []*frontmatter.FrontmatterExternal{
			{
				Path: extPath,
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "3-5"},
				},
			},
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Modify line 4 (within range 3-5).
	modLines := []string{"l1", "l2", "l3", "CHANGED_L4", "l5", "l6", "l7", "l8", "l9", "l10"}
	testMakeExternalFile(t, "external/frag.txt", modLines)

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hash to change when line within declared fragment range changes")
	}
}

func TestExternalWithFragmentsChangeOutsideRangeIgnored(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	lines := []string{"l1", "l2", "l3", "l4", "l5", "l6", "l7", "l8", "l9", "l10"}
	targetCfsPath := testMakeNodeFile(t, "target", "target", "", "")
	extPath := testMakeExternalFile(t, "external/frag2.txt", lines)

	targetItem := testChainItem(t, targetCfsPath, nil)
	chain := &chainresolver.Chain{
		Target: targetItem,
		External: []*frontmatter.FrontmatterExternal{
			{
				Path: extPath,
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "3-5"},
				},
			},
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Modify line 8 (outside range 3-5).
	modLines := []string{"l1", "l2", "l3", "l4", "l5", "l6", "l7", "CHANGED_L8", "l9", "l10"}
	testMakeExternalFile(t, "external/frag2.txt", modLines)

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 != hash2 {
		t.Error("expected hash to remain the same when change is outside declared fragment range")
	}
}

func TestExternalWithMultipleFragmentsDeclarationOrderMatters(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	lines := []string{"l1", "l2", "l3", "l4", "l5", "l6", "l7", "l8", "l9", "l10"}
	targetCfsPath := testMakeNodeFile(t, "target", "target", "", "")
	extPath := testMakeExternalFile(t, "external/frag3.txt", lines)

	targetItem := testChainItem(t, targetCfsPath, nil)

	chain1 := &chainresolver.Chain{
		Target: targetItem,
		External: []*frontmatter.FrontmatterExternal{
			{
				Path: extPath,
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "6-8"},
					{Lines: "1-3"},
				},
			},
		},
	}
	chain2 := &chainresolver.Chain{
		Target: targetItem,
		External: []*frontmatter.FrontmatterExternal{
			{
				Path: extPath,
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "1-3"},
					{Lines: "6-8"},
				},
			},
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain1)
	if err != nil {
		t.Fatalf("chain1: %v", err)
	}
	hash2, err := chainhash.ChainHashCompute(chain2)
	if err != nil {
		t.Fatalf("chain2: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ when fragment declaration order differs")
	}
}

// ---------------------------------------------------------------------------
// Target
// ---------------------------------------------------------------------------

func TestTargetPublicAndAgentBothContribute(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMakeNodeFile(t, "", "ROOT", "", "")
	aCfsPath := testMakeNodeFile(t, "a", "a", "some public content", "some agent content")

	aItem := testChainItem(t, aCfsPath, nil)
	chain := &chainresolver.Chain{
		Target: aItem,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Overwrite file without Agent section.
	testMakeNodeFile(t, "a", "a", "some public content", "")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hash to change when Agent section is removed from target")
	}
}

func TestTargetWithoutAgentNoError(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testMakeNodeFile(t, "", "ROOT", "", "")
	aCfsPath := testMakeNodeFile(t, "a", "a", "some public content", "")

	aItem := testChainItem(t, aCfsPath, nil)
	chain := &chainresolver.Chain{
		Target: aItem,
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27-character hash, got %d: %q", len(hash), hash)
	}
}

// ---------------------------------------------------------------------------
// Input
// ---------------------------------------------------------------------------

func TestInputHashesBodyNotFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	targetCfsPath := testMakeNodeFile(t, "target", "target", "", "")
	inputPath := testMakeArtifactFile(t, "artifacts/input.md", "key: value", "original body\n")

	targetItem := testChainItem(t, targetCfsPath, nil)
	inputItem := &chainresolver.ChainItem{
		LogicalName: "ARTIFACT/input",
		FilePath:    &pathutils.PathCfs{Value: inputPath},
		Qualifier:   nil,
	}
	chain := &chainresolver.Chain{
		Target: targetItem,
		Input:  inputItem,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testMakeArtifactFile(t, "artifacts/input.md", "key: value", "modified body\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hash to change when input body changes")
	}
}

func TestNoInputSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	targetCfsPath := testMakeNodeFile(t, "target", "target", "", "")

	targetItem := testChainItem(t, targetCfsPath, nil)
	chain := &chainresolver.Chain{
		Target: targetItem,
		Input:  nil,
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27-character hash, got %d: %q", len(hash), hash)
	}
}

// ---------------------------------------------------------------------------
// Error cases
// ---------------------------------------------------------------------------

func TestUnreadableSpecNodeFileReturnsParseFailure(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	// Point to a file that does not exist.
	item := testChainItem(t, "code-from-spec/nonexistent/_node.md", nil)
	chain := &chainresolver.Chain{
		Target: item,
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainhash.ErrParseFailure) {
		t.Errorf("expected ErrParseFailure, got: %v", err)
	}
}

func TestUnreadableArtifactFileReturnsFileUnreadable(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	targetCfsPath := testMakeNodeFile(t, "target", "target", "", "")

	targetItem := testChainItem(t, targetCfsPath, nil)
	// Artifact file does not exist.
	artifactItem := &chainresolver.ChainItem{
		LogicalName: "ARTIFACT/missing",
		FilePath:    &pathutils.PathCfs{Value: "artifacts/missing.md"},
		Qualifier:   nil,
	}
	chain := &chainresolver.Chain{
		Target:       targetItem,
		Dependencies: []*chainresolver.ChainItem{artifactItem},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainhash.ErrUnreadableFile) {
		t.Errorf("expected ErrUnreadableFile, got: %v", err)
	}
}

func TestUnreadableExternalFileReturnsFileUnreadable(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	targetCfsPath := testMakeNodeFile(t, "target", "target", "", "")

	targetItem := testChainItem(t, targetCfsPath, nil)
	chain := &chainresolver.Chain{
		Target: targetItem,
		External: []*frontmatter.FrontmatterExternal{
			{
				Path:      "external/nonexistent.txt",
				Fragments: nil,
			},
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainhash.ErrUnreadableFile) {
		t.Errorf("expected ErrUnreadableFile, got: %v", err)
	}
}
