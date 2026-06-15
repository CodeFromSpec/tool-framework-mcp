---
name: cfs-spec-review
description: Review a spec node for ambiguities, omissions, and precision gaps before generating artifacts. Dispatches a confined subagent that can only read the chain, not write files.
---

# Spec Review

Review a spec node's completeness and precision before
generating its artifact.

## When invoked

Run this skill when:

- A new node is created and the human wants to verify
  the spec is ready for generation.
- A node was significantly rewritten.
- A generated artifact changed more than expected after
  regeneration (suggests the spec has variability
  points).
- The human explicitly asks to review a spec.

The user provides a logical name (e.g.,
`SPEC/architecture/backend/internal/api/operations/account-close/implementation`).

## Prerequisites

The framework-mcp MCP server must be connected (the
`load_chain` tool must be available).

## Algorithm

1. Dispatch a `cfs-spec-review` subagent with the
   logical name:

   > You are a spec review subagent. Your task is to
   > review the specification for `<logical-name>` and
   > report ambiguities, omissions, and inconsistencies.
   >
   > Call `load_chain` with logical_name `<logical-name>`
   > and analyze the chain.

2. When the subagent completes, present its findings
   to the user organized by severity:
   - **Ambiguities** and **Inconsistencies** first —
     these must be resolved before generation.
   - **Omissions** next — the human decides which
     matter.
   - **Variability points** last — the human decides
     whether to accept or prescribe.

3. Do not automatically fix anything. Present findings
   for the human to review and decide.

## Rules

- Dispatch exactly one subagent per logical name.
- Multiple nodes can be reviewed in parallel (single
  message with multiple Agent tool calls).
- Do not generate artifacts as part of a review. The
  review and generation are separate steps.
- If the subagent reports the spec is ready for
  generation with no findings, report that to the
  user and suggest running `/cfs-generate`.
