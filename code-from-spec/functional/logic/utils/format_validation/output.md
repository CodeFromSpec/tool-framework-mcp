<!-- code-from-spec: ROOT/functional/utils/format_validation@dRO4R1sCCFLDMcTn6iobcnew1t0 -->

# Format Validation

## Records

```
record FormatError
  node:   string   -- logical name of the node where the error was found
  rule:   string   -- name of the rule that was violated
  detail: string   -- human-readable description of the specific violation
```

## Functions

---

### ValidateFormat(discovered_nodes) -> list of FormatError

`discovered_nodes` is a list of records, each containing:
- `logical_name`: string
- `file_path`: string

Returns a (possibly empty) list of FormatError. All errors from
all nodes are collected before returning — validation never stops
at the first error.

Steps:

1. Initialize `errors` as an empty list.

2. For each `node` in `discovered_nodes`:

   2a. Open the file at `node.file_path` using `OpenFileReader`.
       If the file cannot be opened, add a FormatError:
         node:   <node.logical_name>
         rule:   "unreadable node"
         detail: "cannot open file at <node.file_path>"
       Skip remaining steps for this node and continue to the
       next.

   2b. Parse the frontmatter from the file using `ParseFrontmatter`.
       If parsing fails (malformed YAML or unreadable), add a
       FormatError:
         node:   <node.logical_name>
         rule:   "unreadable node"
         detail: "frontmatter parse error: <error message>"
       Close the reader. Skip remaining steps for this node.

   2c. Parse the body using `ParseNode`.
       If parsing fails (unexpected content, duplicate sections,
       etc.), add a FormatError:
         node:   <node.logical_name>
         rule:   "unreadable node"
         detail: "body parse error: <error message>"
       Close the reader. Skip remaining steps for this node.

   2d. Close the reader.

   2e. Determine whether this node is a leaf or intermediate.
       A node is intermediate if any other node in
       `discovered_nodes` has a logical name that starts with
       `node.logical_name` followed by "/".
       Otherwise it is a leaf.

   2f. Run all validation rules below against this node,
       collecting any errors into `errors`.

3. Return `errors`.

---

### Rule: name_verification

Applied to: all nodes.

1. Derive the expected logical name from `node.file_path`
   using `ReverseResolve` from `logical_names`.
   If `ReverseResolve` returns an error, add a FormatError:
     node:   <node.logical_name>
     rule:   "name_verification"
     detail: "cannot derive logical name from path: <error message>"
   Stop this rule for the node.

2. Obtain the actual first heading from `parsed_node.name_section.heading`.

3. Normalize both the derived logical name and the first heading
   using `NormalizeName` from `name_normalization`.

4. If the normalized values do not match, add a FormatError:
     node:   <node.logical_name>
     rule:   "name_verification"
     detail: "heading \"<actual heading>\" does not match expected
              logical name \"<expected logical name>\""

---

### Rule: frontmatter_field_restrictions

Applied to: intermediate nodes only.

1. If the frontmatter has a non-empty `depends_on` list, add a
   FormatError:
     node:   <node.logical_name>
     rule:   "frontmatter_field_restrictions"
     detail: "intermediate node must not have \"depends_on\""

2. If the frontmatter has a non-empty `external` list, add a
   FormatError:
     node:   <node.logical_name>
     rule:   "frontmatter_field_restrictions"
     detail: "intermediate node must not have \"external\""

3. If the frontmatter has a non-empty `input`, add a FormatError:
     node:   <node.logical_name>
     rule:   "frontmatter_field_restrictions"
     detail: "intermediate node must not have \"input\""

4. If the frontmatter has a non-empty `outputs` list, add a
   FormatError:
     node:   <node.logical_name>
     rule:   "frontmatter_field_restrictions"
     detail: "intermediate node must not have \"outputs\""

---

### Rule: agent_section_restrictions

Applied to: intermediate nodes only.

1. If `parsed_node.agent` is present (not empty), add a
   FormatError:
     node:   <node.logical_name>
     rule:   "agent_section_restrictions"
     detail: "intermediate node must not have a \"# Agent\" section"

---

### Rule: dependency_targets

Applied to: leaf nodes only (nodes that declare `depends_on`).

For each `dep` in `frontmatter.depends_on`:

1. Resolve `dep` to a file path using `ResolvePath` (for ROOT/
   references) or `ResolveArtifactReference` (for ARTIFACT/
   references) from `logical_names`.
   If resolution fails, add a FormatError:
     node:   <node.logical_name>
     rule:   "dependency_targets"
     detail: "cannot resolve depends_on entry \"<dep>\": <error message>"
   Continue to the next entry.

2. For ROOT/ references: verify that the resolved `_node.md` file
   exists in `discovered_nodes` (i.e., a node with that file path
   is present).
   If not found, add a FormatError:
     node:   <node.logical_name>
     rule:   "dependency_targets"
     detail: "depends_on target \"<dep>\" does not exist"
   Continue to the next entry.

3. For ROOT/ references: check ancestor relationship.
   Strip any qualifier from `dep` to get the bare logical name.
   If `node.logical_name` starts with `<dep>/` (i.e., `dep` is
   an ancestor of this node), add a FormatError:
     node:   <node.logical_name>
     rule:   "dependency_targets"
     detail: "depends_on \"<dep>\" points to an ancestor (already inherited)"

4. For ROOT/ references: check descendant relationship.
   If `<dep>` starts with `node.logical_name + "/"` (i.e., `dep`
   is a descendant of this node), add a FormatError:
     node:   <node.logical_name>
     rule:   "dependency_targets"
     detail: "depends_on \"<dep>\" points to a descendant
              (creates circular dependency)"

---

### Rule: external_file_existence

Applied to: leaf nodes only (nodes that declare `external`).

For each `ext` in `frontmatter.external`:

1. Validate `ext.path` using `ValidatePath` from `path_validation`.
   If validation fails, add a FormatError:
     node:   <node.logical_name>
     rule:   "external_file_existence"
     detail: "invalid path \"<ext.path>\": <error message>"
   Continue to the next entry.

2. Attempt to open the file at `ext.path` using `OpenFileReader`.
   If it cannot be opened, add a FormatError:
     node:   <node.logical_name>
     rule:   "external_file_existence"
     detail: "external file not found: \"<ext.path>\""
   Continue to the next entry.

3. If `ext.fragments` is empty or absent, close the reader and
   continue to the next entry.

4. If `ext.fragments` is declared, for each `fragment` in
   `ext.fragments`:

   a. Read all lines from the file (or reopen it for each
      fragment to avoid reader state issues). Use `SkipLines` to
      skip to the start of the declared range, then read the
      lines in the range.
      The `fragment.lines` field is a range in the form
      "<start>-<end>" (1-based, inclusive).

   b. Compute SHA-1 of the extracted content (CRLF normalized to
      LF), then encode as base64url (no padding, 27 characters).

   c. If the computed hash does not equal `fragment.hash`, add a
      FormatError:
        node:   <node.logical_name>
        rule:   "external_file_existence"
        detail: "fragment hash mismatch for \"<ext.path>\"
                 lines <fragment.lines>: expected \"<fragment.hash>\",
                 got \"<computed_hash>\""

5. Close the reader.

---

### Rule: output_path_validation

Applied to: leaf nodes only (nodes that declare `outputs`).

For each `out` in `frontmatter.outputs`:

1. Call `ValidatePath(out.path, project_root)` from
   `path_validation`.
   If validation fails, add a FormatError:
     node:   <node.logical_name>
     rule:   "output_path_validation"
     detail: "invalid output path \"<out.path>\": <error message>"

---

### Rule: duplicate_public_subsections

Applied to: all nodes that have a `# Public` section.

1. If `parsed_node.public` is absent, skip this rule.

2. Initialize `seen_headings` as an empty list.

3. For each `subsection` in `parsed_node.public.subsections`:

   a. Normalize `subsection.heading` using `NormalizeName`.

   b. If the normalized heading already exists in `seen_headings`,
      add a FormatError:
        node:   <node.logical_name>
        rule:   "duplicate_public_subsections"
        detail: "duplicate \"##\" heading in \"# Public\":
                 \"<subsection.heading>\" (normalized: \"<normalized>\")"

   c. Otherwise, append the normalized heading to `seen_headings`.
```

## Contracts

- Every discovered node is validated regardless of whether it is a
  leaf or intermediate.
- All errors for all nodes are collected before returning. Validation
  never stops at the first error in a node or across nodes.
- "Unreadable node" is the only error that causes remaining rules to
  be skipped for that node.
