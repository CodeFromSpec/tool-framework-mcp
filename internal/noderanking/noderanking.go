// code-from-spec: ROOT/golang/internal/node_ranking/code@EJSjeOloABdRJoJXkrONS5KgztY

// Package noderanking implements a topological ranking algorithm for spec nodes
// and their artifacts. It detects dependency cycles as a by-product of the
// ranking iteration — no separate graph traversal is performed.
//
// The algorithm works as follows:
//  1. Build an entry_map from all spec nodes and their artifacts.
//  2. Iteratively propagate ranks: each entry's rank = max(dependency ranks) + 1.
//  3. Convergence is reached when a full pass makes no changes.
//  4. If N passes complete without convergence, the extra changed entries in one
//     more pass are the cycle participants.
package noderanking

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/nodediscovery"
)

// ErrUnresolvableRef is returned when a depends_on or input target cannot be
// resolved to a known entry in the entry map.
var ErrUnresolvableRef = errors.New("unresolvable reference")

// RankedEntry holds the final ranking for one spec node or one artifact.
type RankedEntry struct {
	LogicalName string
	Rank        int
}

// rankEntry is the internal working record used while building and iterating.
type rankEntry struct {
	logicalName  string
	rank         int
	dependencies []string // logical names this entry depends on
}

// DetectCycles takes the full set of discovered nodes with their parsed
// frontmatter, assigns ranks via iterative propagation, and detects cycles.
//
// Returns:
//   - rankedEntries: one RankedEntry per spec node and one per artifact output.
//   - cycleParticipants: logical names involved in cycles (empty when none).
//   - error: wraps ErrUnresolvableRef when a depends_on or input ref is unknown.
func DetectCycles(nodes []nodediscovery.DiscoveredNode) ([]RankedEntry, []string, error) {
	// -----------------------------------------------------------------------
	// Step 1 — Discovery: collect all entries and build dependency lists.
	// -----------------------------------------------------------------------

	// entry_map holds one working record per logical name (spec node + artifacts).
	entryMap := make(map[string]*rankEntry)

	// Parse frontmatter for each node.
	fmCache := make(map[string]*frontmatter.Frontmatter, len(nodes))
	for _, node := range nodes {
		fm, err := frontmatter.ParseFrontmatter(node.FilePath)
		if err != nil {
			fm = &frontmatter.Frontmatter{}
		}
		fmCache[node.LogicalName] = fm
	}

	// Pass A: add every spec node and each of its artifact outputs to entryMap.
	for _, node := range nodes {
		lname := node.LogicalName
		fm := fmCache[lname]

		entryMap[lname] = &rankEntry{
			logicalName:  lname,
			rank:         0,
			dependencies: nil,
		}

		for _, out := range fm.Outputs {
			nodePathWithoutRoot := strings.TrimPrefix(lname, "ROOT/")
			artifactKey := "ARTIFACT/" + nodePathWithoutRoot + "(" + out.ID + ")"

			entryMap[artifactKey] = &rankEntry{
				logicalName:  artifactKey,
				rank:         0,
				dependencies: []string{lname},
			}
		}
	}

	// Pass B: populate each spec-node entry's dependency list.
	for _, node := range nodes {
		lname := node.LogicalName
		fm := fmCache[lname]
		entry := entryMap[lname]

		var deps []string

		if lname != "ROOT" {
			parent, ok := logicalnames.ParentLogicalName(lname)
			if ok {
				deps = append(deps, parent)
			}
		}

		for _, ref := range fm.DependsOn {
			if _, found := entryMap[ref]; !found {
				return nil, nil, fmt.Errorf("depends_on %q in node %q: %w", ref, lname, ErrUnresolvableRef)
			}
			deps = append(deps, ref)
		}

		if fm.Input != "" {
			ref := fm.Input
			if _, found := entryMap[ref]; !found {
				return nil, nil, fmt.Errorf("input %q in node %q: %w", ref, lname, ErrUnresolvableRef)
			}
			deps = append(deps, ref)
		}

		entry.dependencies = deps
	}

	// -----------------------------------------------------------------------
	// Step 2 — Initialization: all ranks are already 0 from construction.
	// -----------------------------------------------------------------------

	// -----------------------------------------------------------------------
	// Steps 3–4 — Iterative propagation until convergence or N passes.
	// -----------------------------------------------------------------------

	// N is the total number of entries (spec nodes + artifacts).
	n := len(entryMap)

	passCount := 0
	for {
		changed := runOnePass(entryMap)
		passCount++

		if !changed {
			// Convergence: no rank changed this pass — graph is acyclic.
			break
		}
		if passCount == n {
			// N passes completed without convergence — cycles exist.
			break
		}
	}

	// -----------------------------------------------------------------------
	// Step 5 — Cycle detection.
	// -----------------------------------------------------------------------

	var cycleParticipants []string

	if passCount == n {
		// Run one additional pass; entries whose rank still changes are in a cycle.
		cycleParticipants = collectCycleParticipants(entryMap)
	}
	// If passCount < n, cycleParticipants stays nil (treated as empty).

	// -----------------------------------------------------------------------
	// Step 6 — Build result.
	// -----------------------------------------------------------------------

	rankedEntries := make([]RankedEntry, 0, len(entryMap))
	for _, e := range entryMap {
		rankedEntries = append(rankedEntries, RankedEntry{
			LogicalName: e.logicalName,
			Rank:        e.rank,
		})
	}

	return rankedEntries, cycleParticipants, nil
}

// runOnePass executes a single propagation pass over all entries in entryMap.
// For each entry with dependencies, it sets the entry's rank to
// max(dependency ranks) + 1 if that is larger than the current rank.
// Returns true if any rank changed during this pass.
func runOnePass(entryMap map[string]*rankEntry) bool {
	changed := false

	for _, entry := range entryMap {
		// Entries with no dependencies (e.g., ROOT) always stay at rank 0.
		if len(entry.dependencies) == 0 {
			continue
		}

		// Find the maximum rank among all dependencies.
		maxDepRank := 0
		for _, depName := range entry.dependencies {
			dep := entryMap[depName]
			if dep == nil {
				continue
			}
			if dep.rank > maxDepRank {
				maxDepRank = dep.rank
			}
		}

		// New rank must exceed the maximum dependency rank.
		newRank := maxDepRank + 1

		if newRank > entry.rank {
			entry.rank = newRank
			changed = true
		}
	}

	return changed
}

// collectCycleParticipants runs one more pass and collects the logical names of
// every entry whose rank changes. Those entries are participants in a cycle
// because a DAG would have already converged.
func collectCycleParticipants(entryMap map[string]*rankEntry) []string {
	var participants []string

	for _, entry := range entryMap {
		if len(entry.dependencies) == 0 {
			continue
		}

		maxDepRank := 0
		for _, depName := range entry.dependencies {
			dep := entryMap[depName]
			if dep == nil {
				continue
			}
			if dep.rank > maxDepRank {
				maxDepRank = dep.rank
			}
		}

		newRank := maxDepRank + 1
		if newRank > entry.rank {
			entry.rank = newRank
			participants = append(participants, entry.logicalName)
		}
	}

	return participants
}
