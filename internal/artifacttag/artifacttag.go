// code-from-spec: SPEC/golang/implementation/parsing/artifact_tag@KpMgRYz_KFjRkQcxgHjsrkR-ttc
package artifacttag

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/file"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

var ErrFileUnreadable = errors.New("file cannot be opened or read")
var ErrNoTagFound     = errors.New("file has no code-from-spec: tag")
var ErrMalformedTag   = errors.New("tag exists but cannot be parsed")

const tagPrefix = "code-from-spec: "
const hashLength = 27

type ArtifactTag struct {
	LogicalName string
	Hash        string
}

func ArtifactTagExtract(filePath *pathutils.PathCfs) (*ArtifactTag, error) {
	if filePath == nil {
		return nil, fmt.Errorf("%w: nil file path", ErrFileUnreadable)
	}

	handle, err := file.FileOpen(filePath, "read")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	foundLine := ""

	for {
		line, err := file.FileReadLine(handle)
		if errors.Is(err, file.ErrEndOfFile) {
			break
		}
		if err != nil {
			file.FileClose(handle)
			return nil, fmt.Errorf("%w", err)
		}

		idx := strings.Index(line, tagPrefix)
		if idx != -1 {
			foundLine = line[idx+len(tagPrefix):]
			break
		}
	}

	file.FileClose(handle)

	if foundLine == "" {
		return nil, ErrNoTagFound
	}

	remainder := strings.TrimLeft(foundLine, " \t")

	atIdx := strings.Index(remainder, "@")
	if atIdx == -1 {
		return nil, fmt.Errorf("%w: missing '@' separator", ErrMalformedTag)
	}

	logicalName := remainder[:atIdx]
	if logicalName == "" {
		return nil, fmt.Errorf("%w: empty logical name", ErrMalformedTag)
	}

	hashCandidate := remainder[atIdx+1:]
	if len(hashCandidate) < hashLength {
		return nil, fmt.Errorf("%w: hash too short (got %d, need %d)", ErrMalformedTag, len(hashCandidate), hashLength)
	}

	hash := hashCandidate[:hashLength]

	return &ArtifactTag{
		LogicalName: logicalName,
		Hash:        hash,
	}, nil
}
