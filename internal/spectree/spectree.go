// code-from-spec: ROOT/golang/implementation/spec_tree/scan@OR7c0GXC1xxstOJ7nlaaWzDNg2U
package spectree

import (
	"fmt"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/listfiles"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

var ErrNoNodesFound = fmt.Errorf("no _node.md files found under code-from-spec/")

type SpecTreeNode struct {
	LogicalName string
	FilePath    pathutils.PathCfs
}

func SpecTreeScan() ([]*SpecTreeNode, error) {
	root := &pathutils.PathCfs{Value: "code-from-spec/"}

	files, err := listfiles.ListFiles(root)
	if err != nil {
		return nil, fmt.Errorf("SpecTreeScan: %w", err)
	}

	var nodes []*SpecTreeNode
	for _, f := range files {
		lastSlash := strings.LastIndex(f.Value, "/")
		var fileName string
		if lastSlash == -1 {
			fileName = f.Value
		} else {
			fileName = f.Value[lastSlash+1:]
		}
		if fileName != "_node.md" {
			continue
		}

		logicalName, err := logicalnames.LogicalNameFromPath(f)
		if err != nil {
			return nil, fmt.Errorf("SpecTreeScan: %w", err)
		}

		nodes = append(nodes, &SpecTreeNode{
			LogicalName: logicalName,
			FilePath:    *f,
		})
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].LogicalName < nodes[j].LogicalName
	})

	if len(nodes) == 0 {
		return nil, ErrNoNodesFound
	}

	return nodes, nil
}
