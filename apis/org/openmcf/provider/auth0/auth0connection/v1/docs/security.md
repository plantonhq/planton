# Auth0 Connection - Security

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
