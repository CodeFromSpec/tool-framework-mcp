// code-from-spec: ROOT/golang/implementation/mcp_tools/load_chain@3uJ_noKPpTONl-CYzkjIdnW2pV8
package mcploadchain

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
)

var ErrNoOutput = errors.New("no output")
var ErrInvalidOutputPath = errors.New("invalid output path")

type MCPLoadChainResult struct {
	ChainHash string
	Context   string
	Input     *string
}

func MCPLoadChain(logical_name string) (*MCPLoadChainResult, error) {
	nodePath, err := logicalnames.LogicalNameToPath(logical_name)
	if err != nil {
		return nil, fmt.Errorf("MCPLoadChain: %w", err)
	}

	fm, err := frontmatter.FrontmatterParse(nodePath)
	if err != nil {
		return nil, fmt.Errorf("MCPLoadChain: %w", err)
	}

	if fm.Output == "" {
		return nil, fmt.Errorf("MCPLoadChain: %w", ErrNoOutput)
	}

	if err := pathutils.PathValidateCfs(fm.Output); err != nil {
		return nil, fmt.Errorf("MCPLoadChain: %w: %w", ErrInvalidOutputPath, err)
	}

	chain, err := chainresolver.ChainResolve(logical_name)
	if err != nil {
		return nil, fmt.Errorf("MCPLoadChain: %w", err)
	}

	chainHash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		return nil, fmt.Errorf("MCPLoadChain: %w", err)
	}

	var sb strings.Builder

	for _, ancestor := range chain.Ancestors {
		node, err := parsenode.NodeParse(ancestor.LogicalName)
		if err != nil {
			return nil, fmt.Errorf("MCPLoadChain ancestor %s: %w", ancestor.LogicalName, err)
		}
		if node.Public == nil {
			continue
		}
		if len(node.Public.Content) == 0 && len(node.Public.Subsections) == 0 {
			continue
		}
		sb.WriteString(node.Public.RawHeading + "\n")
		for _, line := range node.Public.Content {
			sb.WriteString(line + "\n")
		}
		for _, sub := range node.Public.Subsections {
			sb.WriteString(sub.RawHeading + "\n")
			for _, line := range sub.Content {
				sb.WriteString(line + "\n")
			}
		}
	}

	for _, dep := range chain.Dependencies {
		if logicalnames.LogicalNameIsArtifact(dep.LogicalName) {
			content, err := readFileStripFrontmatter(&dep.FilePath)
			if err != nil {
				return nil, fmt.Errorf("MCPLoadChain dependency artifact %s: %w", dep.LogicalName, err)
			}
			sb.WriteString(content)
		} else if dep.Qualifier == "" {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return nil, fmt.Errorf("MCPLoadChain dependency %s: %w", dep.LogicalName, err)
			}
			if node.Public == nil {
				continue
			}
			sb.WriteString(node.Public.RawHeading + "\n")
			for _, line := range node.Public.Content {
				sb.WriteString(line + "\n")
			}
			for _, sub := range node.Public.Subsections {
				sb.WriteString(sub.RawHeading + "\n")
				for _, line := range sub.Content {
					sb.WriteString(line + "\n")
				}
			}
		} else {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return nil, fmt.Errorf("MCPLoadChain dependency %s: %w", dep.LogicalName, err)
			}
			if node.Public == nil {
				continue
			}
			normalizedQualifier := textnormalization.NormalizeText(dep.Qualifier)
			for _, sub := range node.Public.Subsections {
				if sub.Heading == normalizedQualifier {
					sb.WriteString(sub.RawHeading + "\n")
					for _, line := range sub.Content {
						sb.WriteString(line + "\n")
					}
					break
				}
			}
		}
	}

	for _, ext := range chain.External {
		cfsPath := &pathutils.PathCfs{Value: ext.Path}
		content, err := readFileAll(cfsPath)
		if err != nil {
			return nil, fmt.Errorf("MCPLoadChain external %s: %w", ext.Path, err)
		}
		sb.WriteString(content)
	}

	if chain.Target != nil {
		sb.WriteString("---\n")
		sb.WriteString("output: " + fm.Output + "\n")
		sb.WriteString("---\n")

		node, err := parsenode.NodeParse(chain.Target.LogicalName)
		if err != nil {
			return nil, fmt.Errorf("MCPLoadChain target %s: %w", chain.Target.LogicalName, err)
		}

		if node.Public != nil {
			sb.WriteString(node.Public.RawHeading + "\n")
			for _, line := range node.Public.Content {
				sb.WriteString(line + "\n")
			}
			for _, sub := range node.Public.Subsections {
				sb.WriteString(sub.RawHeading + "\n")
				for _, line := range sub.Content {
					sb.WriteString(line + "\n")
				}
			}
		}

		if node.Agent != nil {
			sb.WriteString(node.Agent.RawHeading + "\n")
			for _, line := range node.Agent.Content {
				sb.WriteString(line + "\n")
			}
			for _, sub := range node.Agent.Subsections {
				sb.WriteString(sub.RawHeading + "\n")
				for _, line := range sub.Content {
					sb.WriteString(line + "\n")
				}
			}
		}
	}

	var inputPtr *string
	if chain.Input != nil {
		inputContent, err := readFileStripFrontmatter(&chain.Input.FilePath)
		if err != nil {
			return nil, fmt.Errorf("MCPLoadChain input: %w", err)
		}
		inputPtr = &inputContent
	}

	return &MCPLoadChainResult{
		ChainHash: chainHash,
		Context:   sb.String(),
		Input:     inputPtr,
	}, nil
}

func readFileAll(cfsPath *pathutils.PathCfs) (string, error) {
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		return "", fmt.Errorf("readFileAll: %w", err)
	}
	defer filereader.FileClose(reader)

	var sb strings.Builder
	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				break
			}
			return "", fmt.Errorf("readFileAll: %w", err)
		}
		sb.WriteString(line + "\n")
	}
	return sb.String(), nil
}

func readFileStripFrontmatter(cfsPath *pathutils.PathCfs) (string, error) {
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		return "", fmt.Errorf("readFileStripFrontmatter: %w", err)
	}
	defer filereader.FileClose(reader)

	firstLine, err := filereader.FileReadLine(reader)
	if err != nil {
		if errors.Is(err, filereader.ErrEndOfFile) {
			return "", nil
		}
		return "", fmt.Errorf("readFileStripFrontmatter: %w", err)
	}

	var sb strings.Builder

	if firstLine != "---" {
		sb.WriteString(firstLine + "\n")
		for {
			line, err := filereader.FileReadLine(reader)
			if err != nil {
				if errors.Is(err, filereader.ErrEndOfFile) {
					break
				}
				return "", fmt.Errorf("readFileStripFrontmatter: %w", err)
			}
			sb.WriteString(line + "\n")
		}
		return sb.String(), nil
	}

	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				return "", nil
			}
			return "", fmt.Errorf("readFileStripFrontmatter: %w", err)
		}
		if line == "---" {
			break
		}
	}

	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				break
			}
			return "", fmt.Errorf("readFileStripFrontmatter: %w", err)
		}
		sb.WriteString(line + "\n")
	}
	return sb.String(), nil
}
