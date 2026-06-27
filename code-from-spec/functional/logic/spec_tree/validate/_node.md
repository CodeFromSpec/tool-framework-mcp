---
depends_on:
  - SPEC/functional/logic/os/file
  - SPEC/functional/logic/os/path_utils
  - SPEC/functional/logic/os/list_files
  - SPEC/functional/logic/utils/logical_names
  - SPEC/functional/logic/utils/text_normalization
  - SPEC/functional/logic/parsing/frontmatter(interface)
  - SPEC/functional/logic/parsing/node_parsing(interface)
  - EXTERNAL/code-from-spec/_rules/CODE_FROM_SPEC.md
output: code-from-spec/functional/logic/spec_tree/validate/output.md
---

# SPEC/functional/logic/spec_tree/validate

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

function SpecTreeValidate(entries: list of SpecTreeValidateInput, all_dirs: list of string) -> list of FormatError
```

Takes the full set of discovered nodes with their parsed
frontmatter and body, plus a list of all subdirectory
paths found under `code-from-spec/`. Returns a list of
format errors (empty if all nodes are valid).

A node has children if any other entry in the input
list has a logical name that starts with its logical
name followed by `/`. For example, given entries
`SPEC/a` and `SPEC/a/b`, `SPEC/a` has children
(because `SPEC/a/b` starts with `SPEC/a/`). `SPEC/a/b`
is a leaf if no entry starts with `SPEC/a/b/`.

# Agent

## Behavior

Build a set of all known logical names for lookup. Add
every entry's `logical_name` to the set. Then, for each
entry that has output, construct the artifact logical
name — strip `SPEC/` from the entry's logical name,
prepend `ARTIFACT/` — and add it to the set. Example:
entry `SPEC/a/b` with output adds `ARTIFACT/a/b` to
the set.

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

The fields `depends_on`, `input`, and `output` are only
permitted on leaf nodes (nodes without children). If a
node with children has any of these fields non-empty,
report one error per field.

#### leaf_only_agent

Rule name: `leaf_only_agent`.

Only leaf nodes may have an `# Agent` section
(`node.agent` is present). If a node with children has
`# Agent`, it is a format error.

#### dependency_targets

Rule name: `dependency_targets`.

Each `depends_on` entry must be valid:

- **SPEC references** (detected by `LogicalNameIsSpec`):
  strip the qualifier using `LogicalNameStripQualifier`
  to get the bare logical name (e.g.
  `SPEC/a/b(interface)` → `SPEC/a/b`). Verify the bare
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

- **`EXTERNAL/` references**: convert to a path using
  `LogicalNameExternalToPath`. Create a `PathCfs` from
  the result. Open the file with `FileOpen`. If it fails
  (file does not exist or not readable), report a format
  error. If it succeeds, call `FileClose` immediately.

- **Unrecognized prefix**: report a format error.

Report one error per invalid entry.

#### input_target

Rule name: `input_target`.

If `frontmatter.input` is non-empty, verify it starts
with `ARTIFACT/` or `EXTERNAL/`. If neither, report a
format error.

For `ARTIFACT/` references: strip any qualifier using
`LogicalNameStripQualifier` (defensively), and verify the
bare reference exists in the known logical names set.
If not, report a format error.

For `EXTERNAL/` references: convert to a path using
`LogicalNameExternalToPath`. Create a `PathCfs` from the
result. Open the file with `FileOpen`. If it fails,
report a format error. If it succeeds, call `FileClose`
immediately.

#### missing_node_md

Rule name: `missing_node_md`.

Check the `all_dirs` list for subdirectories under
`code-from-spec/` that do not have a corresponding
node in the entries. For each directory path in
`all_dirs`:
- Skip directories whose first path segment after
  `code-from-spec/` starts with `_` (these are ignored
  by the framework).
- Skip the `code-from-spec/` directory itself.
- Check whether a node exists whose file path would be
  `<dir>/_node.md`. If no such node exists in the
  entries, report a format error with the directory path
  and detail "subdirectory has no _node.md".

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
- Subdirectories without `_node.md` are reported as
  format errors (except `_`-prefixed dirs under
  `code-from-spec/`).
