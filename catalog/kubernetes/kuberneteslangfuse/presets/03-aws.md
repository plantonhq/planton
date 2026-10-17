# AWS

Production Langfuse on EKS with events, media and batch exports in an S3 bucket reached through the pod's IAM role, with no static key anywhere. Postgres, ClickHouse and Valkey run in the cluster beside Langfuse; every object Langfuse writes is encrypted with S3-managed keys.

## When to Use

- Langfuse on an EKS cluster where object storage should be S3, not an in-cluster store
- Teams that forbid long-lived access keys in Kubernetes Secrets
- Media uploads from browsers and SDKs outside the cluster, which presigned S3 URLs serve without any extra route

## Key Configuration Choices

- **S3 by reference** (`blob_storage.target.s3`) -- the bucket and its region both come from the `langfuse-bucket` AwsS3Bucket. Media and exports share the target.
- **Keyless access** (`service_account.workload_identity.eks`) -- the ServiceAccount carries the `langfuse-irsa` AwsIamRole's ARN, so the target leaves `access_key` empty and Langfuse uses the AWS credential chain. The role needs List, Get, Put and Delete on the bucket; Delete is what data retention uses.
- **Server-side encryption on every object** (`sse` with `aes256` on events, media and exports) -- switch to `aws_kms` with `kms_key_id` for a customer-managed key, and add KMS permissions to the role.
- **Durable secrets, seeded admin, closed sign-up and production sizing** -- the same reasoning as the in-cluster production preset.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `https://langfuse.replaceme.example.com` | The public URL people open Langfuse at | The host your KubernetesHttpRoute or KubernetesIngress serves over the `service` output |
| `$secret/replace-with-langfuse-salt` | A managed secret holding the API-key salt | Create it with `openssl rand -base64 32` |
| `$secret/replace-with-langfuse-encryption-key` | A managed secret holding the encryption key | Create it with `openssl rand -hex 32` (exactly 64 hex characters) |
| `admin@replaceme.example.com` | The seeded admin's email, the first sign-in name | Your platform team's address |

The IAM role's trust policy must name the ServiceAccount `system:serviceaccount:langfuse:my-langfuse`: the chart names the ServiceAccount after the resource. Add a lifecycle rule that expires the `events/` prefix after about 30 days; never expire `media/`.

## Related Presets

- **02-production-in-cluster** -- the same shape with SeaweedFS in the cluster instead of S3.
- **04-gcp**, **05-azure** -- the same shape on Google Cloud Storage and Azure Blob Storage.
