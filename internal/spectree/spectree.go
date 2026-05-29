// code-from-spec: ROOT/golang/implementation/spec_tree/scan@NbOf-SIZC3nzUVC7FPteHf_MFXw

package spectree

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/listfiles"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

// SpecTreeNode represents a single node discovered in the spec tree.
// It pairs the node's logical name with the CFS path to its _node.md file.
type SpecTreeNode struct {
	// LogicalName is the framework-level identifier for this node,
	// derived from the path of its _node.md file.
	LogicalName string

	// FilePath is the CFS-format path to the _node.md file for this node,
	// relative to the project root.
	FilePath *pathutils.PathCfs
}

var (
	// ErrNoNodesFound is returned when no _node.md files are found
	// under the code-from-spec/ directory.
	ErrNoNodesFound = errors.New("no nodes found")
)

// SpecTreeScan scans the code-from-spec/ directory relative to the
// project root and returns all discovered spec tree nodes, sorted
// alphabetically by logical name.
//
// Each node corresponds to a _node.md file found in the tree. The
// logical name is derived from the file's path, and the file path is
// stored as a PathCfs.
//
// Possible errors:
//   - ErrNoNodesFound — no _node.md files were found under code-from-spec/
//   - errors propagated from ListFiles
//   - errors propagated from LogicalNameFromPath
func SpecTreeScan() ([]*SpecTreeNode, error) {
	// Step 1: list all files under code-from-spec/
	root := &pathutils.PathCfs{Value: "code-from-spec"}
	allFiles, err := listfiles.ListFiles(root)
	if err != nil {
		return nil, fmt.Errorf("listing files: %w", err)
	}

	// Step 2: filter to only _node.md files
	var nodePaths []*pathutils.PathCfs
	for _, filePath := range allFiles {
		lastSlash := strings.LastIndex(filePath.Value, "/")
		var fileName string
		if lastSlash == -1 {
			fileName = filePath.Value
		} else {
			fileName = filePath.Value[lastSlash+1:]
		}
		if fileName == "_node.md" {
			nodePaths = append(nodePaths, filePath)
		}
	}

	// Step 3: build SpecTreeNode records
	nodes := make([]*SpecTreeNode, 0, len(nodePaths))
	for _, filePath := range nodePaths {
		logicalName, err := logicalnames.LogicalNameFromPath(filePath)
		if err != nil {
			return nil, fmt.Errorf("deriving logical name from %q: %w", filePath.Value, err)
		}
		nodes = append(nodes, &SpecTreeNode{
			LogicalName: logicalName,
			FilePath:    filePath,
		})
	}

	// Step 4: sort alphabetically by logical name
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].LogicalName < nodes[j].LogicalName
	})

	// Step 5: return error if no nodes found
	if len(nodes) == 0 {
		return nil, ErrNoNodesFound
	}

	// Step 6: return the sorted list
	return nodes, nil
}
