# Tooling

Operations that a Code from Spec tool must implement.

This document assumes familiarity with
CODE_FROM_SPEC.md and MANIFEST.md.

---

## validate_specs

Validate the spec tree and report the status of all
artifacts.

**Parameters:** none.

**Returns:** a report containing format errors, cycles,
and artifact status. Always returns a report — never
raises an error. Problems are collected in the report.

**Behavior:**

1. Walk the spec tree and check every node for format
   errors (see FILE_FORMAT.md and CODE_FROM_SPEC.md).
2. Detect circular references across `depends_on`,
   `input`, and inheritance. Report cycle participants.
3. For each node that declares `output`, determine the
   artifact status by comparing the manifest against
   the current spec tree and file system (see
   MANIFEST.md, "Artifact status"). Each entry includes
   the node's rank — entries with equal rank have no
   dependency between them and can be processed in
   parallel.
4. Report all findings: format errors, cycles, and
   artifact status (stale, modified, missing, orphan).

Nodes without `output` are not checked for staleness —
they do not generate artifacts.

---

## load_chain

Load the complete spec chain for a given node.

**Parameters:**

- `logical_name` (string, required) — the logical name
  of the target node. The node must declare `output`.

**Returns:** an XML document as defined in
CHAIN_ASSEMBLY.md. The document contains up to seven
sections: `<previous_constraints>`,
`<previous_instructions>`, `<previous_input>`,
`<existing_artifact>`, `<constraints>`,
`<instructions>`, and `<input>`.

The `<previous_*>` sections are populated from the
cache when available (see CACHE.md). The
`<existing_artifact>` section is present only when the
output file exists on disk — if the file does not
exist or cannot be read, the section is omitted
silently.

The content within `<constraints>` entries matches
exactly what is hashed — hash and delivery never
diverge (see FILE_FORMAT.md, "Block extraction").

If the artifact is modified (checksum in the manifest
does not match the file on disk), returns an error.
The artifact must be accepted or deleted before
regeneration.

If any file in the chain (other than the existing
artifact) is unreadable, returns an error.

---

## write_file

Write a generated artifact to disk and update the
manifest.

**Parameters:**

- `logical_name` (string, required) — the logical name
  of the node whose `output` authorizes the write.
  Must not contain a parenthetical qualifier.
- `path` (string, required) — file path relative to the
  project root. Must match the node's declared `output`.
- `content` (string, required) — complete file content.

**Behavior:**

1. Validate that `logical_name` has no qualifier and
   that `path` matches the `output` declared in the
   node's frontmatter.
2. Write the file to disk.
3. Compute the checksum (hash of the written content)
   and the current chain hash.
4. Update the manifest entry for this node with the
   new checksum and chain hash.

The manifest must be updated atomically. See
MANIFEST.md ("Concurrency") for locking requirements.

---

## reconstruct_cache

Populate the cache from the current state of the
repository.

**Parameters:** none.

**Behavior:**

For each entry in the manifest, resolve the chain and
populate `.cache/.content/` with the processed content
of each position, and `.cache/.chains/` with the chain
structure. Idempotent — skips files that already exist
in the cache.

See CACHE.md for details on the cache structure.

---

## prune_cache

Remove unreferenced files from the cache.

**Parameters:** none.

**Behavior:**

Delete content files in `.cache/.content/` whose hash
is not referenced by any chain file in
`.cache/.chains/`. Delete chain files in
`.cache/.chains/` whose hash is not referenced by any
manifest entry.

See CACHE.md for details on the cache structure.

---

## accept

Accept a modified artifact without regenerating it.
Updates the manifest checksum to match the current
file on disk.

**Parameters:**

- `logical_name` (string, required) — the logical name
  of the node whose artifact was modified.

**Behavior:**

1. Verify the artifact is in "modified" status
   (checksum mismatch). If not, return an error.
2. Compute the hash of the file on disk.
3. Update the manifest entry's checksum to match.

The chain hash in the manifest is not changed — the
artifact is accepted as-is against the same spec
version that produced it.

---

## dump_chain

Save the spec chain for a given node to a file for
inspection.

**Parameters:**

- `logical_name` (string, required) — the logical name
  of the target node. The node must declare `output`.

**Behavior:**

Assemble the spec chain exactly as `load_chain` would,
and write it to `<project root>/dump_chain.xml`. This
produces the same document the generation subagent
would receive, allowing the orchestrator or the human
to inspect it.

---

## version

Report the tool version.

**Parameters:** none.

**Returns:** the version string.

---

## Resources

| Document | Description |
|---|---|
| [CODE_FROM_SPEC.md](https://github.com/CodeFromSpec/framework/blob/main/CODE_FROM_SPEC.md) | Full methodology specification |
| [CHAIN_ASSEMBLY.md](https://github.com/CodeFromSpec/framework/blob/main/rules/CHAIN_ASSEMBLY.md) | Chain format, assembly order, and delivery |
| [MANIFEST.md](https://github.com/CodeFromSpec/framework/blob/main/rules/MANIFEST.md) | Manifest format and artifact status |
| [CACHE.md](https://github.com/CodeFromSpec/framework/blob/main/rules/CACHE.md) | Cache structure for disposition computation |
| [tool-framework-mcp](https://github.com/CodeFromSpec/tool-framework-mcp) | Reference implementation |
