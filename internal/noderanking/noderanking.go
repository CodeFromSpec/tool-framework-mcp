package noderanking

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"
)

var ErrUnresolvableReference = errors.New("unresolvable reference")

type NodeRankEntry struct {
	Reference parsing.CfsReference
	Rank      int
}

type rankEntry struct {
	ref  parsing.CfsReference
	deps []string
	rank int
}

func unqualifiedName(ref string) string {
	idx := strings.Index(ref, "(")
	if idx >= 0 {
		return ref[:idx]
	}
	return ref
}

func NodeRankCompute(entries []parsing.Node) ([]NodeRankEntry, []string, error) {
	entryMap := make(map[string]*rankEntry)

	for _, node := range entries {
		entryMap[node.Reference.LogicalName] = &rankEntry{
			ref:  node.Reference,
			deps: []string{},
			rank: 0,
		}

		if node.Frontmatter != nil && node.Frontmatter.Output != nil {
			bare := strings.TrimPrefix(node.Reference.LogicalName, "SPEC/")
			artifactName := "ARTIFACT/" + bare
			parentName := node.Reference.LogicalName
			artifactRef := parsing.CfsReference{
				NodeType:    parsing.CfsNodeTypeArtifact,
				LogicalName: artifactName,
				Qualifier:   nil,
				Path:        *node.Frontmatter.Output,
				ParentName:  &parentName,
			}
			entryMap[artifactName] = &rankEntry{
				ref:  artifactRef,
				deps: []string{node.Reference.LogicalName},
				rank: 0,
			}
		}
	}

	for _, node := range entries {
		entry := entryMap[node.Reference.LogicalName]

		if node.Reference.ParentName != nil {
			entry.deps = append(entry.deps, *node.Reference.ParentName)
		}

		if node.Frontmatter == nil {
			continue
		}

		for _, ref := range node.Frontmatter.DependsOn {
			if strings.HasPrefix(ref, "SPEC/") {
				unqualified := unqualifiedName(ref)
				if _, ok := entryMap[unqualified]; !ok {
					return nil, nil, fmt.Errorf("%w: %s", ErrUnresolvableReference, ref)
				}
				entry.deps = append(entry.deps, unqualified)
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

		if node.Frontmatter.Input != nil {
			inp := *node.Frontmatter.Input
			if strings.HasPrefix(inp, "SPEC/") {
				unqualified := unqualifiedName(inp)
				if _, ok := entryMap[unqualified]; !ok {
					return nil, nil, fmt.Errorf("%w: %s", ErrUnresolvableReference, inp)
				}
				entry.deps = append(entry.deps, unqualified)
			} else if strings.HasPrefix(inp, "ARTIFACT/") {
				if _, ok := entryMap[inp]; !ok {
					return nil, nil, fmt.Errorf("%w: %s", ErrUnresolvableReference, inp)
				}
				entry.deps = append(entry.deps, inp)
			} else if strings.HasPrefix(inp, "EXTERNAL/") {
				continue
			}
		}
	}

	n := len(entryMap)
	cycleCandidates := []string{}
	changed := false

	for i := 1; i <= n; i++ {
		changed = false

		for logicalName, entry := range entryMap {
			if len(entry.deps) == 0 {
				continue
			}

			maxDepRank := 0
			for _, dep := range entry.deps {
				if depEntry, ok := entryMap[dep]; ok {
					if depEntry.rank > maxDepRank {
						maxDepRank = depEntry.rank
					}
				}
			}

			newRank := 1 + maxDepRank

			if newRank > entry.rank {
				entry.rank = newRank
				changed = true
				if i == n {
					cycleCandidates = append(cycleCandidates, logicalName)
				}
			}
		}

		if !changed {
			break
		}
	}

	cycles := []string{}
	if changed {
		cycles = cycleCandidates
	}

	ranked := make([]NodeRankEntry, 0, len(entryMap))
	for _, entry := range entryMap {
		ranked = append(ranked, NodeRankEntry{
			Reference: entry.ref,
			Rank:      entry.rank,
		})
	}

	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Rank != ranked[j].Rank {
			return ranked[i].Rank < ranked[j].Rank
		}
		return ranked[i].Reference.LogicalName < ranked[j].Reference.LogicalName
	})

	return ranked, cycles, nil
}
</content>
</invoke>
