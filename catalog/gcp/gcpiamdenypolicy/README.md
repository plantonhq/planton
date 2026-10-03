# GCP IAM Deny Policy

Creates an IAM deny policy — rules that BLOCK principals from using specific permissions regardless of any role grants they hold. Deny always outranks allow: a permission denied here cannot be used even by a project owner, which is what makes deny policies the guardrail layer (protect break-glass secrets, forbid destructive APIs org-wide) rather than another access grant. The policy attaches to a project, folder, or organization and applies to the attach point and everything below it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Deny Policy** -- a `google_iam_deny_policy` attached to the configured parent, carrying the deny rules (denied principals, denied permissions, exceptions, and optional conditions)

No API-enablement resource is created — deny policies live on the IAM v2 surface, which is always on.

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target attach point.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Organization

- **IAM**: the deploying principal's permissions are listed in [`iac/permissions.yaml`](iac/permissions.yaml) — note they must be granted at the ORGANIZATION level even for project-attached policies (the manifest notes explain why).
- **Supported permissions**: only permissions on [Google's supported-permissions list](https://cloud.google.com/iam/docs/deny-permissions-support) can be denied — the deny API rejects others.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpIamDenyPolicy
metadata:
  name: guard-secret-access
spec:
  rules:
    - description: Nobody reads break-glass secrets except the break-glass account
      denyRule:
        deniedPrincipals:
          - principalSet://goog/public:all
        exceptionPrincipals:
          - principal://iam.googleapis.com/projects/-/serviceAccounts/break-glass@my-project.iam.gserviceaccount.com
        deniedPermissions:
          - secretmanager.googleapis.com/versions.access
```

```shell
planton apply -f deny-policy.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `rules` | `[]object` | The deny rules — each names the principals denied, the permissions they are denied, any exceptions, and an optional condition. | Min 1 entry; each entry requires `denyRule` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `parent` | `object` | provider default project | Attach point: exactly one of `projectId` (StringValueOrRef, can reference a GcpProject), `folderId` (StringValueOrRef, can reference a GcpFolder), or `organizationId`. Empty means the provider's default project. The module renders the URL-ENCODED full resource name GCP's API expects, so manifests never hand-assemble it. |
| `policyName` | `string` | `metadata.name` | The policy's resource ID. Immutable: changing it destroys and recreates the policy. |
| `displayName` | `string` | `""` | Human-readable name shown in consoles. |
| `deletionPolicy` | `string` | `DELETE` | What destroy does: `DELETE`, `PREVENT` (refuse — protects a guardrail whose silent removal re-opens the surface it guards), or `ABANDON` (keep denying, drop from management). |

### Rule Fields (`rules[].denyRule`)

| Field | Type | Description |
|-------|------|-------------|
| `deniedPrincipals` | `[]string` | Identities denied, in the v2 principal formats: `principalSet://goog/public:all` (everyone), `principal://goog/subject/{email}` (one user), `principalSet://goog/group/{group-email}` (a group), `principal://iam.googleapis.com/projects/-/serviceAccounts/{email}` (a service account), `principalSet://cloudresourcemanager.googleapis.com/organizations/{org-id}` (everyone in the org). |
| `exceptionPrincipals` | `[]string` | Identities EXCLUDED from the rule even when `deniedPrincipals` covers them — the break-glass carve-out. |
| `deniedPermissions` | `[]string` | Permissions denied, as `{service-fqdn}/{resource}.{verb}` — e.g. `secretmanager.googleapis.com/versions.access`. Only permissions on Google's supported-permissions list. |
| `exceptionPermissions` | `[]string` | Permissions EXCLUDED from `deniedPermissions` — a permission in both lists is NOT denied. |
| `denialCondition` | `object` | Optional CEL condition on resource tags scoping when the denial applies (e.g. `!resource.matchTag('12345678/env', 'sandbox')` denies everywhere EXCEPT tagged sandboxes). Requires `expression`; `title` and `description` are optional. |

### Validation Rules

- **At most one parent arm**: set at most one of `parent.projectId`, `parent.folderId`, `parent.organizationId`; all empty means the provider's default project.
- **Every rule needs a `denyRule`** body; a `denialCondition`, when present, needs an `expression`.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `policy_name` | `string` | `{url-encoded-parent}/{policy_name}` — the handle gcloud and the v2 policies API reference the policy by |
| `etag` | `string` | The policy's current etag — changes on every update; useful for optimistic-concurrency tooling reading the policy out of band |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Deny always outranks allow** — even project owners are blocked. Test the exception list before applying a broad denial; a deny policy with a wrong `exceptionPrincipals` entry locks out the account meant to keep access.
- **Deny policies are managed with org-scoped credentials** — see [`iac/permissions.yaml`](iac/permissions.yaml) for the deploying principal's permissions and why project-level credentials cannot manage them.
- **Only supported permissions can be denied** — check Google's supported-permissions list; most but not all IAM permissions are deniable.
- **Prefer `deletionPolicy: PREVENT` for production guardrails** — silently removing a deny policy re-opens the surface it guards, with no incident-side symptom until someone uses the re-opened permission.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- [GcpProject](/docs/catalog/gcp/gcpproject) — provides a project attach point via ValueFromRef
- [GcpProjectIamMember](/docs/catalog/gcp/gcpprojectiammember) — the allow-policy side; deny policies override whatever it grants
- [GcpSecretManagerSecret](/docs/catalog/gcp/gcpsecretmanagersecret) — the break-glass secrets a deny policy typically guards

## Additional Resources

- [IAM Deny Policies Overview](https://cloud.google.com/iam/docs/deny-overview)
- [Permissions Supported in Deny Policies](https://cloud.google.com/iam/docs/deny-permissions-support)
- [Principal Identifiers for Deny Policies](https://cloud.google.com/iam/docs/principal-identifiers)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
