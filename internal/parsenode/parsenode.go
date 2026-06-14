// code-from-spec: ROOT/golang/implementation/parsing/node_parsing@JHyQOIHlTAT4juTYwmEWB7i6eOs
package parsenode

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
)

// NodeSubsection represents a level-2 heading and its content within a section.
type NodeSubsection struct {
	Heading    string
	RawHeading string
	Content    []string
}

// NodeSection represents a level-1 heading and its content within a node file.
type NodeSection struct {
	Heading     string
	RawHeading  string
	Content     []string
	Subsections []*NodeSubsection
}

// Node represents a parsed spec node file.
type Node struct {
	NameSection *NodeSection
	Public      *NodeSection
	Agent       *NodeSection
	Private     *NodeSection
}

var ErrNotASpecReference                   = errors.New("logical name is not a SPEC/ reference")
var ErrHasQualifier                        = errors.New("logical name contains a parenthetical qualifier")
var ErrFileUnreadable                      = errors.New("file cannot be opened or read")
var ErrUnexpectedContentBeforeFirstHeading = errors.New("file has non-blank content before the first level-1 heading, or has no level-1 heading")
var ErrNodeNameDoesNotMatch                = errors.New("first heading does not match the logical name after normalization")
var ErrDuplicatePublicSection              = errors.New("more than one Public section exists")
var ErrDuplicateAgentSection               = errors.New("more than one Agent section exists")
var ErrDuplicatePrivateSection             = errors.New("more than one Private section exists")
var ErrUnrecognizedSection                 = errors.New("unrecognized level-1 heading")
var ErrDuplicateSubsection                 = errors.New("two level-2 headings within the same section normalize to the same text")

// NodeParse reads and parses the spec node file identified by logical_name.
// The logical name must be a SPEC/ reference and must not contain a parenthetical qualifier.
// Returns the parsed Node or an error describing the failure.
func NodeParse(logicalName string) (*Node, error) {
	if !logicalnames.LogicalNameIsSpec(logicalName) {
		return nil, fmt.Errorf("%w", ErrNotASpecReference)
	}

	if logicalnames.LogicalNameHasQualifier(logicalName) {
		return nil, fmt.Errorf("%w", ErrHasQualifier)
	}

	cfsPath, err := logicalnames.LogicalNameToPath(logicalName)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	firstLine, firstLineErr := filereader.FileReadLine(reader)

	var leftoverLine string
	hasLeftover := false

	if firstLineErr == nil {
		if firstLine == "---" {
			for {
				line, err := filereader.FileReadLine(reader)
				if errors.Is(err, filereader.ErrEndOfFile) {
					filereader.FileClose(reader)
					return nil, fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
				}
				if err != nil {
					filereader.FileClose(reader)
					return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
				}
				if line == "---" {
					break
				}
			}
		} else {
			leftoverLine = firstLine
			hasLeftover = true
		}
	} else if !errors.Is(firstLineErr, filereader.ErrEndOfFile) {
		filereader.FileClose(reader)
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, firstLineErr)
	}

	sectionsSeenPublic := false
	sectionsSeenAgent := false
	sectionsSeenPrivate := false

	var currentSection *NodeSection
	var currentSubsection *NodeSubsection
	var resultNameSection *NodeSection
	var resultPublic *NodeSection
	var resultAgent *NodeSection
	var resultPrivate *NodeSection

	insideFence := false
	fenceChar := byte(0)
	fenceLength := 0

	normalizedLogicalName := textnormalization.NormalizeText(logicalName)

	finalizeSubsection := func() {
		if currentSubsection != nil && currentSection != nil {
			currentSection.Subsections = append(currentSection.Subsections, currentSubsection)
			currentSubsection = nil
		}
	}

	appendToContent := func(line string) {
		if currentSubsection != nil {
			currentSubsection.Content = append(currentSubsection.Content, line)
		} else if currentSection != nil {
			currentSection.Content = append(currentSection.Content, line)
		}
	}

	processLine := func(line string) error {
		if insideFence {
			if isFenceClose(line, fenceChar, fenceLength) {
				insideFence = false
				fenceChar = 0
				fenceLength = 0
			}
			appendToContent(line)
			return nil
		}

		if ch, length, ok := detectFenceOpen(line); ok {
			insideFence = true
			fenceChar = ch
			fenceLength = length
			appendToContent(line)
			return nil
		}

		headingLevel, headingText, isHeading := parseATXHeading(line)
		if !isHeading {
			if currentSubsection != nil {
				currentSubsection.Content = append(currentSubsection.Content, line)
			} else if currentSection != nil {
				currentSection.Content = append(currentSection.Content, line)
			} else {
				if strings.TrimSpace(line) != "" {
					filereader.FileClose(reader)
					return fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
				}
			}
			return nil
		}

		normalizedHeading := textnormalization.NormalizeText(headingText)

		if headingLevel == 1 {
			finalizeSubsection()

			if resultNameSection == nil {
				if normalizedHeading != normalizedLogicalName {
					filereader.FileClose(reader)
					return fmt.Errorf("%w", ErrNodeNameDoesNotMatch)
				}
				newSection := &NodeSection{
					Heading:     normalizedHeading,
					RawHeading:  line,
					Content:     []string{},
					Subsections: []*NodeSubsection{},
				}
				currentSection = newSection
				resultNameSection = newSection
			} else if normalizedHeading == "public" {
				if sectionsSeenPublic {
					filereader.FileClose(reader)
					return fmt.Errorf("%w", ErrDuplicatePublicSection)
				}
				sectionsSeenPublic = true
				newSection := &NodeSection{
					Heading:     normalizedHeading,
					RawHeading:  line,
					Content:     []string{},
					Subsections: []*NodeSubsection{},
				}
				currentSection = newSection
				resultPublic = newSection
			} else if normalizedHeading == "agent" {
				if sectionsSeenAgent {
					filereader.FileClose(reader)
					return fmt.Errorf("%w", ErrDuplicateAgentSection)
				}
				sectionsSeenAgent = true
				newSection := &NodeSection{
					Heading:     normalizedHeading,
					RawHeading:  line,
					Content:     []string{},
					Subsections: []*NodeSubsection{},
				}
				currentSection = newSection
				resultAgent = newSection
			} else if normalizedHeading == "private" {
				if sectionsSeenPrivate {
					filereader.FileClose(reader)
					return fmt.Errorf("%w", ErrDuplicatePrivateSection)
				}
				sectionsSeenPrivate = true
				newSection := &NodeSection{
					Heading:     normalizedHeading,
					RawHeading:  line,
					Content:     []string{},
					Subsections: []*NodeSubsection{},
				}
				currentSection = newSection
				resultPrivate = newSection
			} else {
				filereader.FileClose(reader)
				return fmt.Errorf("%w", ErrUnrecognizedSection)
			}
		} else if headingLevel == 2 {
			if currentSection == nil {
				if strings.TrimSpace(normalizedHeading) == "" {
					return nil
				}
				filereader.FileClose(reader)
				return fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
			}
			finalizeSubsection()
			for _, existing := range currentSection.Subsections {
				if existing.Heading == normalizedHeading {
					filereader.FileClose(reader)
					return fmt.Errorf("%w", ErrDuplicateSubsection)
				}
			}
			newSubsection := &NodeSubsection{
				Heading:    normalizedHeading,
				RawHeading: line,
				Content:    []string{},
			}
			currentSubsection = newSubsection
		} else {
			if currentSubsection != nil {
				currentSubsection.Content = append(currentSubsection.Content, line)
			} else if currentSection != nil {
				currentSection.Content = append(currentSection.Content, line)
			} else {
				if strings.TrimSpace(line) != "" {
					filereader.FileClose(reader)
					return fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
				}
			}
		}

		return nil
	}

	if hasLeftover {
		if err := processLine(leftoverLine); err != nil {
			return nil, err
		}
	}

	for {
		line, err := filereader.FileReadLine(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			break
		}
		if err != nil {
			filereader.FileClose(reader)
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		if err := processLine(line); err != nil {
			return nil, err
		}
	}

	finalizeSubsection()

	if resultNameSection == nil {
		filereader.FileClose(reader)
		return nil, fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
	}

	filereader.FileClose(reader)

	return &Node{
		NameSection: resultNameSection,
		Public:      resultPublic,
		Agent:       resultAgent,
		Private:     resultPrivate,
	}, nil
}

func parseATXHeading(line string) (level int, text string, ok bool) {
	if len(line) == 0 || line[0] != '#' {
		return 0, "", false
	}

	i := 0
	for i < len(line) && line[i] == '#' {
		i++
	}

	if i >= len(line) || line[i] != ' ' {
		return 0, "", false
	}

	level = i
	text = line[i+1:]
	text = strings.TrimSpace(text)

	trimmed := strings.TrimRight(text, " ")
	if len(trimmed) > 0 && trimmed[len(trimmed)-1] == '#' {
		idx := strings.LastIndex(trimmed, " #")
		if idx >= 0 {
			candidate := strings.TrimRight(trimmed[:idx], " ")
			text = candidate
		}
	}

	return level, text, true
}

func detectFenceOpen(line string) (ch byte, length int, ok bool) {
	if len(line) == 0 {
		return 0, 0, false
	}

	c := line[0]
	if c != '`' && c != '~' {
		return 0, 0, false
	}

	count := 0
	for count < len(line) && line[count] == c {
		count++
	}

	if count < 3 {
		return 0, 0, false
	}

	return c, count, true
}

func isFenceClose(line string, fenceChar byte, fenceLength int) bool {
	count := 0
	for count < len(line) && line[count] == fenceChar {
		count++
	}

	if count < fenceLength {
		return false
	}

	rest := strings.TrimSpace(line[count:])
	return rest == ""
}
