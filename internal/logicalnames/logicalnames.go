// code-from-spec: SPEC/golang/implementation/utils/logical_names@oS3s58V9VJ-n64I2fL6k-gkR2gQ
package logicalnames

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

var ErrUnrecognizedPrefix = errors.New("unrecognized logical name prefix")
var ErrInvalidName = errors.New("invalid logical name")
var ErrNoOutput = errors.New("generator node has no output field")
var ErrInvalidPath = errors.New("path does not match expected _node.md pattern")

type NodeType int

const (
	NodeTypeSpec     NodeType = iota
	NodeTypeArtifact
	NodeTypeExternal
)

type LogicalName struct {
	Type      NodeType
	Name      string
	Qualifier *string
	Path      string
	Parent    *string
}

func stringPtr(s string) *string {
	return &s
}

func LogicalNameParse(logicalName string) (*LogicalName, error) {
	var qualifier *string
	stripped := logicalName

	openIdx := strings.Index(logicalName, "(")
	if openIdx != -1 {
		closeIdx := strings.Index(logicalName[openIdx:], ")")
		if closeIdx != -1 {
			q := logicalName[openIdx+1 : openIdx+closeIdx]
			qualifier = stringPtr(q)
		}
		stripped = logicalName[:openIdx]
	}

	if stripped == "SPEC" {
		return &LogicalName{
			Type:      NodeTypeSpec,
			Name:      "SPEC",
			Qualifier: qualifier,
			Path:      "code-from-spec/_node.md",
			Parent:    nil,
		}, nil
	}

	if strings.HasPrefix(stripped, "SPEC/") {
		relative := strings.TrimPrefix(stripped, "SPEC/")
		if relative == "" {
			return nil, fmt.Errorf("%w: %s", ErrInvalidName, logicalName)
		}

		path := "code-from-spec/" + relative + "/_node.md"

		var parent string
		lastSlash := strings.LastIndex(relative, "/")
		if lastSlash == -1 {
			parent = "SPEC"
		} else {
			parent = "SPEC/" + relative[:lastSlash]
		}

		return &LogicalName{
			Type:      NodeTypeSpec,
			Name:      stripped,
			Qualifier: qualifier,
			Path:      path,
			Parent:    stringPtr(parent),
		}, nil
	}

	if strings.HasPrefix(stripped, "ARTIFACT/") {
		relative := strings.TrimPrefix(stripped, "ARTIFACT/")
		if relative == "" {
			return nil, fmt.Errorf("%w: %s", ErrInvalidName, logicalName)
		}

		generatorName := "SPEC/" + relative
		generatorPath := "code-from-spec/" + relative + "/_node.md"

		fm, err := frontmatter.FrontmatterParse(pathutils.PathCfs{Value: generatorPath})
		if err != nil {
			return nil, fmt.Errorf("parsing frontmatter for %s: %w", generatorPath, err)
		}

		if fm.Output == "" {
			return nil, fmt.Errorf("%w: %s", ErrNoOutput, generatorName)
		}

		return &LogicalName{
			Type:      NodeTypeArtifact,
			Name:      stripped,
			Qualifier: nil,
			Path:      fm.Output,
			Parent:    stringPtr(generatorName),
		}, nil
	}

	if strings.HasPrefix(stripped, "EXTERNAL/") {
		relative := strings.TrimPrefix(stripped, "EXTERNAL/")
		if relative == "" {
			return nil, fmt.Errorf("%w: %s", ErrInvalidName, logicalName)
		}

		return &LogicalName{
			Type:      NodeTypeExternal,
			Name:      stripped,
			Qualifier: nil,
			Path:      relative,
			Parent:    nil,
		}, nil
	}

	return nil, fmt.Errorf("%w: %s", ErrUnrecognizedPrefix, logicalName)
}

func LogicalNameFromPath(cfsPath pathutils.PathCfs) (*LogicalName, error) {
	value := cfsPath.Value

	if !strings.HasPrefix(value, "code-from-spec/") {
		return nil, fmt.Errorf("%w: %s", ErrInvalidPath, value)
	}

	if value == "code-from-spec/_node.md" {
		return &LogicalName{
			Type:      NodeTypeSpec,
			Name:      "SPEC",
			Qualifier: nil,
			Path:      "code-from-spec/_node.md",
			Parent:    nil,
		}, nil
	}

	if !strings.HasSuffix(value, "/_node.md") {
		return nil, fmt.Errorf("%w: %s", ErrInvalidPath, value)
	}

	relative := strings.TrimPrefix(value, "code-from-spec/")
	relative = strings.TrimSuffix(relative, "/_node.md")

	if relative == "" {
		return nil, fmt.Errorf("%w: %s", ErrInvalidPath, value)
	}

	name := "SPEC/" + relative

	var parent string
	lastSlash := strings.LastIndex(relative, "/")
	if lastSlash == -1 {
		parent = "SPEC"
	} else {
		parent = "SPEC/" + relative[:lastSlash]
	}

	return &LogicalName{
		Type:      NodeTypeSpec,
		Name:      name,
		Qualifier: nil,
		Path:      value,
		Parent:    stringPtr(parent),
	}, nil
}
