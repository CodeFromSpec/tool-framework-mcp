// code-from-spec: ROOT/golang/implementation/mcp_tools/chain_hash@gV6LCfMu26LY46qAA4Xaz4Gw_Nc
package mcpchainhash

import (
	"errors"
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
)

var ErrNoOutput = errors.New("no output")

func MCPChainHash(logical_name string) (string, error) {
	filePath, err := logicalnames.LogicalNameToPath(logical_name)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	fm, err := frontmatter.FrontmatterParse(filePath)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	if fm.Output == "" {
		return "", ErrNoOutput
	}

	chain, err := chainresolver.ChainResolve(logical_name)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	return hash, nil
}
