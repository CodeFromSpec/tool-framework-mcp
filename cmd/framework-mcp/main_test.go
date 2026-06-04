// code-from-spec: ROOT/golang/tests/server@da-jvw5gKNVPJuYdc3JQKWnrFps
package main_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var testBinaryPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "framework-mcp-test-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, "TestMain: failed to create temp dir:", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	binaryName := "framework-mcp"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	testBinaryPath = filepath.Join(dir, binaryName)

	cmd := exec.Command("go", "build", "-o", testBinaryPath, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "TestMain: failed to build binary:", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestHelpFlagPrintsUsageToStdout(t *testing.T) {
	cmd := exec.Command(testBinaryPath, "--help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0, got: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("expected usage message in stdout, got: %s", string(out))
	}
}

func TestHelpWordPrintsUsageToStdout(t *testing.T) {
	cmd := exec.Command(testBinaryPath, "help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0, got: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("expected usage message in stdout, got: %s", string(out))
	}
}

func TestShortHelpFlagPrintsUsageToStdout(t *testing.T) {
	cmd := exec.Command(testBinaryPath, "-h")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0, got: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("expected usage message in stdout, got: %s", string(out))
	}
}

func TestUnrecognizedArgumentPrintsUsageToStderr(t *testing.T) {
	cmd := exec.Command(testBinaryPath, "something")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected exit 1, got exit 0")
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Fatalf("expected exit code 1, got %d", exitErr.ExitCode())
		}
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Errorf("expected usage message in stderr, got: %s", stderr.String())
	}
}

func TestMultipleArgumentsPrintsUsageToStderr(t *testing.T) {
	cmd := exec.Command(testBinaryPath, "foo", "bar")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected exit 1, got exit 0")
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Fatalf("expected exit code 1, got %d", exitErr.ExitCode())
		}
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Errorf("expected usage message in stderr, got: %s", stderr.String())
	}
}

type testMCPProcess struct {
	cmd    *exec.Cmd
	stdin  *strings.Reader
	stdout *bufio.Reader
	writer *os.File
	reader *os.File
}

func testStartMCPProcess(t *testing.T) (*exec.Cmd, *bufio.Writer, *bufio.Reader) {
	t.Helper()
	cmd := exec.Command(testBinaryPath)

	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatalf("testStartMCPProcess: pipe: %v", err)
	}
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("testStartMCPProcess: pipe: %v", err)
	}

	cmd.Stdin = stdinR
	cmd.Stdout = stdoutW
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("testStartMCPProcess: start: %v", err)
	}

	stdinR.Close()
	stdoutW.Close()

	t.Cleanup(func() {
		stdinW.Close()
		cmd.Wait()
		stdoutR.Close()
	})

	return cmd, bufio.NewWriter(stdinW), bufio.NewReader(stdoutR)
}

func testSendJSON(t *testing.T, w *bufio.Writer, v any) {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("testSendJSON: marshal: %v", err)
	}
	if _, err := w.Write(data); err != nil {
		t.Fatalf("testSendJSON: write: %v", err)
	}
	if err := w.WriteByte('\n'); err != nil {
		t.Fatalf("testSendJSON: write newline: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("testSendJSON: flush: %v", err)
	}
}

func testReadJSON(t *testing.T, r *bufio.Reader) map[string]any {
	t.Helper()
	line, err := r.ReadString('\n')
	if err != nil {
		t.Fatalf("testReadJSON: read: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(line), &result); err != nil {
		t.Fatalf("testReadJSON: unmarshal: %v", err)
	}
	return result
}

func testMCPHandshake(t *testing.T, w *bufio.Writer, r *bufio.Reader) {
	t.Helper()

	testSendJSON(t, w, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2024-11-05",
			"clientInfo": map[string]any{
				"name":    "test-client",
				"version": "1.0.0",
			},
			"capabilities": map[string]any{},
		},
	})

	testReadJSON(t, r)

	testSendJSON(t, w, map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	})
}

func TestToolsListAdvertisesMaxResultSizeCharsForLoadChain(t *testing.T) {
	_, w, r := testStartMCPProcess(t)
	testMCPHandshake(t, w, r)

	testSendJSON(t, w, map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
	})

	resp := testReadJSON(t, r)

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result object, got: %v", resp["result"])
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("expected tools array, got: %v", result["tools"])
	}

	var found bool
	for _, toolAny := range tools {
		tool, ok := toolAny.(map[string]any)
		if !ok {
			continue
		}
		if tool["name"] != "load_chain" {
			continue
		}
		found = true
		meta, ok := tool["_meta"].(map[string]any)
		if !ok {
			t.Fatalf("load_chain tool missing _meta, got: %v", tool["_meta"])
		}
		maxSize, ok := meta["anthropic/maxResultSizeChars"]
		if !ok {
			t.Fatal("load_chain _meta missing anthropic/maxResultSizeChars")
		}
		maxSizeFloat, ok := maxSize.(float64)
		if !ok {
			t.Fatalf("anthropic/maxResultSizeChars expected float64, got: %T", maxSize)
		}
		if int(maxSizeFloat) != 500000 {
			t.Errorf("expected anthropic/maxResultSizeChars = 500000, got %v", maxSizeFloat)
		}
	}
	if !found {
		t.Error("load_chain tool not found in tools/list response")
	}
}

func TestToolsListAdvertisesAllTools(t *testing.T) {
	_, w, r := testStartMCPProcess(t)
	testMCPHandshake(t, w, r)

	testSendJSON(t, w, map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
	})

	resp := testReadJSON(t, r)

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result object, got: %v", resp["result"])
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("expected tools array, got: %v", result["tools"])
	}

	expected := []string{"load_chain", "write_file", "validate_specs", "chain_hash", "version"}
	found := make(map[string]bool)
	for _, toolAny := range tools {
		tool, ok := toolAny.(map[string]any)
		if !ok {
			continue
		}
		if name, ok := tool["name"].(string); ok {
			found[name] = true
		}
	}

	for _, name := range expected {
		if !found[name] {
			t.Errorf("expected tool %q not found in tools/list response", name)
		}
	}
}
