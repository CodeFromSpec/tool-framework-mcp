// code-from-spec: ROOT/golang/implementation/parsing/artifact_tag@NhcKg9teavpiFA3aHe-hus7vtBQ

package artifacttag

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

var ErrFileUnreadable = errors.New("file cannot be opened or read")
var ErrNoTagFound = errors.New("file has no code-from-spec: substring")
var ErrMalformedTag = errors.New("tag exists but cannot be parsed")

type ArtifactTag struct {
	LogicalName string
	Hash        string
}

const tagPrefix = "code-from-spec: "

func ArtifactTagExtract(filePath *pathutils.PathCfs) (*ArtifactTag, error) {
	reader, err := filereader.FileOpen(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	var matchedLine string
	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				filereader.FileClose(reader)
				return nil, ErrNoTagFound
			}
			filereader.FileClose(reader)
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		if strings.Contains(line, tagPrefix) {
			matchedLine = line
			filereader.FileClose(reader)
			break
		}
	}

	idx := strings.Index(matchedLine, tagPrefix)
	portion := matchedLine[idx+len(tagPrefix):]
	portion = strings.TrimLeft(portion, " \t")

	atIdx := strings.Index(portion, "@")
	if atIdx < 0 {
		return nil, ErrMalformedTag
	}

	logicalName := portion[:atIdx]
	if logicalName == "" {
		return nil, ErrMalformedTag
	}

	afterAt := portion[atIdx+1:]
	if len(afterAt) < 27 {
		return nil, ErrMalformedTag
	}
	hash := afterAt[:27]

	return &ArtifactTag{
		LogicalName: logicalName,
		Hash:        hash,
	}, nil
}
