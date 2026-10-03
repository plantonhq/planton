// Package catalogkindreflect provides runtime access to CatalogKind metadata
// encoded as protobuf enum value options.
package catalogkindreflect

import (
	"github.com/plantonhq/planton/shared/catalogkind"
)

// Prerequisites returns the direct prerequisites for the given kind.
// Returns nil if the kind has no prerequisites or if kind meta is unavailable.
func Prerequisites(kind catalogkind.CatalogKind) []catalogkind.CatalogKind {
	meta, err := KindMeta(kind)
	if err != nil {
		return nil
	}
	return meta.GetPrerequisites()
}

// TransitivePrerequisites returns all prerequisites in topological order
// (deploy first to last). Resolves the full transitive dependency graph:
// if A depends on B and B depends on C, returns [C, B].
//
// Returns an error if a cycle is detected (indicates a modeling mistake).
func TransitivePrerequisites(kind catalogkind.CatalogKind) ([]catalogkind.CatalogKind, error) {
	return TransitiveClosure(Prerequisites(kind))
}

// TransitiveClosure returns the given kinds plus all of their transitive
// prerequisites in topological order (deploy first to last), deduplicated.
// Every kind appears after everything it depends on, so callers can deploy
// the result front to back. Accepts multiple roots so a caller can expand a
// combined set (e.g. a kind's registry prerequisites merged with extras a
// test scenario composes) in one pass.
//
// Returns an error if a cycle is detected (indicates a modeling mistake).
func TransitiveClosure(kinds []catalogkind.CatalogKind) ([]catalogkind.CatalogKind, error) {
	var result []catalogkind.CatalogKind
	visited := make(map[catalogkind.CatalogKind]bool)
	inStack := make(map[catalogkind.CatalogKind]bool)

	var root catalogkind.CatalogKind
	var visit func(k catalogkind.CatalogKind) error
	visit = func(k catalogkind.CatalogKind) error {
		if inStack[k] {
			return cycleError(root, k)
		}
		if visited[k] {
			return nil
		}

		inStack[k] = true
		for _, prereq := range Prerequisites(k) {
			if err := visit(prereq); err != nil {
				return err
			}
		}
		inStack[k] = false
		visited[k] = true
		result = append(result, k)
		return nil
	}

	for _, k := range kinds {
		root = k
		if err := visit(k); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// HasPrerequisites returns true if the kind has any direct prerequisites.
func HasPrerequisites(kind catalogkind.CatalogKind) bool {
	return len(Prerequisites(kind)) > 0
}

func cycleError(root, cycleAt catalogkind.CatalogKind) error {
	return &CycleError{Root: root, CycleAt: cycleAt}
}

// CycleError indicates a circular dependency in the prerequisite graph.
type CycleError struct {
	Root    catalogkind.CatalogKind
	CycleAt catalogkind.CatalogKind
}

func (e *CycleError) Error() string {
	return "prerequisite cycle detected: " + e.Root.String() + " -> ... -> " + e.CycleAt.String()
}
