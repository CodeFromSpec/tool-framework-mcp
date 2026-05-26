// code-from-spec: ROOT/golang/internal/tools/load_chain/code@PENDING
package load_chain

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/normalizename"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathvalidation"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LoadChainArgs defines the input parameters for the load_chain tool.
type LoadChainArgs struct {
	LogicalName string `json:"logical_name" jsonschema:"Logical name of the node to generate code for."`
}

// HandleLoadChain validates the logical name, loads the spec chain,
// and returns the chain hash, context stream, and input as separate
// text content items.
func HandleLoadChain(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args LoadChainArgs,
) (*mcp.CallToolResult, any, error) {
	// Step 1: Validate logical name starts with ROOT/.
	if !strings.HasPrefix(args.LogicalName, "ROOT/") && args.LogicalName != "ROOT" {
		return toolError("invalid logical name"), nil, nil
	}

	// Step 2: Resolve logical name to file path and parse frontmatter.
	nodePath, ok := logicalnames.PathFromLogicalName(args.LogicalName)
	if !ok {
		return toolError("invalid logical name"), nil, nil
	}

	fm, err := frontmatter.ParseFrontmatter(nodePath)
	if err != nil {
		return toolError(fmt.Sprintf("unreadable file: %v", err)), nil, nil
	}

	// Step 3: Check outputs not empty.
	if len(fm.Outputs) == 0 {
		return toolError("no outputs"), nil, nil
	}

	// Step 4: Validate output paths.
	for _, out := range fm.Outputs {
		if err := pathvalidation.ValidatePath(out.Path, "."); err != nil {
			return toolError(fmt.Sprintf("invalid output path: %v", err)), nil, nil
		}
	}

	// Step 5: Resolve chain.
	chain, err := chainresolver.ResolveChain(args.LogicalName)
	if err != nil {
		return toolError(fmt.Sprintf("chain resolution failure: %v", err)), nil, nil
	}

	// Build context stream and hash parts.
	var hashParts [][]byte
	var contextBuf strings.Builder

	// Step 1 -- Ancestors (root to target's parent).
	for _, ancestor := range chain.Ancestors {
		parsed, err := parsenode.ParseNode(ancestor.LogicalName)
		if err != nil {
			return toolError(fmt.Sprintf("unreadable file: %v", err)), nil, nil
		}

		if parsed.Public == nil {
			continue
		}
		// Check if # Public has any content (body or subsections).
		hasContent := strings.TrimSpace(parsed.Public.Content) != ""
		if !hasContent {
			for _, sub := range parsed.Public.Subsections {
				if strings.TrimSpace(sub.Content) != "" {
					hasContent = true
					break
				}
			}
		}
		if !hasContent {
			continue
		}

		// Compute SHA-1 of the # Public section including the heading.
		publicWithHeading := sectionWithHeading(parsed.Public)
		hash := sha1.Sum([]byte(publicWithHeading))
		hashParts = append(hashParts, hash[:])

		// Append content without the heading.
		appendContent(&contextBuf, sectionContentWithoutHeading(parsed.Public))
	}

	// Step 2 -- Dependencies (depends_on), sorted alphabetically.
	for _, dep := range chain.Dependencies {
		if logicalnames.IsArtifactRef(dep.LogicalName) {
			// ARTIFACT reference: resolve to artifact file, read excluding frontmatter.
			content, err := readArtifactContent(dep.LogicalName)
			if err != nil {
				return toolError(fmt.Sprintf("unreadable file: %v", err)), nil, nil
			}
			hash := sha1.Sum([]byte(content))
			hashParts = append(hashParts, hash[:])
			appendContent(&contextBuf, content)
		} else if dep.Qualifier != nil {
			// Subsection reference: ROOT/x/y(z). Strip qualifier for parsing.
			baseLogicalName := stripQualifier(dep.LogicalName)
			parsed, err := parsenode.ParseNode(baseLogicalName)
			if err != nil {
				return toolError(fmt.Sprintf("unreadable file: %v", err)), nil, nil
			}

			subsection := findSubsection(parsed.Public, *dep.Qualifier)
			if subsection == nil {
				return toolError(fmt.Sprintf("chain resolution failure: subsection %q not found", *dep.Qualifier)), nil, nil
			}

			subWithHeading := "## " + subsection.Heading + "\n" + subsection.Content
			hash := sha1.Sum([]byte(subWithHeading))
			hashParts = append(hashParts, hash[:])
			appendContent(&contextBuf, subsection.Content)
		} else {
			// Plain node reference: ROOT/x/y.
			parsed, err := parsenode.ParseNode(dep.LogicalName)
			if err != nil {
				return toolError(fmt.Sprintf("unreadable file: %v", err)), nil, nil
			}

			if parsed.Public == nil {
				continue
			}

			publicWithHeading := sectionWithHeading(parsed.Public)
			hash := sha1.Sum([]byte(publicWithHeading))
			hashParts = append(hashParts, hash[:])
			appendContent(&contextBuf, sectionContentWithoutHeading(parsed.Public))
		}
	}

	// Step 3 -- External files, sorted alphabetically by path.
	externals := make([]frontmatter.External, len(fm.External))
	copy(externals, fm.External)
	sort.Slice(externals, func(i, j int) bool {
		return externals[i].Path < externals[j].Path
	})

	for _, ext := range externals {
		content, err := readExternalContent(ext)
		if err != nil {
			return toolError(fmt.Sprintf("unreadable file: %v", err)), nil, nil
		}
		hash := sha1.Sum([]byte(content))
		hashParts = append(hashParts, hash[:])
		appendContent(&contextBuf, content)
	}

	// Step 4 -- Target # Public.
	targetParsed, err := parsenode.ParseNode(chain.Target.LogicalName)
	if err != nil {
		return toolError(fmt.Sprintf("unreadable file: %v", err)), nil, nil
	}

	if targetParsed.Public != nil {
		publicWithHeading := sectionWithHeading(targetParsed.Public)
		hash := sha1.Sum([]byte(publicWithHeading))
		hashParts = append(hashParts, hash[:])

		// Build reduced frontmatter with only outputs.
		reducedFM := buildReducedFrontmatter(fm.Outputs)
		appendContent(&contextBuf, reducedFM+sectionContentWithoutHeading(targetParsed.Public))
	}

	// Step 5 -- Target # Agent.
	if targetParsed.Agent != nil {
		agentWithHeading := sectionWithHeading(targetParsed.Agent)
		hash := sha1.Sum([]byte(agentWithHeading))
		hashParts = append(hashParts, hash[:])
		appendContent(&contextBuf, sectionContentWithoutHeading(targetParsed.Agent))
	}

	// Step 6 -- Input separation.
	var inputContent string
	if fm.Input != "" {
		content, err := readArtifactContent(fm.Input)
		if err != nil {
			return toolError(fmt.Sprintf("unreadable file: %v", err)), nil, nil
		}
		hash := sha1.Sum([]byte(content))
		hashParts = append(hashParts, hash[:])
		inputContent = content
	}

	// Step 7 -- Chain hash computation.
	var concatenated []byte
	for _, h := range hashParts {
		concatenated = append(concatenated, h...)
	}
	finalHash := sha1.Sum(concatenated)
	chainHash := base64.RawURLEncoding.EncodeToString(finalHash[:])
	if len(chainHash) > 27 {
		chainHash = chainHash[:27]
	}

	// Step 8 -- Build result as a single text block.
	var result strings.Builder
	result.WriteString("chain_hash: ")
	result.WriteString(chainHash)
	result.WriteString("\n\n")
	result.WriteString(contextBuf.String())
	if inputContent != "" {
		result.WriteString("\n--- input ---\n")
		result.WriteString(inputContent)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
	}, nil, nil
}

// sectionWithHeading returns the full section content including its heading.
func sectionWithHeading(s *parsenode.Section) string {
	var buf strings.Builder
	buf.WriteString("# ")
	buf.WriteString(s.Heading)
	buf.WriteString("\n")
	if s.Content != "" {
		buf.WriteString(s.Content)
	}
	for _, sub := range s.Subsections {
		buf.WriteString("\n## ")
		buf.WriteString(sub.Heading)
		buf.WriteString("\n")
		if sub.Content != "" {
			buf.WriteString(sub.Content)
		}
	}
	return buf.String()
}

// sectionContentWithoutHeading returns the section content (including subsections)
// but without the top-level heading.
func sectionContentWithoutHeading(s *parsenode.Section) string {
	var buf strings.Builder
	if s.Content != "" {
		buf.WriteString(s.Content)
	}
	for _, sub := range s.Subsections {
		if buf.Len() > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString("## ")
		buf.WriteString(sub.Heading)
		buf.WriteString("\n")
		if sub.Content != "" {
			buf.WriteString(sub.Content)
		}
	}
	return buf.String()
}

// findSubsection finds a subsection within the Public section by normalized name.
func findSubsection(public *parsenode.Section, qualifier string) *parsenode.Subsection {
	if public == nil {
		return nil
	}
	normalizedQualifier := normalizename.NormalizeName(qualifier)
	for i := range public.Subsections {
		if normalizename.NormalizeName(public.Subsections[i].Heading) == normalizedQualifier {
			return &public.Subsections[i]
		}
	}
	return nil
}

// readArtifactContent resolves an ARTIFACT/ reference to its artifact file
// and reads the content excluding frontmatter. The reference format is
// ARTIFACT/x/y(id) where id identifies the output in the node's frontmatter.
func readArtifactContent(artifactRef string) (string, error) {
	nodePath, artifactID, ok := logicalnames.ArtifactRefParts(artifactRef)
	if !ok {
		return "", fmt.Errorf("cannot resolve artifact reference: %s", artifactRef)
	}

	// Read the node's frontmatter to find the output path for this artifact ID.
	fm, err := frontmatter.ParseFrontmatter(nodePath)
	if err != nil {
		return "", fmt.Errorf("cannot read node %s: %w", nodePath, err)
	}

	var artifactPath string
	for _, out := range fm.Outputs {
		if out.ID == artifactID {
			artifactPath = out.Path
			break
		}
	}
	if artifactPath == "" {
		return "", fmt.Errorf("artifact ID %q not found in outputs of %s", artifactID, nodePath)
	}

	return readFileExcludingFrontmatter(artifactPath)
}

// readFileExcludingFrontmatter reads a file and returns its content
// with YAML frontmatter stripped. CRLF is normalized to LF.
func readFileExcludingFrontmatter(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	return stripFrontmatter(content), nil
}

// stripFrontmatter removes YAML frontmatter delimited by --- lines.
func stripFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---") {
		return content
	}
	// Find end of first --- line.
	idx := strings.Index(content[3:], "\n")
	if idx < 0 {
		return content
	}
	rest := content[3+idx+1:]
	// Find closing ---.
	closingIdx := strings.Index(rest, "---")
	if closingIdx < 0 {
		return content
	}
	afterClosing := rest[closingIdx+3:]
	nlIdx := strings.Index(afterClosing, "\n")
	if nlIdx < 0 {
		return ""
	}
	return afterClosing[nlIdx+1:]
}

// readExternalContent reads the content of an external file entry,
// handling optional fragment extraction.
func readExternalContent(ext frontmatter.External) (string, error) {
	data, err := os.ReadFile(ext.Path)
	if err != nil {
		return "", err
	}
	content := strings.ReplaceAll(string(data), "\r\n", "\n")

	if len(ext.Fragments) == 0 {
		return content, nil
	}

	// Extract fragments.
	lines := strings.Split(content, "\n")
	var result strings.Builder
	for _, frag := range ext.Fragments {
		// Parse the Lines field as "start-end".
		parts := strings.SplitN(frag.Lines, "-", 2)
		if len(parts) != 2 {
			continue
		}
		start := 0
		end := 0
		fmt.Sscanf(parts[0], "%d", &start)
		fmt.Sscanf(parts[1], "%d", &end)
		if start < 1 || end < start || end > len(lines) {
			continue
		}
		extracted := lines[start-1 : end]
		if result.Len() > 0 {
			result.WriteString("\n")
		}
		result.WriteString(strings.Join(extracted, "\n"))
	}

	return result.String(), nil
}

// buildReducedFrontmatter builds a YAML frontmatter block containing only outputs.
func buildReducedFrontmatter(outputs []frontmatter.Output) string {
	var buf strings.Builder
	buf.WriteString("---\noutputs:\n")
	for _, out := range outputs {
		buf.WriteString("  - id: ")
		buf.WriteString(out.ID)
		buf.WriteString("\n    path: ")
		buf.WriteString(filepath.ToSlash(out.Path))
		buf.WriteString("\n")
	}
	buf.WriteString("---\n\n")
	return buf.String()
}

// appendContent appends content to the builder, adding a blank line separator
// if the builder already has content.
func appendContent(buf *strings.Builder, content string) {
	if buf.Len() > 0 && content != "" {
		buf.WriteString("\n")
	}
	buf.WriteString(content)
}

// stripQualifier removes the parenthetical qualifier from a logical name.
func stripQualifier(logicalName string) string {
	idx := strings.Index(logicalName, "(")
	if idx < 0 {
		return logicalName
	}
	return logicalName[:idx]
}

// toolError returns a CallToolResult with IsError set to true.
func toolError(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		IsError: true,
	}
}
