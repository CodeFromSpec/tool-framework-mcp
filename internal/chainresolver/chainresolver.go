// code-from-spec: ROOT/golang/implementation/chain/resolver@k_gm1NlUS-8ZUFPhJC2kVDh1Ykw

package chainresolver

import (
	"errors"
	"fmt"
	"sort"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ErrUnreadableFrontmatter is returned when a node's frontmatter
// cannot be parsed.
var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")

// ErrUnresolvableArtifact is returned when an ARTIFACT/ reference's
// output id does not match any declared output.
var ErrUnresolvableArtifact = errors.New("unresolvable artifact reference")

// ChainItem represents a single node entry within a resolved chain.
type ChainItem struct {
	// LogicalName is the logical name of the node.
	LogicalName string

	// FilePath is the CFS path to the node's file.
	FilePath pathutils.PathCfs

	// Qualifier is an optional qualifier for the entry.
	// Empty string when absent.
	Qualifier string
}

// Chain holds the fully resolved chain for a target node.
// Items are ordered as described in the chain assembly specification.
type Chain struct {
	// Ancestors is the ordered list of ancestor nodes, from root down
	// to (but not including) the target node.
	Ancestors []*ChainItem

	// Dependencies is the list of entries from the target's depends_on,
	// sorted alphabetically by file path then by qualifier, each with
	// its resolved file path and an optional qualifier.
	Dependencies []*ChainItem

	// External is the list of external file references from the target's
	// frontmatter, sorted alphabetically by path, including fragment
	// declarations when present.
	External []*frontmatter.FrontmatterExternal

	// Target is the target node itself.
	Target *ChainItem

	// Input is the target's input artifact. Nil when absent.
	Input *ChainItem
}

// ChainResolve returns the resolved chain for the given target logical
// name. The chain contains ancestors, dependencies, external references,
// the target node, and optionally an input artifact — in the order
// required for context assembly or chain hash computation.
func ChainResolve(targetLogicalName string) (*Chain, error) {
	// Step 1 — Resolve ancestors and target
	ancestors, target, err := resolveAncestorsAndTarget(targetLogicalName)
	if err != nil {
		return nil, err
	}

	// Step 2 — Resolve dependencies
	targetFrontmatter, err := parseFrontmatter(target.FilePath)
	if err != nil {
		return nil, err
	}

	dependencies, err := resolveDependencies(targetFrontmatter)
	if err != nil {
		return nil, err
	}

	// Step 3 — Deduplicate dependencies
	dependencies = deduplicateDependencies(dependencies)

	// Step 4 — Collect external
	external := collectExternal(targetFrontmatter)

	// Step 5 — Resolve input
	input, err := resolveInput(targetFrontmatter)
	if err != nil {
		return nil, err
	}

	// Step 6 — Return
	return &Chain{
		Ancestors:    ancestors,
		Dependencies: dependencies,
		External:     external,
		Target:       target,
		Input:        input,
	}, nil
}

// resolveAncestorsAndTarget builds the ancestors list and target ChainItem.
func resolveAncestorsAndTarget(targetLogicalName string) ([]*ChainItem, *ChainItem, error) {
	if targetLogicalName == "ROOT" {
		filePath, err := logicalnames.LogicalNameToPath("ROOT")
		if err != nil {
			return nil, nil, err
		}
		target := &ChainItem{
			LogicalName: "ROOT",
			FilePath:    *filePath,
		}
		return []*ChainItem{}, target, nil
	}

	// Collect all names from target up to ROOT
	collectedNames := []string{targetLogicalName}
	current := targetLogicalName

	for {
		parent, err := logicalnames.LogicalNameGetParent(current)
		if err != nil {
			return nil, nil, err
		}
		collectedNames = append(collectedNames, parent)
		if parent == "ROOT" {
			break
		}
		current = parent
	}

	// Sort alphabetically — produces root-first order
	sort.Strings(collectedNames)

	// Build ChainItems for each name
	items := make([]*ChainItem, 0, len(collectedNames))
	for _, name := range collectedNames {
		stripped := logicalnames.LogicalNameStripQualifier(name)
		filePath, err := logicalnames.LogicalNameToPath(stripped)
		if err != nil {
			return nil, nil, err
		}
		item := &ChainItem{
			LogicalName: stripped,
			FilePath:    *filePath,
		}
		items = append(items, item)
	}

	// Last item is the target; everything before it is ancestors
	target := items[len(items)-1]
	ancestors := items[:len(items)-1]

	return ancestors, target, nil
}

// parseFrontmatter parses a node's frontmatter and wraps errors with ErrUnreadableFrontmatter.
func parseFrontmatter(filePath pathutils.PathCfs) (*frontmatter.Frontmatter, error) {
	fm, err := frontmatter.FrontmatterParse(&filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnreadableFrontmatter, err)
	}
	return fm, nil
}

// resolveDependencies processes the depends_on list from the frontmatter.
func resolveDependencies(fm *frontmatter.Frontmatter) ([]*ChainItem, error) {
	dependencies := make([]*ChainItem, 0, len(fm.DependsOn))

	for _, entry := range fm.DependsOn {
		var item *ChainItem
		var err error

		switch {
		case len(entry) >= 5 && entry[:5] == "ROOT/":
			item, err = resolveRootDependency(entry)
		case len(entry) >= 9 && entry[:9] == "ARTIFACT/":
			item, err = resolveArtifactDependency(entry)
		default:
			return nil, fmt.Errorf("%w: unrecognized reference prefix in %q", ErrUnresolvableArtifact, entry)
		}

		if err != nil {
			return nil, err
		}

		dependencies = append(dependencies, item)
	}

	// Sort dependencies: primary by file path, secondary by qualifier
	sort.SliceStable(dependencies, func(i, j int) bool {
		pathI := dependencies[i].FilePath.Value
		pathJ := dependencies[j].FilePath.Value
		if pathI != pathJ {
			return pathI < pathJ
		}
		// qualifier absent sorts before present
		qualI := dependencies[i].Qualifier
		qualJ := dependencies[j].Qualifier
		if qualI == "" && qualJ != "" {
			return true
		}
		if qualI != "" && qualJ == "" {
			return false
		}
		return qualI < qualJ
	})

	return dependencies, nil
}

// resolveRootDependency resolves a ROOT/ dependency entry.
func resolveRootDependency(entry string) (*ChainItem, error) {
	qualifier, _ := logicalnames.LogicalNameGetQualifier(entry)
	bareName := logicalnames.LogicalNameStripQualifier(entry)

	filePath, err := logicalnames.LogicalNameToPath(bareName)
	if err != nil {
		return nil, err
	}

	return &ChainItem{
		LogicalName: bareName,
		FilePath:    *filePath,
		Qualifier:   qualifier,
	}, nil
}

// resolveArtifactDependency resolves an ARTIFACT/ dependency entry.
func resolveArtifactDependency(entry string) (*ChainItem, error) {
	qualifier, present := logicalnames.LogicalNameGetQualifier(entry)
	if !present {
		return nil, fmt.Errorf("%w: ARTIFACT/ reference %q has no qualifier", ErrUnresolvableArtifact, entry)
	}

	generatorName, err := logicalnames.LogicalNameGetArtifactGenerator(entry)
	if err != nil {
		return nil, err
	}

	generatorPath, err := logicalnames.LogicalNameToPath(generatorName)
	if err != nil {
		return nil, err
	}

	generatorFrontmatter, err := parseFrontmatter(*generatorPath)
	if err != nil {
		return nil, err
	}

	artifactPath, err := findOutputPath(generatorFrontmatter, qualifier, entry)
	if err != nil {
		return nil, err
	}

	return &ChainItem{
		LogicalName: entry,
		FilePath:    pathutils.PathCfs{Value: artifactPath},
		Qualifier:   qualifier,
	}, nil
}

// findOutputPath searches for an output by id in a frontmatter and returns its path.
func findOutputPath(fm *frontmatter.Frontmatter, id string, ref string) (string, error) {
	for _, output := range fm.Outputs {
		if output.ID == id {
			return output.Path, nil
		}
	}
	return "", fmt.Errorf("%w: no output with id %q found for reference %q", ErrUnresolvableArtifact, id, ref)
}

// deduplicateDependencies removes duplicate entries from dependencies.
func deduplicateDependencies(dependencies []*ChainItem) []*ChainItem {
	deduped := make([]*ChainItem, 0, len(dependencies))

	for _, item := range dependencies {
		if logicalnames.LogicalNameIsArtifact(item.LogicalName) {
			// ARTIFACT/ entry: duplicate if same logical_name (including qualifier)
			if !containsArtifact(deduped, item.LogicalName) {
				deduped = append(deduped, item)
			}
		} else {
			// ROOT/ entry
			if shouldSkipRootEntry(deduped, item) {
				continue
			}
			// If this item has no qualifier, remove all previously added items
			// with the same file_path but a non-absent qualifier
			if item.Qualifier == "" {
				deduped = removeQualifiedEntriesWithPath(deduped, item.FilePath.Value)
			}
			deduped = append(deduped, item)
		}
	}

	return deduped
}

// containsArtifact checks if deduped already has an item with the given logical name.
func containsArtifact(deduped []*ChainItem, logicalName string) bool {
	for _, existing := range deduped {
		if existing.LogicalName == logicalName {
			return true
		}
	}
	return false
}

// shouldSkipRootEntry returns true if the item should be skipped during deduplication.
func shouldSkipRootEntry(deduped []*ChainItem, item *ChainItem) bool {
	for _, existing := range deduped {
		if logicalnames.LogicalNameIsArtifact(existing.LogicalName) {
			continue
		}
		if existing.FilePath.Value != item.FilePath.Value {
			continue
		}
		// Same file path — check qualifier rules
		if existing.Qualifier == item.Qualifier {
			// Exact duplicate — skip
			return true
		}
		if existing.Qualifier == "" {
			// Existing has no qualifier (full section) — current item is subsumed — skip
			return true
		}
	}
	return false
}

// removeQualifiedEntriesWithPath removes all ROOT/ items from deduped that have
// the specified file path and a non-empty qualifier.
func removeQualifiedEntriesWithPath(deduped []*ChainItem, filePath string) []*ChainItem {
	result := make([]*ChainItem, 0, len(deduped))
	for _, item := range deduped {
		if !logicalnames.LogicalNameIsArtifact(item.LogicalName) &&
			item.FilePath.Value == filePath &&
			item.Qualifier != "" {
			// Remove this entry — it's subsumed by the new unqualified entry
			continue
		}
		result = append(result, item)
	}
	return result
}

// collectExternal gathers and sorts external references from frontmatter.
func collectExternal(fm *frontmatter.Frontmatter) []*frontmatter.FrontmatterExternal {
	external := make([]*frontmatter.FrontmatterExternal, len(fm.External))
	copy(external, fm.External)

	sort.SliceStable(external, func(i, j int) bool {
		return external[i].Path < external[j].Path
	})

	return external
}

// resolveInput resolves the input artifact if declared in the frontmatter.
func resolveInput(fm *frontmatter.Frontmatter) (*ChainItem, error) {
	if fm.Input == "" {
		return nil, nil
	}

	qualifier, present := logicalnames.LogicalNameGetQualifier(fm.Input)
	if !present {
		return nil, fmt.Errorf("%w: input reference %q has no qualifier", ErrUnresolvableArtifact, fm.Input)
	}

	inputGeneratorName, err := logicalnames.LogicalNameGetArtifactGenerator(fm.Input)
	if err != nil {
		return nil, err
	}

	inputGeneratorPath, err := logicalnames.LogicalNameToPath(inputGeneratorName)
	if err != nil {
		return nil, err
	}

	inputGeneratorFrontmatter, err := parseFrontmatter(*inputGeneratorPath)
	if err != nil {
		return nil, err
	}

	inputArtifactPath, err := findOutputPath(inputGeneratorFrontmatter, qualifier, fm.Input)
	if err != nil {
		return nil, err
	}

	return &ChainItem{
		LogicalName: fm.Input,
		FilePath:    pathutils.PathCfs{Value: inputArtifactPath},
		Qualifier:   qualifier,
	}, nil
}
