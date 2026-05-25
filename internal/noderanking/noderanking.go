// code-from-spec: ROOT/golang/internal/node_ranking/code@PENDING
package noderanking

import (
	"errors"
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/nodediscovery"
)

// RankedEntry pairs a logical name with its computed rank.
type RankedEntry struct {
	LogicalName string
	Rank        int
}

// ErrUnresolvableRef is returned when a depends_on or input target cannot be resolved.
var ErrUnresolvableRef = errors.New("unresolvable reference")

// DetectCycles takes the full set of discovered nodes and returns the ranked
// entries and a slice of logical names involved in cycles (empty if no cycles).
func DetectCycles(nodes []nodediscovery.DiscoveredNode) ([]RankedEntry, []string, error) {
	// Build a set of known node logical names.
	nodeSet := make(map[string]bool, len(nodes))
	for _, n := range nodes {
		nodeSet[n.LogicalName] = true
	}

	// Parse frontmatter for each node.
	fmCache := make(map[string]*frontmatter.Frontmatter, len(nodes))
	for _, n := range nodes {
		fm, err := frontmatter.ParseFrontmatter(n.FilePath)
		if err != nil {
			// Use empty frontmatter on parse failure.
			fm = &frontmatter.Frontmatter{
				DependsOn: []string{},
				External:  []frontmatter.External{},
				Outputs:   []frontmatter.Output{},
			}
		}
		fmCache[n.LogicalName] = fm
	}

	// Step 1: Build entries and dependency map.
	entries := make([]string, 0, len(nodes)*2)
	deps := make(map[string][]string)

	// All known entries (nodes + artifacts).
	allEntries := make(map[string]bool)

	// First pass: add all node entries so we can validate deps.
	for _, n := range nodes {
		allEntries[n.LogicalName] = true
	}
	// Also pre-register artifact entries from outputs.
	for _, n := range nodes {
		fm := fmCache[n.LogicalName]
		for _, out := range fm.Outputs {
			allEntries[out.Path] = true
		}
	}

	// Second pass: build entries list and dep lists.
	for _, n := range nodes {
		entries = append(entries, n.LogicalName)
		fm := fmCache[n.LogicalName]

		var nodeDeps []string

		// Parent dependency.
		if parent, ok := logicalnames.ParentLogicalName(n.LogicalName); ok {
			nodeDeps = append(nodeDeps, parent)
		}

		// depends_on entries.
		for _, dep := range fm.DependsOn {
			if !allEntries[dep] {
				return nil, nil, fmt.Errorf("%w: %s depends on %s", ErrUnresolvableRef, n.LogicalName, dep)
			}
			nodeDeps = append(nodeDeps, dep)
		}

		// input artifact.
		if fm.Input != "" {
			if !allEntries[fm.Input] {
				return nil, nil, fmt.Errorf("%w: %s input %s", ErrUnresolvableRef, n.LogicalName, fm.Input)
			}
			nodeDeps = append(nodeDeps, fm.Input)
		}

		deps[n.LogicalName] = nodeDeps

		// Register artifact outputs.
		for _, out := range fm.Outputs {
			entries = append(entries, out.Path)
			deps[out.Path] = []string{n.LogicalName}
		}
	}

	// Step 2: Initialize ranks.
	ranks := make(map[string]int, len(entries))
	for _, e := range entries {
		ranks[e] = 0
	}

	// Step 3-4: Iterate until convergence.
	totalEntries := len(entries)
	converged := false
	for pass := 0; pass < totalEntries; pass++ {
		changed := false
		for _, e := range entries {
			depList := deps[e]
			if len(depList) == 0 {
				continue
			}
			maxDepRank := 0
			for _, d := range depList {
				if ranks[d] > maxDepRank {
					maxDepRank = ranks[d]
				}
			}
			newRank := 1 + maxDepRank
			if newRank > ranks[e] {
				ranks[e] = newRank
				changed = true
			}
		}
		if !changed {
			converged = true
			break
		}
	}

	// Step 5: Cycle detection.
	var cycleParticipants []string
	if !converged {
		// One more pass to identify cycle participants.
		for _, e := range entries {
			depList := deps[e]
			if len(depList) == 0 {
				continue
			}
			maxDepRank := 0
			for _, d := range depList {
				if ranks[d] > maxDepRank {
					maxDepRank = ranks[d]
				}
			}
			newRank := 1 + maxDepRank
			if newRank > ranks[e] {
				ranks[e] = newRank
				cycleParticipants = append(cycleParticipants, e)
			}
		}
	}

	// Build result.
	result := make([]RankedEntry, len(entries))
	for i, e := range entries {
		result[i] = RankedEntry{
			LogicalName: e,
			Rank:        ranks[e],
		}
	}

	if cycleParticipants == nil {
		cycleParticipants = []string{}
	}

	return result, cycleParticipants, nil
}
