// code-from-spec: SPEC/golang/implementation/spec_tree/scan@msbGllC7OUA75o1CEHkq_RTE9rM
package spectree

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"
)

var ErrNoNodesFound = errors.New("no nodes found")

func SpecTreeScan() ([]parsing.CfsReference, error) {
	files, err := oslayer.ListAllFiles("code-from-spec/")
	if err != nil {
		return nil, fmt.Errorf("listing files: %w", err)
	}

	var kept []oslayer.CfsPath
	for _, f := range files {
		path := string(f)

		lastSlash := strings.LastIndex(path, "/")
		var fileName string
		if lastSlash == -1 {
			fileName = path
		} else {
			fileName = path[lastSlash+1:]
		}
		if fileName != "_node.md" {
			continue
		}

		remainder := strings.TrimPrefix(path, "code-from-spec/")
		segments := strings.Split(remainder, "/")
		dirSegments := segments[:len(segments)-1]

		if len(dirSegments) == 0 {
			continue
		}

		excluded := false
		for _, seg := range dirSegments {
			if strings.HasPrefix(seg, ".") {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		kept = append(kept, f)
	}

	var refs []parsing.CfsReference
	for _, f := range kept {
		ref, err := parsing.CfsReferenceFromPath(f)
		if err != nil {
			return nil, fmt.Errorf("resolving reference from %s: %w", f, err)
		}
		refs = append(refs, *ref)
	}

	sort.Slice(refs, func(i, j int) bool {
		return refs[i].LogicalName < refs[j].LogicalName
	})

	if len(refs) == 0 {
		return nil, ErrNoNodesFound
	}

	return refs, nil
}
