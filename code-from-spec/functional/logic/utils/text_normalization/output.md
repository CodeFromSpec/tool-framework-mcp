<!-- code-from-spec: ROOT/functional/logic/utils/text_normalization@h-507JpxgN_pgSJkzCIAgRvbtcQ -->

function NormalizeText(raw_string) -> string

  1. If raw_string is empty, return "".

  2. Trim leading and trailing whitespace characters
     (space U+0020 and horizontal tab U+0009) from raw_string.

  3. Collapse each run of one or more consecutive whitespace
     characters (space U+0020 and horizontal tab U+0009) within
     the string to a single space (U+0020).

  4. Apply Unicode simple case folding to the resulting string,
     converting each character to its lowercase equivalent
     (e.g., "Straße" -> "strasse", "PUBLIC" -> "public").

  5. Return the resulting string.
