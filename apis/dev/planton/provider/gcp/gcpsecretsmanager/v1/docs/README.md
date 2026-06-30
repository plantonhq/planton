# Deploying Secrets with GCP Secret Manager

## The Challenge of Secret Management

Hardcoding API keys in source code. Storing passwords in environment variables. Committing certificates to Git. These aren't just rookie mistakes—they're patterns that persist even in mature engineering organizations, often discovered only after a security incident or compliance audit.

The conventional wisdom has long been clear: don't put secrets in code. But the **how** of managing secrets securely at scale—across teams, environments, and cloud platforms—remains a challenge that developers navigate daily. Google Cloud Secret Manager emerged as GCP's answer to this problem: a fully-managed service for storing API keys, passwords, certificates, and other sensitive data with built-in versioning, encryption, and audit logging.

This document explores the spectrum of approaches for deploying and managing secrets using GCP Secret Manager, from manual console clicks to fully-automated infrastructure-as-code patterns. We'll examine what works in production, what doesn't, and how Planton provides a declarative, protobuf-defined interface that makes secret management both secure and maintainable.

## Understanding GCP Secret Manager

Before diving into deployment methods, it's worth understanding what GCP Secret Manager provides—and more importantly, what makes it different from alternatives.

### The Core Concept: Secrets as Versioned Containers

GCP Secret Manager treats secrets not as simple key-value pairs but as **containers for versioned data**. When you create a secret, you're creating a namespace that can hold multiple versions over time. Each version is immutable, numbered sequentially (1, 2, 3...), and can be in one of three states: **Enabled** (accessible), **Disabled** (blocked from access), or **Destroyed** (permanently deleted).

This versioning model solves a critical problem: **how do you rotate credentials without breaking running systems?** With versioning, you can add version 2 (the new credential), migrate applications to use it, then disable and destroy version 1 (the old credential). No atomic cutover required.

### Encryption and Replication

Every secret is encrypted at rest using AES-256. By default, Google manages the encryption keys (handling rotation transparently on a 90-day cycle). For organizations with compliance requirements around key control, you can provide your own **Customer-Managed Encryption Keys (CMEK)** from Cloud KMS, giving you the ability to revoke access by disabling the key.

Secrets can be replicated across regions in two ways:
- **Automatic replication**: Google handles multi-region distribution for high availability (recommended for most use cases)
- **User-managed replication**: You explicitly specify which GCP regions should store the secret (useful for data residency compliance)

### IAM and Audit Logging

Access control operates at two levels: project-wide IAM policies and per-secret IAM bindings. The principle of least privilege applies: grant `roles/secretmanager.secretAccessor` to read specific secrets, `roles/secretmanager.secretVersionAdder` to add new versions, and `roles/secretmanager.admin` sparingly.

When Cloud Audit Logs are enabled (specifically Data Access logs), every secret read is recorded—capturing who accessed which secret version and when. This audit trail is invaluable for compliance and forensic analysis.

### When to Use Secret Manager vs. Alternatives

**Use GCP Secret Manager when:**
- You're already on GCP and want tight integration (Cloud Run, Cloud Functions, GKE)
- You need a fully-managed solution with zero operational overhead
- Compliance requires audit trails and encryption control
- Secrets are relatively static (API keys, database passwords, certificates)

**Consider alternatives when:**
- **HashiCorp Vault**: You need dynamic secrets (on-the-fly credential generation), multi-cloud secret management, or advanced features like secret leasing and revocation. Vault offers unparalleled flexibility but requires running and operating the Vault cluster itself.
- **AWS Secrets Manager**: You operate primarily in AWS (similar feature set to GCP SM, with built-in rotation via Lambda)
- **Kubernetes Secrets**: You need ephemeral, cluster-scoped secrets with minimal external dependencies (though K8s Secrets are only base64-encoded by default and lack the security controls of Secret Manager—best used as a delivery mechanism populated from an external source)

The sweet spot for GCP Secret Manager: **GCP-native applications that need secure, centralized secret storage without the operational burden of self-hosted solutions**.

## Deployment Methods: From Manual to Fully Automated

### Level 0: The Console (Manual Management)

The GCP Console provides a straightforward UI for creating and managing secrets under **Security > Secret Manager**. You provide a secret name, an initial value, and optional settings (replication policy, labels, rotation schedule). The console handles version creation automatically.

**When this makes sense:**
- One-off secret creation during initial setup
- Debugging or inspecting secret versions
- Organizations with strict change control requiring manual steps

**The limitations:**
- Not repeatable or version-controlled
- Prone to human error (typos, wrong region, forgotten IAM grants)
- Common gotcha: forgetting that a secret is a *container*—creating the secret without adding a version means there's no data to retrieve
- Another pitfall: not enabling the Secret Manager API first, resulting in `PERMISSION_DENIED` errors

**Verdict:** Acceptable for exploratory work and emergency fixes, but not a foundation for production secret management.

### Level 1: CLI Scripting (gcloud)

The `gcloud secrets` command family enables scripted secret management. Creating a secret with an initial value:

```bash
echo -n "my-api-key-value" | gcloud secrets create my-api-key \
  --replication-policy="automatic" \
  --data-file=-
```

Adding a new version to rotate a secret:

```bash
echo -n "new-api-key-value" | gcloud secrets versions add my-api-key \
  --data-file=-
```

Accessing a specific version:

```bash
gcloud secrets versions access 2 --secret="my-api-key"
```

**Authentication pattern:** Use Application Default Credentials (ADC) for local development (`gcloud auth application-default login`). On GCP compute platforms (GCE, GKE, Cloud Run), rely on the instance metadata service—no long-lived keys needed.

**When this makes sense:**
- CI/CD pipelines that need to create or rotate secrets
- Migration scripts moving secrets from one system to another
- Break-glass procedures for emergency secret rotation

**Best practices:**
- Handle exit codes and errors (attempting to create an existing secret fails)
- Avoid logging secret values (redirect stderr, mask outputs)
- Consider idempotency (check if secret exists before attempting creation)

**Verdict:** Powerful for automation, but requires careful handling of credentials, error cases, and secret exposure in logs. Suitable for scripted workflows, but doesn't provide the declarative benefits of IaC.

### Level 2: Infrastructure-as-Code (Terraform, Pulumi)

This is where secret management becomes maintainable at scale. IaC tools allow you to **declare the desired state** of secrets and let the tool handle creation, updates, and lifecycle management.

#### Terraform Pattern

Terraform (and its open-source fork OpenTofu) provides two resources:
- `google_secret_manager_secret`: The secret container (name, replication, labels, rotation config)
- `google_secret_manager_secret_version`: The secret data (the actual sensitive value)

This separation is intentional and important—it allows managing metadata separately from the secret payload.

```hcl
resource "google_secret_manager_secret" "api_key" {
  secret_id = "api-key-prod"
  project   = var.project_id
  
  replication {
    automatic {}  # Multi-region automatic replication
  }
  
  labels = {
    environment = "production"
    team        = "platform"
  }
}

resource "google_secret_manager_secret_version" "api_key_v1" {
  secret      = google_secret_manager_secret.api_key.id
  secret_data = var.api_key_value  # Supplied via TF variable, not hardcoded
}

# Grant access to a specific service account
resource "google_secret_manager_secret_iam_member" "accessor" {
  secret_id = google_secret_manager_secret.api_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:app@${var.project_id}.iam.gserviceaccount.com"
}
```

**Critical best practices:**
1. **Never hardcode secret values in .tf files**—supply via variables or CI secrets
2. **Mark variables as sensitive** (`sensitive = true`) to keep them out of logs
3. **Encrypt Terraform state** (use GCS with encryption or Terraform Cloud)
4. **Manage IAM in code** (explicitly grant least-privilege access, don't rely on project-level roles)
5. **Use lifecycle blocks** to prevent accidental deletion (`prevent_destroy = true`)

**Rotation in Terraform:**

```hcl
resource "google_secret_manager_secret" "rotating_password" {
  # ... basic config ...
  
  rotation {
    rotation_period = "2592000s"  # 30 days in seconds
    next_rotation_time = "2024-12-01T00:00:00Z"
  }
  
  topics {
    name = google_pubsub_topic.secret_rotation.id
  }
}
```

This doesn't rotate the secret automatically—it sends a notification to the Pub/Sub topic at the specified interval. You handle the rotation logic (typically a Cloud Function subscribed to the topic that generates a new credential and adds a secret version).

#### Pulumi Pattern

Pulumi uses general-purpose programming languages (TypeScript, Python, Go) rather than HCL. The secret management pattern is similar, but Pulumi has built-in support for marking values as secrets:

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as gcp from "@pulumi/gcp";

const apiKeyValue = pulumi.secret(process.env.API_KEY!);

const apiKeySecret = new gcp.secretmanager.Secret("api-key", {
    secretId: "api-key-prod",
    replication: { automatic: true },
});

new gcp.secretmanager.SecretVersion("api-key-version", {
    secret: apiKeySecret.name,
    secretData: apiKeyValue,  // Encrypted in Pulumi state
});
```

Pulumi automatically encrypts secret values in its state file (using a passphrase or cloud KMS), making it safer to include secret material in your program. Still, best practice: **retrieve secrets from a secure source at runtime** rather than embedding them.

**When IaC makes sense:**
- Production environments requiring reproducibility
- Multi-environment deployments (dev, staging, prod with consistent patterns)
- GitOps workflows where infrastructure is version-controlled
- Teams that need code review and approval for secret creation

**Verdict:** This is the production-grade approach. IaC provides repeatability, version control, and integration with broader infrastructure management. The separation of secret metadata (in code) and secret values (supplied at runtime) balances transparency with security.

### Level 3: Kubernetes Integration (External Secrets Operator)

For applications running on GKE (or any Kubernetes cluster), the **External Secrets Operator (ESO)** bridges GCP Secret Manager with Kubernetes Secrets. Rather than provisioning secrets, ESO **syncs existing secrets from Secret Manager into Kubernetes**.

#### The Pattern

1. **Create a SecretStore** pointing to GCP Secret Manager
2. **Configure Workload Identity** (no static credentials)
3. **Declare ExternalSecret resources** referencing secrets by name
4. **ESO fetches and creates Kubernetes Secrets** for your pods to consume

```yaml
# SecretStore configuration
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: gcp-secret-store
  namespace: production
spec:
  provider:
    gcpsm:
      projectID: "my-gcp-project"
      auth:
        workloadIdentity:
          clusterLocation: us-central1
          clusterName: prod-cluster
          serviceAccountRef:
            name: external-secrets-sa
```

```yaml
# ExternalSecret that syncs a secret
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: database-credentials
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: gcp-secret-store
  target:
    name: db-creds  # Kubernetes Secret to create
  data:
    - secretKey: password
      remoteRef:
        key: prod-db-password
        version: latest
```

**Workload Identity Setup:**

1. Create a GCP service account with `roles/secretmanager.secretAccessor`
2. Create a Kubernetes service account for ESO
3. Bind them: `gcloud iam service-accounts add-iam-policy-binding ...`
4. Annotate the K8s SA: `iam.gke.io/gcp-service-account: eso@project.iam.gserviceaccount.com`

**Why this matters:** No static GCP credentials stored in your cluster. The ESO pods use Workload Identity to exchange K8s tokens for GCP tokens dynamically.

**When this makes sense:**
- GKE applications that need secrets but shouldn't store them in etcd
- Multi-cluster deployments pulling from a centralized Secret Manager
- GitOps workflows where ExternalSecret resources live in Git, but values come from Secret Manager

**Alternative: Secret Manager CSI Driver**

Google also provides a CSI driver that mounts secrets directly as volumes (bypassing Kubernetes Secrets entirely). This is useful if you want to avoid secrets touching etcd at all.

**Verdict:** ESO is the preferred pattern for Kubernetes integration. It decouples application config (K8s Secrets) from the source of truth (GCP Secret Manager), enables centralized rotation, and avoids static credentials.

## Production Essentials

### IAM: Principle of Least Privilege

The most common security mistake: granting broad IAM roles like `roles/secretmanager.admin` at the project level.

**Best practices:**
1. **Grant access per secret**, not per project: Use `google_secret_manager_secret_iam_member` to bind specific identities to specific secrets
2. **Use narrow roles**: `secretAccessor` for reading, `secretVersionAdder` for rotation, `admin` sparingly
3. **Separate environments**: Keep prod secrets in a separate GCP project from dev/staging
4. **Avoid long-lived credentials**: Use Workload Identity (GKE), instance metadata (GCE), or Workload Identity Federation (external systems) instead of service account JSON keys

### Secret Rotation: Automatic Notification, Manual Implementation

GCP Secret Manager doesn't rotate secrets for you—it **notifies you when rotation is due**. You handle the actual rotation logic.

**The pattern:**
1. Set a rotation schedule on the secret (e.g., every 90 days)
2. Attach a Pub/Sub topic to receive `SECRET_ROTATE` events
3. Subscribe a Cloud Function to that topic
4. The function: generates a new credential, updates the target system (e.g., database), adds a new secret version
5. Applications using the secret: fetch the new version (or restart if pinned to `latest`)

**Why not automatic?** Because rotation logic is context-specific. Rotating a database password requires updating the database itself. Rotating an API key might require calling a third-party API. Secret Manager can't know your domain logic.

### Audit Logging and Monitoring

Enable **Cloud Audit Logs (Data Access)** for Secret Manager to record every secret read. These logs answer critical questions during incident response:
- Who accessed which secret?
- When did the access occur?
- Were there any unexpected access patterns?

Set up alerts for:
- Excessive access errors (permission denied—might indicate misconfigured IAM)
- Unusual access frequency (sudden spike could indicate a compromised credential being exploited)
- Secret rotation failures (if using automated rotation)

### Encryption: Default vs. CMEK

By default, secrets are encrypted with Google-managed keys. For most organizations, this is sufficient.

**Use CMEK (Customer-Managed Encryption Keys) when:**
- Compliance mandates control over encryption keys
- You need the ability to revoke access by disabling the key
- Audit requirements demand separate logs for key usage

**CMEK trade-offs:**
- More setup complexity (grant `roles/cloudkms.cryptoKeyEncrypterDecrypter` to Secret Manager's service account)
- Operational risk (disabling the KMS key makes secrets inaccessible—an instant outage)
- Additional cost (KMS operations are billed separately)
- For automatic replication, the key must be in the "global" location; for user-managed, you need a key per region

**Verdict:** Only use CMEK if you have a strong requirement. The operational risk outweighs the benefit for most use cases.

### Common Anti-Patterns to Avoid

1. **Hardcoding secrets in code**: The problem Secret Manager exists to solve. Never commit secrets to Git, Docker images, or config files.

2. **Over-broad IAM**: Granting `roles/editor` or `roles/owner` gives implicit secret access. Always grant explicit, narrow roles.

3. **Not disabling old versions after rotation**: When you rotate a credential, disable the old version after confirming the new one works. Otherwise, a leaked old credential remains valid.

4. **Using the `latest` alias blindly**: Always pin applications to specific version numbers. Using `latest` means a bad secret version pushed by mistake will instantly break production. With pinned versions, you control when updates happen.

5. **Setting secret expiration in production**: GCP allows setting an expiration timestamp that auto-destroys the secret. This can cause unplanned outages. Use manual lifecycle management instead.

6. **Ignoring service account scopes**: On GCE, ensure instances have the `cloud-platform` or `secretmanager` OAuth scope. On GKE without Workload Identity, node-level scopes apply—which can be overly broad.

7. **Storing non-secrets in Secret Manager**: Avoid storing general config (non-sensitive data) in Secret Manager. It incurs unnecessary cost (\$0.06/version/month) and complicates deployments. Use Cloud Storage, Config Maps, or parameter stores for non-sensitive config.

## The 80/20 Configuration Philosophy

Most secret management needs are simple: create a secret, store a value, grant access. The complexity comes from advanced features most users don't need.

**Core fields (80% of use cases):**
- **Project ID**: Where to create the secret
- **Secret ID**: The name (unique per project)
- **Replication**: Automatic (multi-region) vs. user-managed (specific regions)
- **Initial value**: The secret data (handled securely, not in plaintext manifests)
- **Labels**: For organization and ownership tracking

**Advanced fields (20% of use cases):**
- Rotation schedule (period and start time)
- Pub/Sub topics for notifications
- CMEK (customer-managed encryption keys)
- Expiration timestamps
- Version aliases

This is where Planton's design philosophy shines: **expose the 80% in a simple protobuf spec, make the 20% optional**.

### Planton's Approach

The `GcpSecretsManagerSpec` in Planton is deliberately minimal:

```protobuf
message GcpSecretsManagerSpec {
  string project_id = 1;          // Required: target GCP project
  repeated string secret_names = 2; // List of secrets to create
}
```

This design embraces the 80/20 principle:
- **Default to automatic replication** (recommended by Google for most cases)
- **Default to Google-managed encryption** (sufficient for most compliance needs)
- **No rotation schedule required** (most teams handle rotation manually or via separate automation)
- **Secret values provided out-of-band** (via CLI flags or secure files, not in manifests)

For advanced use cases, the underlying Pulumi or Terraform modules support the full feature set—but the declarative API focuses on simplicity.

**Example: Basic secret creation**

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpSecretsManager
metadata:
  name: app-secrets
spec:
  projectId: my-gcp-project
  secretNames:
    - api-key-prod
    - database-password
    - oauth-client-secret
```

This creates three empty secrets with automatic replication. Values are supplied separately (perhaps via `planton apply --set-secret api-key-prod=...`), keeping sensitive data out of version control.

**Why this matters:** The manifest can live in Git (it contains no secrets), be reviewed in pull requests, and deployed via CI/CD—while the actual secret values flow through a secure channel.

## Integration Patterns

### Application Runtime Access

Applications fetch secrets at runtime using the Secret Manager client library:

```python
from google.cloud import secretmanager

client = secretmanager.SecretManagerServiceClient()
secret_name = "projects/my-project/secrets/api-key/versions/3"  # Pin to version 3
response = client.access_secret_version(name=secret_name)
api_key = response.payload.data.decode('UTF-8')
```

**Best practice:** Fetch once at startup and cache in memory. Don't call Secret Manager on every request (it's an API call with latency and rate limits).

**Cloud Run and Cloud Functions integration:** These platforms allow mounting secrets as environment variables or files directly in the console or deployment config. The platform fetches the secret on your behalf using the service's identity.

### CI/CD Pipelines

In Google Cloud Build:

```yaml
availableSecrets:
  secretManager:
    - versionName: "projects/my-project/secrets/ci-api-key/versions/1"
      env: "API_KEY"
```

Cloud Build fetches the secret and makes it available as `$API_KEY` in build steps.

For GitHub Actions, use the `google-github-actions/get-secretmanager-secrets` action:

```yaml
- uses: google-github-actions/auth@v1
  with:
    workload_identity_provider: 'projects/.../providers/...'
    service_account: 'ci-deployer@project.iam.gserviceaccount.com'

- uses: google-github-actions/get-secretmanager-secrets@v1
  with:
    secrets: |-
      api_key:projects/my-project/secrets/api-key/versions/latest
```

**Benefit:** Secrets are fetched just-in-time, not stored permanently in your CI system. Rotation happens in Secret Manager, and pipelines automatically pick up new values.

### Multi-Cloud Scenarios

For applications running outside GCP (AWS, Azure, on-prem), use **Workload Identity Federation** to avoid static credentials. This allows, for example, an AWS Lambda to impersonate a GCP service account via OIDC, fetch secrets from Secret Manager, then execute.

Alternatively, some organizations sync secrets between cloud providers (e.g., copy from GCP Secret Manager to AWS Secrets Manager) for local access patterns, accepting the complexity of keeping them in sync.

## Cost and Operational Considerations

**Pricing breakdown:**
- **Active secret versions**: \$0.06/version/month/location (first 6 versions free)
- **Access operations**: \$0.03 per 10,000 reads/writes
- **Rotation notifications**: \$0.05 per rotation event (first 3/month free)

**Typical costs:** 50 secrets with 2 versions each, automatic replication (~3 regions), 1,000 accesses/day:
- Versions: 50 * 2 * 3 = 300 - 6 = 294 * \$0.06 = **\$17.64/month**
- Access: 50 * 30,000 = 1.5M / 10,000 * \$0.03 = **\$4.50/month**
- **Total: ~\$22/month**

For most organizations, this is negligible compared to compute and storage costs.

**Cost optimization:**
1. **Limit active versions**: Destroy or disable old versions after rotation (disabled versions still count as "active" for billing—destroy them after a grace period)
2. **Cache secret values**: Fetch once at startup, not on every request
3. **Consolidate where possible**: Fewer high-use secrets vs. many rarely-accessed secrets

**Operational best practices:**
- Audit IAM permissions regularly (use IAM Recommender to identify unused access)
- Label secrets with ownership (`team: platform`, `env: production`)
- Set up alerts for permission errors and access spikes
- Document rotation procedures (even if manual) in runbooks

## Conclusion: The Production-Ready Pattern

The evolution of secret management mirrors a broader shift in infrastructure: from manual, error-prone processes to declarative, auditable systems.

**The pattern that works in production:**
1. **Store secrets in GCP Secret Manager** (centralized, encrypted, audited)
2. **Manage secret resources via IaC** (Terraform/Pulumi for repeatability, version control)
3. **Supply secret values securely** (via CI secrets, CLI flags, not in manifests)
4. **Grant least-privilege access** (per-secret IAM bindings, Workload Identity)
5. **Rotate on a schedule** (automated notifications, scripted rotation logic)
6. **Sync to Kubernetes via ESO** (for GKE apps, using Workload Identity)
7. **Monitor and audit** (Cloud Audit Logs, alerts on anomalies)

Planton's declarative API abstracts this complexity: define the secrets you need, let the platform handle provisioning, and supply values through secure channels. The result: secret management that's both secure by default and maintainable at scale.

The shift from "where do I put this API key?" to "how do I declaratively manage secret lifecycle?" represents a maturation in how we think about security infrastructure. GCP Secret Manager—when integrated thoughtfully—makes that shift achievable.

