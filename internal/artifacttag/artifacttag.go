// code-from-spec: ROOT/golang/implementation/parsing/artifact_tag@zcpELs5g-jysroUXSfhjouQPhDc

package artifacttag

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ArtifactTag holds the parsed components of a code-from-spec tag found in a file.
//
// The tag has the format:
//
//	code-from-spec: <logical-name>@<hash>
//
// It may appear inside any comment syntax (//, #, /* */, --, <!-- -->).
type ArtifactTag struct {
	LogicalName string
	Hash        string
}

// ErrFileUnreadable is returned when the file cannot be opened or read.
var ErrFileUnreadable = errors.New("file unreadable")

// ErrNoTagFound is returned when the file contains no code-from-spec: substring.
var ErrNoTagFound = errors.New("no tag found")

// ErrMalformedTag is returned when a code-from-spec: tag exists but cannot be
// parsed — missing @, empty logical name, or wrong hash length.
var ErrMalformedTag = errors.New("malformed tag")

const tagPrefix = "code-from-spec: "

// ArtifactTagExtract opens the file at filePath, scans each line for the
// code-from-spec: pattern, and returns the parsed ArtifactTag.
//
// The tag format is:
//
//	code-from-spec: <logical-name>@<hash>
//
// Comment syntax is ignored — any line containing the substring is considered.
//
// Errors:
//   - ErrFileUnreadable: the file cannot be opened or read.
//   - ErrNoTagFound: the file has no code-from-spec: substring.
//   - ErrMalformedTag: the tag exists but cannot be parsed (no @, empty name,
//     wrong hash length).
//   - (FileReader.*): propagated from FileOpen.
func ArtifactTagExtract(filePath *pathutils.PathCfs) (*ArtifactTag, error) {
	// Step 1: Open the file.
	reader, err := filereader.FileOpen(filePath)
	if err != nil {
		if errors.Is(err, filereader.ErrFileUnreadable) {
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		return nil, fmt.Errorf("opening file: %w", err)
	}

	// Step 2: Set found_line to empty.
	foundLine := ""

	// Step 3: Loop reading lines until EOF or match.
	for {
		line, err := filereader.FileReadLine(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			break
		}
		if err != nil {
			filereader.FileClose(reader)
			return nil, fmt.Errorf("reading file: %w", err)
		}
		if strings.Contains(line, tagPrefix) {
			foundLine = line
			break
		}
	}

	// Step 4: Close the reader.
	filereader.FileClose(reader)

	// Step 5: If no match was found, return ErrNoTagFound.
	if foundLine == "" {
		return nil, fmt.Errorf("%w", ErrNoTagFound)
	}

	// Step 6: Extract the content after the prefix and trim leading whitespace.
	idx := strings.Index(foundLine, tagPrefix)
	tagContent := strings.TrimLeft(foundLine[idx+len(tagPrefix):], " \t")

	// Step 7: Find the first '@'.
	atIdx := strings.Index(tagContent, "@")
	if atIdx == -1 {
		return nil, fmt.Errorf("%w: tag missing '@' separator", ErrMalformedTag)
	}

	// Step 8: Extract logical name.
	logicalName := tagContent[:atIdx]
	if logicalName == "" {
		return nil, fmt.Errorf("%w: logical name is empty", ErrMalformedTag)
	}

	// Step 9: Extract hash candidate.
	hashCandidate := tagContent[atIdx+1:]
	if len(hashCandidate) < 27 {
		return nil, fmt.Errorf("%w: hash must be at least 27 characters", ErrMalformedTag)
	}

	// Step 10: Take the first 27 characters as the hash.
	hash := hashCandidate[:27]

	// Step 11: Return the parsed tag.
	return &ArtifactTag{
		LogicalName: logicalName,
		Hash:        hash,
	}, nil
}
