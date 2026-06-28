// code-from-spec: SPEC/golang/implementation/mcp_tools/chain_hash@oUGFSHF0qQmG240rxqBHBuSkJvs
package mcpchainhash

import (
	"errors"
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

var ErrNoOutput = errors.New("no output")

func MCPChainHash(logicalName string) (string, error) {
	ln, err := logicalnames.LogicalNameParse(logicalName)
	if err != nil {
		return "", fmt.Errorf("parsing logical name: %w", err)
	}

	fm, err := frontmatter.FrontmatterParse(pathutils.PathCfs{Value: ln.Path})
	if err != nil {
		return "", fmt.Errorf("parsing frontmatter: %w", err)
	}

	if fm.Output == "" {
		return "", ErrNoOutput
	}

	chain, err := chainresolver.ChainResolve(logicalName)
	if err != nil {
		return "", fmt.Errorf("resolving chain: %w", err)
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		return "", fmt.Errorf("computing chain hash: %w", err)
	}

	return hash, nil
}
