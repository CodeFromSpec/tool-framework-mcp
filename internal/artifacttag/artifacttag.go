// code-from-spec: ROOT/golang/implementation/parsing/artifact_tag@mBc8vndsHL5_kjg9v3qeTDDhdbk
package artifacttag

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ArtifactTag holds the parsed contents of a code-from-spec tag
// found in a generated source file.
//
// The tag format is:
//
//	code-from-spec: <logical-name>@<hash>
//
// Example tag (inside a comment):
//
//	// code-from-spec: ROOT/golang/interfaces/parsing/artifact_tag@g1QvX6BKXj44-GjQn0opXX3yXbE
type ArtifactTag struct {
	LogicalName string
	Hash        string
}

var (
	// ErrFileUnreadable is returned when the file cannot be opened or read.
	ErrFileUnreadable = errors.New("file unreadable")

	// ErrNoTagFound is returned when the file contains no code-from-spec: substring.
	ErrNoTagFound = errors.New("no tag found")

	// ErrMalformedTag is returned when a code-from-spec: substring is found
	// but cannot be fully parsed (missing "@", empty logical name, or wrong
	// hash length).
	ErrMalformedTag = errors.New("malformed tag")
)

const tagPrefix = "code-from-spec: "

// ArtifactTagExtract opens the file at file_path, scans each line for the
// pattern "code-from-spec: <logical-name>@<hash>", and returns the parsed
// tag on the first match.
//
// The tag may appear inside any comment syntax (//, #, /* */, --, <!-- -->).
// Comment syntax is not parsed — the function scans raw lines for the
// substring "code-from-spec:".
//
// Returns:
//   - (*ArtifactTag, nil) on success.
//   - (nil, ErrFileUnreadable) if the file cannot be opened or read.
//   - (nil, ErrNoTagFound) if no line contains "code-from-spec:".
//   - (nil, ErrMalformedTag) if the tag exists but cannot be parsed
//     (no "@" separator, empty logical name, or wrong hash length).
//
// Path errors from opening the file are propagated directly.
func ArtifactTagExtract(file_path *pathutils.PathCfs) (*ArtifactTag, error) {
	// Step 1: Open the file.
	reader, err := filereader.FileOpen(file_path)
	if err != nil {
		if errors.Is(err, filereader.ErrFileUnreadable) {
			return nil, fmt.Errorf("%w", ErrFileUnreadable)
		}
		// Propagate path errors directly.
		return nil, err
	}

	// Step 2: found_line starts as empty (no value yet).
	foundLine := ""
	found := false

	// Step 3: Read lines until we find the tag or reach end of file.
	for {
		line, err := filereader.FileReadLine(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			break
		}
		if err != nil {
			filereader.FileClose(reader)
			return nil, fmt.Errorf("%w", ErrFileUnreadable)
		}
		if strings.Contains(line, tagPrefix) {
			foundLine = line
			found = true
			break
		}
	}

	// Step 4: Always close the reader.
	filereader.FileClose(reader)

	// Step 5: If no matching line was found, return ErrNoTagFound.
	if !found {
		return nil, fmt.Errorf("%w", ErrNoTagFound)
	}

	// Step 6: Extract the tag value from foundLine.
	idx := strings.Index(foundLine, tagPrefix)
	tagValue := foundLine[idx+len(tagPrefix):]
	tagValue = strings.TrimLeft(tagValue, " \t")

	// Step 7: Find the first "@" in tagValue.
	atIdx := strings.Index(tagValue, "@")
	if atIdx < 0 {
		return nil, fmt.Errorf("%w", ErrMalformedTag)
	}

	// Step 8: Everything before "@" is the logical name.
	logicalName := tagValue[:atIdx]
	if logicalName == "" {
		return nil, fmt.Errorf("%w", ErrMalformedTag)
	}

	// Step 9: Everything after "@" must be at least 27 characters.
	afterAt := tagValue[atIdx+1:]
	if len(afterAt) < 27 {
		return nil, fmt.Errorf("%w", ErrMalformedTag)
	}

	// Step 10: Hash is exactly the first 27 characters after "@".
	hash := afterAt[:27]

	// Step 11: Return the parsed tag.
	return &ArtifactTag{
		LogicalName: logicalName,
		Hash:        hash,
	}, nil
}
