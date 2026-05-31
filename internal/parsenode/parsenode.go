// code-from-spec: ROOT/golang/implementation/parsing/node_parsing@WTh3Fs6NJebqr2R8TtWgJIujRvc
package parsenode

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
)

// ErrNotARootReference is returned when the logical name does not
// start with ROOT/.
var ErrNotARootReference = errors.New("logical name does not start with ROOT/")

// ErrHasQualifier is returned when the logical name contains a
// parenthetical qualifier.
var ErrHasQualifier = errors.New("logical name contains a parenthetical qualifier")

// ErrFileUnreadable is returned when the node file cannot be opened
// or read.
var ErrFileUnreadable = errors.New("node file cannot be opened or read")

// ErrUnexpectedContentBeforeFirstHeading is returned when the file
// body has non-blank content before the first level-1 heading, or has
// no level-1 heading at all.
var ErrUnexpectedContentBeforeFirstHeading = errors.New("unexpected content before first heading or no level-1 heading found")

// ErrNodeNameDoesNotMatch is returned when the first heading does not
// match the logical name after normalization.
var ErrNodeNameDoesNotMatch = errors.New("first heading does not match the logical name after normalization")

// ErrDuplicatePublicSection is returned when more than one Public
// section exists in the file.
var ErrDuplicatePublicSection = errors.New("more than one Public section exists")

// ErrDuplicateAgentSection is returned when more than one Agent
// section exists in the file.
var ErrDuplicateAgentSection = errors.New("more than one Agent section exists")

// ErrDuplicateSubsection is returned when two level-2 headings within
// the same section normalize to the same text.
var ErrDuplicateSubsection = errors.New("duplicate subsection heading within a section")

// NodeSubsection represents a level-2 (##) heading block within a section.
// Heading is the normalized form used for comparisons and lookups.
// RawHeading is the original line as read from the file, preserved for hashing.
// Content holds each line of the subsection body as read from the file.
type NodeSubsection struct {
	Heading    string
	RawHeading string
	Content    []string
}

// NodeSection represents a level-1 (#) heading block within a node file.
// Heading is the normalized form used for comparisons and lookups.
// RawHeading is the original line as read from the file, preserved for hashing.
// Content holds each line of the section body before the first level-2 heading.
// Subsections holds the ordered list of level-2 blocks within this section.
type NodeSection struct {
	Heading     string
	RawHeading  string
	Content     []string
	Subsections []*NodeSubsection
}

// Node is the parsed representation of a framework node file.
// NameSection is always present (the first level-1 heading).
// Public and Agent are optional named sections.
// Private holds all other sections in file order.
type Node struct {
	NameSection *NodeSection
	Public      *NodeSection
	Agent       *NodeSection
	Private     []*NodeSection
}

// atxHeading holds the result of parsing an ATX heading line.
type atxHeading struct {
	level int
	text  string
	raw   string
}

// fenceInfo holds the result of detecting a fenced code block opening.
type fenceInfo struct {
	char   rune
	length int
}

// parseAtxHeading attempts to parse line as an ATX heading.
// Returns nil if the line is not a heading.
func parseAtxHeading(line string) *atxHeading {
	// Count leading '#' characters.
	level := 0
	for _, ch := range line {
		if ch == '#' {
			level++
		} else {
			break
		}
	}
	if level == 0 {
		return nil
	}

	// The character immediately after the '#' characters must be a space.
	if len(line) <= level {
		return nil
	}
	if line[level] != ' ' {
		return nil
	}

	// Extract text after the first space following the '#' characters.
	text := strings.TrimSpace(line[level+1:])

	// Strip optional closing '#' sequence preceded by at least one space.
	if idx := strings.LastIndex(text, " #"); idx >= 0 {
		suffix := strings.TrimLeft(text[idx+1:], "#")
		if suffix == "" {
			// Everything after the last space is all '#' characters.
			text = strings.TrimSpace(text[:idx])
		}
	}

	return &atxHeading{
		level: level,
		text:  text,
		raw:   line,
	}
}

// isFenceOpening checks if line opens a fenced code block.
// Returns nil if the line is not a fence opening.
func isFenceOpening(line string) *fenceInfo {
	if len(line) == 0 {
		return nil
	}

	// Check for backtick fence.
	if line[0] == '`' {
		count := 0
		for _, ch := range line {
			if ch == '`' {
				count++
			} else {
				break
			}
		}
		if count >= 3 {
			// No backtick may appear in the rest of the line.
			rest := line[count:]
			if !strings.ContainsRune(rest, '`') {
				return &fenceInfo{char: '`', length: count}
			}
		}
	}

	// Check for tilde fence.
	if line[0] == '~' {
		count := 0
		for _, ch := range line {
			if ch == '~' {
				count++
			} else {
				break
			}
		}
		if count >= 3 {
			return &fenceInfo{char: '~', length: count}
		}
	}

	return nil
}

// isFenceClosing checks if line closes the current fence.
func isFenceClosing(line string, fenceChar rune, fenceLength int) bool {
	trimmed := strings.TrimRight(line, " \t")
	if len(trimmed) == 0 {
		return false
	}
	// Every character in the trimmed line must equal fenceChar.
	count := 0
	for _, ch := range trimmed {
		if ch != fenceChar {
			return false
		}
		count++
	}
	return count >= fenceLength
}

// runeCount returns the number of runes in s.
func runeCount(s string) int {
	return utf8.RuneCountInString(s)
}

// NodeParse parses the node file identified by logicalName and returns
// a structured Node.
func NodeParse(logicalName string) (*Node, error) {
	// Step 1: Check for ARTIFACT reference (not a ROOT reference).
	if logicalnames.LogicalNameIsArtifact(logicalName) {
		return nil, fmt.Errorf("%w", ErrNotARootReference)
	}

	// Also check that it actually starts with ROOT/.
	if !strings.HasPrefix(logicalName, "ROOT/") && logicalName != "ROOT" {
		return nil, fmt.Errorf("%w", ErrNotARootReference)
	}

	// Step 2: Check for qualifier.
	if logicalnames.LogicalNameHasQualifier(logicalName) {
		return nil, fmt.Errorf("%w", ErrHasQualifier)
	}

	// Step 3: Resolve logical name to file path.
	filePath, err := logicalnames.LogicalNameToPath(logicalName)
	if err != nil {
		return nil, fmt.Errorf("resolving logical name: %w", err)
	}

	// Step 4: Open the file.
	reader, err := filereader.FileOpen(filePath)
	if err != nil {
		if errors.Is(err, filereader.ErrFileUnreadable) {
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		return nil, fmt.Errorf("opening node file: %w", err)
	}

	// Step 5: Skip frontmatter.
	// Read the first line to check for "---".
	var pendingLine *string
	firstLine, err := filereader.FileReadLine(reader)
	if err != nil {
		if errors.Is(err, filereader.ErrEndOfFile) {
			// Empty file — no lines. Proceed to step 6 with no lines.
		} else {
			filereader.FileClose(reader)
			return nil, fmt.Errorf("reading first line: %w", err)
		}
	} else {
		if firstLine == "---" {
			// Skip frontmatter until closing "---".
			foundClose := false
			for {
				line, err := filereader.FileReadLine(reader)
				if err != nil {
					if errors.Is(err, filereader.ErrEndOfFile) {
						filereader.FileClose(reader)
						return nil, fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
					}
					filereader.FileClose(reader)
					return nil, fmt.Errorf("reading frontmatter: %w", err)
				}
				if line == "---" {
					foundClose = true
					break
				}
			}
			_ = foundClose
			// Frontmatter skipped; pendingLine remains nil.
		} else {
			// Not frontmatter — treat as first body line.
			pendingLine = &firstLine
		}
	}

	// Step 6: Parse the body into sections.
	var (
		nameSection      *NodeSection
		publicSection    *NodeSection
		agentSection     *NodeSection
		privateSections  []*NodeSection
		currentSection   *NodeSection
		currentSubsection *NodeSubsection
		inCodeFence      bool
		fenceChar        rune
		fenceLength      int
	)

	// classifyAndStoreSection assigns the section to the appropriate slot.
	classifyAndStoreSection := func(section *NodeSection) error {
		if nameSection == nil {
			nameSection = section
			return nil
		}
		if section.Heading == "public" {
			if publicSection != nil {
				filereader.FileClose(reader)
				return fmt.Errorf("%w", ErrDuplicatePublicSection)
			}
			publicSection = section
			return nil
		}
		if section.Heading == "agent" {
			if agentSection != nil {
				filereader.FileClose(reader)
				return fmt.Errorf("%w", ErrDuplicateAgentSection)
			}
			agentSection = section
			return nil
		}
		privateSections = append(privateSections, section)
		return nil
	}

	// finalizeSubsection moves currentSubsection into currentSection.
	finalizeSubsection := func() {
		if currentSubsection == nil {
			return
		}
		currentSection.Subsections = append(currentSection.Subsections, currentSubsection)
		currentSubsection = nil
	}

	for {
		// Step 6a: Obtain the next line.
		var line string
		if pendingLine != nil {
			line = *pendingLine
			pendingLine = nil
		} else {
			var err error
			line, err = filereader.FileReadLine(reader)
			if err != nil {
				if errors.Is(err, filereader.ErrEndOfFile) {
					break
				}
				filereader.FileClose(reader)
				return nil, fmt.Errorf("reading body: %w", err)
			}
		}

		// Step 6b: Check for code fence transitions.
		if !inCodeFence {
			if fence := isFenceOpening(line); fence != nil {
				inCodeFence = true
				fenceChar = fence.char
				fenceLength = fence.length
				// Treat line as content (step f).
				goto addContent
			}
			// Not a fence opening; continue to step c.
		} else {
			// We are inside a code fence.
			if isFenceClosing(line, fenceChar, fenceLength) {
				inCodeFence = false
				fenceChar = 0
				fenceLength = 0
			}
			// Treat line as content (step f) regardless.
			goto addContent
		}

		// Step 6c: Attempt to parse as ATX heading (only when not in code fence).
		{
			heading := parseAtxHeading(line)
			if heading == nil {
				// Not a heading; treat as content.
				goto addContent
			}

			if heading.level == 1 {
				// Step 6d: Level-1 heading.
				finalizeSubsection()
				if currentSection != nil {
					if err := classifyAndStoreSection(currentSection); err != nil {
						return nil, err
					}
				}
				normalizedHeading := textnormalization.NormalizeText(heading.text)
				currentSection = &NodeSection{
					Heading:     normalizedHeading,
					RawHeading:  heading.raw,
					Content:     []string{},
					Subsections: []*NodeSubsection{},
				}
				currentSubsection = nil
				continue
			}

			if heading.level == 2 && currentSection != nil {
				// Step 6e: Level-2 heading.
				finalizeSubsection()
				normalizedHeading := textnormalization.NormalizeText(heading.text)
				// Check for duplicate subsection.
				for _, existing := range currentSection.Subsections {
					if existing.Heading == normalizedHeading {
						filereader.FileClose(reader)
						return nil, fmt.Errorf("%w", ErrDuplicateSubsection)
					}
				}
				currentSubsection = &NodeSubsection{
					Heading:    normalizedHeading,
					RawHeading: heading.raw,
					Content:    []string{},
				}
				continue
			}

			// Level-2+ heading with no current section, or level >= 3: treat as content.
		}

	addContent:
		// Step 6f: Add line to content.
		if currentSubsection != nil {
			currentSubsection.Content = append(currentSubsection.Content, line)
		} else if currentSection != nil {
			currentSection.Content = append(currentSection.Content, line)
		} else {
			// Before first heading.
			if strings.TrimSpace(line) != "" {
				filereader.FileClose(reader)
				return nil, fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
			}
			// Blank lines before first heading are silently discarded.
		}
	}

	// After loop: finalize remaining section/subsection.
	finalizeSubsection()
	if currentSection != nil {
		if err := classifyAndStoreSection(currentSection); err != nil {
			return nil, err
		}
	}

	// Step 7: Validate name section.
	if nameSection == nil {
		filereader.FileClose(reader)
		return nil, fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
	}

	normalizedLogicalName := textnormalization.NormalizeText(logicalName)
	if nameSection.Heading != normalizedLogicalName {
		filereader.FileClose(reader)
		return nil, fmt.Errorf("%w", ErrNodeNameDoesNotMatch)
	}

	// Step 8: Close the file.
	filereader.FileClose(reader)

	// Step 9: Return the Node.
	return &Node{
		NameSection: nameSection,
		Public:      publicSection,
		Agent:       agentSection,
		Private:     privateSections,
	}, nil
}

// ensure runeCount is used to avoid unused import warning.
var _ = runeCount
