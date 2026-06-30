package mcpreconstructcache

import (
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

func MCPReconstructCache() (string, error) {
	m, err := manifest.OpenManifest(true)
	if err != nil {
		return "", fmt.Errorf("failed to open manifest: %w", err)
	}

	entriesProcessed := 0
	contentWritten := 0
	chainWritten := 0

	for key := range m.Entries {
		specName := "SPEC/" + strings.TrimPrefix(key, "ARTIFACT/")

		node, err := parsing.ParseNode(specName)
		if err != nil {
			continue
		}

		if node.Frontmatter == nil || node.Frontmatter.Output == nil {
			continue
		}

		chain, err := chainresolver.ChainResolve(specName)
		if err != nil {
			continue
		}

		chainHash, positions, err := chainhash.ChainHashCompute(chain)
		if err != nil {
			continue
		}

		for _, position := range positions {
			content, err := extractPositionContent(position)
			if err != nil {
				continue
			}

			wrote, err := writeContentIfNew(position.Hash, content)
			if err != nil {
				continue
			}
			if wrote {
				contentWritten++
			}
		}

		wrote, err := writeChainIfNew(chainHash, positions)
		if err == nil && wrote {
			chainWritten++
		}

		entriesProcessed++
	}

	return fmt.Sprintf(
		"reconstructed cache: %d entries processed, %d content files written, %d chain files written",
		entriesProcessed, contentWritten, chainWritten,
	), nil
}

func extractPositionContent(position chainhash.ContentHash) (string, error) {
	label := position.Label

	if strings.HasPrefix(label, "AGENT[") {
		innerName := label[len("AGENT[") : len(label)-1]
		node, err := parsing.ParseNode(innerName)
		if err != nil {
			return "", fmt.Errorf("failed to parse node %s: %w", innerName, err)
		}
		return parsing.ExtractAgentContent(node), nil
	}

	if strings.HasPrefix(label, "INPUT[") {
		innerLabel := label[len("INPUT[") : len(label)-1]
		return resolveReferenceContent(innerLabel)
	}

	return resolveReferenceContent(label)
}

func resolveReferenceContent(label string) (string, error) {
	if strings.HasPrefix(label, "ARTIFACT/") || strings.HasPrefix(label, "EXTERNAL/") {
		ref, err := parsing.CfsReferenceFromName(label)
		if err != nil {
			return "", fmt.Errorf("failed to resolve reference %s: %w", label, err)
		}
		content, err := parsing.ReadFileContent(oslayer.CfsPath(ref.Path))
		if err != nil {
			return "", fmt.Errorf("failed to read content for %s: %w", label, err)
		}
		return content, nil
	}

	if strings.HasPrefix(label, "SPEC/") {
		logicalName := label
		var qualifier *string
		if strings.HasSuffix(label, ")") {
			if idx := strings.LastIndex(label, "("); idx != -1 {
				q := label[idx+1 : len(label)-1]
				logicalName = label[:idx]
				qualifier = &q
			}
		}

		node, err := parsing.ParseNode(logicalName)
		if err != nil {
			return "", fmt.Errorf("failed to parse node %s: %w", logicalName, err)
		}

		if node.Public == nil {
			return "", fmt.Errorf("node %s has no public section", logicalName)
		}

		if qualifier == nil {
			return parsing.ConcatenateSubsections(node.Public.Subsections), nil
		}

		normQualifier := parsing.NormalizeText(*qualifier)
		for _, sub := range node.Public.Subsections {
			if sub == nil {
				continue
			}
			if sub.Heading == normQualifier {
				return parsing.FormatSection(sub.RawHeading, sub.Content), nil
			}
		}
		return "", fmt.Errorf("no subsection matching qualifier %q in node %s", normQualifier, logicalName)
	}

	return "", fmt.Errorf("unrecognized label %q", label)
}

func writeContentIfNew(contentHash string, content string) (bool, error) {
	existed := true
	_, err := cache.ReadContent(contentHash)
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			existed = false
		}
	}

	if err := cache.WriteContent(contentHash, content); err != nil {
		return false, fmt.Errorf("failed to write content %s: %w", contentHash, err)
	}

	return !existed, nil
}

func writeChainIfNew(chainHash string, positions []chainhash.ContentHash) (bool, error) {
	existed := true
	_, err := cache.ReadChain(chainHash)
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			existed = false
		}
	}

	if err := cache.WriteChain(chainHash, positions); err != nil {
		return false, fmt.Errorf("failed to write chain %s: %w", chainHash, err)
	}

	return !existed, nil
}
