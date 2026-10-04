# Auth0 Client Guide

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

## Client-Specific Security Notes

### Client Secrets

Client secrets are sensitive credentials that must be protected. How secrets apply depends on the application type:

- **SPA applications**: Use PKCE (Proof Key for Code Exchange) and do not have a client secret. This is the recommended approach for public clients.
- **Web applications**: Have a client secret that must be stored server-side. Never expose this in client-side code.
- **M2M applications**: Authenticate entirely via client_id and client_secret. These credentials grant full API access as configured.
- **Native applications**: Treated as public clients. Use PKCE instead of client secrets.

### Token Security

- Access tokens should use short expiration times (default: 86400 seconds).
- Refresh tokens should be configured with rotation and absolute lifetime limits.
- ID tokens contain user claims and should not be sent to APIs as access credentials.

### Callback URL Validation

Auth0 validates redirect URIs strictly. Only registered callback URLs are accepted during authentication flows. Wildcard subdomains are supported but should be used cautiously.

## Permissions
## Management API Scopes

Auth0 client resources require the following Management API scopes for CRUD operations. These scopes must be granted to the M2M application used for infrastructure automation.

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read | `read:clients` | List and retrieve application configurations |
| Create | `create:clients` | Create new applications (SPA, Web, M2M, Native) |
| Update | `update:clients` | Modify application settings, callbacks, and grants |
| Delete | `delete:clients` | Remove applications from the tenant |

## Additional Scopes

Depending on the client configuration, these supplementary scopes may also be required:

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read grants | `read:client_grants` | List client grant associations |
| Create grants | `create:client_grants` | Associate clients with APIs |
| Update grants | `update:client_grants` | Modify granted scopes for a client-API pair |
| Delete grants | `delete:client_grants` | Remove client grant associations |
| Read keys | `read:client_keys` | Retrieve client signing credentials |
| Update keys | `update:client_keys` | Rotate client signing credentials |

## Minimum Required Scopes

For basic lifecycle management (create, read, update, delete), the minimum required scopes are:

```
read:clients create:clients update:clients delete:clients
```

If the automation also manages which APIs a client can access, add:

```
read:client_grants create:client_grants update:client_grants delete:client_grants
```

## Compliance
## Regulatory Frameworks

The table below records Auth0's own published compliance posture for their hosted platform, as Auth0 states it (verify current status in Auth0's trust documentation). These are vendor facts about the service this catalog kind configures — not properties of the catalog kind or of your deployment, and nothing here transfers to your application without your own assessment.

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
| Privacy Shield | Historical | Deprecated framework, replaced by SCCs |

## Client-Specific Compliance Notes

### Data Residency

Auth0 tenants are region-locked at creation. Client resources inherit the tenant's data residency region (US, EU, or AU). Authentication data processed by clients stays within the designated region.

### Audit Logging

All client CRUD operations are recorded in Auth0 tenant logs. These logs capture the Management API actor, timestamp, and operation details. Log retention depends on the plan tier (2 days free, 30 days enterprise).

### Token Data

Tokens issued by Auth0 clients may contain user PII (name, email) in ID token claims. Ensure downstream systems handling these tokens comply with applicable data protection regulations.

### Consent Management

For GDPR-regulated applications, Auth0 clients can be configured to display consent prompts during authentication. This is managed through the client's login flow configuration.

## Cost
## Pricing Model

Auth0 pricing is based on Monthly Active Users (MAUs), not on the number of resources created. Auth0 Applications (clients) are free API objects with no per-resource cost.

## Free Tier

The Auth0 Free plan includes:

- 25,000 MAUs
- 1 tenant
- Unlimited social connections
- Unlimited applications (clients)

## Cost Impact

Creating, updating, or deleting Auth0 client resources has no direct billing impact. There is no charge per application regardless of type (SPA, Web, M2M, Native).

The only cost driver is the number of monthly active users authenticating through your tenant. If your MAU count stays within the free tier, all client resources remain free.

## M2M Token Considerations

Machine-to-machine (M2M) applications consume M2M token grants. The free plan includes 1,000 M2M tokens per month. Exceeding this limit requires a paid plan. Each M2M client credentials grant counts against this quota regardless of which M2M application issues the request.

## Paid Plan Thresholds

| Plan | MAU Limit | M2M Tokens/Month |
|------|-----------|-------------------|
| Free | 25,000 | 1,000 |
| Essentials | Custom | 5,000 |
| Professional | Custom | 10,000 |
| Enterprise | Custom | Negotiated |
