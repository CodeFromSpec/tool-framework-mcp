// code-from-spec: ROOT/golang/internal/tools/hash_fragment/code@_jlZY8-gvoMCkABeQregl6nHPK4

// Package hash_fragment implements the hash_fragment MCP tool.
//
// The tool accepts a file path and a line range (e.g., "150-210"), reads the
// specified lines from the file, computes a SHA-1 hash of the joined content,
// and returns the hash as a 27-character base64url string (RFC 4648 §5, no
// padding). This is the same algorithm used by load_chain to produce chain
// hashes, so callers can use the output directly in external: fragment
// declarations.
package hash_fragment

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathvalidation"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// HashFragmentArgs holds the input parameters for the hash_fragment tool.
// Field tags drive the MCP input-schema inference.
type HashFragmentArgs struct {
	// Path is the file path relative to the project root.
	Path string `json:"path" jsonschema:"File path relative to project root."`
	// Lines is the line range in "start-end" format (1-indexed, inclusive),
	// for example "150-210".
	Lines string `json:"lines" jsonschema:"Line range (e.g., 150-210)."`
}

// HandleHashFragment is the MCP tool handler for hash_fragment.
//
// Steps (as specified):
//  1. Validate args.Path using pathvalidation.ValidatePath.
//  2. Parse args.Lines as "start-end" (1-indexed, inclusive integers).
//  3. Read the file using filereader.
//  4. Extract lines [start, end]; error if end > line count.
//  5. Join extracted lines with LF.
//  6. Compute SHA-1 of joined content.
//  7. Encode with base64url, no padding (RawURLEncoding) → 27 chars.
//  8. Return hash as success text.
//
// All expected error conditions return an MCP tool error (IsError: true).
// The returned Go error is reserved for catastrophic server failures and is
// always nil here.
func HandleHashFragment(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args HashFragmentArgs,
) (*mcp.CallToolResult, any, error) {
	// -------------------------------------------------------------------------
	// Step 1 — Validate the path against the project root (working directory).
	// -------------------------------------------------------------------------
	projectRoot, err := os.Getwd()
	if err != nil {
		return toolError(fmt.Sprintf("cannot determine project root: %v", err)), nil, nil
	}
	if err := pathvalidation.ValidatePath(args.Path, projectRoot); err != nil {
		return toolError(err.Error()), nil, nil
	}

	// -------------------------------------------------------------------------
	// Step 2 — Parse the line range.
	//
	// Expected format: "<start>-<end>" where both are integers, start >= 1,
	// and start <= end.
	// -------------------------------------------------------------------------
	start, end, ok := parseLineRange(args.Lines)
	if !ok {
		return toolError(fmt.Sprintf("invalid line range: %s", args.Lines)), nil, nil
	}

	// -------------------------------------------------------------------------
	// Step 3 — Open and read the file using the filereader package.
	// -------------------------------------------------------------------------
	fr, err := filereader.OpenFileReader(args.Path)
	if err != nil {
		// OpenFileReader wraps ErrOpen; treat any open failure as "not found".
		if errors.Is(err, filereader.ErrOpen) {
			return toolError(fmt.Sprintf("file not found: %s", args.Path)), nil, nil
		}
		// Unexpected error — still a tool error, not a Go error.
		return toolError(fmt.Sprintf("file not found: %s", args.Path)), nil, nil
	}
	defer fr.Close()

	// -------------------------------------------------------------------------
	// Step 4 — Extract lines [start, end] (1-indexed, inclusive).
	//
	// We read the entire file so we can also report the total line count when
	// the range is out of bounds.
	// -------------------------------------------------------------------------
	var allLines []string
	for {
		line, readErr := fr.ReadLine()
		if errors.Is(readErr, filereader.ErrEndOfFile) {
			break
		}
		if readErr != nil {
			return toolError(fmt.Sprintf("error reading file: %s", args.Path)), nil, nil
		}
		allLines = append(allLines, line)
	}

	totalLines := len(allLines)

	// Validate that the requested range fits within the file.
	if end > totalLines {
		return toolError(fmt.Sprintf(
			"invalid line range: %s (file has %d lines)",
			args.Lines, totalLines,
		)), nil, nil
	}

	// Slice is 0-indexed; start/end are 1-indexed inclusive.
	extracted := allLines[start-1 : end]

	// -------------------------------------------------------------------------
	// Step 5 — Join extracted lines with LF.
	// Line endings are already normalised by filereader (CRLF → LF stripped).
	// -------------------------------------------------------------------------
	joined := strings.Join(extracted, "\n")

	// -------------------------------------------------------------------------
	// Step 6 — Compute SHA-1 of the joined content.
	// -------------------------------------------------------------------------
	sum := sha1.Sum([]byte(joined)) // [20]byte

	// -------------------------------------------------------------------------
	// Step 7 — Encode as base64url without padding (RFC 4648 §5).
	// sha1 produces 20 bytes → ceil(20*8/6) = 27 base64url characters.
	// -------------------------------------------------------------------------
	hash := base64.RawURLEncoding.EncodeToString(sum[:])

	// -------------------------------------------------------------------------
	// Step 8 — Return the hash as a success result.
	// -------------------------------------------------------------------------
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: hash}},
	}, nil, nil
}

// parseLineRange parses a "start-end" line range string.
// Both start and end are 1-indexed integers (inclusive).
// Returns (start, end, true) on success, or (0, 0, false) on any failure.
//
// Failure cases:
//   - Not exactly two parts separated by "-"
//   - Either part is not a valid integer
//   - start < 1
//   - start > end
func parseLineRange(s string) (start, end int, ok bool) {
	// Split on the first hyphen only — negative numbers are not valid line
	// numbers, so a simple split by "-" into exactly 2 parts is correct.
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}

	var err error
	start, err = strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, false
	}
	end, err = strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, false
	}

	// start must be at least 1 (lines are 1-indexed) and must not exceed end.
	if start < 1 || start > end {
		return 0, 0, false
	}

	return start, end, true
}

// toolError is a convenience helper that returns an MCP tool error result.
// Using a helper keeps the handler body readable and consistent with the
// contract: IsError: true, single TextContent entry.
func toolError(message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: message}},
		IsError: true,
	}
}
