package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	// OpHostProjectId is the project ID that is now the Shared VPC host --
	// the resolved value, populated even when the spec named no project.
	OpHostProjectId = "host_project_id"
)
