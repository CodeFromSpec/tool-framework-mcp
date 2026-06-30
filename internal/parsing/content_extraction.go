package parsing

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
)

func isBlankLineCE(line string) bool {
	return strings.Trim(line, " \t") == ""
}

func ExtractBlock(content []string) string {
	start := 0
	end := len(content)
	for start < end && isBlankLineCE(content[start]) {
		start++
	}
	for end > start && isBlankLineCE(content[end-1]) {
		end--
	}
	if start == end {
		return ""
	}
	return strings.Join(content[start:end], "\n") + "\n"
}

func FormatSection(rawHeading string, content []string) string {
	head := strings.TrimRight(rawHeading, " \t") + "\n"
	body := ExtractBlock(content)
	return head + body
}

func ConcatenateSubsections(subsections []*NodeSubsection) string {
	result := ""
	for _, subsection := range subsections {
		if subsection == nil {
			continue
		}
		block := FormatSection(subsection.RawHeading, subsection.Content)
		if result != "" && block != "" {
			result += "\n"
		}
		result += block
	}
	return result
}

func ExtractAgentContent(node *Node) string {
	if node == nil || node.Agent == nil {
		return ""
	}
	text := ExtractBlock(node.Agent.Content)
	for _, subsection := range node.Agent.Subsections {
		if subsection == nil {
			continue
		}
		subBlock := FormatSection(subsection.RawHeading, subsection.Content)
		if text != "" && subBlock != "" {
			text += "\n"
		}
		text += subBlock
	}
	if text == "" {
		return ""
	}
	return text
}

func ReadFileContent(cfsPath oslayer.CfsPath) (string, error) {
	handle, err := oslayer.OpenFile(cfsPath, "read", 30000)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}

	var lines []string
	for {
		line, err := handle.ReadLine()
		if err != nil {
			if errors.Is(err, oslayer.ErrEndOfFile) {
				break
			}
			handle.Close()
			return "", fmt.Errorf("failed to read line: %w", err)
		}
		lines = append(lines, line)
	}
	handle.Close()

	text := strings.Join(lines, "\n") + "\n"
	return text, nil
}
