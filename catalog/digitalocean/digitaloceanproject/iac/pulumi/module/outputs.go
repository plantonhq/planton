package module

const (
	// OpProjectId is the project UUID (the API identity, and the import id).
	OpProjectId = "project_id"
	// OpOwnerUuid is the UUID of the account or team that owns the project.
	OpOwnerUuid = "owner_uuid"
	// OpOwnerId is the numeric id of the owning account or team, as a string.
	OpOwnerId = "owner_id"
	// OpResourceUrns is the sorted list of member resource URNs DigitalOcean
	// reports for the project after apply.
	OpResourceUrns = "resource_urns"
)
