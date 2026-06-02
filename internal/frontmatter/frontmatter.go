// code-from-spec: ROOT/golang/implementation/parsing/frontmatter@b6FiWCzdItRlxB2jEWdAjKb3Pn4
package frontmatter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/goccy/go-yaml"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

var ErrFileUnreadable = errors.New("file unreadable")
var ErrMalformedYAML = errors.New("malformed YAML")

type FrontmatterExternal struct {
	Path string
}

type Frontmatter struct {
	DependsOn []string
	External  []*FrontmatterExternal
	Input     string
	Output    string
}

type rawExternal struct {
	Path *string `yaml:"path"`
}

type rawFrontmatter struct {
	DependsOn []string      `yaml:"depends_on"`
	External  []rawExternal `yaml:"external"`
	Input     string        `yaml:"input"`
	Output    string        `yaml:"output"`
}

func FrontmatterParse(file_path *pathutils.PathCfs) (*Frontmatter, error) {
	reader, err := filereader.FileOpen(file_path)
	if err != nil {
		if errors.Is(err, filereader.ErrFileUnreadable) {
			return nil, fmt.Errorf("%w: %s", ErrFileUnreadable, err)
		}
		return nil, err
	}

	firstLine, err := filereader.FileReadLine(reader)
	if err != nil {
		if errors.Is(err, filereader.ErrEndOfFile) {
			filereader.FileClose(reader)
			return &Frontmatter{}, nil
		}
		filereader.FileClose(reader)
		return nil, fmt.Errorf("%w: %s", ErrFileUnreadable, err)
	}

	if firstLine != "---" {
		filereader.FileClose(reader)
		return &Frontmatter{}, nil
	}

	var lines []string
	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				filereader.FileClose(reader)
				return nil, fmt.Errorf("%w: frontmatter not closed", ErrMalformedYAML)
			}
			filereader.FileClose(reader)
			return nil, fmt.Errorf("%w: %s", ErrFileUnreadable, err)
		}
		if line == "---" {
			break
		}
		lines = append(lines, line)
	}

	filereader.FileClose(reader)

	yamlContent := strings.Join(lines, "\n")

	var raw rawFrontmatter
	if err := yaml.Unmarshal([]byte(yamlContent), &raw); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrMalformedYAML, err)
	}

	fm := &Frontmatter{
		DependsOn: raw.DependsOn,
		Input:     raw.Input,
		Output:    raw.Output,
	}

	if fm.DependsOn == nil {
		fm.DependsOn = []string{}
	}

	for _, ext := range raw.External {
		if ext.Path == nil {
			return nil, fmt.Errorf("%w: external entry missing path field", ErrMalformedYAML)
		}
		fm.External = append(fm.External, &FrontmatterExternal{Path: *ext.Path})
	}

	if fm.External == nil {
		fm.External = []*FrontmatterExternal{}
	}

	return fm, nil
}
