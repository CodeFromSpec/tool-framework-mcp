package parsingglobexpansiontest

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/parsing"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/testutils"
)

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestExpandGlob_ValidPatterns(t *testing.T) {
	cases := []struct {
		name          string
		pattern       string
		knownNodes    []string
		declaringNode *string
		want          []string
	}{
		{
			name:       "spec glob matches descendants",
			pattern:    "SPEC/a/*",
			knownNodes: []string{"SPEC/a", "SPEC/a/b", "SPEC/a/b/c", "SPEC/a/d"},
			want:       []string{"SPEC/a/b", "SPEC/a/b/c", "SPEC/a/d"},
		},
		{
			name:       "artifact glob converts prefix",
			pattern:    "ARTIFACT/a/*",
			knownNodes: []string{"SPEC/a", "SPEC/a/b", "SPEC/a/c"},
			want:       []string{"ARTIFACT/a/b", "ARTIFACT/a/c"},
		},
		{
			name:       "verdict glob converts prefix",
			pattern:    "VERDICT/a/*",
			knownNodes: []string{"SPEC/a", "SPEC/a/b", "SPEC/a/c"},
			want:       []string{"VERDICT/a/b", "VERDICT/a/c"},
		},
		{
			name:          "glob excludes declaring node",
			pattern:       "SPEC/a/*",
			knownNodes:    []string{"SPEC/a", "SPEC/a/b", "SPEC/a/c"},
			declaringNode: testutils.Ptr("SPEC/a/b"),
			want:          []string{"SPEC/a/c"},
		},
		{
			name:          "glob excludes declaring node ancestors",
			pattern:       "SPEC/a/*",
			knownNodes:    []string{"SPEC/a", "SPEC/a/b", "SPEC/a/b/c", "SPEC/a/d"},
			declaringNode: testutils.Ptr("SPEC/a/b/c"),
			want:          []string{"SPEC/a/d"},
		},
		{
			name:          "artifact glob excludes declaring node counterpart",
			pattern:       "ARTIFACT/a/*",
			knownNodes:    []string{"SPEC/a", "SPEC/a/b", "SPEC/a/c"},
			declaringNode: testutils.Ptr("SPEC/a/b"),
			want:          []string{"ARTIFACT/a/c"},
		},
		{
			name:       "empty match no error",
			pattern:    "SPEC/a/*",
			knownNodes: []string{"SPEC/a", "SPEC/b"},
			want:       []string{},
		},
		{
			name:       "result is sorted",
			pattern:    "SPEC/a/*",
			knownNodes: []string{"SPEC/a", "SPEC/a/z", "SPEC/a/m", "SPEC/a/b"},
			want:       []string{"SPEC/a/b", "SPEC/a/m", "SPEC/a/z"},
		},
		{
			name:          "declaring node nil no exclusion",
			pattern:       "SPEC/a/*",
			knownNodes:    []string{"SPEC/a", "SPEC/a/b"},
			declaringNode: nil,
			want:          []string{"SPEC/a/b"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parsing.ExpandGlob(tc.pattern, tc.knownNodes, tc.declaringNode)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !slicesEqual(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestExpandGlob_InvalidParameters(t *testing.T) {
	cases := []struct {
		name          string
		pattern       string
		knownNodes    []string
		declaringNode *string
		wantErr       error
	}{
		{
			name:    "knownNodes nil",
			pattern: "SPEC/a/*",
			wantErr: parsing.ErrEmptyNodeList,
		},
		{
			name:       "knownNodes empty",
			pattern:    "SPEC/a/*",
			knownNodes: []string{},
			wantErr:    parsing.ErrEmptyNodeList,
		},
		{
			name:          "declaring node not spec prefix",
			pattern:       "SPEC/a/*",
			knownNodes:    []string{"SPEC/a", "SPEC/a/b"},
			declaringNode: testutils.Ptr("ARTIFACT/a"),
			wantErr:       parsing.ErrInvalidName,
		},
		{
			name:          "declaring node empty string",
			pattern:       "SPEC/a/*",
			knownNodes:    []string{"SPEC/a", "SPEC/a/b"},
			declaringNode: testutils.Ptr(""),
			wantErr:       parsing.ErrInvalidName,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parsing.ExpandGlob(tc.pattern, tc.knownNodes, tc.declaringNode)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got error %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestExpandGlob_InvalidPatterns(t *testing.T) {
	knownNodes := []string{"SPEC/dummy"}

	cases := []struct {
		name    string
		pattern string
	}{
		{
			name:    "external glob",
			pattern: "EXTERNAL/docs/*",
		},
		{
			name:    "partial wildcard",
			pattern: "SPEC/a/foo*",
		},
		{
			name:    "multiple wildcards",
			pattern: "SPEC/*/a/*",
		},
		{
			name:    "qualifier on glob",
			pattern: "SPEC/a/*(interface)",
		},
		{
			name:    "missing relative path",
			pattern: "SPEC/*",
		},
		{
			name:    "unrecognized prefix",
			pattern: "UNKNOWN/a/*",
		},
		{
			name:    "pattern without glob suffix",
			pattern: "SPEC/a/b",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parsing.ExpandGlob(tc.pattern, knownNodes, nil)
			if !errors.Is(err, parsing.ErrInvalidGlob) {
				t.Errorf("got error %v, want ErrInvalidGlob", err)
			}
		})
	}
}
