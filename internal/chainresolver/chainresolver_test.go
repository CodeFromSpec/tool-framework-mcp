// code-from-spec: ROOT/golang/tests/chain/resolver@1RcdT-7p2xN4tiBVL_pGi54bb_I
package chainresolver_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
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

func testWriteNodeFile(t *testing.T, logicalName string, content string) {
	t.Helper()
	var relPath string
	if logicalName == "ROOT" {
		relPath = "code-from-spec/_node.md"
	} else {
		suffix := logicalName[len("ROOT/"):]
		relPath = "code-from-spec/" + suffix + "/_node.md"
	}

	dir := relPath[:len(relPath)-len("_node.md")]
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testWriteNodeFile mkdir: %v", err)
	}
	if err := os.WriteFile(relPath, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteNodeFile write: %v", err)
	}
}

func testNodeContent(logicalName string) string {
	return "# " + logicalName + "\n"
}

func testNodeWithFrontmatter(logicalName string, fm string) string {
	return "---\n" + fm + "\n---\n\n# " + logicalName + "\n"
}

func testFindChainItem(items []*chainresolver.ChainItem, logicalName string) *chainresolver.ChainItem {
	for _, item := range items {
		if item.LogicalName == logicalName {
			return item
		}
	}
	return nil
}

func testFindChainItemByPath(items []*chainresolver.ChainItem, filePath string) *chainresolver.ChainItem {
	for _, item := range items {
		if item.FilePath.Value == filePath {
			return item
		}
	}
	return nil
}

func TestRootAsTarget(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", testNodeContent("ROOT"))

	chain, err := chainresolver.ChainResolve("ROOT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Ancestors) != 0 {
		t.Errorf("expected empty ancestors, got %d", len(chain.Ancestors))
	}
	if chain.Target == nil {
		t.Fatal("expected non-nil target")
	}
	if chain.Target.LogicalName != "ROOT" {
		t.Errorf("expected target logical name ROOT, got %s", chain.Target.LogicalName)
	}
	if chain.Target.Qualifier != "" {
		t.Errorf("expected absent qualifier, got %q", chain.Target.Qualifier)
	}
	if len(chain.Dependencies) != 0 {
		t.Errorf("expected empty dependencies, got %d", len(chain.Dependencies))
	}
	if len(chain.External) != 0 {
		t.Errorf("expected empty external, got %d", len(chain.External))
	}
	if chain.Input != nil {
		t.Errorf("expected absent input, got %v", chain.Input)
	}
}

func TestLinearChainAncestorsOrder(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", testNodeContent("ROOT"))
	testWriteNodeFile(t, "ROOT/a", testNodeContent("ROOT/a"))
	testWriteNodeFile(t, "ROOT/a/b", testNodeContent("ROOT/a/b"))

	chain, err := chainresolver.ChainResolve("ROOT/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Ancestors) != 2 {
		t.Fatalf("expected 2 ancestors, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].LogicalName != "ROOT" {
		t.Errorf("expected first ancestor ROOT, got %s", chain.Ancestors[0].LogicalName)
	}
	if chain.Ancestors[1].LogicalName != "ROOT/a" {
		t.Errorf("expected second ancestor ROOT/a, got %s", chain.Ancestors[1].LogicalName)
	}
	if chain.Target == nil || chain.Target.LogicalName != "ROOT/a/b" {
		t.Errorf("expected target ROOT/a/b")
	}
}

func TestSingleParent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", testNodeContent("ROOT"))
	testWriteNodeFile(t, "ROOT/a", testNodeContent("ROOT/a"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Ancestors) != 1 {
		t.Fatalf("expected 1 ancestor, got %d", len(chain.Ancestors))
	}
	if chain.Ancestors[0].LogicalName != "ROOT" {
		t.Errorf("expected ancestor ROOT, got %s", chain.Ancestors[0].LogicalName)
	}
	if chain.Target == nil || chain.Target.LogicalName != "ROOT/a" {
		t.Errorf("expected target ROOT/a")
	}
}

func TestTargetWithEmptyFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", testNodeContent("ROOT"))
	testWriteNodeFile(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", ""))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Ancestors) != 1 {
		t.Errorf("expected 1 ancestor, got %d", len(chain.Ancestors))
	}
	if chain.Target == nil || chain.Target.LogicalName != "ROOT/a" {
		t.Errorf("expected target ROOT/a")
	}
	if len(chain.Dependencies) != 0 {
		t.Errorf("expected empty dependencies, got %d", len(chain.Dependencies))
	}
	if len(chain.External) != 0 {
		t.Errorf("expected empty external, got %d", len(chain.External))
	}
	if chain.Input != nil {
		t.Errorf("expected absent input")
	}
}

func TestDependencyWithoutQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", testNodeContent("ROOT"))
	testWriteNodeFile(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ROOT/b"))
	testWriteNodeFile(t, "ROOT/b", testNodeContent("ROOT/b"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ROOT/b" {
		t.Errorf("expected ROOT/b, got %s", dep.LogicalName)
	}
	if dep.Qualifier != "" {
		t.Errorf("expected absent qualifier, got %q", dep.Qualifier)
	}
}

func TestDependencyWithQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", testNodeContent("ROOT"))
	testWriteNodeFile(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ROOT/b(interface)"))
	testWriteNodeFile(t, "ROOT/b", testNodeContent("ROOT/b"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ROOT/b" {
		t.Errorf("expected logical name ROOT/b, got %s", dep.LogicalName)
	}
	if dep.Qualifier != "interface" {
		t.Errorf("expected qualifier interface, got %q", dep.Qualifier)
	}
}

func TestDependenciesSortedByFilePath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", testNodeContent("ROOT"))
	testWriteNodeFile(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ROOT/z\n  - ROOT/m\n  - ROOT/b"))
	testWriteNodeFile(t, "ROOT/z", testNodeContent("ROOT/z"))
	testWriteNodeFile(t, "ROOT/m", testNodeContent("ROOT/m"))
	testWriteNodeFile(t, "ROOT/b", testNodeContent("ROOT/b"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 3 {
		t.Fatalf("expected 3 dependencies, got %d", len(chain.Dependencies))
	}

	bPath := pathutils.PathCfs{Value: "code-from-spec/b/_node.md"}
	mPath := pathutils.PathCfs{Value: "code-from-spec/m/_node.md"}
	zPath := pathutils.PathCfs{Value: "code-from-spec/z/_node.md"}

	if chain.Dependencies[0].FilePath.Value != bPath.Value {
		t.Errorf("expected first dep path %s, got %s", bPath.Value, chain.Dependencies[0].FilePath.Value)
	}
	if chain.Dependencies[1].FilePath.Value != mPath.Value {
		t.Errorf("expected second dep path %s, got %s", mPath.Value, chain.Dependencies[1].FilePath.Value)
	}
	if chain.Dependencies[2].FilePath.Value != zPath.Value {
		t.Errorf("expected third dep path %s, got %s", zPath.Value, chain.Dependencies[2].FilePath.Value)
	}
}

func TestArtifactDependencyResolved(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", testNodeContent("ROOT"))
	testWriteNodeFile(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ARTIFACT/b"))
	testWriteNodeFile(t, "ROOT/b", testNodeWithFrontmatter("ROOT/b", "output: out/lib.go"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	dep := chain.Dependencies[0]
	if dep.LogicalName != "ARTIFACT/b" {
		t.Errorf("expected ARTIFACT/b, got %s", dep.LogicalName)
	}
	if dep.FilePath.Value != "out/lib.go" {
		t.Errorf("expected file_path out/lib.go, got %s", dep.FilePath.Value)
	}
}

func TestArtifactDependencyNoOutput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", testNodeContent("ROOT"))
	testWriteNodeFile(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ARTIFACT/b"))
	testWriteNodeFile(t, "ROOT/b", testNodeWithFrontmatter("ROOT/b", ""))

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestArtifactDependencyFileNotExistOnDisk(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", testNodeContent("ROOT"))
	testWriteNodeFile(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ARTIFACT/b"))
	testWriteNodeFile(t, "ROOT/b", testNodeWithFrontmatter("ROOT/b", "output: out/lib.go"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].FilePath.Value != "out/lib.go" {
		t.Errorf("expected file_path out/lib.go, got %s", chain.Dependencies[0].FilePath.Value)
	}
}

func TestMixedRootAndArtifactDependencies(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", testNodeContent("ROOT"))
	testWriteNodeFile(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ROOT/c\n  - ARTIFACT/b"))
	testWriteNodeFile(t, "ROOT/b", testNodeWithFrontmatter("ROOT/b", "output: out/lib.go"))
	testWriteNodeFile(t, "ROOT/c", testNodeContent("ROOT/c"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}

	artifactDep := testFindChainItemByPath(chain.Dependencies, "out/lib.go")
	if artifactDep == nil {
		t.Error("expected ARTIFACT/b dependency with path out/lib.go")
	}

	rootDep := testFindChainItem(chain.Dependencies, "ROOT/c")
	if rootDep == nil {
		t.Error("expected ROOT/c dependency")
	}

	if chain.Dependencies[0].FilePath.Value >= chain.Dependencies[1].FilePath.Value {
		t.Errorf("expected dependencies sorted by file path, got %s then %s",
			chain.Dependencies[0].FilePath.Value, chain.Dependencies[1].FilePath.Value)
	}
}

func TestDedupExactDuplicate(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", testNodeContent("ROOT"))
	testWriteNodeFile(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ROOT/b\n  - ROOT/b"))
	testWriteNodeFile(t, "ROOT/b", testNodeContent("ROOT/b"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency (deduped), got %d", len(chain.Dependencies))
	}
}

func TestDedupNoQualifierSubsumesQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", testNodeContent("ROOT"))
	testWriteNodeFile(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ROOT/b\n  - ROOT/b(interface)"))
	testWriteNodeFile(t, "ROOT/b", testNodeContent("ROOT/b"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].Qualifier != "" {
		t.Errorf("expected absent qualifier (no-qualifier wins), got %q", chain.Dependencies[0].Qualifier)
	}
}

func TestDedupQualifierBeforeNoQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", testNodeContent("ROOT"))
	testWriteNodeFile(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ROOT/b(interface)\n  - ROOT/b"))
	testWriteNodeFile(t, "ROOT/b", testNodeContent("ROOT/b"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(chain.Dependencies))
	}
	if chain.Dependencies[0].Qualifier != "" {
		t.Errorf("expected absent qualifier (no-qualifier wins), got %q", chain.Dependencies[0].Qualifier)
	}
}

func TestDedupSameFileDifferentQualifiers(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", testNodeContent("ROOT"))
	testWriteNodeFile(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ROOT/b(interface)\n  - ROOT/b(constraints)"))
	testWriteNodeFile(t, "ROOT/b", testNodeContent("ROOT/b"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(chain.Dependencies))
	}

	qualifiers := []string{chain.Dependencies[0].Qualifier, chain.Dependencies[1].Qualifier}
	foundConstraints := false
	foundInterface := false
	for _, q := range qualifiers {
		if q == "constraints" {
			foundConstraints = true
		}
		if q == "interface" {
			foundInterface = true
		}
	}
	if !foundConstraints || !foundInterface {
		t.Errorf("expected qualifiers constraints and interface, got %v", qualifiers)
	}
}

func TestDedupDuplicateArtifact(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", testNodeContent("ROOT"))
	testWriteNodeFile(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - ARTIFACT/b\n  - ARTIFACT/b"))
	testWriteNodeFile(t, "ROOT/b", testNodeWithFrontmatter("ROOT/b", "output: out/lib.go"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.Dependencies) != 1 {
		t.Errorf("expected 1 dependency (deduped), got %d", len(chain.Dependencies))
	}
}

func TestExternalEntriesSorted(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", testNodeContent("ROOT"))
	testWriteNodeFile(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "external:\n  - path: docs/api.yaml\n  - path: proto/v1.proto"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.External) != 2 {
		t.Fatalf("expected 2 external entries, got %d", len(chain.External))
	}
	if chain.External[0].Path != "docs/api.yaml" {
		t.Errorf("expected first external docs/api.yaml, got %s", chain.External[0].Path)
	}
	if chain.External[1].Path != "proto/v1.proto" {
		t.Errorf("expected second external proto/v1.proto, got %s", chain.External[1].Path)
	}
}

func TestExternalEmpty(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", testNodeContent("ROOT"))
	testWriteNodeFile(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", ""))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain.External) != 0 {
		t.Errorf("expected empty external, got %d", len(chain.External))
	}
}

func TestInputResolved(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", testNodeContent("ROOT"))
	testWriteNodeFile(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "input: ARTIFACT/b"))
	testWriteNodeFile(t, "ROOT/b", testNodeWithFrontmatter("ROOT/b", "output: out/data.json"))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chain.Input == nil {
		t.Fatal("expected non-nil input")
	}
	if chain.Input.LogicalName != "ARTIFACT/b" {
		t.Errorf("expected ARTIFACT/b, got %s", chain.Input.LogicalName)
	}
	if chain.Input.FilePath.Value != "out/data.json" {
		t.Errorf("expected file_path out/data.json, got %s", chain.Input.FilePath.Value)
	}
}

func TestInputAbsent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", testNodeContent("ROOT"))
	testWriteNodeFile(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", ""))

	chain, err := chainresolver.ChainResolve("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chain.Input != nil {
		t.Errorf("expected absent input, got %v", chain.Input)
	}
}

func TestUnrecognizedPrefixInDependsOn(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", testNodeContent("ROOT"))
	testWriteNodeFile(t, "ROOT/a", testNodeWithFrontmatter("ROOT/a", "depends_on:\n  - UNKNOWN/something"))

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnresolvableArtifact) {
		t.Errorf("expected ErrUnresolvableArtifact, got %v", err)
	}
}

func TestInvalidTargetLogicalName(t *testing.T) {
	_, err := chainresolver.ChainResolve("INVALID/something")
	if err == nil {
		t.Fatal("expected error for invalid logical name, got nil")
	}
}

func TestUnreadableFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNodeFile(t, "ROOT", testNodeContent("ROOT"))

	invalidYAML := "---\nkey: [unclosed bracket\n---\n\n# ROOT/a\n"
	if err := os.MkdirAll("code-from-spec/a", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/a/_node.md", []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, err := chainresolver.ChainResolve("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, chainresolver.ErrUnreadableFrontmatter) {
		t.Errorf("expected ErrUnreadableFrontmatter, got %v", err)
	}
}
