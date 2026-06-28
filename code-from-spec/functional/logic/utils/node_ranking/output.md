<!-- code-from-spec: SPEC/functional/logic/utils/node_ranking@lTNwbyKWux4M8C_n8nUshKUS4rg -->

record NodeRankInput
  logical_name: string
  frontmatter: frontmatter.Frontmatter

record NodeRankEntry
  logical_name: string
  rank: integer

function NodeRankCompute(entries: list of NodeRankInput) -> (ranked: list of NodeRankEntry, cycles: list of string)
  errors:
    - UnresolvableReference: a depends_on or input target cannot be resolved.

  1. Build an entry map keyed by logical name.

     For each NodeRankInput in entries:
       Add a spec entry keyed by logical_name with:
         - deps: empty list (to be filled in step 2)
         - rank: 0

       If frontmatter.output is non-empty:
         Construct artifact logical name:
           Strip "SPEC/" prefix from logical_name and prepend "ARTIFACT/".
           Example: "SPEC/a/b" -> "ARTIFACT/a/b"
         Add an artifact entry keyed by that artifact logical name with:
           - deps: list containing the generating node's logical_name
           - rank: 0

  2. Build dependency edges for each spec node entry.

     For each spec node entry in the entry map:
       If logical_name is "SPEC":
         Skip — root node has no dependencies.

       Else:
         a. Parent dependency:
            Call LogicalNameGetParent(logical_name) to get the parent.
            Add the parent to the entry's deps list.

         b. depends_on dependencies:
            For each reference in frontmatter.depends_on:
              If LogicalNameIsSpec(reference) is true:
                Call LogicalNameStripQualifier(reference) to get bare_name.
                If bare_name is not a key in the entry map:
                  Raise error "UnresolvableReference"
                Add bare_name to the entry's deps list.
              Else if reference starts with "ARTIFACT/":
                If reference is not a key in the entry map:
                  Raise error "UnresolvableReference"
                Add reference to the entry's deps list.
              Else if reference starts with "EXTERNAL/":
                Skip — external files have no rank.
              Else:
                Raise error "UnresolvableReference"

         c. input dependency:
            If frontmatter.input is non-empty:
              If frontmatter.input starts with "ARTIFACT/":
                If frontmatter.input is not a key in the entry map:
                  Raise error "UnresolvableReference"
                Add frontmatter.input to the entry's deps list.
              Else if frontmatter.input starts with "EXTERNAL/":
                Skip — external files have no rank.

  3. Initialize ranks.

     Set rank of entry keyed "SPEC" to 0 (fixed, never updated).
     All other entries already have rank 0 from step 1.

  4. Iterate and detect cycles.

     Let N = total number of entries in the entry map.
     Let cycle_candidates = empty list.

     Repeat up to N times, tracking iteration index i from 1 to N:
       Let changed = false.

       For each entry in the entry map (excluding "SPEC"):
         Let max_dep_rank = maximum rank among all entries in the entry's deps list.
         Let new_rank = 1 + max_dep_rank.

         If new_rank > entry's current rank:
           Update entry's rank to new_rank.
           Set changed = true.
           If i equals N:
             Add entry's logical_name to cycle_candidates.

       If changed is false:
         Stop iteration (converged, no cycles).

     If iteration completed all N passes and changed was still true on pass N:
       Set cycles = cycle_candidates.
     Else:
       Set cycles = empty list.

  5. Collect and return results.

     Build ranked list:
       For each entry in the entry map:
         Append NodeRankEntry with logical_name and rank.

     Sort ranked list:
       Primary sort: rank ascending.
       Secondary sort: logical_name ascending.

     Return (ranked: ranked list, cycles: cycles).
