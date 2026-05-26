// code-from-spec: ROOT/golang/internal/frontmatter/code@LABQubMqG5YZvza9cUUS89hjTlI

// Package frontmatter parses the optional YAML frontmatter block at the top
// of a spec node file (_node.md). The parser stops as soon as the closing
// "---" delimiter is found and never reads the file body, keeping memory use
// proportional to the frontmatter block only.
package frontmatter

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

// ---------------------------------------------------------------------------
// Sentinel errors
// ---------------------------------------------------------------------------

// ErrRead is returned (wrapped) when the target file cannot be read.
var ErrRead = errors.New("error reading file")

// ErrFrontmatterParse is returned (wrapped) when the YAML between the "---"
// delimiters is malformed.
var ErrFrontmatterParse = errors.New("error parsing frontmatter")

// ---------------------------------------------------------------------------
// Public types
// ---------------------------------------------------------------------------

// Output represents one artifact that a spec leaf node produces.
type Output struct {
	ID   string
	Path string
}

// ExternalFragment identifies a sub-range of an external file that the spec
// node depends on.
type ExternalFragment struct {
	Description string
	Lines       string
	Hash        string
}

// External describes an external file dependency, optionally restricted to
// specific fragments.
type External struct {
	Path      string
	Fragments []ExternalFragment
}

// Frontmatter holds all structured metadata extracted from a spec node file.
// Fields absent in the YAML default to their zero values (empty slice / empty
// string).
type Frontmatter struct {
	DependsOn []string
	External  []External
	Input     string
	Outputs   []Output
}

// ---------------------------------------------------------------------------
// Unexported YAML mirror types
// ---------------------------------------------------------------------------
// We unmarshal into these private structs (with yaml tags) and then convert to
// the exported types. This keeps the public API clean.

type yamlExternalFragment struct {
	Description string `yaml:"description"`
	Lines       string `yaml:"lines"`
	Hash        string `yaml:"hash"`
}

type yamlExternal struct {
	Path      string                 `yaml:"path"`
	Fragments []yamlExternalFragment `yaml:"fragments"`
}

type yamlOutput struct {
	ID   string `yaml:"id"`
	Path string `yaml:"path"`
}

type yamlFrontmatter struct {
	DependsOn []string       `yaml:"depends_on"`
	External  []yamlExternal `yaml:"external"`
	Input     string         `yaml:"input"`
	Outputs   []yamlOutput   `yaml:"outputs"`
}

// ---------------------------------------------------------------------------
// ParseFrontmatter
// ---------------------------------------------------------------------------

// ParseFrontmatter reads the file at filePath, extracts the YAML frontmatter
// block (if present), and returns the parsed result.
//
// Rules:
//   - If the file has no opening "---" delimiter, an empty Frontmatter is
//     returned (not an error).
//   - If the opening delimiter is present but the closing "---" is never
//     found (EOF reached first), ErrFrontmatterParse is returned.
//   - If the YAML between the delimiters is malformed, ErrFrontmatterParse is
//     returned.
//   - If the file cannot be read, ErrRead is returned.
//
// All returned errors wrap one of the two sentinels so callers can match with
// errors.Is().
func ParseFrontmatter(filePath string) (*Frontmatter, error) {
	// Step 1: Read the file.
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrRead, filePath, err)
	}

	// Split the entire file into lines for scanning. We only process lines up
	// to and including the closing "---", so memory overhead is minimal.
	lines := strings.Split(string(data), "\n")

	// Use a scanner-style index so the logic mirrors the pseudocode closely.
	scanner := &lineScanner{lines: lines}

	// Step 2: Read the first line.
	firstLine, ok := scanner.next()
	if !ok {
		// EOF on first read — file is empty; return empty Frontmatter.
		return &Frontmatter{}, nil
	}

	// Normalize the line (trim trailing \r from CRLF files).
	if strings.TrimSpace(firstLine) != "---" {
		// No opening delimiter — return empty Frontmatter; not an error.
		return &Frontmatter{}, nil
	}

	// Step 3: Collect lines until the closing "---" or EOF.
	var buf []string
	for {
		line, ok := scanner.next()
		if !ok {
			// EOF reached without a closing delimiter — malformed.
			return nil, fmt.Errorf("%w: %s: missing closing '---' delimiter", ErrFrontmatterParse, filePath)
		}
		if strings.TrimRight(line, "\r") == "---" {
			// Closing delimiter found — stop collecting.
			break
		}
		// Normalize CRLF line endings.
		buf = append(buf, strings.TrimRight(line, "\r"))
	}

	// Step 4: Build the YAML string from collected lines.
	// An empty buffer (block was ---\n---) is valid; yaml.Unmarshal handles it
	// gracefully (nothing to parse → all fields remain zero-valued).
	yamlStr := strings.Join(buf, "\n")

	// Step 5 & 6: Parse the YAML into the mirror struct, then convert.
	var raw yamlFrontmatter
	if len(strings.TrimSpace(yamlStr)) > 0 {
		if err := yaml.Unmarshal([]byte(yamlStr), &raw); err != nil {
			return nil, fmt.Errorf("%w: %s: %w", ErrFrontmatterParse, filePath, err)
		}
	}

	// Step 7: Convert the mirror struct to the exported Frontmatter type.
	fm := &Frontmatter{
		DependsOn: raw.DependsOn,
		Input:     raw.Input,
	}

	// Ensure slice fields are never nil for consistent caller behaviour.
	if fm.DependsOn == nil {
		fm.DependsOn = []string{}
	}

	// Convert External entries.
	fm.External = make([]External, 0, len(raw.External))
	for _, re := range raw.External {
		ext := External{
			Path: re.Path,
		}
		if len(re.Fragments) > 0 {
			ext.Fragments = make([]ExternalFragment, 0, len(re.Fragments))
			for _, rf := range re.Fragments {
				ext.Fragments = append(ext.Fragments, ExternalFragment{
					Description: rf.Description,
					Lines:       rf.Lines,
					Hash:        rf.Hash,
				})
			}
		} else {
			ext.Fragments = []ExternalFragment{}
		}
		fm.External = append(fm.External, ext)
	}

	// Convert Output entries.
	fm.Outputs = make([]Output, 0, len(raw.Outputs))
	for _, ro := range raw.Outputs {
		fm.Outputs = append(fm.Outputs, Output{
			ID:   ro.ID,
			Path: ro.Path,
		})
	}

	return fm, nil
}

// ---------------------------------------------------------------------------
// lineScanner — a minimal line-by-line cursor over a pre-split slice.
// ---------------------------------------------------------------------------

// lineScanner iterates over a slice of lines without reading from disk again.
// It mirrors the "ReadLine" abstraction described in the pseudocode.
type lineScanner struct {
	lines []string
	pos   int
}

// next returns the next line and true, or ("", false) when all lines have been
// consumed (equivalent to "end of file" in the pseudocode).
func (s *lineScanner) next() (string, bool) {
	if s.pos >= len(s.lines) {
		return "", false
	}
	line := s.lines[s.pos]
	s.pos++
	return line, true
}

// bufioScannerExample shows an alternative implementation approach using
// bufio.Scanner (kept as a reference comment only; not used at runtime).
var _ = bufio.NewScanner // import kept to satisfy potential future uses
