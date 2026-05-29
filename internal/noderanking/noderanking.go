// code-from-spec: ROOT/golang/implementation/utils/node_ranking@kx762p-lmfd7j3n0TGb_Lk7pOEY

package noderanking

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
)

// ErrUnresolvableReference is returned when a depends_on or input
// target cannot be resolved to any known node in the input set.
var ErrUnresolvableReference = errors.New("unresolvable reference")

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

// internalEntry holds the working state for a node or artifact during ranking.
type internalEntry struct {
	logicalName  string
	rank         int
	dependencies []string
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
		// Add spec node entry
		entryMap[input.LogicalName] = &internalEntry{
			logicalName:  input.LogicalName,
			rank:         0,
			dependencies: []string{},
		}

		// Add artifact entries for each output
		if input.Frontmatter != nil {
			for _, output := range input.Frontmatter.Outputs {
				// Construct artifact logical name:
				// Strip "ROOT/" prefix, prepend "ARTIFACT/", append "(id)"
				stripped := strings.TrimPrefix(input.LogicalName, "ROOT/")
				artifactName := "ARTIFACT/" + stripped + "(" + output.ID + ")"
				entryMap[artifactName] = &internalEntry{
					logicalName:  artifactName,
					rank:         0,
					dependencies: []string{},
				}
			}
		}
	}

	// Step 2 — Build dependency edges

	// Process spec node entries
	for _, input := range entries {
		entry := entryMap[input.LogicalName]

		// ROOT has no parent dependencies
		if input.LogicalName == "ROOT" {
			entry.dependencies = []string{}
			continue
		}

		// Add parent as dependency
		parent, err := logicalnames.LogicalNameGetParent(input.LogicalName)
		if err != nil {
			return nil, nil, fmt.Errorf("getting parent of %q: %w", input.LogicalName, err)
		}
		entry.dependencies = append(entry.dependencies, parent)

		if input.Frontmatter == nil {
			continue
		}

		// Process depends_on entries
		for _, dep := range input.Frontmatter.DependsOn {
			if dep == nil {
				continue
			}
			depName := *dep
			var lookupKey string

			if strings.HasPrefix(depName, "ARTIFACT/") {
				lookupKey = depName
			} else if strings.HasPrefix(depName, "ROOT/") {
				// Strip any parenthetical qualifier
				if idx := strings.Index(depName, "("); idx >= 0 {
					lookupKey = depName[:idx]
				} else {
					lookupKey = depName
				}
			} else {
				lookupKey = depName
			}

			if _, found := entryMap[lookupKey]; !found {
				return nil, nil, fmt.Errorf("%w: %q referenced by %q", ErrUnresolvableReference, lookupKey, input.LogicalName)
			}
			entry.dependencies = append(entry.dependencies, lookupKey)
		}

		// Process input dependency
		if input.Frontmatter.Input != "" {
			lookupKey := input.Frontmatter.Input
			if _, found := entryMap[lookupKey]; !found {
				return nil, nil, fmt.Errorf("%w: input %q referenced by %q", ErrUnresolvableReference, lookupKey, input.LogicalName)
			}
			entry.dependencies = append(entry.dependencies, lookupKey)
		}
	}

	// Process artifact entries
	for key, entry := range entryMap {
		if !strings.HasPrefix(key, "ARTIFACT/") {
			continue
		}

		generatorName, err := logicalnames.LogicalNameGetArtifactGenerator(key)
		if err != nil {
			return nil, nil, fmt.Errorf("getting artifact generator for %q: %w", key, err)
		}

		if _, found := entryMap[generatorName]; !found {
			return nil, nil, fmt.Errorf("%w: artifact generator %q not found for %q", ErrUnresolvableReference, generatorName, key)
		}

		entry.dependencies = []string{generatorName}
	}

	// Step 3 — Initialize ranks
	// ROOT stays at 0; all others already initialized to 0.

	// Step 4 — Iterate and detect cycles
	n := len(entryMap)
	changedInLastPass := []string{}

	for pass := 0; pass < n; pass++ {
		changedThisPass := []string{}

		for key, entry := range entryMap {
			if key == "ROOT" {
				continue
			}

			// Compute candidate_rank = 1 + max(rank of all dependencies)
			candidateRank := 1
			for _, depName := range entry.dependencies {
				dep, found := entryMap[depName]
				if !found {
					// Should have been caught in step 2, but guard anyway
					return nil, nil, fmt.Errorf("%w: %q", ErrUnresolvableReference, depName)
				}
				if 1+dep.rank > candidateRank {
					candidateRank = 1 + dep.rank
				}
			}

			if candidateRank > entry.rank {
				entry.rank = candidateRank
				changedThisPass = append(changedThisPass, entry.logicalName)
			}
		}

		if len(changedThisPass) == 0 {
			// Convergence reached
			changedInLastPass = []string{}
			break
		}

		changedInLastPass = changedThisPass
	}

	// Step 5 — Output

	// Collect all entries
	result := make([]*NodeRankEntry, 0, len(entryMap))
	for _, entry := range entryMap {
		result = append(result, &NodeRankEntry{
			LogicalName: entry.logicalName,
			Rank:        entry.rank,
		})
	}

	// Sort: primary by rank ascending, secondary by logical name ascending
	sort.Slice(result, func(i, j int) bool {
		if result[i].Rank != result[j].Rank {
			return result[i].Rank < result[j].Rank
		}
		return result[i].LogicalName < result[j].LogicalName
	})

	return result, changedInLastPass, nil
}
