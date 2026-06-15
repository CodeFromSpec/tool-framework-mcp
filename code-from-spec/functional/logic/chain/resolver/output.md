<!-- code-from-spec: SPEC/functional/logic/chain/resolver@IoM0RbJaIR89y1FKHWluHRGg5yw -->

namespace: chainresolver

---

record ChainItem
  unqualified_logical_name: string
  file_path: pathutils.PathCfs
  qualifier: optional string

record Chain
  ancestors: list of ChainItem
  dependencies: list of ChainItem
  target: ChainItem
  input: optional ChainItem

---

function ChainResolve(target_logical_name: string) -> Chain
  errors:
    - UnreadableFrontmatter
    - UnresolvableArtifact
    - (LogicalNames.*): propagated from LogicalNameToPath, LogicalNameGetParent
    - (Frontmatter.*): propagated from FrontmatterParse

  1. Resolve ancestors and target.

     If target_logical_name is "SPEC":
       Resolve the file path by calling LogicalNameToPath(target_logical_name).
       If LogicalNameToPath fails, propagate the error.
       Create a ChainItem:
         unqualified_logical_name = target_logical_name
         file_path = resolved path
         qualifier = absent
       Set ancestors to an empty list.
       Set target to that ChainItem.
       Skip to step 2.

     Otherwise:
       Initialize a name list containing target_logical_name.
       Set current_name = target_logical_name.
       Loop:
         Call LogicalNameGetParent(current_name).
         If LogicalNameGetParent fails, propagate the error.
         Add the returned parent name to the name list.
         Set current_name = parent name.
         If current_name is "SPEC", stop the loop.

       Sort the name list alphabetically.
       This produces root-first order (e.g. "SPEC", "SPEC/a", "SPEC/a/b").

       For each name in the sorted name list:
         Call LogicalNameToPath(name).
         If LogicalNameToPath fails, propagate the error.
         Create a ChainItem:
           unqualified_logical_name = name
           file_path = resolved path
           qualifier = absent

       The last item in the sorted list becomes the target ChainItem.
       All preceding items form the ancestors list.

  2. Resolve dependencies.

     Call FrontmatterParse(target.file_path).
     If FrontmatterParse fails, raise error "UnreadableFrontmatter".

     Initialize an empty dependency list.

     For each entry in frontmatter.depends_on:

       If LogicalNameIsSpec(entry) is true:
         Call LogicalNameGetQualifier(entry) to get the qualifier (absent if none).
         Call LogicalNameStripQualifier(entry) to get the bare logical name.
         Call LogicalNameToPath(bare logical name).
         If LogicalNameToPath fails, propagate the error.
         Create a ChainItem:
           unqualified_logical_name = bare logical name
           file_path = resolved path
           qualifier = extracted qualifier (absent if none)
         Add the ChainItem to the dependency list.

       Else if LogicalNameIsArtifact(entry) is true:
         Call LogicalNameGetArtifactGenerator(entry) to get the generating node's logical name.
         If LogicalNameGetArtifactGenerator fails, propagate the error.
         Call LogicalNameToPath(generating node's logical name).
         If LogicalNameToPath fails, propagate the error.
         Call FrontmatterParse(generating node's file path).
         If FrontmatterParse fails, raise error "UnreadableFrontmatter".
         If generating node's frontmatter.output is empty,
           raise error "UnresolvableArtifact".
         Create a ChainItem:
           unqualified_logical_name = entry (the ARTIFACT/ logical name as-is)
           file_path = generating node's frontmatter.output as PathCfs
           qualifier = absent
         Add the ChainItem to the dependency list.

       Else if LogicalNameIsExternal(entry) is true:
         Call LogicalNameExternalToPath(entry).
         If LogicalNameExternalToPath fails, propagate the error.
         Create a ChainItem:
           unqualified_logical_name = entry (the EXTERNAL/ logical name as-is)
           file_path = resolved path
           qualifier = absent
         Add the ChainItem to the dependency list.

       Else:
         raise error "UnresolvableArtifact".

     Sort the dependency list alphabetically by unqualified_logical_name,
     then by qualifier (absent sorts before present), in a single pass.

  3. Deduplicate dependencies.

     Initialize an empty seen-entries tracking structure.
     Initialize an empty deduplicated dependency list.

     For each entry in the sorted dependency list:

       If LogicalNameIsSpec(entry.unqualified_logical_name) is true:
         Check if an entry with the same unqualified_logical_name and the same qualifier
         already exists in the deduplicated list.
         If yes, skip this entry (duplicate).
         Also check if an entry with the same unqualified_logical_name and no qualifier
         already exists in the deduplicated list.
         If yes, skip this entry (the full section covers every subsection).
         Otherwise, add this entry to the deduplicated list.

       Else if LogicalNameIsArtifact(entry.unqualified_logical_name) is true:
         Check if an entry with the same unqualified_logical_name already exists
         in the deduplicated list.
         If yes, skip this entry (duplicate).
         Otherwise, add this entry to the deduplicated list.

       Else if LogicalNameIsExternal(entry.unqualified_logical_name) is true:
         Check if an entry with the same unqualified_logical_name already exists
         in the deduplicated list.
         If yes, skip this entry (duplicate).
         Otherwise, add this entry to the deduplicated list.

     Replace the dependency list with the deduplicated list.

  4. Resolve input.

     If frontmatter.input is empty:
       Set the Chain's input field to absent.
     Else:
       Set input_entry = frontmatter.input.

       If LogicalNameIsArtifact(input_entry) is true:
         Call LogicalNameGetArtifactGenerator(input_entry) to get the generating node's logical name.
         If LogicalNameGetArtifactGenerator fails, propagate the error.
         Call LogicalNameToPath(generating node's logical name).
         If LogicalNameToPath fails, propagate the error.
         Call FrontmatterParse(generating node's file path).
         If FrontmatterParse fails, raise error "UnreadableFrontmatter".
         If generating node's frontmatter.output is empty,
           raise error "UnresolvableArtifact".
         Create a ChainItem:
           unqualified_logical_name = input_entry (the ARTIFACT/ logical name as-is)
           file_path = generating node's frontmatter.output as PathCfs
           qualifier = absent
         Set the Chain's input field to that ChainItem.

       Else if LogicalNameIsExternal(input_entry) is true:
         Call LogicalNameExternalToPath(input_entry).
         If LogicalNameExternalToPath fails, propagate the error.
         Create a ChainItem:
           unqualified_logical_name = input_entry (the EXTERNAL/ logical name as-is)
           file_path = resolved path
           qualifier = absent
         Set the Chain's input field to that ChainItem.

  5. Return a Chain:
       ancestors = ancestors list (root-first, qualifier absent for all)
       dependencies = deduplicated dependency list (sorted)
       target = target ChainItem
       input = resolved input ChainItem or absent
