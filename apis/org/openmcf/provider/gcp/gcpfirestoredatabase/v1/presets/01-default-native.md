# Default Firestore Native Database

This preset creates the project's default Firestore Native database in the US multi-region (nam5) with delete protection enabled. This is the primary database that client libraries connect to when no database ID is specified.

## When to Use

- Getting started with Firestore for the first time
- Any project that needs a primary document database
- Mobile and web applications using Firebase/Firestore SDKs
- Applications that don't need a custom-named database

## Key Configuration

- **`(default)` database** -- the primary database; client libraries connect here by default
- **FIRESTORE_NATIVE type** -- modern Firestore with real-time listeners and offline support
- **nam5 location** -- US multi-region for high availability
- **Delete protection enabled** -- prevents accidental deletion in production
- **Google-managed encryption** -- no CMEK required for most workloads

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | GCP Console or `GcpProject` outputs |

## Important Notes

- Only one `(default)` database can exist per project
- The `(default)` name is special -- do not confuse it with a regular name
- Location is immutable after creation

## Related Presets

- **02-named-native-pitr** -- Named database with point-in-time recovery
- **03-enterprise-cmek** -- Enterprise edition with CMEK encryption
