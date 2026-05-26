# Artifact Generation with Subagents

How to generate artifacts for a given logical name using a
confined subagent. This document assumes familiarity with
[CODE_FROM_SPEC.md](CODE_FROM_SPEC.md).

---

## Overview

Artifact generation should be performed by confined subagents.
Given a logical name, the orchestrator dispatches a subagent that
receives the spec chain, reviews the specification, and either
generates the artifacts or reports gaps.

---

## Confinement

Ideally, a subagent should only have access to the spec chain
for the target node and the ability to write the declared output
files. It should not explore the filesystem, read unrelated
files, or fetch external information. If the chain is
insufficient, the correct action is to report what is missing.

The `framework-mcp` tool (see Resources in
[CODE_FROM_SPEC.md](CODE_FROM_SPEC.md)) enforces this
confinement. Its tools include:

- `load_chain` — returns the complete spec chain for a logical
  name, including the current chain hash
- `write_file` — writes a file to disk, validated against the
  node's `outputs` list

When `framework-mcp` is available, the orchestrator should
configure the subagent with access to only these tools and no
other filesystem access. A reference subagent definition is
provided at [subagents/code-from-spec-code-generation.md](../subagents/code-from-spec-code-generation.md).

When it is not available, the orchestrator is responsible for
assembling the chain and delivering it to the subagent by other
means (e.g., in the prompt), and for restricting the subagent's
write access (if possible).

---

## How to generate

Given a logical name:

1. Dispatch a subagent with that logical name in the prompt.

2. The subagent obtains the spec chain, reviews the
   specification, and produces one of two results:

   - **Generated artifacts** — written to disk. Each file
     contains a artifact tag identifying the source node and
     chain hash.

   - **Findings report** — the specification is ambiguous,
     incomplete, or contradictory. The subagent reports exactly
     what is wrong. This is correct output — fix the spec and
     retry.

   Both outcomes are equally valid. The subagent may be
   dispatched during specification design specifically to find
   gaps, or during artifact generation to produce files.
