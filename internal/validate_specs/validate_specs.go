// code-from-spec: ROOT/golang/internal/tools/validate_specs/code@PENDING
package validate_specs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/artifacttag"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/formatvalidation"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/nodediscovery"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/noderanking"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/parsenode"
	"github.com/goccy/go-yaml"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ValidateSpecsArgs defines the input parameters for the validate_specs tool.
// No input parameters; scans the entire spec tree.
type ValidateSpecsArgs struct{}

// FormatErrorEntry represents a single format error in the report.
type FormatErrorEntry struct {
	Node    string `yaml:"node"`
	Rule    string `yaml:"rule,omitempty"`
	Message string `yaml:"message,omitempty"`
	Detail  string `yaml:"detail,omitempty"`
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

	// Step 2: Parse all nodes (cache frontmatter and parsed bodies).
	fmCache := make(map[string]*frontmatter.Frontmatter)
	parsedCache := make(map[string]*parsenode.ParsedNode)
	var parseErrors []FormatErrorEntry

	for _, node := range nodes {
		fm, err := frontmatter.ParseFrontmatter(node.FilePath)
		if err != nil {
			parseErrors = append(parseErrors, FormatErrorEntry{
				Node:    node.LogicalName,
				Message: fmt.Sprintf("unreadable file: %v", err),
			})
			continue
		}
		fmCache[node.LogicalName] = fm

		parsed, err := parsenode.ParseNode(node.LogicalName)
		if err != nil {
			parseErrors = append(parseErrors, FormatErrorEntry{
				Node:    node.LogicalName,
				Message: fmt.Sprintf("parse failure: %v", err),
			})
			continue
		}
		parsedCache[node.LogicalName] = parsed
	}

	// Step 3: Format validation.
	var formatErrors []FormatErrorEntry
	formatErrors = append(formatErrors, parseErrors...)

	fmtErrors, err := formatvalidation.ValidateFormat(nodes)
	if err != nil {
		// Non-fatal: include as a format error.
		formatErrors = append(formatErrors, FormatErrorEntry{
			Node:    "",
			Message: fmt.Sprintf("format validation error: %v", err),
		})
	}
	for _, fe := range fmtErrors {
		formatErrors = append(formatErrors, FormatErrorEntry{
			Node:   fe.Node,
			Rule:   fe.Rule,
			Detail: fe.Detail,
		})
	}

	// Step 4: Ranking and cycle detection.
	ranked, cycles, err := noderanking.DetectCycles(nodes)
	if err != nil && !errors.Is(err, noderanking.ErrUnresolvableRef) {
		// Non-fatal for unresolvable refs.
		formatErrors = append(formatErrors, FormatErrorEntry{
			Node:    "",
			Message: fmt.Sprintf("ranking error: %v", err),
		})
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

	// Step 5: Staleness detection.
	// Collect nodes with outputs, sorted by rank.
	type nodeWithRank struct {
		logicalName string
		fm          *frontmatter.Frontmatter
		rank        int
	}
	var outputNodes []nodeWithRank
	for _, node := range nodes {
		fm, ok := fmCache[node.LogicalName]
		if !ok || len(fm.Outputs) == 0 {
			continue
		}
		rank, ok := rankMap[node.LogicalName]
		if !ok {
			rank = 999999 // Unranked nodes go last.
		}
		outputNodes = append(outputNodes, nodeWithRank{
			logicalName: node.LogicalName,
			fm:          fm,
			rank:        rank,
		})
	}
	sort.Slice(outputNodes, func(i, j int) bool {
		return outputNodes[i].rank < outputNodes[j].rank
	})

	var staleness []StalenessEntry
	for _, on := range outputNodes {
		chainHash, err := chainhash.ComputeChainHash(on.logicalName)
		if err != nil {
			// treat as stale — cannot compute hash
			chainHash = ""
		}

		for _, out := range on.fm.Outputs {
			if _, err := os.Stat(out.Path); os.IsNotExist(err) {
				staleness = append(staleness, StalenessEntry{
					Node:         on.logicalName,
					ArtifactPath: out.Path,
					Status:       "missing",
				})
				continue
			}
			tag, err := artifacttag.ExtractArtifactTag(out.Path)
			if err != nil || tag.Hash != chainHash {
				staleness = append(staleness, StalenessEntry{
					Node:         on.logicalName,
					ArtifactPath: out.Path,
					Status:       "stale",
				})
			}
		}
	}

	// Step 6: Build report.
	report := ValidationReport{
		FormatErrors:       formatErrors,
		CircularReferences: circularRefs,
		Staleness:          staleness,
	}

	yamlBytes, err := yaml.Marshal(report)
	if err != nil {
		return toolError(fmt.Sprintf("report serialization failure: %v", err)), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(yamlBytes)}},
	}, nil, nil
}

// toolError returns a CallToolResult with IsError set to true.
func toolError(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		IsError: true,
	}
}
