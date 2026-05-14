# Auth0 Client - Compliance

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

## Client-Specific Compliance Notes

### Data Residency

Auth0 tenants are region-locked at creation. Client resources inherit the tenant's data residency region (US, EU, or AU). Authentication data processed by clients stays within the designated region.

### Audit Logging

All client CRUD operations are recorded in Auth0 tenant logs. These logs capture the Management API actor, timestamp, and operation details. Log retention depends on the plan tier (2 days free, 30 days enterprise).

### Token Data

Tokens issued by Auth0 clients may contain user PII (name, email) in ID token claims. Ensure downstream systems handling these tokens comply with applicable data protection regulations.

### Consent Management

For GDPR-regulated applications, Auth0 clients can be configured to display consent prompts during authentication. This is managed through the client's login flow configuration.
