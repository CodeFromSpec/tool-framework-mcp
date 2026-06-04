---
depends_on:
  - ROOT/functional/logic/os/file_reader
  - ROOT/functional/logic/os/path_utils
  - ROOT/functional/logic/utils/logical_names
  - ROOT/functional/logic/utils/text_normalization
  - ROOT/functional/logic/parsing/frontmatter(interface)
  - ROOT/functional/logic/parsing/node_parsing(interface)
external:
  - path: CODE_FROM_SPEC.md
output: code-from-spec/functional/logic/spec_tree/validate/output.md
---

# ROOT/functional/logic/spec_tree/validate

Linter for the spec tree. Receives discovered nodes with
their parsed frontmatter and body, checks structural rules
defined by the framework, and reports all violations found.

# Public

## Namespace

    namespace: spectreevalidate

## Interface

```
record SpecTreeValidateInput
  logical_name: string
  frontmatter: frontmatter.Frontmatter
  node: parsenode.Node

record FormatError
  node: string
  rule: string
  detail: string

function SpecTreeValidate(entries: list of SpecTreeValidateInput) -> list of FormatError
```

Takes the full set of discovered nodes with their parsed
frontmatter and body. Returns a list of format errors
(empty if all nodes are valid).

A node has children if any other entry in the input
list has a logical name that starts with its logical
name followed by `/`. For example, given entries
`ROOT/a` and `ROOT/a/b`, `ROOT/a` has children
(because `ROOT/a/b` starts with `ROOT/a/`). `ROOT/a/b`
is a leaf if no entry starts with `ROOT/a/b/`.

# Agent

## Behavior

Build a set of all known logical names for lookup. Add
every entry's `logical_name` to the set. Then, for each
entry that has output, construct the artifact logical
name — strip `ROOT/` from the entry's logical name,
prepend `ARTIFACT/` — and add it to the set. Example:
entry `ROOT/a/b` with output adds `ARTIFACT/a/b` to the
set.

For each entry, determine whether it has children: a
node has children if any other entry's logical name
starts with its logical name followed by `/`. Then run
all validation rules below. Collect all errors — do not
stop at the first.

### Validation rules

#### name_heading

Rule name: `name_heading`.

The first section heading (`node.name_section.heading`)
must match the entry's `logical_name`. Comparison uses
`NormalizeText` on both values.

#### leaf_only_fields

Rule name: `leaf_only_fields`.

The fields `depends_on`, `external`, `input`, and
`output` are only permitted on leaf nodes (nodes without
children). If a node with children has any of these fields
non-empty, report one error per field.

#### leaf_only_agent

Rule name: `leaf_only_agent`.

Only leaf nodes may have an `# Agent` section
(`node.agent` is present). If a node with children has
`# Agent`, it is a format error.

#### dependency_targets

Rule name: `dependency_targets`.

Each `depends_on` entry must be valid:

- **ROOT references** (the entry is exactly `ROOT` or
  starts with `ROOT/`): strip the qualifier using
  `LogicalNameStripQualifier` to get the bare logical
  name (e.g. `ROOT/a/b(interface)` → `ROOT/a/b`; if no
  qualifier, returned unchanged). Verify the bare
  logical name exists in the known logical names set.
  Also verify it does not point to the node itself
  (bare name equals the current node's logical name),
  an ancestor (the bare name followed by `/` is a prefix
  of the current node's logical name), or a descendant
  (the current node's logical name followed by `/` is a
  prefix of the bare name).

- **`ARTIFACT/` references**: strip any qualifier using
  `LogicalNameStripQualifier` (there should not be one,
  but do so defensively), then verify the bare reference
  exists in the known logical names set.

Report one error per invalid entry.

#### input_target

Rule name: `input_target`.

If `frontmatter.input` is non-empty, verify it starts
with `ARTIFACT/`. If not, report a format error. Then
strip any qualifier using `LogicalNameStripQualifier`
(defensively), and verify the bare reference exists in
the known logical names set. If not,
report a format error.

#### external_files

Rule name: `external_files`.

For each `external` entry, create a `PathCfs` with the
entry's `path` as its value (external paths are relative
to the project root).

Open the file with `FileOpen`. If it fails (invalid
path, file does not exist, or not readable), report a
format error and skip to the next external entry. If
it succeeds, call `FileClose` immediately.

#### output_paths

Rule name: `output_paths`.

The output path must pass `PathValidateCfs` (no traversal,
no absolute paths, no backslashes, within project root).

#### public_subsection_required

Rule name: `public_subsection_required`.

If `node.public` is present, all content must be under
a `##` subsection. If `node.public.content` is
non-empty (has any non-blank line), report a format
error with detail: "content in # Public must be under
a ## subsection". If `node.public` is absent, skip.

A line is blank if it contains only whitespace or is
empty.

#### duplicate_subsections

Rule name: `duplicate_subsections`.

If `node.public` is present and has subsections, all
`##` subsection headings must be unique after
`NormalizeText`. If `node.public` is absent, skip. Track
seen headings; if a heading has already been seen,
report a format error for that occurrence. The first
occurrence is not an error — only subsequent repeats
are reported.

## Contracts

- All nodes are validated — not just leaf nodes.
- All errors are collected — validation does not stop at
  the first error.
- Each FormatError includes the rule name, the node's
  logical name, and a detail message.
