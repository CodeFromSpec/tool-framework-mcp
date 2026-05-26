// code-from-spec: ROOT/golang/internal/node_discovery/tests@tt4faZQ-zixNhCBTMvRDfs4gMPs

// Package nodediscovery provides tests for the DiscoverNodes function.
//
// Each test creates an isolated temporary directory via t.TempDir(), changes
// the working directory into it with os.Chdir (restoring the original on
// cleanup), and builds a code-from-spec/ tree as needed. This mirrors the
// production contract: DiscoverNodes always resolves paths relative to the
// current working directory.
package nodediscovery

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

// testChdir sets the working directory to dir for the duration of the test.
// It restores the original directory via t.Cleanup so every sub-test is
// isolated even when tests run in parallel or share a process.
func testChdir(t *testing.T, dir string) {
	t.Helper()

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChdir: cannot get current directory: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChdir: cannot chdir to %q: %v", dir, err)
	}

	t.Cleanup(func() {
		// Restore the original directory after each test. Ignore errors here
		// because the temporary directory may already have been cleaned up by
		// the time Cleanup runs (t.TempDir cleanup order is LIFO).
		_ = os.Chdir(original)
	})
}

// testMkNodeFile creates a _node.md file at the given path relative to the
// current directory. All parent directories are created as needed.
func testMkNodeFile(t *testing.T, relPath string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(relPath), 0o755); err != nil {
		t.Fatalf("testMkNodeFile: cannot create directories for %q: %v", relPath, err)
	}

	if err := os.WriteFile(relPath, []byte("# node\n"), 0o644); err != nil {
		t.Fatalf("testMkNodeFile: cannot write %q: %v", relPath, err)
	}
}

// testMkFile creates an arbitrary file at the given path relative to the
// current directory. Used to verify that non-node files are ignored.
func testMkFile(t *testing.T, relPath string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(relPath), 0o755); err != nil {
		t.Fatalf("testMkFile: cannot create directories for %q: %v", relPath, err)
	}

	if err := os.WriteFile(relPath, []byte("content\n"), 0o644); err != nil {
		t.Fatalf("testMkFile: cannot write %q: %v", relPath, err)
	}
}

// testMkDir creates a directory at the given path relative to the current
// directory (and any parents required).
func testMkDir(t *testing.T, relPath string) {
	t.Helper()

	if err := os.MkdirAll(relPath, 0o755); err != nil {
		t.Fatalf("testMkDir: cannot create directory %q: %v", relPath, err)
	}
}

// testNodeNames extracts only the LogicalName fields from a []DiscoveredNode
// slice so table-driven comparisons stay concise.
func testNodeNames(nodes []DiscoveredNode) []string {
	names := make([]string, len(nodes))
	for i, n := range nodes {
		names[i] = n.LogicalName
	}
	return names
}

// testSliceEqual returns true when a and b have identical length and contents.
func testSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ----------------------------------------------------------------------------
// Happy-path tests
// ----------------------------------------------------------------------------

// TestDiscoverNodes_SimpleTree verifies the baseline case: two _node.md files
// at different depths are both discovered, and the result is sorted by logical
// name.
func TestDiscoverNodes_SimpleTree(t *testing.T) {
	// Arrange: isolated temp directory with two nodes.
	dir := t.TempDir()
	testChdir(t, dir)

	testMkNodeFile(t, filepath.Join("code-from-spec", "_node.md"))
	testMkNodeFile(t, filepath.Join("code-from-spec", "sub", "_node.md"))

	// Act
	nodes, err := DiscoverNodes()

	// Assert: no error.
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Assert: exactly two entries.
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d: %v", len(nodes), testNodeNames(nodes))
	}

	// Assert: sorted alphabetically — ROOT comes before ROOT/sub.
	wantNames := []string{"ROOT", "ROOT/sub"}
	gotNames := testNodeNames(nodes)
	if !testSliceEqual(gotNames, wantNames) {
		t.Errorf("logical names mismatch\nwant: %v\n got: %v", wantNames, gotNames)
	}

	// Assert: FilePath values use forward slashes.
	for _, n := range nodes {
		for _, ch := range n.FilePath {
			if ch == '\\' {
				t.Errorf("FilePath %q contains backslash — must use forward slashes", n.FilePath)
			}
		}
	}
}

// TestDiscoverNodes_IgnoresNonNodeFiles checks that files whose name is not
// exactly "_node.md" are silently skipped.
func TestDiscoverNodes_IgnoresNonNodeFiles(t *testing.T) {
	// Arrange: one real node and one non-node file.
	dir := t.TempDir()
	testChdir(t, dir)

	testMkNodeFile(t, filepath.Join("code-from-spec", "_node.md"))
	testMkFile(t, filepath.Join("code-from-spec", "sub", "README.md"))
	testMkFile(t, filepath.Join("code-from-spec", "sub", "node.md"))  // similar but not _node.md
	testMkFile(t, filepath.Join("code-from-spec", "sub", "_node.md.bak")) // backup-style

	// Act
	nodes, err := DiscoverNodes()

	// Assert: no error.
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Assert: only the root node was discovered.
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d: %v", len(nodes), testNodeNames(nodes))
	}

	if nodes[0].LogicalName != "ROOT" {
		t.Errorf("expected logical name ROOT, got %q", nodes[0].LogicalName)
	}
}

// TestDiscoverNodes_SortedByLogicalName places several nodes at varying depths
// in a non-alphabetical order on disk and confirms the output is always sorted
// alphabetically by LogicalName, regardless of filesystem traversal order.
func TestDiscoverNodes_SortedByLogicalName(t *testing.T) {
	// Arrange: nodes whose alphabetical order differs from depth-first order.
	dir := t.TempDir()
	testChdir(t, dir)

	// Create nodes in a deliberately non-sorted physical layout.
	//   ROOT/z         → code-from-spec/z/_node.md
	//   ROOT/a/b       → code-from-spec/a/b/_node.md
	//   ROOT           → code-from-spec/_node.md
	//   ROOT/a         → code-from-spec/a/_node.md
	//   ROOT/m         → code-from-spec/m/_node.md
	testMkNodeFile(t, filepath.Join("code-from-spec", "z", "_node.md"))
	testMkNodeFile(t, filepath.Join("code-from-spec", "a", "b", "_node.md"))
	testMkNodeFile(t, filepath.Join("code-from-spec", "_node.md"))
	testMkNodeFile(t, filepath.Join("code-from-spec", "a", "_node.md"))
	testMkNodeFile(t, filepath.Join("code-from-spec", "m", "_node.md"))

	// Act
	nodes, err := DiscoverNodes()

	// Assert: no error.
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Assert: five nodes found.
	if len(nodes) != 5 {
		t.Fatalf("expected 5 nodes, got %d: %v", len(nodes), testNodeNames(nodes))
	}

	// Assert: exactly this sorted order.
	wantNames := []string{
		"ROOT",
		"ROOT/a",
		"ROOT/a/b",
		"ROOT/m",
		"ROOT/z",
	}
	gotNames := testNodeNames(nodes)
	if !testSliceEqual(gotNames, wantNames) {
		t.Errorf("sorted order mismatch\nwant: %v\n got: %v", wantNames, gotNames)
	}
}

// TestDiscoverNodes_FilePathMatchesLogicalName checks that each DiscoveredNode
// has a FilePath that corresponds to its LogicalName according to the
// code-from-spec/<path>/_node.md convention.
func TestDiscoverNodes_FilePathMatchesLogicalName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testMkNodeFile(t, filepath.Join("code-from-spec", "_node.md"))
	testMkNodeFile(t, filepath.Join("code-from-spec", "x", "y", "_node.md"))

	nodes, err := DiscoverNodes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Map logical name → expected file path (forward slashes).
	wantPaths := map[string]string{
		"ROOT":     "code-from-spec/_node.md",
		"ROOT/x/y": "code-from-spec/x/y/_node.md",
	}

	for _, n := range nodes {
		want, ok := wantPaths[n.LogicalName]
		if !ok {
			t.Errorf("unexpected logical name %q", n.LogicalName)
			continue
		}
		if n.FilePath != want {
			t.Errorf("node %q: want FilePath %q, got %q", n.LogicalName, want, n.FilePath)
		}
	}
}

// ----------------------------------------------------------------------------
// Failure-case tests
// ----------------------------------------------------------------------------

// TestDiscoverNodes_NoDirNotFound verifies that ErrDirNotFound is returned
// when code-from-spec/ does not exist in the working directory.
func TestDiscoverNodes_NoDirNotFound(t *testing.T) {
	// Arrange: temp directory with NO code-from-spec/ subdirectory.
	dir := t.TempDir()
	testChdir(t, dir)
	// Deliberately do not create code-from-spec/.

	// Act
	_, err := DiscoverNodes()

	// Assert: error must be non-nil and wrap ErrDirNotFound.
	if err == nil {
		t.Fatal("expected an error but got nil")
	}

	if !errors.Is(err, ErrDirNotFound) {
		t.Errorf("expected errors.Is(err, ErrDirNotFound) but got: %v", err)
	}

	// Assert: must NOT be ErrWalk or ErrNoNodesFound.
	if errors.Is(err, ErrWalk) {
		t.Errorf("error should not wrap ErrWalk, got: %v", err)
	}
	if errors.Is(err, ErrNoNodesFound) {
		t.Errorf("error should not wrap ErrNoNodesFound, got: %v", err)
	}
}

// TestDiscoverNodes_EmptyDirNoNodesFound verifies that ErrNoNodesFound is
// returned when code-from-spec/ exists but contains no _node.md files.
func TestDiscoverNodes_EmptyDirNoNodesFound(t *testing.T) {
	// Arrange: code-from-spec/ exists but is empty.
	dir := t.TempDir()
	testChdir(t, dir)
	testMkDir(t, "code-from-spec")

	// Act
	_, err := DiscoverNodes()

	// Assert: error must be non-nil and wrap ErrNoNodesFound.
	if err == nil {
		t.Fatal("expected an error but got nil")
	}

	if !errors.Is(err, ErrNoNodesFound) {
		t.Errorf("expected errors.Is(err, ErrNoNodesFound) but got: %v", err)
	}

	// Assert: must NOT be ErrDirNotFound or ErrWalk.
	if errors.Is(err, ErrDirNotFound) {
		t.Errorf("error should not wrap ErrDirNotFound, got: %v", err)
	}
	if errors.Is(err, ErrWalk) {
		t.Errorf("error should not wrap ErrWalk, got: %v", err)
	}
}

// TestDiscoverNodes_EmptyDirWithSubdirsNoNodesFound verifies that even when
// code-from-spec/ contains subdirectories (but no _node.md files),
// ErrNoNodesFound is still returned. Non-node files must not count.
func TestDiscoverNodes_EmptyDirWithSubdirsNoNodesFound(t *testing.T) {
	// Arrange: code-from-spec/ with subdirectories and non-node files.
	dir := t.TempDir()
	testChdir(t, dir)

	testMkFile(t, filepath.Join("code-from-spec", "README.md"))
	testMkFile(t, filepath.Join("code-from-spec", "sub", "README.md"))
	testMkFile(t, filepath.Join("code-from-spec", "sub", "output.md"))

	// Act
	_, err := DiscoverNodes()

	// Assert: ErrNoNodesFound — non-node files do not count.
	if err == nil {
		t.Fatal("expected an error but got nil")
	}

	if !errors.Is(err, ErrNoNodesFound) {
		t.Errorf("expected errors.Is(err, ErrNoNodesFound) but got: %v", err)
	}
}
