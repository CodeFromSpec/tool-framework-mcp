// code-from-spec: ROOT/golang/tests/chain/hash@hqjE0Rx8n-Hj1NV5_qy3PQI4Swo
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

func testWriteNodeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath_dir(path), 0755); err != nil {
		t.Fatalf("testWriteNodeFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNodeFile write: %v", err)
	}
}

func filepath_dir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return "."
}

func testNodeContent(logicalName string, extraSections string) string {
	return fmt.Sprintf("# %s\n%s", logicalName, extraSections)
}

func testChainItem(logicalName string, filePath string) *chainresolver.ChainItem {
	return &chainresolver.ChainItem{
		LogicalName: logicalName,
		FilePath:    pathutils.PathCfs{Value: filePath},
	}
}

func testChainItemWithQualifier(logicalName string, filePath string, qualifier string) *chainresolver.ChainItem {
	return &chainresolver.ChainItem{
		LogicalName: logicalName,
		FilePath:    pathutils.PathCfs{Value: filePath},
		Qualifier:   qualifier,
	}
}

func TestChainHashCompute_Deterministic(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/_node.md", testNodeContent("ROOT", "# Public\nsome content\n"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\ntarget content\n"))

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

func TestChainHashCompute_Is27Characters(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/_node.md", testNodeContent("ROOT", "# Public\nsome content\n"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\ntarget content\n"))

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

func TestChainHashCompute_ChangesWhenAncestorContentChanges(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/_node.md", testNodeContent("ROOT", "# Public\noriginal content\n"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\ntarget content\n"))

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

	testWriteNodeFile(t, "code-from-spec/_node.md", testNodeContent("ROOT", "# Public\nmodified content\n"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change when ancestor content changes, but got same hash: %q", hash1)
	}
}

func TestChainHashCompute_ChangesWhenDependencyContentChanges(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/_node.md", testNodeContent("ROOT", ""))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\ntarget\n"))
	testWriteNodeFile(t, "code-from-spec/b/_node.md", testNodeContent("ROOT/b", "# Public\ndep content\n"))

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ROOT/b", "code-from-spec/b/_node.md"),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "code-from-spec/b/_node.md", testNodeContent("ROOT/b", "# Public\nmodified dep content\n"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change when dependency content changes, but got same hash: %q", hash1)
	}
}

func TestChainHashCompute_ChangesWhenTargetPublicChanges(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\noriginal public\n"))

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\nmodified public\n"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change when target Public changes, but got same hash: %q", hash1)
	}
}

func TestChainHashCompute_ChangesWhenTargetAgentChanges(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\nsome public\n# Agent\noriginal agent\n"))

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\nsome public\n# Agent\nmodified agent\n"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change when target Agent changes, but got same hash: %q", hash1)
	}
}

func TestChainHashCompute_AncestorWithPublicContributesHash(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/_node.md", testNodeContent("ROOT", "# Public\nsome public content\n"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\ntarget content\n"))

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

func TestChainHashCompute_AncestorWithoutPublicSkipped(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/_node.md", testNodeContent("ROOT", ""))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\ntarget content\n"))
	testWriteNodeFile(t, "code-from-spec/z/_node.md", testNodeContent("ROOT/z", "# Public\nsome z content\n"))

	chainA := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md"),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	chainB := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT/z", "code-from-spec/z/_node.md"),
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
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
		t.Errorf("expected hashA and hashB to differ when one ancestor has no Public section, but got same hash: %q", hashA)
	}
}

func TestChainHashCompute_MultipleAncestorsOrderMatters(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/_node.md", testNodeContent("ROOT", "# Public\nroot public\n"))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\na public\n"))
	testWriteNodeFile(t, "code-from-spec/a/b/_node.md", testNodeContent("ROOT/a/b", "# Public\nab target\n"))

	chainX := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md"),
			testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		},
		Target: testChainItem("ROOT/a/b", "code-from-spec/a/b/_node.md"),
	}

	chainY := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
			testChainItem("ROOT", "code-from-spec/_node.md"),
		},
		Target: testChainItem("ROOT/a/b", "code-from-spec/a/b/_node.md"),
	}

	hashX, err := chainhash.ChainHashCompute(chainX)
	if err != nil {
		t.Fatalf("unexpected error for chainX: %v", err)
	}

	hashY, err := chainhash.ChainHashCompute(chainY)
	if err != nil {
		t.Fatalf("unexpected error for chainY: %v", err)
	}

	if hashX == hashY {
		t.Errorf("expected hashX and hashY to differ when ancestor order differs, but got same hash: %q", hashX)
	}
}

func TestChainHashCompute_RootDependencyWithoutQualifierHashesPublic(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\ntarget\n"))
	testWriteNodeFile(t, "code-from-spec/b/_node.md", testNodeContent("ROOT/b", "# Public\ndep public\n"))

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ROOT/b", "code-from-spec/b/_node.md"),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "code-from-spec/b/_node.md", testNodeContent("ROOT/b", "# Public\nmodified dep public\n"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change when dependency Public changes, but got same hash: %q", hash1)
	}
}

func TestChainHashCompute_RootDependencyWithQualifierHashesSubsection(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\ntarget\n"))
	testWriteNodeFile(t, "code-from-spec/b/_node.md", testNodeContent("ROOT/b", "# Public\n## Interface\ninterface content\n"))

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItemWithQualifier("ROOT/b", "code-from-spec/b/_node.md", "interface"),
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "code-from-spec/b/_node.md", testNodeContent("ROOT/b", "# Public\n## Interface\nmodified interface content\n"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change when qualified subsection changes, but got same hash: %q", hash1)
	}
}

func TestChainHashCompute_QualifierCaseNormalization(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\ntarget\n"))
	testWriteNodeFile(t, "code-from-spec/b/_node.md", testNodeContent("ROOT/b", "# Public\n## Interface\ninterface content\n"))

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItemWithQualifier("ROOT/b", "code-from-spec/b/_node.md", "INTERFACE"),
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error with uppercase qualifier: %v", err)
	}
}

func TestChainHashCompute_ArtifactDependencyHashesBodyNotFrontmatter(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\ntarget\n"))

	artifactContent := "---\noutput: some/path.go\n---\nbody content line one\nbody content line two\n"
	if err := os.WriteFile("artifact.md", []byte(artifactContent), 0644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItemWithQualifier("ARTIFACT/a", "artifact.md", ""),
		},
	}
	chain.Dependencies[0].LogicalName = "ARTIFACT/a"

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	artifactContent2 := "---\noutput: some/path.go\n---\nbody content line one\nmodified body content\n"
	if err := os.WriteFile("artifact.md", []byte(artifactContent2), 0644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change when artifact body changes, but got same hash: %q", hash1)
	}
}

func TestChainHashCompute_ArtifactDependencyFrontmatterChangeIgnored(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\ntarget\n"))

	artifactContent := "---\noutput: some/path.go\n---\nbody content\n"
	if err := os.WriteFile("artifact.md", []byte(artifactContent), 0644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			{
				LogicalName: "ARTIFACT/a",
				FilePath:    pathutils.PathCfs{Value: "artifact.md"},
			},
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	artifactContent2 := "---\noutput: different/path.go\nextra: field\n---\nbody content\n"
	if err := os.WriteFile("artifact.md", []byte(artifactContent2), 0644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("expected hash to remain same when only frontmatter changes, got %q and %q", hash1, hash2)
	}
}

func TestChainHashCompute_ArtifactDependencyTagHashChangeIgnored(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\ntarget\n"))

	artifactContent := "---\noutput: some/path.go\n---\n// code-from-spec: ROOT/x/y@aAbBcCdDeEfFgGhHiIjJkKlLmMn\nbody content\n"
	if err := os.WriteFile("artifact.md", []byte(artifactContent), 0644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			{
				LogicalName: "ARTIFACT/a",
				FilePath:    pathutils.PathCfs{Value: "artifact.md"},
			},
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	artifactContent2 := "---\noutput: some/path.go\n---\n// code-from-spec: ROOT/x/y@zZyYxXwWvVuUtTsSrRqQpPoOnNm\nbody content\n"
	if err := os.WriteFile("artifact.md", []byte(artifactContent2), 0644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("expected hash to remain same when only artifact tag hash changes, got %q and %q", hash1, hash2)
	}
}

func TestChainHashCompute_ExternalFileHashesAllContent(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\ntarget\n"))

	if err := os.WriteFile("external.txt", []byte("external content\n"), 0644); err != nil {
		t.Fatalf("write external: %v", err)
	}

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		External: []*frontmatter.FrontmatterExternal{
			{Path: "external.txt"},
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := os.WriteFile("external.txt", []byte("modified external content\n"), 0644); err != nil {
		t.Fatalf("write external: %v", err)
	}

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change when external file changes, but got same hash: %q", hash1)
	}
}

func TestChainHashCompute_TargetPublicAndAgentBothContribute(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\npublic content\n# Agent\nagent content\n"))

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\npublic content\n"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change when Agent section removed, but got same hash: %q", hash1)
	}
}

func TestChainHashCompute_TargetWithoutAgentSucceeds(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\npublic only\n"))

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChainHashCompute_InputHashesBodyNotFrontmatter(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\ntarget\n"))

	inputContent := "---\noutput: some/path.go\n---\ninput body content\n"
	if err := os.WriteFile("input.md", []byte(inputContent), 0644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Input: &chainresolver.ChainItem{
			LogicalName: "ARTIFACT/a",
			FilePath:    pathutils.PathCfs{Value: "input.md"},
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	inputContent2 := "---\noutput: some/path.go\n---\nmodified input body content\n"
	if err := os.WriteFile("input.md", []byte(inputContent2), 0644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("expected hash to change when input body changes, but got same hash: %q", hash1)
	}
}

func TestChainHashCompute_NoInputSkipped(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\ntarget\n"))

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Input:  nil,
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChainHashCompute_UnreadableSpecNodeFileReturnsParseFailure(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/nonexistent", "code-from-spec/nonexistent/_node.md"),
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for nonexistent spec node file, got nil")
	}

	if !errors.Is(err, chainhash.ErrParseFailure) {
		t.Errorf("expected ErrParseFailure, got: %v", err)
	}
}

func TestChainHashCompute_UnreadableArtifactFileReturnsFileUnreadable(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\ntarget\n"))

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			{
				LogicalName: "ARTIFACT/nonexistent",
				FilePath:    pathutils.PathCfs{Value: "nonexistent/artifact.md"},
			},
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for nonexistent artifact file, got nil")
	}

	if !errors.Is(err, chainhash.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestChainHashCompute_UnreadableExternalFileReturnsFileUnreadable(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", testNodeContent("ROOT/a", "# Public\ntarget\n"))

	chain := &chainresolver.Chain{
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md"),
		External: []*frontmatter.FrontmatterExternal{
			{Path: "nonexistent/external.txt"},
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for nonexistent external file, got nil")
	}

	if !errors.Is(err, chainhash.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}
