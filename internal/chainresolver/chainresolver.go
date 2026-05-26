// code-from-spec: ROOT/golang/internal/chain_resolver/code@2JlJSfB8ovKs5RLPE5GbJ9b5CIg

// Package chainresolver assembles the spec chain for a target logical name.
// The chain contains all context a generation subagent needs: ancestor nodes
// (from ROOT down to the target's parent), cross-tree dependencies, external
// file references, and the target node itself, plus an optional input artifact.
package chainresolver

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
)

// ChainItem represents one node in the chain that contributes spec content.
// When Qualifier is nil, the caller uses the node's entire "# Public" section.
// When Qualifier is non-nil, the caller uses only the "## <qualifier>"
// subsection within "# Public".
type ChainItem struct {
	LogicalName string
	FilePath    string
	Qualifier   *string
}

// ExternalItem represents a file referenced via the "external" frontmatter
// field. The caller is responsible for reading and including its content.
type ExternalItem struct {
	Path string
}

// Chain is the complete assembled context for a generation subagent targeting
// a specific node.
type Chain struct {
	// Ancestors holds nodes from ROOT down to (but not including) the target.
	// Each contributes its "# Public" section.
	Ancestors []ChainItem

	// Target is the node being generated.
	// It contributes both its "# Public" and "# Agent" sections.
	Target ChainItem

	// Dependencies holds nodes declared in the target's "depends_on" field,
	// in alphabetical order by FilePath (then Qualifier).
	Dependencies []ChainItem

	// External holds file paths from the target's "external" frontmatter field.
	External []ExternalItem

	// Input is the logical name of the artifact declared in the target's
	// "input" frontmatter field, or empty if none.
	Input string
}

// ResolveChain builds and returns the complete chain for the given target
// logical name. Returns an error if any part of the chain cannot be resolved.
func ResolveChain(targetLogicalName string) (*Chain, error) {
	// -------------------------------------------------------------------------
	// Step 1 — Collect ancestors and target
	// -------------------------------------------------------------------------
	// Walk upward from the target, collecting every logical name in the
	// ancestry chain (including the target itself).
	allNames, err := collectAncestry(targetLogicalName)
	if err != nil {
		return nil, err
	}

	// Sort alphabetically. Because logical names use path-like segments
	// (ROOT < ROOT/a < ROOT/a/b), lexicographic order gives us ancestor order.
	sort.Strings(allNames)

	// Convert each logical name to a ChainItem (Qualifier = nil for ancestors
	// and the target — they contribute their full "# Public" section).
	var allItems []ChainItem
	for _, name := range allNames {
		path, ok := logicalnames.PathFromLogicalName(name)
		if !ok {
			return nil, fmt.Errorf("cannot resolve logical name: %s", name)
		}
		allItems = append(allItems, ChainItem{
			LogicalName: name,
			FilePath:    filepath.ToSlash(path),
			Qualifier:   nil,
		})
	}

	// The last item (deepest path alphabetically) is the target; the rest are
	// ancestors.
	target := allItems[len(allItems)-1]
	ancestors := allItems[:len(allItems)-1]

	// -------------------------------------------------------------------------
	// Step 2 — Dependencies
	// -------------------------------------------------------------------------
	// Parse the target node's frontmatter to read depends_on, external, and
	// input fields.
	fm, err := frontmatter.ParseFrontmatter(target.FilePath)
	if err != nil {
		return nil, fmt.Errorf("parsing frontmatter for %s: %w", targetLogicalName, err)
	}

	var dependencies []ChainItem
	for _, dep := range fm.DependsOn {
		depItem, err := resolveDependency(dep)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, depItem)
	}

	// Sort dependencies alphabetically by FilePath, then by Qualifier.
	// nil Qualifier sorts before any non-nil Qualifier.
	sort.Slice(dependencies, func(i, j int) bool {
		if dependencies[i].FilePath != dependencies[j].FilePath {
			return dependencies[i].FilePath < dependencies[j].FilePath
		}
		// Same file path: nil qualifier sorts first.
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

	// -------------------------------------------------------------------------
	// Step 3 — External files
	// -------------------------------------------------------------------------
	// Collect the file paths from the "external" frontmatter field.
	// The spec says to include the path; fragment details are handled by the
	// caller (load_chain), not here.
	var external []ExternalItem
	for _, ext := range fm.External {
		external = append(external, ExternalItem{Path: filepath.ToSlash(ext.Path)})
	}

	// -------------------------------------------------------------------------
	// Step 4 — Normalize file paths (already done via filepath.ToSlash above)
	// -------------------------------------------------------------------------
	// All ChainItem.FilePath values were passed through filepath.ToSlash when
	// created. Nothing additional to do here.

	// -------------------------------------------------------------------------
	// Step 5 — Deduplicate
	// -------------------------------------------------------------------------
	ancestors = deduplicateChainItems(ancestors)
	dependencies = deduplicateChainItems(dependencies)

	return &Chain{
		Ancestors:    ancestors,
		Target:       target,
		Dependencies: dependencies,
		External:     external,
		Input:        fm.Input,
	}, nil
}

// collectAncestry walks upward from targetLogicalName to ROOT and returns a
// slice containing the target and all its ancestors (logical names only).
func collectAncestry(targetLogicalName string) ([]string, error) {
	names := []string{targetLogicalName}

	current := targetLogicalName
	for {
		parent, ok := logicalnames.ParentLogicalName(current)
		if !ok {
			// No parent — we have reached ROOT or the name is invalid.
			break
		}
		names = append(names, parent)
		current = parent
	}

	return names, nil
}

// resolveDependency converts a single depends_on entry (a logical name that
// may be ROOT/ or ARTIFACT/) into a ChainItem.
func resolveDependency(depLogicalName string) (ChainItem, error) {
	// ARTIFACT/ references are resolved differently: we look up the node path
	// and artifact ID via ArtifactRefParts.
	if logicalnames.IsArtifactRef(depLogicalName) {
		nodePath, artifactID, ok := logicalnames.ArtifactRefParts(depLogicalName)
		if !ok {
			return ChainItem{}, fmt.Errorf("cannot resolve logical name: %s", depLogicalName)
		}
		// Verify the node file exists on disk.
		if _, err := os.Stat(nodePath); err != nil {
			return ChainItem{}, fmt.Errorf("cannot resolve logical name: %s", depLogicalName)
		}
		q := artifactID
		return ChainItem{
			LogicalName: depLogicalName,
			FilePath:    filepath.ToSlash(nodePath),
			Qualifier:   &q,
		}, nil
	}

	// ROOT/ reference: resolve to a file path.
	filePath, ok := logicalnames.PathFromLogicalName(depLogicalName)
	if !ok {
		return ChainItem{}, fmt.Errorf("cannot resolve logical name: %s", depLogicalName)
	}

	// Verify the file exists on disk.
	if _, err := os.Stat(filePath); err != nil {
		return ChainItem{}, fmt.Errorf("cannot resolve logical name: %s", depLogicalName)
	}

	// Determine whether the logical name carries a qualifier (e.g. ROOT/x(y)).
	var qualifier *string
	if hasQ, qok := logicalnames.HasQualifier(depLogicalName); qok && hasQ {
		q, _ := logicalnames.QualifierName(depLogicalName)
		qualifier = &q
	}

	return ChainItem{
		LogicalName: depLogicalName,
		FilePath:    filepath.ToSlash(filePath),
		Qualifier:   qualifier,
	}, nil
}

// deduplicateChainItems removes duplicate entries from a ChainItem slice.
// Two entries are duplicates when they share the same FilePath and Qualifier.
//
// Additionally, if an entry with Qualifier == nil already exists for a given
// FilePath, any entry with the same FilePath and a non-nil Qualifier is
// removed — the full "# Public" section already covers every subsection.
//
// The first occurrence of each entry is kept.
func deduplicateChainItems(items []ChainItem) []ChainItem {
	// Track which (FilePath, Qualifier) pairs we have already emitted,
	// and which FilePaths have been emitted with a nil Qualifier.
	type key struct {
		FilePath  string
		Qualifier string // empty string stands for nil
		HasQual   bool   // distinguishes nil from ""
	}

	seen := make(map[key]struct{})
	// nilQualSeen tracks file paths for which we have already emitted an entry
	// with Qualifier == nil (entire # Public section).
	nilQualSeen := make(map[string]struct{})

	var result []ChainItem
	for _, item := range items {
		// If the full "# Public" for this file is already in the result, any
		// qualified reference to the same file is redundant.
		if item.Qualifier != nil {
			if _, ok := nilQualSeen[item.FilePath]; ok {
				continue
			}
		}

		// Build the deduplication key.
		k := key{FilePath: item.FilePath}
		if item.Qualifier != nil {
			k.Qualifier = *item.Qualifier
			k.HasQual = true
		}

		if _, exists := seen[k]; exists {
			continue
		}

		seen[k] = struct{}{}
		if item.Qualifier == nil {
			nilQualSeen[item.FilePath] = struct{}{}
		}
		result = append(result, item)
	}

	return result
}
