package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	// OpName is the key's full resource name
	// (projects/{project}/locations/global/keys/{key_id}).
	OpName = "name"
	// OpUid is the key's unique id -- what a Firebase app registration's
	// api_key_id references.
	OpUid = "uid"
	// OpKeyString is the key string the client presents -- exported as a
	// Pulumi secret so it never prints in previews or logs.
	OpKeyString = "key_string"
)
