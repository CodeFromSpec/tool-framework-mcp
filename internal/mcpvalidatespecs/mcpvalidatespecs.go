// code-from-spec: ROOT/golang/implementation/mcp_tools/validate_specs@5viw21rnR8x0xOsxxlkk_c70-7k
package mcpvalidatespecs

import (
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
// unparseable, or whose hash does not match the current chain hash.
type StalenessEntry struct {
	// Node is the logical name of the node that owns the output.
	Node string

	// OutputID is the id field from the node's outputs frontmatter.
	OutputID string

	// ArtifactPath is the relative path to the generated file.
	ArtifactPath string

	// Status is one of "missing", "stale", or "malformed tag".
	Status string

	// Detail provides additional context about the staleness condition.
	Detail string

	// Rank is the dependency rank of the node as returned by NodeRankCompute.
	// Entries with equal rank have no dependency between them and can be
	// processed in parallel.
	Rank int
}

// ValidationReport is the result of MCPValidateSpecs. It collects all
// problems found across the spec tree.
type ValidationReport struct {
	// FormatErrors lists nodes that violate spec-tree structural rules.
	FormatErrors []*spectreevalidate.FormatError

	// Cycles is a flat list of logical names involved in non-convergence
	// during ranking, as returned by NodeRankCompute.
	Cycles []string

	// Staleness lists output files whose artifact tag is missing, stale,
	// or malformed. Files whose hash matches are not included.
	Staleness []*StalenessEntry
}

// parsedEntry holds the cached parse results for a single node.
type parsedEntry struct {
	fm   *frontmatter.Frontmatter
	node *parsenode.Node
}

// MCPValidateSpecs scans the entire spec tree rooted at code-from-spec/,
// validates node structure, computes dependency ranks, and checks every
// declared output file for a current artifact tag.
//
// It always returns a report. Callers should inspect FormatErrors, Cycles,
// and Staleness to determine whether any problems were found.
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
	parsedNodes := make(map[string]*parsedEntry, len(allNodes))
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

		node, err := parsenode.NodeParse(entry.LogicalName)
		if err != nil {
			formatErrors = append(formatErrors, &spectreevalidate.FormatError{
				Node:   entry.LogicalName,
				Rule:   "parse",
				Detail: err.Error(),
			})
			continue
		}

		parsedNodes[entry.LogicalName] = &parsedEntry{fm: fm, node: node}
	}

	// Step 3 — Format validation.
	validateInputs := make([]*spectreevalidate.SpecTreeValidateInput, 0, len(parsedNodes))
	for logicalName, pe := range parsedNodes {
		validateInputs = append(validateInputs, &spectreevalidate.SpecTreeValidateInput{
			LogicalName: logicalName,
			Frontmatter: pe.fm,
			Node:        pe.node,
		})
	}

	validationErrs := spectreevalidate.SpecTreeValidate(validateInputs)
	formatErrors = append(formatErrors, validationErrs...)

	// Step 4 — Ranking and cycle detection.
	rankMap := make(map[string]int)
	var cycles []string

	if len(formatErrors) == 0 {
		rankInputs := make([]*noderanking.NodeRankInput, 0, len(parsedNodes))
		for logicalName, pe := range parsedNodes {
			rankInputs = append(rankInputs, &noderanking.NodeRankInput{
				LogicalName: logicalName,
				Frontmatter: pe.fm,
			})
		}

		ranked, rankCycles, err := noderanking.NodeRankCompute(rankInputs)
		if err != nil {
			formatErrors = append(formatErrors, &spectreevalidate.FormatError{
				Node:   "",
				Rule:   "ranking",
				Detail: err.Error(),
			})
			// Leave rankMap empty and cycles empty.
		} else {
			for _, entry := range ranked {
				rankMap[entry.LogicalName] = entry.Rank
			}
			cycles = rankCycles
		}
	}

	// Step 5 — Staleness detection.
	// Collect nodes that have outputs.
	type nodeWithOutputs struct {
		logicalName string
		pe          *parsedEntry
	}

	nodesWithOutputs := make([]nodeWithOutputs, 0)
	for logicalName, pe := range parsedNodes {
		if len(pe.fm.Outputs) > 0 {
			nodesWithOutputs = append(nodesWithOutputs, nodeWithOutputs{
				logicalName: logicalName,
				pe:          pe,
			})
		}
	}

	// Determine processing order.
	if len(rankMap) > 0 {
		sort.Slice(nodesWithOutputs, func(i, j int) bool {
			ri := rankMap[nodesWithOutputs[i].logicalName]
			rj := rankMap[nodesWithOutputs[j].logicalName]
			if ri != rj {
				return ri < rj
			}
			return nodesWithOutputs[i].logicalName < nodesWithOutputs[j].logicalName
		})
	} else {
		sort.Slice(nodesWithOutputs, func(i, j int) bool {
			return nodesWithOutputs[i].logicalName < nodesWithOutputs[j].logicalName
		})
	}

	var stalenessEntries []*StalenessEntry

	for _, n := range nodesWithOutputs {
		logicalName := n.logicalName
		outputs := n.pe.fm.Outputs
		nodeRank := rankMap[logicalName]

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

		for _, output := range outputs {
			pathCfs := &pathutils.PathCfs{Value: output.Path}

			tag, err := artifacttag.ArtifactTagExtract(pathCfs)
			if err != nil {
				var status string
				if err == artifacttag.ErrFileUnreadable {
					status = "missing"
				} else {
					// ErrNoTagFound or ErrMalformedTag
					status = "malformed tag"
				}
				stalenessEntries = append(stalenessEntries, &StalenessEntry{
					Node:         logicalName,
					OutputID:     output.ID,
					ArtifactPath: output.Path,
					Status:       status,
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

	// Step 6 — Assemble and return report.
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
