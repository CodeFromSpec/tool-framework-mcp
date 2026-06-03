// code-from-spec: ROOT/golang/tests/chain/hash@fhY8mm66Tyt-eSkSNHchCiDwDrY
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

func testWriteNodeFile(t *testing.T, logicalName string, extra string) {
	t.Helper()
	path := "code-from-spec/" + logicalName[len("ROOT/"):] + "/_node.md"
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := fmt.Sprintf("# %s\n%s", logicalName, extra)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func testNodePath(logicalName string) pathutils.PathCfs {
	return pathutils.PathCfs{Value: "code-from-spec/" + logicalName[len("ROOT/"):] + "/_node.md"}
}

func testChainItem(logicalName string) *chainresolver.ChainItem {
	return &chainresolver.ChainItem{
		LogicalName: logicalName,
		FilePath:    testNodePath(logicalName),
	}
}

func testChainItemArtifact(logicalName string, filePath string) *chainresolver.ChainItem {
	return &chainresolver.ChainItem{
		LogicalName: logicalName,
		FilePath:    pathutils.PathCfs{Value: filePath},
	}
}

func TestHashIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT/a", "\n# Public\n\nSome content.\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash1 != hash2 {
		t.Errorf("hashes differ: %q vs %q", hash1, hash2)
	}
}

func TestHashIs27Characters(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT/a", "\n# Public\n\nSome content.\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a"),
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

	testWriteNodeFile(t, "ROOT", "\n# Public\n\nOriginal content.\n")
	testWriteNodeFile(t, "ROOT/a", "\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{testChainItem("ROOT")},
		Target:    testChainItem("ROOT/a"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "ROOT", "\n# Public\n\nModified content.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash1 == hash2 {
		t.Error("expected hashes to differ after ancestor content change")
	}
}

func TestHashChangesWhenDependencyContentChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", "\n")
	testWriteNodeFile(t, "ROOT/a", "\n")
	testWriteNodeFile(t, "ROOT/b", "\n# Public\n\nOriginal dep content.\n")

	chain := &chainresolver.Chain{
		Target:       testChainItem("ROOT/a"),
		Dependencies: []*chainresolver.ChainItem{testChainItem("ROOT/b")},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "ROOT/b", "\n# Public\n\nModified dep content.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash1 == hash2 {
		t.Error("expected hashes to differ after dependency content change")
	}
}

func TestHashChangesWhenTargetPublicChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT/a", "\n# Public\n\nOriginal target public.\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "ROOT/a", "\n# Public\n\nModified target public.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash1 == hash2 {
		t.Error("expected hashes to differ after target Public change")
	}
}

func TestHashChangesWhenTargetAgentChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT/a", "\n# Agent\n\nOriginal agent content.\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "ROOT/a", "\n# Agent\n\nModified agent content.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash1 == hash2 {
		t.Error("expected hashes to differ after target Agent change")
	}
}

func TestAncestorWithPublicContributesHash(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", "\n# Public\n\nAncestor public content.\n")
	testWriteNodeFile(t, "ROOT/a", "\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{testChainItem("ROOT")},
		Target:    testChainItem("ROOT/a"),
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27 characters, got %d: %q", len(hash), hash)
	}
}

func TestAncestorWithoutPublicSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", "\n")
	testWriteNodeFile(t, "ROOT/a", "\n# Public\n\nTarget public.\n")
	testWriteNodeFile(t, "ROOT/z", "\n# Public\n\nZ public content.\n")

	chainA := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{testChainItem("ROOT")},
		Target:    testChainItem("ROOT/a"),
	}
	chainB := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{testChainItem("ROOT/z")},
		Target:    testChainItem("ROOT/a"),
	}

	hashA, err := chainhash.ChainHashCompute(chainA)
	if err != nil {
		t.Fatalf("unexpected error chainA: %v", err)
	}
	hashB, err := chainhash.ChainHashCompute(chainB)
	if err != nil {
		t.Fatalf("unexpected error chainB: %v", err)
	}
	if hashA == hashB {
		t.Error("expected hashes to differ: ancestor without Public should contribute nothing")
	}
}

func TestMultipleAncestorsOrderMatters(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", "\n# Public\n\nRoot public.\n")
	testWriteNodeFile(t, "ROOT/a", "\n# Public\n\nA public.\n")
	testWriteNodeFile(t, "ROOT/a/b", "\n")

	chainX := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT"),
			testChainItem("ROOT/a"),
		},
		Target: testChainItem("ROOT/a/b"),
	}
	chainY := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT/a"),
			testChainItem("ROOT"),
		},
		Target: testChainItem("ROOT/a/b"),
	}

	hashX, err := chainhash.ChainHashCompute(chainX)
	if err != nil {
		t.Fatalf("unexpected error chainX: %v", err)
	}
	hashY, err := chainhash.ChainHashCompute(chainY)
	if err != nil {
		t.Fatalf("unexpected error chainY: %v", err)
	}
	if hashX == hashY {
		t.Error("expected hashes to differ when ancestor order differs")
	}
}

func TestRootDependencyWithoutQualifierHashesPublic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT/a", "\n")
	testWriteNodeFile(t, "ROOT/b", "\n# Public\n\nOriginal dep public.\n")

	chain := &chainresolver.Chain{
		Target:       testChainItem("ROOT/a"),
		Dependencies: []*chainresolver.ChainItem{testChainItem("ROOT/b")},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "ROOT/b", "\n# Public\n\nModified dep public.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash1 == hash2 {
		t.Error("expected hashes to differ after dependency Public change")
	}
}

func TestRootDependencyWithQualifierHashesSubsection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT/a", "\n")
	testWriteNodeFile(t, "ROOT/b", "\n# Public\n\n## Interface\n\nOriginal interface content.\n")

	depItem := &chainresolver.ChainItem{
		LogicalName: "ROOT/b",
		FilePath:    testNodePath("ROOT/b"),
		Qualifier:   "interface",
	}
	chain := &chainresolver.Chain{
		Target:       testChainItem("ROOT/a"),
		Dependencies: []*chainresolver.ChainItem{depItem},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "ROOT/b", "\n# Public\n\n## Interface\n\nModified interface content.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash1 == hash2 {
		t.Error("expected hashes to differ after qualified subsection change")
	}
}

func TestQualifierCaseNormalization(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT/a", "\n")
	testWriteNodeFile(t, "ROOT/b", "\n# Public\n\n## Interface\n\nInterface content.\n")

	depItem := &chainresolver.ChainItem{
		LogicalName: "ROOT/b",
		FilePath:    testNodePath("ROOT/b"),
		Qualifier:   "INTERFACE",
	}
	chain := &chainresolver.Chain{
		Target:       testChainItem("ROOT/a"),
		Dependencies: []*chainresolver.ChainItem{depItem},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error with uppercase qualifier: %v", err)
	}
}

func TestArtifactDependencyHashesFileMinusFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT/a", "\n")

	artifactPath := "artifacts/dep.go"
	if err := os.MkdirAll("artifacts", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	artifactContent := "---\noutput: artifacts/dep.go\n---\npackage main\n\n// original body\n"
	if err := os.WriteFile(artifactPath, []byte(artifactContent), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	depItem := testChainItemArtifact("ARTIFACT/a", artifactPath)
	chain := &chainresolver.Chain{
		Target:       testChainItem("ROOT/a"),
		Dependencies: []*chainresolver.ChainItem{depItem},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	artifactContent2 := "---\noutput: artifacts/dep.go\n---\npackage main\n\n// modified body\n"
	if err := os.WriteFile(artifactPath, []byte(artifactContent2), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash1 == hash2 {
		t.Error("expected hashes to differ after artifact body change")
	}
}

func TestArtifactDependencyFrontmatterChangeIgnored(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT/a", "\n")

	artifactPath := "artifacts/dep2.go"
	if err := os.MkdirAll("artifacts", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	artifactContent := "---\noutput: artifacts/dep2.go\n---\npackage main\n\n// body stays same\n"
	if err := os.WriteFile(artifactPath, []byte(artifactContent), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	depItem := testChainItemArtifact("ARTIFACT/a", artifactPath)
	chain := &chainresolver.Chain{
		Target:       testChainItem("ROOT/a"),
		Dependencies: []*chainresolver.ChainItem{depItem},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	artifactContent2 := "---\noutput: artifacts/dep2.go\ndepends_on: [ROOT/x]\n---\npackage main\n\n// body stays same\n"
	if err := os.WriteFile(artifactPath, []byte(artifactContent2), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash1 != hash2 {
		t.Error("expected hashes to be identical after frontmatter-only change")
	}
}

func TestArtifactDependencyTagHashChangeIgnored(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT/a", "\n")

	artifactPath := "artifacts/dep3.go"
	if err := os.MkdirAll("artifacts", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	artifactContent := "---\noutput: artifacts/dep3.go\n---\n// code-from-spec: ROOT/x/y@aAbBcCdDeEfFgGhHiIjJkKlLm\npackage main\n"
	if err := os.WriteFile(artifactPath, []byte(artifactContent), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	depItem := testChainItemArtifact("ARTIFACT/a", artifactPath)
	chain := &chainresolver.Chain{
		Target:       testChainItem("ROOT/a"),
		Dependencies: []*chainresolver.ChainItem{depItem},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	artifactContent2 := "---\noutput: artifacts/dep3.go\n---\n// code-from-spec: ROOT/x/y@zZyYxXwWvVuUtTsSrRqQpPoOnNm\npackage main\n"
	if err := os.WriteFile(artifactPath, []byte(artifactContent2), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash1 != hash2 {
		t.Error("expected hashes to be identical after tag hash change")
	}
}

func TestExternalFileHashesAllContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT/a", "\n")

	extPath := "external/file.txt"
	if err := os.MkdirAll("external", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(extPath, []byte("original external content\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	chain := &chainresolver.Chain{
		Target:   testChainItem("ROOT/a"),
		External: []*frontmatter.FrontmatterExternal{{Path: extPath}},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := os.WriteFile(extPath, []byte("modified external content\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash1 == hash2 {
		t.Error("expected hashes to differ after external file change")
	}
}

func TestTargetPublicAndAgentBothContribute(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT/a", "\n# Public\n\nPublic content.\n\n# Agent\n\nAgent content.\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "ROOT/a", "\n# Public\n\nPublic content.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash1 == hash2 {
		t.Error("expected hashes to differ after removing Agent section")
	}
}

func TestTargetWithoutAgentSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT/a", "\n# Public\n\nPublic only content.\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a"),
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInputHashesFileMinusFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT/a", "\n")

	inputPath := "inputs/input.md"
	if err := os.MkdirAll("inputs", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	inputContent := "---\nsome: frontmatter\n---\nOriginal input body.\n"
	if err := os.WriteFile(inputPath, []byte(inputContent), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a"),
		Input:  testChainItemArtifact("ARTIFACT/a/input", inputPath),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	inputContent2 := "---\nsome: frontmatter\n---\nModified input body.\n"
	if err := os.WriteFile(inputPath, []byte(inputContent2), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash1 == hash2 {
		t.Error("expected hashes to differ after input body change")
	}
}

func TestNoInputSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT/a", "\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a"),
		Input:  nil,
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnreadableSpecNodeFileReturnsParseFailure(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	chain := &chainresolver.Chain{
		Target: &chainresolver.ChainItem{
			LogicalName: "ROOT/nonexistent",
			FilePath:    pathutils.PathCfs{Value: "code-from-spec/nonexistent/_node.md"},
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if !errors.Is(err, chainhash.ErrParseFailure) {
		t.Errorf("expected ErrParseFailure, got: %v", err)
	}
}

func TestUnreadableArtifactFileReturnsFileUnreadable(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT/a", "\n")

	depItem := testChainItemArtifact("ARTIFACT/a", "artifacts/nonexistent.go")
	chain := &chainresolver.Chain{
		Target:       testChainItem("ROOT/a"),
		Dependencies: []*chainresolver.ChainItem{depItem},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if !errors.Is(err, chainhash.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestUnreadableExternalFileReturnsFileUnreadable(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT/a", "\n")

	chain := &chainresolver.Chain{
		Target:   testChainItem("ROOT/a"),
		External: []*frontmatter.FrontmatterExternal{{Path: "external/nonexistent.txt"}},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if !errors.Is(err, chainhash.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}
