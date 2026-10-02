# GCP Project

Deploys a Google Cloud project — the Layer-0 container every other GCP resource lives in. Configures resource hierarchy placement (organization, folder, or standalone), billing account attachment, labels and resource-manager tags, the default-network posture, pre-enabled Cloud APIs, and the deletion policy. The `projectId` is globally unique across all of GCP and immutable — the one naming decision this page exists to get right.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **GCP Project** -- a new project with the immutable `projectId`, the mutable `displayName` (defaults to the resource name), and the configured hierarchy parent
- **Hierarchy Placement** -- created under the organization or folder named by `parentType` + `parentId`, or inside a folder declared in the same chart by referencing a `GcpFolder` in `folderId`; standalone (no parent) is supported for accounts without a GCP Organization
- **Billing Link** -- when `billingAccountId` is set, the project links to that billing account (the deploying identity needs `roles/billing.user` on it)
- **Default Network Suppression** -- unless `autoCreateNetwork` is true, the auto-created "default" VPC with its permissive firewall rules never persists — the standard security-hardening posture
- **Cloud API Enablement** -- each entry in `enabledApis` is activated as a project service (e.g., `compute.googleapis.com`); component kinds also enable the APIs they need at their own deploy time
- **Resource Manager Tags** -- entries in `tags` (`tagKeys/{id}` → `tagValues/{id}`) bind at CREATE TIME for org-policy and IAM-condition targeting; changing them later recreates the project
- **Deletion Policy** -- `deletionPolicy` is GCP's real three-way destroy switch: `DELETE` (default, 30-day restore window), `PREVENT` (destroy fails — foundation protection), or `ABANDON` (unmanage without touching GCP)
- **GCP Labels** -- your `labels` merge beneath Planton's attribution labels (platform keys win on conflicts); project labels are the primary cost-allocation dimension in billing exports

IAM grants are deliberately NOT part of this component — model each grant as a first-class `GcpProjectIamMember` resource.

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials that have permission to create projects under the target organization or folder. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Organization

- **An organization or folder** in the GCP resource hierarchy (unless creating a standalone project on an organization-less account). Provide the NUMERIC organization or folder ID in `parentId` and set `parentType` accordingly -- or, for a folder the chart also declares, reference the `GcpFolder` in `folderId` and leave both empty.
- **A billing account** in the format `0123AB-4567CD-89EFGH`. The deploying identity needs `roles/billing.user` on the account to link it.

## Deploy

### Console

Open the deployment store, find **GCP Project**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Production Project** preset in the [Presets](#presets) tab to pre-populate a production-ready project with essential APIs enabled and destroy blocked by `deletionPolicy: PREVENT`.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpProject
metadata:
  name: platform-prod
  org: acme-corp
  env: prod
spec:
  projectId: "acme-platform-prod"
  parentType: folder
  parentId: "123456789012"
  billingAccountId: "0123AB-4567CD-89EFGH"
  deletionPolicy: PREVENT
  enabledApis:
    - "compute.googleapis.com"
    - "container.googleapis.com"
```

```shell
planton apply -f gcp-project.yaml
```

This creates a project under the specified folder with compute and container APIs enabled, no auto-created default network, and destroy blocked while `PREVENT` is set. A Stack Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a GCP project. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Project ID immutability** -- `projectId` is globally unique across ALL of GCP and can never change; deleted IDs stay reserved for up to 30 days. Encode org, workload, and environment (e.g. `acme-payments-prod`) so IDs stay readable in billing exports and collisions stay improbable.

**Resource hierarchy placement** -- Set `parentType` to `organization` or `folder` and provide the numeric ID in `parentId`, or reference a `GcpFolder` declared in the same chart with `folderId` (its `folder_id` output; `parentType` and `parentId` stay empty). Folder-based placement is the recommended production practice — folders carry environment- and team-scoped org policies and IAM. Changing the parent later migrates the project. Accounts without a GCP Organization omit the parent entirely.

**Default network posture** -- Leave `autoCreateNetwork` unset (or false): the auto-created VPC ships permissive allow-internal/allow-ssh firewall rules in every region, and suppressing it is a standard hardening step. Model networks as explicit `GcpVpcNetwork` resources instead. Note the project still needs one free network slot of quota during creation.

**API pre-enablement** -- The `enabledApis` list activates Cloud APIs at creation time; component kinds enable the APIs they need on their own, so this is a pre-warming convenience. Include `cloudresourcemanager.googleapis.com` and `serviceusage.googleapis.com` so IaC tooling can manage the project and enable further APIs; add `servicenetworking.googleapis.com` when Cloud SQL or Memorystore private IP is planned.

**Deletion policy** -- Set `deletionPolicy: PREVENT` for foundation projects whose accidental destruction would be catastrophic. Use `ABANDON` to hand a project off to another owner or tool: destroy removes it from Planton's state and the project lives on unmanaged.

**Create-time tags** -- `tags` bind resource-manager tag values at creation only — the levers org policies and IAM conditions target. Changing this map after creation RECREATES the project; bind additional tag values to an existing project out-of-band.

## Outputs and Dependencies

### What This Component Consumes

This component has no foreign key dependencies — it is the root of the GCP composition graph.

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `project_id` | Immutable project ID | The `projectId` field of GcpVpcNetwork, GcpGkeCluster, GcpCloudSql, GcpGcsBucket, and every other GCP resource |
| `project_number` | Numeric project number assigned by Google | Workload Identity pool providers, service-agent emails, IAM principal identifiers |
| `name` | Display name of the project | Monitoring dashboards, documentation |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard production project** -- Folder-based hierarchy, billing linked, a hardening baseline of 8 APIs pre-enabled, and destroy blocked by `deletionPolicy: PREVENT`. Start from the **Standard Production Project** preset.

**Development project** -- Lightweight project with billing linked, a minimal 3-API set, and the default `DELETE` policy so teardown is one command. Bake uniqueness into the ID itself (a team or ticket suffix) — IDs are globally unique and reserved after deletion. Start from the **Development Project** preset.

## Works With

Every other GCP component consumes this one: reference `status.outputs.project_id` from any GCP kind's project field. Pair it with `GcpProjectIamMember` for additive IAM grants and `GcpVpcNetwork` for explicit networking.
