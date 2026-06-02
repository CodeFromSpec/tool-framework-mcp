// code-from-spec: ROOT/golang/tests/chain/hash@T34I2qo29E8v86N8vtQYAIDlfLU
package chainhash_test

import (
	"errors"
	"fmt"
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

func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(dirOf(path), 0755); err != nil {
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return "."
}

func testMakeNodeFile(logicalName string, publicContent string, agentContent string) string {
	content := fmt.Sprintf("# %s\n", logicalName)
	if publicContent != "" {
		content += fmt.Sprintf("\n# Public\n\n%s\n", publicContent)
	}
	if agentContent != "" {
		content += fmt.Sprintf("\n# Agent\n\n%s\n", agentContent)
	}
	return content
}

func testMakeNodeFileWithSubsection(logicalName string, subsectionName string, subsectionContent string) string {
	content := fmt.Sprintf("# %s\n\n# Public\n\n## %s\n\n%s\n", logicalName, subsectionName, subsectionContent)
	return content
}

func testMakeChainItem(logicalName string, filePath string, qualifier string) *chainresolver.ChainItem {
	return &chainresolver.ChainItem{
		LogicalName: logicalName,
		FilePath:    pathutils.PathCfs{Value: filePath},
		Qualifier:   qualifier,
	}
}

func testMakeArtifactFile(frontmatterContent string, bodyContent string) string {
	if frontmatterContent == "" {
		return bodyContent
	}
	return fmt.Sprintf("---\n%s\n---\n%s", frontmatterContent, bodyContent)
}

func TestHashIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testMakeNodeFile("ROOT", "root public", ""))
	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "a public", ""))

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testMakeChainItem("ROOT", "code-from-spec/_node.md", ""),
		},
		Target: testMakeChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
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
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testMakeNodeFile("ROOT", "root public", ""))
	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "a public", ""))

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testMakeChainItem("ROOT", "code-from-spec/_node.md", ""),
		},
		Target: testMakeChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27 characters, got %d: %q", len(hash), hash)
	}
}

func TestHashChangesWhenAncestorContentChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testMakeNodeFile("ROOT", "original content", ""))
	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "a public", ""))

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testMakeChainItem("ROOT", "code-from-spec/_node.md", ""),
		},
		Target: testMakeChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "code-from-spec/_node.md", testMakeNodeFile("ROOT", "modified content", ""))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after ancestor content change")
	}
}

func TestHashChangesWhenDependencyContentChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testMakeNodeFile("ROOT", "root public", ""))
	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "a public", ""))
	testWriteFile(t, "code-from-spec/b/_node.md", testMakeNodeFile("ROOT/b", "dep content", ""))

	chain := &chainresolver.Chain{
		Target: testMakeChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Dependencies: []*chainresolver.ChainItem{
			testMakeChainItem("ROOT/b", "code-from-spec/b/_node.md", ""),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "code-from-spec/b/_node.md", testMakeNodeFile("ROOT/b", "modified dep content", ""))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after dependency content change")
	}
}

func TestHashChangesWhenTargetPublicChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testMakeNodeFile("ROOT", "root public", ""))
	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "original", ""))

	chain := &chainresolver.Chain{
		Target: testMakeChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "changed", ""))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after target Public change")
	}
}

func TestHashChangesWhenTargetAgentChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testMakeNodeFile("ROOT", "root public", ""))
	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "a public", "original agent"))

	chain := &chainresolver.Chain{
		Target: testMakeChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "a public", "changed agent"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after target Agent change")
	}
}

func TestAncestorWithPublicSectionContributesHash(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testMakeNodeFile("ROOT", "ancestor content", ""))
	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "a public", ""))

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testMakeChainItem("ROOT", "code-from-spec/_node.md", ""),
		},
		Target: testMakeChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27 characters, got %d", len(hash))
	}
}

func TestAncestorWithoutPublicSectionSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "a public", ""))

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testMakeChainItem("ROOT", "code-from-spec/_node.md", ""),
		},
		Target: testMakeChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
	}

	hashNoPublic, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "code-from-spec/b/_node.md", testMakeNodeFile("ROOT/b", "some content", ""))

	chain2 := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testMakeChainItem("ROOT/b", "code-from-spec/b/_node.md", ""),
		},
		Target: testMakeChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
	}

	hashWithPublic, err := chainhash.ChainHashCompute(chain2)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hashNoPublic == hashWithPublic {
		t.Error("expected hashes to differ between ancestor with and without Public")
	}
}

func TestMultipleAncestorsOrderMatters(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testMakeNodeFile("ROOT", "root", ""))
	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "mid", ""))
	testWriteFile(t, "code-from-spec/a/b/_node.md", testMakeNodeFile("ROOT/a/b", "target", ""))

	chainNatural := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testMakeChainItem("ROOT", "code-from-spec/_node.md", ""),
			testMakeChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		},
		Target: testMakeChainItem("ROOT/a/b", "code-from-spec/a/b/_node.md", ""),
	}

	hashNatural, err := chainhash.ChainHashCompute(chainNatural)
	if err != nil {
		t.Fatalf("natural order call: %v", err)
	}

	chainReversed := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testMakeChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
			testMakeChainItem("ROOT", "code-from-spec/_node.md", ""),
		},
		Target: testMakeChainItem("ROOT/a/b", "code-from-spec/a/b/_node.md", ""),
	}

	hashReversed, err := chainhash.ChainHashCompute(chainReversed)
	if err != nil {
		t.Fatalf("reversed order call: %v", err)
	}

	if hashNatural == hashReversed {
		t.Error("expected hashes to differ for different ancestor orders")
	}
}

func TestRootDependencyWithoutQualifierHashesPublic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testMakeNodeFile("ROOT", "root public", ""))
	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "a public", ""))
	testWriteFile(t, "code-from-spec/b/_node.md", testMakeNodeFile("ROOT/b", "dep content", ""))

	chain := &chainresolver.Chain{
		Target: testMakeChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Dependencies: []*chainresolver.ChainItem{
			testMakeChainItem("ROOT/b", "code-from-spec/b/_node.md", ""),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "code-from-spec/b/_node.md", testMakeNodeFile("ROOT/b", "modified dep content", ""))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after dependency Public change")
	}
}

func TestRootDependencyWithQualifierHashesSubsection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testMakeNodeFile("ROOT", "root public", ""))
	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "a public", ""))
	testWriteFile(t, "code-from-spec/b/_node.md", testMakeNodeFileWithSubsection("ROOT/b", "Interface", "original interface"))

	chain := &chainresolver.Chain{
		Target: testMakeChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Dependencies: []*chainresolver.ChainItem{
			testMakeChainItem("ROOT/b", "code-from-spec/b/_node.md", "interface"),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "code-from-spec/b/_node.md", testMakeNodeFileWithSubsection("ROOT/b", "Interface", "modified interface"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after subsection content change")
	}
}

func TestQualifierCaseNormalization(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testMakeNodeFile("ROOT", "root public", ""))
	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "a public", ""))
	testWriteFile(t, "code-from-spec/b/_node.md", testMakeNodeFileWithSubsection("ROOT/b", "Interface", "interface content"))

	chain := &chainresolver.Chain{
		Target: testMakeChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Dependencies: []*chainresolver.ChainItem{
			testMakeChainItem("ROOT/b", "code-from-spec/b/_node.md", "INTERFACE"),
		},
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27 characters, got %d", len(hash))
	}
}

func TestArtifactDependencyHashesFileMinusFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testMakeNodeFile("ROOT", "root public", ""))
	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "a public", ""))
	testWriteFile(t, "artifacts/out.go", testMakeArtifactFile("output: artifacts/out.go", "body line\n"))

	chain := &chainresolver.Chain{
		Target: testMakeChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Dependencies: []*chainresolver.ChainItem{
			testMakeChainItem("ARTIFACT/some/node", "artifacts/out.go", ""),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "artifacts/out.go", testMakeArtifactFile("output: artifacts/out.go", "modified body line\n"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after artifact body change")
	}
}

func TestArtifactDependencyFrontmatterChangeIgnored(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testMakeNodeFile("ROOT", "root public", ""))
	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "a public", ""))
	testWriteFile(t, "artifacts/out.go", testMakeArtifactFile("output: artifacts/out.go", "body line\n"))

	chain := &chainresolver.Chain{
		Target: testMakeChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Dependencies: []*chainresolver.ChainItem{
			testMakeChainItem("ARTIFACT/some/node", "artifacts/out.go", ""),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "artifacts/out.go", testMakeArtifactFile("output: artifacts/out.go\nextra: field", "body line\n"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 != hash2 {
		t.Error("expected hashes to be equal after frontmatter-only change")
	}
}

func TestExternalFileHashesAllContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testMakeNodeFile("ROOT", "root public", ""))
	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "a public", ""))
	testWriteFile(t, "external/file.txt", "external content\n")

	chain := &chainresolver.Chain{
		Target: testMakeChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		External: []*frontmatter.FrontmatterExternal{
			{Path: "external/file.txt"},
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "external/file.txt", "modified external content\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after external file change")
	}
}

func TestTargetPublicAndAgentBothContribute(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testMakeNodeFile("ROOT", "root public", ""))
	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "pub", "agent"))

	chain := &chainresolver.Chain{
		Target: testMakeChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "pub", ""))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after removing Agent section")
	}
}

func TestTargetWithoutAgentSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testMakeNodeFile("ROOT", "root public", ""))
	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "a public", ""))

	chain := &chainresolver.Chain{
		Target: testMakeChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27 characters, got %d", len(hash))
	}
}

func TestInputHashesFileMinusFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testMakeNodeFile("ROOT", "root public", ""))
	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "a public", ""))
	testWriteFile(t, "artifacts/input.md", testMakeArtifactFile("output: artifacts/input.md", "input body\n"))

	chain := &chainresolver.Chain{
		Target: testMakeChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Input:  testMakeChainItem("ARTIFACT/input", "artifacts/input.md", ""),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	testWriteFile(t, "artifacts/input.md", testMakeArtifactFile("output: artifacts/input.md", "modified input body\n"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after input body change")
	}
}

func TestNoInputSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testMakeNodeFile("ROOT", "root public", ""))
	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "a public", ""))

	chain := &chainresolver.Chain{
		Target: testMakeChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27 characters, got %d", len(hash))
	}
}

func TestUnreadableSpecNodeFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testMakeNodeFile("ROOT", "root public", ""))

	chain := &chainresolver.Chain{
		Target: testMakeChainItem("ROOT/nonexistent", "code-from-spec/nonexistent/_node.md", ""),
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for missing spec node file")
	}
	if !errors.Is(err, chainhash.ErrParseFailure) {
		t.Errorf("expected ErrParseFailure, got: %v", err)
	}
}

func TestUnreadableArtifactFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testMakeNodeFile("ROOT", "root public", ""))
	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "a public", ""))

	chain := &chainresolver.Chain{
		Target: testMakeChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Dependencies: []*chainresolver.ChainItem{
			testMakeChainItem("ARTIFACT/missing", "artifacts/nonexistent.go", ""),
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for missing artifact file")
	}
	if !errors.Is(err, chainhash.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestUnreadableExternalFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", testMakeNodeFile("ROOT", "root public", ""))
	testWriteFile(t, "code-from-spec/a/_node.md", testMakeNodeFile("ROOT/a", "a public", ""))

	chain := &chainresolver.Chain{
		Target: testMakeChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		External: []*frontmatter.FrontmatterExternal{
			{Path: "external/nonexistent.txt"},
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for missing external file")
	}
	if !errors.Is(err, chainhash.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}
