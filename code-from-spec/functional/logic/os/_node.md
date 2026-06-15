# SPEC/functional/logic/os

Operating system abstractions — file I/O, path handling,
and directory listing.

# Private

## Design intent

This is the OS abstraction layer. All filesystem
interaction in the framework goes through components
in this subtree. Other functional components must not
interact with the OS directly — they depend on this
layer instead.

## Motivation

Different operating systems have incompatible filesystem
conventions:
- Path separators (`/` on Unix, `\` on Windows)
- Path resolution (case sensitivity, symlink behavior)
- File APIs (different syscalls, encoding defaults)

By centralizing OS interaction here, the rest of the
framework works with a single canonical path format
(`CfsPath`, forward slashes, relative to project root)
and never deals with OS-specific details. The `os/`
layer handles the translation.
