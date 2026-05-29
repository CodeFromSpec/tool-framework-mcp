// code-from-spec: ROOT/golang/implementation/utils/node_ranking@mDCxxcEljTdvETDrHeW57HUhjKA

package noderanking

import (
	"fmt"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
)

// NodeRankInput represents a single discovered node and its parsed
// frontmatter, used as input to the ranking computation.
type NodeRankInput struct {
	LogicalName string
	Frontmatter *frontmatter.Frontmatter
}

// NodeRankEntry represents a single ranked node or artifact, identified
// by its logical name and assigned a numeric rank.
type NodeRankEntry struct {
	LogicalName string
	Rank        int
}

// ErrUnresolvableReference is returned when a depends_on or input
// target cannot be resolved to any known node in the input set.
var ErrUnresolvableReference = fmt.Errorf("unresolvable reference")

// internalEntry is used internally during ranking computation.
type internalEntry struct {
	logicalName  string
	dependencies []string
	rank         int
}

// NodeRankCompute takes the full set of discovered nodes with their
// parsed frontmatter and computes a topological ranking. It returns
// the ranked entries (nodes and artifacts) and a list of logical names
// involved in dependency cycles (empty if no cycles exist).
//
// Possible errors:
//   - ErrUnresolvableReference: a depends_on or input target cannot be
//     resolved to any known node in the input set.
func NodeRankCompute(entries []*NodeRankInput) (ranked []*NodeRankEntry, cycles []string, err error) {
	// Step 1 — Build entry map
	entryMap := make(map[string]*internalEntry)

	for _, input := range entries {
		entryMap[input.LogicalName] = &internalEntry{
			logicalName:  input.LogicalName,
			dependencies: []string{},
			rank:         0,
		}

		if input.Frontmatter != nil {
			for _, output := range input.Frontmatter.Outputs {
				// Construct artifact logical name: ARTIFACT/<suffix>(<id>)
				// Strip "ROOT/" prefix from logical_name
				suffix := strings.TrimPrefix(input.LogicalName, "ROOT/")
				artifactName := "ARTIFACT/" + suffix + "(" + output.ID + ")"

				entryMap[artifactName] = &internalEntry{
					logicalName:  artifactName,
					dependencies: []string{},
					rank:         0,
				}
			}
		}
	}

	// Step 2 — Build dependency edges

	// Step 2a: spec node entries (keys starting with "ROOT/")
	for _, input := range entries {
		entry := entryMap[input.LogicalName]

		if input.LogicalName == "ROOT" {
			// Skip ROOT — no dependencies
			continue
		}

		// Parent dependency
		parent, err := logicalnames.LogicalNameGetParent(input.LogicalName)
		if err != nil {
			return nil, nil, fmt.Errorf("getting parent of %q: %w", input.LogicalName, err)
		}
		entry.dependencies = append(entry.dependencies, parent)

		if input.Frontmatter == nil {
			continue
		}

		// depends_on dependencies
		for _, dep := range input.Frontmatter.DependsOn {
			if dep == nil {
				continue
			}
			depValue := *dep
			var lookupKey string
			if strings.HasPrefix(depValue, "ARTIFACT/") {
				lookupKey = depValue
			} else if strings.HasPrefix(depValue, "ROOT/") {
				lookupKey = logicalnames.LogicalNameStripQualifier(depValue)
			} else {
				lookupKey = depValue
			}

			if _, found := entryMap[lookupKey]; !found {
				return nil, nil, fmt.Errorf("%w: %q", ErrUnresolvableReference, lookupKey)
			}
			entry.dependencies = append(entry.dependencies, lookupKey)
		}

		// input dependency
		if input.Frontmatter.Input != "" {
			lookupKey := input.Frontmatter.Input
			if _, found := entryMap[lookupKey]; !found {
				return nil, nil, fmt.Errorf("%w: %q", ErrUnresolvableReference, lookupKey)
			}
			entry.dependencies = append(entry.dependencies, lookupKey)
		}
	}

	// Step 2b: artifact entries (keys starting with "ARTIFACT/")
	for key, entry := range entryMap {
		if !strings.HasPrefix(key, "ARTIFACT/") {
			continue
		}

		// Derive the generating node's logical name
		// Strip "ARTIFACT/" prefix and the trailing "(<id>)" qualifier
		withoutPrefix := strings.TrimPrefix(key, "ARTIFACT/")
		// Remove trailing qualifier
		qualifierStart := strings.LastIndex(withoutPrefix, "(")
		if qualifierStart >= 0 {
			withoutPrefix = withoutPrefix[:qualifierStart]
		}
		generatingNode := "ROOT/" + withoutPrefix

		if _, found := entryMap[generatingNode]; !found {
			return nil, nil, fmt.Errorf("%w: %q", ErrUnresolvableReference, generatingNode)
		}
		entry.dependencies = append(entry.dependencies, generatingNode)
	}

	// Step 3 — Initialize ranks
	// "ROOT" rank is fixed at 0; all others are already 0 from Step 1.

	// Step 4 — Iterate and detect cycles
	n := len(entryMap)
	var changedInLastPass []string

	for i := 0; i < n; i++ {
		changedThisPass := []string{}

		for key, entry := range entryMap {
			if key == "ROOT" {
				continue
			}

			if len(entry.dependencies) == 0 {
				newRank := 1
				if newRank > entry.rank {
					entry.rank = newRank
					changedThisPass = append(changedThisPass, key)
				}
				continue
			}

			maxDepRank := 0
			for _, depName := range entry.dependencies {
				dep, found := entryMap[depName]
				if !found {
					// Should not happen — already validated in Step 2
					continue
				}
				if dep.rank > maxDepRank {
					maxDepRank = dep.rank
				}
			}

			newRank := 1 + maxDepRank
			if newRank > entry.rank {
				entry.rank = newRank
				changedThisPass = append(changedThisPass, key)
			}
		}

		if len(changedThisPass) == 0 {
			// Converged — no cycles
			changedInLastPass = []string{}
			break
		}
		changedInLastPass = changedThisPass
	}

	// Step 5 — Output
	ranked = make([]*NodeRankEntry, 0, len(entryMap))
	for _, entry := range entryMap {
		ranked = append(ranked, &NodeRankEntry{
			LogicalName: entry.logicalName,
			Rank:        entry.rank,
		})
	}

	// Sort: primary by rank ascending, secondary by logical name ascending
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Rank != ranked[j].Rank {
			return ranked[i].Rank < ranked[j].Rank
		}
		return ranked[i].LogicalName < ranked[j].LogicalName
	})

	return ranked, changedInLastPass, nil
}
