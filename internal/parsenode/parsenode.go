// code-from-spec: ROOT/golang/implementation/parsing/node_parsing@iWjS28F45U2LILWg-1Wdxrr_oIs
package parsenode

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
)

// ErrNotARootReference is returned when the logical name does not start with ROOT/.
var ErrNotARootReference = errors.New("logical name does not start with ROOT/")

// ErrHasQualifier is returned when the logical name contains a parenthetical qualifier.
var ErrHasQualifier = errors.New("logical name contains a parenthetical qualifier")

// ErrFileUnreadable is returned when the file cannot be opened or read.
var ErrFileUnreadable = errors.New("file cannot be opened or read")

// ErrUnexpectedContentBeforeFirstHeading is returned when the file body has
// non-blank content before the first level-1 heading, or has no level-1 heading at all.
// Blank lines before the first heading are not an error.
var ErrUnexpectedContentBeforeFirstHeading = errors.New("unexpected content before first heading")

// ErrNodeNameDoesNotMatch is returned when the first heading does not match
// the logical name after normalization.
var ErrNodeNameDoesNotMatch = errors.New("node name does not match first heading")

// ErrDuplicatePublicSection is returned when more than one Public section exists.
var ErrDuplicatePublicSection = errors.New("duplicate Public section")

// ErrDuplicateAgentSection is returned when more than one Agent section exists.
var ErrDuplicateAgentSection = errors.New("duplicate Agent section")

// ErrDuplicateSubsection is returned when two level-2 headings within the same
// section normalize to the same text.
var ErrDuplicateSubsection = errors.New("duplicate subsection")

// NodeSubsection represents a level-2 heading section (##) within a node file.
type NodeSubsection struct {
	Heading    string
	RawHeading string
	Content    []string
}

// NodeSection represents a level-1 heading section (#) within a node file.
type NodeSection struct {
	Heading     string
	RawHeading  string
	Content     []string
	Subsections []*NodeSubsection
}

// Node represents the parsed structure of a node file.
type Node struct {
	NameSection *NodeSection
	Public      *NodeSection
	Agent       *NodeSection
	Private     []*NodeSection
}

// NodeParse parses the node file identified by logicalName and returns a Node.
func NodeParse(logicalName string) (*Node, error) {
	// Step 1: reject ARTIFACT/ references.
	if logicalnames.LogicalNameIsArtifact(logicalName) {
		return nil, fmt.Errorf("%w", ErrNotARootReference)
	}

	// Step 2: reject names with qualifiers.
	if logicalnames.LogicalNameHasQualifier(logicalName) {
		return nil, fmt.Errorf("%w", ErrHasQualifier)
	}

	// Step 3: resolve the logical name to a file path.
	filePath, err := logicalnames.LogicalNameToPath(logicalName)
	if err != nil {
		return nil, fmt.Errorf("LogicalNameToPath: %w", err)
	}

	// Step 4: open the file.
	reader, err := filereader.FileOpen(filePath)
	if err != nil {
		if errors.Is(err, filereader.ErrFileUnreadable) {
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		return nil, fmt.Errorf("FileOpen: %w", err)
	}

	// Step 5: skip frontmatter. All error paths must call FileClose.
	var heldLine *string

	firstLine, err := filereader.FileReadLine(reader)
	if err != nil {
		filereader.FileClose(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			return nil, fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
		}
		return nil, fmt.Errorf("FileReadLine: %w", err)
	}

	if firstLine == "---" {
		// Consume lines until closing "---".
		for {
			line, err := filereader.FileReadLine(reader)
			if err != nil {
				filereader.FileClose(reader)
				if errors.Is(err, filereader.ErrEndOfFile) {
					return nil, fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
				}
				return nil, fmt.Errorf("FileReadLine: %w", err)
			}
			if line == "---" {
				break
			}
		}
	} else {
		// Hold this line for the body parsing loop.
		heldLine = &firstLine
	}

	// Step 6: parse the body into sections.
	type sectionsSeenState struct {
		nameSeen   bool
		publicSeen bool
		agentSeen  bool
	}

	var seen sectionsSeenState
	var currentSection *NodeSection
	var currentSubsection *NodeSubsection
	var privateSections []*NodeSection

	// Stored final results.
	var nameSection *NodeSection
	var publicSection *NodeSection
	var agentSection *NodeSection

	// Fence state.
	inFence := false
	fenceChar := ""
	fenceLength := 0

	// storeSection stores a finalized section into the appropriate slot.
	expectedNorm := textnormalization.NormalizeText(logicalName)
	storeSection := func(sec *NodeSection) {
		if sec.Heading == expectedNorm {
			nameSection = sec
		} else if sec.Heading == "public" {
			publicSection = sec
		} else if sec.Heading == "agent" {
			agentSection = sec
		} else {
			privateSections = append(privateSections, sec)
		}
	}

	// finalizeCurrentSubsection appends currentSubsection to currentSection if set.
	finalizeCurrentSubsection := func() {
		if currentSubsection != nil {
			currentSection.Subsections = append(currentSection.Subsections, currentSubsection)
			currentSubsection = nil
		}
	}

	// finalizeCurrentSection stores currentSection if set.
	finalizeCurrentSection := func() {
		if currentSection != nil {
			storeSection(currentSection)
			currentSection = nil
		}
	}

	// processLine handles a single body line.
	processLine := func(line string) error {
		// Step 6a: fence tracking.
		if !inFence {
			backtickCount := countLeadingChar(line, '`')
			if backtickCount >= 3 && !strings.ContainsRune(line[backtickCount:], '`') {
				inFence = true
				fenceChar = "`"
				fenceLength = backtickCount
				// Treat as content — fall through to step 6f.
			} else {
				tildeCount := countLeadingChar(line, '~')
				if tildeCount >= 3 && !strings.ContainsRune(line[tildeCount:], '~') {
					inFence = true
					fenceChar = "~"
					fenceLength = tildeCount
					// Treat as content — fall through to step 6f.
				}
			}
		} else {
			// Already in fence — check for closing delimiter.
			fc := rune(fenceChar[0])
			count := countLeadingChar(line, fc)
			if count >= fenceLength && strings.TrimSpace(line[count:]) == "" {
				inFence = false
			}
			// Treat as content regardless.
			appendContent(currentSection, currentSubsection, line)
			return nil
		}

		// Step 6b: if now in fence, treat as content.
		if inFence {
			appendContent(currentSection, currentSubsection, line)
			return nil
		}

		// Step 6c: attempt to parse as ATX heading.
		level, headingText, isHeading := parseATXHeading(line)

		if !isHeading {
			// Step 6f: content line.
			return handleContentLine(currentSection, currentSubsection, line)
		}

		if level == 1 {
			// Step 6d: new level-1 section.
			finalizeCurrentSubsection()
			finalizeCurrentSection()

			norm := textnormalization.NormalizeText(headingText)

			if !seen.nameSeen {
				if norm != expectedNorm {
					return fmt.Errorf("%w", ErrNodeNameDoesNotMatch)
				}
				currentSection = &NodeSection{
					Heading:     norm,
					RawHeading:  line,
					Content:     []string{},
					Subsections: []*NodeSubsection{},
				}
				seen.nameSeen = true
			} else if norm == "public" {
				if seen.publicSeen {
					return fmt.Errorf("%w", ErrDuplicatePublicSection)
				}
				currentSection = &NodeSection{
					Heading:     norm,
					RawHeading:  line,
					Content:     []string{},
					Subsections: []*NodeSubsection{},
				}
				seen.publicSeen = true
			} else if norm == "agent" {
				if seen.agentSeen {
					return fmt.Errorf("%w", ErrDuplicateAgentSection)
				}
				currentSection = &NodeSection{
					Heading:     norm,
					RawHeading:  line,
					Content:     []string{},
					Subsections: []*NodeSubsection{},
				}
				seen.agentSeen = true
			} else {
				currentSection = &NodeSection{
					Heading:     norm,
					RawHeading:  line,
					Content:     []string{},
					Subsections: []*NodeSubsection{},
				}
			}
			currentSubsection = nil
			return nil
		}

		if level == 2 {
			// Step 6e: new level-2 subsection.
			if currentSection == nil {
				// A ## heading before any # heading is non-blank content.
				return fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
			}

			subNorm := textnormalization.NormalizeText(headingText)

			// Check for duplicate in already-stored subsections.
			for _, existing := range currentSection.Subsections {
				if existing.Heading == subNorm {
					return fmt.Errorf("%w", ErrDuplicateSubsection)
				}
			}
			// Check against in-progress subsection.
			if currentSubsection != nil && currentSubsection.Heading == subNorm {
				return fmt.Errorf("%w", ErrDuplicateSubsection)
			}

			finalizeCurrentSubsection()

			currentSubsection = &NodeSubsection{
				Heading:    subNorm,
				RawHeading: line,
				Content:    []string{},
			}
			return nil
		}

		// level >= 3 or level == 0 falls through to content.
		return handleContentLine(currentSection, currentSubsection, line)
	}

	// Process held line first (if any), then read remaining lines.
	if heldLine != nil {
		if err := processLine(*heldLine); err != nil {
			filereader.FileClose(reader)
			return nil, err
		}
	}

	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				break
			}
			filereader.FileClose(reader)
			return nil, fmt.Errorf("FileReadLine: %w", err)
		}
		if err := processLine(line); err != nil {
			filereader.FileClose(reader)
			return nil, err
		}
	}

	// Step 7: finalize remaining state.
	finalizeCurrentSubsection()
	finalizeCurrentSection()

	if !seen.nameSeen {
		filereader.FileClose(reader)
		return nil, fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
	}

	// Step 8: close the reader.
	filereader.FileClose(reader)

	// Step 9: return the Node.
	return &Node{
		NameSection: nameSection,
		Public:      publicSection,
		Agent:       agentSection,
		Private:     privateSections,
	}, nil
}

// countLeadingChar counts how many consecutive occurrences of ch appear at the
// start of s.
func countLeadingChar(s string, ch rune) int {
	count := 0
	for _, r := range s {
		if r != ch {
			break
		}
		count++
	}
	return count
}

// parseATXHeading attempts to parse line as an ATX heading.
// Returns (level, headingText, true) on success, or (0, "", false) if not a heading.
func parseATXHeading(line string) (int, string, bool) {
	level := countLeadingChar(line, '#')
	if level == 0 {
		return 0, "", false
	}
	// Character after the hashes must be a space.
	if len(line) <= level || line[level] != ' ' {
		return 0, "", false
	}

	headingText := strings.TrimSpace(line[level+1:])

	// Strip optional closing "#" sequence: ends with one or more "#"
	// preceded by at least one space.
	if idx := strings.LastIndex(headingText, " "); idx >= 0 {
		tail := headingText[idx+1:]
		allHash := len(tail) > 0
		for _, r := range tail {
			if r != '#' {
				allHash = false
				break
			}
		}
		if allHash {
			headingText = strings.TrimSpace(headingText[:idx])
		}
	}

	return level, headingText, true
}

// appendContent appends line to the appropriate content slice.
func appendContent(currentSection *NodeSection, currentSubsection *NodeSubsection, line string) {
	if currentSubsection != nil {
		currentSubsection.Content = append(currentSubsection.Content, line)
	} else if currentSection != nil {
		currentSection.Content = append(currentSection.Content, line)
	}
}

// handleContentLine handles a non-heading line (step 6f).
func handleContentLine(currentSection *NodeSection, currentSubsection *NodeSubsection, line string) error {
	if currentSection == nil {
		if strings.TrimSpace(line) == "" {
			// Blank line before first heading — discard.
			return nil
		}
		return fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
	}
	appendContent(currentSection, currentSubsection, line)
	return nil
}
