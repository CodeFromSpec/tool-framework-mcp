// code-from-spec: ROOT/golang/internal/node_discovery/code@L2slOOYxLeXfj7j-t1QsrQgNQvA

// Package nodediscovery walks the code-from-spec/ directory tree and returns
// every _node.md file found, each paired with its logical name. The logical
// name is derived by reverse-resolving the file path through the logicalnames
// package.
package nodediscovery

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
)

// DiscoveredNode holds all information about a single discovered spec node.
type DiscoveredNode struct {
	// LogicalName is the ROOT/ logical name derived from the file path,
	// e.g. "ROOT/golang/internal/node_discovery/code".
	LogicalName string

	// FilePath is the path to the _node.md file relative to the project root,
	// e.g. "code-from-spec/golang/internal/node_discovery/code/_node.md".
	FilePath string
}

// Sentinel errors that callers can match with errors.Is().
var (
	// ErrDirNotFound is returned when code-from-spec/ does not exist.
	ErrDirNotFound = errors.New("directory not found")

	// ErrWalk is returned when a filesystem error is encountered while traversing.
	ErrWalk = errors.New("walk error")

	// ErrNoNodesFound is returned when the walk completes but no _node.md files
	// were found anywhere under code-from-spec/.
	ErrNoNodesFound = errors.New("no nodes found")
)

// rootDir is the top-level directory that DiscoverNodes searches.
// It is defined as a constant here so it is easy to spot and change in tests.
const rootDir = "code-from-spec"

// nodeFileName is the exact file name that identifies a spec node.
const nodeFileName = "_node.md"

// DiscoverNodes walks code-from-spec/ relative to the current working
// directory (which must be the project root) and returns every _node.md
// file found. The returned slice is sorted alphabetically by LogicalName.
//
// Errors:
//   - ErrDirNotFound  — code-from-spec/ does not exist.
//   - ErrWalk         — a filesystem error occurred during traversal.
//   - ErrNoNodesFound — the walk completed but no _node.md files were found.
func DiscoverNodes() ([]DiscoveredNode, error) {
	// Step 1-2: Verify that the root directory exists before walking it.
	// os.Stat lets us distinguish "not found" from other I/O errors cleanly.
	info, err := os.Stat(rootDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%w: %s", ErrDirNotFound, rootDir)
		}
		// Some other I/O error (permissions, etc.) — treat as walk error.
		return nil, fmt.Errorf("%w: stat %s: %v", ErrWalk, rootDir, err)
	}
	if !info.IsDir() {
		// The path exists but is not a directory.
		return nil, fmt.Errorf("%w: %s is not a directory", ErrDirNotFound, rootDir)
	}

	// Step 3: Prepare the collector slice.
	var collected []DiscoveredNode

	// Step 4-5: Walk the directory tree. WalkDir visits directories before
	// their contents, and calls the callback for every entry.
	walkErr := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		// Step 4 (continued): Any filesystem error passed into the callback
		// means the walk encountered a problem — propagate it as ErrWalk.
		if err != nil {
			return fmt.Errorf("%w: %v", ErrWalk, err)
		}

		// Step 5a: Skip directories — we only care about files.
		if d.IsDir() {
			return nil
		}

		// Step 5b: Skip any file whose name is not exactly "_node.md".
		if d.Name() != nodeFileName {
			return nil
		}

		// Step 5c: This is a _node.md file.
		// Normalise the path to use forward slashes so it matches what
		// logicalnames expects regardless of the host OS (Windows uses '\').
		filePath := filepath.ToSlash(path)

		// Derive the logical name from the file path via the logicalnames package.
		// LogicalNameFromPath implements the reverse of PathFromLogicalName.
		logicalName, ok := logicalnames.LogicalNameFromPath(filePath)
		if !ok {
			// The path does not match the expected pattern — treat as a walk error
			// so the caller knows something unexpected happened.
			return fmt.Errorf("%w: cannot derive logical name from path %q", ErrWalk, filePath)
		}

		// Append the discovered node to the collector.
		collected = append(collected, DiscoveredNode{
			LogicalName: logicalName,
			FilePath:    filePath,
		})
		return nil
	})

	if walkErr != nil {
		// Unwrap to check whether we already wrapped with ErrWalk above;
		// if not (e.g. filepath.WalkDir itself failed before calling our
		// callback), wrap it now so callers can always use errors.Is(ErrWalk).
		if !errors.Is(walkErr, ErrWalk) {
			return nil, fmt.Errorf("%w: %v", ErrWalk, walkErr)
		}
		return nil, walkErr
	}

	// Step 6: At least one node must exist.
	if len(collected) == 0 {
		return nil, fmt.Errorf("%w: no %s files found under %s", ErrNoNodesFound, nodeFileName, rootDir)
	}

	// Step 7: Sort alphabetically by LogicalName (ascending, case-sensitive).
	sort.Slice(collected, func(i, j int) bool {
		return collected[i].LogicalName < collected[j].LogicalName
	})

	// Step 8: Return the sorted slice.
	return collected, nil
}
