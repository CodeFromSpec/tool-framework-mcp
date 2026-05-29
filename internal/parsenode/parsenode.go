// code-from-spec: ROOT/golang/implementation/parsing/node_parsing@hNT4dtvm77gTSFY_aDE0KV7-MhA

package parsenode

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
)

// NodeSubsection represents a level-2 (##) heading section within a Node section.
// Heading is the normalized form of the heading text, used for comparisons and
// lookups. RawHeading is the original heading line as read from the file, preserved
// for hashing. Content holds each line of the subsection body exactly as read.
type NodeSubsection struct {
	Heading    string
	RawHeading string
	Content    []string
}

// NodeSection represents a level-1 (#) heading section within a Node.
// Heading is the normalized form of the heading text, used for comparisons and
// lookups. RawHeading is the original heading line as read from the file, preserved
// for hashing. Content holds each line of the section body before the first ##
// heading, exactly as read. Subsections holds the ordered list of ## subsections.
type NodeSection struct {
	Heading     string
	RawHeading  string
	Content     []string
	Subsections []*NodeSubsection
}

// Node represents a parsed spec node file. NameSection is the first level-1
// heading section, whose heading matches the logical name. Public is the optional
// "# Public" section. Agent is the optional "# Agent" section. Private holds all
// other sections in the order they appear in the file.
type Node struct {
	NameSection *NodeSection
	Public      *NodeSection
	Agent       *NodeSection
	Private     []*NodeSection
}

var (
	// ErrNotARootReference is returned when the logical name does not start with "ROOT/".
	ErrNotARootReference = errors.New("not a ROOT reference")

	// ErrHasQualifier is returned when the logical name contains a parenthetical qualifier.
	ErrHasQualifier = errors.New("has qualifier")

	// ErrFileUnreadable is returned when the file cannot be opened or read.
	ErrFileUnreadable = errors.New("file unreadable")

	// ErrUnexpectedContentBeforeFirstHeading is returned when the file body has
	// non-blank content before the first level-1 heading, or has no level-1 heading at all.
	ErrUnexpectedContentBeforeFirstHeading = errors.New("unexpected content before first heading")

	// ErrNodeNameDoesNotMatch is returned when the first heading does not match
	// the logical name after normalization.
	ErrNodeNameDoesNotMatch = errors.New("node name does not match")

	// ErrDuplicatePublicSection is returned when more than one "# Public" section exists.
	ErrDuplicatePublicSection = errors.New("duplicate public section")

	// ErrDuplicateAgentSection is returned when more than one "# Agent" section exists.
	ErrDuplicateAgentSection = errors.New("duplicate agent section")

	// ErrDuplicateSubsection is returned when two ## headings within the same section
	// normalize to the same text.
	ErrDuplicateSubsection = errors.New("duplicate subsection")
)

// headingResult holds the parsed fields of a heading line.
type headingResult struct {
	level      int
	text       string // raw extracted text (trimmed)
	normalized string
	rawLine    string // original line unchanged
}

// isBlank returns true if the line contains only whitespace characters.
func isBlank(line string) bool {
	return strings.TrimSpace(line) == ""
}

// parseHeading attempts to parse a Markdown heading from line.
// Returns nil if the line is not a valid heading.
func parseHeading(line string) *headingResult {
	if len(line) == 0 {
		return nil
	}

	// Count leading '#' characters.
	level := 0
	for level < len(line) && line[level] == '#' {
		level++
	}
	if level == 0 {
		return nil
	}

	// The character immediately after the '#' sequence must be a space.
	if level >= len(line) || line[level] != ' ' {
		return nil
	}

	// Extract the text after the leading "### " prefix.
	text := line[level+1:]

	// Strip trailing ATX closing sequence: optional spaces followed by one or more '#'.
	trimmed := strings.TrimRight(text, " ")
	if strings.HasSuffix(trimmed, "#") {
		// Find the last space before trailing '#' characters.
		i := len(trimmed) - 1
		for i >= 0 && trimmed[i] == '#' {
			i--
		}
		if i >= 0 && trimmed[i] == ' ' {
			text = strings.TrimRight(trimmed[:i], " ")
		}
	}

	text = strings.TrimSpace(text)
	normalized := textnormalization.NormalizeText(text)

	return &headingResult{
		level:      level,
		text:       text,
		normalized: normalized,
		rawLine:    line,
	}
}

// NodeParse parses the spec file for the given logical name and returns a Node.
//
// The logical name must start with "ROOT/" and must not contain a parenthetical
// qualifier. The corresponding file is located via FileOpen. The file is parsed
// into sections and subsections according to level-1 (#) and level-2 (##) headings.
func NodeParse(logical_name string) (*Node, error) {
	// Step 1: Check for artifact reference.
	if logicalnames.LogicalNameIsArtifact(logical_name) {
		return nil, fmt.Errorf("%w", ErrNotARootReference)
	}

	// Step 2: Check for qualifier.
	if logicalnames.LogicalNameHasQualifier(logical_name) {
		return nil, fmt.Errorf("%w", ErrHasQualifier)
	}

	// Step 3: Resolve path.
	cfsPath, err := logicalnames.LogicalNameToPath(logical_name)
	if err != nil {
		return nil, fmt.Errorf("resolving logical name: %w", err)
	}

	// Step 4: Open file.
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		if errors.Is(err, filereader.ErrFileUnreadable) {
			return nil, fmt.Errorf("%w", ErrFileUnreadable)
		}
		return nil, fmt.Errorf("opening file: %w", err)
	}

	// close_and_raise helper: closes reader then returns the given error.
	closeAndRaise := func(e error) error {
		filereader.FileClose(reader)
		return e
	}

	// Step 6: Skip frontmatter.
	firstLine, err := filereader.FileReadLine(reader)
	if err != nil {
		if errors.Is(err, filereader.ErrEndOfFile) {
			return nil, closeAndRaise(ErrUnexpectedContentBeforeFirstHeading)
		}
		return nil, closeAndRaise(fmt.Errorf("%w", ErrFileUnreadable))
	}

	var pendingLine string
	hasPending := false

	if firstLine == "---" {
		// Skip frontmatter block: read until closing "---".
		for {
			line, err := filereader.FileReadLine(reader)
			if err != nil {
				if errors.Is(err, filereader.ErrEndOfFile) {
					return nil, closeAndRaise(ErrUnexpectedContentBeforeFirstHeading)
				}
				return nil, closeAndRaise(fmt.Errorf("%w", ErrFileUnreadable))
			}
			if line == "---" {
				break
			}
		}
	} else {
		// Not frontmatter: first line is the first body line.
		pendingLine = firstLine
		hasPending = true
	}

	// Step 7: Parse the body into sections.
	var (
		currentSection    *NodeSection
		currentSubsection *NodeSubsection
		sections          []*NodeSection
		hasPublic         bool
		hasAgent          bool
		foundFirstHeading bool
		inFence           bool
		fenceChar         byte
		fenceLength       int
		preHeadingLines   []string
	)

	appendContent := func(line string) {
		if currentSubsection != nil {
			currentSubsection.Content = append(currentSubsection.Content, line)
		} else if currentSection != nil {
			currentSection.Content = append(currentSection.Content, line)
		}
	}

	finalizeSubsection := func() {
		if currentSubsection != nil {
			currentSection.Subsections = append(currentSection.Subsections, currentSubsection)
			currentSubsection = nil
		}
	}

	finalizeSection := func() {
		finalizeSubsection()
		if currentSection != nil {
			sections = append(sections, currentSection)
			currentSection = nil
		}
	}

	openSubsection := func(rawLine, normalized string) error {
		if currentSubsection != nil {
			currentSection.Subsections = append(currentSection.Subsections, currentSubsection)
			currentSubsection = nil
		}
		for _, sub := range currentSection.Subsections {
			if sub.Heading == normalized {
				return ErrDuplicateSubsection
			}
		}
		currentSubsection = &NodeSubsection{
			Heading:    normalized,
			RawHeading: rawLine,
			Content:    []string{},
		}
		return nil
	}

	openSection := func(rawLine, normalized string) {
		finalizeSection()
		currentSection = &NodeSection{
			Heading:     normalized,
			RawHeading:  rawLine,
			Content:     []string{},
			Subsections: []*NodeSubsection{},
		}
	}

	for {
		// Read next line.
		var line string
		if hasPending {
			line = pendingLine
			hasPending = false
		} else {
			line, err = filereader.FileReadLine(reader)
			if err != nil {
				if errors.Is(err, filereader.ErrEndOfFile) {
					break
				}
				return nil, closeAndRaise(fmt.Errorf("%w", ErrFileUnreadable))
			}
		}

		// Fence tracking.
		if !inFence {
			// Count leading backticks.
			bt := 0
			for bt < len(line) && line[bt] == '`' {
				bt++
			}
			if bt >= 3 {
				rest := line[bt:]
				if !strings.ContainsRune(rest, '`') {
					inFence = true
					fenceChar = '`'
					fenceLength = bt
					appendContent(line)
					continue
				}
			} else {
				// Count leading tildes.
				ti := 0
				for ti < len(line) && line[ti] == '~' {
					ti++
				}
				if ti >= 3 {
					inFence = true
					fenceChar = '~'
					fenceLength = ti
					appendContent(line)
					continue
				}
			}
			// Not a fence opener — fall through to heading detection.
		} else {
			// Inside a fence: look for closing fence.
			cl := 0
			for cl < len(line) && line[cl] == fenceChar {
				cl++
			}
			if cl >= fenceLength {
				rest := line[cl:]
				if !strings.ContainsRune(rest, rune(fenceChar)) {
					inFence = false
				}
			}
			appendContent(line)
			continue
		}

		// Heading detection (only when not in fence).
		h := parseHeading(line)
		if h == nil {
			// Not a heading — treat as content.
			if !foundFirstHeading {
				if isBlank(line) {
					preHeadingLines = append(preHeadingLines, line)
					continue
				}
				return nil, closeAndRaise(ErrUnexpectedContentBeforeFirstHeading)
			}
			appendContent(line)
			continue
		}

		// Process heading by level.
		switch {
		case h.level == 1:
			if !foundFirstHeading {
				// This is the name section heading.
				expected := textnormalization.NormalizeText(logical_name)
				if h.normalized != expected {
					return nil, closeAndRaise(ErrNodeNameDoesNotMatch)
				}
				foundFirstHeading = true
				openSection(h.rawLine, h.normalized)
				preHeadingLines = nil // discard blank lines before first heading
			} else {
				openSection(h.rawLine, h.normalized)
				if h.normalized == "public" {
					if hasPublic {
						return nil, closeAndRaise(ErrDuplicatePublicSection)
					}
					hasPublic = true
				} else if h.normalized == "agent" {
					if hasAgent {
						return nil, closeAndRaise(ErrDuplicateAgentSection)
					}
					hasAgent = true
				}
			}

		case h.level == 2:
			if !foundFirstHeading {
				return nil, closeAndRaise(ErrUnexpectedContentBeforeFirstHeading)
			}
			if currentSection == nil {
				return nil, closeAndRaise(ErrUnexpectedContentBeforeFirstHeading)
			}
			if err := openSubsection(h.rawLine, h.normalized); err != nil {
				return nil, closeAndRaise(err)
			}

		default:
			// Level >= 3: treat as content.
			if !foundFirstHeading {
				return nil, closeAndRaise(ErrUnexpectedContentBeforeFirstHeading)
			}
			appendContent(line)
		}
	}

	// Step 8: End of file reached.
	finalizeSection()

	if !foundFirstHeading {
		filereader.FileClose(reader)
		return nil, ErrUnexpectedContentBeforeFirstHeading
	}

	// Step 9: Close the file.
	filereader.FileClose(reader)

	// Step 10: Assemble the Node record.
	if len(sections) == 0 {
		return nil, ErrUnexpectedContentBeforeFirstHeading
	}

	node := &Node{
		NameSection: sections[0],
		Private:     []*NodeSection{},
	}

	for _, section := range sections[1:] {
		switch section.Heading {
		case "public":
			node.Public = section
		case "agent":
			node.Agent = section
		default:
			node.Private = append(node.Private, section)
		}
	}

	return node, nil
}
