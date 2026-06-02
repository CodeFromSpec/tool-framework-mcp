// code-from-spec: ROOT/golang/implementation/parsing/node_parsing@g8K4SlVAFNy_lRvz7EiomphORhk
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
var ErrUnexpectedContentBeforeFirstHeading = errors.New("file body has non-blank content before the first level-1 heading, or has no level-1 heading at all")
var ErrNodeNameDoesNotMatch = errors.New("first heading does not match the logical name after normalization")
var ErrDuplicatePublicSection = errors.New("more than one Public section exists")
var ErrDuplicateAgentSection = errors.New("more than one Agent section exists")
var ErrDuplicateSubsection = errors.New("two level-2 headings within the same section normalize to the same text")

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

type sectionKind int

const (
	kindName sectionKind = iota
	kindPublic
	kindAgent
	kindPrivate
)

type parsedSection struct {
	section *NodeSection
	kind    sectionKind
}

func parseHeading(line string) (level int, text string, isHeading bool) {
	i := 0
	for i < len(line) && line[i] == '#' {
		i++
	}
	if i == 0 || i >= len(line) || line[i] != ' ' {
		return 0, "", false
	}
	raw := strings.TrimSpace(line[i+1:])
	for strings.HasSuffix(raw, "#") {
		trimmed := strings.TrimRight(raw, "#")
		trimmed = strings.TrimRight(trimmed, " ")
		if trimmed == raw {
			break
		}
		raw = trimmed
	}
	return i, raw, true
}

func isFenceOpener(line string) (bool, byte, int) {
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
	if count >= 3 {
		return true, ch, count
	}
	return false, 0, 0
}

func isFenceCloser(line string, fenceChar byte, fenceLength int) bool {
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

func NodeParse(logical_name string) (*Node, error) {
	if logicalnames.LogicalNameIsArtifact(logical_name) {
		return nil, fmt.Errorf("%w", ErrNotARootReference)
	}

	if logicalnames.LogicalNameHasQualifier(logical_name) {
		return nil, fmt.Errorf("%w", ErrHasQualifier)
	}

	filePath, err := logicalnames.LogicalNameToPath(logical_name)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	reader, err := filereader.FileOpen(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}
	defer filereader.FileClose(reader)

	var pendingLine *string

	readLine := func() (string, error) {
		if pendingLine != nil {
			line := *pendingLine
			pendingLine = nil
			return line, nil
		}
		return filereader.FileReadLine(reader)
	}

	firstLine, err := readLine()
	if err != nil {
		if errors.Is(err, filereader.ErrEndOfFile) {
			return nil, fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
		}
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	if firstLine == "---" {
		for {
			line, err := readLine()
			if err != nil {
				if errors.Is(err, filereader.ErrEndOfFile) {
					return nil, fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
				}
				return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
			}
			if line == "---" {
				break
			}
		}
	} else if firstLine != "" {
		pendingLine = &firstLine
	}

	var sections []parsedSection
	var currentSection *NodeSection
	var currentSectionKind sectionKind
	var currentSubsection *NodeSubsection
	foundNameSection := false
	publicCount := 0
	agentCount := 0
	inFence := false
	var fenceChar byte
	fenceLength := 0

	for {
		line, err := readLine()
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				break
			}
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}

		if !inFence {
			if ok, ch, length := isFenceOpener(line); ok {
				inFence = true
				fenceChar = ch
				fenceLength = length
				if currentSection != nil {
					if currentSubsection != nil {
						currentSubsection.Content = append(currentSubsection.Content, line)
					} else {
						currentSection.Content = append(currentSection.Content, line)
					}
				}
				continue
			}
		} else {
			if isFenceCloser(line, fenceChar, fenceLength) {
				inFence = false
			}
			if currentSection != nil {
				if currentSubsection != nil {
					currentSubsection.Content = append(currentSubsection.Content, line)
				} else {
					currentSection.Content = append(currentSection.Content, line)
				}
			}
			continue
		}

		level, headingText, isHeading := parseHeading(line)
		if !isHeading {
			level = 0
		}

		if isHeading && level == 1 {
			if currentSection != nil {
				sections = append(sections, parsedSection{section: currentSection, kind: currentSectionKind})
			}
			normalized := textnormalization.NormalizeText(headingText)
			newSection := &NodeSection{
				Heading:     normalized,
				RawHeading:  line,
				Content:     []string{},
				Subsections: []*NodeSubsection{},
			}
			currentSubsection = nil

			if !foundNameSection {
				foundNameSection = true
				expected := textnormalization.NormalizeText(logical_name)
				if normalized != expected {
					return nil, fmt.Errorf("%w", ErrNodeNameDoesNotMatch)
				}
				currentSectionKind = kindName
			} else if normalized == "public" {
				publicCount++
				if publicCount > 1 {
					return nil, fmt.Errorf("%w", ErrDuplicatePublicSection)
				}
				currentSectionKind = kindPublic
			} else if normalized == "agent" {
				agentCount++
				if agentCount > 1 {
					return nil, fmt.Errorf("%w", ErrDuplicateAgentSection)
				}
				currentSectionKind = kindAgent
			} else {
				currentSectionKind = kindPrivate
			}

			currentSection = newSection
			continue
		}

		if isHeading && level == 2 {
			if currentSection == nil {
				if strings.TrimSpace(line) != "" {
					return nil, fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
				}
				continue
			}
			normalized := textnormalization.NormalizeText(headingText)
			for _, sub := range currentSection.Subsections {
				if sub.Heading == normalized {
					return nil, fmt.Errorf("%w", ErrDuplicateSubsection)
				}
			}
			newSub := &NodeSubsection{
				Heading:    normalized,
				RawHeading: line,
				Content:    []string{},
			}
			currentSection.Subsections = append(currentSection.Subsections, newSub)
			currentSubsection = newSub
			continue
		}

		if currentSection == nil {
			if strings.TrimSpace(line) != "" {
				return nil, fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
			}
			continue
		}

		if currentSubsection != nil {
			currentSubsection.Content = append(currentSubsection.Content, line)
		} else {
			currentSection.Content = append(currentSection.Content, line)
		}
	}

	if currentSection != nil {
		sections = append(sections, parsedSection{section: currentSection, kind: currentSectionKind})
	}

	if !foundNameSection {
		return nil, fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
	}

	node := &Node{
		Private: []*NodeSection{},
	}

	for _, ps := range sections {
		switch ps.kind {
		case kindName:
			node.NameSection = ps.section
		case kindPublic:
			node.Public = ps.section
		case kindAgent:
			node.Agent = ps.section
		case kindPrivate:
			node.Private = append(node.Private, ps.section)
		}
	}

	return node, nil
}
