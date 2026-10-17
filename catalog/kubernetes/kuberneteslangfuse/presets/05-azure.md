# Azure

Production Langfuse on AKS with events, media and batch exports in an Azure Blob Storage container, through Langfuse's native Azure driver. Postgres, ClickHouse and Valkey run in the cluster beside Langfuse.

## When to Use

- Langfuse on an AKS cluster where object storage should be Blob Storage, not an in-cluster store
- Media uploads from browsers and SDKs outside the cluster, which presigned Blob Storage URLs serve without any extra route

## Key Configuration Choices

- **Native Azure driver** (`blob_storage.target.azure_blob`) -- the storage account, its blob endpoint and the container all come by reference (the `langfusestorage` AzureStorageAccount and the `langfuse` AzureStorageContainer). Langfuse switches every upload type to Azure at once, so media and exports must use Azure too (here they share the target).
- **The account key as a managed secret** (`account_key`) -- Langfuse authenticates to Blob Storage with the shared key only; it has no managed-identity path there. The key is written to the module's instance Secret and reaches Langfuse through a secretKeyRef. Rotate it by updating the managed secret and applying.
- **Durable secrets, seeded admin, closed sign-up and production sizing** -- the same reasoning as the in-cluster production preset.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `https://langfuse.replaceme.example.com` | The public URL people open Langfuse at | The host your KubernetesHttpRoute or KubernetesIngress serves over the `service` output |
| `$secret/replace-with-azure-storage-account-key` | A managed secret holding the storage account's access key | Azure portal, Storage account, Security + networking, Access keys |
| `$secret/replace-with-langfuse-salt` | A managed secret holding the API-key salt | Create it with `openssl rand -base64 32` |
| `$secret/replace-with-langfuse-encryption-key` | A managed secret holding the encryption key | Create it with `openssl rand -hex 32` (exactly 64 hex characters) |
| `admin@replaceme.example.com` | The seeded admin's email, the first sign-in name | Your platform team's address |

Add a lifecycle management rule that deletes blobs under `events/` after about 30 days; never expire `media/`.

## Related Presets

- **02-production-in-cluster** -- the same shape with SeaweedFS in the cluster instead of Blob Storage.
- **03-aws**, **04-gcp** -- the same shape on S3 and Cloud Storage.
