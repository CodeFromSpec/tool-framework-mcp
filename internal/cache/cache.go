package cache

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
)

const (
	contentDir = "code-from-spec/.cache/.content"
	chainDir   = "code-from-spec/.cache/.chains"
)

type ChainPosition struct {
	Label       string
	ContentHash string
}

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

func WriteChain(chainHash string, positions []ChainPosition) error {
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
		line := position.Label + ": " + position.ContentHash + "\n"
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

func ReadContent(contentHash string) (string, bool, error) {
	targetPath := oslayer.CfsPath(contentDir + "/." + contentHash)
	f, err := oslayer.OpenFile(targetPath, "read", 30000)
	if err != nil {
		if errors.Is(err, oslayer.ErrFileUnreadable) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("opening content file: %w", err)
	}
	defer f.Close()

	var lines []string
	for {
		line, err := f.ReadLine()
		if err != nil {
			if errors.Is(err, oslayer.ErrEndOfFile) {
				break
			}
			return "", false, fmt.Errorf("reading content file: %w", err)
		}
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n") + "\n"
	return content, true, nil
}

func ReadChain(chainHash string) ([]ChainPosition, bool, error) {
	targetPath := oslayer.CfsPath(chainDir + "/." + chainHash)
	f, err := oslayer.OpenFile(targetPath, "read", 30000)
	if err != nil {
		if errors.Is(err, oslayer.ErrFileUnreadable) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("opening chain file: %w", err)
	}
	defer f.Close()

	var positions []ChainPosition
	for {
		line, err := f.ReadLine()
		if err != nil {
			if errors.Is(err, oslayer.ErrEndOfFile) {
				break
			}
			return nil, false, fmt.Errorf("reading chain file: %w", err)
		}
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) != 2 {
			continue
		}
		positions = append(positions, ChainPosition{
			Label:       parts[0],
			ContentHash: parts[1],
		})
	}

	return positions, true, nil
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
