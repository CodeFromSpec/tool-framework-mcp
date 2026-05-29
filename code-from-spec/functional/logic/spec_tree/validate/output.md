<!-- code-from-spec: ROOT/functional/logic/spec_tree/validate@0tXQaFHwa74i2m2o3YWI4I1ROTU -->

# SpecTreeValidate

## Types

```
record SpecTreeValidateInput
  logical_name: string
  frontmatter: Frontmatter
  node: Node

record FormatError
  node: string
  rule: string
  detail: string
```

---

## function SpecTreeValidate(entries: list of SpecTreeValidateInput) -> list of FormatError

Takes the full set of discovered nodes with their parsed frontmatter and body.
Returns a list of format errors (empty if all nodes are valid).
All nodes are validated. All errors are collected — validation does not stop at the first error.

### Step 1 — Build the known logical names set

1. Create an empty set called `known_names`.

2. For each entry in `entries`:
   - Add `entry.logical_name` to `known_names`.
   - For each output in `entry.frontmatter.outputs`:
     - Strip the `ROOT/` prefix from `entry.logical_name` to get the bare path.
     - Construct the artifact name: prepend `"ARTIFACT/"`, append `"(" + output.id + ")"`.
     - Example: entry `ROOT/a/b` with output id `foo` → `"ARTIFACT/a/b(foo)"`.
     - Add the constructed artifact name to `known_names`.

### Step 2 — For each entry, determine if it has children

For each entry in `entries`:
- A node has children if any other entry's `logical_name` starts with
  `entry.logical_name + "/"`.
- Record this as a boolean `has_children` associated with the entry.

### Step 3 — Run all validation rules for each entry

For each entry in `entries`:
- Initialize an empty list `errors` to collect errors for this entry.
- Run each rule below. Append any errors found to `errors`.
- Do not stop at the first error — run all rules for all entries.

After processing all entries, return the combined flat list of all errors.

---

### Rule: name_heading

Rule name: `"name_heading"`.

1. Apply `NormalizeText` to `entry.node.name_section.heading` → `normalized_heading`.
2. Apply `NormalizeText` to `entry.logical_name` → `normalized_name`.
3. If `normalized_heading` does not equal `normalized_name`:
   - Append a FormatError:
     - node: `entry.logical_name`
     - rule: `"name_heading"`
     - detail: `"first heading \"<normalized_heading>\" does not match logical name \"<normalized_name>\""`

---

### Rule: leaf_only_fields

Rule name: `"leaf_only_fields"`.

Only applies when `has_children` is true.

1. If `has_children` is false, skip this rule entirely.

2. If `entry.frontmatter.depends_on` is non-empty:
   - Append a FormatError:
     - node: `entry.logical_name`
     - rule: `"leaf_only_fields"`
     - detail: `"field \"depends_on\" is only permitted on leaf nodes"`

3. If `entry.frontmatter.external` is non-empty:
   - Append a FormatError:
     - node: `entry.logical_name`
     - rule: `"leaf_only_fields"`
     - detail: `"field \"external\" is only permitted on leaf nodes"`

4. If `entry.frontmatter.input` is non-empty:
   - Append a FormatError:
     - node: `entry.logical_name`
     - rule: `"leaf_only_fields"`
     - detail: `"field \"input\" is only permitted on leaf nodes"`

5. If `entry.frontmatter.outputs` is non-empty:
   - Append a FormatError:
     - node: `entry.logical_name`
     - rule: `"leaf_only_fields"`
     - detail: `"field \"outputs\" is only permitted on leaf nodes"`

---

### Rule: leaf_only_agent

Rule name: `"leaf_only_agent"`.

1. If `has_children` is false, skip this rule entirely.

2. If `entry.node.agent` is present:
   - Append a FormatError:
     - node: `entry.logical_name`
     - rule: `"leaf_only_agent"`
     - detail: `"\"# Agent\" section is only permitted on leaf nodes"`

---

### Rule: dependency_targets

Rule name: `"dependency_targets"`.

For each `dep` in `entry.frontmatter.depends_on`:

  **If `dep` starts with `"ROOT/":`**

  1. Call `LogicalNameStripQualifier(dep)` → `bare_name`.
  2. Check if `bare_name` exists in `known_names`.
     If not:
     - Append a FormatError:
       - node: `entry.logical_name`
       - rule: `"dependency_targets"`
       - detail: `"depends_on target \"<bare_name>\" does not exist"`
     - Skip remaining checks for this entry.
  3. Check if `bare_name` equals `entry.logical_name` (self-reference).
     If so:
     - Append a FormatError:
       - node: `entry.logical_name`
       - rule: `"dependency_targets"`
       - detail: `"depends_on \"<bare_name>\" points to the node itself"`
     - Skip remaining checks for this entry.
  4. Check if `bare_name + "/"` is a prefix of `entry.logical_name` (ancestor reference).
     If so:
     - Append a FormatError:
       - node: `entry.logical_name`
       - rule: `"dependency_targets"`
       - detail: `"depends_on \"<bare_name>\" points to an ancestor node"`
     - Skip remaining checks for this entry.
  5. Check if `entry.logical_name + "/"` is a prefix of `bare_name` (descendant reference).
     If so:
     - Append a FormatError:
       - node: `entry.logical_name`
       - rule: `"dependency_targets"`
       - detail: `"depends_on \"<bare_name>\" points to a descendant node"`

  **If `dep` starts with `"ARTIFACT/":`**

  1. Check if `dep` exists in `known_names`.
     If not:
     - Append a FormatError:
       - node: `entry.logical_name`
       - rule: `"dependency_targets"`
       - detail: `"depends_on artifact target \"<dep>\" does not exist"`

  **Otherwise (unrecognized prefix):**

  - Append a FormatError:
    - node: `entry.logical_name`
    - rule: `"dependency_targets"`
    - detail: `"depends_on \"<dep>\" has unrecognized prefix (expected ROOT/ or ARTIFACT/)"`

---

### Rule: input_target

Rule name: `"input_target"`.

1. If `entry.frontmatter.input` is empty, skip this rule entirely.

2. If `entry.frontmatter.input` does not start with `"ARTIFACT/"`:
   - Append a FormatError:
     - node: `entry.logical_name`
     - rule: `"input_target"`
     - detail: `"input \"<entry.frontmatter.input>\" must be an ARTIFACT/ reference"`
   - Stop checking this rule (the existence check is not meaningful without a valid prefix).

3. Check if `entry.frontmatter.input` exists in `known_names`.
   If not:
   - Append a FormatError:
     - node: `entry.logical_name`
     - rule: `"input_target"`
     - detail: `"input artifact \"<entry.frontmatter.input>\" does not exist"`

---

### Rule: external_files

Rule name: `"external_files"`.

For each `ext` in `entry.frontmatter.external`:

  **Step 1 — Verify existence.**

  1. Create a `PathCfs` with `ext.path` as its value.
  2. Call `FileOpen(path_cfs)`.
     If it fails (invalid path, file does not exist, or not readable):
     - Append a FormatError:
       - node: `entry.logical_name`
       - rule: `"external_files"`
       - detail: `"external file \"<ext.path>\" could not be opened: <error>"`
     - Skip to the next external entry (do not process fragments for this entry).
  3. Call `FileClose` immediately.

  **Step 2 — Verify fragments (only if `ext.fragments` is present and non-empty).**

  For each `fragment` in `ext.fragments`:

  a. Parse `fragment.lines` as `"<start>-<end>"` (both integers, 1-based, inclusive).
     If the format is invalid, or `start < 1`, or `start > end`:
     - Append a FormatError:
       - node: `entry.logical_name`
       - rule: `"external_files"`
       - detail: `"external file \"<ext.path>\" fragment has invalid lines range \"<fragment.lines>\""`
     - Skip to the next fragment.

  b. Call `FileOpen(path_cfs)`.
     If it fails:
     - Append a FormatError:
       - node: `entry.logical_name`
       - rule: `"external_files"`
       - detail: `"external file \"<ext.path>\" could not be re-opened for fragment verification"`
     - Skip to the next fragment.

  c. Call `FileSkipLines(reader, start - 1)` to skip to the target range.

  d. Read `end - start + 1` lines using `FileReadLine`.
     - After each successful read, append `"\n"` (LF) to form the line content.
     - If `FileReadLine` raises "end of file" before all lines are read:
       - Call `FileClose(reader)`.
       - Append a FormatError:
         - node: `entry.logical_name`
         - rule: `"external_files"`
         - detail: `"external file \"<ext.path>\" fragment lines <fragment.lines> is out of range"`
       - Skip to the next fragment.

  e. Call `FileClose(reader)`.

  f. Join all lines (each already suffixed with `"\n"`) to form `content`.

  g. Compute SHA-1 of `content` (UTF-8 bytes).

  h. Encode the 20-byte SHA-1 digest as base64url (RFC 4648 §5, no padding) → 27-character string `computed_hash`.

  i. If `computed_hash` does not equal `fragment.hash`:
     - Append a FormatError:
       - node: `entry.logical_name`
       - rule: `"external_files"`
       - detail: `"external file \"<ext.path>\" fragment lines <fragment.lines> hash mismatch: expected \"<fragment.hash>\", got \"<computed_hash>\""`

---

### Rule: output_paths

Rule name: `"output_paths"`.

For each `output` in `entry.frontmatter.outputs`:

  1. Call `PathValidateCfs(output.path)`.
     If it raises an error:
     - Append a FormatError:
       - node: `entry.logical_name`
       - rule: `"output_paths"`
       - detail: `"output path \"<output.path>\" is invalid: <error>"`

---

### Rule: duplicate_subsections

Rule name: `"duplicate_subsections"`.

1. If `entry.node.public` is absent, skip this rule entirely.

2. If `entry.node.public.subsections` is empty, skip this rule entirely.

3. Create an empty set `seen_headings`.

4. For each `subsection` in `entry.node.public.subsections`:
   - Apply `NormalizeText` to `subsection.heading` → `normalized`.
   - If `normalized` is already in `seen_headings`:
     - Append a FormatError:
       - node: `entry.logical_name`
       - rule: `"duplicate_subsections"`
       - detail: `"duplicate subsection heading \"<normalized>\" in # Public"`
   - Else:
     - Add `normalized` to `seen_headings`.
