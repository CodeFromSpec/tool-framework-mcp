// code-from-spec: SPEC/golang/implementation/parsing/frontmatter@0QVTBn-vNbNOgysfrOsEloJnhOE
package frontmatter

import (
	"errors"
	"fmt"
	"strings"

	yaml "github.com/goccy/go-yaml"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/filereader"
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

	reader, err := filereader.FileOpen(*filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	firstLine, err := filereader.FileReadLine(reader)
	if err != nil {
		if errors.Is(err, filereader.ErrEndOfFile) {
			filereader.FileClose(reader)
			return &Frontmatter{DependsOn: []string{}}, nil
		}
		filereader.FileClose(reader)
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	if firstLine != "---" {
		filereader.FileClose(reader)
		return &Frontmatter{DependsOn: []string{}}, nil
	}

	var yamlLines []string
	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				filereader.FileClose(reader)
				return nil, fmt.Errorf("%w: missing closing delimiter", ErrMalformedYAML)
			}
			filereader.FileClose(reader)
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		if line == "---" {
			break
		}
		yamlLines = append(yamlLines, line)
	}

	filereader.FileClose(reader)

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
