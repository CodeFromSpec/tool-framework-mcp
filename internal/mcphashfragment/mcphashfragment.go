// code-from-spec: ROOT/golang/implementation/mcp_tools/hash_fragment@Ws2PAf68eaVICexS_-KdmPRiffE
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

// ErrInvalidLineRange is returned when the lines parameter has an invalid
// format, start < 1, start > end, or end exceeds the file's line count.
var ErrInvalidLineRange = errors.New("invalid line range")

// MCPHashFragment computes a SHA-1 digest of the specified line range within
// the given file. The digest is base64url encoded (RFC 4648 §5, no padding),
// producing a 27-character string.
//
// path must be a file path relative to the project root using forward slashes.
// lines must be a line range in the form "start-end" (e.g., "150-210").
//
// Errors:
//   - ErrInvalidLineRange: the range format is invalid, start < 1,
//     start > end, or end exceeds the file's line count.
//   - PathUtils errors: propagated from PathValidateCfs.
//   - FileReader errors: propagated from FileOpen.
func MCPHashFragment(path string, lines string) (string, error) {
	// Step 1: Validate path.
	if err := pathutils.PathValidateCfs(path); err != nil {
		return "", fmt.Errorf("MCPHashFragment: %w", err)
	}

	// Step 2: Parse line range.
	parts := strings.Split(lines, "-")
	if len(parts) != 2 {
		return "", fmt.Errorf("%w: invalid line range format: expected <start>-<end>", ErrInvalidLineRange)
	}

	start, errStart := strconv.Atoi(parts[0])
	end, errEnd := strconv.Atoi(parts[1])
	if errStart != nil || errEnd != nil {
		return "", fmt.Errorf("%w: invalid line range format: start and end must be integers", ErrInvalidLineRange)
	}

	if start < 1 {
		return "", fmt.Errorf("%w: invalid line range: start must be >= 1", ErrInvalidLineRange)
	}
	if start > end {
		return "", fmt.Errorf("%w: invalid line range: start must be <= end", ErrInvalidLineRange)
	}

	// Step 3: Read lines from the file.
	cfsPath := &pathutils.PathCfs{Value: path}

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		return "", fmt.Errorf("MCPHashFragment: %w", err)
	}

	filereader.FileSkipLines(reader, start-1)

	lineCount := end - start + 1
	collectedLines := make([]string, 0, lineCount)

	for i := 0; i < lineCount; i++ {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				filereader.FileClose(reader)
				return "", fmt.Errorf("%w: invalid line range: end exceeds the file's line count", ErrInvalidLineRange)
			}
			filereader.FileClose(reader)
			return "", fmt.Errorf("MCPHashFragment: %w", err)
		}
		collectedLines = append(collectedLines, line)
	}

	filereader.FileClose(reader)

	// Step 4: Compute the hash.
	var sb strings.Builder
	for _, line := range collectedLines {
		sb.WriteString(line)
		sb.WriteByte('\n')
	}

	content := sb.String()
	digest := sha1.Sum([]byte(content))
	hash := base64.RawURLEncoding.EncodeToString(digest[:])

	return hash, nil
}
