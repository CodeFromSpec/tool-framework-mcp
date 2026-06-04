// code-from-spec: ROOT/golang/implementation/mcp_tools/load_chain@RlQzIhRbaFbHVyOez94Sck0B_T0
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

func MCPLoadChain(logical_name string) (string, error) {
	nodePath, err := logicalnames.LogicalNameToPath(logical_name)
	if err != nil {
		return "", fmt.Errorf("MCPLoadChain: %w", err)
	}

	fm, err := frontmatter.FrontmatterParse(nodePath)
	if err != nil {
		return "", fmt.Errorf("MCPLoadChain: %w", err)
	}
	if fm.Output == "" {
		return "", fmt.Errorf("MCPLoadChain: %w", ErrNoOutput)
	}
	if err := pathutils.PathValidateCfs(fm.Output); err != nil {
		return "", fmt.Errorf("MCPLoadChain: %w", ErrInvalidOutputPath)
	}

	chain, err := chainresolver.ChainResolve(logical_name)
	if err != nil {
		return "", fmt.Errorf("MCPLoadChain: %w", err)
	}

	chainHash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		return "", fmt.Errorf("MCPLoadChain: %w", err)
	}

	var ctx strings.Builder

	for _, ancestor := range chain.Ancestors {
		node, err := parsenode.NodeParse(ancestor.LogicalName)
		if err != nil {
			return "", fmt.Errorf("MCPLoadChain: %w", err)
		}
		if node.Public == nil || len(node.Public.Subsections) == 0 {
			continue
		}
		for _, sub := range node.Public.Subsections {
			ctx.WriteString(sub.RawHeading + "\n")
			for _, line := range sub.Content {
				ctx.WriteString(line + "\n")
			}
		}
	}

	for _, dep := range chain.Dependencies {
		if logicalnames.LogicalNameIsArtifact(dep.LogicalName) {
			depPath := &pathutils.PathCfs{Value: dep.FilePath.Value}
			reader, err := filereader.FileOpen(depPath)
			if err != nil {
				return "", fmt.Errorf("MCPLoadChain: %w", err)
			}
			lines, err := readAllLines(reader)
			filereader.FileClose(reader)
			if err != nil {
				return "", fmt.Errorf("MCPLoadChain: %w", err)
			}
			lines = stripFrontmatter(lines)
			for _, line := range lines {
				ctx.WriteString(line + "\n")
			}
		} else if dep.Qualifier == "" {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return "", fmt.Errorf("MCPLoadChain: %w", err)
			}
			if node.Public != nil && len(node.Public.Subsections) > 0 {
				for _, sub := range node.Public.Subsections {
					ctx.WriteString(sub.RawHeading + "\n")
					for _, line := range sub.Content {
						ctx.WriteString(line + "\n")
					}
				}
			}
		} else {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return "", fmt.Errorf("MCPLoadChain: %w", err)
			}
			normalizedQualifier := textnormalization.NormalizeText(dep.Qualifier)
			if node.Public != nil {
				for _, sub := range node.Public.Subsections {
					if sub.Heading == normalizedQualifier {
						ctx.WriteString(sub.RawHeading + "\n")
						for _, line := range sub.Content {
							ctx.WriteString(line + "\n")
						}
						break
					}
				}
			}
		}
	}

	for _, ext := range chain.External {
		extPath := &pathutils.PathCfs{Value: ext.Path}
		reader, err := filereader.FileOpen(extPath)
		if err != nil {
			return "", fmt.Errorf("MCPLoadChain: %w", err)
		}
		lines, err := readAllLines(reader)
		filereader.FileClose(reader)
		if err != nil {
			return "", fmt.Errorf("MCPLoadChain: %w", err)
		}
		for _, line := range lines {
			ctx.WriteString(line + "\n")
		}
	}

	if chain.Target != nil {
		ctx.WriteString("---\n")
		ctx.WriteString("output: " + fm.Output + "\n")
		ctx.WriteString("---\n")

		node, err := parsenode.NodeParse(chain.Target.LogicalName)
		if err != nil {
			return "", fmt.Errorf("MCPLoadChain: %w", err)
		}
		if node.Public != nil && len(node.Public.Subsections) > 0 {
			for _, sub := range node.Public.Subsections {
				ctx.WriteString(sub.RawHeading + "\n")
				for _, line := range sub.Content {
					ctx.WriteString(line + "\n")
				}
			}
		}
		if node.Agent != nil {
			ctx.WriteString(node.Agent.RawHeading + "\n")
			for _, line := range node.Agent.Content {
				ctx.WriteString(line + "\n")
			}
			for _, sub := range node.Agent.Subsections {
				ctx.WriteString(sub.RawHeading + "\n")
				for _, line := range sub.Content {
					ctx.WriteString(line + "\n")
				}
			}
		}
	}

	var result strings.Builder
	result.WriteString("chain_hash: " + chainHash + "\n")
	result.WriteString("--- context ---\n")
	result.WriteString(ctx.String())

	if chain.Input != nil {
		result.WriteString("--- input ---\n")
		inputPath := &pathutils.PathCfs{Value: chain.Input.FilePath.Value}
		reader, err := filereader.FileOpen(inputPath)
		if err != nil {
			return "", fmt.Errorf("MCPLoadChain: %w", err)
		}
		lines, err := readAllLines(reader)
		filereader.FileClose(reader)
		if err != nil {
			return "", fmt.Errorf("MCPLoadChain: %w", err)
		}
		lines = stripFrontmatter(lines)
		for _, line := range lines {
			result.WriteString(line + "\n")
		}
	}

	outputPath := &pathutils.PathCfs{Value: fm.Output}
	outputReader, err := filereader.FileOpen(outputPath)
	if err == nil {
		result.WriteString("--- existing artifact ---\n")
		lines, err := readAllLines(outputReader)
		filereader.FileClose(outputReader)
		if err == nil {
			for _, line := range lines {
				result.WriteString(line + "\n")
			}
		}
	}

	return result.String(), nil
}

func readAllLines(reader *filereader.FileReader) ([]string, error) {
	var lines []string
	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				break
			}
			return nil, err
		}
		lines = append(lines, line)
	}
	return lines, nil
}

func stripFrontmatter(lines []string) []string {
	firstNonBlank := -1
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			firstNonBlank = i
			break
		}
	}
	if firstNonBlank == -1 || lines[firstNonBlank] != "---" {
		return lines
	}
	for i := firstNonBlank + 1; i < len(lines); i++ {
		if lines[i] == "---" {
			return lines[i+1:]
		}
	}
	return lines
}
