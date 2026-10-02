# GCP Organization Policy Custom Constraint

Defines one custom organization-policy constraint: a rule the organization writes itself, in Common Expression Language, over the fields of a Google Cloud resource -- "GKE node pools must auto-upgrade", "Cloud SQL instances must not have a public IP", "Compute instances must use Shielded VM". Google's predefined constraints cover the common guardrails; a custom constraint covers the ones specific to your organization. It is a DEFINITION: nothing is enforced until a `GcpOrgPolicy` references it -- at the organization, a folder, or a project -- so one rule is written once and enforced from as many places as the hierarchy needs.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Custom constraint** -- the `org_policy_custom_constraint` named `custom.{constraintName}` in the organization, with its resource types, method types, condition, and action

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials that can administer organization policies at the organization. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.

### GCP Organization

- **A Google Cloud Organization** -- a custom constraint's parent is always the organization; there is no project- or folder-scoped form.
- **IAM**: the deploying identity needs `roles/orgpolicy.policyAdmin` on the organization.
- **Supported services** -- each Google Cloud service publishes the resource types, CEL-visible fields, and methods a custom constraint can test (Google's "Custom constraints supported services" list). A constraint over an unsupported type or field is rejected at apply time.

## Deploy

### Console

Open the deployment store, find **GCP Organization Policy Custom Constraint**, and click **Deploy**. The creation wizard walks you through the organization, the resource types, the condition, and the action. Start from the **Deny Public GKE Nodes** preset in the [Presets](#presets) tab for a worked example.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpOrgPolicyCustomConstraint
metadata:
  name: disable-gke-auto-upgrade-off
  org: acme-corp
  env: prod
spec:
  organizationId: "123456789012"
  constraintName: disableGkeAutoUpgradeOff
  displayName: GKE node pools must auto-upgrade
  description: Node pools must enable auto-upgrade; set management.autoUpgrade to true
  resourceTypes:
    - container.googleapis.com/NodePool
  methodTypes:
    - CREATE
    - UPDATE
  condition: resource.management.autoUpgrade == false
  actionType: DENY
```

```shell
planton apply -f custom-constraint.yaml
```

This defines `custom.disableGkeAutoUpgradeOff` in the organization. Nothing is blocked yet -- a `GcpOrgPolicy` at some scope enforces it. A Stack Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a custom constraint. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Constraint name** -- the bare name (`disableGkeAutoUpgradeOff`); the module adds Google's `custom.` prefix, and `custom.{constraintName}` is the `constraint` output a `GcpOrgPolicy` references. Defaults to the manifest name. Immutable: a rename is a delete and a create, and every policy enforcing the old name lapses.

**Resource types and methods** -- the fully qualified REST resource types the condition is evaluated against (`container.googleapis.com/NodePool`) and the operations it runs on (`CREATE`, `UPDATE`; a few services support `DELETE`, `REMOVE_GRANT`, `GOVERN_TAGS`). Resource types are immutable.

**Condition and action** -- the CEL test over the resource's fields and what to do when it is true: `DENY` blocks the request (the condition describes the forbidden shape); `ALLOW` permits it and implicitly denies everything else (an allow-list). Both update in place.

**Description** -- the violation message the blocked engineer sees. Write it as the instruction they need.

**Deletion policy** -- `DELETE` (default) removes the constraint and every policy still enforcing it starts failing to apply -- destroy the policies first; `PREVENT` fails the destroy; `ABANDON` leaves the constraint in place.

## Outputs and Dependencies

### What This Component Consumes

This component has no foreign key dependencies: the organization is named by its numeric ID.

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `constraint` | `custom.{constraintName}` -- the handle a policy enforces | A `GcpOrgPolicy`'s `customConstraint` |
| `name` | `organizations/{org}/customConstraints/custom.{constraintName}` | Auditing; addressing the constraint in tooling |
| `update_time` | RFC 3339 last-update timestamp | Auditing |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Deny public GKE nodes** -- forbid node pools that expose public IPs. Start from the **Deny Public GKE Nodes** preset.

**Require Shielded VM** -- allow only Compute instances with Secure Boot on (an `ALLOW` constraint). Start from the **Require Shielded VM** preset.

## Works With

- [**GCP Organization Policy**](/cloud-catalog/gcp-org-policy) -- enforces this constraint at a project, folder, or the organization by reference
- [**GCP Folder**](/cloud-catalog/gcp-folder) -- the scopes the enforcing policies typically sit on
