// code-from-spec: ROOT/golang/tests/server@y5-Vp3ZTyUaPZKWnE8aHdY_5k-s
package main_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
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
		fmt.Fprintf(os.Stderr, "TestMain: failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	binName := "framework-mcp"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binaryPath = filepath.Join(tmpDir, binName)

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: failed to build binary: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// runBinary executes the binary with the given arguments and returns
// the combined stdout, stderr, and exit code.
func runBinary(args ...string) (stdout string, stderr string, exitCode int) {
	cmd := exec.Command(binaryPath, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return stdout, stderr, exitCode
}

// TestHelpFlag verifies that --help prints usage to stdout and exits 0.
func TestHelpFlag(t *testing.T) {
	stdout, _, exitCode := runBinary("--help")
	if exitCode != 0 {
		t.Errorf("expected exit 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("expected stdout to contain usage message, got: %q", stdout)
	}
}

// TestHelpWord verifies that the word "help" prints usage to stdout and exits 0.
func TestHelpWord(t *testing.T) {
	stdout, _, exitCode := runBinary("help")
	if exitCode != 0 {
		t.Errorf("expected exit 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("expected stdout to contain usage message, got: %q", stdout)
	}
}

// TestShortHelpFlag verifies that -h prints usage to stdout and exits 0.
func TestShortHelpFlag(t *testing.T) {
	stdout, _, exitCode := runBinary("-h")
	if exitCode != 0 {
		t.Errorf("expected exit 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("expected stdout to contain usage message, got: %q", stdout)
	}
}

// TestUnrecognizedArgument verifies that an unrecognized argument prints usage
// to stderr and exits 1.
func TestUnrecognizedArgument(t *testing.T) {
	_, stderr, exitCode := runBinary("something")
	if exitCode != 1 {
		t.Errorf("expected exit 1, got %d", exitCode)
	}
	if !strings.Contains(stderr, "Usage:") {
		t.Errorf("expected stderr to contain usage message, got: %q", stderr)
	}
}

// TestMultipleArguments verifies that multiple arguments print usage to stderr
// and exit 1.
func TestMultipleArguments(t *testing.T) {
	_, stderr, exitCode := runBinary("foo", "bar")
	if exitCode != 1 {
		t.Errorf("expected exit 1, got %d", exitCode)
	}
	if !strings.Contains(stderr, "Usage:") {
		t.Errorf("expected stderr to contain usage message, got: %q", stderr)
	}
}

// sendMCPRequests starts the binary as a subprocess, sends the given JSON-RPC
// lines over stdin, and reads JSON-RPC response lines from stdout. It returns
// all response lines read before the process is killed.
func sendMCPRequests(t *testing.T, requests []string) []string {
	t.Helper()
	cmd := exec.Command(binaryPath)
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("sendMCPRequests: StdinPipe: %v", err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("sendMCPRequests: StdoutPipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("sendMCPRequests: Start: %v", err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})

	for _, req := range requests {
		if _, err := fmt.Fprintln(stdinPipe, req); err != nil {
			t.Fatalf("sendMCPRequests: write request: %v", err)
		}
	}

	lines := make(chan string, 64)
	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		close(lines)
	}()

	// Collect lines until we have enough responses or give up.
	var result []string
	for line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		result = append(result, line)
		if len(result) >= len(requests) {
			break
		}
	}
	return result
}

// buildMCPSession returns the JSON-RPC messages needed to initialize an MCP
// session and list tools.
func buildMCPSession() []string {
	initReq := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.0.1"}}}`
	toolsReq := `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`
	return []string{initReq, toolsReq}
}

// parseToolsListResponse finds the tools/list response line among the given
// JSON-RPC response lines (the one with id=2) and returns the parsed result.
func parseToolsListResponse(t *testing.T, lines []string) map[string]any {
	t.Helper()
	for _, line := range lines {
		var msg map[string]any
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		id, ok := msg["id"]
		if !ok {
			continue
		}
		// JSON numbers unmarshal as float64.
		idFloat, ok := id.(float64)
		if !ok || idFloat != 2 {
			continue
		}
		result, ok := msg["result"].(map[string]any)
		if !ok {
			t.Fatalf("tools/list response result is not an object: %v", msg)
		}
		return result
	}
	t.Fatalf("tools/list response (id=2) not found in lines: %v", lines)
	return nil
}

// TestToolsListMaxResultSizeChars verifies that load_chain advertises
// anthropic/maxResultSizeChars = 500000 in its _meta field.
func TestToolsListMaxResultSizeChars(t *testing.T) {
	lines := sendMCPRequests(t, buildMCPSession())
	result := parseToolsListResponse(t, lines)

	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("tools field is not an array: %v", result)
	}

	for _, item := range tools {
		tool, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if tool["name"] != "load_chain" {
			continue
		}
		meta, ok := tool["_meta"].(map[string]any)
		if !ok {
			t.Fatalf("load_chain tool has no _meta field or it is not an object: %v", tool)
		}
		val, ok := meta["anthropic/maxResultSizeChars"]
		if !ok {
			t.Fatalf("load_chain _meta missing anthropic/maxResultSizeChars: %v", meta)
		}
		// JSON numbers unmarshal as float64.
		valFloat, ok := val.(float64)
		if !ok || valFloat != 500000 {
			t.Errorf("expected anthropic/maxResultSizeChars = 500000, got %v", val)
		}
		return
	}
	t.Fatal("load_chain tool not found in tools/list response")
}

// TestToolsListAllFourTools verifies that tools/list returns all four expected
// tools.
func TestToolsListAllFourTools(t *testing.T) {
	lines := sendMCPRequests(t, buildMCPSession())
	result := parseToolsListResponse(t, lines)

	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("tools field is not an array: %v", result)
	}

	want := []string{"load_chain", "write_file", "validate_specs", "hash_fragment"}
	found := make(map[string]bool)
	for _, item := range tools {
		tool, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name, _ := tool["name"].(string)
		found[name] = true
	}

	for _, name := range want {
		if !found[name] {
			t.Errorf("tool %q not found in tools/list response; found: %v", name, found)
		}
	}
}
