# Auth0 Role - Compliance

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

## Role-Specific Compliance Notes

### Access Control as Auditable Configuration

Roles and their permission sets are access-control configuration. Managing them as version-controlled manifests in the openmcf repository gives a complete change history (who changed which role's permissions, and when) that satisfies change-management and access-review requirements common to SOC 2 and ISO 27001.

### Least Privilege and Access Reviews

Periodic access reviews are a recurring control in most frameworks. Because each role's permissions are declared explicitly in its manifest, reviewers can audit exactly what every role grants without querying the live tenant. Keep roles narrow to make reviews tractable.

### Authoritative Reconciliation

The authoritative permission model means the deployed state matches the reviewed manifest after each apply. Unapproved changes made directly in the Auth0 dashboard are corrected on the next deployment, supporting the integrity of the documented access-control baseline.

### Audit Trail

All role CRUD operations (create, update, delete) and permission changes are recorded in Auth0 tenant logs. Log retention depends on plan tier (2 days free, up to 30 days enterprise). For long-term retention, stream tenant logs to an external SIEM (see the `Auth0EventStream` component).

### Separation of Definition and Assignment

This component defines roles and their permissions but does not assign roles to users. User-to-role assignment is governed separately, supporting separation-of-duties controls between infrastructure and identity administration.
