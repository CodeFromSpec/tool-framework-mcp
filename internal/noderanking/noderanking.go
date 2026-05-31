// code-from-spec: ROOT/golang/implementation/utils/node_ranking@9BM2xJl1j-0FUiL5V4wuoSx5wK8

package noderanking

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
)

// ErrUnresolvableReference is returned when a depends_on or input
// target in the frontmatter cannot be resolved to a known node.
var ErrUnresolvableReference = errors.New("unresolvable reference")

// NodeRankInput represents a discovered node with its parsed
// frontmatter, used as input to the ranking computation.
type NodeRankInput struct {
	// LogicalName is the ROOT/ logical name of the node.
	LogicalName string

	// Frontmatter holds the parsed frontmatter for the node,
	// including its dependency declarations.
	Frontmatter *frontmatter.Frontmatter
}

// NodeRankEntry represents a node or artifact in the ranked output.
type NodeRankEntry struct {
	// LogicalName is the ROOT/ or ARTIFACT/ logical name.
	LogicalName string

	// Rank is the computed topological rank of this entry.
	// Lower values come earlier in the dependency order.
	Rank int
}

// entryState holds the internal state for each node/artifact during ranking.
type entryState struct {
	logicalName  string
	dependencies []string
	rank         int
}

// NodeRankCompute takes the full set of discovered nodes with their
// parsed frontmatter and returns a topologically ranked list of entries
// (nodes and artifacts) along with the logical names of any nodes
// involved in dependency cycles.
//
// The returned ranked slice contains one entry per node/artifact in
// dependency order (lowest rank first). The cycles slice contains the
// logical names of all nodes that participate in a cycle; it is empty
// when no cycles are detected.
//
// Errors:
//   - ErrUnresolvableReference: a depends_on or input target in the
//     frontmatter cannot be resolved to a known node.
func NodeRankCompute(entries []*NodeRankInput) (ranked []*NodeRankEntry, cycles []string, err error) {
	// Step 1 — Build entry map.
	entryMap := make(map[string]*entryState)

	// Track which logical names belong to spec nodes (vs artifacts).
	specNodes := make(map[string]*NodeRankInput)

	for _, input := range entries {
		// Add the spec node entry.
		entryMap[input.LogicalName] = &entryState{
			logicalName:  input.LogicalName,
			dependencies: []string{},
			rank:         0,
		}
		specNodes[input.LogicalName] = input

		// Add artifact entries for each output.
		for _, output := range input.Frontmatter.Outputs {
			artifactLogicalName := buildArtifactLogicalName(input.LogicalName, output.ID)
			entryMap[artifactLogicalName] = &entryState{
				logicalName:  artifactLogicalName,
				dependencies: []string{},
				rank:         0,
			}
		}
	}

	// Step 2 — Build dependency edges.

	// Step 2, part 3: for each spec node entry.
	for logicalName, input := range specNodes {
		entry := entryMap[logicalName]

		// Skip ROOT (no dependencies).
		if logicalName == "ROOT" {
			continue
		}

		// Derive and add parent dependency.
		parent, err := logicalnames.LogicalNameGetParent(logicalName)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %s (getting parent of %s)", ErrUnresolvableReference, logicalName, logicalName)
		}
		if _, ok := entryMap[parent]; !ok {
			return nil, nil, fmt.Errorf("%w: %s", ErrUnresolvableReference, parent)
		}
		entry.dependencies = append(entry.dependencies, parent)

		// Add depends_on edges.
		for _, dep := range input.Frontmatter.DependsOn {
			var lookupKey string
			if strings.HasPrefix(dep, "ARTIFACT/") {
				lookupKey = dep
			} else {
				lookupKey = logicalnames.LogicalNameStripQualifier(dep)
			}
			if _, ok := entryMap[lookupKey]; !ok {
				return nil, nil, fmt.Errorf("%w: %s", ErrUnresolvableReference, lookupKey)
			}
			entry.dependencies = append(entry.dependencies, lookupKey)
		}

		// Add input dependency.
		if input.Frontmatter.Input != "" {
			lookupKey := input.Frontmatter.Input
			if _, ok := entryMap[lookupKey]; !ok {
				return nil, nil, fmt.Errorf("%w: %s", ErrUnresolvableReference, lookupKey)
			}
			entry.dependencies = append(entry.dependencies, lookupKey)
		}
	}

	// Step 2, part 4: for each artifact entry.
	for logicalName, entry := range entryMap {
		if !strings.HasPrefix(logicalName, "ARTIFACT/") {
			continue
		}
		generatorLogicalName, err := logicalnames.LogicalNameGetArtifactGenerator(logicalName)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %s (getting generator for %s)", ErrUnresolvableReference, logicalName, logicalName)
		}
		if _, ok := entryMap[generatorLogicalName]; !ok {
			return nil, nil, fmt.Errorf("%w: %s", ErrUnresolvableReference, generatorLogicalName)
		}
		entry.dependencies = append(entry.dependencies, generatorLogicalName)
	}

	// Step 3 — Initialize ranks.
	// ROOT rank is 0 (fixed). All others start at 0 (already set during construction).

	// Step 4 — Iterate and detect cycles.
	n := len(entryMap)
	var changedInLastPass []string

	for pass := 0; pass < n; pass++ {
		var changedThisPass []string

		for logicalName, entry := range entryMap {
			if logicalName == "ROOT" {
				continue
			}

			// Compute new rank as 1 + max(rank of dependencies).
			newRank := 1
			for _, dep := range entry.dependencies {
				depEntry := entryMap[dep]
				candidate := depEntry.rank + 1
				if candidate > newRank {
					newRank = candidate
				}
			}

			if newRank > entry.rank {
				entry.rank = newRank
				changedThisPass = append(changedThisPass, logicalName)
			}
		}

		changedInLastPass = changedThisPass

		if len(changedThisPass) == 0 {
			// Converged — no cycles.
			break
		}
	}

	// Determine cycle participants.
	if len(changedInLastPass) > 0 {
		cycles = changedInLastPass
	} else {
		cycles = []string{}
	}

	// Step 5 — Output.
	ranked = make([]*NodeRankEntry, 0, len(entryMap))
	for _, entry := range entryMap {
		ranked = append(ranked, &NodeRankEntry{
			LogicalName: entry.logicalName,
			Rank:        entry.rank,
		})
	}

	// Sort: primary by rank ascending, secondary by logical name ascending.
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Rank != ranked[j].Rank {
			return ranked[i].Rank < ranked[j].Rank
		}
		return ranked[i].LogicalName < ranked[j].LogicalName
	})

	return ranked, cycles, nil
}

// buildArtifactLogicalName derives the ARTIFACT/ logical name for a
// given node logical name and output id.
//
// Examples:
//   - "ROOT/a/b" + "foo" → "ARTIFACT/a/b(foo)"
//   - "ROOT"     + "bar" → "ARTIFACT/(bar)"
func buildArtifactLogicalName(nodeLogicalName string, outputID string) string {
	var pathSegment string
	if nodeLogicalName == "ROOT" {
		pathSegment = ""
	} else {
		// Strip the "ROOT/" prefix.
		pathSegment = strings.TrimPrefix(nodeLogicalName, "ROOT/")
	}
	return "ARTIFACT/" + pathSegment + "(" + outputID + ")"
}
