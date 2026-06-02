<!-- code-from-spec: ROOT/functional/logic/chain/resolver@PSK7DCdLeV7LgXTk8S3OCsjZzoQ -->

## Namespace

    namespace: chainresolver

## Records

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

## Functions

function ChainResolve(target_logical_name: string) -> Chain
  errors:
    - UnreadableFrontmatter: a node's frontmatter cannot be parsed.
    - UnresolvableArtifact: an ARTIFACT/ reference cannot be resolved.
    - (LogicalNames.*): propagated from LogicalNameToPath, LogicalNameGetParent.
    - (Frontmatter.*): propagated from FrontmatterParse.

  1. Resolve ancestors and target.

     If target_logical_name is "ROOT":
       Call LogicalNameToPath with "ROOT". If it fails, propagate the error.
       Create a ChainItem with logical_name "ROOT", the resolved file path, and qualifier absent.
       Set ancestors to an empty list and target to this ChainItem.
       Skip to step 2.

     Otherwise:
       Create a collection starting with target_logical_name.
       Call LogicalNameGetParent on target_logical_name to get the parent.
       If it fails, propagate the error.
       Add the parent to the collection.
       Repeat calling LogicalNameGetParent on the most recent addition until "ROOT" is reached (inclusive).
       If LogicalNameGetParent fails at any point, propagate the error.

       Sort the collected logical names alphabetically. This produces root-first order.

       For each name in the sorted list:
         Call LogicalNameToPath on the name. If it fails, propagate the error.
         Create a ChainItem with that logical name, the resolved file path, and qualifier absent.

       The last ChainItem in the list is the target.
       The remaining ChainItems form the ancestors list.

  2. Resolve dependencies.

     Call FrontmatterParse with the target's file path.
     If parsing fails, raise error "unreadable frontmatter".

     For each entry in frontmatter.depends_on:

       If the entry does not start with "ROOT/" and does not start with "ARTIFACT/":
         Raise error "unresolvable artifact".

       If the entry starts with "ROOT/":
         Call LogicalNameGetQualifier on the entry to extract the qualifier (absent if none).
         Call LogicalNameStripQualifier on the entry to get the bare logical name.
         Call LogicalNameToPath on the bare logical name. If it fails, propagate the error.
         Create a ChainItem with the bare logical name, the resolved file path, and the qualifier.

       If the entry starts with "ARTIFACT/":
         Call LogicalNameGetArtifactGenerator on the entry. If it fails, propagate the error.
         Call LogicalNameToPath on the generating node's logical name. If it fails, propagate the error.
         Call FrontmatterParse on the generating node's file path.
         If parsing fails, raise error "unreadable frontmatter".
         The output path is the generating node's frontmatter.output value, used as a PathCfs.
         Create a ChainItem with the ARTIFACT/ logical name, that output path, and qualifier absent.

     Sort the resulting dependency ChainItems alphabetically by file_path value,
     then by qualifier (absent sorts before present).

  3. Deduplicate dependencies.

     For each pair of entries in the sorted dependencies list:

       For ROOT/ entries (determined using LogicalNameIsArtifact returning false):
         Two entries are duplicates if they have the same file_path and the same qualifier.
         If an entry exists with a given file_path and qualifier absent, any other entry
         with the same file_path and a qualifier present is redundant — remove it.
         Keep the first occurrence when removing duplicates.

       For ARTIFACT/ entries (LogicalNameIsArtifact returns true):
         Two entries are duplicates if they have the same logical_name.
         Keep the first occurrence when removing duplicates.

  4. Collect external.

     Copy the external list from the target's frontmatter into the chain.
     Sort entries alphabetically by path.

  5. Resolve input.

     If frontmatter.input is non-empty:
       Call LogicalNameGetArtifactGenerator on frontmatter.input. If it fails, propagate the error.
       Call LogicalNameToPath on the generating node's logical name. If it fails, propagate the error.
       Call FrontmatterParse on the generating node's file path.
       If parsing fails, raise error "unreadable frontmatter".
       The output path is the generating node's frontmatter.output value, used as a PathCfs.
       Create a ChainItem with the ARTIFACT/ logical name, that output path, and qualifier absent.
       Set the chain's input to this ChainItem.

     If frontmatter.input is empty:
       Set the chain's input to absent.

  6. Return the Chain with:
     - ancestors: the ancestors list (root-first order)
     - dependencies: the deduplicated, sorted dependencies list
     - external: the sorted external list
     - target: the target ChainItem
     - input: the resolved input ChainItem or absent
