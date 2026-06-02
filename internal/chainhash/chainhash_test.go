// code-from-spec: ROOT/golang/tests/chain/hash@oPNLRtwdybUtqjGbeFK7uxM-1Bo
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

func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath(path), 0755); err != nil {
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

func filepath(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' {
			return p[:i]
		}
	}
	return "."
}

func testMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("testMkdirAll: %v", err)
	}
}

func testRootNode(t *testing.T) {
	t.Helper()
	testMkdirAll(t, "code-from-spec")
	testWriteNode(t, "code-from-spec/_node.md", "# ROOT\n")
}

func testWriteNode(t *testing.T, path string, content string) {
	t.Helper()
	dir := filepath(path)
	if dir != "." {
		testMkdirAll(t, dir)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNode: %v", err)
	}
}

func testChainItem(logicalName, filePath string) *chainresolver.ChainItem {
	return &chainresolver.ChainItem{
		LogicalName: logicalName,
		FilePath:    pathutils.PathCfs{Value: filePath},
	}
}

func testChainItemWithQualifier(logicalName, filePath, qualifier string) *chainresolver.ChainItem {
	return &chainresolver.ChainItem{
		LogicalName: logicalName,
		FilePath:    pathutils.PathCfs{Value: filePath},
		Qualifier:   qualifier,
	}
}

func TestHashIsDeterministic(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n\nsome content\n")
	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md"),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
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

	testWriteNode(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n\nsome content\n")
	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\ntarget content\n")

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
		t.Errorf("expected 27-character hash, got %d characters: %q", len(hash), hash)
	}
}

func TestHashChangesWhenAncestorContentChanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n\noriginal content\n")
	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md"),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNode(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n\nmodified content\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after ancestor content change")
	}
}

func TestHashChangesWhenDependencyContentChanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteNode(t, "code-from-spec/b/_node.md", "# ROOT/b\n\n# Public\n\noriginal b content\n")
	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md"),
		},
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ROOT/b", "code-from-spec/b/_node.md"),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNode(t, "code-from-spec/b/_node.md", "# ROOT/b\n\n# Public\n\nmodified b content\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after dependency content change")
	}
}

func TestHashChangesWhenTargetPublicChanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\noriginal public\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\nmodified public\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after target Public change")
	}
}

func TestHashChangesWhenTargetAgentChanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\npublic content\n\n# Agent\n\noriginal agent\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\npublic content\n\n# Agent\n\nmodified agent\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after target Agent change")
	}
}

func TestAncestorWithPublicContributesHash(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n\nsome ancestor content\n")
	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n")

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
		t.Errorf("expected 27-character hash, got %d: %q", len(hash), hash)
	}
}

func TestAncestorWithoutPublicSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, "code-from-spec/_node.md", "# ROOT\n")
	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\ntarget content\n")
	testWriteNode(t, "code-from-spec/c/_node.md", "# ROOT/c\n\n# Public\n\nc public content\n")
	testWriteNode(t, "code-from-spec/a2/_node.md", "# ROOT/a2\n\n# Public\n\ntarget content\n")

	chainA := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md"),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	chainB := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT/c", "code-from-spec/c/_node.md"),
		},
		Target: testChainItem("ROOT/a2", "code-from-spec/a2/_node.md"),
	}

	hashA, err := chainhash.ChainHashCompute(chainA)
	if err != nil {
		t.Fatalf("unexpected error for chain A: %v", err)
	}
	hashB, err := chainhash.ChainHashCompute(chainB)
	if err != nil {
		t.Fatalf("unexpected error for chain B: %v", err)
	}

	if hashA == hashB {
		t.Error("expected hashes to differ: ROOT without Public vs ROOT/c with Public")
	}
}

func TestMultipleAncestorsOrderMatters(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, "code-from-spec/_node.md", "# ROOT\n\n# Public\n\nroot content\n")
	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\na content\n")
	testWriteNode(t, "code-from-spec/a/b/_node.md", "# ROOT/a/b\n")

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
		t.Fatalf("unexpected error for forward chain: %v", err)
	}
	hashReversed, err := chainhash.ChainHashCompute(chainReversed)
	if err != nil {
		t.Fatalf("unexpected error for reversed chain: %v", err)
	}

	if hashForward == hashReversed {
		t.Error("expected hashes to differ when ancestor order is swapped")
	}
}

func TestRootDependencyNoQualifierHashesPublic(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, "code-from-spec/b/_node.md", "# ROOT/b\n\n# Public\n\noriginal b public\n")
	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n")

	chain := &chainresolver.Chain{
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ROOT/b", "code-from-spec/b/_node.md"),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNode(t, "code-from-spec/b/_node.md", "# ROOT/b\n\n# Public\n\nmodified b public\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after ROOT dependency Public change")
	}
}

func TestRootDependencyWithQualifierHashesSubsection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, "code-from-spec/b/_node.md", "# ROOT/b\n\n# Public\n\n## Interface\n\noriginal interface content\n")
	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n")

	chain := &chainresolver.Chain{
		Dependencies: []*chainresolver.ChainItem{
			testChainItemWithQualifier("ROOT/b", "code-from-spec/b/_node.md", "interface"),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNode(t, "code-from-spec/b/_node.md", "# ROOT/b\n\n# Public\n\n## Interface\n\nmodified interface content\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after qualified subsection change")
	}
}

func TestQualifierCaseNormalization(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, "code-from-spec/b/_node.md", "# ROOT/b\n\n# Public\n\n## Interface\n\ninterface content\n")
	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n")

	chain := &chainresolver.Chain{
		Dependencies: []*chainresolver.ChainItem{
			testChainItemWithQualifier("ROOT/b", "code-from-spec/b/_node.md", "INTERFACE"),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Errorf("expected no error with uppercase qualifier, got: %v", err)
	}
}

func TestArtifactDependencyHashesFileMinusFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	artifactPath := "artifacts/dep.md"
	testMkdirAll(t, "artifacts")
	if err := os.WriteFile(artifactPath, []byte("---\noutput: some/path\n---\n\noriginal body\n"), 0644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n")

	chain := &chainresolver.Chain{
		Dependencies: []*chainresolver.ChainItem{
			testChainItemWithQualifier("ARTIFACT/x", artifactPath, ""),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := os.WriteFile(artifactPath, []byte("---\noutput: some/path\n---\n\nmodified body\n"), 0644); err != nil {
		t.Fatalf("write artifact: %v", err)
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
	tmp := t.TempDir()
	testChdir(t, tmp)

	artifactPath := "artifacts/dep2.md"
	testMkdirAll(t, "artifacts")
	if err := os.WriteFile(artifactPath, []byte("---\noutput: some/path\n---\n\nstable body\n"), 0644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n")

	chain := &chainresolver.Chain{
		Dependencies: []*chainresolver.ChainItem{
			testChainItemWithQualifier("ARTIFACT/x", artifactPath, ""),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := os.WriteFile(artifactPath, []byte("---\noutput: different/path\n---\n\nstable body\n"), 0644); err != nil {
		t.Fatalf("write artifact: %v", err)
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
	tmp := t.TempDir()
	testChdir(t, tmp)

	artifactPath := "artifacts/dep3.go"
	testMkdirAll(t, "artifacts")
	body1 := "---\noutput: some/path\n---\n// code-from-spec: ROOT/x/y@aAbBcCdDeEfFgGhHiIjJkKlLmMn\n\nsome code\n"
	if err := os.WriteFile(artifactPath, []byte(body1), 0644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n")

	chain := &chainresolver.Chain{
		Dependencies: []*chainresolver.ChainItem{
			testChainItemWithQualifier("ARTIFACT/x", artifactPath, ""),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body2 := "---\noutput: some/path\n---\n// code-from-spec: ROOT/x/y@zZyYxXwWvVuUtTsSrRqQpPoOnNm\n\nsome code\n"
	if err := os.WriteFile(artifactPath, []byte(body2), 0644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 != hash2 {
		t.Error("expected hashes to be identical after artifact tag hash change only")
	}
}

func TestExternalFileHashesAllContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	extPath := "external/data.txt"
	testMkdirAll(t, "external")
	if err := os.WriteFile(extPath, []byte("original external content\n"), 0644); err != nil {
		t.Fatalf("write external file: %v", err)
	}

	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n")

	chain := &chainresolver.Chain{
		External: []*frontmatter.FrontmatterExternal{
			{Path: extPath},
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := os.WriteFile(extPath, []byte("modified external content\n"), 0644); err != nil {
		t.Fatalf("write external file: %v", err)
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
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\npublic content\n\n# Agent\n\nagent content\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\npublic content\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after removing Agent section")
	}
}

func TestTargetWithoutAgentSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\npublic only\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27-character hash, got %d: %q", len(hash), hash)
	}
}

func TestInputHashesFileMinusFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	inputPath := "inputs/input.md"
	testMkdirAll(t, "inputs")
	if err := os.WriteFile(inputPath, []byte("---\noutput: some/path\n---\n\noriginal input body\n"), 0644); err != nil {
		t.Fatalf("write input file: %v", err)
	}

	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Input:  testChainItemWithQualifier("ARTIFACT/x", inputPath, ""),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := os.WriteFile(inputPath, []byte("---\noutput: some/path\n---\n\nmodified input body\n"), 0644); err != nil {
		t.Fatalf("write input file: %v", err)
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
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n\n# Public\n\npublic content\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Input:  nil,
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27-character hash, got %d: %q", len(hash), hash)
	}
}

func TestUnreadableSpecNodeFile(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for unreadable spec node file, got nil")
	}
	if !errors.Is(err, chainhash.ErrParseFailure) {
		t.Errorf("expected ErrParseFailure, got: %v", err)
	}
}

func TestUnreadableArtifactFile(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n")

	chain := &chainresolver.Chain{
		Dependencies: []*chainresolver.ChainItem{
			testChainItemWithQualifier("ARTIFACT/x", "nonexistent/artifact.md", ""),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for unreadable artifact file, got nil")
	}
	if !errors.Is(err, chainhash.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestUnreadableExternalFile(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNode(t, "code-from-spec/a/_node.md", "# ROOT/a\n")

	chain := &chainresolver.Chain{
		External: []*frontmatter.FrontmatterExternal{
			{Path: "nonexistent/external.txt"},
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for unreadable external file, got nil")
	}
	if !errors.Is(err, chainhash.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}
