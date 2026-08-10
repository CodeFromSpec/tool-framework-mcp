# Chain Assembly

How the spec chain is assembled and delivered to
generation subagents. This level of detail is primarily
relevant for tool implementors.

This document assumes familiarity with
CODE_FROM_SPEC.md and CACHE.md.

---

## Format

The spec chain is an XML document delivered via
`load_chain`. The XML is designed to be consumed by a
generation subagent that has no knowledge of Code from
Spec. The element names (`<constraints>`,
`<references>`, `<instructions>`, etc.) were chosen to
be self-explanatory to the subagent, and differ from
the terminology used elsewhere in the framework.

The document has these sections, in this order:

1. **`<previous_constraints>`** — old content for
   `<constraints>` positions that changed or were
   removed, as recorded in the cache. Present only when
   the cache is available and the existing artifact is
   present on disk. Contains only entries with
   `disposition="changed"` or `disposition="removed"`,
   each with their old content. Positions that did not
   change are not listed — their `unchanged`
   disposition is on the corresponding entry in
   `<constraints>`.

2. **`<previous_references>`** — old content for
   `<references>` positions (imports) that changed or
   were removed, as recorded in the cache. Present only
   when the cache is available and the existing artifact
   is present on disk. Same rules as
   `<previous_constraints>`, scoped to imports.

3. **`<previous_instructions>`** — the previous
   `# Agent` section (excluding the `# Agent` heading),
   as recorded in the cache. Present only when the
   cache is available, the existing artifact is present
   on disk, and the instructions changed or were
   removed. Carries `disposition="changed"` or
   `disposition="removed"`. Contains the old content.

4. **`<previous_input>`** — old content for `input`
   entries that changed or were removed, as recorded
   in the cache. Present only when the cache is
   available, the existing artifact is present on disk,
   and at least one `input` entry changed or was
   removed. Contains an `<entry>` element per affected
   entry, with a `name` attribute and
   `disposition="changed"` or `disposition="removed"`.
   Entries that did not change are not listed — their
   `unchanged` disposition is on the corresponding entry
   in `<input>`.

5. **`<existing_artifact>`** — the current content of
   the artifact file on disk. Present only when the
   file exists.

6. **`<constraints>`** — the inheritance line: ancestors
   from root to the target's parent, plus the target
   node's own `# Public` as the last entry. Each position
   is an `<entry>` element with a `name` attribute
   identifying the source. When cache is available and
   the existing artifact is present, each entry carries a
   `disposition` attribute:
   - `unchanged` — same name, same content hash as
     the previous generation.
   - `changed` — same name, different content hash.
   - `added` — no counterpart in the previous chain.
   When cache is not available, entries have no
   `disposition`.

7. **`<references>`** — the target node's `imports`,
   in alphabetical order. Same `<entry>` shape and
   disposition rules as `<constraints>`, applied to a
   different relation: material to consult, not
   convention to adopt. Present only when the node
   declares `imports`.

8. **`<instructions>`** — the target node's `# Agent`
   section. The `# Agent` heading is not included.
   Present only when the node has an `# Agent` section.
   When cache is available and the existing artifact
   is present, carries a `disposition` attribute:
   `unchanged`, `changed`, or `added`.

9. **`<input>`** — the content referenced by the target
   node's `input` field. Each position is an `<entry>`
   element with a `name` attribute, one per `input`
   reference (a single value or each item in a list).
   For `SPEC/` references, the `# Public` content is
   extracted using the same rules as `<constraints>`
   entries. Present only when the node declares `input`.
   When cache is available and the existing artifact is
   present, each entry carries a `disposition` attribute:
   `unchanged`, `changed`, or `added`.

Sections 1–5 provide context from the previous
generation: what the spec said before, what input was
used, and what was generated from them. Sections 6–7 are
the current spec content, authoritative over everything
before them. Section 8 is the generation guidance.
Section 9 is the material to transform.

This split is purely how the chain is delivered — it has
no effect on the chain hash, which continues to identify
each position by its full logical name regardless of
which section carries it (see CHAIN_HASH.md).

## Generation scenarios

Three scenarios determine which sections are present:

- **First generation** (no existing artifact): the
  spec chain contains `<constraints>` and
  `<instructions>`, plus `<references>` and `<input>`
  when the node declares them. Even if the cache has
  data, it is not used — there is no existing code to
  compare against.

- **Regeneration without cache** (existing artifact,
  no cache): the spec chain adds
  `<existing_artifact>` to the above. No `<previous_*>`
  sections — the subagent compares the existing
  artifact directly against the current spec.

- **Regeneration with cache** (existing artifact and
  cache available): all sections may be present. The
  `<previous_*>` sections and the current
  `<constraints>`, `<references>`, `<instructions>`,
  and `<input>` carry disposition attributes showing
  exactly what changed.

---

## Constraints assembly order

Positions within `<constraints>` appear in this order:

1. Ancestors from root to the target node's parent.
2. The target node's `# Public`.

---

## References assembly order

Positions within `<references>` appear in alphabetical
order by the full logical name (including prefix and
qualifier) — the target node's `imports` entries.

---

## Input assembly order

Entries within `<input>` appear in alphabetical order by
the full logical name (including prefix and qualifier),
using the same ordering and deduplication rules as
`imports` (see CHAIN_HASH.md). This order is independent
of the `imports` list — the two are deduplicated and
sorted separately.

---

## Content extraction

All content is boundary-normalized using the block
extraction rules defined in FILE_FORMAT.md ("Block
extraction"). The extracted form is what is delivered
— hash and delivery never diverge.

For `# Agent`, the `# Agent` heading is not included
— only the content within it.

---

## Example

Generating an artifact for
`SPEC/payments/fees/calculation`.

Previous frontmatter (at the time of last generation):

```yaml
---
type: artifact
imports:
  - SPEC/legacy/old-fees
input: ARTIFACT/functional/fees/calculation
output: internal/fees/calculation.go
---
```

Current frontmatter:

```yaml
---
type: artifact
imports:
  - EXTERNAL/proto/payments/v1/transfers.proto
  - SPEC/integrations/database
input:
  - ARTIFACT/functional/fees/calculation
  - ARTIFACT/functional/fees/rounding
output: internal/fees/calculation.go
---
```

The resulting spec chain:

```xml
<chain>
  <previous_constraints>
    <entry name="SPEC/payments/fees" disposition="changed">
    ...old content...
    </entry>
  </previous_constraints>

  <previous_references>
    <entry name="SPEC/legacy/old-fees" disposition="removed">
    ...old content...
    </entry>
  </previous_references>

  <previous_instructions disposition="changed">
  ...previous # Agent content...
  </previous_instructions>

  <existing_artifact>
  ...current file on disk...
  </existing_artifact>

  <constraints>
    <entry name="SPEC/payments" disposition="unchanged">...</entry>
    <entry name="SPEC/payments/fees" disposition="changed">...</entry>
    <entry name="SPEC/payments/fees/calculation" disposition="unchanged">...</entry>
  </constraints>

  <references>
    <entry name="EXTERNAL/proto/payments/v1/transfers.proto" disposition="added">...</entry>
    <entry name="SPEC/integrations/database" disposition="added">...</entry>
  </references>

  <instructions disposition="changed">
  ...generation guidance...
  </instructions>

  <input>
    <entry name="ARTIFACT/functional/fees/calculation" disposition="unchanged">
    ...material to transform...
    </entry>
    <entry name="ARTIFACT/functional/fees/rounding" disposition="added">
    ...material to transform...
    </entry>
  </input>
</chain>
```

`ARTIFACT/functional/fees/rounding` is new in this
generation, so it has no counterpart in the previous chain
and no entry in `<previous_input>` — `added` entries carry
no old content to show.

---

## Resources

| Document | Description |
|---|---|
| [CODE_FROM_SPEC.md](https://github.com/CodeFromSpec/framework/blob/main/CODE_FROM_SPEC.md) | Full methodology specification |
| [CHAIN_HASH.md](https://github.com/CodeFromSpec/framework/blob/main/rules/CHAIN_HASH.md) | Chain hash algorithm for staleness detection |
| [CACHE.md](https://github.com/CodeFromSpec/framework/blob/main/rules/CACHE.md) | Cache structure for disposition computation |
| [FILE_FORMAT.md](https://github.com/CodeFromSpec/framework/blob/main/rules/FILE_FORMAT.md) | Block extraction and normalization rules |
| [TOOLING.md](https://github.com/CodeFromSpec/framework/blob/main/rules/TOOLING.md) | Operations a tool must implement |
