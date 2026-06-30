# Auth0 Action - Security

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

## Action-Specific Security Notes

### Sandboxed Runtime

Action code runs in Auth0's sandboxed Node.js runtime environment. Each Action execution is isolated -- Actions cannot access the filesystem, spawn processes, or communicate with other Actions except through the event/API objects provided by Auth0.

### Secrets Management

Actions support injecting secrets via environment variables configured in the Auth0 dashboard or Management API. These secrets are:

- Encrypted at rest in Auth0's configuration store
- Available to Action code via `event.secrets`
- Not logged or exposed in Action execution logs
- Scoped to a single Action (not shared across Actions)

Never hardcode credentials in Action source code. Always use the secrets mechanism.

### Code Execution Context

Actions in the post-login trigger receive the full user profile and authentication context. The Action code has access to:

- User profile data (email, name, metadata)
- Authentication method details (connection, MFA status)
- Client information (application making the request)
- The ability to modify tokens, deny access, or redirect users

### Supply Chain Security

Actions can include npm dependencies. Auth0 resolves and bundles these at deploy time. Pin dependency versions to avoid unintended updates. Review dependencies for known vulnerabilities before deploying to production.

### Action Versioning

Auth0 maintains a version history for each Action. Only the deployed version executes in the authentication pipeline. Draft versions can be tested before deployment.
