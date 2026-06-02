// code-from-spec: ROOT/golang/implementation/parsing/artifact_tag@HVkKJ-4EvWuVLelTQy7EJKo5Wjg
package artifacttag

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

var ErrFileUnreadable = errors.New("file cannot be opened or read")
var ErrNoTagFound = errors.New("no code-from-spec tag found")
var ErrMalformedTag = errors.New("tag is malformed")

type ArtifactTag struct {
	LogicalName string
	Hash        string
}

const tagPrefix = "code-from-spec: "

func ArtifactTagExtract(file_path *pathutils.PathCfs) (*ArtifactTag, error) {
	reader, err := filereader.FileOpen(file_path)
	if err != nil {
		if errors.Is(err, filereader.ErrFileUnreadable) {
			return nil, fmt.Errorf("%w: %s", ErrFileUnreadable, err)
		}
		return nil, err
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
			return nil, fmt.Errorf("%w: %s", ErrFileUnreadable, err)
		}

		idx := strings.Index(line, tagPrefix)
		if idx >= 0 {
			matchedLine = line[idx+len(tagPrefix):]
			filereader.FileClose(reader)
			break
		}
	}

	portion := strings.TrimLeft(matchedLine, " \t")

	atIdx := strings.Index(portion, "@")
	if atIdx < 0 {
		return nil, fmt.Errorf("%w: missing @", ErrMalformedTag)
	}

	logicalName := portion[:atIdx]
	if logicalName == "" {
		return nil, fmt.Errorf("%w: empty logical name", ErrMalformedTag)
	}

	afterAt := portion[atIdx+1:]
	if len(afterAt) < 27 {
		return nil, fmt.Errorf("%w: hash too short", ErrMalformedTag)
	}

	hash := afterAt[:27]

	return &ArtifactTag{
		LogicalName: logicalName,
		Hash:        hash,
	}, nil
}
