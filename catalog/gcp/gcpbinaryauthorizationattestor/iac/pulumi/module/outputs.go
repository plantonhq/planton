package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpAttestorId                    = "attestor_id"
	OpAttestorName                  = "attestor_name"
	OpNoteReference                 = "note_reference"
	OpDelegationServiceAccountEmail = "delegation_service_account_email"
)
