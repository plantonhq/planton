# Auth0 Event Stream - Compliance

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

## Event Stream-Specific Compliance Notes

### PII in Event Data

Auth0 log events contain user PII including email addresses, IP addresses, and user agent strings. When streaming events to external systems, ensure the destination complies with applicable data protection regulations (GDPR, CCPA). Data Processing Agreements may be required with third-party log aggregation providers.

### Cross-Border Data Transfer

Event streams may deliver data to destinations outside your Auth0 tenant's region. If your tenant is in the EU region and events stream to a US-based service, this constitutes a cross-border data transfer subject to GDPR Chapter V requirements. Use Standard Contractual Clauses (SCCs) or other approved transfer mechanisms.

### Log Retention and Right to Erasure

Auth0's built-in log retention is time-limited. Event streams export logs to external systems where retention policies are independently managed. For GDPR right-to-erasure compliance, ensure your external log storage can identify and delete user-specific log entries upon request.

### Audit Trail

All event stream CRUD operations are recorded in Auth0 tenant logs. Stream delivery status (success/failure) is also logged. These operational logs provide evidence of log pipeline health for compliance audits.

### Data Minimization

Auth0 streams all tenant log event types by default. Event streams support filtering by event type. For compliance-sensitive environments, configure filters to exclude event types containing unnecessary PII.
