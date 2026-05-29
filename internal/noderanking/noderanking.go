// code-from-spec: ROOT/golang/implementation/utils/node_ranking@3uaGDxgoOElAMkt0fub5og5Y15Q
package noderanking

import (
	"errors"
	"fmt"
	"sort"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
)

// NodeRankInput represents a discovered node with its logical name and
// parsed frontmatter, used as input to the ranking computation.
type NodeRankInput struct {
	LogicalName string
	Frontmatter *frontmatter.Frontmatter
}

// NodeRankEntry represents a node or artifact with its computed
// topological rank.
type NodeRankEntry struct {
	LogicalName string
	Rank        int
}

// ErrUnresolvableReference is returned when a depends_on or input
// target cannot be resolved to a known node.
var ErrUnresolvableReference = errors.New("unresolvable reference")

// internalEntry is used internally during rank computation.
type internalEntry struct {
	logicalName  string
	dependencies []string
	rank         int
}

// NodeRankCompute takes the full set of discovered nodes with their
// parsed frontmatter and computes a topological rank for each node
// and artifact. It returns the ranked entries and a list of logical
// names involved in dependency cycles (empty if no cycles exist).
//
// Returns ErrUnresolvableReference if a depends_on or input target
// cannot be resolved to any entry in the provided set.
func NodeRankCompute(entries []*NodeRankInput) (ranked []*NodeRankEntry, cycles []string, err error) {
	// Step 1: Build entry map
	entryMap := make(map[string]*internalEntry)

	for _, input := range entries {
		// Add the spec node entry
		entryMap[input.LogicalName] = &internalEntry{
			logicalName:  input.LogicalName,
			dependencies: []string{},
			rank:         0,
		}

		// Add artifact entries for each output
		if input.Frontmatter != nil {
			for _, output := range input.Frontmatter.Outputs {
				// Derive artifact logical name: strip "ROOT/" prefix, prepend "ARTIFACT/", append "(<id>)"
				suffix := input.LogicalName[len("ROOT/"):]
				artifactName := fmt.Sprintf("ARTIFACT/%s(%s)", suffix, output.ID)

				entryMap[artifactName] = &internalEntry{
					logicalName:  artifactName,
					dependencies: []string{},
					rank:         0,
				}
			}
		}
	}

	// Step 2: Build dependency edges

	// For each spec node entry
	for _, input := range entries {
		entry := entryMap[input.LogicalName]

		// a. Parent dependency
		if input.LogicalName != "ROOT" {
			parent, parentErr := logicalnames.LogicalNameGetParent(input.LogicalName)
			if parentErr != nil {
				return nil, nil, fmt.Errorf("%w: parent of %s: %v", ErrUnresolvableReference, input.LogicalName, parentErr)
			}
			if _, found := entryMap[parent]; !found {
				return nil, nil, fmt.Errorf("%w: parent %q of node %q not found", ErrUnresolvableReference, parent, input.LogicalName)
			}
			entry.dependencies = append(entry.dependencies, parent)
		}

		if input.Frontmatter == nil {
			continue
		}

		// b. depends_on dependencies
		for _, ref := range input.Frontmatter.DependsOn {
			var lookupKey string
			if len(ref) >= 9 && ref[:9] == "ARTIFACT/" {
				lookupKey = ref
			} else {
				lookupKey = logicalnames.LogicalNameStripQualifier(ref)
			}
			if _, found := entryMap[lookupKey]; !found {
				return nil, nil, fmt.Errorf("%w: depends_on reference %q from node %q not found", ErrUnresolvableReference, lookupKey, input.LogicalName)
			}
			entry.dependencies = append(entry.dependencies, lookupKey)
		}

		// c. input dependency
		if input.Frontmatter.Input != "" {
			inputRef := input.Frontmatter.Input
			if _, found := entryMap[inputRef]; !found {
				return nil, nil, fmt.Errorf("%w: input reference %q from node %q not found", ErrUnresolvableReference, inputRef, input.LogicalName)
			}
			entry.dependencies = append(entry.dependencies, inputRef)
		}
	}

	// For each artifact entry, add the generating node as a dependency
	for key, entry := range entryMap {
		if len(key) < 9 || key[:9] != "ARTIFACT/" {
			continue
		}
		generator, genErr := logicalnames.LogicalNameGetArtifactGenerator(key)
		if genErr != nil {
			return nil, nil, fmt.Errorf("%w: cannot get generator for artifact %q: %v", ErrUnresolvableReference, key, genErr)
		}
		if _, found := entryMap[generator]; !found {
			return nil, nil, fmt.Errorf("%w: generator node %q for artifact %q not found", ErrUnresolvableReference, generator, key)
		}
		entry.dependencies = append(entry.dependencies, generator)
	}

	// Step 3: Initialize ranks — all entries start at 0 (already set above)

	// Step 4: Iterate and detect cycles
	n := len(entryMap)
	cycleCandidates := []string{}

	for pass := 1; pass <= n; pass++ {
		changed := false
		currentPassCandidates := []string{}

		for key, entry := range entryMap {
			if key == "ROOT" {
				continue
			}

			// Compute candidate_rank = 1 + max(rank of dependencies)
			maxDepRank := -1
			for _, dep := range entry.dependencies {
				depEntry, found := entryMap[dep]
				if !found {
					// Should not happen since we validated above, but guard anyway
					return nil, nil, fmt.Errorf("%w: dependency %q not found during rank computation", ErrUnresolvableReference, dep)
				}
				if depEntry.rank > maxDepRank {
					maxDepRank = depEntry.rank
				}
			}

			candidateRank := 0
			if maxDepRank >= 0 {
				candidateRank = 1 + maxDepRank
			}

			if candidateRank > entry.rank {
				entry.rank = candidateRank
				changed = true
				currentPassCandidates = append(currentPassCandidates, entry.logicalName)
			}
		}

		if !changed {
			// Converged — no cycles
			cycleCandidates = []string{}
			break
		}

		if pass == n {
			// Did not converge — cycles detected
			cycleCandidates = currentPassCandidates
		}
	}

	// Step 5: Collect, sort, and return
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

	return result, cycleCandidates, nil
}
