# CMEK-Encrypted Database

This preset creates a Cloud Spanner database with customer-managed encryption (CMEK), GCP API-level drop protection, a 3-day version retention period, and an explicit UTC time zone. Designed for regulated and enterprise environments.

## When to Use

- Compliance requirements mandate customer-managed encryption keys (HIPAA, PCI-DSS, FedRAMP)
- Production databases that need protection against accidental deletion
- Enterprise environments with key management policies
- Workloads where the default time zone should be UTC for consistency

## Key Configuration

- **CMEK encryption** -- database encrypted with a customer-managed KMS key; key must be in the same GCP location as the Spanner instance
- **Drop protection enabled** -- prevents deletion of the database and its parent instance through any interface
- **3-day version retention** -- balanced recovery window providing multi-day point-in-time recovery
- **UTC time zone** -- explicit default for SQL timestamp functions, avoiding GCP's default of `America/Los_Angeles`
- **GoogleSQL dialect** -- default dialect with full Spanner feature support

## Important Notes

- The KMS key **must exist in the same location** as the Spanner instance
- The Spanner service account needs `cloudkms.cryptoKeyEncrypterDecrypter` role on the key
- CMEK configuration is **immutable** -- changing the key requires recreating the database
- To delete this database, you must first set `enableDropProtection` to `false`

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the Spanner instance lives | GCP Console or `GcpProject` outputs |
| `<spanner-instance-name>` | Name of the existing Spanner instance | `GcpSpannerInstance` outputs (`instance_name`) |
| `<database-name>` | Name for this database (2-30 chars, lowercase, hyphens/underscores allowed) | Choose a descriptive name (e.g., `secure-db`) |
| `<kms-key-fully-qualified-name>` | Fully qualified KMS key path | `GcpKmsKey` outputs (`key_id`) or format: `projects/{p}/locations/{l}/keyRings/{r}/cryptoKeys/{k}` |

## Related Presets

- **01-basic-database** -- GoogleSQL with minimal configuration (no encryption, no protection)
- **02-postgresql-database** -- PostgreSQL dialect with extended retention
