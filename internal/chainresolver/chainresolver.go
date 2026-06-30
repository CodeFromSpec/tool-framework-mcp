package chainresolver

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"
)

var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")
var ErrUnresolvableArtifact = errors.New("unresolvable artifact")

type Chain struct {
	Ancestors    []parsing.CfsReference
	Dependencies []parsing.CfsReference
	Target       parsing.CfsReference
	Input        *parsing.CfsReference
}

func ChainResolve(targetLogicalName string) (Chain, error) {
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

	deps, err := resolveDependencies(fm)
	if err != nil {
		return Chain{}, err
	}

	deps = deduplicateDependencies(deps)

	input, err := resolveInput(fm)
	if err != nil {
		return Chain{}, err
	}

	return Chain{
		Ancestors:    ancestors,
		Dependencies: deps,
		Target:       target,
		Input:        input,
	}, nil
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

func resolveDependencies(fm *parsing.NodeFrontmatter) ([]parsing.CfsReference, error) {
	if fm == nil {
		return []parsing.CfsReference{}, nil
	}

	deps := make([]parsing.CfsReference, 0, len(fm.DependsOn))

	for _, entry := range fm.DependsOn {
		ref, err := parsing.CfsReferenceFromName(entry)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrUnresolvableArtifact, err)
		}
		deps = append(deps, *ref)
	}

	sort.Slice(deps, func(i, j int) bool {
		if deps[i].LogicalName != deps[j].LogicalName {
			return deps[i].LogicalName < deps[j].LogicalName
		}
		if deps[i].Qualifier == nil && deps[j].Qualifier != nil {
			return true
		}
		if deps[i].Qualifier != nil && deps[j].Qualifier == nil {
			return false
		}
		if deps[i].Qualifier != nil && deps[j].Qualifier != nil {
			return *deps[i].Qualifier < *deps[j].Qualifier
		}
		return false
	})

	return deps, nil
}

func deduplicateDependencies(deps []parsing.CfsReference) []parsing.CfsReference {
	result := make([]parsing.CfsReference, 0, len(deps))

	for _, dep := range deps {
		if strings.HasPrefix(dep.LogicalName, "SPEC/") {
			if isSpecDuplicate(result, dep) {
				continue
			}
		} else {
			if isNameDuplicate(result, dep.LogicalName) {
				continue
			}
		}
		result = append(result, dep)
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

func resolveInput(fm *parsing.NodeFrontmatter) (*parsing.CfsReference, error) {
	if fm == nil || fm.Input == nil {
		return nil, nil
	}

	ref, err := parsing.CfsReferenceFromName(*fm.Input)
	if err != nil {
		return nil, err
	}

	return ref, nil
}
