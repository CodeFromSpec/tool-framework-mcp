package cache

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
)

const (
	contentDir = "code-from-spec/.cache/.content"
	chainDir   = "code-from-spec/.cache/.chains"
)

var (
	ErrNotFound            = errors.New("cache entry not found")
	ErrChainFileCorrupted  = errors.New("chain file corrupted")
)

func fileExists(targetPath oslayer.CfsPath) (bool, error) {
	f, err := oslayer.OpenFile(targetPath, "read", 0)
	if err == nil {
		f.Close()
		return true, nil
	}
	if errors.Is(err, oslayer.ErrFileUnreadable) {
		return false, nil
	}
	if errors.Is(err, oslayer.ErrLockTimeout) {
		return false, nil
	}
	return false, err
}

func WriteContent(contentHash string, content string) error {
	targetPath := oslayer.CfsPath(contentDir + "/." + contentHash)
	exists, err := fileExists(targetPath)
	if err != nil {
		return fmt.Errorf("checking existing content file: %w", err)
	}
	if exists {
		return nil
	}

	tempPath := oslayer.CfsPath(contentDir + "/._tmp_" + contentHash)
	f, err := oslayer.OpenFile(tempPath, "overwrite", 30000)
	if err != nil {
		return fmt.Errorf("opening temporary content file: %w", err)
	}
	if err := f.Write(content); err != nil {
		f.Close()
		return fmt.Errorf("writing content file: %w", err)
	}
	f.Close()

	if err := oslayer.RenameFile(tempPath, targetPath); err != nil {
		return fmt.Errorf("renaming content file: %w", err)
	}
	return nil
}

func WriteChain(chainHash string, positions []chainhash.ContentHash) error {
	targetPath := oslayer.CfsPath(chainDir + "/." + chainHash)
	exists, err := fileExists(targetPath)
	if err != nil {
		return fmt.Errorf("checking existing chain file: %w", err)
	}
	if exists {
		return nil
	}

	tempPath := oslayer.CfsPath(chainDir + "/._tmp_" + chainHash)
	f, err := oslayer.OpenFile(tempPath, "overwrite", 30000)
	if err != nil {
		return fmt.Errorf("opening temporary chain file: %w", err)
	}
	for _, position := range positions {
		line := position.Label + ": " + position.Hash + "\n"
		if err := f.Write(line); err != nil {
			f.Close()
			return fmt.Errorf("writing chain file: %w", err)
		}
	}
	f.Close()

	if err := oslayer.RenameFile(tempPath, targetPath); err != nil {
		return fmt.Errorf("renaming chain file: %w", err)
	}
	return nil
}

func ReadContent(contentHash string) (string, error) {
	targetPath := oslayer.CfsPath(contentDir + "/." + contentHash)
	f, err := oslayer.OpenFile(targetPath, "read", 30000)
	if err != nil {
		if errors.Is(err, oslayer.ErrFileUnreadable) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("opening content file: %w", err)
	}
	defer f.Close()

	var lines []string
	for {
		line, err := f.ReadLine()
		if err != nil {
			if errors.Is(err, oslayer.ErrEndOfFile) {
				break
			}
			return "", fmt.Errorf("reading content file: %w", err)
		}
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n") + "\n"
	return content, nil
}

func ReadChain(chainHash string) ([]chainhash.ContentHash, error) {
	targetPath := oslayer.CfsPath(chainDir + "/." + chainHash)
	f, err := oslayer.OpenFile(targetPath, "read", 30000)
	if err != nil {
		if errors.Is(err, oslayer.ErrFileUnreadable) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("opening chain file: %w", err)
	}
	defer f.Close()

	var positions []chainhash.ContentHash
	for {
		line, err := f.ReadLine()
		if err != nil {
			if errors.Is(err, oslayer.ErrEndOfFile) {
				break
			}
			return nil, fmt.Errorf("reading chain file: %w", err)
		}
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) != 2 {
			return nil, ErrChainFileCorrupted
		}
		positions = append(positions, chainhash.ContentHash{
			Label: parts[0],
			Hash:  parts[1],
		})
	}

	return positions, nil
}

func extractHashes(files []oslayer.CfsPath) []string {
	var hashes []string
	for _, file := range files {
		name := path.Base(string(file))
		if !strings.HasPrefix(name, ".") {
			continue
		}
		if strings.HasPrefix(name, "._tmp_") {
			continue
		}
		hashes = append(hashes, strings.TrimPrefix(name, "."))
	}
	return hashes
}

func ListContentHashes() ([]string, error) {
	files, err := oslayer.ListAllFiles(oslayer.CfsPath(contentDir))
	if err != nil {
		if errors.Is(err, oslayer.ErrDirectoryNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("listing content files: %w", err)
	}
	return extractHashes(files), nil
}

func ListChainHashes() ([]string, error) {
	files, err := oslayer.ListAllFiles(oslayer.CfsPath(chainDir))
	if err != nil {
		if errors.Is(err, oslayer.ErrDirectoryNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("listing chain files: %w", err)
	}
	return extractHashes(files), nil
}

func DeleteContent(contentHash string) error {
	targetPath := oslayer.CfsPath(contentDir + "/." + contentHash)
	if err := oslayer.DeleteFile(targetPath); err != nil {
		return fmt.Errorf("deleting content file: %w", err)
	}
	return nil
}

func DeleteChain(chainHash string) error {
	targetPath := oslayer.CfsPath(chainDir + "/." + chainHash)
	if err := oslayer.DeleteFile(targetPath); err != nil {
		return fmt.Errorf("deleting chain file: %w", err)
	}
	return nil
}
