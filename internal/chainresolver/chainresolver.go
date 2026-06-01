// code-from-spec: ROOT/golang/implementation/chain/resolver@lXA8tT_fJqejzd9sknrk2wsWFBE
package chainresolver

import (
	"errors"
	"fmt"
	"sort"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")

var ErrUnresolvableArtifact = errors.New("unresolvable artifact")

type ChainItem struct {
	LogicalName string
	FilePath    pathutils.PathCfs
	Qualifier   *string
}

type Chain struct {
	Ancestors    []*ChainItem
	Dependencies []*ChainItem
	External     []*frontmatter.FrontmatterExternal
	Target       *ChainItem
	Input        *ChainItem
}

func ChainResolve(targetLogicalName string) (*Chain, error) {
	ancestors, target, err := resolveAncestorsAndTarget(targetLogicalName)
	if err != nil {
		return nil, err
	}

	targetFrontmatter, err := frontmatter.FrontmatterParse(&target.FilePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnreadableFrontmatter, err)
	}

	dependencies, err := resolveDependencies(targetFrontmatter)
	if err != nil {
		return nil, err
	}

	external := collectExternal(targetFrontmatter)

	input, err := resolveInput(targetFrontmatter)
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

func resolveAncestorsAndTarget(targetLogicalName string) ([]*ChainItem, *ChainItem, error) {
	if targetLogicalName == "ROOT" {
		path, err := logicalnames.LogicalNameToPath("ROOT")
		if err != nil {
			return nil, nil, err
		}
		item := &ChainItem{LogicalName: "ROOT", FilePath: *path, Qualifier: nil}
		return []*ChainItem{}, item, nil
	}

	nameList := []string{targetLogicalName}
	current := targetLogicalName
	for {
		parent, err := logicalnames.LogicalNameGetParent(current)
		if err != nil {
			return nil, nil, err
		}
		nameList = append(nameList, parent)
		if parent == "ROOT" {
			break
		}
		current = parent
	}

	sort.Strings(nameList)

	items := make([]*ChainItem, 0, len(nameList))
	for _, name := range nameList {
		path, err := logicalnames.LogicalNameToPath(name)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, &ChainItem{LogicalName: name, FilePath: *path, Qualifier: nil})
	}

	target := items[len(items)-1]
	ancestors := items[:len(items)-1]
	return ancestors, target, nil
}

func resolveDependencies(targetFrontmatter *frontmatter.Frontmatter) ([]*ChainItem, error) {
	var dependencies []*ChainItem

	for _, entry := range targetFrontmatter.DependsOn {
		switch {
		case len(entry) >= 5 && entry[:5] == "ROOT/":
			item, err := resolveRootDependency(entry)
			if err != nil {
				return nil, err
			}
			dependencies = append(dependencies, item)

		case len(entry) >= 9 && entry[:9] == "ARTIFACT/":
			item, err := resolveArtifactDependency(entry)
			if err != nil {
				return nil, err
			}
			dependencies = append(dependencies, item)

		default:
			return nil, fmt.Errorf("%w: unrecognized prefix in depends_on entry %q", ErrUnresolvableArtifact, entry)
		}
	}

	sort.Slice(dependencies, func(i, j int) bool {
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

	dependencies = deduplicateDependencies(dependencies)
	return dependencies, nil
}

func resolveRootDependency(entry string) (*ChainItem, error) {
	qualifier, present := logicalnames.LogicalNameGetQualifier(entry)
	bareName := logicalnames.LogicalNameStripQualifier(entry)

	path, err := logicalnames.LogicalNameToPath(bareName)
	if err != nil {
		return nil, err
	}

	item := &ChainItem{LogicalName: bareName, FilePath: *path}
	if present {
		q := qualifier
		item.Qualifier = &q
	}
	return item, nil
}

func resolveArtifactDependency(entry string) (*ChainItem, error) {
	qualifier, present := logicalnames.LogicalNameGetQualifier(entry)
	if !present {
		return nil, fmt.Errorf("%w: ARTIFACT/ reference %q is missing a qualifier", ErrUnresolvableArtifact, entry)
	}

	generatorName, err := logicalnames.LogicalNameGetArtifactGenerator(entry)
	if err != nil {
		return nil, err
	}

	generatorPath, err := logicalnames.LogicalNameToPath(generatorName)
	if err != nil {
		return nil, err
	}

	generatorFrontmatter, err := frontmatter.FrontmatterParse(generatorPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnreadableFrontmatter, err)
	}

	outputPath, err := findOutputPath(generatorFrontmatter, qualifier, entry)
	if err != nil {
		return nil, err
	}

	q := qualifier
	return &ChainItem{
		LogicalName: entry,
		FilePath:    pathutils.PathCfs{Value: outputPath},
		Qualifier:   &q,
	}, nil
}

func findOutputPath(fm *frontmatter.Frontmatter, qualifier string, ref string) (string, error) {
	for _, o := range fm.Outputs {
		if o.ID == qualifier {
			return o.Path, nil
		}
	}
	return "", fmt.Errorf("%w: no output with id %q in %q", ErrUnresolvableArtifact, qualifier, ref)
}

func deduplicateDependencies(dependencies []*ChainItem) []*ChainItem {
	var deduped []*ChainItem

	for _, entry := range dependencies {
		if logicalnames.LogicalNameIsArtifact(entry.LogicalName) {
			if !containsArtifact(deduped, entry.LogicalName) {
				deduped = append(deduped, entry)
			}
		} else {
			if !containsRootEntry(deduped, entry) {
				deduped = append(deduped, entry)
			}
		}
	}

	return deduped
}

func containsArtifact(items []*ChainItem, logicalName string) bool {
	for _, item := range items {
		if item.LogicalName == logicalName {
			return true
		}
	}
	return false
}

func containsRootEntry(items []*ChainItem, candidate *ChainItem) bool {
	for _, item := range items {
		if item.FilePath.Value != candidate.FilePath.Value {
			continue
		}
		if item.Qualifier == nil {
			return true
		}
		if candidate.Qualifier != nil && *item.Qualifier == *candidate.Qualifier {
			return true
		}
	}
	return false
}

func collectExternal(targetFrontmatter *frontmatter.Frontmatter) []*frontmatter.FrontmatterExternal {
	external := make([]*frontmatter.FrontmatterExternal, len(targetFrontmatter.External))
	copy(external, targetFrontmatter.External)

	sort.Slice(external, func(i, j int) bool {
		return external[i].Path < external[j].Path
	})

	return external
}

func resolveInput(targetFrontmatter *frontmatter.Frontmatter) (*ChainItem, error) {
	if targetFrontmatter.Input == "" {
		return nil, nil
	}

	entry := targetFrontmatter.Input

	qualifier, present := logicalnames.LogicalNameGetQualifier(entry)
	if !present {
		return nil, fmt.Errorf("%w: input reference %q is missing a qualifier", ErrUnresolvableArtifact, entry)
	}

	generatorName, err := logicalnames.LogicalNameGetArtifactGenerator(entry)
	if err != nil {
		return nil, err
	}

	generatorPath, err := logicalnames.LogicalNameToPath(generatorName)
	if err != nil {
		return nil, err
	}

	generatorFrontmatter, err := frontmatter.FrontmatterParse(generatorPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnreadableFrontmatter, err)
	}

	outputPath, err := findOutputPath(generatorFrontmatter, qualifier, entry)
	if err != nil {
		return nil, err
	}

	q := qualifier
	return &ChainItem{
		LogicalName: entry,
		FilePath:    pathutils.PathCfs{Value: outputPath},
		Qualifier:   &q,
	}, nil
}
