# GcpProject

## Overview

`GcpProject` manages one Google Cloud project — the Layer-0 container every
other GCP resource lives in. Every other GCP kind's `projectId` field can
reference a `GcpProject` resource, making the project the root node of the
GCP composition graph.

The component owns exactly what the project resource owns: identity
(`projectId`), hierarchy placement (organization or folder), billing
linkage, labels and resource-manager tags, the default-network decision,
pre-enabled Cloud APIs, and the deletion policy. IAM grants are
deliberately NOT bundled here — model each grant as a first-class
`GcpProjectIamMember` resource.

## Key Features

- **Hierarchy placement**: create under an organization or a folder --
  by numeric ID, or by referencing a `GcpFolder` declared in the same
  chart (`folderId`); changing the parent migrates the project.
- **Billing linkage**: attach a billing account at creation
  ([`iac/permissions.yaml`](iac/permissions.yaml) lists the exact
  least-privilege grants, including the billing-account-scoped one).
- **Honest deletion semantics**: `deletionPolicy` is the provider's real
  three-way switch — `DELETE` (default), `PREVENT` (destroy fails), or
  `ABANDON` (remove from state, project lives on) — instead of a boolean
  that hides the ABANDON path.
- **Hardened by default**: the auto-created "default" VPC network is
  suppressed unless `autoCreateNetwork: true`; explicit `GcpVpcNetwork`
  resources are the composable path.
- **Cost attribution**: user labels merge beneath the platform's
  attribution labels identically on both engines; labels are the primary
  cost-allocation dimension in billing exports.
- **Create-time tags**: resource-manager tags (`tagKeys/… → tagValues/…`)
  bind at creation for org-policy and IAM-condition targeting; tag an
  existing project with `GcpTagBinding` instead.

## Example Usage

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpProject
metadata:
  name: prod-workloads
spec:
  projectId: acme-prod-workloads
  parentType: folder
  parentId: "123456789012"
  billingAccountId: 0123AB-4567CD-89EFGH
  deletionPolicy: PREVENT
  enabledApis:
    - compute.googleapis.com
    - container.googleapis.com
```

Grant access with first-class IAM members:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpProjectIamMember
metadata:
  name: platform-admins
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: prod-workloads
      fieldPath: status.outputs.project_id
  role:
    value: roles/editor
  member:
    value: group:platform-admins@acme.com
```

### Deploy with CLI

```bash
planton pulumi up --manifest project.yaml
# or
planton tofu apply --manifest project.yaml
```

## Best Practices

1. **Project IDs are permanent, globally-unique identity** — deleted IDs
   stay reserved for ~30 days. Choose them deliberately; uniqueness
   belongs in the ID itself.
2. **Use `PREVENT` on foundation projects** and flip to `DELETE`
   deliberately when decommissioning; use `ABANDON` when ownership moves
   out of IaC.
3. **Keep IAM out of the project resource** — one `GcpProjectIamMember`
   per grant composes without clobbering anyone else's grants.
4. **Reference, don't repeat**: downstream kinds resolve
   `status.outputs.project_id` via `valueFrom`, so a project rename never
   ripples.

## Stack Outputs

| Output | Description |
|--------|-------------|
| `project_id` | The unique project ID — the value every other kind's `projectId` reference resolves |
| `project_number` | The numeric identifier assigned by Google |
| `name` | The display name |

## Related Components

- **GcpFolder** — the folder the project is placed in by reference
- **GcpOrgPolicy** — guardrails scoped to the project
- **GcpTagBinding** — tags on an existing project
- **GcpProjectIamMember** — additive IAM grants on the project
- **GcpVpcNetwork** — explicit networks (instead of the default network)
- **GcpServiceAccount** — workload identities inside the project

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
