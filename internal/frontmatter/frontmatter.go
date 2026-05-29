// code-from-spec: ROOT/golang/implementation/parsing/frontmatter@fSpWawzneVqIzgYOIdcNHaM7sRo

package frontmatter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
	"github.com/goccy/go-yaml"
)

var (
	// ErrFileUnreadable is returned when the file cannot be opened or read.
	ErrFileUnreadable = errors.New("file unreadable")

	// ErrMalformedYAML is returned when the content between --- delimiters
	// is not valid YAML.
	ErrMalformedYAML = errors.New("malformed YAML")
)

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

// rawFragment is the unexported struct used for YAML unmarshalling of a fragment entry.
type rawFragment struct {
	Description string `yaml:"description"`
	Lines       string `yaml:"lines"`
	Hash        string `yaml:"hash"`
}

// rawExternal is the unexported struct used for YAML unmarshalling of an external entry.
type rawExternal struct {
	Path      string        `yaml:"path"`
	Fragments []rawFragment `yaml:"fragments"`
}

// rawOutput is the unexported struct used for YAML unmarshalling of an output entry.
type rawOutput struct {
	ID   string `yaml:"id"`
	Path string `yaml:"path"`
}

// rawFrontmatter is the unexported struct used for YAML unmarshalling of the frontmatter block.
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
	if errors.Is(err, filereader.ErrEndOfFile) {
		filereader.FileClose(reader)
		return emptyFrontmatter(), nil
	}
	if err != nil {
		filereader.FileClose(reader)
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	// Step 3: Check for opening "---".
	if firstLine != "---" {
		filereader.FileClose(reader)
		return emptyFrontmatter(), nil
	}

	// Step 4: Collect YAML lines until closing "---".
	var yamlLines []string
	for {
		line, err := filereader.FileReadLine(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			filereader.FileClose(reader)
			return nil, fmt.Errorf("%w: missing closing ---", ErrMalformedYAML)
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

	// Step 5: Close the reader.
	filereader.FileClose(reader)

	// Step 6: Join and parse YAML.
	yamlContent := strings.Join(yamlLines, "\n")

	var raw rawFrontmatter
	if err := yaml.Unmarshal([]byte(yamlContent), &raw); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMalformedYAML, err)
	}

	// Step 7: If the YAML is empty (all zero values), return empty Frontmatter.
	// We check by seeing if there's no meaningful content — yaml.Unmarshal
	// on an empty or whitespace-only string will leave raw at its zero value.
	// We proceed to step 8 regardless; zero values simply yield empty slices/strings.

	// Step 8: Extract and validate fields.

	// Extract depends_on.
	dependsOn := make([]*string, 0, len(raw.DependsOn))
	for i := range raw.DependsOn {
		s := raw.DependsOn[i]
		dependsOn = append(dependsOn, &s)
	}

	// Extract external.
	external := make([]*FrontmatterExternal, 0, len(raw.External))
	for _, rawExt := range raw.External {
		if rawExt.Path == "" {
			return nil, fmt.Errorf("%w: external entry missing required field \"path\"", ErrMalformedYAML)
		}
		fragments := make([]*FrontmatterExternalFragment, 0, len(rawExt.Fragments))
		for _, rawFrag := range rawExt.Fragments {
			if rawFrag.Lines == "" {
				return nil, fmt.Errorf("%w: fragment entry missing required field \"lines\"", ErrMalformedYAML)
			}
			if rawFrag.Hash == "" {
				return nil, fmt.Errorf("%w: fragment entry missing required field \"hash\"", ErrMalformedYAML)
			}
			frag := &FrontmatterExternalFragment{
				Description: rawFrag.Description,
				Lines:       rawFrag.Lines,
				Hash:        rawFrag.Hash,
			}
			fragments = append(fragments, frag)
		}
		ext := &FrontmatterExternal{
			Path:      rawExt.Path,
			Fragments: fragments,
		}
		external = append(external, ext)
	}

	// Extract outputs.
	outputs := make([]*FrontmatterOutput, 0, len(raw.Outputs))
	for _, rawOut := range raw.Outputs {
		if rawOut.ID == "" {
			return nil, fmt.Errorf("%w: output entry missing required field \"id\"", ErrMalformedYAML)
		}
		if rawOut.Path == "" {
			return nil, fmt.Errorf("%w: output entry missing required field \"path\"", ErrMalformedYAML)
		}
		out := &FrontmatterOutput{
			ID:   rawOut.ID,
			Path: rawOut.Path,
		}
		outputs = append(outputs, out)
	}

	// Step 9: Construct and return the Frontmatter.
	return &Frontmatter{
		DependsOn: dependsOn,
		External:  external,
		Input:     raw.Input,
		Outputs:   outputs,
	}, nil
}

// emptyFrontmatter returns a Frontmatter with all fields set to their zero/empty values.
func emptyFrontmatter() *Frontmatter {
	return &Frontmatter{
		DependsOn: []*string{},
		External:  []*FrontmatterExternal{},
		Input:     "",
		Outputs:   []*FrontmatterOutput{},
	}
}
