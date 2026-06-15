// code-from-spec: SPEC/golang/implementation/utils/node_ranking@WMufa4SxI52My2ZW_4q-Lqc-TAU
package noderanking

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/logicalnames"
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
	logicalName string
	deps        []string
	rank        int
}

func NodeRankCompute(entries []*NodeRankInput) ([]*NodeRankEntry, []string, error) {
	entryMap := make(map[string]*rankEntry)

	for _, input := range entries {
		entryMap[input.LogicalName] = &rankEntry{
			logicalName: input.LogicalName,
			deps:        []string{},
			rank:        0,
		}

		if input.Frontmatter != nil && input.Frontmatter.Output != "" {
			bare := strings.TrimPrefix(input.LogicalName, "SPEC/")
			artifactName := "ARTIFACT/" + bare
			entryMap[artifactName] = &rankEntry{
				logicalName: artifactName,
				deps:        []string{input.LogicalName},
				rank:        0,
			}
		}
	}

	for _, input := range entries {
		entry := entryMap[input.LogicalName]

		if input.LogicalName == "SPEC" {
			continue
		}

		parent, err := logicalnames.LogicalNameGetParent(input.LogicalName)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %w", ErrUnresolvableReference, err)
		}
		entry.deps = append(entry.deps, parent)

		if input.Frontmatter == nil {
			continue
		}

		for _, ref := range input.Frontmatter.DependsOn {
			if logicalnames.LogicalNameIsSpec(ref) {
				bareName := logicalnames.LogicalNameStripQualifier(ref)
				if _, ok := entryMap[bareName]; !ok {
					return nil, nil, fmt.Errorf("%w: %s", ErrUnresolvableReference, ref)
				}
				entry.deps = append(entry.deps, bareName)
			} else if strings.HasPrefix(ref, "ARTIFACT/") {
				if _, ok := entryMap[ref]; !ok {
					return nil, nil, fmt.Errorf("%w: %s", ErrUnresolvableReference, ref)
				}
				entry.deps = append(entry.deps, ref)
			} else if strings.HasPrefix(ref, "EXTERNAL/") {
				continue
			} else {
				return nil, nil, fmt.Errorf("%w: %s", ErrUnresolvableReference, ref)
			}
		}

		if input.Frontmatter.Input != "" {
			inp := input.Frontmatter.Input
			if strings.HasPrefix(inp, "ARTIFACT/") {
				if _, ok := entryMap[inp]; !ok {
					return nil, nil, fmt.Errorf("%w: %s", ErrUnresolvableReference, inp)
				}
				entry.deps = append(entry.deps, inp)
			} else if strings.HasPrefix(inp, "EXTERNAL/") {
				// skip
			}
		}
	}

	n := len(entryMap)
	cycles := []string{}
	changed := false

	for i := 1; i <= n; i++ {
		changed = false

		for _, entry := range entryMap {
			if entry.logicalName == "SPEC" {
				continue
			}

			maxDepRank := -1
			for _, dep := range entry.deps {
				if depEntry, ok := entryMap[dep]; ok {
					if depEntry.rank > maxDepRank {
						maxDepRank = depEntry.rank
					}
				}
			}

			newRank := 1 + maxDepRank
			if maxDepRank == -1 {
				newRank = 1
			}

			if newRank > entry.rank {
				entry.rank = newRank
				changed = true
				if i == n {
					cycles = append(cycles, entry.logicalName)
				}
			}
		}

		if !changed {
			break
		}
	}

	if !changed {
		cycles = []string{}
	}

	ranked := make([]*NodeRankEntry, 0, len(entryMap))
	for _, entry := range entryMap {
		ranked = append(ranked, &NodeRankEntry{
			LogicalName: entry.logicalName,
			Rank:        entry.rank,
		})
	}

	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Rank != ranked[j].Rank {
			return ranked[i].Rank < ranked[j].Rank
		}
		return ranked[i].LogicalName < ranked[j].LogicalName
	})

	return ranked, cycles, nil
}
