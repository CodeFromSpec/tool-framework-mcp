// code-from-spec: ROOT/golang/implementation/chain/resolver@itzBOrRn6Alfg3obRuWEGVOU7Og
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
	Qualifier   string
}

type Chain struct {
	Ancestors    []*ChainItem
	Dependencies []*ChainItem
	External     []*frontmatter.FrontmatterExternal
	Target       *ChainItem
	Input        *ChainItem
}

func ChainResolve(target_logical_name string) (*Chain, error) {
	ancestors, target, err := resolveAncestorsAndTarget(target_logical_name)
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

	ext := make([]*frontmatter.FrontmatterExternal, len(fm.External))
	copy(ext, fm.External)
	sort.Slice(ext, func(i, j int) bool {
		return ext[i].Path < ext[j].Path
	})

	var input *ChainItem
	if fm.Input != "" {
		input, err = resolveArtifactItem(fm.Input)
		if err != nil {
			return nil, err
		}
	}

	return &Chain{
		Ancestors:    ancestors,
		Dependencies: deps,
		External:     ext,
		Target:       target,
		Input:        input,
	}, nil
}

func resolveAncestorsAndTarget(target_logical_name string) ([]*ChainItem, *ChainItem, error) {
	if target_logical_name == "ROOT" {
		path, err := logicalnames.LogicalNameToPath("ROOT")
		if err != nil {
			return nil, nil, err
		}
		item := &ChainItem{LogicalName: "ROOT", FilePath: *path}
		return []*ChainItem{}, item, nil
	}

	collected := []string{target_logical_name}
	current := target_logical_name
	for {
		parent, err := logicalnames.LogicalNameGetParent(current)
		if err != nil {
			return nil, nil, err
		}
		collected = append(collected, parent)
		if parent == "ROOT" {
			break
		}
		current = parent
	}

	sort.Strings(collected)

	items := make([]*ChainItem, 0, len(collected))
	for _, name := range collected {
		path, err := logicalnames.LogicalNameToPath(name)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, &ChainItem{LogicalName: name, FilePath: *path})
	}

	target := items[len(items)-1]
	ancestors := items[:len(items)-1]
	return ancestors, target, nil
}

func resolveDependencies(fm *frontmatter.Frontmatter) ([]*ChainItem, error) {
	var deps []*ChainItem

	for _, entry := range fm.DependsOn {
		if !hasPrefix(entry, "ROOT/") && !hasPrefix(entry, "ARTIFACT/") {
			return nil, fmt.Errorf("%w: %s", ErrUnresolvableArtifact, entry)
		}

		if hasPrefix(entry, "ROOT/") {
			qualifier, _ := logicalnames.LogicalNameGetQualifier(entry)
			bare := logicalnames.LogicalNameStripQualifier(entry)
			path, err := logicalnames.LogicalNameToPath(bare)
			if err != nil {
				return nil, err
			}
			deps = append(deps, &ChainItem{
				LogicalName: bare,
				FilePath:    *path,
				Qualifier:   qualifier,
			})
			continue
		}

		item, err := resolveArtifactItem(entry)
		if err != nil {
			return nil, err
		}
		deps = append(deps, item)
	}

	sort.Slice(deps, func(i, j int) bool {
		if deps[i].FilePath.Value != deps[j].FilePath.Value {
			return deps[i].FilePath.Value < deps[j].FilePath.Value
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

func resolveArtifactItem(entry string) (*ChainItem, error) {
	generator, err := logicalnames.LogicalNameGetArtifactGenerator(entry)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnresolvableArtifact, err)
	}
	genPath, err := logicalnames.LogicalNameToPath(generator)
	if err != nil {
		return nil, err
	}
	genFM, err := frontmatter.FrontmatterParse(genPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnreadableFrontmatter, err)
	}
	if genFM.Output == "" {
		return nil, fmt.Errorf("%w: node %q has no output", ErrUnresolvableArtifact, generator)
	}
	outputPath := &pathutils.PathCfs{Value: genFM.Output}
	return &ChainItem{
		LogicalName: entry,
		FilePath:    *outputPath,
	}, nil
}

func deduplicateDependencies(deps []*ChainItem) []*ChainItem {
	var result []*ChainItem

	for _, candidate := range deps {
		if logicalnames.LogicalNameIsArtifact(candidate.LogicalName) {
			if !containsArtifact(result, candidate.LogicalName) {
				result = append(result, candidate)
			}
			continue
		}

		if candidate.Qualifier == "" {
			removeRedundant(&result, candidate.FilePath.Value)
			if !containsRoot(result, candidate.FilePath.Value, "") {
				result = append(result, candidate)
			}
		} else {
			if !rootWithNoQualifierExists(result, candidate.FilePath.Value) {
				if !containsRoot(result, candidate.FilePath.Value, candidate.Qualifier) {
					result = append(result, candidate)
				}
			}
		}
	}

	return result
}

func containsArtifact(deps []*ChainItem, logicalName string) bool {
	for _, d := range deps {
		if d.LogicalName == logicalName {
			return true
		}
	}
	return false
}

func containsRoot(deps []*ChainItem, filePath string, qualifier string) bool {
	for _, d := range deps {
		if !logicalnames.LogicalNameIsArtifact(d.LogicalName) && d.FilePath.Value == filePath && d.Qualifier == qualifier {
			return true
		}
	}
	return false
}

func rootWithNoQualifierExists(deps []*ChainItem, filePath string) bool {
	for _, d := range deps {
		if !logicalnames.LogicalNameIsArtifact(d.LogicalName) && d.FilePath.Value == filePath && d.Qualifier == "" {
			return true
		}
	}
	return false
}

func removeRedundant(deps *[]*ChainItem, filePath string) {
	var filtered []*ChainItem
	for _, d := range *deps {
		if !logicalnames.LogicalNameIsArtifact(d.LogicalName) && d.FilePath.Value == filePath && d.Qualifier != "" {
			continue
		}
		filtered = append(filtered, d)
	}
	*deps = filtered
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
