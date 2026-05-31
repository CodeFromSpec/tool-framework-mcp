// code-from-spec: ROOT/golang/implementation/mcp_tools/validate_specs@YJZaOxOT1bVREZ_G_eqLFIYVd2A

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

// StalenessEntry describes a single output file whose artifact tag is missing,
// malformed, or does not match the current chain hash.
type StalenessEntry struct {
	// Node is the logical name of the owning node (e.g. "ROOT/foo/bar").
	Node string

	// OutputID is the output id declared in the node's frontmatter (e.g. "interface").
	OutputID string

	// ArtifactPath is the relative path of the output file from the project root.
	ArtifactPath string

	// Status is one of "missing", "stale", or "malformed tag".
	Status string

	// Detail provides additional context about the staleness condition.
	Detail string

	// Rank is the topological rank of the node as returned by NodeRankCompute.
	// Entries with equal rank have no dependency between them and can be
	// processed in parallel.
	Rank int
}

// ValidationReport is the result returned by MCPValidateSpecs.
type ValidationReport struct {
	// FormatErrors lists all node format violations found during spec tree
	// validation.
	FormatErrors []*spectreevalidate.FormatError

	// Cycles is a flat list of logical names involved in non-convergence during
	// ranking, as returned by NodeRankCompute.
	Cycles []string

	// Staleness lists all output files that are missing, stale, or have a
	// malformed artifact tag. Entries where the hash matches are not included.
	Staleness []*StalenessEntry
}

// parsedNode holds the parsed frontmatter and node structure for a single node.
type parsedNode struct {
	fm   *frontmatter.Frontmatter
	node *parsenode.Node
}

// MCPValidateSpecs scans the entire spec tree starting from code-from-spec/,
// validates node format, computes dependency ranks, and checks every declared
// output file for a current and matching artifact tag.
//
// It always returns a report. Problems are collected inside the report rather
// than returned as errors.
func MCPValidateSpecs() *ValidationReport {
	// Step 1 — Discover nodes.
	allNodes, err := spectree.SpecTreeScan()
	if err != nil {
		return &ValidationReport{
			FormatErrors: []*spectreevalidate.FormatError{
				{Node: "", Rule: "scan", Detail: err.Error()},
			},
			Cycles:    []string{},
			Staleness: []*StalenessEntry{},
		}
	}

	// Step 2 — Parse all nodes.
	parsedNodes := make(map[string]*parsedNode, len(allNodes))
	var formatErrors []*spectreevalidate.FormatError

	for _, entry := range allNodes {
		fm, err := frontmatter.FrontmatterParse(entry.FilePath)
		if err != nil {
			formatErrors = append(formatErrors, &spectreevalidate.FormatError{
				Node:   entry.LogicalName,
				Rule:   "parse",
				Detail: err.Error(),
			})
			continue
		}

		nd, err := parsenode.NodeParse(entry.LogicalName)
		if err != nil {
			formatErrors = append(formatErrors, &spectreevalidate.FormatError{
				Node:   entry.LogicalName,
				Rule:   "parse",
				Detail: err.Error(),
			})
			continue
		}

		parsedNodes[entry.LogicalName] = &parsedNode{fm: fm, node: nd}
	}

	// Step 3 — Format validation.
	validateInputs := make([]*spectreevalidate.SpecTreeValidateInput, 0, len(parsedNodes))
	for logicalName, pn := range parsedNodes {
		validateInputs = append(validateInputs, &spectreevalidate.SpecTreeValidateInput{
			LogicalName: logicalName,
			Frontmatter: pn.fm,
			Node:        pn.node,
		})
	}
	validationErrs := spectreevalidate.SpecTreeValidate(validateInputs)
	formatErrors = append(formatErrors, validationErrs...)

	// Step 4 — Ranking and cycle detection.
	rankMap := make(map[string]int)
	var cycles []string

	if len(formatErrors) == 0 {
		rankInputs := make([]*noderanking.NodeRankInput, 0, len(parsedNodes))
		for logicalName, pn := range parsedNodes {
			rankInputs = append(rankInputs, &noderanking.NodeRankInput{
				LogicalName: logicalName,
				Frontmatter: pn.fm,
			})
		}

		ranked, detectedCycles, err := noderanking.NodeRankCompute(rankInputs)
		if err != nil {
			if errors.Is(err, noderanking.ErrUnresolvableReference) {
				formatErrors = append(formatErrors, &spectreevalidate.FormatError{
					Node:   "",
					Rule:   "ranking",
					Detail: err.Error(),
				})
				// Leave rankMap empty and cycles empty.
			} else {
				formatErrors = append(formatErrors, &spectreevalidate.FormatError{
					Node:   "",
					Rule:   "ranking",
					Detail: fmt.Sprintf("unexpected ranking error: %v", err),
				})
			}
		} else {
			for _, entry := range ranked {
				rankMap[entry.LogicalName] = entry.Rank
			}
			cycles = detectedCycles
		}
	}

	// Step 5 — Staleness detection.
	// Collect nodes that have outputs.
	type nodeWithOutputs struct {
		logicalName string
		pn          *parsedNode
		rank        int
	}

	var nodesWithOutputs []nodeWithOutputs
	for logicalName, pn := range parsedNodes {
		if len(pn.fm.Outputs) == 0 {
			continue
		}
		rank := 0
		if r, ok := rankMap[logicalName]; ok {
			rank = r
		}
		nodesWithOutputs = append(nodesWithOutputs, nodeWithOutputs{
			logicalName: logicalName,
			pn:          pn,
			rank:        rank,
		})
	}

	// Sort: if rankMap is non-empty, sort by rank then logical name; otherwise sort by logical name.
	if len(rankMap) > 0 {
		sort.Slice(nodesWithOutputs, func(i, j int) bool {
			if nodesWithOutputs[i].rank != nodesWithOutputs[j].rank {
				return nodesWithOutputs[i].rank < nodesWithOutputs[j].rank
			}
			return nodesWithOutputs[i].logicalName < nodesWithOutputs[j].logicalName
		})
	} else {
		sort.Slice(nodesWithOutputs, func(i, j int) bool {
			return nodesWithOutputs[i].logicalName < nodesWithOutputs[j].logicalName
		})
	}

	var stalenessEntries []*StalenessEntry

	for _, nwo := range nodesWithOutputs {
		logicalName := nwo.logicalName
		outputs := nwo.pn.fm.Outputs
		nodeRank := nwo.rank

		// Resolve chain.
		chain, err := chainresolver.ChainResolve(logicalName)
		if err != nil {
			for _, output := range outputs {
				stalenessEntries = append(stalenessEntries, &StalenessEntry{
					Node:         logicalName,
					OutputID:     output.ID,
					ArtifactPath: output.Path,
					Status:       "missing",
					Detail:       err.Error(),
					Rank:         nodeRank,
				})
			}
			continue
		}

		// Compute expected hash.
		expectedHash, err := chainhash.ChainHashCompute(chain)
		if err != nil {
			for _, output := range outputs {
				stalenessEntries = append(stalenessEntries, &StalenessEntry{
					Node:         logicalName,
					OutputID:     output.ID,
					ArtifactPath: output.Path,
					Status:       "missing",
					Detail:       err.Error(),
					Rank:         nodeRank,
				})
			}
			continue
		}

		// Check each output file.
		for _, output := range outputs {
			pathCfs := &pathutils.PathCfs{Value: output.Path}

			tag, err := artifacttag.ArtifactTagExtract(pathCfs)
			if err != nil {
				if errors.Is(err, artifacttag.ErrFileUnreadable) {
					stalenessEntries = append(stalenessEntries, &StalenessEntry{
						Node:         logicalName,
						OutputID:     output.ID,
						ArtifactPath: output.Path,
						Status:       "missing",
						Detail:       err.Error(),
						Rank:         nodeRank,
					})
					continue
				}
				if errors.Is(err, artifacttag.ErrNoTagFound) || errors.Is(err, artifacttag.ErrMalformedTag) {
					stalenessEntries = append(stalenessEntries, &StalenessEntry{
						Node:         logicalName,
						OutputID:     output.ID,
						ArtifactPath: output.Path,
						Status:       "malformed tag",
						Detail:       err.Error(),
						Rank:         nodeRank,
					})
					continue
				}
				// Unexpected error — treat as missing.
				stalenessEntries = append(stalenessEntries, &StalenessEntry{
					Node:         logicalName,
					OutputID:     output.ID,
					ArtifactPath: output.Path,
					Status:       "missing",
					Detail:       err.Error(),
					Rank:         nodeRank,
				})
				continue
			}

			fileHash := tag.Hash
			if fileHash != expectedHash {
				stalenessEntries = append(stalenessEntries, &StalenessEntry{
					Node:         logicalName,
					OutputID:     output.ID,
					ArtifactPath: output.Path,
					Status:       "stale",
					Detail:       fmt.Sprintf("file hash %s does not match expected hash %s", fileHash, expectedHash),
					Rank:         nodeRank,
				})
			}
			// If hashes match, skip — do not add an entry.
		}
	}

	// Step 6 — Sort staleness entries and assemble report.
	sort.Slice(stalenessEntries, func(i, j int) bool {
		if stalenessEntries[i].Rank != stalenessEntries[j].Rank {
			return stalenessEntries[i].Rank < stalenessEntries[j].Rank
		}
		return stalenessEntries[i].Node < stalenessEntries[j].Node
	})

	if formatErrors == nil {
		formatErrors = []*spectreevalidate.FormatError{}
	}
	if cycles == nil {
		cycles = []string{}
	}
	if stalenessEntries == nil {
		stalenessEntries = []*StalenessEntry{}
	}

	return &ValidationReport{
		FormatErrors: formatErrors,
		Cycles:       cycles,
		Staleness:    stalenessEntries,
	}
}
