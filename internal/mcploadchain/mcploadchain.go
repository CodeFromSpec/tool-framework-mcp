// code-from-spec: SPEC/golang/implementation/mcp_tools/load_chain@wwwtnIv_FapPoH-JqkFI89Le2PA
package mcploadchain

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/textnormalization"
)

var ErrNoOutput = errors.New("no output")
var ErrInvalidOutputPath = errors.New("invalid output path")

func MCPLoadChain(logicalName string) (string, error) {
	targetFilePath, err := logicalnames.LogicalNameToPath(logicalName)
	if err != nil {
		return "", fmt.Errorf("resolving logical name: %w", err)
	}

	fm, err := frontmatter.FrontmatterParse(targetFilePath)
	if err != nil {
		return "", fmt.Errorf("parsing frontmatter: %w", err)
	}
	if fm.Output == "" {
		return "", fmt.Errorf("%w", ErrNoOutput)
	}
	if err := pathutils.PathValidateCfs(fm.Output); err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidOutputPath, err)
	}

	chain, err := chainresolver.ChainResolve(logicalName)
	if err != nil {
		return "", fmt.Errorf("resolving chain: %w", err)
	}

	chainHash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		return "", fmt.Errorf("computing chain hash: %w", err)
	}

	var contextParts []string

	for _, ancestor := range chain.Ancestors {
		if ancestor == nil {
			continue
		}
		node, err := parsenode.NodeParse(ancestor.UnqualifiedLogicalName)
		if err != nil {
			return "", fmt.Errorf("parsing ancestor node %s: %w", ancestor.UnqualifiedLogicalName, err)
		}
		if node.Public == nil || len(node.Public.Subsections) == 0 {
			continue
		}
		block := buildPublicBlock(node.Public)
		contextParts = append(contextParts, block)
	}

	for _, dep := range chain.Dependencies {
		if dep == nil {
			continue
		}
		if logicalnames.LogicalNameIsArtifact(dep.UnqualifiedLogicalName) {
			content, err := readFileContentSkipTag(dep.FilePath)
			if err != nil {
				return "", fmt.Errorf("reading artifact dep %s: %w", dep.UnqualifiedLogicalName, err)
			}
			contextParts = append(contextParts, content)
		} else if logicalnames.LogicalNameIsExternal(dep.UnqualifiedLogicalName) {
			content, err := readFileContent(dep.FilePath)
			if err != nil {
				return "", fmt.Errorf("reading external dep %s: %w", dep.UnqualifiedLogicalName, err)
			}
			contextParts = append(contextParts, content)
		} else if logicalnames.LogicalNameIsSpec(dep.UnqualifiedLogicalName) {
			node, err := parsenode.NodeParse(dep.UnqualifiedLogicalName)
			if err != nil {
				return "", fmt.Errorf("parsing dep node %s: %w", dep.UnqualifiedLogicalName, err)
			}
			if dep.Qualifier == nil {
				if node.Public == nil || len(node.Public.Subsections) == 0 {
					continue
				}
				block := buildPublicBlock(node.Public)
				contextParts = append(contextParts, block)
			} else {
				if node.Public == nil {
					continue
				}
				normalizedQualifier := textnormalization.NormalizeText(*dep.Qualifier)
				var found *parsenode.NodeSubsection
				for _, sub := range node.Public.Subsections {
					if sub.Heading == normalizedQualifier {
						found = sub
						break
					}
				}
				if found == nil {
					continue
				}
				block := buildSubsectionBlock(found)
				contextParts = append(contextParts, block)
			}
		}
	}

	if chain.Target != nil {
		frontmatterBlock := fmt.Sprintf("---\noutput: %s\n---\n", fm.Output)
		contextParts = append(contextParts, frontmatterBlock)

		targetNode, err := parsenode.NodeParse(chain.Target.UnqualifiedLogicalName)
		if err != nil {
			return "", fmt.Errorf("parsing target node: %w", err)
		}

		if targetNode.Public != nil && len(targetNode.Public.Subsections) > 0 {
			block := buildPublicBlock(targetNode.Public)
			contextParts = append(contextParts, block)
		}

		if targetNode.Agent != nil {
			agentBlock := buildAgentBlock(targetNode.Agent)
			contextParts = append(contextParts, agentBlock)
		}
	}

	var sb strings.Builder
	sb.WriteString("chain_hash: ")
	sb.WriteString(chainHash)
	sb.WriteString("\n")
	sb.WriteString("--- context ---\n")
	sb.WriteString(strings.Join(contextParts, "\n"))

	if chain.Input != nil {
		sb.WriteString("--- input ---\n")
		var inputContent string
		if logicalnames.LogicalNameIsArtifact(chain.Input.UnqualifiedLogicalName) {
			inputContent, err = readFileContentSkipTag(chain.Input.FilePath)
			if err != nil {
				return "", fmt.Errorf("reading input artifact: %w", err)
			}
		} else {
			inputContent, err = readFileContent(chain.Input.FilePath)
			if err != nil {
				return "", fmt.Errorf("reading input: %w", err)
			}
		}
		sb.WriteString(inputContent)
	}

	outputCfsPath := &pathutils.PathCfs{Value: fm.Output}
	outputOsPath, pathErr := pathutils.PathCfsToOs(outputCfsPath)
	if pathErr == nil {
		_, statErr := os.Stat(outputOsPath.Value)
		if statErr == nil {
			existingContent, readErr := readFileContent(*outputCfsPath)
			if readErr == nil {
				sb.WriteString("--- existing artifact ---\n")
				sb.WriteString(existingContent)
			}
		}
	}

	return sb.String(), nil
}

func readFileContentSkipTag(cfsPath pathutils.PathCfs) (string, error) {
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		return "", fmt.Errorf("opening file %s: %w", cfsPath.Value, err)
	}
	defer filereader.FileClose(reader)

	var sb strings.Builder
	tagSkipped := false
	for {
		line, err := filereader.FileReadLine(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("reading file %s: %w", cfsPath.Value, err)
		}
		if !tagSkipped && strings.Contains(line, "code-from-spec:") {
			tagSkipped = true
			continue
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	return sb.String(), nil
}

func readFileContent(cfsPath pathutils.PathCfs) (string, error) {
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		return "", fmt.Errorf("opening file %s: %w", cfsPath.Value, err)
	}
	defer filereader.FileClose(reader)

	var sb strings.Builder
	for {
		line, err := filereader.FileReadLine(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("reading file %s: %w", cfsPath.Value, err)
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	return sb.String(), nil
}

func trimLeadingBlankLines(lines []string) []string {
	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	return lines[start:]
}

func trimTrailingBlankLines(lines []string) []string {
	end := len(lines)
	for end > 0 && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	return lines[:end]
}

func buildSubsectionBlock(sub *parsenode.NodeSubsection) string {
	var sb strings.Builder
	sb.WriteString(strings.TrimRight(sub.RawHeading, " \t"))
	sb.WriteString("\n")

	trimmed := trimTrailingBlankLines(trimLeadingBlankLines(sub.Content))
	for _, line := range trimmed {
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	return sb.String()
}

func buildPublicBlock(section *parsenode.NodeSection) string {
	var subsectionBlocks []string
	for _, sub := range section.Subsections {
		subsectionBlocks = append(subsectionBlocks, buildSubsectionBlock(sub))
	}
	return strings.Join(subsectionBlocks, "\n")
}

func buildAgentBlock(section *parsenode.NodeSection) string {
	var sb strings.Builder
	sb.WriteString(strings.TrimRight(section.RawHeading, " \t"))
	sb.WriteString("\n")

	trimmedContent := trimTrailingBlankLines(trimLeadingBlankLines(section.Content))
	for _, line := range trimmedContent {
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	for _, sub := range section.Subsections {
		sb.WriteString("\n")
		sb.WriteString(strings.TrimRight(sub.RawHeading, " \t"))
		sb.WriteString("\n")
		trimmed := trimTrailingBlankLines(trimLeadingBlankLines(sub.Content))
		for _, line := range trimmed {
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	result := sb.String()
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return result
}
