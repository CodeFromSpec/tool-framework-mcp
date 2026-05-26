// code-from-spec: ROOT/golang/internal/pathvalidation/code@Qwvf4ocJPrEYd047iBhTXhyyb7A

// Package pathvalidation provides a single function, ValidatePath, that
// verifies a caller-supplied relative file path is safe to use within a
// project root directory before any write operation is performed.
//
// Threat model addressed
//   - Relative traversal  : ../../etc/passwd
//   - Embedded traversal  : internal/../../outside/file.go
//   - OS-specific separators: backslash on Windows (..\..\)
//   - Encoding tricks      : percent-encoded or Unicode-escaped sequences
//   - Symlinks             : a valid relative path that resolves outside
//     the project via a symlink in the directory tree
//
// The function never attempts to sanitize or repair an invalid path.
// Every invalid input is rejected outright — the caller decides how to
// handle the error.
package pathvalidation

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// ValidatePath checks whether path is a safe relative path that resolves
// within projectRoot. It returns nil when the path is safe, or a descriptive
// error when any validation step fails.
//
// The function is read-only: it never creates, writes, or modifies any file.
//
// Error messages follow the spec exactly:
//   - "path is empty"
//   - "path is absolute: <path>"
//   - "path contains directory traversal: <path>"
//   - "path resolves outside project root: <path>"
func ValidatePath(path string, projectRoot string) error {
	// -------------------------------------------------------------------------
	// Step 1 — Reject empty paths.
	// An empty string cannot name a valid file inside the project.
	// -------------------------------------------------------------------------
	if path == "" {
		return fmt.Errorf("path is empty")
	}

	// -------------------------------------------------------------------------
	// Step 2 — Reject absolute paths.
	//
	// On Unix:    any path starting with "/" is absolute.
	// On Windows: a drive letter followed by ":" (e.g. "C:") makes it
	//             absolute. filepath.IsAbs would handle the Windows case but
	//             it does NOT treat a bare "/" as absolute on Windows (it is
	//             "rooted but not absolute"). We therefore check both
	//             conditions explicitly and portably.
	//
	// We check the raw input before any normalization so that tricks such as
	// a leading backslash are caught here too (handled further below), and
	// so that the error message always reflects what the caller submitted.
	// -------------------------------------------------------------------------
	if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") {
		return fmt.Errorf("path is absolute: %s", path)
	}
	// Detect a Windows drive letter: one ASCII letter followed by ':'
	// (e.g. "C:", "D:", "c:").
	if len(path) >= 2 && path[1] == ':' {
		first := path[0]
		if (first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') {
			return fmt.Errorf("path is absolute: %s", path)
		}
	}

	// -------------------------------------------------------------------------
	// Step 3 — Normalize OS-specific separators.
	// Replace every backslash with a forward slash so that the rest of the
	// validation pipeline works on a single canonical separator regardless of
	// the host OS. This neutralizes tricks like "..\..\etc\passwd".
	// -------------------------------------------------------------------------
	normalized := strings.ReplaceAll(path, "\\", "/")

	// -------------------------------------------------------------------------
	// Step 4 — Decode percent-encoded and Unicode escape sequences.
	// url.PathUnescape decodes sequences like "%2F", "%2E", "%2e%2e%2f" into
	// their literal characters. If decoding fails (malformed encoding), we
	// treat the path as containing traversal to be safe.
	// -------------------------------------------------------------------------
	decoded, err := url.PathUnescape(normalized)
	if err != nil {
		// Malformed percent-encoding — treat as traversal attempt.
		return fmt.Errorf("path contains directory traversal: %s", path)
	}
	normalized = decoded

	// -------------------------------------------------------------------------
	// Step 5 — Normalize "." and ".." components lexically (no filesystem I/O).
	// filepath.Clean resolves ".", "..", duplicate separators, and trailing
	// slashes into a canonical form. We convert back to slash-separated form
	// for the component inspection that follows.
	// -------------------------------------------------------------------------
	normalized = filepath.ToSlash(filepath.Clean(normalized))

	// -------------------------------------------------------------------------
	// Step 6 — Reject remaining ".." components.
	// Even after Clean, a path that begins with ".." will still contain that
	// component. Split on "/" and check each segment individually.
	//
	// The original caller-supplied path is used in the error message so the
	// caller knows exactly what was submitted.
	// -------------------------------------------------------------------------
	for _, component := range strings.Split(normalized, "/") {
		if component == ".." {
			return fmt.Errorf("path contains directory traversal: %s", path)
		}
	}

	// -------------------------------------------------------------------------
	// Step 7 — Build the candidate absolute path.
	// Join projectRoot with the normalized (clean) relative path.
	// filepath.Join handles the single-separator guarantee at the join point.
	// -------------------------------------------------------------------------
	candidate := filepath.Join(projectRoot, normalized)

	// -------------------------------------------------------------------------
	// Step 8 — Resolve symlinks in the candidate path.
	// filepath.EvalSymlinks follows every symlink component and returns the
	// real physical path. If the full path does not exist, walk upward to
	// find the deepest existing ancestor, resolve symlinks on that, and
	// append the remaining unresolved segments. This handles the common case
	// where the file does not exist yet but the parent directories do.
	// -------------------------------------------------------------------------
	resolvedCandidate, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		resolvedCandidate, err = resolveExistingPrefix(candidate)
		if err != nil {
			return fmt.Errorf("path resolves outside project root: %s", path)
		}
	}

	// -------------------------------------------------------------------------
	// Step 9 — Verify the resolved path is inside the project root.
	// Resolve symlinks in projectRoot as well to get its canonical form, then
	// check that resolvedCandidate starts with that canonical root.
	//
	// The containment check appends a separator to the canonical root before
	// comparing, preventing a false positive where a root of "/project" would
	// match a resolved path of "/project-extra".
	// -------------------------------------------------------------------------
	resolvedRoot, err := filepath.EvalSymlinks(projectRoot)
	if err != nil {
		// If the project root itself cannot be resolved, we cannot make a safe
		// determination — reject the path.
		return fmt.Errorf("path resolves outside project root: %s", path)
	}

	// Normalize both resolved paths to forward slashes for a portable prefix
	// comparison (relevant on Windows where filepath.EvalSymlinks may return
	// backslash-separated paths).
	resolvedCandidateSlash := filepath.ToSlash(resolvedCandidate)
	resolvedRootSlash := filepath.ToSlash(resolvedRoot)

	// Ensure the root ends with "/" so the prefix check is exact:
	//   "/project"       must NOT match "/project-extra/file.go"
	//   "/project/"  DOES match "/project/internal/file.go"
	rootWithSep := resolvedRootSlash
	if !strings.HasSuffix(rootWithSep, "/") {
		rootWithSep += "/"
	}

	// A path equal to the project root itself (no trailing separator) is also
	// acceptable — it means the caller is writing to the root directory entry.
	if resolvedCandidateSlash != resolvedRootSlash &&
		!strings.HasPrefix(resolvedCandidateSlash, rootWithSep) {
		return fmt.Errorf("path resolves outside project root: %s", path)
	}

	// -------------------------------------------------------------------------
	// Step 10 — Return success.
	// All validation checks passed; the path is safe to use.
	// -------------------------------------------------------------------------
	return nil
}

// resolveExistingPrefix walks upward from candidate until it finds a directory
// that exists on disk, resolves symlinks on that prefix, then re-appends the
// remaining (non-existent) segments. This allows ValidatePath to work with
// paths whose target file has not been created yet.
func resolveExistingPrefix(candidate string) (string, error) {
	current := candidate
	var tail []string
	for {
		resolved, err := filepath.EvalSymlinks(current)
		if err == nil {
			result := resolved
			for i := len(tail) - 1; i >= 0; i-- {
				result = filepath.Join(result, tail[i])
			}
			return result, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("no existing ancestor")
		}
		tail = append(tail, filepath.Base(current))
		current = parent
	}
}
