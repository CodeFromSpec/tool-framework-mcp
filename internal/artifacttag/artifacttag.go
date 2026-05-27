// code-from-spec: ROOT/golang/internal/artifact_tag/code@IOf0-OmoNz5aWow6IBZ9VjS0x_E
package artifacttag

import (
	"errors"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filereader"
)

// ErrFileUnreadable is returned when the target file cannot be opened or read.
var ErrFileUnreadable = errors.New("file unreadable")

// ErrNoTagFound is returned when no artifact tag line is present in the file.
var ErrNoTagFound = errors.New("no artifact tag found")

// ErrMalformedTag is returned when a tag line is found but its structure is invalid.
var ErrMalformedTag = errors.New("malformed artifact tag")

const (
	tagPrefix  = "code-from-spec: "
	hashLength = 27
)

// ArtifactTag holds the parsed contents of a code-from-spec tag.
type ArtifactTag struct {
	// LogicalName is the part before the first '@' in the tag value.
	LogicalName string
	// Hash is the 27-character hash that follows the first '@'.
	Hash string
}

// ExtractArtifactTag scans filePath line by line and returns the first
// artifact tag it finds. The tag format is:
//
//	code-from-spec: <logical-name>@<27-char-hash>
//
// The function uses the FIRST '@' after "code-from-spec: " to split the
// logical name from the hash, and takes exactly the first 27 characters
// after that '@' as the hash value, ignoring any trailing comment syntax.
//
// Errors:
//   - ErrFileUnreadable  – the file could not be opened or read.
//   - ErrNoTagFound      – no tag line was found in the file.
//   - ErrMalformedTag    – a tag line was found but is structurally invalid
//     (missing '@' separator or hash shorter than 27 characters).
func ExtractArtifactTag(filePath string) (*ArtifactTag, error) {
	reader, err := filereader.OpenFileReader(filePath)
	if err != nil {
		return nil, ErrFileUnreadable
	}
	defer reader.Close()

	for {
		line, err := reader.ReadLine()
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				return nil, ErrNoTagFound
			}
			return nil, ErrFileUnreadable
		}

		tag, parseErr := extractFromLine(line)
		if parseErr == nil {
			return tag, nil
		}
		if errors.Is(parseErr, ErrMalformedTag) {
			return nil, parseErr
		}
		// parseErr == errNoTagOnLine — continue scanning
	}
}

// errNoTagOnLine is an internal sentinel used only within this package
// to signal that the current line does not contain a tag prefix.
var errNoTagOnLine = errors.New("no tag on this line")

// extractFromLine looks for the tag prefix in a single line and parses it.
// Returns errNoTagOnLine when the prefix is absent, ErrMalformedTag when the
// prefix is present but the structure is invalid.
func extractFromLine(line string) (*ArtifactTag, error) {
	idx := strings.Index(line, tagPrefix)
	if idx == -1 {
		return nil, errNoTagOnLine
	}

	// Everything after the prefix is the tag value.
	value := line[idx+len(tagPrefix):]

	// Find the FIRST '@' in the value to split logical name from hash.
	atIdx := strings.Index(value, "@")
	if atIdx == -1 {
		return nil, ErrMalformedTag
	}

	logicalName := value[:atIdx]
	remainder := value[atIdx+1:]

	if len(remainder) < hashLength {
		return nil, ErrMalformedTag
	}

	hash := remainder[:hashLength]

	return &ArtifactTag{
		LogicalName: logicalName,
		Hash:        hash,
	}, nil
}
