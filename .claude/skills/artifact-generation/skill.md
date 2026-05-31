---
name: artifact-generation
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
   `code-from-spec-artifact-generation` subagent with the following
   prompt:

   > You are a confined artifact generation subagent.
   > Your only task is to generate the artifact
   > `<artifact-id>` for the node `<logical-name>`.
   >
   > Steps:
   > 1. Call `load_chain` with logical_name `<logical-name>` to
   >    receive the complete spec chain. The first line of the
   >    response is `chain_hash: <hash>` — extract this hash.
   > 2. Read the chain carefully. Identify the target node's
   >    spec (its intent, contracts, and interface), the
   >    constraints from ancestor nodes, and any dependency
   >    specs.
   > 3. Generate the artifact content. The artifact must
   >    contain the artifact tag:
   >    `code-from-spec: <logical-name>@<chain-hash>`
   >    where `<chain-hash>` is the hash extracted in step 1.
   >    Place the tag as early in the file as practical, inside
   >    a comment appropriate for the file type.
   > 4. Call `write_file` with the complete file content
   >    (including the artifact tag with the correct hash).
   > 5. If the spec has gaps or contradictions that prevent
   >    generation, do not guess — report the problem clearly
   >    instead of writing a file.
   > 6. After generating, list any assumptions you made where
   >    the spec was silent or ambiguous. Label this section
   >    `## Assumptions`. Include: format choices, field
   >    mappings you inferred, interpretations of ambiguous
   >    wording. If there are none, omit the section.

4. After all subagents complete, run `validate_specs` again.
   Report the remaining stale items (if any) to the user.

## Rules

- Dispatch one subagent per artifact.
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
