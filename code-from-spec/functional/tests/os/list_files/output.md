<!-- code-from-spec: ROOT/functional/tests/os/list_files@ACYOXZUrctRuc1kLRnno50tYN5Q -->

## Test cases for ListFiles

---

### TC-1: Lists files in a flat directory

Setup:
  Create a temporary directory.
  Inside it, create three files: "a.txt", "b.txt", "c.txt".
  Construct a PathCfs pointing to the temporary directory.

Action:
  Call ListFiles with the PathCfs of the temporary directory.

Expected outcome:
  Returns a list of three PathCfs values.
  The list is in alphabetical order: "a.txt", "b.txt", "c.txt"
  (each expressed relative to the project root).

---

### TC-2: Lists files recursively

Setup:
  Create a temporary directory named "dir".
  Inside "dir", create file "alpha.txt".
  Inside "dir", create subdirectory "sub".
  Inside "dir/sub", create file "beta.txt".
  Inside "dir/sub", create subdirectory "deep".
  Inside "dir/sub/deep", create file "gamma.txt".
  Construct a PathCfs pointing to "dir".

Action:
  Call ListFiles with the PathCfs of "dir".

Expected outcome:
  Returns a list of three PathCfs values in alphabetical order:
    "dir/alpha.txt"
    "dir/sub/beta.txt"
    "dir/sub/deep/gamma.txt"
  (each expressed relative to the project root).

---

### TC-3: Results are sorted alphabetically

Setup:
  Create a temporary directory.
  Inside it, create three files: "z.txt", "a.txt", "m.txt"
  (created in that non-alphabetical order).
  Construct a PathCfs pointing to the temporary directory.

Action:
  Call ListFiles with the PathCfs of the temporary directory.

Expected outcome:
  Returns a list of three PathCfs values in alphabetical order:
    "a.txt", "m.txt", "z.txt"
  (each expressed relative to the project root).

---

### TC-4: Empty directory

Setup:
  Create an empty temporary directory.
  Construct a PathCfs pointing to it.

Action:
  Call ListFiles with the PathCfs of the empty directory.

Expected outcome:
  Returns an empty list.
  No error is raised.

---

### TC-5: Directory with only subdirectories

Setup:
  Create a temporary directory.
  Inside it, create one or more subdirectories, each potentially
  containing further subdirectories, but no files at any level.
  Construct a PathCfs pointing to the temporary directory.

Action:
  Call ListFiles with the PathCfs of the temporary directory.

Expected outcome:
  Returns an empty list.
  No error is raised.

---

### TC-6: Directory does not exist

Setup:
  Construct a PathCfs pointing to a path that does not exist on disk.

Action:
  Call ListFiles with the non-existent PathCfs.

Expected outcome:
  Raises error DirectoryNotFound.

---

### TC-7: Propagates validation errors from PathCfsToOs

Setup:
  Construct a PathCfs with a value that would traverse outside the
  project root, for example "../../outside".

Action:
  Call ListFiles with the invalid PathCfs.

Expected outcome:
  Raises error DirectoryTraversal (propagated from PathUtils).

---

### TC-8: Propagates conversion errors from PathOsToCfs

Skip this test on platforms where symlinks are not supported.

Setup:
  Create a temporary directory inside the project root.
  Inside it, create a regular file.
  Inside it, create a symlink that resolves to a target outside
  the project root.
  Construct a PathCfs pointing to the temporary directory.

Action:
  Call ListFiles with that PathCfs.

Expected outcome:
  Raises error ResolvesOutsideRoot (propagated from PathUtils).

---

### TC-9: Walk error

Skip this test on platforms where directory permissions cannot
prevent traversal.

Setup:
  Create a temporary directory.
  Inside it, create a subdirectory.
  Set permissions on the subdirectory so that it cannot be read
  or listed.
  Construct a PathCfs pointing to the temporary directory.

Action:
  Call ListFiles with the PathCfs of the parent directory.

Expected outcome:
  Raises error WalkError.
