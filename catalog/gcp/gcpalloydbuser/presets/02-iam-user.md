# IAM User (ALLOYDB_IAM_USER)

This preset maps a GCP IAM principal (here, a service account email) to an AlloyDB database user with no stored password.

## When to Use

- Workloads that already authenticate through AlloyDB Auth Proxy or Language Connectors with IAM
- Eliminating long-lived database passwords from application configuration

## Key Configuration Choices

- **userId is the IAM principal email** — must match the identity presented at connect time
- **No password** — spec CEL rejects passwords on ALLOYDB_IAM_USER

## Related Kinds

- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — the identity behind IAM users
