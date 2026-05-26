<!-- code-from-spec: ROOT/functional/utils/format_validation@iV-EZdiHc5Y2OTWBjSlJUHO11Tk -->

# format_validation

Validates the format of all discovered spec nodes. Reads and parses
each node file, applies all structural rules, and returns a list of
errors. All nodes are validated and all errors are collected —
validation never stops at the first error.


## Records

```
record FormatError
  node:   string   — logical name of the node where the error was found
  rule:   string   — name of the rule that was violated
  detail: string   — human-readable description of the specific violation
```


## Functions

---

### ValidateFormat

```
function ValidateFormat(discovered_nodes) -> list of FormatError
  errors:
    - unreadable node: a node file cannot be read or parsed.
```

`discovered_nodes` is a list of records, each containing:
- `logical_name`: the node's logical name (e.g. `ROOT/x/y`)
- `file_path`: absolute or project-relative path to the node's `_node.md` file

Returns a list of `FormatError` records. Returns an empty list when
all nodes are valid.

**Step-by-step logic:**

1. Initialize `errors` as an empty list.

2. Build a set of all known logical names from `discovered_nodes`.
   This set is used in later steps to check whether a node has children
   and whether dependency targets exist.

3. For each node in `discovered_nodes`:

   a. Open the node file using `OpenFileReader(file_path)`.
      If the file cannot be opened, append a `FormatError`:
        node   = logical_name
        rule   = "unreadable node"
        detail = "cannot open file at <file_path>"
      Skip all further checks for this node and continue to the next.

   b. Parse frontmatter from the file using `ParseFrontmatter(file_path)`.
      If parsing fails (malformed YAML), append a `FormatError`:
        node   = logical_name
        rule   = "unreadable node"
        detail = "malformed frontmatter: <error message>"
      Skip all further checks for this node and continue to the next.

   c. Parse the body using `ParseNode(logical_name)`.
      If parsing fails, append a `FormatError`:
        node   = logical_name
        rule   = "unreadable node"
        detail = "body parse error: <error message>"
      Skip all further checks for this node and continue to the next.

   d. Determine whether this node has children:
      A node has children if any other logical name in the known set
      starts with `<logical_name> + "/"`.
      Set `is_leaf` = true if no children exist, false otherwise.

   e. Run all validation rules (steps 4–10) for this node.
      Append every error found to `errors`. Do not stop early.

4. Return `errors`.


## Validation Rules

The rules below (steps e.1 through e.7) are applied for every node
inside step 3e above. All errors are collected; none causes an early
exit for the current node.

---

### Rule 1 — Name Verification

Verify that the first `#` heading in the parsed body matches the
node's logical name.

1. Take the `heading` from `parsed_node.name_section`.
2. Compute `expected` = `NormalizeName(logical_name)`.
3. Compute `actual`   = `NormalizeName(heading)`.
4. If `actual` != `expected`, append a `FormatError`:
     node   = logical_name
     rule   = "name verification"
     detail = "first heading \"<heading>\" does not match logical name \"<logical_name>\""

---

### Rule 2 — Frontmatter Field Restrictions

The fields `depends_on`, `external`, `input`, and `outputs` are only
permitted on leaf nodes.

1. If `is_leaf` is false:
   a. If `frontmatter.depends_on` is non-empty, append a `FormatError`:
        node   = logical_name
        rule   = "frontmatter field restrictions"
        detail = "field \"depends_on\" is not permitted on intermediate nodes"
   b. If `frontmatter.external` is non-empty, append a `FormatError`:
        node   = logical_name
        rule   = "frontmatter field restrictions"
        detail = "field \"external\" is not permitted on intermediate nodes"
   c. If `frontmatter.input` is non-empty, append a `FormatError`:
        node   = logical_name
        rule   = "frontmatter field restrictions"
        detail = "field \"input\" is not permitted on intermediate nodes"
   d. If `frontmatter.outputs` is non-empty, append a `FormatError`:
        node   = logical_name
        rule   = "frontmatter field restrictions"
        detail = "field \"outputs\" is not permitted on intermediate nodes"

---

### Rule 3 — Agent Section Restrictions

Only leaf nodes may have a `# Agent` section.

1. If `is_leaf` is false and `parsed_node.agent` is present:
   Append a `FormatError`:
     node   = logical_name
     rule   = "agent section restrictions"
     detail = "\"# Agent\" section is not permitted on intermediate nodes"

---

### Rule 4 — Dependency Targets

For each entry in `frontmatter.depends_on`:

1. Resolve the entry to a file path using `ResolvePath` (for `ROOT/`
   entries) or `ResolveArtifactReference` (for `ARTIFACT/` entries).
   If resolution raises an error, append a `FormatError`:
     node   = logical_name
     rule   = "dependency targets"
     detail = "cannot resolve dependency \"<entry>\": <error message>"
   Continue to the next entry.

2. For `ROOT/` entries only — check whether the resolved `_node.md`
   file exists among the discovered nodes (by file path or logical name).
   If it does not exist, append a `FormatError`:
     node   = logical_name
     rule   = "dependency targets"
     detail = "dependency \"<entry>\" points to a non-existent node"
   Continue to the next entry.

3. Check that the entry does not point to an ancestor of the current node.
   A node `A` is an ancestor of `B` if `B`'s logical name starts with
   `A`'s logical name followed by `"/"`.
   If the dependency target is an ancestor, append a `FormatError`:
     node   = logical_name
     rule   = "dependency targets"
     detail = "dependency \"<entry>\" points to an ancestor (content already inherited)"

4. Check that the entry does not point to a descendant of the current node.
   A node `D` is a descendant of `C` if `D`'s logical name starts with
   `C`'s logical name followed by `"/"`.
   If the dependency target is a descendant, append a `FormatError`:
     node   = logical_name
     rule   = "dependency targets"
     detail = "dependency \"<entry>\" points to a descendant (circular dependency)"

---

### Rule 5 — External File Existence and Fragment Hashes

For each entry in `frontmatter.external`:

1. Check that the file at `entry.path` exists.
   If it does not exist, append a `FormatError`:
     node   = logical_name
     rule   = "external file existence"
     detail = "external file \"<entry.path>\" does not exist"
   Continue to the next external entry (skip fragment checks).

2. If `entry.fragments` is present and non-empty:
   For each fragment in `entry.fragments`:
   a. Open the file using `OpenFileReader(entry.path)`.
   b. Parse `fragment.lines` as a range `<start>-<end>` (1-based, inclusive).
   c. Read lines from the file. Skip to line `start`, then collect
      lines `start` through `end` inclusive.
   d. Join the collected lines with LF (`"\n"`) as the separator.
      Normalize CRLF to LF before joining.
   e. Compute the SHA-1 digest of the joined content, encoded as
      base64url (RFC 4648 §5, no padding, 27 characters).
   f. If the computed hash does not equal `fragment.hash`, append a `FormatError`:
        node   = logical_name
        rule   = "external file existence"
        detail = "fragment hash mismatch for \"<entry.path>\" lines <fragment.lines>: expected \"<fragment.hash>\", got \"<computed_hash>\""

---

### Rule 6 — Output Path Validation

For each entry in `frontmatter.outputs`:

1. Call `ValidatePath(entry.path, project_root)`.
   If it raises an error, append a `FormatError`:
     node   = logical_name
     rule   = "output path validation"
     detail = "invalid output path \"<entry.path>\": <error message>"

---

### Rule 7 — Duplicate Public Subsections

Within the `# Public` section, all `##` subsection headings must be
unique after normalization.

1. If `parsed_node.public` is absent, skip this rule.

2. Initialize `seen_headings` as an empty list of normalized strings.

3. For each subsection in `parsed_node.public.subsections`:
   a. Compute `normalized` = `NormalizeName(subsection.heading)`.
   b. If `normalized` is already in `seen_headings`, append a `FormatError`:
        node   = logical_name
        rule   = "duplicate public subsections"
        detail = "subsection heading \"<subsection.heading>\" is a duplicate after normalization"
   c. Otherwise, add `normalized` to `seen_headings`.


## Contracts

- Every node in `discovered_nodes` is validated — not just leaf nodes.
- Every violated rule produces an error entry — validation never stops
  at the first error for any given node.
- The returned list may be empty (all nodes valid) or contain many
  entries (one per violation across all nodes).
- Nodes that cannot be read or parsed are reported with rule
  `"unreadable node"` and are skipped for all other rule checks.
