---
name: cfs-generate
description: Generates or regenerates artifacts from the Code from Spec tree. Use when stale artifacts exist, or when the user asks to generate or regenerate artifacts.
---

# Artifact Generation

Generate artifacts for stale entries reported by
`validate_specs`.

## When invoked

Run this skill when the user asks to generate or regenerate
artifacts, or when stale artifacts exist.

## Prerequisites

1. Verify the framework-mcp MCP server is connected (the
   `validate_specs`, `load_chain`, and `write_file` tools must
   be available).

2. Run `validate_specs`. If `format_errors` are reported, stop
   and tell the user to fix them first — artifact generation
   requires a clean spec tree.

## Algorithm

1. Run `validate_specs` and collect all stale/missing artifacts.
2. If no stale artifacts, report that everything is up to date
   and stop.
3. Group stale artifacts by rank. The rank (returned by
   `validate_specs`) reflects dependency depth — artifacts
   with lower rank must be generated before artifacts with
   higher rank, because higher-rank artifacts may depend on
   them. Process ranks in ascending order. Within the same
   rank, artifacts are independent and should be dispatched
   in parallel. For each artifact, dispatch a
   `cfs-artifact-generation` subagent.

   Prompt:

   > You are a confined artifact generation subagent.
   > Your only task is to generate the artifact
   > for the node `<logical-name>`.

4. **After each rank completes, run `validate_specs` again
   before starting the next rank.** This is mandatory, not
   an optimization to skip. Regenerating rank N changes
   artifact content, which may cause rank N+1 artifacts
   that depend on them to become stale. Without
   re-validating, newly stale artifacts are missed. The
   `validate_specs` call between ranks is what keeps the
   generation session consistent.
5. After all ranks are processed, run `validate_specs` a
   final time. Report the remaining stale items (if any)
   to the user.

## Rules

- Dispatch one subagent per artifact.
- **Do not add guidance, hints, or corrections to the subagent
  prompt beyond the template above.** The subagent must
  generate from the chain alone. If a previous generation
  produced a wrong result, the fix belongs in the spec — not
  in an ad-hoc instruction injected into the prompt. Prompt
  additions bypass the chain, are not versioned, do not
  participate in the hash, and will not reproduce on the next
  regeneration.
- Artifacts with the same rank are independent — dispatch them
  in parallel (single message with multiple Agent tool calls).
  Wait for all artifacts in a rank to complete before starting
  the next rank.
- Never edit generated files manually — always regenerate via
  a subagent.
- After each subagent completes, check its output for an
  `## Assumptions` section or any language indicating the spec
  was ambiguous, silent, or required interpretation (e.g.,
  "the spec does not specify", "chose", "assumed", "not
  defined"). Collect all such items and present them to the
  user **before** reporting success. These are potential spec
  gaps that need confirmation.
- If a subagent reports a spec gap that prevented generation,
  surface it to the user. Do not attempt to fill the gap by
  reading the codebase yourself.
- After generation, do not automatically run build or tests
  unless the user asks — report what was generated and let the
  user decide.
- Track and report token usage. After each rank completes,
  report the cumulative subagent tokens spent in this
  generation session. At the end, report the total.
