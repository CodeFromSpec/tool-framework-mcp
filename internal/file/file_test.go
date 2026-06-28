// code-from-spec: SPEC/golang/tests/os/file@CaUhUINZlfk8YInz08jyMtdPO-c
package file_test

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/file"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
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

func TestFileOpenReadsAllLines(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "alpha\nbeta\ngamma\n"
	if err := os.WriteFile("test.txt", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "read", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer file.FileClose(handle)

	for _, want := range []string{"alpha", "beta", "gamma"} {
		got, err := file.FileReadLine(handle)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	_, err = file.FileReadLine(handle)
	if !errors.Is(err, file.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileReadLineNormalizesCRLF(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "alpha\r\nbeta\r\n"
	if err := os.WriteFile("test.txt", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "read", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer file.FileClose(handle)

	for _, want := range []string{"alpha", "beta"} {
		got, err := file.FileReadLine(handle)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFileReadLineNoTrailingNewline(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "alpha\nbeta"
	if err := os.WriteFile("test.txt", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "read", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer file.FileClose(handle)

	for _, want := range []string{"alpha", "beta"} {
		got, err := file.FileReadLine(handle)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	_, err = file.FileReadLine(handle)
	if !errors.Is(err, file.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileSkipLinesAdvancesReader(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "one\ntwo\nthree\nfour\nfive\n"
	if err := os.WriteFile("test.txt", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "read", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer file.FileClose(handle)

	if err := file.FileSkipLines(handle, 2); err != nil {
		t.Fatalf("FileSkipLines: %v", err)
	}

	got, err := file.FileReadLine(handle)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if got != "three" {
		t.Errorf("got %q, want %q", got, "three")
	}
}

func TestFileSkipLinesPastEndOfFile(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "one\ntwo\n"
	if err := os.WriteFile("test.txt", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "read", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer file.FileClose(handle)

	if err := file.FileSkipLines(handle, 10); err != nil {
		t.Fatalf("FileSkipLines: %v", err)
	}

	_, err = file.FileReadLine(handle)
	if !errors.Is(err, file.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileReadLinePreservesLeadingWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "  alpha\n    beta\n"
	if err := os.WriteFile("test.txt", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "read", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer file.FileClose(handle)

	for _, want := range []string{"  alpha", "    beta"} {
		got, err := file.FileReadLine(handle)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFileReadLinePreservesTrailingWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "alpha  \nbeta   \n"
	if err := os.WriteFile("test.txt", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "read", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer file.FileClose(handle)

	for _, want := range []string{"alpha  ", "beta   "} {
		got, err := file.FileReadLine(handle)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFileReadLinePreservesInternalWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "alpha   beta\none\ttwo\n"
	if err := os.WriteFile("test.txt", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "read", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer file.FileClose(handle)

	for _, want := range []string{"alpha   beta", "one\ttwo"} {
		got, err := file.FileReadLine(handle)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFileReadLinePreservesEmptyLines(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "alpha\n\n\nbeta\n"
	if err := os.WriteFile("test.txt", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "read", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer file.FileClose(handle)

	for _, want := range []string{"alpha", "", "", "beta"} {
		got, err := file.FileReadLine(handle)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFileReadLinePreservesNonASCII(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "café\n日本語\n🎉🚀\n"
	if err := os.WriteFile("test.txt", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "read", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer file.FileClose(handle)

	for _, want := range []string{"café", "日本語", "🎉🚀"} {
		got, err := file.FileReadLine(handle)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFileOpenEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("test.txt", []byte{}, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "read", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer file.FileClose(handle)

	_, err = file.FileReadLine(handle)
	if !errors.Is(err, file.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileReadLineSingleLineWithoutNewline(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("test.txt", []byte("hello"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "read", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer file.FileClose(handle)

	got, err := file.FileReadLine(handle)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}

	_, err = file.FileReadLine(handle)
	if !errors.Is(err, file.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileOpenFileDoesNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	_, err := file.FileOpen(&pathutils.PathCfs{Value: "nonexistent/file.txt"}, "read", 500)
	if !errors.Is(err, file.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestFileReadLineAfterClose(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("test.txt", []byte("alpha\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "read", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	file.FileClose(handle)

	_, err = file.FileReadLine(handle)
	if !errors.Is(err, file.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileSkipLinesAfterClose(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("test.txt", []byte("alpha\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "read", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	file.FileClose(handle)

	if err := file.FileSkipLines(handle, 1); err != nil {
		t.Errorf("expected no error after close, got %v", err)
	}
}

func TestFileWriteNewFile(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "newfile.txt"}, "overwrite", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	if err := file.FileWrite(handle, "hello world"); err != nil {
		t.Fatalf("FileWrite: %v", err)
	}
	file.FileClose(handle)

	got, err := os.ReadFile("newfile.txt")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "hello world" {
		t.Errorf("got %q, want %q", string(got), "hello world")
	}
}

func TestFileWriteOverwritesExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("test.txt", []byte("old"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "overwrite", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	if err := file.FileWrite(handle, "new"); err != nil {
		t.Fatalf("FileWrite: %v", err)
	}
	file.FileClose(handle)

	got, err := os.ReadFile("test.txt")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "new" {
		t.Errorf("got %q, want %q", string(got), "new")
	}
}

func TestFileOpenCreatesIntermediateDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "a/b/c/file.txt"}, "overwrite", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	if err := file.FileWrite(handle, "data"); err != nil {
		t.Fatalf("FileWrite: %v", err)
	}
	file.FileClose(handle)

	got, err := os.ReadFile(filepath.Join("a", "b", "c", "file.txt"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "data" {
		t.Errorf("got %q, want %q", string(got), "data")
	}
}

func TestFileWritePreservesUTF8(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "overwrite", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	want := "café 日本語 🎉"
	if err := file.FileWrite(handle, want); err != nil {
		t.Fatalf("FileWrite: %v", err)
	}
	file.FileClose(handle)

	got, err := os.ReadFile("test.txt")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != want {
		t.Errorf("got %q, want %q", string(got), want)
	}
}

func TestFileWritePreservesLineEndings(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "overwrite", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	want := "alpha\r\nbeta\r\n"
	if err := file.FileWrite(handle, want); err != nil {
		t.Fatalf("FileWrite: %v", err)
	}
	file.FileClose(handle)

	got, err := os.ReadFile("test.txt")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != want {
		t.Errorf("got %q, want %q", string(got), want)
	}
}

func TestFileWriteEmptyContent(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "overwrite", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	if err := file.FileWrite(handle, ""); err != nil {
		t.Fatalf("FileWrite: %v", err)
	}
	file.FileClose(handle)

	info, err := os.Stat("test.txt")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size() != 0 {
		t.Errorf("expected zero bytes, got %d", info.Size())
	}
}

func TestFileOpenPropagatesValidationErrors(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	_, err := file.FileOpen(&pathutils.PathCfs{Value: "../../outside"}, "overwrite", 500)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFileOpenCannotCreateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("blockingfile", []byte("content"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := file.FileOpen(&pathutils.PathCfs{Value: "blockingfile/subdir/file.txt"}, "overwrite", 500)
	if !errors.Is(err, file.ErrCannotCreateDirectory) {
		t.Errorf("expected ErrCannotCreateDirectory, got %v", err)
	}
}

func TestFileOpenCannotOpenFilePathIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.Mkdir("mydir", 0755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	_, err := file.FileOpen(&pathutils.PathCfs{Value: "mydir"}, "overwrite", 500)
	if !errors.Is(err, file.ErrCannotOpenFile) {
		t.Errorf("expected ErrCannotOpenFile, got %v", err)
	}
}

func TestFileAppendDoesNotTruncate(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("test.txt", []byte("old"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "append", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	file.FileClose(handle)

	got, err := os.ReadFile("test.txt")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "old" {
		t.Errorf("got %q, want %q", string(got), "old")
	}
}

func TestFileAppendCreatesFileIfNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "newfile.txt"}, "append", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	file.FileClose(handle)

	if _, err := os.Stat("newfile.txt"); err != nil {
		t.Errorf("expected file to exist, got: %v", err)
	}
}

func TestFileReadLineFailsInOverwriteMode(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "overwrite", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer file.FileClose(handle)

	_, err = file.FileReadLine(handle)
	if !errors.Is(err, file.ErrWrongMode) {
		t.Errorf("expected ErrWrongMode, got %v", err)
	}
}

func TestFileReadLineFailsInAppendMode(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "append", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer file.FileClose(handle)

	_, err = file.FileReadLine(handle)
	if !errors.Is(err, file.ErrWrongMode) {
		t.Errorf("expected ErrWrongMode, got %v", err)
	}
}

func TestFileWriteFailsInReadMode(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("test.txt", []byte("content"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "read", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer file.FileClose(handle)

	if err := file.FileWrite(handle, "anything"); !errors.Is(err, file.ErrWrongMode) {
		t.Errorf("expected ErrWrongMode, got %v", err)
	}
}

func TestFileSkipLinesFailsInOverwriteMode(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	handle, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "overwrite", 500)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer file.FileClose(handle)

	if err := file.FileSkipLines(handle, 1); !errors.Is(err, file.ErrWrongMode) {
		t.Errorf("expected ErrWrongMode, got %v", err)
	}
}

func TestFileOpenRejectsUnknownMode(t *testing.T) {
	_, err := file.FileOpen(&pathutils.PathCfs{Value: "any.txt"}, "invalid", 500)
	if !errors.Is(err, file.ErrInvalidMode) {
		t.Errorf("expected ErrInvalidMode, got %v", err)
	}
}

func TestFileRenameMovesFile(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("a.txt", []byte("data"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := file.FileRename(&pathutils.PathCfs{Value: "a.txt"}, &pathutils.PathCfs{Value: "b.txt"}); err != nil {
		t.Fatalf("FileRename: %v", err)
	}

	got, err := os.ReadFile("b.txt")
	if err != nil {
		t.Fatalf("ReadFile b.txt: %v", err)
	}
	if string(got) != "data" {
		t.Errorf("got %q, want %q", string(got), "data")
	}

	if _, err := os.Stat("a.txt"); !os.IsNotExist(err) {
		t.Errorf("expected a.txt to not exist, got: %v", err)
	}
}

func TestFileRenameOverwritesDestination(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("dest.txt", []byte("old"), 0644); err != nil {
		t.Fatalf("WriteFile dest: %v", err)
	}
	if err := os.WriteFile("src.txt", []byte("new"), 0644); err != nil {
		t.Fatalf("WriteFile src: %v", err)
	}

	if err := file.FileRename(&pathutils.PathCfs{Value: "src.txt"}, &pathutils.PathCfs{Value: "dest.txt"}); err != nil {
		t.Fatalf("FileRename: %v", err)
	}

	got, err := os.ReadFile("dest.txt")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "new" {
		t.Errorf("got %q, want %q", string(got), "new")
	}
}

func TestFileRenameNonExistentSource(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	err := file.FileRename(&pathutils.PathCfs{Value: "nonexistent.txt"}, &pathutils.PathCfs{Value: "dest.txt"})
	if !errors.Is(err, file.ErrCannotRename) {
		t.Errorf("expected ErrCannotRename, got %v", err)
	}
}

func TestFileDeleteRemovesFile(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("target.txt", []byte("content"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := file.FileDelete(&pathutils.PathCfs{Value: "target.txt"}); err != nil {
		t.Fatalf("FileDelete: %v", err)
	}

	if _, err := os.Stat("target.txt"); !os.IsNotExist(err) {
		t.Errorf("expected target.txt to not exist, got: %v", err)
	}
}

func TestFileDeleteNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	err := file.FileDelete(&pathutils.PathCfs{Value: "nonexistent.txt"})
	if !errors.Is(err, file.ErrCannotDelete) {
		t.Errorf("expected ErrCannotDelete, got %v", err)
	}
}

func TestFileSharedLockAllowsConcurrentReaders(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("test.txt", []byte("data\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	handle1, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "read", 500)
	if err != nil {
		t.Fatalf("FileOpen handle1: %v", err)
	}

	handle2, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "read", 500)
	if err != nil {
		t.Fatalf("FileOpen handle2 (shared lock should not block): %v", err)
	}

	file.FileClose(handle1)
	file.FileClose(handle2)
}

func TestFileExclusiveLockBlocksOtherExclusiveLocks(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("test.txt", []byte("data\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	handle1, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "overwrite", 500)
	if err != nil {
		t.Fatalf("FileOpen handle1: %v", err)
	}

	var handle2 *file.FileHandle
	var handle2Err error
	var wg sync.WaitGroup
	opened := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		handle2, handle2Err = file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "overwrite", 5000)
		close(opened)
	}()

	select {
	case <-opened:
		t.Error("second exclusive open returned before first was closed")
		file.FileClose(handle1)
	case <-time.After(100 * time.Millisecond):
		file.FileClose(handle1)
		wg.Wait()
	}

	if handle2Err != nil {
		t.Fatalf("FileOpen handle2: %v", handle2Err)
	}
	file.FileClose(handle2)
}

func TestFileExclusiveLockBlocksSharedLocks(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("test.txt", []byte("data\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	handle1, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "overwrite", 500)
	if err != nil {
		t.Fatalf("FileOpen handle1: %v", err)
	}

	var handle2 *file.FileHandle
	var handle2Err error
	var wg sync.WaitGroup
	opened := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		handle2, handle2Err = file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "read", 5000)
		close(opened)
	}()

	select {
	case <-opened:
		t.Error("read open returned before exclusive lock was released")
		file.FileClose(handle1)
	case <-time.After(100 * time.Millisecond):
		file.FileClose(handle1)
		wg.Wait()
	}

	if handle2Err != nil {
		t.Fatalf("FileOpen handle2: %v", handle2Err)
	}
	file.FileClose(handle2)
}

func TestFileAppendModeAcquiresExclusiveLock(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("test.txt", []byte("data\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	handle1, err := file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "append", 500)
	if err != nil {
		t.Fatalf("FileOpen handle1: %v", err)
	}

	var handle2 *file.FileHandle
	var handle2Err error
	var wg sync.WaitGroup
	opened := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		handle2, handle2Err = file.FileOpen(&pathutils.PathCfs{Value: "test.txt"}, "read", 5000)
		close(opened)
	}()

	select {
	case <-opened:
		t.Error("read open returned before append exclusive lock was released")
		file.FileClose(handle1)
	case <-time.After(100 * time.Millisecond):
		file.FileClose(handle1)
		wg.Wait()
	}

	if handle2Err != nil {
		t.Fatalf("FileOpen handle2: %v", handle2Err)
	}
	file.FileClose(handle2)
}
