// code-from-spec: SPEC/golang/implementation/manifest@T0ZodL6lfBB_Jwi-d_b_2aKiDt4
package manifest

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
)

var ErrLockTimeout = errors.New("lock timeout")
var ErrReadOnly = errors.New("read only")
var ErrManifestClosed = errors.New("manifest closed")
var ErrManifestFormatError = errors.New("manifest format error")

type ManifestEntry struct {
	Path      string
	Checksum  string
	ChainHash string
}

type Manifest struct {
	Version  string
	Entries  map[string]ManifestEntry
	readOnly bool
	closed   bool
	lockFile *oslayer.File
}

func parseManifest(f *oslayer.File) (map[string]ManifestEntry, error) {
	header, err := f.ReadLine()
	if err != nil {
		return nil, fmt.Errorf("reading manifest header: %w", err)
	}
	if header != "code-from-spec: v5" {
		return nil, ErrManifestFormatError
	}

	entries := make(map[string]ManifestEntry)
	for {
		line, err := f.ReadLine()
		if errors.Is(err, oslayer.ErrEndOfFile) {
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

func OpenManifest(readOnly bool) (*Manifest, error) {
	if readOnly {
		manifestFH, err := oslayer.OpenFile(oslayer.CfsPath("code-from-spec/.manifest"), "read", 30000)
		if err != nil {
			if errors.Is(err, oslayer.ErrFileUnreadable) {
				return &Manifest{
					readOnly: true,
					Version:  "v5",
					Entries:  make(map[string]ManifestEntry),
				}, nil
			}
			return nil, fmt.Errorf("opening manifest: %w", err)
		}

		lockPath := oslayer.CfsPath("code-from-spec/.manifest.lock")
		lockFH, err := oslayer.OpenFile(lockPath, "read", 30000)
		if err != nil {
			if errors.Is(err, oslayer.ErrFileUnreadable) {
				appendFH, appendErr := oslayer.OpenFile(lockPath, "append", 0)
				if appendErr == nil {
					appendFH.Close()
				}

				lockFH, err = oslayer.OpenFile(lockPath, "read", 30000)
				if err != nil {
					return nil, fmt.Errorf("opening lock file (retry): %w", err)
				}
			} else if errors.Is(err, oslayer.ErrLockTimeout) {
				return nil, ErrLockTimeout
			} else {
				return nil, fmt.Errorf("opening lock file: %w", err)
			}
		}

		entries, err := parseManifest(manifestFH)
		if err != nil {
			lockFH.Close()
			manifestFH.Close()
			return nil, err
		}

		lockFH.Close()
		manifestFH.Close()

		return &Manifest{
			readOnly: true,
			Version:  "v5",
			Entries:  entries,
		}, nil
	}

	lockPath := oslayer.CfsPath("code-from-spec/.manifest.lock")
	lockFH, err := oslayer.OpenFile(lockPath, "append", 30000)
	if err != nil {
		if errors.Is(err, oslayer.ErrLockTimeout) {
			return nil, ErrLockTimeout
		}
		return nil, fmt.Errorf("acquiring write lock: %w", err)
	}

	manifestFH, err := oslayer.OpenFile(oslayer.CfsPath("code-from-spec/.manifest"), "read", 30000)
	if err != nil {
		if errors.Is(err, oslayer.ErrFileUnreadable) {
			return &Manifest{
				readOnly: false,
				Version:  "v5",
				Entries:  make(map[string]ManifestEntry),
				lockFile: lockFH,
			}, nil
		}
		lockFH.Close()
		return nil, fmt.Errorf("opening manifest for write: %w", err)
	}

	entries, err := parseManifest(manifestFH)
	if err != nil {
		manifestFH.Close()
		lockFH.Close()
		return nil, err
	}
	manifestFH.Close()

	return &Manifest{
		readOnly: false,
		Version:  "v5",
		Entries:  entries,
		lockFile: lockFH,
	}, nil
}

func (m *Manifest) Save() error {
	if m.readOnly {
		return ErrReadOnly
	}
	if m.closed {
		return ErrManifestClosed
	}

	fh, err := oslayer.OpenFile(oslayer.CfsPath("code-from-spec/.manifest"), "overwrite", 30000)
	if err != nil {
		return fmt.Errorf("opening manifest for save: %w", err)
	}

	if err := fh.Write("code-from-spec: v5\n"); err != nil {
		fh.Close()
		return fmt.Errorf("writing manifest header: %w", err)
	}

	keys := make([]string, 0, len(m.Entries))
	for k := range m.Entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		entry := m.Entries[key]
		line := fmt.Sprintf("%s;path:%s;checksum:%s;chain:%s\n", key, entry.Path, entry.Checksum, entry.ChainHash)
		if err := fh.Write(line); err != nil {
			fh.Close()
			return fmt.Errorf("writing manifest entry: %w", err)
		}
	}

	fh.Close()
	m.lockFile.Close()
	m.closed = true

	return nil
}

func (m *Manifest) Discard() error {
	if m.readOnly {
		return ErrReadOnly
	}
	if m.closed {
		return ErrManifestClosed
	}

	m.lockFile.Close()
	m.closed = true

	return nil
}
