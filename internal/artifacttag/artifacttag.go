// code-from-spec: ROOT/golang/implementation/parsing/artifact_tag@MPUokAIeJZnNLWxjOMSksNs5azg

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
		if errors.Is(err, filereader.ErrFileUnreadable) {
			return nil, fmt.Errorf("%w", filereader.ErrFileUnreadable)
		}
		return nil, fmt.Errorf("opening file: %w", err)
	}

	// Step 2: Set found_line to empty (no match yet).
	foundLine := ""

	// Step 3: Loop over lines.
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

	// Step 6: Take the portion after the first occurrence of the prefix.
	idx := strings.Index(foundLine, tagPrefix)
	rawTag := foundLine[idx+len(tagPrefix):]

	// Step 7: Trim leading whitespace from rawTag.
	rawTag = strings.TrimLeft(rawTag, " \t")

	// Step 8: Find the first "@" in rawTag.
	atIdx := strings.Index(rawTag, "@")
	if atIdx < 0 {
		return nil, fmt.Errorf("%w", ErrMalformedTag)
	}

	// Step 9: Extract the logical name (before "@").
	logicalName := rawTag[:atIdx]
	if logicalName == "" {
		return nil, fmt.Errorf("%w", ErrMalformedTag)
	}

	// Step 10: Extract the hash as exactly 27 characters after "@".
	afterAt := rawTag[atIdx+1:]
	if len(afterAt) < 27 {
		return nil, fmt.Errorf("%w", ErrMalformedTag)
	}
	hash := afterAt[:27]

	// Step 11: Return the ArtifactTag.
	return &ArtifactTag{
		LogicalName: logicalName,
		Hash:        hash,
	}, nil
}
