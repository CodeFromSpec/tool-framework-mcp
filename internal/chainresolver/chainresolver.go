// code-from-spec: SPEC/golang/implementation/chain/resolver@Qz-P-pfgA7bY5mrs3imTcIItLtQ
package chainresolver

import (
	"errors"
	"fmt"
	"sort"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")
var ErrUnresolvableArtifact = errors.New("unresolvable artifact")

type ChainItem struct {
	UnqualifiedLogicalName string
	FilePath               pathutils.PathCfs
	Qualifier              string
}

type Chain struct {
	Ancestors    []*ChainItem
	Dependencies []*ChainItem
	Target       *ChainItem
	Input        *ChainItem
}

func ChainResolve(targetLogicalName string) (*Chain, error) {
	ancestors, target, err := resolveAncestorsAndTarget(targetLogicalName)
	if err != nil {
		return nil, err
	}

	fm, err := frontmatter.FrontmatterParse(&target.FilePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnreadableFrontmatter, err)
	}

	deps, err := resolveDependencies(fm)
	if err != nil {
		return nil, err
	}

	deps = deduplicateDependencies(deps)

	input, err := resolveInput(fm)
	if err != nil {
		return nil, err
	}

	return &Chain{
		Ancestors:    ancestors,
		Dependencies: deps,
		Target:       target,
		Input:        input,
	}, nil
}

func resolveAncestorsAndTarget(targetLogicalName string) ([]*ChainItem, *ChainItem, error) {
	if targetLogicalName == "SPEC" {
		path, err := logicalnames.LogicalNameToPath(targetLogicalName)
		if err != nil {
			return nil, nil, err
		}
		item := &ChainItem{
			UnqualifiedLogicalName: targetLogicalName,
			FilePath:               *path,
			Qualifier:              "",
		}
		return []*ChainItem{}, item, nil
	}

	names := []string{targetLogicalName}
	current := targetLogicalName
	for {
		parent, err := logicalnames.LogicalNameGetParent(current)
		if err != nil {
			return nil, nil, err
		}
		names = append(names, parent)
		current = parent
		if current == "SPEC" {
			break
		}
	}

	sort.Strings(names)

	items := make([]*ChainItem, 0, len(names))
	for _, name := range names {
		path, err := logicalnames.LogicalNameToPath(name)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, &ChainItem{
			UnqualifiedLogicalName: name,
			FilePath:               *path,
			Qualifier:              "",
		})
	}

	target := items[len(items)-1]
	ancestors := items[:len(items)-1]
	return ancestors, target, nil
}

func resolveDependencies(fm *frontmatter.Frontmatter) ([]*ChainItem, error) {
	deps := make([]*ChainItem, 0, len(fm.DependsOn))

	for _, entry := range fm.DependsOn {
		item, err := resolveEntry(entry)
		if err != nil {
			return nil, err
		}
		deps = append(deps, item)
	}

	sort.Slice(deps, func(i, j int) bool {
		if deps[i].UnqualifiedLogicalName != deps[j].UnqualifiedLogicalName {
			return deps[i].UnqualifiedLogicalName < deps[j].UnqualifiedLogicalName
		}
		if deps[i].Qualifier == "" && deps[j].Qualifier != "" {
			return true
		}
		if deps[i].Qualifier != "" && deps[j].Qualifier == "" {
			return false
		}
		return deps[i].Qualifier < deps[j].Qualifier
	})

	return deps, nil
}

func resolveEntry(entry string) (*ChainItem, error) {
	if logicalnames.LogicalNameIsSpec(entry) {
		qualifier, hasQualifier := logicalnames.LogicalNameGetQualifier(entry)
		bare := logicalnames.LogicalNameStripQualifier(entry)
		path, err := logicalnames.LogicalNameToPath(bare)
		if err != nil {
			return nil, err
		}
		item := &ChainItem{
			UnqualifiedLogicalName: bare,
			FilePath:               *path,
			Qualifier:              "",
		}
		if hasQualifier {
			item.Qualifier = qualifier
		}
		return item, nil
	}

	if logicalnames.LogicalNameIsArtifact(entry) {
		return resolveArtifactEntry(entry)
	}

	if logicalnames.LogicalNameIsExternal(entry) {
		path, err := logicalnames.LogicalNameExternalToPath(entry)
		if err != nil {
			return nil, err
		}
		return &ChainItem{
			UnqualifiedLogicalName: entry,
			FilePath:               *path,
			Qualifier:              "",
		}, nil
	}

	return nil, fmt.Errorf("%w: %s", ErrUnresolvableArtifact, entry)
}

func resolveArtifactEntry(entry string) (*ChainItem, error) {
	generatorName, err := logicalnames.LogicalNameGetArtifactGenerator(entry)
	if err != nil {
		return nil, err
	}
	generatorPath, err := logicalnames.LogicalNameToPath(generatorName)
	if err != nil {
		return nil, err
	}
	generatorFm, err := frontmatter.FrontmatterParse(generatorPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnreadableFrontmatter, err)
	}
	if generatorFm.Output == "" {
		return nil, fmt.Errorf("%w: %s has no output", ErrUnresolvableArtifact, entry)
	}
	return &ChainItem{
		UnqualifiedLogicalName: entry,
		FilePath:               pathutils.PathCfs{Value: generatorFm.Output},
		Qualifier:              "",
	}, nil
}

func deduplicateDependencies(deps []*ChainItem) []*ChainItem {
	result := make([]*ChainItem, 0, len(deps))

	for _, dep := range deps {
		if logicalnames.LogicalNameIsSpec(dep.UnqualifiedLogicalName) {
			if isSpecDuplicate(result, dep) {
				continue
			}
		} else {
			if isNameDuplicate(result, dep.UnqualifiedLogicalName) {
				continue
			}
		}
		result = append(result, dep)
	}

	return result
}

func isSpecDuplicate(existing []*ChainItem, candidate *ChainItem) bool {
	for _, e := range existing {
		if e.UnqualifiedLogicalName != candidate.UnqualifiedLogicalName {
			continue
		}
		if e.Qualifier == "" {
			return true
		}
		if candidate.Qualifier != "" && e.Qualifier == candidate.Qualifier {
			return true
		}
	}
	return false
}

func isNameDuplicate(existing []*ChainItem, name string) bool {
	for _, e := range existing {
		if e.UnqualifiedLogicalName == name {
			return true
		}
	}
	return false
}

func resolveInput(fm *frontmatter.Frontmatter) (*ChainItem, error) {
	if fm.Input == "" {
		return nil, nil
	}

	entry := fm.Input

	if logicalnames.LogicalNameIsArtifact(entry) {
		return resolveArtifactEntry(entry)
	}

	if logicalnames.LogicalNameIsExternal(entry) {
		path, err := logicalnames.LogicalNameExternalToPath(entry)
		if err != nil {
			return nil, err
		}
		return &ChainItem{
			UnqualifiedLogicalName: entry,
			FilePath:               *path,
			Qualifier:              "",
		}, nil
	}

	return nil, nil
}
