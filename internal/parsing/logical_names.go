// code-from-spec: SPEC/golang/implementation/parsing/logical_names@oYuD1Ar58tSrQ9mA-4hF9eEWyJ4
package parsing

import (
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
)

type CfsNodeType int

const (
	CfsNodeTypeSpec     CfsNodeType = iota
	CfsNodeTypeArtifact CfsNodeType = iota
	CfsNodeTypeExternal CfsNodeType = iota
)

type CfsReference struct {
	NodeType    CfsNodeType
	LogicalName string
	Qualifier   *string
	Path        string
	ParentName  *string
}

func stringPtrLN(s string) *string {
	return &s
}

func extractQualifierLN(logicalName string) (stripped string, qualifier *string) {
	openIdx := strings.Index(logicalName, "(")
	if openIdx == -1 {
		return logicalName, nil
	}
	closeIdx := strings.Index(logicalName[openIdx:], ")")
	if closeIdx == -1 {
		return logicalName, nil
	}
	q := logicalName[openIdx+1 : openIdx+closeIdx]
	return logicalName[:openIdx], stringPtrLN(q)
}

func computeParentLN(prefix, relative string) *string {
	lastSlash := strings.LastIndex(relative, "/")
	if lastSlash == -1 {
		return nil
	}
	return stringPtrLN(prefix + relative[:lastSlash])
}

func CfsReferenceFromName(logicalName string) (*CfsReference, error) {
	stripped, qualifier := extractQualifierLN(logicalName)

	switch {
	case strings.HasPrefix(stripped, "SPEC/"):
		relative := strings.TrimPrefix(stripped, "SPEC/")
		if relative == "" {
			return nil, ErrInvalidName
		}
		path := "code-from-spec/" + relative + "/_node.md"
		parent := computeParentLN("SPEC/", relative)
		return &CfsReference{
			NodeType:    CfsNodeTypeSpec,
			LogicalName: stripped,
			Qualifier:   qualifier,
			Path:        path,
			ParentName:  parent,
		}, nil

	case strings.HasPrefix(stripped, "ARTIFACT/"):
		relative := strings.TrimPrefix(stripped, "ARTIFACT/")
		if relative == "" {
			return nil, ErrInvalidName
		}
		generatorName := "SPEC/" + relative
		node, err := ParseNode(generatorName)
		if err != nil {
			return nil, fmt.Errorf("resolving artifact %q: %w", stripped, err)
		}
		if node.Frontmatter == nil || node.Frontmatter.Output == nil {
			return nil, ErrNoOutput
		}
		return &CfsReference{
			NodeType:    CfsNodeTypeArtifact,
			LogicalName: stripped,
			Qualifier:   nil,
			Path:        *node.Frontmatter.Output,
			ParentName:  stringPtrLN(generatorName),
		}, nil

	case strings.HasPrefix(stripped, "EXTERNAL/"):
		relative := strings.TrimPrefix(stripped, "EXTERNAL/")
		if relative == "" {
			return nil, ErrInvalidName
		}
		return &CfsReference{
			NodeType:    CfsNodeTypeExternal,
			LogicalName: stripped,
			Qualifier:   nil,
			Path:        relative,
			ParentName:  nil,
		}, nil

	default:
		return nil, ErrUnrecognizedPrefix
	}
}

func CfsReferenceFromPath(cfsPath oslayer.CfsPath) (*CfsReference, error) {
	value := string(cfsPath)

	if !strings.HasPrefix(value, "code-from-spec/") {
		return nil, ErrInvalidPath
	}
	if !strings.HasSuffix(value, "/_node.md") {
		return nil, ErrInvalidPath
	}

	relative := strings.TrimPrefix(value, "code-from-spec/")
	relative = strings.TrimSuffix(relative, "/_node.md")
	if relative == "" {
		return nil, ErrInvalidPath
	}

	logicalName := "SPEC/" + relative
	parent := computeParentLN("SPEC/", relative)

	return &CfsReference{
		NodeType:    CfsNodeTypeSpec,
		LogicalName: logicalName,
		Qualifier:   nil,
		Path:        value,
		ParentName:  parent,
	}, nil
}
