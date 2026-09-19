# GcpOrgPolicyCustomConstraint

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpOrgPolicyCustomConstraintSpec defines ONE custom organization-policy
constraint: a rule the organization writes itself, in Common Expression
Language, over the fields of a Google Cloud resource -- "GKE node pools
must have auto-upgrade on", "Cloud SQL instances must not have a public
IP", "Compute instances must use Shielded VM". Google's predefined
constraints cover the common guardrails; a custom constraint covers the
ones specific to your organization.

A custom constraint is a DEFINITION, not an enforcement. It lives at the
organization (the only parent Google allows) and does nothing until a
GcpOrgPolicy enforces it -- at the organization, at a folder, or at a
project -- by referencing this resource's `constraint` output. Define
the rule once here; enforce it from as many policies as the hierarchy
needs. That split is why this is its own kind: the constraint outlives
any single policy, and deleting one policy must not delete the rule
two other folders rely on.

Which resources and fields a constraint can test is Google's list
("Custom constraints supported services"): each service publishes the
resource types and the CEL-visible fields, and which methods (create,
update, ...) it evaluates. A constraint over an unsupported resource
type or field is rejected at apply time.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpOrgPolicyCustomConstraint
metadata:
  name: disable-gke-auto-upgrade-off
spec:
  # The owning organization (numeric). A custom constraint's parent is
  # always the organization. Immutable.
  organizationId: "123456789012"

  # The bare name; the module adds Google's `custom.` prefix and the
  # `constraint` output is `custom.disableGkeAutoUpgradeOff` -- what a
  # GcpOrgPolicy's customConstraint references. Defaults to metadata.name.
  # Immutable.
  constraintName: disableGkeAutoUpgradeOff

  displayName: GKE node pools must auto-upgrade
  # The violation message a blocked engineer reads.
  description: Node pools must enable auto-upgrade; set management.autoUpgrade to true

  # The REST resource types the condition is evaluated against (one
  # service). Immutable.
  resourceTypes:
    - container.googleapis.com/NodePool

  # The operations evaluated: CREATE, UPDATE (all services); DELETE,
  # REMOVE_GRANT, GOVERN_TAGS (some).
  methodTypes:
    - CREATE
    - UPDATE

  # CEL over the resource's fields; with DENY, true means "blocked".
  condition: resource.management.autoUpgrade == false
  actionType: DENY

  # DELETE (default; policies still enforcing it start failing), PREVENT,
  # ABANDON.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.organizationId` | `string` | yes |  |  |
| `spec.constraintName` | `string` |  |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.resourceTypes` | `[]string` | yes |  |  |
| `spec.methodTypes` | `[]string` | yes |  |  |
| `spec.condition` | `string` | yes |  |  |
| `spec.actionType` | `string` | yes |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.organizationId

`string` · required

The organization that owns the constraint: the numeric organization
ID (from `gcloud organizations list`), without the `organizations/`
prefix. Immutable -- a constraint cannot move between organizations.

- rule: {"required":true,"string":{"pattern":"^[0-9]+$"}}

### spec.constraintName

`string`

The constraint's name WITHOUT the `custom.` prefix -- the module adds
it, so Google knows the constraint as `custom.<constraint_name>` and
that full handle is the `constraint` output a GcpOrgPolicy references.
Defaults to metadata.name when empty. Google's rule: starts with a
letter, then letters and digits, at most 62 characters in all;
unique within the organization. Immutable: a rename is a delete and a
create, and every policy enforcing the old name would lapse.

- rule: constraint_name must start with a letter and contain only letters and digits (at most 62 characters), written WITHOUT the custom. prefix -- e.g. disableGkeAutoUpgrade

### spec.displayName

`string`

A human-friendly name for the constraint, shown in the console's
organization-policy pages. Mutable.

- rule: {"string":{"maxLen":"200"}}

### spec.description

`string`

What Google shows a user whose request the constraint blocks -- the
violation message. Write it as the instruction the blocked engineer
needs ("Node pools must enable auto-upgrade; set
management.autoUpgrade to true"). Mutable.

- rule: {"string":{"maxLen":"2000"}}

### spec.resourceTypes

`[]string` · required

The Google Cloud REST resource types the condition is evaluated
against, fully qualified: `container.googleapis.com/NodePool`,
`compute.googleapis.com/Instance`, `sqladmin.googleapis.com/Instance`,
`storage.googleapis.com/Bucket`. At least one; every type in one
constraint must belong to the same service. Immutable: changing the
list recreates the constraint.

- rule: {"repeated":{"minItems":"1","items":{"string":{"pattern":"^[a-z][a-z0-9]*\\.googleapis\\.com/[A-Za-z][A-Za-z0-9]*$"}}}}

### spec.methodTypes

`[]string` · required

The operations the constraint is evaluated on. CREATE and UPDATE are
supported by every service that supports custom constraints; DELETE,
REMOVE_GRANT, and GOVERN_TAGS are supported by a few (the supported
services list says which). Google rejects a method a service does not
support at apply time. Mutable.

- rule: method_types entries must each be one of: CREATE, UPDATE, DELETE, REMOVE_GRANT, GOVERN_TAGS
- rule: {"repeated":{"minItems":"1"}}

### spec.condition

`string` · required

The Common Expression Language test over the resource, evaluated on
each method in method_types: `resource.management.autoUpgrade == false`,
`resource.settings.ipConfiguration.ipv4Enabled == true`,
`!has(resource.shieldedInstanceConfig) || resource.shieldedInstanceConfig.enableSecureBoot == false`.
The fields are the service's own REST resource fields; Google's per-
service pages list the ones the constraint engine can see. Together
with action_type it reads: "when this is true, ALLOW or DENY the
request". Mutable.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.actionType

`string` · required

What happens when the condition is true for a request: DENY blocks it
(the usual guardrail -- the condition describes the forbidden shape),
ALLOW permits it and implicitly denies everything the condition does
not match (an allow-list -- the condition describes the only
acceptable shape). Mutable.

- rule: action_type must be one of: ALLOW, DENY
- rule: {"required":true}

### spec.deletionPolicy

`string`

What destroying this resource does to the constraint in GCP:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the constraint is deleted; every GcpOrgPolicy still
               enforcing it starts failing to apply, so destroy the
               policies first (a chart's dependency order does this
               when the policies reference this resource)
  "PREVENT" -- destroy FAILS; the guard for a rule many policies rely on
  "ABANDON" -- the constraint is removed from management but keeps
               existing in GCP, still enforceable by policies

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpOrgPolicyCustomConstraint, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | The constraint's full resource name, `organizations/{org_id}/customConstraints/custom.{constraint_name}`. |
| `status.outputs.constraint` | `string` | The handle a policy enforces the constraint by: `custom.{constraint_name}`. This is what a GcpOrgPolicy's `custom_constraint` references, and what `gcloud org-policies` and the console call the constraint. |
| `status.outputs.update_time` | `string` | When the constraint was last updated (RFC 3339 UTC). |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpOrgPolicy | `spec.customConstraint` | `status.outputs.constraint` |

## See Also

- [Overview](../README.md)
