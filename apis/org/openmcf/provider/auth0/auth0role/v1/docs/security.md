# Auth0 Role - Security

## Platform Security Posture

Auth0 maintains the following certifications and security standards:

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
