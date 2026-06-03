// code-from-spec: ROOT/golang/implementation/parsing/frontmatter@qTNGMpnpZnYkFN_W0OaI0M5gI3U
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

type Frontmatter struct {
	DependsOn []string
	External  []*FrontmatterExternal
	Input     string
	Output    string
}

type rawExternal struct {
	Path string `yaml:"path"`
}

type rawFrontmatter struct {
	DependsOn []string      `yaml:"depends_on"`
	External  []rawExternal `yaml:"external"`
	Input     string        `yaml:"input"`
	Output    string        `yaml:"output"`
}

func FrontmatterParse(filePath *pathutils.PathCfs) (*Frontmatter, error) {
	reader, err := filereader.FileOpen(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	firstLine, err := filereader.FileReadLine(reader)
	if err != nil {
		if errors.Is(err, filereader.ErrEndOfFile) {
			filereader.FileClose(reader)
			return &Frontmatter{}, nil
		}
		filereader.FileClose(reader)
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
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
				return nil, fmt.Errorf("%w: unterminated frontmatter block", ErrMalformedYAML)
			}
			filereader.FileClose(reader)
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		if line == "---" {
			break
		}
		lines = append(lines, line)
	}

	filereader.FileClose(reader)

	raw := rawFrontmatter{}
	if err := yaml.Unmarshal([]byte(strings.Join(lines, "\n")), &raw); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMalformedYAML, err)
	}

	fm := &Frontmatter{
		DependsOn: raw.DependsOn,
		Input:     raw.Input,
		Output:    raw.Output,
	}

	if fm.DependsOn == nil {
		fm.DependsOn = []string{}
	}

	for _, e := range raw.External {
		if e.Path == "" {
			return nil, fmt.Errorf("%w: external entry missing required path field", ErrMalformedYAML)
		}
		fm.External = append(fm.External, &FrontmatterExternal{Path: e.Path})
	}

	if fm.External == nil {
		fm.External = []*FrontmatterExternal{}
	}

	return fm, nil
}
