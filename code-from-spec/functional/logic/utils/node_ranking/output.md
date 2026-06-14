<!-- code-from-spec: ROOT/functional/logic/utils/node_ranking@i7UwY37_ciHXtGeMUf5LLnTtXEk -->

namespace: noderanking

record NodeRankInput
  logical_name: string
  frontmatter: frontmatter.Frontmatter

record NodeRankEntry
  logical_name: string
  rank: integer

record RankingEntry
  logical_name: string
  dependencies: list of strings
  rank: integer

function NodeRankCompute(entries: list of NodeRankInput) -> (ranked: list of NodeRankEntry, cycles: list of string)

  1. Build entry map.

     Create an empty map from logical_name to RankingEntry.

     For each NodeRankInput in entries:
       Add a RankingEntry with:
         logical_name = NodeRankInput.logical_name
         dependencies = empty list
         rank = 0

       If NodeRankInput.frontmatter.output is non-empty:
         Construct the artifact logical name:
           Strip the "SPEC/" prefix from NodeRankInput.logical_name.
           Prepend "ARTIFACT/" to the result.
         Add another RankingEntry with:
           logical_name = the artifact logical name
           dependencies = empty list
           rank = 0

  2. Build dependency edges.

     For each RankingEntry that is a spec node (logical_name is a SPEC/ reference):

       If the spec node is not "SPEC" (the root):
         Call LogicalNameGetParent(logical_name).
         Add the returned parent logical name to the entry's dependencies.

       For each string dep in the corresponding frontmatter.depends_on:
         If dep starts with "ARTIFACT/":
           Add dep as-is to the entry's dependencies.
         Else if LogicalNameIsSpec(dep) is true:
           Call LogicalNameStripQualifier(dep) to get the bare name.
           Add the bare name to the entry's dependencies.
         Else if dep starts with "EXTERNAL/":
           Skip — external files have no rank.

       If the corresponding frontmatter.input is non-empty:
         If frontmatter.input starts with "ARTIFACT/":
           Add frontmatter.input to the entry's dependencies.
         Else if frontmatter.input starts with "EXTERNAL/":
           Skip — external files have no rank.

     For each RankingEntry that is an artifact (logical_name starts with "ARTIFACT/"):
       Derive the generating node logical name:
         Strip the "ARTIFACT/" prefix.
         Prepend "SPEC/".
       Add the generating node logical name to the artifact entry's dependencies.

     After building all dependency edges:
       For each RankingEntry, for each dependency in its dependencies list:
         If the dependency is not a key in the entry map:
           Raise error "unresolvable reference".

  3. Initialize ranks.

     Set the rank of the "SPEC" entry to 0.
     Set the rank of all other entries to 0.

  4. Iterate and detect cycles.

     Let N = total number of entries in the map.
     Let cycle_participants = empty list.

     Repeat up to N times:
       Let changed = false.

       For each RankingEntry (excluding "SPEC"):
         Let max_dep_rank = 0.
         For each dependency in the entry's dependencies:
           Look up the dependency's current rank in the entry map.
           If that rank exceeds max_dep_rank, set max_dep_rank to that rank.
         Let new_rank = 1 + max_dep_rank.
         If new_rank exceeds the entry's current rank:
           Update the entry's rank to new_rank.
           Set changed = true.

       If changed is false:
         Stop iterating (converged, no cycles).

     If after N full passes the loop exited because changed was still true:
       Set cycle_participants to the list of logical names of all entries
       whose rank changed in the last pass.

  5. Build and return output.

     Create ranked as an empty list of NodeRankEntry.

     For each RankingEntry in the entry map:
       Append a NodeRankEntry with:
         logical_name = RankingEntry.logical_name
         rank = RankingEntry.rank

     Sort ranked first by rank ascending, then by logical_name ascending.

     Return (ranked, cycle_participants).
