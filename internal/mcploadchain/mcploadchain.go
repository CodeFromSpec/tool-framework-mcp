// code-from-spec: SPEC/golang/implementation/mcp_tools/load_chain@tK1-eqlF_FwNqPgGj-49G1HXagU
package mcploadchain

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/file"
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
			return "", fmt.Errorf("parsing ancestor node %q: %w", ancestor.UnqualifiedLogicalName, err)
		}
		if node.Public == nil || len(node.Public.Subsections) == 0 {
			continue
		}
		block := buildPublicSubsectionsBlock(node.Public.Subsections)
		contextParts = append(contextParts, block)
	}

	for _, dep := range chain.Dependencies {
		if dep == nil {
			continue
		}
		if logicalnames.LogicalNameIsArtifact(dep.UnqualifiedLogicalName) {
			content, err := readFileSkippingArtifactTag(&dep.FilePath)
			if err != nil {
				return "", fmt.Errorf("reading artifact dependency %q: %w", dep.UnqualifiedLogicalName, err)
			}
			contextParts = append(contextParts, content)
		} else if logicalnames.LogicalNameIsExternal(dep.UnqualifiedLogicalName) {
			content, err := readFileAll(&dep.FilePath)
			if err != nil {
				return "", fmt.Errorf("reading external dependency %q: %w", dep.UnqualifiedLogicalName, err)
			}
			contextParts = append(contextParts, content)
		} else if logicalnames.LogicalNameIsSpec(dep.UnqualifiedLogicalName) {
			node, err := parsenode.NodeParse(dep.UnqualifiedLogicalName)
			if err != nil {
				return "", fmt.Errorf("parsing dependency node %q: %w", dep.UnqualifiedLogicalName, err)
			}
			if dep.Qualifier == nil {
				if node.Public == nil || len(node.Public.Subsections) == 0 {
					continue
				}
				block := buildPublicSubsectionsBlock(node.Public.Subsections)
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
				if found != nil {
					block := buildSubsectionBlock(found)
					contextParts = append(contextParts, block)
				}
			}
		}
	}

	if chain.Target != nil {
		frontmatterBlock := "---\noutput: " + fm.Output + "\n---\n"
		contextParts = append(contextParts, frontmatterBlock)

		targetNode, err := parsenode.NodeParse(chain.Target.UnqualifiedLogicalName)
		if err != nil {
			return "", fmt.Errorf("parsing target node: %w", err)
		}

		if targetNode.Public != nil && len(targetNode.Public.Subsections) > 0 {
			block := buildPublicSubsectionsBlock(targetNode.Public.Subsections)
			contextParts = append(contextParts, block)
		}

		if targetNode.Agent != nil {
			agentBlock := buildAgentBlock(targetNode)
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
		sb.WriteString("\n--- input ---\n")
		var inputContent string
		if logicalnames.LogicalNameIsArtifact(chain.Input.UnqualifiedLogicalName) {
			inputContent, err = readFileSkippingArtifactTag(&chain.Input.FilePath)
			if err != nil {
				return "", fmt.Errorf("reading input artifact: %w", err)
			}
		} else {
			inputContent, err = readFileAll(&chain.Input.FilePath)
			if err != nil {
				return "", fmt.Errorf("reading input file: %w", err)
			}
		}
		sb.WriteString(inputContent)
	}

	existingPath := &pathutils.PathCfs{Value: fm.Output}
	existingContent, readErr := readFileAll(existingPath)
	if readErr == nil {
		sb.WriteString("\n--- existing artifact ---\n")
		sb.WriteString(existingContent)
	}

	return sb.String(), nil
}

func readFileAll(cfsPath *pathutils.PathCfs) (string, error) {
	handle, err := file.FileOpen(cfsPath, "read", 30000)
	if err != nil {
		return "", err
	}
	defer file.FileClose(handle)

	var lines []string
	for {
		line, err := file.FileReadLine(handle)
		if errors.Is(err, file.ErrEndOfFile) {
			break
		}
		if err != nil {
			return "", err
		}
		lines = append(lines, line)
	}

	var sb strings.Builder
	for _, line := range lines {
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	return sb.String(), nil
}

func readFileSkippingArtifactTag(cfsPath *pathutils.PathCfs) (string, error) {
	handle, err := file.FileOpen(cfsPath, "read", 30000)
	if err != nil {
		return "", err
	}
	defer file.FileClose(handle)

	var lines []string
	artifactTagSkipped := false
	for {
		line, err := file.FileReadLine(handle)
		if errors.Is(err, file.ErrEndOfFile) {
			break
		}
		if err != nil {
			return "", err
		}
		if !artifactTagSkipped && strings.Contains(line, "code-from-spec:") {
			artifactTagSkipped = true
			continue
		}
		lines = append(lines, line)
	}

	var sb strings.Builder
	for _, line := range lines {
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	return sb.String(), nil
}

func buildSubsectionBlock(sub *parsenode.NodeSubsection) string {
	var sb strings.Builder
	sb.WriteString(strings.TrimRight(sub.RawHeading, " \t"))
	sb.WriteString("\n")

	content := trimLeadingBlankLines(sub.Content)
	content = trimTrailingBlankLines(content)
	for _, line := range content {
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	return sb.String()
}

func buildPublicSubsectionsBlock(subsections []*parsenode.NodeSubsection) string {
	var blocks []string
	for _, sub := range subsections {
		blocks = append(blocks, buildSubsectionBlock(sub))
	}
	return strings.Join(blocks, "\n")
}

func buildAgentBlock(node *parsenode.Node) string {
	agent := node.Agent
	var sb strings.Builder

	sb.WriteString(strings.TrimRight(agent.RawHeading, " \t"))
	sb.WriteString("\n")

	content := trimLeadingBlankLines(agent.Content)
	content = trimTrailingBlankLines(content)
	for _, line := range content {
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	for _, sub := range agent.Subsections {
		sb.WriteString("\n")
		sb.WriteString(strings.TrimRight(sub.RawHeading, " \t"))
		sb.WriteString("\n")

		subContent := trimLeadingBlankLines(sub.Content)
		subContent = trimTrailingBlankLines(subContent)
		for _, line := range subContent {
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	return sb.String()
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
