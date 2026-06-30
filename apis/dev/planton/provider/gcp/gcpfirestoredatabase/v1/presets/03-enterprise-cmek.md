# Enterprise Firestore Database with CMEK

This preset creates an Enterprise-edition Firestore Native database with customer-managed encryption (CMEK), point-in-time recovery, and delete protection. Designed for regulated and enterprise environments that require maximum control over data encryption.

## When to Use

- Compliance requirements mandate customer-managed encryption keys (HIPAA, PCI-DSS, FedRAMP)
- Enterprise SLA and advanced security features are needed
- Production databases handling sensitive data
- Organizations with centralized key management policies

## Key Configuration

- **ENTERPRISE edition** -- enhanced SLA and advanced security features
- **FIRESTORE_NATIVE type** -- required for ENTERPRISE edition
- **CMEK encryption** -- database encrypted with a customer-managed KMS key
- **Point-in-time recovery** -- 7-day version history for disaster recovery
- **Delete protection enabled** -- prevents accidental deletion through any interface

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | GCP Console or `GcpProject` outputs |
| `<location-id>` | Database location (e.g., `nam5`, `eur3`, `us-east1`) | [Firestore locations](https://cloud.google.com/firestore/docs/locations) |
| `<database-name>` | Database name (4-63 chars, lowercase, hyphens) | Choose a descriptive name (e.g., `secure-db`) |
| `<kms-key-fully-qualified-name>` | Fully qualified KMS key path | `GcpKmsKey` outputs (`key_id`) or `projects/{p}/locations/{l}/keyRings/{r}/cryptoKeys/{k}` |

## Important Notes

- The KMS key **must be in the same location** as the database (nam5 → us, eur3 → europe)
- CMEK configuration is **immutable** -- changing the key requires recreating the database
- ENTERPRISE edition is **immutable** -- cannot downgrade to STANDARD after creation
- ENTERPRISE requires FIRESTORE_NATIVE type (DATASTORE_MODE is not supported)
- To delete this database, first set `deleteProtectionState` to `DELETE_PROTECTION_DISABLED`

## Related Presets

- **01-default-native** -- Default database with minimal configuration
- **02-named-native-pitr** -- Named database with PITR but without CMEK
