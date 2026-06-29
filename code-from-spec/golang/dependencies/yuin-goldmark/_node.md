# SPEC/golang/dependencies/yuin-goldmark

CommonMark Markdown parser for Go:
`github.com/yuin/goldmark`.

MIT licensed. Produces an AST from Markdown source,
enabling structured traversal of headings, paragraphs,
code blocks, and other elements without implementing
parsing logic.

# Public

## Import

```go
import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)
```

## Parsing into AST

```go
md := goldmark.New()
source := []byte(markdownContent)
doc := md.Parser().Parse(text.NewReader(source))
```

`Parse` returns an `ast.Node` representing the document
root. The `source` byte slice must be retained — AST
nodes reference positions within it.

## AST structure

The document root is a container. Its direct children
are block-level nodes: `*ast.Heading`,
`*ast.Paragraph`, `*ast.FencedCodeBlock`,
`*ast.List`, `*ast.ThematicBreak`, etc.

Headings are containers for inline nodes (their text
content), but they do **not** contain the blocks that
follow them. A heading and the paragraphs "under" it
are siblings, not parent-child:

```
Document
├── Heading(1)       ← children are inline
├── Paragraph        ← sibling, not child
├── FencedCodeBlock  ← sibling
├── Heading(2)
├── Paragraph
└── Heading(1)
```

## Heading

```go
type Heading struct {
	ast.BaseBlock
	Level int // 1–6
}
```

Kind: `ast.KindHeading`.

Check and cast:

```go
if heading, ok := n.(*ast.Heading); ok {
	level := heading.Level
}
```

## Block position in source — Lines()

Block nodes inherit `Lines()` from `ast.BaseBlock`.
It returns `*text.Segments` — a collection of
`text.Segment` values, each with `Start` and `Stop`
byte offsets into the source.

```go
lines := node.Lines()
lines.Len()          // number of segments
lines.At(i)          // returns text.Segment at index i
```

For an ATX heading (`# Foo`), `Lines()` contains one
segment covering only the heading **text content** —
it does **not** include the `#` prefix or the space
after it. For `# Foo`, `Lines().At(0).Start` points
to `F`, not `#`.

## Extracting heading text

The text content of a heading is stored in its inline
children. Walk the children and concatenate `*ast.Text`
segments:

```go
func headingText(h *ast.Heading, source []byte) string {
	var buf bytes.Buffer
	for c := h.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*ast.Text); ok {
			buf.Write(t.Segment.Value(source))
		}
	}
	return buf.String()
}
```

Returns the text without the `#` prefix.

## Extracting raw heading line

Since `Lines().At(0).Start` points to the text content
(after `# `), not the `#` itself, scan backward through
the source to find the preceding `\n`. The byte after
that `\n` is the start of the heading line. If no `\n`
is found, the heading is at the start of the source
(offset 0).

```go
func headingLineStart(h *ast.Heading, source []byte) int {
	pos := h.Lines().At(0).Start
	for pos > 0 && source[pos-1] != '\n' {
		pos--
	}
	return pos
}
```

The raw heading line is then:
`source[headingLineStart:headingContentStop]` where
`headingContentStop` is `Lines().At(0).Stop`.

## Extracting raw source between headings

- **Start of section content**: `Lines().At(0).Stop`
  of the heading (first byte after the heading line).
- **End of section**: `headingLineStart` of the next
  heading, or `len(source)` if no next heading.

```go
start := headingA.Lines().At(0).Stop
stop := headingLineStart(headingB, source)
content := source[start:stop]
```

## AST traversal

```go
ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
	if heading, ok := n.(*ast.Heading); ok && entering {
		// heading.Level is 1–6
	}
	return ast.WalkContinue, nil
})
```

`Walk` visits every node depth-first. Each node is
visited twice: `entering=true`, then `entering=false`.

## Iterating direct children

```go
for child := node.FirstChild(); child != nil; child = child.NextSibling() {
	// process child
}
```
