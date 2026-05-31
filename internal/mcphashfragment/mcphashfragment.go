// code-from-spec: ROOT/golang/implementation/mcp_tools/hash_fragment@gEuAiA4DvQvgI14AtMu31JnKI14

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

// ErrInvalidLineRange is returned when the lines parameter does not
// match the expected format, when start < 1, when start > end, or
// when end exceeds the total number of lines in the file.
var ErrInvalidLineRange = errors.New("invalid line range")

// MCPHashFragment reads the lines denoted by the range string from the
// file at path (relative to the project root) and returns a SHA-1
// digest of that fragment encoded as a 27-character base64url string
// (RFC 4648 §5, no padding).
//
// path must be a forward-slash relative path validated by PathValidateCfs.
// lines must be a range of the form "start-end" (e.g. "150-210") where
// start >= 1 and end <= the total number of lines in the file.
//
// Errors:
//   - ErrInvalidLineRange: the range format is invalid, start < 1,
//     start > end, or end exceeds the file's line count.
//   - PathUtils errors propagated from PathValidateCfs.
//   - FileReader errors propagated from FileOpen.
func MCPHashFragment(path string, lines string) (string, error) {
	// Step 1: Validate path.
	if err := pathutils.PathValidateCfs(path); err != nil {
		return "", fmt.Errorf("path validation failed: %w", err)
	}

	// Step 2: Parse line range.
	parts := strings.SplitN(lines, "-", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("%w: expected format start-end", ErrInvalidLineRange)
	}

	start, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", fmt.Errorf("%w: start is not an integer", ErrInvalidLineRange)
	}

	end, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("%w: end is not an integer", ErrInvalidLineRange)
	}

	if start < 1 {
		return "", fmt.Errorf("%w: start must be >= 1", ErrInvalidLineRange)
	}

	if start > end {
		return "", fmt.Errorf("%w: start must be <= end", ErrInvalidLineRange)
	}

	// Step 3: Read lines.
	cfsPath := &pathutils.PathCfs{Value: path}

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		return "", fmt.Errorf("could not open file: %w", err)
	}

	filereader.FileSkipLines(reader, start-1)

	linesToRead := end - start + 1
	collectedLines := make([]string, 0, linesToRead)

	for i := 0; i < linesToRead; i++ {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			filereader.FileClose(reader)
			if errors.Is(err, filereader.ErrEndOfFile) {
				return "", fmt.Errorf("%w: end exceeds the file's line count", ErrInvalidLineRange)
			}
			return "", fmt.Errorf("error reading line: %w", err)
		}
		collectedLines = append(collectedLines, line)
	}

	filereader.FileClose(reader)

	// Step 4: Compute hash.
	var sb strings.Builder
	for _, line := range collectedLines {
		sb.WriteString(line)
		sb.WriteByte('\n')
	}

	digest := sha1.Sum([]byte(sb.String()))
	hash := base64.RawURLEncoding.EncodeToString(digest[:])

	return hash, nil
}
