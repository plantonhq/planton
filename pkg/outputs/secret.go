//go:build !codegen
// +build !codegen

package outputs

import (
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/shared/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// SecretOutputs maps each top-level field of the kind's stack outputs to whether its schema marks
// it `sensitive` -- a secret the resource generates (a client secret, an access key, an admin
// password). The schema is the one rule for which outputs are secrets: the CLI masks them when it
// renders captured outputs, the platform stores them in the organization's secret store and keeps
// only a reference, and `planton module verify` holds both engines to exporting exactly these as
// secrets. Secret outputs are top-level by the catalog's shape rule, because both engines decide
// secrecy per top-level output.
func SecretOutputs(kind cloudresourcekind.CloudResourceKind) (map[string]bool, error) {
	message, err := resolveStackOutputsMessage(kind)
	if err != nil {
		return nil, err
	}
	fields := message.ProtoReflect().Descriptor().Fields()
	secret := make(map[string]bool, fields.Len())
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		secret[string(field.Name())] = isSensitive(field)
	}
	return secret, nil
}

func isSensitive(field protoreflect.FieldDescriptor) bool {
	opts, ok := field.Options().(proto.Message)
	if !ok || opts == nil || !proto.HasExtension(opts, options.E_Sensitive) {
		return false
	}
	return proto.GetExtension(opts, options.E_Sensitive).(bool)
}
