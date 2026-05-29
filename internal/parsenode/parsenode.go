// code-from-spec: ROOT/golang/implementation/parsing/node_parsing@aKtZJsLzv_roopmBlt_MgfJ2P8Y

package parsenode

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
)

// NodeSubsection represents a level-2 heading and its content within
// a section. The heading is stored in normalized form (after NormalizeText).
// Content contains raw markdown text with only leading/trailing blank lines
// trimmed.
type NodeSubsection struct {
	Heading string
	Content string
}

// NodeSection represents a level-1 heading section within a node file.
// The heading is stored in normalized form (after NormalizeText). Content
// contains raw markdown text between the section heading and the first
// subsection (or end of section), with only leading/trailing blank lines
// trimmed. A section that exists in the file but has no content is present
// with an empty Content and an empty Subsections slice.
type NodeSection struct {
	Heading     string
	Content     string
	Subsections []*NodeSubsection
}

// Node represents a parsed _node.md file. Each field corresponds to a
// well-known section of the file, except Private which collects all
// remaining sections in the order they appear.
type Node struct {
	// NameSection is the first level-1 section, whose heading matches
	// the logical name after normalization.
	NameSection *NodeSection

	// Public is the `# Public` section, if present.
	Public *NodeSection

	// Agent is the `# Agent` section, if present.
	Agent *NodeSection

	// Private contains all other level-1 sections, in file order.
	Private []*NodeSection
}

var (
	// ErrNotRootReference is returned when the logical name does not
	// start with "ROOT/".
	ErrNotRootReference = errors.New("not a ROOT reference")

	// ErrHasQualifier is returned when the logical name contains a
	// parenthetical qualifier.
	ErrHasQualifier = errors.New("has qualifier")

	// ErrFileUnreadable is returned when the node file cannot be opened
	// or read.
	ErrFileUnreadable = errors.New("file unreadable")

	// ErrUnexpectedContentBeforeFirstHeading is returned when the file
	// body has non-blank content before the first level-1 heading, or
	// has no level-1 heading at all.
	ErrUnexpectedContentBeforeFirstHeading = errors.New("unexpected content before first heading")

	// ErrNodeNameDoesNotMatch is returned when the first heading does
	// not match the logical name after normalization.
	ErrNodeNameDoesNotMatch = errors.New("node name does not match")

	// ErrDuplicatePublicSection is returned when more than one `# Public`
	// section exists in the file.
	ErrDuplicatePublicSection = errors.New("duplicate public section")

	// ErrDuplicateAgentSection is returned when more than one `# Agent`
	// section exists in the file.
	ErrDuplicateAgentSection = errors.New("duplicate agent section")

	// ErrDuplicateSubsection is returned when two `##` headings within
	// the same section normalize to the same text.
	ErrDuplicateSubsection = errors.New("duplicate subsection")
)

// event kinds used during tokenization
const (
	eventContent  = iota
	eventHeading1
	eventHeading2
)

type event struct {
	kind int
	text string // normalized heading text for headings; raw line for content
}

// NodeParse parses the _node.md file associated with the given logical name
// and returns a Node representation of its contents.
func NodeParse(logical_name string) (*Node, error) {
	// Step 1: check for artifact reference
	if logicalnames.LogicalNameIsArtifact(logical_name) {
		return nil, ErrNotRootReference
	}

	// Step 2: check for qualifier
	if logicalnames.LogicalNameHasQualifier(logical_name) {
		return nil, ErrHasQualifier
	}

	// Step 3: resolve logical name to file path
	cfsPath, err := logicalnames.LogicalNameToPath(logical_name)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	// Step 4: open file
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		if errors.Is(err, filereader.ErrFileUnreadable) {
			return nil, ErrFileUnreadable
		}
		return nil, fmt.Errorf("%w", err)
	}

	// Step 5: skip frontmatter
	var bodyLines []string
	firstLine, err := filereader.FileReadLine(reader)
	if err != nil {
		filereader.FileClose(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			return nil, ErrUnexpectedContentBeforeFirstHeading
		}
		return nil, fmt.Errorf("%w", err)
	}

	if firstLine == "---" {
		// consume lines until closing "---"
		foundClosing := false
		for {
			line, err := filereader.FileReadLine(reader)
			if err != nil {
				filereader.FileClose(reader)
				if errors.Is(err, filereader.ErrEndOfFile) {
					return nil, ErrUnexpectedContentBeforeFirstHeading
				}
				return nil, fmt.Errorf("%w", err)
			}
			if line == "---" {
				foundClosing = true
				break
			}
		}
		_ = foundClosing
	} else {
		// first line is body content — carry it forward
		bodyLines = append(bodyLines, firstLine)
	}

	// Step 6: read remaining body lines
	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				break
			}
			filereader.FileClose(reader)
			return nil, fmt.Errorf("%w", err)
		}
		bodyLines = append(bodyLines, line)
	}
	filereader.FileClose(reader)

	// Step 7: tokenize lines into events
	events := tokenize(bodyLines)

	// Step 8: build sections from event stream
	sections, foundNonBlank, err := buildSections(events)
	if err != nil {
		return nil, err
	}

	// Step 9: validate pre-heading content
	if foundNonBlank {
		return nil, ErrUnexpectedContentBeforeFirstHeading
	}

	// Step 10: validate at least one section
	if len(sections) == 0 {
		return nil, ErrUnexpectedContentBeforeFirstHeading
	}

	// Step 11: classify sections
	node, err := classifySections(sections, logical_name)
	if err != nil {
		return nil, err
	}

	return node, nil
}

// tokenize converts lines into a stream of events, respecting fenced code blocks.
func tokenize(lines []string) []event {
	events := make([]event, 0, len(lines))

	inFence := false
	fenceChar := rune(0)
	fenceLength := 0

	for _, line := range lines {
		if inFence {
			// check if this is the closing fence
			if isClosingFence(line, fenceChar, fenceLength) {
				inFence = false
			}
			// emit as content regardless
			events = append(events, event{kind: eventContent, text: line})
			continue
		}

		// check for opening fence
		if fc, fl, ok := parseOpeningFence(line); ok {
			inFence = true
			fenceChar = fc
			fenceLength = fl
			events = append(events, event{kind: eventContent, text: line})
			continue
		}

		// try to parse as ATX heading
		if level, heading, ok := parseATXHeading(line); ok {
			if level == 1 {
				events = append(events, event{kind: eventHeading1, text: heading})
			} else if level == 2 {
				events = append(events, event{kind: eventHeading2, text: heading})
			} else {
				// level 3+ becomes content
				events = append(events, event{kind: eventContent, text: line})
			}
			continue
		}

		events = append(events, event{kind: eventContent, text: line})
	}

	return events
}

// isClosingFence returns true if line is a valid closing fence for the given
// fence character and minimum length.
func isClosingFence(line string, fenceChar rune, fenceLength int) bool {
	trimmed := strings.TrimRight(line, " \t")
	if len(trimmed) < fenceLength {
		return false
	}
	for _, ch := range trimmed {
		if ch != fenceChar {
			return false
		}
	}
	return true
}

// parseOpeningFence detects an opening fence line. Returns the fence character,
// length, and true on success.
func parseOpeningFence(line string) (rune, int, bool) {
	if len(line) == 0 {
		return 0, 0, false
	}

	var fenceChar rune
	if line[0] == '`' {
		fenceChar = '`'
	} else if line[0] == '~' {
		fenceChar = '~'
	} else {
		return 0, 0, false
	}

	count := 0
	for _, ch := range line {
		if ch == fenceChar {
			count++
		} else {
			break
		}
	}

	if count < 3 {
		return 0, 0, false
	}

	return fenceChar, count, true
}

// parseATXHeading parses an ATX heading line. Returns level, normalized heading
// text, and true on success.
func parseATXHeading(line string) (int, string, bool) {
	if len(line) == 0 || line[0] != '#' {
		return 0, "", false
	}

	// count leading '#' characters
	level := 0
	for level < len(line) && line[level] == '#' {
		level++
	}

	// must be followed by at least one space
	if level >= len(line) || line[level] != ' ' {
		return 0, "", false
	}

	// extract text after the leading "# " prefix
	text := line[level+1:]
	text = strings.TrimSpace(text)

	// strip optional closing '#' sequence: must be preceded by a space
	if idx := strings.LastIndex(text, " "); idx >= 0 {
		suffix := text[idx+1:]
		allHash := len(suffix) > 0
		for _, ch := range suffix {
			if ch != '#' {
				allHash = false
				break
			}
		}
		if allHash {
			text = strings.TrimSpace(text[:idx])
		}
	}

	normalized := textnormalization.NormalizeText(text)
	return level, normalized, true
}

// buildSections processes the event stream and builds NodeSection records.
// Returns the sections, whether non-blank content was found before the first
// heading, and any error.
func buildSections(events []event) ([]*NodeSection, bool, error) {
	var sections []*NodeSection
	var currentSection *NodeSection
	var currentSubsection *NodeSubsection
	var currentSectionContentLines []string
	var currentSubsectionContentLines []string
	foundNonBlankBeforeFirstHeading := false

	flushSubsection := func() error {
		if currentSubsection != nil {
			currentSubsection.Content = trimBlankLines(currentSubsectionContentLines)
			currentSection.Subsections = append(currentSection.Subsections, currentSubsection)
			currentSubsection = nil
			currentSubsectionContentLines = nil
		}
		return nil
	}

	flushSection := func() error {
		if err := flushSubsection(); err != nil {
			return err
		}
		if currentSection != nil {
			currentSection.Content = trimBlankLines(currentSectionContentLines)
			sections = append(sections, currentSection)
			currentSection = nil
			currentSectionContentLines = nil
		}
		return nil
	}

	for _, ev := range events {
		switch ev.kind {
		case eventHeading1:
			if err := flushSection(); err != nil {
				return nil, false, err
			}
			currentSection = &NodeSection{
				Heading:     ev.text,
				Subsections: []*NodeSubsection{},
			}
			currentSectionContentLines = nil

		case eventHeading2:
			if currentSection == nil {
				// no level-1 heading yet — treat as content
				if !isBlankLine(ev.text) {
					foundNonBlankBeforeFirstHeading = true
				}
			} else {
				if err := flushSubsection(); err != nil {
					return nil, false, err
				}
				// check for duplicate subsection
				for _, existing := range currentSection.Subsections {
					if existing.Heading == ev.text {
						return nil, false, ErrDuplicateSubsection
					}
				}
				currentSubsection = &NodeSubsection{
					Heading: ev.text,
				}
				currentSubsectionContentLines = nil
			}

		case eventContent:
			if currentSection == nil {
				if !isBlankLine(ev.text) {
					foundNonBlankBeforeFirstHeading = true
				}
			} else if currentSubsection != nil {
				currentSubsectionContentLines = append(currentSubsectionContentLines, ev.text)
			} else {
				currentSectionContentLines = append(currentSectionContentLines, ev.text)
			}
		}
	}

	if err := flushSection(); err != nil {
		return nil, false, err
	}

	return sections, foundNonBlankBeforeFirstHeading, nil
}

// classifySections maps sections into the Node struct fields.
func classifySections(sections []*NodeSection, logicalName string) (*Node, error) {
	normalizedLogicalName := textnormalization.NormalizeText(logicalName)

	var nameSection *NodeSection
	var publicSection *NodeSection
	var agentSection *NodeSection
	var privateSections []*NodeSection

	for _, section := range sections {
		if nameSection == nil {
			// first section — must match the logical name
			if section.Heading != normalizedLogicalName {
				return nil, ErrNodeNameDoesNotMatch
			}
			nameSection = section
			continue
		}

		switch section.Heading {
		case "public":
			if publicSection != nil {
				return nil, ErrDuplicatePublicSection
			}
			publicSection = section

		case "agent":
			if agentSection != nil {
				return nil, ErrDuplicateAgentSection
			}
			agentSection = section

		default:
			privateSections = append(privateSections, section)
		}
	}

	return &Node{
		NameSection: nameSection,
		Public:      publicSection,
		Agent:       agentSection,
		Private:     privateSections,
	}, nil
}

// trimBlankLines removes leading and trailing blank lines from a slice,
// then joins the remainder with newlines.
func trimBlankLines(lines []string) string {
	start := 0
	for start < len(lines) && isBlankLine(lines[start]) {
		start++
	}
	end := len(lines)
	for end > start && isBlankLine(lines[end-1]) {
		end--
	}
	return strings.Join(lines[start:end], "\n")
}

// isBlankLine returns true if the line is empty or contains only whitespace.
func isBlankLine(line string) bool {
	for _, ch := range line {
		if !unicode.IsSpace(ch) {
			return false
		}
	}
	return true
}
