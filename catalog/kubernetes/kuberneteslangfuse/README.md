# Overview

The KubernetesLangfuse API resource installs Langfuse, the open-source LLM engineering platform for tracing, prompt management, evaluations and datasets, on a Kubernetes cluster from the official `langfuse` Helm chart (chart 2.1.x, Langfuse v4).

## Why We Created This API Resource

Langfuse is one application with four stores behind it: Postgres for users, projects, prompts and API keys; ClickHouse for traces, observations and scores; Redis or Valkey for the ingestion queues and caches; and object storage for every ingested event, media attachment and batch export. The official chart can bundle all four, but a bundled store is a store nobody sized, backed up or upgraded on purpose.

Planton already has a first-class kind for each store, so this kind composes them by reference instead. Each store keeps its own lifecycle, backups and sizing, and each one renders as its own node on the architecture diagram:

- **`KubernetesPostgres`** for the Postgres database, wired to its read-write Service and its application credential Secret.
- **`KubernetesClickHouse`** for ClickHouse, wired to its client Service, its auth Secret and its logical cluster name.
- **`KubernetesValkey`** for the queues and caches, wired to its Service, its ACL user and its password Secret.
- **`KubernetesSeaweedFs`** (or an S3-compatible store, AWS S3, Cloudflare R2, Google Cloud Storage or Azure Blob Storage) for blob storage.

The `self-hosted-langfuse` InfraChart composes the whole set in one deployment.

## Key Features

### One Release, Two Workloads

- **Web server and worker**: the web server serves the UI, the public API and the OTLP ingestion endpoint; the worker consumes the queues, loads events into ClickHouse and runs evaluations, exports and retention. Each has its own replicas, resources, probes, autoscaling (HPA, KEDA or VPA), disruption budget, scheduling and rollout strategy.
- **The chart's bundled stores never ship**: the module pins `postgresql`, `redis`, `clickhouse` and `s3` off after every merge, including the `helm_values` escape hatch.

### Secured by Default

- **No secret ever renders into Helm values**: every sensitive value the spec carries (store connection URLs, the SMTP URL, single sign-on client secrets, blob storage keys, the AI provider key, the license key) is written into the module-owned `<name>-secrets` Secret, and Langfuse reads it through a secretKeyRef.
- **Generated instance secrets**: the API-key salt, the 64-hex-character encryption key and the NextAuth session secret are generated once when the spec leaves them empty, and kept across applies.
- **Headless initialization**: by default Langfuse starts with a seeded organization, a seeded project and a generated project key pair, so an SDK can send traces the moment the install finishes. An optional seeded admin gets a generated password.

### The Full Configuration Surface

- **Sign-in**: email and password, sign-up, email verification, session lifetime, SSO-enforced domains, automatic organization and project membership for new users, and fifteen single sign-on providers (Google, GitHub, GitHub Enterprise, GitLab, Azure AD, Okta, Authentik, OneLogin, Auth0, Cognito, Keycloak, JumpCloud, WorkOS, WordPress and a custom OIDC provider).
- **Blob storage**: one target for events, with optional separate targets for media and batch exports, server-side encryption on S3, public endpoints for presigned URLs, and concurrency tuning.
- **Product settings**: SMTP, telemetry, experimental features, caching, ingestion queue sharding, Langfuse AI (the Assistant and Ask AI), evaluation runtime, the Slack app, Enterprise Edition features (license, Instance Management API, UI customization, ingestion masking), Langfuse's own logs and traces, and outbound proxies, private certificate authorities and host allowlists.
- **Cloud identity**: `service_account.workload_identity` federates the ServiceAccount with an AWS IAM role (EKS IRSA), a Google service account (GKE Workload Identity) or an Azure managed identity, for keyless access to S3, GCS, Bedrock, Vertex and SES.
- **Escape hatches**: `extra_env` and `extra_env_from_secret` for any variable the spec does not type, and `helm_values` for any chart key it does not type.

## Benefits

- **Composable stores**: each of the four stores is a separate resource with its own backups and sizing, wired by `valueFrom`, so the diagram shows exactly what Langfuse depends on.
- **Durable secrets**: the salt and encryption key are bound to the Postgres data. Declare both as managed secrets for a durable install, and a recreate never invalidates existing API keys or stored credentials.
- **Consistency**: the OpenTofu and Pulumi modules render the same values document and the same Secret.

## Outputs

- `namespace`: the namespace Langfuse runs in
- `service`: the web Service name
- `endpoint`: the in-cluster address of the web server (`http://<service>.<namespace>.svc.cluster.local:<port>`)
- `external_url`: the public URL Langfuse is configured for
- `otel_endpoint`: the OTLP traces endpoint (`<endpoint>/api/public/otel`)
- `port_forward_command`: the `kubectl port-forward` command that opens the UI on localhost
- `instance_secret_name`: the module-owned Secret holding every declared and generated secret
- `organization_id`, `project_id`: the seeded organization and project
- `project_public_key`: the seeded project's public key
- `project_secret_key_secret`: where the seeded project's secret key lives (Secret name and key)
- `admin_email`, `admin_password_secret`: the seeded admin's sign-in name and where the password lives

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
