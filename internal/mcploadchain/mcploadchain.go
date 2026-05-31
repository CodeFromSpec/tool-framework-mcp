// code-from-spec: ROOT/golang/implementation/mcp_tools/load_chain@X7lROk45mSOZE6YdwCFrtDFJfjg

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

// MCPLoadChainResult holds the output of a successful MCPLoadChain call.
type MCPLoadChainResult struct {
	// ChainHash is the 27-character base64url chain hash.
	ChainHash string

	// Context contains all chain content concatenated as a single stream.
	Context string

	// Input contains the content of the input artifact, excluding frontmatter.
	// It is nil when the target node has no input field.
	Input *string
}

// MCPLoadChain resolves and assembles the full spec chain for the given
// logical name and returns the chain hash, concatenated context, and
// optional input content.
//
// The function resolves the chain for the target node, computes its hash,
// and concatenates all chain positions into a single context string. If
// the target node declares an input artifact, its content (excluding
// frontmatter) is returned in the Input field of the result.
//
// Errors:
//   - ErrNoOutputs: the target node has no outputs field.
//   - ErrInvalidOutputPath: an output path fails path validation.
//   - (LogicalNames.*): propagated from LogicalNameToPath.
//   - (ChainResolver.*): propagated from ChainResolve.
//   - (ChainHash.*): propagated from ChainHashCompute.
//   - (NodeParsing.*): propagated from NodeParse.
//   - (FileReader.*): propagated from FileOpen.
func MCPLoadChain(logical_name string) (*MCPLoadChainResult, error) {
	// Step 1 — Validate and resolve

	// 1. Resolve the target node's file path.
	filePath, err := logicalnames.LogicalNameToPath(logical_name)
	if err != nil {
		return nil, fmt.Errorf("LogicalNameToPath: %w", err)
	}

	// 2. Parse frontmatter and verify outputs exist.
	fm, err := frontmatter.FrontmatterParse(filePath)
	if err != nil {
		return nil, fmt.Errorf("FrontmatterParse: %w", err)
	}
	if len(fm.Outputs) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrNoOutputs, logical_name)
	}

	// 3. Validate each output path.
	for _, output := range fm.Outputs {
		if err := pathutils.PathValidateCfs(output.Path); err != nil {
			return nil, fmt.Errorf("%w: %s: %w", ErrInvalidOutputPath, output.Path, err)
		}
	}

	// 4. Resolve the chain.
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

	// 6. Initialize context.
	var contextBuilder strings.Builder

	// 7. Ancestors — emit public sections.
	for _, ancestor := range chain.Ancestors {
		node, err := parsenode.NodeParse(ancestor.LogicalName)
		if err != nil {
			return nil, fmt.Errorf("NodeParse(%s): %w", ancestor.LogicalName, err)
		}

		if node.Public == nil {
			continue
		}
		if len(node.Public.Content) == 0 && len(node.Public.Subsections) == 0 {
			continue
		}

		contextBuilder.WriteString(node.Public.RawHeading)
		contextBuilder.WriteString("\n")
		for _, line := range node.Public.Content {
			contextBuilder.WriteString(line)
			contextBuilder.WriteString("\n")
		}
		for _, sub := range node.Public.Subsections {
			contextBuilder.WriteString(sub.RawHeading)
			contextBuilder.WriteString("\n")
			for _, line := range sub.Content {
				contextBuilder.WriteString(line)
				contextBuilder.WriteString("\n")
			}
		}
	}

	// 8. Dependencies — emit dependency content.
	for _, dep := range chain.Dependencies {
		if logicalnames.LogicalNameIsArtifact(dep.LogicalName) {
			// 8a. Artifact dependency — read file content, strip frontmatter.
			reader, err := filereader.FileOpen(dep.FilePath)
			if err != nil {
				return nil, fmt.Errorf("FileOpen(%s): %w", dep.FilePath.Value, err)
			}
			content, err := readAllStrippingFrontmatter(reader)
			filereader.FileClose(reader)
			if err != nil {
				return nil, fmt.Errorf("reading artifact %s: %w", dep.FilePath.Value, err)
			}
			contextBuilder.WriteString(content)
		} else if dep.Qualifier == nil {
			// 8b. Node dependency without qualifier — emit full public section.
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return nil, fmt.Errorf("NodeParse(%s): %w", dep.LogicalName, err)
			}

			if node.Public != nil {
				contextBuilder.WriteString(node.Public.RawHeading)
				contextBuilder.WriteString("\n")
				for _, line := range node.Public.Content {
					contextBuilder.WriteString(line)
					contextBuilder.WriteString("\n")
				}
				for _, sub := range node.Public.Subsections {
					contextBuilder.WriteString(sub.RawHeading)
					contextBuilder.WriteString("\n")
					for _, line := range sub.Content {
						contextBuilder.WriteString(line)
						contextBuilder.WriteString("\n")
					}
				}
			}
		} else {
			// 8c. Node dependency with qualifier — emit matching subsection only.
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return nil, fmt.Errorf("NodeParse(%s): %w", dep.LogicalName, err)
			}

			normalizedQualifier := textnormalization.NormalizeText(*dep.Qualifier)
			if node.Public != nil {
				for _, sub := range node.Public.Subsections {
					if sub.Heading == normalizedQualifier {
						contextBuilder.WriteString(sub.RawHeading)
						contextBuilder.WriteString("\n")
						for _, line := range sub.Content {
							contextBuilder.WriteString(line)
							contextBuilder.WriteString("\n")
						}
						break
					}
				}
			}
		}
	}

	// 9. External — emit external file content.
	for _, ext := range chain.External {
		extPath := &pathutils.PathCfs{Value: ext.Path}

		if len(ext.Fragments) == 0 {
			// 9b. No fragments — emit entire file.
			reader, err := filereader.FileOpen(extPath)
			if err != nil {
				return nil, fmt.Errorf("FileOpen(%s): %w", ext.Path, err)
			}
			for {
				line, err := filereader.FileReadLine(reader)
				if errors.Is(err, filereader.ErrEndOfFile) {
					break
				}
				if err != nil {
					filereader.FileClose(reader)
					return nil, fmt.Errorf("FileReadLine(%s): %w", ext.Path, err)
				}
				contextBuilder.WriteString(line)
				contextBuilder.WriteString("\n")
			}
			filereader.FileClose(reader)
		} else {
			// 9c. Fragments present — emit each fragment's line range.
			for _, frag := range ext.Fragments {
				start, end, err := parseLineRange(frag.Lines)
				if err != nil {
					return nil, fmt.Errorf("parseLineRange(%q): %w", frag.Lines, err)
				}

				reader, err := filereader.FileOpen(extPath)
				if err != nil {
					return nil, fmt.Errorf("FileOpen(%s): %w", ext.Path, err)
				}

				filereader.FileSkipLines(reader, start-1)

				linesToRead := end - start + 1
				for i := 0; i < linesToRead; i++ {
					line, err := filereader.FileReadLine(reader)
					if errors.Is(err, filereader.ErrEndOfFile) {
						break
					}
					if err != nil {
						filereader.FileClose(reader)
						return nil, fmt.Errorf("FileReadLine(%s): %w", ext.Path, err)
					}
					contextBuilder.WriteString(line)
					contextBuilder.WriteString("\n")
				}
				filereader.FileClose(reader)
			}
		}
	}

	// 10. Target Public — emit reduced frontmatter block and public section.

	// 10a. Emit reduced frontmatter block.
	contextBuilder.WriteString("---\n")
	contextBuilder.WriteString("outputs:\n")
	for _, output := range fm.Outputs {
		contextBuilder.WriteString("  - id: ")
		contextBuilder.WriteString(output.ID)
		contextBuilder.WriteString("\n")
		contextBuilder.WriteString("    path: ")
		contextBuilder.WriteString(output.Path)
		contextBuilder.WriteString("\n")
	}
	contextBuilder.WriteString("---\n")

	// 10b. Parse the target node.
	targetNode, err := parsenode.NodeParse(chain.Target.LogicalName)
	if err != nil {
		return nil, fmt.Errorf("NodeParse(%s): %w", chain.Target.LogicalName, err)
	}

	// 10c. Emit public section if present.
	if targetNode.Public != nil {
		contextBuilder.WriteString(targetNode.Public.RawHeading)
		contextBuilder.WriteString("\n")
		for _, line := range targetNode.Public.Content {
			contextBuilder.WriteString(line)
			contextBuilder.WriteString("\n")
		}
		for _, sub := range targetNode.Public.Subsections {
			contextBuilder.WriteString(sub.RawHeading)
			contextBuilder.WriteString("\n")
			for _, line := range sub.Content {
				contextBuilder.WriteString(line)
				contextBuilder.WriteString("\n")
			}
		}
	}

	// 11. Target Agent — emit agent section if present.
	if targetNode.Agent != nil {
		contextBuilder.WriteString(targetNode.Agent.RawHeading)
		contextBuilder.WriteString("\n")
		for _, line := range targetNode.Agent.Content {
			contextBuilder.WriteString(line)
			contextBuilder.WriteString("\n")
		}
		for _, sub := range targetNode.Agent.Subsections {
			contextBuilder.WriteString(sub.RawHeading)
			contextBuilder.WriteString("\n")
			for _, line := range sub.Content {
				contextBuilder.WriteString(line)
				contextBuilder.WriteString("\n")
			}
		}
	}

	// Step 4 — Extract input

	// 12. If chain.input is present, read its content stripping frontmatter.
	var inputContent *string
	if chain.Input != nil {
		reader, err := filereader.FileOpen(chain.Input.FilePath)
		if err != nil {
			return nil, fmt.Errorf("FileOpen(%s): %w", chain.Input.FilePath.Value, err)
		}
		content, err := readAllStrippingFrontmatter(reader)
		filereader.FileClose(reader)
		if err != nil {
			return nil, fmt.Errorf("reading input %s: %w", chain.Input.FilePath.Value, err)
		}
		inputContent = &content
	}

	// Step 5 — Return result

	// 13. Build and return the result.
	result := &MCPLoadChainResult{
		ChainHash: chainHash,
		Context:   contextBuilder.String(),
		Input:     inputContent,
	}
	return result, nil
}

// readAllStrippingFrontmatter reads all lines from reader, stripping any
// leading frontmatter block delimited by "---" markers, and returns the
// remaining content as a single string with each line followed by "\n".
//
// The reader must already be open. The caller is responsible for closing it.
func readAllStrippingFrontmatter(reader *filereader.FileReader) (string, error) {
	var builder strings.Builder

	// Read the first line to check for frontmatter.
	firstLine, err := filereader.FileReadLine(reader)
	if errors.Is(err, filereader.ErrEndOfFile) {
		// Empty file.
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("FileReadLine: %w", err)
	}

	if firstLine == "---" {
		// Frontmatter present — discard lines until the closing "---".
		for {
			line, err := filereader.FileReadLine(reader)
			if errors.Is(err, filereader.ErrEndOfFile) {
				// No closing delimiter found — treat rest as content (already consumed).
				return builder.String(), nil
			}
			if err != nil {
				return "", fmt.Errorf("FileReadLine: %w", err)
			}
			if line == "---" {
				// Closing delimiter found — content starts after this.
				break
			}
		}
	} else {
		// No frontmatter — the first line is content.
		builder.WriteString(firstLine)
		builder.WriteString("\n")
	}

	// Read all remaining lines as content.
	for {
		line, err := filereader.FileReadLine(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("FileReadLine: %w", err)
		}
		builder.WriteString(line)
		builder.WriteString("\n")
	}

	return builder.String(), nil
}

// parseLineRange parses a "start-end" string into integer start and end values.
func parseLineRange(lines string) (start, end int, err error) {
	_, err = fmt.Sscanf(lines, "%d-%d", &start, &end)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid line range %q: %w", lines, err)
	}
	return start, end, nil
}
