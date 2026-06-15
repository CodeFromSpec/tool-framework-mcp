<!-- code-from-spec: SPEC/functional/tests/parsing/frontmatter@RregwZnV3STv3VBd-rm6kDBAink -->

## Test Suite: FrontmatterParse

---

### Happy Path

#### TC-1: Parses complete frontmatter (all fields)

Setup:
  Create a file with frontmatter containing all fields:
    depends_on:
      - "SPEC/some/node"
      - "ARTIFACT/some/output"
      - "EXTERNAL/proto/api.proto"
    input: "some/input/path"
    output: "some/output/path"

Action:
  Call FrontmatterParse with the file's PathCfs.

Expected outcome:
  No error.
  Result depends_on contains "SPEC/some/node", "ARTIFACT/some/output", "EXTERNAL/proto/api.proto".
  Result input equals "some/input/path".
  Result output equals "some/output/path".

---

#### TC-2: Parses frontmatter with only output

Setup:
  Create a file with frontmatter containing only:
    output: "only/output/path"

Action:
  Call FrontmatterParse with the file's PathCfs.

Expected outcome:
  No error.
  Result depends_on is empty.
  Result input is empty.
  Result output equals "only/output/path".

---

#### TC-3: Parses frontmatter with only depends_on

Setup:
  Create a file with frontmatter containing only:
    depends_on:
      - "SPEC/some/dep"

Action:
  Call FrontmatterParse with the file's PathCfs.

Expected outcome:
  No error.
  Result depends_on contains "SPEC/some/dep".
  Result input is empty.
  Result output is empty.

---

#### TC-4: Parses frontmatter with EXTERNAL/ in depends_on

Setup:
  Create a file with frontmatter containing:
    depends_on:
      - "EXTERNAL/proto/api.proto"

Action:
  Call FrontmatterParse with the file's PathCfs.

Expected outcome:
  No error.
  Result depends_on contains "EXTERNAL/proto/api.proto".

---

#### TC-5: Parses frontmatter with input field

Setup:
  Create a file with frontmatter containing only:
    input: "some/input/file"

Action:
  Call FrontmatterParse with the file's PathCfs.

Expected outcome:
  No error.
  Result input equals "some/input/file".
  Result depends_on is empty.
  Result output is empty.

---

#### TC-6: Ignores unknown frontmatter fields

Setup:
  Create a file with frontmatter containing:
    output: "some/output/path"
    custom_field: "unexpected value"

Action:
  Call FrontmatterParse with the file's PathCfs.

Expected outcome:
  No error.
  Result output equals "some/output/path".
  Unknown field custom_field is silently ignored.

---

#### TC-7: File with no frontmatter returns empty result

Setup:
  Create a file with only body content and no "---" delimiter.

Action:
  Call FrontmatterParse with the file's PathCfs.

Expected outcome:
  No error.
  Result depends_on is empty.
  Result input is empty.
  Result output is empty.

---

### Edge Cases

#### TC-8: Empty frontmatter

Setup:
  Create a file with opening and closing "---" and nothing between them.

Action:
  Call FrontmatterParse with the file's PathCfs.

Expected outcome:
  No error.
  Result depends_on is empty.
  Result input is empty.
  Result output is empty.

---

#### TC-9: File with only frontmatter, nothing after

Setup:
  Create a file with frontmatter containing:
    output: "some/output/path"
  No body content after the closing "---".

Action:
  Call FrontmatterParse with the file's PathCfs.

Expected outcome:
  No error.
  Result output equals "some/output/path".

---

#### TC-10: Delimiter with trailing whitespace is not recognized

Setup:
  Create a file whose first line is "---   " (three dashes followed by spaces).

Action:
  Call FrontmatterParse with the file's PathCfs.

Expected outcome:
  No error.
  The line is not treated as a frontmatter delimiter.
  Result has all fields empty.

---

### Failure Cases

#### TC-11: File does not exist

Setup:
  No file is created. Use a PathCfs pointing to a non-existent location.

Action:
  Call FrontmatterParse with the non-existent PathCfs.

Expected outcome:
  Error FileUnreadable is raised.

---

#### TC-12: Propagates path errors

Setup:
  Construct an invalid PathCfs such as "../../outside".

Action:
  Call FrontmatterParse with the invalid PathCfs.

Expected outcome:
  Error DirectoryTraversal is raised, propagated from FileReader/PathUtils via FileOpen.

---

#### TC-13: Malformed YAML

Setup:
  Create a file with invalid YAML between frontmatter delimiters, for example:
    ---
    depends_on: [unclosed bracket
    ---

Action:
  Call FrontmatterParse with the file's PathCfs.

Expected outcome:
  Error MalformedYAML is raised.

---

#### TC-14: Unclosed frontmatter block

Setup:
  Create a file that starts with "---" but has no closing "---".

Action:
  Call FrontmatterParse with the file's PathCfs.

Expected outcome:
  Error MalformedYAML is raised.

---

#### TC-15: Unknown field 'external' is silently ignored

Setup:
  Create a file with frontmatter containing an "external" field (v3 format):
    output: "some/output/path"
    external: "some/external/ref"

Action:
  Call FrontmatterParse with the file's PathCfs.

Expected outcome:
  No error.
  The external field is ignored.
  Result output equals "some/output/path".
  Result depends_on is empty.
  Result input is empty.
