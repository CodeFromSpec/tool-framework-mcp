package mcploadchain

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/manifest"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"
)

var (
	ErrNoOutput          = errors.New("target node has no output field")
	ErrInvalidOutputPath = errors.New("output path is invalid")
	ErrArtifactModified  = errors.New("artifact file was modified outside the framework")
)

func MCPLoadChain(logicalName string) (string, error) {
	node, err := parsing.ParseNode(logicalName)
	if err != nil {
		return "", err
	}

	if node.Frontmatter == nil || node.Frontmatter.Output == nil {
		return "", ErrNoOutput
	}

	outputPath := *node.Frontmatter.Output

	if err := oslayer.ValidateStringIsCfsPath(outputPath); err != nil {
		return "", ErrInvalidOutputPath
	}

	artifactLogicalName := "ARTIFACT/" + strings.TrimPrefix(logicalName, "SPEC/")

	m, err := manifest.OpenManifest(true)
	if err == nil {
		if entry, ok := m.Entries[artifactLogicalName]; ok {
			fileChecksum, readErr := computeFileChecksum(oslayer.CfsPath(outputPath))
			if readErr == nil {
				if fileChecksum != entry.Checksum {
					return "", ErrArtifactModified
				}
			}
		}
	}

	chain, err := chainresolver.ChainResolve(logicalName)
	if err != nil {
		return "", err
	}

	chainHash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		return "", err
	}

	var sb strings.Builder

	sb.WriteString("chain_hash: ")
	sb.WriteString(chainHash)
	sb.WriteString("\n")
	sb.WriteString("<chain>\n")

	existingContent, readErr := readFileContent(oslayer.CfsPath(outputPath))
	if readErr == nil {
		sb.WriteString("<existing_artifact>\n")
		sb.WriteString(existingContent)
		sb.WriteString("</existing_artifact>\n")
	}

	sb.WriteString("<constraints>\n")

	for _, ancestor := range chain.Ancestors {
		ancestorNode, parseErr := parsing.ParseNode(ancestor.LogicalName)
		if parseErr != nil {
			return "", parseErr
		}
		if ancestorNode.Public == nil || len(ancestorNode.Public.Subsections) == 0 {
			continue
		}
		content := extractPublicContent(ancestorNode, nil)
		if content == "" {
			continue
		}
		sb.WriteString("<entry name=\"")
		sb.WriteString(ancestor.LogicalName)
		sb.WriteString("\">\n")
		sb.WriteString(content)
		sb.WriteString("</entry>\n")
	}

	for _, dep := range chain.Dependencies {
		switch {
		case strings.HasPrefix(dep.LogicalName, "ARTIFACT/"):
			fileContent, readErr := readFileContent(oslayer.CfsPath(dep.Path))
			if readErr != nil {
				return "", readErr
			}
			sb.WriteString("<entry name=\"")
			sb.WriteString(dep.LogicalName)
			sb.WriteString("\">\n")
			sb.WriteString(fileContent)
			sb.WriteString("</entry>\n")

		case strings.HasPrefix(dep.LogicalName, "EXTERNAL/"):
			fileContent, readErr := readFileContent(oslayer.CfsPath(dep.Path))
			if readErr != nil {
				return "", readErr
			}
			sb.WriteString("<entry name=\"")
			sb.WriteString(dep.LogicalName)
			sb.WriteString("\">\n")
			sb.WriteString(fileContent)
			sb.WriteString("</entry>\n")

		case strings.HasPrefix(dep.LogicalName, "SPEC/"):
			depNode, parseErr := parsing.ParseNode(dep.LogicalName)
			if parseErr != nil {
				return "", parseErr
			}
			content := extractPublicContent(depNode, dep.Qualifier)
			if content == "" {
				continue
			}
			entryName := dep.LogicalName
			if dep.Qualifier != nil {
				entryName = entryName + "(" + *dep.Qualifier + ")"
			}
			sb.WriteString("<entry name=\"")
			sb.WriteString(entryName)
			sb.WriteString("\">\n")
			sb.WriteString(content)
			sb.WriteString("</entry>\n")
		}
	}

	targetNode, err := parsing.ParseNode(chain.Target.LogicalName)
	if err != nil {
		return "", err
	}
	if targetNode.Public != nil && len(targetNode.Public.Subsections) > 0 {
		content := extractPublicContent(targetNode, nil)
		if content != "" {
			sb.WriteString("<entry name=\"")
			sb.WriteString(chain.Target.LogicalName)
			sb.WriteString("\">\n")
			sb.WriteString(content)
			sb.WriteString("</entry>\n")
		}
	}

	sb.WriteString("</constraints>\n")

	if targetNode.Agent != nil {
		agentContent := buildAgentContent(targetNode.Agent)
		sb.WriteString("<instructions>\n")
		sb.WriteString(agentContent)
		sb.WriteString("</instructions>\n")
	}

	if chain.Input != nil {
		sb.WriteString("<input>\n")
		inputContent, inputErr := resolveInputContent(chain.Input)
		if inputErr != nil {
			return "", inputErr
		}
		sb.WriteString(inputContent)
		sb.WriteString("</input>\n")
	}

	sb.WriteString("</chain>\n")

	return sb.String(), nil
}

func computeFileChecksum(cfsPath oslayer.CfsPath) (string, error) {
	handle, err := oslayer.OpenFile(cfsPath, "read", 30000)
	if err != nil {
		return "", err
	}
	defer handle.Close()

	var sb strings.Builder
	for {
		line, err := handle.ReadLine()
		if err != nil {
			if errors.Is(err, oslayer.ErrEndOfFile) {
				break
			}
			return "", err
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	content := sb.String()
	sum := sha1.Sum([]byte(content))
	checksum := base64.RawURLEncoding.EncodeToString(sum[:])
	return checksum, nil
}

func readFileContent(cfsPath oslayer.CfsPath) (string, error) {
	handle, err := oslayer.OpenFile(cfsPath, "read", 30000)
	if err != nil {
		return "", err
	}
	defer handle.Close()

	var sb strings.Builder
	for {
		line, err := handle.ReadLine()
		if err != nil {
			if errors.Is(err, oslayer.ErrEndOfFile) {
				break
			}
			return "", err
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

func extractPublicContent(node *parsing.Node, qualifier *string) string {
	if node.Public == nil {
		return ""
	}

	if qualifier != nil {
		normalizedQualifier := parsing.NormalizeText(*qualifier)
		for _, sub := range node.Public.Subsections {
			if parsing.NormalizeText(sub.Heading) == normalizedQualifier {
				return renderSubsection(sub)
			}
		}
		return ""
	}

	var parts []string
	for _, sub := range node.Public.Subsections {
		parts = append(parts, renderSubsection(sub))
	}
	return strings.Join(parts, "\n")
}

func renderSubsection(sub *parsing.NodeSubsection) string {
	var sb strings.Builder
	sb.WriteString(strings.TrimRight(sub.RawHeading, " \t"))
	sb.WriteString("\n")
	for _, line := range sub.Content {
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	return sb.String()
}

func buildAgentContent(agentSection *parsing.NodeSection) string {
	var blocks []string

	trimmedContent := trimBlankLines(agentSection.Content)
	if len(trimmedContent) > 0 {
		var sb strings.Builder
		for _, line := range trimmedContent {
			sb.WriteString(line)
			sb.WriteString("\n")
		}
		blocks = append(blocks, sb.String())
	}

	for _, sub := range agentSection.Subsections {
		var sb strings.Builder
		sb.WriteString(strings.TrimRight(sub.RawHeading, " \t"))
		sb.WriteString("\n")
		for _, line := range sub.Content {
			sb.WriteString(line)
			sb.WriteString("\n")
		}
		blocks = append(blocks, sb.String())
	}

	result := strings.Join(blocks, "\n")
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return result
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

func resolveInputContent(ref *parsing.CfsReference) (string, error) {
	switch {
	case strings.HasPrefix(ref.LogicalName, "ARTIFACT/"):
		return readFileContent(oslayer.CfsPath(ref.Path))
	case strings.HasPrefix(ref.LogicalName, "EXTERNAL/"):
		return readFileContent(oslayer.CfsPath(ref.Path))
	case strings.HasPrefix(ref.LogicalName, "SPEC/"):
		inputNode, err := parsing.ParseNode(ref.LogicalName)
		if err != nil {
			return "", err
		}
		return extractPublicContent(inputNode, ref.Qualifier), nil
	}
	return "", nil
}
