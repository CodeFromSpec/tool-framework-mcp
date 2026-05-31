// code-from-spec: ROOT/golang/implementation/parsing/frontmatter@EyNEULEEZ0AiSRGOLyoq70vHC3o
package frontmatter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/goccy/go-yaml"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ErrFileUnreadable is returned when the spec file cannot be opened or read.
var ErrFileUnreadable = errors.New("file unreadable")

// ErrMalformedYAML is returned when the content between --- delimiters
// is not valid YAML.
var ErrMalformedYAML = errors.New("malformed YAML")

// FrontmatterExternalFragment represents an optional annotated fragment
// of an external file referenced in the frontmatter.
type FrontmatterExternalFragment struct {
	// Description is an optional human-readable label for the fragment.
	Description string

	// Lines is the line range or content selector string.
	Lines string

	// Hash is a content hash for the fragment.
	Hash string
}

// FrontmatterExternal represents a single external file reference,
// optionally broken into named fragments.
type FrontmatterExternal struct {
	// Path is the CFS-format path to the external file.
	Path string

	// Fragments is an optional list of fragments within the external file.
	Fragments []*FrontmatterExternalFragment
}

// FrontmatterOutput represents a single declared output in the frontmatter.
type FrontmatterOutput struct {
	// ID is the logical identifier for the output artifact.
	ID string

	// Path is the CFS-format path where the output file should be written.
	Path string
}

// Frontmatter holds the parsed contents of the YAML frontmatter block
// from a spec node file.
type Frontmatter struct {
	// DependsOn is the list of logical names this node depends on.
	DependsOn []string

	// External is the list of external file references.
	External []*FrontmatterExternal

	// Input is the CFS-format path to the input source material file.
	Input string

	// Outputs is the list of declared output artifacts.
	Outputs []*FrontmatterOutput
}

// rawFragment is the unexported struct used for YAML unmarshalling of fragments.
type rawFragment struct {
	Description string `yaml:"description"`
	Lines       string `yaml:"lines"`
	Hash        string `yaml:"hash"`
}

// rawExternal is the unexported struct used for YAML unmarshalling of external entries.
type rawExternal struct {
	Path      string        `yaml:"path"`
	Fragments []rawFragment `yaml:"fragments"`
}

// rawOutput is the unexported struct used for YAML unmarshalling of output entries.
type rawOutput struct {
	ID   string `yaml:"id"`
	Path string `yaml:"path"`
}

// rawFrontmatter is the unexported struct used for YAML unmarshalling of the full frontmatter block.
type rawFrontmatter struct {
	DependsOn []string      `yaml:"depends_on"`
	External  []rawExternal `yaml:"external"`
	Input     string        `yaml:"input"`
	Outputs   []rawOutput   `yaml:"outputs"`
}

// FrontmatterParse opens and parses the spec node file at filePath,
// extracting the YAML frontmatter block delimited by --- markers.
//
// All fields default to their zero values (empty list, empty string)
// when absent from the YAML.
//
// Errors:
//   - ErrFileUnreadable: the file cannot be opened or read.
//   - ErrMalformedYAML: the content between --- delimiters is not valid YAML.
//   - (FileReader.*): propagated from FileOpen.
func FrontmatterParse(filePath *pathutils.PathCfs) (*Frontmatter, error) {
	// Step 1: Open the file.
	reader, err := filereader.FileOpen(filePath)
	if err != nil {
		if errors.Is(err, filereader.ErrFileUnreadable) {
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		return nil, fmt.Errorf("opening file: %w", err)
	}

	// Step 2: Read the first line.
	firstLine, err := filereader.FileReadLine(reader)
	if err != nil {
		filereader.FileClose(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			return emptyFrontmatter(), nil
		}
		return nil, fmt.Errorf("%w: reading first line: %w", ErrFileUnreadable, err)
	}
	if firstLine != "---" {
		filereader.FileClose(reader)
		return emptyFrontmatter(), nil
	}

	// Step 3: Collect lines until closing "---".
	var yamlLines []string
	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			filereader.FileClose(reader)
			if errors.Is(err, filereader.ErrEndOfFile) {
				return nil, fmt.Errorf("%w: missing closing ---", ErrMalformedYAML)
			}
			return nil, fmt.Errorf("%w: reading frontmatter: %w", ErrFileUnreadable, err)
		}
		if line == "---" {
			break
		}
		yamlLines = append(yamlLines, line)
	}

	// Step 4: Close the reader.
	filereader.FileClose(reader)

	// Step 5: Join lines into YAML text.
	yamlText := strings.Join(yamlLines, "\n")

	// Step 6: Parse YAML.
	var raw rawFrontmatter
	if len(strings.TrimSpace(yamlText)) > 0 {
		if err := yaml.Unmarshal([]byte(yamlText), &raw); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrMalformedYAML, err)
		}
	}

	// Step 7: Build and return the Frontmatter record.
	fm := &Frontmatter{}

	// a. depends_on
	if raw.DependsOn != nil {
		fm.DependsOn = raw.DependsOn
	} else {
		fm.DependsOn = []string{}
	}

	// b. external
	fm.External = make([]*FrontmatterExternal, 0, len(raw.External))
	for _, re := range raw.External {
		if re.Path == "" {
			return nil, fmt.Errorf("%w: external entry missing path", ErrMalformedYAML)
		}
		ext := &FrontmatterExternal{
			Path:      re.Path,
			Fragments: make([]*FrontmatterExternalFragment, 0, len(re.Fragments)),
		}
		for _, rf := range re.Fragments {
			if rf.Lines == "" {
				return nil, fmt.Errorf("%w: external fragment missing lines", ErrMalformedYAML)
			}
			if rf.Hash == "" {
				return nil, fmt.Errorf("%w: external fragment missing hash", ErrMalformedYAML)
			}
			frag := &FrontmatterExternalFragment{
				Description: rf.Description,
				Lines:       rf.Lines,
				Hash:        rf.Hash,
			}
			ext.Fragments = append(ext.Fragments, frag)
		}
		fm.External = append(fm.External, ext)
	}

	// c. input
	fm.Input = raw.Input

	// d. outputs
	fm.Outputs = make([]*FrontmatterOutput, 0, len(raw.Outputs))
	for _, ro := range raw.Outputs {
		if ro.ID == "" {
			return nil, fmt.Errorf("%w: output entry missing id", ErrMalformedYAML)
		}
		if ro.Path == "" {
			return nil, fmt.Errorf("%w: output entry missing path", ErrMalformedYAML)
		}
		fm.Outputs = append(fm.Outputs, &FrontmatterOutput{
			ID:   ro.ID,
			Path: ro.Path,
		})
	}

	return fm, nil
}

// emptyFrontmatter returns a Frontmatter record with all fields at their defaults.
func emptyFrontmatter() *Frontmatter {
	return &Frontmatter{
		DependsOn: []string{},
		External:  []*FrontmatterExternal{},
		Outputs:   []*FrontmatterOutput{},
	}
}
