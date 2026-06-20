package runner

import (
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"github.com/plantonhq/openmcf/internal/manifest"
	"github.com/plantonhq/openmcf/pkg/outputs"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"sigs.k8s.io/yaml"
)

const stringValueOrRefFullName = "org.openmcf.shared.foreignkey.v1.StringValueOrRef"

// ResolveManifestRefs implements, for the standalone E2E harness, the foreign-key
// resolution the Planton orchestrator performs in production: it replaces each
// value_from reference in the component manifest whose default_kind matches a
// deployed prerequisite with the literal value read from that prerequisite's
// outputs (via the field's default_kind_field_path). Standalone OpenMCF otherwise
// requires literal values -- the tofu generator errors on an unresolved ref and
// the pulumi modules drop it -- so this is the step that makes a composed
// (e.g. subnet -> vpc) topology testable end to end.
//
// The resolved manifest is written to a temp file whose path is returned; the
// original is left untouched. When there is nothing to resolve, the original path
// is returned unchanged.
//
// Scope: singular StringValueOrRef fields directly on the spec (the single-value
// case, e.g. a subnet's vpc_id). Repeated ([*]) expansion is intentionally not
// handled yet -- it lands with the slice that first needs it.
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
	tmpFile, err := os.CreateTemp("", "openmcf-e2e-resolved-*.yaml")
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

// resolveRefsInMessage replaces value_from arms on the message's singular
// StringValueOrRef fields with literals from the matching prerequisite's outputs.
// Returns whether any field was resolved.
func resolveRefsInMessage(msg protoreflect.Message, flattened map[cloudresourcekind.CloudResourceKind]map[string]string) (bool, error) {
	resolvedAny := false
	fields := msg.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		if fd.Kind() != protoreflect.MessageKind || fd.IsList() || fd.IsMap() {
			continue
		}
		if string(fd.Message().FullName()) != stringValueOrRefFullName {
			continue
		}
		if !msg.Has(fd) {
			continue
		}
		ref, ok := msg.Get(fd).Message().Interface().(*foreignkeyv1.StringValueOrRef)
		if !ok || ref.GetValueFrom() == nil {
			continue
		}

		opts := fd.Options()
		if opts == nil {
			continue
		}
		kind, _ := proto.GetExtension(opts, foreignkeyv1.E_DefaultKind).(cloudresourcekind.CloudResourceKind)
		if kind == cloudresourcekind.CloudResourceKind_unspecified {
			continue
		}
		outs, ok := flattened[kind]
		if !ok {
			// No prerequisite of this kind was deployed; leave the ref untouched.
			continue
		}
		path, _ := proto.GetExtension(opts, foreignkeyv1.E_DefaultKindFieldPath).(string)
		key := strings.TrimPrefix(path, "status.outputs.")
		val, ok := outs[key]
		if !ok {
			return false, errors.Errorf("prerequisite %s has no output %q to resolve field %q", kind, key, fd.Name())
		}

		resolved := &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
		}
		msg.Set(fd, protoreflect.ValueOfMessage(resolved.ProtoReflect()))
		resolvedAny = true
	}
	return resolvedAny, nil
}
