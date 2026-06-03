// code-from-spec: ROOT/golang/tests/server@VFEb0ev5LZzbIlGkTYaWqOSLBoo
package main_test

import (
	"bufio"
	"context"
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
	tmpDir, err := os.MkdirTemp("", "framework-mcp-test-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create temp dir:", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	binaryName := "framework-mcp"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	testBinaryPath = filepath.Join(tmpDir, binaryName)

	cmd := exec.Command("go", "build", "-o", testBinaryPath, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to build binary:", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestHelpFlag(t *testing.T) {
	cmd := exec.Command(testBinaryPath, "--help")
	stdout, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0, got error: %v", err)
	}
	if !strings.Contains(string(stdout), "Usage:") {
		t.Errorf("expected usage message in stdout, got: %s", stdout)
	}
}

func TestHelpWord(t *testing.T) {
	cmd := exec.Command(testBinaryPath, "help")
	stdout, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0, got error: %v", err)
	}
	if !strings.Contains(string(stdout), "Usage:") {
		t.Errorf("expected usage message in stdout, got: %s", stdout)
	}
}

func TestShortHelpFlag(t *testing.T) {
	cmd := exec.Command(testBinaryPath, "-h")
	stdout, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0, got error: %v", err)
	}
	if !strings.Contains(string(stdout), "Usage:") {
		t.Errorf("expected usage message in stdout, got: %s", stdout)
	}
}

func TestUnrecognizedArgument(t *testing.T) {
	cmd := exec.Command(testBinaryPath, "something")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected exit 1, got exit 0")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 1 {
		t.Fatalf("expected exit code 1, got: %v", err)
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Errorf("expected usage message in stderr, got: %s", stderr.String())
	}
}

func TestMultipleArguments(t *testing.T) {
	cmd := exec.Command(testBinaryPath, "foo", "bar")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected exit 1, got exit 0")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 1 {
		t.Fatalf("expected exit code 1, got: %v", err)
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Errorf("expected usage message in stderr, got: %s", stderr.String())
	}
}

func testMCPHandshake(t *testing.T, stdin *bufio.Writer, stdout *bufio.Reader) {
	t.Helper()

	initReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2024-11-05",
			"clientInfo": map[string]any{
				"name":    "test-client",
				"version": "0.0.1",
			},
		},
	}
	line, err := json.Marshal(initReq)
	if err != nil {
		t.Fatalf("marshal initialize: %v", err)
	}
	if _, err := fmt.Fprintf(stdin, "%s\n", line); err != nil {
		t.Fatalf("write initialize: %v", err)
	}
	if err := stdin.Flush(); err != nil {
		t.Fatalf("flush initialize: %v", err)
	}

	respLine, err := stdout.ReadString('\n')
	if err != nil {
		t.Fatalf("read initialize response: %v", err)
	}
	var initResp map[string]any
	if err := json.Unmarshal([]byte(respLine), &initResp); err != nil {
		t.Fatalf("unmarshal initialize response: %v", err)
	}

	notif := map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	}
	notifLine, err := json.Marshal(notif)
	if err != nil {
		t.Fatalf("marshal initialized notification: %v", err)
	}
	if _, err := fmt.Fprintf(stdin, "%s\n", notifLine); err != nil {
		t.Fatalf("write initialized notification: %v", err)
	}
	if err := stdin.Flush(); err != nil {
		t.Fatalf("flush initialized notification: %v", err)
	}
}

func testStartMCPServer(t *testing.T) (*bufio.Writer, *bufio.Reader, context.CancelFunc) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, testBinaryPath)

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		t.Fatalf("stdin pipe: %v", err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		t.Fatalf("stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatalf("start binary: %v", err)
	}

	t.Cleanup(func() {
		cancel()
		cmd.Wait()
	})

	return bufio.NewWriter(stdinPipe), bufio.NewReader(stdoutPipe), cancel
}

func TestToolsListMaxResultSizeChars(t *testing.T) {
	stdin, stdout, _ := testStartMCPServer(t)
	testMCPHandshake(t, stdin, stdout)

	listReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
	}
	line, err := json.Marshal(listReq)
	if err != nil {
		t.Fatalf("marshal tools/list: %v", err)
	}
	if _, err := fmt.Fprintf(stdin, "%s\n", line); err != nil {
		t.Fatalf("write tools/list: %v", err)
	}
	if err := stdin.Flush(); err != nil {
		t.Fatalf("flush tools/list: %v", err)
	}

	respLine, err := stdout.ReadString('\n')
	if err != nil {
		t.Fatalf("read tools/list response: %v", err)
	}

	var resp map[string]any
	if err := json.Unmarshal([]byte(respLine), &resp); err != nil {
		t.Fatalf("unmarshal tools/list response: %v", err)
	}

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result object, got: %v", resp)
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("expected tools array, got: %v", result)
	}

	for _, tool := range tools {
		toolMap, ok := tool.(map[string]any)
		if !ok {
			continue
		}
		if toolMap["name"] == "load_chain" {
			meta, ok := toolMap["_meta"].(map[string]any)
			if !ok {
				t.Fatalf("load_chain tool missing _meta field")
			}
			val, ok := meta["anthropic/maxResultSizeChars"]
			if !ok {
				t.Fatalf("load_chain _meta missing anthropic/maxResultSizeChars")
			}
			numVal, ok := val.(float64)
			if !ok {
				t.Fatalf("anthropic/maxResultSizeChars is not a number: %T", val)
			}
			if int(numVal) != 500000 {
				t.Errorf("expected anthropic/maxResultSizeChars = 500000, got %v", numVal)
			}
			return
		}
	}
	t.Error("load_chain tool not found in tools/list response")
}

func TestToolsListAdvertisesAllTools(t *testing.T) {
	stdin, stdout, _ := testStartMCPServer(t)
	testMCPHandshake(t, stdin, stdout)

	listReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
	}
	line, err := json.Marshal(listReq)
	if err != nil {
		t.Fatalf("marshal tools/list: %v", err)
	}
	if _, err := fmt.Fprintf(stdin, "%s\n", line); err != nil {
		t.Fatalf("write tools/list: %v", err)
	}
	if err := stdin.Flush(); err != nil {
		t.Fatalf("flush tools/list: %v", err)
	}

	respLine, err := stdout.ReadString('\n')
	if err != nil {
		t.Fatalf("read tools/list response: %v", err)
	}

	var resp map[string]any
	if err := json.Unmarshal([]byte(respLine), &resp); err != nil {
		t.Fatalf("unmarshal tools/list response: %v", err)
	}

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result object, got: %v", resp)
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("expected tools array, got: %v", result)
	}

	expected := []string{"load_chain", "write_file", "validate_specs", "chain_hash", "version"}
	found := map[string]bool{}
	for _, tool := range tools {
		toolMap, ok := tool.(map[string]any)
		if !ok {
			continue
		}
		if name, ok := toolMap["name"].(string); ok {
			found[name] = true
		}
	}

	for _, name := range expected {
		if !found[name] {
			t.Errorf("expected tool %q not found in tools/list response", name)
		}
	}
}
