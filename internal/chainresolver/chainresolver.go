// code-from-spec: ROOT/golang/implementation/chain/resolver@o1DttcJ985duoGFzuuyCDCh3zIs
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

func ChainResolve(targetLogicalName string) (*Chain, error) {
	var ancestors []*ChainItem
	var target *ChainItem

	if targetLogicalName == "ROOT" {
		filePath, err := logicalnames.LogicalNameToPath("ROOT")
		if err != nil {
			return nil, err
		}
		target = &ChainItem{LogicalName: "ROOT", FilePath: *filePath}
		ancestors = []*ChainItem{}
	} else {
		nameList := []string{targetLogicalName}
		current := targetLogicalName

		for {
			parent, err := logicalnames.LogicalNameGetParent(current)
			if err != nil {
				if errors.Is(err, logicalnames.ErrNoParent) {
					break
				}
				return nil, err
			}
			nameList = append(nameList, parent)
			current = parent
			if current == "ROOT" {
				break
			}
		}

		sort.Strings(nameList)

		items := make([]*ChainItem, 0, len(nameList))
		for _, name := range nameList {
			filePath, err := logicalnames.LogicalNameToPath(name)
			if err != nil {
				return nil, err
			}
			items = append(items, &ChainItem{LogicalName: name, FilePath: *filePath})
		}

		target = items[len(items)-1]
		ancestors = items[:len(items)-1]
	}

	targetFrontmatter, err := frontmatter.FrontmatterParse(&target.FilePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnreadableFrontmatter, err)
	}

	dependencies := []*ChainItem{}

	for _, entry := range targetFrontmatter.DependsOn {
		dep, err := resolveDependency(entry)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, dep)
	}

	sort.Slice(dependencies, func(i, j int) bool {
		pi := dependencies[i].FilePath.Value
		pj := dependencies[j].FilePath.Value
		if pi != pj {
			return pi < pj
		}
		qi := dependencies[i].Qualifier
		qj := dependencies[j].Qualifier
		if qi == "" && qj != "" {
			return true
		}
		if qi != "" && qj == "" {
			return false
		}
		return qi < qj
	})

	deduplicated := []*ChainItem{}
	for _, entry := range dependencies {
		if logicalnames.LogicalNameIsArtifact(entry.LogicalName) {
			if !containsArtifact(deduplicated, entry.LogicalName) {
				deduplicated = append(deduplicated, entry)
			}
		} else {
			if containsPathNoQualifier(deduplicated, entry.FilePath.Value) {
				continue
			} else if containsPathAndQualifier(deduplicated, entry.FilePath.Value, entry.Qualifier) {
				continue
			} else {
				deduplicated = append(deduplicated, entry)
			}
		}
	}
	dependencies = deduplicated

	external := make([]*frontmatter.FrontmatterExternal, len(targetFrontmatter.External))
	copy(external, targetFrontmatter.External)
	sort.Slice(external, func(i, j int) bool {
		return external[i].Path < external[j].Path
	})

	var input *ChainItem
	if targetFrontmatter.Input != "" {
		inputItem, err := resolveArtifactToChainItem(targetFrontmatter.Input)
		if err != nil {
			return nil, err
		}
		input = inputItem
	}

	return &Chain{
		Ancestors:    ancestors,
		Dependencies: dependencies,
		External:     external,
		Target:       target,
		Input:        input,
	}, nil
}

func resolveDependency(entry string) (*ChainItem, error) {
	switch {
	case len(entry) >= 5 && entry[:5] == "ROOT/":
		qualifier, _ := logicalnames.LogicalNameGetQualifier(entry)
		bareName := logicalnames.LogicalNameStripQualifier(entry)
		filePath, err := logicalnames.LogicalNameToPath(bareName)
		if err != nil {
			return nil, err
		}
		return &ChainItem{LogicalName: bareName, FilePath: *filePath, Qualifier: qualifier}, nil

	case len(entry) >= 9 && entry[:9] == "ARTIFACT/":
		return resolveArtifactToChainItem(entry)

	default:
		return nil, fmt.Errorf("%w: %q", ErrUnresolvableArtifact, entry)
	}
}

func resolveArtifactToChainItem(entry string) (*ChainItem, error) {
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
	if generatorFrontmatter.Output == "" {
		return nil, fmt.Errorf("%w: %q has no output", ErrUnresolvableArtifact, entry)
	}
	artifactPath := pathutils.PathCfs{Value: generatorFrontmatter.Output}
	return &ChainItem{LogicalName: entry, FilePath: artifactPath}, nil
}

func containsArtifact(items []*ChainItem, logicalName string) bool {
	for _, item := range items {
		if item.LogicalName == logicalName {
			return true
		}
	}
	return false
}

func containsPathNoQualifier(items []*ChainItem, filePath string) bool {
	for _, item := range items {
		if item.FilePath.Value == filePath && item.Qualifier == "" {
			return true
		}
	}
	return false
}

func containsPathAndQualifier(items []*ChainItem, filePath string, qualifier string) bool {
	for _, item := range items {
		if item.FilePath.Value == filePath && item.Qualifier == qualifier {
			return true
		}
	}
	return false
}
