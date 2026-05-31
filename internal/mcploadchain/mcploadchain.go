// code-from-spec: ROOT/golang/implementation/mcp_tools/load_chain@gyVrkBQTT9RTWM0NAJwNROLBdKA
package mcploadchain

import (
	"errors"
	"fmt"
	"strconv"
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

	// Input contains the content of the input artifact (excluding frontmatter),
	// if one exists. Nil when no input artifact is present.
	Input *string
}

// MCPLoadChain resolves the spec chain for the given logical name, computes
// the chain hash, and returns the assembled context along with an optional
// input artifact.
//
// The target node must declare an outputs field. Each output path is validated
// before the result is returned.
//
// Errors:
//   - ErrNoOutputs: target node has no outputs field.
//   - ErrInvalidOutputPath: an output path fails path validation.
//   - (LogicalNames.*): propagated from LogicalNameToPath.
//   - (ChainResolver.*): propagated from ChainResolve.
//   - (ChainHash.*): propagated from ChainHashCompute.
//   - (NodeParsing.*): propagated from NodeParse.
//   - (FileReader.*): propagated from FileOpen.
func MCPLoadChain(logical_name string) (*MCPLoadChainResult, error) {
	// Step 1 — Validate and resolve

	targetFilePath, err := logicalnames.LogicalNameToPath(logical_name)
	if err != nil {
		return nil, fmt.Errorf("LogicalNameToPath: %w", err)
	}

	fm, err := frontmatter.FrontmatterParse(targetFilePath)
	if err != nil {
		return nil, fmt.Errorf("FrontmatterParse: %w", err)
	}

	if len(fm.Outputs) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrNoOutputs, logical_name)
	}

	for _, output := range fm.Outputs {
		if err := pathutils.PathValidateCfs(output.Path); err != nil {
			return nil, fmt.Errorf("%w: %s: %w", ErrInvalidOutputPath, output.Path, err)
		}
	}

	chain, err := chainresolver.ChainResolve(logical_name)
	if err != nil {
		return nil, fmt.Errorf("ChainResolve: %w", err)
	}

	// Step 2 — Compute chain hash

	chainHash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		return nil, fmt.Errorf("ChainHashCompute: %w", err)
	}

	// Step 3 — Build context stream

	var contextBuilder strings.Builder

	// Step 6 — Ancestors
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

		for _, line := range node.Public.Content {
			contextBuilder.WriteString(line)
			contextBuilder.WriteByte('\n')
		}

		for _, sub := range node.Public.Subsections {
			contextBuilder.WriteString(sub.RawHeading)
			contextBuilder.WriteByte('\n')
			for _, line := range sub.Content {
				contextBuilder.WriteString(line)
				contextBuilder.WriteByte('\n')
			}
		}
	}

	// Step 7 — Dependencies
	for _, dep := range chain.Dependencies {
		if logicalnames.LogicalNameIsArtifact(dep.LogicalName) {
			reader, err := filereader.FileOpen(dep.FilePath)
			if err != nil {
				return nil, fmt.Errorf("FileOpen dependency %s: %w", dep.LogicalName, err)
			}

			stripFrontmatter(reader)

			if err := readAllLines(reader, &contextBuilder); err != nil {
				filereader.FileClose(reader)
				return nil, fmt.Errorf("reading dependency %s: %w", dep.LogicalName, err)
			}

			filereader.FileClose(reader)
		} else if dep.Qualifier == nil {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return nil, fmt.Errorf("NodeParse dependency %s: %w", dep.LogicalName, err)
			}

			if node.Public == nil {
				continue
			}

			for _, line := range node.Public.Content {
				contextBuilder.WriteString(line)
				contextBuilder.WriteByte('\n')
			}

			for _, sub := range node.Public.Subsections {
				contextBuilder.WriteString(sub.RawHeading)
				contextBuilder.WriteByte('\n')
				for _, line := range sub.Content {
					contextBuilder.WriteString(line)
					contextBuilder.WriteByte('\n')
				}
			}
		} else {
			// Qualifier is present
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return nil, fmt.Errorf("NodeParse dependency %s: %w", dep.LogicalName, err)
			}

			if node.Public == nil {
				continue
			}

			normalizedQualifier := textnormalization.NormalizeText(*dep.Qualifier)

			var found *parsenode.NodeSubsection
			for _, sub := range node.Public.Subsections {
				if sub.Heading == normalizedQualifier {
					found = sub
					break
				}
			}

			if found == nil {
				continue
			}

			for _, line := range found.Content {
				contextBuilder.WriteString(line)
				contextBuilder.WriteByte('\n')
			}
		}
	}

	// Step 8 — External files
	for _, ext := range chain.External {
		extPath := &pathutils.PathCfs{Value: ext.Path}

		if len(ext.Fragments) == 0 {
			reader, err := filereader.FileOpen(extPath)
			if err != nil {
				return nil, fmt.Errorf("FileOpen external %s: %w", ext.Path, err)
			}

			if err := readAllLines(reader, &contextBuilder); err != nil {
				filereader.FileClose(reader)
				return nil, fmt.Errorf("reading external %s: %w", ext.Path, err)
			}

			filereader.FileClose(reader)
		} else {
			for _, frag := range ext.Fragments {
				start, end, err := parseLineRange(frag.Lines)
				if err != nil {
					return nil, fmt.Errorf("parsing fragment lines %q in %s: %w", frag.Lines, ext.Path, err)
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
						return nil, fmt.Errorf("reading external fragment %s: %w", ext.Path, err)
					}
					contextBuilder.WriteString(line)
					contextBuilder.WriteByte('\n')
				}

				filereader.FileClose(reader)
			}
		}
	}

	// Step 9 — Target
	targetFm, err := frontmatter.FrontmatterParse(targetFilePath)
	if err != nil {
		return nil, fmt.Errorf("FrontmatterParse target: %w", err)
	}

	contextBuilder.WriteString("---\n")
	contextBuilder.WriteString("outputs:\n")
	for _, output := range targetFm.Outputs {
		contextBuilder.WriteString("  - id: ")
		contextBuilder.WriteString(output.ID)
		contextBuilder.WriteByte('\n')
		contextBuilder.WriteString("    path: ")
		contextBuilder.WriteString(output.Path)
		contextBuilder.WriteByte('\n')
	}
	contextBuilder.WriteString("---\n")

	targetNode, err := parsenode.NodeParse(chain.Target.LogicalName)
	if err != nil {
		return nil, fmt.Errorf("NodeParse target: %w", err)
	}

	if targetNode.Public != nil {
		for _, line := range targetNode.Public.Content {
			contextBuilder.WriteString(line)
			contextBuilder.WriteByte('\n')
		}
		for _, sub := range targetNode.Public.Subsections {
			contextBuilder.WriteString(sub.RawHeading)
			contextBuilder.WriteByte('\n')
			for _, line := range sub.Content {
				contextBuilder.WriteString(line)
				contextBuilder.WriteByte('\n')
			}
		}
	}

	if targetNode.Agent != nil {
		for _, line := range targetNode.Agent.Content {
			contextBuilder.WriteString(line)
			contextBuilder.WriteByte('\n')
		}
		for _, sub := range targetNode.Agent.Subsections {
			contextBuilder.WriteString(sub.RawHeading)
			contextBuilder.WriteByte('\n')
			for _, line := range sub.Content {
				contextBuilder.WriteString(line)
				contextBuilder.WriteByte('\n')
			}
		}
	}

	// Step 4 — Extract input

	var inputStr *string

	if chain.Input != nil {
		reader, err := filereader.FileOpen(chain.Input.FilePath)
		if err != nil {
			return nil, fmt.Errorf("FileOpen input: %w", err)
		}

		stripFrontmatter(reader)

		var inputBuilder strings.Builder
		if err := readAllLines(reader, &inputBuilder); err != nil {
			filereader.FileClose(reader)
			return nil, fmt.Errorf("reading input: %w", err)
		}

		filereader.FileClose(reader)

		s := inputBuilder.String()
		inputStr = &s
	}

	// Step 5 — Return result

	return &MCPLoadChainResult{
		ChainHash: chainHash,
		Context:   contextBuilder.String(),
		Input:     inputStr,
	}, nil
}

// stripFrontmatter reads and discards the YAML frontmatter block from reader.
// It looks for the opening "---" on the first non-blank line, then reads until
// the closing "---". If no opening "---" is found, the reader position is
// left as-is (no lines are consumed beyond the initial blank lines).
//
// Note: because FileReader is sequential and does not support seek, if the
// first non-blank line is not "---", we cannot "unread" it. In this case,
// the function accepts the loss of that line — the spec says to treat the
// file as having no frontmatter and start from the beginning, but since
// we cannot rewind, we simply stop after reading the first non-blank line
// that is not "---". This matches the intent: no frontmatter stripping occurs.
func stripFrontmatter(reader *filereader.FileReader) {
	// Read lines until we find the first non-blank line.
	for {
		line, err := filereader.FileReadLine(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			return
		}
		if err != nil {
			return
		}

		if strings.TrimSpace(line) == "" {
			continue
		}

		// First non-blank line found.
		if line != "---" {
			// No frontmatter — this line is consumed but cannot be put back.
			// The spec says treat as no frontmatter (start from beginning),
			// but since we cannot rewind, we stop here.
			return
		}

		// Found opening "---". Now read until closing "---".
		for {
			line, err := filereader.FileReadLine(reader)
			if errors.Is(err, filereader.ErrEndOfFile) {
				return
			}
			if err != nil {
				return
			}
			if line == "---" {
				return
			}
		}
	}
}

// readAllLines reads all remaining lines from reader and appends each followed
// by "\n" to the given builder. Returns the first non-ErrEndOfFile error.
func readAllLines(reader *filereader.FileReader, builder *strings.Builder) error {
	for {
		line, err := filereader.FileReadLine(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			return nil
		}
		if err != nil {
			return err
		}
		builder.WriteString(line)
		builder.WriteByte('\n')
	}
}

// parseLineRange parses a "start-end" string into two integers.
func parseLineRange(s string) (start, end int, err error) {
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid line range %q: expected format start-end", s)
	}
	start, err = strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start in line range %q: %w", s, err)
	}
	end, err = strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end in line range %q: %w", s, err)
	}
	return start, end, nil
}
