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

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"`

### NodeBuilder

```go
type NodeBuilder struct { /* unexported fields */ }

func CreateSpecNode(t *testing.T, logicalName string) *NodeBuilder
func (b *NodeBuilder) SetOutput(value string)
func (b *NodeBuilder) SetInput(value string)
func (b *NodeBuilder) AddDependsOn(value string)
func (b *NodeBuilder) SetPublic(content string)
func (b *NodeBuilder) SetAgent(content string)
func (b *NodeBuilder) SetPrivate(content string)
func (b *NodeBuilder) Write()
```

#### CreateSpecNode

Creates a `NodeBuilder` for the given logical name
(e.g. `SPEC/a/b`). The builder stores `t` and the
logical name for later use by `Write`.

#### SetOutput, SetInput

Set the `output` or `input` frontmatter field.

#### AddDependsOn

Appends a `depends_on` entry. Can be called multiple
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
   `input`, `depends_on` between `---` delimiters.
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
	input       *string
	dependsOn   []string
	public      *string
	agent       *string
	private     *string
}

func CreateSpecNode(t *testing.T, logicalName string) *NodeBuilder {
	t.Helper()
	return &NodeBuilder{t: t, logicalName: logicalName}
}

func (b *NodeBuilder) SetOutput(value string)    { b.output = &value }
func (b *NodeBuilder) SetInput(value string)     { b.input = &value }
func (b *NodeBuilder) AddDependsOn(value string) { b.dependsOn = append(b.dependsOn, value) }
func (b *NodeBuilder) SetPublic(content string)  { b.public = &content }
func (b *NodeBuilder) SetAgent(content string)   { b.agent = &content }
func (b *NodeBuilder) SetPrivate(content string) { b.private = &content }

func (b *NodeBuilder) Write() {
	b.t.Helper()
	var buf strings.Builder

	if b.output != nil || b.input != nil || len(b.dependsOn) > 0 {
		buf.WriteString("---\n")
		if len(b.dependsOn) > 0 {
			buf.WriteString("depends_on:\n")
			for _, dep := range b.dependsOn {
				buf.WriteString("  - " + dep + "\n")
			}
		}
		if b.input != nil {
			buf.WriteString("input: " + *b.input + "\n")
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
