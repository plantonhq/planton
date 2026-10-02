package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	// OpName is the binding's resource name,
	// tagBindings/{encoded full resource name}/{tagValues/id}.
	OpName = "name"
	// OpParent is the full resource name the tag is bound to, as sent.
	OpParent = "parent"
	// OpTagValue is the bound value's resource name, tagValues/{id}.
	OpTagValue = "tag_value"
)
