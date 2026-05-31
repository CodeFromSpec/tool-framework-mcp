// code-from-spec: ROOT/golang/implementation/parsing/frontmatter@0TsuUXXV2SlJwckixpG7FegxhFE

package frontmatter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
	"github.com/goccy/go-yaml"
)

// ErrFileUnreadable is returned when the file at the given path cannot
// be opened or read.
var ErrFileUnreadable = errors.New("file unreadable")

// ErrMalformedYAML is returned when the content between the --- delimiters
// is not valid YAML.
var ErrMalformedYAML = errors.New("malformed YAML in frontmatter")

// FrontmatterExternalFragment represents a single fragment entry
// within an external dependency declaration.
type FrontmatterExternalFragment struct {
	// Description is an optional human-readable description of
	// the fragment. Empty string when absent.
	Description string

	// Lines holds the content lines of the fragment.
	Lines string

	// Hash is a content hash for the fragment.
	Hash string
}

// FrontmatterExternal represents an external file referenced in the
// frontmatter, along with its optional fragment list.
type FrontmatterExternal struct {
	// Path is the CFS path to the external file.
	Path string

	// Fragments is the list of fragments within the external file.
	// Empty slice when absent.
	Fragments []*FrontmatterExternalFragment
}

// FrontmatterOutput represents a single entry in the outputs list of
// the frontmatter.
type FrontmatterOutput struct {
	// ID is the identifier for this output artifact.
	ID string

	// Path is the CFS path where the output file should be written.
	Path string
}

// Frontmatter holds all parsed fields from a spec node's YAML
// frontmatter block. All fields default to their zero value (empty
// slice or empty string) when absent from the YAML.
type Frontmatter struct {
	// DependsOn is the list of logical names this node depends on.
	DependsOn []string

	// External is the list of external file references.
	External []*FrontmatterExternal

	// Input is the CFS path to the input material for transformation.
	// Empty string when absent.
	Input string

	// Outputs is the list of output descriptors declared by this node.
	Outputs []*FrontmatterOutput
}

// rawFragment is the unexported struct used to unmarshal a fragment
// from YAML.
type rawFragment struct {
	Description string `yaml:"description"`
	Lines       string `yaml:"lines"`
	Hash        string `yaml:"hash"`
}

// rawExternal is the unexported struct used to unmarshal an external
// entry from YAML.
type rawExternal struct {
	Path      string        `yaml:"path"`
	Fragments []rawFragment `yaml:"fragments"`
}

// rawOutput is the unexported struct used to unmarshal an output entry
// from YAML.
type rawOutput struct {
	ID   string `yaml:"id"`
	Path string `yaml:"path"`
}

// rawFrontmatter is the unexported struct used to unmarshal the full
// frontmatter block from YAML.
type rawFrontmatter struct {
	DependsOn []string      `yaml:"depends_on"`
	External  []rawExternal `yaml:"external"`
	Input     string        `yaml:"input"`
	Outputs   []rawOutput   `yaml:"outputs"`
}

// emptyFrontmatter returns a Frontmatter with all fields set to their
// zero/empty values.
func emptyFrontmatter() *Frontmatter {
	return &Frontmatter{
		DependsOn: []string{},
		External:  []*FrontmatterExternal{},
		Input:     "",
		Outputs:   []*FrontmatterOutput{},
	}
}

// FrontmatterParse opens the file at filePath, extracts the YAML
// frontmatter delimited by --- lines, and unmarshals it into a
// Frontmatter struct. All missing fields are returned as empty slices
// or empty strings.
//
// Errors:
//   - ErrFileUnreadable: the file cannot be opened or read.
//   - ErrMalformedYAML: the content between --- delimiters is not
//     valid YAML.
//   - (FileReader.*): propagated from FileOpen.
func FrontmatterParse(filePath *pathutils.PathCfs) (*Frontmatter, error) {
	// Step 1: Open the file.
	reader, err := filereader.FileOpen(filePath)
	if err != nil {
		if errors.Is(err, filereader.ErrFileUnreadable) {
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		// Propagate PathUtils errors as-is.
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

	// Step 4: Collect YAML lines until the closing "---".
	var yamlLines []string
	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				filereader.FileClose(reader)
				return nil, fmt.Errorf("%w: unexpected end of file while reading frontmatter", ErrMalformedYAML)
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

	// Step 6: Join lines into a single YAML string.
	yamlText := strings.Join(yamlLines, "\n")

	// Step 7: Return empty Frontmatter if the YAML block is blank.
	if strings.TrimSpace(yamlText) == "" {
		return emptyFrontmatter(), nil
	}

	// Step 8: Parse the YAML text.
	var raw rawFrontmatter
	if err := yaml.Unmarshal([]byte(yamlText), &raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMalformedYAML, err)
	}

	// Step 9: Extract depends_on.
	dependsOn := raw.DependsOn
	if dependsOn == nil {
		dependsOn = []string{}
	}

	// Step 10: Extract input.
	input := raw.Input

	// Step 11: Extract and validate outputs.
	outputsList := make([]*FrontmatterOutput, 0, len(raw.Outputs))
	for _, entry := range raw.Outputs {
		if entry.ID == "" {
			return nil, fmt.Errorf("%w: output entry missing required field \"id\"", ErrMalformedYAML)
		}
		if entry.Path == "" {
			return nil, fmt.Errorf("%w: output entry missing required field \"path\"", ErrMalformedYAML)
		}
		outputsList = append(outputsList, &FrontmatterOutput{
			ID:   entry.ID,
			Path: entry.Path,
		})
	}

	// Step 12: Extract and validate external entries.
	externalList := make([]*FrontmatterExternal, 0, len(raw.External))
	for _, entry := range raw.External {
		if entry.Path == "" {
			return nil, fmt.Errorf("%w: external entry missing required field \"path\"", ErrMalformedYAML)
		}

		var fragmentsList []*FrontmatterExternalFragment
		if entry.Fragments != nil {
			fragmentsList = make([]*FrontmatterExternalFragment, 0, len(entry.Fragments))
			for _, frag := range entry.Fragments {
				if frag.Lines == "" {
					return nil, fmt.Errorf("%w: fragment entry missing required field \"lines\"", ErrMalformedYAML)
				}
				if frag.Hash == "" {
					return nil, fmt.Errorf("%w: fragment entry missing required field \"hash\"", ErrMalformedYAML)
				}
				fragmentsList = append(fragmentsList, &FrontmatterExternalFragment{
					Description: frag.Description,
					Lines:       frag.Lines,
					Hash:        frag.Hash,
				})
			}
		} else {
			fragmentsList = []*FrontmatterExternalFragment{}
		}

		externalList = append(externalList, &FrontmatterExternal{
			Path:      entry.Path,
			Fragments: fragmentsList,
		})
	}

	// Step 13: Return the populated Frontmatter.
	return &Frontmatter{
		DependsOn: dependsOn,
		External:  externalList,
		Input:     input,
		Outputs:   outputsList,
	}, nil
}
