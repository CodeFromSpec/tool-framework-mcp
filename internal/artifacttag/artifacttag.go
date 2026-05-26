// code-from-spec: ROOT/golang/internal/artifact_tag/code@PENDING
package artifacttag

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filereader"
)

// ArtifactTag holds the logical name and hash extracted from a generated file.
type ArtifactTag struct {
	LogicalName string
	Hash        string
}

// Sentinel errors for ExtractArtifactTag.
var (
	ErrFileUnreadable = errors.New("file unreadable")
	ErrNoTagFound     = errors.New("no tag found")
	ErrMalformedTag   = errors.New("malformed tag")
)

const tagMarker = "code-from-spec: "

// ExtractArtifactTag opens a file and scans line by line for the
// "code-from-spec: <logical-name>@<hash>" pattern. Returns the first match.
func ExtractArtifactTag(filePath string) (*ArtifactTag, error) {
	reader, err := filereader.OpenFileReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrFileUnreadable, err)
	}

	for {
		line, err := reader.ReadLine()
		if errors.Is(err, filereader.ErrEndOfFile) {
			break
		}

		idx := strings.Index(line, tagMarker)
		if idx < 0 {
			continue
		}

		// Extract tag value: everything after the marker, trimmed.
		tagValue := strings.TrimRight(line[idx+len(tagMarker):], " \t\r\n")

		// Find the last "@" to split logical name and hash.
		atIdx := strings.LastIndex(tagValue, "@")
		if atIdx < 0 {
			return nil, fmt.Errorf("%w: no @ separator in tag", ErrMalformedTag)
		}

		logicalName := tagValue[:atIdx]
		rawHash := tagValue[atIdx+1:]

		// The hash is exactly 27 base64url characters. Anything after
		// (e.g., comment closing syntax like "-->") is ignored.
		hash := extractBase64URL(rawHash)

		if logicalName == "" {
			return nil, fmt.Errorf("%w: empty logical name", ErrMalformedTag)
		}

		if len(hash) != 27 {
			return nil, fmt.Errorf("%w: hash length is %d, expected 27", ErrMalformedTag, len(hash))
		}

		return &ArtifactTag{
			LogicalName: logicalName,
			Hash:        hash,
		}, nil
	}

	return nil, fmt.Errorf("%w", ErrNoTagFound)
}

// extractBase64URL extracts the leading run of base64url characters
// (A-Z, a-z, 0-9, -, _) from a string, stopping at the first
// character that is not part of the base64url alphabet.
func extractBase64URL(s string) string {
	for i, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_') {
			return s[:i]
		}
	}
	return s
}
