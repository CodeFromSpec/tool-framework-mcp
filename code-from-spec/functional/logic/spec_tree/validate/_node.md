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
outputs:
  - id: format_validation
    path: code-from-spec/functional/logic/spec_tree/validate/output.md
---

# ROOT/functional/logic/spec_tree/validate

Linter for the spec tree. Receives discovered nodes with
their parsed frontmatter and body, checks structural rules
defined by the framework, and reports all violations found.

# Public

## Interface

```
record SpecTreeValidateInput
  logical_name: string
  frontmatter: Frontmatter
  node: Node

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
entry that has outputs, construct the artifact logical
name for each output — strip `ROOT/` from the entry's
logical name, prepend `ARTIFACT/`, append `(id)` — and
add it to the set. Example: entry `ROOT/a/b` with
output id `foo` adds `ARTIFACT/a/b(foo)` to the set.

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
`outputs` are only permitted on leaf nodes (nodes without
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

- **`ROOT/` references**: strip the qualifier using
  `LogicalNameStripQualifier` to get the bare logical
  name (e.g. `ROOT/a/b(interface)` → `ROOT/a/b`; if no
  qualifier, returned unchanged). Verify the bare
  logical name exists in the known logical names set. Also verify it does not point to the node itself
  (bare name equals the current node's logical name),
  an ancestor (the bare name followed by `/` is a prefix
  of the current node's logical name), or a descendant
  (the current node's logical name followed by `/` is a
  prefix of the bare name).

- **`ARTIFACT/` references**: verify the reference exists
  in the known logical names set.

Report one error per invalid entry.

#### input_target

Rule name: `input_target`.

If `frontmatter.input` is non-empty, verify it starts
with `ARTIFACT/`. If not, report a format error. Then
verify it exists in the known logical names set. If not,
report a format error.

#### external_files

Rule name: `external_files`.

For each `external` entry, create a `PathCfs` with the
entry's `path` as its value (external paths are relative
to the project root).

**Step 1 — Verify existence.** Open the file with
`FileOpen`. If it fails (invalid path, file does not
exist, or not readable), report a format error and skip
to the next external entry. If it succeeds, call
`FileClose` immediately.

**Step 2 — Verify fragments.** If `fragments` are
declared, process each fragment independently:
- Parse the `lines` field as `start-end` (both 1-based,
  inclusive). Example: `"150-210"` means lines 150
  through 210. If the format is invalid, `start < 1`,
  or `start > end`, report a format error and skip this
  fragment.
- Open the file again with `FileOpen`. If it fails,
  report a format error and skip this fragment.
- Use `FileSkipLines` to skip `start - 1` lines, then
  read `end - start + 1` lines with `FileReadLine`. If
  `FileReadLine` returns "end of file" before all lines
  are read, call `FileClose`, report a format error
  (fragment out of range), and skip this fragment.
- Call `FileClose`.
- Append `\n` (LF) after each read line, including the
  last, to form the content. `FileReadLine` already
  normalizes CRLF, so the result is platform-independent.
- Compute SHA-1 of the joined content.
- Encode the 20-byte SHA-1 digest as base64url (RFC 4648
  §5, no padding) — 27 characters.
- Compare with the declared `hash`. If mismatch, report
  a format error.

#### output_paths

Rule name: `output_paths`.

Each `outputs` entry's `path` must pass `PathValidateCfs`
(no traversal, no absolute paths, no backslashes, within
project root).

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
