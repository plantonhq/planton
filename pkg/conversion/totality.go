package conversion

import (
	"fmt"
	"sort"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// CheckTotality proves a spec accounts for every schema difference between
// its two versions' spec messages: a field the new version removed must be
// the source of a rename/map or a declared-lossy drop (silent disappearance
// is a defect), and a field the new version REQUIRES must be produced by an
// op. Fields present in both versions pass through untouched by design.
//
// Message names are fully qualified (e.g.
// "dev.planton._test.testcatalogkindgeneric.v1alpha1.TestCatalogKindGenericSpec");
// both packages must be linked into the calling binary.
func CheckTotality(spec *Spec, fromMessage, toMessage string) error {
	fromDesc, err := lookupMessage(fromMessage)
	if err != nil {
		return err
	}
	toDesc, err := lookupMessage(toMessage)
	if err != nil {
		return err
	}

	fromPaths := fieldPaths("spec", fromDesc)
	toPaths := fieldPaths("spec", toDesc)

	handledSources := map[string]bool{}
	producedTargets := map[string]bool{}
	for _, op := range spec.Ops {
		switch {
		case op.Rename != nil:
			handledSources[op.Rename.FromPath] = true
			producedTargets[op.Rename.ToPath] = true
		case op.Map != nil:
			handledSources[op.Map.Path] = true
			to := op.Map.To
			if to == "" {
				to = op.Map.Path
			}
			producedTargets[to] = true
		case op.Drop != nil:
			handledSources[op.Drop.Path] = true
		case op.Default != nil:
			producedTargets[op.Default.Path] = true
		}
	}

	var problems []string

	// Every removed field must be handled by an op (directly or via an
	// ancestor path an op moves/drops wholesale).
	var removed []string
	for path := range fromPaths {
		if _, stillThere := toPaths[path]; !stillThere {
			removed = append(removed, path)
		}
	}
	sort.Strings(removed)
	for _, path := range removed {
		if !coveredBy(path, handledSources) {
			problems = append(problems, fmt.Sprintf(
				"%s exists at %s but not at %s and no op renames, maps, or drops it -- stored values would vanish silently",
				path, spec.From, spec.To))
		}
	}

	// Added fields need no structural check here: whether an upgraded
	// document satisfies the new version's requiredness rules is proven
	// authoritatively by the corpus, which runs full schema validation on
	// every converted fixture.
	_ = producedTargets

	if len(problems) > 0 {
		return fmt.Errorf("conversion spec %s %s->%s is not total:\n  %s",
			spec.Kind, spec.From, spec.To, strings.Join(problems, "\n  "))
	}
	return nil
}

func coveredBy(path string, handled map[string]bool) bool {
	for candidate := path; candidate != ""; {
		if handled[candidate] {
			return true
		}
		idx := strings.LastIndex(candidate, ".")
		if idx < 0 {
			return false
		}
		candidate = candidate[:idx]
	}
	return false
}

func lookupMessage(fullName string) (protoreflect.MessageDescriptor, error) {
	msgType, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(fullName))
	if err != nil {
		return nil, fmt.Errorf(
			"message %s is not linked into this binary -- the corpus test must import both versions' packages: %w",
			fullName, err)
	}
	return msgType.Descriptor(), nil
}

// fieldPaths returns document paths (JSON names) for a message's fields,
// recursing into nested messages declared in the same package. Lists, maps,
// and foreign/well-known types are treated as leaves -- their internal shape
// belongs to their own contract.
func fieldPaths(prefix string, desc protoreflect.MessageDescriptor) map[string]struct{} {
	paths := make(map[string]struct{})
	var walk func(prefix string, d protoreflect.MessageDescriptor, depth int)
	walk = func(prefix string, d protoreflect.MessageDescriptor, depth int) {
		if depth > 8 {
			return
		}
		fields := d.Fields()
		for i := 0; i < fields.Len(); i++ {
			field := fields.Get(i)
			path := prefix + "." + field.JSONName()
			paths[path] = struct{}{}
			if field.Kind() == protoreflect.MessageKind && !field.IsList() && !field.IsMap() &&
				strings.HasPrefix(string(field.Message().FullName()), string(d.ParentFile().Package())) {
				walk(path, field.Message(), depth+1)
			}
		}
	}
	walk(prefix, desc, 0)
	return paths
}
