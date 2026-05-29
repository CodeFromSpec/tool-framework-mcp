// code-from-spec: ROOT/golang/implementation/spec_tree/scan@desNcchrAq_mK5BqNlx04C6ty98

package spectree

import (
	"fmt"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/listfiles"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// SpecTreeNode represents a single node discovered in the spec tree.
// Each node corresponds to a _node.md file found under code-from-spec/.
type SpecTreeNode struct {
	// LogicalName is the logical name derived from the node's file path.
	LogicalName string

	// FilePath is the CFS path to the _node.md file.
	FilePath *pathutils.PathCfs
}

// SpecTreeScan scans the code-from-spec/ directory for all _node.md
// files and returns a SpecTreeNode for each one found.
//
// The returned slice is sorted alphabetically by logical name.
//
// Returns an error if:
//   - listing files fails (errors propagated from ListFiles).
//   - deriving a logical name from a path fails (errors propagated
//     from LogicalNameFromPath).
//   - no _node.md files are found under code-from-spec/.
func SpecTreeScan() ([]*SpecTreeNode, error) {
	dir := &pathutils.PathCfs{Value: "code-from-spec"}

	files, err := listfiles.ListFiles(dir)
	if err != nil {
		return nil, err
	}

	var nodes []*SpecTreeNode

	for _, file := range files {
		// Extract the file name: the portion after the last "/".
		lastSlash := strings.LastIndex(file.Value, "/")
		var fileName string
		if lastSlash < 0 {
			fileName = file.Value
		} else {
			fileName = file.Value[lastSlash+1:]
		}

		if fileName != "_node.md" {
			continue
		}

		logicalName, err := logicalnames.LogicalNameFromPath(file)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, &SpecTreeNode{
			LogicalName: logicalName,
			FilePath:    file,
		})
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes found")
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].LogicalName < nodes[j].LogicalName
	})

	return nodes, nil
}
