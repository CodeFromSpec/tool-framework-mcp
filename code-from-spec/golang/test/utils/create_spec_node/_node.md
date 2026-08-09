---
output: internal/testutils/create_spec_node.go
---

# SPEC/golang/test/utils/create_spec_node

Test helpers for creating `_node.md` files on disk.
Two modes: a builder for valid nodes (correct format
guaranteed), and a raw writer for arbitrary content
(for testing parse error cases).

# Public

## Package

`package testutils`

## Interface

`import "github.com/CodeFromSpec/tool-framework-mcp/v6/internal/testutils"`

### NodeBuilder

```go
type NodeBuilder struct { /* unexported fields */ }

func CreateSpecNode(t *testing.T, logicalName string) *NodeBuilder
func (b *NodeBuilder) SetOutput(value string)
func (b *NodeBuilder) SetInputScalar(value string)
func (b *NodeBuilder) SetInputList(values []string)
func (b *NodeBuilder) AddImport(value string)
func (b *NodeBuilder) SetPublic(content string)
func (b *NodeBuilder) SetAgent(content string)
func (b *NodeBuilder) SetPrivate(content string)
func (b *NodeBuilder) Write()
```

#### CreateSpecNode

Creates a `NodeBuilder` for the given logical name
(e.g. `SPEC/a/b`). The builder stores `t` and the
logical name for later use by `Write`.

#### SetOutput

Set the `output` frontmatter field.

#### SetInputScalar

Set the `input` frontmatter field as a single scalar
string (`input: value`). Mutually exclusive with
`SetInputList` — calling both on the same builder before
`Write` is a test-authoring error; `Write` calls
`t.Fatalf`.

#### SetInputList

Set the `input` frontmatter field as a YAML list
(`input:` followed by one `- value` line per entry),
even when given a single value. Use this to exercise the
list-shaped form of `input` explicitly, including the
single-element case. Mutually exclusive with
`SetInputScalar`.

#### AddImport

Appends an `imports` entry. Can be called multiple
times.

#### SetPublic, SetAgent, SetPrivate

Set the content for the `# Public`, `# Agent`, or
`# Private` section. The content is placed after the
section heading — the heading itself is added
automatically by `Write`.

#### Write

Writes the `_node.md` file to disk. Derives the file
path from the logical name (`SPEC/a/b` →
`code-from-spec/a/b/_node.md`). Creates intermediate
directories. Assembles the file content:

1. Frontmatter block (if any field was set): `output`,
   `input`, `imports` between `---` delimiters.
2. Node name heading: `# <logicalName>`.
3. `# Public` section (if set).
4. `# Agent` section (if set).
5. `# Private` section (if set).

Calls `t.Helper()`. Calls `t.Fatalf` on failure.

### WriteRawNode

```go
func WriteRawNode(t *testing.T, logicalName string, content string)
```

Creates a `_node.md` file at the path derived from
the logical name, with `content` written exactly as
provided — no heading, no frontmatter, no validation.
Creates intermediate directories.

Use this for tests that need malformed content:
missing headings, invalid frontmatter, wrong node
names, etc.

Calls `t.Helper()`. Calls `t.Fatalf` on failure.

# Agent

## Ownership

This file declares and implements:
- Types: `NodeBuilder`
- Functions: `CreateSpecNode`, `WriteRawNode`

To avoid name collisions with other files in this
package, all identifiers you declare beyond the ones
listed in the Ownership section (functions, variables,
types) must use the suffix `CSN`.

## Reference implementation

```go
type NodeBuilder struct {
	t           *testing.T
	logicalName string
	output      *string
	inputScalar *string
	inputList   []string
	imports     []string
	public      *string
	agent       *string
	private     *string
}

func CreateSpecNode(t *testing.T, logicalName string) *NodeBuilder {
	t.Helper()
	return &NodeBuilder{t: t, logicalName: logicalName}
}

func (b *NodeBuilder) SetOutput(value string)        { b.output = &value }
func (b *NodeBuilder) SetInputScalar(value string)   { b.inputScalar = &value }
func (b *NodeBuilder) SetInputList(values []string)  { b.inputList = values }
func (b *NodeBuilder) AddImport(value string)        { b.imports = append(b.imports, value) }
func (b *NodeBuilder) SetPublic(content string)      { b.public = &content }
func (b *NodeBuilder) SetAgent(content string)       { b.agent = &content }
func (b *NodeBuilder) SetPrivate(content string)     { b.private = &content }

func (b *NodeBuilder) Write() {
	b.t.Helper()

	if b.inputScalar != nil && len(b.inputList) > 0 {
		b.t.Fatalf("CreateSpecNode.Write: SetInputScalar and SetInputList are mutually exclusive")
	}

	var buf strings.Builder

	if b.output != nil || b.inputScalar != nil || len(b.inputList) > 0 || len(b.imports) > 0 {
		buf.WriteString("---\n")
		if len(b.imports) > 0 {
			buf.WriteString("imports:\n")
			for _, dep := range b.imports {
				buf.WriteString("  - " + dep + "\n")
			}
		}
		if b.inputScalar != nil {
			buf.WriteString("input: " + *b.inputScalar + "\n")
		} else if len(b.inputList) > 0 {
			buf.WriteString("input:\n")
			for _, v := range b.inputList {
				buf.WriteString("  - " + v + "\n")
			}
		}
		if b.output != nil {
			buf.WriteString("output: " + *b.output + "\n")
		}
		buf.WriteString("---\n")
	}

	buf.WriteString("# " + b.logicalName + "\n")

	if b.public != nil {
		buf.WriteString("\n# Public\n")
		buf.WriteString(*b.public + "\n")
	}
	if b.agent != nil {
		buf.WriteString("\n# Agent\n")
		buf.WriteString(*b.agent + "\n")
	}
	if b.private != nil {
		buf.WriteString("\n# Private\n")
		buf.WriteString(*b.private + "\n")
	}

	relative := strings.TrimPrefix(b.logicalName, "SPEC/")
	path := filepath.Join("code-from-spec", filepath.FromSlash(relative), "_node.md")
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		b.t.Fatalf("CreateSpecNode.Write: mkdir %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(buf.String()), 0644); err != nil {
		b.t.Fatalf("CreateSpecNode.Write: write %s: %v", path, err)
	}
}

func WriteRawNode(t *testing.T, logicalName string, content string) {
	t.Helper()
	relative := strings.TrimPrefix(logicalName, "SPEC/")
	path := filepath.Join("code-from-spec", filepath.FromSlash(relative), "_node.md")
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("WriteRawNode: mkdir %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteRawNode: write %s: %v", path, err)
	}
}
```
