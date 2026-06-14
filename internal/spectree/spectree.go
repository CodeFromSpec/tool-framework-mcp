// code-from-spec: ROOT/golang/implementation/spec_tree/scan@uRTaU1GekXBr8Fkq1m6-SXpeD_U
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
	rootDir := &pathutils.PathCfs{Value: "code-from-spec/"}

	allFiles, err := listfiles.ListFiles(rootDir)
	if err != nil {
		return nil, fmt.Errorf("listing files: %w", err)
	}

	var nodeFiles []*pathutils.PathCfs
	for _, f := range allFiles {
		lastSlash := strings.LastIndex(f.Value, "/")
		var fileName string
		if lastSlash < 0 {
			fileName = f.Value
		} else {
			fileName = f.Value[lastSlash+1:]
		}
		if fileName != "_node.md" {
			continue
		}

		remainder := strings.TrimPrefix(f.Value, "code-from-spec/")
		firstSlash := strings.Index(remainder, "/")
		if firstSlash >= 0 {
			firstSegment := remainder[:firstSlash]
			if strings.HasPrefix(firstSegment, "_") {
				continue
			}
		}

		nodeFiles = append(nodeFiles, f)
	}

	var nodes []*SpecTreeNode
	for _, f := range nodeFiles {
		logicalName, err := logicalnames.LogicalNameFromPath(f)
		if err != nil {
			return nil, fmt.Errorf("deriving logical name from %s: %w", f.Value, err)
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
