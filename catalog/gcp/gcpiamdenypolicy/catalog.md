# GCP IAM Deny Policy

Creates an IAM deny policy — rules that BLOCK principals from using specific permissions regardless of any role grants they hold. Deny always outranks allow: a permission denied here cannot be used even by a project owner, which makes deny policies the guardrail layer (protect break-glass secrets, forbid destructive APIs org-wide) rather than another access grant. The policy attaches to a project, folder, or organization and applies to everything below the attach point.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Deny Policy** -- a `google_iam_deny_policy` attached to the configured parent, carrying the deny rules (denied principals, denied permissions, exceptions, and optional tag-based conditions)

No API-enablement resource — deny policies live on the always-on IAM v2 surface.

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target attach point.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **IAM**: creating deny policies requires ORG-LEVEL `roles/iam.denyAdmin` — even for project-attached policies. Deny policies are platform-team infrastructure by Google's own permission design.
- **Supported permissions**: only permissions on Google's supported-permissions list can be denied.

## Deploy

### Console

Open the deployment store, find **GCP IAM Deny Policy**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the attach point, and an inline builder for the deny rules. Start from the **Guard Secret Access** preset in the [Presets](#presets) tab for the most common shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpIamDenyPolicy
metadata:
  name: guard-secret-access
  org: acme-corp
  env: prod
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
  deletionPolicy: PREVENT
```

```shell
planton apply -f deny-policy.yaml
```

This blocks secret-version access for everyone in the provider's default project except the break-glass service account — no role grant can override it, and `PREVENT` makes the guardrail fail destroys instead of silently vanishing. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the attach point can reference a project created in the same chart:

```yaml
spec:
  parent:
    projectId:
      valueFrom:
        kind: GcpProject
        name: my-app-project
        fieldPath: status.outputs.project_id
```

The InfraPipeline creates the project first, then attaches the guardrail to it.

## Key Configuration

These are the most important decisions when configuring a deny policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Attach point** -- `parent` takes exactly one of `projectId`, `folderId`, or `organizationId`; empty means the provider's default project. The module renders the URL-ENCODED full resource name GCP's API expects (e.g. `cloudresourcemanager.googleapis.com%2Fprojects%2Fmy-project`) so manifests never hand-assemble it.

**Rules** -- each rule names the denied principals (v2 principal formats like `principalSet://goog/public:all`), the denied permissions (`{service-fqdn}/{resource}.{verb}`, from Google's supported list), exception lists for both, and an optional CEL `denialCondition` evaluated on resource tags.

**Exceptions are the break-glass path** -- `exceptionPrincipals` carves identities out of the denial even when `deniedPrincipals` covers them; a permission in both `deniedPermissions` and `exceptionPermissions` is NOT denied.

**Deletion policy** -- prefer `PREVENT` for production guardrails: silently removing a deny policy re-opens the surface it guards.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `parent.projectId` | `status.outputs.project_id` |
| **GcpFolder** (optional) | `parent.folderId` | `status.outputs.folder_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `policy_name` | `{url-encoded-parent}/{policy_name}` | Addressing the policy in gcloud and the v2 policies API |
| `etag` | Current etag, changes on every update | Optimistic-concurrency tooling reading the policy out of band |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Guard break-glass secrets** -- deny secret-version access to everyone except the break-glass account, with `PREVENT` so the guardrail cannot vanish silently. Start from the **Guard Secret Access** preset.

**Block destructive APIs org-wide** -- deny project deletion across the organization, with a tag condition exempting sandbox resources. Start from the **Block Destructive APIs** preset.

## Works With

- [**GCP Project**](/cloud-catalog/gcp-project) -- provides a project attach point
- [**GCP Project IAM Member**](/cloud-catalog/gcp-project-iam-member) -- the allow side of IAM; deny policies override its grants
- [**GCP Secret Manager Secret**](/cloud-catalog/gcp-secret-manager-secret) -- the break-glass secrets deny policies typically guard
