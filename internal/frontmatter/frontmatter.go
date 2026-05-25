// code-from-spec: ROOT/golang/internal/frontmatter/code@PENDING
package frontmatter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/goccy/go-yaml"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filereader"
)

// Sentinel errors for frontmatter operations.
var (
	ErrRead             = errors.New("error reading file")
	ErrFrontmatterParse = errors.New("error parsing frontmatter")
)

// Output represents an output artifact declared in frontmatter.
type Output struct {
	ID   string
	Path string
}

// ExternalFragment describes a specific fragment within an external file.
type ExternalFragment struct {
	Description string
	Lines       string
	Hash        string
}

// External represents an external file reference in frontmatter.
type External struct {
	Path      string
	Fragments []ExternalFragment
}

// Frontmatter holds the parsed content of a spec node's YAML frontmatter.
type Frontmatter struct {
	DependsOn []string
	External  []External
	Input     string
	Outputs   []Output
}

// yamlFrontmatter is an unexported struct with yaml tags for unmarshalling.
type yamlFrontmatter struct {
	DependsOn []string       `yaml:"depends_on"`
	External  []yamlExternal `yaml:"external"`
	Input     string         `yaml:"input"`
	Outputs   []yamlOutput   `yaml:"outputs"`
}

type yamlOutput struct {
	ID   string `yaml:"id"`
	Path string `yaml:"path"`
}

type yamlExternalFragment struct {
	Description string `yaml:"description"`
	Lines       string `yaml:"lines"`
	Hash        string `yaml:"hash"`
}

type yamlExternal struct {
	Path      string                 `yaml:"path"`
	Fragments []yamlExternalFragment `yaml:"fragments"`
}

// ParseFrontmatter reads a spec node file, extracts the YAML frontmatter
// block (delimited by "---" lines), and returns the parsed result. If the
// file has no frontmatter delimiters, it returns an empty Frontmatter
// (not an error).
func ParseFrontmatter(filePath string) (*Frontmatter, error) {
	reader, err := filereader.OpenFileReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %s", ErrRead, filePath, err)
	}

	// Read the first line to check for opening delimiter.
	firstLine, err := reader.ReadLine()
	if errors.Is(err, filereader.ErrEndOfFile) {
		// Empty file: no frontmatter.
		return &Frontmatter{}, nil
	}

	if firstLine != "---" {
		// No frontmatter block.
		return &Frontmatter{}, nil
	}

	// Collect YAML lines until closing "---".
	var yamlLines []string
	foundClosing := false
	for {
		line, err := reader.ReadLine()
		if errors.Is(err, filereader.ErrEndOfFile) {
			break
		}
		if line == "---" {
			foundClosing = true
			break
		}
		yamlLines = append(yamlLines, line)
	}

	if !foundClosing {
		return nil, fmt.Errorf("%w: %s: unclosed frontmatter block", ErrFrontmatterParse, filePath)
	}

	yamlText := strings.Join(yamlLines, "\n")

	var raw yamlFrontmatter
	if err := yaml.Unmarshal([]byte(yamlText), &raw); err != nil {
		return nil, fmt.Errorf("%w: %s: %s", ErrFrontmatterParse, filePath, err)
	}

	return convertFrontmatter(&raw), nil
}

// convertFrontmatter maps the unexported YAML struct to the exported types.
func convertFrontmatter(raw *yamlFrontmatter) *Frontmatter {
	fm := &Frontmatter{
		DependsOn: raw.DependsOn,
		Input:     raw.Input,
	}

	// Ensure slices are non-nil for consistent usage.
	if fm.DependsOn == nil {
		fm.DependsOn = []string{}
	}

	for _, ext := range raw.External {
		e := External{Path: ext.Path}
		for _, frag := range ext.Fragments {
			e.Fragments = append(e.Fragments, ExternalFragment{
				Description: frag.Description,
				Lines:       frag.Lines,
				Hash:        frag.Hash,
			})
		}
		if e.Fragments == nil {
			e.Fragments = []ExternalFragment{}
		}
		fm.External = append(fm.External, e)
	}
	if fm.External == nil {
		fm.External = []External{}
	}

	for _, out := range raw.Outputs {
		fm.Outputs = append(fm.Outputs, Output{
			ID:   out.ID,
			Path: out.Path,
		})
	}
	if fm.Outputs == nil {
		fm.Outputs = []Output{}
	}

	return fm
}
