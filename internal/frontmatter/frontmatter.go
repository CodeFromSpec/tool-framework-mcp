// code-from-spec: ROOT/golang/implementation/internal/frontmatter/code@tZW47GZDHusiIYHZfJZqkD7KCik

package frontmatter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
	"github.com/goccy/go-yaml"
)

// ErrMalformedYAML is returned when the YAML block between --- delimiters
// is not valid YAML or a required field in a sub-record is missing.
var ErrMalformedYAML = errors.New("malformed YAML")

// FrontmatterExternalFragment represents a single fragment of an external dependency.
type FrontmatterExternalFragment struct {
	Description *string
	Lines       string
	Hash        string
}

// FrontmatterExternal represents an external dependency entry.
type FrontmatterExternal struct {
	Path      string
	Fragments *[]FrontmatterExternalFragment
}

// FrontmatterOutput represents a single output entry.
type FrontmatterOutput struct {
	ID   string
	Path string
}

// Frontmatter holds all parsed frontmatter data from a spec file.
type Frontmatter struct {
	DependsOn []string
	External  []FrontmatterExternal
	Input     string
	Outputs   []FrontmatterOutput
}

// rawFragment is used for YAML unmarshalling of fragment entries.
type rawFragment struct {
	Description *string `yaml:"description"`
	Lines       *string `yaml:"lines"`
	Hash        *string `yaml:"hash"`
}

// rawExternal is used for YAML unmarshalling of external entries.
type rawExternal struct {
	Path      string        `yaml:"path"`
	Fragments []rawFragment `yaml:"fragments"`
}

// rawOutput is used for YAML unmarshalling of output entries.
type rawOutput struct {
	ID   *string `yaml:"id"`
	Path *string `yaml:"path"`
}

// rawFrontmatter is used for YAML unmarshalling of the full frontmatter block.
type rawFrontmatter struct {
	DependsOn []string      `yaml:"depends_on"`
	External  []rawExternal `yaml:"external"`
	Input     string        `yaml:"input"`
	Outputs   []rawOutput   `yaml:"outputs"`
}

// FrontmatterParse opens the file at file_path and parses any YAML frontmatter
// delimited by --- markers at the top of the file.
//
// Possible errors:
//   - Path errors propagated from FileOpen (ErrPathEmpty, ErrPathAbsolute, etc.)
//   - filereader.ErrFileUnreadable if the file cannot be opened.
//   - ErrMalformedYAML if the YAML block is invalid or required fields are missing.
func FrontmatterParse(file_path *pathutils.PathCfs) (Frontmatter, error) {
	// Step 1: Open the file.
	reader, err := filereader.FileOpen(file_path)
	if err != nil {
		return Frontmatter{}, fmt.Errorf("FrontmatterParse: %w", err)
	}

	// Step 2: Read the first line.
	firstLine, err := filereader.FileReadLine(reader)
	if err != nil {
		if errors.Is(err, filereader.ErrEndOfFile) {
			filereader.FileClose(reader)
			return Frontmatter{}, nil
		}
		filereader.FileClose(reader)
		return Frontmatter{}, fmt.Errorf("FrontmatterParse: %w", err)
	}

	// Step 3: Check that the first line is exactly "---".
	if firstLine != "---" {
		filereader.FileClose(reader)
		return Frontmatter{}, nil
	}

	// Step 4: Collect YAML lines until the closing "---".
	var yamlLines []string
	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				filereader.FileClose(reader)
				return Frontmatter{}, fmt.Errorf("FrontmatterParse: %w", ErrMalformedYAML)
			}
			filereader.FileClose(reader)
			return Frontmatter{}, fmt.Errorf("FrontmatterParse: %w", err)
		}
		if line == "---" {
			break
		}
		yamlLines = append(yamlLines, line)
	}

	// Step 5: Close the reader.
	filereader.FileClose(reader)

	// Step 6: If no YAML lines were collected, return an empty Frontmatter.
	if len(yamlLines) == 0 {
		return Frontmatter{}, nil
	}

	// Step 7: Join and parse the YAML.
	yamlText := strings.Join(yamlLines, "\n")
	var raw rawFrontmatter
	if err := yaml.Unmarshal([]byte(yamlText), &raw); err != nil {
		return Frontmatter{}, fmt.Errorf("FrontmatterParse: %w: %w", ErrMalformedYAML, err)
	}

	// Step 8: Build the Frontmatter record from parsed YAML.

	// depends_on
	dependsOn := raw.DependsOn
	if dependsOn == nil {
		dependsOn = []string{}
	}

	// external
	external := []FrontmatterExternal{}
	for _, rawExt := range raw.External {
		if rawExt.Path == "" {
			return Frontmatter{}, fmt.Errorf("FrontmatterParse: %w: external entry missing required field \"path\"", ErrMalformedYAML)
		}

		var fragments *[]FrontmatterExternalFragment
		if rawExt.Fragments != nil {
			built := make([]FrontmatterExternalFragment, 0, len(rawExt.Fragments))
			for _, rawFrag := range rawExt.Fragments {
				if rawFrag.Lines == nil {
					return Frontmatter{}, fmt.Errorf("FrontmatterParse: %w: fragment entry missing required field \"lines\"", ErrMalformedYAML)
				}
				if rawFrag.Hash == nil {
					return Frontmatter{}, fmt.Errorf("FrontmatterParse: %w: fragment entry missing required field \"hash\"", ErrMalformedYAML)
				}
				built = append(built, FrontmatterExternalFragment{
					Description: rawFrag.Description,
					Lines:       *rawFrag.Lines,
					Hash:        *rawFrag.Hash,
				})
			}
			fragments = &built
		}

		external = append(external, FrontmatterExternal{
			Path:      rawExt.Path,
			Fragments: fragments,
		})
	}

	// outputs
	outputs := []FrontmatterOutput{}
	for _, rawOut := range raw.Outputs {
		if rawOut.ID == nil {
			return Frontmatter{}, fmt.Errorf("FrontmatterParse: %w: output entry missing required field \"id\"", ErrMalformedYAML)
		}
		if rawOut.Path == nil {
			return Frontmatter{}, fmt.Errorf("FrontmatterParse: %w: output entry missing required field \"path\"", ErrMalformedYAML)
		}
		outputs = append(outputs, FrontmatterOutput{
			ID:   *rawOut.ID,
			Path: *rawOut.Path,
		})
	}

	// Step 9: Return the Frontmatter record.
	return Frontmatter{
		DependsOn: dependsOn,
		External:  external,
		Input:     raw.Input,
		Outputs:   outputs,
	}, nil
}
