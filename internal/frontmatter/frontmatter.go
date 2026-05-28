// code-from-spec: ROOT/golang/implementation/parsing/frontmatter@ZmuTyUn3ro-jvpCA5kYOCU-esvA

package frontmatter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
	"github.com/goccy/go-yaml"
)

// ErrFileUnreadable is returned when the file cannot be opened or read.
var ErrFileUnreadable = errors.New("file unreadable")

// ErrMalformedYAML is returned when the content between --- delimiters
// is not valid YAML.
var ErrMalformedYAML = errors.New("malformed YAML")

// FrontmatterExternalFragment represents a single fragment of an external
// dependency, with an optional description, the raw lines content, and a
// hash identifying the fragment.
type FrontmatterExternalFragment struct {
	Description string
	Lines       string
	Hash        string
}

// FrontmatterExternal represents an external dependency referenced in a
// spec file, identified by its path and containing zero or more fragments.
type FrontmatterExternal struct {
	Path      string
	Fragments []*FrontmatterExternalFragment
}

// FrontmatterOutput represents a single output entry declared in a spec
// file's frontmatter, with an id and a target path.
type FrontmatterOutput struct {
	ID   string
	Path string
}

// Frontmatter holds all structured data parsed from the YAML frontmatter
// block of a spec file. All fields default to their zero value (empty
// string, empty slice) when absent from the YAML.
type Frontmatter struct {
	DependsOn []*string
	External  []*FrontmatterExternal
	Input     string
	Outputs   []*FrontmatterOutput
}

// rawFragment is the unexported struct used to unmarshal YAML fragment entries.
type rawFragment struct {
	Description string `yaml:"description"`
	Lines       string `yaml:"lines"`
	Hash        string `yaml:"hash"`
}

// rawExternal is the unexported struct used to unmarshal YAML external entries.
type rawExternal struct {
	Path      string        `yaml:"path"`
	Fragments []rawFragment `yaml:"fragments"`
}

// rawOutput is the unexported struct used to unmarshal YAML output entries.
type rawOutput struct {
	ID   string `yaml:"id"`
	Path string `yaml:"path"`
}

// rawFrontmatter is the unexported struct used to unmarshal the full YAML block.
type rawFrontmatter struct {
	DependsOn []string      `yaml:"depends_on"`
	External  []rawExternal `yaml:"external"`
	Input     string        `yaml:"input"`
	Outputs   []rawOutput   `yaml:"outputs"`
}

// FrontmatterParse opens and parses the YAML frontmatter of the spec file
// at the given CFS path, returning a populated Frontmatter. All fields
// default to empty when absent from the YAML.
//
// Possible errors:
//   - Path errors propagated from FileOpen (ErrPathEmpty, ErrPathAbsolute,
//     ErrPathContainsBackslash, ErrDirectoryTraversal, ErrResolvesOutsideRoot,
//     ErrCannotDetermineRoot)
//   - ErrFileUnreadable
//   - ErrMalformedYAML
func FrontmatterParse(file_path *pathutils.PathCfs) (*Frontmatter, error) {
	// Step 1: Open the file.
	reader, err := filereader.FileOpen(file_path)
	if err != nil {
		if errors.Is(err, filereader.ErrFileUnreadable) {
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		return nil, err
	}

	// Step 2: Read the first line.
	firstLine, err := filereader.FileReadLine(reader)
	if err != nil {
		if errors.Is(err, filereader.ErrEndOfFile) {
			filereader.FileClose(reader)
			return emptyFrontmatter(), nil
		}
		filereader.FileClose(reader)
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	// Step 3: Check that the first line is exactly "---".
	if firstLine != "---" {
		filereader.FileClose(reader)
		return emptyFrontmatter(), nil
	}

	// Step 4: Collect YAML lines until closing "---" delimiter.
	var yamlLines []string
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
		yamlLines = append(yamlLines, line)
	}

	// Step 5: Close the reader.
	filereader.FileClose(reader)

	// Step 6: If yaml_lines is empty, return an empty Frontmatter.
	if len(yamlLines) == 0 {
		return emptyFrontmatter(), nil
	}

	// Step 7: Join and parse YAML.
	yamlText := strings.Join(yamlLines, "\n")
	var raw rawFrontmatter
	if err := yaml.Unmarshal([]byte(yamlText), &raw); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMalformedYAML, err)
	}

	// Step 8: Build the Frontmatter record from parsed YAML.

	// depends_on
	dependsOn := make([]*string, 0, len(raw.DependsOn))
	for i := range raw.DependsOn {
		s := raw.DependsOn[i]
		dependsOn = append(dependsOn, &s)
	}

	// external
	external := make([]*FrontmatterExternal, 0, len(raw.External))
	for _, rawExt := range raw.External {
		if rawExt.Path == "" {
			return nil, fmt.Errorf("%w: external entry missing required field 'path'", ErrMalformedYAML)
		}

		var fragments []*FrontmatterExternalFragment
		if rawExt.Fragments != nil {
			fragments = make([]*FrontmatterExternalFragment, 0, len(rawExt.Fragments))
			for _, rawFrag := range rawExt.Fragments {
				if rawFrag.Lines == "" {
					return nil, fmt.Errorf("%w: fragment entry missing required field 'lines'", ErrMalformedYAML)
				}
				if rawFrag.Hash == "" {
					return nil, fmt.Errorf("%w: fragment entry missing required field 'hash'", ErrMalformedYAML)
				}
				fragments = append(fragments, &FrontmatterExternalFragment{
					Description: rawFrag.Description,
					Lines:       rawFrag.Lines,
					Hash:        rawFrag.Hash,
				})
			}
		}

		external = append(external, &FrontmatterExternal{
			Path:      rawExt.Path,
			Fragments: fragments,
		})
	}

	// outputs
	outputs := make([]*FrontmatterOutput, 0, len(raw.Outputs))
	for _, rawOut := range raw.Outputs {
		if rawOut.ID == "" {
			return nil, fmt.Errorf("%w: output entry missing required field 'id'", ErrMalformedYAML)
		}
		if rawOut.Path == "" {
			return nil, fmt.Errorf("%w: output entry missing required field 'path'", ErrMalformedYAML)
		}
		outputs = append(outputs, &FrontmatterOutput{
			ID:   rawOut.ID,
			Path: rawOut.Path,
		})
	}

	// Step 9: Return the Frontmatter record.
	return &Frontmatter{
		DependsOn: dependsOn,
		External:  external,
		Input:     raw.Input,
		Outputs:   outputs,
	}, nil
}

// emptyFrontmatter returns a Frontmatter with all fields set to their zero values.
func emptyFrontmatter() *Frontmatter {
	return &Frontmatter{
		DependsOn: []*string{},
		External:  []*FrontmatterExternal{},
		Input:     "",
		Outputs:   []*FrontmatterOutput{},
	}
}
