<!-- code-from-spec: ROOT/functional/logic/utils/node_ranking@6kqVyq52zQSlETyDNKaRPwedc-k -->

## Data Structures

record NodeRankInput
  logical_name: string
  frontmatter: Frontmatter

record NodeRankEntry
  logical_name: string
  rank: integer

record InternalEntry
  logical_name: string
  dependencies: list of strings
  rank: integer


## Functions

---

function NodeRankCompute(entries: list of NodeRankInput) -> (ranked: list of NodeRankEntry, cycles: list of string)
  errors:
    - "unresolvable reference": a depends_on or input target cannot be resolved.

  1. **Build entry map** (Step 1)

     Initialize an empty entry map keyed by logical name,
     where each value is an InternalEntry.

     For each NodeRankInput in entries:

       a. Add a spec node entry:
          - key: the input's logical_name
          - dependencies: empty list (populated in Step 2)
          - rank: 0

       b. For each output in the input's frontmatter.outputs:
          - Derive the artifact logical name:
            Strip the "ROOT/" prefix from the node's logical_name,
            prepend "ARTIFACT/", and append "(<id>)" where <id>
            is the output's id field.
            Example: node "ROOT/a/b" with output id "foo"
            → "ARTIFACT/a/b(foo)"
          - Add an artifact entry:
            - key: the derived artifact logical name
            - dependencies: empty list (populated in Step 2)
            - rank: 0

  2. **Build dependency edges** (Step 2)

     For each spec node entry in the entry map:

       a. Parent dependency:
          If the logical_name is "ROOT", skip (ROOT has no parent).
          Otherwise, call LogicalNameGetParent(logical_name)
          to get the parent logical name. Add the parent logical
          name to this entry's dependencies.
          If the parent is not found in the entry map,
          raise error "unresolvable reference".

       b. depends_on dependencies:
          For each ref in frontmatter.depends_on:
            - If ref starts with "ARTIFACT/", use ref as-is as the
              lookup key.
            - If ref starts with "ROOT/", call
              LogicalNameStripQualifier(ref) to get the bare
              logical name; use that as the lookup key.
          If the lookup key is not found in the entry map,
          raise error "unresolvable reference".
          Add the lookup key to this entry's dependencies.

       c. input dependency:
          If frontmatter.input is non-empty, add frontmatter.input
          as a dependency (it is an "ARTIFACT/" reference, used as-is).
          If frontmatter.input is not found in the entry map,
          raise error "unresolvable reference".

     For each artifact entry in the entry map:
       - The generating node logical name is derived by:
         stripping the "ARTIFACT/" prefix, stripping any qualifier,
         and prepending "ROOT/".
         Equivalently: call LogicalNameGetArtifactGenerator on the
         artifact's logical name to obtain the generator node's
         logical name.
       - Add the generator's logical name to this artifact entry's
         dependencies.
       - If the generator is not found in the entry map,
         raise error "unresolvable reference".

  3. **Initialize ranks** (Step 3)

     Set rank of entry keyed "ROOT" to 0.
     Set rank of all other entries to 0.
     (All entries begin at 0; ROOT is fixed and excluded from
     iteration.)

  4. **Iterate and detect cycles** (Step 4)

     Let N = total number of entries in the entry map.

     Initialize cycle_candidates as an empty list.

     Repeat up to N times (pass index from 1 to N):

       a. Set changed = false.

       b. For each entry in the entry map, excluding "ROOT":
            Compute candidate_rank = 1 + max(rank of each entry
            in this entry's dependencies list).
            If candidate_rank > entry's current rank:
              Update entry's rank to candidate_rank.
              Set changed = true.
              Record this entry's logical_name as a potential
              cycle candidate.

       c. If changed is false, stop iterating (converged, no cycles).

       d. If this is pass number N and changed is still true:
            A cycle exists.
            Set cycle_candidates to the list of logical names
            that changed in this final pass.
            Stop iterating.

     If the loop stopped due to non-convergence (pass N completed
     with changes), the cycle_candidates list is non-empty.
     Otherwise, cycle_candidates remains empty.

  5. **Output** (Step 5)

     Collect all entries from the entry map as NodeRankEntry records
     (logical_name + rank).

     Sort the collected entries:
       - Primary sort: rank ascending.
       - Secondary sort: logical_name ascending (lexicographic).

     Return:
       - ranked: the sorted list of NodeRankEntry records.
       - cycles: the cycle_candidates list (empty if no cycle
         was detected).
