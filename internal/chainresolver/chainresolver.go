// code-from-spec: SPEC/golang/implementation/chain/resolver@h6W4UGmwIZ33dImxZlUjHPgfC6o
package chainresolver

import (
	"errors"
	"fmt"
	"sort"
	"strings"

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
	ancestors, target, targetLn, err := resolveAncestorsAndTarget(targetLogicalName)
	if err != nil {
		return nil, err
	}

	fm, err := frontmatter.FrontmatterParse(pathutils.PathCfs{Value: targetLn.Path})
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

func resolveAncestorsAndTarget(targetLogicalName string) ([]*ChainItem, *ChainItem, *logicalnames.LogicalName, error) {
	targetLn, err := logicalnames.LogicalNameParse(targetLogicalName)
	if err != nil {
		return nil, nil, nil, err
	}

	if targetLn.Parent == nil {
		target := &ChainItem{
			UnqualifiedLogicalName: targetLn.Name,
			FilePath:               pathutils.PathCfs{Value: targetLn.Path},
			Qualifier:              "",
		}
		return []*ChainItem{}, target, targetLn, nil
	}

	names := []string{targetLogicalName}
	currentLn := targetLn
	for {
		if currentLn.Parent == nil {
			break
		}
		names = append(names, *currentLn.Parent)
		parentLn, err := logicalnames.LogicalNameParse(*currentLn.Parent)
		if err != nil {
			return nil, nil, nil, err
		}
		currentLn = parentLn
	}

	sort.Strings(names)

	items := make([]*ChainItem, 0, len(names))
	for _, name := range names {
		ln, err := logicalnames.LogicalNameParse(name)
		if err != nil {
			return nil, nil, nil, err
		}
		items = append(items, &ChainItem{
			UnqualifiedLogicalName: ln.Name,
			FilePath:               pathutils.PathCfs{Value: ln.Path},
			Qualifier:              "",
		})
	}

	target := items[len(items)-1]
	ancestors := items[:len(items)-1]
	return ancestors, target, targetLn, nil
}

func resolveDependencies(fm *frontmatter.Frontmatter) ([]*ChainItem, error) {
	deps := make([]*ChainItem, 0, len(fm.DependsOn))

	for _, entry := range fm.DependsOn {
		ln, err := logicalnames.LogicalNameParse(entry)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrUnresolvableArtifact, err)
		}

		var item *ChainItem
		switch ln.Type {
		case logicalnames.NodeTypeSpec:
			qualifier := ""
			if ln.Qualifier != nil {
				qualifier = *ln.Qualifier
			}
			item = &ChainItem{
				UnqualifiedLogicalName: ln.Name,
				FilePath:               pathutils.PathCfs{Value: ln.Path},
				Qualifier:              qualifier,
			}
		case logicalnames.NodeTypeArtifact:
			item = &ChainItem{
				UnqualifiedLogicalName: ln.Name,
				FilePath:               pathutils.PathCfs{Value: ln.Path},
				Qualifier:              "",
			}
		case logicalnames.NodeTypeExternal:
			item = &ChainItem{
				UnqualifiedLogicalName: ln.Name,
				FilePath:               pathutils.PathCfs{Value: ln.Path},
				Qualifier:              "",
			}
		default:
			return nil, fmt.Errorf("%w: %s", ErrUnresolvableArtifact, entry)
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

func deduplicateDependencies(deps []*ChainItem) []*ChainItem {
	result := make([]*ChainItem, 0, len(deps))

	for _, dep := range deps {
		if strings.HasPrefix(dep.UnqualifiedLogicalName, "SPEC/") || dep.UnqualifiedLogicalName == "SPEC" {
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

	ln, err := logicalnames.LogicalNameParse(fm.Input)
	if err != nil {
		return nil, err
	}

	return &ChainItem{
		UnqualifiedLogicalName: ln.Name,
		FilePath:               pathutils.PathCfs{Value: ln.Path},
		Qualifier:              "",
	}, nil
}
