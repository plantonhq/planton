# Terraform Module for GCP Artifact Registry Repository

This Terraform module deploys a Google Cloud Artifact Registry repository with automated service account management for read/write access.

## Overview

This module creates:
- **Artifact Registry Repository** - Supports multiple formats (Docker, Maven, NPM, Python, Go, etc.)
- **Reader Service Account** - For pulling artifacts (public repos) or dedicated read access
- **Writer Service Account** - For pushing artifacts and repository administration
- **IAM Bindings** - Automatic permission assignments based on public/private configuration

## Features

- Multi-format repository support (Docker, Maven, NPM, Python, Go, Generic, YUM, Kubeflow)
- Automatic service account creation with proper IAM roles
- Public or private access configuration
- Regional deployment for optimal performance and cost
- Complete outputs for CI/CD integration

## Usage

### Initialize Terraform State Backend

```shell
planton tofu init --manifest hack/manifest.yaml --backend-type s3 \
  --backend-config="bucket=planton-tf-state-backend" \
  --backend-config="dynamodb_table=planton-tf-state-backend-lock" \
  --backend-config="region=ap-south-2" \
  --backend-config="key=planton/gcp-stacks/test-gcp-artifact-registry.tfstate"
```

### Plan Changes

```shell
planton tofu plan --manifest hack/manifest.yaml
```

### Apply Configuration

```shell
planton tofu apply --manifest hack/manifest.yaml --auto-approve
```

### Destroy Resources

**Warning:** This will delete the repository and all artifacts permanently.

```shell
planton tofu destroy --manifest hack/manifest.yaml --auto-approve
```

## Module Inputs

The module reads configuration from a YAML manifest with the following structure:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpArtifactRegistryRepo
metadata:
  name: my-docker-registry
spec:
  repoFormat: DOCKER
  projectId: my-gcp-project-id
  region: us-central1
  enablePublicAccess: false
```

See `hack/manifest.yaml` for a complete example.

## Module Outputs

After successful deployment, the module provides:

- `repo_name` - Repository identifier
- `repo_hostname` - Repository hostname (e.g., `us-central1-docker.pkg.dev`)
- `repo_url` - Full repository URL for pushing/pulling artifacts
- `reader_service_account_email` - Email of the reader service account
- `reader_service_account_key_base64` - Base64-encoded private key (sensitive)
- `writer_service_account_email` - Email of the writer service account  
- `writer_service_account_key_base64` - Base64-encoded private key (sensitive)

## Service Account Usage

### Reader Service Account

Use for:
- **CI/CD pipelines** pulling artifacts
- **GKE workloads** reading container images
- **Cloud Run services** accessing images

Permissions: `roles/artifactregistry.reader`

### Writer Service Account  

Use for:
- **CI/CD pipelines** pushing artifacts
- **Build systems** uploading packages
- **Repository management** (cleanup, configuration)

Permissions: `roles/artifactregistry.writer` and `roles/artifactregistry.repoAdmin`

## Architecture

```
GCP Artifact Registry Repository
│
├─ Repository Resource
│  ├─ Format: DOCKER/MAVEN/NPM/etc
│  ├─ Region: us-central1 (configurable)
│  └─ Labels: Auto-generated from metadata
│
├─ IAM Bindings (Public)
│  └─ allUsers → roles/artifactregistry.reader
│
├─ IAM Bindings (Private)
│  ├─ Reader SA → roles/artifactregistry.reader
│  └─ Writer SA → roles/artifactregistry.writer + repoAdmin
│
└─ Service Accounts
   ├─ {name}-{suffix}-ro (Reader)
   └─ {name}-{suffix}-rw (Writer)
```

## Best Practices

1. **Co-locate with consumers** - Deploy in the same region as GKE/Cloud Run for free egress
2. **Use descriptive names** - Include format and purpose in metadata.name
3. **Secure credentials** - Store service account keys in secrets management (Vault, GCP Secret Manager)
4. **Prefer Workload Identity** - For GKE workloads, use Workload Identity instead of keys
5. **Enable cleanup policies** - (Future enhancement) Configure retention to manage costs

## Troubleshooting

### Permission Denied Errors

Ensure the Terraform service account has:
- `roles/artifactregistry.admin` on the project
- `roles/iam.serviceAccountAdmin` for service account creation
- `roles/iam.serviceAccountKeyAdmin` for key creation

### Repository Already Exists

If you get "already exists" errors, either:
1. Import existing repository: `terraform import google_artifact_registry_repository.repo projects/{project}/locations/{location}/repositories/{name}`
2. Use a different `metadata.name` in your manifest

## Additional Resources

- [GCP Artifact Registry Documentation](https://cloud.google.com/artifact-registry/docs)
- [Terraform GCP Provider - Artifact Registry](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/artifact_registry_repository)
- [Planton CLI Documentation](https://planton.dev)
