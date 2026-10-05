# Auth0 Connection Guide

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

## Connection-Specific Security Notes

### Social Connection Secrets

Social connections require OAuth client_id and client_secret pairs from each identity provider (Google, GitHub, Facebook, etc.). These credentials are stored encrypted in Auth0's configuration store. Treat them as sensitive secrets -- leaking a social provider's client_secret allows impersonation of your application to that provider.

### Database Connections

Database connections store password hashes using bcrypt (10 salt rounds by default). Auth0 never stores plaintext passwords. Custom database connections that delegate to an external user store must use secure HTTPS endpoints for the connection scripts.

### Enterprise Connection Security

- **SAML connections**: Require X.509 certificate configuration for signature validation. SAML assertions are validated against the configured certificate.
- **OIDC connections**: Validate tokens using the provider's JWKS endpoint. Ensure the provider's discovery document is served over HTTPS.
- **Azure AD connections**: Use Microsoft's OAuth 2.0 endpoints with tenant-specific configuration. Multi-tenant configurations should restrict to specific Azure AD tenants.

### Brute Force Protection

Database connections support brute force protection, which blocks login attempts after repeated failures. This is configured at the connection level and applies to all clients using the connection.

## Permissions
## Management API Scopes

Auth0 connection resources require the following Management API scopes for CRUD operations. These scopes must be granted to the M2M application used for infrastructure automation.

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read | `read:connections` | List and retrieve connection configurations |
| Create | `create:connections` | Create new connections (Database, Social, Enterprise) |
| Update | `update:connections` | Modify connection settings, enabled clients, and options |
| Delete | `delete:connections` | Remove connections from the tenant |

## Additional Scopes

Depending on the connection type and operations performed, these supplementary scopes may be required:

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read users | `read:users` | List users associated with a connection |
| Delete users | `delete:users` | Remove users from a database connection |
| Read stats | `read:stats` | Retrieve connection usage statistics |

## Minimum Required Scopes

For basic lifecycle management (create, read, update, delete), the minimum required scopes are:

```
read:connections create:connections update:connections delete:connections
```

## Connection-Client Association

Connections are enabled for specific clients via the `enabled_clients` property on the connection object. Modifying this association requires `update:connections` scope. No additional client-specific scopes are needed for this operation.

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

## Connection-Specific Compliance Notes

### Password Storage

Database connections store credentials using bcrypt hashing. Auth0's password storage practices comply with NIST SP 800-63B guidelines for memorized secret verifiers. Password history and complexity policies are configurable per connection.

### Social Identity Data

Social connections import user profile data from external identity providers. This data is subject to the privacy policies of both Auth0 and the upstream provider. Ensure your application's privacy policy covers data received from social logins.

### Enterprise Federation

SAML and OIDC connections federate authentication to external identity providers. Auth0 acts as a relying party. Compliance responsibility for the upstream identity provider's security posture remains with the organization operating that provider.

### Audit Trail

All connection CRUD operations and authentication events are captured in Auth0 tenant logs. Login events include connection name, client, and result. Log retention depends on plan tier.

### Data Residency

Connection configurations are stored in the tenant's designated region (US, EU, or AU). User profile data from database connections is stored in the same region.

## Cost
## Pricing Model

Auth0 pricing is based on Monthly Active Users (MAUs), not on the number of resources created. Authentication connections are free API objects with no per-resource cost.

## Free Tier

The Auth0 Free plan includes:

- 25,000 MAUs
- 1 tenant
- Unlimited social connections
- Unlimited database connections

## Cost Impact

Creating, updating, or deleting Auth0 connection resources has no direct billing impact. There is no charge per connection regardless of type (Database, Social, Enterprise, SAML, OIDC, Azure AD).

The only cost driver is the number of monthly active users authenticating through your tenant.

## Enterprise Connection Considerations

While connections themselves are free objects, certain enterprise connection types (SAML, OIDC, Azure AD) are only available on paid plans:

| Connection Type | Minimum Plan |
|----------------|--------------|
| Database | Free |
| Social (Google, GitHub, etc.) | Free |
| SAML | Essentials |
| OIDC | Essentials |
| Azure AD | Essentials |
| LDAP/AD | Enterprise |

## Social Connection Limits

The free plan allows unlimited social connections, but each social provider requires its own OAuth app registration (e.g., Google Cloud Console, GitHub OAuth App). Auth0 provides default development keys for testing, but production deployments should use custom keys.
