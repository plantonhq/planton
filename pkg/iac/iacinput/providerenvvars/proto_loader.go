package providerenvvars

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/protobufyaml"
	"google.golang.org/protobuf/proto"
)

// loadProviderConfigProto loads provider config YAML bytes into a proto message.
func loadProviderConfigProto(providerConfigYaml []byte, protoMsg proto.Message) error {
	if err := protobufyaml.LoadYamlBytes(providerConfigYaml, protoMsg); err != nil {
		return errors.Wrap(err, "failed to load yaml bytes into provider config proto")
	}
	return nil
}
