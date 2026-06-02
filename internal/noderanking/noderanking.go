// code-from-spec: ROOT/golang/implementation/utils/node_ranking@NmAoLFbec81I0BpMD0y0qlAdP3A

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
	logicalName  string
	dependencies []string
	rank         int
}

func NodeRankCompute(entries []*NodeRankInput) ([]*NodeRankEntry, []string, error) {
	entryMap := make(map[string]*rankEntry)

	for _, item := range entries {
		entryMap[item.LogicalName] = &rankEntry{
			logicalName:  item.LogicalName,
			dependencies: []string{},
			rank:         0,
		}

		if item.Frontmatter.Output != "" {
			stripped := strings.TrimPrefix(item.LogicalName, "ROOT/")
			artifactKey := "ARTIFACT/" + stripped
			entryMap[artifactKey] = &rankEntry{
				logicalName:  artifactKey,
				dependencies: []string{},
				rank:         0,
			}
		}
	}

	for _, item := range entries {
		if item.LogicalName != "ROOT" {
			parentName, err := logicalnames.LogicalNameGetParent(item.LogicalName)
			if err != nil {
				return nil, nil, fmt.Errorf("%w: %s", ErrUnresolvableReference, item.LogicalName)
			}
			if _, ok := entryMap[parentName]; !ok {
				return nil, nil, fmt.Errorf("%w: parent %s of %s not found", ErrUnresolvableReference, parentName, item.LogicalName)
			}
			entryMap[item.LogicalName].dependencies = append(entryMap[item.LogicalName].dependencies, parentName)
		}

		for _, dep := range item.Frontmatter.DependsOn {
			var lookupKey string
			if strings.HasPrefix(dep, "ARTIFACT/") {
				lookupKey = dep
			} else {
				lookupKey = logicalnames.LogicalNameStripQualifier(dep)
			}
			if _, ok := entryMap[lookupKey]; !ok {
				return nil, nil, fmt.Errorf("%w: depends_on target %s not found", ErrUnresolvableReference, lookupKey)
			}
			entryMap[item.LogicalName].dependencies = append(entryMap[item.LogicalName].dependencies, lookupKey)
		}

		if item.Frontmatter.Input != "" {
			if _, ok := entryMap[item.Frontmatter.Input]; !ok {
				return nil, nil, fmt.Errorf("%w: input target %s not found", ErrUnresolvableReference, item.Frontmatter.Input)
			}
			entryMap[item.LogicalName].dependencies = append(entryMap[item.LogicalName].dependencies, item.Frontmatter.Input)
		}
	}

	for artifactKey := range entryMap {
		if !strings.HasPrefix(artifactKey, "ARTIFACT/") {
			continue
		}
		withoutPrefix := strings.TrimPrefix(artifactKey, "ARTIFACT/")
		stripped := logicalnames.LogicalNameStripQualifier(withoutPrefix)
		generatorName := "ROOT/" + stripped
		if _, ok := entryMap[generatorName]; !ok {
			return nil, nil, fmt.Errorf("%w: generator node %s for artifact %s not found", ErrUnresolvableReference, generatorName, artifactKey)
		}
		entryMap[artifactKey].dependencies = append(entryMap[artifactKey].dependencies, generatorName)
	}

	n := len(entryMap)
	var changedInLastPass []string

	for i := 1; i <= n; i++ {
		changedThisPass := []string{}

		for _, entry := range entryMap {
			if entry.logicalName == "ROOT" {
				continue
			}
			maxDepRank := 0
			for _, dep := range entry.dependencies {
				if depEntry, ok := entryMap[dep]; ok {
					if depEntry.rank > maxDepRank {
						maxDepRank = depEntry.rank
					}
				}
			}
			newRank := 1 + maxDepRank
			if newRank > entry.rank {
				entry.rank = newRank
				changedThisPass = append(changedThisPass, entry.logicalName)
			}
		}

		if len(changedThisPass) == 0 {
			changedInLastPass = nil
			break
		}
		changedInLastPass = changedThisPass
	}

	result := make([]*NodeRankEntry, 0, len(entryMap))
	for _, entry := range entryMap {
		result = append(result, &NodeRankEntry{
			LogicalName: entry.logicalName,
			Rank:        entry.rank,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Rank != result[j].Rank {
			return result[i].Rank < result[j].Rank
		}
		return result[i].LogicalName < result[j].LogicalName
	})

	cycleList := []string{}
	if changedInLastPass != nil {
		cycleList = changedInLastPass
	}

	return result, cycleList, nil
}
