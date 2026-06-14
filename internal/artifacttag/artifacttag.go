// code-from-spec: ROOT/golang/implementation/parsing/artifact_tag@_Vp8KQedKa6loC8WBZr2ZT7VhWY
package artifacttag

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

var ErrFileUnreadable = errors.New("file cannot be opened or read")
var ErrMalformedTag   = errors.New("tag exists but cannot be parsed")
var ErrNoTagFound     = errors.New("no artifact tag found in file")

const tagPrefix = "code-from-spec: "
const hashLength = 27

type ArtifactTag struct {
	LogicalName string
	Hash        string
}

func ArtifactTagExtract(filePath *pathutils.PathCfs) (*ArtifactTag, error) {
	reader, err := filereader.FileOpen(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	foundLine := ""
	done := false

	for !done {
		line, err := filereader.FileReadLine(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			done = true
			break
		}
		if err != nil {
			filereader.FileClose(reader)
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}

		if strings.Contains(line, tagPrefix) {
			foundLine = line
			done = true
		}
	}

	filereader.FileClose(reader)

	if foundLine == "" {
		return nil, ErrNoTagFound
	}

	idx := strings.Index(foundLine, tagPrefix)
	rawTag := strings.TrimLeft(foundLine[idx+len(tagPrefix):], " \t")

	atIdx := strings.Index(rawTag, "@")
	if atIdx == -1 {
		return nil, fmt.Errorf("%w: missing @ separator", ErrMalformedTag)
	}

	logicalName := rawTag[:atIdx]
	if logicalName == "" {
		return nil, fmt.Errorf("%w: empty logical name", ErrMalformedTag)
	}

	hashCandidate := rawTag[atIdx+1:]
	if len(hashCandidate) < hashLength {
		return nil, fmt.Errorf("%w: hash too short", ErrMalformedTag)
	}

	hash := hashCandidate[:hashLength]

	return &ArtifactTag{
		LogicalName: logicalName,
		Hash:        hash,
	}, nil
}
