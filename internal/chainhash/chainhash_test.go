// code-from-spec: ROOT/golang/tests/chain/hash@0rP0byDeFTDIrLhOkwU1BpoRoDA
package chainhash_test

import (
	"errors"
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

func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

func testChainItem(logicalName string, cfsPath string) *chainresolver.ChainItem {
	return &chainresolver.ChainItem{
		LogicalName: logicalName,
		FilePath:    pathutils.PathCfs{Value: cfsPath},
	}
}

func testChainItemQualified(logicalName string, cfsPath string, qualifier string) *chainresolver.ChainItem {
	q := qualifier
	return &chainresolver.ChainItem{
		LogicalName: logicalName,
		FilePath:    pathutils.PathCfs{Value: cfsPath},
		Qualifier:   &q,
	}
}

func TestHashIsDeterministic(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n\nSome content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nSome content.\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md"),
		},
		Dependencies: []*chainresolver.ChainItem{},
		External:     []*frontmatter.FrontmatterExternal{},
		Target:       testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
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
		t.Errorf("expected deterministic hash, got %q and %q", hash1, hash2)
	}
}

func TestHashIs27Characters(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n\nSome content.\n")

	chain := &chainresolver.Chain{
		Ancestors:    []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{},
		External:     []*frontmatter.FrontmatterExternal{},
		Target:       testChainItem("ROOT", "code-from-spec/_node.md"),
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
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n\nOriginal content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nTarget.\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md"),
		},
		Dependencies: []*chainresolver.ChainItem{},
		External:     []*frontmatter.FrontmatterExternal{},
		Target:       testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n\nModified content.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change when ancestor content changes, got same hash: %q", hash1)
	}
}

func TestHashChangesWhenDependencyContentChanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n\nRoot content.\n")
	testWriteFile(t, "code-from-spec/b/_node.md", "# ROOT/b\n\n# Public\n\nDependency original.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nTarget.\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md"),
		},
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ROOT/b", "code-from-spec/b/_node.md"),
		},
		External: []*frontmatter.FrontmatterExternal{},
		Target:   testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "code-from-spec/b/_node.md", "# ROOT/b\n\n# Public\n\nDependency modified.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change when dependency content changes, got same hash: %q", hash1)
	}
}

func TestHashChangesWhenTargetPublicChanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n\nRoot content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nTarget original.\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md"),
		},
		Dependencies: []*chainresolver.ChainItem{},
		External:     []*frontmatter.FrontmatterExternal{},
		Target:       testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nTarget modified.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change when target Public changes, got same hash: %q", hash1)
	}
}

func TestHashChangesWhenTargetAgentChanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n\nRoot content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nSome public.\n\n# Agent\n\nOriginal agent.\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md"),
		},
		Dependencies: []*chainresolver.ChainItem{},
		External:     []*frontmatter.FrontmatterExternal{},
		Target:       testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nSome public.\n\n# Agent\n\nModified agent.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change when target Agent changes, got same hash: %q", hash1)
	}
}

func TestAncestorWithPublicSectionContributesHash(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n\nSome content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nTarget.\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md"),
		},
		Dependencies: []*chainresolver.ChainItem{},
		External:     []*frontmatter.FrontmatterExternal{},
		Target:       testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27 characters, got %d: %q", len(hash), hash)
	}
	if hash == "" {
		t.Error("expected non-empty hash")
	}
}

func TestAncestorWithoutPublicSectionSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\nJust a description.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\nJust a description.\n")

	chainA := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md"),
		},
		Dependencies: []*chainresolver.ChainItem{},
		External:     []*frontmatter.FrontmatterExternal{},
		Target:       testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	testWriteFile(t, "code-from-spec/root2/_node.md", "# ROOT/root2\n\n# Public\n\nSome content.\n")
	testWriteFile(t, "code-from-spec/root2/a/_node.md", "# ROOT/root2/a\n\n# Public\n\nTarget.\n")

	chainB := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT/root2", "code-from-spec/root2/_node.md"),
		},
		Dependencies: []*chainresolver.ChainItem{},
		External:     []*frontmatter.FrontmatterExternal{},
		Target:       testChainItem("ROOT/root2/a", "code-from-spec/root2/a/_node.md"),
	}

	hashA, err := chainhash.ChainHashCompute(chainA)
	if err != nil {
		t.Fatalf("unexpected error for chainA: %v", err)
	}
	hashB, err := chainhash.ChainHashCompute(chainB)
	if err != nil {
		t.Fatalf("unexpected error for chainB: %v", err)
	}

	if hashA == hashB {
		t.Errorf("expected different hashes when ancestor has no Public section vs has one, got same: %q", hashA)
	}
}

func TestMultipleAncestorsOrderMatters(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n\nRoot public.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nIntermediate public.\n")
	testWriteFile(t, "code-from-spec/a/b/_node.md", "# ROOT/a/b\n\n# Public\n\nTarget.\n")

	chainForward := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md"),
			testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		},
		Dependencies: []*chainresolver.ChainItem{},
		External:     []*frontmatter.FrontmatterExternal{},
		Target:       testChainItem("ROOT/a/b", "code-from-spec/a/b/_node.md"),
	}

	chainSwapped := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
			testChainItem("ROOT", "code-from-spec/_node.md"),
		},
		Dependencies: []*chainresolver.ChainItem{},
		External:     []*frontmatter.FrontmatterExternal{},
		Target:       testChainItem("ROOT/a/b", "code-from-spec/a/b/_node.md"),
	}

	hashForward, err := chainhash.ChainHashCompute(chainForward)
	if err != nil {
		t.Fatalf("unexpected error for chainForward: %v", err)
	}
	hashSwapped, err := chainhash.ChainHashCompute(chainSwapped)
	if err != nil {
		t.Fatalf("unexpected error for chainSwapped: %v", err)
	}

	if hashForward == hashSwapped {
		t.Errorf("expected different hashes for different ancestor order, got same: %q", hashForward)
	}
}

func TestRootDependencyWithoutQualifierHashesPublic(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/b/_node.md", "# ROOT/b\n\n# Public\n\nOriginal dep public.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nTarget.\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ROOT/b", "code-from-spec/b/_node.md"),
		},
		External: []*frontmatter.FrontmatterExternal{},
		Target:   testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "code-from-spec/b/_node.md", "# ROOT/b\n\n# Public\n\nModified dep public.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change when dependency Public changes, got same hash: %q", hash1)
	}
}

func TestRootDependencyWithQualifierHashesSubsection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/b/_node.md", "# ROOT/b\n\n# Public\n\n## Interface\n\nOriginal interface content.\n\n## Other\n\nOther content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nTarget.\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{
			testChainItemQualified("ROOT/b", "code-from-spec/b/_node.md", "interface"),
		},
		External: []*frontmatter.FrontmatterExternal{},
		Target:   testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "code-from-spec/b/_node.md", "# ROOT/b\n\n# Public\n\n## Interface\n\nModified interface content.\n\n## Other\n\nOther content.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change when qualified subsection changes, got same hash: %q", hash1)
	}
}

func TestQualifierCaseNormalization(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/b/_node.md", "# ROOT/b\n\n# Public\n\n## Interface\n\nSome interface content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nTarget.\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{
			testChainItemQualified("ROOT/b", "code-from-spec/b/_node.md", "INTERFACE"),
		},
		External: []*frontmatter.FrontmatterExternal{},
		Target:   testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27 characters, got %d: %q", len(hash), hash)
	}
}

func TestArtifactDependencyHashesFileMinusFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "artifacts/some_artifact.md", "---\nsome: value\n---\n\nOriginal body content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nTarget.\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ARTIFACT/some_id", "artifacts/some_artifact.md"),
		},
		External: []*frontmatter.FrontmatterExternal{},
		Target:   testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "artifacts/some_artifact.md", "---\nsome: value\n---\n\nModified body content.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change when artifact body changes, got same hash: %q", hash1)
	}
}

func TestArtifactDependencyFrontmatterChangeIgnored(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "artifacts/some_artifact.md", "---\nsome: value\n---\n\nBody content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nTarget.\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ARTIFACT/some_id", "artifacts/some_artifact.md"),
		},
		External: []*frontmatter.FrontmatterExternal{},
		Target:   testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "artifacts/some_artifact.md", "---\nother: changed\n---\n\nBody content.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("expected hash to stay same when only artifact frontmatter changes, got %q and %q", hash1, hash2)
	}
}

func TestExternalFileHashesAllContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "external/file.txt", "Original external content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nTarget.\n")

	chain := &chainresolver.Chain{
		Ancestors:    []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{},
		External: []*frontmatter.FrontmatterExternal{
			{Path: "external/file.txt"},
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "external/file.txt", "Modified external content.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change when external file changes, got same hash: %q", hash1)
	}
}

func TestTargetPublicAndAgentBothContribute(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nPublic content.\n\n# Agent\n\nAgent content.\n")

	chain := &chainresolver.Chain{
		Ancestors:    []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{},
		External:     []*frontmatter.FrontmatterExternal{},
		Target:       testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nPublic content.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change when Agent section removed, got same hash: %q", hash1)
	}
}

func TestTargetWithoutAgentSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nPublic only content.\n")

	chain := &chainresolver.Chain{
		Ancestors:    []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{},
		External:     []*frontmatter.FrontmatterExternal{},
		Target:       testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27 characters, got %d: %q", len(hash), hash)
	}
}

func TestInputHashesFileMinusFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "artifacts/input_artifact.md", "---\nsome: value\n---\n\nOriginal input body.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nTarget.\n")

	chain := &chainresolver.Chain{
		Ancestors:    []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{},
		External:     []*frontmatter.FrontmatterExternal{},
		Target:       testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Input:        testChainItem("ARTIFACT/input_id", "artifacts/input_artifact.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "artifacts/input_artifact.md", "---\nsome: value\n---\n\nModified input body.\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change when input body changes, got same hash: %q", hash1)
	}
}

func TestNoInputSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nTarget.\n")

	chain := &chainresolver.Chain{
		Ancestors:    []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{},
		External:     []*frontmatter.FrontmatterExternal{},
		Target:       testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Input:        nil,
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27 characters, got %d: %q", len(hash), hash)
	}
}

func TestUnreadableSpecNodeFileReturnsParseFailure(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	chain := &chainresolver.Chain{
		Ancestors:    []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{},
		External:     []*frontmatter.FrontmatterExternal{},
		Target:       testChainItem("ROOT/missing", "code-from-spec/missing/_node.md"),
		Input:        nil,
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

	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nTarget.\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ARTIFACT/missing", "nonexistent/missing.md"),
		},
		External: []*frontmatter.FrontmatterExternal{},
		Target:   testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Input:    nil,
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainhash.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestUnreadableExternalFileReturnsFileUnreadable(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nTarget.\n")

	chain := &chainresolver.Chain{
		Ancestors:    []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{},
		External: []*frontmatter.FrontmatterExternal{
			{Path: "nonexistent/file.txt"},
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Input:  nil,
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainhash.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}
