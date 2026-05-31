// code-from-spec: ROOT/golang/implementation/mcp_tools/hash_fragment@4l7xmNsgD1lLddjwlVZmF9nCMbE
package mcphashfragment

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ErrInvalidLineRange is returned when the line range format is invalid,
// the start line is less than 1, start exceeds end, or end exceeds the
// file's total line count.
var ErrInvalidLineRange = errors.New("invalid line range")

// MCPHashFragment reads lines [start, end] (inclusive, 1-based) from the
// file at path (relative to the project root, forward slashes), computes
// a SHA-1 digest of those lines, and returns it as a base64url-encoded
// string (RFC 4648 §5, no padding, 27 characters).
//
// path must pass PathUtils.PathValidateCfs validation.
// lines must be a range string of the form "start-end" (e.g. "150-210").
//
// Errors:
//   - ErrInvalidLineRange: range format is invalid, start < 1,
//     start > end, or end exceeds the file's line count.
//   - PathUtils errors propagated from PathValidateCfs.
//   - FileReader errors propagated from FileOpen.
func MCPHashFragment(path string, lines string) (string, error) {
	// Step 1: Validate the CFS path.
	if err := pathutils.PathValidateCfs(path); err != nil {
		return "", err
	}

	// Step 2: Parse the line range "start-end".
	parts := strings.SplitN(lines, "-", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("%w: invalid line range format", ErrInvalidLineRange)
	}
	start, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", fmt.Errorf("%w: invalid line range format", ErrInvalidLineRange)
	}
	end, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("%w: invalid line range format", ErrInvalidLineRange)
	}
	if start < 1 {
		return "", fmt.Errorf("%w: start line must be >= 1", ErrInvalidLineRange)
	}
	if start > end {
		return "", fmt.Errorf("%w: start line must be <= end line", ErrInvalidLineRange)
	}

	// Step 3: Open the file.
	cfsPath := &pathutils.PathCfs{Value: path}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		return "", err
	}

	// Step 4: Skip lines before the requested range.
	filereader.FileSkipLines(reader, start-1)

	// Step 5: Read the lines in the range.
	linesToRead := end - start + 1
	collectedLines := make([]string, 0, linesToRead)
	for i := 0; i < linesToRead; i++ {
		line, err := filereader.FileReadLine(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			filereader.FileClose(reader)
			return "", fmt.Errorf("%w: end line exceeds file line count", ErrInvalidLineRange)
		}
		if err != nil {
			filereader.FileClose(reader)
			return "", err
		}
		collectedLines = append(collectedLines, line)
	}

	// Step 6: Close the reader.
	filereader.FileClose(reader)

	// Step 7: Build the hash input string, appending "\n" after each line.
	var sb strings.Builder
	for _, line := range collectedLines {
		sb.WriteString(line)
		sb.WriteByte('\n')
	}
	hashInput := sb.String()

	// Step 8: Compute SHA-1 digest and encode as base64url (no padding).
	digest := sha1.Sum([]byte(hashInput))
	encoded := base64.RawURLEncoding.EncodeToString(digest[:])

	// Step 9: Return the 27-character base64url-encoded hash string.
	return encoded, nil
}
