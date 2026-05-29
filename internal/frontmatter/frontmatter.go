// code-from-spec: ROOT/golang/implementation/parsing/frontmatter@ls_tSbB_Zep-Xc-Ydf9Ej4FaS5A

package frontmatter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
	"github.com/goccy/go-yaml"
)

// ErrFileUnreadable is returned when the file cannot be opened or read.
var ErrFileUnreadable = errors.New("file unreadable")

// ErrMalformedYAML is returned when the content between --- delimiters
// is not valid YAML.
var ErrMalformedYAML = errors.New("malformed YAML")

// FrontmatterExternalFragment represents a single fragment entry within
// an external dependency, identified by an optional description, its
// raw content lines, and a hash.
type FrontmatterExternalFragment struct {
	Description string
	Lines       string
	Hash        string
}

// FrontmatterExternal represents an external dependency declared in a
// spec file's frontmatter, consisting of a path and an optional list
// of fragments.
type FrontmatterExternal struct {
	Path      string
	Fragments []*FrontmatterExternalFragment
}

// FrontmatterOutput represents a single output entry declared in a
// spec file's frontmatter, with an id and a target path.
type FrontmatterOutput struct {
	ID   string
	Path string
}

// Frontmatter holds the parsed contents of a spec file's YAML
// frontmatter block. All fields default to empty (empty list,
// empty string) when absent from the YAML.
type Frontmatter struct {
	DependsOn []string
	External  []*FrontmatterExternal
	Input     string
	Outputs   []*FrontmatterOutput
}

// yamlFragment is the unexported struct used for YAML unmarshalling of a fragment entry.
type yamlFragment struct {
	Description string `yaml:"description"`
	Lines       string `yaml:"lines"`
	Hash        string `yaml:"hash"`
}

// yamlExternal is the unexported struct used for YAML unmarshalling of an external entry.
type yamlExternal struct {
	Path      string         `yaml:"path"`
	Fragments []yamlFragment `yaml:"fragments"`
}

// yamlOutput is the unexported struct used for YAML unmarshalling of an output entry.
type yamlOutput struct {
	ID   string `yaml:"id"`
	Path string `yaml:"path"`
}

// yamlFrontmatter is the unexported struct used for YAML unmarshalling of the frontmatter block.
type yamlFrontmatter struct {
	DependsOn []string       `yaml:"depends_on"`
	External  []yamlExternal `yaml:"external"`
	Input     string         `yaml:"input"`
	Outputs   []yamlOutput   `yaml:"outputs"`
}

// FrontmatterParse opens and parses the YAML frontmatter of the spec file
// at the given CFS path. It returns a populated Frontmatter with all
// declared fields. Fields absent from the YAML default to empty values
// (empty string, empty list).
//
// Returns an error if:
//   - the path is invalid or cannot be resolved (path errors propagated
//     from the underlying file open operation).
//   - the file cannot be opened or read (ErrFileUnreadable).
//   - the content between --- delimiters is not valid YAML (ErrMalformedYAML).
func FrontmatterParse(file_path *pathutils.PathCfs) (*Frontmatter, error) {
	// Step 1: Open the file.
	reader, err := filereader.FileOpen(file_path)
	if err != nil {
		return nil, err
	}

	// Step 2: Read the first line and check for opening delimiter.
	firstLine, err := filereader.FileReadLine(reader)
	if errors.Is(err, filereader.ErrEndOfFile) {
		filereader.FileClose(reader)
		return &Frontmatter{}, nil
	}
	if err != nil {
		filereader.FileClose(reader)
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}
	if firstLine != "---" {
		filereader.FileClose(reader)
		return &Frontmatter{}, nil
	}

	// Step 3: Collect YAML lines until closing delimiter.
	var yamlLines []string
	for {
		line, err := filereader.FileReadLine(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			filereader.FileClose(reader)
			return nil, fmt.Errorf("%w: missing closing delimiter", ErrMalformedYAML)
		}
		if err != nil {
			filereader.FileClose(reader)
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		if line == "---" {
			break
		}
		yamlLines = append(yamlLines, line)
	}

	// Step 4: Close the file.
	filereader.FileClose(reader)

	// Step 5: If no YAML lines collected, return empty Frontmatter.
	if len(yamlLines) == 0 {
		return &Frontmatter{}, nil
	}

	// Step 6: Parse the collected YAML.
	yamlText := strings.Join(yamlLines, "\n")
	var raw yamlFrontmatter
	if err := yaml.Unmarshal([]byte(yamlText), &raw); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMalformedYAML, err)
	}

	// Step 7: Convert parsed YAML into exported types, validating required fields.
	fm := &Frontmatter{}

	// depends_on
	if raw.DependsOn != nil {
		fm.DependsOn = raw.DependsOn
	} else {
		fm.DependsOn = []string{}
	}

	// external
	fm.External = make([]*FrontmatterExternal, 0, len(raw.External))
	for _, extRaw := range raw.External {
		if extRaw.Path == "" {
			return nil, fmt.Errorf("%w: external entry missing required field 'path'", ErrMalformedYAML)
		}
		ext := &FrontmatterExternal{
			Path:      extRaw.Path,
			Fragments: make([]*FrontmatterExternalFragment, 0, len(extRaw.Fragments)),
		}
		for _, fragRaw := range extRaw.Fragments {
			if fragRaw.Lines == "" {
				return nil, fmt.Errorf("%w: fragment entry missing required field 'lines'", ErrMalformedYAML)
			}
			if fragRaw.Hash == "" {
				return nil, fmt.Errorf("%w: fragment entry missing required field 'hash'", ErrMalformedYAML)
			}
			ext.Fragments = append(ext.Fragments, &FrontmatterExternalFragment{
				Description: fragRaw.Description,
				Lines:       fragRaw.Lines,
				Hash:        fragRaw.Hash,
			})
		}
		fm.External = append(fm.External, ext)
	}

	// input
	fm.Input = raw.Input

	// outputs
	fm.Outputs = make([]*FrontmatterOutput, 0, len(raw.Outputs))
	for _, outRaw := range raw.Outputs {
		if outRaw.ID == "" {
			return nil, fmt.Errorf("%w: output entry missing required field 'id'", ErrMalformedYAML)
		}
		if outRaw.Path == "" {
			return nil, fmt.Errorf("%w: output entry missing required field 'path'", ErrMalformedYAML)
		}
		fm.Outputs = append(fm.Outputs, &FrontmatterOutput{
			ID:   outRaw.ID,
			Path: outRaw.Path,
		})
	}

	// Step 8: Return the populated Frontmatter.
	return fm, nil
}
