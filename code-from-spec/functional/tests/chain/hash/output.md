<!-- code-from-spec: ROOT/functional/tests/chain/hash@2_rqZ3hgkmCSWQT2xUGvKhXDzn0 -->

## Properties

### Hash is deterministic

Setup:
  Create a spec node file for ROOT (_node.md) with `# Public` content.
  Create a spec node file for ROOT/a (_node.md) with `# Public` content.
  Build a Chain directly:
    ancestors = [ChainItem(logical_name="ROOT", file_path=<root_node_path>, qualifier=absent)]
    dependencies = []
    external = []
    target = ChainItem(logical_name="ROOT/a", file_path=<root_a_node_path>, qualifier=absent)
    input = absent

Actions:
  Call ChainHashCompute with the Chain. Record result as hash1.
  Call ChainHashCompute again with the same Chain. Record result as hash2.

Expected outcome:
  hash1 equals hash2.

---

### Hash is 27 characters

Setup:
  Create a spec node file for ROOT (_node.md) with `# Public` content.
  Build a Chain directly:
    ancestors = []
    dependencies = []
    external = []
    target = ChainItem(logical_name="ROOT", file_path=<root_node_path>, qualifier=absent)
    input = absent

Actions:
  Call ChainHashCompute with the Chain.

Expected outcome:
  The result is exactly 27 characters long.

---

### Hash changes when ancestor content changes

Setup:
  Create a spec node file for ROOT (_node.md) with `# Public\n\nOriginal content.`.
  Create a spec node file for ROOT/a (_node.md) with `# Public\n\nTarget.`.
  Build a Chain directly:
    ancestors = [ChainItem(logical_name="ROOT", file_path=<root_node_path>, qualifier=absent)]
    dependencies = []
    external = []
    target = ChainItem(logical_name="ROOT/a", file_path=<root_a_node_path>, qualifier=absent)
    input = absent

Actions:
  Call ChainHashCompute with the Chain. Record result as hash1.
  Overwrite ROOT's _node.md, replacing `# Public` content with `# Public\n\nModified content.`.
  Call ChainHashCompute again with the same Chain. Record result as hash2.

Expected outcome:
  hash1 does not equal hash2.

---

### Hash changes when dependency content changes

Setup:
  Create a spec node file for ROOT (_node.md) with `# Public` content.
  Create a spec node file for ROOT/b (_node.md) with `# Public\n\nDependency original.`.
  Create a spec node file for ROOT/a (_node.md) with `# Public` content.
  Build a Chain directly:
    ancestors = [ChainItem(logical_name="ROOT", file_path=<root_node_path>, qualifier=absent)]
    dependencies = [ChainItem(logical_name="ROOT/b", file_path=<root_b_node_path>, qualifier=absent)]
    external = []
    target = ChainItem(logical_name="ROOT/a", file_path=<root_a_node_path>, qualifier=absent)
    input = absent

Actions:
  Call ChainHashCompute with the Chain. Record result as hash1.
  Overwrite ROOT/b's _node.md with `# Public\n\nDependency modified.`.
  Call ChainHashCompute again with the same Chain. Record result as hash2.

Expected outcome:
  hash1 does not equal hash2.

---

### Hash changes when target Public changes

Setup:
  Create a spec node file for ROOT (_node.md) with `# Public` content.
  Create a spec node file for ROOT/a (_node.md) with `# Public\n\nTarget original.`.
  Build a Chain directly:
    ancestors = [ChainItem(logical_name="ROOT", file_path=<root_node_path>, qualifier=absent)]
    dependencies = []
    external = []
    target = ChainItem(logical_name="ROOT/a", file_path=<root_a_node_path>, qualifier=absent)
    input = absent

Actions:
  Call ChainHashCompute with the Chain. Record result as hash1.
  Overwrite ROOT/a's _node.md with `# Public\n\nTarget modified.`.
  Call ChainHashCompute again with the same Chain. Record result as hash2.

Expected outcome:
  hash1 does not equal hash2.

---

### Hash changes when target Agent changes

Setup:
  Create a spec node file for ROOT (_node.md) with `# Public` content.
  Create a spec node file for ROOT/a (_node.md) with `# Public\n\nSome public.\n\n# Agent\n\nOriginal agent.`.
  Build a Chain directly:
    ancestors = [ChainItem(logical_name="ROOT", file_path=<root_node_path>, qualifier=absent)]
    dependencies = []
    external = []
    target = ChainItem(logical_name="ROOT/a", file_path=<root_a_node_path>, qualifier=absent)
    input = absent

Actions:
  Call ChainHashCompute with the Chain. Record result as hash1.
  Overwrite ROOT/a's _node.md with `# Public\n\nSome public.\n\n# Agent\n\nModified agent.`.
  Call ChainHashCompute again with the same Chain. Record result as hash2.

Expected outcome:
  hash1 does not equal hash2.

---

## Ancestors

### Ancestor with Public section contributes hash

Setup:
  Create a spec node file for ROOT (_node.md) with `# Public\n\nSome content.`.
  Create a spec node file for ROOT/a (_node.md) with `# Public\n\nTarget.`.
  Build a Chain directly:
    ancestors = [ChainItem(logical_name="ROOT", file_path=<root_node_path>, qualifier=absent)]
    dependencies = []
    external = []
    target = ChainItem(logical_name="ROOT/a", file_path=<root_a_node_path>, qualifier=absent)
    input = absent

Actions:
  Call ChainHashCompute with the Chain.

Expected outcome:
  No error. Result is exactly 27 characters long and non-empty.

---

### Ancestor without Public section — skipped

Setup:
  Create a spec node file for ROOT (_node.md) with content that has no `# Public` section
  (e.g., only a name heading and a description).
  Create a spec node file for ROOT/a (_node.md) with content that has no `# Public` section.
  Build Chain-A directly:
    ancestors = [ChainItem(logical_name="ROOT", file_path=<root_node_path>, qualifier=absent)]
    dependencies = []
    external = []
    target = ChainItem(logical_name="ROOT/a", file_path=<root_a_node_path>, qualifier=absent)
    input = absent
  Create a second spec node file for ROOT2 (_node.md) with `# Public\n\nSome content.`.
  Create a spec node file for ROOT2/a (_node.md) with `# Public\n\nTarget.`.
  Build Chain-B directly:
    ancestors = [ChainItem(logical_name="ROOT2", file_path=<root2_node_path>, qualifier=absent)]
    dependencies = []
    external = []
    target = ChainItem(logical_name="ROOT2/a", file_path=<root2_a_node_path>, qualifier=absent)
    input = absent

Actions:
  Call ChainHashCompute with Chain-A. Record result as hashA.
  Call ChainHashCompute with Chain-B. Record result as hashB.

Expected outcome:
  hashA does not equal hashB.

---

### Multiple ancestors — order matters

Setup:
  Create a spec node file for ROOT (_node.md) with `# Public\n\nRoot public.`.
  Create a spec node file for ROOT/a (_node.md) with `# Public\n\nIntermediate public.`.
  Create a spec node file for ROOT/a/b (_node.md) with `# Public\n\nTarget.`.
  Build Chain-Forward directly:
    ancestors = [
      ChainItem(logical_name="ROOT", file_path=<root_node_path>, qualifier=absent),
      ChainItem(logical_name="ROOT/a", file_path=<root_a_node_path>, qualifier=absent)
    ]
    dependencies = []
    external = []
    target = ChainItem(logical_name="ROOT/a/b", file_path=<root_a_b_node_path>, qualifier=absent)
    input = absent
  Build Chain-Swapped directly with ancestors in reversed order:
    ancestors = [
      ChainItem(logical_name="ROOT/a", file_path=<root_a_node_path>, qualifier=absent),
      ChainItem(logical_name="ROOT", file_path=<root_node_path>, qualifier=absent)
    ]
    dependencies = []
    external = []
    target = ChainItem(logical_name="ROOT/a/b", file_path=<root_a_b_node_path>, qualifier=absent)
    input = absent

Actions:
  Call ChainHashCompute with Chain-Forward. Record result as hashForward.
  Call ChainHashCompute with Chain-Swapped. Record result as hashSwapped.

Expected outcome:
  hashForward does not equal hashSwapped.

---

## Dependencies

### ROOT dependency without qualifier — hashes Public

Setup:
  Create a spec node file for ROOT/b (_node.md) with `# Public\n\nOriginal dep public.`.
  Create a spec node file for ROOT/a (_node.md) with `# Public\n\nTarget.`.
  Build a Chain directly:
    ancestors = []
    dependencies = [ChainItem(logical_name="ROOT/b", file_path=<root_b_node_path>, qualifier=absent)]
    external = []
    target = ChainItem(logical_name="ROOT/a", file_path=<root_a_node_path>, qualifier=absent)
    input = absent

Actions:
  Call ChainHashCompute with the Chain. Record result as hash1.
  Overwrite ROOT/b's _node.md with `# Public\n\nModified dep public.`.
  Call ChainHashCompute again with the same Chain. Record result as hash2.

Expected outcome:
  hash1 does not equal hash2.

---

### ROOT dependency with qualifier — hashes subsection

Setup:
  Create a spec node file for ROOT/b (_node.md) with:
    `# Public\n\n## Interface\n\nOriginal interface content.\n\n## Other\n\nOther content.`
  Create a spec node file for ROOT/a (_node.md) with `# Public\n\nTarget.`.
  Build a Chain directly:
    ancestors = []
    dependencies = [ChainItem(logical_name="ROOT/b", file_path=<root_b_node_path>, qualifier="interface")]
    external = []
    target = ChainItem(logical_name="ROOT/a", file_path=<root_a_node_path>, qualifier=absent)
    input = absent

Actions:
  Call ChainHashCompute with the Chain. Record result as hash1.
  Overwrite ROOT/b's _node.md changing the `## Interface` content to `## Interface\n\nModified interface content.`.
  Call ChainHashCompute again with the same Chain. Record result as hash2.

Expected outcome:
  hash1 does not equal hash2.

---

### Qualifier case normalization

Setup:
  Create a spec node file for ROOT/b (_node.md) with:
    `# Public\n\n## Interface\n\nSome interface content.`
  Create a spec node file for ROOT/a (_node.md) with `# Public\n\nTarget.`.
  Build a Chain directly:
    ancestors = []
    dependencies = [ChainItem(logical_name="ROOT/b", file_path=<root_b_node_path>, qualifier="INTERFACE")]
    external = []
    target = ChainItem(logical_name="ROOT/a", file_path=<root_a_node_path>, qualifier=absent)
    input = absent

Actions:
  Call ChainHashCompute with the Chain.

Expected outcome:
  No error. The result is exactly 27 characters long.

---

### ARTIFACT dependency — hashes file minus frontmatter

Setup:
  Create an artifact file with frontmatter followed by body content:
    `---\nsome: value\n---\n\nOriginal body content.`
  Create a spec node file for ROOT/a (_node.md) with `# Public\n\nTarget.`.
  Build a Chain directly:
    ancestors = []
    dependencies = [ChainItem(logical_name="ARTIFACT/some_id", file_path=<artifact_file_path>, qualifier=absent)]
    external = []
    target = ChainItem(logical_name="ROOT/a", file_path=<root_a_node_path>, qualifier=absent)
    input = absent

Actions:
  Call ChainHashCompute with the Chain. Record result as hash1.
  Overwrite the artifact file with `---\nsome: value\n---\n\nModified body content.`.
  Call ChainHashCompute again with the same Chain. Record result as hash2.

Expected outcome:
  hash1 does not equal hash2.

---

### ARTIFACT dependency — frontmatter change ignored

Setup:
  Create an artifact file with frontmatter followed by body content:
    `---\nsome: value\n---\n\nBody content.`
  Create a spec node file for ROOT/a (_node.md) with `# Public\n\nTarget.`.
  Build a Chain directly:
    ancestors = []
    dependencies = [ChainItem(logical_name="ARTIFACT/some_id", file_path=<artifact_file_path>, qualifier=absent)]
    external = []
    target = ChainItem(logical_name="ROOT/a", file_path=<root_a_node_path>, qualifier=absent)
    input = absent

Actions:
  Call ChainHashCompute with the Chain. Record result as hash1.
  Overwrite the artifact file with `---\nother: changed\n---\n\nBody content.`.
  Call ChainHashCompute again with the same Chain. Record result as hash2.

Expected outcome:
  hash1 equals hash2.

---

## External files

### External file — hashes all content

Setup:
  Create an external file with content `Original external content.`.
  Create a spec node file for ROOT/a (_node.md) with `# Public\n\nTarget.`.
  Build a Chain directly:
    ancestors = []
    dependencies = []
    external = [FrontmatterExternal(path=<external_file_path>)]
    target = ChainItem(logical_name="ROOT/a", file_path=<root_a_node_path>, qualifier=absent)
    input = absent

Actions:
  Call ChainHashCompute with the Chain. Record result as hash1.
  Overwrite the external file with `Modified external content.`.
  Call ChainHashCompute again with the same Chain. Record result as hash2.

Expected outcome:
  hash1 does not equal hash2.

---

## Target

### Target Public and Agent both contribute

Setup:
  Create a spec node file for ROOT/a (_node.md) with:
    `# Public\n\nPublic content.\n\n# Agent\n\nAgent content.`
  Build a Chain directly:
    ancestors = []
    dependencies = []
    external = []
    target = ChainItem(logical_name="ROOT/a", file_path=<root_a_node_path>, qualifier=absent)
    input = absent

Actions:
  Call ChainHashCompute with the Chain. Record result as hash1.
  Overwrite ROOT/a's _node.md with `# Public\n\nPublic content.` (removing `# Agent`).
  Call ChainHashCompute again with the same Chain. Record result as hash2.

Expected outcome:
  hash1 does not equal hash2.

---

### Target without Agent — Agent skipped

Setup:
  Create a spec node file for ROOT/a (_node.md) with `# Public\n\nPublic only content.` and no `# Agent` section.
  Build a Chain directly:
    ancestors = []
    dependencies = []
    external = []
    target = ChainItem(logical_name="ROOT/a", file_path=<root_a_node_path>, qualifier=absent)
    input = absent

Actions:
  Call ChainHashCompute with the Chain.

Expected outcome:
  No error. Result is exactly 27 characters long.

---

## Input

### Input hashes file minus frontmatter

Setup:
  Create an artifact file with frontmatter followed by body content:
    `---\nsome: value\n---\n\nOriginal input body.`
  Create a spec node file for ROOT/a (_node.md) with `# Public\n\nTarget.`.
  Build a Chain directly:
    ancestors = []
    dependencies = []
    external = []
    target = ChainItem(logical_name="ROOT/a", file_path=<root_a_node_path>, qualifier=absent)
    input = ChainItem(logical_name="ARTIFACT/input_id", file_path=<artifact_file_path>, qualifier=absent)

Actions:
  Call ChainHashCompute with the Chain. Record result as hash1.
  Overwrite the artifact file with `---\nsome: value\n---\n\nModified input body.`.
  Call ChainHashCompute again with the same Chain. Record result as hash2.

Expected outcome:
  hash1 does not equal hash2.

---

### No input — skipped

Setup:
  Create a spec node file for ROOT/a (_node.md) with `# Public\n\nTarget.`.
  Build a Chain directly:
    ancestors = []
    dependencies = []
    external = []
    target = ChainItem(logical_name="ROOT/a", file_path=<root_a_node_path>, qualifier=absent)
    input = absent

Actions:
  Call ChainHashCompute with the Chain.

Expected outcome:
  No error. Result is exactly 27 characters long.

---

## Error cases

### Unreadable spec node file

Setup:
  Build a Chain directly referencing a spec node file that does not exist on disk:
    ancestors = []
    dependencies = []
    external = []
    target = ChainItem(logical_name="ROOT/missing", file_path=<nonexistent_path>, qualifier=absent)
    input = absent

Actions:
  Call ChainHashCompute with the Chain.

Expected outcome:
  Error ParseFailure is returned.

---

### Unreadable artifact file

Setup:
  Build a Chain directly with an ARTIFACT dependency pointing to a file that does not exist on disk:
    ancestors = []
    dependencies = [ChainItem(logical_name="ARTIFACT/missing", file_path=<nonexistent_path>, qualifier=absent)]
    external = []
    target = ChainItem(logical_name="ROOT/a", file_path=<root_a_node_path>, qualifier=absent)
    input = absent
  (ROOT/a node file exists on disk with `# Public` content.)

Actions:
  Call ChainHashCompute with the Chain.

Expected outcome:
  Error FileUnreadable is returned.

---

### Unreadable external file

Setup:
  Create a spec node file for ROOT/a (_node.md) with `# Public\n\nTarget.`.
  Build a Chain directly with an external entry pointing to a file that does not exist on disk:
    ancestors = []
    dependencies = []
    external = [FrontmatterExternal(path=<nonexistent_path>)]
    target = ChainItem(logical_name="ROOT/a", file_path=<root_a_node_path>, qualifier=absent)
    input = absent

Actions:
  Call ChainHashCompute with the Chain.

Expected outcome:
  Error FileUnreadable is returned.
