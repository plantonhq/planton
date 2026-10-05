# GCP IAM Custom Role

Defines a project-scoped IAM custom role — a named, reusable bundle of permissions tailored to exactly what a workload or team needs, instead of granting one of Google's broad predefined roles. The role is the least-privilege building block: define the bundle once, grant it with GcpProjectIamMember or GcpServiceAccountIamMember (both reference this role's `name` output), and every permission edit propagates to every grant immediately. The `roleId` and project are immutable — changing either destroys and recreates the role, breaking every grant that references the old name.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Project IAM Custom Role** -- a `projects.IAMCustomRole` in the specified GCP project with the given role ID, title, description, permission list, and launch stage

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** that will own the role. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef. The role can only be granted on resources within this project.
- **IAM API** (`iam.googleapis.com`) enabled in the target project.

## Deploy

### Console

Open the deployment store, find **GCP IAM Custom Role**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Workload Least-Privilege Role** preset in the [Presets](#presets) tab to pre-populate a minimal permission bundle.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpIamCustomRole
metadata:
  name: log-bucket-writer
  org: acme-corp
  env: prod
spec:
  roleId: logBucketWriter
  projectId:
    value: "acme-prod-12345"
  title: Log Bucket Writer
  description: Grants exactly the permissions this workload needs, nothing more
  permissions:
    - storage.objects.create
    - storage.objects.get
```

```shell
planton apply -f gcp-iam-custom-role.yaml
```

This mints a two-permission role. Grant it by wiring a GcpProjectIamMember's `role` field to this resource's `name` output. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the role to a GCP project deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
```

The InfraPipeline resolves the dependency graph, deploys the project first, then provisions the role with the resolved project ID.

## Key Configuration

These are the most important decisions when configuring a custom role. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Role ID** -- `roleId` forms the grantable name `projects/<project>/roles/<role_id>`. 3-64 characters of letters, digits, underscores, and periods -- hyphens are NOT allowed (camelCase is the convention). Immutable: a changed ID is a NEW role name, and every grant referencing the old one breaks.

**Permissions** -- the bundle IS the role. Each entry follows `<service>.<resource>.<verb>` (e.g. `storage.objects.get`); at least one is required. Mutable: adding or removing permissions updates the role in place, and every existing grant reflects the change immediately -- that is the point of defining the bundle once. Discover valid values with `gcloud iam list-testable-permissions`; some permissions are NOT_SUPPORTED in custom roles and rejected at deploy.

**Launch stage** -- `stage` is a lifecycle label (leave it unset and GCP applies GA, the right label for production). DISABLED is the reversible IAM kill switch: the role stays defined while every grant of it is rejected.

**Soft-delete semantics** -- GCP retains deleted custom roles (and reserves their IDs) for up to 14 days. Creating a role with a soft-deleted ID undeletes and updates it in place; the `deleted` output reports the state. Because deletion is soft, `deletionPolicy: DELETE` and `ABANDON` differ less here than on hard-delete resources — ABANDON keeps every binding working, while DELETE makes bindings stop granting access; `PREVENT` protects a role that live bindings depend on.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | Fully-qualified role name (`projects/<project>/roles/<role_id>`) | GcpProjectIamMember / GcpServiceAccountIamMember `role` fields -- the grantable handle |
| `role_id` | The bare role ID within the project | Tooling that addresses the role by short ID |
| `deleted` | Whether the role is currently soft-deleted (14-day window) | Health dashboards, undelete runbooks |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Workload least privilege** -- A minimal bundle scoped to exactly what one workload calls (e.g. write-only log sink access). Start from the **Workload Least-Privilege Role** preset.

**Read-only auditor** -- The `*.get` / `*.list` / `*.getIamPolicy` family for audit dashboards and access reviews, with no mutating verbs. Start from the **Read-Only Auditor Role** preset.

**CI/CD deployer** -- Deploys new Cloud Run revisions from CI/CD without broader project access. Start from the **CI/CD Deployer Role** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project that owns the role
- [**GCP Project IAM Member**](/infra-catalog/gcp-project-iam-member) -- grants this role at project scope by referencing the `name` output
- [**GCP Service Account IAM Member**](/infra-catalog/gcp-service-account-iam-member) -- grants this role ON a service account by referencing the `name` output
- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- its inline `projectIamRoles` list also accepts this role's fully-qualified name
