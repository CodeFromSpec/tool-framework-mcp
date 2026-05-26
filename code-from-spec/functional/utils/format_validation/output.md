<!-- code-from-spec: ROOT/functional/utils/format_validation@iV-EZdiHc5Y2OTWBjSlJUHO11Tk -->

# ValidateFormat

## Records

```
record FormatError
  node:   string   -- logical name of the node where the error was found
  rule:   string   -- name of the rule that was violated
  detail: string   -- human-readable description of the violation
```

---

## Helper: IsLeaf(logical_name, all_logical_names) -> boolean

Determines whether a node is a leaf (has no children).

1. For each name in all_logical_names:
     If name starts with <logical_name + "/"> then
       return false
2. Return true.

---

## Helper: IsAncestor(candidate, logical_name) -> boolean

Determines whether candidate is an ancestor of logical_name.

1. If logical_name starts with <candidate + "/"> then
     return true.
2. Return false.

---

## Helper: IsDescendant(candidate, logical_name) -> boolean

Determines whether candidate is a descendant of logical_name.

1. If candidate starts with <logical_name + "/"> then
     return true.
2. Return false.

---

## Function: ValidateFormat(discovered_nodes) -> list of FormatError

`discovered_nodes` is a list of records, each with:
  - logical_name: string
  - file_path: string

Returns a list of FormatError records. The list is empty when
all nodes pass every rule.

### Steps

1. Collect all_logical_names: extract the logical_name field
   from every entry in discovered_nodes into a flat list.

2. Initialize errors as an empty list.

3. For each node in discovered_nodes (logical_name, file_path):

   a. Open the file at file_path using OpenFileReader.
      If the file cannot be opened, append a FormatError:
        node:   <logical_name>
        rule:   "unreadable node"
        detail: "cannot open file at <file_path>"
      Skip all remaining rules for this node and continue
      to the next node.

   b. Parse frontmatter by calling ParseFrontmatter(file_path).
      If parsing fails with "file unreadable", treat the same
      as step (a).
      If parsing fails with "malformed YAML", append a FormatError:
        node:   <logical_name>
        rule:   "malformed frontmatter"
        detail: "YAML between --- delimiters is invalid"
      Continue with remaining rules using an empty frontmatter.

   c. Parse the node body by calling ParseNode(logical_name).
      If ParseNode raises any error, append a FormatError for
      each raised condition:
        node:   <logical_name>
        rule:   <error name, e.g. "unexpected content before first heading">
        detail: <error message from ParseNode>
      Continue with remaining rules where possible.

   d. Determine leaf status:
      is_leaf = IsLeaf(logical_name, all_logical_names)

   -- Rule: Name verification --
   e. Use ReverseResolve(file_path) to derive the expected
      logical name from the filesystem path.
      Normalize both the derived name and the heading text from
      the parsed node's name_section using NormalizeName.
      If the normalized values do not match, append a FormatError:
        node:   <logical_name>
        rule:   "name verification"
        detail: "first heading <heading text> does not match
                 expected logical name <derived name>"

   -- Rule: Frontmatter field restrictions --
   f. If is_leaf is false:
      If frontmatter has any non-empty depends_on, append:
        node:   <logical_name>
        rule:   "frontmatter field restrictions"
        detail: "depends_on is not permitted on intermediate nodes"
      If frontmatter has any non-empty external, append:
        node:   <logical_name>
        rule:   "frontmatter field restrictions"
        detail: "external is not permitted on intermediate nodes"
      If frontmatter has a non-empty input, append:
        node:   <logical_name>
        rule:   "frontmatter field restrictions"
        detail: "input is not permitted on intermediate nodes"
      If frontmatter has any non-empty outputs, append:
        node:   <logical_name>
        rule:   "frontmatter field restrictions"
        detail: "outputs is not permitted on intermediate nodes"

   -- Rule: Agent section restrictions --
   g. If is_leaf is false:
      If parsed node has a non-empty agent section, append:
        node:   <logical_name>
        rule:   "agent section restrictions"
        detail: "# Agent section is not permitted on intermediate nodes"

   -- Rule: Dependency targets --
   h. For each dep_name in frontmatter.depends_on:
      i.  Call ResolveArtifactReference or ResolvePath to obtain
          the target node path or artifact reference, based on
          whether dep_name starts with "ARTIFACT/" or "ROOT/".

          For ROOT/ references:
            Call ResolvePath(dep_name) to get the file path.
            Check whether the file at that path exists.
            If it does not exist, append:
              node:   <logical_name>
              rule:   "dependency targets"
              detail: "depends_on entry <dep_name> does not resolve
                       to an existing _node.md file"

          For ARTIFACT/ references:
            Call ResolveArtifactReference(dep_name).
            No file-existence check is required here (artifact
            existence is outside the scope of format validation).

      ii. If dep_name is a ROOT/ reference:
          Strip any parenthetical qualifier using ExtractQualifier
          to get the bare target logical name.
          If IsAncestor(bare_target, logical_name) is true, append:
            node:   <logical_name>
            rule:   "dependency targets"
            detail: "depends_on entry <dep_name> points to an ancestor
                     (ancestor content is already inherited)"
          If IsDescendant(bare_target, logical_name) is true, append:
            node:   <logical_name>
            rule:   "dependency targets"
            detail: "depends_on entry <dep_name> points to a descendant
                     (would create a circular dependency)"

   -- Rule: External file existence --
   i. For each ext in frontmatter.external:
      Check whether the file at ext.path exists.
      If it does not exist, append:
        node:   <logical_name>
        rule:   "external file existence"
        detail: "external file <ext.path> does not exist"
        Skip fragment checks for this entry and continue.

      If ext.fragments is non-empty:
        Open a FileReader for ext.path using OpenFileReader.
        For each fragment in ext.fragments:
          Parse fragment.lines as a range "<start>-<end>"
          (both values are 1-based line numbers, inclusive).
          Use SkipLines to skip to line <start>, then read
          lines from <start> to <end> inclusive using ReadLine.
          Concatenate those lines with LF terminators.
          Compute the SHA-1 digest of the concatenated content
          (after normalizing CRLF to LF) and encode as base64url
          (no padding, 27 characters).
          If the computed hash does not equal fragment.hash, append:
            node:   <logical_name>
            rule:   "external file existence"
            detail: "fragment hash mismatch for <ext.path> lines
                     <fragment.lines>: expected <fragment.hash>,
                     got <computed hash>"

   -- Rule: Output path validation --
   j. For each out in frontmatter.outputs:
      Call ValidatePath(out.path, project_root).
      If ValidatePath raises any error, append:
        node:   <logical_name>
        rule:   "output path validation"
        detail: <error message from ValidatePath, including out.path>

   -- Rule: Duplicate public subsections --
   k. If the parsed node has a public section:
      Collect all subsection headings from public.subsections.
      Normalize each heading using NormalizeName.
      Build a list of seen normalized headings (initially empty).
      For each normalized heading:
        If it already appears in seen, append:
          node:   <logical_name>
          rule:   "duplicate public subsections"
          detail: "subsection heading <original heading> in # Public
                   duplicates an earlier heading after normalization"
        Else add it to seen.

4. Return errors.

---

## Contracts

- Every discovered node is validated regardless of whether it is
  a leaf or intermediate node.
- All rules are checked for every node — validation does not stop
  at the first error, neither within a node nor across nodes.
- project_root is the root directory of the project, provided by
  the caller and passed through to ValidatePath and file-existence
  checks.
