// code-from-spec: ROOT/golang/internal/artifact_tag/code@d8JPsz2xvc5a2kzp7f_fTNsQvJA

// Package artifacttag provides utilities for locating and parsing the
// "code-from-spec" tag embedded in generated files.
//
// The tag format is:
//
//	code-from-spec: <logical-name>@<hash>
//
// where <logical-name> is the spec node name (e.g. "ROOT/golang/server")
// and <hash> is exactly 27 characters (a base64url chain hash).
//
// The scan is purely substring-based — comment syntax is irrelevant.
// Only the first matching line in the file is used.
package artifacttag

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filereader"
)

// tagPrefix is the fixed substring we search for in each line.
const tagPrefix = "code-from-spec: "

// hashLength is the required exact length of the hash portion.
const hashLength = 27

// ArtifactTag holds the parsed contents of a code-from-spec tag.
type ArtifactTag struct {
	// LogicalName is the spec node path, e.g. "ROOT/golang/server".
	LogicalName string
	// Hash is the 27-character base64url chain hash.
	Hash string
}

// Sentinel errors — callers may match these with errors.Is().
var (
	// ErrFileUnreadable is returned when the file cannot be opened or read.
	ErrFileUnreadable = errors.New("file unreadable")
	// ErrNoTagFound is returned when no "code-from-spec:" substring appears in the file.
	ErrNoTagFound = errors.New("no tag found")
	// ErrMalformedTag is returned when the tag is found but cannot be parsed
	// (missing "@", empty logical name, or hash length != 27).
	ErrMalformedTag = errors.New("malformed tag")
)

// ExtractArtifactTag opens the file at filePath and scans it line by line
// for the first occurrence of the "code-from-spec: <name>@<hash>" pattern.
//
// On success it returns a populated *ArtifactTag.
// On failure it returns one of ErrFileUnreadable, ErrNoTagFound, or
// ErrMalformedTag (all wrapped so errors.Is() works).
func ExtractArtifactTag(filePath string) (*ArtifactTag, error) {
	// Step 1 — Open the file. Any open failure is reported as ErrFileUnreadable.
	r, err := filereader.OpenFileReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrFileUnreadable, err)
	}

	// Step 2 — Read lines one by one, looking for the tag prefix.
	for {
		line, err := r.ReadLine()
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				// Step 3 — Exhausted all lines with no match.
				return nil, fmt.Errorf("%w: %s", ErrNoTagFound, filePath)
			}
			// Any other read error is treated as file unreadable.
			return nil, fmt.Errorf("%w: %s", ErrFileUnreadable, err)
		}

		// Step 2a — Check whether this line contains the tag prefix.
		idx := strings.Index(line, tagPrefix)
		if idx == -1 {
			// No match on this line; move on.
			continue
		}

		// Step 2b — Extract the raw value: everything after the prefix,
		// with trailing whitespace trimmed.
		rawValue := strings.TrimRight(line[idx+len(tagPrefix):], " \t\r\n")

		// Step 4 — Parse the raw value.
		return parseRawValue(rawValue)
	}
}

// parseRawValue splits rawValue on the last "@" and validates the parts.
//
// It returns ErrMalformedTag (wrapped) for any of these conditions:
//   - no "@" present
//   - logical name (part before last "@") is empty
//   - hash (part after last "@") is not exactly 27 characters
func parseRawValue(rawValue string) (*ArtifactTag, error) {
	// Step 4a — Find the last "@".
	// Using the last occurrence allows logical names that themselves contain "@".
	lastAt := strings.LastIndex(rawValue, "@")
	if lastAt == -1 {
		return nil, fmt.Errorf("%w: no '@' separator in %q", ErrMalformedTag, rawValue)
	}

	// Step 4b — Split on the last "@".
	logicalName := rawValue[:lastAt]
	hash := rawValue[lastAt+1:]

	// Step 4c — Logical name must not be empty.
	if logicalName == "" {
		return nil, fmt.Errorf("%w: empty logical name in %q", ErrMalformedTag, rawValue)
	}

	// Step 4d — Hash must be exactly 27 characters.
	if len(hash) != hashLength {
		return nil, fmt.Errorf("%w: hash length %d (want %d) in %q",
			ErrMalformedTag, len(hash), hashLength, rawValue)
	}

	// Step 5 — Return the populated record.
	return &ArtifactTag{
		LogicalName: logicalName,
		Hash:        hash,
	}, nil
}
