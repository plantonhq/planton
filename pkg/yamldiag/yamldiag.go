// Package yamldiag diagnoses why a YAML manifest does not fit its kind's
// proto schema -- with real YAML line numbers, field paths, expected shapes,
// and fix suggestions.
//
// It exists because the authoritative parse destroys exactly the information
// an author needs: manifests are flattened to single-line JSON before
// protojson unmarshals them, so parser errors report offsets into an
// internal translation ("line 1:553") of a file the author has never seen.
// This package re-reads the ORIGINAL bytes into a yaml.v3 node tree (which
// preserves source positions) and walks it against the kind's message
// descriptor, reporting every structural mismatch in one pass.
//
// It is a sad-path diagnoser, never a parser: protojson remains the single
// authority on what loads. The walk only reports error classes that are
// structurally certain (unknown fields, composite-vs-scalar shape errors,
// quoting errors on string fields, unknown enum names, duplicate oneof
// members). Anything it does not fully model is skipped in the safe
// direction: alias values and merge-key expansions are not walked (their
// content is checked at the anchor's definition site), and scalar coercions
// protojson owns are never second-guessed. A wrong diagnosis is worse than
// a missing one; when in doubt, say nothing about that node.
//
// Error messages reuse the explain package's exported schema vocabulary
// (TypeLabel, ValueFromContract) so the reference surface and the error
// surface teach the same language.
package yamldiag

import (
	"fmt"
	"strings"

	"github.com/plantonhq/planton/pkg/explain"
	"github.com/plantonhq/planton/pkg/refannotations"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/yaml.v3"
)

const stringValueOrRefFullName = "dev.planton.shared.foreignkey.v1.StringValueOrRef"

// Mismatch is one located authoring error.
type Mismatch struct {
	// Path is the dotted field path in canonical YAML (protojson) spelling --
	// the same spelling `planton explain <kind>.<path>` drills into.
	Path string
	// Line is the 1-based line in the diagnosed YAML source.
	Line int
	// Problem states what is wrong and what the schema expects instead.
	Problem string
	// Suggestion optionally names the likely fix ("did you mean ...").
	Suggestion string
	// See overrides the schema-reference pointer path when the offense path
	// itself is not explainable -- an unknown field points at its PARENT,
	// which is where the valid fields are listed. Empty means Path.
	See string
}

// Format renders the mismatch for terminals and error text. kindName scopes
// the explain pointer; pass the manifest's kind.
func (m Mismatch) Format(kindName string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s (line %d): %s", m.Path, m.Line, m.Problem)
	if m.Suggestion != "" {
		fmt.Fprintf(&b, "\n  did you mean: %s", m.Suggestion)
	}
	if kindName != "" {
		see := m.See
		if see == "" {
			see = m.Path
		}
		if see != "" {
			fmt.Fprintf(&b, "\n  see: planton explain %s.%s", kindName, see)
		} else {
			fmt.Fprintf(&b, "\n  see: planton explain %s", kindName)
		}
	}
	return b.String()
}

// FormatAll renders every mismatch as one plain-text block.
func FormatAll(mismatches []Mismatch, kindName string) string {
	parts := make([]string, 0, len(mismatches))
	for _, m := range mismatches {
		parts = append(parts, m.Format(kindName))
	}
	return strings.Join(parts, "\n")
}

// Diagnose walks manifest YAML against the kind's envelope descriptor and
// returns every structural mismatch it is certain about, in source order.
// Empty means "no diagnosis" -- the manifest fits structurally and the
// failure is something protojson owns. Callers must keep the original
// parser error as the floor.
func Diagnose(manifestYamlBytes []byte, md protoreflect.MessageDescriptor) []Mismatch {
	var root yaml.Node
	// Decode errors are pure YAML syntax problems; the YAML layer's own
	// errors already carry real line numbers, so there is nothing to add.
	if err := yaml.Unmarshal(manifestYamlBytes, &root); err != nil {
		return nil
	}
	// yaml.Unmarshal yields a DocumentNode wrapping the first document --
	// mirroring the loader's own first-document semantics.
	if root.Kind != yaml.DocumentNode || len(root.Content) != 1 {
		return nil
	}
	body := root.Content[0]
	if body.Kind != yaml.MappingNode {
		return nil
	}

	w := &walker{}
	w.walkMessage(body, md, "")
	return w.mismatches
}

// walker accumulates mismatches over one document walk. Constructs the walk
// does not model (alias values, merge-key expansions) are skipped LOCALLY:
// a diagnosis elsewhere in the document is provable from what is physically
// present, so under-reporting inside the skipped subtree is the only cost.
type walker struct {
	mismatches []Mismatch
}

func (w *walker) report(path string, line int, problem, suggestion string) {
	w.mismatches = append(w.mismatches, Mismatch{Path: path, Line: line, Problem: problem, Suggestion: suggestion})
}

// reportUnknown records an unknown-field mismatch whose reference pointer
// targets the parent path -- the field itself cannot be explained, its
// legitimate siblings can.
func (w *walker) reportUnknown(path, parent string, line int, problem, suggestion string) {
	w.mismatches = append(w.mismatches, Mismatch{Path: path, Line: line, Problem: problem, Suggestion: suggestion, See: parent})
}

// walkMessage checks a mapping node against a message descriptor.
func (w *walker) walkMessage(node *yaml.Node, md protoreflect.MessageDescriptor, path string) {
	// A merge key (<<) injects keys the walk cannot see. Field-existence
	// checks on the EXPLICIT keys stay provable (a merge can never make an
	// unknown field known), but oneof-duplication is not -- the merged
	// content could carry the other member -- so that check disarms here.
	mergePresent := false
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Tag == "!!merge" {
			mergePresent = true
		}
	}

	// Track oneof membership: setting two members of the same exclusive
	// group is always an error, and protojson reports it without positions.
	seenOneofs := map[protoreflect.FullName]string{}

	for i := 0; i+1 < len(node.Content); i += 2 {
		keyNode, valueNode := node.Content[i], node.Content[i+1]
		if keyNode.Kind != yaml.ScalarNode || keyNode.Tag == "!!merge" {
			continue
		}
		key := keyNode.Value

		fd := fieldByEitherName(md, key)
		if fd == nil {
			w.reportUnknown(joinPath(path, key), path, keyNode.Line,
				fmt.Sprintf("unknown field %q on %s", key, md.Name()),
				suggestField(key, md))
			continue
		}
		fieldPath := joinPath(path, fd.JSONName())

		if oneof := fd.ContainingOneof(); oneof != nil && !oneof.IsSynthetic() && !mergePresent {
			if prior, ok := seenOneofs[oneof.FullName()]; ok {
				w.report(fieldPath, keyNode.Line,
					fmt.Sprintf("%q and %q are alternatives (oneof %s) -- set exactly one", prior, fd.JSONName(), oneof.Name()),
					"")
				continue
			}
			seenOneofs[oneof.FullName()] = fd.JSONName()
		}

		w.walkValue(valueNode, fd, fieldPath)
	}
}

// walkValue checks one field's value node against its descriptor.
func (w *walker) walkValue(node *yaml.Node, fd protoreflect.FieldDescriptor, path string) {
	// An alias value's content lives at its anchor definition, where it is
	// (or will be) checked in place; re-walking it here would double-report
	// the same defect under two paths. Skip locally -- under-reporting
	// inside the alias is the safe direction.
	if node.Kind == yaml.AliasNode {
		return
	}
	// Null loads as the field's default everywhere; nothing to check.
	if node.Kind == yaml.ScalarNode && node.Tag == "!!null" {
		return
	}

	switch {
	case fd.IsMap():
		if node.Kind != yaml.MappingNode {
			w.report(path, node.Line, fmt.Sprintf("expects a map of %s values; got %s",
				explain.TypeLabel(fd.MapValue()), nodeLabel(node)), "")
			return
		}
		for i := 0; i+1 < len(node.Content); i += 2 {
			w.walkElement(node.Content[i+1], fd.MapValue(), joinPath(path, node.Content[i].Value))
		}
	case fd.IsList():
		if node.Kind != yaml.SequenceNode {
			w.report(path, node.Line, fmt.Sprintf("expects a list of %s; got %s",
				explain.TypeLabel(fd), nodeLabel(node)), "")
			return
		}
		for i, item := range node.Content {
			w.walkElement(item, fd, fmt.Sprintf("%s[%d]", path, i))
		}
	default:
		w.walkElement(node, fd, path)
	}
}

// walkElement checks a single (non-repeated) value against the field's
// element type: message fields recurse, enums check the value name, string
// fields check quoting. Everything else is protojson's territory.
func (w *walker) walkElement(node *yaml.Node, fd protoreflect.FieldDescriptor, path string) {
	if node.Kind == yaml.AliasNode {
		return
	}
	if node.Kind == yaml.ScalarNode && node.Tag == "!!null" {
		return
	}

	if md := fd.Message(); md != nil {
		w.walkMessageElement(node, fd, md, path)
		return
	}

	switch fd.Kind() {
	case protoreflect.EnumKind:
		w.checkEnum(node, fd, path)
	case protoreflect.StringKind:
		// An unquoted number/bool where the schema says string is the classic
		// YAML quoting trap ("1.29" vs 1.29). Only string fields are checked:
		// numeric fields legally accept string forms, so flagging the inverse
		// would contradict the parser.
		if node.Kind == yaml.ScalarNode &&
			(node.Tag == "!!int" || node.Tag == "!!float" || node.Tag == "!!bool") {
			w.report(path, node.Line,
				fmt.Sprintf("expects a string; got an unquoted %s (%s) -- quote it: \"%s\"",
					strings.TrimPrefix(node.Tag, "!!"), node.Value, node.Value), "")
		} else if node.Kind != yaml.ScalarNode {
			w.report(path, node.Line, fmt.Sprintf("expects a string; got %s", nodeLabel(node)), "")
		}
	default:
		// Numeric/bool/bytes scalars: protojson owns their coercions; only a
		// composite where a scalar belongs is structurally certain.
		if node.Kind != yaml.ScalarNode {
			w.report(path, node.Line, fmt.Sprintf("expects a %s; got %s", fd.Kind(), nodeLabel(node)), "")
		}
	}
}

// walkMessageElement checks a value against a message-typed field.
func (w *walker) walkMessageElement(node *yaml.Node, fd protoreflect.FieldDescriptor, md protoreflect.MessageDescriptor, path string) {
	full := string(md.FullName())
	switch {
	// Untyped payloads accept anything.
	case full == "google.protobuf.Struct" || full == "google.protobuf.Value" || full == "google.protobuf.ListValue":
		return
	// Scalar-shaped well-known types (wrappers, time): protojson accepts
	// scalars; a composite is certainly wrong, the scalar formats are its
	// business.
	case strings.HasPrefix(full, "google.protobuf."):
		if node.Kind != yaml.ScalarNode {
			w.report(path, node.Line, fmt.Sprintf("expects a %s value; got %s",
				explain.TypeLabel(fd), nodeLabel(node)), "")
		}
		return
	}

	if node.Kind != yaml.MappingNode {
		// THE classic: a bare string where the foreign-key wrapper expects
		// {value: ...} or {valueFrom: ...}. Teach the exact contract, using
		// the field's declared reference target.
		if full == stringValueOrRefFullName {
			w.report(path, node.Line,
				fmt.Sprintf("got a bare %s -- %s", nodeLabel(node), explain.ValueFromContract(refTarget(fd))),
				"")
			return
		}
		w.report(path, node.Line, fmt.Sprintf("expects an object (%s); got %s",
			explain.TypeLabel(fd), nodeLabel(node)), "")
		return
	}
	w.walkMessage(node, md, path)
}

// checkEnum verifies a scalar names a real enum value. Numbers are legal
// protojson enum forms and pass through.
func (w *walker) checkEnum(node *yaml.Node, fd protoreflect.FieldDescriptor, path string) {
	if node.Kind != yaml.ScalarNode {
		w.report(path, node.Line, fmt.Sprintf("expects one of the %s values; got %s",
			fd.Enum().Name(), nodeLabel(node)), "")
		return
	}
	if node.Tag == "!!int" {
		return
	}
	values := fd.Enum().Values()
	names := make([]string, 0, values.Len())
	for i := 0; i < values.Len(); i++ {
		name := string(values.Get(i).Name())
		if name == node.Value {
			return
		}
		names = append(names, name)
	}
	w.report(path, node.Line,
		fmt.Sprintf("%q is not a value of enum %s (values: %s)",
			node.Value, fd.Enum().Name(), strings.Join(names, ", ")),
		nearest(node.Value, names))
}

// refTarget reads the foreign-key field's declared reference target -- its
// default, else its first candidate -- so the taught contract is concrete,
// not placeholder-shaped.
func refTarget(fd protoreflect.FieldDescriptor) (refKind, refFieldPath string) {
	annotations := refannotations.Of(fd)
	if keys := annotations.Keys(); len(keys) > 0 {
		return keys[0].Kind.String(), keys[0].FieldPath
	}
	if annotations.DefaultKind != cloudresourcekind.CloudResourceKind_unspecified {
		return annotations.DefaultKind.String(), ""
	}
	return "", ""
}

func fieldByEitherName(md protoreflect.MessageDescriptor, name string) protoreflect.FieldDescriptor {
	fields := md.Fields()
	for i := 0; i < fields.Len(); i++ {
		if fields.Get(i).JSONName() == name {
			return fields.Get(i)
		}
	}
	return fields.ByName(protoreflect.Name(name))
}

func nodeLabel(node *yaml.Node) string {
	switch node.Kind {
	case yaml.MappingNode:
		return "an object"
	case yaml.SequenceNode:
		return "a list"
	case yaml.ScalarNode:
		return fmt.Sprintf("%s %q", strings.TrimPrefix(node.Tag, "!!"), node.Value)
	default:
		return "an unsupported YAML construct"
	}
}

func joinPath(path, segment string) string {
	if path == "" {
		return segment
	}
	return path + "." + segment
}
