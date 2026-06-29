// code-from-spec: SPEC/golang/implementation/parsing/node_parsing@ER6DaHwQeusnu2_BjH066zF60do
package parsing

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
	"github.com/goccy/go-yaml"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type NodeFrontmatter struct {
	DependsOn []string
	Input     *string
	Output    *string
}

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
	Reference   CfsReference
	Frontmatter *NodeFrontmatter
	NameSection NodeSection
	Public      *NodeSection
	Agent       *NodeSection
	Private     *NodeSection
}

type rawFrontmatterNP struct {
	DependsOn []string `yaml:"depends_on"`
	Input     *string  `yaml:"input"`
	Output    *string  `yaml:"output"`
}

type headingRecordNP struct {
	level      int
	normalized string
	raw        string
	content    []string
}

func ParseNode(logicalName string) (*Node, error) {
	ref, err := CfsReferenceFromName(logicalName)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNotASpecReference, err)
	}
	if ref.NodeType != CfsNodeTypeSpec {
		return nil, fmt.Errorf("%w", ErrNotASpecReference)
	}
	if ref.Qualifier != nil {
		return nil, fmt.Errorf("%w", ErrHasQualifier)
	}

	handle, err := oslayer.OpenFile(oslayer.CfsPath(ref.Path), "read", 30000)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	var lines []string
	for {
		line, err := handle.ReadLine()
		if err != nil {
			if errors.Is(err, oslayer.ErrEndOfFile) {
				break
			}
			handle.Close()
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		lines = append(lines, line)
	}
	handle.Close()

	joined := strings.Join(lines, "\n") + "\n"
	source := []byte(joined)

	frontmatter, body, err := extractFrontmatterNP(source)
	if err != nil {
		return nil, err
	}

	headings, err := collectHeadingsNP(body)
	if err != nil {
		return nil, err
	}

	node, err := buildNodeNP(logicalName, ref, frontmatter, headings)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func extractFrontmatterNP(source []byte) (*NodeFrontmatter, []byte, error) {
	if !bytes.HasPrefix(source, []byte("---\n")) {
		return nil, source, nil
	}

	rest := source[4:]
	idx := bytes.Index(rest, []byte("\n---\n"))
	if idx < 0 {
		return nil, nil, fmt.Errorf("%w", ErrMalformedYAML)
	}

	yamlText := rest[:idx]
	body := rest[idx+5:]

	if len(yamlText) == 0 {
		return nil, body, nil
	}

	var raw rawFrontmatterNP
	if err := yaml.Unmarshal(yamlText, &raw); err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrMalformedYAML, err)
	}

	fm := &NodeFrontmatter{
		DependsOn: raw.DependsOn,
		Input:     raw.Input,
		Output:    raw.Output,
	}

	return fm, body, nil
}

func extractHeadingTextNP(h *ast.Heading, source []byte) string {
	var buf bytes.Buffer
	for c := h.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*ast.Text); ok {
			buf.Write(t.Segment.Value(source))
		}
	}
	return buf.String()
}

func headingLineStartNP(h *ast.Heading, source []byte) int {
	pos := h.Lines().At(0).Start
	for pos > 0 && source[pos-1] != '\n' {
		pos--
	}
	return pos
}

func headingLineEndNP(h *ast.Heading, source []byte) int {
	pos := h.Lines().At(0).Stop
	for pos < len(source) && source[pos] != '\n' {
		pos++
	}
	return pos
}

func contentLinesNP(body []byte, contentStart, contentEnd int) []string {
	if contentStart >= contentEnd || contentStart >= len(body) {
		return nil
	}
	raw := string(body[contentStart:contentEnd])
	parts := strings.Split(raw, "\n")
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	if len(parts) == 0 {
		return nil
	}
	return parts
}

func hasNonBlankContentNP(body []byte, start, end int) bool {
	if start >= end || start >= len(body) {
		return false
	}
	segment := body[start:end]
	for _, b := range segment {
		if b != ' ' && b != '\t' && b != '\n' && b != '\r' {
			return true
		}
	}
	return false
}

func collectHeadingsNP(body []byte) ([]headingRecordNP, error) {
	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(body))

	type rawHeadingNP struct {
		heading   *ast.Heading
		lineStart int
		lineEnd   int
	}

	var structuralHeadings []rawHeadingNP

	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		h, ok := child.(*ast.Heading)
		if !ok {
			continue
		}
		if h.Level != 1 && h.Level != 2 {
			continue
		}
		if h.Lines().Len() == 0 {
			continue
		}
		ls := headingLineStartNP(h, body)
		le := headingLineEndNP(h, body)
		structuralHeadings = append(structuralHeadings, rawHeadingNP{h, ls, le})
	}

	if len(structuralHeadings) > 0 {
		firstHeadingStart := structuralHeadings[0].lineStart
		if hasNonBlankContentNP(body, 0, firstHeadingStart) {
			return nil, fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
		}
	} else {
		if hasNonBlankContentNP(body, 0, len(body)) {
			return nil, fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
		}
	}

	records := make([]headingRecordNP, 0, len(structuralHeadings))
	for i, sh := range structuralHeadings {
		textPart := extractHeadingTextNP(sh.heading, body)
		rawLine := strings.TrimRight(string(body[sh.lineStart:sh.lineEnd]), " \t")
		normalized := NormalizeText(textPart)

		contentStart := sh.lineEnd
		if contentStart < len(body) && body[contentStart] == '\n' {
			contentStart++
		}

		var contentEnd int
		if i+1 < len(structuralHeadings) {
			contentEnd = structuralHeadings[i+1].lineStart
		} else {
			contentEnd = len(body)
		}

		cl := contentLinesNP(body, contentStart, contentEnd)

		records = append(records, headingRecordNP{
			level:      sh.heading.Level,
			normalized: normalized,
			raw:        rawLine,
			content:    cl,
		})
	}

	return records, nil
}

func buildNodeNP(logicalName string, ref *CfsReference, frontmatter *NodeFrontmatter, headings []headingRecordNP) (*Node, error) {
	var nameSection *NodeSection
	var public *NodeSection
	var agent *NodeSection
	var private *NodeSection

	var currentSection *NodeSection
	var currentSubsection *NodeSubsection

	finalizeSubsectionNP := func() {
		if currentSubsection != nil && currentSection != nil {
			currentSection.Subsections = append(currentSection.Subsections, currentSubsection)
			currentSubsection = nil
		}
	}

	finalizeSectionNP := func() {
		currentSection = nil
	}

	for _, rec := range headings {
		if rec.level == 1 {
			finalizeSubsectionNP()
			finalizeSectionNP()

			sec := &NodeSection{
				Heading:    rec.normalized,
				RawHeading: rec.raw,
				Content:    rec.content,
			}

			if nameSection == nil {
				expectedName := NormalizeText(logicalName)
				if rec.normalized != expectedName {
					return nil, fmt.Errorf("%w: got %q, want %q", ErrNodeNameDoesNotMatch, rec.normalized, expectedName)
				}
				nameSection = sec
				currentSection = sec
			} else {
				switch rec.normalized {
				case "public":
					if public != nil {
						return nil, fmt.Errorf("%w", ErrDuplicatePublicSection)
					}
					public = sec
					currentSection = sec
				case "agent":
					if agent != nil {
						return nil, fmt.Errorf("%w", ErrDuplicateAgentSection)
					}
					agent = sec
					currentSection = sec
				case "private":
					if private != nil {
						return nil, fmt.Errorf("%w", ErrDuplicatePrivateSection)
					}
					private = sec
					currentSection = sec
				default:
					return nil, fmt.Errorf("%w: %q", ErrUnrecognizedSection, rec.normalized)
				}
			}
		} else if rec.level == 2 {
			if currentSection == nil {
				return nil, fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
			}
			finalizeSubsectionNP()

			for _, existing := range currentSection.Subsections {
				if existing.Heading == rec.normalized {
					return nil, fmt.Errorf("%w: %q", ErrDuplicateSubsection, rec.normalized)
				}
			}

			currentSubsection = &NodeSubsection{
				Heading:    rec.normalized,
				RawHeading: rec.raw,
				Content:    rec.content,
			}
		}
	}

	finalizeSubsectionNP()

	if nameSection == nil {
		return nil, fmt.Errorf("%w", ErrUnexpectedContentBeforeFirstHeading)
	}

	return &Node{
		Reference:   *ref,
		Frontmatter: frontmatter,
		NameSection: *nameSection,
		Public:      public,
		Agent:       agent,
		Private:     private,
	}, nil
}
