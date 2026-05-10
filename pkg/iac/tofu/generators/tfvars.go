package generators

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/pkg/fileutil"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ProtoToTFVars converts a protobuf message into a Terraform tfvars-compatible
// string. The conversion applies OpenMCF type rules to flatten wrapper types
// (like StringValueOrRef) to primitives and omit orchestrator-only fields
// (like KubernetesClusterSelector).
//
// Pipeline: protojson.Marshal -> JSON map -> Flatten (type rules) -> HCL string.
func ProtoToTFVars(msg proto.Message) (string, error) {
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

	Flatten(data, msg.ProtoReflect().Descriptor(), DefaultRules())

	var buf bytes.Buffer
	if err := WriteMapToHCL(&buf, data, 0); err != nil {
		return "", errors.Wrap(err, "failed to convert map to hcl")
	}

	return buf.String(), nil
}

// WriteVarFile converts a protobuf message to tfvars and writes it to a file.
// Creates parent directories if they do not exist.
func WriteVarFile(msg proto.Message, tfvarsFile string) error {
	tfvarsString, err := ProtoToTFVars(msg)
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
