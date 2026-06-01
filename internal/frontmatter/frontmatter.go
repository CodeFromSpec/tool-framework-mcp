// code-from-spec: ROOT/golang/implementation/parsing/frontmatter@4ILIQmvKEglClmtp8jPSfm9HNlw
package frontmatter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
	"github.com/goccy/go-yaml"
)

var ErrFileUnreadable = errors.New("file unreadable")
var ErrMalformedYAML = errors.New("malformed YAML")

type FrontmatterExternal struct {
	Path string
}

type FrontmatterOutput struct {
	ID   string
	Path string
}

type Frontmatter struct {
	DependsOn []string
	External  []*FrontmatterExternal
	Input     string
	Outputs   []*FrontmatterOutput
}

type rawExternal struct {
	Path string `yaml:"path"`
}

type rawOutput struct {
	ID   string `yaml:"id"`
	Path string `yaml:"path"`
}

type rawFrontmatter struct {
	DependsOn []string      `yaml:"depends_on"`
	External  []rawExternal `yaml:"external"`
	Input     string        `yaml:"input"`
	Outputs   []rawOutput   `yaml:"outputs"`
}

func FrontmatterParse(filePath *pathutils.PathCfs) (*Frontmatter, error) {
	reader, err := filereader.FileOpen(filePath)
	if err != nil {
		if errors.Is(err, filereader.ErrFileUnreadable) {
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		return nil, err
	}

	firstLine, err := filereader.FileReadLine(reader)
	if err != nil {
		if errors.Is(err, filereader.ErrEndOfFile) {
			filereader.FileClose(reader)
			return emptyFrontmatter(), nil
		}
		filereader.FileClose(reader)
		return nil, err
	}
	if firstLine != "---" {
		filereader.FileClose(reader)
		return emptyFrontmatter(), nil
	}

	var lines []string
	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				filereader.FileClose(reader)
				return nil, fmt.Errorf("%w: missing closing --- delimiter", ErrMalformedYAML)
			}
			filereader.FileClose(reader)
			return nil, err
		}
		if line == "---" {
			break
		}
		lines = append(lines, line)
	}

	filereader.FileClose(reader)

	if len(lines) == 0 {
		return emptyFrontmatter(), nil
	}

	raw := &rawFrontmatter{}
	if err := yaml.Unmarshal([]byte(strings.Join(lines, "\n")), raw); err != nil {
		return nil, fmt.Errorf("%w: invalid YAML in frontmatter block", ErrMalformedYAML)
	}

	fm := &Frontmatter{}

	if raw.DependsOn != nil {
		fm.DependsOn = raw.DependsOn
	} else {
		fm.DependsOn = []string{}
	}

	fm.Input = raw.Input

	fm.External = make([]*FrontmatterExternal, 0, len(raw.External))
	for _, e := range raw.External {
		if e.Path == "" {
			return nil, fmt.Errorf("%w: external entry missing required field: path", ErrMalformedYAML)
		}
		fm.External = append(fm.External, &FrontmatterExternal{Path: e.Path})
	}

	fm.Outputs = make([]*FrontmatterOutput, 0, len(raw.Outputs))
	for _, o := range raw.Outputs {
		if o.ID == "" {
			return nil, fmt.Errorf("%w: outputs entry missing required field: id", ErrMalformedYAML)
		}
		if o.Path == "" {
			return nil, fmt.Errorf("%w: outputs entry missing required field: path", ErrMalformedYAML)
		}
		fm.Outputs = append(fm.Outputs, &FrontmatterOutput{ID: o.ID, Path: o.Path})
	}

	return fm, nil
}

func emptyFrontmatter() *Frontmatter {
	return &Frontmatter{
		DependsOn: []string{},
		External:  []*FrontmatterExternal{},
		Input:     "",
		Outputs:   []*FrontmatterOutput{},
	}
}
