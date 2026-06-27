// code-from-spec: SPEC/golang/tests/chain/hash@RSaKM1-BNUFCklJCHMSr16_XGes
package chainhash_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
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
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return "."
}

func testChainItem(logicalName string, filePath string) *chainresolver.ChainItem {
	return &chainresolver.ChainItem{
		UnqualifiedLogicalName: logicalName,
		FilePath:               pathutils.PathCfs{Value: filePath},
	}
}

func testChainItemWithQualifier(logicalName string, filePath string, qualifier string) *chainresolver.ChainItem {
	q := qualifier
	return &chainresolver.ChainItem{
		UnqualifiedLogicalName: logicalName,
		FilePath:               pathutils.PathCfs{Value: filePath},
		Qualifier:              &q,
	}
}

func TestChainHashCompute_Deterministic(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface content\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
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
		t.Errorf("expected deterministic hashes, got %q and %q", hash1, hash2)
	}
}

func TestChainHashCompute_Is27Characters(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface content\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hash) != 27 {
		t.Errorf("expected 27 characters, got %d: %q", len(hash), hash)
	}
}

func TestChainHashCompute_HashChangesWhenAncestorContentChanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n\n# Public\n\n## Context\n\ninitial context\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface content\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("SPEC", "code-from-spec/_node.md"),
		},
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n\n# Public\n\n## Context\n\nmodified context\n")

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change when ancestor content changes")
	}
}

func TestChainHashCompute_HashChangesWhenDependencyContentChanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n\n# Public\n\n## Context\n\nroot context\n")
	testWriteFile(t, "code-from-spec/b/_node.md", "# SPEC/b\n\n# Public\n\n## Interface\n\ninitial b interface\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("SPEC/b", "code-from-spec/b/_node.md"),
		},
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "code-from-spec/b/_node.md", "# SPEC/b\n\n# Public\n\n## Interface\n\nmodified b interface\n")

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change when dependency content changes")
	}
}

func TestChainHashCompute_HashChangesWhenTargetPublicChanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n\n# Public\n\n## Context\n\nroot context\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\ninitial interface\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nmodified interface\n")

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change when target Public changes")
	}
}

func TestChainHashCompute_HashChangesWhenTargetAgentChanges(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n\n# Agent\n\ninitial agent content\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n\n# Agent\n\nmodified agent content\n")

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change when target Agent changes")
	}
}

func TestChainHashCompute_AncestorWithPublicSubsectionsContributesHash(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n\n# Public\n\n## Context\n\nroot context\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("SPEC", "code-from-spec/_node.md"),
		},
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hash) != 27 {
		t.Errorf("expected 27 characters, got %d: %q", len(hash), hash)
	}
}

func TestChainHashCompute_AncestorWithoutPublicSection_Skipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n\n# Public\n\n## Context\n\nroot context\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("SPEC", "code-from-spec/_node.md"),
		},
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}

	hashWithPublic, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n")

	hashWithoutPublic, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hashWithPublic == hashWithoutPublic {
		t.Error("expected hash to change when ancestor Public section is removed")
	}
}

func TestChainHashCompute_MultipleAncestors_OrderMatters(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n\n# Public\n\n## Context\n\nroot context\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Context\n\na context\n")
	testWriteFile(t, "code-from-spec/a/b/_node.md", "# SPEC/a/b\n\n# Public\n\n## Interface\n\nsome interface\n")

	chainA := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("SPEC", "code-from-spec/_node.md"),
			testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
		},
		Target: testChainItem("SPEC/a/b", "code-from-spec/a/b/_node.md"),
	}

	chainB := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
			testChainItem("SPEC", "code-from-spec/_node.md"),
		},
		Target: testChainItem("SPEC/a/b", "code-from-spec/a/b/_node.md"),
	}

	hashA, err := chainhash.ChainHashCompute(chainA)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hashB, err := chainhash.ChainHashCompute(chainB)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hashA == hashB {
		t.Error("expected different hashes for different ancestor orders")
	}
}

func TestChainHashCompute_SpecDependencyWithoutQualifier_HashesPublicSubsections(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/b/_node.md", "# SPEC/b\n\n# Public\n\n## Interface\n\ninitial b interface\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("SPEC/b", "code-from-spec/b/_node.md"),
		},
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "code-from-spec/b/_node.md", "# SPEC/b\n\n# Public\n\n## Interface\n\nmodified b interface\n")

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change when SPEC dependency content changes")
	}
}

func TestChainHashCompute_SpecDependencyWithQualifier_HashesSubsection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/b/_node.md", "# SPEC/b\n\n# Public\n\n## Interface\n\ninitial b interface\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItemWithQualifier("SPEC/b", "code-from-spec/b/_node.md", "interface"),
		},
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "code-from-spec/b/_node.md", "# SPEC/b\n\n# Public\n\n## Interface\n\nmodified b interface\n")

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change when qualified dependency subsection content changes")
	}
}

func TestChainHashCompute_QualifierCaseNormalization(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/b/_node.md", "# SPEC/b\n\n# Public\n\n## Interface\n\nsome interface\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItemWithQualifier("SPEC/b", "code-from-spec/b/_node.md", "INTERFACE"),
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("expected no error with uppercase qualifier, got: %v", err)
	}
}

func TestChainHashCompute_ArtifactDependency_HashesFullFileContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")
	testWriteFile(t, "out/artifact.txt", "initial artifact content\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ARTIFACT/out", "out/artifact.txt"),
		},
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "out/artifact.txt", "modified artifact content\n")

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change when artifact file content changes")
	}
}

func TestChainHashCompute_ArtifactDependency_TagHashChangeIgnored(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")
	testWriteFile(t, "out/artifact.txt", "// code-from-spec: SPEC/a@C59GaYbNt2Nw-HgaePlvyK1XUMU\n\nsome content\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ARTIFACT/out", "out/artifact.txt"),
		},
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "out/artifact.txt", "// code-from-spec: SPEC/a@zZyYxXwWvVuUtTsSrRqQpPoOnNm\n\nsome content\n")

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hashBefore != hashAfter {
		t.Error("expected hash to remain the same when only artifact tag hash changes")
	}
}

func TestChainHashCompute_ExternalDependency_HashesAllContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")
	testWriteFile(t, "somefile.proto", "initial external content\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("EXTERNAL/somefile", "somefile.proto"),
		},
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "somefile.proto", "modified external content\n")

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change when external file content changes")
	}
}

func TestChainHashCompute_LeadingBlankLinesRemovedFromSubsection(t *testing.T) {
	tmpA := t.TempDir()
	tmpB := t.TempDir()

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	if err := os.MkdirAll(tmpA+"/code-from-spec/a", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tmpA+"/code-from-spec/a/_node.md", []byte("# SPEC/a\n\n# Public\n\n## Interface\n\n\ncontent line\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(tmpB+"/code-from-spec/a", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tmpB+"/code-from-spec/a/_node.md", []byte("# SPEC/a\n\n# Public\n\n## Interface\n\ncontent line\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tmpA); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	chainA := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}
	hashA, err := chainhash.ChainHashCompute(chainA)
	if err != nil {
		t.Fatalf("unexpected error for chain A: %v", err)
	}

	if err := os.Chdir(tmpB); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	chainB := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}
	hashB, err := chainhash.ChainHashCompute(chainB)
	if err != nil {
		t.Fatalf("unexpected error for chain B: %v", err)
	}

	if hashA != hashB {
		t.Errorf("expected equal hashes when leading blanks differ, got %q and %q", hashA, hashB)
	}
}

func TestChainHashCompute_TrailingBlankLinesRemovedFromSubsection(t *testing.T) {
	tmpA := t.TempDir()
	tmpB := t.TempDir()

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	if err := os.MkdirAll(tmpA+"/code-from-spec/a", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tmpA+"/code-from-spec/a/_node.md", []byte("# SPEC/a\n\n# Public\n\n## Interface\n\ncontent line\n\n\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(tmpB+"/code-from-spec/a", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tmpB+"/code-from-spec/a/_node.md", []byte("# SPEC/a\n\n# Public\n\n## Interface\n\ncontent line\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tmpA); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	chainA := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}
	hashA, err := chainhash.ChainHashCompute(chainA)
	if err != nil {
		t.Fatalf("unexpected error for chain A: %v", err)
	}

	if err := os.Chdir(tmpB); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	chainB := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}
	hashB, err := chainhash.ChainHashCompute(chainB)
	if err != nil {
		t.Fatalf("unexpected error for chain B: %v", err)
	}

	if hashA != hashB {
		t.Errorf("expected equal hashes when trailing blanks differ, got %q and %q", hashA, hashB)
	}
}

func TestChainHashCompute_InteriorBlankLinesPreserved(t *testing.T) {
	tmpA := t.TempDir()
	tmpB := t.TempDir()

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	if err := os.MkdirAll(tmpA+"/code-from-spec/a", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tmpA+"/code-from-spec/a/_node.md", []byte("# SPEC/a\n\n# Public\n\n## Interface\n\nline one\n\nline two\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(tmpB+"/code-from-spec/a", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tmpB+"/code-from-spec/a/_node.md", []byte("# SPEC/a\n\n# Public\n\n## Interface\n\nline one\nline two\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tmpA); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	chainA := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}
	hashA, err := chainhash.ChainHashCompute(chainA)
	if err != nil {
		t.Fatalf("unexpected error for chain A: %v", err)
	}

	if err := os.Chdir(tmpB); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	chainB := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}
	hashB, err := chainhash.ChainHashCompute(chainB)
	if err != nil {
		t.Fatalf("unexpected error for chain B: %v", err)
	}

	if hashA == hashB {
		t.Error("expected different hashes when interior blank lines differ")
	}
}

func TestChainHashCompute_TargetPublicAndAgentBothContribute(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n\n# Agent\n\nagent instructions\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change when Agent section is removed")
	}
}

func TestChainHashCompute_TargetWithoutAgent_AgentSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hash) != 27 {
		t.Errorf("expected 27 characters, got %d: %q", len(hash), hash)
	}
}

func TestChainHashCompute_InputHashesFullFileContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")
	testWriteFile(t, "input/artifact.txt", "initial input content\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
		Input:  testChainItem("ARTIFACT/input", "input/artifact.txt"),
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "input/artifact.txt", "modified input content\n")

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change when input file content changes")
	}
}

func TestChainHashCompute_NoInput_Skipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hash) != 27 {
		t.Errorf("expected 27 characters, got %d: %q", len(hash), hash)
	}
}

func TestChainHashCompute_UnreadableSpecNodeFile(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for non-existent spec file")
	}

	if !errors.Is(err, chainhash.ErrParseFailure) {
		t.Errorf("expected ErrParseFailure, got: %v", err)
	}
}

func TestChainHashCompute_UnreadableArtifactFile(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ARTIFACT/out", "nonexistent/artifact.txt"),
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for non-existent artifact file")
	}

	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestChainHashCompute_UnreadableExternalFile(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("EXTERNAL/somefile", "nonexistent/somefile.proto"),
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for non-existent external file")
	}

	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}
