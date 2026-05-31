// code-from-spec: ROOT/golang/implementation/chain/resolver@rrBfl4xNCIMDUXXbkZhwmf7-Ppk
package chainresolver

import (
	"errors"
	"fmt"
	"sort"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ErrUnreadableFrontmatter is returned when a node's frontmatter cannot be parsed.
var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")

// ErrUnresolvableArtifact is returned when an ARTIFACT/ reference's output id
// does not match any declared output in the referenced node.
var ErrUnresolvableArtifact = errors.New("unresolvable artifact")

// ChainItem represents a single node position within a resolved chain.
type ChainItem struct {
	// LogicalName is the logical name of the node.
	LogicalName string

	// FilePath is the CFS-format path to the node's spec file.
	FilePath *pathutils.PathCfs

	// Qualifier is an optional qualifier string, used for dependency entries.
	Qualifier *string
}

// Chain is the fully resolved chain for a target node, in assembly order.
type Chain struct {
	// Ancestors holds the nodes from root down to (but not including) the target.
	Ancestors []*ChainItem

	// Dependencies holds entries from the target's depends_on, sorted
	// alphabetically by file path then by qualifier.
	Dependencies []*ChainItem

	// External holds external file references from the target's frontmatter,
	// sorted alphabetically by path, including any fragment declarations.
	External []*frontmatter.FrontmatterExternal

	// Target is the target node itself.
	Target *ChainItem

	// Input is the target's input artifact, if present.
	Input *ChainItem
}

// ChainResolve returns the chain for the given target logical name.
func ChainResolve(targetLogicalName string) (*Chain, error) {
	// Step 1 — Resolve ancestors and target.
	ancestors, target, err := resolveAncestorsAndTarget(targetLogicalName)
	if err != nil {
		return nil, err
	}

	// Step 2 — Resolve dependencies.
	fm, err := parseFrontmatter(target.FilePath)
	if err != nil {
		return nil, err
	}

	deps, err := resolveDependencies(fm)
	if err != nil {
		return nil, err
	}

	// Step 3 — Deduplicate dependencies.
	dependencies := deduplicateDeps(deps)

	// Step 4 — Collect external.
	external := collectExternal(fm)

	// Step 5 — Resolve input.
	input, err := resolveInput(fm)
	if err != nil {
		return nil, err
	}

	return &Chain{
		Ancestors:    ancestors,
		Dependencies: dependencies,
		External:     external,
		Target:       target,
		Input:        input,
	}, nil
}

// resolveAncestorsAndTarget builds the ancestors list and the target ChainItem
// for the given logical name.
func resolveAncestorsAndTarget(targetLogicalName string) ([]*ChainItem, *ChainItem, error) {
	if targetLogicalName == "ROOT" {
		filePath, err := logicalnames.LogicalNameToPath("ROOT")
		if err != nil {
			return nil, nil, err
		}
		target := &ChainItem{
			LogicalName: "ROOT",
			FilePath:    filePath,
			Qualifier:   nil,
		}
		return []*ChainItem{}, target, nil
	}

	// Collect all names from target up to ROOT.
	nameList := []string{targetLogicalName}
	current := targetLogicalName
	for {
		parent, err := logicalnames.LogicalNameGetParent(current)
		if err != nil {
			return nil, nil, err
		}
		current = parent
		nameList = append(nameList, current)
		if current == "ROOT" {
			break
		}
	}

	// Sort alphabetically — this gives root-first order.
	sort.Strings(nameList)

	// Build ChainItems for each name.
	items := make([]*ChainItem, 0, len(nameList))
	for _, name := range nameList {
		filePath, err := logicalnames.LogicalNameToPath(name)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, &ChainItem{
			LogicalName: name,
			FilePath:    filePath,
			Qualifier:   nil,
		})
	}

	// Last item is the target; everything before it are ancestors.
	target := items[len(items)-1]
	ancestors := items[:len(items)-1]
	return ancestors, target, nil
}

// parseFrontmatter wraps FrontmatterParse and maps errors to ErrUnreadableFrontmatter.
func parseFrontmatter(filePath *pathutils.PathCfs) (*frontmatter.Frontmatter, error) {
	fm, err := frontmatter.FrontmatterParse(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnreadableFrontmatter, err)
	}
	return fm, nil
}

// resolveDependencies resolves the depends_on entries in the frontmatter into
// a sorted list of ChainItems.
func resolveDependencies(fm *frontmatter.Frontmatter) ([]*ChainItem, error) {
	deps := make([]*ChainItem, 0, len(fm.DependsOn))

	for _, ref := range fm.DependsOn {
		item, err := resolveSingleDep(ref)
		if err != nil {
			return nil, err
		}
		deps = append(deps, item)
	}

	// Sort by file_path value, then by qualifier (absent before present).
	sort.SliceStable(deps, func(i, j int) bool {
		pi := deps[i].FilePath.Value
		pj := deps[j].FilePath.Value
		if pi != pj {
			return pi < pj
		}
		// absent qualifier sorts before present.
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

// resolveSingleDep resolves a single depends_on reference into a ChainItem.
func resolveSingleDep(ref string) (*ChainItem, error) {
	if len(ref) >= 5 && ref[:5] == "ROOT/" {
		qualifier, hasQualifier := logicalnames.LogicalNameGetQualifier(ref)
		bare := logicalnames.LogicalNameStripQualifier(ref)

		filePath, err := logicalnames.LogicalNameToPath(bare)
		if err != nil {
			return nil, err
		}

		item := &ChainItem{
			LogicalName: bare,
			FilePath:    filePath,
			Qualifier:   nil,
		}
		if hasQualifier {
			q := qualifier
			item.Qualifier = &q
		}
		return item, nil
	}

	if logicalnames.LogicalNameIsArtifact(ref) {
		qualifier, hasQualifier := logicalnames.LogicalNameGetQualifier(ref)
		if !hasQualifier {
			return nil, fmt.Errorf("%w: ARTIFACT/ reference %q is missing an id qualifier", ErrUnresolvableArtifact, ref)
		}

		generatorName, err := logicalnames.LogicalNameGetArtifactGenerator(ref)
		if err != nil {
			return nil, err
		}

		generatorPath, err := logicalnames.LogicalNameToPath(generatorName)
		if err != nil {
			return nil, err
		}

		generatorFM, err := parseFrontmatter(generatorPath)
		if err != nil {
			return nil, err
		}

		artifactPath, err := findOutputPath(generatorFM, qualifier, ref)
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

	return nil, fmt.Errorf("%w: dependency %q does not start with ROOT/ or ARTIFACT/", ErrUnresolvableArtifact, ref)
}

// findOutputPath searches the frontmatter outputs for an entry with the given id.
func findOutputPath(fm *frontmatter.Frontmatter, id string, ref string) (string, error) {
	for _, output := range fm.Outputs {
		if output.ID == id {
			return output.Path, nil
		}
	}
	return "", fmt.Errorf("%w: no output with id %q found for reference %q", ErrUnresolvableArtifact, id, ref)
}

// deduplicateDeps removes duplicate entries from deps according to the
// deduplication rules specified in the pseudocode.
func deduplicateDeps(deps []*ChainItem) []*ChainItem {
	deduped := make([]*ChainItem, 0, len(deps))

	for _, entry := range deps {
		if logicalnames.LogicalNameIsArtifact(entry.LogicalName) {
			// ARTIFACT/ entry: deduplicate by exact logical name (including qualifier).
			if !containsArtifact(deduped, entry.LogicalName) {
				deduped = append(deduped, entry)
			}
		} else {
			// ROOT/ entry.
			if entry.Qualifier == nil {
				// Unqualified: add if no unqualified entry with same file path exists,
				// and remove any existing qualified entries with same file path.
				if !containsUnqualifiedPath(deduped, entry.FilePath.Value) {
					deduped = removeQualifiedByPath(deduped, entry.FilePath.Value)
					deduped = append(deduped, entry)
				}
			} else {
				// Qualified: discard if an unqualified entry with same path exists.
				if containsUnqualifiedPath(deduped, entry.FilePath.Value) {
					continue
				}
				// Discard if same path + same qualifier already exists.
				if containsQualifiedPath(deduped, entry.FilePath.Value, *entry.Qualifier) {
					continue
				}
				deduped = append(deduped, entry)
			}
		}
	}

	return deduped
}

// containsArtifact returns true if deduped already has an entry with the given logical name.
func containsArtifact(deduped []*ChainItem, logicalName string) bool {
	for _, d := range deduped {
		if d.LogicalName == logicalName {
			return true
		}
	}
	return false
}

// containsUnqualifiedPath returns true if deduped has a ROOT/ entry with the given
// file path and no qualifier.
func containsUnqualifiedPath(deduped []*ChainItem, filePath string) bool {
	for _, d := range deduped {
		if !logicalnames.LogicalNameIsArtifact(d.LogicalName) && d.FilePath.Value == filePath && d.Qualifier == nil {
			return true
		}
	}
	return false
}

// containsQualifiedPath returns true if deduped has a ROOT/ entry with the given
// file path and matching qualifier value.
func containsQualifiedPath(deduped []*ChainItem, filePath string, qualifier string) bool {
	for _, d := range deduped {
		if !logicalnames.LogicalNameIsArtifact(d.LogicalName) && d.FilePath.Value == filePath && d.Qualifier != nil && *d.Qualifier == qualifier {
			return true
		}
	}
	return false
}

// removeQualifiedByPath returns a new slice with all ROOT/ entries having the
// given file path and a present qualifier removed.
func removeQualifiedByPath(deduped []*ChainItem, filePath string) []*ChainItem {
	result := make([]*ChainItem, 0, len(deduped))
	for _, d := range deduped {
		if !logicalnames.LogicalNameIsArtifact(d.LogicalName) && d.FilePath.Value == filePath && d.Qualifier != nil {
			continue
		}
		result = append(result, d)
	}
	return result
}

// collectExternal copies and sorts the external entries from the frontmatter.
func collectExternal(fm *frontmatter.Frontmatter) []*frontmatter.FrontmatterExternal {
	external := make([]*frontmatter.FrontmatterExternal, len(fm.External))
	copy(external, fm.External)
	sort.SliceStable(external, func(i, j int) bool {
		return external[i].Path < external[j].Path
	})
	return external
}

// resolveInput resolves the input artifact reference from the frontmatter.
func resolveInput(fm *frontmatter.Frontmatter) (*ChainItem, error) {
	if fm.Input == "" {
		return nil, nil
	}

	ref := fm.Input

	qualifier, hasQualifier := logicalnames.LogicalNameGetQualifier(ref)
	if !hasQualifier {
		return nil, fmt.Errorf("%w: input reference %q is missing an id qualifier", ErrUnresolvableArtifact, ref)
	}

	generatorName, err := logicalnames.LogicalNameGetArtifactGenerator(ref)
	if err != nil {
		return nil, err
	}

	generatorPath, err := logicalnames.LogicalNameToPath(generatorName)
	if err != nil {
		return nil, err
	}

	generatorFM, err := parseFrontmatter(generatorPath)
	if err != nil {
		return nil, err
	}

	artifactPath, err := findOutputPath(generatorFM, qualifier, ref)
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
