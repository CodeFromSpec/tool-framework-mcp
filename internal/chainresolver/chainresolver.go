// code-from-spec: ROOT/golang/internal/chain_resolver/code@PENDING
package chainresolver

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
)

// ChainItem represents a single node in the chain with its file path
// and an optional qualifier indicating a specific subsection.
type ChainItem struct {
	LogicalName string
	FilePath    string
	Qualifier   *string
}

// ExternalItem represents an external file referenced by the target node.
type ExternalItem struct {
	Path string
}

// Chain holds the fully resolved chain for a target node.
type Chain struct {
	Ancestors    []ChainItem
	Target       ChainItem
	Dependencies []ChainItem
	External     []ExternalItem
	Input        string
}

// ResolveChain builds the chain for the given target logical name by
// collecting ancestors, dependencies, external references, and input.
func ResolveChain(targetLogicalName string) (*Chain, error) {
	// Step 1: Walk up parents to collect ancestors and target.
	var allItems []ChainItem

	current := targetLogicalName
	for {
		filePath, ok := logicalnames.PathFromLogicalName(current)
		if !ok {
			return nil, fmt.Errorf("cannot resolve logical name: %s", current)
		}
		allItems = append(allItems, ChainItem{
			LogicalName: current,
			FilePath:    filePath,
			Qualifier:   nil,
		})

		hasParent, ok := logicalnames.HasParent(current)
		if !ok || !hasParent {
			break
		}
		parent, ok := logicalnames.ParentLogicalName(current)
		if !ok {
			break
		}
		current = parent
	}

	// Sort all items alphabetically by logical name.
	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].LogicalName < allItems[j].LogicalName
	})

	// The last item (alphabetically) is the target; the rest are ancestors.
	var ancestors []ChainItem
	var target ChainItem
	if len(allItems) > 0 {
		target = allItems[len(allItems)-1]
		ancestors = allItems[:len(allItems)-1]
	}

	// Step 2: Read target frontmatter and process dependencies.
	fm, err := frontmatter.ParseFrontmatter(target.FilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading frontmatter for %s: %w", targetLogicalName, err)
	}

	var dependencies []ChainItem
	for _, dep := range fm.DependsOn {
		// Handle ARTIFACT/ references.
		if logicalnames.IsArtifactRef(dep) {
			nodePath, _, ok := logicalnames.ArtifactRefParts(dep)
			if !ok {
				return nil, fmt.Errorf("cannot resolve logical name: %s", dep)
			}
			// For artifact refs, the file path is the node path.
			dependencies = append(dependencies, ChainItem{
				LogicalName: dep,
				FilePath:    nodePath,
				Qualifier:   nil,
			})
			continue
		}

		depPath, ok := logicalnames.PathFromLogicalName(dep)
		if !ok {
			return nil, fmt.Errorf("cannot resolve logical name: %s", dep)
		}

		// Determine qualifier.
		var qualifier *string
		hasQual, ok := logicalnames.HasQualifier(dep)
		if ok && hasQual {
			q, ok := logicalnames.QualifierName(dep)
			if ok {
				qualifier = &q
			}
		}

		// Verify file exists on disk.
		if _, err := os.Stat(depPath); err != nil {
			return nil, fmt.Errorf("cannot resolve logical name: %s", dep)
		}

		dependencies = append(dependencies, ChainItem{
			LogicalName: dep,
			FilePath:    depPath,
			Qualifier:   qualifier,
		})
	}

	// Sort dependencies by FilePath, then by Qualifier (nil before non-nil).
	sort.Slice(dependencies, func(i, j int) bool {
		if dependencies[i].FilePath != dependencies[j].FilePath {
			return dependencies[i].FilePath < dependencies[j].FilePath
		}
		// nil sorts before non-nil.
		if dependencies[i].Qualifier == nil && dependencies[j].Qualifier != nil {
			return true
		}
		if dependencies[i].Qualifier != nil && dependencies[j].Qualifier == nil {
			return false
		}
		if dependencies[i].Qualifier != nil && dependencies[j].Qualifier != nil {
			return *dependencies[i].Qualifier < *dependencies[j].Qualifier
		}
		return false
	})

	// Step 3: Process external entries.
	var external []ExternalItem
	for _, ext := range fm.External {
		external = append(external, ExternalItem{
			Path: ext.Path,
		})
	}

	// Step 4: Normalize all file paths with filepath.ToSlash.
	for i := range ancestors {
		ancestors[i].FilePath = filepath.ToSlash(ancestors[i].FilePath)
	}
	target.FilePath = filepath.ToSlash(target.FilePath)
	for i := range dependencies {
		dependencies[i].FilePath = filepath.ToSlash(dependencies[i].FilePath)
	}

	// Step 5: Deduplicate dependencies.
	// A nil qualifier subsumes specific qualifiers for the same FilePath.
	dependencies = deduplicateItems(dependencies)

	// Also deduplicate across ancestors and dependencies.
	allChain := append(ancestors, dependencies...)
	allChain = deduplicateItems(allChain)
	// Split back: ancestors come first (same count unless deduplicated).
	if len(ancestors) > 0 {
		// Re-split based on whether items were originally ancestors.
		ancestorSet := make(map[string]bool)
		for _, a := range ancestors {
			ancestorSet[a.LogicalName] = true
		}
		var newAncestors, newDeps []ChainItem
		for _, item := range allChain {
			if ancestorSet[item.LogicalName] {
				newAncestors = append(newAncestors, item)
			} else {
				newDeps = append(newDeps, item)
			}
		}
		ancestors = newAncestors
		dependencies = newDeps
	}

	return &Chain{
		Ancestors:    ancestors,
		Target:       target,
		Dependencies: dependencies,
		External:     external,
		Input:        fm.Input,
	}, nil
}

// deduplicateItems removes duplicate ChainItems. Two items are duplicates
// when they share the same FilePath and Qualifier. A nil qualifier subsumes
// any specific qualifier for the same FilePath.
func deduplicateItems(items []ChainItem) []ChainItem {
	// First pass: identify FilePaths that have a nil-qualifier entry.
	nilQualPaths := make(map[string]bool)
	for _, item := range items {
		if item.Qualifier == nil {
			nilQualPaths[item.FilePath] = true
		}
	}

	type key struct {
		filePath  string
		qualifier string
		hasQual   bool
	}

	seen := make(map[key]bool)
	var result []ChainItem

	for _, item := range items {
		// If a nil-qualifier entry exists for this path, skip non-nil qualifiers.
		if item.Qualifier != nil && nilQualPaths[item.FilePath] {
			continue
		}

		k := key{filePath: item.FilePath, hasQual: item.Qualifier != nil}
		if item.Qualifier != nil {
			k.qualifier = *item.Qualifier
		}

		if seen[k] {
			continue
		}
		seen[k] = true
		result = append(result, item)
	}

	return result
}
