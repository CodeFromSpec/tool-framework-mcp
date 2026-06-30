package parsing_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"
)

func TestExtractBlock(t *testing.T) {
	tests := []struct {
		name    string
		content []string
		want    string
	}{
		{
			name:    "removes leading blank lines",
			content: []string{"", "  ", "hello"},
			want:    "hello\n",
		},
		{
			name:    "removes trailing blank lines",
			content: []string{"hello", "", "  "},
			want:    "hello\n",
		},
		{
			name:    "preserves interior blank lines",
			content: []string{"a", "", "b"},
			want:    "a\n\nb\n",
		},
		{
			name:    "empty input returns empty string",
			content: []string{},
			want:    "",
		},
		{
			name:    "all blank lines returns empty string",
			content: []string{"", "  ", "\t"},
			want:    "",
		},
		{
			name:    "ends with exactly one LF",
			content: []string{"hello"},
			want:    "hello\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsing.ExtractBlock(tt.content)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatSection(t *testing.T) {
	tests := []struct {
		name       string
		rawHeading string
		content    []string
		want       string
	}{
		{
			name:       "heading plus content",
			rawHeading: "## Interface",
			content:    []string{"content line"},
			want:       "## Interface\ncontent line\n",
		},
		{
			name:       "strips trailing whitespace from heading",
			rawHeading: "## Interface   ",
			content:    []string{"content"},
			want:       "## Interface\ncontent\n",
		},
		{
			name:       "empty content heading only",
			rawHeading: "## Interface",
			content:    []string{},
			want:       "## Interface\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsing.FormatSection(tt.rawHeading, tt.content)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConcatenateSubsections(t *testing.T) {
	t.Run("single subsection", func(t *testing.T) {
		subs := []*parsing.NodeSubsection{
			{RawHeading: "## A", Content: []string{"line1"}},
		}

		got := parsing.ConcatenateSubsections(subs)
		want := "## A\nline1\n"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("multiple subsections separated by blank line", func(t *testing.T) {
		subs := []*parsing.NodeSubsection{
			{RawHeading: "## A", Content: []string{"a1"}},
			{RawHeading: "## B", Content: []string{"b1"}},
		}

		got := parsing.ConcatenateSubsections(subs)
		want := "## A\na1\n\n## B\nb1\n"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("empty subsection skipped in separation", func(t *testing.T) {
		subs := []*parsing.NodeSubsection{
			{RawHeading: "## A", Content: []string{"a1"}},
			{RawHeading: "## B", Content: []string{}},
			{RawHeading: "## C", Content: []string{"c1"}},
		}

		got := parsing.ConcatenateSubsections(subs)
		want := "## A\na1\n\n## B\n\n## C\nc1\n"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestExtractAgentContent(t *testing.T) {
	t.Run("agent with direct content and subsections", func(t *testing.T) {
		node := &parsing.Node{
			Agent: &parsing.NodeSection{
				Content: []string{"direct line"},
				Subsections: []*parsing.NodeSubsection{
					{RawHeading: "## Logic", Content: []string{"step 1"}},
				},
			},
		}

		got := parsing.ExtractAgentContent(node)
		want := "direct line\n\n## Logic\nstep 1\n"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("agent with no content returns empty", func(t *testing.T) {
		node := &parsing.Node{}

		got := parsing.ExtractAgentContent(node)
		want := ""
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("agent with only blank content lines and no subsections", func(t *testing.T) {
		node := &parsing.Node{
			Agent: &parsing.NodeSection{
				Content: []string{"", "  "},
			},
		}

		got := parsing.ExtractAgentContent(node)
		want := ""
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestReadFileContent(t *testing.T) {
	t.Run("reads file content", func(t *testing.T) {
		testutils.Chdir(t)

		if err := os.WriteFile("test.txt", []byte("line1\nline2\n"), 0o644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		got, err := parsing.ReadFileContent(oslayer.CfsPath("test.txt"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := "line1\nline2\n"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		testutils.Chdir(t)

		_, err := parsing.ReadFileContent(oslayer.CfsPath("missing.txt"))
		if !errors.Is(err, oslayer.ErrFileUnreadable) {
			t.Errorf("expected error wrapping ErrFileUnreadable, got %v", err)
		}
	})
}
