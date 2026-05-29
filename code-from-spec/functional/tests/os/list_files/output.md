<!-- code-from-spec: ROOT/functional/tests/os/list_files@8a__BUD_FR1ccWj6DOUOPwrZVLM -->

# Test Specification: ListFiles

## Function Under Test

```
function ListFiles(cfs_path: PathCfs) -> list of PathCfs
```

---

## Test Cases

### Happy Path

---

#### TC-01: Lists files in a flat directory

Setup:
  - Create a temporary directory within the project root, referred to as <base>.
  - Inside <base>, create three files: `a.txt`, `b.txt`, `c.txt`.

Actions:
  1. Call `ListFiles` with the `PathCfs` pointing to <base>.

Expected outcome:
  - Returns a list of three `PathCfs` values.
  - The list is sorted alphabetically.
  - The values correspond to `<base>/a.txt`, `<base>/b.txt`, `<base>/c.txt`
    (each expressed as a path relative to the project root).

---

#### TC-02: Lists files recursively

Setup:
  - Create the following directory structure within the project root:
    ```
    <base>/
      alpha.txt
      sub/
        beta.txt
        deep/
          gamma.txt
    ```

Actions:
  1. Call `ListFiles` with the `PathCfs` pointing to <base>.

Expected outcome:
  - Returns a list of three `PathCfs` values.
  - The list is sorted alphabetically.
  - The values correspond to:
    - `<base>/alpha.txt`
    - `<base>/sub/beta.txt`
    - `<base>/sub/deep/gamma.txt`
    each expressed as a path relative to the project root.

---

#### TC-03: Results are sorted alphabetically

Setup:
  - Create a temporary directory within the project root, referred to as <base>.
  - Inside <base>, create three files in this order: `z.txt`, `a.txt`, `m.txt`.

Actions:
  1. Call `ListFiles` with the `PathCfs` pointing to <base>.

Expected outcome:
  - Returns a list of three `PathCfs` values in the order:
    `<base>/a.txt`, `<base>/m.txt`, `<base>/z.txt`.

---

### Edge Cases

---

#### TC-04: Empty directory

Setup:
  - Create a temporary directory within the project root, referred to as <base>.
  - Do not create any files or subdirectories inside <base>.

Actions:
  1. Call `ListFiles` with the `PathCfs` pointing to <base>.

Expected outcome:
  - Returns an empty list.
  - No error is raised.

---

#### TC-05: Directory with only subdirectories

Setup:
  - Create a temporary directory within the project root, referred to as <base>.
  - Inside <base>, create one or more subdirectories, but no files at any level.

Actions:
  1. Call `ListFiles` with the `PathCfs` pointing to <base>.

Expected outcome:
  - Returns an empty list.
  - No error is raised.

---

### Failure Cases

---

#### TC-06: Directory does not exist

Setup:
  - Identify a path within the project root that does not exist on the filesystem.

Actions:
  1. Call `ListFiles` with a `PathCfs` pointing to the non-existent directory.

Expected outcome:
  - Raises error "directory not found".

---

#### TC-07: Propagates validation errors from PathCfsToOs

Setup:
  - No filesystem setup required.

Actions:
  1. Call `ListFiles` with an invalid `PathCfs` value such as `"../../outside"`.

Expected outcome:
  - Raises error "directory traversal", propagated from `PathCfsToOs`.

---

#### TC-08: Propagates conversion errors from PathOsToCfs

Precondition:
  - Skip this test on platforms where symlinks are not supported.

Setup:
  - Create a temporary directory within the project root, referred to as <base>.
  - Inside <base>, create a regular file.
  - Inside <base>, create a symlink that resolves to a target outside the project root.

Actions:
  1. Call `ListFiles` with the `PathCfs` pointing to <base>.

Expected outcome:
  - Raises error "resolves outside root", propagated from `PathOsToCfs`.

---

#### TC-09: Walk error

Precondition:
  - Skip this test on platforms where directory permissions cannot prevent traversal
    (e.g., when running as a superuser).

Setup:
  - Create a temporary directory within the project root, referred to as <base>.
  - Inside <base>, create a subdirectory referred to as <restricted>.
  - Inside <restricted>, create at least one file.
  - Set the permissions on <restricted> so that its contents cannot be read.

Actions:
  1. Call `ListFiles` with the `PathCfs` pointing to <base>.

Expected outcome:
  - Raises error "walk error".
