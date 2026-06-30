// code-from-spec: SPEC/golang/test/cases/manifest@Raupyxs5pgJuDmgyuO0zHpvvAw8
package manifest_test

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/manifest"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"
)

const manifestPath = "code-from-spec/.manifest"
const manifestDir = "code-from-spec"

func writeManifestFile(t *testing.T, content string) {
	t.Helper()
	if err := os.MkdirAll(manifestDir, 0755); err != nil {
		t.Fatalf("failed to create manifest dir: %v", err)
	}
	if err := os.WriteFile(manifestPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write manifest file: %v", err)
	}
}

func readManifestFile(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to read manifest file: %v", err)
	}
	return string(data)
}

func manifestEntry(logicalName, path, checksum, chainHash string) string {
	return fmt.Sprintf("%s;path:%s;checksum:%s;chain:%s\n", logicalName, path, checksum, chainHash)
}

func TestOpenManifest_ReadOnly_ExistingManifest(t *testing.T) {
	testutils.Chdir(t)

	entry1 := manifestEntry("ARTIFACT/alpha", "internal/alpha/alpha.go", "checksum1111111111111111111", "chainhash1111111111111111111")
	entry2 := manifestEntry("ARTIFACT/beta", "internal/beta/beta.go", "checksum2222222222222222222", "chainhash2222222222222222222")
	writeManifestFile(t, "code-from-spec: v5\n"+entry1+entry2)

	m, err := manifest.OpenManifest(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Version != "v5" {
		t.Errorf("expected version v5, got %q", m.Version)
	}
	if len(m.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(m.Entries))
	}

	e1, ok := m.Entries["ARTIFACT/alpha"]
	if !ok {
		t.Fatal("missing entry ARTIFACT/alpha")
	}
	if e1.Path != "internal/alpha/alpha.go" {
		t.Errorf("wrong path: %q", e1.Path)
	}
	if e1.Checksum != "checksum1111111111111111111" {
		t.Errorf("wrong checksum: %q", e1.Checksum)
	}
	if e1.ChainHash != "chainhash1111111111111111111" {
		t.Errorf("wrong chain hash: %q", e1.ChainHash)
	}

	e2, ok := m.Entries["ARTIFACT/beta"]
	if !ok {
		t.Fatal("missing entry ARTIFACT/beta")
	}
	if e2.Path != "internal/beta/beta.go" {
		t.Errorf("wrong path: %q", e2.Path)
	}
	if e2.Checksum != "checksum2222222222222222222" {
		t.Errorf("wrong checksum: %q", e2.Checksum)
	}
	if e2.ChainHash != "chainhash2222222222222222222" {
		t.Errorf("wrong chain hash: %q", e2.ChainHash)
	}
}

func TestOpenManifest_ReadOnly_HeaderOnly(t *testing.T) {
	testutils.Chdir(t)

	writeManifestFile(t, "code-from-spec: v5\n")

	m, err := manifest.OpenManifest(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(m.Entries) != 0 {
		t.Errorf("expected empty entries map, got %d entries", len(m.Entries))
	}
}

func TestOpenManifest_ReadOnly_MissingManifest(t *testing.T) {
	testutils.Chdir(t)

	m, err := manifest.OpenManifest(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(m.Entries) != 0 {
		t.Errorf("expected empty entries map, got %d entries", len(m.Entries))
	}

	if _, statErr := os.Stat(manifestPath); !os.IsNotExist(statErr) {
		t.Error("manifest file should not have been created")
	}
}

func TestOpenManifest_Writable_LoadsExistingEntries(t *testing.T) {
	testutils.Chdir(t)

	entry1 := manifestEntry("ARTIFACT/alpha", "internal/alpha/alpha.go", "checksum1111111111111111111", "chainhash1111111111111111111")
	writeManifestFile(t, "code-from-spec: v5\n"+entry1)

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(m.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(m.Entries))
	}
	if _, ok := m.Entries["ARTIFACT/alpha"]; !ok {
		t.Error("missing entry ARTIFACT/alpha")
	}

	if err := m.Discard(); err != nil {
		t.Errorf("Discard failed: %v", err)
	}
}

func TestOpenManifest_Writable_MissingManifest(t *testing.T) {
	testutils.Chdir(t)

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(m.Entries) != 0 {
		t.Errorf("expected empty entries map, got %d entries", len(m.Entries))
	}

	if err := m.Discard(); err != nil {
		t.Errorf("Discard failed: %v", err)
	}
}

func TestSave_CreatesManifestFromScratch(t *testing.T) {
	testutils.Chdir(t)

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m.Entries["ARTIFACT/beta"] = manifest.ManifestEntry{
		Path:      "internal/beta/beta.go",
		Checksum:  "checksum2222222222222222222",
		ChainHash: "chainhash2222222222222222222",
	}
	m.Entries["ARTIFACT/alpha"] = manifest.ManifestEntry{
		Path:      "internal/alpha/alpha.go",
		Checksum:  "checksum1111111111111111111",
		ChainHash: "chainhash1111111111111111111",
	}

	if err := m.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	content := readManifestFile(t)
	expected := "code-from-spec: v5\n" +
		"ARTIFACT/alpha;path:internal/alpha/alpha.go;checksum:checksum1111111111111111111;chain:chainhash1111111111111111111\n" +
		"ARTIFACT/beta;path:internal/beta/beta.go;checksum:checksum2222222222222222222;chain:chainhash2222222222222222222\n"
	if content != expected {
		t.Errorf("file content mismatch.\ngot:  %q\nwant: %q", content, expected)
	}
}

func TestSave_OverwritesExistingManifest(t *testing.T) {
	testutils.Chdir(t)

	entry1 := manifestEntry("ARTIFACT/alpha", "internal/alpha/alpha.go", "checksum1111111111111111111", "chainhash1111111111111111111")
	writeManifestFile(t, "code-from-spec: v5\n"+entry1)

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m.Entries["ARTIFACT/beta"] = manifest.ManifestEntry{
		Path:      "internal/beta/beta.go",
		Checksum:  "checksum2222222222222222222",
		ChainHash: "chainhash2222222222222222222",
	}

	if err := m.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	content := readManifestFile(t)
	expected := "code-from-spec: v5\n" +
		"ARTIFACT/alpha;path:internal/alpha/alpha.go;checksum:checksum1111111111111111111;chain:chainhash1111111111111111111\n" +
		"ARTIFACT/beta;path:internal/beta/beta.go;checksum:checksum2222222222222222222;chain:chainhash2222222222222222222\n"
	if content != expected {
		t.Errorf("file content mismatch.\ngot:  %q\nwant: %q", content, expected)
	}
}

func TestSave_ModifiedEntry(t *testing.T) {
	testutils.Chdir(t)

	entry1 := manifestEntry("ARTIFACT/alpha", "internal/alpha/alpha.go", "old-checksum1111111111111111", "chainhash1111111111111111111")
	writeManifestFile(t, "code-from-spec: v5\n"+entry1)

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	e := m.Entries["ARTIFACT/alpha"]
	e.Checksum = "new-checksum1111111111111111"
	m.Entries["ARTIFACT/alpha"] = e

	if err := m.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	content := readManifestFile(t)
	expected := "code-from-spec: v5\n" +
		"ARTIFACT/alpha;path:internal/alpha/alpha.go;checksum:new-checksum1111111111111111;chain:chainhash1111111111111111111\n"
	if content != expected {
		t.Errorf("file content mismatch.\ngot:  %q\nwant: %q", content, expected)
	}
}

func TestSave_RemovedEntry(t *testing.T) {
	testutils.Chdir(t)

	entry1 := manifestEntry("ARTIFACT/alpha", "internal/alpha/alpha.go", "checksum1111111111111111111", "chainhash1111111111111111111")
	entry2 := manifestEntry("ARTIFACT/beta", "internal/beta/beta.go", "checksum2222222222222222222", "chainhash2222222222222222222")
	writeManifestFile(t, "code-from-spec: v5\n"+entry1+entry2)

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	delete(m.Entries, "ARTIFACT/beta")

	if err := m.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	content := readManifestFile(t)
	expected := "code-from-spec: v5\n" +
		"ARTIFACT/alpha;path:internal/alpha/alpha.go;checksum:checksum1111111111111111111;chain:chainhash1111111111111111111\n"
	if content != expected {
		t.Errorf("file content mismatch.\ngot:  %q\nwant: %q", content, expected)
	}
}

func TestSave_EmptyEntries(t *testing.T) {
	testutils.Chdir(t)

	entry1 := manifestEntry("ARTIFACT/alpha", "internal/alpha/alpha.go", "checksum1111111111111111111", "chainhash1111111111111111111")
	entry2 := manifestEntry("ARTIFACT/beta", "internal/beta/beta.go", "checksum2222222222222222222", "chainhash2222222222222222222")
	writeManifestFile(t, "code-from-spec: v5\n"+entry1+entry2)

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m.Entries = map[string]manifest.ManifestEntry{}

	if err := m.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	content := readManifestFile(t)
	expected := "code-from-spec: v5\n"
	if content != expected {
		t.Errorf("file content mismatch.\ngot:  %q\nwant: %q", content, expected)
	}
}

func TestDiscard_DoesNotModifyFile(t *testing.T) {
	testutils.Chdir(t)

	entry1 := manifestEntry("ARTIFACT/alpha", "internal/alpha/alpha.go", "checksum1111111111111111111", "chainhash1111111111111111111")
	writeManifestFile(t, "code-from-spec: v5\n"+entry1)

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m.Entries["ARTIFACT/beta"] = manifest.ManifestEntry{
		Path:      "internal/beta/beta.go",
		Checksum:  "checksum2222222222222222222",
		ChainHash: "chainhash2222222222222222222",
	}

	if err := m.Discard(); err != nil {
		t.Fatalf("Discard failed: %v", err)
	}

	content := readManifestFile(t)
	expected := "code-from-spec: v5\n" + entry1
	if content != expected {
		t.Errorf("file content mismatch.\ngot:  %q\nwant: %q", content, expected)
	}
}

func TestOpenManifest_InvalidHeader(t *testing.T) {
	testutils.Chdir(t)

	writeManifestFile(t, "invalid-header\n")

	_, err := manifest.OpenManifest(true)
	if !errors.Is(err, manifest.ErrManifestFormatError) {
		t.Errorf("expected ErrManifestFormatError, got %v", err)
	}
}

func TestSave_OnReadOnly(t *testing.T) {
	testutils.Chdir(t)

	m, err := manifest.OpenManifest(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = m.Save()
	if !errors.Is(err, manifest.ErrReadOnly) {
		t.Errorf("expected ErrReadOnly, got %v", err)
	}
}

func TestDiscard_OnReadOnly(t *testing.T) {
	testutils.Chdir(t)

	m, err := manifest.OpenManifest(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = m.Discard()
	if !errors.Is(err, manifest.ErrReadOnly) {
		t.Errorf("expected ErrReadOnly, got %v", err)
	}
}

func TestDiscard_AfterSave(t *testing.T) {
	testutils.Chdir(t)

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := m.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	err = m.Discard()
	if !errors.Is(err, manifest.ErrManifestClosed) {
		t.Errorf("expected ErrManifestClosed, got %v", err)
	}
}

func TestSave_AfterDiscard(t *testing.T) {
	testutils.Chdir(t)

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := m.Discard(); err != nil {
		t.Fatalf("Discard failed: %v", err)
	}

	err = m.Save()
	if !errors.Is(err, manifest.ErrManifestClosed) {
		t.Errorf("expected ErrManifestClosed, got %v", err)
	}
}

func TestSave_AfterSave(t *testing.T) {
	testutils.Chdir(t)

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := m.Save(); err != nil {
		t.Fatalf("first Save failed: %v", err)
	}

	err = m.Save()
	if !errors.Is(err, manifest.ErrManifestClosed) {
		t.Errorf("expected ErrManifestClosed, got %v", err)
	}
}

func TestDiscard_AfterDiscard(t *testing.T) {
	testutils.Chdir(t)

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := m.Discard(); err != nil {
		t.Fatalf("first Discard failed: %v", err)
	}

	err = m.Discard()
	if !errors.Is(err, manifest.ErrManifestClosed) {
		t.Errorf("expected ErrManifestClosed, got %v", err)
	}
}

func TestConcurrency_ReadersDoNotBlock(t *testing.T) {
	testutils.Chdir(t)

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := manifest.OpenManifest(true)
			if err != nil {
				errCh <- err
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("concurrent readers timed out")
	}

	close(errCh)
	for err := range errCh {
		t.Errorf("reader error: %v", err)
	}
}

func TestConcurrency_WriterBlocksReader(t *testing.T) {
	testutils.Chdir(t)

	writer, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("writer OpenManifest failed: %v", err)
	}

	readerStarted := make(chan struct{})
	readerDone := make(chan error, 1)

	go func() {
		close(readerStarted)
		_, err := manifest.OpenManifest(true)
		readerDone <- err
	}()

	<-readerStarted
	time.Sleep(50 * time.Millisecond)

	if err := writer.Discard(); err != nil {
		t.Fatalf("writer Discard failed: %v", err)
	}

	select {
	case err := <-readerDone:
		if err != nil {
			t.Errorf("reader error after lock release: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("reader timed out after writer released lock")
	}
}

func TestConcurrency_WriterBlocksWriter(t *testing.T) {
	testutils.Chdir(t)

	first, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("first writer OpenManifest failed: %v", err)
	}

	secondStarted := make(chan struct{})
	secondDone := make(chan error, 1)

	go func() {
		close(secondStarted)
		_, err := manifest.OpenManifest(false)
		secondDone <- err
	}()

	<-secondStarted
	time.Sleep(50 * time.Millisecond)

	if err := first.Discard(); err != nil {
		t.Fatalf("first writer Discard failed: %v", err)
	}

	select {
	case err := <-secondDone:
		if err != nil {
			t.Errorf("second writer error after lock release: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("second writer timed out after first writer released lock")
	}
}
