<!-- code-from-spec: ROOT/functional/utils/format_validation@uxv3UKsmqN0cbX6TW_NVjQp0OKc -->

# ValidateFormat

## Records

```
record FormatError
  node:   string   -- logical name of the node that failed validation
  rule:   string   -- short identifier for the rule that was violated
  detail: string   -- human-readable explanation of the violation
```

---

## Functions

### ValidateFormat(discovered_nodes) -> list of FormatError

`discovered_nodes` is a list of records, each containing:
- `logical_name`: string
- `file_path`: string

Returns a list of FormatError records. Returns an empty list when all
nodes are valid. All errors across all nodes are collected before
returning — validation does not stop at the first error.

Errors:
- "unreadable node": a node file cannot be opened or read.

**Steps:**

1. Initialize `all_errors` as an empty list.

2. Build a set `all_logical_names` containing the `logical_name` of
   every entry in `discovered_nodes`.
   This set is used to classify nodes as leaf or intermediate, and
   to verify dependency targets.

3. For each `node` in `discovered_nodes`:

   3a. Determine whether `node` is an **intermediate node** (has
       children) or a **leaf node** (has no children).
       A node is intermediate if any other entry in `all_logical_names`
       starts with the node's `logical_name` followed by "/".
       Otherwise the node is a leaf.

   3b. Open the file at `node.file_path` using `file_reader`.
       If the file cannot be read, append a FormatError:
         node:   <node.logical_name>
         rule:   "unreadable node"
         detail: "file at <node.file_path> cannot be opened"
       Skip all remaining checks for this node and continue to the
       next node.

   3c. Parse frontmatter from the file using `frontmatter`.
       Store result as `fm` (fields: `depends_on`, `external`,
       `input`, `outputs`).

   3d. Parse the node body using `node_parsing`.
       Store result as `parsed` (fields: `name_section`,
       `public`, `agent`, `private`).

   3e. **Rule: name verification.**
       Use `logical_names` reverse resolution on `node.file_path`
       to obtain the expected logical name `expected_name`.
       Normalize both `parsed.name_section.heading` and
       `expected_name` using `name_normalization`.
       If the normalized values differ, append a FormatError:
         node:   <node.logical_name>
         rule:   "name mismatch"
         detail: "heading \"<parsed.name_section.heading>\" does not
                  match expected logical name \"<expected_name>\""

   3f. **Rule: frontmatter field restrictions (intermediate nodes
       only).**
       If the node is intermediate:

       If `fm.depends_on` is not empty, append a FormatError:
         node:   <node.logical_name>
         rule:   "depends_on on intermediate node"
         detail: "field depends_on is not permitted on nodes with children"

       If `fm.external` is not empty, append a FormatError:
         node:   <node.logical_name>
         rule:   "external on intermediate node"
         detail: "field external is not permitted on nodes with children"

       If `fm.input` is not empty, append a FormatError:
         node:   <node.logical_name>
         rule:   "input on intermediate node"
         detail: "field input is not permitted on nodes with children"

       If `fm.outputs` is not empty, append a FormatError:
         node:   <node.logical_name>
         rule:   "outputs on intermediate node"
         detail: "field outputs is not permitted on nodes with children"

   3g. **Rule: agent section restriction (intermediate nodes only).**
       If the node is intermediate and `parsed.agent` is present,
       append a FormatError:
         node:   <node.logical_name>
         rule:   "agent section on intermediate node"
         detail: "# Agent section is not permitted on nodes with children"

   3h. **Rule: dependency targets.**
       For each `dep` in `fm.depends_on`:

       i. Resolve the logical name in `dep` to a file path using
          `logical_names`.
          If `dep` starts with "ARTIFACT/", use
          `ResolveArtifactReference` to extract the node path, then
          resolve that to a file path using `ResolvePath`.
          If `dep` starts with "ROOT/", use `ResolvePath` directly.
          If the resolved file does not exist in `all_logical_names`
          (i.e., there is no discovered node whose `file_path` matches),
          append a FormatError:
            node:   <node.logical_name>
            rule:   "missing dependency target"
            detail: "depends_on entry \"<dep>\" does not resolve to a
                     known node"

       ii. Strip any qualifier from the target's ROOT logical name to
           get `target_bare`.
           If `target_bare` is a prefix of `node.logical_name`
           (i.e., `node.logical_name` starts with `target_bare`
           followed by "/" or equals `target_bare`), the target is an
           ancestor. Append a FormatError:
             node:   <node.logical_name>
             rule:   "depends_on ancestor"
             detail: "depends_on entry \"<dep>\" points to an ancestor;
                      ancestor content is already inherited"

       iii. If `node.logical_name` is a prefix of `target_bare`
            (i.e., `target_bare` starts with `node.logical_name`
            followed by "/" or equals `node.logical_name`), the target
            is a descendant. Append a FormatError:
              node:   <node.logical_name>
              rule:   "depends_on descendant"
              detail: "depends_on entry \"<dep>\" points to a descendant;
                       this would create a circular dependency"

   3i. **Rule: external file existence and fragment hash verification.**
       For each `ext` in `fm.external`:

       i. Check whether the file at `ext.path` exists and can be
          opened using `file_reader`.
          If the file does not exist or cannot be read, append a
          FormatError:
            node:   <node.logical_name>
            rule:   "missing external file"
            detail: "external path \"<ext.path>\" does not exist or
                     cannot be read"
          Skip fragment checks for this entry and continue to the next.

       ii. If `ext.fragments` is not empty, for each `frag` in
           `ext.fragments`:

           - Open the file at `ext.path` using `file_reader`.
           - Parse `frag.lines` to obtain `start_line` and `end_line`
             (the range is inclusive on both ends).
           - Read lines from `start_line` to `end_line` (1-based).
             Collect them as `extracted_content`.
           - Normalize CRLF to LF in `extracted_content`.
           - Compute the SHA-1 digest of `extracted_content`, encode
             as base64url (RFC 4648 §5, no padding, 27 characters).
           - If the computed hash does not equal `frag.hash`, append
             a FormatError:
               node:   <node.logical_name>
               rule:   "fragment hash mismatch"
               detail: "fragment at lines <frag.lines> of
                        \"<ext.path>\" has hash <computed_hash> but
                        declared hash is <frag.hash>"

   3j. **Rule: output path validation.**
       For each `out` in `fm.outputs`:
       Call `ValidatePath(out.path, project_root)`.
       If `ValidatePath` raises an error, append a FormatError:
         node:   <node.logical_name>
         rule:   "invalid output path"
         detail: "<the error message returned by ValidatePath>"

   3k. **Rule: duplicate public subsections.**
       If `parsed.public` is present:
       Collect the headings of all `##` subsections within
       `parsed.public.subsections`.
       Normalize each heading using `name_normalization`.
       Build a list of seen normalized headings. For each normalized
       heading, if it has already been seen, append a FormatError:
         node:   <node.logical_name>
         rule:   "duplicate public subsection"
         detail: "subsection heading \"<original_heading>\" in
                  # Public normalizes to \"<normalized>\" which
                  conflicts with a previous subsection"

4. Return `all_errors`.
