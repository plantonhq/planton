# GcpIamDenyPolicy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpIamDenyPolicySpec defines an IAM DENY policy — rules that BLOCK
principals from using specific permissions regardless of any role
grants they hold. Deny always outranks allow: a permission denied here
cannot be used even by a project owner, which is what makes deny
policies the guardrail layer (protect break-glass secrets, forbid
destructive APIs org-wide) rather than another access grant.

Attach point: a project, folder, or organization. The policy applies
to the attach point and everything below it. Creating deny policies
requires org-level permissions (roles/iam.denyAdmin) even for
project-attached policies.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpIamDenyPolicy
metadata:
  name: my-sample-deny-policy
spec:
  # Where the policy attaches — at most ONE of projectId, folderId, or
  # organizationId; the policy applies to the attach point and everything
  # below it. Omit the whole message to attach to the provider's default
  # project. The module renders the URL-ENCODED full resource name GCP's
  # API expects, so manifests never hand-assemble it.
  parent:
    projectId:
      value: my-gcp-project-123
    # folderId:                         # attach to a folder instead
    #   value: "987654321098"
    # organizationId: "123456789012"    # attach to the organization instead

  # The policy's resource ID (the last segment of its name). Defaults to
  # metadata.name when omitted. Immutable: changing it destroys and
  # recreates the policy.
  policyName: my-sample-deny-policy

  # Shown in consoles.
  displayName: Sample deny policy

  # The deny rules (min 1). Deny always outranks allow — a permission
  # denied here cannot be used even by a project owner.
  rules:
    - # What this rule guards and why — for the operator auditing later.
      description: Nobody reads secret versions except the break-glass account
      denyRule:
        # Identities denied, in the v2 principal formats
        # (principalSet://goog/public:all, principal://goog/subject/{email},
        # principalSet://goog/group/{email}, ...).
        deniedPrincipals:
          - principalSet://goog/public:all
        # Identities EXCLUDED from the rule even when deniedPrincipals
        # covers them — the break-glass carve-out.
        exceptionPrincipals:
          - principal://iam.googleapis.com/projects/-/serviceAccounts/break-glass@my-gcp-project-123.iam.gserviceaccount.com
        # Permissions denied, as {service-fqdn}/{resource}.{verb}. Only
        # permissions on Google's supported-permissions list.
        deniedPermissions:
          - secretmanager.googleapis.com/versions.access
          - secretmanager.googleapis.com/versions.destroy
        # Permissions EXCLUDED from deniedPermissions — a permission in
        # both lists is NOT denied.
        exceptionPermissions:
          - secretmanager.googleapis.com/versions.destroy
        # Optional CEL condition on resource tags scoping when the denial
        # applies. The tag key is namespaced by the org's numeric ID.
        denialCondition:
          expression: "!resource.matchTag('12345678/env', 'sandbox')"
          title: exempt-sandboxes
          description: Denial applies everywhere except sandbox-tagged resources

  # What a destroy does: DELETE (default), PREVENT, or ABANDON. Prefer
  # PREVENT for production guardrails — silently removing a deny policy
  # re-opens the surface it guards.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.parent` | `GcpIamDenyPolicyParent` |  |  |  |
| `spec.parent.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.parent.folderId` | `string \| valueFrom` |  |  | GcpFolder (`status.outputs.folder_id`) |
| `spec.parent.organizationId` | `string` |  |  |  |
| `spec.policyName` | `string` |  |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.rules` | `[]GcpIamDenyPolicyRule` | yes |  |  |
| `spec.rules[].description` | `string` |  |  |  |
| `spec.rules[].denyRule` | `GcpIamDenyPolicyDenyRule` | yes |  |  |
| `spec.rules[].denyRule.deniedPrincipals` | `[]string` |  |  |  |
| `spec.rules[].denyRule.exceptionPrincipals` | `[]string` |  |  |  |
| `spec.rules[].denyRule.deniedPermissions` | `[]string` |  |  |  |
| `spec.rules[].denyRule.exceptionPermissions` | `[]string` |  |  |  |
| `spec.rules[].denyRule.denialCondition` | `GcpIamDenyPolicyCondition` |  |  |  |
| `spec.rules[].denyRule.denialCondition.expression` | `string` | yes |  |  |
| `spec.rules[].denyRule.denialCondition.title` | `string` |  |  |  |
| `spec.rules[].denyRule.denialCondition.description` | `string` |  |  |  |
| `spec.rules[].denyRule.denialCondition.location` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.parent

`GcpIamDenyPolicyParent`

Where the policy attaches. Omit entirely to attach to the
provider's default project.

- rule: set at most one of project_id, folder_id, or organization_id (empty means the provider's default project)

### spec.parent.projectId

`string | valueFrom`

Attach to a project — a literal project ID or a reference to a
GcpProject resource.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.parent.folderId

`string | valueFrom`

Attach to a folder: the folder's numeric ID (with or without the
"folders/" prefix) -- a literal, or a reference to a GcpFolder resource
(its folder_id output), so a chart that creates a folder can guard it
with a deny policy in the same deploy.

- references: GcpFolder (`status.outputs.folder_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpFolder, name: <that resource's name>, fieldPath: status.outputs.folder_id}} -- a bare string does not parse

### spec.parent.organizationId

`string`

Attach to an organization: the numeric organization ID.

### spec.policyName

`string`

The policy's resource ID (the last segment of its name). Defaults
to metadata.name when left empty. Immutable: changing it destroys
and recreates the policy.

### spec.displayName

`string`

Human-readable name shown in consoles.

### spec.rules

`[]GcpIamDenyPolicyRule` · required

The deny rules — each names the principals denied, the permissions
they are denied, any exceptions, and an optional condition.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].description

`string`

What this rule guards and why — for the operator auditing the
policy later.

### spec.rules[].denyRule

`GcpIamDenyPolicyDenyRule` · required

The rule body.

- rule: {"required":true}

### spec.rules[].denyRule.deniedPrincipals

`[]string`

The identities denied, in the v2 principal formats:
  principalSet://goog/public:all                       -- everyone
  principal://goog/subject/{email}                     -- one user
  principalSet://goog/group/{group-email}              -- a group
  principal://iam.googleapis.com/projects/-/serviceAccounts/{email}
                                                       -- a service
                                                          account
  principalSet://cloudresourcemanager.googleapis.com/organizations/{org-id}
                                                       -- everyone in
                                                          the org

### spec.rules[].denyRule.exceptionPrincipals

`[]string`

Identities EXCLUDED from the rule even when denied_principals
covers them — the break-glass carve-out (e.g. deny a group but
exempt the on-call account).

### spec.rules[].denyRule.deniedPermissions

`[]string`

The permissions denied, as {service-fqdn}/{resource}.{verb} — e.g.
"secretmanager.googleapis.com/versions.access",
"iam.googleapis.com/roles.delete". Only permissions the deny API
supports may be listed (see Google's supported-permissions list).

### spec.rules[].denyRule.exceptionPermissions

`[]string`

Permissions EXCLUDED from denied_permissions — a permission
appearing in both lists is NOT denied.

### spec.rules[].denyRule.denialCondition

`GcpIamDenyPolicyCondition`

Optional CEL condition scoping when the denial applies, evaluated
on resource tags (e.g.
!resource.matchTag('12345678/env', 'production') denies everywhere
EXCEPT tagged production resources).

### spec.rules[].denyRule.denialCondition.expression

`string` · required

The CEL expression, e.g.
resource.matchTag('12345678/env', 'sandbox').

- rule: {"required":true}

### spec.rules[].denyRule.denialCondition.title

`string`

Short title identifying the condition's purpose.

### spec.rules[].denyRule.denialCondition.description

`string`

What the condition scopes and why.

### spec.rules[].denyRule.denialCondition.location

`string`

Where the expression came from, for error attribution in UIs (e.g.
a file/line marker). Rarely set by hand.

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the policy is deleted; the permissions it denied
               become usable again wherever roles allow them
  "PREVENT" -- destroy FAILS; protects a guardrail whose silent
               removal would re-open the surface it guards
  "ABANDON" -- the policy is removed from management but keeps
               denying in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpIamDenyPolicy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.policy_name` | `string` | The policy's identifier as {url-encoded-parent}/{policy_name} — the handle gcloud and the v2 policies API reference the policy by. |
| `status.outputs.etag` | `string` | The policy's current etag — changes on every update; useful for optimistic-concurrency tooling reading the policy out of band. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.parent.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.parent.folderId` | GcpFolder | `status.outputs.folder_id` |

## See Also

- [Overview](../README.md)
