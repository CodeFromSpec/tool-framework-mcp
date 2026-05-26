// code-from-spec: ROOT/golang/internal/node_discovery/tests@3Ee_FIpvdJdh1pN3pWY7Rt08brU

// Package nodediscovery provides tests for DiscoverNodes.
// Each test uses t.TempDir() for isolation and os.Chdir to set the working
// directory, restoring it via t.Cleanup.
package nodediscovery

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// testChdir changes the working directory to dir and registers a cleanup
// function to restore the original directory when the test ends.
func testChdir(t *testing.T, dir string) {
	t.Helper()

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChdir: failed to get current working directory: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChdir: failed to chdir to %q: %v", dir, err)
	}

	t.Cleanup(func() {
		// Restore the original working directory after the test.
		if err := os.Chdir(original); err != nil {
			// This is a best-effort cleanup; log but don't fail the test.
			t.Logf("testChdir cleanup: failed to restore working directory to %q: %v", original, err)
		}
	})
}

// testMakeNodeFile creates a _node.md file at the given path inside base,
// creating any necessary parent directories.
func testMakeNodeFile(t *testing.T, base, relPath string) {
	t.Helper()

	full := filepath.Join(base, relPath)

	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("testMakeNodeFile: failed to create directories for %q: %v", full, err)
	}

	if err := os.WriteFile(full, []byte("# node\n"), 0o644); err != nil {
		t.Fatalf("testMakeNodeFile: failed to write file %q: %v", full, err)
	}
}

// ---------------------------------------------------------------------------
// Happy-path tests
// ---------------------------------------------------------------------------

// TestDiscoverNodes_SimpleTree verifies that DiscoverNodes returns all
// _node.md files in a two-level tree, sorted by logical name.
func TestDiscoverNodes_SimpleTree(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// Create:
	//   code-from-spec/_node.md          -> logical name "ROOT"
	//   code-from-spec/sub/_node.md      -> logical name "ROOT/sub"
	testMakeNodeFile(t, dir, filepath.Join("code-from-spec", "_node.md"))
	testMakeNodeFile(t, dir, filepath.Join("code-from-spec", "sub", "_node.md"))

	nodes, err := DiscoverNodes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d: %v", len(nodes), nodes)
	}

	// Verify both entries are present and sorted.
	// The root node maps to the logical name that represents the root of the
	// code-from-spec directory. We check relative ordering rather than
	// hard-coding the exact logical names, since the logicalnames package
	// determines the exact string representation.
	for i := 1; i < len(nodes); i++ {
		if nodes[i].LogicalName < nodes[i-1].LogicalName {
			t.Errorf("nodes not sorted: nodes[%d].LogicalName %q < nodes[%d].LogicalName %q",
				i, nodes[i].LogicalName, i-1, nodes[i-1].LogicalName)
		}
	}

	// Each node must have a non-empty FilePath ending in _node.md.
	for _, n := range nodes {
		if n.FilePath == "" {
			t.Errorf("node %q has empty FilePath", n.LogicalName)
		}
		if filepath.Base(n.FilePath) != "_node.md" {
			t.Errorf("node %q: FilePath %q does not end in _node.md", n.LogicalName, n.FilePath)
		}
	}
}

// TestDiscoverNodes_IgnoresNonNodeFiles verifies that files other than
// _node.md (e.g., README.md) are not returned as discovered nodes.
func TestDiscoverNodes_IgnoresNonNodeFiles(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// Create one valid node and one unrelated file.
	testMakeNodeFile(t, dir, filepath.Join("code-from-spec", "_node.md"))

	// Create a non-node file that must be ignored.
	readmePath := filepath.Join(dir, "code-from-spec", "sub", "README.md")
	if err := os.MkdirAll(filepath.Dir(readmePath), 0o755); err != nil {
		t.Fatalf("failed to create sub directory: %v", err)
	}
	if err := os.WriteFile(readmePath, []byte("# readme\n"), 0o644); err != nil {
		t.Fatalf("failed to write README.md: %v", err)
	}

	nodes, err := DiscoverNodes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d: %v", len(nodes), nodes)
	}

	if filepath.Base(nodes[0].FilePath) != "_node.md" {
		t.Errorf("expected FilePath to end in _node.md, got %q", nodes[0].FilePath)
	}
}

// TestDiscoverNodes_SortedByLogicalName verifies that DiscoverNodes always
// returns nodes sorted alphabetically by LogicalName, regardless of the
// filesystem traversal order.
func TestDiscoverNodes_SortedByLogicalName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// Create several nodes at varying depths. The directory names are chosen
	// so that lexicographic order of logical names differs from creation order.
	paths := []string{
		filepath.Join("code-from-spec", "zzz", "_node.md"),
		filepath.Join("code-from-spec", "aaa", "_node.md"),
		filepath.Join("code-from-spec", "mmm", "deep", "_node.md"),
		filepath.Join("code-from-spec", "_node.md"),
	}

	for _, p := range paths {
		testMakeNodeFile(t, dir, p)
	}

	nodes, err := DiscoverNodes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != len(paths) {
		t.Fatalf("expected %d nodes, got %d: %v", len(paths), len(nodes), nodes)
	}

	// Verify strict ascending order.
	for i := 1; i < len(nodes); i++ {
		if nodes[i].LogicalName < nodes[i-1].LogicalName {
			t.Errorf("sort violation at index %d: %q < %q",
				i, nodes[i].LogicalName, nodes[i-1].LogicalName)
		}
	}
}

// ---------------------------------------------------------------------------
// Failure-case tests
// ---------------------------------------------------------------------------

// TestDiscoverNodes_NoDirReturnsErrDirNotFound verifies that DiscoverNodes
// returns an error wrapping ErrDirNotFound when code-from-spec/ does not exist.
func TestDiscoverNodes_NoDirReturnsErrDirNotFound(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// Intentionally do NOT create code-from-spec/.
	_, err := DiscoverNodes()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if !errors.Is(err, ErrDirNotFound) {
		t.Errorf("expected errors.Is(err, ErrDirNotFound), got: %v", err)
	}
}

// TestDiscoverNodes_EmptyDirReturnsErrNoNodesFound verifies that DiscoverNodes
// returns an error wrapping ErrNoNodesFound when code-from-spec/ exists but
// contains no _node.md files.
func TestDiscoverNodes_EmptyDirReturnsErrNoNodesFound(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// Create the directory but populate it only with a non-node file.
	cfsDir := filepath.Join(dir, "code-from-spec")
	if err := os.MkdirAll(cfsDir, 0o755); err != nil {
		t.Fatalf("failed to create code-from-spec directory: %v", err)
	}

	// Write an unrelated file to confirm it is not mistaken for a node.
	if err := os.WriteFile(filepath.Join(cfsDir, "some-file.md"), []byte("# x\n"), 0o644); err != nil {
		t.Fatalf("failed to write dummy file: %v", err)
	}

	_, err := DiscoverNodes()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if !errors.Is(err, ErrNoNodesFound) {
		t.Errorf("expected errors.Is(err, ErrNoNodesFound), got: %v", err)
	}
}
