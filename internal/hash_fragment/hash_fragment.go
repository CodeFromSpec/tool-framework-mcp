// code-from-spec: ROOT/golang/internal/tools/hash_fragment/code@PENDING
package hash_fragment

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathvalidation"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// HashFragmentArgs defines the input parameters for the hash_fragment tool.
type HashFragmentArgs struct {
	Path  string `json:"path" jsonschema:"File path relative to project root."`
	Lines string `json:"lines" jsonschema:"Line range (e.g., 150-210)."`
}

// HandleHashFragment validates the file path, reads the specified line range,
// and returns a SHA-1 hash (base64url encoded, 27 chars) of the extracted content.
func HandleHashFragment(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args HashFragmentArgs,
) (*mcp.CallToolResult, any, error) {
	// Step 1: Validate path against working directory.
	if err := pathvalidation.ValidatePath(args.Path, "."); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			IsError: true,
		}, nil, nil
	}

	// Step 2: Parse the line range.
	start, end, err := parseLineRange(args.Lines)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("invalid line range: %s", args.Lines)}},
			IsError: true,
		}, nil, nil
	}

	// Step 3: Read the file using filereader.
	fr, err := filereader.OpenFileReader(args.Path)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("file not found: %s", args.Path)}},
			IsError: true,
		}, nil, nil
	}

	// Step 4: Read all lines.
	var allLines []string
	for {
		line, err := fr.ReadLine()
		if errors.Is(err, filereader.ErrEndOfFile) {
			break
		}
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("error reading file: %s", args.Path)}},
				IsError: true,
			}, nil, nil
		}
		allLines = append(allLines, line)
	}

	// Step 5: Validate line range against file length.
	if end > len(allLines) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("invalid line range: %s (file has %d lines)", args.Lines, len(allLines))}},
			IsError: true,
		}, nil, nil
	}

	// Step 6: Extract lines (1-indexed, inclusive).
	extracted := allLines[start-1 : end]

	// Step 7: Join with LF.
	content := strings.Join(extracted, "\n")

	// Step 8: Compute SHA-1 and encode as base64url (no padding, 27 chars).
	hash := sha1.Sum([]byte(content))
	encoded := base64.RawURLEncoding.EncodeToString(hash[:])
	// Truncate to 27 characters.
	if len(encoded) > 27 {
		encoded = encoded[:27]
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: encoded}},
	}, nil, nil
}

// parseLineRange parses a "start-end" string into two 1-indexed integers.
func parseLineRange(lines string) (int, int, error) {
	parts := strings.SplitN(lines, "-", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid format")
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start: %w", err)
	}

	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end: %w", err)
	}

	if start < 1 {
		return 0, 0, fmt.Errorf("start must be >= 1")
	}

	if start > end {
		return 0, 0, fmt.Errorf("start must be <= end")
	}

	return start, end, nil
}
