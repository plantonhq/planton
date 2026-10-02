# GCP Organization Policy

Sets one Google Cloud organization policy: the rules for one constraint at one point in the resource hierarchy -- a project, a folder, or the organization. Organization policies are Google Cloud's guardrails ("no serial-port access on VMs", "resources only in these regions", "no public buckets"), enforced by the platform on every create and update beneath the scope regardless of who makes the call. Boolean constraints take an `enforce` rule; list constraints take `values`, `allowAll`, or `denyAll`; the organization's own rules (`GcpOrgPolicyCustomConstraint`) are enforced by reference.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Organization policy** -- the `org_policy_policy` named `{scope}/policies/{constraint}`, with its enforced rule set (`policy`) and, when given, its audit-only rule set (`dryRunPolicy`)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials that can administer organization policies at the target scope. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.

### GCP Scope

- **IAM**: the deploying identity needs `roles/orgpolicy.policyAdmin` on the scope -- on the project for a project-scoped policy (the one shape that needs no organization-level grant), on the folder or organization otherwise.
- **The constraint's type decides the rule shape** -- a boolean constraint (`compute.disableSerialPortAccess`) takes `enforce`; a list constraint (`gcp.resourceLocations`) takes `values`, `allowAll`, or `denyAll`. Google rejects a mismatch at apply time; `gcloud org-policies list-constraints` shows the type.
- **One policy per constraint per scope** -- the name is the identity; a second manifest on the same pair collides with the first.

## Deploy

### Console

Open the deployment store, find **GCP Organization Policy**, and click **Deploy**. The creation wizard walks you through the scope, the constraint, and the rules. Start from the **Disable Serial Port (Project)** preset in the [Presets](#presets) tab for the simplest guardrail.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpOrgPolicy
metadata:
  name: disable-serial-port
  org: acme-corp
  env: prod
spec:
  scope:
    projectId:
      value: acme-prod-workloads
  constraint: compute.disableSerialPortAccess
  policy:
    rules:
      - enforce: true
```

```shell
planton apply -f org-policy.yaml
```

This enforces the boolean constraint on the project: from now on no VM in it can enable serial-port access. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the policy references its scope (a `GcpProject` or `GcpFolder`) and, for an organization-defined rule, its `GcpOrgPolicyCustomConstraint` via ValueFromRef:

```yaml
spec:
  scope:
    folderId:
      valueFrom:
        kind: GcpFolder
        name: production
        fieldPath: status.outputs.folder_id
  customConstraint:
    valueFrom:
      kind: GcpOrgPolicyCustomConstraint
      name: deny-gke-auto-upgrade-off
      fieldPath: status.outputs.constraint
  policy:
    rules:
      - enforce: true
```

The InfraPipeline deploys the folder and the constraint first, then the policy that enforces the constraint on the folder.

## Key Configuration

These are the most important decisions when configuring an organization policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scope** -- at most one of `projectId`, `folderId`, `organizationId`; empty means the provider's default project. Everything beneath the scope inherits the policy. Immutable: a change recreates the policy.

**Constraint** -- exactly one of `constraint` (a predefined or managed constraint by name) or `customConstraint` (a reference to a `GcpOrgPolicyCustomConstraint`). Immutable.

**Rules** -- each rule carries exactly one verdict: `enforce` (boolean constraints), or `values` / `allowAll` / `denyAll` (list constraints), optionally scoped by a tag `condition` (`resource.matchTag(...)`). A boolean constraint needs exactly one unconditional rule; conditional rules set `enforce` to the opposite and refine it. `enforce: false` is a real rule -- the way a child scope relaxes a guardrail its parent enforces.

**Dry run first** -- `dryRunPolicy` evaluates the same rule shape in audit mode: violations are logged, nothing is blocked. Roll out a new guardrail as a dry-run policy alone, read the violations it would have caused, then add the rules to `policy`.

**Inherit or reset** -- `inheritFromParent` (list constraints) keeps the values set higher in the hierarchy in effect beside this policy's; `reset` discards everything inherited and restores the constraint's default -- and must stand alone.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `scope.projectId` | `status.outputs.project_id` |
| **GcpFolder** (optional) | `scope.folderId` | `status.outputs.folder_id` |
| **GcpOrgPolicyCustomConstraint** (optional) | `customConstraint` | `status.outputs.constraint` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | `{scope}/policies/{constraint}` -- the policy's full resource name | Auditing; addressing the policy in tooling |
| `etag` | Opaque version marker Google changes on every update | Detecting out-of-band edits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Disable serial port (project)** -- a boolean constraint enforced on one project; the shape that proves live without organization credentials. Start from the **Disable Serial Port (Project)** preset.

**Restrict locations (folder)** -- the `gcp.resourceLocations` list constraint on a folder with a value group. Start from the **Restrict Locations (Folder)** preset.

**Tag-conditioned dry run** -- a guardrail with a sandbox exemption by tag, mirrored in a dry-run policy for a safe rollout. Start from the **Tag-Conditioned Dry Run** preset.

## Works With

- [**GCP Folder**](/cloud-catalog/gcp-folder) -- the folder scope, inherited by everything beneath it
- [**GCP Project**](/cloud-catalog/gcp-project) -- the project scope
- [**GCP Organization Policy Custom Constraint**](/cloud-catalog/gcp-org-policy-custom-constraint) -- the organization's own rule this policy enforces
- [**GCP Tag Key**](/cloud-catalog/gcp-tag-key) -- the tags a rule's `condition` tests
