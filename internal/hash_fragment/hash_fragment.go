// code-from-spec: ROOT/golang/internal/tools/hash_fragment/code@ZMGDUERO3vWP5m2-mYkPlifghRU

// Package hash_fragment implements the hash_fragment MCP tool.
//
// The tool accepts a file path and a line range (e.g. "150-210"), extracts
// the specified lines from the file, and returns a SHA-1 hash of the joined
// content encoded as base64url (RFC 4648 §5, no padding). This produces a
// 27-character string suitable for use in external fragment declarations.
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
type HashFragmentArgs struct {
	// Path is the file path relative to the project root.
	Path string `json:"path" jsonschema:"File path relative to project root."`

	// Lines is the line range in "start-end" format (1-indexed, inclusive).
	// Example: "150-210"
	Lines string `json:"lines" jsonschema:"Line range (e.g., 150-210)."`
}

// HandleHashFragment is the MCP tool handler for hash_fragment.
//
// It validates the path, parses the line range, reads the specified lines
// from the file, and returns a SHA-1/base64url hash of the joined content.
//
// All expected error conditions return an MCP tool error (IsError: true).
// The Go error return is reserved for catastrophic server failures and is
// always nil here.
func HandleHashFragment(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args HashFragmentArgs,
) (*mcp.CallToolResult, any, error) {

	// Step 1: Validate the path against the working directory (project root).
	// This guards against directory traversal, absolute paths, etc.
	wd, err := os.Getwd()
	if err != nil {
		// If we cannot determine the working directory the server is misconfigured;
		// treat as a tool error so the server keeps running.
		return toolError("could not determine working directory: " + err.Error()), nil, nil
	}

	if err := pathvalidation.ValidatePath(args.Path, wd); err != nil {
		return toolError(err.Error()), nil, nil
	}

	// Step 2: Parse the line range "start-end".
	// Both start and end are 1-indexed and inclusive.
	// Conditions that make the range invalid:
	//   - Not exactly two integers separated by a hyphen
	//   - start < 1
	//   - start > end
	start, end, ok := parseLineRange(args.Lines)
	if !ok {
		return toolError(fmt.Sprintf("invalid line range: %s", args.Lines)), nil, nil
	}

	// Step 3: Open the file for sequential reading using filereader.
	// filereader normalises CRLF line endings, so we do not need to handle
	// them here.
	r, err := filereader.OpenFileReader(args.Path)
	if err != nil {
		if errors.Is(err, filereader.ErrOpen) {
			return toolError(fmt.Sprintf("file not found: %s", args.Path)), nil, nil
		}
		return toolError(fmt.Sprintf("could not open file %s: %v", args.Path, err)), nil, nil
	}

	// Step 4: Collect lines start..end (1-indexed, inclusive).
	// We read through the file line by line, keeping only the lines we need,
	// and track the total line count to produce a useful error if end exceeds it.
	var extracted []string
	lineNum := 0 // current 1-indexed line number (incremented before use)

	for {
		line, err := r.ReadLine()
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				break
			}
			// Unexpected read error.
			return toolError(fmt.Sprintf("error reading file %s: %v", args.Path, err)), nil, nil
		}
		lineNum++

		if lineNum >= start && lineNum <= end {
			extracted = append(extracted, line)
		}

		// Once we have read past the end of the requested range we can stop
		// reading — unless we still need to count total lines for a potential
		// bounds error. We must continue to know the total line count only if
		// end is beyond the file; but we do not know that yet.  Keep reading
		// until EOF so we always have the accurate total.
	}

	// Check whether end exceeds the file's total line count.
	if end > lineNum {
		return toolError(fmt.Sprintf(
			"invalid line range: %s (file has %d lines)", args.Lines, lineNum,
		)), nil, nil
	}

	// Step 5: Join the extracted lines with LF.
	joined := strings.Join(extracted, "\n")

	// Step 6: Compute SHA-1 of the joined content.
	h := sha1.New()
	h.Write([]byte(joined))
	digest := h.Sum(nil)

	// Step 7: Encode the hash as base64url without padding (RFC 4648 §5).
	// This produces a 27-character string for a 20-byte SHA-1 digest.
	encoded := base64.RawURLEncoding.EncodeToString(digest)

	// Step 8: Return the hash as a success result.
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: encoded}},
	}, nil, nil
}

// parseLineRange parses a "start-end" string into two 1-indexed integers.
// Returns (start, end, true) on success.
// Returns (0, 0, false) if:
//   - the string does not contain exactly one hyphen separating two integers
//   - either part is not a valid positive integer
//   - start < 1
//   - start > end
func parseLineRange(s string) (start, end int, ok bool) {
	// Split on the first hyphen only; this means a range like "10-20" splits
	// cleanly, while "10-20-30" is deliberately rejected (more than one part
	// after the first hyphen).
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}

	startStr := strings.TrimSpace(parts[0])
	endStr := strings.TrimSpace(parts[1])

	// Reject empty halves or values containing additional hyphens (e.g. "10-20-30").
	if startStr == "" || endStr == "" || strings.Contains(endStr, "-") {
		return 0, 0, false
	}

	s64, err := strconv.ParseInt(startStr, 10, 64)
	if err != nil {
		return 0, 0, false
	}
	e64, err := strconv.ParseInt(endStr, 10, 64)
	if err != nil {
		return 0, 0, false
	}

	start = int(s64)
	end = int(e64)

	if start < 1 || start > end {
		return 0, 0, false
	}

	return start, end, true
}

// toolError is a convenience function that builds an MCP tool error result.
func toolError(message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: message}},
		IsError: true,
	}
}
