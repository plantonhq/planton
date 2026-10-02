package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	// OpServiceProjectId is the project now attached as a service project.
	OpServiceProjectId = "service_project_id"
	// OpHostProjectId is the host it is attached to.
	OpHostProjectId = "host_project_id"
)
