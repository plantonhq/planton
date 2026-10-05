# GCP Service Account

Deploys a GCP service account with optional JSON key generation and configurable IAM role bindings at both the project and organization levels. The service account can be used for GKE Workload Identity (keyless), CI/CD pipelines (with exported key), or any workload that needs programmatic access to GCP APIs. Keyless is the default posture: no key material exists unless the `userManagedKey` block explicitly asks for it, and the account's `member` output feeds first-class IAM grant resources directly.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Service Account** -- a `serviceaccount.Account` in the specified GCP project with the given account ID and display name
- **Service Account Key** -- created only when the `userManagedKey` block is present; either generates a private key (exported base64) shaped by algorithm/format fields, or registers your own uploaded public key
- **Project IAM Bindings** -- one `projects.IAMMember` per entry in `projectIamRoles`, granting the specified role to the service account in the target project
- **Organization IAM Bindings** -- created only when `orgIamRoles` is non-empty and `orgId` is set; one `organizations.IAMMember` per entry, granting the specified role at the organization level

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the service account will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **IAM API** (`iam.googleapis.com`) enabled in the target project.
- **Organization ID** (if using `orgIamRoles`) -- the numeric GCP organization ID for organization-level role bindings.

## Deploy

### Console

Open the deployment store, find **GCP Service Account**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Workload Identity Service Account** preset in the [Presets](#presets) tab to pre-populate a keyless service account for GKE workloads.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpServiceAccount
metadata:
  name: app-workload
  org: acme-corp
  env: prod
spec:
  serviceAccountId: app-workload-prod
  projectId:
    value: "acme-prod-12345"
  projectIamRoles:
    - roles/logging.logWriter
    - roles/monitoring.metricWriter
```

```shell
planton apply -f gcp-service-account.yaml
```

This creates a service account with logging and monitoring permissions, no JSON key. Pair with GKE Workload Identity for keyless pod authentication. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the service account to a GCP project deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
```

The InfraPipeline resolves the dependency graph, deploys the project first, then provisions the service account with the resolved project ID.

## Key Configuration

These are the most important decisions when configuring a service account. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Key management** -- Add a `userManagedKey` block to create a key for CI/CD pipelines or external systems; omit it (the default) for Workload Identity or other keyless authentication patterns. An empty block generates the classic 2048-bit RSA JSON key; `keepers` gives declarative rotation (bump a value, the key is replaced); `publicKeyData` uploads your own public key so the private half never leaves your custody; `deletionPolicy: PREVENT` guards against accidental destroys. Exported keys are a security risk -- prefer keyless authentication when possible.

**Project IAM roles** -- `projectIamRoles` lists roles granted at the project level: predefined roles (e.g., `roles/logging.logWriter`) or a custom role's fully-qualified name (`projects/<project>/roles/<role_id>` -- a GcpIamCustomRole's `name` output is exactly this value). Follow the least-privilege principle -- grant only the roles the workload needs. Each role creates a separate IAM member binding. For grants other teams manage independently (or grants needing IAM Conditions), compose GcpProjectIamMember resources against this account's `member` output instead.

**Organization IAM roles** -- `orgIamRoles` lists roles granted at the GCP organization level. Requires `orgId` to be set. Use sparingly -- organization-level roles are broad and affect all projects in the organization.

**Service account ID** -- `serviceAccountId` is the short ID (6-30 characters) used to form the email address (`{id}@{project}.iam.gserviceaccount.com`). Choose a descriptive, environment-specific ID to avoid collisions across environments sharing a project. Immutable -- recreating an account invalidates every grant that referenced it (the new account gets a different unique ID).

**Console profile** -- `displayName` (what GCP console IAM lists show; falls back to the resource name) and `description` (the note an auditor reads) stay editable after creation.

**Lifecycle kill switch** -- `disabled: true` keeps every grant but rejects all authentication -- the instant, reversible first response to a suspected credential leak. Re-enabling restores everything; no grant is touched at any point.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `email` | Service account email address (`{id}@{project}.iam.gserviceaccount.com`) | Workload Identity bindings, workload attachment, application configuration |
| `member` | IAM member string (`serviceAccount:{email}`) | Feeds directly into GcpProjectIamMember / GcpServiceAccountIamMember `member` fields |
| `unique_id` | Stable numeric ID, never reused across delete/recreate | Audit log correlation, principal pinning |
| `name` | Full resource name (`projects/{project}/serviceAccounts/{email}`) | A GcpServiceAccountIamMember's `service_account_id` -- grants ON this account |
| `key_base64` | Base64-encoded private key (generate flow only; empty for keyless and uploaded-key accounts) | CI/CD secret stores, external system authentication |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Workload Identity** -- Keyless service account for GKE pods, with logging, monitoring, and secret access roles. No JSON key is generated -- pods authenticate via KSA-to-GSA binding. The recommended pattern for runtime workloads. Start from the **Workload Identity Service Account** preset.

**CI/CD pipeline** -- Service account with a generated JSON key for external CI/CD systems (GitHub Actions, GitLab CI). Includes Artifact Registry, GKE, and Cloud Run deployment roles. Start from the **CI/CD Pipeline Service Account** preset.

**Identity with first-class grants** -- A pure identity with no inline role lists; access arrives via separate GcpProjectIamMember nodes wired to the `member` output, so each grant is independently owned, reviewed, and removable. Start from the **Pure Identity with First-Class Grants** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the service account is created
- [**GCP Project IAM Member**](/infra-catalog/gcp-project-iam-member) -- consumes the `member` output to grant this account project-scope roles as first-class resources
- [**GCP Service Account IAM Member**](/infra-catalog/gcp-service-account-iam-member) -- consumes the `name` output to control who may impersonate or act as this account
- [**GCP IAM Custom Role**](/infra-catalog/gcp-iam-custom-role) -- its `name` output slots into `projectIamRoles` for least-privilege custom bundles