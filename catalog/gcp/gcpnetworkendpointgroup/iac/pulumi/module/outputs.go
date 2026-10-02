package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	// OpSelfLink is the group's self link: a backend service's
	// backends[].group.
	OpSelfLink = "self_link"
	// OpNegName is the resolved cloud-side name.
	OpNegName = "neg_name"
	// OpNegId is Google's numeric identifier (zonal groups; empty globally).
	OpNegId = "neg_id"
	// OpZone is the zone of a zonal group; empty for a global one.
	OpZone = "zone"
	// OpSize is the number of endpoints declared in the group.
	OpSize = "size"
)
