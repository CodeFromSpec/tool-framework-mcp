// code-from-spec: ROOT/golang/implementation/spec_tree/scan@8sm51EI5DKRsIhauxrp5mNN9idA
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

var ErrNoNodesFound = errors.New("no _node.md files found under code-from-spec/")

type SpecTreeNode struct {
	LogicalName string
	FilePath    pathutils.PathCfs
}

func SpecTreeScan() ([]*SpecTreeNode, error) {
	dir := &pathutils.PathCfs{Value: "code-from-spec/"}

	files, err := listfiles.ListFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("SpecTreeScan: %w", err)
	}

	var nodes []*SpecTreeNode

	for _, f := range files {
		value := f.Value
		lastSlash := strings.LastIndex(value, "/")
		var fileName string
		if lastSlash == -1 {
			fileName = value
		} else {
			fileName = value[lastSlash+1:]
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
