// code-from-spec: SPEC/golang/tests/oslayer/file@CW1qa95QLgp58lNCpyIwv5ktQEo
package oslayer_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
)

func testChdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChdir: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("testChdir cleanup: %v", err)
		}
	})
}

func TestOpenFileRead_OpensAndReadsAllLines(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("file.txt", []byte("alpha\nbeta\ngamma\n"), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := oslayer.OpenFile("file.txt", "read", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()

	lines := []string{"alpha", "beta", "gamma"}
	for _, want := range lines {
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

func TestOpenFileRead_NormalizesCRLF(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("file.txt", []byte("alpha\r\nbeta\r\n"), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := oslayer.OpenFile("file.txt", "read", 1000)
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

func TestOpenFileRead_NoTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("file.txt", []byte("alpha\nbeta"), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := oslayer.OpenFile("file.txt", "read", 1000)
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

func TestOpenFileRead_SkipLinesAdvancesReader(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	content := "one\ntwo\nthree\nfour\nfive\n"
	if err := os.WriteFile("file.txt", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := oslayer.OpenFile("file.txt", "read", 1000)
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

func TestOpenFileRead_SkipLinesPastEndOfFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("file.txt", []byte("one\ntwo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := oslayer.OpenFile("file.txt", "read", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()

	if err := f.SkipLines(10); err != nil {
		t.Fatalf("SkipLines: %v", err)
	}
	_, err = f.ReadLine()
	if !errors.Is(err, oslayer.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestOpenFileRead_PreservesLeadingWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("file.txt", []byte("  alpha\n    beta\n"), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := oslayer.OpenFile("file.txt", "read", 1000)
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

func TestOpenFileRead_PreservesTrailingWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("file.txt", []byte("alpha  \nbeta   \n"), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := oslayer.OpenFile("file.txt", "read", 1000)
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

func TestOpenFileRead_PreservesInternalWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("file.txt", []byte("alpha   beta\none\ttwo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := oslayer.OpenFile("file.txt", "read", 1000)
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

func TestOpenFileRead_PreservesEmptyLines(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("file.txt", []byte("alpha\n\n\nbeta\n"), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := oslayer.OpenFile("file.txt", "read", 1000)
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

func TestOpenFileRead_PreservesNonASCII(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("file.txt", []byte("café\n日本語\n🎉🚀\n"), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := oslayer.OpenFile("file.txt", "read", 1000)
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

func TestOpenFileRead_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("file.txt", []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := oslayer.OpenFile("file.txt", "read", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()

	_, err = f.ReadLine()
	if !errors.Is(err, oslayer.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestOpenFileRead_SingleLineWithoutNewline(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("file.txt", []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := oslayer.OpenFile("file.txt", "read", 1000)
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

func TestOpenFileRead_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	_, err := oslayer.OpenFile("nonexistent/file.txt", "read", 1000)
	if !errors.Is(err, oslayer.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestOpenFileRead_ReadAfterClose(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("file.txt", []byte("alpha\n"), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := oslayer.OpenFile("file.txt", "read", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	f.Close()

	_, err = f.ReadLine()
	if !errors.Is(err, oslayer.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestOpenFileRead_SkipAfterClose(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("file.txt", []byte("alpha\n"), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := oslayer.OpenFile("file.txt", "read", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	f.Close()

	if err := f.SkipLines(1); err != nil {
		t.Errorf("SkipLines after close: %v", err)
	}
}

func TestOpenFileOverwrite_WritesContentToNewFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	f, err := oslayer.OpenFile("file.txt", "overwrite", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	if err := f.Write("hello world"); err != nil {
		t.Fatalf("Write: %v", err)
	}
	f.Close()

	data, err := os.ReadFile("file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello world" {
		t.Errorf("got %q, want %q", string(data), "hello world")
	}
}

func TestOpenFileOverwrite_OverwritesExistingFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("file.txt", []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := oslayer.OpenFile("file.txt", "overwrite", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	if err := f.Write("new"); err != nil {
		t.Fatalf("Write: %v", err)
	}
	f.Close()

	data, err := os.ReadFile("file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new" {
		t.Errorf("got %q, want %q", string(data), "new")
	}
}

func TestOpenFileOverwrite_CreatesIntermediateDirectories(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	f, err := oslayer.OpenFile("a/b/c/file.txt", "overwrite", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	if err := f.Write("content"); err != nil {
		t.Fatalf("Write: %v", err)
	}
	f.Close()

	data, err := os.ReadFile("a/b/c/file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "content" {
		t.Errorf("got %q, want %q", string(data), "content")
	}
}

func TestOpenFileOverwrite_PreservesUTF8Content(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	content := "café 日本語 🎉"
	f, err := oslayer.OpenFile("file.txt", "overwrite", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	if err := f.Write(content); err != nil {
		t.Fatalf("Write: %v", err)
	}
	f.Close()

	data, err := os.ReadFile("file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != content {
		t.Errorf("got %q, want %q", string(data), content)
	}
}

func TestOpenFileOverwrite_PreservesLineEndings(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	content := "alpha\r\nbeta\r\n"
	f, err := oslayer.OpenFile("file.txt", "overwrite", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	if err := f.Write(content); err != nil {
		t.Fatalf("Write: %v", err)
	}
	f.Close()

	data, err := os.ReadFile("file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != content {
		t.Errorf("got %q, want %q", string(data), content)
	}
}

func TestOpenFileOverwrite_WritesEmptyContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	f, err := oslayer.OpenFile("file.txt", "overwrite", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	if err := f.Write(""); err != nil {
		t.Fatalf("Write: %v", err)
	}
	f.Close()

	data, err := os.ReadFile("file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 0 {
		t.Errorf("expected empty file, got %d bytes", len(data))
	}
}

func TestOpenFileOverwrite_PropagatesValidationErrors(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	_, err := oslayer.OpenFile("../../outside", "overwrite", 1000)
	if !errors.Is(err, oslayer.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got %v", err)
	}
}

func TestOpenFileOverwrite_CannotCreateDirectory(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("notadir", []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := oslayer.OpenFile("notadir/file.txt", "overwrite", 1000)
	if !errors.Is(err, oslayer.ErrCannotCreateDirectory) {
		t.Errorf("expected ErrCannotCreateDirectory, got %v", err)
	}
}

func TestOpenFileOverwrite_CannotOpenFilePathIsDirectory(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.Mkdir("mydir", 0755); err != nil {
		t.Fatal(err)
	}
	_, err := oslayer.OpenFile("mydir", "overwrite", 1000)
	if !errors.Is(err, oslayer.ErrCannotOpenFile) {
		t.Errorf("expected ErrCannotOpenFile, got %v", err)
	}
}

func TestOpenFileAppend_OpensWithoutTruncating(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("file.txt", []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := oslayer.OpenFile("file.txt", "append", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	f.Close()

	data, err := os.ReadFile("file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "old" {
		t.Errorf("got %q, want %q", string(data), "old")
	}
}

func TestOpenFileAppend_CreatesFileIfNotExists(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	f, err := oslayer.OpenFile("newfile.txt", "append", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	f.Close()

	if _, err := os.Stat("newfile.txt"); err != nil {
		t.Errorf("expected file to exist: %v", err)
	}
}

func TestOpenFileAppend_WriteSucceeds(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	f, err := oslayer.OpenFile("file.txt", "append", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	if err := f.Write("content"); err != nil {
		t.Fatalf("Write: %v", err)
	}
	f.Close()

	data, err := os.ReadFile("file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "content" {
		t.Errorf("got %q, want %q", string(data), "content")
	}
}

func TestOpenFileAppend_ActuallyAppendsContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("file.txt", []byte("old\n"), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := oslayer.OpenFile("file.txt", "append", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	if err := f.Write("new\n"); err != nil {
		t.Fatalf("Write: %v", err)
	}
	f.Close()

	data, err := os.ReadFile("file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "old\nnew\n" {
		t.Errorf("got %q, want %q", string(data), "old\nnew\n")
	}
}

func TestOpenFileAppend_CreatesIntermediateDirectories(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	f, err := oslayer.OpenFile("x/y/z/file.txt", "append", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	f.Close()

	if _, err := os.Stat("x/y/z/file.txt"); err != nil {
		t.Errorf("expected file to exist: %v", err)
	}
}

func TestOpenFileWrongMode_ReadLineFailsInOverwriteMode(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	f, err := oslayer.OpenFile("file.txt", "overwrite", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()

	_, err = f.ReadLine()
	if !errors.Is(err, oslayer.ErrWrongMode) {
		t.Errorf("expected ErrWrongMode, got %v", err)
	}
}

func TestOpenFileWrongMode_ReadLineFailsInAppendMode(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	f, err := oslayer.OpenFile("file.txt", "append", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()

	_, err = f.ReadLine()
	if !errors.Is(err, oslayer.ErrWrongMode) {
		t.Errorf("expected ErrWrongMode, got %v", err)
	}
}

func TestOpenFileWrongMode_WriteFailsInReadMode(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("file.txt", []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := oslayer.OpenFile("file.txt", "read", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()

	err = f.Write("new content")
	if !errors.Is(err, oslayer.ErrWrongMode) {
		t.Errorf("expected ErrWrongMode, got %v", err)
	}
}

func TestOpenFileWrongMode_SkipLinesFailsInOverwriteMode(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	f, err := oslayer.OpenFile("file.txt", "overwrite", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()

	err = f.SkipLines(1)
	if !errors.Is(err, oslayer.ErrWrongMode) {
		t.Errorf("expected ErrWrongMode, got %v", err)
	}
}

func TestOpenFileWrongMode_SkipLinesFailsInAppendMode(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	f, err := oslayer.OpenFile("file.txt", "append", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()

	err = f.SkipLines(1)
	if !errors.Is(err, oslayer.ErrWrongMode) {
		t.Errorf("expected ErrWrongMode, got %v", err)
	}
}

func TestOpenFileInvalidMode(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	_, err := oslayer.OpenFile("file.txt", "invalid", 1000)
	if !errors.Is(err, oslayer.ErrInvalidMode) {
		t.Errorf("expected ErrInvalidMode, got %v", err)
	}
}

func TestRenameFile_RenamesFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("a.txt", []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := oslayer.RenameFile("a.txt", "b.txt"); err != nil {
		t.Fatalf("RenameFile: %v", err)
	}

	if _, err := os.Stat("a.txt"); !os.IsNotExist(err) {
		t.Error("expected a.txt to be gone")
	}
	data, err := os.ReadFile("b.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "data" {
		t.Errorf("got %q, want %q", string(data), "data")
	}
}

func TestRenameFile_OverwritesDestination(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("dest.txt", []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("src.txt", []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := oslayer.RenameFile("src.txt", "dest.txt"); err != nil {
		t.Fatalf("RenameFile: %v", err)
	}

	data, err := os.ReadFile("dest.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new" {
		t.Errorf("got %q, want %q", string(data), "new")
	}
}

func TestRenameFile_NonExistentSource(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	err := oslayer.RenameFile("nonexistent.txt", "dest.txt")
	if !errors.Is(err, oslayer.ErrCannotRename) {
		t.Errorf("expected ErrCannotRename, got %v", err)
	}
}

func TestRenameFile_InvalidCfsPath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	err := oslayer.RenameFile("../../outside", "dest.txt")
	if !errors.Is(err, oslayer.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got %v", err)
	}
}

func TestDeleteFile_DeletesFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("target.txt", []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := oslayer.DeleteFile("target.txt"); err != nil {
		t.Fatalf("DeleteFile: %v", err)
	}

	if _, err := os.Stat("target.txt"); !os.IsNotExist(err) {
		t.Error("expected target.txt to be gone")
	}
}

func TestDeleteFile_NonExistentFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	err := oslayer.DeleteFile("nonexistent.txt")
	if !errors.Is(err, oslayer.ErrCannotDelete) {
		t.Errorf("expected ErrCannotDelete, got %v", err)
	}
}

func TestDeleteFile_InvalidCfsPath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	err := oslayer.DeleteFile("../../outside")
	if !errors.Is(err, oslayer.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got %v", err)
	}
}

func TestLocking_SharedLockAllowsConcurrentReaders(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("file.txt", []byte("content\n"), 0644); err != nil {
		t.Fatal(err)
	}

	f1, err := oslayer.OpenFile("file.txt", "read", 1000)
	if err != nil {
		t.Fatalf("OpenFile f1: %v", err)
	}
	defer f1.Close()

	f2, err := oslayer.OpenFile("file.txt", "read", 1000)
	if err != nil {
		t.Fatalf("OpenFile f2: %v", err)
	}
	defer f2.Close()
}

func TestLocking_ExclusiveLockBlocksOtherExclusiveLocks(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	f1, err := oslayer.OpenFile("file.txt", "overwrite", 5000)
	if err != nil {
		t.Fatalf("OpenFile f1: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		f2, err := oslayer.OpenFile("file.txt", "overwrite", 5000)
		if err != nil {
			done <- err
			return
		}
		f2.Close()
		done <- nil
	}()

	time.Sleep(100 * time.Millisecond)
	f1.Close()

	if err := <-done; err != nil {
		t.Errorf("second OpenFile failed: %v", err)
	}
}

func TestLocking_ExclusiveLockBlocksSharedLocks(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("file.txt", []byte("content\n"), 0644); err != nil {
		t.Fatal(err)
	}

	f1, err := oslayer.OpenFile("file.txt", "overwrite", 5000)
	if err != nil {
		t.Fatalf("OpenFile f1: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		f2, err := oslayer.OpenFile("file.txt", "read", 5000)
		if err != nil {
			done <- err
			return
		}
		f2.Close()
		done <- nil
	}()

	time.Sleep(100 * time.Millisecond)
	f1.Close()

	if err := <-done; err != nil {
		t.Errorf("second OpenFile failed: %v", err)
	}
}

func TestLocking_AppendModeAcquiresExclusiveLock(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("file.txt", []byte("content\n"), 0644); err != nil {
		t.Fatal(err)
	}

	f1, err := oslayer.OpenFile("file.txt", "append", 5000)
	if err != nil {
		t.Fatalf("OpenFile f1: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		f2, err := oslayer.OpenFile("file.txt", "read", 5000)
		if err != nil {
			done <- err
			return
		}
		f2.Close()
		done <- nil
	}()

	time.Sleep(100 * time.Millisecond)
	f1.Close()

	if err := <-done; err != nil {
		t.Errorf("second OpenFile failed: %v", err)
	}
}

func TestLocking_LockTimeout(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	f1, err := oslayer.OpenFile("file.txt", "overwrite", 5000)
	if err != nil {
		t.Fatalf("OpenFile f1: %v", err)
	}
	defer f1.Close()

	_, err = oslayer.OpenFile("file.txt", "overwrite", 50)
	if !errors.Is(err, oslayer.ErrLockTimeout) {
		t.Errorf("expected ErrLockTimeout, got %v", err)
	}
}
