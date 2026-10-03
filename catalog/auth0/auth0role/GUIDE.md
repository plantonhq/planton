# Auth0 Role Guide

## Security
## Platform Security Posture

The certifications below are Auth0's own published claims about their hosted platform (verify current status on Auth0's compliance page). They describe the vendor's service — never this catalog kind, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Auth0's published certifications and security standards:

- SOC 2 Type II (annual audit)
- ISO 27001, ISO 27018 (privacy controls)
- HIPAA BAA available on enterprise plans
- PCI DSS Level 1 Service Provider
- FedRAMP Authorized (moderate baseline)
- CSA STAR Level 2
- GDPR compliant with Data Processing Agreement

## Data Protection

- **Data residency**: US, EU, AU regions
- **Encryption in transit**: TLS 1.2+
- **Encryption at rest**: AES-256
- **Penetration testing**: Annual third-party assessments

## Role-Specific Security Notes

### Least Privilege

Roles are the primary mechanism for enforcing least privilege in Auth0 RBAC. Grant each role only the scopes its users genuinely need. Avoid broad "superuser" roles that aggregate every scope; prefer several focused roles that can be combined per user.

### Authoritative Permission Management

This component manages a role's permission set authoritatively. A permission removed from the spec is removed from the role on the next apply. This is a security strength: the manifest is the single source of truth, so out-of-band privilege escalation (a scope added directly in the dashboard) is reconciled away on the next deployment. Review changes to the `permissions` list with the same rigor as any access-control change.

### Permissions Are References, Not Grants of New Capability

A role permission references a scope already defined on a resource server. Adding a permission to a role does not create new capability in the API — it grants users with that role the existing scope. The actual enforcement happens at the API, which must validate the scope claim in the access token.

### Token Exposure

When a resource server enables RBAC with an `_authz` token dialect, a user's role permissions are embedded as scopes in their access token. Access tokens are JWTs and their claims are readable by anyone who holds the token. Grant only the scopes downstream APIs require; do not use roles to carry sensitive metadata.

### Separation of Duties

Defining roles (this component) is separate from assigning roles to users (a runtime identity operation). This separation supports least-privilege review: infrastructure reviewers govern what a role can do, while identity/admin processes govern who holds it.

## Permissions
## Management API Scopes

Auth0 Role resources require the following Management API scopes for CRUD operations. These scopes must be granted to the M2M application used for infrastructure automation.

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read | `read:roles` | List and retrieve roles and their permissions |
| Create | `create:roles` | Create new roles |
| Update | `update:roles` | Modify role name, description, and permission assignments |
| Delete | `delete:roles` | Remove roles from the tenant |

## Permission Assignment Scopes

Setting a role's permissions reads the resource servers that own the referenced scopes. The following scope is required in addition to the role scopes above:

| Operation | Required Scope | Description |
|-----------|---------------|-------------|
| Resolve scopes for assignment | `read:resource_servers` | Read resource servers to validate the scopes assigned to the role |

Managing the permission assignments themselves (add/remove permissions on a role) is covered by `update:roles`.

## Minimum Required Scopes

For basic lifecycle management (create, read, update, delete) including permission assignment, the minimum required scopes are:

```
read:roles create:roles update:roles delete:roles read:resource_servers
```

## Prerequisite Scopes Must Exist

The scopes referenced by a role's permissions must already be defined on their resource servers before they can be assigned. This kind does not create scopes — use the `Auth0ResourceServer` component (or define scopes directly in Auth0) first. The M2M application does not need write access to resource servers to assign existing scopes to a role; `read:resource_servers` is sufficient.

## Compliance
## Regulatory Frameworks

The table below records Auth0's own published compliance posture for their hosted platform, as Auth0 states it (verify current status in Auth0's trust documentation). These are vendor facts about the service this component configures — not properties of the component or of your deployment, and nothing here transfers to your application without your own assessment.

Auth0's published framework posture:

| Framework | Status | Notes |
|-----------|--------|-------|
| SOC 2 Type II | Certified | Annual audit cycle |
| ISO 27001:2022 | Certified | Information security management |
| ISO 27018:2019 | Certified | PII protection in public clouds |
| HIPAA | BAA Available | Enterprise plans only |
| PCI DSS Level 1 | Certified | Service provider certification |
| FedRAMP Moderate | Authorized | US government workloads |
| CSA STAR Level 2 | Certified | Cloud security assurance |
| GDPR | Compliant | Data Processing Agreement available |
| CCPA | Compliant | California consumer privacy |

## Role-Specific Compliance Notes

### Access Control as Auditable Configuration

Roles and their permission sets are access-control configuration. Managing them as version-controlled manifests in the planton repository gives a complete change history (who changed which role's permissions, and when) that satisfies change-management and access-review requirements common to SOC 2 and ISO 27001.

### Least Privilege and Access Reviews

Periodic access reviews are a recurring control in most frameworks. Because each role's permissions are declared explicitly in its manifest, reviewers can audit exactly what every role grants without querying the live tenant. Keep roles narrow to make reviews tractable.

### Authoritative Reconciliation

The authoritative permission model means the deployed state matches the reviewed manifest after each apply. Unapproved changes made directly in the Auth0 dashboard are corrected on the next deployment, supporting the integrity of the documented access-control baseline.

### Audit Trail

All role CRUD operations (create, update, delete) and permission changes are recorded in Auth0 tenant logs. Log retention depends on plan tier (2 days free, up to 30 days enterprise). For long-term retention, stream tenant logs to an external SIEM (see the `Auth0EventStream` component).

### Separation of Definition and Assignment

This kind defines roles and their permissions but does not assign roles to users. User-to-role assignment is governed separately, supporting separation-of-duties controls between infrastructure and identity administration.

## Cost
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
