<!-- code-from-spec: ROOT/functional/logic/os/file_writer@q-3YkwzF852pRRy5itHEhAj2UrA -->

# FileWrite

```
function FileWrite(cfs_path: PathCfs, content: string)

  1. Convert cfs_path to an OS path by calling PathCfsToOs(cfs_path).
     If PathCfsToOs raises an error, propagate it to the caller.
     Let os_path be the resulting PathOs.

  2. Determine the parent directory of os_path.
     If the parent directory does not exist, create it along with
     any missing intermediate directories.
     If any directory cannot be created, raise error "cannot create directory".

  3. Write content to the file at os_path, encoded as UTF-8 text.
     If the file already exists, overwrite it without warning.
     If the file does not exist, create it.
     If the file cannot be written, raise error "cannot write file".
```

## Error Conditions

| Error | Cause |
|---|---|
| path is empty | `cfs_path` value is empty (from `PathValidateCfs`) |
| path is absolute | `cfs_path` starts with `/` or a drive letter (from `PathValidateCfs`) |
| path contains backslash | `cfs_path` contains `\` characters (from `PathValidateCfs`) |
| directory traversal | `cfs_path` contains `..` components after normalization (from `PathValidateCfs`) |
| cannot determine root | working directory cannot be read (from `PathGetProjectRoot`) |
| resolves outside root | resolved path is outside the project root (from `PathCfsToOs`) |
| cannot create directory | an intermediate directory cannot be created |
| cannot write file | the file cannot be written |
