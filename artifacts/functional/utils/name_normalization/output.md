<!-- code-from-spec: ROOT/functional/utils/name_normalization@PENDING -->

function NormalizeName(raw_string) -> string

  1. Trim leading and trailing whitespace from raw_string.
     Whitespace characters are space (U+0020) and horizontal
     tab (U+0009).

  2. For each run of consecutive whitespace characters within
     the string, replace the entire run with a single space
     (U+0020).

  3. Apply Unicode simple case folding to the entire string.

  4. Return the result.
