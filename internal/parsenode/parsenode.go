// code-from-spec: ROOT/golang/internal/parsenode/code@5Oz0AIxAkUbUQUPEntNOyQHLtMY

// Package parsenode parses a spec node file (_node.md) into a structured
// representation of its level-1 sections and level-2 subsections, given a
// logical name.
//
// It uses goldmark to parse the CommonMark markdown body into an AST, then
// walks the top-level headings to extract sections. Fenced code blocks are
// handled correctly — headings inside fenced code blocks are treated as plain
// content, not as structural headings.
//
// The entry point is ParseNode. All other exported symbols are the data types
// and sentinel errors documented in the spec interface.
package parsenode

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/normalizename"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// ---------------------------------------------------------------------------
// Public data types
// ---------------------------------------------------------------------------

// Subsection represents a level-2 heading (##) and the content that follows
// it until the next structural heading.
type Subsection struct {
	// Heading is the raw text of the ## heading, without the "## " prefix.
	Heading string
	// Content is the trimmed body text that belongs to this subsection.
	Content string
}

// Section represents a level-1 heading (#) and all the content (including
// level-2 subsections) that follows it until the next level-1 heading.
type Section struct {
	// Heading is the raw text of the # heading, without the "# " prefix.
	Heading string
	// Content is the trimmed body text directly under the # heading and
	// before the first ## subsection heading (or before end-of-section if
	// there are no subsections).
	Content string
	// Subsections holds the level-2 subsections within this section.
	Subsections []Subsection
}

// ParsedNode is the fully structured result of parsing a spec node file.
type ParsedNode struct {
	// NameSection is the first level-1 section. Its heading equals (after
	// normalization) the last path segment of the logical name. Always present.
	NameSection Section
	// Public is the "# Public" section, or nil when the file has no such section.
	Public *Section
	// Agent is the "# Agent" section, or nil when the file has no such section.
	Agent *Section
	// Private holds all other level-1 sections (not name, public, or agent).
	Private []Section
}

// ---------------------------------------------------------------------------
// Sentinel errors
// ---------------------------------------------------------------------------

var (
	// ErrRead is returned when the spec file cannot be read from disk.
	ErrRead = errors.New("error reading file")

	// ErrUnexpectedContent is returned when non-blank content appears in the
	// file body before the very first level-1 heading.
	ErrUnexpectedContent = errors.New("unexpected content before first heading")

	// ErrInvalidNodeName is returned when the first level-1 heading does not
	// match the logical name after normalization.
	ErrInvalidNodeName = errors.New("node name section does not match logical name")

	// ErrDuplicatePublic is returned when the file contains more than one
	// level-1 heading that normalizes to "public".
	ErrDuplicatePublic = errors.New("duplicate public section")

	// ErrDuplicateSubsection is returned when two level-2 headings within the
	// "# Public" section normalize to the same text.
	ErrDuplicateSubsection = errors.New("duplicate subsection in public")
)

// ---------------------------------------------------------------------------
// ParseNode
// ---------------------------------------------------------------------------

// ParseNode reads and parses the spec node file identified by logicalName.
//
// The file path is resolved via logicalnames.PathFromLogicalName. The file is
// read with os.ReadFile. The markdown body (after frontmatter) is parsed with
// goldmark to produce an AST, which is then walked to collect sections and
// subsections. All heading comparisons use normalizename.NormalizeName.
//
// Returned errors always wrap one of the ErrXxx sentinel values so callers
// can use errors.Is for matching.
func ParseNode(logicalName string) (*ParsedNode, error) {
	// Step 1 — resolve logical name to a file path.
	filePath, ok := logicalnames.PathFromLogicalName(logicalName)
	if !ok {
		return nil, fmt.Errorf("%w: cannot resolve %q", ErrInvalidNodeName, logicalName)
	}

	// Step 2 — read the file.
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %s", ErrRead, filePath, err.Error())
	}

	// Step 3 — strip YAML frontmatter to obtain the markdown body.
	body := stripFrontmatter(raw)

	// Steps 4–6 — parse sections from the markdown body.
	sections, parseErr := parseSections(body)
	if parseErr != nil {
		return nil, fmt.Errorf("%s: %w", filePath, parseErr)
	}

	// Step 7 — at least one section (the name section) must be present.
	if len(sections) == 0 {
		return nil, fmt.Errorf("%s: %w: got no level-1 headings", filePath, ErrInvalidNodeName)
	}

	// Step 8 — validate the first section's heading against the logical name.
	expectedName := lastSegment(logicalName)
	if normalizename.NormalizeName(sections[0].Heading) != normalizename.NormalizeName(expectedName) {
		return nil, fmt.Errorf(
			"%s: %w: first heading is %q, expected last segment of %q",
			filePath, ErrInvalidNodeName, sections[0].Heading, logicalName,
		)
	}

	// Steps 9–10 — classify sections.
	nameSection := sections[0]
	var publicSection *Section
	var agentSection *Section
	var privateSections []Section

	for _, sec := range sections[1:] {
		norm := normalizename.NormalizeName(sec.Heading)
		switch norm {
		case "public":
			if publicSection != nil {
				return nil, fmt.Errorf("%s: %w", filePath, ErrDuplicatePublic)
			}
			// Detect duplicate subsection headings within Public.
			seen := make(map[string]struct{}, len(sec.Subsections))
			for _, sub := range sec.Subsections {
				key := normalizename.NormalizeName(sub.Heading)
				if _, exists := seen[key]; exists {
					return nil, fmt.Errorf("%s: %w: %q", filePath, ErrDuplicateSubsection, sub.Heading)
				}
				seen[key] = struct{}{}
			}
			// Copy to heap so the pointer remains valid.
			copied := sec
			publicSection = &copied

		case "agent":
			copied := sec
			agentSection = &copied

		default:
			privateSections = append(privateSections, sec)
		}
	}

	// Step 11 — assemble and return the ParsedNode.
	return &ParsedNode{
		NameSection: nameSection,
		Public:      publicSection,
		Agent:       agentSection,
		Private:     privateSections,
	}, nil
}

// ---------------------------------------------------------------------------
// Frontmatter stripping
// ---------------------------------------------------------------------------

// stripFrontmatter removes the YAML frontmatter block from raw file content
// and returns only the markdown body bytes.
//
// The frontmatter is present when the file begins with a line containing
// exactly "---" (optionally followed by "\r\n" or "\n"). The frontmatter ends
// at the second such "---" line. If the opening delimiter is not found, or
// the closing delimiter is not found, the entire content is treated as body.
func stripFrontmatter(raw []byte) []byte {
	// Normalise CRLF → LF for delimiter detection only; we return a sub-slice
	// of the (CRLF-preserved) original so that goldmark byte offsets are valid.
	//
	// However, goldmark itself handles CRLF correctly, so returning the
	// original bytes is safe.
	lines := splitLines(raw)
	if len(lines) == 0 {
		return raw
	}

	// The first line must be exactly "---" to start frontmatter.
	if strings.TrimRight(lines[0].text, "\r") != "---" {
		return raw
	}

	// Scan for the closing "---" line, starting from line index 1.
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i].text, "\r") == "---" {
			// The body starts at the byte offset immediately after this line.
			bodyStart := lines[i].end
			if bodyStart >= len(raw) {
				return []byte{}
			}
			return raw[bodyStart:]
		}
	}

	// Closing delimiter not found — treat whole file as body.
	return raw
}

// lineRecord holds a single line's text and the byte offset where the next
// line begins (i.e., the offset after the terminating '\n', or len(raw) if
// this is the last line without a trailing newline).
type lineRecord struct {
	text string // line content including '\n' if present
	end  int    // byte offset of first byte of the NEXT line
}

// splitLines splits raw bytes into a slice of lineRecords.
func splitLines(raw []byte) []lineRecord {
	var records []lineRecord
	pos := 0
	for pos < len(raw) {
		end := bytes.IndexByte(raw[pos:], '\n')
		if end < 0 {
			// Last line with no trailing newline.
			records = append(records, lineRecord{
				text: string(raw[pos:]),
				end:  len(raw),
			})
			break
		}
		absEnd := pos + end + 1 // inclusive of '\n'
		records = append(records, lineRecord{
			text: string(raw[pos : pos+end]), // text without '\n'
			end:  absEnd,
		})
		pos = absEnd
	}
	return records
}

// ---------------------------------------------------------------------------
// Section parsing via goldmark AST
// ---------------------------------------------------------------------------

// headingEntry is an internal record representing one structural heading
// found during the goldmark AST walk.
type headingEntry struct {
	level     int    // 1 or 2
	text      string // raw heading text (without # prefix)
	lineStart int    // byte offset of the '#' in source
	lineEnd   int    // byte offset of the first byte after the heading line
}

// parseSections parses a markdown body byte slice and returns a flat list of
// Sections, each populated with its Subsections and trimmed Content.
//
// Structural headings inside fenced code blocks are detected and skipped by
// post-processing the raw source between headings: if a heading's line-start
// is inside a fenced block (odd number of fence markers before it), it is
// excluded from the heading list.
func parseSections(body []byte) ([]Section, error) {
	if len(bytes.TrimSpace(body)) == 0 {
		return nil, nil
	}

	// Parse the body with goldmark.
	md := goldmark.New()
	source := body
	doc := md.Parser().Parse(text.NewReader(source))

	// Collect all level-1 and level-2 headings from the document's direct
	// children. The goldmark AST structure places headings as siblings of the
	// blocks that follow them — not as parents.
	var headings []headingEntry
	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		h, ok := child.(*ast.Heading)
		if !ok {
			continue
		}
		if h.Level != 1 && h.Level != 2 {
			continue
		}
		headings = append(headings, headingEntry{
			level:     h.Level,
			text:      headingText(h, source),
			lineStart: headingLineStart(h, source),
			lineEnd:   headingLineEnd(h, source),
		})
	}

	// Goldmark already skips headings inside fenced code blocks during parsing,
	// so the headings slice only contains structural headings. No additional
	// fence detection is needed.

	if len(headings) == 0 {
		// No headings found. Check for unexpected content.
		if hasNonBlankContent(source) {
			return nil, ErrUnexpectedContent
		}
		return nil, nil
	}

	// Verify there is no non-blank content before the first heading.
	if hasNonBlankContent(source[:headings[0].lineStart]) {
		return nil, ErrUnexpectedContent
	}

	// Build Section values from the heading list.
	//
	// Algorithm:
	//   - Iterate headings in order.
	//   - On a level-1 heading:
	//       - Flush any open subsection into the current section.
	//       - Flush the current section into the result list.
	//       - Start a new current section.
	//   - On a level-2 heading:
	//       - Flush any open subsection into the current section.
	//       - Record where the new subsection's content begins.
	//   - After all headings: flush the remaining open subsection and section.
	//
	// Content end for heading[i] = headings[i+1].lineStart (or len(source)).

	type sectionState struct {
		sec          Section
		contentStart int // byte offset where section-level content begins
		contentEnd   int // byte offset where section-level content ends
	}

	var result []Section
	var cur *sectionState // currently open section

	// State for the open subsection within cur.
	var subHeading string // empty means no open subsection
	var subStart int      // byte offset where subsection content begins

	// endOfContent returns the byte offset where heading[i]'s content ends.
	endOfContent := func(i int) int {
		if i+1 < len(headings) {
			return headings[i+1].lineStart
		}
		return len(source)
	}

	for i, h := range headings {
		switch h.level {
		case 1:
			// Close any open subsection.
			if cur != nil && subHeading != "" {
				subContent := trimContent(source[subStart:cur.contentEnd])
				cur.sec.Subsections = append(cur.sec.Subsections, Subsection{
					Heading: subHeading,
					Content: subContent,
				})
				subHeading = ""
			}
			// Close any open section.
			if cur != nil {
				cur.sec.Content = trimContent(source[cur.contentStart:cur.contentEnd])
				result = append(result, cur.sec)
			}
			// Open a new section.
			//
			// contentEnd is set to the start of the NEXT heading (level-1 or
			// level-2). This will be overridden when we encounter the first ##
			// subsection inside this section.
			cur = &sectionState{
				sec:          Section{Heading: h.text},
				contentStart: h.lineEnd,
				contentEnd:   endOfContent(i),
			}
			subHeading = ""

		case 2:
			if cur == nil {
				// Level-2 heading before any level-1 heading — ignore.
				continue
			}
			if subHeading == "" {
				// First subsection in this section: the section's direct content
				// ends where this ## heading line begins.
				cur.contentEnd = h.lineStart
			} else {
				// Close the previously open subsection.
				subEnd := h.lineStart
				subContent := trimContent(source[subStart:subEnd])
				cur.sec.Subsections = append(cur.sec.Subsections, Subsection{
					Heading: subHeading,
					Content: subContent,
				})
			}
			// Open the new subsection.
			subHeading = h.text
			subStart = h.lineEnd
		}
	}

	// Flush the last open subsection and section.
	if cur != nil {
		if subHeading != "" {
			// Subsection content runs to the end of the containing section.
			subContent := trimContent(source[subStart:endOfContent(len(headings)-1)])
			cur.sec.Subsections = append(cur.sec.Subsections, Subsection{
				Heading: subHeading,
				Content: subContent,
			})
		}
		cur.sec.Content = trimContent(source[cur.contentStart:cur.contentEnd])
		result = append(result, cur.sec)
	}

	return result, nil
}

// ---------------------------------------------------------------------------
// goldmark AST helpers
// ---------------------------------------------------------------------------

// headingText returns the plain text content of a goldmark Heading node by
// walking its inline children and concatenating all *ast.Text segments.
func headingText(h *ast.Heading, source []byte) string {
	var buf bytes.Buffer
	for c := h.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*ast.Text); ok {
			buf.Write(t.Segment.Value(source))
		}
	}
	return buf.String()
}

// headingLineStart returns the byte offset of the '#' character that begins
// the heading line in source.
//
// goldmark's Lines().At(0).Start points to the first character of the heading
// text (i.e., after the "# " prefix), so we scan backward to find the
// preceding '\n' (or the start of source for the very first line).
func headingLineStart(h *ast.Heading, source []byte) int {
	pos := h.Lines().At(0).Start
	for pos > 0 && source[pos-1] != '\n' {
		pos--
	}
	return pos
}

// headingLineEnd returns the byte offset of the first byte after the heading
// line — i.e., the offset immediately after the terminating '\n' (or '\r\n').
func headingLineEnd(h *ast.Heading, source []byte) int {
	// Lines().At(0).Stop is the byte after the last text character, typically
	// pointing at '\n' or '\r'.
	stop := h.Lines().At(0).Stop
	if stop < len(source) {
		if source[stop] == '\r' && stop+1 < len(source) && source[stop+1] == '\n' {
			return stop + 2
		}
		if source[stop] == '\n' || source[stop] == '\r' {
			return stop + 1
		}
	}
	return stop
}

// ---------------------------------------------------------------------------
// Content helper: trimContent
// ---------------------------------------------------------------------------

// trimContent converts a raw byte slice (a contiguous range of source bytes)
// into a trimmed content string:
//
//  1. Normalize CRLF → LF.
//  2. Split into lines.
//  3. Remove all leading blank lines.
//  4. Remove all trailing blank lines.
//  5. Join remaining lines with "\n".
//
// A blank line is one that is empty or contains only whitespace characters.
func trimContent(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	// Normalize line endings.
	normalized := bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n"))
	// Remove a single trailing newline added by the heading-line-end offset so
	// we don't introduce a spurious leading blank line.
	lines := strings.Split(string(normalized), "\n")

	// Remove leading blank lines.
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	// Remove trailing blank lines.
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(lines, "\n")
}

// hasNonBlankContent reports whether the byte slice contains at least one line
// with non-whitespace content.
func hasNonBlankContent(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	for _, line := range strings.Split(string(b), "\n") {
		if strings.TrimSpace(line) != "" {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Logical name helpers
// ---------------------------------------------------------------------------

// lastSegment derives the expected heading text from a logical name:
//  1. Strip any parenthetical qualifier: "ROOT/x/y(z)" → "ROOT/x/y".
//  2. Return the last path segment after the final '/'. For "ROOT" with no
//     '/', return "ROOT" itself.
func lastSegment(logicalName string) string {
	name := logicalName
	// Strip qualifier.
	if idx := strings.Index(name, "("); idx >= 0 {
		name = name[:idx]
	}
	// Take last segment.
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		return name[idx+1:]
	}
	return name
}
