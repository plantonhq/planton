# Auth0 Connection - Compliance

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
