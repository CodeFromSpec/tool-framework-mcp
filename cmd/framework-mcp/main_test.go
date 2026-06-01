// code-from-spec: ROOT/golang/tests/server@BR93iyQ1lhNmXUC-2JYXIQaFRv0
package main_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "framework-mcp-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	binName := "framework-mcp"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binaryPath = filepath.Join(tmpDir, binName)

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: build binary: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func testRunBinary(args ...string) (stdout, stderr string, exitCode int) {
	cmd := exec.Command(binaryPath, args...)
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

func TestHelpFlag(t *testing.T) {
	stdout, _, exitCode := testRunBinary("--help")
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("expected stdout to contain usage message, got: %q", stdout)
	}
}

func TestHelpWord(t *testing.T) {
	stdout, _, exitCode := testRunBinary("help")
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("expected stdout to contain usage message, got: %q", stdout)
	}
}

func TestShortHelpFlag(t *testing.T) {
	stdout, _, exitCode := testRunBinary("-h")
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("expected stdout to contain usage message, got: %q", stdout)
	}
}

func TestUnrecognizedArgument(t *testing.T) {
	_, stderr, exitCode := testRunBinary("something")
	if exitCode != 1 {
		t.Fatalf("expected exit 1, got %d", exitCode)
	}
	if !strings.Contains(stderr, "Usage:") {
		t.Errorf("expected stderr to contain usage message, got: %q", stderr)
	}
}

func TestMultipleArguments(t *testing.T) {
	_, stderr, exitCode := testRunBinary("foo", "bar")
	if exitCode != 1 {
		t.Fatalf("expected exit 1, got %d", exitCode)
	}
	if !strings.Contains(stderr, "Usage:") {
		t.Errorf("expected stderr to contain usage message, got: %q", stderr)
	}
}

func testStartMCPServer(t *testing.T) (*exec.Cmd, io.WriteCloser, *bufio.Scanner) {
	t.Helper()
	cmd := exec.Command(binaryPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("stdin pipe: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	t.Cleanup(func() {
		stdin.Close()
		cmd.Wait()
	})
	scanner := bufio.NewScanner(stdout)
	return cmd, stdin, scanner
}

func testSendRequest(t *testing.T, w io.Writer, id int, method string, params any) {
	t.Helper()
	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
	}
	if params != nil {
		req["params"] = params
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	if err != nil {
		t.Fatalf("write request: %v", err)
	}
}

func testReadResponse(t *testing.T, scanner *bufio.Scanner, id int) map[string]any {
	t.Helper()
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		var resp map[string]any
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		respID, ok := resp["id"]
		if !ok {
			continue
		}
		var rid float64
		switch v := respID.(type) {
		case float64:
			rid = v
		case int:
			rid = float64(v)
		default:
			continue
		}
		if int(rid) == id {
			return resp
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner error: %v", err)
	}
	t.Fatalf("no response found for id %d", id)
	return nil
}

func testInitializeServer(t *testing.T, w io.Writer, scanner *bufio.Scanner) {
	t.Helper()
	testSendRequest(t, w, 1, "initialize", map[string]any{
		"protocolVersion": "2024-11-05",
		"clientInfo": map[string]any{
			"name":    "test-client",
			"version": "1.0",
		},
		"capabilities": map[string]any{},
	})
	testReadResponse(t, scanner, 1)
}

func testListTools(t *testing.T, w io.Writer, scanner *bufio.Scanner) map[string]any {
	t.Helper()
	testSendRequest(t, w, 2, "tools/list", nil)
	return testReadResponse(t, scanner, 2)
}

func testFindTool(resp map[string]any, name string) map[string]any {
	result, ok := resp["result"].(map[string]any)
	if !ok {
		return nil
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		return nil
	}
	for _, item := range tools {
		tool, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if tool["name"] == name {
			return tool
		}
	}
	return nil
}

func TestToolsListMaxResultSizeCharsForLoadChain(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = ctx

	_, stdin, scanner := testStartMCPServer(t)
	testInitializeServer(t, stdin, scanner)
	resp := testListTools(t, stdin, scanner)

	tool := testFindTool(resp, "load_chain")
	if tool == nil {
		t.Fatal("load_chain tool not found in tools/list response")
	}

	meta, ok := tool["_meta"].(map[string]any)
	if !ok {
		t.Fatal("load_chain tool has no _meta field")
	}

	val, ok := meta["anthropic/maxResultSizeChars"]
	if !ok {
		t.Fatal("load_chain _meta missing anthropic/maxResultSizeChars")
	}

	numVal, ok := val.(float64)
	if !ok {
		t.Fatalf("anthropic/maxResultSizeChars is not a number, got %T", val)
	}

	if int(numVal) != 500000 {
		t.Errorf("expected anthropic/maxResultSizeChars=500000, got %v", numVal)
	}
}

func TestToolsListAdvertisesAllTools(t *testing.T) {
	_, stdin, scanner := testStartMCPServer(t)
	testInitializeServer(t, stdin, scanner)
	resp := testListTools(t, stdin, scanner)

	expectedTools := []string{"load_chain", "write_file", "validate_specs"}
	for _, name := range expectedTools {
		if tool := testFindTool(resp, name); tool == nil {
			t.Errorf("expected tool %q not found in tools/list response", name)
		}
	}
}
