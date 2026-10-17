# Google Cloud

Production Langfuse on GKE with events, media and batch exports in a Cloud Storage bucket, reached through Langfuse's native GCS driver and the pod's Google service account, with no key file anywhere. Postgres, ClickHouse and Valkey run in the cluster beside Langfuse.

## When to Use

- Langfuse on a GKE cluster where object storage should be Cloud Storage, not an in-cluster store
- Teams that forbid service account key files in Kubernetes Secrets
- Media uploads from browsers and SDKs outside the cluster, which presigned GCS URLs serve without any extra route

## Key Configuration Choices

- **Native GCS driver** (`blob_storage.target.gcs`) -- more reliable than Cloud Storage's S3 interoperability mode, which mishandles deletes. Langfuse switches every upload type to GCS at once, so media and exports must use GCS too (here they share the target).
- **Keyless access** (`service_account.workload_identity.gke`) -- the ServiceAccount carries the `langfuse` GcpServiceAccount's email, so `credentials_json` stays empty and Langfuse uses application default credentials. That service account needs `roles/storage.objectAdmin` on the bucket.
- **Durable secrets, seeded admin, closed sign-up and production sizing** -- the same reasoning as the in-cluster production preset.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `https://langfuse.replaceme.example.com` | The public URL people open Langfuse at | The host your KubernetesHttpRoute or KubernetesIngress serves over the `service` output |
| `$secret/replace-with-langfuse-salt` | A managed secret holding the API-key salt | Create it with `openssl rand -base64 32` |
| `$secret/replace-with-langfuse-encryption-key` | A managed secret holding the encryption key | Create it with `openssl rand -hex 32` (exactly 64 hex characters) |
| `admin@replaceme.example.com` | The seeded admin's email, the first sign-in name | Your platform team's address |

The Google service account must grant `roles/iam.workloadIdentityUser` to `serviceAccount:<project>.svc.id.goog[langfuse/my-langfuse]`: the chart names the Kubernetes ServiceAccount after the resource. Add a lifecycle rule that deletes objects under `events/` after about 30 days; never expire `media/`.

## Related Presets

- **02-production-in-cluster** -- the same shape with SeaweedFS in the cluster instead of Cloud Storage.
- **03-aws**, **05-azure** -- the same shape on S3 and Azure Blob Storage.
