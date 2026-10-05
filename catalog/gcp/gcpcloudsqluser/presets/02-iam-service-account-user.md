# IAM Service Account User (Passwordless)

This preset creates a passwordless database user for a workload's service account. Authentication flows through IAM — no credential to store, leak, or rotate — which is the strongest auth posture Cloud SQL offers.

## When to Use

- Workloads that already run as a GCP service account (GKE with Workload Identity, Cloud Run, CI runners on WIF)
- Eliminating database passwords from your secret inventory entirely

## Prerequisites

- **PostgreSQL**: the instance must carry the database flag `cloudsql.iam_authentication = "on"` (set it in the `GcpCloudSql` spec's `databaseFlags`)
- The service account also needs the `roles/cloudsql.instanceUser` IAM role and connects through the Auth Proxy / connectors

## Key Configuration Choices

- **`type: CLOUD_IAM_SERVICE_ACCOUNT`** — IAM authentication; setting a password is rejected pre-deploy
- **`userName` is the SA email** — on MySQL, GCP stores it truncated before the `@` (the `user_name` output reflects what clients authenticate with)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `my-gcp-project-123` | GCP project ID | GCP Console or `GcpProject` outputs |
| `my-postgres-prod` | Your instance's resource name | The instance manifest |
| `userName` | The service account's email | `GcpServiceAccount` outputs (`email`) |

## Related Presets

- **01-application-user** — the classic password credential when IAM auth is not available

## Related Kinds

- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — the identity this user maps to
- [GcpCloudSql](/docs/catalog/gcp/gcpcloudsql) — the instance (remember the PostgreSQL IAM flag)
