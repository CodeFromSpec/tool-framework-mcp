// code-from-spec: SPEC/golang/implementation/spec_tree/scan@tiiekRPyfGpBsQubg9n-cKKGkSY
package spectree

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/listfiles"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

var ErrNoNodesFound = errors.New("no _node.md files found under code-from-spec/")

type SpecTreeNode struct {
	LogicalName string
	FilePath    pathutils.PathCfs
}

func SpecTreeScan() ([]*SpecTreeNode, error) {
	dir := &pathutils.PathCfs{Value: "code-from-spec"}

	files, err := listfiles.ListFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("listing files: %w", err)
	}

	var kept []*pathutils.PathCfs
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

		remainder := strings.TrimPrefix(f.Value, "code-from-spec/")
		slashIdx := strings.Index(remainder, "/")
		if slashIdx != -1 {
			firstSegment := remainder[:slashIdx]
			if strings.HasPrefix(firstSegment, "_") {
				continue
			}
		}

		kept = append(kept, f)
	}

	var nodes []*SpecTreeNode
	for _, f := range kept {
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
