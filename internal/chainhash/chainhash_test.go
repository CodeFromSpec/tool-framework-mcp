// code-from-spec: ROOT/golang/internal/chain_hash/tests@CIVzZwY5fTYcuJhnLGxorT-o8yc
package chainhash

// Tests for ComputeChainHash. These are internal tests (same package as the
// implementation) so they share access to unexported identifiers.
//
// Spec constraints verified here:
//   - The returned hash is always exactly 27 characters (base64url, no padding).
//   - The hash is deterministic: calling ComputeChainHash twice on the same
//     tree produces the same result.
//   - The hash changes when any file in the chain changes.
//
// Each test builds a minimal spec tree inside t.TempDir() so tests are
// completely isolated from the real repository layout.

import (
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// testWriteFile creates all parent directories and writes content to path.
// It is a test helper — failures are reported via t.Fatal so the calling test
// stops immediately.
func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("testWriteFile: mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("testWriteFile: write %s: %v", path, err)
	}
}

// testMakeMinimalChain builds the smallest valid spec chain for a given
// logical name inside root and returns root.
//
// The chain for "A/B/C" is made up of three files:
//
//	<root>/code-from-spec/A/spec.md          (root node)
//	<root>/code-from-spec/A/B/spec.md        (intermediate node)
//	<root>/code-from-spec/A/B/C/spec.md      (leaf / target node)
//
// Each file contains just enough content to be a real spec entry.
func testMakeMinimalChain(t *testing.T, root, logicalName string) {
	t.Helper()
	// Build each segment of the path as its own spec.md file.
	// The content does not need to be valid YAML for hashing purposes;
	// it just needs to be non-empty so the files exist on disk.
	base := filepath.Join(root, "code-from-spec")
	parts := splitLogicalName(logicalName) // uses unexported helper from impl
	for i := range parts {
		dir := filepath.Join(append([]string{base}, parts[:i+1]...)...)
		testWriteFile(t, filepath.Join(dir, "spec.md"), "# spec for "+filepath.Join(parts[:i+1]...)+"\n")
	}
}

// testChainPaths returns the list of absolute file paths that form the chain
// for logicalName rooted at root, in order from root to leaf.
// This mirrors the same walk that ComputeChainHash performs, letting tests
// mutate specific positions in the chain.
func testChainPaths(t *testing.T, root, logicalName string) []string {
	t.Helper()
	base := filepath.Join(root, "code-from-spec")
	parts := splitLogicalName(logicalName)
	paths := make([]string, len(parts))
	for i := range parts {
		dir := filepath.Join(append([]string{base}, parts[:i+1]...)...)
		paths[i] = filepath.Join(dir, "spec.md")
	}
	return paths
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// TestComputeChainHash_Length verifies that the returned hash is exactly
// 27 characters, which is the correct length for a 20-byte SHA-1 value
// encoded as base64url without padding ( ceil(20*8/6) = 27 ).
func TestComputeChainHash_Length(t *testing.T) {
	root := t.TempDir()
	logicalName := "FOO/BAR/BAZ"
	testMakeMinimalChain(t, root, logicalName)

	hash, err := computeChainHashInDir(root, logicalName)
	if err != nil {
		t.Fatalf("ComputeChainHash returned unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected hash length 27, got %d (hash=%q)", len(hash), hash)
	}
}

// TestComputeChainHash_Deterministic verifies that calling ComputeChainHash
// twice on exactly the same spec tree returns the same result.
func TestComputeChainHash_Deterministic(t *testing.T) {
	root := t.TempDir()
	logicalName := "ALPHA/BETA"
	testMakeMinimalChain(t, root, logicalName)

	first, err := computeChainHashInDir(root, logicalName)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	second, err := computeChainHashInDir(root, logicalName)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if first != second {
		t.Errorf("hash is not deterministic: first=%q second=%q", first, second)
	}
}

// TestComputeChainHash_ChangesOnLeafEdit verifies that modifying the leaf
// spec file (the node itself) produces a different hash.
func TestComputeChainHash_ChangesOnLeafEdit(t *testing.T) {
	root := t.TempDir()
	logicalName := "NS/PARENT/CHILD"
	testMakeMinimalChain(t, root, logicalName)

	before, err := computeChainHashInDir(root, logicalName)
	if err != nil {
		t.Fatalf("before edit: %v", err)
	}

	// Modify the leaf (last path in the chain).
	paths := testChainPaths(t, root, logicalName)
	leafPath := paths[len(paths)-1]
	testWriteFile(t, leafPath, "# modified leaf content\n")

	after, err := computeChainHashInDir(root, logicalName)
	if err != nil {
		t.Fatalf("after edit: %v", err)
	}

	if before == after {
		t.Errorf("hash did not change after modifying the leaf file")
	}
}

// TestComputeChainHash_ChangesOnRootEdit verifies that modifying the root
// (first) spec file in the chain also changes the hash, confirming that all
// chain positions contribute to the hash, not just the leaf.
func TestComputeChainHash_ChangesOnRootEdit(t *testing.T) {
	root := t.TempDir()
	logicalName := "NS/PARENT/CHILD"
	testMakeMinimalChain(t, root, logicalName)

	before, err := computeChainHashInDir(root, logicalName)
	if err != nil {
		t.Fatalf("before edit: %v", err)
	}

	// Modify the root (first path in the chain).
	paths := testChainPaths(t, root, logicalName)
	testWriteFile(t, paths[0], "# modified root content\n")

	after, err := computeChainHashInDir(root, logicalName)
	if err != nil {
		t.Fatalf("after edit: %v", err)
	}

	if before == after {
		t.Errorf("hash did not change after modifying the root file")
	}
}

// TestComputeChainHash_ChangesOnIntermediateEdit verifies that modifying an
// intermediate spec file (neither root nor leaf) also changes the hash.
func TestComputeChainHash_ChangesOnIntermediateEdit(t *testing.T) {
	// Need at least three segments so there is a true intermediate node.
	root := t.TempDir()
	logicalName := "X/Y/Z"
	testMakeMinimalChain(t, root, logicalName)

	before, err := computeChainHashInDir(root, logicalName)
	if err != nil {
		t.Fatalf("before edit: %v", err)
	}

	// Modify the middle path (index 1 out of 0,1,2).
	paths := testChainPaths(t, root, logicalName)
	testWriteFile(t, paths[1], "# modified intermediate content\n")

	after, err := computeChainHashInDir(root, logicalName)
	if err != nil {
		t.Fatalf("after edit: %v", err)
	}

	if before == after {
		t.Errorf("hash did not change after modifying an intermediate file")
	}
}

// TestComputeChainHash_CRLFNormalization verifies that a file written with
// CRLF line endings produces the same hash as the same file written with LF
// line endings, because the implementation normalises CRLF → LF before hashing.
func TestComputeChainHash_CRLFNormalization(t *testing.T) {
	logicalName := "CRLF/TEST"

	// Build one tree with LF content.
	rootLF := t.TempDir()
	base := filepath.Join(rootLF, "code-from-spec")
	parts := splitLogicalName(logicalName)
	for i := range parts {
		dir := filepath.Join(append([]string{base}, parts[:i+1]...)...)
		// Explicit LF content.
		testWriteFile(t, filepath.Join(dir, "spec.md"), "line one\nline two\n")
	}

	// Build another tree with CRLF content that should normalise to the same.
	rootCRLF := t.TempDir()
	base2 := filepath.Join(rootCRLF, "code-from-spec")
	for i := range parts {
		dir := filepath.Join(append([]string{base2}, parts[:i+1]...)...)
		// Explicit CRLF content.
		testWriteFile(t, filepath.Join(dir, "spec.md"), "line one\r\nline two\r\n")
	}

	hashLF, err := computeChainHashInDir(rootLF, logicalName)
	if err != nil {
		t.Fatalf("LF tree: %v", err)
	}
	hashCRLF, err := computeChainHashInDir(rootCRLF, logicalName)
	if err != nil {
		t.Fatalf("CRLF tree: %v", err)
	}

	if hashLF != hashCRLF {
		t.Errorf("CRLF normalization failed: LF hash=%q, CRLF hash=%q", hashLF, hashCRLF)
	}
}

// TestComputeChainHash_MissingFile verifies that ComputeChainHash returns an
// error when a required spec file is missing from disk.
func TestComputeChainHash_MissingFile(t *testing.T) {
	root := t.TempDir()
	// Do NOT create any spec files — the chain is entirely absent.
	logicalName := "MISSING/NODE"

	_, err := computeChainHashInDir(root, logicalName)
	if err == nil {
		t.Error("expected an error for a missing spec tree, got nil")
	}
}

// TestComputeChainHash_SingleSegment verifies that a single-segment logical
// name (only a root node, no separators) works correctly and still produces
// a 27-character hash.
func TestComputeChainHash_SingleSegment(t *testing.T) {
	root := t.TempDir()
	logicalName := "STANDALONE"
	testMakeMinimalChain(t, root, logicalName)

	hash, err := computeChainHashInDir(root, logicalName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27-character hash, got %d (hash=%q)", len(hash), hash)
	}
}

// TestComputeChainHash_DifferentNamesProduceDifferentHashes verifies that two
// distinct logical names backed by distinct content do not accidentally
// collide on the same hash value.
func TestComputeChainHash_DifferentNamesProduceDifferentHashes(t *testing.T) {
	root := t.TempDir()
	// Build two separate chains in the same root.
	testMakeMinimalChain(t, root, "GROUP/NODE_A")
	testMakeMinimalChain(t, root, "GROUP/NODE_B")

	// Make their leaf content explicitly different so there is no chance of
	// accidental equality.
	paths_a := testChainPaths(t, root, "GROUP/NODE_A")
	paths_b := testChainPaths(t, root, "GROUP/NODE_B")
	testWriteFile(t, paths_a[len(paths_a)-1], "# content for NODE_A\n")
	testWriteFile(t, paths_b[len(paths_b)-1], "# content for NODE_B\n")

	hashA, err := computeChainHashInDir(root, "GROUP/NODE_A")
	if err != nil {
		t.Fatalf("NODE_A: %v", err)
	}
	hashB, err := computeChainHashInDir(root, "GROUP/NODE_B")
	if err != nil {
		t.Fatalf("NODE_B: %v", err)
	}

	if hashA == hashB {
		t.Errorf("expected different hashes for different nodes, both returned %q", hashA)
	}
}
