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

	existingContent, existingReadErr := parsing.ReadFileContent(oslayer.CfsPath(outputPath))
	artifactExists := existingReadErr == nil

	var cachedPositions []chainhash.ContentHash
	var cachedChainHash string
	cacheAvailable := false

	if artifactExists && m != nil {
		if entry, ok := m.Entries[artifactLogicalName]; ok {
			cachedChainHash = entry.ChainHash
			if cachedChainHash != "" {
				var readErr error
				cachedPositions, readErr = cache.ReadChain(cachedChainHash)
				if readErr == nil {
					cacheAvailable = true
				}
			}
		}
	}

	currentHashByLabel := make(map[string]string)
	for _, pos := range positions {
		currentHashByLabel[pos.Label] = pos.Hash
	}

	cachedHashByLabel := make(map[string]string)
	if cacheAvailable {
		for _, pos := range cachedPositions {
			cachedHashByLabel[pos.Label] = pos.Hash
		}
	}

	contentByLabel := make(map[string]string)

	var sb strings.Builder

	sb.WriteString("<chain>\n")

	if artifactExists && cacheAvailable {
		prevConstraintsEntries := buildPreviousConstraintsEntries(chain, cachedHashByLabel, currentHashByLabel)
		if len(prevConstraintsEntries) > 0 {
			sb.WriteString("<previous_constraints>\n")
			for _, e := range prevConstraintsEntries {
				sb.WriteString(e)
			}
			sb.WriteString("</previous_constraints>\n")
		}

		prevInstructions := buildPreviousInstructions(logicalName, cachedHashByLabel, currentHashByLabel)
		if prevInstructions != "" {
			sb.WriteString("<previous_instructions>\n")
			sb.WriteString(prevInstructions)
			sb.WriteString("</previous_instructions>\n")
		}

		prevInput := buildPreviousInput(chain, cachedHashByLabel, currentHashByLabel)
		if prevInput != "" {
			sb.WriteString("<previous_input>\n")
			sb.WriteString(prevInput)
			sb.WriteString("</previous_input>\n")
		}
	}

	if artifactExists {
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
		disposition := computeDisposition(ancestor.LogicalName, currentHashByLabel, cachedHashByLabel, artifactExists && cacheAvailable)
		sb.WriteString("<entry name=\"")
		sb.WriteString(ancestor.LogicalName)
		sb.WriteString("\"")
		if disposition != "" {
			sb.WriteString(" disposition=\"")
			sb.WriteString(disposition)
			sb.WriteString("\"")
		}
		sb.WriteString(">\n")
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
			disposition := computeDisposition(dep.LogicalName, currentHashByLabel, cachedHashByLabel, artifactExists && cacheAvailable)
			sb.WriteString("<entry name=\"")
			sb.WriteString(dep.LogicalName)
			sb.WriteString("\"")
			if disposition != "" {
				sb.WriteString(" disposition=\"")
				sb.WriteString(disposition)
				sb.WriteString("\"")
			}
			sb.WriteString(">\n")
			sb.WriteString(fileContent)
			sb.WriteString("</entry>\n")

		case strings.HasPrefix(dep.LogicalName, "EXTERNAL/"):
			fileContent, readErr := parsing.ReadFileContent(oslayer.CfsPath(dep.Path))
			if readErr != nil {
				return "", fmt.Errorf("reading dependency %s: %w", dep.LogicalName, readErr)
			}
			contentByLabel[dep.LogicalName] = fileContent
			disposition := computeDisposition(dep.LogicalName, currentHashByLabel, cachedHashByLabel, artifactExists && cacheAvailable)
			sb.WriteString("<entry name=\"")
			sb.WriteString(dep.LogicalName)
			sb.WriteString("\"")
			if disposition != "" {
				sb.WriteString(" disposition=\"")
				sb.WriteString(disposition)
				sb.WriteString("\"")
			}
			sb.WriteString(">\n")
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
			disposition := computeDisposition(entryName, currentHashByLabel, cachedHashByLabel, artifactExists && cacheAvailable)
			sb.WriteString("<entry name=\"")
			sb.WriteString(entryName)
			sb.WriteString("\"")
			if disposition != "" {
				sb.WriteString(" disposition=\"")
				sb.WriteString(disposition)
				sb.WriteString("\"")
			}
			sb.WriteString(">\n")
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
			disposition := computeDisposition(chain.Target.LogicalName, currentHashByLabel, cachedHashByLabel, artifactExists && cacheAvailable)
			sb.WriteString("<entry name=\"")
			sb.WriteString(chain.Target.LogicalName)
			sb.WriteString("\"")
			if disposition != "" {
				sb.WriteString(" disposition=\"")
				sb.WriteString(disposition)
				sb.WriteString("\"")
			}
			sb.WriteString(">\n")
			sb.WriteString(content)
			sb.WriteString("</entry>\n")
		}
	}

	sb.WriteString("</constraints>\n")

	if targetNode.Agent != nil {
		agentContent := parsing.ExtractAgentContent(targetNode)
		agentLabel := "AGENT[" + chain.Target.LogicalName + "]"
		contentByLabel[agentLabel] = agentContent
		disposition := computeDisposition(agentLabel, currentHashByLabel, cachedHashByLabel, artifactExists && cacheAvailable)
		sb.WriteString("<instructions")
		if disposition != "" {
			sb.WriteString(" disposition=\"")
			sb.WriteString(disposition)
			sb.WriteString("\"")
		}
		sb.WriteString(">\n")
		sb.WriteString(agentContent)
		sb.WriteString("</instructions>\n")
	}

	if chain.Input != nil {
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
		disposition := computeDisposition(inputLabel, currentHashByLabel, cachedHashByLabel, artifactExists && cacheAvailable)
		sb.WriteString("<input")
		if disposition != "" {
			sb.WriteString(" disposition=\"")
			sb.WriteString(disposition)
			sb.WriteString("\"")
		}
		sb.WriteString(">\n")
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

func computeDisposition(label string, currentHashByLabel, cachedHashByLabel map[string]string, diffEnabled bool) string {
	if !diffEnabled {
		return ""
	}
	currentHash, inCurrent := currentHashByLabel[label]
	cachedHash, inCached := cachedHashByLabel[label]
	if !inCurrent {
		return ""
	}
	if !inCached {
		return "added"
	}
	if currentHash == cachedHash {
		return "unchanged"
	}
	return "changed"
}

func buildPreviousConstraintsEntries(chain chainresolver.Chain, cachedHashByLabel, currentHashByLabel map[string]string) []string {
	var entries []string

	for _, ancestor := range chain.Ancestors {
		label := ancestor.LogicalName
		cachedHash, inCached := cachedHashByLabel[label]
		if !inCached {
			continue
		}
		currentHash, inCurrent := currentHashByLabel[label]
		if inCurrent && currentHash == cachedHash {
			continue
		}
		content, readErr := cache.ReadContent(cachedHash)
		if readErr != nil {
			continue
		}
		disposition := "changed"
		if !inCurrent {
			disposition = "removed"
		}
		entries = append(entries, buildPreviousEntry(label, disposition, content))
	}

	for _, dep := range chain.Dependencies {
		entryName := dep.LogicalName
		if dep.Qualifier != nil {
			entryName = entryName + "(" + *dep.Qualifier + ")"
		}
		cachedHash, inCached := cachedHashByLabel[entryName]
		if !inCached {
			continue
		}
		currentHash, inCurrent := currentHashByLabel[entryName]
		if inCurrent && currentHash == cachedHash {
			continue
		}
		content, readErr := cache.ReadContent(cachedHash)
		if readErr != nil {
			continue
		}
		disposition := "changed"
		if !inCurrent {
			disposition = "removed"
		}
		entries = append(entries, buildPreviousEntry(entryName, disposition, content))
	}

	targetLabel := chain.Target.LogicalName
	cachedHash, inCached := cachedHashByLabel[targetLabel]
	if inCached {
		currentHash, inCurrent := currentHashByLabel[targetLabel]
		if !(inCurrent && currentHash == cachedHash) {
			content, readErr := cache.ReadContent(cachedHash)
			if readErr == nil {
				disposition := "changed"
				if !inCurrent {
					disposition = "removed"
				}
				entries = append(entries, buildPreviousEntry(targetLabel, disposition, content))
			}
		}
	}

	for label, cachedHash := range cachedHashByLabel {
		if strings.HasPrefix(label, "AGENT[") || strings.HasPrefix(label, "INPUT[") {
			continue
		}
		if _, inCurrent := currentHashByLabel[label]; inCurrent {
			continue
		}
		inAnchors := false
		for _, ancestor := range chain.Ancestors {
			if ancestor.LogicalName == label {
				inAnchors = true
				break
			}
		}
		if inAnchors {
			continue
		}
		for _, dep := range chain.Dependencies {
			depLabel := dep.LogicalName
			if dep.Qualifier != nil {
				depLabel = depLabel + "(" + *dep.Qualifier + ")"
			}
			if depLabel == label {
				inAnchors = true
				break
			}
		}
		if inAnchors {
			continue
		}
		if label == chain.Target.LogicalName {
			continue
		}
		content, readErr := cache.ReadContent(cachedHash)
		if readErr != nil {
			continue
		}
		entries = append(entries, buildPreviousEntry(label, "removed", content))
	}

	return entries
}

func buildPreviousEntry(label, disposition, content string) string {
	var sb strings.Builder
	sb.WriteString("<entry name=\"")
	sb.WriteString(label)
	sb.WriteString("\" disposition=\"")
	sb.WriteString(disposition)
	sb.WriteString("\">\n")
	sb.WriteString(content)
	sb.WriteString("</entry>\n")
	return sb.String()
}

func buildPreviousInstructions(logicalName string, cachedHashByLabel, currentHashByLabel map[string]string) string {
	agentLabel := "AGENT[" + logicalName + "]"
	cachedHash, inCached := cachedHashByLabel[agentLabel]
	if !inCached {
		return ""
	}
	currentHash, inCurrent := currentHashByLabel[agentLabel]
	if inCurrent && currentHash == cachedHash {
		return ""
	}
	content, readErr := cache.ReadContent(cachedHash)
	if readErr != nil {
		return ""
	}
	disposition := "changed"
	if !inCurrent {
		disposition = "removed"
	}
	var sb strings.Builder
	sb.WriteString("<instructions disposition=\"")
	sb.WriteString(disposition)
	sb.WriteString("\">\n")
	sb.WriteString(content)
	sb.WriteString("</instructions>\n")
	return sb.String()
}

func buildPreviousInput(chain chainresolver.Chain, cachedHashByLabel, currentHashByLabel map[string]string) string {
	if chain.Input == nil {
		for label, cachedHash := range cachedHashByLabel {
			if !strings.HasPrefix(label, "INPUT[") {
				continue
			}
			if _, inCurrent := currentHashByLabel[label]; inCurrent {
				continue
			}
			content, readErr := cache.ReadContent(cachedHash)
			if readErr != nil {
				continue
			}
			var sb strings.Builder
			sb.WriteString("<input disposition=\"removed\">\n")
			sb.WriteString(content)
			sb.WriteString("</input>\n")
			return sb.String()
		}
		return ""
	}

	inputLabel := "INPUT[" + chain.Input.LogicalName
	if chain.Input.Qualifier != nil {
		inputLabel = inputLabel + "(" + *chain.Input.Qualifier + ")"
	}
	inputLabel = inputLabel + "]"

	cachedHash, inCached := cachedHashByLabel[inputLabel]
	if !inCached {
		return ""
	}
	currentHash, inCurrent := currentHashByLabel[inputLabel]
	if inCurrent && currentHash == cachedHash {
		return ""
	}
	content, readErr := cache.ReadContent(cachedHash)
	if readErr != nil {
		return ""
	}
	disposition := "changed"
	if !inCurrent {
		disposition = "removed"
	}
	var sb strings.Builder
	sb.WriteString("<input disposition=\"")
	sb.WriteString(disposition)
	sb.WriteString("\">\n")
	sb.WriteString(content)
	sb.WriteString("</input>\n")
	return sb.String()
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
