// code-from-spec: ROOT/golang/implementation/utils/node_ranking@pDW5r3J6IqfCrR8ybNfWD7OulNw

package noderanking

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
)

// ErrUnresolvableReference is returned when a depends_on or input target
// cannot be resolved to a known node.
var ErrUnresolvableReference = errors.New("unresolvable reference")

// NodeRankInput holds a discovered node's logical name and its parsed frontmatter,
// used as input to NodeRankCompute.
type NodeRankInput struct {
	// LogicalName is the ROOT/ logical name of the node.
	LogicalName string

	// Frontmatter is the parsed frontmatter of the node's spec file.
	Frontmatter *frontmatter.Frontmatter
}

// NodeRankEntry holds a node or artifact's logical name and its computed rank.
type NodeRankEntry struct {
	// LogicalName is the ROOT/ or ARTIFACT/ logical name.
	LogicalName string

	// Rank is the computed topological rank (lower rank = fewer dependencies).
	Rank int
}

// rankMapEntry is the internal tracking structure used during rank computation.
type rankMapEntry struct {
	logicalName  string
	rank         int
	dependencies []string
}

// NodeRankCompute takes the full set of discovered nodes with their parsed
// frontmatter and computes a topological ranking for all nodes and artifacts.
//
// It returns:
//   - ranked: a list of NodeRankEntry values, one per node and artifact,
//     ordered by ascending rank.
//   - cycles: a list of logical names that are part of dependency cycles.
//     Empty if no cycles are detected.
//
// Errors:
//   - ErrUnresolvableReference: a depends_on or input target cannot be resolved.
func NodeRankCompute(entries []*NodeRankInput) (ranked []*NodeRankEntry, cycles []string, err error) {
	// Step 1 — Build entry map.
	entryMap := make(map[string]*rankMapEntry)

	for _, item := range entries {
		// Add spec node entry.
		entryMap[item.LogicalName] = &rankMapEntry{
			logicalName:  item.LogicalName,
			rank:         0,
			dependencies: []string{},
		}

		// Add artifact entries for each output.
		for _, output := range item.Frontmatter.Outputs {
			// Derive artifact logical name: strip "ROOT/" prefix, prepend "ARTIFACT/",
			// append "(<id>)".
			bare := strings.TrimPrefix(item.LogicalName, "ROOT/")
			artifactName := "ARTIFACT/" + bare + "(" + output.ID + ")"
			entryMap[artifactName] = &rankMapEntry{
				logicalName:  artifactName,
				rank:         0,
				dependencies: []string{},
			}
		}
	}

	// Step 2 — Build dependency edges.

	// Build a lookup map from LogicalName to Frontmatter for spec nodes.
	fmMap := make(map[string]*frontmatter.Frontmatter, len(entries))
	for _, item := range entries {
		fmMap[item.LogicalName] = item.Frontmatter
	}

	// Process spec node entries.
	for _, item := range entries {
		mapEntry := entryMap[item.LogicalName]

		// Skip ROOT — it has no parent and no automatic dependencies.
		if item.LogicalName == "ROOT" {
			continue
		}

		// Add parent as a dependency.
		parent, err := logicalnames.LogicalNameGetParent(item.LogicalName)
		if err != nil {
			return nil, nil, fmt.Errorf("LogicalNameGetParent(%q): %w", item.LogicalName, err)
		}
		mapEntry.dependencies = append(mapEntry.dependencies, parent)

		// Add depends_on references.
		for _, ref := range item.Frontmatter.DependsOn {
			var depKey string
			if strings.HasPrefix(ref, "ARTIFACT/") {
				depKey = ref
			} else if strings.HasPrefix(ref, "ROOT/") {
				depKey = logicalnames.LogicalNameStripQualifier(ref)
			} else {
				depKey = ref
			}
			mapEntry.dependencies = append(mapEntry.dependencies, depKey)
		}

		// Add input reference if non-empty.
		if item.Frontmatter.Input != "" {
			mapEntry.dependencies = append(mapEntry.dependencies, item.Frontmatter.Input)
		}
	}

	// Process artifact entries.
	for logicalName, mapEntry := range entryMap {
		if !logicalnames.LogicalNameIsArtifact(logicalName) {
			continue
		}
		generator, err := logicalnames.LogicalNameGetArtifactGenerator(logicalName)
		if err != nil {
			return nil, nil, fmt.Errorf("LogicalNameGetArtifactGenerator(%q): %w", logicalName, err)
		}
		mapEntry.dependencies = append(mapEntry.dependencies, generator)
	}

	// Step 5 — Validate that all dependencies exist in the map.
	for _, mapEntry := range entryMap {
		for _, dep := range mapEntry.dependencies {
			if _, ok := entryMap[dep]; !ok {
				return nil, nil, fmt.Errorf("%w: %q referenced by %q not found",
					ErrUnresolvableReference, dep, mapEntry.logicalName)
			}
		}
	}

	// Step 3 — Initialize ranks.
	// ROOT is already 0; all others are already 0 from initialization.
	// (Nothing extra needed — map entries are initialized with rank 0.)

	// Step 4 — Iterate and detect cycles.
	n := len(entryMap)
	var lastChangedNames []string
	converged := false

	for pass := 1; pass <= n; pass++ {
		changed := false
		changedNames := []string{}

		for logicalName, mapEntry := range entryMap {
			if logicalName == "ROOT" {
				continue
			}

			// Compute candidate rank = 1 + max(rank of dependencies).
			maxDepRank := -1
			for _, dep := range mapEntry.dependencies {
				depEntry := entryMap[dep]
				if depEntry.rank > maxDepRank {
					maxDepRank = depEntry.rank
				}
			}

			candidateRank := 0
			if maxDepRank >= 0 {
				candidateRank = 1 + maxDepRank
			}

			if candidateRank > mapEntry.rank {
				mapEntry.rank = candidateRank
				changed = true
				changedNames = append(changedNames, logicalName)
			}
		}

		lastChangedNames = changedNames

		if !changed {
			converged = true
			break
		}
	}

	// Step 10 — Collect cycle participants if not converged.
	if !converged {
		cycles = lastChangedNames
	} else {
		cycles = []string{}
	}

	// Step 5 — Output.
	ranked = make([]*NodeRankEntry, 0, len(entryMap))
	for _, mapEntry := range entryMap {
		ranked = append(ranked, &NodeRankEntry{
			LogicalName: mapEntry.logicalName,
			Rank:        mapEntry.rank,
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
