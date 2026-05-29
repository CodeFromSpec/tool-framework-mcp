// code-from-spec: ROOT/golang/implementation/chain/resolver@YVgNXPoyfciYLcOe-TZ4T7VKbVU

package chainresolver

import (
	"errors"
	"fmt"
	"sort"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ErrUnresolvableArtifact is returned when an ARTIFACT/ reference's
// output id does not match any output declared in the referenced
// node's frontmatter.
var ErrUnresolvableArtifact = errors.New("unresolvable artifact")

// ChainItem represents a single node entry within a resolved chain.
// It holds the logical name, the CFS file path to the node's spec
// file, and an optional qualifier (used when the item originates
// from an ARTIFACT/ reference that targets a specific output id).
type ChainItem struct {
	LogicalName string
	FilePath    *pathutils.PathCfs
	Qualifier   *string
}

// Chain holds the fully resolved chain for a target logical name.
// It contains the ordered lists of ancestors, dependencies, external
// files, the target itself, and an optional input artifact.
//
// Assembly order mirrors the order in which a downstream tool should
// concatenate context:
//  1. Ancestors — root down to (but not including) the target.
//  2. Dependencies — target's depends_on, sorted alphabetically by
//     file path then qualifier.
//  3. External — target's external files, sorted alphabetically by
//     path.
//  4. Target — the target node itself.
//  5. Input — the target's input artifact, if present.
type Chain struct {
	Ancestors    []*ChainItem
	Dependencies []*ChainItem
	External     []*frontmatter.FrontmatterExternal
	Target       *ChainItem
	Input        *ChainItem
}

// ChainResolve builds and returns the full chain for the given target
// logical name. It walks the ancestor hierarchy, collects depends_on
// entries, gathers external file references, and resolves the optional
// input artifact.
//
// The returned Chain fields are ordered as described in the Chain struct
// documentation.
//
// Returns an error if:
//   - the target logical name is invalid or cannot be converted to a
//     path (errors propagated from LogicalNameToPath / LogicalNameGetParent).
//   - any node's frontmatter cannot be parsed (frontmatter parse errors
//     are propagated).
//   - an ARTIFACT/ reference in depends_on specifies an output id that
//     does not match any declared output in the referenced node's
//     frontmatter (ErrUnresolvableArtifact).
func ChainResolve(target_logical_name string) (*Chain, error) {
	// Step 1 — Resolve ancestors and target
	var ancestors []*ChainItem
	var target *ChainItem

	if target_logical_name == "ROOT" {
		filePath, err := logicalnames.LogicalNameToPath("ROOT")
		if err != nil {
			return nil, fmt.Errorf("resolving ROOT path: %w", err)
		}
		target = &ChainItem{
			LogicalName: "ROOT",
			FilePath:    filePath,
			Qualifier:   nil,
		}
		ancestors = []*ChainItem{}
	} else {
		// Collect the target and all ancestors into nameList
		nameList := []string{}
		nameList = append(nameList, target_logical_name)

		currentName := target_logical_name
		for {
			parentName, err := logicalnames.LogicalNameGetParent(currentName)
			if err != nil {
				return nil, fmt.Errorf("getting parent of %q: %w", currentName, err)
			}
			nameList = append(nameList, parentName)
			if parentName == "ROOT" {
				break
			}
			currentName = parentName
		}

		// Sort alphabetically — this produces root-first order
		sort.Strings(nameList)

		// Convert each name to a ChainItem
		items := make([]*ChainItem, 0, len(nameList))
		for _, name := range nameList {
			filePath, err := logicalnames.LogicalNameToPath(name)
			if err != nil {
				return nil, fmt.Errorf("resolving path for %q: %w", name, err)
			}
			items = append(items, &ChainItem{
				LogicalName: name,
				FilePath:    filePath,
				Qualifier:   nil,
			})
		}

		// Last item is the target; the rest are ancestors
		target = items[len(items)-1]
		ancestors = items[:len(items)-1]
	}

	// Step 2 — Resolve dependencies
	targetFrontmatter, err := frontmatter.FrontmatterParse(target.FilePath)
	if err != nil {
		return nil, fmt.Errorf("parsing frontmatter for %q: %w", target.LogicalName, err)
	}

	dependencies := []*ChainItem{}

	for _, entry := range targetFrontmatter.DependsOn {
		if len(entry) >= 5 && entry[:5] == "ROOT/" {
			qualifier, hasQualifier := logicalnames.LogicalNameGetQualifier(entry)
			bareName := logicalnames.LogicalNameStripQualifier(entry)

			depPath, err := logicalnames.LogicalNameToPath(bareName)
			if err != nil {
				return nil, fmt.Errorf("resolving dependency path for %q: %w", bareName, err)
			}

			item := &ChainItem{
				LogicalName: bareName,
				FilePath:    depPath,
				Qualifier:   nil,
			}
			if hasQualifier {
				q := qualifier
				item.Qualifier = &q
			}
			dependencies = append(dependencies, item)

		} else if len(entry) >= 9 && entry[:9] == "ARTIFACT/" {
			qualifier, hasQualifier := logicalnames.LogicalNameGetQualifier(entry)
			if !hasQualifier {
				return nil, fmt.Errorf("dependency %q: %w", entry, ErrUnresolvableArtifact)
			}
			artifactID := qualifier

			generatorName, err := logicalnames.LogicalNameGetArtifactGenerator(entry)
			if err != nil {
				return nil, fmt.Errorf("getting artifact generator for %q: %w", entry, err)
			}

			generatorPath, err := logicalnames.LogicalNameToPath(generatorName)
			if err != nil {
				return nil, fmt.Errorf("resolving generator path for %q: %w", generatorName, err)
			}

			generatorFrontmatter, err := frontmatter.FrontmatterParse(generatorPath)
			if err != nil {
				return nil, fmt.Errorf("parsing frontmatter for generator %q: %w", generatorName, err)
			}

			artifactPath := ""
			for _, output := range generatorFrontmatter.Outputs {
				if output.ID == artifactID {
					artifactPath = output.Path
					break
				}
			}
			if artifactPath == "" {
				return nil, fmt.Errorf("artifact %q output id %q not found: %w", entry, artifactID, ErrUnresolvableArtifact)
			}

			q := artifactID
			item := &ChainItem{
				LogicalName: entry,
				FilePath:    &pathutils.PathCfs{Value: artifactPath},
				Qualifier:   &q,
			}
			dependencies = append(dependencies, item)

		} else {
			return nil, fmt.Errorf("dependency %q has unrecognized prefix: %w", entry, ErrUnresolvableArtifact)
		}
	}

	// Sort dependencies by file_path, then qualifier (absent before present, then alphabetical)
	sort.SliceStable(dependencies, func(i, j int) bool {
		pi := dependencies[i].FilePath.Value
		pj := dependencies[j].FilePath.Value
		if pi != pj {
			return pi < pj
		}
		qi := dependencies[i].Qualifier
		qj := dependencies[j].Qualifier
		if qi == nil && qj == nil {
			return false
		}
		if qi == nil {
			return true
		}
		if qj == nil {
			return false
		}
		return *qi < *qj
	})

	// Step 3 — Deduplicate dependencies
	deduped := []*ChainItem{}
	for _, entry := range dependencies {
		if logicalnames.LogicalNameIsArtifact(entry.LogicalName) {
			// ARTIFACT/ entry: deduplicate by logical_name (including qualifier)
			duplicate := false
			for _, existing := range deduped {
				if existing.LogicalName == entry.LogicalName {
					duplicate = true
					break
				}
			}
			if !duplicate {
				deduped = append(deduped, entry)
			}
		} else {
			// ROOT/ entry: deduplicate by file_path and qualifier
			duplicate := false
			for _, existing := range deduped {
				if existing.FilePath.Value == entry.FilePath.Value {
					// Same file_path and same qualifier → skip
					if qualifierEqual(existing.Qualifier, entry.Qualifier) {
						duplicate = true
						break
					}
					// Existing has absent qualifier → existing subsumes any qualified entry
					if existing.Qualifier == nil {
						duplicate = true
						break
					}
				}
			}
			if !duplicate {
				deduped = append(deduped, entry)
			}
		}
	}
	dependencies = deduped

	// Step 4 — Collect external
	external := make([]*frontmatter.FrontmatterExternal, len(targetFrontmatter.External))
	copy(external, targetFrontmatter.External)
	sort.SliceStable(external, func(i, j int) bool {
		return external[i].Path < external[j].Path
	})

	// Step 5 — Resolve input
	var input *ChainItem

	if targetFrontmatter.Input != "" {
		inputEntry := targetFrontmatter.Input

		qualifier, hasQualifier := logicalnames.LogicalNameGetQualifier(inputEntry)
		if !hasQualifier {
			return nil, fmt.Errorf("input %q missing qualifier: %w", inputEntry, ErrUnresolvableArtifact)
		}
		artifactID := qualifier

		generatorName, err := logicalnames.LogicalNameGetArtifactGenerator(inputEntry)
		if err != nil {
			return nil, fmt.Errorf("getting input artifact generator for %q: %w", inputEntry, err)
		}

		generatorPath, err := logicalnames.LogicalNameToPath(generatorName)
		if err != nil {
			return nil, fmt.Errorf("resolving input generator path for %q: %w", generatorName, err)
		}

		generatorFrontmatter, err := frontmatter.FrontmatterParse(generatorPath)
		if err != nil {
			return nil, fmt.Errorf("parsing frontmatter for input generator %q: %w", generatorName, err)
		}

		artifactPath := ""
		for _, output := range generatorFrontmatter.Outputs {
			if output.ID == artifactID {
				artifactPath = output.Path
				break
			}
		}
		if artifactPath == "" {
			return nil, fmt.Errorf("input %q output id %q not found: %w", inputEntry, artifactID, ErrUnresolvableArtifact)
		}

		q := artifactID
		input = &ChainItem{
			LogicalName: inputEntry,
			FilePath:    &pathutils.PathCfs{Value: artifactPath},
			Qualifier:   &q,
		}
	}

	return &Chain{
		Ancestors:    ancestors,
		Dependencies: dependencies,
		External:     external,
		Target:       target,
		Input:        input,
	}, nil
}

// qualifierEqual returns true if both qualifier pointers are equal
// (both nil, or both point to the same string value).
func qualifierEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
