package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpName                 = "name"
	OpReservationGroupName = "reservation_group_name"
	OpLocation             = "location"
)
