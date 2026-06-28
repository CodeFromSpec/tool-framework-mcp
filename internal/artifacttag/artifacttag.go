// code-from-spec: SPEC/golang/implementation/parsing/artifact_tag@SpaLfL2cooO3k5p2429vSl1O1l4
package artifacttag

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/file"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

var ErrNoTagFound   = errors.New("no code-from-spec tag found in file")
var ErrMalformedTag = errors.New("code-from-spec tag is malformed")

type ArtifactTag struct {
	LogicalName string
	Hash        string
}

func ArtifactTagExtract(filePath *pathutils.PathCfs) (*ArtifactTag, error) {
	handle, err := file.FileOpen(filePath, "read", 30000)
	if err != nil {
		return nil, err
	}

	foundLine := ""
	for {
		line, err := file.FileReadLine(handle)
		if errors.Is(err, file.ErrEndOfFile) {
			break
		}
		if err != nil {
			file.FileClose(handle)
			return nil, err
		}
		if strings.Contains(line, "code-from-spec: ") {
			foundLine = line
			break
		}
	}

	file.FileClose(handle)

	if foundLine == "" {
		return nil, ErrNoTagFound
	}

	idx := strings.Index(foundLine, "code-from-spec: ")
	remainder := foundLine[idx+len("code-from-spec: "):]
	remainder = strings.TrimLeft(remainder, " \t")

	atIdx := strings.Index(remainder, "@")
	if atIdx == -1 {
		return nil, fmt.Errorf("%w", ErrMalformedTag)
	}

	logicalName := remainder[:atIdx]
	if logicalName == "" {
		return nil, fmt.Errorf("%w", ErrMalformedTag)
	}

	afterAt := remainder[atIdx+1:]
	if len(afterAt) < 27 {
		return nil, fmt.Errorf("%w", ErrMalformedTag)
	}

	hash := afterAt[:27]

	return &ArtifactTag{
		LogicalName: logicalName,
		Hash:        hash,
	}, nil
}
