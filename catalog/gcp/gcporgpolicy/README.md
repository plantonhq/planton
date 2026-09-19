# GCP Organization Policy

Sets one Google Cloud organization policy — the rules for one constraint at one point in the resource hierarchy (a project, a folder, or the organization). Organization policies are Google Cloud's guardrails: "no serial-port access on VMs", "resources only in these regions", "no public buckets", enforced by the platform on every create and update beneath the scope regardless of who makes the call or which tool they use. Boolean constraints take an `enforce` rule; list constraints take `values`, `allowAll`, or `denyAll`; the organization's own rules (`GcpOrgPolicyCustomConstraint`) are enforced by reference.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Organization policy** -- the `org_policy_policy` named `{scope}/policies/{constraint}`, with its enforced rule set (`policy`) and, when given, its audit-only rule set (`dryRunPolicy`)

## Before You Deploy

### The Constraint's Type Decides the Rule Shape — Read This First

- **Boolean constraints take `enforce`; list constraints take `values`, `allowAll`, or `denyAll`.** The type is not knowable offline, so the spec accepts every shape and Google rejects a mismatch at apply time with a clear message. `gcloud org-policies list-constraints --project <id>` shows each constraint's type.
- **A boolean constraint needs exactly one unconditional rule.** Conditional rules (by tag) must set `enforce` to the opposite of that rule and refine it.
- **One policy per constraint per scope.** The name is the identity, and both the scope and the constraint are immutable — changing either recreates the policy.

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials that can administer organization policies at the target scope.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Scope

- **IAM**: the deploying identity needs `roles/orgpolicy.policyAdmin` on the scope — on the project for a project-scoped policy (the one shape that needs no organization-level grant), on the folder or organization otherwise.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpOrgPolicy
metadata:
  name: disable-serial-port
spec:
  constraint: compute.disableSerialPortAccess
  policy:
    rules:
      - enforce: true
```

```shell
planton apply -f org-policy.yaml
```

With no `scope`, the policy applies to the project the credentials are configured for.

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `constraint` / `customConstraint` | `string` / `StringValueOrRef` | Exactly one: a predefined or managed constraint by name (`compute.disableSerialPortAccess`), or a reference to a `GcpOrgPolicyCustomConstraint`. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `scope` | `message` | provider default project | At most one of `projectId` (a `GcpProject` reference), `folderId` (a `GcpFolder` reference), `organizationId`. Immutable. |
| `policy` | `message` | — | The ENFORCED rule set: `inheritFromParent`, `reset`, `rules[]`. |
| `dryRunPolicy` | `message` | — | The same shape, evaluated in audit mode (violations logged, nothing blocked). |
| `policy.rules[].enforce` | `bool` | — | Boolean constraints: enforced (`true`) or explicitly not (`false`). |
| `policy.rules[].values` | `message` | — | List constraints: `allowedValues[]`, `deniedValues[]`. |
| `policy.rules[].allowAll` / `denyAll` | `bool` | — | List constraints: every value allowed / denied. |
| `policy.rules[].condition` | `message` | — | `expression` (CEL of `resource.matchTag(...)` tests), `title`, `description`, `location`. |
| `policy.rules[].parameters` | `string` | — | JSON object of parameter values for managed constraints that declare parameters. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT` (destroy fails), or `ABANDON` (unmanaged, stays enforced). |

### Validation Rules

- **Exactly one constraint**: `constraint` or `customConstraint`.
- **At most one scope arm**; `organizationId` is numeric.
- **Exactly one verdict per rule**: `allowAll`, `denyAll`, `enforce`, or `values` — the API's own one-of, enforced by shape.
- **`reset` stands alone**: a reset rule set carries no rules and does not set `inheritFromParent`.
- **A condition needs an `expression`**; list values are non-empty strings.
- **`parameters`** is a JSON object (`{...}`).
- **`deletionPolicy`**: `DELETE`, `PREVENT`, or `ABANDON`.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `{scope}/policies/{constraint}` — a project scope is reported as the project number |
| `etag` | `string` | Opaque version marker Google changes on every update |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Dry run first.** Roll out a new guardrail as `dryRunPolicy` alone, read the violations it would have caused in the audit log (`dryRunPolicyViolation`), then add the same rules to `policy`.
- **`enforce: false` is a real rule** — the way a child scope relaxes a guardrail its parent enforces. Both engines send Google's `"FALSE"` for it; an unset verdict is never sent.
- **Rule order never matters** — Google compares rules as a set, so re-plans stay clean whatever order the manifest lists them in.
- **Deleting the policy** returns the scope to what it inherits from above (or the constraint's default); `PREVENT` guards the guardrails a landing zone depends on.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- [GcpOrgPolicyCustomConstraint](/docs/catalog/gcp/gcporgpolicycustomconstraint) — the organization's own rule this policy enforces
- [GcpFolder](/docs/catalog/gcp/gcpfolder) — the folder scope
- [GcpProject](/docs/catalog/gcp/gcpproject) — the project scope
- [GcpTagKey](/docs/catalog/gcp/gcptagkey) — the tags a rule's condition tests

## Additional Resources

- [Organization Policy Service overview](https://cloud.google.com/resource-manager/docs/organization-policy/overview)
- [Constraints list](https://cloud.google.com/resource-manager/docs/organization-policy/org-policy-constraints)
- [Tags and conditional policies](https://cloud.google.com/resource-manager/docs/organization-policy/tags-organization-policy)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
