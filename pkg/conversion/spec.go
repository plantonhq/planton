// Package conversion is the declarative api-version conversion engine.
//
// A kind that serves a new api version owes a ConversionSpec: a data file
// authored once beside the kind's protos (conversions/<from>_to_<to>.yaml)
// and executed by every runtime that needs the bridge -- the CLI's offline
// upgrade-manifest, the server boundary, and stored-document migration. The
// spec operates on the DOCUMENT (JSON field names -- the durable contract),
// never on generated structs, so one spec serves every language and every
// document store.
//
// The format is deliberately small. Four operations cover the breaking-change
// classes a version bump can introduce:
//
//	rename:  { from: spec.oldName, to: spec.newName }
//	map:     { path: spec.old, to: spec.new, expr: <CEL over `value`> }
//	default: { path: spec.field, value: <literal> }        # set when absent
//	drop:    { path: spec.field, lossy: { reason: ... } }  # loss is DECLARED
//
// `ops` converts from -> to (the upgrade); `reverse` converts to -> from
// (the downgrade) and must be authored explicitly -- round-trips are law:
// downgrade(upgrade(doc)) == doc except for losses the spec declares.
// `outputPaths` declares output renames so referrer documents (which
// hold other kinds' output paths as data) can be rewritten in the same
// migration run.
package conversion

import (
	"fmt"
	"os"
	"strings"

	"sigs.k8s.io/yaml"
)

// Spec is one kind's conversion between two adjacent api versions.
type Spec struct {
	Kind string `json:"kind"`
	From string `json:"from"`
	To   string `json:"to"`

	// Ops convert a document from `From` to `To` (the upgrade).
	Ops []Op `json:"ops"`

	// Reverse converts a document from `To` back to `From` (the downgrade).
	Reverse []Op `json:"reverse"`

	// OutputPaths declares output field renames between the versions.
	OutputPaths []OutputPath `json:"outputPaths,omitempty"`
}

// Op is exactly one of the four operations. Paths address the document with
// dot-separated JSON field names (e.g. "spec.displayName").
type Op struct {
	Rename  *RenameOp  `json:"rename,omitempty"`
	Map     *MapOp     `json:"map,omitempty"`
	Default *DefaultOp `json:"default,omitempty"`
	Drop    *DropOp    `json:"drop,omitempty"`
}

type RenameOp struct {
	FromPath string `json:"from"`
	ToPath   string `json:"to"`
}

type MapOp struct {
	// Path is the source node; To is the destination (defaults to Path).
	Path string `json:"path"`
	To   string `json:"to,omitempty"`
	// Expr is a CEL expression over `value`, the node at Path. Absent source
	// nodes are a no-op.
	Expr string `json:"expr"`
}

type DefaultOp struct {
	Path  string `json:"path"`
	Value any    `json:"value"`
}

type DropOp struct {
	Path  string `json:"path"`
	Lossy Lossy  `json:"lossy"`
}

// Lossy declares WHY dropping the value is acceptable. Undeclared loss is a
// defect; declared loss surfaces to users.
type Lossy struct {
	Reason string `json:"reason"`
}

type OutputPath struct {
	FromPath string `json:"from"`
	ToPath   string `json:"to"`
}

// LoadSpec reads and structurally validates one conversion spec file.
func LoadSpec(path string) (*Spec, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading conversion spec %s: %w", path, err)
	}
	spec, err := parseSpec(raw)
	if err != nil {
		return nil, fmt.Errorf("conversion spec %s: %w", path, err)
	}
	return spec, nil
}

func parseSpec(raw []byte) (*Spec, error) {
	var spec Spec
	if err := yaml.UnmarshalStrict(raw, &spec); err != nil {
		return nil, fmt.Errorf("malformed: %w", err)
	}
	if err := spec.validate(); err != nil {
		return nil, err
	}
	return &spec, nil
}

func (s *Spec) validate() error {
	var problems []string
	if s.Kind == "" {
		problems = append(problems, "kind is required")
	}
	if s.From == "" || s.To == "" {
		problems = append(problems, "from and to versions are required")
	}
	if len(s.Reverse) == 0 {
		problems = append(problems, "reverse ops are required -- downgrades are part of the contract (declare losses instead of omitting the direction)")
	}
	for i, op := range append(append([]Op{}, s.Ops...), s.Reverse...) {
		set := 0
		if op.Rename != nil {
			set++
			if op.Rename.FromPath == "" || op.Rename.ToPath == "" {
				problems = append(problems, fmt.Sprintf("op %d: rename requires from and to", i))
			}
		}
		if op.Map != nil {
			set++
			if op.Map.Path == "" || op.Map.Expr == "" {
				problems = append(problems, fmt.Sprintf("op %d: map requires path and expr", i))
			}
		}
		if op.Default != nil {
			set++
			if op.Default.Path == "" {
				problems = append(problems, fmt.Sprintf("op %d: default requires path", i))
			}
		}
		if op.Drop != nil {
			set++
			if op.Drop.Path == "" {
				problems = append(problems, fmt.Sprintf("op %d: drop requires path", i))
			}
			if op.Drop.Lossy.Reason == "" {
				problems = append(problems, fmt.Sprintf("op %d: drop must DECLARE its loss with a reason", i))
			}
		}
		if set != 1 {
			problems = append(problems, fmt.Sprintf("op %d: exactly one of rename|map|default|drop must be set (got %d)", i, set))
		}
	}
	for i, op := range s.OutputPaths {
		if op.FromPath == "" || op.ToPath == "" {
			problems = append(problems, fmt.Sprintf("outputPaths %d: from and to are required", i))
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("%s", strings.Join(problems, "; "))
	}
	return nil
}
