# GCP Organization Policy Custom Constraint

Defines one custom organization-policy constraint — a rule the organization writes itself, in Common Expression Language, over the fields of a Google Cloud resource: "GKE node pools must auto-upgrade", "Cloud SQL instances must not have a public IP", "Compute instances must use Shielded VM". Google's predefined constraints cover the common guardrails; a custom constraint covers the ones specific to your organization. It is a DEFINITION: nothing is enforced until a `GcpOrgPolicy` references it, so one rule is written once and enforced from as many places in the hierarchy as needed.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Custom constraint** -- the `org_policy_custom_constraint` named `custom.{constraintName}` in the organization, with its resource types, method types, condition, and action

## Before You Deploy

### A Constraint Is a Definition, Not an Enforcement — Read This First

- **Nothing happens until a policy references it.** Enforce the constraint at a project, folder, or the organization with a `GcpOrgPolicy` whose `customConstraint` points at this resource's `constraint` output.
- **The name, the organization, and the resource types are immutable.** Changing any recreates the constraint, and every policy enforcing the old name lapses. The condition, action, methods, display name, and description update in place.
- **Destroy the policies first.** Deleting a constraint that policies still enforce makes those policies fail to apply; a chart whose policies reference this resource destroys them in the right order.

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials that can administer organization policies at the organization.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Organization

- **A Google Cloud Organization** -- a custom constraint's parent is always the organization.
- **IAM**: the deploying identity needs `roles/orgpolicy.policyAdmin` on the organization.
- **Supported services** -- each service publishes the resource types, CEL-visible fields, and methods a constraint can test; an unsupported one is rejected at apply time.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpOrgPolicyCustomConstraint
metadata:
  name: disable-gke-auto-upgrade-off
spec:
  organizationId: "123456789012"
  constraintName: disableGkeAutoUpgradeOff
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

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `organizationId` | `string` | The numeric organization ID. Immutable. |
| `resourceTypes` | `list<string>` | Fully qualified REST resource types (`container.googleapis.com/NodePool`), at least one, all from one service. Immutable. |
| `methodTypes` | `list<string>` | `CREATE`, `UPDATE`, `DELETE`, `REMOVE_GRANT`, `GOVERN_TAGS` (per the service's support), at least one. |
| `condition` | `string` | The CEL test over the resource's fields. |
| `actionType` | `string` | `DENY` (block when the condition is true) or `ALLOW` (permit only when true). |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `constraintName` | `string` | `metadata.name` | The bare name; the module adds the `custom.` prefix. Letter first, letters and digits, at most 62 characters. Immutable. |
| `displayName` | `string` | — | Console label. |
| `description` | `string` | — | The violation message a blocked user sees. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT` (destroy fails), or `ABANDON` (unmanaged, stays enforceable). |

### Validation Rules

- **`organizationId`** is numeric, without the `organizations/` prefix.
- **`constraintName`**: `^[A-Za-z][A-Za-z0-9]{0,61}$`, never with the `custom.` prefix (the module adds it).
- **`resourceTypes`**: each `^[a-z][a-z0-9]*\.googleapis\.com/[A-Za-z][A-Za-z0-9]*$`.
- **`methodTypes`**: each one of `CREATE`, `UPDATE`, `DELETE`, `REMOVE_GRANT`, `GOVERN_TAGS`.
- **`actionType`**: `ALLOW` or `DENY`.
- **`deletionPolicy`**: `DELETE`, `PREVENT`, or `ABANDON`.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `organizations/{org}/customConstraints/custom.{constraintName}` |
| `constraint` | `string` | `custom.{constraintName}` — what a `GcpOrgPolicy`'s `customConstraint` references |
| `update_time` | `string` | RFC 3339 last-update timestamp |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **`DENY` versus `ALLOW`** — with `DENY` the condition describes the forbidden shape; with `ALLOW` it describes the only acceptable shape and everything else is implicitly denied. Most guardrails are `DENY`.
- **Write the description for the person it blocks** — it is the error message they read; make it the instruction they need.
- **The custom. prefix is the module's** — the spec's `constraintName` is the bare name and the CEL rejects the prefix, so it can never be doubled.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- [GcpOrgPolicy](/docs/catalog/gcp/gcporgpolicy) — enforces the constraint at a scope by reference
- [GcpFolder](/docs/catalog/gcp/gcpfolder) — the scopes the enforcing policies typically sit on

## Additional Resources

- [Creating and managing custom constraints](https://cloud.google.com/resource-manager/docs/organization-policy/creating-managing-custom-constraints)
- [Custom constraints supported services](https://cloud.google.com/resource-manager/docs/organization-policy/custom-constraint-supported-services)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
