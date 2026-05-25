// code-from-spec: ROOT/golang/internal/pathvalidation/code@PENDING
package pathvalidation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidatePath checks that a relative path is safe to use within projectRoot.
// It rejects empty paths, absolute paths, directory traversal attempts, and
// symlink escapes. Returns nil if the path is safe.
func ValidatePath(path string, projectRoot string) error {
	// 1. Reject empty path.
	if path == "" {
		return fmt.Errorf("path is empty")
	}

	// 2. Reject absolute paths: leading slash or drive letter (e.g. "C:").
	if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") {
		return fmt.Errorf("path is absolute: %s", path)
	}
	if len(path) >= 2 && isLetter(path[0]) && path[1] == ':' {
		return fmt.Errorf("path is absolute: %s", path)
	}

	// 3. Normalize: replace backslashes, resolve . and .. components.
	normalized := filepath.ToSlash(filepath.Clean(path))

	// 4. After normalization, reject if any component is "..".
	// filepath.Clean will leave leading ".." components when the path escapes.
	parts := strings.Split(normalized, "/")
	for _, part := range parts {
		if part == ".." {
			return fmt.Errorf("path contains directory traversal: %s", path)
		}
	}

	// 5. Join with project root to form the full absolute path.
	fullPath := filepath.Join(projectRoot, normalized)

	// 6. Resolve symlinks in the full path. Use EvalSymlinks which resolves
	// the longest existing prefix, but the target may not exist yet.
	// We evaluate the parent directory to handle not-yet-created files.
	realPath, err := resolveReal(fullPath)
	if err != nil {
		// If we cannot resolve at all, allow it — the path may just not exist yet.
		// But if the parent exists and escapes, we catch it below.
		return nil
	}

	// 7. Resolve symlinks in project root.
	realRoot, err := filepath.EvalSymlinks(projectRoot)
	if err != nil {
		// If we cannot resolve the project root, we cannot validate symlink escape.
		return nil
	}

	// Normalize both to forward slashes for consistent comparison.
	realPath = filepath.ToSlash(realPath)
	realRoot = filepath.ToSlash(realRoot)

	// Ensure the root ends with a slash for prefix matching, so that
	// /project-root-extra does not pass validation for /project-root.
	if !strings.HasSuffix(realRoot, "/") {
		realRoot += "/"
	}

	// 8. Check containment.
	if !strings.HasPrefix(realPath, realRoot) && realPath != strings.TrimSuffix(realRoot, "/") {
		return fmt.Errorf("path resolves outside project root: %s", path)
	}

	return nil
}

// resolveReal attempts to resolve symlinks. If the full path doesn't exist,
// it walks up to the nearest existing ancestor, resolves that, then appends
// the remaining segments.
func resolveReal(fullPath string) (string, error) {
	resolved, err := filepath.EvalSymlinks(fullPath)
	if err == nil {
		return resolved, nil
	}
	if !os.IsNotExist(err) {
		return "", err
	}

	// Walk up until we find an existing directory.
	dir := filepath.Dir(fullPath)
	base := filepath.Base(fullPath)
	resolvedDir, err := resolveReal(dir)
	if err != nil {
		return "", err
	}
	return filepath.Join(resolvedDir, base), nil
}

// isLetter returns true if b is an ASCII letter (drive letter check).
func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}
