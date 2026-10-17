# Production In-Cluster

Langfuse for a team that keeps its data: two replicas of the web server and the worker at Langfuse's recommended sizing, a replicated ClickHouse with Keeper, the salt and encryption key declared as managed secrets, a seeded admin and sign-up closed. Every store runs in the cluster beside Langfuse.

## When to Use

- A shared Langfuse instance for one or more teams on your own cluster
- Data that must survive upgrades, recreates and node loss
- An exposed instance with a public URL and single sign-on to follow

## Key Configuration Choices

- **Durable instance secrets** (`secrets.salt`, `secrets.encryption_key`) -- both results live in Postgres. Declared as `$secret/` references, they survive a delete and recreate of this resource; generated values would not, and every API key and stored credential would break.
- **Cluster mode on** (`clickhouse.cluster` with the name from the ClickHouse resource) -- replicated tables from the first start, so ClickHouse can add replicas later without a data migration. The KubernetesClickHouse needs a managed Keeper.
- **Public endpoints for presigned URLs** (`blob_storage.media_upload.public_endpoint`, `blob_storage.batch_export.public_endpoint`) -- browsers upload media and download exports straight from the store, so URLs are signed for the address a route serves SeaweedFS at.
- **A seeded admin and closed sign-up** (`init.admin`, `auth.disable_sign_up`) -- the admin owns the organization from first start; nobody else can create an account until you invite them or add single sign-on.
- **Langfuse's production sizing** (`web.resources`, `worker.resources`) -- requests of 2 CPU and 4 GiB per pod, two replicas each, a disruption budget and zone spreading.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `https://langfuse.replaceme.example.com` | The public URL people open Langfuse at | The host your KubernetesHttpRoute or KubernetesIngress serves over the `service` output |
| `https://s3.langfuse.replaceme.example.com` | The public address of the SeaweedFS S3 API | The host a route serves over the SeaweedFS S3 Service |
| `$secret/replace-with-langfuse-salt` | A managed secret holding the API-key salt | Create it with `openssl rand -base64 32` |
| `$secret/replace-with-langfuse-encryption-key` | A managed secret holding the encryption key | Create it with `openssl rand -hex 32` (exactly 64 hex characters) |
| `admin@replaceme.example.com` | The seeded admin's email, the first sign-in name | Your platform team's address |

## Related Presets

- **01-evaluation** -- the same stores at evaluation sizing with everything generated.
- **03-aws**, **04-gcp**, **05-azure** -- this shape with events, media and exports in the cloud's object storage.
