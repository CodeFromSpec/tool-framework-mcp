// code-from-spec: SPEC/golang/implementation/mcp_tools/chain_hash@uLsEDPHIi2pyqT9ArUq5ikaB2vY
package mcpchainhash

import (
	"errors"
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/logicalnames"
)

var ErrNoOutput = errors.New("no output")

func MCPChainHash(logicalName string) (string, error) {
	filePath, err := logicalnames.LogicalNameToPath(logicalName)
	if err != nil {
		return "", fmt.Errorf("resolving logical name: %w", err)
	}

	fm, err := frontmatter.FrontmatterParse(filePath)
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
