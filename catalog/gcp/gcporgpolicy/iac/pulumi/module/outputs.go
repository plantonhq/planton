package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	// OpName is the policy's full resource name, {parent}/policies/{constraint}.
	OpName = "name"
	// OpEtag is the policy's opaque version marker.
	OpEtag = "etag"
)
