package cache_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/cache"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"
)

const testTimeoutMs = 1000

func readFileLines(t *testing.T, path oslayer.CfsPath) []string {
	t.Helper()
	f, err := oslayer.OpenFile(path, "read", testTimeoutMs)
	if err != nil {
		t.Fatalf("OpenFile(%q) error: %v", path, err)
	}
	defer f.Close()

	var lines []string
	for {
		line, err := f.ReadLine()
		if err != nil {
			if errors.Is(err, oslayer.ErrEndOfFile) {
				break
			}
			t.Fatalf("ReadLine error: %v", err)
		}
		lines = append(lines, line)
	}
	return lines
}

func writeRawFile(t *testing.T, path oslayer.CfsPath, content string) {
	t.Helper()
	f, err := oslayer.OpenFile(path, "overwrite", testTimeoutMs)
	if err != nil {
		t.Fatalf("OpenFile(%q) error: %v", path, err)
	}
	defer f.Close()
	if err := f.Write(content); err != nil {
		t.Fatalf("Write error: %v", err)
	}
}

func assertStringSlicesEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d lines, want %d: got=%v want=%v", len(got), len(want), got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("line %d: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestWriteContentWritesContentFile(t *testing.T) {
	testutils.Chdir(t)

	if err := cache.WriteContent("abcdefghijklmnopqrstuvwxyza", "hello world\n"); err != nil {
		t.Fatalf("WriteContent error: %v", err)
	}

	lines := readFileLines(t, oslayer.CfsPath("code-from-spec/.cache/.content/.abcdefghijklmnopqrstuvwxyza"))
	assertStringSlicesEqual(t, lines, []string{"hello world"})
}

func TestWriteContentSkipsExisting(t *testing.T) {
	testutils.Chdir(t)

	if err := cache.WriteContent("abcdefghijklmnopqrstuvwxyza", "first"); err != nil {
		t.Fatalf("WriteContent error: %v", err)
	}

	if err := cache.WriteContent("abcdefghijklmnopqrstuvwxyza", "second"); err != nil {
		t.Fatalf("WriteContent error: %v", err)
	}

	lines := readFileLines(t, oslayer.CfsPath("code-from-spec/.cache/.content/.abcdefghijklmnopqrstuvwxyza"))
	assertStringSlicesEqual(t, lines, []string{"first"})
}

func TestWriteChainWritesChainFile(t *testing.T) {
	testutils.Chdir(t)

	positions := []chainhash.ContentHash{
		{Label: "SPEC/root", Hash: "aaaaaaaaaaaaaaaaaaaaaaaaaa1"},
		{Label: "SPEC/root/a", Hash: "bbbbbbbbbbbbbbbbbbbbbbbbbbb"},
		{Label: "AGENT[SPEC/root/a]", Hash: "ccccccccccccccccccccccccccc"},
	}

	if err := cache.WriteChain("zyxwvutsrqponmlkjihgfedcbaz", positions); err != nil {
		t.Fatalf("WriteChain error: %v", err)
	}

	lines := readFileLines(t, oslayer.CfsPath("code-from-spec/.cache/.chains/.zyxwvutsrqponmlkjihgfedcbaz"))
	assertStringSlicesEqual(t, lines, []string{
		"SPEC/root: aaaaaaaaaaaaaaaaaaaaaaaaaa1",
		"SPEC/root/a: bbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"AGENT[SPEC/root/a]: ccccccccccccccccccccccccccc",
	})
}

func TestWriteChainSkipsExisting(t *testing.T) {
	testutils.Chdir(t)

	original := []chainhash.ContentHash{
		{Label: "SPEC/root", Hash: "aaaaaaaaaaaaaaaaaaaaaaaaaa1"},
	}
	if err := cache.WriteChain("zyxwvutsrqponmlkjihgfedcbaz", original); err != nil {
		t.Fatalf("WriteChain error: %v", err)
	}

	different := []chainhash.ContentHash{
		{Label: "SPEC/other", Hash: "ddddddddddddddddddddddddddd"},
		{Label: "SPEC/other/b", Hash: "eeeeeeeeeeeeeeeeeeeeeeeeeee"},
	}
	if err := cache.WriteChain("zyxwvutsrqponmlkjihgfedcbaz", different); err != nil {
		t.Fatalf("WriteChain error: %v", err)
	}

	lines := readFileLines(t, oslayer.CfsPath("code-from-spec/.cache/.chains/.zyxwvutsrqponmlkjihgfedcbaz"))
	assertStringSlicesEqual(t, lines, []string{
		"SPEC/root: aaaaaaaaaaaaaaaaaaaaaaaaaa1",
	})
}

func TestReadContentReadsExisting(t *testing.T) {
	testutils.Chdir(t)

	if err := cache.WriteContent("abcdefghijklmnopqrstuvwxyza", "test content\n"); err != nil {
		t.Fatalf("WriteContent error: %v", err)
	}

	content, err := cache.ReadContent("abcdefghijklmnopqrstuvwxyza")
	if err != nil {
		t.Fatalf("ReadContent error: %v", err)
	}
	if content != "test content\n" {
		t.Fatalf("got %q, want %q", content, "test content\n")
	}
}

func TestReadContentReturnsErrNotFoundForMissing(t *testing.T) {
	testutils.Chdir(t)

	_, err := cache.ReadContent("nonexistenthashvalue12345ab")
	if !errors.Is(err, cache.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestReadChainReadsExisting(t *testing.T) {
	testutils.Chdir(t)

	positions := []chainhash.ContentHash{
		{Label: "SPEC/root", Hash: "aaaaaaaaaaaaaaaaaaaaaaaaaa1"},
		{Label: "SPEC/root/a", Hash: "bbbbbbbbbbbbbbbbbbbbbbbbbbb"},
		{Label: "AGENT[SPEC/root/a]", Hash: "ccccccccccccccccccccccccccc"},
	}
	if err := cache.WriteChain("zyxwvutsrqponmlkjihgfedcbaz", positions); err != nil {
		t.Fatalf("WriteChain error: %v", err)
	}

	got, err := cache.ReadChain("zyxwvutsrqponmlkjihgfedcbaz")
	if err != nil {
		t.Fatalf("ReadChain error: %v", err)
	}
	if len(got) != len(positions) {
		t.Fatalf("got %d positions, want %d", len(got), len(positions))
	}
	for i := range positions {
		if got[i] != positions[i] {
			t.Fatalf("position %d: got %v, want %v", i, got[i], positions[i])
		}
	}
}

func TestReadChainReturnsErrNotFoundForMissing(t *testing.T) {
	testutils.Chdir(t)

	_, err := cache.ReadChain("nonexistenthashvalue12345ab")
	if !errors.Is(err, cache.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestReadChainReturnsErrChainFileCorruptedForMalformedLine(t *testing.T) {
	testutils.Chdir(t)

	writeRawFile(t, oslayer.CfsPath("code-from-spec/.cache/.chains/.malformedhashvalue12345ab"), "bad line\n")

	_, err := cache.ReadChain("malformedhashvalue12345ab")
	if !errors.Is(err, cache.ErrChainFileCorrupted) {
		t.Fatalf("got %v, want ErrChainFileCorrupted", err)
	}
}

func TestListContentHashesListsHashes(t *testing.T) {
	testutils.Chdir(t)

	if err := cache.WriteContent("aaaaaaaaaaaaaaaaaaaaaaaaaaa", "one"); err != nil {
		t.Fatalf("WriteContent error: %v", err)
	}
	if err := cache.WriteContent("bbbbbbbbbbbbbbbbbbbbbbbbbbb", "two"); err != nil {
		t.Fatalf("WriteContent error: %v", err)
	}

	hashes, err := cache.ListContentHashes()
	if err != nil {
		t.Fatalf("ListContentHashes error: %v", err)
	}

	want := map[string]bool{
		"aaaaaaaaaaaaaaaaaaaaaaaaaaa": true,
		"bbbbbbbbbbbbbbbbbbbbbbbbbbb": true,
	}
	if len(hashes) != len(want) {
		t.Fatalf("got %d hashes, want %d: %v", len(hashes), len(want), hashes)
	}
	for _, h := range hashes {
		if !want[h] {
			t.Fatalf("unexpected hash %q in %v", h, hashes)
		}
	}
}

func TestListContentHashesEmptyDirectoryReturnsNil(t *testing.T) {
	testutils.Chdir(t)

	hashes, err := cache.ListContentHashes()
	if err != nil {
		t.Fatalf("ListContentHashes error: %v", err)
	}
	if hashes != nil {
		t.Fatalf("got %v, want nil", hashes)
	}
}

func TestListChainHashesListsHashes(t *testing.T) {
	testutils.Chdir(t)

	positions := []chainhash.ContentHash{
		{Label: "SPEC/root", Hash: "aaaaaaaaaaaaaaaaaaaaaaaaaa1"},
	}
	if err := cache.WriteChain("aaaaaaaaaaaaaaaaaaaaaaaaaaa", positions); err != nil {
		t.Fatalf("WriteChain error: %v", err)
	}
	if err := cache.WriteChain("bbbbbbbbbbbbbbbbbbbbbbbbbbb", positions); err != nil {
		t.Fatalf("WriteChain error: %v", err)
	}

	hashes, err := cache.ListChainHashes()
	if err != nil {
		t.Fatalf("ListChainHashes error: %v", err)
	}

	want := map[string]bool{
		"aaaaaaaaaaaaaaaaaaaaaaaaaaa": true,
		"bbbbbbbbbbbbbbbbbbbbbbbbbbb": true,
	}
	if len(hashes) != len(want) {
		t.Fatalf("got %d hashes, want %d: %v", len(hashes), len(want), hashes)
	}
	for _, h := range hashes {
		if !want[h] {
			t.Fatalf("unexpected hash %q in %v", h, hashes)
		}
	}
}

func TestDeleteContentDeletesFile(t *testing.T) {
	testutils.Chdir(t)

	if err := cache.WriteContent("abcdefghijklmnopqrstuvwxyza", "content"); err != nil {
		t.Fatalf("WriteContent error: %v", err)
	}

	if err := cache.DeleteContent("abcdefghijklmnopqrstuvwxyza"); err != nil {
		t.Fatalf("DeleteContent error: %v", err)
	}

	_, err := cache.ReadContent("abcdefghijklmnopqrstuvwxyza")
	if !errors.Is(err, cache.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestDeleteChainDeletesFile(t *testing.T) {
	testutils.Chdir(t)

	positions := []chainhash.ContentHash{
		{Label: "SPEC/root", Hash: "aaaaaaaaaaaaaaaaaaaaaaaaaa1"},
	}
	if err := cache.WriteChain("zyxwvutsrqponmlkjihgfedcbaz", positions); err != nil {
		t.Fatalf("WriteChain error: %v", err)
	}

	if err := cache.DeleteChain("zyxwvutsrqponmlkjihgfedcbaz"); err != nil {
		t.Fatalf("DeleteChain error: %v", err)
	}

	_, err := cache.ReadChain("zyxwvutsrqponmlkjihgfedcbaz")
	if !errors.Is(err, cache.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestListContentHashesFiltersTemporaryFiles(t *testing.T) {
	testutils.Chdir(t)

	if err := cache.WriteContent("abcdefghijklmnopqrstuvwxyza", "content"); err != nil {
		t.Fatalf("WriteContent error: %v", err)
	}

	writeRawFile(t, oslayer.CfsPath("code-from-spec/.cache/.content/._tmp_somehash"), "temp")

	hashes, err := cache.ListContentHashes()
	if err != nil {
		t.Fatalf("ListContentHashes error: %v", err)
	}

	assertStringSlicesEqual(t, hashes, []string{"abcdefghijklmnopqrstuvwxyza"})
}
