package chainresolver

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/parsing"
)

var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")
var ErrUnresolvableArtifact = errors.New("unresolvable artifact")

type Chain struct {
	Ancestors []parsing.CfsReference
	Imports   []parsing.CfsReference
	Target    parsing.CfsReference
	Input     []parsing.CfsReference
}

func ChainResolve(targetLogicalName string, knownSpecNodes []string) (Chain, error) {
	targetRef, err := parsing.CfsReferenceFromName(targetLogicalName)
	if err != nil {
		return Chain{}, err
	}

	ancestors, target, err := resolveAncestorsAndTarget(targetRef)
	if err != nil {
		return Chain{}, err
	}

	node, err := parsing.ParseNode(targetLogicalName)
	if err != nil {
		return Chain{}, fmt.Errorf("%w: %w", ErrUnreadableFrontmatter, err)
	}

	var fm *parsing.NodeFrontmatter
	if node.Frontmatter != nil {
		fm = node.Frontmatter
	}

	expandedImports, err := expandGlobs(getImports(fm), knownSpecNodes, targetLogicalName)
	if err != nil {
		return Chain{}, err
	}

	imports, err := resolveAndDeduplicateRefs(expandedImports)
	if err != nil {
		return Chain{}, err
	}

	expandedInput, err := expandGlobs(getInput(fm), knownSpecNodes, targetLogicalName)
	if err != nil {
		return Chain{}, err
	}

	input, err := resolveAndDeduplicateRefs(expandedInput)
	if err != nil {
		return Chain{}, err
	}

	return Chain{
		Ancestors: ancestors,
		Imports:   imports,
		Target:    target,
		Input:     input,
	}, nil
}

func getImports(fm *parsing.NodeFrontmatter) []string {
	if fm == nil {
		return nil
	}
	return fm.Imports
}

func getInput(fm *parsing.NodeFrontmatter) []string {
	if fm == nil {
		return nil
	}
	return fm.Input
}

func expandGlobs(entries []string, knownSpecNodes []string, declaringNode string) ([]string, error) {
	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		if strings.HasSuffix(entry, "/*") {
			expanded, err := parsing.ExpandGlob(entry, knownSpecNodes, &declaringNode)
			if err != nil {
				return nil, err
			}
			result = append(result, expanded...)
		} else {
			result = append(result, entry)
		}
	}
	return result, nil
}

func resolveAncestorsAndTarget(targetRef *parsing.CfsReference) ([]parsing.CfsReference, parsing.CfsReference, error) {
	if targetRef.ParentName == nil {
		return []parsing.CfsReference{}, *targetRef, nil
	}

	refs := []parsing.CfsReference{*targetRef}
	currentRef := targetRef

	for {
		if currentRef.ParentName == nil {
			break
		}
		parentRef, err := parsing.CfsReferenceFromName(*currentRef.ParentName)
		if err != nil {
			return nil, parsing.CfsReference{}, err
		}
		refs = append(refs, *parentRef)
		currentRef = parentRef
	}

	sort.Slice(refs, func(i, j int) bool {
		return refs[i].LogicalName < refs[j].LogicalName
	})

	target := refs[len(refs)-1]
	ancestors := refs[:len(refs)-1]
	return ancestors, target, nil
}

func resolveAndDeduplicateRefs(entries []string) ([]parsing.CfsReference, error) {
	refs := make([]parsing.CfsReference, 0, len(entries))

	for _, entry := range entries {
		ref, err := parsing.CfsReferenceFromName(entry)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrUnresolvableArtifact, err)
		}
		refs = append(refs, *ref)
	}

	sort.Slice(refs, func(i, j int) bool {
		if refs[i].LogicalName != refs[j].LogicalName {
			return refs[i].LogicalName < refs[j].LogicalName
		}
		if refs[i].Qualifier == nil && refs[j].Qualifier != nil {
			return true
		}
		if refs[i].Qualifier != nil && refs[j].Qualifier == nil {
			return false
		}
		if refs[i].Qualifier != nil && refs[j].Qualifier != nil {
			return *refs[i].Qualifier < *refs[j].Qualifier
		}
		return false
	})

	return deduplicateRefs(refs), nil
}

func deduplicateRefs(refs []parsing.CfsReference) []parsing.CfsReference {
	result := make([]parsing.CfsReference, 0, len(refs))

	for _, ref := range refs {
		if strings.HasPrefix(ref.LogicalName, "SPEC/") {
			if isSpecDuplicate(result, ref) {
				continue
			}
		} else {
			if isNameDuplicate(result, ref.LogicalName) {
				continue
			}
		}
		result = append(result, ref)
	}

	return result
}

func isSpecDuplicate(existing []parsing.CfsReference, candidate parsing.CfsReference) bool {
	for _, e := range existing {
		if e.LogicalName != candidate.LogicalName {
			continue
		}
		if e.Qualifier == nil {
			return true
		}
		if candidate.Qualifier != nil && e.Qualifier != nil && *e.Qualifier == *candidate.Qualifier {
			return true
		}
	}
	return false
}

func isNameDuplicate(existing []parsing.CfsReference, name string) bool {
	for _, e := range existing {
		if e.LogicalName == name {
			return true
		}
	}
	return false
}
