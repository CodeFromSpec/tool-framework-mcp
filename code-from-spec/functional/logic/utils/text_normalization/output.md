<!-- code-from-spec: ROOT/functional/logic/utils/text_normalization@Oso--NTz8yT0bgPb2Dokj1uJOuo -->

# Text Normalization

## Functions

---

function NormalizeText(raw_string) -> string

  1. If raw_string is empty, return "".

  2. Trim leading and trailing whitespace characters from raw_string.
     Whitespace characters are: space (U+0020) and horizontal tab (U+0009).

  3. Scan the trimmed string for runs of consecutive whitespace characters.
     For each run of one or more whitespace characters, replace the run
     with a single space (U+0020).

  4. Apply Unicode simple case folding to the resulting string.
     This converts each character to its case-folded equivalent
     (e.g., "A" -> "a", "Straße" -> "strasse").

  5. Return the transformed string.

## Contracts

- Pure function — performs no I/O and raises no errors.
- Deterministic — identical inputs always produce identical outputs.

## Examples

| Input                    | Output                  |
|--------------------------|-------------------------|
| "  Interface  "          | "interface"             |
| "PUBLIC"                 | "public"                |
| "Straße"                 | "strasse"               |
| "Testes   de   aceitação" | "testes de aceitação"  |
| ""                       | ""                      |
