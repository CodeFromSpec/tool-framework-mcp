// code-from-spec: ROOT/golang/internal/tools/validate_specs/code@JEI0T3s4rXrqbitJ3psC7jUbndw

// Package validate_specs implements the validate_specs MCP tool.
//
// The tool scans the entire spec tree and reports:
//   - Format errors (structural rule violations per node)
//   - Circular references between nodes
//   - Stale or missing generated artifacts (hash mismatch)
//
// It uses the following internal packages:
//   - nodediscovery    — walk the spec tree and enumerate all nodes
//   - frontmatter      — parse YAML frontmatter from node files
//   - parsenode        — parse the markdown body into sections
//   - formatvalidation — validate structural rules per node
//   - noderanking      — topological sort + cycle detection
//   - artifacttag      — extract the code-from-spec tag from generated files
//   - chainhash        — compute the expected chain hash for a node
package validate_specs

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/artifacttag"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/formatvalidation"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/nodediscovery"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/noderanking"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/parsenode"
)

// ValidateSpecsArgs is the (empty) argument struct for the validate_specs tool.
// No input parameters are required — the tool always scans the full spec tree.
type ValidateSpecsArgs struct{}

// stalenessEntry records a single artifact staleness finding.
// Promoted to package level so it can be shared between the handler and
// buildReport without requiring an anonymous struct conversion.
type stalenessEntry struct {
	logicalName string
	outputID    string
	outputPath  string
	rank        int
	status      string // "missing" or "stale"
	detail      string // extra context (e.g., hash found vs expected)
}

// HandleValidateSpecs is the MCP tool handler for validate_specs.
//
// Flow:
//  1. Discover all _node.md files in the spec tree.
//  2. Validate format rules for every node.
//  3. Detect cycles using noderanking.
//  4. For nodes with outputs, check artifact staleness in rank order.
//  5. Return a human-readable report as a success result.
//
// All errors are collected across steps; the handler never stops at the first
// failure (per spec: "Reports all errors found — does not stop at the first").
func HandleValidateSpecs(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args ValidateSpecsArgs,
) (*mcp.CallToolResult, any, error) {

	// -------------------------------------------------------------------------
	// Step 1 — Discover all spec nodes
	// -------------------------------------------------------------------------
	nodes, err := nodediscovery.DiscoverNodes()
	if err != nil {
		// Discovery failure is fatal: we cannot proceed without the node list.
		// Provide an actionable message so the agent knows what to fix.
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf(
					"node discovery failed: %v\n\nEnsure the tool is run from the project root and that code-from-spec/ exists.",
					err,
				),
			}},
			IsError: true,
		}, nil, nil
	}

	// -------------------------------------------------------------------------
	// Step 2 — Format validation
	//
	// ValidateFormat checks structural rules (frontmatter, headings, paths, …)
	// for every node. Unreadable nodes are wrapped in FormatError entries;
	// processing continues for the remaining nodes.
	// -------------------------------------------------------------------------
	formatErrors, fvErr := formatvalidation.ValidateFormat(nodes)
	if fvErr != nil && !errors.Is(fvErr, formatvalidation.ErrUnreadableNode) {
		// An unexpected internal error (not a per-node problem).
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("format validation encountered an unexpected error: %v", fvErr),
			}},
			IsError: true,
		}, nil, nil
	}

	// -------------------------------------------------------------------------
	// Step 3 — Cycle detection and topological ranking
	//
	// Skipped entirely when format errors exist — ranking depends on valid
	// frontmatter, so the results would be unreliable.
	// -------------------------------------------------------------------------
	var rankedEntries []noderanking.RankedEntry
	var cycleParticipants []string
	var rankErr error

	if len(formatErrors) == 0 {
		rankedEntries, cycleParticipants, rankErr = noderanking.DetectCycles(nodes)
		if rankErr != nil && !errors.Is(rankErr, noderanking.ErrUnresolvableRef) {
			// Unexpected internal error from noderanking.
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("node ranking encountered an unexpected error: %v", rankErr),
				}},
				IsError: true,
			}, nil, nil
		}
	}

	// Build a rank map (logicalName → rank) for ordering staleness entries.
	// Nodes that did not appear in rankedEntries (e.g., cycle participants)
	// will be absent from the map and treated as MaxInt rank.
	rankMap := make(map[string]int, len(rankedEntries))
	for _, entry := range rankedEntries {
		rankMap[entry.LogicalName] = entry.Rank
	}

	// -------------------------------------------------------------------------
	// Step 4 — Staleness check
	//
	// For each node that declares outputs:
	//   a. Compute the expected chain hash.
	//   b. For each output path, extract the artifact tag from the file.
	//   c. If the file is missing or has no tag → report "missing".
	//      If the hash in the tag differs from expected → report "stale".
	//      If the hashes match → skip (up to date).
	//
	// Nodes are processed in rank order (lowest rank first) so the final report
	// lists staleness entries in dependency order.
	// -------------------------------------------------------------------------

	// Cache frontmatter and parsed node bodies to avoid re-reading files.
	// Each entry either holds valid data or records the error encountered.
	type parsedCacheEntry struct {
		fm  *frontmatter.Frontmatter
		pn  *parsenode.ParsedNode
		err error // non-nil means the node could not be fully parsed
	}
	parsedCache := make(map[string]parsedCacheEntry, len(nodes))

	for _, node := range nodes {
		fm, fmErr := frontmatter.ParseFrontmatter(node.FilePath)
		if fmErr != nil {
			// Unreadable/unparseable — already captured by formatvalidation.
			// Record the error so staleness check can skip this node.
			parsedCache[node.LogicalName] = parsedCacheEntry{err: fmErr}
			continue
		}

		pn, pnErr := parsenode.ParseNode(node.LogicalName)
		if pnErr != nil {
			// Body parse error — also captured by formatvalidation.
			parsedCache[node.LogicalName] = parsedCacheEntry{fm: fm, err: pnErr}
			continue
		}

		parsedCache[node.LogicalName] = parsedCacheEntry{fm: fm, pn: pn}
	}

	// Sort a copy of the node list by rank (ascending) for ordered processing.
	const maxInt = int(^uint(0) >> 1)
	nodesSortedByRank := make([]nodediscovery.DiscoveredNode, len(nodes))
	copy(nodesSortedByRank, nodes)
	sort.Slice(nodesSortedByRank, func(i, j int) bool {
		ri, oki := rankMap[nodesSortedByRank[i].LogicalName]
		rj, okj := rankMap[nodesSortedByRank[j].LogicalName]
		if !oki {
			ri = maxInt
		}
		if !okj {
			rj = maxInt
		}
		return ri < rj
	})

	var stalenessEntries []stalenessEntry

	for _, node := range nodesSortedByRank {
		cached, ok := parsedCache[node.LogicalName]
		if !ok || cached.err != nil || cached.fm == nil {
			// Node had a parse error; skip staleness check for it.
			continue
		}

		if len(cached.fm.Outputs) == 0 {
			// No outputs declared on this node — nothing to check.
			continue
		}

		// Compute the expected chain hash for this node.
		expectedHash, hashErr := chainhash.ComputeChainHash(node.LogicalName)
		if hashErr != nil {
			// Cannot determine expected hash — report each output as missing.
			for _, out := range cached.fm.Outputs {
				stalenessEntries = append(stalenessEntries, stalenessEntry{
					logicalName: node.LogicalName,
					outputID:    out.ID,
					outputPath:  out.Path,
					rank:        rankMap[node.LogicalName],
					status:      "missing",
					detail:      fmt.Sprintf("chain hash computation failed: %v", hashErr),
				})
			}
			continue
		}

		// Check each declared output file.
		for _, out := range cached.fm.Outputs {
			tag, tagErr := artifacttag.ExtractArtifactTag(out.Path)
			if tagErr != nil {
				status := "missing"
				detail := ""

				switch {
				case errors.Is(tagErr, artifacttag.ErrFileUnreadable):
					// File does not exist or cannot be opened.
					detail = "file not found or unreadable"

				case errors.Is(tagErr, artifacttag.ErrNoTagFound):
					// File exists but carries no code-from-spec tag.
					detail = "file exists but contains no code-from-spec tag"

				case errors.Is(tagErr, artifacttag.ErrMalformedTag):
					// Tag present but cannot be parsed — treat as stale, not missing.
					status = "stale"
					detail = "code-from-spec tag is malformed"

				default:
					// Unexpected error from the artifact tag extractor.
					detail = fmt.Sprintf("artifact tag extraction error: %v", tagErr)
				}

				stalenessEntries = append(stalenessEntries, stalenessEntry{
					logicalName: node.LogicalName,
					outputID:    out.ID,
					outputPath:  out.Path,
					rank:        rankMap[node.LogicalName],
					status:      status,
					detail:      detail,
				})
				continue
			}

			// Tag extracted successfully — compare hashes.
			if tag.Hash != expectedHash {
				stalenessEntries = append(stalenessEntries, stalenessEntry{
					logicalName: node.LogicalName,
					outputID:    out.ID,
					outputPath:  out.Path,
					rank:        rankMap[node.LogicalName],
					status:      "stale",
					detail:      fmt.Sprintf("file hash %q does not match expected %q", tag.Hash, expectedHash),
				})
			}
			// Hash matches — artifact is current; no entry needed.
		}
	}

	// -------------------------------------------------------------------------
	// Step 5 — Assemble and return the validation report
	// -------------------------------------------------------------------------
	report := buildReport(formatErrors, cycleParticipants, stalenessEntries)

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: report}},
	}, nil, nil
}

// buildReport assembles a human-readable validation report from the three
// categories of findings: format errors, cycle participants, and staleness
// entries. Each section is only included if there are findings in it.
func buildReport(
	formatErrors []formatvalidation.FormatError,
	cycles []string,
	staleEntries []stalenessEntry,
) string {
	var sb strings.Builder

	// --- Format errors ---
	if len(formatErrors) > 0 {
		fmt.Fprintf(&sb, "FORMAT ERRORS (%d)\n", len(formatErrors))
		sb.WriteString(strings.Repeat("-", 40) + "\n")
		for _, fe := range formatErrors {
			fmt.Fprintf(&sb, "  node:   %s\n", fe.Node)
			fmt.Fprintf(&sb, "  rule:   %s\n", fe.Rule)
			if fe.Detail != "" {
				fmt.Fprintf(&sb, "  detail: %s\n", fe.Detail)
			}
			sb.WriteString("\n")
		}
	}

	// --- Cycle participants ---
	if len(cycles) > 0 {
		fmt.Fprintf(&sb, "CIRCULAR REFERENCES (%d nodes involved)\n", len(cycles))
		sb.WriteString(strings.Repeat("-", 40) + "\n")
		for _, name := range cycles {
			fmt.Fprintf(&sb, "  %s\n", name)
		}
		sb.WriteString("\n")
	}

	// --- Stale / missing artifacts ---
	if len(staleEntries) > 0 {
		fmt.Fprintf(&sb, "STALE / MISSING ARTIFACTS (%d)\n", len(staleEntries))
		sb.WriteString(strings.Repeat("-", 40) + "\n")
		for _, se := range staleEntries {
			fmt.Fprintf(&sb, "  [%s] %s  output: %s  path: %s\n",
				se.status, se.logicalName, se.outputID, se.outputPath)
			if se.detail != "" {
				fmt.Fprintf(&sb, "    %s\n", se.detail)
			}
		}
		sb.WriteString("\n")
	}

	// --- Summary line ---
	if len(formatErrors) == 0 && len(cycles) == 0 && len(staleEntries) == 0 {
		sb.WriteString("All spec nodes are valid and all artifacts are up to date.\n")
	} else {
		var parts []string
		if n := len(formatErrors); n > 0 {
			parts = append(parts, fmt.Sprintf("%d format error(s)", n))
		}
		if n := len(cycles); n > 0 {
			parts = append(parts, fmt.Sprintf("%d node(s) in cycle(s)", n))
		}
		if n := len(staleEntries); n > 0 {
			parts = append(parts, fmt.Sprintf("%d stale/missing artifact(s)", n))
		}
		fmt.Fprintf(&sb, "Summary: %s.\n", strings.Join(parts, ", "))
	}

	return sb.String()
}

// RegisterTool registers the validate_specs tool on the given MCP server.
// Call this once during server startup before calling server.Run.
func RegisterTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "validate_specs",
		Description: "Validate the spec tree for format errors, circular references, and artifact staleness.",
	}, HandleValidateSpecs)
}
