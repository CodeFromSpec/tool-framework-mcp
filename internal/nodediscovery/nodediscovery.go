// code-from-spec: ROOT/golang/internal/node_discovery/code@PENDING
package nodediscovery

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
)

// DiscoveredNode pairs a logical name with its file path.
type DiscoveredNode struct {
	LogicalName string
	FilePath    string
}

// Sentinel errors for DiscoverNodes.
var (
	ErrDirNotFound  = errors.New("directory not found")
	ErrWalk         = errors.New("walk error")
	ErrNoNodesFound = errors.New("no nodes found")
)

// DiscoverNodes walks the code-from-spec/ directory relative to the working
// directory and returns every _node.md file found, with its logical name
// derived via logicalnames reverse resolution. The returned slice is sorted
// alphabetically by logical name.
func DiscoverNodes() ([]DiscoveredNode, error) {
	const specDir = "code-from-spec"

	// Check that code-from-spec/ exists.
	info, err := os.Stat(specDir)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("%w: %s", ErrDirNotFound, specDir)
	}

	var nodes []DiscoveredNode

	walkErr := filepath.WalkDir(specDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%w: %s", ErrWalk, err)
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() == "_node.md" {
			// Normalize to forward slashes.
			normalized := filepath.ToSlash(path)
			logicalName, ok := logicalnames.LogicalNameFromPath(normalized)
			if ok {
				nodes = append(nodes, DiscoveredNode{
					LogicalName: logicalName,
					FilePath:    normalized,
				})
			}
		}
		return nil
	})
	if walkErr != nil {
		// If the error is already wrapped with ErrWalk, return as-is.
		if errors.Is(walkErr, ErrWalk) {
			return nil, walkErr
		}
		return nil, fmt.Errorf("%w: %s", ErrWalk, walkErr)
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("%w", ErrNoNodesFound)
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].LogicalName < nodes[j].LogicalName
	})

	return nodes, nil
}
