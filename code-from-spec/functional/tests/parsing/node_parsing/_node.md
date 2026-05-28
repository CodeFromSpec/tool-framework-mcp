---
depends_on:
  - ROOT/functional/logic/parsing/node_parsing(interface)
outputs:
  - id: node_parsing_tests
    path: code-from-spec/functional/tests/parsing/node_parsing/output.md
---

# ROOT/functional/tests/parsing/node_parsing

Test cases for the node parsing component.

# Public

## Test cases

### Happy path

#### Minimal node -- name section only

Create a node file for logical name ROOT/x with only a
heading and description. Call ParseNode with "ROOT/x".

Expect name section heading = "root/x" (normalized),
name section content = "This node has only a name section.",
no subsections, public = empty, private = empty.

#### Full node -- name, public, private sections

Create a node file for ROOT/payments/fees with frontmatter
(depends_on, outputs), a name section, a public section
with Interface and Constraints subsections, and private
sections (Implementation, Decisions). Call ParseNode.

Expect name section heading = "root/payments/fees",
name section content = "Calculates transaction fees.",
public section with heading = "public", two subsections
(interface with content, constraints with content), and
two private sections (implementation, decisions) each with
content.

#### Node with no public section

Create a node file for ROOT/decisions with only a name
section and a private section (Rationale). Call ParseNode.

Expect public = empty, name section heading =
"root/decisions", one private section with heading =
"rationale".

#### Public section with content before first subsection

Create a node file for ROOT/a with a public section that
has direct content before any subsection, plus one
subsection. Call ParseNode.

Expect public content = "This is direct content of the
public section.", and one subsection with heading =
"interface".

### Heading normalization

#### Case insensitive public detection

Create a node with "# PUBLIC" as the public heading. Call
ParseNode. Expect public not empty, public heading =
"public".

#### Public with mixed case and extra whitespace

Create a node with "#   PuBLiC" as the public heading. Call
ParseNode. Expect public not empty, public heading =
"public".

#### Node name with varied whitespace

Create a node with "#    ROOT/e" as the name heading. Call
ParseNode. Expect name section heading = "root/e".

#### Subsection headings are normalized

Create a node with subsections "##   Interface" and
"## CONSTRAINTS". Call ParseNode. Expect subsection
headings = "interface" and "constraints".

#### Tab characters in heading whitespace

Create a node with tab characters around the subsection
name. Call ParseNode. Expect subsection heading =
"interface".

### Content extraction

#### Level-3 and deeper headings are content

Create a node with subsections that contain level-3 and
level-4 headings. Call ParseNode.

Expect the deeper headings and their text are included as
raw content within the subsection, not treated as
structural boundaries.

#### Fenced code blocks with heading-like content

Create a node with a fenced code block inside a subsection
that contains lines starting with # and ##. Call ParseNode.

Expect the heading-like lines inside the code block are
treated as content, not as structural headings.

#### Content between sections is trimmed

Create a node with blank lines surrounding content in the
public section and subsections. Call ParseNode.

Expect leading and trailing blank lines are trimmed from
all content strings.

### Failure cases

#### File does not exist

Call ParseNode with a logical name whose file does not
exist. Expect "read error".

#### No frontmatter delimiters

Create a node file with no frontmatter delimiters. Call
ParseNode. Expect no error -- frontmatter is optional.
Name section heading and content parsed correctly.

#### Content before first heading

Create a node file with text before any heading. Call
ParseNode. Expect "unexpected content".

#### Level-2 heading before any level-1 heading

Create a node file with a level-2 heading before the first
level-1 heading. Call ParseNode. Expect "unexpected
content".

#### Node name does not match logical name

Create a node file where the first heading does not match
the logical name. Call ParseNode. Expect "invalid node
name".

#### Node name case mismatch is not an error

Create a node file where the heading uses different casing
than the logical name. Call ParseNode. Expect no error --
normalization makes them equal.

#### Duplicate public section -- same case

Create a node with two "# Public" sections. Call ParseNode.
Expect "duplicate public section".

#### Duplicate public section -- different case

Create a node with "# Public" and "# PUBLIC" sections. Call
ParseNode. Expect "duplicate public section".

#### Duplicate subsection in public -- same case

Create a node with two "## Interface" subsections under
public. Call ParseNode. Expect "duplicate subsection".

#### Duplicate subsection in public -- different case

Create a node with "## Interface" and "## INTERFACE"
subsections under public. Call ParseNode. Expect "duplicate
subsection".

#### Duplicate subsection in public -- whitespace variation

Create a node with "## Interface" and "##   Interface"
subsections under public. Call ParseNode. Expect "duplicate
subsection".

#### First element is a paragraph -- missing node name

Create a node file with a paragraph instead of a heading as
the first element. Call ParseNode. Expect "unexpected
content".

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Describe tests in terms of the functional interface —
  use function names and error names from the interface,
  not language-specific constructs.
- Each test case has: a description, setup (what files to
  create and with what content), actions (what functions
  to call), and expected outcome.
- Do not prescribe how to create test files or assert
  results — those are implementation details for the
  language layer.
