// code-from-spec: ROOT/golang/internal/parsenode/code@PENDING
package parsenode

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/normalizename"
)

// Subsection represents a level-2 heading and its content within a section.
type Subsection struct {
	Heading string
	Content string
}

// Section represents a level-1 heading, its content, and any level-2 subsections.
type Section struct {
	Heading     string
	Content     string
	Subsections []Subsection
}

// ParsedNode is the structured representation of a spec node file.
type ParsedNode struct {
	NameSection Section
	Public      *Section
	Agent       *Section
	Private     []Section
}

// Error sentinels for ParseNode.
var (
	ErrRead                = errors.New("error reading file")
	ErrUnexpectedContent   = errors.New("unexpected content before first heading")
	ErrInvalidNodeName     = errors.New("node name section does not match logical name")
	ErrDuplicatePublic     = errors.New("duplicate public section")
	ErrDuplicateSubsection = errors.New("duplicate subsection in public")
)

// ParseNode reads a spec node file identified by logicalName and returns
// a structured ParsedNode with its sections parsed from the markdown body.
func ParseNode(logicalName string) (*ParsedNode, error) {
	// Step 1: Resolve logical name to file path.
	filePath, ok := logicalnames.PathFromLogicalName(logicalName)
	if !ok {
		return nil, fmt.Errorf("%w: cannot resolve logical name %q", ErrRead, logicalName)
	}

	// Step 2: Read the file.
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrRead, filePath, err)
	}

	// Step 3: Skip frontmatter.
	body := skipFrontmatter(data)

	// Step 4: Parse the body with goldmark.
	md := goldmark.New()
	source := body
	doc := md.Parser().Parse(text.NewReader(source))

	// Collect all level-1 and level-2 headings in document order.
	var headings []headingInfo

	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		if h, ok := child.(*ast.Heading); ok && (h.Level == 1 || h.Level == 2) {
			headings = append(headings, headingInfo{
				node:  h,
				text:  headingText(h, source),
				level: h.Level,
			})
		}
	}

	// Check for unexpected content before the first level-1 heading.
	if len(headings) == 0 || headings[0].level != 1 {
		// Check if there is any non-blank content at all.
		if strings.TrimSpace(string(source)) != "" {
			return nil, fmt.Errorf("%w", ErrUnexpectedContent)
		}
	} else {
		// Check for non-blank content before the first heading.
		firstHeadingStart := headingLineStart(headings[0].node, source)
		before := source[:firstHeadingStart]
		if strings.TrimSpace(string(before)) != "" {
			return nil, fmt.Errorf("%w", ErrUnexpectedContent)
		}
	}

	// Build sections from headings and their content regions.
	sections := buildSections(headings, source)

	// Trim leading/trailing blank lines from all section and subsection content.
	for i := range sections {
		sections[i].Content = trimBlankLines(sections[i].Content)
		for j := range sections[i].Subsections {
			sections[i].Subsections[j].Content = trimBlankLines(sections[i].Subsections[j].Content)
		}
	}

	if len(sections) == 0 {
		return nil, fmt.Errorf("%w", ErrUnexpectedContent)
	}

	// Validate: first heading must match logical name.
	normalizedFirst := normalizename.NormalizeName(sections[0].Heading)
	normalizedLogical := normalizename.NormalizeName(logicalName)
	if normalizedFirst != normalizedLogical {
		return nil, fmt.Errorf("%w: heading %q does not match %q", ErrInvalidNodeName, sections[0].Heading, logicalName)
	}

	result := &ParsedNode{
		NameSection: sections[0],
	}

	// Walk remaining sections, classifying as public, agent, or private.
	for i := 1; i < len(sections); i++ {
		s := sections[i]
		normalized := normalizename.NormalizeName(s.Heading)

		switch normalized {
		case normalizename.NormalizeName("Public"):
			if result.Public != nil {
				return nil, fmt.Errorf("%w", ErrDuplicatePublic)
			}
			// Check for duplicate subsections.
			seen := make(map[string]bool)
			for _, sub := range s.Subsections {
				normalizedSub := normalizename.NormalizeName(sub.Heading)
				if seen[normalizedSub] {
					return nil, fmt.Errorf("%w: %q", ErrDuplicateSubsection, sub.Heading)
				}
				seen[normalizedSub] = true
			}
			sec := s
			result.Public = &sec

		case normalizename.NormalizeName("Agent"):
			sec := s
			result.Agent = &sec

		default:
			result.Private = append(result.Private, s)
		}
	}

	return result, nil
}

// skipFrontmatter removes YAML frontmatter delimited by --- lines.
func skipFrontmatter(data []byte) []byte {
	s := string(data)
	if !strings.HasPrefix(s, "---") {
		return data
	}
	// Find the closing ---.
	rest := s[3:]
	// Skip the rest of the first --- line.
	idx := strings.Index(rest, "\n")
	if idx < 0 {
		return data
	}
	rest = rest[idx+1:]
	// Find next ---.
	closingIdx := strings.Index(rest, "---")
	if closingIdx < 0 {
		return data
	}
	// Skip past the closing --- line.
	afterClosing := rest[closingIdx+3:]
	nlIdx := strings.Index(afterClosing, "\n")
	if nlIdx < 0 {
		return []byte{}
	}
	return []byte(afterClosing[nlIdx+1:])
}

// headingText extracts the inline text content of a heading node.
func headingText(h *ast.Heading, source []byte) string {
	var buf bytes.Buffer
	for c := h.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*ast.Text); ok {
			buf.Write(t.Segment.Value(source))
		}
	}
	return buf.String()
}

// headingLineStart returns the byte offset where the heading line begins
// (the # character), scanning backward from the text content start.
func headingLineStart(h *ast.Heading, source []byte) int {
	pos := h.Lines().At(0).Start
	for pos > 0 && source[pos-1] != '\n' {
		pos--
	}
	return pos
}

// buildSections constructs Section values from heading info and source bytes.
func buildSections(headings []headingInfo, source []byte) []Section {
	// First pass: identify level-1 section boundaries.
	var l1Indices []int
	for i, h := range headings {
		if h.level == 1 {
			l1Indices = append(l1Indices, i)
		}
	}

	var sections []Section
	for idx, l1Pos := range l1Indices {
		h := headings[l1Pos]

		// Content starts after the heading line.
		contentStart := h.node.Lines().At(0).Stop

		// Content ends at the next L1 heading's line start, or end of source.
		var contentEnd int
		if idx+1 < len(l1Indices) {
			nextH := headings[l1Indices[idx+1]]
			contentEnd = headingLineStart(nextH.node, source)
		} else {
			contentEnd = len(source)
		}

		// Find level-2 headings within this section's range.
		var subsections []Subsection
		sectionContent := string(source[contentStart:contentEnd])

		// Collect L2 headings that belong to this L1 section.
		var l2InSection []int
		for i := l1Pos + 1; i < len(headings); i++ {
			if headings[i].level == 1 {
				break
			}
			if headings[i].level == 2 {
				l2InSection = append(l2InSection, i)
			}
		}

		if len(l2InSection) > 0 {
			// Content before the first L2 heading is the section's own content.
			firstL2Start := headingLineStart(headings[l2InSection[0]].node, source)
			sectionContent = string(source[contentStart:firstL2Start])

			// Build subsections.
			for j, l2Idx := range l2InSection {
				subH := headings[l2Idx]
				subContentStart := subH.node.Lines().At(0).Stop
				var subContentEnd int
				if j+1 < len(l2InSection) {
					subContentEnd = headingLineStart(headings[l2InSection[j+1]].node, source)
				} else {
					subContentEnd = contentEnd
				}
				subsections = append(subsections, Subsection{
					Heading: subH.text,
					Content: string(source[subContentStart:subContentEnd]),
				})
			}
		}

		sections = append(sections, Section{
			Heading:     h.text,
			Content:     sectionContent,
			Subsections: subsections,
		})
	}

	return sections
}

// headingInfo stores information about a heading encountered during parsing.
type headingInfo struct {
	node  *ast.Heading
	text  string
	level int
}

// trimBlankLines removes leading and trailing blank lines from content.
func trimBlankLines(s string) string {
	lines := strings.Split(s, "\n")

	// Trim leading blank lines.
	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}

	// Trim trailing blank lines.
	end := len(lines)
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}

	if start >= end {
		return ""
	}

	return strings.Join(lines[start:end], "\n")
}
