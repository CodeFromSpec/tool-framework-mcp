// code-from-spec: ROOT/golang/implementation/mcp_tools/load_chain@YtDW0V9n9v2FWAcddObFCo_1TWw

package mcploadchain

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
)

// ErrNoOutputs is returned when the target node has no outputs field.
var ErrNoOutputs = errors.New("no outputs")

// ErrInvalidOutputPath is returned when an output path fails path validation.
var ErrInvalidOutputPath = errors.New("invalid output path")

// MCPLoadChainResult holds the result returned by MCPLoadChain.
type MCPLoadChainResult struct {
	// ChainHash is the 27-character base64url chain hash for the target node.
	ChainHash string

	// Context is all chain content concatenated as a single stream.
	Context string

	// Input is the content of the input artifact (excluding frontmatter),
	// present only when the target node declares an input field.
	Input *string
}

// MCPLoadChain builds the full chain context for the given logical name and
// returns a result containing the chain hash, concatenated context, and
// optional input content.
func MCPLoadChain(logical_name string) (*MCPLoadChainResult, error) {
	// Step 1 — Validate and resolve

	// 1. Convert logical name to file path.
	targetFilePath, err := logicalnames.LogicalNameToPath(logical_name)
	if err != nil {
		return nil, fmt.Errorf("LogicalNameToPath: %w", err)
	}

	// 2. Parse frontmatter and check outputs.
	fm, err := frontmatter.FrontmatterParse(targetFilePath)
	if err != nil {
		return nil, fmt.Errorf("FrontmatterParse: %w", err)
	}
	if len(fm.Outputs) == 0 {
		return nil, fmt.Errorf("%w", ErrNoOutputs)
	}

	// 3. Validate each output path.
	for _, out := range fm.Outputs {
		if err := pathutils.PathValidateCfs(out.Path); err != nil {
			return nil, fmt.Errorf("%w: %s: %v", ErrInvalidOutputPath, out.Path, err)
		}
	}

	// 4. Resolve the full chain.
	chain, err := chainresolver.ChainResolve(logical_name)
	if err != nil {
		return nil, fmt.Errorf("ChainResolve: %w", err)
	}

	// Step 2 — Compute chain hash

	// 5. Compute the chain hash.
	chainHash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		return nil, fmt.Errorf("ChainHashCompute: %w", err)
	}

	// Step 3 — Build context stream

	var contextBuilder strings.Builder

	// Ancestors
	for _, ancestor := range chain.Ancestors {
		node, err := parsenode.NodeParse(ancestor.LogicalName)
		if err != nil {
			return nil, fmt.Errorf("NodeParse ancestor %s: %w", ancestor.LogicalName, err)
		}
		if node.Public == nil {
			continue
		}
		if len(node.Public.Content) == 0 && len(node.Public.Subsections) == 0 {
			continue
		}
		appendLines(&contextBuilder, node.Public.Content)
		for _, sub := range node.Public.Subsections {
			contextBuilder.WriteString(sub.RawHeading)
			contextBuilder.WriteString("\n")
			appendLines(&contextBuilder, sub.Content)
		}
	}

	// Dependencies
	for _, dep := range chain.Dependencies {
		if logicalnames.LogicalNameIsArtifact(dep.LogicalName) {
			// Read file, strip frontmatter, append remaining lines.
			content, err := readFileStripFrontmatter(dep.FilePath)
			if err != nil {
				return nil, fmt.Errorf("reading dependency artifact %s: %w", dep.LogicalName, err)
			}
			contextBuilder.WriteString(content)
		} else if dep.Qualifier == nil {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return nil, fmt.Errorf("NodeParse dependency %s: %w", dep.LogicalName, err)
			}
			if node.Public == nil {
				continue
			}
			appendLines(&contextBuilder, node.Public.Content)
			for _, sub := range node.Public.Subsections {
				contextBuilder.WriteString(sub.RawHeading)
				contextBuilder.WriteString("\n")
				appendLines(&contextBuilder, sub.Content)
			}
		} else {
			// Qualifier present — emit only the matching subsection.
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return nil, fmt.Errorf("NodeParse dependency %s: %w", dep.LogicalName, err)
			}
			if node.Public == nil {
				continue
			}
			normalizedQualifier := textnormalization.NormalizeText(*dep.Qualifier)
			for _, sub := range node.Public.Subsections {
				if sub.Heading == normalizedQualifier {
					appendLines(&contextBuilder, sub.Content)
					break
				}
			}
		}
	}

	// External
	for _, ext := range chain.External {
		extPath := &pathutils.PathCfs{Value: ext.Path}
		if len(ext.Fragments) == 0 {
			reader, err := filereader.FileOpen(extPath)
			if err != nil {
				return nil, fmt.Errorf("FileOpen external %s: %w", ext.Path, err)
			}
			if err := appendAllLines(&contextBuilder, reader); err != nil {
				filereader.FileClose(reader)
				return nil, fmt.Errorf("reading external %s: %w", ext.Path, err)
			}
			filereader.FileClose(reader)
		} else {
			for _, frag := range ext.Fragments {
				start, end, err := parseLineRange(frag.Lines)
				if err != nil {
					return nil, fmt.Errorf("parsing fragment lines %q for %s: %w", frag.Lines, ext.Path, err)
				}
				reader, err := filereader.FileOpen(extPath)
				if err != nil {
					return nil, fmt.Errorf("FileOpen external fragment %s: %w", ext.Path, err)
				}
				filereader.FileSkipLines(reader, start-1)
				count := end - start + 1
				for i := 0; i < count; i++ {
					line, err := filereader.FileReadLine(reader)
					if errors.Is(err, filereader.ErrEndOfFile) {
						break
					}
					if err != nil {
						filereader.FileClose(reader)
						return nil, fmt.Errorf("reading external fragment %s line %d: %w", ext.Path, start+i, err)
					}
					contextBuilder.WriteString(line)
					contextBuilder.WriteString("\n")
				}
				filereader.FileClose(reader)
			}
		}
	}

	// Target Public and Frontmatter — emit reduced frontmatter block
	contextBuilder.WriteString("---\n")
	contextBuilder.WriteString("outputs:\n")
	for _, out := range fm.Outputs {
		contextBuilder.WriteString(fmt.Sprintf("  - id: %s\n", out.ID))
		contextBuilder.WriteString(fmt.Sprintf("    path: %s\n", out.Path))
	}
	contextBuilder.WriteString("---\n")

	// Target public section
	targetNode, err := parsenode.NodeParse(chain.Target.LogicalName)
	if err != nil {
		return nil, fmt.Errorf("NodeParse target %s: %w", chain.Target.LogicalName, err)
	}
	if targetNode.Public != nil {
		if len(targetNode.Public.Content) > 0 || len(targetNode.Public.Subsections) > 0 {
			appendLines(&contextBuilder, targetNode.Public.Content)
			for _, sub := range targetNode.Public.Subsections {
				contextBuilder.WriteString(sub.RawHeading)
				contextBuilder.WriteString("\n")
				appendLines(&contextBuilder, sub.Content)
			}
		}
	}

	// Target Agent
	if targetNode.Agent != nil {
		appendLines(&contextBuilder, targetNode.Agent.Content)
		for _, sub := range targetNode.Agent.Subsections {
			contextBuilder.WriteString(sub.RawHeading)
			contextBuilder.WriteString("\n")
			appendLines(&contextBuilder, sub.Content)
		}
	}

	// Step 4 — Extract input

	var inputContent *string
	if chain.Input != nil {
		content, err := readFileStripFrontmatter(chain.Input.FilePath)
		if err != nil {
			return nil, fmt.Errorf("reading input artifact: %w", err)
		}
		inputContent = &content
	}

	// Step 5 — Return result

	context := contextBuilder.String()
	return &MCPLoadChainResult{
		ChainHash: chainHash,
		Context:   context,
		Input:     inputContent,
	}, nil
}

// appendLines appends each line to the builder followed by "\n".
func appendLines(b *strings.Builder, lines []string) {
	for _, line := range lines {
		b.WriteString(line)
		b.WriteString("\n")
	}
}

// appendAllLines reads all lines from the reader and appends them to the builder.
// Returns an error for unexpected read errors (ErrEndOfFile is treated as normal termination).
func appendAllLines(b *strings.Builder, reader *filereader.FileReader) error {
	for {
		line, err := filereader.FileReadLine(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			break
		}
		if err != nil {
			return fmt.Errorf("FileReadLine: %w", err)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	return nil
}

// readFileStripFrontmatter opens the file at cfsPath, strips the leading
// YAML frontmatter (if present), and returns the remaining content as a string.
// Each line is followed by "\n".
func readFileStripFrontmatter(cfsPath *pathutils.PathCfs) (string, error) {
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		return "", fmt.Errorf("FileOpen %s: %w", cfsPath.Value, err)
	}
	defer filereader.FileClose(reader)

	var builder strings.Builder

	// Find the first non-blank line.
	// If it is "---", skip until the closing "---", then read the rest.
	// Otherwise, include it and read the rest normally.
	inFrontmatter := false
	frontmatterDone := false
	foundFirstNonBlank := false

	for {
		line, err := filereader.FileReadLine(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("FileReadLine %s: %w", cfsPath.Value, err)
		}

		if !foundFirstNonBlank {
			if strings.TrimSpace(line) == "" {
				// Blank line before any content — skip it while looking for first non-blank.
				continue
			}
			foundFirstNonBlank = true
			if line == "---" {
				inFrontmatter = true
				continue
			}
			// Not frontmatter, emit this line.
			builder.WriteString(line)
			builder.WriteString("\n")
			continue
		}

		if inFrontmatter && !frontmatterDone {
			if line == "---" {
				frontmatterDone = true
				inFrontmatter = false
			}
			// Discard frontmatter lines.
			continue
		}

		builder.WriteString(line)
		builder.WriteString("\n")
	}

	return builder.String(), nil
}

// parseLineRange parses a "start-end" fragment lines string.
func parseLineRange(lines string) (start, end int, err error) {
	_, err = fmt.Sscanf(lines, "%d-%d", &start, &end)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid line range %q: %w", lines, err)
	}
	return start, end, nil
}
