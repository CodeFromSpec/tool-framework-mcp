package mcppruneorphans

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/manifest"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/parsing"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/spectree"
)

func MCPPruneOrphans() (string, error) {
	refs, err := spectree.SpecTreeScan()
	if err != nil && !errors.Is(err, spectree.ErrNoNodesFound) {
		return "", fmt.Errorf("scanning spec tree: %w", err)
	}

	parsedNodes := make(map[string]*parsing.Node)
	for _, ref := range refs {
		node, parseErr := parsing.ParseNode(ref.LogicalName)
		if parseErr != nil {
			continue
		}
		parsedNodes[ref.LogicalName] = node
	}

	m, err := manifest.OpenManifest(false)
	if err != nil {
		return "", fmt.Errorf("opening manifest: %w", err)
	}
	defer func() { _ = m.Discard() }()

	var orphanKeys []string
	for key := range m.Entries {
		generatingNode := "SPEC/" + strings.TrimPrefix(key, "ARTIFACT/")
		node, found := parsedNodes[generatingNode]
		if !found {
			orphanKeys = append(orphanKeys, key)
			continue
		}
		if node.Frontmatter == nil || node.Frontmatter.Type == nil {
			orphanKeys = append(orphanKeys, key)
		}
	}
	sort.Strings(orphanKeys)

	var sb strings.Builder
	handled := make(map[string]bool)

	for _, key := range orphanKeys {
		entry := m.Entries[key]
		path := oslayer.CfsPath(entry.Path)

		f, openErr := oslayer.OpenFile(path, "read", 0)
		if openErr != nil {
			if errors.Is(openErr, oslayer.ErrFileUnreadable) {
				sb.WriteString(fmt.Sprintf("pruned %s (%s) — not found on disk\n", key, entry.Path))
				handled[key] = true
			}
			continue
		}
		f.Close()

		if deleteErr := oslayer.DeleteFile(path); deleteErr != nil {
			continue
		}
		sb.WriteString(fmt.Sprintf("pruned %s (%s) — deleted from disk\n", key, entry.Path))
		handled[key] = true
	}

	for key := range handled {
		delete(m.Entries, key)
	}

	if err := m.Save(); err != nil {
		return "", fmt.Errorf("saving manifest: %w", err)
	}

	sb.WriteString(fmt.Sprintf("pruned orphans: %d entries removed", len(handled)))

	return sb.String(), nil
}
