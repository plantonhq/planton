# Overview

The **GcpServiceAccount** API resource provides an easy, opinionated way to create and manage Google Cloud projects under your
organization or folder. By focusing on essential configurations (such as project billing, labels, and pre-enabling
APIs), it streamlines the process of project creation within the Planton multi-cloud deployment framework.

## Purpose

Provisioning a GCP project often involves juggling multiple steps: setting up a unique project ID, adding labels for
cost
allocation, linking a billing account, assigning it to the correct folder or organization, disabling the default
network,
and enabling various Cloud APIs. The **GcpServiceAccount** resource aims to consolidate these tasks by:

- **Automating GCP Service Account Creation**: Provide a single, consistent interface for spinning up new GCP projects.
- **Centralizing Common Configurations**: Handle billing account linkage, default network removal, and label assignments
  without requiring repetitive manual steps.
- **Enforcing Best Practices**: Offer validated fields for project ID naming, label constraints, and correct API service
  format (e.g., `compute.googleapis.com`).

## Key Features

### Simplified Project Setup

- **One Resource, One Manifest**: Create a project with a unique project ID, display name, and optional folder or
  organization structure.
- **Billing Account Attachment**: Automatically attach your project to a billing account for immediate usage of GCP
  services.

### Flexible Parent Handling

- **Organization or Folder**: Choose whether to attach the project directly under an organization or under a specific
  folder. Exactly one of `orgId` or `folderId` must be provided.

### Labels and Governance

- **Cost Allocation Labels**: Apply key-value labels for cost tracking or compliance. These labels help categorize
  projects for cost visibility and resource grouping.
- **Disabling Default Network**: Optionally remove the default VPC network immediately after project creation, a common
  security best practice.

### Pre-Enabled APIs

- **Batch Enablement**: Enable a list of required Google APIs (e.g., `compute.googleapis.com`,
  `cloudrun.googleapis.com`)
  at creation time, ensuring your project is fully functional from the start.

### Multi-Cloud Alignment

- **Planton CLI**: Deploy and manage your new GCP project alongside other cloud resources using either Pulumi or
  Terraform under the hood.
- **Consistent Resource Model**: Rely on the same structure (apiVersion, kind, metadata, spec) that Planton
  provides across all providers.

## Benefits

- **Unified Provisioning**: Reduce the complexity of creating GCP projects by using the same framework that handles
  resources in AWS, Kubernetes, or other clouds.
- **Repeatable & Validated**: Leverage built-in Protobuf validations that catch invalid project IDs, API formats, or
  missing parents (folder/organization) before deployment.
- **Security & Governance**: Adopt recommended best practices like removing the default network and enforcing billing
  account usage from the beginning.
- **Scalable Structure**: Expand your infrastructure with additional GCP resources—shared VPC networks, Cloud Storage
  buckets, GKE clusters—on top of the newly created project, within the same Planton manifest.

## Example Usage

Below is a minimal YAML snippet demonstrating how to create a GCP project using Planton. Notice that all fields
are
in camelCase for consistency.

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpServiceAccount
metadata:
  name: my-gcp-service-account
  version:
    message: "Initial project creation"
spec:
  projectId: "example-project-123"
  name: "Example Project"
  orgId: "1234567890123"
  # folderId: "4567890123456" # Uncomment if you prefer folder-based hierarchy instead of orgId
  billingAccountId: "0123AB-4567CD-89EFGH"
  labels:
    environment: "dev"
    owner: "team-infra"
  disableDefaultNetwork: true
  enabledApis:
    - compute.googleapis.com
    - storage.googleapis.com
  ownerMember: "user:alice@example.com"
```

> **Note**: Exactly one of `orgId` or `folderId` must be set. If `orgId` is provided, leave `folderId` empty, and vice
> versa.

### Deploying with Planton

Once your YAML manifest is ready, you can deploy using Planton's CLI. Planton will validate the manifest
against the Protobuf schema and handle provisioning via Pulumi or Terraform under the hood.

- **Using Pulumi**:
  ```bash
  planton pulumi up --manifest gcp_service_account.yaml --stack org/project/my-gcp-stack
  ```

- **Using Terraform**:
  ```bash
  planton terraform apply --manifest gcp_service_account.yaml --stack org/project/my-gcp-stack
  ```

The CLI automatically configures the corresponding IaC modules, ensuring your new Google Cloud project is created with
the specified parent, billing configuration, default network settings, and any pre-enabled APIs.

---

Happy deploying! If you have questions or encounter any issues, feel free to open an issue on our GitHub repository or
reach out through our community channels.
