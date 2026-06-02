package generators

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/pkg/crkreflect"
	"github.com/plantonhq/openmcf/pkg/fileutil"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// RenderTFVars converts an OpenMCF manifest proto into a Terraform
// tfvars-compatible string, choosing the emission shape from the kind's
// metadata. This is the kind-aware entry point every runtime caller should use
// so the converter and the shipped module always agree on the wire format:
//
//   - Manifest-projection kinds (CloudResourceKindMeta.kubernetes_manifest_projection
//     set) -> ProtoToManifestTFVars: camelCase, pruned, fed verbatim to a
//     kubernetes_manifest passthrough module.
//   - All other (provider-abstraction) kinds -> ProtoToTFVars: snake_case typed
//     variables consumed by hand-written HCL.
//
// On any metadata-resolution failure it falls back to the snake_case path, which
// is the format the overwhelming majority of modules use.
func RenderTFVars(msg proto.Message) (string, error) {
	if isManifestProjectionMessage(msg) {
		return ProtoToManifestTFVars(msg)
	}
	return ProtoToTFVars(msg)
}

// isManifestProjectionMessage reports whether the manifest's kind is a thin
// projection of a single Kubernetes custom resource (so its tfvars must be the
// camelCase manifest spec). The projection flag lives on the kind enum's
// CloudResourceKindMeta, so we resolve kind -> meta via crkreflect rather than
// inspecting the message body.
func isManifestProjectionMessage(msg proto.Message) bool {
	kindStr, err := crkreflect.ExtractKindFromProto(msg)
	if err != nil {
		return false
	}
	meta, err := crkreflect.KindMeta(crkreflect.KindFromString(kindStr))
	if err != nil {
		return false
	}
	return meta.GetKubernetesManifestProjection() != nil
}

// ProtoToTFVars converts a protobuf message into a Terraform tfvars-compatible
// string. The conversion applies OpenMCF type rules to flatten wrapper types
// (like StringValueOrRef) to primitives and omit orchestrator-only fields
// (like KubernetesClusterSelector), and renames keys to snake_case to match the
// generated snake_case variables.tf.
//
// Pipeline: protojson.Marshal -> JSON map -> Flatten (type rules) -> HCL string.
func ProtoToTFVars(msg proto.Message) (string, error) {
	return protoToTFVars(msg, flattenOpts{})
}

// ProtoToManifestTFVars is the manifest-projection variant of ProtoToTFVars: it
// keeps the protojson camelCase keys (the CRD's own JSON keys) instead of
// renaming to snake_case, so the emitted `spec` can be handed verbatim to a
// kubernetes_manifest resource. Unset fields are already omitted by protojson,
// so the result carries no nulls -- which is exactly why the projection module
// needs no oneOf/required-subfield pruning in HCL. Wrapper flattening and
// orchestrator-field skipping still apply (they key on message type, not case).
func ProtoToManifestTFVars(msg proto.Message) (string, error) {
	return protoToTFVars(msg, flattenOpts{preserveJSONNames: true})
}

func protoToTFVars(msg proto.Message, opts flattenOpts) (string, error) {
	jsonBytes, err := protojson.MarshalOptions{
		EmitUnpopulated: false,
	}.Marshal(msg)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal proto to json")
	}

	var data map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal json")
	}

	flattenWithOpts(data, msg.ProtoReflect().Descriptor(), DefaultRules(), opts)

	var buf bytes.Buffer
	if err := WriteMapToHCL(&buf, data, 0); err != nil {
		return "", errors.Wrap(err, "failed to convert map to hcl")
	}

	return buf.String(), nil
}

// WriteVarFile renders a manifest proto to tfvars (kind-aware via RenderTFVars)
// and writes it to a file. Creates parent directories if they do not exist.
func WriteVarFile(msg proto.Message, tfvarsFile string) error {
	tfvarsString, err := RenderTFVars(msg)
	if err != nil {
		return errors.Wrap(err, "failed to convert manifest proto to tfvars")
	}

	if !fileutil.IsDirExists(filepath.Dir(tfvarsFile)) {
		if err := os.MkdirAll(filepath.Dir(tfvarsFile), 0755); err != nil {
			return errors.Wrapf(err, "failed to create directory %s", filepath.Dir(tfvarsFile))
		}
	}

	if err := os.WriteFile(tfvarsFile, []byte(tfvarsString), 0644); err != nil {
		return errors.Wrapf(err, "failed to write tfvars file %s", tfvarsFile)
	}
	return nil
}
