// code-from-spec: SPEC/golang/tests/oslayer/file@45UeSXzqIbpUfpCAprmpd4fkq5w
package oslayerfiletest_test

import (
	"errors"
	"os"
	"testing"

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

func TestReadMode_OpensAndReadsAllLines(t *testing.T) {
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

func TestReadMode_NoTrailingNewline(t *testing.T) {
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

func TestReadMode_SkipLinesAdvancesReader(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("file.txt", []byte("one\ntwo\nthree\nfour\nfive\n"), 0644); err != nil {
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

func TestReadMode_SkipLinesPastEndOfFile(t *testing.T) {
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
		t.Errorf("SkipLines past EOF should not error, got %v", err)
	}
	_, err = f.ReadLine()
	if !errors.Is(err, oslayer.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestReadMode_PreservesLeadingWhitespace(t *testing.T) {
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

func TestReadMode_PreservesTrailingWhitespace(t *testing.T) {
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

func TestReadMode_PreservesInternalWhitespace(t *testing.T) {
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

func TestReadMode_PreservesEmptyLines(t *testing.T) {
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

func TestReadMode_PreservesNonASCII(t *testing.T) {
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

func TestReadMode_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("file.txt", []byte{}, 0644); err != nil {
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

func TestReadMode_SingleLineWithoutNewline(t *testing.T) {
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

func TestReadMode_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	_, err := oslayer.OpenFile("nonexistent/file.txt", "read", 1000)
	if !errors.Is(err, oslayer.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestReadMode_ReadAfterClose(t *testing.T) {
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
		t.Errorf("expected ErrEndOfFile after close, got %v", err)
	}
}

func TestReadMode_SkipAfterClose(t *testing.T) {
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
		t.Errorf("SkipLines after close should not error, got %v", err)
	}
}

func TestOverwriteMode_WritesContentToNewFile(t *testing.T) {
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

func TestOverwriteMode_OverwritesExistingFile(t *testing.T) {
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

func TestOverwriteMode_CreatesIntermediateDirectories(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	f, err := oslayer.OpenFile("a/b/c/file.txt", "overwrite", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	f.Close()
	if _, err := os.Stat("a/b/c/file.txt"); err != nil {
		t.Errorf("file not found: %v", err)
	}
}

func TestOverwriteMode_PreservesUTF8Content(t *testing.T) {
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

func TestOverwriteMode_PreservesLineEndingsAsReceived(t *testing.T) {
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

func TestOverwriteMode_WritesEmptyContent(t *testing.T) {
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
		t.Errorf("expected empty file, got %q", string(data))
	}
}

func TestOverwriteMode_PropagatesValidationErrors(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	_, err := oslayer.OpenFile("../../outside", "overwrite", 1000)
	if !errors.Is(err, oslayer.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got %v", err)
	}
}

func TestOverwriteMode_CannotCreateDirectory(t *testing.T) {
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

func TestOverwriteMode_CannotOpenFilePathIsDirectory(t *testing.T) {
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

func TestAppendMode_OpensWithoutTruncating(t *testing.T) {
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

func TestAppendMode_CreatesFileIfNotExists(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	f, err := oslayer.OpenFile("newfile.txt", "append", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	f.Close()
	if _, err := os.Stat("newfile.txt"); err != nil {
		t.Errorf("file not found: %v", err)
	}
}

func TestAppendMode_WriteSucceeds(t *testing.T) {
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

func TestAppendMode_ActuallyAppendsContent(t *testing.T) {
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

func TestAppendMode_CreatesIntermediateDirectories(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	f, err := oslayer.OpenFile("x/y/z/file.txt", "append", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	f.Close()
	if _, err := os.Stat("x/y/z/file.txt"); err != nil {
		t.Errorf("file not found: %v", err)
	}
}

func TestWrongMode_ReadLineFailsInOverwriteMode(t *testing.T) {
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

func TestWrongMode_ReadLineFailsInAppendMode(t *testing.T) {
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

func TestWrongMode_WriteFailsInReadMode(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("file.txt", []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := oslayer.OpenFile("file.txt", "read", 1000)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()
	err = f.Write("something")
	if !errors.Is(err, oslayer.ErrWrongMode) {
		t.Errorf("expected ErrWrongMode, got %v", err)
	}
}

func TestWrongMode_SkipLinesFailsInOverwriteMode(t *testing.T) {
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

func TestWrongMode_SkipLinesFailsInAppendMode(t *testing.T) {
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

func TestInvalidMode_RejectsUnknownMode(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	_, err := oslayer.OpenFile("file.txt", "invalid", 1000)
	if !errors.Is(err, oslayer.ErrInvalidMode) {
		t.Errorf("expected ErrInvalidMode, got %v", err)
	}
}

func TestRenameFile_RenamesAFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("a.txt", []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := oslayer.RenameFile("a.txt", "b.txt"); err != nil {
		t.Fatalf("RenameFile: %v", err)
	}
	if _, err := os.Stat("a.txt"); !os.IsNotExist(err) {
		t.Error("a.txt should not exist")
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

func TestDeleteFile_DeletesAFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("target.txt", []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := oslayer.DeleteFile("target.txt"); err != nil {
		t.Fatalf("DeleteFile: %v", err)
	}
	if _, err := os.Stat("target.txt"); !os.IsNotExist(err) {
		t.Error("target.txt should not exist")
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
	if err := os.WriteFile("file.txt", []byte("data\n"), 0644); err != nil {
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

	acquired := make(chan struct{})
	go func() {
		f2, err := oslayer.OpenFile("file.txt", "overwrite", 5000)
		if err != nil {
			t.Errorf("OpenFile f2: %v", err)
			return
		}
		close(acquired)
		f2.Close()
	}()

	f1.Close()
	<-acquired
}

func TestLocking_ExclusiveLockBlocksSharedLocks(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	f1, err := oslayer.OpenFile("file.txt", "overwrite", 5000)
	if err != nil {
		t.Fatalf("OpenFile f1: %v", err)
	}

	acquired := make(chan struct{})
	go func() {
		f2, err := oslayer.OpenFile("file.txt", "read", 5000)
		if err != nil {
			t.Errorf("OpenFile f2: %v", err)
			return
		}
		close(acquired)
		f2.Close()
	}()

	f1.Close()
	<-acquired
}

func TestLocking_AppendModeAcquiresExclusiveLock(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	f1, err := oslayer.OpenFile("file.txt", "append", 5000)
	if err != nil {
		t.Fatalf("OpenFile f1: %v", err)
	}

	acquired := make(chan struct{})
	go func() {
		f2, err := oslayer.OpenFile("file.txt", "read", 5000)
		if err != nil {
			t.Errorf("OpenFile f2: %v", err)
			return
		}
		close(acquired)
		f2.Close()
	}()

	f1.Close()
	<-acquired
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
