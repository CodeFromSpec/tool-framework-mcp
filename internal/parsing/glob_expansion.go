package parsing

import (
	"sort"
	"strings"
)

func ExpandGlob(pattern string, knownNodes []string, declaringNode *string) ([]string, error) {
	if len(knownNodes) == 0 {
		return nil, ErrEmptyNodeList
	}

	if declaringNode != nil && !strings.HasPrefix(*declaringNode, "SPEC/") {
		return nil, ErrInvalidName
	}

	if !strings.HasSuffix(pattern, "/*") {
		return nil, ErrInvalidGlob
	}

	if strings.Contains(pattern, "(") {
		return nil, ErrInvalidGlob
	}

	basePath := strings.TrimSuffix(pattern, "/*")

	var prefix string
	var relative string

	if strings.HasPrefix(basePath, "SPEC/") {
		prefix = "SPEC/"
		relative = strings.TrimPrefix(basePath, "SPEC/")
	} else if strings.HasPrefix(basePath, "ARTIFACT/") {
		prefix = "ARTIFACT/"
		relative = strings.TrimPrefix(basePath, "ARTIFACT/")
	} else {
		return nil, ErrInvalidGlob
	}

	if relative == "" {
		return nil, ErrInvalidGlob
	}

	if strings.Contains(relative, "*") {
		return nil, ErrInvalidGlob
	}

	matchPrefix := "SPEC/" + relative + "/"

	var results []string

	for _, name := range knownNodes {
		if strings.HasPrefix(name, matchPrefix) {
			if prefix == "SPEC/" {
				results = append(results, name)
			} else {
				specRelative := strings.TrimPrefix(name, "SPEC/")
				results = append(results, "ARTIFACT/"+specRelative)
			}
		}
	}

	if declaringNode != nil {
		declRelative := strings.TrimPrefix(*declaringNode, "SPEC/")

		excludedSet := make(map[string]bool)
		current := declRelative
		for current != "" {
			excludedSet[prefix+current] = true
			idx := strings.LastIndex(current, "/")
			if idx < 0 {
				break
			}
			current = current[:idx]
		}

		filtered := results[:0]
		for _, name := range results {
			if !excludedSet[name] {
				filtered = append(filtered, name)
			}
		}
		results = filtered
	}

	sort.Strings(results)
	return results, nil
}
