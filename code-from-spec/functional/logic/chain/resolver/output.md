<!-- code-from-spec: ROOT/functional/logic/chain/resolver@OsgJFsBgFzznudWeW-QgI1fqCdY -->

# Chain Resolver

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

---

### ChainResolve

```
function ChainResolve(target_logical_name: string) -> Chain
  errors:
    - UnreadableFrontmatter: a node's frontmatter cannot be parsed.
    - UnresolvableArtifact: an ARTIFACT/ reference's output id does
        not match any declared output, or an ARTIFACT/ reference is
        missing its qualifier, or the reference uses an unrecognized
        prefix.
    - (LogicalNames.*): propagated from LogicalNameToPath,
        LogicalNameGetParent.
    - (Frontmatter.*): propagated from FrontmatterParse.
```

#### Step 1 — Resolve ancestors and target

  1. If target_logical_name equals "ROOT":
       a. Resolve file path using LogicalNameToPath("ROOT").
          If it fails, propagate the error.
       b. Create a ChainItem with logical_name "ROOT", the
          resolved file path, and qualifier absent.
       c. Set ancestors to an empty list.
       d. Set target to that ChainItem.
       e. Continue to Step 2.

  2. Otherwise:
       a. Create a new list called collected_names.
       b. Add target_logical_name to collected_names.
       c. Set current to target_logical_name.
       d. Repeat:
            i.  Call LogicalNameGetParent(current).
                If it fails, propagate the error.
            ii. Add the returned parent name to collected_names.
            iii. If the returned parent name equals "ROOT", stop.
            iv. Otherwise, set current to the returned parent
                name and continue repeating.
       e. Sort collected_names alphabetically. This produces
          root-first order (e.g. "ROOT", "ROOT/a", "ROOT/a/b").
       f. For each name in sorted collected_names:
            i.  Strip any qualifier using LogicalNameStripQualifier.
            ii. Resolve the stripped name to a file path using
                LogicalNameToPath. If it fails, propagate the error.
            iii. Create a ChainItem with the stripped logical name,
                 the resolved file path, and qualifier absent.
       g. The last ChainItem in the resulting list is the target.
       h. The remaining ChainItems (all but the last) form the
          ancestors list.

#### Step 2 — Resolve dependencies

  1. Call FrontmatterParse(target.file_path).
     If parsing fails, raise error "UnreadableFrontmatter".
     Store the result as target_frontmatter.

  2. Create an empty list called dependencies.

  3. For each entry in target_frontmatter.depends_on:
       a. If entry starts with "ROOT/":
            i.   Extract qualifier using LogicalNameGetQualifier(entry).
                 Store as dep_qualifier (absent if none).
            ii.  Strip qualifier using LogicalNameStripQualifier(entry).
                 Store as bare_name.
            iii. Resolve bare_name to a file path using
                 LogicalNameToPath(bare_name).
                 If it fails, propagate the error.
            iv.  Create a ChainItem with bare_name, the resolved file
                 path, and dep_qualifier.
            v.   Add the ChainItem to dependencies.

       b. Else if entry starts with "ARTIFACT/":
            i.   Extract qualifier (artifact id) using
                 LogicalNameGetQualifier(entry).
                 If qualifier is absent, raise error "UnresolvableArtifact".
            ii.  Derive the generating node's logical name using
                 LogicalNameGetArtifactGenerator(entry).
                 If it fails, propagate the error.
                 Store as generator_name.
            iii. Resolve generator_name to a file path using
                 LogicalNameToPath(generator_name).
                 If it fails, propagate the error.
                 Store as generator_path.
            iv.  Call FrontmatterParse(generator_path).
                 If parsing fails, raise error "UnreadableFrontmatter".
                 Store result as generator_frontmatter.
            v.   Search generator_frontmatter.outputs for an entry
                 whose id equals the qualifier.
                 If no match is found, raise error "UnresolvableArtifact".
                 Store the matching output's path as artifact_path.
            vi.  Create a PathCfs from artifact_path.
            vii. Create a ChainItem with logical_name equal to entry
                 (the original "ARTIFACT/..." string), the artifact
                 file path, and the qualifier.
            viii. Add the ChainItem to dependencies.

       c. Else (neither "ROOT/" nor "ARTIFACT/"):
            Raise error "UnresolvableArtifact".

  4. Sort dependencies alphabetically:
       - Primary sort key: file_path value (ascending).
       - Secondary sort key: qualifier (absent sorts before present;
         when both present, sort alphabetically).

#### Step 3 — Deduplicate dependencies

  1. Create an empty list called deduped.

  2. For each item in dependencies (in current sorted order):
       a. If LogicalNameIsArtifact(item.logical_name) is true
          (ARTIFACT/ entry):
            - An entry is a duplicate if deduped already contains
              an item with the exact same logical_name (including
              qualifier). If so, skip. Otherwise add to deduped.

       b. If LogicalNameIsArtifact(item.logical_name) is false
          (ROOT/ entry):
            - Check if deduped already contains an item with the
              same file_path and the same qualifier. If so, skip.
            - Check if deduped already contains an item with the
              same file_path and qualifier absent. If so, the
              current item's subsection is already included — skip.
            - Otherwise add item to deduped.
            - After adding, if the new item has qualifier absent,
              remove all previously added items in deduped that
              have the same file_path and a non-absent qualifier,
              because the full section subsumes them.

  3. Replace dependencies with deduped.

#### Step 4 — Collect external

  1. Copy target_frontmatter.external into a list called external.
  2. Sort external alphabetically by path value.
  3. Fragments within each entry retain their original declaration order.

#### Step 5 — Resolve input

  1. If target_frontmatter.input is non-empty:
       a. Extract qualifier (artifact id) using
          LogicalNameGetQualifier(target_frontmatter.input).
          If qualifier is absent, raise error "UnresolvableArtifact".
       b. Derive the generating node's logical name using
          LogicalNameGetArtifactGenerator(target_frontmatter.input).
          If it fails, propagate the error.
          Store as input_generator_name.
       c. Resolve input_generator_name to a file path using
          LogicalNameToPath(input_generator_name).
          If it fails, propagate the error.
          Store as input_generator_path.
       d. Call FrontmatterParse(input_generator_path).
          If parsing fails, raise error "UnreadableFrontmatter".
          Store result as input_generator_frontmatter.
       e. Search input_generator_frontmatter.outputs for an entry
          whose id equals the qualifier.
          If no match is found, raise error "UnresolvableArtifact".
          Store the matching output's path as input_artifact_path.
       f. Create a PathCfs from input_artifact_path.
       g. Create a ChainItem with logical_name equal to
          target_frontmatter.input (the original "ARTIFACT/..."
          string), the input artifact file path, and the qualifier.
       h. Set the chain's input field to that ChainItem.

  2. If target_frontmatter.input is empty:
       Set the chain's input field to absent.

#### Step 6 — Return

  1. Return a Chain record with:
       - ancestors: the ancestors list from Step 1.
       - dependencies: the deduped list from Step 3.
       - external: the sorted list from Step 4.
       - target: the target ChainItem from Step 1.
       - input: the resolved input ChainItem or absent from Step 5.

## Contracts

- The chain is fully resolved — all file paths are derived from
  logical names and frontmatter. Existence is not verified; the
  caller handles missing files.
- File paths are pathutils.PathCfs values (forward slashes).
- No duplicate entries in the dependencies list.
- Ancestors are in root-first order.
- Dependencies are sorted by file path then qualifier.
- External entries are sorted by path.
