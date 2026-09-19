# GcpOrgPolicy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpOrgPolicySpec sets ONE organization policy: the rules for ONE
constraint at ONE point in the resource hierarchy (a project, a folder,
or the organization). An organization policy is Google Cloud's guardrail
mechanism -- "no serial-port access on VMs", "resources only in these
regions", "no public Cloud Storage buckets" -- enforced by the platform
on every create and update beneath the scope, regardless of who makes
the call or which tool they use.

Two kinds of constraint exist and the rule shape follows the constraint:
  - a BOOLEAN constraint (`compute.disableSerialPortAccess`,
    `iam.disableServiceAccountKeyCreation`) takes a rule with `enforce`;
  - a LIST constraint (`gcp.resourceLocations`,
    `iam.allowedPolicyMemberDomains`) takes a rule with `values`,
    `allow_all`, or `deny_all`.
The constraint's type is not knowable offline, so the spec accepts every
rule shape and Google rejects a mismatch at apply time with a clear
message. Constraints beginning `custom.` are the organization's own
rules, defined once with GcpOrgPolicyCustomConstraint and enforced from
any number of policies by reference (`custom_constraint`).

One policy per constraint per scope: the policy's name IS
`{scope}/policies/{constraint}`, so a second GcpOrgPolicy on the same
pair collides with the first. Both the scope and the constraint are
immutable -- changing either recreates the policy (and the API serializes
the create behind the delete, so the old rules lapse for a moment).

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpOrgPolicy
metadata:
  name: disable-serial-port
spec:
  # Where the policy applies: at most one of projectId (a literal or a
  # GcpProject reference), folderId (a GcpFolder reference), organizationId.
  # Omit the whole block to use the provider's default project. Immutable.
  scope:
    projectId:
      value: my-gcp-project-123

  # Exactly one of constraint (a predefined or managed constraint by name)
  # or customConstraint (a reference to a GcpOrgPolicyCustomConstraint's
  # constraint output). Immutable. A boolean constraint takes `enforce`; a
  # list constraint takes `values`, `allowAll`, or `denyAll`.
  constraint: compute.disableSerialPortAccess

  # The ENFORCED rules. A boolean constraint needs exactly one unconditional
  # rule; conditional rules (by tag) set the opposite and refine it.
  policy:
    rules:
      - enforce: true
      - enforce: false
        condition:
          title: sandbox exemption
          expression: resource.matchTag('123456789012/environment', 'sandbox')

  # The same shape in AUDIT mode: violations are logged, nothing is blocked.
  # Roll a new guardrail out here first.
  dryRunPolicy:
    rules:
      - enforce: true

  # DELETE (default), PREVENT (destroy fails), ABANDON (unmanaged, stays
  # enforced).
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.scope` | `GcpOrgPolicyScope` |  |  |  |
| `spec.scope.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.scope.folderId` | `string \| valueFrom` |  |  | GcpFolder (`status.outputs.folder_id`) |
| `spec.scope.organizationId` | `string` |  |  |  |
| `spec.constraint` | `string` |  |  |  |
| `spec.customConstraint` | `string \| valueFrom` |  |  | GcpOrgPolicyCustomConstraint (`status.outputs.constraint`) |
| `spec.policy` | `GcpOrgPolicyRuleSet` |  |  |  |
| `spec.policy.inheritFromParent` | `bool` |  |  |  |
| `spec.policy.reset` | `bool` |  |  |  |
| `spec.policy.rules` | `[]GcpOrgPolicyRule` |  |  |  |
| `spec.policy.rules[].allowAll` | `bool` |  |  |  |
| `spec.policy.rules[].denyAll` | `bool` |  |  |  |
| `spec.policy.rules[].enforce` | `bool` |  |  |  |
| `spec.policy.rules[].values` | `GcpOrgPolicyRuleValues` |  |  |  |
| `spec.policy.rules[].values.allowedValues` | `[]string` |  |  |  |
| `spec.policy.rules[].values.deniedValues` | `[]string` |  |  |  |
| `spec.policy.rules[].condition` | `GcpOrgPolicyRuleCondition` |  |  |  |
| `spec.policy.rules[].condition.expression` | `string` | yes |  |  |
| `spec.policy.rules[].condition.title` | `string` |  |  |  |
| `spec.policy.rules[].condition.description` | `string` |  |  |  |
| `spec.policy.rules[].condition.location` | `string` |  |  |  |
| `spec.policy.rules[].parameters` | `string` |  |  |  |
| `spec.dryRunPolicy` | `GcpOrgPolicyRuleSet` |  |  |  |
| `spec.dryRunPolicy.inheritFromParent` | `bool` |  |  |  |
| `spec.dryRunPolicy.reset` | `bool` |  |  |  |
| `spec.dryRunPolicy.rules` | `[]GcpOrgPolicyRule` |  |  |  |
| `spec.dryRunPolicy.rules[].allowAll` | `bool` |  |  |  |
| `spec.dryRunPolicy.rules[].denyAll` | `bool` |  |  |  |
| `spec.dryRunPolicy.rules[].enforce` | `bool` |  |  |  |
| `spec.dryRunPolicy.rules[].values` | `GcpOrgPolicyRuleValues` |  |  |  |
| `spec.dryRunPolicy.rules[].values.allowedValues` | `[]string` |  |  |  |
| `spec.dryRunPolicy.rules[].values.deniedValues` | `[]string` |  |  |  |
| `spec.dryRunPolicy.rules[].condition` | `GcpOrgPolicyRuleCondition` |  |  |  |
| `spec.dryRunPolicy.rules[].condition.expression` | `string` | yes |  |  |
| `spec.dryRunPolicy.rules[].condition.title` | `string` |  |  |  |
| `spec.dryRunPolicy.rules[].condition.description` | `string` |  |  |  |
| `spec.dryRunPolicy.rules[].condition.location` | `string` |  |  |  |
| `spec.dryRunPolicy.rules[].parameters` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.scope

`GcpOrgPolicyScope`

Where the policy applies: a project, a folder, or the organization.
At most one arm; all empty means the provider's default project (the
project the deploying credentials are configured for). Everything
beneath the scope inherits the policy unless a lower policy overrides
it (`inherit_from_parent`) or resets it (`reset`). Immutable.

- rule: set at most one of project_id, folder_id, or organization_id (empty means the provider's default project)

### spec.scope.projectId

`string | valueFrom`

Project scope: a literal project ID (Google also accepts the project
number) or a reference to a GcpProject resource. The API stores the
policy under the project NUMBER and the provider treats the two forms
as equal, so a literal ID never shows a spurious change.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.scope.folderId

`string | valueFrom`

Folder scope: the folder's numeric ID -- a literal, or a reference to
a GcpFolder resource (its folder_id output). Every project and folder
beneath it inherits the policy.

- references: GcpFolder (`status.outputs.folder_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpFolder, name: <that resource's name>, fieldPath: status.outputs.folder_id}} -- a bare string does not parse

### spec.scope.organizationId

`string`

Organization scope: the numeric organization ID, without the
`organizations/` prefix. The root of inheritance for the whole estate.

- rule: organization_id must be the numeric organization ID, without the organizations/ prefix

### spec.constraint

`string`

A predefined or managed constraint by name, exactly as Google lists it
(`gcloud org-policies list-constraints`): `compute.disableSerialPortAccess`,
`gcp.resourceLocations`, `iam.managed.disableServiceAccountKeyCreation`,
`storage.publicAccessPrevention`. Immutable. Exactly one of this and
custom_constraint.

- rule: constraint must be a dotted constraint name such as compute.disableSerialPortAccess or gcp.resourceLocations

### spec.customConstraint

`string | valueFrom`

The organization's own constraint to enforce: a reference to a
GcpOrgPolicyCustomConstraint resource (its `constraint` output, the
`custom.<name>` handle), or that handle as a literal. Immutable.
Exactly one of this and constraint. A custom constraint is always
organization-defined, but the policy that enforces it may sit at any
scope -- this is how one rule is written once and applied to three
folders by three small policies.

- references: GcpOrgPolicyCustomConstraint (`status.outputs.constraint`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpOrgPolicyCustomConstraint, name: <that resource's name>, fieldPath: status.outputs.constraint}} -- a bare string does not parse

### spec.policy

`GcpOrgPolicyRuleSet`

The rules Google ENFORCES at this scope. Omit to set no live rules --
meaningful only together with dry_run_policy, the audit-first rollout:
first a dry-run policy alone, read the violations it would have caused
in the audit log, then add the same rules here.

- rule: a rule set with reset: true must carry no rules and must not set inherit_from_parent -- reset restores the constraint's default and discards everything inherited

### spec.policy.inheritFromParent

`bool`

For LIST constraints only: whether the values allowed or denied by
policies higher in the hierarchy stay in effect here alongside this
policy's rules (true), or this policy becomes the new root of
evaluation and nothing above it applies (false, the default). Google
ignores it on boolean constraints. Cannot be true together with reset.

### spec.policy.reset

`bool`

Discard every rule inherited from above and restore the constraint's
own default behavior at this scope (the enforcement Google applies
when nobody has set a policy). Mutually exclusive with rules and with
inherit_from_parent: a reset policy carries nothing else. Works for
both list and boolean constraints.

### spec.policy.rules

`[]GcpOrgPolicyRule`

The rules. For a BOOLEAN constraint there must be exactly one rule
without a condition, and every conditional rule must set `enforce` to
the opposite of that unconditional rule (Google's rule; violations are
rejected at apply time). For a LIST constraint the rules combine:
unconditional rules set the baseline, conditional rules refine it for
resources whose tags match. Google compares rules as a set, so their
order never causes a spurious change.

### spec.policy.rules[].allowAll

`bool`

LIST constraints: every value is allowed at this scope (`true`).
Setting it to false is meaningful only as the unconditional baseline
that a conditional rule overrides. Both engines send Google's string
form ("TRUE"/"FALSE").

### spec.policy.rules[].denyAll

`bool`

LIST constraints: every value is denied at this scope (`true`).
Both engines send Google's string form ("TRUE"/"FALSE").

### spec.policy.rules[].enforce

`bool`

BOOLEAN constraints: the constraint is enforced here (`true`) or
explicitly NOT enforced (`false`) -- the way a child scope relaxes
a guardrail its parent enforces. Both engines send Google's string
form ("TRUE"/"FALSE").

### spec.policy.rules[].values

`GcpOrgPolicyRuleValues`

LIST constraints: the explicit allow and deny lists.

### spec.policy.rules[].values.allowedValues

`[]string`

Values permitted at this scope.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.policy.rules[].values.deniedValues

`[]string`

Values forbidden at this scope.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.policy.rules[].condition

`GcpOrgPolicyRuleCondition`

Apply this rule only to resources whose Resource Manager tags match.
Omit for an unconditional rule. A boolean constraint needs exactly one
unconditional rule; conditional rules refine it.

### spec.policy.rules[].condition.expression

`string` · required

A Common Expression Language expression of 1 to 10 tag tests joined by
`||` or `&&`. Each test is either
`resource.matchTag('<org_id>/<tag_key_short_name>', '<tag_value_short_name>')`
(by short names) or `resource.matchTagId('tagKeys/<id>', 'tagValues/<id>')`
(by the ids GcpTagKey and GcpTagValue output). Example:
`resource.matchTag('123456789012/environment', 'prod')`.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.policy.rules[].condition.title

`string`

A short label for the condition, shown in the console and audit logs.

### spec.policy.rules[].condition.description

`string`

A longer explanation of what the condition selects and why.

### spec.policy.rules[].condition.location

`string`

Where the expression came from, for error reporting (a file name and
position, or any free text).

### spec.policy.rules[].parameters

`string`

For MANAGED constraints that declare parameters (the `*.managed.*`
family): the parameter values as one JSON object, typed as the
constraint defines them -- e.g.
`{"allowedLocations": ["us-east1", "us-west1"], "allowAll": true}`.
Google validates the object against the constraint at apply time;
the provider validates only that it parses as JSON.

- rule: parameters must be a JSON object, e.g. {"allowedLocations": ["us-east1"]}

### spec.dryRunPolicy

`GcpOrgPolicyRuleSet`

The same rule shape, evaluated in AUDIT mode: violations are logged
(`cloudaudit.googleapis.com/policy`, `dryRunPolicyViolation`), nothing
is blocked. Set it beside `policy` to preview a tightening before it
bites, or alone to measure a new guardrail against live traffic first.

- rule: a rule set with reset: true must carry no rules and must not set inherit_from_parent -- reset restores the constraint's default and discards everything inherited

### spec.dryRunPolicy.inheritFromParent

`bool`

For LIST constraints only: whether the values allowed or denied by
policies higher in the hierarchy stay in effect here alongside this
policy's rules (true), or this policy becomes the new root of
evaluation and nothing above it applies (false, the default). Google
ignores it on boolean constraints. Cannot be true together with reset.

### spec.dryRunPolicy.reset

`bool`

Discard every rule inherited from above and restore the constraint's
own default behavior at this scope (the enforcement Google applies
when nobody has set a policy). Mutually exclusive with rules and with
inherit_from_parent: a reset policy carries nothing else. Works for
both list and boolean constraints.

### spec.dryRunPolicy.rules

`[]GcpOrgPolicyRule`

The rules. For a BOOLEAN constraint there must be exactly one rule
without a condition, and every conditional rule must set `enforce` to
the opposite of that unconditional rule (Google's rule; violations are
rejected at apply time). For a LIST constraint the rules combine:
unconditional rules set the baseline, conditional rules refine it for
resources whose tags match. Google compares rules as a set, so their
order never causes a spurious change.

### spec.dryRunPolicy.rules[].allowAll

`bool`

LIST constraints: every value is allowed at this scope (`true`).
Setting it to false is meaningful only as the unconditional baseline
that a conditional rule overrides. Both engines send Google's string
form ("TRUE"/"FALSE").

### spec.dryRunPolicy.rules[].denyAll

`bool`

LIST constraints: every value is denied at this scope (`true`).
Both engines send Google's string form ("TRUE"/"FALSE").

### spec.dryRunPolicy.rules[].enforce

`bool`

BOOLEAN constraints: the constraint is enforced here (`true`) or
explicitly NOT enforced (`false`) -- the way a child scope relaxes
a guardrail its parent enforces. Both engines send Google's string
form ("TRUE"/"FALSE").

### spec.dryRunPolicy.rules[].values

`GcpOrgPolicyRuleValues`

LIST constraints: the explicit allow and deny lists.

### spec.dryRunPolicy.rules[].values.allowedValues

`[]string`

Values permitted at this scope.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.dryRunPolicy.rules[].values.deniedValues

`[]string`

Values forbidden at this scope.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.dryRunPolicy.rules[].condition

`GcpOrgPolicyRuleCondition`

Apply this rule only to resources whose Resource Manager tags match.
Omit for an unconditional rule. A boolean constraint needs exactly one
unconditional rule; conditional rules refine it.

### spec.dryRunPolicy.rules[].condition.expression

`string` · required

A Common Expression Language expression of 1 to 10 tag tests joined by
`||` or `&&`. Each test is either
`resource.matchTag('<org_id>/<tag_key_short_name>', '<tag_value_short_name>')`
(by short names) or `resource.matchTagId('tagKeys/<id>', 'tagValues/<id>')`
(by the ids GcpTagKey and GcpTagValue output). Example:
`resource.matchTag('123456789012/environment', 'prod')`.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.dryRunPolicy.rules[].condition.title

`string`

A short label for the condition, shown in the console and audit logs.

### spec.dryRunPolicy.rules[].condition.description

`string`

A longer explanation of what the condition selects and why.

### spec.dryRunPolicy.rules[].condition.location

`string`

Where the expression came from, for error reporting (a file name and
position, or any free text).

### spec.dryRunPolicy.rules[].parameters

`string`

For MANAGED constraints that declare parameters (the `*.managed.*`
family): the parameter values as one JSON object, typed as the
constraint defines them -- e.g.
`{"allowedLocations": ["us-east1", "us-west1"], "allowAll": true}`.
Google validates the object against the constraint at apply time;
the provider validates only that it parses as JSON.

- rule: parameters must be a JSON object, e.g. {"allowedLocations": ["us-east1"]}

### spec.deletionPolicy

`string`

What destroying this resource does to the policy in GCP:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the policy is deleted and the scope falls back to what
               it inherits from above (or the constraint's default)
  "PREVENT" -- destroy FAILS; the guard for the guardrails a landing
               zone depends on
  "ABANDON" -- the policy is removed from management but keeps being
               enforced in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `exactly_one_constraint`: set exactly one of constraint (a predefined or managed constraint such as compute.disableSerialPortAccess) or custom_constraint (a reference to a GcpOrgPolicyCustomConstraint)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpOrgPolicy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | The policy's full resource name, `{scope}/policies/{constraint}` -- e.g. `projects/123456789012/policies/compute.disableSerialPortAccess`. Google returns the project scope as the project NUMBER even when the policy was created with the project ID. |
| `status.outputs.etag` | `string` | The policy's etag: an opaque version marker Google changes on every update, useful for detecting out-of-band edits. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.scope.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.scope.folderId` | GcpFolder | `status.outputs.folder_id` |
| `spec.customConstraint` | GcpOrgPolicyCustomConstraint | `status.outputs.constraint` |

## See Also

- [Overview](../README.md)
