// code-from-spec: ROOT/golang/implementation/chain/chain_resolver@jXrL6eFeIOcDo6a0VeG4dzwFrTI

package chainresolver

import (
	"errors"
	"fmt"
	"sort"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

var (
	// ErrInvalidLogicalName is returned when the target logical name
	// cannot be converted to a file path or its parent cannot be derived.
	ErrInvalidLogicalName = errors.New("invalid logical name")

	// ErrUnreadableFrontmatter is returned when a node's spec file cannot
	// be opened or its frontmatter cannot be parsed.
	ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")

	// ErrUnresolvableArtifact is returned when an ARTIFACT/ reference
	// specifies an output id that does not appear in the node's declared
	// outputs list.
	ErrUnresolvableArtifact = errors.New("unresolvable artifact")
)

// ChainItem represents a single node in the resolved chain, identified
// by its logical name and the CFS path to its spec file. An optional
// qualifier is set when the item was referenced via an ARTIFACT/ prefix
// and points to a specific output id.
type ChainItem struct {
	LogicalName string
	FilePath    *pathutils.PathCfs
	Qualifier   *string
}

// Chain holds the fully resolved chain for a target node.
//
// Assembly order:
//  1. Ancestors — from root down to (but not including) the target node.
//  2. Dependencies — from the target's depends_on, sorted alphabetically
//     by file path then by qualifier.
//  3. External — from the target's external, sorted alphabetically by path.
//  4. Target — the target node itself.
//  5. Input — the target's input artifact, if present.
type Chain struct {
	Ancestors    []*ChainItem
	Dependencies []*ChainItem
	External     []*frontmatter.FrontmatterExternal
	Target       *ChainItem
	Input        *ChainItem
}

// ChainResolve builds and returns the full Chain for the given target
// logical name.
func ChainResolve(target_logical_name string) (*Chain, error) {
	// Step 1 — Resolve ancestors and target
	ancestors, target, err := resolveAncestorsAndTarget(target_logical_name)
	if err != nil {
		return nil, err
	}

	// Step 2 — Resolve dependencies
	targetFrontmatter, err := frontmatter.FrontmatterParse(target.FilePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnreadableFrontmatter, err)
	}

	deps, err := resolveDependencies(targetFrontmatter)
	if err != nil {
		return nil, err
	}

	// Step 3 — Deduplicate dependencies
	deps = deduplicateDependencies(deps)

	// Step 4 — Collect external
	external := collectExternal(targetFrontmatter)

	// Step 5 — Resolve input
	inputItem, err := resolveInput(targetFrontmatter)
	if err != nil {
		return nil, err
	}

	// Step 6 — Return
	return &Chain{
		Ancestors:    ancestors,
		Dependencies: deps,
		External:     external,
		Target:       target,
		Input:        inputItem,
	}, nil
}

// resolveAncestorsAndTarget builds the ancestors list and target ChainItem.
func resolveAncestorsAndTarget(target_logical_name string) ([]*ChainItem, *ChainItem, error) {
	if target_logical_name == "ROOT" {
		path, err := logicalnames.LogicalNameToPath("ROOT")
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %w", ErrInvalidLogicalName, err)
		}
		item := &ChainItem{
			LogicalName: "ROOT",
			FilePath:    path,
			Qualifier:   nil,
		}
		return []*ChainItem{}, item, nil
	}

	// Collect all logical names from target up to ROOT
	names := []string{target_logical_name}
	current := target_logical_name
	for {
		parent, err := logicalnames.LogicalNameGetParent(current)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %w", ErrInvalidLogicalName, err)
		}
		names = append(names, parent)
		current = parent
		if current == "ROOT" {
			break
		}
	}

	// Sort alphabetically — ROOT first, deepest path last
	sort.Strings(names)

	// Build ChainItems for each name
	items := make([]*ChainItem, 0, len(names))
	for _, name := range names {
		path, err := logicalnames.LogicalNameToPath(name)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %w", ErrInvalidLogicalName, err)
		}
		items = append(items, &ChainItem{
			LogicalName: name,
			FilePath:    path,
			Qualifier:   nil,
		})
	}

	// Last item is the target; all preceding are ancestors
	target := items[len(items)-1]
	ancestors := items[:len(items)-1]

	return ancestors, target, nil
}

// resolveDependencies resolves each entry in depends_on into ChainItems,
// then sorts them by file path then qualifier.
func resolveDependencies(fm *frontmatter.Frontmatter) ([]*ChainItem, error) {
	var deps []*ChainItem

	for _, entry := range fm.DependsOn {
		if entry == nil {
			continue
		}
		ref := *entry

		if logicalnames.LogicalNameIsArtifact(ref) {
			item, err := resolveArtifactRef(ref)
			if err != nil {
				return nil, err
			}
			deps = append(deps, item)
		} else if len(ref) >= 5 && ref[:5] == "ROOT/" || ref == "ROOT" {
			qualifier, hasQualifier := logicalnames.LogicalNameGetQualifier(ref)
			bare := logicalnames.LogicalNameStripQualifier(ref)

			path, err := logicalnames.LogicalNameToPath(bare)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", ErrInvalidLogicalName, err)
			}

			item := &ChainItem{
				LogicalName: bare,
				FilePath:    path,
				Qualifier:   nil,
			}
			if hasQualifier {
				q := qualifier
				item.Qualifier = &q
			}
			deps = append(deps, item)
		} else {
			return nil, fmt.Errorf("%w: unsupported reference %q", ErrUnresolvableArtifact, ref)
		}
	}

	// Sort by file path, then qualifier (absent before present)
	sort.SliceStable(deps, func(i, j int) bool {
		pi := deps[i].FilePath.Value
		pj := deps[j].FilePath.Value
		if pi != pj {
			return pi < pj
		}
		// absent qualifier sorts before present
		if deps[i].Qualifier == nil && deps[j].Qualifier != nil {
			return true
		}
		if deps[i].Qualifier != nil && deps[j].Qualifier == nil {
			return false
		}
		if deps[i].Qualifier != nil && deps[j].Qualifier != nil {
			return *deps[i].Qualifier < *deps[j].Qualifier
		}
		return false
	})

	return deps, nil
}

// resolveArtifactRef resolves an ARTIFACT/ reference into a ChainItem.
func resolveArtifactRef(ref string) (*ChainItem, error) {
	qualifier, hasQualifier := logicalnames.LogicalNameGetQualifier(ref)
	if !hasQualifier {
		return nil, fmt.Errorf("%w: ARTIFACT/ reference %q missing qualifier", ErrUnresolvableArtifact, ref)
	}

	generatorLogicalName, err := logicalnames.LogicalNameGetArtifactGenerator(ref)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidLogicalName, err)
	}

	generatorFilePath, err := logicalnames.LogicalNameToPath(generatorLogicalName)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidLogicalName, err)
	}

	generatorFrontmatter, err := frontmatter.FrontmatterParse(generatorFilePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnreadableFrontmatter, err)
	}

	artifactPath, err := findOutputPath(generatorFrontmatter, qualifier, ref)
	if err != nil {
		return nil, err
	}

	q := qualifier
	return &ChainItem{
		LogicalName: ref,
		FilePath:    &pathutils.PathCfs{Value: artifactPath},
		Qualifier:   &q,
	}, nil
}

// findOutputPath searches a frontmatter's outputs for an entry with the given id.
func findOutputPath(fm *frontmatter.Frontmatter, id string, ref string) (string, error) {
	for _, out := range fm.Outputs {
		if out != nil && out.ID == id {
			return out.Path, nil
		}
	}
	return "", fmt.Errorf("%w: output id %q not found for %q", ErrUnresolvableArtifact, id, ref)
}

// deduplicateDependencies removes redundant entries from sorted deps.
func deduplicateDependencies(deps []*ChainItem) []*ChainItem {
	result := make([]*ChainItem, 0, len(deps))

	for _, item := range deps {
		if logicalnames.LogicalNameIsArtifact(item.LogicalName) {
			// For ARTIFACT/ entries: skip exact duplicates by logical name
			if containsArtifact(result, item.LogicalName) {
				continue
			}
			result = append(result, item)
		} else {
			// For ROOT/ entries
			filePath := item.FilePath.Value

			if item.Qualifier == nil {
				// No qualifier — check if already present with no qualifier
				if containsRootWithPath(result, filePath, nil) {
					continue
				}
				// Add it and remove any previously added entries with same path + qualifier
				result = removeRootWithPathAndQualifier(result, filePath)
				result = append(result, item)
			} else {
				// Has qualifier — skip if same path+qualifier already present
				if containsRootWithPath(result, filePath, item.Qualifier) {
					continue
				}
				// Skip if same path with no qualifier already present (subsumed)
				if containsRootWithPath(result, filePath, nil) {
					continue
				}
				result = append(result, item)
			}
		}
	}

	return result
}

// containsArtifact reports whether result already has an ARTIFACT/ item with the given logical name.
func containsArtifact(result []*ChainItem, logicalName string) bool {
	for _, r := range result {
		if r.LogicalName == logicalName {
			return true
		}
	}
	return false
}

// containsRootWithPath reports whether result has a ROOT/ item matching the given path and qualifier.
// If qualifier is nil, it looks for an entry with no qualifier.
func containsRootWithPath(result []*ChainItem, filePath string, qualifier *string) bool {
	for _, r := range result {
		if logicalnames.LogicalNameIsArtifact(r.LogicalName) {
			continue
		}
		if r.FilePath.Value != filePath {
			continue
		}
		if qualifier == nil && r.Qualifier == nil {
			return true
		}
		if qualifier != nil && r.Qualifier != nil && *qualifier == *r.Qualifier {
			return true
		}
	}
	return false
}

// removeRootWithPathAndQualifier removes from result any ROOT/ entries that have
// the given file path AND a non-nil qualifier (they are subsumed by a no-qualifier entry).
func removeRootWithPathAndQualifier(result []*ChainItem, filePath string) []*ChainItem {
	filtered := result[:0]
	for _, r := range result {
		if !logicalnames.LogicalNameIsArtifact(r.LogicalName) &&
			r.FilePath.Value == filePath &&
			r.Qualifier != nil {
			continue
		}
		filtered = append(filtered, r)
	}
	return filtered
}

// collectExternal copies and sorts the external list from frontmatter.
func collectExternal(fm *frontmatter.Frontmatter) []*frontmatter.FrontmatterExternal {
	if len(fm.External) == 0 {
		return []*frontmatter.FrontmatterExternal{}
	}

	external := make([]*frontmatter.FrontmatterExternal, len(fm.External))
	copy(external, fm.External)

	sort.SliceStable(external, func(i, j int) bool {
		return external[i].Path < external[j].Path
	})

	return external
}

// resolveInput resolves the input artifact reference from frontmatter.
func resolveInput(fm *frontmatter.Frontmatter) (*ChainItem, error) {
	if fm.Input == "" {
		return nil, nil
	}

	ref := fm.Input
	qualifier, hasQualifier := logicalnames.LogicalNameGetQualifier(ref)
	if !hasQualifier {
		return nil, fmt.Errorf("%w: input reference %q missing qualifier", ErrUnresolvableArtifact, ref)
	}

	generatorLogicalName, err := logicalnames.LogicalNameGetArtifactGenerator(ref)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidLogicalName, err)
	}

	generatorFilePath, err := logicalnames.LogicalNameToPath(generatorLogicalName)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidLogicalName, err)
	}

	generatorFrontmatter, err := frontmatter.FrontmatterParse(generatorFilePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnreadableFrontmatter, err)
	}

	artifactPath, err := findOutputPath(generatorFrontmatter, qualifier, ref)
	if err != nil {
		return nil, err
	}

	q := qualifier
	return &ChainItem{
		LogicalName: ref,
		FilePath:    &pathutils.PathCfs{Value: artifactPath},
		Qualifier:   &q,
	}, nil
}
