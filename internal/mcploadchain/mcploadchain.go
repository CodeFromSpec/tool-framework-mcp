// code-from-spec: ROOT/golang/implementation/mcp_tools/load_chain@Y0MqXxUDwQin1yTpKFVbPsQUgYI
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

func MCPLoadChain(logical_name string) (string, error) {
	nodePath, err := logicalnames.LogicalNameToPath(logical_name)
	if err != nil {
		return "", fmt.Errorf("resolving logical name: %w", err)
	}

	fm, err := frontmatter.FrontmatterParse(nodePath)
	if err != nil {
		return "", fmt.Errorf("parsing frontmatter: %w", err)
	}

	if fm.Output == "" {
		return "", ErrNoOutput
	}

	if err := pathutils.PathValidateCfs(fm.Output); err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidOutputPath, err)
	}

	chain, err := chainresolver.ChainResolve(logical_name)
	if err != nil {
		return "", fmt.Errorf("resolving chain: %w", err)
	}

	chainHash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		return "", fmt.Errorf("computing chain hash: %w", err)
	}

	contextBlocks, err := buildContextStream(chain, fm)
	if err != nil {
		return "", fmt.Errorf("building context stream: %w", err)
	}

	var sb strings.Builder

	sb.WriteString("chain_hash: ")
	sb.WriteString(chainHash)
	sb.WriteString("\n")
	sb.WriteString("--- context ---\n")

	for i, block := range contextBlocks {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(block)
	}

	if chain.Input != nil {
		sb.WriteString("--- input ---\n")

		inputLines, err := readFileLines(chain.Input.FilePath, logicalnames.LogicalNameIsArtifact(chain.Input.UnqualifiedLogicalName))
		if err != nil {
			return "", fmt.Errorf("reading input file: %w", err)
		}
		for _, line := range inputLines {
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	outputCfsPath := &pathutils.PathCfs{Value: fm.Output}
	existingReader, err := filereader.FileOpen(outputCfsPath)
	if err == nil {
		var existingLines []string
		for {
			line, readErr := filereader.FileReadLine(existingReader)
			if errors.Is(readErr, filereader.ErrEndOfFile) {
				break
			}
			if readErr != nil {
				filereader.FileClose(existingReader)
				return "", fmt.Errorf("reading existing artifact: %w", readErr)
			}
			existingLines = append(existingLines, line)
		}
		filereader.FileClose(existingReader)

		sb.WriteString("--- existing artifact ---\n")
		for _, line := range existingLines {
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}

func buildContextStream(chain *chainresolver.Chain, fm *frontmatter.Frontmatter) ([]string, error) {
	var blocks []string

	for _, ancestor := range chain.Ancestors {
		node, err := parsenode.NodeParse(ancestor.UnqualifiedLogicalName)
		if err != nil {
			return nil, fmt.Errorf("parsing ancestor %s: %w", ancestor.UnqualifiedLogicalName, err)
		}

		if node.Public == nil || len(node.Public.Subsections) == 0 {
			continue
		}

		subsectionBlocks := extractSubsectionBlocks(node.Public.Subsections)
		if len(subsectionBlocks) > 0 {
			blocks = append(blocks, joinBlocksWithBlankLine(subsectionBlocks))
		}
	}

	for _, dep := range chain.Dependencies {
		if logicalnames.LogicalNameIsArtifact(dep.UnqualifiedLogicalName) {
			lines, err := readFileLines(dep.FilePath, true)
			if err != nil {
				return nil, fmt.Errorf("reading artifact dependency %s: %w", dep.UnqualifiedLogicalName, err)
			}
			block := linesToBlock(lines)
			if block != "" {
				blocks = append(blocks, block)
			}
		} else if logicalnames.LogicalNameIsExternal(dep.UnqualifiedLogicalName) {
			lines, err := readFileLines(dep.FilePath, false)
			if err != nil {
				return nil, fmt.Errorf("reading external dependency %s: %w", dep.UnqualifiedLogicalName, err)
			}
			block := linesToBlock(lines)
			if block != "" {
				blocks = append(blocks, block)
			}
		} else if logicalnames.LogicalNameIsSpec(dep.UnqualifiedLogicalName) {
			node, err := parsenode.NodeParse(dep.UnqualifiedLogicalName)
			if err != nil {
				return nil, fmt.Errorf("parsing dependency %s: %w", dep.UnqualifiedLogicalName, err)
			}

			if node.Public == nil || len(node.Public.Subsections) == 0 {
				continue
			}

			if dep.Qualifier == nil {
				subsectionBlocks := extractSubsectionBlocks(node.Public.Subsections)
				if len(subsectionBlocks) > 0 {
					blocks = append(blocks, joinBlocksWithBlankLine(subsectionBlocks))
				}
			} else {
				targetHeading := textnormalization.NormalizeText(*dep.Qualifier)
				var found *parsenode.NodeSubsection
				for _, sub := range node.Public.Subsections {
					if sub.Heading == targetHeading {
						found = sub
						break
					}
				}
				if found != nil {
					block := extractSubsectionBlock(found)
					blocks = append(blocks, block)
				}
			}
		}
	}

	frontmatterBlock := "---\noutput: " + fm.Output + "\n---\n"
	blocks = append(blocks, frontmatterBlock)

	if chain.Target != nil {
		node, err := parsenode.NodeParse(chain.Target.UnqualifiedLogicalName)
		if err != nil {
			return nil, fmt.Errorf("parsing target node: %w", err)
		}

		if node.Public != nil && len(node.Public.Subsections) > 0 {
			subsectionBlocks := extractSubsectionBlocks(node.Public.Subsections)
			if len(subsectionBlocks) > 0 {
				blocks = append(blocks, joinBlocksWithBlankLine(subsectionBlocks))
			}
		}

		if node.Agent != nil {
			agentBlock := buildAgentBlock(node.Agent)
			blocks = append(blocks, agentBlock)
		}
	}

	return blocks, nil
}

func extractSubsectionBlock(sub *parsenode.NodeSubsection) string {
	var sb strings.Builder

	heading := strings.TrimRight(sub.RawHeading, " \t")
	sb.WriteString(heading)
	sb.WriteString("\n")

	content := trimBlankLines(sub.Content)
	for _, line := range content {
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	return sb.String()
}

func extractSubsectionBlocks(subsections []*parsenode.NodeSubsection) []string {
	var blocks []string
	for _, sub := range subsections {
		blocks = append(blocks, extractSubsectionBlock(sub))
	}
	return blocks
}

func joinBlocksWithBlankLine(blocks []string) string {
	if len(blocks) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, block := range blocks {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(block)
	}
	return sb.String()
}

func trimBlankLines(lines []string) []string {
	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}

	end := len(lines)
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}

	return lines[start:end]
}

func readFileLines(filePath pathutils.PathCfs, stripArtifactTag bool) ([]string, error) {
	reader, err := filereader.FileOpen(&filePath)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	defer filereader.FileClose(reader)

	var lines []string
	for {
		line, readErr := filereader.FileReadLine(reader)
		if errors.Is(readErr, filereader.ErrEndOfFile) {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("%w", readErr)
		}

		if stripArtifactTag && isArtifactTagLine(line) {
			continue
		}

		lines = append(lines, line)
	}

	return lines, nil
}

func isArtifactTagLine(line string) bool {
	return strings.Contains(line, "code-from-spec:")
}

func linesToBlock(lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, line := range lines {
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	result := sb.String()
	result = strings.TrimRight(result, "\n")
	result += "\n"

	return result
}

func buildAgentBlock(agentSection *parsenode.NodeSection) string {
	var sb strings.Builder

	heading := strings.TrimRight(agentSection.RawHeading, " \t")
	sb.WriteString(heading)
	sb.WriteString("\n")

	content := trimBlankLines(agentSection.Content)
	for _, line := range content {
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	if len(agentSection.Subsections) > 0 {
		subsectionBlocks := extractSubsectionBlocks(agentSection.Subsections)
		for _, subBlock := range subsectionBlocks {
			sb.WriteString("\n")
			sb.WriteString(subBlock)
		}
	}

	result := sb.String()
	result = strings.TrimRight(result, "\n")
	result += "\n"

	return result
}
