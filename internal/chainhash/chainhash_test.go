// code-from-spec: ROOT/golang/tests/chain/hash@L3-hz1QNB4Y0inQPTXfULE5zXR4
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

// testChdir changes the working directory to dir for the duration of the test,
// restoring it via t.Cleanup.
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

// testWriteNodeFile creates a _node.md file at the given CFS path (relative to cwd).
// heading is the first-level heading (e.g. "ROOT/a"), sections is optional additional content.
func testWriteNodeFile(t *testing.T, cfsPath string, heading string, extraContent string) {
	t.Helper()
	content := fmt.Sprintf("# %s\n%s", heading, extraContent)
	dir := filepath.Dir(cfsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testWriteNodeFile mkdir: %v", err)
	}
	if err := os.WriteFile(cfsPath, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNodeFile write: %v", err)
	}
}

// testWriteFile writes arbitrary content to a file at the given path relative to cwd.
func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile write: %v", err)
	}
}

// testChainItem creates a ChainItem from a logical name and cfs path string.
func testChainItem(logicalName, cfsPath string, qualifier string) *chainresolver.ChainItem {
	return &chainresolver.ChainItem{
		LogicalName: logicalName,
		FilePath:    pathutils.PathCfs{Value: cfsPath},
		Qualifier:   qualifier,
	}
}

// testMinimalChain builds a Chain with only a target and all other fields empty/nil.
func testMinimalChain(target *chainresolver.ChainItem) *chainresolver.Chain {
	return &chainresolver.Chain{
		Ancestors:    nil,
		Dependencies: nil,
		External:     nil,
		Target:       target,
		Input:        nil,
	}
}

// --- Properties ---

func TestChainHashCompute_Deterministic(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "\n# Public\nSome public content\n")

	chain := testMinimalChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""))

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("hash not deterministic: %q != %q", hash1, hash2)
	}
}

func TestChainHashCompute_Is27Characters(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "\n# Public\nSome content\n")

	chain := testMinimalChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""))

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hash) != 27 {
		t.Errorf("expected 27 characters, got %d: %q", len(hash), hash)
	}
}

func TestChainHashCompute_ChangesWhenAncestorContentChanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", "ROOT", "\n# Public\nRoot public content\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md", ""),
		},
		Dependencies: nil,
		External:     nil,
		Target:       testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Input:        nil,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "code-from-spec/_node.md", "ROOT", "\n# Public\nModified root public content\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error after modify: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after ancestor content change")
	}
}

func TestChainHashCompute_ChangesWhenDependencyContentChanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", "ROOT", "")
	testWriteNodeFile(t, "code-from-spec/b/_node.md", "ROOT/b", "\n# Public\nDependency content\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "\n# Public\nTarget public\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md", ""),
		},
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ROOT/b", "code-from-spec/b/_node.md", ""),
		},
		External: nil,
		Target:   testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Input:    nil,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "code-from-spec/b/_node.md", "ROOT/b", "\n# Public\nModified dependency content\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error after modify: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after dependency content change")
	}
}

func TestChainHashCompute_ChangesWhenTargetPublicChanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", "ROOT", "")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "\n# Public\nTarget public\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md", ""),
		},
		Dependencies: nil,
		External:     nil,
		Target:       testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Input:        nil,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "\n# Public\nModified target public\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error after modify: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after target Public change")
	}
}

func TestChainHashCompute_ChangesWhenTargetAgentChanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", "ROOT", "")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "\n# Public\nPublic content\n\n# Agent\nAgent content\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md", ""),
		},
		Dependencies: nil,
		External:     nil,
		Target:       testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Input:        nil,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "\n# Public\nPublic content\n\n# Agent\nDifferent agent content\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error after modify: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after target Agent change")
	}
}

// --- Ancestors ---

func TestChainHashCompute_AncestorWithPublicContributesHash(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", "ROOT", "\n# Public\nAncestor public content\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md", ""),
		},
		Dependencies: nil,
		External:     nil,
		Target:       testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Input:        nil,
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash == "" {
		t.Error("expected non-empty hash")
	}
	if len(hash) != 27 {
		t.Errorf("expected 27 characters, got %d: %q", len(hash), hash)
	}
}

func TestChainHashCompute_AncestorWithoutPublicDifferentFromAncestorWithPublic(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	// ROOT has no Public section
	testWriteNodeFile(t, "code-from-spec/_node.md", "ROOT", "")
	// ROOT/a is the target with a Public section
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "\n# Public\nTarget public\n")
	// ROOT/c has a Public section
	testWriteNodeFile(t, "code-from-spec/c/_node.md", "ROOT/c", "\n# Public\nC public content\n")

	chainA := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md", ""),
		},
		Dependencies: nil,
		External:     nil,
		Target:       testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Input:        nil,
	}

	chainB := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT/c", "code-from-spec/c/_node.md", ""),
		},
		Dependencies: nil,
		External:     nil,
		Target:       testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Input:        nil,
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
		t.Error("expected hashes to differ: ancestor without Public vs ancestor with Public")
	}
}

func TestChainHashCompute_MultipleAncestorsOrderMatters(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/_node.md", "ROOT", "\n# Public\nRoot public\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "\n# Public\nA public\n")
	testWriteNodeFile(t, "code-from-spec/a/b/_node.md", "ROOT/a/b", "")

	chainA := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT", "code-from-spec/_node.md", ""),
			testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		},
		Dependencies: nil,
		External:     nil,
		Target:       testChainItem("ROOT/a/b", "code-from-spec/a/b/_node.md", ""),
		Input:        nil,
	}

	chainB := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
			testChainItem("ROOT", "code-from-spec/_node.md", ""),
		},
		Dependencies: nil,
		External:     nil,
		Target:       testChainItem("ROOT/a/b", "code-from-spec/a/b/_node.md", ""),
		Input:        nil,
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
		t.Error("expected hashes to differ when ancestor order is reversed")
	}
}

// --- Dependencies ---

func TestChainHashCompute_RootDepWithoutQualifierHashesPublic(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/b/_node.md", "ROOT/b", "\n# Public\nDep public\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "")

	chain := &chainresolver.Chain{
		Ancestors: nil,
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ROOT/b", "code-from-spec/b/_node.md", ""),
		},
		External: nil,
		Target:   testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Input:    nil,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "code-from-spec/b/_node.md", "ROOT/b", "\n# Public\nModified dep public\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error after modify: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after dependency Public change")
	}
}

func TestChainHashCompute_RootDepWithQualifierHashesSubsection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/b/_node.md", "ROOT/b", "\n# Public\n\n## Interface\nInterface content\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "")

	chain := &chainresolver.Chain{
		Ancestors: nil,
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ROOT/b", "code-from-spec/b/_node.md", "interface"),
		},
		External: nil,
		Target:   testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Input:    nil,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "code-from-spec/b/_node.md", "ROOT/b", "\n# Public\n\n## Interface\nModified interface\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error after modify: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after qualified subsection change")
	}
}

func TestChainHashCompute_QualifierCaseNormalization(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/b/_node.md", "ROOT/b", "\n# Public\n\n## Interface\nInterface content\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "")

	chain := &chainresolver.Chain{
		Ancestors: nil,
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ROOT/b", "code-from-spec/b/_node.md", "INTERFACE"),
		},
		External: nil,
		Target:   testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Input:    nil,
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hash) != 27 {
		t.Errorf("expected 27-character hash, got %d: %q", len(hash), hash)
	}
}

func TestChainHashCompute_ArtifactDepHashesBodyMinusFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "artifacts/dep.md", "---\noutputs:\n  - id: foo\n    path: foo.go\n---\nArtifact body\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "")

	chain := &chainresolver.Chain{
		Ancestors: nil,
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ARTIFACT/dep", "artifacts/dep.md", ""),
		},
		External: nil,
		Target:   testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Input:    nil,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "artifacts/dep.md", "---\noutputs:\n  - id: foo\n    path: foo.go\n---\nModified artifact body\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error after modify: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after artifact body change")
	}
}

func TestChainHashCompute_ArtifactDepFrontmatterChangeIgnored(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "artifacts/dep.md", "---\noutputs:\n  - id: foo\n    path: foo.go\n---\nStable body\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "")

	chain := &chainresolver.Chain{
		Ancestors: nil,
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ARTIFACT/dep", "artifacts/dep.md", ""),
		},
		External: nil,
		Target:   testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Input:    nil,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Modify only frontmatter, leave body unchanged
	testWriteFile(t, "artifacts/dep.md", "---\noutputs:\n  - id: foo\n    path: foo.go\nextra: field\n---\nStable body\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error after frontmatter modify: %v", err)
	}

	if hash1 != hash2 {
		t.Error("expected hashes to be equal after frontmatter-only change")
	}
}

// --- External files ---

func TestChainHashCompute_ExternalWholeFileHashesAllContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "external/file.txt", "External file content\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "")

	chain := &chainresolver.Chain{
		Ancestors:    nil,
		Dependencies: nil,
		External: []*frontmatter.FrontmatterExternal{
			{Path: "external/file.txt", Fragments: nil},
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Input:  nil,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "external/file.txt", "Modified external content\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error after modify: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after external file change")
	}
}

// testMakeLines creates a string with n lines, each containing "line N".
func testMakeLines(n int) string {
	var sb strings.Builder
	for i := 1; i <= n; i++ {
		sb.WriteString(fmt.Sprintf("line %d\n", i))
	}
	return sb.String()
}

// testMakeLinesWithModifiedLine creates a string with n lines, overriding one line.
func testMakeLinesWithModifiedLine(n, modLine int, modContent string) string {
	var sb strings.Builder
	for i := 1; i <= n; i++ {
		if i == modLine {
			sb.WriteString(modContent + "\n")
		} else {
			sb.WriteString(fmt.Sprintf("line %d\n", i))
		}
	}
	return sb.String()
}

func TestChainHashCompute_ExternalWithFragmentsHashesDeclaredRange(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "external/file.txt", testMakeLines(10))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "")

	chain := &chainresolver.Chain{
		Ancestors:    nil,
		Dependencies: nil,
		External: []*frontmatter.FrontmatterExternal{
			{
				Path: "external/file.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "3-5"},
				},
			},
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Input:  nil,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Modify line 4 (within range 3-5)
	testWriteFile(t, "external/file.txt", testMakeLinesWithModifiedLine(10, 4, "modified line 4"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error after modify: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after modifying line within fragment range")
	}
}

func TestChainHashCompute_ExternalWithFragmentsChangeOutsideRangeIgnored(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "external/file.txt", testMakeLines(10))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "")

	chain := &chainresolver.Chain{
		Ancestors:    nil,
		Dependencies: nil,
		External: []*frontmatter.FrontmatterExternal{
			{
				Path: "external/file.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "3-5"},
				},
			},
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Input:  nil,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Modify line 8 (outside range 3-5)
	testWriteFile(t, "external/file.txt", testMakeLinesWithModifiedLine(10, 8, "modified line 8"))

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error after modify: %v", err)
	}

	if hash1 != hash2 {
		t.Error("expected hashes to be equal after modifying line outside fragment range")
	}
}

func TestChainHashCompute_ExternalMultipleFragmentsOrderMatters(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "external/file.txt", testMakeLines(10))
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "")

	chainA := &chainresolver.Chain{
		Ancestors:    nil,
		Dependencies: nil,
		External: []*frontmatter.FrontmatterExternal{
			{
				Path: "external/file.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "6-8"},
					{Lines: "1-3"},
				},
			},
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Input:  nil,
	}

	chainB := &chainresolver.Chain{
		Ancestors:    nil,
		Dependencies: nil,
		External: []*frontmatter.FrontmatterExternal{
			{
				Path: "external/file.txt",
				Fragments: []*frontmatter.FrontmatterExternalFragment{
					{Lines: "1-3"},
					{Lines: "6-8"},
				},
			},
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Input:  nil,
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
		t.Error("expected hashes to differ when fragment order changes")
	}
}

// --- Target ---

func TestChainHashCompute_TargetPublicAndAgentBothContribute(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "\n# Public\nPublic content\n\n# Agent\nAgent content\n")

	chain := testMinimalChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""))

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Remove Agent section
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "\n# Public\nPublic content\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error after removing Agent: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ when Agent section is removed")
	}
}

func TestChainHashCompute_TargetWithoutAgentIsValid(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "\n# Public\nPublic only content\n")

	chain := testMinimalChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""))

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hash) != 27 {
		t.Errorf("expected 27-character hash, got %d: %q", len(hash), hash)
	}
}

// --- Input ---

func TestChainHashCompute_InputHashesBodyMinusFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "artifacts/input.md", "---\noutputs:\n  - id: in\n    path: in.go\n---\nInput body content\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "")

	chain := &chainresolver.Chain{
		Ancestors:    nil,
		Dependencies: nil,
		External:     nil,
		Target:       testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Input:        testChainItem("ARTIFACT/input", "artifacts/input.md", ""),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "artifacts/input.md", "---\noutputs:\n  - id: in\n    path: in.go\n---\nModified input body\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error after modify: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ after input body change")
	}
}

func TestChainHashCompute_NoInputSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "")

	chain := testMinimalChain(testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""))

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hash) != 27 {
		t.Errorf("expected 27-character hash, got %d: %q", len(hash), hash)
	}
}

// --- Error cases ---

func TestChainHashCompute_UnreadableSpecNodeFile(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	chain := testMinimalChain(testChainItem("ROOT/nonexistent", "code-from-spec/nonexistent/_node.md", ""))

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for missing spec node file, got nil")
	}
	if !errors.Is(err, chainhash.ErrParseFailure) {
		t.Errorf("expected ErrParseFailure, got: %v", err)
	}
}

func TestChainHashCompute_UnreadableArtifactFile(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "")

	chain := &chainresolver.Chain{
		Ancestors: nil,
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ARTIFACT/missing", "artifacts/missing.md", ""),
		},
		External: nil,
		Target:   testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Input:    nil,
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for missing artifact file, got nil")
	}
	if !errors.Is(err, chainhash.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestChainHashCompute_UnreadableExternalFile(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", "ROOT/a", "")

	chain := &chainresolver.Chain{
		Ancestors:    nil,
		Dependencies: nil,
		External: []*frontmatter.FrontmatterExternal{
			{Path: "external/missing.txt", Fragments: nil},
		},
		Target: testChainItem("ROOT/a", "code-from-spec/a/_node.md", ""),
		Input:  nil,
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for missing external file, got nil")
	}
	if !errors.Is(err, chainhash.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}
