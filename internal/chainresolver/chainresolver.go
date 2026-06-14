// code-from-spec: ROOT/golang/implementation/chain/resolver@HywIiX4W_gKf9rqBpHWBE_6R5gE
package chainresolver

import (
	"errors"
	"fmt"
	"sort"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

var ErrUnreadableFrontmatter = errors.New("a node's frontmatter cannot be parsed")
var ErrUnresolvableArtifact = errors.New("an ARTIFACT/ reference cannot be resolved")

type ChainItem struct {
	UnqualifiedLogicalName string
	FilePath               pathutils.PathCfs
	Qualifier              *string
}

type Chain struct {
	Ancestors    []*ChainItem
	Dependencies []*ChainItem
	Target       *ChainItem
	Input        *ChainItem
}

func ChainResolve(targetLogicalName string) (*Chain, error) {
	var ancestors []*ChainItem
	var target *ChainItem

	if targetLogicalName == "SPEC" {
		path, err := logicalnames.LogicalNameToPath("SPEC")
		if err != nil {
			return nil, fmt.Errorf("resolving SPEC path: %w", err)
		}
		target = &ChainItem{
			UnqualifiedLogicalName: "SPEC",
			FilePath:               *path,
			Qualifier:              nil,
		}
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
				return nil, fmt.Errorf("getting parent of %q: %w", current, err)
			}
			nameList = append(nameList, parent)
			current = parent
		}

		sort.Strings(nameList)

		var itemsList []*ChainItem
		for _, name := range nameList {
			path, err := logicalnames.LogicalNameToPath(name)
			if err != nil {
				return nil, fmt.Errorf("resolving path for %q: %w", name, err)
			}
			itemsList = append(itemsList, &ChainItem{
				UnqualifiedLogicalName: name,
				FilePath:               *path,
				Qualifier:              nil,
			})
		}

		target = itemsList[len(itemsList)-1]
		ancestors = itemsList[:len(itemsList)-1]
	}

	targetFrontmatter, err := frontmatter.FrontmatterParse(&target.FilePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnreadableFrontmatter, err)
	}

	var dependencies []*ChainItem

	for _, entry := range targetFrontmatter.DependsOn {
		item, err := resolveDependencyEntry(entry)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, item)
	}

	sort.SliceStable(dependencies, func(i, j int) bool {
		ni := dependencies[i].UnqualifiedLogicalName
		nj := dependencies[j].UnqualifiedLogicalName
		if ni != nj {
			return ni < nj
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

	var chainInput *ChainItem
	if targetFrontmatter.Input != "" {
		chainInput, err = resolveInputEntry(targetFrontmatter.Input)
		if err != nil {
			return nil, err
		}
	}

	return &Chain{
		Ancestors:    ancestors,
		Dependencies: dependencies,
		Target:       target,
		Input:        chainInput,
	}, nil
}

func resolveDependencyEntry(entry string) (*ChainItem, error) {
	if logicalnames.LogicalNameIsSpec(entry) {
		qualifier, hasQualifier := logicalnames.LogicalNameGetQualifier(entry)
		var qualifierPtr *string
		if hasQualifier {
			q := qualifier
			qualifierPtr = &q
		}
		bareName := logicalnames.LogicalNameStripQualifier(entry)
		path, err := logicalnames.LogicalNameToPath(bareName)
		if err != nil {
			return nil, fmt.Errorf("resolving path for dependency %q: %w", entry, err)
		}
		return &ChainItem{
			UnqualifiedLogicalName: bareName,
			FilePath:               *path,
			Qualifier:              qualifierPtr,
		}, nil
	}

	if logicalnames.LogicalNameIsArtifact(entry) {
		return resolveArtifactEntry(entry)
	}

	if logicalnames.LogicalNameIsExternal(entry) {
		path, err := logicalnames.LogicalNameExternalToPath(entry)
		if err != nil {
			return nil, fmt.Errorf("resolving external path for %q: %w", entry, err)
		}
		return &ChainItem{
			UnqualifiedLogicalName: entry,
			FilePath:               *path,
			Qualifier:              nil,
		}, nil
	}

	return nil, fmt.Errorf("%w: %q", ErrUnresolvableArtifact, entry)
}

func resolveArtifactEntry(entry string) (*ChainItem, error) {
	generatorName, err := logicalnames.LogicalNameGetArtifactGenerator(entry)
	if err != nil {
		return nil, fmt.Errorf("getting artifact generator for %q: %w", entry, err)
	}
	generatorPath, err := logicalnames.LogicalNameToPath(generatorName)
	if err != nil {
		return nil, fmt.Errorf("resolving generator path for %q: %w", entry, err)
	}
	generatorFrontmatter, err := frontmatter.FrontmatterParse(generatorPath)
	if err != nil {
		return nil, fmt.Errorf("%w: parsing generator frontmatter for %q: %w", ErrUnreadableFrontmatter, entry, err)
	}
	if generatorFrontmatter.Output == "" {
		return nil, fmt.Errorf("%w: generator %q has no output declared", ErrUnresolvableArtifact, generatorName)
	}
	return &ChainItem{
		UnqualifiedLogicalName: entry,
		FilePath:               pathutils.PathCfs{Value: generatorFrontmatter.Output},
		Qualifier:              nil,
	}, nil
}

func resolveInputEntry(inputEntry string) (*ChainItem, error) {
	if logicalnames.LogicalNameIsArtifact(inputEntry) {
		return resolveArtifactEntry(inputEntry)
	}

	if logicalnames.LogicalNameIsExternal(inputEntry) {
		path, err := logicalnames.LogicalNameExternalToPath(inputEntry)
		if err != nil {
			return nil, fmt.Errorf("resolving external input path for %q: %w", inputEntry, err)
		}
		return &ChainItem{
			UnqualifiedLogicalName: inputEntry,
			FilePath:               *path,
			Qualifier:              nil,
		}, nil
	}

	return nil, fmt.Errorf("%w: input %q is not ARTIFACT/ or EXTERNAL/", ErrUnresolvableArtifact, inputEntry)
}

func deduplicateDependencies(dependencies []*ChainItem) []*ChainItem {
	var deduped []*ChainItem

	for _, item := range dependencies {
		name := item.UnqualifiedLogicalName

		if logicalnames.LogicalNameIsSpec(name) {
			if item.Qualifier == nil {
				alreadyHasUnqualified := false
				for _, d := range deduped {
					if d.UnqualifiedLogicalName == name && d.Qualifier == nil {
						alreadyHasUnqualified = true
						break
					}
				}
				if alreadyHasUnqualified {
					continue
				}
				var filtered []*ChainItem
				for _, d := range deduped {
					if d.UnqualifiedLogicalName == name && d.Qualifier != nil {
						continue
					}
					filtered = append(filtered, d)
				}
				deduped = append(filtered, item)
			} else {
				skip := false
				for _, d := range deduped {
					if d.UnqualifiedLogicalName == name {
						if d.Qualifier == nil {
							skip = true
							break
						}
						if d.Qualifier != nil && *d.Qualifier == *item.Qualifier {
							skip = true
							break
						}
					}
				}
				if skip {
					continue
				}
				deduped = append(deduped, item)
			}
			continue
		}

		if logicalnames.LogicalNameIsArtifact(name) || logicalnames.LogicalNameIsExternal(name) {
			found := false
			for _, d := range deduped {
				if d.UnqualifiedLogicalName == name {
					found = true
					break
				}
			}
			if found {
				continue
			}
			deduped = append(deduped, item)
		}
	}

	return deduped
}
