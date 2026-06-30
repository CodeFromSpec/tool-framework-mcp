// code-from-spec: SPEC/golang/implementation/oslayer/path@dyKT71lYlHm9lMdA2HfeCSbSGG8
package oslayer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type CfsPath string
type OsPath string

func GetProjectRoot() (OsPath, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrCannotDetermineRoot, err)
	}
	return OsPath(wd), nil
}

func ValidateStringIsCfsPath(value string) error {
	if value == "" {
		return fmt.Errorf("%w", ErrPathEmpty)
	}
	if strings.HasPrefix(value, "/") || strings.Contains(value, ":") {
		return fmt.Errorf("%w", ErrPathAbsolute)
	}
	if strings.Contains(value, "\\") {
		return fmt.Errorf("%w", ErrPathContainsBackslash)
	}
	normalized := filepath.ToSlash(filepath.Clean(value))
	for _, component := range strings.Split(normalized, "/") {
		if component == ".." {
			return fmt.Errorf("%w", ErrDirectoryTraversal)
		}
	}
	return nil
}

func containedInRootPath(resolved, root string) bool {
	if resolved == root {
		return true
	}
	return strings.HasPrefix(resolved, root+string(filepath.Separator))
}

func CfsPathToOs(cfsPath CfsPath) (OsPath, error) {
	if err := ValidateStringIsCfsPath(string(cfsPath)); err != nil {
		return "", err
	}
	root, err := GetProjectRoot()
	if err != nil {
		return "", err
	}
	osRelative := filepath.FromSlash(string(cfsPath))
	absolutePath := filepath.Join(string(root), osRelative)
	if _, statErr := os.Stat(absolutePath); statErr == nil {
		resolved, resolveErr := filepath.EvalSymlinks(absolutePath)
		if resolveErr != nil {
			return "", fmt.Errorf("%w: %w", ErrResolvesOutsideRoot, resolveErr)
		}
		if !containedInRootPath(resolved, string(root)) {
			return "", fmt.Errorf("%w", ErrResolvesOutsideRoot)
		}
		absolutePath = resolved
	}
	return OsPath(absolutePath), nil
}

func OsPathToCfs(osPath OsPath) (CfsPath, error) {
	root, err := GetProjectRoot()
	if err != nil {
		return "", err
	}
	resolved := string(osPath)
	if _, statErr := os.Stat(resolved); statErr == nil {
		resolvedSymlink, resolveErr := filepath.EvalSymlinks(resolved)
		if resolveErr != nil {
			return "", fmt.Errorf("%w: %w", ErrResolvesOutsideRoot, resolveErr)
		}
		resolved = resolvedSymlink
	}
	if !containedInRootPath(resolved, string(root)) {
		return "", fmt.Errorf("%w", ErrResolvesOutsideRoot)
	}
	relativePath := strings.TrimPrefix(resolved, string(root))
	relativePath = strings.TrimPrefix(relativePath, string(filepath.Separator))
	cfsValue := filepath.ToSlash(relativePath)
	return CfsPath(cfsValue), nil
}
