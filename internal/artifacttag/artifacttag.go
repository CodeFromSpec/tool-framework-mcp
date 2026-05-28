// code-from-spec: ROOT/golang/implementation/internal/artifact_tag/code@s_hBcQQP7MLOcVF_4t1gMU7S7-U

package artifacttag

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

// ArtifactTag holds the parsed logical name and hash from a code-from-spec tag.
type ArtifactTag struct {
	LogicalName string
	Hash        string
}

var (
	// ErrNoTagFound is returned when the file contains no "code-from-spec: " substring.
	ErrNoTagFound = errors.New("no tag found")

	// ErrMalformedTag is returned when the tag exists but cannot be parsed.
	ErrMalformedTag = errors.New("malformed tag")
)

const tagPrefix = "code-from-spec: "
const hashLength = 27

// ArtifactTagExtract scans file_path line by line and returns the first
// ArtifactTag found. The tag must be in the form:
//
//	code-from-spec: <logical_name>@<27-character-hash>
//
// Possible errors:
//   - Path errors propagated from FileOpen
//   - filereader.ErrFileUnreadable if the file cannot be opened
//   - ErrNoTagFound if no matching line is found
//   - ErrMalformedTag if the tag line cannot be parsed
func ArtifactTagExtract(filePath *pathutils.PathCfs) (*ArtifactTag, error) {
	// Step 1: open the file.
	reader, err := filereader.FileOpen(filePath)
	if err != nil {
		return nil, fmt.Errorf("ArtifactTagExtract: %w", err)
	}

	// Step 2: found_line starts empty.
	foundLine := ""

	// Step 3: scan lines until a match or EOF.
	for {
		line, err := filereader.FileReadLine(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			break
		}
		if err != nil {
			filereader.FileClose(reader)
			return nil, fmt.Errorf("ArtifactTagExtract: %w", err)
		}
		if strings.Contains(line, tagPrefix) {
			foundLine = line
			break
		}
	}

	// Step 4: close the reader.
	filereader.FileClose(reader)

	// Step 5: no match found.
	if foundLine == "" {
		return nil, fmt.Errorf("ArtifactTagExtract: %w", ErrNoTagFound)
	}

	// Step 6: extract the portion after "code-from-spec: ".
	idx := strings.Index(foundLine, tagPrefix)
	rawTag := foundLine[idx+len(tagPrefix):]

	// Step 7: trim leading whitespace.
	rawTag = strings.TrimLeft(rawTag, " \t")

	// Step 8: find the first "@".
	atIdx := strings.Index(rawTag, "@")
	if atIdx == -1 {
		return nil, fmt.Errorf("ArtifactTagExtract: %w", ErrMalformedTag)
	}

	// Step 9: extract logical name (before "@").
	logicalName := rawTag[:atIdx]
	if logicalName == "" {
		return nil, fmt.Errorf("ArtifactTagExtract: %w", ErrMalformedTag)
	}

	// Step 10: extract the 27-character hash after "@".
	afterAt := rawTag[atIdx+1:]
	if len(afterAt) < hashLength {
		return nil, fmt.Errorf("ArtifactTagExtract: %w", ErrMalformedTag)
	}
	hash := afterAt[:hashLength]

	// Step 11: return the parsed tag.
	return &ArtifactTag{
		LogicalName: logicalName,
		Hash:        hash,
	}, nil
}
