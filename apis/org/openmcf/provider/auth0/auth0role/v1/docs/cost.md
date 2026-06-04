# Auth0 Role - Cost

## Pricing Model

Auth0 pricing is based on Monthly Active Users (MAUs), not on the number of resources created. Roles are free API objects with no per-resource cost.

## Free Tier

The Auth0 Free plan includes:

- 25,000 MAUs
- 1 tenant
- RBAC with unlimited roles and permissions

Note: some advanced RBAC and organizations features are gated to paid plans, but creating roles and assigning permissions to them is available broadly. Verify against your plan if you rely on advanced RBAC behavior.

## Cost Impact

Creating, updating, or deleting Auth0 Role resources (and their permissions) has no direct billing impact. There is no charge per role, per permission, or per assignment.

The only cost driver is the number of monthly active users authenticating through your tenant. Role evaluation during login adds no additional cost.

## Operational Considerations

| Factor | Impact |
|--------|--------|
| Number of roles | No cost; affects manageability, not billing |
| Permissions per role | No cost; very large permission sets can enlarge tokens |
| Token size | Many permissions embedded via an `_authz` dialect increase access-token size |

## Rate Limits

The Auth0 Management API enforces rate limits on role operations. Creating roles and setting permissions counts against the Management API rate limit. For bulk role provisioning, apply changes in batches to stay within limits.
