// code-from-spec: ROOT/golang/internal/artifact_tag/code@wnJl_i9eb3zBB_k3Ezvw_c4Xs54

// Package artifacttag locates and extracts the "code-from-spec: <logical-name>@<hash>"
// tag from a generated source file. The scanner searches every line for the
// substring — it does not interpret comment delimiters, so the tag may appear
// inside any comment syntax (// # /* */ -- etc.).
package artifacttag

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filereader"
)

// hashLength is the exact number of characters expected in a valid hash.
// All hashes are base64url-encoded and always 27 characters long.
const hashLength = 27

// tagPrefix is the substring that every artifact tag line must contain.
const tagPrefix = "code-from-spec: "

// ArtifactTag holds the two components of a parsed artifact tag.
type ArtifactTag struct {
	// LogicalName is the spec node path that produced the file (e.g. "ROOT/golang/internal/artifact_tag/code").
	LogicalName string
	// Hash is the 27-character chain hash embedded in the tag.
	Hash string
}

// Sentinel errors — callers should use errors.Is() to inspect returned errors.
var (
	// ErrFileUnreadable is returned when the file cannot be opened or a read
	// error occurs while scanning it.
	ErrFileUnreadable = errors.New("file unreadable")

	// ErrNoTagFound is returned when the file was fully read and no line
	// contained the "code-from-spec: " substring.
	ErrNoTagFound = errors.New("no tag found")

	// ErrMalformedTag is returned when a tag line is found but its value does
	// not satisfy the format rules (missing "@", empty logical name, or hash
	// length != 27).
	ErrMalformedTag = errors.New("malformed tag")
)

// ExtractArtifactTag opens the file at filePath, scans it line by line, and
// returns the first artifact tag found.
//
// The function always closes the file before returning, regardless of outcome.
//
// Error cases:
//   - ErrFileUnreadable — file cannot be opened or a read error occurs.
//   - ErrNoTagFound     — no line in the file contains "code-from-spec: ".
//   - ErrMalformedTag   — a tag line was found but could not be parsed.
func ExtractArtifactTag(filePath string) (*ArtifactTag, error) {
	// Step 1: Open the file for sequential line-by-line reading.
	reader, err := filereader.OpenFileReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	// Steps 2-4: Scan lines until we find the tag or reach end-of-file.
	tagLine := ""
	found := false

	for {
		// Step 3a: Read the next line.
		line, readErr := reader.ReadLine()
		if readErr != nil {
			if errors.Is(readErr, filereader.ErrEndOfFile) {
				// Normal end of file — exit the scan loop.
				break
			}
			// Unexpected read error — close before returning.
			reader.Close()
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, readErr)
		}

		// Step 3b: Check whether this line carries the artifact tag.
		if strings.Contains(line, tagPrefix) {
			tagLine = line
			found = true
			// Stop at the first match; do not read the rest of the file.
			break
		}
	}

	// Step 4: Always close the reader before returning.
	reader.Close()

	// Step 5: Report if no tag line was found at all.
	if !found {
		return nil, ErrNoTagFound
	}

	// Step 6: Extract the raw value that follows the prefix.
	//   a. Find the position of the prefix in the line.
	prefixIdx := strings.Index(tagLine, tagPrefix)
	//   b. Take the substring immediately after the prefix.
	raw := tagLine[prefixIdx+len(tagPrefix):]
	//   c. Trim trailing whitespace (handles CRLF remnants or trailing spaces).
	raw = strings.TrimRight(raw, " \t\r\n")

	// Step 7: Locate the last "@" that separates the logical name from the hash.
	atIdx := strings.LastIndex(raw, "@")
	if atIdx == -1 {
		return nil, fmt.Errorf("%w: no '@' separator in %q", ErrMalformedTag, raw)
	}

	// Step 8: Split at the last "@".
	logicalName := raw[:atIdx]
	hash := raw[atIdx+1:]

	// Step 9: Validate the two components.
	if logicalName == "" {
		return nil, fmt.Errorf("%w: logical name is empty in %q", ErrMalformedTag, raw)
	}
	if len(hash) != hashLength {
		return nil, fmt.Errorf("%w: hash %q has length %d, want %d", ErrMalformedTag, hash, len(hash), hashLength)
	}

	// Step 10: Return the populated record.
	return &ArtifactTag{
		LogicalName: logicalName,
		Hash:        hash,
	}, nil
}
