package specprojection

import (
	"encoding/json"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// KeyStyle selects how the projected map spells proto field keys.
type KeyStyle int

const (
	// SnakeCaseKeys renames every proto field key to its proto field name,
	// matching the snake_case variables.tf of provider-abstraction modules.
	SnakeCaseKeys KeyStyle = iota
	// JSONKeys keeps protojson's names (lowerCamelCase or an explicit
	// json_name), which for a Kubernetes-manifest-projection kind are the
	// custom resource's own keys.
	JSONKeys
)

// Project renders a manifest message as the plain value map an IaC engine
// consumes: protojson with unpopulated fields omitted, then Flatten with the
// default type rules. Map keys a user authored (labels, environment variable
// names) are never renamed; only proto field keys follow the style.
func Project(msg proto.Message, style KeyStyle) (map[string]interface{}, error) {
	jsonBytes, err := protojson.MarshalOptions{EmitUnpopulated: false}.Marshal(msg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal proto to json")
	}

	var data map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal json")
	}
	if data == nil {
		data = map[string]interface{}{}
	}

	flattenWithOpts(data, msg.ProtoReflect().Descriptor(), DefaultRules(), flattenOpts{preserveJSONNames: style == JSONKeys})
	return data, nil
}
