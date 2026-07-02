package runner

import (
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/outputs"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"sigs.k8s.io/yaml"
)

const stringValueOrRefFullName = "dev.planton.shared.foreignkey.v1.StringValueOrRef"

// ResolveManifestRefs implements, for the standalone E2E harness, the foreign-key
// resolution the Planton orchestrator performs in production: it replaces each
// value_from reference in the component manifest whose default_kind matches a
// deployed prerequisite with the literal value read from that prerequisite's
// outputs (via the field's default_kind_field_path). Standalone Planton otherwise
// requires literal values -- the tofu generator errors on an unresolved ref and
// the pulumi modules drop it -- so this is the step that makes a composed
// (e.g. subnet -> vpc) topology testable end to end.
//
// The resolved manifest is written to a temp file whose path is returned; the
// original is left untouched. When there is nothing to resolve, the original path
// is returned unchanged.
//
// Scope: singular and repeated StringValueOrRef fields directly on the spec
// (e.g. a subnet's vpc_id, a role's managed_policy_arns). Each element of a
// repeated field resolves independently, so a list can mix literals (say, an
// AWS-managed policy ARN) with references to deployed prerequisites.
func ResolveManifestRefs(manifestPath string, depOutputs map[cloudresourcekind.CloudResourceKind]map[string]interface{}) (string, error) {
	if len(depOutputs) == 0 {
		return manifestPath, nil
	}

	manifestObject, err := manifest.LoadManifest(manifestPath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to load manifest for ref resolution from %s", manifestPath)
	}

	// Flatten each prerequisite's outputs to dotted keys so a default_kind_field_path
	// like "status.outputs.vpc_id" resolves the way the platform flattens outputs.
	flattened := make(map[cloudresourcekind.CloudResourceKind]map[string]string, len(depOutputs))
	for kind, out := range depOutputs {
		flattened[kind] = outputs.Flatten(out)
	}

	top := manifestObject.ProtoReflect()
	specFd := top.Descriptor().Fields().ByName("spec")
	if specFd == nil || specFd.Kind() != protoreflect.MessageKind {
		return manifestPath, nil
	}

	resolvedAny, err := resolveRefsInMessage(top.Mutable(specFd).Message(), flattened)
	if err != nil {
		return "", err
	}
	if !resolvedAny {
		return manifestPath, nil
	}

	jsonBytes, err := protojson.Marshal(manifestObject)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal resolved manifest")
	}
	yamlBytes, err := yaml.JSONToYAML(jsonBytes)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert resolved manifest to yaml")
	}

	// A temp file (not next to the scenario) so scenario discovery never picks it up.
	tmpFile, err := os.CreateTemp("", "planton-e2e-resolved-*.yaml")
	if err != nil {
		return "", errors.Wrap(err, "failed to create temp file for resolved manifest")
	}
	if _, err := tmpFile.Write(yamlBytes); err != nil {
		tmpFile.Close()
		return "", errors.Wrap(err, "failed to write resolved manifest")
	}
	if err := tmpFile.Close(); err != nil {
		return "", errors.Wrap(err, "failed to close resolved manifest")
	}
	return tmpFile.Name(), nil
}

// resolveRefsInMessage replaces value_from arms on the message's singular and
// repeated StringValueOrRef fields with literals from the matching
// prerequisite's outputs. Returns whether any field was resolved.
func resolveRefsInMessage(msg protoreflect.Message, flattened map[cloudresourcekind.CloudResourceKind]map[string]string) (bool, error) {
	resolvedAny := false
	fields := msg.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		if fd.Kind() != protoreflect.MessageKind || fd.IsMap() {
			continue
		}
		if string(fd.Message().FullName()) != stringValueOrRefFullName {
			continue
		}

		if fd.IsList() {
			// Repeated refs (e.g. a role's managed_policy_arns): each element
			// resolves independently, so literals and references can mix in one
			// list.
			if !msg.Has(fd) {
				continue
			}
			list := msg.Mutable(fd).List()
			for j := 0; j < list.Len(); j++ {
				ref, ok := list.Get(j).Message().Interface().(*foreignkeyv1.StringValueOrRef)
				if !ok || ref.GetValueFrom() == nil {
					continue
				}
				val, resolved, err := lookupRefValue(fd, flattened)
				if err != nil {
					return false, err
				}
				if !resolved {
					break // no prerequisite of this kind was deployed; leave the list untouched
				}
				list.Set(j, protoreflect.ValueOfMessage((&foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
				}).ProtoReflect()))
				resolvedAny = true
			}
			continue
		}

		if !msg.Has(fd) {
			continue
		}
		ref, ok := msg.Get(fd).Message().Interface().(*foreignkeyv1.StringValueOrRef)
		if !ok || ref.GetValueFrom() == nil {
			continue
		}
		val, resolved, err := lookupRefValue(fd, flattened)
		if err != nil {
			return false, err
		}
		if !resolved {
			continue
		}
		msg.Set(fd, protoreflect.ValueOfMessage((&foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
		}).ProtoReflect()))
		resolvedAny = true
	}
	return resolvedAny, nil
}

// lookupRefValue resolves a field's default_kind annotation against the deployed
// prerequisites' flattened outputs. Returns (value, true, nil) on success and
// (_, false, nil) when the field has no default_kind or no prerequisite of that
// kind was deployed -- in which case the ref is left untouched. A deployed
// prerequisite that is missing the annotated output is an error, so a
// misdeclared field path fails loudly rather than silently skipping.
func lookupRefValue(fd protoreflect.FieldDescriptor, flattened map[cloudresourcekind.CloudResourceKind]map[string]string) (string, bool, error) {
	opts := fd.Options()
	if opts == nil {
		return "", false, nil
	}
	kind, _ := proto.GetExtension(opts, foreignkeyv1.E_DefaultKind).(cloudresourcekind.CloudResourceKind)
	if kind == cloudresourcekind.CloudResourceKind_unspecified {
		return "", false, nil
	}
	outs, ok := flattened[kind]
	if !ok {
		return "", false, nil
	}
	path, _ := proto.GetExtension(opts, foreignkeyv1.E_DefaultKindFieldPath).(string)
	key := strings.TrimPrefix(path, "status.outputs.")
	val, ok := outs[key]
	if !ok {
		return "", false, errors.Errorf("prerequisite %s has no output %q to resolve field %q", kind, key, fd.Name())
	}
	return val, true, nil
}
