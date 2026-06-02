<!-- code-from-spec: ROOT/functional/logic/chain/resolver@6mQ-AhKXvN9XWXHB3I-xfGKp7OM -->

## Namespace

    namespace: chainresolver

## Records

```
record ChainItem
  logical_name: string
  file_path: pathutils.PathCfs
  qualifier: optional string

record Chain
  ancestors: list of ChainItem
  dependencies: list of ChainItem
  external: list of frontmatter.FrontmatterExternal
  target: ChainItem
  input: optional ChainItem
```

## Functions

```
function ChainResolve(target_logical_name: string) -> Chain
  errors:
    - UnreadableFrontmatter: a node's frontmatter cannot be parsed.
    - UnresolvableArtifact: an ARTIFACT/ reference cannot be resolved.
    - (LogicalNames.*): propagated from LogicalNameToPath, LogicalNameGetParent.
    - (Frontmatter.*): propagated from FrontmatterParse.
```

### Step 1 — Resolve ancestors and target

1. If target_logical_name equals "ROOT":
     Resolve file path using LogicalNameToPath("ROOT").
     If it fails, propagate the error.
     Create a ChainItem with logical_name "ROOT", the resolved file path, and qualifier absent.
     Set ancestors to empty list.
     Set target to this ChainItem.
     Proceed to Step 2.

2. Otherwise:
     Initialize name_list to an empty list.
     Add target_logical_name to name_list.
     Set current to target_logical_name.
     Loop:
       Call LogicalNameGetParent(current).
       If it fails with NoParent, stop the loop.
       If it fails with another error, propagate the error.
       Add the parent to name_list.
       Set current to the parent.
       If current equals "ROOT", add "ROOT" to name_list and stop the loop.
     Sort name_list alphabetically.
     For each name in name_list:
       Call LogicalNameToPath(name).
       If it fails, propagate the error.
       Create a ChainItem with logical_name set to name, the resolved file path, and qualifier absent.
     The last item in the resulting list is the target.
     The remaining items form the ancestors list.

### Step 2 — Resolve dependencies

1. Call FrontmatterParse(target.file_path).
   If parsing fails, raise "UnreadableFrontmatter".
   Store result as target_frontmatter.

2. Initialize dependencies to an empty list.

3. For each entry in target_frontmatter.depends_on:
     If entry starts with "ROOT/":
       a. Call LogicalNameGetQualifier(entry) to get qualifier (absent if none).
       b. Call LogicalNameStripQualifier(entry) to get bare_name.
       c. Call LogicalNameToPath(bare_name).
          If it fails, propagate the error.
       d. Create a ChainItem with logical_name bare_name, the resolved file path, and the qualifier.
       e. Add the ChainItem to dependencies.
     Else if entry starts with "ARTIFACT/":
       a. Call LogicalNameGetArtifactGenerator(entry).
          If it fails, propagate the error.
          Store result as generator_name.
       b. Call LogicalNameToPath(generator_name).
          If it fails, propagate the error.
          Store result as generator_path.
       c. Call FrontmatterParse(generator_path).
          If parsing fails, raise "UnreadableFrontmatter".
          Store result as generator_frontmatter.
       d. If generator_frontmatter.output is empty, raise "UnresolvableArtifact".
       e. Create a PathCfs from generator_frontmatter.output.
       f. Create a ChainItem with logical_name entry, file_path from step (e), and qualifier absent.
       g. Add the ChainItem to dependencies.
     Else:
       Raise "UnresolvableArtifact".

4. Sort dependencies alphabetically by file_path.value, then by qualifier
   (absent sorts before present).

### Step 3 — Deduplicate dependencies

1. Initialize deduplicated to an empty list.

2. For each entry in dependencies:
     If LogicalNameIsArtifact(entry.logical_name) is true:
       If no entry with the same logical_name exists in deduplicated:
         Add entry to deduplicated.
     Else:
       If an entry with the same file_path and no qualifier exists in deduplicated:
         Skip this entry (the full section already covers it).
       Else if an entry with the same file_path and the same qualifier exists in deduplicated:
         Skip this entry (exact duplicate).
       Else:
         Add entry to deduplicated.

3. Replace dependencies with deduplicated.

### Step 4 — Collect external

1. Copy target_frontmatter.external into the chain's external list.
2. Sort external entries alphabetically by path.

### Step 5 — Resolve input

1. If target_frontmatter.input is non-empty:
     a. Call LogicalNameGetArtifactGenerator(target_frontmatter.input).
        If it fails, propagate the error.
        Store result as input_generator_name.
     b. Call LogicalNameToPath(input_generator_name).
        If it fails, propagate the error.
        Store result as input_generator_path.
     c. Call FrontmatterParse(input_generator_path).
        If parsing fails, raise "UnreadableFrontmatter".
        Store result as input_generator_frontmatter.
     d. If input_generator_frontmatter.output is empty, raise "UnresolvableArtifact".
     e. Create a PathCfs from input_generator_frontmatter.output.
     f. Create a ChainItem with logical_name target_frontmatter.input, file_path from step (e),
        and qualifier absent.
     g. Set the chain's input field to this ChainItem.

2. If target_frontmatter.input is empty:
     Set the chain's input field to absent.

### Return

Return a Chain with:
  ancestors: the ancestors list (root-first order)
  dependencies: the deduplicated, sorted dependencies list
  external: the sorted external list
  target: the target ChainItem
  input: the resolved input ChainItem, or absent
