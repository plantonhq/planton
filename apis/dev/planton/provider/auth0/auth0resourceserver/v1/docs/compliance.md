# Auth0 Resource Server - Compliance

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

## Resource Server-Specific Compliance Notes

### Access Token Content

Access tokens issued for resource servers may contain user identifiers and granted scopes. If your API's access tokens include custom claims with PII (via Actions or Rules), ensure token handling complies with GDPR data minimization principles. Only include claims that the API needs.

### Scope-Based Access Control

Resource server scopes provide the foundation for authorization decisions. For compliance-sensitive workloads, document the mapping between scopes and business-level access rights. Auditors may request evidence that scope assignments follow least-privilege principles.

### Audit Trail

All resource server CRUD operations are logged in Auth0 tenant logs. Token issuance events for each resource server are also logged, providing an audit trail of API access. Log retention depends on plan tier (2 days free, 30 days enterprise).

### API Identifier Stability

The resource server identifier (audience) is immutable after creation. Changing an API's audience requires creating a new resource server. This is relevant for compliance documentation that references specific API identifiers.
