// code-from-spec: ROOT/golang/internal/chain_hash/tests@Fl5aMLYgfb7sqDTevYFkpcrtjqk

// Package chainhash_test verifies ComputeChainHash behaviour.
//
// Test strategy (from spec):
//   - The hash must be deterministic across calls.
//   - The hash must always be exactly 27 characters.
//   - The hash must change when any file in the chain changes.
//
// Every test builds its own isolated spec tree under t.TempDir() so tests
// are fully independent and do not touch real project files.
//
// Helper naming: all helpers are prefixed with "test" to avoid collisions
// with unexported symbols in the implementation package.
package chainhash

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ----------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------

// testWriteFile creates the file at path (and any missing parent directories)
// with the given content. It calls t.Fatal on any error.
func testWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("testWriteFile: MkdirAll(%s): %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("testWriteFile: WriteFile(%s): %v", path, err)
	}
}

// testNodePath returns the _node.md path for a logical name under the given
// root directory. It mirrors the PathFromLogicalName resolution:
//
//	ROOT                → <root>/code-from-spec/_node.md
//	ROOT/<path>         → <root>/code-from-spec/<path>/_node.md
func testNodePath(root, logicalName string) string {
	const prefix = "ROOT"
	if logicalName == prefix {
		return filepath.Join(root, "code-from-spec", "_node.md")
	}
	rel := strings.TrimPrefix(logicalName, prefix+"/")
	// Replace forward slashes with the OS path separator.
	rel = filepath.FromSlash(rel)
	return filepath.Join(root, "code-from-spec", rel, "_node.md")
}

// testChdir changes the working directory to dir for the duration of the test.
// ComputeChainHash uses the working directory as the project root (all paths
// it constructs are relative to it), so each test must chdir into its temp tree.
func testChdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChdir: Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChdir: Chdir(%s): %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Fatalf("testChdir cleanup: Chdir(%s): %v", orig, err)
		}
	})
}

// testMakeMinimalTree creates the minimal spec tree required for the logical
// name and returns the path to the target node file. The tree contains only
// the target node and its ancestors, each with an empty frontmatter.
//
// "Minimal" means no depends_on, no external, no input. Each node file
// contains a single # Public section so there is content to hash.
func testMakeMinimalTree(t *testing.T, root, logicalName string) string {
	t.Helper()

	// Build the ancestor chain manually (root-first, not including target).
	// e.g. ROOT/a/b → ["ROOT", "ROOT/a"]
	var ancestors []string
	parts := strings.Split(logicalName, "/")
	for i := 1; i < len(parts); i++ {
		ancestors = append(ancestors, strings.Join(parts[:i], "/"))
	}

	// Write each ancestor with a simple # Public section.
	for _, anc := range ancestors {
		content := "# Public\n\nAncestor: " + anc + "\n"
		testWriteFile(t, testNodePath(root, anc), content)
	}

	// Write the target node.
	targetContent := "# Public\n\nTarget: " + logicalName + "\n"
	targetPath := testNodePath(root, logicalName)
	testWriteFile(t, targetPath, targetContent)

	return targetPath
}

// ----------------------------------------------------------------
// Tests
// ----------------------------------------------------------------

// TestComputeChainHash_Length verifies that the hash is always exactly 27
// characters long (the spec mandates this for base64url-encoded SHA-1 without
// padding).
func TestComputeChainHash_Length(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	testMakeMinimalTree(t, root, "ROOT/a/b")

	hash, err := ComputeChainHash("ROOT/a/b")
	if err != nil {
		t.Fatalf("ComputeChainHash returned unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected hash length 27, got %d (hash=%q)", len(hash), hash)
	}
}

// TestComputeChainHash_Deterministic verifies that calling ComputeChainHash
// twice on the same unchanged tree returns identical hashes.
func TestComputeChainHash_Deterministic(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	testMakeMinimalTree(t, root, "ROOT/x/y")

	hash1, err := ComputeChainHash("ROOT/x/y")
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	hash2, err := ComputeChainHash("ROOT/x/y")
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if hash1 != hash2 {
		t.Errorf("hash is not deterministic: first=%q, second=%q", hash1, hash2)
	}
}

// TestComputeChainHash_ChangesWhenTargetChanges verifies that the hash changes
// when the target node file is modified.
func TestComputeChainHash_ChangesWhenTargetChanges(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	targetPath := testMakeMinimalTree(t, root, "ROOT/mynode")

	before, err := ComputeChainHash("ROOT/mynode")
	if err != nil {
		t.Fatalf("before modification: %v", err)
	}

	// Modify the target file.
	testWriteFile(t, targetPath, "# Public\n\nTarget: modified content\n")

	after, err := ComputeChainHash("ROOT/mynode")
	if err != nil {
		t.Fatalf("after modification: %v", err)
	}

	if before == after {
		t.Errorf("hash did not change after target file modification (hash=%q)", before)
	}
}

// TestComputeChainHash_ChangesWhenAncestorChanges verifies that the hash
// changes when an ancestor's # Public section is modified.
func TestComputeChainHash_ChangesWhenAncestorChanges(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	testMakeMinimalTree(t, root, "ROOT/parent/child")
	// The parent is ROOT/parent.
	parentPath := testNodePath(root, "ROOT/parent")

	before, err := ComputeChainHash("ROOT/parent/child")
	if err != nil {
		t.Fatalf("before modification: %v", err)
	}

	// Modify the parent's # Public section.
	testWriteFile(t, parentPath, "# Public\n\nParent: modified ancestor content\n")

	after, err := ComputeChainHash("ROOT/parent/child")
	if err != nil {
		t.Fatalf("after modification: %v", err)
	}

	if before == after {
		t.Errorf("hash did not change after ancestor modification (hash=%q)", before)
	}
}

// TestComputeChainHash_ROOTNode verifies that ComputeChainHash works for ROOT
// itself (no ancestors, no parent chain).
func TestComputeChainHash_ROOTNode(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	// Write ROOT node.
	testWriteFile(t, testNodePath(root, "ROOT"), "# Public\n\nRoot content\n")

	hash, err := ComputeChainHash("ROOT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected hash length 27, got %d (hash=%q)", len(hash), hash)
	}
}

// TestComputeChainHash_DependsOnROOTRef verifies that a depends_on entry
// pointing to a ROOT/ node contributes to the hash via its # Public section.
// Changing the depended-upon node's # Public section must change the hash.
func TestComputeChainHash_DependsOnROOTRef(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	// Write the dependency node.
	depPath := testNodePath(root, "ROOT/dep")
	testWriteFile(t, depPath, "# Public\n\nDep content v1\n")

	// Write the target with a depends_on frontmatter block.
	targetPath := testNodePath(root, "ROOT/target")
	targetContent := "---\ndepends_on:\n  - ROOT/dep\n---\n# Public\n\nTarget content\n"
	testWriteFile(t, targetPath, targetContent)

	// Write ROOT ancestor.
	testWriteFile(t, testNodePath(root, "ROOT"), "# Public\n\nRoot\n")

	before, err := ComputeChainHash("ROOT/target")
	if err != nil {
		t.Fatalf("before: %v", err)
	}

	// Change the dependency.
	testWriteFile(t, depPath, "# Public\n\nDep content v2 (changed)\n")

	after, err := ComputeChainHash("ROOT/target")
	if err != nil {
		t.Fatalf("after: %v", err)
	}

	if before == after {
		t.Errorf("hash did not change after depends_on node modification")
	}
}

// TestComputeChainHash_DependsOnQualifier verifies that a ROOT/x(qualifier)
// depends_on entry hashes only the ## qualifier subsection within # Public.
// Changing only the subsection must change the hash; changing other content
// outside that subsection must also not affect it (the subsection drives it).
func TestComputeChainHash_DependsOnQualifier(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	// Write ROOT ancestor.
	testWriteFile(t, testNodePath(root, "ROOT"), "# Public\n\nRoot\n")

	// Write the dependency node with a named subsection.
	depPath := testNodePath(root, "ROOT/dep")
	depContent := "# Public\n\n## mysection\n\nSection content v1\n\n## other\n\nOther content\n"
	testWriteFile(t, depPath, depContent)

	// Target depends on only the "mysection" subsection.
	targetPath := testNodePath(root, "ROOT/target")
	testWriteFile(t, targetPath, "---\ndepends_on:\n  - ROOT/dep(mysection)\n---\n# Public\n\nTarget\n")

	hash1, err := ComputeChainHash("ROOT/target")
	if err != nil {
		t.Fatalf("hash1: %v", err)
	}

	// Change only the "other" subsection — hash must NOT change.
	depContent2 := "# Public\n\n## mysection\n\nSection content v1\n\n## other\n\nOther content CHANGED\n"
	testWriteFile(t, depPath, depContent2)

	hash2, err := ComputeChainHash("ROOT/target")
	if err != nil {
		t.Fatalf("hash2: %v", err)
	}
	if hash1 != hash2 {
		t.Errorf("hash changed when only unrelated subsection changed (expected stable)")
	}

	// Change the "mysection" subsection — hash must change.
	depContent3 := "# Public\n\n## mysection\n\nSection content v2 CHANGED\n\n## other\n\nOther content CHANGED\n"
	testWriteFile(t, depPath, depContent3)

	hash3, err := ComputeChainHash("ROOT/target")
	if err != nil {
		t.Fatalf("hash3: %v", err)
	}
	if hash1 == hash3 {
		t.Errorf("hash did not change when depends_on qualified subsection changed")
	}
}

// TestComputeChainHash_ExternalFile verifies that a node with an external
// file dependency has a hash that changes when the external file changes.
func TestComputeChainHash_ExternalFile(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	// Write ROOT.
	testWriteFile(t, testNodePath(root, "ROOT"), "# Public\n\nRoot\n")

	// Write an external file at a known path (relative to project root).
	extPath := filepath.Join(root, "somefile.txt")
	testWriteFile(t, extPath, "external content v1\n")

	// Write the target with an external frontmatter entry.
	// The path in the frontmatter must be relative to the project root (which
	// is root after testChdir).
	targetPath := testNodePath(root, "ROOT/target")
	targetContent := "---\nexternal:\n  - path: somefile.txt\n---\n# Public\n\nTarget\n"
	testWriteFile(t, targetPath, targetContent)

	before, err := ComputeChainHash("ROOT/target")
	if err != nil {
		t.Fatalf("before: %v", err)
	}

	// Change the external file.
	testWriteFile(t, extPath, "external content v2 (changed)\n")

	after, err := ComputeChainHash("ROOT/target")
	if err != nil {
		t.Fatalf("after: %v", err)
	}

	if before == after {
		t.Errorf("hash did not change after external file modification")
	}
}

// TestComputeChainHash_InputArtifact verifies that a node with an input
// artifact has a hash that changes when the artifact content changes.
//
// The input frontmatter field points to an ARTIFACT/ reference. The artifact
// is resolved via the node's outputs list, then hashed (with frontmatter stripped).
func TestComputeChainHash_InputArtifact(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	// Write ROOT.
	testWriteFile(t, testNodePath(root, "ROOT"), "# Public\n\nRoot\n")

	// Write the producer node — declares an output artifact.
	producerPath := testNodePath(root, "ROOT/producer")
	producerContent := "---\noutputs:\n  - id: code\n    path: some/output.go\n---\n# Public\n\nProducer\n"
	testWriteFile(t, producerPath, producerContent)

	// Write the artifact file itself.
	artifactPath := filepath.Join(root, "some", "output.go")
	testWriteFile(t, artifactPath, "// artifact content v1\npackage foo\n")

	// Write the consumer node — has input pointing to the artifact.
	targetPath := testNodePath(root, "ROOT/consumer")
	targetContent := "---\ninput: ARTIFACT/producer(code)\n---\n# Public\n\nConsumer\n"
	testWriteFile(t, targetPath, targetContent)

	before, err := ComputeChainHash("ROOT/consumer")
	if err != nil {
		t.Fatalf("before: %v", err)
	}

	// Modify the artifact (no frontmatter in this file, so stripped == original).
	testWriteFile(t, artifactPath, "// artifact content v2 (changed)\npackage foo\n")

	after, err := ComputeChainHash("ROOT/consumer")
	if err != nil {
		t.Fatalf("after: %v", err)
	}

	if before == after {
		t.Errorf("hash did not change after input artifact modification")
	}
}

// TestComputeChainHash_CRLFNormalization verifies that CRLF and LF line endings
// produce the same hash (normalization happens before hashing).
func TestComputeChainHash_CRLFNormalization(t *testing.T) {
	root1 := t.TempDir()
	root2 := t.TempDir()

	// LF version.
	testChdir(t, root1)
	testWriteFile(t, testNodePath(root1, "ROOT"), "# Public\n\nRoot\n")
	testWriteFile(t, testNodePath(root1, "ROOT/node"), "# Public\n\nContent\n")
	hashLF, err := ComputeChainHash("ROOT/node")
	if err != nil {
		t.Fatalf("LF version: %v", err)
	}

	// CRLF version — same logical content, different line endings.
	testChdir(t, root2)
	testWriteFile(t, testNodePath(root2, "ROOT"), "# Public\r\n\r\nRoot\r\n")
	testWriteFile(t, testNodePath(root2, "ROOT/node"), "# Public\r\n\r\nContent\r\n")
	hashCRLF, err := ComputeChainHash("ROOT/node")
	if err != nil {
		t.Fatalf("CRLF version: %v", err)
	}

	if hashLF != hashCRLF {
		t.Errorf("CRLF and LF produced different hashes: LF=%q, CRLF=%q", hashLF, hashCRLF)
	}
}

// TestComputeChainHash_NoPublicSection verifies that an ancestor with no
// # Public section is silently skipped (contributes nothing to the hash).
// The hash should still be 27 characters and not an error.
func TestComputeChainHash_NoPublicSection(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	// Write ROOT with no # Public section.
	testWriteFile(t, testNodePath(root, "ROOT"), "No public section here\n")

	// Write target with a # Public section.
	testWriteFile(t, testNodePath(root, "ROOT/child"), "# Public\n\nChild content\n")

	hash, err := ComputeChainHash("ROOT/child")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected hash length 27, got %d", len(hash))
	}
}

// TestComputeChainHash_AgentSection verifies that the # Agent section of the
// target also contributes to the hash. Changing it must change the hash.
func TestComputeChainHash_AgentSection(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	testWriteFile(t, testNodePath(root, "ROOT"), "# Public\n\nRoot\n")

	targetPath := testNodePath(root, "ROOT/node")
	testWriteFile(t, targetPath, "# Public\n\nPublic content\n\n# Agent\n\nAgent content v1\n")

	before, err := ComputeChainHash("ROOT/node")
	if err != nil {
		t.Fatalf("before: %v", err)
	}

	// Modify only the # Agent section.
	testWriteFile(t, targetPath, "# Public\n\nPublic content\n\n# Agent\n\nAgent content v2 (changed)\n")

	after, err := ComputeChainHash("ROOT/node")
	if err != nil {
		t.Fatalf("after: %v", err)
	}

	if before == after {
		t.Errorf("hash did not change after # Agent section modification")
	}
}

// TestComputeChainHash_InvalidLogicalName verifies that an error is returned
// for unsupported logical name formats.
func TestComputeChainHash_InvalidLogicalName(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	invalidNames := []string{
		"",
		"ARTIFACT/something(id)",
		"notROOT/foo",
		"random string",
	}

	for _, name := range invalidNames {
		_, err := ComputeChainHash(name)
		if err == nil {
			t.Errorf("expected error for logical name %q, got nil", name)
		}
	}
}

// TestComputeChainHash_MissingNodeFile verifies that an error is returned when
// the target node file does not exist on disk.
func TestComputeChainHash_MissingNodeFile(t *testing.T) {
	root := t.TempDir()
	testChdir(t, root)

	// Do NOT write the node file — just chdir.
	_, err := ComputeChainHash("ROOT/missing")
	if err == nil {
		t.Fatal("expected error for missing node file, got nil")
	}
	// The error message should mention the file or "unreadable".
	if !strings.Contains(err.Error(), "unreadable") {
		t.Errorf("expected error to contain %q, got: %v", "unreadable", err)
	}
}

// TestComputeChainHash_DependsOnSortOrder verifies that the order of
// depends_on entries in the frontmatter does not affect the hash (they are
// sorted alphabetically before hashing).
func TestComputeChainHash_DependsOnSortOrder(t *testing.T) {
	// Build two trees that are identical except the depends_on list order.

	buildTree := func(t *testing.T, root, order1, order2 string) {
		t.Helper()
		testWriteFile(t, testNodePath(root, "ROOT"), "# Public\n\nRoot\n")
		testWriteFile(t, testNodePath(root, "ROOT/aaaa"), "# Public\n\nDep aaaa\n")
		testWriteFile(t, testNodePath(root, "ROOT/zzzz"), "# Public\n\nDep zzzz\n")

		targetContent := "---\ndepends_on:\n  - " + order1 + "\n  - " + order2 + "\n---\n# Public\n\nTarget\n"
		testWriteFile(t, testNodePath(root, "ROOT/target"), targetContent)
	}

	root1 := t.TempDir()
	root2 := t.TempDir()

	// Tree 1: aaaa first.
	testChdir(t, root1)
	buildTree(t, root1, "ROOT/aaaa", "ROOT/zzzz")
	hash1, err := ComputeChainHash("ROOT/target")
	if err != nil {
		t.Fatalf("tree1: %v", err)
	}

	// Tree 2: zzzz first.
	testChdir(t, root2)
	buildTree(t, root2, "ROOT/zzzz", "ROOT/aaaa")
	hash2, err := ComputeChainHash("ROOT/target")
	if err != nil {
		t.Fatalf("tree2: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("hash differs based on depends_on list order: %q vs %q", hash1, hash2)
	}
}
