// code-from-spec: ROOT/golang/internal/node_ranking/code@ONLVRRWAr3i8d32vU6fn89LDQE4

// Package noderanking computes an integer rank for every spec node and
// artifact in the discovered set. Rank determines processing order: lower-
// ranked entries must be processed before higher-ranked ones. The package also
// identifies entries that participate in dependency cycles.
//
// The ranking algorithm is iterative (Bellman-Ford style): all ranks start at
// 0 and are updated in repeated passes until convergence. If ranks are still
// changing after N passes (where N = number of entries), a cycle is present
// and affected entries are collected.
package noderanking

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/nodediscovery"
)

// RankedEntry pairs a logical name with its computed rank.
// Entries with lower ranks must be processed before entries with higher ranks.
// Entries with the same rank have no dependency relationship and may be
// processed in any order (including in parallel).
type RankedEntry struct {
	LogicalName string
	Rank        int
}

// ErrUnresolvableRef is returned by DetectCycles when a depends_on or input
// reference cannot be resolved to any known entry in the entry map.
var ErrUnresolvableRef = errors.New("unresolvable reference")

// ─── internal types ──────────────────────────────────────────────────────────

// entryKind distinguishes spec nodes from generated artifacts.
type entryKind int

const (
	kindNode     entryKind = iota // a spec node (ROOT/…)
	kindArtifact                  // a generated artifact (ARTIFACT/…)
)

// entry is the internal working record used during rank computation.
type entry struct {
	logicalName    string    // key in the entry map (bare, no qualifier for nodes)
	kind           entryKind
	generatingNode string   // for kindArtifact: the node logical name that produces it
	dependencies   []string // logical names this entry directly depends on
	rank           int
}

// ─── public API ──────────────────────────────────────────────────────────────

// DetectCycles takes the full set of discovered nodes, parses their frontmatter
// internally, computes ranks via iterative propagation, and returns:
//
//   - rankedEntries: one RankedEntry per spec node and per declared artifact.
//   - cycleNames:    logical names of entries involved in dependency cycles
//     (empty if no cycles exist).
//   - error:         wraps ErrUnresolvableRef if any depends_on / input target
//     cannot be resolved.
// nodeInfo pairs a logical name with its parsed frontmatter.
type nodeInfo struct {
	logicalName string
	fm          *frontmatter.Frontmatter
}

func DetectCycles(nodes []nodediscovery.DiscoveredNode) ([]RankedEntry, []string, error) {
	// ── Step 0: Parse frontmatter for every discovered node ──────────────────
	// DiscoveredNode only carries LogicalName and FilePath; we must parse the
	// frontmatter ourselves before we can inspect depends_on / input / outputs.
	nodeInfos := make([]nodeInfo, 0, len(nodes))
	for _, n := range nodes {
		fm, err := frontmatter.ParseFrontmatter(n.FilePath)
		if err != nil {
			return nil, nil, fmt.Errorf("noderanking: parsing frontmatter for %q: %w", n.LogicalName, err)
		}
		nodeInfos = append(nodeInfos, nodeInfo{logicalName: n.LogicalName, fm: fm})
	}

	// ── Step 1: Build entry map ───────────────────────────────────────────────
	entryMap := buildEntryMap(nodeInfos)

	// ── Step 2: Build dependency edges ───────────────────────────────────────
	var unresolvableRefs []string
	entryMap, unresolvableRefs = buildDependencies(entryMap, nodeInfos)

	if len(unresolvableRefs) > 0 {
		return nil, nil, fmt.Errorf("%w: %s", ErrUnresolvableRef, strings.Join(unresolvableRefs, ", "))
	}

	// ── Step 3: Iterative rank propagation ───────────────────────────────────
	totalEntries := len(entryMap)
	passNumber := 0
	var cycleParticipants []string

	for {
		changed := false
		passNumber++

		for key, e := range entryMap {
			computedRank := computeRank(e, entryMap)
			if computedRank > e.rank {
				e.rank = computedRank
				entryMap[key] = e
				changed = true
			}
		}

		if !changed {
			// Ranks have converged — no cycles detected.
			break
		}

		if passNumber >= totalEntries {
			// Ranks are still changing after N passes — a cycle exists.
			// Collect entries whose rank is still increasing in this final pass.
			cycleParticipants = []string{}
			for key, e := range entryMap {
				if len(e.dependencies) == 0 {
					continue
				}
				computedRank := computeRank(e, entryMap)
				if computedRank > e.rank {
					cycleParticipants = append(cycleParticipants, key)
				}
			}
			break
		}
	}

	// ── Step 4: Assemble results ──────────────────────────────────────────────
	rankedEntries := make([]RankedEntry, 0, len(entryMap))
	for _, e := range entryMap {
		rankedEntries = append(rankedEntries, RankedEntry{
			LogicalName: e.logicalName,
			Rank:        e.rank,
		})
	}

	if cycleParticipants == nil {
		cycleParticipants = []string{}
	}

	return rankedEntries, cycleParticipants, nil
}

// ─── BuildEntryMap ────────────────────────────────────────────────────────────

// buildEntryMap creates the initial entry map from the list of node infos.
// For each node it creates:
//   - one "node" entry keyed by the node's logical name, and
//   - one "artifact" entry per output declared in the node's frontmatter,
//     keyed as "ARTIFACT/<path-suffix>(<output-id>)".
//
// All ranks are initialised to 0; dependencies are populated later.
func buildEntryMap(nodeInfos []nodeInfo) map[string]entry {
	em := make(map[string]entry, len(nodeInfos)*2)

	for _, ni := range nodeInfos {
		// ── Node entry ────────────────────────────────────────────────────────
		em[ni.logicalName] = entry{
			logicalName:  ni.logicalName,
			kind:         kindNode,
			dependencies: []string{},
			rank:         0,
		}

		// ── Artifact entries ──────────────────────────────────────────────────
		for _, out := range ni.fm.Outputs {
			// Build the ARTIFACT/ key.
			// "ROOT/x/y" + output id "code" → "ARTIFACT/x/y(code)"
			pathSuffix := stripRootPrefix(ni.logicalName)
			var artifactKey string
			if pathSuffix == "" {
				// Edge case: ROOT node itself declares an output.
				artifactKey = "ARTIFACT/(" + out.ID + ")"
			} else {
				artifactKey = "ARTIFACT/" + pathSuffix + "(" + out.ID + ")"
			}

			em[artifactKey] = entry{
				logicalName:    artifactKey,
				kind:           kindArtifact,
				generatingNode: ni.logicalName,
				dependencies:   []string{},
				rank:           0,
			}
		}
	}

	return em
}

// ─── BuildDependencies ────────────────────────────────────────────────────────

// buildDependencies populates the dependencies slice for every entry in the map.
//
// For node entries, dependencies are:
//  1. The parent node (all nodes except ROOT).
//  2. Each ref in depends_on (after NormalizeRef).
//  3. The input artifact ref (if non-empty), used as-is.
//
// For artifact entries, the only dependency is the node that generates them.
//
// Any reference that cannot be resolved in the entry map is collected in the
// returned unresolvable slice so the caller can surface a single combined error.
func buildDependencies(
	em map[string]entry,
	nodeInfos []nodeInfo,
) (map[string]entry, []string) {
	var unresolvable []string

	// ── Node entries ──────────────────────────────────────────────────────────
	for _, ni := range nodeInfos {
		e := em[ni.logicalName]

		// a. Parent dependency (every node except ROOT depends on its parent).
		if parent, ok := logicalnames.ParentLogicalName(ni.logicalName); ok && parent != "" {
			if _, exists := em[parent]; !exists {
				unresolvable = append(unresolvable, parent)
			} else {
				e.dependencies = append(e.dependencies, parent)
			}
		}

		// b. depends_on references — strip qualifier from ROOT/ refs before lookup.
		for _, ref := range ni.fm.DependsOn {
			lookupKey := normalizeRef(ref)
			if _, exists := em[lookupKey]; !exists {
				unresolvable = append(unresolvable, ref)
			} else {
				e.dependencies = append(e.dependencies, lookupKey)
			}
		}

		// c. input reference — always an artifact ref; used as-is.
		if ni.fm.Input != "" {
			if _, exists := em[ni.fm.Input]; !exists {
				unresolvable = append(unresolvable, ni.fm.Input)
			} else {
				e.dependencies = append(e.dependencies, ni.fm.Input)
			}
		}

		em[ni.logicalName] = e
	}

	// ── Artifact entries ──────────────────────────────────────────────────────
	// Each artifact depends exclusively on the node that generates it.
	for key, e := range em {
		if e.kind != kindArtifact {
			continue
		}
		gen := e.generatingNode
		if _, exists := em[gen]; !exists {
			unresolvable = append(unresolvable, gen)
		} else {
			e.dependencies = append(e.dependencies, gen)
		}
		em[key] = e
	}

	return em, unresolvable
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// computeRank returns the rank an entry should have based on its dependencies'
// current ranks. An entry with no dependencies has rank 0; otherwise it is
// 1 + max(dependency ranks).
func computeRank(e entry, em map[string]entry) int {
	if len(e.dependencies) == 0 {
		return 0
	}
	maxDepRank := 0
	for _, depName := range e.dependencies {
		dep := em[depName]
		if dep.rank > maxDepRank {
			maxDepRank = dep.rank
		}
	}
	return 1 + maxDepRank
}

// normalizeRef strips the parenthetical qualifier from ROOT/ references so they
// can be looked up in the entry map (where node keys are stored without
// qualifiers). ARTIFACT/ references are returned unchanged because their
// qualifier (output ID) is a meaningful part of the key.
//
// Examples:
//
//	"ROOT/x/y(z)"            → "ROOT/x/y"
//	"ARTIFACT/x/y(code)"     → "ARTIFACT/x/y(code)"   (unchanged)
//	"ROOT/x/y"               → "ROOT/x/y"              (no qualifier)
func normalizeRef(ref string) string {
	if strings.HasPrefix(ref, "ARTIFACT/") {
		// Artifact refs include the qualifier as part of the key — leave intact.
		return ref
	}
	if strings.HasPrefix(ref, "ROOT/") || ref == "ROOT" {
		// Strip any trailing "(qualifier)" from ROOT references.
		stripped, _ := logicalnames.HasQualifier(ref)
		if stripped {
			// QualifierName is not needed here; we just want to remove "(…)".
			if idx := strings.LastIndex(ref, "("); idx != -1 {
				return ref[:idx]
			}
		}
		return ref
	}
	// Unrecognized prefix — return as-is; caller will report it unresolvable.
	return ref
}

// stripRootPrefix removes the leading "ROOT/" from a logical name, returning
// only the path suffix. The root node itself ("ROOT") yields an empty string.
//
// Examples:
//
//	"ROOT/x/y" → "x/y"
//	"ROOT"     → ""
func stripRootPrefix(logicalName string) string {
	if logicalName == "ROOT" {
		return ""
	}
	if after, ok := strings.CutPrefix(logicalName, "ROOT/"); ok {
		return after
	}
	// Not a ROOT/ name — return unchanged.
	return logicalName
}
