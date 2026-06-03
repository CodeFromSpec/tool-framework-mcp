// code-from-spec: ROOT/golang/implementation/utils/node_ranking@MkQbXkMeoJSHoLTZmPF1a4rERNY
package noderanking

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
)

var ErrUnresolvableReference = errors.New("unresolvable reference")

type NodeRankInput struct {
	LogicalName string
	Frontmatter *frontmatter.Frontmatter
}

type NodeRankEntry struct {
	LogicalName string
	Rank        int
}

type rankEntry struct {
	logicalName    string
	dependencyKeys []string
	rank           int
}

func NodeRankCompute(entries []*NodeRankInput) (ranked []*NodeRankEntry, cycles []string, err error) {
	entryMap := make(map[string]*rankEntry)

	for _, input := range entries {
		entryMap[input.LogicalName] = &rankEntry{
			logicalName: input.LogicalName,
		}
		if input.Frontmatter != nil && input.Frontmatter.Output != "" {
			bare := strings.TrimPrefix(input.LogicalName, "ROOT/")
			artifactName := "ARTIFACT/" + bare
			entryMap[artifactName] = &rankEntry{
				logicalName: artifactName,
			}
		}
	}

	for _, input := range entries {
		e := entryMap[input.LogicalName]

		if input.LogicalName != "ROOT" {
			parent, err := logicalnames.LogicalNameGetParent(input.LogicalName)
			if err != nil {
				return nil, nil, fmt.Errorf("%w: %w", ErrUnresolvableReference, err)
			}
			e.dependencyKeys = append(e.dependencyKeys, parent)
		}

		if input.Frontmatter != nil {
			for _, dep := range input.Frontmatter.DependsOn {
				if strings.HasPrefix(dep, "ARTIFACT/") {
					e.dependencyKeys = append(e.dependencyKeys, dep)
				} else {
					bare := logicalnames.LogicalNameStripQualifier(dep)
					e.dependencyKeys = append(e.dependencyKeys, bare)
				}
			}

			if input.Frontmatter.Input != "" {
				e.dependencyKeys = append(e.dependencyKeys, input.Frontmatter.Input)
			}
		}
	}

	for key, e := range entryMap {
		if strings.HasPrefix(key, "ARTIFACT/") {
			generatorName := "ROOT/" + strings.TrimPrefix(key, "ARTIFACT/")
			e.dependencyKeys = append(e.dependencyKeys, generatorName)
		}
	}

	for _, e := range entryMap {
		for _, dep := range e.dependencyKeys {
			if _, ok := entryMap[dep]; !ok {
				return nil, nil, fmt.Errorf("%w: %q not found", ErrUnresolvableReference, dep)
			}
		}
	}

	n := len(entryMap)
	var lastUpdated []string

	for pass := 0; pass < n; pass++ {
		changed := false
		lastUpdated = nil

		for _, e := range entryMap {
			if e.logicalName == "ROOT" {
				continue
			}

			maxDepRank := 0
			for _, dep := range e.dependencyKeys {
				depEntry := entryMap[dep]
				if depEntry.rank > maxDepRank {
					maxDepRank = depEntry.rank
				}
			}

			candidate := 1 + maxDepRank
			if candidate > e.rank {
				e.rank = candidate
				changed = true
				lastUpdated = append(lastUpdated, e.logicalName)
			}
		}

		if !changed {
			lastUpdated = nil
			break
		}
	}

	ranked = make([]*NodeRankEntry, 0, len(entryMap))
	for _, e := range entryMap {
		ranked = append(ranked, &NodeRankEntry{
			LogicalName: e.logicalName,
			Rank:        e.rank,
		})
	}

	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Rank != ranked[j].Rank {
			return ranked[i].Rank < ranked[j].Rank
		}
		return ranked[i].LogicalName < ranked[j].LogicalName
	})

	return ranked, lastUpdated, nil
}
