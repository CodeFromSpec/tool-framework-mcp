package chainhash

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"
)

var ErrParseFailure = errors.New("parse failure")

type ContentHash struct {
	Label string
	Hash  string
}

func hashPublicSubsections(node *parsing.Node) []byte {
	if node.Public == nil {
		return nil
	}
	if len(node.Public.Subsections) == 0 {
		return nil
	}
	text := parsing.ConcatenateSubsections(node.Public.Subsections)
	sum := sha1.Sum([]byte(text))
	return sum[:]
}

func hashQualifiedSubsection(node *parsing.Node, qualifier string) []byte {
	normalizedQualifier := parsing.NormalizeText(qualifier)
	if node.Public == nil {
		return nil
	}
	for _, sub := range node.Public.Subsections {
		if sub.Heading == normalizedQualifier {
			text := parsing.FormatSection(sub.RawHeading, sub.Content)
			sum := sha1.Sum([]byte(text))
			return sum[:]
		}
	}
	return nil
}

func hashAgentSection(node *parsing.Node) []byte {
	text := parsing.ExtractAgentContent(node)
	if text == "" {
		return nil
	}
	sum := sha1.Sum([]byte(text))
	return sum[:]
}

func hashFileContent(filePath oslayer.CfsPath) ([]byte, error) {
	text, err := parsing.ReadFileContent(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", filePath, err)
	}
	sum := sha1.Sum([]byte(text))
	return sum[:], nil
}

func referenceLabel(ref parsing.CfsReference) string {
	label := ref.LogicalName
	if ref.Qualifier != nil {
		label += "(" + *ref.Qualifier + ")"
	}
	return label
}

func processSpecDep(ref parsing.CfsReference) ([]byte, error) {
	node, err := parsing.ParseNode(ref.LogicalName)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %s", ErrParseFailure, ref.LogicalName, err)
	}
	if ref.Qualifier == nil {
		return hashPublicSubsections(node), nil
	}
	return hashQualifiedSubsection(node, *ref.Qualifier), nil
}

func ChainHashCompute(chain chainresolver.Chain) (string, []ContentHash, error) {
	var hashes [][]byte
	var positions []ContentHash

	recordPosition := func(label string, rawHash []byte) {
		encoded := base64.RawURLEncoding.EncodeToString(rawHash)
		positions = append(positions, ContentHash{Label: label, Hash: encoded})
		hashes = append(hashes, rawHash)
	}

	for _, ancestor := range chain.Ancestors {
		node, err := parsing.ParseNode(ancestor.LogicalName)
		if err != nil {
			return "", nil, fmt.Errorf("%w: ancestor %s: %s", ErrParseFailure, ancestor.LogicalName, err)
		}
		h := hashPublicSubsections(node)
		if h != nil {
			recordPosition(ancestor.LogicalName, h)
		}
	}

	for _, dep := range chain.Dependencies {
		label := referenceLabel(dep)
		if strings.HasPrefix(dep.LogicalName, "ARTIFACT/") {
			h, err := hashFileContent(oslayer.CfsPath(dep.Path))
			if err != nil {
				return "", nil, err
			}
			recordPosition(label, h)
		} else if strings.HasPrefix(dep.LogicalName, "EXTERNAL/") {
			h, err := hashFileContent(oslayer.CfsPath(dep.Path))
			if err != nil {
				return "", nil, err
			}
			recordPosition(label, h)
		} else if strings.HasPrefix(dep.LogicalName, "SPEC/") {
			h, err := processSpecDep(dep)
			if err != nil {
				return "", nil, err
			}
			if h != nil {
				recordPosition(label, h)
			}
		}
	}

	targetNode, err := parsing.ParseNode(chain.Target.LogicalName)
	if err != nil {
		return "", nil, fmt.Errorf("%w: target %s: %s", ErrParseFailure, chain.Target.LogicalName, err)
	}

	h := hashPublicSubsections(targetNode)
	if h != nil {
		recordPosition(chain.Target.LogicalName, h)
	}

	agentHash := hashAgentSection(targetNode)
	if agentHash != nil {
		recordPosition("AGENT["+chain.Target.LogicalName+"]", agentHash)
	}

	if chain.Input != nil {
		hashes = append(hashes, []byte{0x49})
		input := chain.Input
		inputLabel := "INPUT[" + referenceLabel(*input) + "]"
		if strings.HasPrefix(input.LogicalName, "ARTIFACT/") {
			h, err := hashFileContent(oslayer.CfsPath(input.Path))
			if err != nil {
				return "", nil, err
			}
			recordPosition(inputLabel, h)
		} else if strings.HasPrefix(input.LogicalName, "EXTERNAL/") {
			h, err := hashFileContent(oslayer.CfsPath(input.Path))
			if err != nil {
				return "", nil, err
			}
			recordPosition(inputLabel, h)
		} else if strings.HasPrefix(input.LogicalName, "SPEC/") {
			h, err := processSpecDep(*input)
			if err != nil {
				return "", nil, err
			}
			if h != nil {
				recordPosition(inputLabel, h)
			}
		}
	}

	var concatenated []byte
	for _, h := range hashes {
		concatenated = append(concatenated, h...)
	}

	finalSum := sha1.Sum(concatenated)
	encoded := base64.RawURLEncoding.EncodeToString(finalSum[:])
	return encoded, positions, nil
}
