// code-from-spec: ROOT/golang/internal/pathvalidation/code@PENDING
package pathvalidation

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidatePath checks that a relative path is safe to use within projectRoot.
func ValidatePath(path string, projectRoot string) error {
	if path == "" {
		return fmt.Errorf("path is empty")
	}

	if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") {
		return fmt.Errorf("path is absolute: %s", path)
	}
	if len(path) >= 2 && isLetter(path[0]) && path[1] == ':' {
		return fmt.Errorf("path is absolute: %s", path)
	}

	cleaned := filepath.Clean(path)
	parts := strings.Split(filepath.ToSlash(cleaned), "/")
	for _, part := range parts {
		if part == ".." {
			return fmt.Errorf("path contains directory traversal: %s", path)
		}
	}

	absPath, err := filepath.Abs(filepath.Join(projectRoot, cleaned))
	if err != nil {
		return fmt.Errorf("path resolves outside project root: %s", path)
	}

	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return fmt.Errorf("path resolves outside project root: %s", path)
	}

	// Resolve symlinks on the longest existing prefix of the target path.
	realPath := resolveExistingPrefix(absPath)
	realRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		realRoot = absRoot
	}

	// Normalize for comparison.
	realPath = normalizePath(realPath)
	realRoot = normalizePath(realRoot)

	if !strings.HasPrefix(realPath, realRoot+"/") && realPath != realRoot {
		return fmt.Errorf("path resolves outside project root: %s", path)
	}

	return nil
}

// resolveExistingPrefix resolves symlinks on the longest existing ancestor,
// then appends the remaining non-existent segments.
func resolveExistingPrefix(absPath string) string {
	// Try resolving the full path first.
	resolved, err := filepath.EvalSymlinks(absPath)
	if err == nil {
		return resolved
	}

	// Walk up to find the deepest existing ancestor.
	var trailing []string
	current := absPath
	for {
		parent := filepath.Dir(current)
		if parent == current {
			// Reached filesystem root without finding an existing dir.
			return absPath
		}
		trailing = append([]string{filepath.Base(current)}, trailing...)
		current = parent

		resolved, err := filepath.EvalSymlinks(current)
		if err == nil {
			// Found an existing ancestor. Append trailing segments.
			for _, seg := range trailing {
				resolved = filepath.Join(resolved, seg)
			}
			return resolved
		}
	}
}

// normalizePath converts a path to lowercase forward-slash form for comparison.
func normalizePath(p string) string {
	p = filepath.ToSlash(p)
	if filepath.Separator == '\\' {
		p = strings.ToLower(p)
	}
	return p
}

func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}
