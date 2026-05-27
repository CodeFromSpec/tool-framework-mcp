// code-from-spec: ROOT/golang/internal/tools/load_chain/code@yQhE6xcGfD_cOSvPktyyoIBUffI

// Package load_chain implements the MCP tool handler for the load_chain tool.
// The tool takes a logical name (e.g. ROOT/x/y) and returns the assembled
// spec chain as a single text response, including a content hash for staleness
// detection.
//
// The chain contains:
//  1. A 27-character base64url SHA-1 hash (chain_hash)
//  2. The concatenated spec context (ancestors, dependencies, external files,
//     reduced frontmatter, target public+agent sections)
//  3. The input artifact content (if the target declares one)
package load_chain

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/normalizename"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathvalidation"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LoadChainArgs holds the typed input parameters for the load_chain tool.
type LoadChainArgs struct {
	LogicalName string `json:"logical_name" jsonschema:"Logical name of the node to generate code for."`
}

// HandleLoadChain is the MCP tool handler for load_chain.
// It assembles the full spec chain for a target logical name and returns it
// as a single text response, prefixed with the chain hash.
//
// The returned Go error is reserved for catastrophic server failures.
// All expected error conditions use IsError: true on the result.
func HandleLoadChain(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args LoadChainArgs,
) (*mcp.CallToolResult, any, error) {

	// --- Phase 1: Validation ---

	// Step 1 — Validate the logical name: it must be a ROOT/ reference.
	// PathFromLogicalName handles ROOT/ references only; ARTIFACT/ references
	// and anything else return ("", false).
	if args.LogicalName == "" {
		return toolError("invalid logical name: logical_name is required and must be a ROOT/ reference"), nil, nil
	}
	targetPath, ok := logicalnames.PathFromLogicalName(args.LogicalName)
	if !ok {
		return toolError(fmt.Sprintf("invalid logical name: %q is not a recognized ROOT/ reference", args.LogicalName)), nil, nil
	}

	// Step 2 — Parse the target node's frontmatter.
	fm, err := frontmatter.ParseFrontmatter(targetPath)
	if err != nil {
		return toolError(fmt.Sprintf("unreadable file: cannot parse frontmatter for %q: %v", targetPath, err)), nil, nil
	}

	// Step 3 — Check that outputs is present and non-empty.
	if len(fm.Outputs) == 0 {
		return toolError(fmt.Sprintf("no outputs: target node %q has no outputs field in its frontmatter", args.LogicalName)), nil, nil
	}

	// Step 4 — Validate each output path.
	// We use the current working directory as the project root (the tool is
	// always executed from the project root per the framework spec).
	projectRoot, err := os.Getwd()
	if err != nil {
		return toolError(fmt.Sprintf("unreadable file: cannot determine project root: %v", err)), nil, nil
	}
	for _, output := range fm.Outputs {
		if err := pathvalidation.ValidatePath(output.Path, projectRoot); err != nil {
			return toolError(fmt.Sprintf("invalid output path: output %q in %q: %v", output.Path, args.LogicalName, err)), nil, nil
		}
	}

	// --- Phase 2: Assemble the Context Stream ---

	// contextParts accumulates all content for the final context string.
	var contextParts []string

	// --- Step 1: Ancestors ---
	// Collect ancestors from ROOT down to the target's direct parent.
	// We walk up from the target using ParentLogicalName, then reverse the
	// list so we iterate root-first.
	var ancestors []string
	current := args.LogicalName
	for {
		parent, ok := logicalnames.ParentLogicalName(current)
		if !ok {
			// No parent (we are at ROOT, or invalid) — stop.
			break
		}
		ancestors = append(ancestors, parent)
		current = parent
	}
	// Reverse: ancestors[0] is ROOT, ancestors[len-1] is the direct parent.
	for i, j := 0, len(ancestors)-1; i < j; i, j = i+1, j-1 {
		ancestors[i], ancestors[j] = ancestors[j], ancestors[i]
	}

	for _, ancestorName := range ancestors {
		parsed, err := parsenode.ParseNode(ancestorName)
		if err != nil {
			return toolError(fmt.Sprintf("unreadable file: cannot parse ancestor %q: %v", ancestorName, err)), nil, nil
		}
		full := sectionFullContent(parsed.Public)
		if full == "" {
			continue
		}
		contextParts = append(contextParts, full)
	}

	// --- Step 2: Dependencies (depends_on) ---
	// Sort alphabetically by logical name string.
	sortedDeps := make([]string, len(fm.DependsOn))
	copy(sortedDeps, fm.DependsOn)
	sort.Strings(sortedDeps)

	for _, dep := range sortedDeps {
		switch {
		case logicalnames.IsArtifactRef(dep):
			// Case C: ARTIFACT/ reference
			nodePath, artifactID, ok := logicalnames.ArtifactRefParts(dep)
			if !ok {
				return toolError(fmt.Sprintf("chain resolution failure: cannot resolve ARTIFACT/ reference %q in depends_on of %q", dep, args.LogicalName)), nil, nil
			}
			depFM, err := frontmatter.ParseFrontmatter(nodePath)
			if err != nil {
				return toolError(fmt.Sprintf("chain resolution failure: cannot parse frontmatter for dependency %q: %v", dep, err)), nil, nil
			}
			// Find the output whose ID matches artifactID.
			var artifactPath string
			for _, out := range depFM.Outputs {
				if out.ID == artifactID {
					artifactPath = out.Path
					break
				}
			}
			if artifactPath == "" {
				return toolError(fmt.Sprintf("chain resolution failure: artifact %q not found in outputs of %q", artifactID, nodePath)), nil, nil
			}
			raw, err := os.ReadFile(artifactPath)
			if err != nil {
				return toolError(fmt.Sprintf("unreadable file: cannot read artifact %q for dependency %q: %v", artifactPath, dep, err)), nil, nil
			}
			stripped := stripFrontmatter(string(raw))
			contextParts = append(contextParts, stripped)

		default:
			// Case A or B: ROOT/ reference (with or without qualifier).
			hasQ, _ := logicalnames.HasQualifier(dep)
			parsed, err := parsenode.ParseNode(dep)
			if err != nil {
				return toolError(fmt.Sprintf("chain resolution failure: cannot parse dependency %q: %v", dep, err)), nil, nil
			}

			if !hasQ {
				// Case A: no qualifier — use full # Public section.
				if parsed.Public == nil {
					continue
				}
				var depContent strings.Builder
				if strings.TrimSpace(parsed.Public.Content) != "" {
					depContent.WriteString(parsed.Public.Content)
				}
				for _, sub := range parsed.Public.Subsections {
					if depContent.Len() > 0 {
						depContent.WriteString("\n")
					}
					depContent.WriteString("## " + sub.Heading + "\n\n")
					depContent.WriteString(sub.Content)
				}
				if depContent.Len() == 0 {
					continue
				}
				contextParts = append(contextParts, depContent.String())
			} else {
				// Case B: qualifier — use the matching subsection of # Public.
				qualifier, _ := logicalnames.QualifierName(dep)
				normalizedQ := normalizename.NormalizeName(qualifier)
				if parsed.Public == nil {
					return toolError(fmt.Sprintf("chain resolution failure: dependency %q has no # Public section", dep)), nil, nil
				}
				var matched *parsenode.Subsection
				for i := range parsed.Public.Subsections {
					sub := &parsed.Public.Subsections[i]
					if normalizename.NormalizeName(sub.Heading) == normalizedQ {
						matched = sub
						break
					}
				}
				if matched == nil {
					return toolError(fmt.Sprintf("chain resolution failure: subsection %q not found in # Public of dependency %q", qualifier, dep)), nil, nil
				}
				// Append content WITHOUT the heading.
				contextParts = append(contextParts, matched.Content)
			}
		}
	}

	// --- Step 3: External files ---
	// Sort alphabetically by path.
	sortedExt := make([]frontmatter.External, len(fm.External))
	copy(sortedExt, fm.External)
	sort.Slice(sortedExt, func(i, j int) bool {
		return sortedExt[i].Path < sortedExt[j].Path
	})

	for _, ext := range sortedExt {
		if len(ext.Fragments) == 0 {
			// Case A: no fragments — read and include the entire file.
			raw, err := os.ReadFile(ext.Path)
			if err != nil {
				return toolError(fmt.Sprintf("unreadable file: cannot read external file %q: %v", ext.Path, err)), nil, nil
			}
			content := string(raw)
			contextParts = append(contextParts, content)
		} else {
			// Case B: specific fragments declared.
			fr, err := filereader.OpenFileReader(ext.Path)
			if err != nil {
				return toolError(fmt.Sprintf("unreadable file: cannot open external file %q: %v", ext.Path, err)), nil, nil
			}
			var fragmentContent strings.Builder
			currentLine := 0
			for _, frag := range ext.Fragments {
				start, end, err := parseLineRange(frag.Lines)
				if err != nil {
					return toolError(fmt.Sprintf("unreadable file: invalid line range %q in external file %q: %v", frag.Lines, ext.Path, err)), nil, nil
				}
				skip := start - 1 - currentLine
				if skip > 0 {
					fr.SkipLines(skip)
					currentLine += skip
				}
				var fragLines []string
				for i := start; i <= end; i++ {
					line, readErr := fr.ReadLine()
					if readErr != nil {
						return toolError(fmt.Sprintf("unreadable file: cannot read line %d from %q: %v", i, ext.Path, readErr)), nil, nil
					}
					fragLines = append(fragLines, line)
					currentLine++
				}
				fragmentContent.WriteString(strings.Join(fragLines, "\n"))
			}
			content := fragmentContent.String()
			contextParts = append(contextParts, content)
		}
	}

	// --- Step 4: Target's reduced frontmatter and # Public section ---
	targetParsed, err := parsenode.ParseNode(args.LogicalName)
	if err != nil {
		return toolError(fmt.Sprintf("unreadable file: cannot parse target node %q: %v", args.LogicalName, err)), nil, nil
	}

	// Build the reduced frontmatter YAML block (only the outputs field).
	// This appears in the context stream but does NOT contribute to the hash.
	reducedFM := buildReducedFrontmatter(fm.Outputs)
	contextParts = append(contextParts, reducedFM)
	// (NOT added to hashInputs per spec)

	if pubFull := sectionFullContent(targetParsed.Public); pubFull != "" {
		contextParts = append(contextParts, pubFull)
	}

	// --- Step 5: Target's # Agent section ---
	if agentFull := sectionFullContent(targetParsed.Agent); agentFull != "" {
		contextParts = append(contextParts, agentFull)
	}

	// --- Phase 3: Input artifact (if declared) ---
	var inputContent string
	if fm.Input != "" {
		// Resolve the ARTIFACT/ reference declared in the input field.
		nodePath, artifactID, ok := logicalnames.ArtifactRefParts(fm.Input)
		if !ok {
			return toolError(fmt.Sprintf("chain resolution failure: cannot resolve input reference %q for %q", fm.Input, args.LogicalName)), nil, nil
		}
		inputFM, err := frontmatter.ParseFrontmatter(nodePath)
		if err != nil {
			return toolError(fmt.Sprintf("chain resolution failure: cannot parse frontmatter for input node %q: %v", nodePath, err)), nil, nil
		}
		var inputArtifactPath string
		for _, out := range inputFM.Outputs {
			if out.ID == artifactID {
				inputArtifactPath = out.Path
				break
			}
		}
		if inputArtifactPath == "" {
			return toolError(fmt.Sprintf("chain resolution failure: artifact %q not found in outputs of input node %q", artifactID, nodePath)), nil, nil
		}
		raw, err := os.ReadFile(inputArtifactPath)
		if err != nil {
			return toolError(fmt.Sprintf("unreadable file: cannot read input artifact %q: %v", inputArtifactPath, err)), nil, nil
		}
		inputContent = stripFrontmatter(string(raw))
	}

	// --- Phase 4: Compute the chain hash ---
	chainHash, err := chainhash.ComputeChainHash(args.LogicalName)
	if err != nil {
		return toolError(fmt.Sprintf("chain hash computation failed for %q: %v", args.LogicalName, err)), nil, nil
	}

	// --- Phase 5: Assemble and return the result ---
	// Join all context parts (no separator — the spec says plain concatenation).
	contextStr := strings.Join(contextParts, "")

	// Build the final response text:
	//   Line 1: "chain_hash: <hash>"
	//   Blank line
	//   Context content
	//   (blank line + "--- input ---" + blank line + input content, if present)
	var sb strings.Builder
	sb.WriteString("chain_hash: ")
	sb.WriteString(chainHash)
	sb.WriteString("\n\n")
	sb.WriteString(contextStr)
	if inputContent != "" {
		sb.WriteString("\n--- input ---\n")
		sb.WriteString(inputContent)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
	}, nil, nil
}

// --- Helpers ---

// toolError constructs a tool-error result (IsError: true) with an actionable
// message. Using this instead of returning a Go error keeps the server alive.
func toolError(message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: message}},
		IsError: true,
	}
}

// stripFrontmatter removes the leading frontmatter block (delimited by "---"
// lines) from content, if present. The content between the opening and closing
// "---" delimiters is discarded, along with both delimiter lines.
//
// If no frontmatter is present the original content is returned unchanged.
func stripFrontmatter(content string) string {
	// Frontmatter must start at the very beginning of the file.
	if !strings.HasPrefix(content, "---") {
		return content
	}
	// Find the end of the first line (the opening "---").
	firstNewline := strings.Index(content, "\n")
	if firstNewline == -1 {
		// Single line consisting only of "---" — not valid frontmatter.
		return content
	}
	// The first line must be exactly "---" (possibly with trailing CR).
	firstLine := strings.TrimRight(content[:firstNewline], "\r")
	if firstLine != "---" {
		return content
	}
	// Find the closing "---" delimiter.
	rest := content[firstNewline+1:]
	closingIdx := -1
	searchIn := rest
	offset := 0
	for {
		idx := strings.Index(searchIn, "---")
		if idx == -1 {
			break
		}
		// The "---" must appear at the start of a line.
		if idx == 0 || searchIn[idx-1] == '\n' {
			// Verify the line is exactly "---".
			lineEnd := strings.Index(searchIn[idx:], "\n")
			var line string
			if lineEnd == -1 {
				line = strings.TrimRight(searchIn[idx:], "\r")
			} else {
				line = strings.TrimRight(searchIn[idx:idx+lineEnd], "\r")
			}
			if line == "---" {
				closingIdx = offset + idx
				break
			}
		}
		// Advance past this occurrence.
		advance := idx + 3
		offset += advance
		searchIn = searchIn[advance:]
	}
	if closingIdx == -1 {
		// No closing delimiter — not valid frontmatter, return as-is.
		return content
	}
	// Skip past the closing "---" line (include the trailing newline if any).
	afterClosing := rest[closingIdx+3:]
	if strings.HasPrefix(afterClosing, "\r\n") {
		afterClosing = afterClosing[2:]
	} else if strings.HasPrefix(afterClosing, "\n") {
		afterClosing = afterClosing[1:]
	}
	return afterClosing
}

func sectionFullContent(s *parsenode.Section) string {
	if s == nil {
		return ""
	}
	var sb strings.Builder
	if strings.TrimSpace(s.Content) != "" {
		sb.WriteString(s.Content)
	}
	for _, sub := range s.Subsections {
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString("## ")
		sb.WriteString(sub.Heading)
		sb.WriteString("\n\n")
		sb.WriteString(sub.Content)
	}
	return sb.String()
}

// buildReducedFrontmatter constructs the YAML frontmatter block containing
// only the outputs field, wrapped in "---" delimiters. This is appended to
// the context stream so the agent can see what artifacts the target declares,
// but it does NOT contribute to the chain hash.
func buildReducedFrontmatter(outputs []frontmatter.Output) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString("outputs:\n")
	for _, out := range outputs {
		sb.WriteString(fmt.Sprintf("  - id: %s\n    path: %s\n", out.ID, out.Path))
	}
	sb.WriteString("---\n")
	return sb.String()
}

// parseLineRange parses a "start-end" line range string and returns the
// 1-based start and end line numbers.
func parseLineRange(lines string) (start, end int, err error) {
	parts := strings.SplitN(lines, "-", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected format start-end, got %q", lines)
	}
	_, err = fmt.Sscanf(parts[0], "%d", &start)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start line in %q: %w", lines, err)
	}
	_, err = fmt.Sscanf(parts[1], "%d", &end)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end line in %q: %w", lines, err)
	}
	if start < 1 || end < start {
		return 0, 0, fmt.Errorf("invalid range %q: start must be >= 1 and end must be >= start", lines)
	}
	return start, end, nil
}

// RegisterTool registers the load_chain MCP tool on the given server.
// Call this from the main server setup.
func RegisterTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "load_chain",
		Description: "Load the spec chain context for a given logical name. Returns all relevant spec files concatenated in a single response.",
		// Increase max result size to accommodate large spec chains.
		Meta: mcp.Meta{"anthropic/maxResultSizeChars": 500000},
	}, HandleLoadChain)
}

// Ensure the chainresolver package is referenced — the chain is resolved via
// chainresolver.ResolveChain which is used indirectly through parsenode,
// frontmatter, and logicalnames above. The import below keeps the dependency
// explicit and visible.
var _ = chainresolver.ResolveChain
