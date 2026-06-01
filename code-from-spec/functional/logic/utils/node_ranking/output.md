<!-- code-from-spec: ROOT/functional/logic/utils/node_ranking@pLWNUkgid17GjA_z6uJ_kDVq34o -->

## Records

record NodeRankInput
  logical_name: string
  frontmatter: frontmatter.Frontmatter

record NodeRankEntry
  logical_name: string
  rank: integer

record RankEntry
  logical_name: string
  dependencies: list of strings
  rank: integer

## Functions

function NodeRankCompute(entries: list of NodeRankInput) -> (ranked: list of NodeRankEntry, cycles: list of string)
  errors:
    - UnresolvableReference: a depends_on or input target cannot be resolved.

  1. Build entry map (keyed by logical name):

     Initialize an empty map called entry_map.

     For each item in entries:
       a. Add a RankEntry to entry_map keyed by item.logical_name
          with rank = 0 and dependencies = empty list.

       b. For each output in item.frontmatter.outputs:
            Construct artifact_key by:
              - Taking item.logical_name and stripping the "ROOT/" prefix.
              - Prepending "ARTIFACT/".
              - Appending "(" + output.id + ")".
            Example: "ROOT/a/b" with output.id "foo" -> "ARTIFACT/a/b(foo)".
            Add a RankEntry to entry_map keyed by artifact_key
            with rank = 0 and dependencies = empty list.

  2. Build dependency edges:

     For each item in entries:
       a. If item.logical_name is not "ROOT":
            Call LogicalNameGetParent(item.logical_name) -> parent_name.
            If parent_name is not found in entry_map,
              raise error "UnresolvableReference".
            Add parent_name to entry_map[item.logical_name].dependencies.

       b. For each dep in item.frontmatter.depends_on:
            If dep starts with "ARTIFACT/":
              lookup_key = dep (use as-is).
            Else:
              lookup_key = LogicalNameStripQualifier(dep).
            If lookup_key is not found in entry_map,
              raise error "UnresolvableReference".
            Add lookup_key to entry_map[item.logical_name].dependencies.

       c. If item.frontmatter.input is non-empty:
            If item.frontmatter.input is not found in entry_map,
              raise error "UnresolvableReference".
            Add item.frontmatter.input to entry_map[item.logical_name].dependencies.

     For each artifact_key in entry_map whose key starts with "ARTIFACT/":
       Determine the generating node's logical name by:
         - Stripping the "ARTIFACT/" prefix from artifact_key.
         - Stripping any qualifier using LogicalNameStripQualifier.
         - Prepending "ROOT/".
       If the generating node logical name is not found in entry_map,
         raise error "UnresolvableReference".
       Add the generating node logical name to entry_map[artifact_key].dependencies.

  3. Initialize ranks:

     Set entry_map["ROOT"].rank = 0.
     For all other entries, rank is already 0 from Step 1.

  4. Iterate and detect cycles:

     Let N = total number of entries in entry_map.
     Initialize changed_in_last_pass = empty list.

     Repeat up to N times (loop index i from 1 to N):
       Set changed_this_pass = empty list.

       For each entry in entry_map (excluding "ROOT"):
         Compute max_dep_rank = maximum of entry_map[dep].rank
           for each dep in entry.dependencies.
         Compute new_rank = 1 + max_dep_rank.
         If new_rank > entry.rank:
           Set entry.rank = new_rank.
           Add entry.logical_name to changed_this_pass.

       If changed_this_pass is empty:
         Stop iterating (converged, no cycles).
       Else:
         Set changed_in_last_pass = changed_this_pass.

       If i equals N and changed_this_pass is not empty:
         Cycles detected. changed_in_last_pass contains the cycle participants.

  5. Output:

     Build result_entries as a list of NodeRankEntry
       by collecting all entries from entry_map as (logical_name, rank).

     Sort result_entries:
       - Primary: rank ascending.
       - Secondary: logical_name ascending.

     If cycles were detected, set cycle_list = changed_in_last_pass.
     Else set cycle_list = empty list.

     Return (ranked = result_entries, cycles = cycle_list).
