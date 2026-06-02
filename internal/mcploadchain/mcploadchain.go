// code-from-spec: ROOT/golang/implementation/mcp_tools/load_chain@670Pfe22hfXfrJ1ZwCy32HcWjrg
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

var ErrNoOutput = errors.New("target node has no output field")
var ErrInvalidOutputPath = errors.New("the output path fails path validation")

type MCPLoadChainResult struct {
	ChainHash string
	Context   string
	Input     *string
}

func MCPLoadChain(logical_name string) (*MCPLoadChainResult, error) {
	filePath, err := logicalnames.LogicalNameToPath(logical_name)
	if err != nil {
		return nil, fmt.Errorf("MCPLoadChain: %w", err)
	}

	fm, err := frontmatter.FrontmatterParse(filePath)
	if err != nil {
		return nil, fmt.Errorf("MCPLoadChain: %w", err)
	}

	if fm.Output == "" {
		return nil, fmt.Errorf("MCPLoadChain: %w", ErrNoOutput)
	}

	if err := pathutils.PathValidateCfs(fm.Output); err != nil {
		return nil, fmt.Errorf("MCPLoadChain: %w", ErrInvalidOutputPath)
	}

	chain, err := chainresolver.ChainResolve(logical_name)
	if err != nil {
		return nil, fmt.Errorf("MCPLoadChain: %w", err)
	}

	chainHash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		return nil, fmt.Errorf("MCPLoadChain: %w", err)
	}

	var ctx strings.Builder

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
		ctx.WriteString(node.Public.RawHeading + "\n")
		for _, line := range node.Public.Content {
			ctx.WriteString(line + "\n")
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
			content, err := readFileWithoutFrontmatter(&dep.FilePath)
			if err != nil {
				return nil, fmt.Errorf("MCPLoadChain artifact dep %s: %w", dep.LogicalName, err)
			}
			ctx.WriteString(content)
		} else if dep.Qualifier == "" {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return nil, fmt.Errorf("MCPLoadChain dep %s: %w", dep.LogicalName, err)
			}
			ctx.WriteString(node.Public.RawHeading + "\n")
			for _, line := range node.Public.Content {
				ctx.WriteString(line + "\n")
			}
			for _, sub := range node.Public.Subsections {
				ctx.WriteString(sub.RawHeading + "\n")
				for _, line := range sub.Content {
					ctx.WriteString(line + "\n")
				}
			}
		} else {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return nil, fmt.Errorf("MCPLoadChain dep %s: %w", dep.LogicalName, err)
			}
			normalizedQualifier := textnormalization.NormalizeText(dep.Qualifier)
			var matched *parsenode.NodeSubsection
			for _, sub := range node.Public.Subsections {
				if sub.Heading == normalizedQualifier {
					matched = sub
					break
				}
			}
			if matched != nil {
				ctx.WriteString(matched.RawHeading + "\n")
				for _, line := range matched.Content {
					ctx.WriteString(line + "\n")
				}
			}
		}
	}

	for _, ext := range chain.External {
		cfsPath := &pathutils.PathCfs{Value: ext.Path}
		r, err := filereader.FileOpen(cfsPath)
		if err != nil {
			return nil, fmt.Errorf("MCPLoadChain external %s: %w", ext.Path, err)
		}
		for {
			line, err := filereader.FileReadLine(r)
			if err != nil {
				if errors.Is(err, filereader.ErrEndOfFile) {
					break
				}
				filereader.FileClose(r)
				return nil, fmt.Errorf("MCPLoadChain external %s: %w", ext.Path, err)
			}
			ctx.WriteString(line + "\n")
		}
		filereader.FileClose(r)
	}

	ctx.WriteString("---\n")
	ctx.WriteString("output: " + fm.Output + "\n")
	ctx.WriteString("---\n")

	targetNode, err := parsenode.NodeParse(chain.Target.LogicalName)
	if err != nil {
		return nil, fmt.Errorf("MCPLoadChain target: %w", err)
	}
	ctx.WriteString(targetNode.Public.RawHeading + "\n")
	for _, line := range targetNode.Public.Content {
		ctx.WriteString(line + "\n")
	}
	for _, sub := range targetNode.Public.Subsections {
		ctx.WriteString(sub.RawHeading + "\n")
		for _, line := range sub.Content {
			ctx.WriteString(line + "\n")
		}
	}

	if targetNode.Agent != nil {
		ctx.WriteString(targetNode.Agent.RawHeading + "\n")
		for _, line := range targetNode.Agent.Content {
			ctx.WriteString(line + "\n")
		}
		for _, sub := range targetNode.Agent.Subsections {
			ctx.WriteString(sub.RawHeading + "\n")
			for _, line := range sub.Content {
				ctx.WriteString(line + "\n")
			}
		}
	}

	var inputPtr *string
	if chain.Input != nil {
		content, err := readFileWithoutFrontmatter(&chain.Input.FilePath)
		if err != nil {
			return nil, fmt.Errorf("MCPLoadChain input: %w", err)
		}
		inputPtr = &content
	}

	return &MCPLoadChainResult{
		ChainHash: chainHash,
		Context:   ctx.String(),
		Input:     inputPtr,
	}, nil
}

func readFileWithoutFrontmatter(cfsPath *pathutils.PathCfs) (string, error) {
	r, err := filereader.FileOpen(cfsPath)
	if err != nil {
		return "", fmt.Errorf("readFileWithoutFrontmatter: %w", err)
	}
	defer filereader.FileClose(r)

	var lines []string
	for {
		line, err := filereader.FileReadLine(r)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				break
			}
			return "", fmt.Errorf("readFileWithoutFrontmatter: %w", err)
		}
		lines = append(lines, line)
	}

	if len(lines) > 0 && lines[0] == "---" {
		closeIdx := -1
		for i := 1; i < len(lines); i++ {
			if lines[i] == "---" {
				closeIdx = i
				break
			}
		}
		if closeIdx >= 0 {
			lines = lines[closeIdx+1:]
		}
	}

	var sb strings.Builder
	for _, line := range lines {
		sb.WriteString(line + "\n")
	}
	return sb.String(), nil
}
