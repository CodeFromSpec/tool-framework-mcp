// code-from-spec: ROOT/golang/implementation/mcp_tools/load_chain@hUvtcTNsCGCbnyLB7kzonfPApJQ
package mcploadchain

import (
	"errors"
	"fmt"

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

// MCPLoadChainResult holds the result of a successful load_chain tool call.
type MCPLoadChainResult struct {
	// ChainHash is the 27-character base64url-encoded SHA-1 chain hash.
	ChainHash string

	// Context is all chain content concatenated as a single stream.
	Context string

	// Input is the content of the input artifact, excluding frontmatter.
	// It is nil when the target node has no input artifact.
	Input *string
}

// MCPLoadChain resolves the spec chain for the given logical name,
// computes the chain hash, and returns the concatenated context and
// optional input artifact content.
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

	// 1. Convert logical name to file path.
	filePath, err := logicalnames.LogicalNameToPath(logical_name)
	if err != nil {
		return nil, fmt.Errorf("MCPLoadChain: resolving logical name: %w", err)
	}

	// 2. Parse frontmatter and check for outputs.
	fm, err := frontmatter.FrontmatterParse(filePath)
	if err != nil {
		return nil, fmt.Errorf("MCPLoadChain: parsing frontmatter: %w", err)
	}
	if len(fm.Outputs) == 0 {
		return nil, fmt.Errorf("MCPLoadChain: %w", ErrNoOutputs)
	}

	// 3. Validate each output path.
	for _, output := range fm.Outputs {
		if err := pathutils.PathValidateCfs(output.Path); err != nil {
			return nil, fmt.Errorf("MCPLoadChain: validating output path %q: %w", output.Path, ErrInvalidOutputPath)
		}
	}

	// 4. Resolve the chain.
	chain, err := chainresolver.ChainResolve(logical_name)
	if err != nil {
		return nil, fmt.Errorf("MCPLoadChain: resolving chain: %w", err)
	}

	// Step 2 — Compute chain hash

	// 5. Compute the chain hash.
	chainHash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		return nil, fmt.Errorf("MCPLoadChain: computing chain hash: %w", err)
	}

	// Step 3 — Build context stream

	// 6. Initialize context.
	var context string

	// 7. Ancestors.
	for _, ancestor := range chain.Ancestors {
		node, err := parsenode.NodeParse(ancestor.LogicalName)
		if err != nil {
			return nil, fmt.Errorf("MCPLoadChain: parsing ancestor %q: %w", ancestor.LogicalName, err)
		}

		// Skip if public is absent or has empty content and subsections.
		if node.Public == nil {
			continue
		}
		if len(node.Public.Content) == 0 && len(node.Public.Subsections) == 0 {
			continue
		}

		context += node.Public.RawHeading + "\n"
		for _, line := range node.Public.Content {
			context += line + "\n"
		}
		for _, sub := range node.Public.Subsections {
			context += sub.RawHeading + "\n"
			for _, line := range sub.Content {
				context += line + "\n"
			}
		}
	}

	// 8. Dependencies.
	for _, dep := range chain.Dependencies {
		if logicalnames.LogicalNameIsArtifact(dep.LogicalName) {
			// 8a. Artifact reference — open the file and strip frontmatter.
			reader, err := filereader.FileOpen(&dep.FilePath)
			if err != nil {
				return nil, fmt.Errorf("MCPLoadChain: opening artifact %q: %w", dep.LogicalName, err)
			}
			firstLine, hasFirstLine, err := stripFrontmatter(reader)
			if err != nil {
				filereader.FileClose(reader)
				return nil, fmt.Errorf("MCPLoadChain: stripping frontmatter from artifact %q: %w", dep.LogicalName, err)
			}
			if hasFirstLine {
				context += firstLine + "\n"
			}
			for {
				line, err := filereader.FileReadLine(reader)
				if err != nil {
					if errors.Is(err, filereader.ErrEndOfFile) {
						break
					}
					filereader.FileClose(reader)
					return nil, fmt.Errorf("MCPLoadChain: reading artifact %q: %w", dep.LogicalName, err)
				}
				context += line + "\n"
			}
			filereader.FileClose(reader)

		} else if dep.Qualifier == "" {
			// 8b. No qualifier — emit full public section.
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return nil, fmt.Errorf("MCPLoadChain: parsing dependency %q: %w", dep.LogicalName, err)
			}
			context += node.Public.RawHeading + "\n"
			for _, line := range node.Public.Content {
				context += line + "\n"
			}
			for _, sub := range node.Public.Subsections {
				context += sub.RawHeading + "\n"
				for _, line := range sub.Content {
					context += line + "\n"
				}
			}

		} else {
			// 8c. Qualifier present — emit the matching subsection.
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return nil, fmt.Errorf("MCPLoadChain: parsing dependency %q: %w", dep.LogicalName, err)
			}
			normalizedQualifier := textnormalization.NormalizeText(dep.Qualifier)
			var matchedSub *parsenode.NodeSubsection
			for _, sub := range node.Public.Subsections {
				if sub.Heading == normalizedQualifier {
					matchedSub = sub
					break
				}
			}
			if matchedSub != nil {
				context += matchedSub.RawHeading + "\n"
				for _, line := range matchedSub.Content {
					context += line + "\n"
				}
			}
		}
	}

	// 9. External.
	for _, ext := range chain.External {
		extPath := &pathutils.PathCfs{Value: ext.Path}

		if len(ext.Fragments) == 0 {
			// 9b. No fragments — read entire file.
			reader, err := filereader.FileOpen(extPath)
			if err != nil {
				return nil, fmt.Errorf("MCPLoadChain: opening external %q: %w", ext.Path, err)
			}
			for {
				line, err := filereader.FileReadLine(reader)
				if err != nil {
					if errors.Is(err, filereader.ErrEndOfFile) {
						break
					}
					filereader.FileClose(reader)
					return nil, fmt.Errorf("MCPLoadChain: reading external %q: %w", ext.Path, err)
				}
				context += line + "\n"
			}
			filereader.FileClose(reader)

		} else {
			// 9c. Fragments present.
			for _, frag := range ext.Fragments {
				start, end, err := parseLineRange(frag.Lines)
				if err != nil {
					return nil, fmt.Errorf("MCPLoadChain: parsing fragment lines %q: %w", frag.Lines, err)
				}
				reader, err := filereader.FileOpen(extPath)
				if err != nil {
					return nil, fmt.Errorf("MCPLoadChain: opening external fragment %q: %w", ext.Path, err)
				}
				filereader.FileSkipLines(reader, start-1)
				count := end - start + 1
				for i := 0; i < count; i++ {
					line, err := filereader.FileReadLine(reader)
					if err != nil {
						if errors.Is(err, filereader.ErrEndOfFile) {
							break
						}
						filereader.FileClose(reader)
						return nil, fmt.Errorf("MCPLoadChain: reading external fragment %q: %w", ext.Path, err)
					}
					context += line + "\n"
				}
				filereader.FileClose(reader)
			}
		}
	}

	// 10. Target Public.

	// 10a. Emit reduced frontmatter block.
	context += "---\n"
	context += "outputs:\n"
	for _, output := range fm.Outputs {
		context += "  - id: " + output.ID + "\n"
		context += "    path: " + output.Path + "\n"
	}
	context += "---\n"

	// 10b. Parse the target node.
	targetNode, err := parsenode.NodeParse(chain.Target.LogicalName)
	if err != nil {
		return nil, fmt.Errorf("MCPLoadChain: parsing target node: %w", err)
	}

	// 10c. Emit public section if present.
	if targetNode.Public != nil {
		context += targetNode.Public.RawHeading + "\n"
		for _, line := range targetNode.Public.Content {
			context += line + "\n"
		}
		for _, sub := range targetNode.Public.Subsections {
			context += sub.RawHeading + "\n"
			for _, line := range sub.Content {
				context += line + "\n"
			}
		}
	}

	// 11. Target Agent.
	if targetNode.Agent != nil {
		context += targetNode.Agent.RawHeading + "\n"
		for _, line := range targetNode.Agent.Content {
			context += line + "\n"
		}
		for _, sub := range targetNode.Agent.Subsections {
			context += sub.RawHeading + "\n"
			for _, line := range sub.Content {
				context += line + "\n"
			}
		}
	}

	// Step 4 — Extract input

	// 12. Read input artifact if present.
	var inputContent *string
	if chain.Input != nil {
		reader, err := filereader.FileOpen(&chain.Input.FilePath)
		if err != nil {
			return nil, fmt.Errorf("MCPLoadChain: opening input artifact: %w", err)
		}
		firstLine, hasFirstLine, err := stripFrontmatter(reader)
		if err != nil {
			filereader.FileClose(reader)
			return nil, fmt.Errorf("MCPLoadChain: stripping frontmatter from input: %w", err)
		}
		var inputBuf string
		if hasFirstLine {
			inputBuf += firstLine + "\n"
		}
		for {
			line, err := filereader.FileReadLine(reader)
			if err != nil {
				if errors.Is(err, filereader.ErrEndOfFile) {
					break
				}
				filereader.FileClose(reader)
				return nil, fmt.Errorf("MCPLoadChain: reading input artifact: %w", err)
			}
			inputBuf += line + "\n"
		}
		filereader.FileClose(reader)
		inputContent = &inputBuf
	}

	// Step 5 — Return result

	return &MCPLoadChainResult{
		ChainHash: chainHash,
		Context:   context,
		Input:     inputContent,
	}, nil
}

// stripFrontmatter reads and discards a YAML frontmatter block (delimited by
// "---" lines) from reader. It returns the first content line after the
// frontmatter, a bool indicating whether a first content line was returned,
// and any read error.
//
// If the first line of the file is "---", the function discards all lines up
// to and including the closing "---". The first content line is then read and
// returned as the first line.
//
// If the first line is not "---", frontmatter is absent. That line is returned
// as-is as the first content line (hasFirstLine = true).
func stripFrontmatter(reader *filereader.FileReader) (firstLine string, hasFirstLine bool, err error) {
	line, err := filereader.FileReadLine(reader)
	if err != nil {
		if errors.Is(err, filereader.ErrEndOfFile) {
			return "", false, nil
		}
		return "", false, err
	}

	if line != "---" {
		// No frontmatter — this line is content.
		return line, true, nil
	}

	// Frontmatter present — discard until closing "---".
	for {
		l, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				// Frontmatter was never closed; no content remains.
				return "", false, nil
			}
			return "", false, err
		}
		if l == "---" {
			// Closing delimiter found; no first content line to return yet.
			return "", false, nil
		}
	}
}

// parseLineRange parses a "start-end" fragment line range string into
// integer start and end values.
func parseLineRange(lines string) (start int, end int, err error) {
	_, err = fmt.Sscanf(lines, "%d-%d", &start, &end)
	if err != nil {
		return 0, 0, fmt.Errorf("parseLineRange: invalid range %q: %w", lines, err)
	}
	return start, end, nil
}
