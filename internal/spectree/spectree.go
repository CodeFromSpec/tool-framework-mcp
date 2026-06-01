// code-from-spec: ROOT/golang/implementation/spec_tree/scan@ZW_XvPVmmFZk0cc5Knxak3pHhmo

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

// ErrNoNodesFound is returned when no _node.md files are found
// under the code-from-spec/ directory.
var ErrNoNodesFound = errors.New("no _node.md files found under code-from-spec/")

// SpecTreeNode represents a single node discovered in the spec tree.
// Each node corresponds to a _node.md file found under code-from-spec/.
type SpecTreeNode struct {
	// LogicalName is the canonical logical name of the node
	// (e.g. "ROOT/functional/logic/os/file_reader").
	LogicalName string

	// FilePath is the CFS-format path to the _node.md file
	// (e.g. "code-from-spec/functional/logic/os/file_reader/_node.md").
	FilePath pathutils.PathCfs
}

// SpecTreeScan scans the code-from-spec/ directory relative to the
// project root and returns all discovered spec nodes sorted
// alphabetically by logical name.
//
// Each node corresponds to a _node.md file. The logical name is
// derived from its path relative to the code-from-spec/ directory.
//
// Errors:
//   - ErrNoNodesFound: no _node.md files found under code-from-spec/.
//   - (ListFiles.*): propagated from the internal file listing operation.
//   - (LogicalNames.*): propagated from LogicalNameFromPath.
func SpecTreeScan() ([]*SpecTreeNode, error) {
	// Step 1: List all files under code-from-spec/.
	dir := &pathutils.PathCfs{Value: "code-from-spec/"}
	allFiles, err := listfiles.ListFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("listing files under code-from-spec/: %w", err)
	}

	// Step 2: Filter for _node.md files only.
	var nodePaths []*pathutils.PathCfs
	for _, filePath := range allFiles {
		lastSlash := strings.LastIndex(filePath.Value, "/")
		var fileName string
		if lastSlash >= 0 {
			fileName = filePath.Value[lastSlash+1:]
		} else {
			fileName = filePath.Value
		}
		if fileName == "_node.md" {
			nodePaths = append(nodePaths, filePath)
		}
	}

	// Step 3: Build SpecTreeNode records.
	var nodes []*SpecTreeNode
	for _, filePath := range nodePaths {
		logicalName, err := logicalnames.LogicalNameFromPath(filePath)
		if err != nil {
			return nil, fmt.Errorf("deriving logical name from %q: %w", filePath.Value, err)
		}
		nodes = append(nodes, &SpecTreeNode{
			LogicalName: logicalName,
			FilePath:    *filePath,
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
