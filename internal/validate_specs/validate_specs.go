// code-from-spec: ROOT/golang/internal/tools/validate_specs/code@DEr0GnERqEHco4xjSDibyK3GR1U

// Package validate_specs implements the validate_specs MCP tool.
//
// The tool scans the entire spec tree and produces a validation report
// covering three categories:
//   1. Format errors   — structural rule violations in individual node files.
//   2. Cycles          — nodes that form circular dependency chains.
//   3. Staleness       — generated output files whose artifact-tag hash does
//                        not match the current chain hash.
//
// All errors are collected; the handler never stops at the first problem.
package validate_specs

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	goyaml "github.com/goccy/go-yaml"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/artifacttag"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/formatvalidation"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/nodediscovery"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/noderanking"
)

// ValidateSpecsArgs holds the (empty) input parameters for the validate_specs tool.
type ValidateSpecsArgs struct{}

// FormatErrorEntry represents a single format error in the report.
type FormatErrorEntry struct {
	Node   string `yaml:"node"`
	Rule   string `yaml:"rule,omitempty"`
	Detail string `yaml:"detail,omitempty"`
}

// StalenessEntry represents a staleness finding for an artifact.
type StalenessEntry struct {
	Node         string `yaml:"node"`
	ArtifactPath string `yaml:"artifact_path"`
	Status       string `yaml:"status"`
}

// ValidationReport is the YAML-serializable validation result.
type ValidationReport struct {
	FormatErrors       []FormatErrorEntry `yaml:"format_errors"`
	CircularReferences [][]string         `yaml:"circular_references"`
	Staleness          []StalenessEntry   `yaml:"staleness"`
}

// ---------------------------------------------------------------------------
// Handler
// ---------------------------------------------------------------------------

// HandleValidateSpecs discovers all spec nodes, validates format,
// detects circular references, and checks artifact staleness.
func HandleValidateSpecs(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args ValidateSpecsArgs,
) (*mcp.CallToolResult, any, error) {
	// Step 1: Discover nodes.
	nodes, err := nodediscovery.DiscoverNodes()
	if err != nil {
		return toolError(fmt.Sprintf("node discovery failure: %v", err)), nil, nil
	}

	// Step 2: Format validation.
	var formatErrors []FormatErrorEntry

	fmtErrors, err := formatvalidation.ValidateFormat(nodes)
	if err != nil {
		formatErrors = append(formatErrors, FormatErrorEntry{
			Node:   "",
			Detail: fmt.Sprintf("format validation error: %v", err),
		})
	}
	for _, fe := range fmtErrors {
		formatErrors = append(formatErrors, FormatErrorEntry{
			Node:   fe.Node,
			Rule:   fe.Rule,
			Detail: fe.Detail,
		})
	}

	// Step 3: Ranking and cycle detection.
	ranked, cycles, rankErr := noderanking.DetectCycles(nodes)

	if rankErr != nil {
		if errors.Is(rankErr, noderanking.ErrUnresolvableRef) {
			formatErrors = append(formatErrors, FormatErrorEntry{
				Node:   "",
				Rule:   "unresolvable_ref",
				Detail: rankErr.Error(),
			})
		} else {
			formatErrors = append(formatErrors, FormatErrorEntry{
				Node:   "",
				Rule:   "ranking_error",
				Detail: fmt.Sprintf("node ranking failed: %v", rankErr),
			})
		}
	}

	var circularRefs [][]string
	if len(cycles) > 0 {
		circularRefs = append(circularRefs, cycles)
	}

	// Build rank map for ordering.
	rankMap := make(map[string]int)
	for _, r := range ranked {
		rankMap[r.LogicalName] = r.Rank
	}

	// Step 4: Staleness detection.
	type nodeWithRank struct {
		logicalName string
		fm          *frontmatter.Frontmatter
		rank        int
	}
	var outputNodes []nodeWithRank
	for _, node := range nodes {
		fm, fmErr := frontmatter.ParseFrontmatter(node.FilePath)
		if fmErr != nil || len(fm.Outputs) == 0 {
			continue
		}
		rank := rankMap[node.LogicalName]
		outputNodes = append(outputNodes, nodeWithRank{
			logicalName: node.LogicalName,
			fm:          fm,
			rank:        rank,
		})
	}
	sort.Slice(outputNodes, func(i, j int) bool {
		if outputNodes[i].rank != outputNodes[j].rank {
			return outputNodes[i].rank < outputNodes[j].rank
		}
		return outputNodes[i].logicalName < outputNodes[j].logicalName
	})

	var staleness []StalenessEntry
	for _, on := range outputNodes {
		chainHash, hashErr := chainhash.ComputeChainHash(on.logicalName)
		if hashErr != nil {
			chainHash = ""
		}

		for _, out := range on.fm.Outputs {
			tag, tagErr := artifacttag.ExtractArtifactTag(out.Path)
			if tagErr != nil {
				status := "missing"
				if !errors.Is(tagErr, artifacttag.ErrFileUnreadable) {
					status = "stale"
				}
				staleness = append(staleness, StalenessEntry{
					Node:         on.logicalName,
					ArtifactPath: out.Path,
					Status:       status,
				})
				continue
			}
			if tag.Hash != chainHash {
				staleness = append(staleness, StalenessEntry{
					Node:         on.logicalName,
					ArtifactPath: out.Path,
					Status:       "stale",
				})
			}
		}
	}

	// Step 5: Build report.
	report := ValidationReport{
		FormatErrors:       formatErrors,
		CircularReferences: circularRefs,
		Staleness:          staleness,
	}

	yamlBytes, err := goyaml.Marshal(report)
	if err != nil {
		return toolError(fmt.Sprintf("report serialization failure: %v", err)), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(yamlBytes)}},
	}, nil, nil
}

func toolError(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		IsError: true,
	}
}
