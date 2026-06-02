// code-from-spec: ROOT/golang/implementation/parsing/node_parsing@7V1G3qnvq9UEARrySqmLXP_5XdQ
package parsenode

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
)

var ErrNotARootReference = errors.New("logical name does not start with ROOT/")
var ErrHasQualifier = errors.New("logical name contains a parenthetical qualifier")
var ErrFileUnreadable = errors.New("file cannot be opened or read")
var ErrUnexpectedContentBeforeFirstHeading = errors.New("unexpected content before first heading")
var ErrNodeNameDoesNotMatch = errors.New("first heading does not match logical name")
var ErrDuplicatePublicSection = errors.New("more than one Public section")
var ErrDuplicateAgentSection = errors.New("more than one Agent section")
var ErrDuplicateSubsection = errors.New("duplicate subsection heading within section")

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
	Private     []*NodeSection
}

type fenceState struct {
	open      bool
	fenceChar byte
	fenceLen  int
}

func parseHeading(line string) (level int, text string, isHeading bool) {
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
	text = strings.TrimSpace(line[i+1:])

	for strings.HasSuffix(text, "#") {
		trimmed := strings.TrimRight(text, "#")
		trimmed = strings.TrimRight(trimmed, " ")
		if len(trimmed) < len(text) {
			text = trimmed
		} else {
			break
		}
	}

	return level, text, true
}

func detectFence(line string) (isFence bool, fenceChar byte, fenceLen int) {
	if len(line) == 0 {
		return false, 0, 0
	}

	ch := line[0]
	if ch != '`' && ch != '~' {
		return false, 0, 0
	}

	count := 0
	for count < len(line) && line[count] == ch {
		count++
	}

	if count < 3 {
		return false, 0, 0
	}

	rest := line[count:]
	if strings.ContainsAny(rest, "`~") && ch == '`' {
		return false, 0, 0
	}

	return true, ch, count
}

func closingFence(line string, fenceChar byte, fenceLen int) bool {
	count := 0
	for count < len(line) && line[count] == fenceChar {
		count++
	}
	if count < fenceLen {
		return false
	}
	rest := strings.TrimSpace(line[count:])
	return rest == ""
}

func NodeParse(logical_name string) (*Node, error) {
	if logicalnames.LogicalNameIsArtifact(logical_name) {
		return nil, ErrNotARootReference
	}

	if logicalnames.LogicalNameHasQualifier(logical_name) {
		return nil, ErrHasQualifier
	}

	filePath, err := logicalnames.LogicalNameToPath(logical_name)
	if err != nil {
		if errors.Is(err, logicalnames.ErrUnsupportedReference) {
			return nil, ErrNotARootReference
		}
		return nil, err
	}

	reader, err := filereader.FileOpen(filePath)
	if err != nil {
		if errors.Is(err, filereader.ErrFileUnreadable) {
			return nil, fmt.Errorf("%w: %s", ErrFileUnreadable, err)
		}
		return nil, err
	}

	firstLine, err := filereader.FileReadLine(reader)
	if err != nil {
		if errors.Is(err, filereader.ErrEndOfFile) {
			filereader.FileClose(reader)
			return nil, ErrUnexpectedContentBeforeFirstHeading
		}
		filereader.FileClose(reader)
		return nil, fmt.Errorf("%w: %s", ErrFileUnreadable, err)
	}

	var heldLine *string
	if firstLine == "---" {
		for {
			line, err := filereader.FileReadLine(reader)
			if err != nil {
				if errors.Is(err, filereader.ErrEndOfFile) {
					filereader.FileClose(reader)
					return nil, ErrUnexpectedContentBeforeFirstHeading
				}
				filereader.FileClose(reader)
				return nil, fmt.Errorf("%w: %s", ErrFileUnreadable, err)
			}
			if line == "---" {
				break
			}
		}
	} else {
		heldLine = &firstLine
	}

	node := &Node{
		Private: []*NodeSection{},
	}

	var currentSection *NodeSection
	var currentSubsection *NodeSubsection
	var fence fenceState

	normalizedLogicalName := textnormalization.NormalizeText(logical_name)

	processLine := func(line string) error {
		if fence.open {
			if closingFence(line, fence.fenceChar, fence.fenceLen) {
				fence.open = false
			}
			if currentSection == nil {
				return nil
			}
			if currentSubsection != nil {
				currentSubsection.Content = append(currentSubsection.Content, line)
			} else {
				currentSection.Content = append(currentSection.Content, line)
			}
			return nil
		}

		isFence, fChar, fLen := detectFence(line)
		if isFence {
			fence.open = true
			fence.fenceChar = fChar
			fence.fenceLen = fLen
			if currentSection == nil {
				return nil
			}
			if currentSubsection != nil {
				currentSubsection.Content = append(currentSubsection.Content, line)
			} else {
				currentSection.Content = append(currentSection.Content, line)
			}
			return nil
		}

		level, headingText, isHeading := parseHeading(line)

		if !isHeading || level > 2 {
			if currentSection == nil {
				if strings.TrimSpace(line) != "" {
					filereader.FileClose(reader)
					return ErrUnexpectedContentBeforeFirstHeading
				}
				return nil
			}
			if currentSubsection != nil {
				currentSubsection.Content = append(currentSubsection.Content, line)
			} else {
				currentSection.Content = append(currentSection.Content, line)
			}
			return nil
		}

		if level == 1 {
			if currentSubsection != nil && currentSection != nil {
				currentSection.Subsections = append(currentSection.Subsections, currentSubsection)
				currentSubsection = nil
			}

			newSection := &NodeSection{
				Heading:     textnormalization.NormalizeText(headingText),
				RawHeading:  line,
				Content:     []string{},
				Subsections: []*NodeSubsection{},
			}

			if node.NameSection == nil {
				if newSection.Heading != normalizedLogicalName {
					filereader.FileClose(reader)
					return ErrNodeNameDoesNotMatch
				}
				node.NameSection = newSection
			} else if newSection.Heading == "public" {
				if node.Public != nil {
					filereader.FileClose(reader)
					return ErrDuplicatePublicSection
				}
				node.Public = newSection
			} else if newSection.Heading == "agent" {
				if node.Agent != nil {
					filereader.FileClose(reader)
					return ErrDuplicateAgentSection
				}
				node.Agent = newSection
			} else {
				node.Private = append(node.Private, newSection)
			}

			currentSection = newSection
			currentSubsection = nil
			return nil
		}

		if level == 2 {
			if currentSection == nil {
				if strings.TrimSpace(line) != "" {
					filereader.FileClose(reader)
					return ErrUnexpectedContentBeforeFirstHeading
				}
				return nil
			}

			if currentSubsection != nil {
				currentSection.Subsections = append(currentSection.Subsections, currentSubsection)
				currentSubsection = nil
			}

			normalizedSub := textnormalization.NormalizeText(headingText)
			for _, existing := range currentSection.Subsections {
				if existing.Heading == normalizedSub {
					filereader.FileClose(reader)
					return ErrDuplicateSubsection
				}
			}

			currentSubsection = &NodeSubsection{
				Heading:    normalizedSub,
				RawHeading: line,
				Content:    []string{},
			}
			return nil
		}

		return nil
	}

	if heldLine != nil {
		if err := processLine(*heldLine); err != nil {
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
			return nil, fmt.Errorf("%w: %s", ErrFileUnreadable, err)
		}
		if err := processLine(line); err != nil {
			return nil, err
		}
	}

	if currentSubsection != nil && currentSection != nil {
		currentSection.Subsections = append(currentSection.Subsections, currentSubsection)
	}

	filereader.FileClose(reader)

	if node.NameSection == nil {
		return nil, ErrUnexpectedContentBeforeFirstHeading
	}

	return node, nil
}
