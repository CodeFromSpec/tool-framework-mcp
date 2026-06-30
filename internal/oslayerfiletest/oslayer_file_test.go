// code-from-spec: SPEC/golang/test/cases/oslayer/file@1LULl25XjfYLmSibwDVWqV2UHks
package oslayerfiletest

import (
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"
)

func writeFile(t *testing.T, name string, content []byte) {
	t.Helper()
	if err := os.WriteFile(name, content, 0644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
}

func readFile(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("readFile: %v", err)
	}
	return data
}

func TestReadMode_OpensAndReadsAllLines(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "f.txt", []byte("alpha\nbeta\ngamma\n"))
	f, err := oslayer.OpenFile("f.txt", "read", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()
	for _, want := range []string{"alpha", "beta", "gamma"} {
		got, err := f.ReadLine()
		if err != nil {
			t.Fatalf("ReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
	_, err = f.ReadLine()
	if !errors.Is(err, oslayer.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestReadMode_NormalizesCRLF(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "f.txt", []byte("alpha\r\nbeta\r\n"))
	f, err := oslayer.OpenFile("f.txt", "read", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()
	for _, want := range []string{"alpha", "beta"} {
		got, err := f.ReadLine()
		if err != nil {
			t.Fatalf("ReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestReadMode_NoTrailingNewline(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "f.txt", []byte("alpha\nbeta"))
	f, err := oslayer.OpenFile("f.txt", "read", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()
	for _, want := range []string{"alpha", "beta"} {
		got, err := f.ReadLine()
		if err != nil {
			t.Fatalf("ReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
	_, err = f.ReadLine()
	if !errors.Is(err, oslayer.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestReadMode_SkipLinesAdvancesReader(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "f.txt", []byte("one\ntwo\nthree\nfour\nfive\n"))
	f, err := oslayer.OpenFile("f.txt", "read", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()
	if err := f.SkipLines(2); err != nil {
		t.Fatalf("SkipLines: %v", err)
	}
	got, err := f.ReadLine()
	if err != nil {
		t.Fatalf("ReadLine: %v", err)
	}
	if got != "three" {
		t.Errorf("got %q, want %q", got, "three")
	}
}

func TestReadMode_SkipLinesPastEndOfFile(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "f.txt", []byte("one\ntwo\n"))
	f, err := oslayer.OpenFile("f.txt", "read", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()
	if err := f.SkipLines(10); err != nil {
		t.Errorf("SkipLines: expected no error, got %v", err)
	}
	_, err = f.ReadLine()
	if !errors.Is(err, oslayer.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestReadMode_PreservesLeadingWhitespace(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "f.txt", []byte("  alpha\n    beta\n"))
	f, err := oslayer.OpenFile("f.txt", "read", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()
	for _, want := range []string{"  alpha", "    beta"} {
		got, err := f.ReadLine()
		if err != nil {
			t.Fatalf("ReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestReadMode_PreservesTrailingWhitespace(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "f.txt", []byte("alpha  \nbeta   \n"))
	f, err := oslayer.OpenFile("f.txt", "read", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()
	for _, want := range []string{"alpha  ", "beta   "} {
		got, err := f.ReadLine()
		if err != nil {
			t.Fatalf("ReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestReadMode_PreservesInternalWhitespace(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "f.txt", []byte("alpha   beta\none\ttwo\n"))
	f, err := oslayer.OpenFile("f.txt", "read", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()
	for _, want := range []string{"alpha   beta", "one\ttwo"} {
		got, err := f.ReadLine()
		if err != nil {
			t.Fatalf("ReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestReadMode_PreservesEmptyLines(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "f.txt", []byte("alpha\n\n\nbeta\n"))
	f, err := oslayer.OpenFile("f.txt", "read", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()
	for _, want := range []string{"alpha", "", "", "beta"} {
		got, err := f.ReadLine()
		if err != nil {
			t.Fatalf("ReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestReadMode_PreservesNonASCII(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "f.txt", []byte("café\n日本語\n🎉🚀\n"))
	f, err := oslayer.OpenFile("f.txt", "read", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()
	for _, want := range []string{"café", "日本語", "🎉🚀"} {
		got, err := f.ReadLine()
		if err != nil {
			t.Fatalf("ReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestReadMode_EmptyFile(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "f.txt", []byte{})
	f, err := oslayer.OpenFile("f.txt", "read", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()
	_, err = f.ReadLine()
	if !errors.Is(err, oslayer.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestReadMode_SingleLineWithoutNewline(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "f.txt", []byte("hello"))
	f, err := oslayer.OpenFile("f.txt", "read", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()
	got, err := f.ReadLine()
	if err != nil {
		t.Fatalf("ReadLine: %v", err)
	}
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
	_, err = f.ReadLine()
	if !errors.Is(err, oslayer.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestReadMode_FileDoesNotExist(t *testing.T) {
	testutils.Chdir(t)
	_, err := oslayer.OpenFile("nonexistent/file.txt", "read", 5000)
	if !errors.Is(err, oslayer.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestReadMode_ReadAfterClose(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "f.txt", []byte("alpha\n"))
	f, err := oslayer.OpenFile("f.txt", "read", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	f.Close()
	_, err = f.ReadLine()
	if !errors.Is(err, oslayer.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestReadMode_SkipAfterClose(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "f.txt", []byte("alpha\n"))
	f, err := oslayer.OpenFile("f.txt", "read", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	f.Close()
	if err := f.SkipLines(1); err != nil {
		t.Errorf("SkipLines after close: expected no error, got %v", err)
	}
}

func TestOverwriteMode_WritesContentToNewFile(t *testing.T) {
	testutils.Chdir(t)
	f, err := oslayer.OpenFile("f.txt", "overwrite", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	if err := f.Write("hello world"); err != nil {
		t.Fatalf("Write: %v", err)
	}
	f.Close()
	got := readFile(t, "f.txt")
	if string(got) != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}

func TestOverwriteMode_OverwritesExistingFile(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "f.txt", []byte("old"))
	f, err := oslayer.OpenFile("f.txt", "overwrite", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	if err := f.Write("new"); err != nil {
		t.Fatalf("Write: %v", err)
	}
	f.Close()
	got := readFile(t, "f.txt")
	if string(got) != "new" {
		t.Errorf("got %q, want %q", got, "new")
	}
}

func TestOverwriteMode_CreatesIntermediateDirectories(t *testing.T) {
	testutils.Chdir(t)
	f, err := oslayer.OpenFile("a/b/c/file.txt", "overwrite", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	if err := f.Write("data"); err != nil {
		t.Fatalf("Write: %v", err)
	}
	f.Close()
	got := readFile(t, "a/b/c/file.txt")
	if string(got) != "data" {
		t.Errorf("got %q, want %q", got, "data")
	}
}

func TestOverwriteMode_PreservesUTF8(t *testing.T) {
	testutils.Chdir(t)
	content := "café 日本語 🎉"
	f, err := oslayer.OpenFile("f.txt", "overwrite", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	if err := f.Write(content); err != nil {
		t.Fatalf("Write: %v", err)
	}
	f.Close()
	got := readFile(t, "f.txt")
	if string(got) != content {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestOverwriteMode_PreservesLineEndingsAsReceived(t *testing.T) {
	testutils.Chdir(t)
	content := "alpha\r\nbeta\r\n"
	f, err := oslayer.OpenFile("f.txt", "overwrite", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	if err := f.Write(content); err != nil {
		t.Fatalf("Write: %v", err)
	}
	f.Close()
	got := readFile(t, "f.txt")
	if string(got) != content {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestOverwriteMode_WritesEmptyContent(t *testing.T) {
	testutils.Chdir(t)
	f, err := oslayer.OpenFile("f.txt", "overwrite", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	if err := f.Write(""); err != nil {
		t.Fatalf("Write: %v", err)
	}
	f.Close()
	got := readFile(t, "f.txt")
	if len(got) != 0 {
		t.Errorf("expected zero bytes, got %d", len(got))
	}
}

func TestOverwriteMode_ValidationError(t *testing.T) {
	testutils.Chdir(t)
	_, err := oslayer.OpenFile("../../outside", "overwrite", 5000)
	if !errors.Is(err, oslayer.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got %v", err)
	}
	if _, statErr := os.Stat("../../outside"); statErr == nil {
		t.Error("file should not have been created")
	}
}

func TestOverwriteMode_CannotCreateDirectory(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "notadir", []byte("data"))
	_, err := oslayer.OpenFile("notadir/file.txt", "overwrite", 5000)
	if !errors.Is(err, oslayer.ErrCannotCreateDirectory) {
		t.Errorf("expected ErrCannotCreateDirectory, got %v", err)
	}
}

func TestOverwriteMode_CannotOpenFilePathIsDirectory(t *testing.T) {
	testutils.Chdir(t)
	if err := os.Mkdir("mydir", 0755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	_, err := oslayer.OpenFile("mydir", "overwrite", 5000)
	if !errors.Is(err, oslayer.ErrCannotOpenFile) {
		t.Errorf("expected ErrCannotOpenFile, got %v", err)
	}
}

func TestAppendMode_OpensWithoutTruncating(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "f.txt", []byte("old"))
	f, err := oslayer.OpenFile("f.txt", "append", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	f.Close()
	got := readFile(t, "f.txt")
	if string(got) != "old" {
		t.Errorf("got %q, want %q", got, "old")
	}
}

func TestAppendMode_CreatesFileIfNotExists(t *testing.T) {
	testutils.Chdir(t)
	f, err := oslayer.OpenFile("new.txt", "append", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	f.Close()
	if _, err := os.Stat("new.txt"); err != nil {
		t.Errorf("file not created: %v", err)
	}
}

func TestAppendMode_WriteSucceeds(t *testing.T) {
	testutils.Chdir(t)
	f, err := oslayer.OpenFile("f.txt", "append", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	if err := f.Write("content"); err != nil {
		t.Errorf("Write: %v", err)
	}
	f.Close()
	got := readFile(t, "f.txt")
	if string(got) != "content" {
		t.Errorf("got %q, want %q", got, "content")
	}
}

func TestAppendMode_ActuallyAppends(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "f.txt", []byte("old\n"))
	f, err := oslayer.OpenFile("f.txt", "append", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	if err := f.Write("new\n"); err != nil {
		t.Fatalf("Write: %v", err)
	}
	f.Close()
	got := readFile(t, "f.txt")
	if string(got) != "old\nnew\n" {
		t.Errorf("got %q, want %q", got, "old\nnew\n")
	}
}

func TestAppendMode_CreatesIntermediateDirectories(t *testing.T) {
	testutils.Chdir(t)
	f, err := oslayer.OpenFile("x/y/z/file.txt", "append", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	f.Close()
	if _, err := os.Stat("x/y/z/file.txt"); err != nil {
		t.Errorf("file not created: %v", err)
	}
}

func TestWrongMode_ReadLineFailsInOverwrite(t *testing.T) {
	testutils.Chdir(t)
	f, err := oslayer.OpenFile("f.txt", "overwrite", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()
	_, err = f.ReadLine()
	if !errors.Is(err, oslayer.ErrWrongMode) {
		t.Errorf("expected ErrWrongMode, got %v", err)
	}
}

func TestWrongMode_ReadLineFailsInAppend(t *testing.T) {
	testutils.Chdir(t)
	f, err := oslayer.OpenFile("f.txt", "append", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()
	_, err = f.ReadLine()
	if !errors.Is(err, oslayer.ErrWrongMode) {
		t.Errorf("expected ErrWrongMode, got %v", err)
	}
}

func TestWrongMode_WriteFailsInRead(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "f.txt", []byte("data"))
	f, err := oslayer.OpenFile("f.txt", "read", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()
	err = f.Write("x")
	if !errors.Is(err, oslayer.ErrWrongMode) {
		t.Errorf("expected ErrWrongMode, got %v", err)
	}
}

func TestWrongMode_SkipLinesFailsInOverwrite(t *testing.T) {
	testutils.Chdir(t)
	f, err := oslayer.OpenFile("f.txt", "overwrite", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()
	err = f.SkipLines(1)
	if !errors.Is(err, oslayer.ErrWrongMode) {
		t.Errorf("expected ErrWrongMode, got %v", err)
	}
}

func TestWrongMode_SkipLinesFailsInAppend(t *testing.T) {
	testutils.Chdir(t)
	f, err := oslayer.OpenFile("f.txt", "append", 5000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()
	err = f.SkipLines(1)
	if !errors.Is(err, oslayer.ErrWrongMode) {
		t.Errorf("expected ErrWrongMode, got %v", err)
	}
}

func TestInvalidMode_OpenFileRejectsUnknownMode(t *testing.T) {
	testutils.Chdir(t)
	_, err := oslayer.OpenFile("f.txt", "invalid", 5000)
	if !errors.Is(err, oslayer.ErrInvalidMode) {
		t.Errorf("expected ErrInvalidMode, got %v", err)
	}
}

func TestRenameFile_RenamesFile(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "a.txt", []byte("data"))
	if err := oslayer.RenameFile("a.txt", "b.txt"); err != nil {
		t.Fatalf("RenameFile: %v", err)
	}
	got := readFile(t, "b.txt")
	if string(got) != "data" {
		t.Errorf("got %q, want %q", got, "data")
	}
	if _, err := os.Stat("a.txt"); err == nil {
		t.Error("source file still exists")
	}
}

func TestRenameFile_OverwritesDestination(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "dest.txt", []byte("old"))
	writeFile(t, "src.txt", []byte("new"))
	if err := oslayer.RenameFile("src.txt", "dest.txt"); err != nil {
		t.Fatalf("RenameFile: %v", err)
	}
	got := readFile(t, "dest.txt")
	if string(got) != "new" {
		t.Errorf("got %q, want %q", got, "new")
	}
}

func TestRenameFile_NonExistentSource(t *testing.T) {
	testutils.Chdir(t)
	err := oslayer.RenameFile("nonexistent.txt", "dest.txt")
	if !errors.Is(err, oslayer.ErrCannotRename) {
		t.Errorf("expected ErrCannotRename, got %v", err)
	}
}

func TestRenameFile_InvalidCfsPath(t *testing.T) {
	testutils.Chdir(t)
	err := oslayer.RenameFile("../../outside", "dest.txt")
	if !errors.Is(err, oslayer.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got %v", err)
	}
}

func TestDeleteFile_DeletesFile(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "target.txt", []byte("data"))
	if err := oslayer.DeleteFile("target.txt"); err != nil {
		t.Fatalf("DeleteFile: %v", err)
	}
	if _, err := os.Stat("target.txt"); err == nil {
		t.Error("file still exists")
	}
}

func TestDeleteFile_NonExistentFile(t *testing.T) {
	testutils.Chdir(t)
	err := oslayer.DeleteFile("nonexistent.txt")
	if !errors.Is(err, oslayer.ErrCannotDelete) {
		t.Errorf("expected ErrCannotDelete, got %v", err)
	}
}

func TestDeleteFile_InvalidCfsPath(t *testing.T) {
	testutils.Chdir(t)
	err := oslayer.DeleteFile("../../outside")
	if !errors.Is(err, oslayer.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got %v", err)
	}
}

func TestLocking_SharedLockAllowsConcurrentReaders(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "f.txt", []byte("data\n"))
	f1, err := oslayer.OpenFile("f.txt", "read", 5000)
	if err != nil {
		t.Fatalf("OpenFile f1: %v", err)
	}
	defer f1.Close()
	f2, err := oslayer.OpenFile("f.txt", "read", 5000)
	if err != nil {
		t.Errorf("OpenFile f2 (shared lock): %v", err)
	} else {
		f2.Close()
	}
}

func TestLocking_ExclusiveLockBlocksOtherExclusive(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "f.txt", []byte("data\n"))
	f1, err := oslayer.OpenFile("f.txt", "overwrite", 5000)
	if err != nil {
		t.Fatalf("OpenFile f1: %v", err)
	}

	ready := make(chan struct{})
	done := make(chan struct{})
	var mu sync.Mutex
	var acquiredAfterClose bool

	go func() {
		close(ready)
		f2, err := oslayer.OpenFile("f.txt", "overwrite", 5000)
		if err == nil {
			mu.Lock()
			acquiredAfterClose = true
			mu.Unlock()
			f2.Close()
		}
		close(done)
	}()

	<-ready
	time.Sleep(50 * time.Millisecond)
	f1.Close()
	<-done

	mu.Lock()
	got := acquiredAfterClose
	mu.Unlock()
	if !got {
		t.Error("second exclusive lock was never acquired after first closed")
	}
}

func TestLocking_ExclusiveLockBlocksSharedLock(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "f.txt", []byte("data\n"))
	f1, err := oslayer.OpenFile("f.txt", "overwrite", 5000)
	if err != nil {
		t.Fatalf("OpenFile f1: %v", err)
	}

	ready := make(chan struct{})
	done := make(chan struct{})
	var mu sync.Mutex
	var acquiredAfterClose bool

	go func() {
		close(ready)
		f2, err := oslayer.OpenFile("f.txt", "read", 5000)
		if err == nil {
			mu.Lock()
			acquiredAfterClose = true
			mu.Unlock()
			f2.Close()
		}
		close(done)
	}()

	<-ready
	time.Sleep(50 * time.Millisecond)
	f1.Close()
	<-done

	mu.Lock()
	got := acquiredAfterClose
	mu.Unlock()
	if !got {
		t.Error("shared lock was never acquired after exclusive closed")
	}
}

func TestLocking_AppendModeAcquiresExclusiveLock(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "f.txt", []byte("data\n"))
	f1, err := oslayer.OpenFile("f.txt", "append", 5000)
	if err != nil {
		t.Fatalf("OpenFile f1: %v", err)
	}

	ready := make(chan struct{})
	done := make(chan struct{})
	var mu sync.Mutex
	var acquiredAfterClose bool

	go func() {
		close(ready)
		f2, err := oslayer.OpenFile("f.txt", "read", 5000)
		if err == nil {
			mu.Lock()
			acquiredAfterClose = true
			mu.Unlock()
			f2.Close()
		}
		close(done)
	}()

	<-ready
	time.Sleep(50 * time.Millisecond)
	f1.Close()
	<-done

	mu.Lock()
	got := acquiredAfterClose
	mu.Unlock()
	if !got {
		t.Error("shared lock was never acquired after append exclusive closed")
	}
}

func TestLocking_LockTimeout(t *testing.T) {
	testutils.Chdir(t)
	writeFile(t, "f.txt", []byte("data\n"))
	f1, err := oslayer.OpenFile("f.txt", "overwrite", 5000)
	if err != nil {
		t.Fatalf("OpenFile f1: %v", err)
	}
	defer f1.Close()

	_, err = oslayer.OpenFile("f.txt", "overwrite", 50)
	if !errors.Is(err, oslayer.ErrLockTimeout) {
		t.Errorf("expected ErrLockTimeout, got %v", err)
	}
}
