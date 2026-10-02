package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	// OpPeeringName is the peering entry's name on this side's network.
	OpPeeringName = "peering_name"
	// OpNetwork is this side's network as a self link.
	OpNetwork = "network"
	// OpState is ACTIVE or INACTIVE as GCP reports it (create form only).
	OpState = "state"
	// OpStateDetails is GCP's explanation of the state (create form only).
	OpStateDetails = "state_details"
)
