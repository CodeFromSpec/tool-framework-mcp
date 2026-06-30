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
`<instructions>`, etc.) were chosen to be
self-explanatory to the subagent, and differ from the
terminology used elsewhere in the framework.

The document has up to seven sections, in this order:

1. **`<previous_constraints>`** — old spec content for
   positions that changed or were removed, as recorded
   in the cache. Present only when the cache is
   available and the existing artifact is present on
   disk. Contains only entries with
   `disposition="changed"` or `disposition="removed"`,
   each with their old content. Positions that did not
   change are not listed — their `unchanged`
   disposition is on the corresponding entry in
   `<constraints>`.

2. **`<previous_instructions>`** — the previous
   `# Agent` section (excluding the `# Agent` heading),
   as recorded in the cache. Present only when the
   cache is available, the existing artifact is present
   on disk, and the instructions changed or were
   removed. Carries `disposition="changed"` or
   `disposition="removed"`. Contains the old content.

3. **`<previous_input>`** — the previous input content,
   as recorded in the cache. Present only when the
   cache is available, the existing artifact is present
   on disk, and the input changed or was removed.
   Carries `disposition="changed"` or
   `disposition="removed"`. Contains the old content.

4. **`<existing_artifact>`** — the current content of
   the artifact file on disk. Present only when the
   file exists.

5. **`<constraints>`** — the current spec content. Each
   position is an `<entry>` element with a `name`
   attribute identifying the source. When cache is
   available and the existing artifact is present,
   each entry carries a `disposition` attribute:
   - `unchanged` — same name, same content hash as
     the previous generation.
   - `changed` — same name, different content hash.
   - `added` — no counterpart in the previous chain.
   When cache is not available, entries have no
   `disposition`.

6. **`<instructions>`** — the target node's `# Agent`
   section. The `# Agent` heading is not included.
   Present only when the node has an `# Agent` section.
   When cache is available and the existing artifact
   is present, carries a `disposition` attribute:
   `unchanged`, `changed`, or `added`.

7. **`<input>`** — the content referenced by the target
   node's `input` field. For `SPEC/` references, the
   `# Public` content is extracted using the same rules
   as `<constraints>` entries. Present only when the
   node declares `input`. When cache is available and
   the existing artifact is present, carries a
   `disposition` attribute: `unchanged`, `changed`,
   or `added`.

Sections 1–4 provide context from the previous
generation: what the spec said before, what input was
used, and what was generated from them. Section 5 is
the current source of truth. Section 6 is the
generation guidance. Section 7 is the material to
transform.

## Generation scenarios

Three scenarios determine which sections are present:

- **First generation** (no existing artifact): the
  spec chain contains only `<constraints>`,
  `<instructions>`, and optionally `<input>`. Even
  if the cache has data, it is not used — there is
  no existing code to compare against.

- **Regeneration without cache** (existing artifact,
  no cache): the spec chain contains
  `<existing_artifact>`, `<constraints>`,
  `<instructions>`, and optionally `<input>`. No
  `<previous_*>` sections — the subagent compares
  the existing artifact directly against the current
  spec.

- **Regeneration with cache** (existing artifact and
  cache available): all seven sections may be present.
  The `<previous_*>` sections and the current
  `<constraints>`, `<instructions>`, and `<input>`
  carry disposition attributes showing exactly what
  changed.

---

## Constraints assembly order

Positions within `<constraints>` appear in this order:

1. Ancestors from root to the target node's parent.
2. `depends_on` entries in alphabetical order by the
   full logical name (including prefix and qualifier).
3. The target node's `# Public`.

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
depends_on:
  - SPEC/legacy/old-fees
input: ARTIFACT/functional/fees/calculation
output: internal/fees/calculation.go
---
```

Current frontmatter:

```yaml
---
depends_on:
  - EXTERNAL/proto/payments/v1/transfers.proto
  - SPEC/integrations/database
input: ARTIFACT/functional/fees/calculation
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
    <entry name="SPEC/legacy/old-fees" disposition="removed">
    ...old content...
    </entry>
  </previous_constraints>

  <previous_instructions disposition="changed">
  ...previous # Agent content...
  </previous_instructions>

  <existing_artifact>
  ...current file on disk...
  </existing_artifact>

  <constraints>
    <entry name="SPEC/payments" disposition="unchanged">...</entry>
    <entry name="SPEC/payments/fees" disposition="changed">...</entry>
    <entry name="EXTERNAL/proto/payments/v1/transfers.proto" disposition="added">...</entry>
    <entry name="SPEC/integrations/database" disposition="added">...</entry>
    <entry name="SPEC/payments/fees/calculation" disposition="unchanged">...</entry>
  </constraints>

  <instructions disposition="changed">
  ...generation guidance...
  </instructions>

  <input disposition="unchanged">
  ...material to transform...
  </input>
</chain>
```

---

## Resources

| Document | Description |
|---|---|
| [CODE_FROM_SPEC.md](https://github.com/CodeFromSpec/framework/blob/main/CODE_FROM_SPEC.md) | Full methodology specification |
| [CHAIN_HASH.md](https://github.com/CodeFromSpec/framework/blob/main/rules/CHAIN_HASH.md) | Chain hash algorithm for staleness detection |
| [CACHE.md](https://github.com/CodeFromSpec/framework/blob/main/rules/CACHE.md) | Cache structure for disposition computation |
| [FILE_FORMAT.md](https://github.com/CodeFromSpec/framework/blob/main/rules/FILE_FORMAT.md) | Block extraction and normalization rules |
| [TOOLING.md](https://github.com/CodeFromSpec/framework/blob/main/rules/TOOLING.md) | Operations a tool must implement |
