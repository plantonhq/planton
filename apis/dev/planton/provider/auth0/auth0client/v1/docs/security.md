# Auth0 Client - Security

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
