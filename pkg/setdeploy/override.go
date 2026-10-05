package setdeploy

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/internal/manifest/manifestprotobuf"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/protobufyaml"
	"github.com/plantonhq/planton/pkg/reflection/metadatareflect"
	"google.golang.org/protobuf/encoding/protojson"
)

// Node-addressed field overrides. In a SET, a bare dotted field path is
// ambiguous — which document's field? — so the override names its node:
// `<kind>/<name>:<fieldPath>=<value>`. This is the one home for the
// semantics: the single-manifest lane's `--set` uses the same dotted-path
// setter underneath (never a second path grammar), and every set-shaped
// consumer (a CLI's set flag, CI-injected image references) parses and
// applies through here.

// NodeOverride is one parsed node-addressed override.
type NodeOverride struct {
	// Kind is the target document's kind, exactly as its manifest spells it.
	Kind string
	// Name is the target document's metadata.name.
	Name string
	// FieldPath is the dotted path to set on that document.
	FieldPath string
	// Value is the value to set.
	Value string
}

// ParseNodeOverride parses `<kind>/<name>:<fieldPath>` (the map key of a
// key=value flag) into its parts. The value rides separately — flag parsing
// already split on '='.
func ParseNodeOverride(key, value string) (NodeOverride, error) {
	slash := strings.IndexByte(key, '/')
	colon := strings.IndexByte(key, ':')
	if slash <= 0 || colon <= slash+1 || colon == len(key)-1 {
		return NodeOverride{}, errors.Errorf(
			"override %q is not node-addressed — a set deploy needs `<kind>/<name>:<fieldPath>=<value>` (a bare field path is ambiguous across documents)", key)
	}
	return NodeOverride{
		Kind:      key[:slash],
		Name:      key[slash+1 : colon],
		FieldPath: key[colon+1:],
		Value:     value,
	}, nil
}

// ApplyNodeOverride sets the override's field on the ONE document it names
// and returns the docs with that document re-serialized. A miss refuses
// naming the documents that do exist — nothing is guessed.
func ApplyNodeOverride(docs []Doc, override NodeOverride) ([]Doc, error) {
	var available []string
	for i, doc := range docs {
		msg, err := manifest.LoadManifestBytes(doc.Bytes, doc.Source)
		if err != nil {
			// A document that fails to load is the preflight wall's refusal
			// to make with full context; the override pass leaves it be.
			continue
		}
		kind, _ := catalogkindreflect.ExtractKindFromProto(msg)
		docName := metadatareflect.ExtractMetadata(msg).GetName()
		available = append(available, kind+"/"+docName)
		if kind != override.Kind || docName != override.Name {
			continue
		}

		updated, err := manifestprotobuf.SetProtoField(msg, override.FieldPath, override.Value)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to set %s=%s on %s/%s",
				override.FieldPath, override.Value, override.Kind, override.Name)
		}
		jsonBytes, err := protojson.Marshal(updated)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to serialize %s/%s after override", override.Kind, override.Name)
		}
		yamlBytes, err := protobufyaml.JSONToYAML(jsonBytes)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to render %s/%s after override", override.Kind, override.Name)
		}
		out := make([]Doc, len(docs))
		copy(out, docs)
		out[i] = Doc{Bytes: yamlBytes, Source: doc.Source}
		return out, nil
	}
	return nil, errors.Errorf("override targets %s/%s, which is not in this set — the set holds: %s",
		override.Kind, override.Name, strings.Join(available, ", "))
}
