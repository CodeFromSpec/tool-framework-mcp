// code-from-spec: SPEC/golang/implementation/parsing/node_parsing@0HEKeWJZ7-tcpQJ8OzekiESTHno
package parsenode

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/file"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/textnormalization"
)

type NodeSubsection struct {
	Heading    string
	RawHeading string
	Content    []string
}

type NodeSection struct {
	Heading     string
	RawHeading  string
	Content     []string
	Subsections []*NodeSubsection
}

type Node struct {
	NameSection *NodeSection
	Public      *NodeSection
	Agent       *NodeSection
	Private     *NodeSection
}

var ErrNotASpecReference                   = errors.New("logical name is not a SPEC/ reference")
var ErrHasQualifier                        = errors.New("logical name contains a parenthetical qualifier")
var ErrFileUnreadable                      = errors.New("file cannot be opened or read")
var ErrUnexpectedContentBeforeFirstHeading = errors.New("file body has non-blank content before the first level-1 heading, or has no level-1 heading at all")
var ErrNodeNameDoesNotMatch                = errors.New("first heading does not match the logical name after normalization")
var ErrDuplicatePublicSection              = errors.New("more than one Public section exists")
var ErrDuplicateAgentSection               = errors.New("more than one Agent section exists")
var ErrDuplicatePrivateSection             = errors.New("more than one Private section exists")
var ErrUnrecognizedSection                 = errors.New("unrecognized level-1 heading")
var ErrDuplicateSubsection                 = errors.New("two level-2 headings within the same section normalize to the same text")

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

	handle, err := file.FileOpen(cfsPath, "read", 30000)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	var firstBodyLine *string
	firstLine, err := file.FileReadLine(handle)
	if err != nil {
		if !errors.Is(err, file.ErrEndOfFile) {
			file.FileClose(handle)
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
	} else {
		if firstLine == "---" {
			for {
				line, err := file.FileReadLine(handle)
				if err != nil {
					if errors.Is(err, file.ErrEndOfFile) {
						file.FileClose(handle)
						return nil, fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
					}
					file.FileClose(handle)
					return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
				}
				if line == "---" {
					break
				}
			}
		} else {
			firstBodyLine = &firstLine
		}
	}

	var nameSection *NodeSection
	var publicSection *NodeSection
	var agentSection *NodeSection
	var privateSection *NodeSection
	var currentSection *NodeSection
	var currentSubsection *NodeSubsection
	inFence := false
	fenceChar := byte(0)
	fenceWidth := 0

	appendToContent := func(line string) {
		if currentSubsection != nil {
			currentSubsection.Content = append(currentSubsection.Content, line)
		} else if currentSection != nil {
			currentSection.Content = append(currentSection.Content, line)
		}
	}

	appendOrError := func(line string) error {
		if currentSubsection != nil || currentSection != nil {
			appendToContent(line)
			return nil
		}
		if strings.TrimSpace(line) != "" {
			return fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
		}
		return nil
	}

	finalizeSubsection := func() {
		if currentSubsection != nil {
			currentSection.Subsections = append(currentSection.Subsections, currentSubsection)
			currentSubsection = nil
		}
	}

	processLine := func(line string) error {
		stripped := strings.TrimSpace(line)

		if !inFence {
			if strings.HasPrefix(stripped, "```") || strings.HasPrefix(stripped, "~~~") {
				ch := stripped[0]
				count := 0
				for count < len(stripped) && stripped[count] == ch {
					count++
				}
				fenceChar = ch
				fenceWidth = count
				inFence = true
				appendToContent(line)
				return nil
			}
		} else {
			ch := fenceChar
			count := 0
			for count < len(stripped) && stripped[count] == ch {
				count++
			}
			if count >= fenceWidth && count == len(stripped) {
				inFence = false
				fenceChar = 0
				fenceWidth = 0
			}
			appendToContent(line)
			return nil
		}

		level, textPart, isHeading := parseATXHeading(line)
		if !isHeading {
			return appendOrError(line)
		}

		heading := textnormalization.NormalizeText(textPart)

		if level == 1 {
			finalizeSubsection()
			currentSection = nil

			if nameSection == nil {
				expected := textnormalization.NormalizeText(logicalName)
				if heading != expected {
					return fmt.Errorf("%w", ErrNodeNameDoesNotMatch)
				}
				nameSection = &NodeSection{
					Heading:     heading,
					RawHeading:  line,
					Content:     []string{},
					Subsections: []*NodeSubsection{},
				}
				currentSection = nameSection
			} else if heading == "public" {
				if publicSection != nil {
					return fmt.Errorf("%w", ErrDuplicatePublicSection)
				}
				publicSection = &NodeSection{
					Heading:     heading,
					RawHeading:  line,
					Content:     []string{},
					Subsections: []*NodeSubsection{},
				}
				currentSection = publicSection
			} else if heading == "agent" {
				if agentSection != nil {
					return fmt.Errorf("%w", ErrDuplicateAgentSection)
				}
				agentSection = &NodeSection{
					Heading:     heading,
					RawHeading:  line,
					Content:     []string{},
					Subsections: []*NodeSubsection{},
				}
				currentSection = agentSection
			} else if heading == "private" {
				if privateSection != nil {
					return fmt.Errorf("%w", ErrDuplicatePrivateSection)
				}
				privateSection = &NodeSection{
					Heading:     heading,
					RawHeading:  line,
					Content:     []string{},
					Subsections: []*NodeSubsection{},
				}
				currentSection = privateSection
			} else {
				return fmt.Errorf("%w", ErrUnrecognizedSection)
			}
		} else if level == 2 {
			if currentSection == nil {
				return appendOrError(line)
			}
			finalizeSubsection()
			for _, sub := range currentSection.Subsections {
				if sub.Heading == heading {
					return fmt.Errorf("%w", ErrDuplicateSubsection)
				}
			}
			currentSubsection = &NodeSubsection{
				Heading:    heading,
				RawHeading: line,
				Content:    []string{},
			}
		} else {
			appendToContent(line)
		}

		return nil
	}

	if firstBodyLine != nil {
		if err := processLine(*firstBodyLine); err != nil {
			file.FileClose(handle)
			return nil, err
		}
	}

	for {
		line, err := file.FileReadLine(handle)
		if err != nil {
			if errors.Is(err, file.ErrEndOfFile) {
				finalizeSubsection()
				break
			}
			file.FileClose(handle)
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		if err := processLine(line); err != nil {
			file.FileClose(handle)
			return nil, err
		}
	}

	if nameSection == nil {
		file.FileClose(handle)
		return nil, fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
	}

	file.FileClose(handle)

	return &Node{
		NameSection: nameSection,
		Public:      publicSection,
		Agent:       agentSection,
		Private:     privateSection,
	}, nil
}

func parseATXHeading(line string) (level int, text string, ok bool) {
	if len(line) == 0 || line[0] != '#' {
		return 0, "", false
	}
	count := 0
	for count < len(line) && line[count] == '#' {
		count++
	}
	if count >= len(line) || line[count] != ' ' {
		return 0, "", false
	}
	textPart := line[count+1:]
	textPart = strings.TrimSpace(textPart)
	for len(textPart) > 0 && textPart[len(textPart)-1] == '#' {
		trimmed := strings.TrimRight(textPart, "#")
		if len(trimmed) > 0 && trimmed[len(trimmed)-1] == ' ' {
			textPart = strings.TrimRight(trimmed, " ")
			break
		}
		break
	}
	return count, textPart, true
}
