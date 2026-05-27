[//]: # (code-from-spec: ROOT/functional/mcp_tools/validate_specs@rX-wkva3YXYnq3S7ZqdODj7z478)

# validate_specs

## Overview

`validate_specs` is an MCP tool that inspects the `code-from-spec/` tree and
reports every structural problem it finds. It is a read-only, side-effect-free
operation: it walks the filesystem, parses files, checks rules, and returns a
human-readable report.

## Inputs

The tool takes no parameters.

## Outputs

A single string: a plain-text report. The report is either

- `"All specs are valid."` when no problems are found, or
- a list of findings, one per line, each prefixed with the node logical name and
  the rule that was violated.

## Algorithm

### Step 1 — Discover nodes

Call `DiscoverNodes()`.

- If it returns a `directory not found` error, return a report with a single
  line: `"Error: code-from-spec/ directory not found."`.
- If it returns a `walk error`, return `"Error: filesystem error while
  traversing code-from-spec/."`.
- If it returns `no nodes found`, return `"Error: no _node.md files found under
  code-from-spec/."`.

Proceed with the returned `list of DiscoveredNode`.

### Step 2 — Validate format

Call `ValidateFormat(discovered_nodes)`.

Collect every `FormatError` returned. Each error has three fields: `node`,
`rule`, and `detail`.

If any node file is unreadable during this step, record a format error with
`rule = "unreadable node"` and `detail` set to the file path.

### Step 3 — Detect cycles and compute ranks

For every discovered node, read its frontmatter with `ParseFrontmatter`. Ignore
nodes whose frontmatter cannot be read (they are already covered by Step 2).

Call `DetectCycles(nodes)` with the set of nodes that parsed successfully.

- If `DetectCycles` returns an `unresolvable reference` error, record a problem:
  `rule = "unresolvable reference"`, with the offending node and detail
  describing the target that could not be resolved.
- For each logical name in `cycle_participants`, record a problem:
  `rule = "dependency cycle"`, with `detail = "node participates in a
  dependency cycle"`.

### Step 4 — Check staleness

For each node that has at least one output declared in its frontmatter, iterate
over the outputs. For each output:

1. Call `ValidatePath(output.path, project_root)`. If it fails, record a
   problem with `rule = "invalid output path"` and the error detail.
2. Attempt to call `ExtractArtifactTag(output.path)`.
   - `file unreadable` — record `rule = "output file missing"`, detail: the
     output path.
   - `no tag found` — record `rule = "no artifact tag"`, detail: the output
     path.
   - `malformed tag` — record `rule = "malformed artifact tag"`, detail: the
     output path.
   - On success, check that `ArtifactTag.logical_name` matches the node's own
     logical name. If not, record `rule = "artifact tag mismatch"`, detail:
     `"expected <node logical name>, got <tag logical name>"`.

### Step 5 — Assemble report

Collect all problems from Steps 2–4. If the list is empty, return `"All specs
are valid."`.

Otherwise, sort the problems: first by `node` alphabetically, then by `rule`
alphabetically. Format each problem as:

```
[<node>] <rule>: <detail>
```

Join with newlines and return the result.

## Error handling principles

- Problems that affect individual nodes are collected and included in the
  report; they do not abort the run.
- Only errors that prevent the entire walk from proceeding (Step 1) cause an
  early return.
- Every error message uses plain English with no trailing punctuation variation
  — messages end with a period only when they form a complete sentence.

## Dependencies

| Utility | Used in step |
|---|---|
| `DiscoverNodes` | 1 |
| `ValidateFormat` | 2 |
| `ParseFrontmatter` | 3 |
| `DetectCycles` | 3 |
| `ValidatePath` | 4 |
| `ExtractArtifactTag` | 4 |
