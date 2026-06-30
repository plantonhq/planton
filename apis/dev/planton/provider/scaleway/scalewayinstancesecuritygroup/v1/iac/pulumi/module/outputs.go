package module

const (
	// OpSecurityGroupId is the exported stack output name that contains
	// the identifier of the created Scaleway Instance Security Group.
	//
	// This is referenced by ScalewayInstance via StringValueOrRef on
	// the security_group_id field.
	OpSecurityGroupId = "security_group_id"
)
