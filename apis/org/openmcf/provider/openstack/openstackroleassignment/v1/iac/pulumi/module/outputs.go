package module

const (
	// OpId is the exported stack output containing the role assignment composite ID.
	OpId = "id"
	// OpRoleId is the exported stack output containing the role UUID.
	OpRoleId = "role_id"
	// OpProjectId is the exported stack output containing the project UUID (if project-scoped).
	OpProjectId = "project_id"
	// OpDomainId is the exported stack output containing the domain UUID (if domain-scoped).
	OpDomainId = "domain_id"
	// OpUserId is the exported stack output containing the user UUID (if user assignment).
	OpUserId = "user_id"
	// OpGroupId is the exported stack output containing the group UUID (if group assignment).
	OpGroupId = "group_id"
	// OpRegion is the exported stack output containing the OpenStack region.
	OpRegion = "region"
)
