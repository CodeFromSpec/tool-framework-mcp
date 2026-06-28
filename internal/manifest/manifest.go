// code-from-spec: SPEC/golang/implementation/manifest@_2-5YIkaqpxrKB0guBG9fyWmHu0
package manifest

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/file"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

var ErrInvalidMode  = errors.New("invalid mode")
var ErrLockTimeout  = errors.New("lock timeout")
var ErrWrongMode    = errors.New("wrong mode")
var ErrHandleClosed = errors.New("handle closed")

type ManifestEntry struct {
	Path      string
	Checksum  string
	ChainHash string
}

type ManifestHandle struct {
	Mode       string
	Version    string
	Entries    map[string]ManifestEntry
	lockHandle *file.FileHandle
	closed     bool
}

func parseManifest(fh *file.FileHandle) (map[string]ManifestEntry, error) {
	header, err := file.FileReadLine(fh)
	if err != nil {
		return nil, fmt.Errorf("reading manifest header: %w", err)
	}
	if header != "code-from-spec: v5" {
		return nil, fmt.Errorf("manifest format error: unexpected header")
	}

	entries := make(map[string]ManifestEntry)
	for {
		line, err := file.FileReadLine(fh)
		if errors.Is(err, file.ErrEndOfFile) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading manifest line: %w", err)
		}

		fields := strings.SplitN(line, ";", 4)
		if len(fields) < 4 {
			continue
		}

		name := fields[0]
		pathVal := strings.TrimPrefix(fields[1], "path:")
		checksum := strings.TrimPrefix(fields[2], "checksum:")
		chain := strings.TrimPrefix(fields[3], "chain:")

		entries[name] = ManifestEntry{
			Path:      pathVal,
			Checksum:  checksum,
			ChainHash: chain,
		}
	}

	return entries, nil
}

func ManifestOpen(mode string) (*ManifestHandle, error) {
	if mode != "read" && mode != "write" {
		return nil, ErrInvalidMode
	}

	if mode == "read" {
		manifestPath := pathutils.PathCfs{Value: "code-from-spec/.manifest"}
		manifestFH, err := file.FileOpen(manifestPath, "read", 30000)
		if err != nil {
			if errors.Is(err, file.ErrFileUnreadable) {
				return &ManifestHandle{
					Mode:    "read",
					Version: "v5",
					Entries: make(map[string]ManifestEntry),
				}, nil
			}
			return nil, fmt.Errorf("opening manifest: %w", err)
		}

		lockPath := pathutils.PathCfs{Value: "code-from-spec/.manifest.lock"}
		lockFH, err := file.FileOpen(lockPath, "read", 30000)
		if err != nil {
			if errors.Is(err, file.ErrFileUnreadable) {
				appendFH, appendErr := file.FileOpen(lockPath, "append", 0)
				if appendErr == nil {
					file.FileClose(appendFH)
				}

				lockFH, err = file.FileOpen(lockPath, "read", 30000)
				if err != nil {
					return nil, fmt.Errorf("opening lock file (retry): %w", err)
				}
			} else if errors.Is(err, file.ErrLockTimeout) {
				return nil, ErrLockTimeout
			} else {
				return nil, fmt.Errorf("opening lock file: %w", err)
			}
		}

		entries, err := parseManifest(manifestFH)
		if err != nil {
			file.FileClose(lockFH)
			file.FileClose(manifestFH)
			return nil, err
		}

		file.FileClose(lockFH)
		file.FileClose(manifestFH)

		return &ManifestHandle{
			Mode:    "read",
			Version: "v5",
			Entries: entries,
		}, nil
	}

	lockPath := pathutils.PathCfs{Value: "code-from-spec/.manifest.lock"}
	lockFH, err := file.FileOpen(lockPath, "append", 30000)
	if err != nil {
		if errors.Is(err, file.ErrLockTimeout) {
			return nil, ErrLockTimeout
		}
		return nil, fmt.Errorf("acquiring write lock: %w", err)
	}

	manifestPath := pathutils.PathCfs{Value: "code-from-spec/.manifest"}
	manifestFH, err := file.FileOpen(manifestPath, "read", 30000)
	if err != nil {
		if errors.Is(err, file.ErrFileUnreadable) {
			return &ManifestHandle{
				Mode:       "write",
				Version:    "v5",
				Entries:    make(map[string]ManifestEntry),
				lockHandle: lockFH,
			}, nil
		}
		file.FileClose(lockFH)
		return nil, fmt.Errorf("opening manifest for write: %w", err)
	}

	entries, err := parseManifest(manifestFH)
	if err != nil {
		file.FileClose(manifestFH)
		file.FileClose(lockFH)
		return nil, err
	}
	file.FileClose(manifestFH)

	return &ManifestHandle{
		Mode:       "write",
		Version:    "v5",
		Entries:    entries,
		lockHandle: lockFH,
	}, nil
}

func ManifestSave(handle *ManifestHandle) error {
	if handle.Mode == "read" {
		return ErrWrongMode
	}
	if handle.lockHandle == nil {
		return ErrHandleClosed
	}

	manifestPath := pathutils.PathCfs{Value: "code-from-spec/.manifest"}
	fh, err := file.FileOpen(manifestPath, "overwrite", 30000)
	if err != nil {
		return fmt.Errorf("opening manifest for save: %w", err)
	}

	if err := file.FileWrite(fh, "code-from-spec: v5\n"); err != nil {
		file.FileClose(fh)
		return fmt.Errorf("writing manifest header: %w", err)
	}

	keys := make([]string, 0, len(handle.Entries))
	for k := range handle.Entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		entry := handle.Entries[key]
		line := fmt.Sprintf("%s;path:%s;checksum:%s;chain:%s\n", key, entry.Path, entry.Checksum, entry.ChainHash)
		if err := file.FileWrite(fh, line); err != nil {
			file.FileClose(fh)
			return fmt.Errorf("writing manifest entry: %w", err)
		}
	}

	file.FileClose(fh)
	file.FileClose(handle.lockHandle)
	handle.closed = true
	handle.lockHandle = nil

	return nil
}

func ManifestDiscard(handle *ManifestHandle) error {
	if handle.Mode == "read" {
		return ErrWrongMode
	}
	if handle.lockHandle == nil {
		return ErrHandleClosed
	}

	file.FileClose(handle.lockHandle)
	handle.closed = true
	handle.lockHandle = nil

	return nil
}
