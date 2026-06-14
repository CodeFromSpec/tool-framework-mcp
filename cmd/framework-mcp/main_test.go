// code-from-spec: ROOT/golang/tests/server@KUfixTlgZBP60kr6UKuSA7Lcn-A
package main_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "framework-mcp-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	name := "framework-mcp"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binaryPath = tmpDir + string(os.PathSeparator) + name

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build binary: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestHelpFlag(t *testing.T) {
	out, err := exec.Command(binaryPath, "--help").Output()
	if err != nil {
		t.Fatalf("expected exit 0, got: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("expected stdout to contain usage message, got: %s", string(out))
	}
}

func TestHelpWord(t *testing.T) {
	out, err := exec.Command(binaryPath, "help").Output()
	if err != nil {
		t.Fatalf("expected exit 0, got: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("expected stdout to contain usage message, got: %s", string(out))
	}
}

func TestShortHelpFlag(t *testing.T) {
	out, err := exec.Command(binaryPath, "-h").Output()
	if err != nil {
		t.Fatalf("expected exit 0, got: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("expected stdout to contain usage message, got: %s", string(out))
	}
}

func TestUnrecognizedArgument(t *testing.T) {
	cmd := exec.Command(binaryPath, "something")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected exit 1, got exit 0")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got: %v", err)
	}
	if exitErr.ExitCode() != 1 {
		t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Errorf("expected stderr to contain usage message, got: %s", stderr.String())
	}
}

func TestMultipleArguments(t *testing.T) {
	cmd := exec.Command(binaryPath, "foo", "bar")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected exit 1, got exit 0")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got: %v", err)
	}
	if exitErr.ExitCode() != 1 {
		t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Errorf("expected stderr to contain usage message, got: %s", stderr.String())
	}
}

type testMCPSession struct {
	cmd    *exec.Cmd
	stdin  *bytes.Buffer
	stdout *bufio.Reader
	cancel context.CancelFunc
}

func testStartMCPSession(t *testing.T) *testMCPSession {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctx, binaryPath)
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		t.Fatalf("failed to get stdin pipe: %v", err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		t.Fatalf("failed to get stdout pipe: %v", err)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatalf("failed to start binary: %v", err)
	}

	sess := &testMCPSession{
		cmd:    cmd,
		stdout: bufio.NewReader(stdoutPipe),
		cancel: cancel,
	}

	t.Cleanup(func() {
		cancel()
		stdinPipe.Close()
		cmd.Wait()
	})

	testMCPHandshake(t, stdinPipe, sess.stdout)

	sess.stdin = &bytes.Buffer{}
	_ = sess.stdin

	stdinRef := stdinPipe
	sess.stdin = &bytes.Buffer{}
	_ = stdinRef

	sendFn := func(msg map[string]any) {
		data, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("failed to marshal message: %v", err)
		}
		data = append(data, '\n')
		if _, err := stdinPipe.Write(data); err != nil {
			t.Fatalf("failed to write to stdin: %v", err)
		}
	}

	sess.stdin = &bytes.Buffer{}
	_ = sendFn

	return &testMCPSession{
		cmd:    cmd,
		stdout: sess.stdout,
		cancel: cancel,
	}
}

func testSendMessage(t *testing.T, stdin interface{ Write([]byte) (int, error) }, msg map[string]any) {
	t.Helper()
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("testSendMessage: marshal: %v", err)
	}
	data = append(data, '\n')
	if _, err := stdin.Write(data); err != nil {
		t.Fatalf("testSendMessage: write: %v", err)
	}
}

func testReadMessage(t *testing.T, reader *bufio.Reader) map[string]any {
	t.Helper()
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("testReadMessage: %v", err)
	}
	var msg map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &msg); err != nil {
		t.Fatalf("testReadMessage: unmarshal: %v", err)
	}
	return msg
}

func testMCPHandshake(t *testing.T, stdin interface{ Write([]byte) (int, error) }, stdout *bufio.Reader) {
	t.Helper()

	initReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2024-11-05",
			"clientInfo": map[string]any{
				"name":    "test-client",
				"version": "1.0",
			},
			"capabilities": map[string]any{},
		},
	}
	testSendMessage(t, stdin, initReq)

	testReadMessage(t, stdout)

	notif := map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	}
	testSendMessage(t, stdin, notif)
}

func TestToolsListAdvertisesMaxResultSizeCharsForLoadChain(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctx, binaryPath)
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		t.Fatalf("failed to get stdin pipe: %v", err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		t.Fatalf("failed to get stdout pipe: %v", err)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatalf("failed to start binary: %v", err)
	}
	t.Cleanup(func() {
		cancel()
		stdinPipe.Close()
		cmd.Wait()
	})

	stdout := bufio.NewReader(stdoutPipe)
	testMCPHandshake(t, stdinPipe, stdout)

	toolsListReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]any{},
	}
	testSendMessage(t, stdinPipe, toolsListReq)

	resp := testReadMessage(t, stdout)

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result object, got: %v", resp)
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("expected tools array, got: %v", result)
	}

	var loadChainTool map[string]any
	for _, toolAny := range tools {
		tool, ok := toolAny.(map[string]any)
		if !ok {
			continue
		}
		if tool["name"] == "load_chain" {
			loadChainTool = tool
			break
		}
	}

	if loadChainTool == nil {
		t.Fatal("load_chain tool not found in tools/list response")
	}

	meta, ok := loadChainTool["_meta"].(map[string]any)
	if !ok {
		t.Fatalf("expected _meta object in load_chain tool, got: %v", loadChainTool)
	}

	maxSize, ok := meta["anthropic/maxResultSizeChars"]
	if !ok {
		t.Fatal("expected anthropic/maxResultSizeChars in _meta")
	}

	maxSizeFloat, ok := maxSize.(float64)
	if !ok {
		t.Fatalf("expected numeric anthropic/maxResultSizeChars, got: %T %v", maxSize, maxSize)
	}

	if int(maxSizeFloat) != 500000 {
		t.Errorf("expected anthropic/maxResultSizeChars to be 500000, got %v", maxSizeFloat)
	}
}

func TestToolsListAdvertisesAllTools(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctx, binaryPath)
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		t.Fatalf("failed to get stdin pipe: %v", err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		t.Fatalf("failed to get stdout pipe: %v", err)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatalf("failed to start binary: %v", err)
	}
	t.Cleanup(func() {
		cancel()
		stdinPipe.Close()
		cmd.Wait()
	})

	stdout := bufio.NewReader(stdoutPipe)
	testMCPHandshake(t, stdinPipe, stdout)

	toolsListReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]any{},
	}
	testSendMessage(t, stdinPipe, toolsListReq)

	resp := testReadMessage(t, stdout)

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result object, got: %v", resp)
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("expected tools array, got: %v", result)
	}

	toolNames := make(map[string]bool)
	for _, toolAny := range tools {
		tool, ok := toolAny.(map[string]any)
		if !ok {
			continue
		}
		if name, ok := tool["name"].(string); ok {
			toolNames[name] = true
		}
	}

	expectedTools := []string{"load_chain", "write_file", "validate_specs", "chain_hash", "version"}
	for _, expected := range expectedTools {
		if !toolNames[expected] {
			t.Errorf("expected tool %q to be present in tools/list response", expected)
		}
	}
}
