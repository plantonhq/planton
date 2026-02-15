# Named Firestore Native Database with PITR

This preset creates a named Firestore Native database with point-in-time recovery enabled (7-day version retention) and delete protection. Suitable for production workloads that need disaster recovery capabilities.

## When to Use

- Production workloads that need 7-day point-in-time recovery
- Additional databases beyond the project's default
- Microservice architectures where each service has its own database
- Applications that need data recovery capabilities

## Key Configuration

- **Named database** -- a custom database separate from `(default)`
- **FIRESTORE_NATIVE type** -- modern Firestore with real-time listeners
- **Point-in-time recovery** -- 7-day version history for disaster recovery
- **Delete protection enabled** -- prevents accidental deletion
- **Google-managed encryption** -- default encryption

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | GCP Console or `GcpProject` outputs |
| `<location-id>` | Database location (e.g., `nam5`, `eur3`, `us-east1`) | [Firestore locations](https://cloud.google.com/firestore/docs/locations) |
| `<database-name>` | Database name (4-63 chars, lowercase, hyphens) | Choose a descriptive name (e.g., `orders-db`) |

## Important Notes

- PITR increases storage costs due to version retention
- Location is immutable after creation
- Named databases require Firestore client libraries that support the `databaseId` parameter

## Related Presets

- **01-default-native** -- Default database with minimal configuration
- **03-enterprise-cmek** -- Enterprise edition with CMEK encryption
