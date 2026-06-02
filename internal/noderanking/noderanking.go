// code-from-spec: ROOT/golang/implementation/utils/node_ranking@YmQSZlncXpZbHzCoW8Obzav_jHQ
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
	Frontmatter frontmatter.Frontmatter
}

type NodeRankEntry struct {
	LogicalName string
	Rank        int
}

type rankState struct {
	dependencyKeys []string
	rank           int
}

func NodeRankCompute(entries []*NodeRankInput) (ranked []*NodeRankEntry, cycles []string, err error) {
	stateMap := make(map[string]*rankState)

	for _, entry := range entries {
		stateMap[entry.LogicalName] = &rankState{}
		if entry.Frontmatter.Output != "" {
			artifactName := "ARTIFACT/" + strings.TrimPrefix(entry.LogicalName, "ROOT/")
			stateMap[artifactName] = &rankState{}
		}
	}

	for _, entry := range entries {
		state := stateMap[entry.LogicalName]

		if entry.LogicalName != "ROOT" {
			parent, err := logicalnames.LogicalNameGetParent(entry.LogicalName)
			if err != nil {
				return nil, nil, fmt.Errorf("%w: %w", ErrUnresolvableReference, err)
			}
			state.dependencyKeys = append(state.dependencyKeys, parent)
		}

		for _, dep := range entry.Frontmatter.DependsOn {
			var key string
			if strings.HasPrefix(dep, "ARTIFACT/") {
				key = dep
			} else {
				key = logicalnames.LogicalNameStripQualifier(dep)
			}
			state.dependencyKeys = append(state.dependencyKeys, key)
		}

		if entry.Frontmatter.Input != "" {
			state.dependencyKeys = append(state.dependencyKeys, entry.Frontmatter.Input)
		}
	}

	for _, entry := range entries {
		if entry.Frontmatter.Output != "" {
			artifactName := "ARTIFACT/" + strings.TrimPrefix(entry.LogicalName, "ROOT/")
			artifactState := stateMap[artifactName]
			artifactState.dependencyKeys = append(artifactState.dependencyKeys, entry.LogicalName)
		}
	}

	for logicalName, state := range stateMap {
		for _, key := range state.dependencyKeys {
			if _, ok := stateMap[key]; !ok {
				return nil, nil, fmt.Errorf("%w: %s depends on unknown %s", ErrUnresolvableReference, logicalName, key)
			}
		}
	}

	n := len(stateMap)
	var cycleParticipants []string

	for pass := 0; pass < n; pass++ {
		changed := false
		var updatedInPass []string

		for logicalName, state := range stateMap {
			if logicalName == "ROOT" {
				continue
			}
			maxDepRank := -1
			for _, key := range state.dependencyKeys {
				depState := stateMap[key]
				if depState.rank > maxDepRank {
					maxDepRank = depState.rank
				}
			}
			candidateRank := maxDepRank + 1
			if candidateRank < 0 {
				candidateRank = 0
			}
			if candidateRank > state.rank {
				state.rank = candidateRank
				changed = true
				updatedInPass = append(updatedInPass, logicalName)
			}
		}

		if !changed {
			cycleParticipants = nil
			break
		}

		if pass == n-1 {
			cycleParticipants = updatedInPass
		}
	}

	var result []*NodeRankEntry
	for logicalName, state := range stateMap {
		result = append(result, &NodeRankEntry{
			LogicalName: logicalName,
			Rank:        state.rank,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Rank != result[j].Rank {
			return result[i].Rank < result[j].Rank
		}
		return result[i].LogicalName < result[j].LogicalName
	})

	sort.Strings(cycleParticipants)

	return result, cycleParticipants, nil
}
