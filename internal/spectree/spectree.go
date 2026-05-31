// code-from-spec: ROOT/golang/implementation/spec_tree/scan@IVvK_w0QFRJiuQzo6KGXLEEhDZ8

package spectree

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/listfiles"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ErrNoNodesFound is returned when no _node.md files are found under code-from-spec/.
var ErrNoNodesFound = errors.New("no nodes found")

// SpecTreeNode represents a single node discovered in the spec tree.
type SpecTreeNode struct {
	// LogicalName is the logical name of the node derived from its path.
	LogicalName string

	// FilePath is the path to the node's _node.md file in CFS format.
	FilePath *pathutils.PathCfs
}

// SpecTreeScan scans the code-from-spec/ directory relative to the project root
// and returns a list of all spec tree nodes found.
//
// Each node corresponds to a _node.md file discovered during the scan.
// The returned list is sorted alphabetically by logical name.
//
// Errors:
//   - ErrNoNodesFound: no _node.md files were found under code-from-spec/.
//   - (ListFiles.*): propagated from ListFiles.
//   - (LogicalNames.*): propagated from LogicalNameFromPath.
func SpecTreeScan() ([]*SpecTreeNode, error) {
	// Step 1: List all files under code-from-spec/.
	dir := &pathutils.PathCfs{Value: "code-from-spec/"}
	allFiles, err := listfiles.ListFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("SpecTreeScan: %w", err)
	}

	// Step 2: Filter to only _node.md files.
	var nodeFiles []*pathutils.PathCfs
	for _, filePath := range allFiles {
		lastSlash := strings.LastIndex(filePath.Value, "/")
		var fileName string
		if lastSlash >= 0 {
			fileName = filePath.Value[lastSlash+1:]
		} else {
			fileName = filePath.Value
		}
		if fileName == "_node.md" {
			nodeFiles = append(nodeFiles, filePath)
		}
	}

	// Step 3: Build SpecTreeNode records.
	var nodes []*SpecTreeNode
	for _, filePath := range nodeFiles {
		logicalName, err := logicalnames.LogicalNameFromPath(filePath)
		if err != nil {
			return nil, fmt.Errorf("SpecTreeScan: %w", err)
		}
		nodes = append(nodes, &SpecTreeNode{
			LogicalName: logicalName,
			FilePath:    filePath,
		})
	}

	// Step 4: Sort alphabetically by logical name.
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].LogicalName < nodes[j].LogicalName
	})

	// Step 5: Return error if no nodes were found.
	if len(nodes) == 0 {
		return nil, ErrNoNodesFound
	}

	// Step 6: Return the sorted list.
	return nodes, nil
}
