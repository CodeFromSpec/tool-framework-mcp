// code-from-spec: ROOT/golang/implementation/mcp_tools/validate_specs@Dd85TKR_5yeUj4wbifTFYVqewH4

package mcpvalidatespecs

import (
	"errors"
	"fmt"
	"sort"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/artifacttag"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/noderanking"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectree"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectreevalidate"
)

// StalenessEntry describes a single output artifact that is missing,
// stale, or has a malformed artifact tag.
//
// Artifacts whose hash matches are not included in the report.
type StalenessEntry struct {
	// Node is the logical name of the spec tree node that owns the output.
	Node string

	// OutputID is the id of the output as declared in the node's frontmatter.
	OutputID string

	// ArtifactPath is the relative path to the artifact file.
	ArtifactPath string

	// Status describes the staleness condition. One of:
	//   "missing"       — the file does not exist.
	//   "stale"         — the file exists but the hash does not match.
	//   "malformed tag" — the file exists but has no artifact tag or the
	//                     tag cannot be parsed.
	Status string

	// Detail provides a human-readable explanation of the staleness condition.
	Detail string

	// Rank is the node's rank as returned by NodeRankCompute. Entries with
	// equal rank have no dependency between them and can be processed in
	// parallel.
	Rank int
}

// ValidationReport is the result returned by MCPValidateSpecs.
// It aggregates all discovered problems across the spec tree.
type ValidationReport struct {
	// FormatErrors is the list of format rule violations found during
	// spec tree validation.
	FormatErrors []*spectreevalidate.FormatError

	// Cycles is a flat list of logical names involved in non-convergence
	// during ranking, as returned by NodeRankCompute.
	Cycles []string

	// Staleness is the list of output artifacts that are missing, stale,
	// or have a malformed artifact tag.
	Staleness []*StalenessEntry
}

// parsedNode holds the cached parse result for a single spec tree node.
type parsedNode struct {
	logicalName string
	fm          *frontmatter.Frontmatter
	node        *parsenode.Node
}

// MCPValidateSpecs scans the entire spec tree starting from
// "code-from-spec/", validates all nodes, and checks each declared
// output artifact for staleness.
//
// It never returns an error. All discovered problems — format errors,
// ranking cycles, and stale or missing artifacts — are collected and
// returned in a ValidationReport.
func MCPValidateSpecs() *ValidationReport {
	var formatErrors []*spectreevalidate.FormatError

	// Step 1 — Discover nodes.
	discoveredNodes, err := spectree.SpecTreeScan()
	if err != nil {
		formatErrors = append(formatErrors, &spectreevalidate.FormatError{
			Node:   "",
			Rule:   "scan",
			Detail: err.Error(),
		})
		return &ValidationReport{
			FormatErrors: formatErrors,
			Cycles:       []string{},
			Staleness:    []*StalenessEntry{},
		}
	}

	// Step 2 — Parse all nodes.
	// Cache keyed by logical name, preserving insertion order via a slice.
	cacheMap := make(map[string]*parsedNode)
	var cacheOrder []string

	for _, n := range discoveredNodes {
		logicalName := n.LogicalName
		filePath := n.FilePath

		// 2a. Parse frontmatter.
		fm, fmErr := frontmatter.FrontmatterParse(&filePath)
		if fmErr != nil {
			formatErrors = append(formatErrors, &spectreevalidate.FormatError{
				Node:   logicalName,
				Rule:   "parse",
				Detail: fmErr.Error(),
			})
			continue
		}

		// 2b. Parse node.
		nd, ndErr := parsenode.NodeParse(logicalName)
		if ndErr != nil {
			formatErrors = append(formatErrors, &spectreevalidate.FormatError{
				Node:   logicalName,
				Rule:   "parse",
				Detail: ndErr.Error(),
			})
			continue
		}

		// 2c. Store in cache.
		cacheMap[logicalName] = &parsedNode{
			logicalName: logicalName,
			fm:          fm,
			node:        nd,
		}
		cacheOrder = append(cacheOrder, logicalName)
	}

	// Step 3 — Format validation.
	var validateInputs []*spectreevalidate.SpecTreeValidateInput
	for _, name := range cacheOrder {
		entry := cacheMap[name]
		validateInputs = append(validateInputs, &spectreevalidate.SpecTreeValidateInput{
			LogicalName: entry.logicalName,
			Frontmatter: entry.fm,
			Node:        entry.node,
		})
	}

	validationFormatErrors := spectreevalidate.SpecTreeValidate(validateInputs)
	formatErrors = append(formatErrors, validationFormatErrors...)

	// Step 4 — Ranking and cycle detection.
	var rankedEntries []*noderanking.NodeRankEntry
	var cycleList []string

	if len(formatErrors) > 0 {
		// Skip ranking when there are format errors.
		rankedEntries = []*noderanking.NodeRankEntry{}
		cycleList = []string{}
	} else {
		// 4a. Build ranking inputs.
		var rankInputs []*noderanking.NodeRankInput
		for _, name := range cacheOrder {
			entry := cacheMap[name]
			rankInputs = append(rankInputs, &noderanking.NodeRankInput{
				LogicalName: entry.logicalName,
				Frontmatter: entry.fm,
			})
		}

		// 4b. Compute ranking.
		ranked, cycles, rankErr := noderanking.NodeRankCompute(rankInputs)
		if rankErr != nil {
			if errors.Is(rankErr, noderanking.ErrUnresolvableReference) {
				formatErrors = append(formatErrors, &spectreevalidate.FormatError{
					Node:   "",
					Rule:   "ranking",
					Detail: rankErr.Error(),
				})
			} else {
				formatErrors = append(formatErrors, &spectreevalidate.FormatError{
					Node:   "",
					Rule:   "ranking",
					Detail: rankErr.Error(),
				})
			}
			rankedEntries = []*noderanking.NodeRankEntry{}
			cycleList = []string{}
		} else {
			rankedEntries = ranked
			cycleList = cycles
		}
	}

	// Step 5 — Staleness detection.

	// Step 7 — Build rank lookup by logical name.
	rankLookup := make(map[string]int)
	for _, entry := range rankedEntries {
		rankLookup[entry.LogicalName] = entry.Rank
	}

	// Step 8 — Collect nodes with outputs and sort them.
	var nodesWithOutputs []*parsedNode
	for _, name := range cacheOrder {
		entry := cacheMap[name]
		if len(entry.fm.Outputs) > 0 {
			nodesWithOutputs = append(nodesWithOutputs, entry)
		}
	}

	sort.Slice(nodesWithOutputs, func(i, j int) bool {
		rankI := rankLookup[nodesWithOutputs[i].logicalName]
		rankJ := rankLookup[nodesWithOutputs[j].logicalName]
		if rankI != rankJ {
			return rankI < rankJ
		}
		return nodesWithOutputs[i].logicalName < nodesWithOutputs[j].logicalName
	})

	// Step 9 — Initialize staleness list.
	var stalenessEntries []*StalenessEntry

	// Step 10 — Check staleness for each node with outputs.
	for _, entry := range nodesWithOutputs {
		logicalName := entry.logicalName
		nodeRank := rankLookup[logicalName]

		// 10a. Resolve the chain.
		chain, chainErr := chainresolver.ChainResolve(logicalName)
		if chainErr != nil {
			for _, output := range entry.fm.Outputs {
				stalenessEntries = append(stalenessEntries, &StalenessEntry{
					Node:         logicalName,
					OutputID:     output.ID,
					ArtifactPath: output.Path,
					Status:       "missing",
					Detail:       chainErr.Error(),
					Rank:         nodeRank,
				})
			}
			continue
		}

		// 10b. Compute the chain hash.
		chainHash, hashErr := chainhash.ChainHashCompute(chain)
		if hashErr != nil {
			for _, output := range entry.fm.Outputs {
				stalenessEntries = append(stalenessEntries, &StalenessEntry{
					Node:         logicalName,
					OutputID:     output.ID,
					ArtifactPath: output.Path,
					Status:       "missing",
					Detail:       hashErr.Error(),
					Rank:         nodeRank,
				})
			}
			continue
		}

		// 10c. Check each output artifact.
		for _, output := range entry.fm.Outputs {
			cfsPath := &pathutils.PathCfs{Value: output.Path}

			tag, tagErr := artifacttag.ArtifactTagExtract(cfsPath)
			if tagErr != nil {
				if errors.Is(tagErr, artifacttag.ErrFileUnreadable) {
					stalenessEntries = append(stalenessEntries, &StalenessEntry{
						Node:         logicalName,
						OutputID:     output.ID,
						ArtifactPath: output.Path,
						Status:       "missing",
						Detail:       tagErr.Error(),
						Rank:         nodeRank,
					})
				} else if errors.Is(tagErr, artifacttag.ErrNoTagFound) || errors.Is(tagErr, artifacttag.ErrMalformedTag) {
					stalenessEntries = append(stalenessEntries, &StalenessEntry{
						Node:         logicalName,
						OutputID:     output.ID,
						ArtifactPath: output.Path,
						Status:       "malformed tag",
						Detail:       tagErr.Error(),
						Rank:         nodeRank,
					})
				} else {
					stalenessEntries = append(stalenessEntries, &StalenessEntry{
						Node:         logicalName,
						OutputID:     output.ID,
						ArtifactPath: output.Path,
						Status:       "missing",
						Detail:       tagErr.Error(),
						Rank:         nodeRank,
					})
				}
				continue
			}

			// Compare the tag hash with the computed chain hash.
			if tag.Hash != chainHash {
				stalenessEntries = append(stalenessEntries, &StalenessEntry{
					Node:         logicalName,
					OutputID:     output.ID,
					ArtifactPath: output.Path,
					Status:       "stale",
					Detail:       fmt.Sprintf("file hash %s does not match expected hash %s", tag.Hash, chainHash),
					Rank:         nodeRank,
				})
			}
			// If hashes match, skip — do not add an entry.
		}
	}

	// Step 11 — Sort staleness entries by rank ascending, then logical name ascending.
	sort.Slice(stalenessEntries, func(i, j int) bool {
		if stalenessEntries[i].Rank != stalenessEntries[j].Rank {
			return stalenessEntries[i].Rank < stalenessEntries[j].Rank
		}
		return stalenessEntries[i].Node < stalenessEntries[j].Node
	})

	// Step 12 — Assemble and return the report.
	if cycleList == nil {
		cycleList = []string{}
	}
	if formatErrors == nil {
		formatErrors = []*spectreevalidate.FormatError{}
	}
	if stalenessEntries == nil {
		stalenessEntries = []*StalenessEntry{}
	}

	return &ValidationReport{
		FormatErrors: formatErrors,
		Cycles:       cycleList,
		Staleness:    stalenessEntries,
	}
}
