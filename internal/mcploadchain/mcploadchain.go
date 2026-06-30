package mcploadchain

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/cache"
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
		return "", fmt.Errorf("parsing target node: %w", err)
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
		return "", fmt.Errorf("resolving chain: %w", err)
	}

	chainHash, positions, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		return "", fmt.Errorf("computing chain hash: %w", err)
	}

	contentByLabel := make(map[string]string)

	var sb strings.Builder

	sb.WriteString("chain_hash: ")
	sb.WriteString(chainHash)
	sb.WriteString("\n")
	sb.WriteString("<chain>\n")

	existingContent, readErr := parsing.ReadFileContent(oslayer.CfsPath(outputPath))
	if readErr == nil {
		sb.WriteString("<existing_artifact>\n")
		sb.WriteString(existingContent)
		sb.WriteString("</existing_artifact>\n")
	}

	sb.WriteString("<constraints>\n")

	for _, ancestor := range chain.Ancestors {
		ancestorNode, parseErr := parsing.ParseNode(ancestor.LogicalName)
		if parseErr != nil {
			return "", fmt.Errorf("parsing ancestor %s: %w", ancestor.LogicalName, parseErr)
		}
		if ancestorNode.Public == nil || len(ancestorNode.Public.Subsections) == 0 {
			continue
		}
		content := extractPublicContent(ancestorNode, nil)
		if content == "" {
			continue
		}
		contentByLabel[ancestor.LogicalName] = content
		sb.WriteString("<entry name=\"")
		sb.WriteString(ancestor.LogicalName)
		sb.WriteString("\">\n")
		sb.WriteString(content)
		sb.WriteString("</entry>\n")
	}

	for _, dep := range chain.Dependencies {
		switch {
		case strings.HasPrefix(dep.LogicalName, "ARTIFACT/"):
			fileContent, readErr := parsing.ReadFileContent(oslayer.CfsPath(dep.Path))
			if readErr != nil {
				return "", fmt.Errorf("reading dependency %s: %w", dep.LogicalName, readErr)
			}
			contentByLabel[dep.LogicalName] = fileContent
			sb.WriteString("<entry name=\"")
			sb.WriteString(dep.LogicalName)
			sb.WriteString("\">\n")
			sb.WriteString(fileContent)
			sb.WriteString("</entry>\n")

		case strings.HasPrefix(dep.LogicalName, "EXTERNAL/"):
			fileContent, readErr := parsing.ReadFileContent(oslayer.CfsPath(dep.Path))
			if readErr != nil {
				return "", fmt.Errorf("reading dependency %s: %w", dep.LogicalName, readErr)
			}
			contentByLabel[dep.LogicalName] = fileContent
			sb.WriteString("<entry name=\"")
			sb.WriteString(dep.LogicalName)
			sb.WriteString("\">\n")
			sb.WriteString(fileContent)
			sb.WriteString("</entry>\n")

		case strings.HasPrefix(dep.LogicalName, "SPEC/"):
			depNode, parseErr := parsing.ParseNode(dep.LogicalName)
			if parseErr != nil {
				return "", fmt.Errorf("parsing dependency %s: %w", dep.LogicalName, parseErr)
			}
			content := extractPublicContent(depNode, dep.Qualifier)
			if content == "" {
				continue
			}
			entryName := dep.LogicalName
			if dep.Qualifier != nil {
				entryName = entryName + "(" + *dep.Qualifier + ")"
			}
			contentByLabel[entryName] = content
			sb.WriteString("<entry name=\"")
			sb.WriteString(entryName)
			sb.WriteString("\">\n")
			sb.WriteString(content)
			sb.WriteString("</entry>\n")
		}
	}

	targetNode, err := parsing.ParseNode(chain.Target.LogicalName)
	if err != nil {
		return "", fmt.Errorf("parsing target node: %w", err)
	}
	if targetNode.Public != nil && len(targetNode.Public.Subsections) > 0 {
		content := extractPublicContent(targetNode, nil)
		if content != "" {
			contentByLabel[chain.Target.LogicalName] = content
			sb.WriteString("<entry name=\"")
			sb.WriteString(chain.Target.LogicalName)
			sb.WriteString("\">\n")
			sb.WriteString(content)
			sb.WriteString("</entry>\n")
		}
	}

	sb.WriteString("</constraints>\n")

	if targetNode.Agent != nil {
		agentContent := parsing.ExtractAgentContent(targetNode)
		contentByLabel["AGENT["+chain.Target.LogicalName+"]"] = agentContent
		sb.WriteString("<instructions>\n")
		sb.WriteString(agentContent)
		sb.WriteString("</instructions>\n")
	}

	if chain.Input != nil {
		sb.WriteString("<input>\n")
		inputContent, inputErr := resolveInputContent(chain.Input)
		if inputErr != nil {
			return "", fmt.Errorf("resolving input: %w", inputErr)
		}
		inputLabel := "INPUT[" + chain.Input.LogicalName
		if chain.Input.Qualifier != nil {
			inputLabel = inputLabel + "(" + *chain.Input.Qualifier + ")"
		}
		inputLabel = inputLabel + "]"
		contentByLabel[inputLabel] = inputContent
		sb.WriteString(inputContent)
		sb.WriteString("</input>\n")
	}

	sb.WriteString("</chain>\n")

	for _, position := range positions {
		if content, ok := contentByLabel[position.Label]; ok {
			_ = cache.WriteContent(position.Hash, content)
		}
	}
	_ = cache.WriteChain(chainHash, positions)

	return sb.String(), nil
}

func computeFileChecksum(cfsPath oslayer.CfsPath) (string, error) {
	content, err := parsing.ReadFileContent(cfsPath)
	if err != nil {
		return "", fmt.Errorf("reading file %s: %w", cfsPath, err)
	}

	sum := sha1.Sum([]byte(content))
	checksum := base64.RawURLEncoding.EncodeToString(sum[:])
	return checksum, nil
}

func extractPublicContent(node *parsing.Node, qualifier *string) string {
	if node.Public == nil {
		return ""
	}

	if qualifier != nil {
		normalizedQualifier := parsing.NormalizeText(*qualifier)
		for _, sub := range node.Public.Subsections {
			if sub.Heading == normalizedQualifier {
				return parsing.FormatSection(sub.RawHeading, sub.Content)
			}
		}
		return ""
	}

	return parsing.ConcatenateSubsections(node.Public.Subsections)
}

func resolveInputContent(ref *parsing.CfsReference) (string, error) {
	switch {
	case strings.HasPrefix(ref.LogicalName, "ARTIFACT/"):
		return parsing.ReadFileContent(oslayer.CfsPath(ref.Path))
	case strings.HasPrefix(ref.LogicalName, "EXTERNAL/"):
		return parsing.ReadFileContent(oslayer.CfsPath(ref.Path))
	case strings.HasPrefix(ref.LogicalName, "SPEC/"):
		inputNode, err := parsing.ParseNode(ref.LogicalName)
		if err != nil {
			return "", fmt.Errorf("parsing input node %s: %w", ref.LogicalName, err)
		}
		return extractPublicContent(inputNode, ref.Qualifier), nil
	}
	return "", nil
}
