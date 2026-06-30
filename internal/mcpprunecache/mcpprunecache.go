package mcpprunecache

import (
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/cache"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/manifest"
)

func MCPPruneCache() (string, error) {
	m, err := manifest.OpenManifest(true)
	if err != nil {
		return "", fmt.Errorf("opening manifest: %w", err)
	}

	referencedChains := make(map[string]bool)
	for _, entry := range m.Entries {
		referencedChains[entry.ChainHash] = true
	}

	allChains, err := cache.ListChainHashes()
	if err != nil {
		return "", fmt.Errorf("listing chain hashes: %w", err)
	}

	chainsDeleted := 0
	for _, hash := range allChains {
		if referencedChains[hash] {
			continue
		}
		if err := cache.DeleteChain(hash); err != nil {
			continue
		}
		chainsDeleted++
	}

	referencedContent := make(map[string]bool)
	for hash := range referencedChains {
		positions, err := cache.ReadChain(hash)
		if err != nil {
			continue
		}
		for _, position := range positions {
			referencedContent[position.Hash] = true
		}
	}

	allContent, err := cache.ListContentHashes()
	if err != nil {
		return "", fmt.Errorf("listing content hashes: %w", err)
	}

	contentDeleted := 0
	for _, hash := range allContent {
		if referencedContent[hash] {
			continue
		}
		if err := cache.DeleteContent(hash); err != nil {
			continue
		}
		contentDeleted++
	}

	return fmt.Sprintf("pruned cache: %d chain files deleted, %d content files deleted", chainsDeleted, contentDeleted), nil
}
