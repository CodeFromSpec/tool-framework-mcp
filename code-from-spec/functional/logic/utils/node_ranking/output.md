<!-- code-from-spec: ROOT/functional/logic/utils/node_ranking@1LHYK3wAj0SBRJf7rXOeXM3a-no -->

## Records

record NodeRankInput
  logical_name: string
  frontmatter: frontmatter.Frontmatter

record NodeRankEntry
  logical_name: string
  rank: integer

## Functions

function NodeRankCompute(entries: list of NodeRankInput) -> (ranked: list of NodeRankEntry, cycles: list of string)
  errors:
    - UnresolvableReference: a depends_on or input target cannot be resolved.

  1. Build entry map.

     Create an empty map keyed by logical name. Each map value tracks:
       - dependency_keys: list of strings (logical names this entry depends on)
       - rank: integer (current rank, initialized in step 3)

     For each NodeRankInput in entries:
       Add a spec node entry keyed by the input's logical_name.
       If frontmatter.output is non-empty:
         Construct the artifact logical name: strip "ROOT/" prefix from logical_name, prepend "ARTIFACT/".
         Add an artifact entry keyed by the artifact logical name.

  2. Build dependency edges.

     For each spec node entry:
       If the entry's logical_name is "ROOT":
         It has no dependencies. Leave its dependency_keys empty.
       Else:
         Call LogicalNameGetParent on the entry's logical_name to get the parent logical name.
         Add the parent logical name to this entry's dependency_keys.

       For each item in frontmatter.depends_on:
         If the item starts with "ARTIFACT/":
           Use the item as-is as the dependency key.
         Else (it is a ROOT/ reference):
           Call LogicalNameStripQualifier on the item to get the bare logical name.
           Use the bare logical name as the dependency key.
         Add the key to this entry's dependency_keys.

       If frontmatter.input is non-empty:
         Add frontmatter.input as-is to this entry's dependency_keys (it is an ARTIFACT/ reference).

     For each artifact entry:
       The artifact depends on the spec node that generates it.
       Derive the generating node's logical name: strip "ARTIFACT/" prefix, prepend "ROOT/".
       Add that logical name to the artifact entry's dependency_keys.

     For each entry (both spec nodes and artifacts), verify all dependency_keys exist in the entry map.
     If any key is not found, raise error "unresolvable reference".

  3. Initialize ranks.

     Set rank of "ROOT" to 0.
     Set rank of all other entries to 0 as the initial value.

  4. Iterate and detect cycles.

     Let N = total number of entries in the map.

     Repeat up to N times:
       Set changed = false for this pass.
       For each entry excluding "ROOT":
         Compute candidate_rank = 1 + max(rank of each entry in dependency_keys).
         If candidate_rank is greater than the entry's current rank:
           Update the entry's rank to candidate_rank.
           Set changed = true.
           Mark this entry as updated in this pass.
       If changed is false:
         Stop iterating (converged, no cycles).

     If the loop completes all N passes and changed is still true after the final pass:
       A cycle exists.
       Collect the logical names of entries whose rank changed in the last pass.
       These are the cycle participants to report.

  5. Output.

     Build the ranked list: for each entry in the map, create a NodeRankEntry with
     the entry's logical_name and final rank.

     Sort the ranked list by rank ascending, then by logical_name ascending.

     Return:
       ranked: the sorted list of NodeRankEntry
       cycles: the list of logical names from cycle participants (empty if no cycle)
