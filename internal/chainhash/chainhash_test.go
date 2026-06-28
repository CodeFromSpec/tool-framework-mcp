// code-from-spec: SPEC/golang/tests/chain/hash@kLviuEI2jKXjGzh-x3K0JoEguvI
package chainhash_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/file"
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
	if err := os.MkdirAll(fileDir(path), 0755); err != nil {
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

func fileDir(path string) string {
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

func testChainItemQualified(logicalName string, filePath string, qualifier string) *chainresolver.ChainItem {
	return &chainresolver.ChainItem{
		UnqualifiedLogicalName: logicalName,
		FilePath:               pathutils.PathCfs{Value: filePath},
		Qualifier:              qualifier,
	}
}

func TestChainHashCompute_Deterministic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

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
		t.Errorf("expected deterministic hash, got %q and %q", hash1, hash2)
	}
}

func TestChainHashCompute_Is27Characters(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface content\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
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
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n\n# Public\n\n## Context\n\ninitial context\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")

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

func TestChainHashCompute_ChangesWhenDependencyContentChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n\n# Public\n\n## Context\n\nroot context\n")
	testWriteFile(t, "code-from-spec/b/_node.md", "# SPEC/b\n\n# Public\n\n## Interface\n\ninitial b interface\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\na interface\n")

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

func TestChainHashCompute_ChangesWhenTargetPublicChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

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

func TestChainHashCompute_ChangesWhenTargetAgentChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

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
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n\n# Public\n\n## Context\n\nsome context\n")
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
		t.Errorf("expected 27-character hash, got %d: %q", len(hash), hash)
	}
}

func TestChainHashCompute_AncestorWithoutPublicSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")
	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n\n# Public\n\n## Context\n\nsome context\n")

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
		t.Error("expected hashes to differ when ancestor Public section is removed")
	}
}

func TestChainHashCompute_MultipleAncestorsOrderMatters(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n\n# Public\n\n## Context\n\nroot context\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Context\n\na context\n")
	testWriteFile(t, "code-from-spec/a/b/_node.md", "# SPEC/a/b\n\n# Public\n\n## Interface\n\nb interface\n")

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
		t.Error("expected hashes to differ when ancestor order differs")
	}
}

func TestChainHashCompute_SpecDependencyWithoutQualifierHashesPublicSubsections(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/b/_node.md", "# SPEC/b\n\n# Public\n\n## Interface\n\ninitial b interface\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\na interface\n")

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

func TestChainHashCompute_SpecDependencyWithQualifierHashesSubsection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/b/_node.md", "# SPEC/b\n\n# Public\n\n## Interface\n\ninitial b interface\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\na interface\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItemQualified("SPEC/b", "code-from-spec/b/_node.md", "interface"),
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
		t.Error("expected hash to change when qualified SPEC dependency subsection changes")
	}
}

func TestChainHashCompute_QualifierCaseNormalization(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/b/_node.md", "# SPEC/b\n\n# Public\n\n## Interface\n\nb interface content\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\na interface\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItemQualified("SPEC/b", "code-from-spec/b/_node.md", "INTERFACE"),
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Errorf("expected no error with uppercase qualifier, got: %v", err)
	}
}

func TestChainHashCompute_ArtifactDependencyHashesFullFileContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")
	testWriteFile(t, "output/artifact.go", "// initial content\npackage main\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ARTIFACT/out", "output/artifact.go"),
		},
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "output/artifact.go", "// modified content\npackage main\n")

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change when artifact file content changes")
	}
}

func TestChainHashCompute_ArtifactDependencyTagHashChangeIgnored(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")
	testWriteFile(t, "output/artifact.go", "// code-from-spec: SPEC/x/y@aAbBcCdDeEfFgGhHiIjJkKlLmMn\npackage main\n\nfunc Foo() {}\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ARTIFACT/out", "output/artifact.go"),
		},
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "output/artifact.go", "// code-from-spec: SPEC/x/y@zZyYxXwWvVuUtTsSrRqQpPoOnNm\npackage main\n\nfunc Foo() {}\n")

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hashBefore != hashAfter {
		t.Error("expected hash to remain unchanged when only artifact tag hash changes")
	}
}

func TestChainHashCompute_ExternalDependencyHashesAllContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")
	testWriteFile(t, "somefile.proto", "syntax = \"proto3\";\nmessage Foo {}\n")

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

	testWriteFile(t, "somefile.proto", "syntax = \"proto3\";\nmessage Foo {}\nmessage Bar {}\n")

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change when external file content changes")
	}
}

func TestChainHashCompute_LeadingBlankLinesRemovedFromSubsection(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()

	contentA := "# SPEC/a\n\n# Public\n\n## Interface\n\n\nactual content\n"
	contentB := "# SPEC/a\n\n# Public\n\n## Interface\nactual content\n"

	if err := os.MkdirAll(dirA+"/code-from-spec/a", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dirA+"/code-from-spec/a/_node.md", []byte(contentA), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(dirB+"/code-from-spec/a", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dirB+"/code-from-spec/a/_node.md", []byte(contentB), 0644); err != nil {
		t.Fatal(err)
	}

	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	if err := os.Chdir(dirA); err != nil {
		t.Fatal(err)
	}
	chainA := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}
	hashA, err := chainhash.ChainHashCompute(chainA)
	if err != nil {
		t.Fatalf("unexpected error for chain A: %v", err)
	}

	if err := os.Chdir(dirB); err != nil {
		t.Fatal(err)
	}
	chainB := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}
	hashB, err := chainhash.ChainHashCompute(chainB)
	if err != nil {
		t.Fatalf("unexpected error for chain B: %v", err)
	}

	if hashA != hashB {
		t.Errorf("expected hashes to be equal after stripping leading blank lines, got %q and %q", hashA, hashB)
	}
}

func TestChainHashCompute_TrailingBlankLinesRemovedFromSubsection(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()

	contentA := "# SPEC/a\n\n# Public\n\n## Interface\nactual content\n\n\n"
	contentB := "# SPEC/a\n\n# Public\n\n## Interface\nactual content\n"

	if err := os.MkdirAll(dirA+"/code-from-spec/a", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dirA+"/code-from-spec/a/_node.md", []byte(contentA), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(dirB+"/code-from-spec/a", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dirB+"/code-from-spec/a/_node.md", []byte(contentB), 0644); err != nil {
		t.Fatal(err)
	}

	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	if err := os.Chdir(dirA); err != nil {
		t.Fatal(err)
	}
	chainA := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}
	hashA, err := chainhash.ChainHashCompute(chainA)
	if err != nil {
		t.Fatalf("unexpected error for chain A: %v", err)
	}

	if err := os.Chdir(dirB); err != nil {
		t.Fatal(err)
	}
	chainB := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}
	hashB, err := chainhash.ChainHashCompute(chainB)
	if err != nil {
		t.Fatalf("unexpected error for chain B: %v", err)
	}

	if hashA != hashB {
		t.Errorf("expected hashes to be equal after stripping trailing blank lines, got %q and %q", hashA, hashB)
	}
}

func TestChainHashCompute_InteriorBlankLinesPreserved(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()

	contentA := "# SPEC/a\n\n# Public\n\n## Interface\nline one\n\nline two\n"
	contentB := "# SPEC/a\n\n# Public\n\n## Interface\nline one\nline two\n"

	if err := os.MkdirAll(dirA+"/code-from-spec/a", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dirA+"/code-from-spec/a/_node.md", []byte(contentA), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(dirB+"/code-from-spec/a", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dirB+"/code-from-spec/a/_node.md", []byte(contentB), 0644); err != nil {
		t.Fatal(err)
	}

	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	if err := os.Chdir(dirA); err != nil {
		t.Fatal(err)
	}
	chainA := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}
	hashA, err := chainhash.ChainHashCompute(chainA)
	if err != nil {
		t.Fatalf("unexpected error for chain A: %v", err)
	}

	if err := os.Chdir(dirB); err != nil {
		t.Fatal(err)
	}
	chainB := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}
	hashB, err := chainhash.ChainHashCompute(chainB)
	if err != nil {
		t.Fatalf("unexpected error for chain B: %v", err)
	}

	if hashA == hashB {
		t.Error("expected hashes to differ when interior blank lines are removed")
	}
}

func TestChainHashCompute_TargetPublicAndAgentBothContribute(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n\n# Agent\n\nagent guidance\n")

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

func TestChainHashCompute_TargetWithoutAgentNoError(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27-character hash, got %d: %q", len(hash), hash)
	}
}

func TestChainHashCompute_InputHashesFullFileContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")
	testWriteFile(t, "input/data.md", "initial input content\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
		Input:  testChainItem("ARTIFACT/input", "input/data.md"),
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteFile(t, "input/data.md", "modified input content\n")

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change when input file content changes")
	}
}

func TestChainHashCompute_NoInputSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
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

func TestChainHashCompute_UnreadableSpecNodeFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for non-existent spec node file, got nil")
	}
	if !errors.Is(err, chainhash.ErrParseFailure) {
		t.Errorf("expected ErrParseFailure, got: %v", err)
	}
}

func TestChainHashCompute_UnreadableArtifactFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ARTIFACT/out", "nonexistent/artifact.go"),
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for non-existent artifact file, got nil")
	}
	if !errors.Is(err, file.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestChainHashCompute_UnreadableExternalFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\n\nsome interface\n")

	chain := &chainresolver.Chain{
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md"),
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("EXTERNAL/somefile", "nonexistent/somefile.proto"),
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for non-existent external file, got nil")
	}
	if !errors.Is(err, file.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}
