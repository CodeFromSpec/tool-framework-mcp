// code-from-spec: ROOT/golang/tests/chain/hash@5j4ND3a_eNe3BVubxapdZNrIou8
package chainhash_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
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

func testWriteNode(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath_dir(path), 0755); err != nil {
		t.Fatalf("testWriteNode mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNode write: %v", err)
	}
}

func filepath_dir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return "."
}

func testMakeRootNode(t *testing.T) {
	t.Helper()
	testWriteNode(t, "code-from-spec/_node.md", "# ROOT\n")
}

func testMakeNodeWithPublic(t *testing.T, logicalName string, cfsPath string, publicContent string) {
	t.Helper()
	content := "# " + logicalName + "\n\n# Public\n\n" + publicContent + "\n"
	testWriteNode(t, cfsPath, content)
}

func testMakeNodeWithPublicAndAgent(t *testing.T, logicalName string, cfsPath string, publicContent string, agentContent string) {
	t.Helper()
	content := "# " + logicalName + "\n\n# Public\n\n" + publicContent + "\n\n# Agent\n\n" + agentContent + "\n"
	testWriteNode(t, cfsPath, content)
}

func testMakeNodeWithPublicAndSubsection(t *testing.T, logicalName string, cfsPath string, subsectionHeading string, subsectionContent string) {
	t.Helper()
	content := "# " + logicalName + "\n\n# Public\n\n## " + subsectionHeading + "\n\n" + subsectionContent + "\n"
	testWriteNode(t, cfsPath, content)
}

func testMakeMinimalNode(t *testing.T, logicalName string, cfsPath string) {
	t.Helper()
	content := "# " + logicalName + "\n"
	testWriteNode(t, cfsPath, content)
}

func testChainItem(logicalName string, cfsPath string, qualifier string) *chainresolver.ChainItem {
	return &chainresolver.ChainItem{
		LogicalName: logicalName,
		FilePath:    pathutils.PathCfs{Value: cfsPath},
		Qualifier:   qualifier,
	}
}

func testMinimalChain(target *chainresolver.ChainItem) *chainresolver.Chain {
	return &chainresolver.Chain{
		Target: target,
	}
}

func TestChainHashCompute_Deterministic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMakeRootNode(t)
	testMakeNodeWithPublic(t, "ROOT/a", "code-from-spec/a/_node.md", "Some content.")

	chain := testMinimalChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""))

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

func TestChainHashCompute_27Characters(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMakeRootNode(t)
	testMakeNodeWithPublic(t, "ROOT/a", "code-from-spec/a/_node.md", "Some content.")

	chain := testMinimalChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""))

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27-character hash, got %d characters: %q", len(hash), hash)
	}
}

func TestChainHashCompute_ChangesWhenAncestorContentChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMakeNodeWithPublic(t, "ROOT", "code-from-spec/_node.md", "Original public content.")
	testMakeMinimalNode(t, "ROOT/a", "code-from-spec/a/_node.md")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md", ""),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteNode(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n\nModified public content.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after ancestor content change")
	}
}

func TestChainHashCompute_ChangesWhenDependencyContentChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMakeRootNode(t)
	testMakeMinimalNode(t, "ROOT/a", "code-from-spec/a/_node.md")
	testMakeNodeWithPublic(t, "ROOT/b", "code-from-spec/b/_node.md", "Original content.")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ROOT/b", "code-from-spec/b/_node.md", ""),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteNode(t, "code-from-spec/b/_node.md", "# ROOT/b\n\n# Public\n\nModified content.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after dependency content change")
	}
}

func TestChainHashCompute_ChangesWhenTargetPublicChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMakeRootNode(t)
	testMakeNodeWithPublic(t, "ROOT/a", "code-from-spec/a/_node.md", "Original public.")

	chain := testMinimalChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""))

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nModified public.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after target Public change")
	}
}

func TestChainHashCompute_ChangesWhenTargetAgentChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMakeRootNode(t)
	testMakeNodeWithPublicAndAgent(t, "ROOT/a", "code-from-spec/a/_node.md", "Public content.", "Original agent.")

	chain := testMinimalChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""))

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nPublic content.\n\n# Agent\n\nModified agent.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after target Agent change")
	}
}

func TestChainHashCompute_AncestorWithPublicContributesHash(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMakeNodeWithPublic(t, "ROOT", "code-from-spec/_node.md", "Root public content.")
	testMakeMinimalNode(t, "ROOT/a", "code-from-spec/a/_node.md")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md", ""),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27-character hash, got %d: %q", len(hash), hash)
	}
}

func TestChainHashCompute_AncestorWithoutPublicSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMakeMinimalNode(t, "ROOT", "code-from-spec/_node.md")
	testMakeNodeWithPublic(t, "ROOT/a", "code-from-spec/a/_node.md", "Target public.")
	testMakeNodeWithPublic(t, "ROOT/z", "code-from-spec/z/_node.md", "Z public content.")

	chainA := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md", ""),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
	}

	chainB := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT/z", "code-from-spec/z/_node.md", ""),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
	}

	hashA, err := chainhash.ChainHashCompute(chainA)
	if err != nil {
		t.Fatalf("chainA: %v", err)
	}
	hashB, err := chainhash.ChainHashCompute(chainB)
	if err != nil {
		t.Fatalf("chainB: %v", err)
	}

	if hashA == hashB {
		t.Error("expected hashes to differ: ancestor without Public should contribute nothing")
	}
}

func TestChainHashCompute_MultipleAncestorsOrderMatters(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMakeNodeWithPublic(t, "ROOT", "code-from-spec/_node.md", "Root public.")
	testMakeNodeWithPublic(t, "ROOT/a", "code-from-spec/a/_node.md", "A public.")
	testMakeMinimalNode(t, "ROOT/a/b", "code-from-spec/a/b/_node.md")

	chainX := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md", ""),
			testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		},
		Target: testChainItem("ROOT/a/b", "code-from-spec/a/b/_node.md", ""),
	}

	chainY := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
			testChainItem("ROOT", "code-from-spec/_node.md", ""),
		},
		Target: testChainItem("ROOT/a/b", "code-from-spec/a/b/_node.md", ""),
	}

	hashX, err := chainhash.ChainHashCompute(chainX)
	if err != nil {
		t.Fatalf("chainX: %v", err)
	}
	hashY, err := chainhash.ChainHashCompute(chainY)
	if err != nil {
		t.Fatalf("chainY: %v", err)
	}

	if hashX == hashY {
		t.Error("expected hashes to differ when ancestor order changes")
	}
}

func TestChainHashCompute_RootDependencyNoQualifierHashesPublic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMakeRootNode(t)
	testMakeMinimalNode(t, "ROOT/a", "code-from-spec/a/_node.md")
	testMakeNodeWithPublic(t, "ROOT/b", "code-from-spec/b/_node.md", "Original public.")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ROOT/b", "code-from-spec/b/_node.md", ""),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteNode(t, "code-from-spec/b/_node.md", "# ROOT/b\n\n# Public\n\nModified public.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after dependency Public change")
	}
}

func TestChainHashCompute_RootDependencyWithQualifierHashesSubsection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMakeRootNode(t)
	testMakeMinimalNode(t, "ROOT/a", "code-from-spec/a/_node.md")
	testMakeNodeWithPublicAndSubsection(t, "ROOT/b", "code-from-spec/b/_node.md", "Interface", "Original interface content.")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ROOT/b", "code-from-spec/b/_node.md", "interface"),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteNode(t, "code-from-spec/b/_node.md", "# ROOT/b\n\n# Public\n\n## Interface\n\nModified interface content.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after subsection content change")
	}
}

func TestChainHashCompute_QualifierCaseNormalization(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMakeRootNode(t)
	testMakeMinimalNode(t, "ROOT/a", "code-from-spec/a/_node.md")
	testMakeNodeWithPublicAndSubsection(t, "ROOT/b", "code-from-spec/b/_node.md", "Interface", "Interface content.")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ROOT/b", "code-from-spec/b/_node.md", "INTERFACE"),
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Errorf("expected no error with uppercase qualifier, got: %v", err)
	}
}

func TestChainHashCompute_ArtifactDependencyHashesBodyNotFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMakeRootNode(t)
	testMakeMinimalNode(t, "ROOT/a", "code-from-spec/a/_node.md")

	artifactContent := "---\noutput: some/path.go\n---\nOriginal body content.\n"
	testWriteNode(t, "artifacts/output.md", artifactContent)

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ARTIFACT/a", "artifacts/output.md", ""),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteNode(t, "artifacts/output.md", "---\noutput: some/path.go\n---\nModified body content.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after artifact body change")
	}
}

func TestChainHashCompute_ArtifactDependencyFrontmatterChangeIgnored(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMakeRootNode(t)
	testMakeMinimalNode(t, "ROOT/a", "code-from-spec/a/_node.md")

	testWriteNode(t, "artifacts/output.md", "---\noutput: some/path.go\n---\nBody content unchanged.\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ARTIFACT/a", "artifacts/output.md", ""),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteNode(t, "artifacts/output.md", "---\noutput: different/path.go\nextra: field\n---\nBody content unchanged.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 != hash2 {
		t.Error("expected hashes to be identical when only frontmatter changes")
	}
}

func TestChainHashCompute_ArtifactDependencyTagHashChangeIgnored(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMakeRootNode(t)
	testMakeMinimalNode(t, "ROOT/a", "code-from-spec/a/_node.md")

	testWriteNode(t, "artifacts/output.md", "---\noutput: some/path.go\n---\n// code-from-spec: ROOT/x/y@aAbBcCdDeEfFgGhHiIjJkKlLmMn\nOther body content.\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ARTIFACT/a", "artifacts/output.md", ""),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteNode(t, "artifacts/output.md", "---\noutput: some/path.go\n---\n// code-from-spec: ROOT/x/y@zZyYxXwWvVuUtTsSrRqQpPoOnNm\nOther body content.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 != hash2 {
		t.Error("expected hashes to be identical when only artifact tag hash changes")
	}
}

func TestChainHashCompute_ExternalFileHashesAllContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMakeRootNode(t)
	testMakeMinimalNode(t, "ROOT/a", "code-from-spec/a/_node.md")

	testWriteNode(t, "external/file.txt", "Original external content.\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		External: []*frontmatter.FrontmatterExternal{
			{Path: "external/file.txt"},
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteNode(t, "external/file.txt", "Modified external content.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after external file content change")
	}
}

func TestChainHashCompute_TargetPublicAndAgentBothContribute(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMakeRootNode(t)
	testMakeNodeWithPublicAndAgent(t, "ROOT/a", "code-from-spec/a/_node.md", "Public content.", "Agent content.")

	chain := testMinimalChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""))

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nPublic content.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after removing Agent section")
	}
}

func TestChainHashCompute_TargetWithoutAgentSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMakeRootNode(t)
	testMakeNodeWithPublic(t, "ROOT/a", "code-from-spec/a/_node.md", "Public content.")

	chain := testMinimalChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""))

	_, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Errorf("expected no error when target has no Agent section, got: %v", err)
	}
}

func TestChainHashCompute_InputHashesBodyNotFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMakeRootNode(t)
	testMakeMinimalNode(t, "ROOT/a", "code-from-spec/a/_node.md")

	testWriteNode(t, "input/file.md", "---\ninput: true\n---\nOriginal input body.\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Input:  testChainItem("ARTIFACT/a", "input/file.md", ""),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteNode(t, "input/file.md", "---\ninput: true\n---\nModified input body.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after input body change")
	}
}

func TestChainHashCompute_NoInputSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMakeRootNode(t)
	testMakeMinimalNode(t, "ROOT/a", "code-from-spec/a/_node.md")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Input:  nil,
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Errorf("expected no error when input is absent, got: %v", err)
	}
}

func TestChainHashCompute_UnreadableSpecNodeFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	chain := testMinimalChain(testChainItem("ROOT/nonexistent", "code-from-spec/nonexistent/_node.md", ""))

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for unreadable spec node file")
	}
	if !errors.Is(err, chainhash.ErrParseFailure) {
		t.Errorf("expected ErrParseFailure, got: %v", err)
	}
}

func TestChainHashCompute_UnreadableArtifactFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMakeRootNode(t)
	testMakeMinimalNode(t, "ROOT/a", "code-from-spec/a/_node.md")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ARTIFACT/missing", "artifacts/nonexistent.md", ""),
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for unreadable artifact file")
	}
	if !errors.Is(err, chainhash.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestChainHashCompute_UnreadableExternalFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMakeRootNode(t)
	testMakeMinimalNode(t, "ROOT/a", "code-from-spec/a/_node.md")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		External: []*frontmatter.FrontmatterExternal{
			{Path: "nonexistent/file.txt"},
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for unreadable external file")
	}
	if !errors.Is(err, chainhash.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}
