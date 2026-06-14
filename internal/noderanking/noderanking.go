// code-from-spec: ROOT/golang/implementation/utils/node_ranking@sD_ro_vqo4S3xmqepL86DF0X2aQ
package noderanking

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
)

var ErrUnresolvableReference = errors.New("a depends_on or input target cannot be resolved")

type NodeRankInput struct {
	LogicalName string
	Frontmatter *frontmatter.Frontmatter
}

type NodeRankEntry struct {
	LogicalName string
	Rank        int
}

type rankingEntry struct {
	logicalName  string
	dependencies []string
	rank         int
}

func NodeRankCompute(entries []*NodeRankInput) (ranked []*NodeRankEntry, cycles []string, err error) {
	entryMap := make(map[string]*rankingEntry)

	for _, input := range entries {
		entryMap[input.LogicalName] = &rankingEntry{
			logicalName:  input.LogicalName,
			dependencies: []string{},
			rank:         0,
		}

		if input.Frontmatter != nil && input.Frontmatter.Output != "" {
			bare := strings.TrimPrefix(input.LogicalName, "SPEC/")
			artifactName := "ARTIFACT/" + bare
			entryMap[artifactName] = &rankingEntry{
				logicalName:  artifactName,
				dependencies: []string{},
				rank:         0,
			}
		}
	}

	frontmatterByLogical := make(map[string]*frontmatter.Frontmatter)
	for _, input := range entries {
		frontmatterByLogical[input.LogicalName] = input.Frontmatter
	}

	for logicalName, entry := range entryMap {
		if !logicalnames.LogicalNameIsSpec(logicalName) {
			continue
		}

		if logicalName != "SPEC" {
			parent, parentErr := logicalnames.LogicalNameGetParent(logicalName)
			if parentErr != nil {
				return nil, nil, fmt.Errorf("getting parent of %s: %w", logicalName, parentErr)
			}
			entry.dependencies = append(entry.dependencies, parent)
		}

		fm := frontmatterByLogical[logicalName]
		if fm == nil {
			continue
		}

		for _, dep := range fm.DependsOn {
			if strings.HasPrefix(dep, "ARTIFACT/") {
				entry.dependencies = append(entry.dependencies, dep)
			} else if logicalnames.LogicalNameIsSpec(dep) {
				bare := logicalnames.LogicalNameStripQualifier(dep)
				entry.dependencies = append(entry.dependencies, bare)
			} else if strings.HasPrefix(dep, "EXTERNAL/") {
				continue
			}
		}

		if fm.Input != "" {
			if strings.HasPrefix(fm.Input, "ARTIFACT/") {
				entry.dependencies = append(entry.dependencies, fm.Input)
			} else if strings.HasPrefix(fm.Input, "EXTERNAL/") {
				continue
			}
		}
	}

	for logicalName, entry := range entryMap {
		if !strings.HasPrefix(logicalName, "ARTIFACT/") {
			continue
		}
		bare := strings.TrimPrefix(logicalName, "ARTIFACT/")
		generatorName := "SPEC/" + bare
		entry.dependencies = append(entry.dependencies, generatorName)
	}

	for _, entry := range entryMap {
		for _, dep := range entry.dependencies {
			if _, exists := entryMap[dep]; !exists {
				return nil, nil, fmt.Errorf("%w: %s", ErrUnresolvableReference, dep)
			}
		}
	}

	n := len(entryMap)
	cycleParticipants := []string{}
	converged := false

	for pass := 0; pass < n; pass++ {
		changed := false

		for logicalName, entry := range entryMap {
			if logicalName == "SPEC" {
				continue
			}

			maxDepRank := 0
			for _, dep := range entry.dependencies {
				depEntry := entryMap[dep]
				if depEntry.rank > maxDepRank {
					maxDepRank = depEntry.rank
				}
			}

			newRank := 1 + maxDepRank
			if newRank > entry.rank {
				entry.rank = newRank
				changed = true
			}
		}

		if !changed {
			converged = true
			break
		}

		if pass == n-1 && !converged {
			for _, entry := range entryMap {
				maxDepRank := 0
				for _, dep := range entry.dependencies {
					depEntry := entryMap[dep]
					if depEntry.rank > maxDepRank {
						maxDepRank = depEntry.rank
					}
				}
				newRank := 1 + maxDepRank
				if newRank > entry.rank {
					cycleParticipants = append(cycleParticipants, entry.logicalName)
				}
			}
		}
	}

	ranked = make([]*NodeRankEntry, 0, len(entryMap))
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

	return ranked, cycleParticipants, nil
}
