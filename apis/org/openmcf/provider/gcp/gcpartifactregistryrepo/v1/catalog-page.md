# GCP Artifact Registry Repo

Deploys a Google Cloud Artifact Registry repository with configurable format, region, and access control. The component creates the repository along with dedicated reader and writer GCP service accounts with appropriate IAM bindings, and optionally enables unauthenticated public access for open-source artifact distribution.

## What Gets Created

When you deploy a GcpArtifactRegistryRepo resource, OpenMCF provisions:

- **Artifact Registry Repository** — a repository in the specified GCP project and region, with the chosen format (Docker, Maven, NPM, Python, Go, etc.) and GCP labels applied
- **Reader Service Account** — a GCP service account with `roles/artifactregistry.reader` on the repository (for private repos) and a JSON key exported as a base64-encoded secret
- **Writer Service Account** — a GCP service account with both `roles/artifactregistry.writer` and `roles/artifactregistry.repoAdmin` on the repository, and a JSON key exported as a base64-encoded secret
- **IAM Bindings** — repository-level IAM members granting the reader and writer service accounts their respective roles
- **Public Access IAM Binding** (conditional) — when `enablePublicAccess` is `true`, an `allUsers` binding with `roles/artifactregistry.reader` replaces the reader service account binding, making the repository publicly readable

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **An existing GCP project** — referenced via `projectId`
- **IAM permissions** to create Artifact Registry repositories and service accounts in the target project
- **Artifact Registry API** enabled on the GCP project (`artifactregistry.googleapis.com`)

## Quick Start

Create a file `artifact-registry.yaml`:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpArtifactRegistryRepo
metadata:
  name: my-docker-repo
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpArtifactRegistryRepo.my-docker-repo
spec:
  repoFormat: DOCKER
  projectId:
    value: my-gcp-project-123
  region: us-central1
```

Deploy:

```shell
openmcf apply -f artifact-registry.yaml
```

This creates a private Docker repository in `us-central1` with dedicated reader and writer service accounts.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `repoFormat` | `GcpArtifactRegistryRepoFormat` | Format of the repository. One of: `DOCKER`, `GENERIC`, `GO`, `KUBEFLOW`, `MAVEN`, `NPM`, `PYTHON`, `YUM`. | Required |
| `projectId` | `StringValueOrRef` | GCP project ID where the repository will be created. Can be a literal value or a reference to a GcpProject resource via `valueFrom`. | Required |
| `region` | `string` | GCP region for the repository (e.g., `us-central1`, `us-west2`). Choosing a region close to your workloads reduces image pull latency. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enablePublicAccess` | `bool` | `false` | When `true`, grants `allUsers` the `roles/artifactregistry.reader` role on the repository, making artifacts publicly readable. Useful for open-source projects. |

## Examples

### Private Docker Repository

A private Docker repository with dedicated service accounts for CI/CD integration:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpArtifactRegistryRepo
metadata:
  name: backend-images
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpArtifactRegistryRepo.backend-images
spec:
  repoFormat: DOCKER
  projectId:
    value: my-gcp-project-123
  region: us-central1
```

### Public NPM Registry for Open-Source Packages

A publicly readable NPM repository for distributing open-source packages:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpArtifactRegistryRepo
metadata:
  name: oss-npm-packages
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpArtifactRegistryRepo.oss-npm-packages
spec:
  repoFormat: NPM
  projectId:
    value: my-oss-project
  region: us-east1
  enablePublicAccess: true
```

### Maven Repository with Foreign Key Reference

A private Maven repository that references an OpenMCF-managed GcpProject instead of hardcoding the project ID:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpArtifactRegistryRepo
metadata:
  name: java-artifacts
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpArtifactRegistryRepo.java-artifacts
spec:
  repoFormat: MAVEN
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      field: status.outputs.project_id
  region: europe-west1
```

### Python Repository in a Shared Services Project

A private Python repository for internal library distribution, deployed to a region close to your compute workloads:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpArtifactRegistryRepo
metadata:
  name: internal-pypi
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: shared-services
    pulumi.openmcf.org/stack.name: prod.GcpArtifactRegistryRepo.internal-pypi
spec:
  repoFormat: PYTHON
  projectId:
    value: shared-services-project
  region: asia-southeast1
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `readerServiceAccount.email` | `string` | Email address of the reader service account |
| `readerServiceAccount.keyBase64` | `string` | Base64-encoded JSON key for the reader service account |
| `writerServiceAccount.email` | `string` | Email address of the writer service account |
| `writerServiceAccount.keyBase64` | `string` | Base64-encoded JSON key for the writer service account |
| `repoName` | `string` | Name of the created Artifact Registry repository |
| `hostname` | `string` | Regional hostname of the repository (e.g., `us-central1-docker.pkg.dev`) |
| `repoUrl` | `string` | Full URL of the repository (e.g., `us-central1-docker.pkg.dev/my-project/my-repo`) |

## Related Components

- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project where the repository is created
- [GcpGkeCluster](/docs/catalog/gcp/gcpgkecluster) — Kubernetes clusters that pull container images from the repository
- [GcpGkeNodePool](/docs/catalog/gcp/gcpgkenodepool) — node pools that can use the reader service account for image pulls
- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — additional service accounts that may need access to the repository
