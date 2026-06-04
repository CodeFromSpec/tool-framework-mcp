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
file. It should not explore the filesystem, read unrelated
files, or fetch external information. If the chain is
insufficient, the correct action is to report what is missing.

The `framework-mcp` tool (see Resources in
[CODE_FROM_SPEC.md](CODE_FROM_SPEC.md)) enforces this
confinement. Its tools include:

- `load_chain` — returns the complete spec chain for a logical
  name, including the current chain hash
- `write_file` — writes a file to disk, validated against the
  node's `output` path

When `framework-mcp` is available, the orchestrator should
configure the subagent with access to only these tools and no
other filesystem access. A reference subagent definition is
provided at [subagents/cfs-artifact-generation.md](../subagents/cfs-artifact-generation.md).

When it is not available, the orchestrator is responsible for
assembling the chain and delivering it to the subagent by other
means (e.g., in the prompt), and for restricting the subagent's
write access (if possible).

---

## Existing artifact as reference

When regenerating a stale artifact, the subagent should
receive the existing artifact content alongside the spec
chain. The `load_chain` tool includes the existing artifact
automatically when the output file exists on disk — the
orchestrator does not need to read or relay it.

The subagent compares the spec with the existing code and
produces minimal changes — preserving what already satisfies
the spec and modifying only what needs to change.

This reduces diff noise, avoids unnecessary churn, and makes
code review practical. Without the existing artifact, the
subagent generates from scratch every time, producing
different variable names, function ordering, and formatting
even when the behavior is identical.

The existing artifact is **not** part of the chain and does
**not** participate in the chain hash. It is delivered
alongside the chain but does not affect staleness detection.

If the subagent anchors on a bug in the existing artifact
(reproducing it instead of following the spec), delete the
artifact and regenerate from scratch. The decision to include
or exclude the existing artifact is the human's, case by
case.

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
