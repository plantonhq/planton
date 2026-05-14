# Auth0 Resource Server - Security

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

## Resource Server-Specific Security Notes

### Signing Algorithms

Resource servers support two token signing algorithms with different security characteristics:

- **RS256 (recommended)**: Asymmetric signing using RSA public/private key pairs. The API validates tokens using Auth0's public JWKS endpoint. No shared secret required. Supports key rotation without coordination.
- **HS256**: Symmetric signing using a shared secret. The API must store the signing secret to validate tokens. Key rotation requires coordinated updates between Auth0 and all consuming APIs.

Use RS256 unless you have a specific requirement for HS256. The signing secret for HS256 resource servers is sensitive and must be treated as a credential.

### Scope Design

Scopes define the permissions available for an API. Follow least-privilege principles:

- Define granular scopes (e.g., `read:orders`, `write:orders`) rather than coarse ones (`admin`).
- Scope names are arbitrary strings but conventionally use `action:resource` format.
- Scopes are requested at authentication time and granted based on client grants or user consent.

### Token Validation

APIs must validate access tokens on every request. Validation must include:

- Signature verification against Auth0's JWKS endpoint (RS256) or shared secret (HS256)
- Issuer (`iss`) claim matches your Auth0 domain
- Audience (`aud`) claim matches the resource server identifier
- Expiration (`exp`) claim is in the future
