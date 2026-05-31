// code-from-spec: ROOT/golang/implementation/mcp_tools/validate_specs@OqzClokBzqW44-QU2XQIwYxEBDc

package mcpvalidatespecs

import (
	"errors"
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

// StalenessEntry describes a single output file whose artifact tag is
// missing, malformed, or has a hash that does not match the current
// chain hash.
//
// Status is one of:
//   - "missing"       — the file does not exist.
//   - "stale"         — the file exists but the hash does not match.
//   - "malformed tag" — the file exists but has no artifact tag or the
//     tag cannot be parsed.
//
// Rank is the value returned by NodeRankCompute for the owning node.
// Entries with equal rank have no dependency between them and can be
// processed in parallel.
type StalenessEntry struct {
	Node         string
	OutputID     string
	ArtifactPath string
	Status       string
	Detail       string
	Rank         int
}

// ValidationReport is the value returned by MCPValidateSpecs.
// It aggregates every category of problem found during the scan.
type ValidationReport struct {
	FormatErrors []*spectreevalidate.FormatError
	Cycles       []string
	Staleness    []*StalenessEntry
}

// parsedNode holds the fully parsed state of a single spec node.
type parsedNode struct {
	logicalName string
	filePath    *pathutils.PathCfs
	frontmatter *frontmatter.Frontmatter
	node        *parsenode.Node
}

// MCPValidateSpecs scans the entire spec tree rooted at code-from-spec/,
// validates node format rules, computes dependency ranks, and checks every
// declared output file for a valid, current artifact tag.
//
// It never returns an error. All problems — format violations, dependency
// cycles, and stale or missing output files — are collected in the returned
// ValidationReport.
func MCPValidateSpecs() *ValidationReport {
	// Step 1 — Discover nodes.
	scannedNodes, err := spectree.SpecTreeScan()
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
	var parsedNodes []*parsedNode
	var formatErrors []*spectreevalidate.FormatError

	for _, entry := range scannedNodes {
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

		parsedNodes = append(parsedNodes, &parsedNode{
			logicalName: entry.LogicalName,
			filePath:    entry.FilePath,
			frontmatter: fm,
			node:        nd,
		})
	}

	// Step 3 — Format validation.
	validateInputs := make([]*spectreevalidate.SpecTreeValidateInput, 0, len(parsedNodes))
	for _, pn := range parsedNodes {
		validateInputs = append(validateInputs, &spectreevalidate.SpecTreeValidateInput{
			LogicalName: pn.logicalName,
			Frontmatter: pn.frontmatter,
			Node:        pn.node,
		})
	}
	formatErrors = append(formatErrors, spectreevalidate.SpecTreeValidate(validateInputs)...)

	// Step 4 — Ranking and cycle detection.
	var rankedEntries []*noderanking.NodeRankEntry
	var cycles []string
	rankingFailed := false

	if len(formatErrors) == 0 {
		rankInputs := make([]*noderanking.NodeRankInput, 0, len(parsedNodes))
		for _, pn := range parsedNodes {
			rankInputs = append(rankInputs, &noderanking.NodeRankInput{
				LogicalName: pn.logicalName,
				Frontmatter: pn.frontmatter,
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
				rankingFailed = true
			} else {
				formatErrors = append(formatErrors, &spectreevalidate.FormatError{
					Node:   "",
					Rule:   "ranking",
					Detail: err.Error(),
				})
				rankingFailed = true
			}
		} else {
			rankedEntries = ranked
			cycles = detectedCycles
		}
	}

	_ = rankingFailed

	// Build a rank lookup map for convenience.
	rankByName := make(map[string]int, len(rankedEntries))
	for _, re := range rankedEntries {
		rankByName[re.LogicalName] = re.Rank
	}

	// Step 5 — Staleness detection.
	var stalenessEntries []*StalenessEntry

	// Filter parsed_nodes to those with at least one output.
	var nodesWithOutputs []*parsedNode
	for _, pn := range parsedNodes {
		if len(pn.frontmatter.Outputs) > 0 {
			nodesWithOutputs = append(nodesWithOutputs, pn)
		}
	}

	// Sort based on whether ranking succeeded.
	if len(rankedEntries) > 0 {
		sort.SliceStable(nodesWithOutputs, func(i, j int) bool {
			ri := rankByName[nodesWithOutputs[i].logicalName]
			rj := rankByName[nodesWithOutputs[j].logicalName]
			if ri != rj {
				return ri < rj
			}
			return nodesWithOutputs[i].logicalName < nodesWithOutputs[j].logicalName
		})
	} else {
		sort.SliceStable(nodesWithOutputs, func(i, j int) bool {
			return nodesWithOutputs[i].logicalName < nodesWithOutputs[j].logicalName
		})
	}

	for _, pn := range nodesWithOutputs {
		nodeRank := rankByName[pn.logicalName]

		chain, err := chainresolver.ChainResolve(pn.logicalName)
		if err != nil {
			for _, output := range pn.frontmatter.Outputs {
				stalenessEntries = append(stalenessEntries, &StalenessEntry{
					Node:         pn.logicalName,
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
			for _, output := range pn.frontmatter.Outputs {
				stalenessEntries = append(stalenessEntries, &StalenessEntry{
					Node:         pn.logicalName,
					OutputID:     output.ID,
					ArtifactPath: output.Path,
					Status:       "missing",
					Detail:       err.Error(),
					Rank:         nodeRank,
				})
			}
			continue
		}

		for _, output := range pn.frontmatter.Outputs {
			outputPath := &pathutils.PathCfs{Value: output.Path}

			tag, err := artifacttag.ArtifactTagExtract(outputPath)
			if err != nil {
				if errors.Is(err, artifacttag.ErrFileUnreadable) {
					stalenessEntries = append(stalenessEntries, &StalenessEntry{
						Node:         pn.logicalName,
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
						Node:         pn.logicalName,
						OutputID:     output.ID,
						ArtifactPath: output.Path,
						Status:       "malformed tag",
						Detail:       err.Error(),
						Rank:         nodeRank,
					})
					continue
				}
				// Unknown error — treat as missing.
				stalenessEntries = append(stalenessEntries, &StalenessEntry{
					Node:         pn.logicalName,
					OutputID:     output.ID,
					ArtifactPath: output.Path,
					Status:       "missing",
					Detail:       err.Error(),
					Rank:         nodeRank,
				})
				continue
			}

			if tag.Hash != expectedHash {
				stalenessEntries = append(stalenessEntries, &StalenessEntry{
					Node:         pn.logicalName,
					OutputID:     output.ID,
					ArtifactPath: output.Path,
					Status:       "stale",
					Detail:       "file hash " + tag.Hash + ", expected " + expectedHash,
					Rank:         nodeRank,
				})
			}
			// If hashes match, no entry is added.
		}
	}

	// Step 6 — Assemble report.
	sort.SliceStable(stalenessEntries, func(i, j int) bool {
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
