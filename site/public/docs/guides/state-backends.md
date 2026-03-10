---
title: "State Backends"
description: "Configure state storage for Pulumi, OpenTofu, and Terraform using manifest labels for automatic backend detection"
icon: "database"
order: 25
---

# State Backends

This guide covers how to configure state storage for Pulumi, OpenTofu, and Terraform deployments. OpenMCF reads backend configuration from manifest labels and CLI flags, eliminating the need for separate backend configuration files.

For the conceptual overview of state management — what state is, why it matters, and how each engine handles it — see [State Management](../concepts/state-management).

---

## Quick Reference

| Provisioner   | Backend Labels                                                                                                      |
| ------------- | ------------------------------------------------------------------------------------------------------------------- |
| **Terraform** | `terraform.openmcf.org/backend.type`, `backend.bucket`, `backend.key`, `backend.region`, `backend.endpoint` |
| **OpenTofu**  | `tofu.openmcf.org/backend.type`, `backend.bucket`, `backend.key`, `backend.region`, `backend.endpoint`      |
| **Pulumi**    | `pulumi.openmcf.org/stack.name` (uses Pulumi Cloud or local)                                                |

### S3 Backend Labels (Complete Example)

```yaml
metadata:
  labels:
    openmcf.org/provisioner: terraform
    terraform.openmcf.org/backend.type: s3
    terraform.openmcf.org/backend.bucket: my-terraform-state
    terraform.openmcf.org/backend.key: path/to/state.tfstate
    terraform.openmcf.org/backend.region: us-west-2
```

---

## CLI Flags for Backend Configuration

In addition to manifest labels, you can configure backends using CLI flags. **CLI flags take precedence over manifest labels**, allowing you to override settings for CI/CD or testing.

### Backend CLI Flags

| Flag                 | Description                                         |
| -------------------- | --------------------------------------------------- |
| `--backend-type`     | Backend type: `local`, `s3`, `gcs`, `azurerm`       |
| `--backend-bucket`   | Bucket/container name for state storage             |
| `--backend-key`      | State file path within the bucket                   |
| `--backend-region`   | AWS region (use `auto` for S3-compatible backends)  |
| `--backend-endpoint` | Custom S3-compatible endpoint URL (R2, MinIO, etc.) |
| `--reconfigure`      | Force backend reconfiguration                       |

### Example: CLI-Based Configuration

```bash
# Full CLI configuration
openmcf apply -f manifest.yaml \
    --backend-type s3 \
    --backend-bucket my-state-bucket \
    --backend-key env/prod/terraform.tfstate \
    --backend-region us-west-2

# Override just the bucket (other values from manifest)
openmcf apply -f manifest.yaml --backend-bucket different-bucket
```

### Configuration Precedence

```
CLI Flags > Manifest Labels > Environment Variables > Interactive Prompts > Defaults
```

This means:
1. CLI flags always win if provided
2. Manifest labels override environment variables
3. Environment variables provide defaults when manifest labels are absent
4. If required values are missing, you'll be prompted interactively
5. Local backend is used if nothing is configured

---

## Environment Variables

For scenarios where CLI flags are cumbersome or manifests can't be modified, you can configure backend settings via environment variables. This is especially useful for CI/CD pipelines and 12-factor app patterns.

### Supported Variables

| Variable                           | Description                                         |
| ---------------------------------- | --------------------------------------------------- |
| `OPENMCF_BACKEND_TYPE`     | Backend type: `s3`, `gcs`, `azurerm`, `local`       |
| `OPENMCF_BACKEND_BUCKET`   | State bucket/container name                         |
| `OPENMCF_BACKEND_REGION`   | AWS region (use `auto` for S3-compatible backends)  |
| `OPENMCF_BACKEND_ENDPOINT` | Custom S3-compatible endpoint URL (R2, MinIO, etc.) |

**Note:** `backend.key` is intentionally NOT configurable via environment variable. State file paths should be explicit and traceable, so they must come from manifest labels or CLI flags.

### Usage Example

```bash
# Set environment variables (e.g., in CI/CD pipeline)
export OPENMCF_BACKEND_TYPE=s3
export OPENMCF_BACKEND_BUCKET=my-state-bucket
export OPENMCF_BACKEND_REGION=auto
export OPENMCF_BACKEND_ENDPOINT=https://account-id.r2.cloudflarestorage.com

# Run with key from manifest
openmcf apply -f manifest.yaml

# Or provide key via CLI flag
openmcf apply -f manifest.yaml --backend-key env/prod/state.tfstate
```

### CI/CD Example (GitHub Actions)

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    env:
      OPENMCF_BACKEND_TYPE: s3
      OPENMCF_BACKEND_BUCKET: ${{ secrets.STATE_BUCKET }}
      OPENMCF_BACKEND_REGION: auto
      OPENMCF_BACKEND_ENDPOINT: ${{ secrets.R2_ENDPOINT }}
    steps:
      - uses: actions/checkout@v4
      - name: Deploy infrastructure
        run: openmcf apply -f manifest.yaml
```

### Override Behavior

Environment variables serve as defaults that can be overridden:

```bash
# Environment sets bucket to "default-bucket"
export OPENMCF_BACKEND_BUCKET=default-bucket

# Manifest label overrides to "manifest-bucket"
# terraform.openmcf.org/backend.bucket: manifest-bucket

# CLI flag overrides to "cli-bucket" (highest priority)
openmcf apply -f manifest.yaml --backend-bucket cli-bucket
```

---

## Pulumi State

Pulumi stores state either in Pulumi Cloud (default) or locally. The stack name label identifies where your state is stored.

### Stack Name Label

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPostgres
metadata:
  name: app-database
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/stack.name: prod.KubernetesPostgres.app-database
spec:
  container:
    replicas: 1
```

### Stack Name Format

The stack name follows the pattern: `<environment>.<project>.<stack>`

- **environment**: Deployment environment (prod, staging, dev)
- **project**: Pulumi project name (usually matches the kind)
- **stack**: Unique identifier for this deployment

### Pulumi Backend Options

**1. Pulumi Cloud (Recommended)**

```bash
# Login to Pulumi Cloud
pulumi login

# State is automatically stored in Pulumi Cloud
openmcf apply -f database.yaml
```

**2. Local Backend**

```bash
# Use local filesystem for state
pulumi login --local

# State stored in ~/.pulumi/
openmcf apply -f database.yaml
```

**3. Self-hosted Backend (S3, GCS, Azure)**

```bash
# Use S3 for state
pulumi login s3://my-pulumi-state-bucket

# Or GCS
pulumi login gs://my-pulumi-state-bucket

# Or Azure Blob
pulumi login azblob://my-container
```

---

## OpenTofu / Terraform State

OpenTofu and Terraform use a backend configuration to store state. OpenMCF reads this configuration from manifest labels.

### Label Format

Each provisioner uses its own label prefix. The backend configuration requires these labels:

| Label            | Description                                      | Required                  |
| ---------------- | ------------------------------------------------ | ------------------------- |
| `backend.type`   | Backend type: `s3`, `gcs`, `azurerm`, or `local` | Yes                       |
| `backend.bucket` | Bucket/container name for remote state           | Yes (for remote backends) |
| `backend.key`    | State file path within the bucket                | Yes                       |
| `backend.region` | AWS region (S3 only)                             | Yes (for S3)              |

**For Terraform (S3 backend):**
```yaml
metadata:
  labels:
    openmcf.org/provisioner: terraform
    terraform.openmcf.org/backend.type: s3
    terraform.openmcf.org/backend.bucket: my-terraform-state
    terraform.openmcf.org/backend.key: vpc/production.tfstate
    terraform.openmcf.org/backend.region: us-west-2
```

**For OpenTofu (GCS backend):**
```yaml
metadata:
  labels:
    openmcf.org/provisioner: tofu
    tofu.openmcf.org/backend.type: gcs
    tofu.openmcf.org/backend.bucket: my-tofu-state
    tofu.openmcf.org/backend.key: vpc/production
```

### Backward Compatibility

For backward compatibility:
- OpenTofu accepts `terraform.openmcf.org/*` labels if provisioner-specific labels are not present
- `backend.object` label is still supported but deprecated in favor of `backend.key`

```yaml
metadata:
  labels:
    openmcf.org/provisioner: tofu
    # Legacy labels - still work
    terraform.openmcf.org/backend.type: s3
    terraform.openmcf.org/backend.bucket: my-bucket
    terraform.openmcf.org/backend.object: path/to/state.tfstate  # deprecated, use backend.key
    terraform.openmcf.org/backend.region: us-west-2
```

We recommend using provisioner-specific labels with `backend.key` for clarity.

---

## Supported Backend Types

### Amazon S3

Store state in an S3 bucket with optional DynamoDB locking.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsVpc
metadata:
  name: production-vpc
  labels:
    openmcf.org/provisioner: terraform
    terraform.openmcf.org/backend.type: s3
    terraform.openmcf.org/backend.bucket: my-terraform-state
    terraform.openmcf.org/backend.key: vpc/production.tfstate
    terraform.openmcf.org/backend.region: us-west-2
spec:
  cidrBlock: 10.0.0.0/16
  region: us-west-2
```

**Required labels:**
- `backend.type`: `s3`
- `backend.bucket`: S3 bucket name
- `backend.key`: State file path within the bucket
- `backend.region`: AWS region where the bucket is located

**Prerequisites:**
- S3 bucket must exist
- IAM permissions: `s3:GetObject`, `s3:PutObject`, `s3:DeleteObject`, `s3:ListBucket`
- Optional: DynamoDB table for state locking (configured via environment)

---

### S3-Compatible Backends (Cloudflare R2, MinIO)

OpenMCF supports S3-compatible backends like Cloudflare R2, MinIO, and other S3-compatible storage services. The CLI automatically detects these backends and configures the necessary compatibility flags.

**Detection signals:**
- `backend.region` set to `auto`
- `backend.endpoint` is specified

When an S3-compatible backend is detected, OpenMCF automatically adds:
- `skip_credentials_validation = true`
- `skip_region_validation = true`
- `skip_requesting_account_id = true`
- `skip_metadata_api_check = true`
- `skip_s3_checksum = true`
- `use_path_style = true`

#### Cloudflare R2 Example

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsVpc
metadata:
  name: production-vpc
  labels:
    openmcf.org/provisioner: terraform
    terraform.openmcf.org/backend.type: s3
    terraform.openmcf.org/backend.bucket: my-r2-state-bucket
    terraform.openmcf.org/backend.key: vpc/production.tfstate
    terraform.openmcf.org/backend.region: auto
    terraform.openmcf.org/backend.endpoint: https://<account-id>.r2.cloudflarestorage.com
spec:
  cidrBlock: 10.0.0.0/16
```

#### CLI-Based R2 Configuration

```bash
openmcf apply -f manifest.yaml \
    --backend-type s3 \
    --backend-bucket my-r2-state-bucket \
    --backend-key vpc/production.tfstate \
    --backend-region auto \
    --backend-endpoint https://<account-id>.r2.cloudflarestorage.com
```

#### MinIO Example

```yaml
metadata:
  labels:
    openmcf.org/provisioner: tofu
    tofu.openmcf.org/backend.type: s3
    tofu.openmcf.org/backend.bucket: terraform-state
    tofu.openmcf.org/backend.key: app/state.tfstate
    tofu.openmcf.org/backend.region: auto
    tofu.openmcf.org/backend.endpoint: https://minio.example.com:9000
```

**Prerequisites for R2:**
- Cloudflare R2 bucket must exist
- R2 API token with read/write permissions
- Configure credentials via AWS CLI profile or environment variables:
  ```bash
  export AWS_ACCESS_KEY_ID=<r2-access-key>
  export AWS_SECRET_ACCESS_KEY=<r2-secret-key>
  ```

---

### Google Cloud Storage

Store state in a GCS bucket.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GkeCluster
metadata:
  name: staging-cluster
  labels:
    openmcf.org/provisioner: tofu
    tofu.openmcf.org/backend.type: gcs
    tofu.openmcf.org/backend.bucket: my-gcs-state-bucket
    tofu.openmcf.org/backend.key: gke/staging-cluster
spec:
  projectId: my-gcp-project
  region: us-central1
```

**Required labels:**
- `backend.type`: `gcs`
- `backend.bucket`: GCS bucket name
- `backend.key`: State prefix/path within the bucket

**Prerequisites:**
- GCS bucket must exist
- IAM permissions: `storage.objects.get`, `storage.objects.create`, `storage.objects.delete`, `storage.objects.list`

---

### Azure Storage

Store state in Azure Blob Storage.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureAksCluster
metadata:
  name: production-aks
  labels:
    openmcf.org/provisioner: terraform
    terraform.openmcf.org/backend.type: azurerm
    terraform.openmcf.org/backend.bucket: tfstate-container
    terraform.openmcf.org/backend.key: aks/production
spec:
  location: eastus
  nodeCount: 3
```

**Required labels:**
- `backend.type`: `azurerm`
- `backend.bucket`: Azure container name
- `backend.key`: State file path within the container

**Prerequisites:**
- Storage account and container must exist
- Storage account name configured via environment
- IAM permissions: Storage Blob Data Contributor

---

### Local Backend

Store state on the local filesystem. **Not recommended for production or team use.**

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesDeployment
metadata:
  name: test-service
  labels:
    openmcf.org/provisioner: tofu
    tofu.openmcf.org/backend.type: local
    tofu.openmcf.org/backend.key: /tmp/test-service.tfstate
spec:
  replicas: 1
```

**Required labels:**
- `backend.type`: `local`
- `backend.key`: Local file path for state

**Use cases:**
- Local development
- Testing
- Single-user scenarios

**Limitations:**
- No state locking
- Not shareable between team members
- Lost if machine is wiped

---

## Complete Examples

### Example 1: AWS Infrastructure with S3 Backend

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRdsInstance
metadata:
  name: app-database
  labels:
    # Provisioner selection
    openmcf.org/provisioner: terraform
    
    # Backend configuration
    terraform.openmcf.org/backend.type: s3
    terraform.openmcf.org/backend.bucket: company-terraform-state
    terraform.openmcf.org/backend.key: rds/app-database/production.tfstate
    terraform.openmcf.org/backend.region: us-west-2
spec:
  engine: postgres
  engineVersion: "15"
  instanceClass: db.t3.medium
  allocatedStorage: 100
  region: us-west-2
```

**Deploy:**
```bash
openmcf apply -f database.yaml
```

---

### Example 2: GCP Infrastructure with GCS Backend (OpenTofu)

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudRun
metadata:
  name: api-service
  labels:
    # Provisioner selection
    openmcf.org/provisioner: tofu
    
    # Backend configuration (OpenTofu-specific)
    tofu.openmcf.org/backend.type: gcs
    tofu.openmcf.org/backend.bucket: company-tofu-state
    tofu.openmcf.org/backend.key: cloud-run/api-service/prod
spec:
  projectId: my-gcp-project
  region: us-central1
  image: gcr.io/my-project/api:latest
```

**Deploy:**
```bash
openmcf apply -f api-service.yaml
```

---

### Example 3: Multi-Environment with Kustomize

**Base manifest** (`base/database.yaml`):
```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRdsInstance
metadata:
  name: app-database
  labels:
    openmcf.org/provisioner: terraform
    terraform.openmcf.org/backend.type: s3
    terraform.openmcf.org/backend.bucket: company-terraform-state
    terraform.openmcf.org/backend.region: us-west-2
    # Key will be patched per environment
spec:
  engine: postgres
  instanceClass: db.t3.small
```

**Production overlay** (`overlays/prod/kustomization.yaml`):
```yaml
patches:
  - patch: |
      - op: add
        path: /metadata/labels/terraform.openmcf.org~1backend.key
        value: rds/app-database/production.tfstate
      - op: replace
        path: /spec/instanceClass
        value: db.t3.large
    target:
      kind: AwsRdsInstance
```

**Deploy:**
```bash
openmcf apply --kustomize-dir . --overlay prod
```

---

### Example 4: Pulumi with Stack Name

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPostgres
metadata:
  name: analytics-db
  labels:
    # Provisioner selection
    openmcf.org/provisioner: pulumi
    
    # Stack name for state identification
    pulumi.openmcf.org/stack.name: production.KubernetesPostgres.analytics-db
spec:
  container:
    replicas: 3
    resources:
      limits:
        cpu: 2000m
        memory: 8Gi
```

**Deploy:**
```bash
# Ensure you're logged into Pulumi
pulumi login

# Deploy
openmcf apply -f analytics-db.yaml
```

---

## Best Practices

### 1. Use Consistent Naming

Establish a naming convention for state paths:

```
<bucket>/<resource-type>/<resource-name>/<environment>
```

Example: `terraform-state/rds/app-database/production`

### 2. Separate State by Environment

Use different buckets or prefixes for different environments:

```yaml
# Production
terraform.openmcf.org/backend.bucket: prod-terraform-state
terraform.openmcf.org/backend.key: vpc/main.tfstate

# Staging
terraform.openmcf.org/backend.bucket: staging-terraform-state
terraform.openmcf.org/backend.key: vpc/main.tfstate

# Development
terraform.openmcf.org/backend.bucket: dev-terraform-state
terraform.openmcf.org/backend.key: vpc/main.tfstate
```

### 3. Enable Versioning

Always enable versioning on your state bucket:

**S3:**
```bash
aws s3api put-bucket-versioning \
  --bucket my-terraform-state \
  --versioning-configuration Status=Enabled
```

**GCS:**
```bash
gsutil versioning set on gs://my-terraform-state
```

### 4. Enable Encryption

Encrypt state at rest:

**S3:**
```bash
aws s3api put-bucket-encryption \
  --bucket my-terraform-state \
  --server-side-encryption-configuration '{
    "Rules": [{"ApplyServerSideEncryptionByDefault": {"SSEAlgorithm": "AES256"}}]
  }'
```

### 5. Restrict Access

Implement least-privilege access to state files:

- Use IAM roles/policies
- Enable bucket logging
- Consider using separate accounts for production state

### 6. Use State Locking

For S3, configure DynamoDB for state locking:

```bash
aws dynamodb create-table \
  --table-name terraform-state-lock \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST
```

---

## Troubleshooting

### "Backend configuration required"

**Error:** Backend type is specified but required labels are missing.

**Solution:** For remote backends, all required labels must be specified:

```yaml
labels:
  # S3 backend - all four labels required
  terraform.openmcf.org/backend.type: s3
  terraform.openmcf.org/backend.bucket: my-terraform-state
  terraform.openmcf.org/backend.key: path/to/state.tfstate
  terraform.openmcf.org/backend.region: us-west-2
```

For GCS or Azure backends, `backend.region` is not required.

---

### "Access Denied" to State Bucket

**Error:** Permission denied when accessing state.

**Solutions:**
1. Verify IAM permissions on the bucket
2. Check credential configuration
3. Ensure bucket exists in the expected region
4. Verify bucket policy allows access

---

### State Lock Timeout

**Error:** Unable to acquire state lock.

**Solutions:**
1. Check if another operation is running
2. Force unlock if previous operation crashed:
   ```bash
   terraform force-unlock <LOCK_ID>
   ```
3. Verify DynamoDB table permissions (for S3 backend)

---

### Wrong Provisioner Labels

**Error:** Backend not detected, using local backend.

**Solution:** Ensure labels match the provisioner:

```yaml
# For Terraform (S3)
openmcf.org/provisioner: terraform
terraform.openmcf.org/backend.type: s3
terraform.openmcf.org/backend.bucket: my-state-bucket
terraform.openmcf.org/backend.key: path/to/state.tfstate
terraform.openmcf.org/backend.region: us-west-2

# For OpenTofu (GCS)
openmcf.org/provisioner: tofu
tofu.openmcf.org/backend.type: gcs
tofu.openmcf.org/backend.bucket: my-state-bucket
tofu.openmcf.org/backend.key: path/to/state
```

### "Backend configuration changed"

**Error:** Terraform/Tofu prompts for backend reconfiguration.

**Solution:** Use the `--reconfigure` flag to accept the new backend configuration:

```bash
openmcf init -f manifest.yaml --reconfigure

# Or with other commands that run init internally
openmcf apply -f manifest.yaml --reconfigure
```

---

### S3-Compatible Backend: "InvalidClientTokenId"

**Error:** `error calling sts:GetCallerIdentity: InvalidClientTokenId`

**Cause:** This error occurs when using S3-compatible backends (R2, MinIO) without the proper endpoint configuration. Terraform tries to validate credentials via AWS STS, which fails with non-AWS credentials.

**Solution:** Set `region` to `auto` and provide the `endpoint`:

```yaml
metadata:
  labels:
    terraform.openmcf.org/backend.region: auto
    terraform.openmcf.org/backend.endpoint: https://<account-id>.r2.cloudflarestorage.com
```

Or via CLI:
```bash
openmcf apply -f manifest.yaml \
    --backend-region auto \
    --backend-endpoint https://<account-id>.r2.cloudflarestorage.com \
    --reconfigure
```

---

### Incomplete Backend Configuration (Interactive Prompt)

**Behavior:** When required backend configuration is missing, OpenMCF prompts for values interactively:

```
ℹ  S3-Compatible Backend Detected
   Region is set to 'auto', indicating an S3-compatible backend

✗  Incomplete Backend Configuration

   S3 backend requires the following configuration:

   • Custom S3-compatible endpoint (required when region is 'auto')
     Flag:    --backend-endpoint
     Example: https://<account-id>.r2.cloudflarestorage.com

Enter endpoint: _
```

**Solution:** Either:
1. Provide values at the prompt
2. Add labels to your manifest
3. Use CLI flags: `--backend-endpoint https://...`

**In CI/CD (non-interactive):** The CLI will exit with an error and show exactly which flags or labels are needed.

---

## What's Next

- [State Management](../concepts/state-management) — Conceptual overview of state in OpenMCF
- [Unified Commands](/docs/cli/unified-commands) — Using apply, destroy, init, plan, refresh
- [Credentials](./credentials) — Setting up cloud provider credentials
- [Kustomize Integration](./kustomize) — Multi-environment deployments with per-environment state
- [CLI Reference](/docs/cli/cli-reference) — Complete command and flag reference
