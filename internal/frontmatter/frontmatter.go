// code-from-spec: ROOT/golang/implementation/parsing/frontmatter@RWVyO815pmIgcbBjLRh7FTRF5uM
package frontmatter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/goccy/go-yaml"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

var ErrFileUnreadable = errors.New("file cannot be opened or read")
var ErrMalformedYAML = errors.New("content between --- delimiters is not valid YAML")

type Frontmatter struct {
	DependsOn []string
	Input     string
	Output    string
}

type frontmatterYAML struct {
	DependsOn []string `yaml:"depends_on"`
	Input     string   `yaml:"input"`
	Output    string   `yaml:"output"`
}

func FrontmatterParse(filePath *pathutils.PathCfs) (*Frontmatter, error) {
	reader, err := filereader.FileOpen(filePath)
	if err != nil {
		if errors.Is(err, filereader.ErrFileUnreadable) {
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		return nil, fmt.Errorf("opening file: %w", err)
	}

	firstLine, err := filereader.FileReadLine(reader)
	if err != nil {
		if errors.Is(err, filereader.ErrEndOfFile) {
			filereader.FileClose(reader)
			return emptyFrontmatter(), nil
		}
		filereader.FileClose(reader)
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	if firstLine != "---" {
		filereader.FileClose(reader)
		return emptyFrontmatter(), nil
	}

	var yamlLines []string
	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				filereader.FileClose(reader)
				return nil, fmt.Errorf("%w: missing closing ---", ErrMalformedYAML)
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
		return emptyFrontmatter(), nil
	}

	yamlText := strings.Join(yamlLines, "\n")

	var parsed frontmatterYAML
	if err := yaml.Unmarshal([]byte(yamlText), &parsed); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMalformedYAML, err)
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

func emptyFrontmatter() *Frontmatter {
	return &Frontmatter{
		DependsOn: []string{},
		Input:     "",
		Output:    "",
	}
}
