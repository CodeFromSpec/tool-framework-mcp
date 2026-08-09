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
	Input     *parsing.CfsReference
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

	imports, err := resolveImports(fm)
	if err != nil {
		return Chain{}, err
	}

	imports = deduplicateImports(imports)

	input, err := resolveInput(fm)
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

func resolveImports(fm *parsing.NodeFrontmatter) ([]parsing.CfsReference, error) {
	if fm == nil {
		return []parsing.CfsReference{}, nil
	}

	imports := make([]parsing.CfsReference, 0, len(fm.Imports))

	for _, entry := range fm.Imports {
		ref, err := parsing.CfsReferenceFromName(entry)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrUnresolvableArtifact, err)
		}
		imports = append(imports, *ref)
	}

	sort.Slice(imports, func(i, j int) bool {
		if imports[i].LogicalName != imports[j].LogicalName {
			return imports[i].LogicalName < imports[j].LogicalName
		}
		if imports[i].Qualifier == nil && imports[j].Qualifier != nil {
			return true
		}
		if imports[i].Qualifier != nil && imports[j].Qualifier == nil {
			return false
		}
		if imports[i].Qualifier != nil && imports[j].Qualifier != nil {
			return *imports[i].Qualifier < *imports[j].Qualifier
		}
		return false
	})

	return imports, nil
}

func deduplicateImports(imports []parsing.CfsReference) []parsing.CfsReference {
	result := make([]parsing.CfsReference, 0, len(imports))

	for _, imp := range imports {
		if strings.HasPrefix(imp.LogicalName, "SPEC/") {
			if isSpecDuplicate(result, imp) {
				continue
			}
		} else {
			if isNameDuplicate(result, imp.LogicalName) {
				continue
			}
		}
		result = append(result, imp)
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
