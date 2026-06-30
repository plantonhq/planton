# Auth0 Action - Compliance

## Regulatory Frameworks

Auth0 supports the following compliance frameworks:

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

## Action-Specific Compliance Notes

### Code as Configuration

Actions contain executable code that runs in Auth0's infrastructure. For regulated environments, treat Action code as auditable configuration. Maintain version control for all Action source code outside Auth0 (e.g., in the planton repository) to satisfy change management requirements.

### Data Access in Actions

Post-login Actions receive the authenticated user's profile data. If your Action logic processes PII (name, email, phone), ensure the Action's behavior complies with applicable data protection regulations. Avoid logging PII in Action console output.

### Custom Claims and Data Minimization

Actions that add custom claims to tokens should follow GDPR data minimization principles. Only include claims that the downstream API requires. Token claims are visible to any party that can decode the token (access tokens are JWTs).

### Audit Trail

All Action CRUD operations (create, update, deploy, delete) are recorded in Auth0 tenant logs. Action execution failures are also logged. Log retention depends on plan tier (2 days free, 30 days enterprise).

### Third-Party Dependencies

Actions that use npm packages introduce third-party code into the authentication pipeline. For compliance-sensitive environments, maintain an approved package list and review transitive dependencies.
