// code-from-spec: SPEC/golang/implementation/parsing/logical_names@3fOgxLOrrAWZ-kTlVXOBSi3D0qo
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

func computeParentLN(prefix, relative string) *string {
	idx := strings.LastIndex(relative, "/")
	if idx < 0 {
		return nil
	}
	return stringPtrLN(prefix + relative[:idx])
}

func CfsReferenceFromName(logicalName string) (*CfsReference, error) {
	var qualifier *string
	stripped := logicalName

	if idx := strings.Index(logicalName, "("); idx >= 0 {
		closing := strings.Index(logicalName[idx:], ")")
		if closing >= 0 {
			q := logicalName[idx+1 : idx+closing]
			qualifier = stringPtrLN(q)
			stripped = logicalName[:idx]
		}
	}

	switch {
	case strings.HasPrefix(stripped, "SPEC/"):
		relative := strings.TrimPrefix(stripped, "SPEC/")
		if relative == "" || strings.HasSuffix(relative, "/") {
			return nil, fmt.Errorf("%w: %q", ErrInvalidName, logicalName)
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
			return nil, fmt.Errorf("%w: %q", ErrInvalidName, logicalName)
		}
		generatorName := "SPEC/" + relative
		node, err := ParseNode(generatorName)
		if err != nil {
			return nil, fmt.Errorf("resolving artifact %q: %w", logicalName, err)
		}
		if node.Frontmatter == nil || node.Frontmatter.Output == nil {
			return nil, fmt.Errorf("%w: %q", ErrNoOutput, logicalName)
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
			return nil, fmt.Errorf("%w: %q", ErrInvalidName, logicalName)
		}
		return &CfsReference{
			NodeType:    CfsNodeTypeExternal,
			LogicalName: stripped,
			Qualifier:   nil,
			Path:        relative,
			ParentName:  nil,
		}, nil

	default:
		return nil, fmt.Errorf("%w: %q", ErrUnrecognizedPrefix, logicalName)
	}
}

func CfsReferenceFromPath(cfsPath oslayer.CfsPath) (*CfsReference, error) {
	value := string(cfsPath)

	const specPrefix = "code-from-spec/"
	const nodeSuffix = "/_node.md"

	if !strings.HasPrefix(value, specPrefix) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidPath, value)
	}
	value = strings.TrimPrefix(value, specPrefix)

	if !strings.HasSuffix(value, nodeSuffix) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidPath, string(cfsPath))
	}
	value = value[:len(value)-len(nodeSuffix)]

	if value == "" {
		return nil, fmt.Errorf("%w: %q", ErrInvalidPath, string(cfsPath))
	}

	logicalName := "SPEC/" + value
	parent := computeParentLN("SPEC/", value)

	return &CfsReference{
		NodeType:    CfsNodeTypeSpec,
		LogicalName: logicalName,
		Qualifier:   nil,
		Path:        string(cfsPath),
		ParentName:  parent,
	}, nil
}
