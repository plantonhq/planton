package conversion

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// Conversion specs live beside the kind's protos:
//
//	<provider>/<kind>/conversions/<from>_to_<to>.yaml
//
// Discovery operates on an fs.FS rooted at the provider base
// (catalog), so the same code serves a repo checkout
// (os.DirFS) and the specs embedded into the CLI binary.

// SpecFiles returns every conversion spec path under the provider base.
func SpecFiles(fsys fs.FS) ([]string, error) {
	matches, err := fs.Glob(fsys, "*/*/conversions/*.yaml")
	if err != nil {
		return nil, fmt.Errorf("globbing conversion specs: %w", err)
	}
	sort.Strings(matches)
	return matches, nil
}

// SpecsForKind loads every conversion spec authored for the kind. The kind's
// directory is located by resolving directory names through the registry --
// never by recomposing names -- so lookup follows whatever the registry says.
func SpecsForKind(fsys fs.FS, kind catalogkind.CatalogKind) ([]*Spec, error) {
	files, err := SpecFiles(fsys)
	if err != nil {
		return nil, err
	}
	var specs []*Spec
	for _, file := range files {
		kindDir := path.Base(path.Dir(path.Dir(file)))
		if catalogkindreflect.KindFromString(kindDir) != kind {
			continue
		}
		spec, err := loadSpecFS(fsys, file)
		if err != nil {
			return nil, err
		}
		specs = append(specs, spec)
	}
	return specs, nil
}

func loadSpecFS(fsys fs.FS, file string) (*Spec, error) {
	raw, err := fs.ReadFile(fsys, file)
	if err != nil {
		return nil, fmt.Errorf("reading conversion spec %s: %w", file, err)
	}
	spec, err := parseSpec(raw)
	if err != nil {
		return nil, fmt.Errorf("conversion spec %s: %w", file, err)
	}
	return spec, nil
}

// Step is one hop of a conversion path: a spec plus the direction to run it.
type Step struct {
	Spec      *Spec
	Direction Direction
}

// Path returns the ordered conversion steps from one version to another,
// walking the kind's specs in either direction (a spec authored as
// v1alpha1->v1alpha2 also serves the downgrade). Returns a plain-language
// error naming the missing bridge when no path exists.
func Path(specs []*Spec, from, to string) ([]Step, error) {
	if from == to {
		return nil, nil
	}
	type edge struct {
		next string
		step Step
	}
	adjacency := map[string][]edge{}
	for _, spec := range specs {
		adjacency[spec.From] = append(adjacency[spec.From], edge{spec.To, Step{spec, Upgrade}})
		adjacency[spec.To] = append(adjacency[spec.To], edge{spec.From, Step{spec, Downgrade}})
	}

	// Breadth-first over the version graph (tiny: one or two versions today,
	// a handful ever).
	type node struct {
		version string
		steps   []Step
	}
	visited := map[string]bool{from: true}
	queue := []node{{from, nil}}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, e := range adjacency[current.version] {
			if visited[e.next] {
				continue
			}
			steps := append(append([]Step{}, current.steps...), e.step)
			if e.next == to {
				return steps, nil
			}
			visited[e.next] = true
			queue = append(queue, node{e.next, steps})
		}
	}

	var available []string
	for _, spec := range specs {
		available = append(available, spec.From+"->"+spec.To)
	}
	if len(available) == 0 {
		return nil, fmt.Errorf("no conversion specs exist for this kind -- a bridge from %s to %s must be authored in the kind's conversions/ directory", from, to)
	}
	return nil, fmt.Errorf("no conversion path from %s to %s -- authored bridges: %s", from, to, strings.Join(available, ", "))
}
