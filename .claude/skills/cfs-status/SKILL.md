---
name: cfs-status
description: Run validate_specs and report the current state of the spec tree — format errors, cycles, and artifact staleness — in a readable format.
---

# Spec Tree Status

Run `validate_specs` and present the results in a clear,
actionable format.

## When invoked

Run this skill when the user asks to check the spec tree
status, verify staleness, or invokes `/cfs-status`.

## Prerequisites

The framework-mcp MCP server must be connected (the
`validate_specs` tool must be available).

## Algorithm

1. Call `validate_specs`.

2. If the spec tree is clean (no format errors, no cycles,
   no staleness), report:

   > Spec tree is clean. All artifacts are up to date.

   and stop.

3. Otherwise, format the report in sections:

   **Format errors** (if any):

   List each error with node, rule, and detail. Group by
   rule if there are many.

   **Cycles** (if any):

   List the logical names involved in cycles.

   **Staleness** (if any):

   Group by rank (ascending). For each rank, list the
   artifacts with their status (`stale`, `modified`,
   `missing`, or `orphan`) and output path. Show counts per
   rank and total.

   Example:

   ```
   ## Staleness (12 artifacts)

   ### Rank 7 (4 artifacts)
   - SPEC/golang/interfaces/os/file_reader → code-from-spec/.../output.md (stale)
   - SPEC/golang/interfaces/os/file_writer → code-from-spec/.../output.md (stale)
   - SPEC/golang/implementation/os/path_utils → internal/pathutils/pathutils.go (stale)
   - SPEC/golang/tests/os/path_utils → internal/pathutils/pathutils_test.go (stale)

   ### Rank 9 (8 artifacts)
   ...
   ```

4. End with a summary line:

   > X format errors, Y cycles, Z stale artifacts.

## Rules

- Do not take any action beyond reporting. Do not
  offer to regenerate or fix — just report the state.
- If `validate_specs` is not available (MCP server not
  connected), report the error and suggest reconnecting.
