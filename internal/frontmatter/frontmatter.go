// code-from-spec: SPEC/golang/implementation/parsing/frontmatter@8_N7YzFZFwopfIXxZIgmM96v1fs
package frontmatter

import (
	"errors"
	"fmt"
	"strings"

	yaml "github.com/goccy/go-yaml"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/file"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

var ErrFileUnreadable = errors.New("file unreadable")
var ErrMalformedYAML = errors.New("malformed YAML")

type Frontmatter struct {
	DependsOn []string
	Input     string
	Output    string
}

type yamlFrontmatter struct {
	DependsOn []string `yaml:"depends_on"`
	Input     string   `yaml:"input"`
	Output    string   `yaml:"output"`
}

func FrontmatterParse(filePath *pathutils.PathCfs) (*Frontmatter, error) {
	if filePath == nil {
		return nil, fmt.Errorf("%w: nil file path", ErrFileUnreadable)
	}

	handle, err := file.FileOpen(filePath, "read")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	firstLine, err := file.FileReadLine(handle)
	if err != nil {
		if errors.Is(err, file.ErrEndOfFile) {
			file.FileClose(handle)
			return &Frontmatter{DependsOn: []string{}}, nil
		}
		file.FileClose(handle)
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	if firstLine != "---" {
		file.FileClose(handle)
		return &Frontmatter{DependsOn: []string{}}, nil
	}

	var yamlLines []string
	for {
		line, err := file.FileReadLine(handle)
		if err != nil {
			if errors.Is(err, file.ErrEndOfFile) {
				file.FileClose(handle)
				return nil, fmt.Errorf("%w: missing closing delimiter", ErrMalformedYAML)
			}
			file.FileClose(handle)
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		if line == "---" {
			break
		}
		yamlLines = append(yamlLines, line)
	}

	file.FileClose(handle)

	if len(yamlLines) == 0 {
		return &Frontmatter{DependsOn: []string{}}, nil
	}

	raw := strings.Join(yamlLines, "\n")

	var parsed yamlFrontmatter
	if err := yaml.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrMalformedYAML, err)
	}

	dependsOn := parsed.DependsOn
	if dependsOn == nil {
		dependsOn = []string{}
	}

	return &Frontmatter{
		DependsOn: dependsOn,
		Input:     parsed.Input,
		Output:    parsed.Output,
	}, nil
}
