# Tooling

Operations that a Code from Spec tool must implement.

This document assumes familiarity with CODE_FROM_SPEC.md and
MANIFEST.md.

---

## validate_specs

Validate the spec tree and report the status of all artifacts.

**Parameters:** none.

**Returns:** a report containing format errors, cycles, and artifact
status. Always returns a report — never raises an error. Problems are
collected in the report.

**Behavior:**

1. Walk the spec tree and check every node for format errors (see
   FILE_FORMAT.md and CODE_FROM_SPEC.md).
2. Detect circular references across `imports`, `input`, and
   inheritance. Report cycle participants.
3. For each node that declares `type`, determine the artifact or
   verdict status by comparing the manifest against the current spec
   tree and file system (see MANIFEST.md, "Artifact and verdict
   status"). Each entry includes the
   node's rank — entries with equal rank have no dependency between
   them and can be processed in parallel.
4. Report all findings: format errors, cycles, and artifact and
   verdict status (stale, modified, missing, orphan). Verdict
   entries include their `result`.

Nodes without `type` are not checked for staleness — they generate
nothing.

---

## load_chain

Load the complete spec chain for a given node.

**Parameters:**

- `token` (string, required) — an opaque token identifying the target
  node, as returned by `create_token`. The node must declare `type`.

**Returns:** an XML document as defined in CHAIN_ASSEMBLY.md. The
document contains these core sections:

- `<existing_artifact>` — present only when the output file exists on
  disk. If the file does not exist or cannot be read, the section is
  omitted silently.
- `<constraints>` — the inheritance line: ancestors plus the target
  node's own `# Public`. Each position is an `<entry>` element with a
  `name` attribute.
- `<references>` — the target node's `imports`, in alphabetical
  order. Present only when the node declares `imports`.
- `<instructions>` — the target node's `# Agent` section.
- `<input>` — the content referenced by the target node's `input`
  field. Each position is an `<entry>` element with a `name`
  attribute, one per `input` reference (a single value or each item
  in a list).

When cache is available, additional sections may appear before the
core sections: `<previous_constraints>`, `<previous_references>`,
`<previous_instructions>`, `<previous_input>` (see CACHE.md).

For nodes with `type: verdict`, the document contains only
`<constraints>`, `<references>`, `<instructions>`, and `<input>` —
never `<existing_artifact>` or `<previous_*>` sections (see
CHAIN_ASSEMBLY.md, "Verdict chains").

The content within `<constraints>` and `<references>` entries matches
exactly what is hashed — hash and delivery never diverge (see
FILE_FORMAT.md, "Block extraction").

If the artifact or verdict is modified (checksum in the manifest does
not match the file on disk), returns an error. It must be accepted or
deleted before regeneration.

If any file in the chain (other than the existing artifact) is
unreadable, returns an error.

If `token` is malformed or was not produced by `create_token`, returns
an error.

---

## write_file

Write a generated artifact to disk and update the manifest.

**Parameters:**

- `token` (string, required) — an opaque token identifying the node
  whose `output` declares the target path, as returned by
  `create_token`.
- `content` (string, required) — complete file content.

**Behavior:**

1. Resolve `token` to the target node's logical name. Read the node's
   frontmatter and derive the output path from its `output` field,
   applying the default when the field is absent.
2. Write the file to disk at the derived path.
3. Compute the checksum (hash of the written content) and the current
   chain hash.
4. Update the manifest entry for this node with the new checksum and
   chain hash.

If `token` is malformed, was not produced by `create_token`, or the
node's `type` is not `artifact`, returns an error.

The manifest must be updated atomically. See MANIFEST.md
("Concurrency") for locking requirements.

---

## write_verdict

Write a verdict document to disk and update the manifest.

**Parameters:**

- `token` (string, required) — an opaque token identifying the target
  node, as returned by `create_token`. The node must declare
  `type: verdict`.
- `passed` (boolean, required) — the verdict: `true` for pass, `false`
  for fail.
- `content` (string, required) — complete verdict document content.

**Behavior:**

1. Resolve `token` to the target node's logical name. Read the node's
   frontmatter and derive the output path from its `output` field,
   applying the default when the field is absent.
2. Write the document to disk at the derived path.
3. Compute the checksum (hash of the written content) and the current
   chain hash.
4. Update the `VERDICT/` manifest entry for this node with the new
   checksum, chain hash, and `result:pass` or `result:fail`.

If `token` is malformed, was not produced by `create_token`, or the
node's `type` is not `verdict`, returns an error.

The manifest must be updated atomically. See MANIFEST.md
("Concurrency") for locking requirements.

---

## create_token

Mint an opaque token for a logical name.

**Parameters:**

- `logical_name` (string, required) — the logical name of the target
  node. Must be a `SPEC/` reference with no parenthetical qualifier.

**Returns:** an opaque token string encoding `logical_name`.

**Behavior:**

Generate a token that `load_chain` and `write_file` accept in place of
a raw logical name, and that only this operation can produce. This
lets an orchestrator hand a generation subagent a token instead of a
logical name: since the subagent has no way to mint a token itself, it
cannot request the chain or write the output of a node other than the
one it was dispatched for, even though the chain it receives from
`load_chain` may reference other logical names (via inheritance,
`imports`, or `input`).

`create_token` is intended for the orchestrator only and must not be
exposed to generation subagents — exposing it defeats the confinement
it provides.

If `logical_name` is not a `SPEC/` reference, or contains a
parenthetical qualifier, returns an error.

---

## reconstruct_cache

Populate the cache from the current state of the repository.

**Parameters:** none.

**Behavior:**

For each `ARTIFACT/` entry in the manifest, resolve the chain and
populate `.cache/.content/` with the processed content of each
position, and `.cache/.chains/` with the chain structure. Verdict
chains are not cached (see CACHE.md). Idempotent — skips files
that already exist in the cache.

See CACHE.md for details on the cache structure.

---

## prune_cache

Remove unreferenced files from the cache.

**Parameters:** none.

**Behavior:**

Delete content files in `.cache/.content/` whose hash is not
referenced by any chain file in `.cache/.chains/`. Delete chain files
in `.cache/.chains/` whose hash is not referenced by any manifest
entry.

See CACHE.md for details on the cache structure.

---

## accept

Accept an artifact or verdict without regenerating it.
Updates the manifest entry to match the current state:
checksum from the file on disk, chain hash from the
current spec tree.

**Parameters:**

- `logical_name` (string, required) — the logical name
  of the artifact or verdict to accept:
  `ARTIFACT/<name>` or `VERDICT/<name>`.

**Behavior:**

1. Compute the hash of the file on disk and the current
   chain hash from the spec tree.
2. Compare both against the manifest entry. If the
   manifest entry exists, both already match, and — for
   `VERDICT/` entries — `result` is already `accepted`,
   the entry is up to date: return an error.
3. Update the manifest entry's checksum and chain hash
   to match the current values. For `VERDICT/` entries,
   set `result:accepted`.

This handles four cases:
- **Modified** (checksum mismatch): the file was edited
  outside the framework.
- **Stale** (chain hash mismatch): the spec changed but
  the content is still correct.
- **Modified and stale**: both changed.
- **Failed verdict** (`VERDICT/` entries): the human
  overrides the verdict — `result` becomes `accepted`
  even when the hashes are current.

---

## dump_chain

Save the spec chain for a given node to a file for inspection.

**Parameters:**

- `logical_name` (string, required) — the logical name of the target
  node. The node must declare `type`.

**Behavior:**

Assemble the spec chain exactly as `load_chain` would, and write it to
a file under `code-from-spec/.dump/`, named after the logical name with
path separators replaced by underscores (e.g.,
`code-from-spec/.dump/SPEC_golang_implementation_chain_hash.xml`). Each
node dumps to its own file, so multiple dumps do not overwrite each
other. This produces the same document the generation subagent would
receive, allowing the orchestrator or the human to inspect it.

---

## prune_orphans

Remove orphan manifest entries and their artifact files.

**Parameters:** none.

**Behavior:**

1. Scan the spec tree and identify manifest entries whose corresponding
   spec node no longer exists or no longer declares `type`.
2. For each orphan entry, delete the artifact or verdict file from disk
   first, then remove the entry from the manifest. This order ensures
   that if the file deletion fails, the manifest entry is preserved —
   the orphan remains trackable.
3. If the file does not exist on disk, remove the manifest entry
   directly.
4. If the file exists but cannot be deleted, skip that entry —
   do not remove it from the manifest.

Returns a report of every entry pruned. See MANIFEST.md ("Artifact and
verdict status") for the orphan status reported by `validate_specs`.

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
