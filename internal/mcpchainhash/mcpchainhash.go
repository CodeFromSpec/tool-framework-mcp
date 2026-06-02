// code-from-spec: ROOT/golang/implementation/mcp_tools/chain_hash@PdO_k_M4sE8XPh3XPOI9w_yW1WI
package mcpchainhash

import (
	"errors"
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
)

var ErrNoOutput = errors.New("target node has no output field")

func MCPChainHash(logical_name string) (string, error) {
	filePath, err := logicalnames.LogicalNameToPath(logical_name)
	if err != nil {
		return "", fmt.Errorf("MCPChainHash: %w", err)
	}

	fm, err := frontmatter.FrontmatterParse(filePath)
	if err != nil {
		return "", fmt.Errorf("MCPChainHash: %w", err)
	}

	if fm.Output == "" {
		return "", fmt.Errorf("MCPChainHash: %w", ErrNoOutput)
	}

	chain, err := chainresolver.ChainResolve(logical_name)
	if err != nil {
		return "", fmt.Errorf("MCPChainHash: %w", err)
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		return "", fmt.Errorf("MCPChainHash: %w", err)
	}

	return hash, nil
}
