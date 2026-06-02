// code-from-spec: ROOT/golang/implementation/mcp_tools/load_chain@LFwdhuR-6-EfxdRAFFYrr4WFG6Y
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
	targetFilePath, err := logicalnames.LogicalNameToPath(logical_name)
	if err != nil {
		return nil, fmt.Errorf("MCPLoadChain: %w", err)
	}

	fm, err := frontmatter.FrontmatterParse(targetFilePath)
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

	var contextBuilder strings.Builder

	for _, ancestor := range chain.Ancestors {
		node, err := parsenode.NodeParse(ancestor.LogicalName)
		if err != nil {
			return nil, fmt.Errorf("MCPLoadChain: ancestor %s: %w", ancestor.LogicalName, err)
		}
		if node.Public == nil {
			continue
		}
		if len(node.Public.Content) == 0 && len(node.Public.Subsections) == 0 {
			continue
		}
		contextBuilder.WriteString(node.Public.RawHeading)
		contextBuilder.WriteString("\n")
		for _, line := range node.Public.Content {
			contextBuilder.WriteString(line)
			contextBuilder.WriteString("\n")
		}
		for _, sub := range node.Public.Subsections {
			contextBuilder.WriteString(sub.RawHeading)
			contextBuilder.WriteString("\n")
			for _, line := range sub.Content {
				contextBuilder.WriteString(line)
				contextBuilder.WriteString("\n")
			}
		}
	}

	for _, dep := range chain.Dependencies {
		if logicalnames.LogicalNameIsArtifact(dep.LogicalName) {
			content, err := readFileStripFrontmatter(&dep.FilePath)
			if err != nil {
				return nil, fmt.Errorf("MCPLoadChain: artifact dep %s: %w", dep.LogicalName, err)
			}
			contextBuilder.WriteString(content)
		} else if dep.Qualifier == nil {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return nil, fmt.Errorf("MCPLoadChain: dep %s: %w", dep.LogicalName, err)
			}
			if node.Public == nil {
				continue
			}
			contextBuilder.WriteString(node.Public.RawHeading)
			contextBuilder.WriteString("\n")
			for _, line := range node.Public.Content {
				contextBuilder.WriteString(line)
				contextBuilder.WriteString("\n")
			}
			for _, sub := range node.Public.Subsections {
				contextBuilder.WriteString(sub.RawHeading)
				contextBuilder.WriteString("\n")
				for _, line := range sub.Content {
					contextBuilder.WriteString(line)
					contextBuilder.WriteString("\n")
				}
			}
		} else {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return nil, fmt.Errorf("MCPLoadChain: dep %s: %w", dep.LogicalName, err)
			}
			if node.Public == nil {
				continue
			}
			normalizedQualifier := textnormalization.NormalizeText(*dep.Qualifier)
			for _, sub := range node.Public.Subsections {
				if sub.Heading == normalizedQualifier {
					contextBuilder.WriteString(sub.RawHeading)
					contextBuilder.WriteString("\n")
					for _, line := range sub.Content {
						contextBuilder.WriteString(line)
						contextBuilder.WriteString("\n")
					}
					break
				}
			}
		}
	}

	for _, ext := range chain.External {
		extPath := &pathutils.PathCfs{Value: ext.Path}
		content, err := readFileAllLines(extPath)
		if err != nil {
			return nil, fmt.Errorf("MCPLoadChain: external %s: %w", ext.Path, err)
		}
		contextBuilder.WriteString(content)
	}

	contextBuilder.WriteString("---\n")
	contextBuilder.WriteString("output: ")
	contextBuilder.WriteString(fm.Output)
	contextBuilder.WriteString("\n")
	contextBuilder.WriteString("---\n")

	targetNode, err := parsenode.NodeParse(chain.Target.LogicalName)
	if err != nil {
		return nil, fmt.Errorf("MCPLoadChain: target node: %w", err)
	}

	if targetNode.Public != nil {
		contextBuilder.WriteString(targetNode.Public.RawHeading)
		contextBuilder.WriteString("\n")
		for _, line := range targetNode.Public.Content {
			contextBuilder.WriteString(line)
			contextBuilder.WriteString("\n")
		}
		for _, sub := range targetNode.Public.Subsections {
			contextBuilder.WriteString(sub.RawHeading)
			contextBuilder.WriteString("\n")
			for _, line := range sub.Content {
				contextBuilder.WriteString(line)
				contextBuilder.WriteString("\n")
			}
		}
	}

	if targetNode.Agent != nil {
		contextBuilder.WriteString(targetNode.Agent.RawHeading)
		contextBuilder.WriteString("\n")
		for _, line := range targetNode.Agent.Content {
			contextBuilder.WriteString(line)
			contextBuilder.WriteString("\n")
		}
		for _, sub := range targetNode.Agent.Subsections {
			contextBuilder.WriteString(sub.RawHeading)
			contextBuilder.WriteString("\n")
			for _, line := range sub.Content {
				contextBuilder.WriteString(line)
				contextBuilder.WriteString("\n")
			}
		}
	}

	var inputContent *string
	if chain.Input != nil {
		content, err := readFileStripFrontmatter(&chain.Input.FilePath)
		if err != nil {
			return nil, fmt.Errorf("MCPLoadChain: input: %w", err)
		}
		inputContent = &content
	}

	return &MCPLoadChainResult{
		ChainHash: chainHash,
		Context:   contextBuilder.String(),
		Input:     inputContent,
	}, nil
}

func readFileStripFrontmatter(cfsPath *pathutils.PathCfs) (string, error) {
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		return "", fmt.Errorf("readFileStripFrontmatter: %w", err)
	}
	defer filereader.FileClose(reader)

	var lines []string
	foundOpeningDelimiter := false

	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				break
			}
			return "", fmt.Errorf("readFileStripFrontmatter: %w", err)
		}

		if !foundOpeningDelimiter {
			if line == "---" {
				foundOpeningDelimiter = true
				for {
					inner, err := filereader.FileReadLine(reader)
					if err != nil {
						if errors.Is(err, filereader.ErrEndOfFile) {
							break
						}
						return "", fmt.Errorf("readFileStripFrontmatter: %w", err)
					}
					if inner == "---" {
						break
					}
				}
				continue
			}
			lines = append(lines, line)
			foundOpeningDelimiter = true
		} else {
			lines = append(lines, line)
		}
	}

	var sb strings.Builder
	for _, l := range lines {
		sb.WriteString(l)
		sb.WriteString("\n")
	}
	return sb.String(), nil
}

func readFileAllLines(cfsPath *pathutils.PathCfs) (string, error) {
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		return "", fmt.Errorf("readFileAllLines: %w", err)
	}
	defer filereader.FileClose(reader)

	var sb strings.Builder
	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				break
			}
			return "", fmt.Errorf("readFileAllLines: %w", err)
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	return sb.String(), nil
}
