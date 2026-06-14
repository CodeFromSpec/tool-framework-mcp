// code-from-spec: ROOT/golang/tests/chain/hash@mXs7TeaKNtU9ylagM6QVGEhSbFM
package chainhash_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
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

func testWriteNodeFile(t *testing.T, relPath string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(relPath), 0755); err != nil {
		t.Fatalf("testWriteNodeFile mkdir: %v", err)
	}
	if err := os.WriteFile(relPath, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNodeFile write: %v", err)
	}
}

func testChainItem(unqualifiedLogicalName string, cfsPath string, qualifier *string) *chainresolver.ChainItem {
	return &chainresolver.ChainItem{
		UnqualifiedLogicalName: unqualifiedLogicalName,
		FilePath:               pathutils.PathCfs{Value: cfsPath},
		Qualifier:              qualifier,
	}
}

func testStrPtr(s string) *string {
	return &s
}

func TestChainHashCompute_Deterministic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "# SPEC\n\n# Public\n\n## Context\nsome ancestor content\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface content\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("SPEC", "code-from-spec/_node.md", nil),
		},
		Dependencies: []*chainresolver.ChainItem{},
		Target:       testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:        nil,
	}

	result1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result1 != result2 {
		t.Errorf("expected deterministic hash, got %q and %q", result1, result2)
	}
}

func TestChainHashCompute_Is27Characters(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface content\n")

	chain := &chainresolver.Chain{
		Ancestors:    []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{},
		Target:       testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:        nil,
	}

	result, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 27 {
		t.Errorf("expected 27-character hash, got %d characters: %q", len(result), result)
	}
}

func TestChainHashCompute_ChangesWhenAncestorContentChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "# SPEC\n\n# Public\n\n## Context\noriginal ancestor content\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface content\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("SPEC", "code-from-spec/_node.md", nil),
		},
		Dependencies: []*chainresolver.ChainItem{},
		Target:       testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:        nil,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "code-from-spec/_node.md", "# SPEC\n\n# Public\n\n## Context\nmodified ancestor content\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hash to change when ancestor content changes")
	}
}

func TestChainHashCompute_ChangesWhenDependencyContentChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "# SPEC\n\n# Public\n\n## Context\nsome context\n")
	testWriteNodeFile(t, "code-from-spec/b/_node.md", "# SPEC/b\n\n# Public\n\n## Interface\noriginal dependency content\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface content\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("SPEC/b", "code-from-spec/b/_node.md", nil),
		},
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:  nil,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "code-from-spec/b/_node.md", "# SPEC/b\n\n# Public\n\n## Interface\nmodified dependency content\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hash to change when dependency content changes")
	}
}

func TestChainHashCompute_ChangesWhenTargetPublicChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\noriginal target interface\n")

	chain := &chainresolver.Chain{
		Ancestors:    []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{},
		Target:       testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:        nil,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nmodified target interface\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hash to change when target Public changes")
	}
}

func TestChainHashCompute_ChangesWhenTargetAgentChanges(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface\n\n# Agent\noriginal agent content\n")

	chain := &chainresolver.Chain{
		Ancestors:    []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{},
		Target:       testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:        nil,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface\n\n# Agent\nmodified agent content\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hash to change when target Agent changes")
	}
}

func TestChainHashCompute_AncestorWithPublicSubsectionsContributesHash(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "# SPEC\n\n# Public\n\n## Context\nsome ancestor context\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface content\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("SPEC", "code-from-spec/_node.md", nil),
		},
		Dependencies: []*chainresolver.ChainItem{},
		Target:       testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:        nil,
	}

	result, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 27 {
		t.Errorf("expected 27-character hash, got %d: %q", len(result), result)
	}
}

func TestChainHashCompute_AncestorWithoutPublicSection_Skipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "# SPEC\n\n# Public\n\n## Context\nsome context\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface content\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("SPEC", "code-from-spec/_node.md", nil),
		},
		Dependencies: []*chainresolver.ChainItem{},
		Target:       testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:        nil,
	}

	hashWithPublic, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "code-from-spec/_node.md", "# SPEC\n")

	hashWithoutPublic, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hashWithPublic == hashWithoutPublic {
		t.Error("expected hash to differ when ancestor has no Public section")
	}
}

func TestChainHashCompute_MultipleAncestors_OrderMatters(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/_node.md", "# SPEC\n\n# Public\n\n## Context\nroot context\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Context\nmid context\n")
	testWriteNodeFile(t, "code-from-spec/a/b/_node.md", "# SPEC/a/b\n\n# Public\n\n## Interface\nsome interface content\n")

	chainForward := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("SPEC", "code-from-spec/_node.md", nil),
			testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		},
		Dependencies: []*chainresolver.ChainItem{},
		Target:       testChainItem("SPEC/a/b", "code-from-spec/a/b/_node.md", nil),
		Input:        nil,
	}

	chainReversed := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{
			testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
			testChainItem("SPEC", "code-from-spec/_node.md", nil),
		},
		Dependencies: []*chainresolver.ChainItem{},
		Target:       testChainItem("SPEC/a/b", "code-from-spec/a/b/_node.md", nil),
		Input:        nil,
	}

	hashForward, err := chainhash.ChainHashCompute(chainForward)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hashReversed, err := chainhash.ChainHashCompute(chainReversed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hashForward == hashReversed {
		t.Error("expected different hashes for different ancestor orderings")
	}
}

func TestChainHashCompute_SpecDependencyWithoutQualifier_HashesPublicSubsections(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/b/_node.md", "# SPEC/b\n\n# Public\n\n## Interface\noriginal interface\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface content\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("SPEC/b", "code-from-spec/b/_node.md", nil),
		},
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:  nil,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "code-from-spec/b/_node.md", "# SPEC/b\n\n# Public\n\n## Interface\nmodified interface\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hash to change when SPEC dependency Public changes")
	}
}

func TestChainHashCompute_SpecDependencyWithQualifier_HashesSubsection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/b/_node.md", "# SPEC/b\n\n# Public\n\n## Interface\noriginal interface\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface content\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("SPEC/b", "code-from-spec/b/_node.md", testStrPtr("interface")),
		},
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:  nil,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "code-from-spec/b/_node.md", "# SPEC/b\n\n# Public\n\n## Interface\nmodified interface\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hash to change when qualified SPEC dependency subsection changes")
	}
}

func TestChainHashCompute_QualifierCaseNormalization(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/b/_node.md", "# SPEC/b\n\n# Public\n\n## Interface\nsome interface content\n")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface content\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("SPEC/b", "code-from-spec/b/_node.md", testStrPtr("INTERFACE")),
		},
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:  nil,
	}

	result, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 27 {
		t.Errorf("expected 27-character hash, got %d: %q", len(result), result)
	}
}

func TestChainHashCompute_ArtifactDependency_HashesFullFileContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "output.md", "original artifact content")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface content\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ARTIFACT/x", "output.md", nil),
		},
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:  nil,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "output.md", "modified artifact content")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hash to change when artifact dependency content changes")
	}
}

func TestChainHashCompute_ArtifactDependency_TagHashChangeIgnored(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	bodyContent := "some body content\n"
	contentWithTagA := "// code-from-spec: ARTIFACT/x@AAAAAAAAAAAAAAAAAAAAAAAAAAA\n" + bodyContent
	contentWithTagB := "// code-from-spec: ARTIFACT/x@zZyYxXwWvVuUtTsSrRqQpPoOnNm\n" + bodyContent

	testWriteNodeFile(t, "output.md", contentWithTagA)
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface content\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ARTIFACT/x", "output.md", nil),
		},
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:  nil,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "output.md", contentWithTagB)

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 != hash2 {
		t.Error("expected hash to be unchanged when only the artifact tag hash changes")
	}
}

func TestChainHashCompute_ExternalDependency_HashesAllContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "rules.md", "original external content")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface content\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("EXTERNAL/rules.md", "rules.md", nil),
		},
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:  nil,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "rules.md", "modified external content")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hash to change when external dependency content changes")
	}
}

func TestChainHashCompute_LeadingBlankLinesRemovedFromSubsection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface content\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("SPEC/b", "code-from-spec/b/_node.md", nil),
		},
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:  nil,
	}

	versionA := "# SPEC/b\n\n# Public\n\n## Interface\n\n\ninterface content line\n"
	testWriteNodeFile(t, "code-from-spec/b/_node.md", versionA)

	hashA, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error for version A: %v", err)
	}

	versionB := "# SPEC/b\n\n# Public\n\n## Interface\ninterface content line\n"
	testWriteNodeFile(t, "code-from-spec/b/_node.md", versionB)

	hashB, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error for version B: %v", err)
	}

	if hashA != hashB {
		t.Error("expected leading blank lines to be stripped so hashes match")
	}
}

func TestChainHashCompute_TrailingBlankLinesRemovedFromSubsection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface content\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("SPEC/b", "code-from-spec/b/_node.md", nil),
		},
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:  nil,
	}

	versionA := "# SPEC/b\n\n# Public\n\n## Interface\ninterface content line\n\n\n"
	testWriteNodeFile(t, "code-from-spec/b/_node.md", versionA)

	hashA, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error for version A: %v", err)
	}

	versionB := "# SPEC/b\n\n# Public\n\n## Interface\ninterface content line\n"
	testWriteNodeFile(t, "code-from-spec/b/_node.md", versionB)

	hashB, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error for version B: %v", err)
	}

	if hashA != hashB {
		t.Error("expected trailing blank lines to be stripped so hashes match")
	}
}

func TestChainHashCompute_InteriorBlankLinesPreserved(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface content\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("SPEC/b", "code-from-spec/b/_node.md", nil),
		},
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:  nil,
	}

	versionA := "# SPEC/b\n\n# Public\n\n## Interface\nfirst line\n\nsecond line\n"
	testWriteNodeFile(t, "code-from-spec/b/_node.md", versionA)

	hashA, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error for version A: %v", err)
	}

	versionB := "# SPEC/b\n\n# Public\n\n## Interface\nfirst line\nsecond line\n"
	testWriteNodeFile(t, "code-from-spec/b/_node.md", versionB)

	hashB, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error for version B: %v", err)
	}

	if hashA == hashB {
		t.Error("expected interior blank lines to be preserved so hashes differ")
	}
}

func TestChainHashCompute_TargetPublicAndAgentBothContribute(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface\n\n# Agent\nsome agent guidance\n")

	chain := &chainresolver.Chain{
		Ancestors:    []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{},
		Target:       testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:        nil,
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface\n")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hash to change when Agent section is removed")
	}
}

func TestChainHashCompute_TargetWithoutAgent_AgentSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface\n")

	chain := &chainresolver.Chain{
		Ancestors:    []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{},
		Target:       testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:        nil,
	}

	result, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 27 {
		t.Errorf("expected 27-character hash, got %d: %q", len(result), result)
	}
}

func TestChainHashCompute_InputHashesFullFileContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "input.md", "original input content")
	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface content\n")

	chain := &chainresolver.Chain{
		Ancestors:    []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{},
		Target:       testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:        testChainItem("ARTIFACT/input", "input.md", nil),
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testWriteNodeFile(t, "input.md", "modified input content")

	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hash to change when input content changes")
	}
}

func TestChainHashCompute_NoInput_Skipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface content\n")

	chain := &chainresolver.Chain{
		Ancestors:    []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{},
		Target:       testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:        nil,
	}

	result, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 27 {
		t.Errorf("expected 27-character hash, got %d: %q", len(result), result)
	}
}

func TestChainHashCompute_UnreadableSpecNodeFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	chain := &chainresolver.Chain{
		Ancestors:    []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{},
		Target:       testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:        nil,
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for unreadable spec node file, got nil")
	}

	if !errors.Is(err, chainhash.ErrParseFailure) {
		t.Errorf("expected ErrParseFailure, got: %v", err)
	}
}

func TestChainHashCompute_UnreadableArtifactFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface content\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("ARTIFACT/x", "nonexistent/artifact.md", nil),
		},
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:  nil,
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for unreadable artifact file, got nil")
	}

	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestChainHashCompute_UnreadableExternalFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n# Public\n\n## Interface\nsome interface content\n")

	chain := &chainresolver.Chain{
		Ancestors: []*chainresolver.ChainItem{},
		Dependencies: []*chainresolver.ChainItem{
			testChainItem("EXTERNAL/rules.md", "nonexistent/rules.md", nil),
		},
		Target: testChainItem("SPEC/a", "code-from-spec/a/_node.md", nil),
		Input:  nil,
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for unreadable external file, got nil")
	}

	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}
