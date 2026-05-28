// code-from-spec: ROOT/golang/implementation/utils/spec_tree@T5ZRSkxuUQzVXDugld-X9hovwxQ

package spectree

import (
	"errors"
	"sort"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/listfiles"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

// ErrNoNodesFound is returned when no _node.md files are found
// under the code-from-spec/ directory.
var ErrNoNodesFound = errors.New("no nodes found")

// SpecTreeNode represents a single spec node discovered on disk.
// It pairs the node's logical name (e.g. "ROOT/golang/interfaces/utils/spec_tree")
// with the CFS-format path to its _node.md file.
type SpecTreeNode struct {
	LogicalName string
	FilePath    *pathutils.PathCfs
}

// SpecTreeScan scans the code-from-spec/ directory relative to the
// project root and returns all discovered spec nodes sorted
// alphabetically by logical name.
//
// Each _node.md file found is converted into a SpecTreeNode containing
// its logical name and its CFS file path.
//
// Possible errors:
//   - ErrNoNodesFound — no _node.md files were found under code-from-spec/.
//   - Errors propagated from ListFiles.
//   - Errors propagated from LogicalNameFromPath.
func SpecTreeScan() ([]*SpecTreeNode, error) {
	// Step 1: list all files under code-from-spec/.
	dir := &pathutils.PathCfs{Value: "code-from-spec"}
	allFiles, err := listfiles.ListFiles(dir)
	if err != nil {
		return nil, err
	}

	// Step 2: filter to only _node.md files.
	var nodeFiles []*pathutils.PathCfs
	for _, f := range allFiles {
		value := f.Value
		fileName := value
		if idx := lastSlashIndex(value); idx >= 0 {
			fileName = value[idx+1:]
		}
		if fileName == "_node.md" {
			nodeFiles = append(nodeFiles, f)
		}
	}

	// Step 3: convert each PathCfs to a SpecTreeNode.
	nodes := make([]*SpecTreeNode, 0, len(nodeFiles))
	for _, f := range nodeFiles {
		logicalName, err := logicalnames.LogicalNameFromPath(f)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, &SpecTreeNode{
			LogicalName: logicalName,
			FilePath:    f,
		})
	}

	// Step 4: sort alphabetically by logical name.
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].LogicalName < nodes[j].LogicalName
	})

	// Step 5: return error if no nodes were found.
	if len(nodes) == 0 {
		return nil, ErrNoNodesFound
	}

	// Step 6: return the sorted list.
	return nodes, nil
}

// lastSlashIndex returns the index of the last '/' in s, or -1 if none.
func lastSlashIndex(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '/' {
			return i
		}
	}
	return -1
}
