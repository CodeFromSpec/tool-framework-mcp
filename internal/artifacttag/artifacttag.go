// code-from-spec: ROOT/golang/implementation/parsing/artifact_tag@5xwnlLy2HPiM_vLnyd1P9Zmb15U

package artifacttag

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

// ArtifactTag holds the parsed contents of a code-from-spec tag
// found inside a generated file.
//
// The tag format is:
//
//	code-from-spec: <logical-name>@<hash>
//
// Example tag line:
//
//	// code-from-spec: ROOT/golang/interfaces/parsing/artifact_tag@axzN9uSE7z30JhEmKNZBEH3ekfg
type ArtifactTag struct {
	// LogicalName is the logical name component of the tag (e.g.
	// "ROOT/golang/interfaces/parsing/artifact_tag").
	LogicalName string

	// Hash is the hash component of the tag (e.g.
	// "axzN9uSE7z30JhEmKNZBEH3ekfg").
	Hash string
}

var (
	// ErrNoTagFound is returned when the file contains no
	// "code-from-spec:" substring on any line.
	ErrNoTagFound = errors.New("no tag found")

	// ErrMalformedTag is returned when the tag string is present but
	// cannot be parsed — for example, the "@" separator is missing,
	// the logical name is empty, or the hash has the wrong length.
	ErrMalformedTag = errors.New("malformed tag")
)

const tagPrefix = "code-from-spec: "

// ArtifactTagExtract opens the file at file_path, scans its lines for
// the first occurrence of the "code-from-spec:" pattern, and returns
// the parsed ArtifactTag.
//
// The tag may appear inside any comment syntax (//, #, /* */, --, <!-- -->).
// Comment syntax is not parsed — every line is scanned for the substring
// "code-from-spec:" regardless of context.
//
// Possible errors:
//   - Path errors propagated from opening the file (e.g. pathutils.ErrPathEmpty,
//     pathutils.ErrPathAbsolute, pathutils.ErrPathContainsBackslash,
//     pathutils.ErrDirectoryTraversal, pathutils.ErrResolvesOutsideRoot,
//     pathutils.ErrCannotDetermineRoot).
//   - ErrNoTagFound — the file was read successfully but contains no tag.
//   - ErrMalformedTag — a tag line was found but could not be parsed.
func ArtifactTagExtract(file_path *pathutils.PathCfs) (*ArtifactTag, error) {
	// Step 1: Open the file.
	reader, err := filereader.FileOpen(file_path)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	// Step 2: Initialize tracking variables.
	foundLine := ""
	var readError error

	// Step 3: Scan lines for the tag prefix.
	for {
		line, err := filereader.FileReadLine(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			break
		}
		if err != nil {
			readError = err
			break
		}
		if strings.Contains(line, tagPrefix) {
			foundLine = line
			break
		}
	}

	// Step 4: Close the reader.
	filereader.FileClose(reader)

	// Step 5: Propagate read errors.
	if readError != nil {
		return nil, fmt.Errorf("file unreadable: %w", readError)
	}

	// Step 6: Check if a tag line was found.
	if foundLine == "" {
		return nil, fmt.Errorf("%w", ErrNoTagFound)
	}

	// Step 7: Extract the raw tag value from the line.
	idx := strings.Index(foundLine, tagPrefix)
	rawTag := foundLine[idx+len(tagPrefix):]
	rawTag = strings.TrimLeft(rawTag, " \t")

	// Step 8: Find the "@" separator.
	atIdx := strings.Index(rawTag, "@")
	if atIdx == -1 {
		return nil, fmt.Errorf("%w", ErrMalformedTag)
	}

	// Step 9: Extract the logical name.
	logicalName := rawTag[:atIdx]
	if logicalName == "" {
		return nil, fmt.Errorf("%w", ErrMalformedTag)
	}

	// Step 10: Extract the remainder after "@".
	remainder := rawTag[atIdx+1:]
	if len(remainder) < 27 {
		return nil, fmt.Errorf("%w", ErrMalformedTag)
	}

	// Step 11: Take the first 27 characters as the hash.
	hash := remainder[:27]

	// Step 12: Return the populated ArtifactTag.
	return &ArtifactTag{
		LogicalName: logicalName,
		Hash:        hash,
	}, nil
}
