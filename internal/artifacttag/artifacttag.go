// code-from-spec: ROOT/golang/implementation/parsing/artifact_tag@d1f2fUv4GKKv0CGMFdpixT5Taig

// Package artifacttag provides functionality for extracting the
// code-from-spec artifact tag from generated source files. The tag
// encodes the logical name and hash of the spec that produced the file.
package artifacttag

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ArtifactTag holds the parsed components of a code-from-spec tag
// found in a generated source file.
//
// Tag format:
//
//	code-from-spec: <logical-name>@<hash>
//
// The tag may appear inside any comment syntax (//, #, /* */, --, <!-- -->).
// Parsing is line-based and does not interpret comment delimiters.
type ArtifactTag struct {
	// LogicalName is the logical node name extracted from the tag,
	// for example "ROOT/golang/interfaces/parsing/artifact_tag".
	LogicalName string

	// Hash is the chain hash extracted from the tag,
	// for example "Na2fdUmffqbI_YdC0liSgTl_-fQ".
	Hash string
}

// ErrFileUnreadable is returned when the file cannot be opened or read.
var ErrFileUnreadable = errors.New("file unreadable")

// ErrNoTagFound is returned when the file contains no code-from-spec: substring.
var ErrNoTagFound = errors.New("no tag found")

// ErrMalformedTag is returned when the tag exists but cannot be parsed
// (missing @, empty logical name, or wrong hash length).
var ErrMalformedTag = errors.New("malformed tag")

const tagPrefix = "code-from-spec: "
const hashLength = 27

// ArtifactTagExtract scans the file at filePath line by line for the
// first occurrence of the substring "code-from-spec:" and parses the
// logical name and hash from it.
//
// The tag format is:
//
//	code-from-spec: <logical-name>@<hash>
//
// Parsing is purely textual — comment delimiters are ignored.
//
// Errors:
//   - ErrFileUnreadable: the file cannot be opened or read.
//   - ErrNoTagFound: no "code-from-spec:" substring was found in the file.
//   - ErrMalformedTag: the tag was found but could not be parsed
//     (e.g. no @ separator, empty logical name, or wrong hash length).
//   - (FileReader.*): propagated from FileOpen.
func ArtifactTagExtract(filePath *pathutils.PathCfs) (*ArtifactTag, error) {
	// Step 1: Open the file for reading.
	reader, err := filereader.FileOpen(filePath)
	if err != nil {
		if errors.Is(err, filereader.ErrFileUnreadable) {
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		return nil, fmt.Errorf("opening file: %w", err)
	}

	// Steps 2–4: Scan lines until a match or EOF.
	foundLine := ""
	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				break
			}
			filereader.FileClose(reader)
			return nil, fmt.Errorf("%w: reading line: %w", ErrFileUnreadable, err)
		}
		if strings.Contains(line, tagPrefix) {
			foundLine = line
			break
		}
	}

	filereader.FileClose(reader)

	// Step 5: Check if tag was found.
	if foundLine == "" {
		return nil, fmt.Errorf("%w: the file has no code-from-spec: tag", ErrNoTagFound)
	}

	// Step 6: Extract the raw tag substring after the prefix.
	idx := strings.Index(foundLine, tagPrefix)
	rawTag := foundLine[idx+len(tagPrefix):]

	// Step 7: Trim leading whitespace.
	rawTag = strings.TrimLeft(rawTag, " \t")

	// Step 8: Find the first "@".
	atIdx := strings.Index(rawTag, "@")
	if atIdx == -1 {
		return nil, fmt.Errorf("%w: no @ separator found in tag", ErrMalformedTag)
	}

	// Step 9: Extract logical name.
	logicalName := rawTag[:atIdx]
	if logicalName == "" {
		return nil, fmt.Errorf("%w: logical name is empty", ErrMalformedTag)
	}

	// Step 10: Extract the hash candidate.
	hashCandidate := rawTag[atIdx+1:]
	if len(hashCandidate) < hashLength {
		return nil, fmt.Errorf("%w: hash must be at least 27 characters", ErrMalformedTag)
	}

	// Step 11: Take the first 27 characters as the hash.
	hash := hashCandidate[:hashLength]

	// Step 12: Return the result.
	return &ArtifactTag{
		LogicalName: logicalName,
		Hash:        hash,
	}, nil
}
