<!-- code-from-spec: ROOT/functional/logic/utils/text_normalization@6vLGbEZG-TgzpuPhdzZd9x2Qi0I -->

function NormalizeText(raw_string) -> string

  1. Trim leading and trailing occurrences of space (U+0020) and
     horizontal tab (U+0009) from raw_string.

  2. Scan the trimmed string for runs of consecutive whitespace
     characters, where whitespace is defined as space (U+0020)
     or horizontal tab (U+0009).
     Replace each such run with a single space (U+0020).

  3. Apply Unicode simple case folding to the result, converting
     each character to its lowercase equivalent.

  4. Return the resulting string.

  Notes:
  - If raw_string is empty or contains only whitespace, return "".
  - This function is pure: no I/O, no side effects, no errors raised.
  - Deterministic: identical inputs always produce identical outputs.
  - Whitespace is strictly U+0020 and U+0009; do not use broad
    library predicates that include U+00A0 or other whitespace-like
    characters.
