<!-- code-from-spec: ROOT/functional/logic/utils/text_normalization@XJ7Af7tH0zQ0nKxUT-OnazWkq7A -->

function NormalizeText(raw_string) -> string

  1. Trim leading and trailing whitespace (U+0020 and U+0009) from raw_string.

  2. Collapse each run of one or more whitespace characters (U+0020, U+0009)
     within the result to a single space (U+0020).

  3. Apply Unicode simple case folding to the result.
     This maps each character to its case-folded equivalent
     (e.g., "A" -> "a", "Straße" -> "strasse").

  4. Return the resulting string.

  Notes:
  - If raw_string is empty or contains only whitespace, return "".
  - This is a pure function: no I/O, no side effects, no errors raised.
  - Deterministic: identical inputs always produce identical outputs.
