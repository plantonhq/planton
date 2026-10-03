# GCP Project IAM Member

Grants one role, to one identity, on one project — the safe, ADDITIVE unit of GCP access control. The grant merges into the project's IAM policy without touching any other member's bindings on the same role, and removal subtracts only this exact (role, member) pair — grants from different charts, teams, and tools never fight over the policy. Authoritative binding/policy management (which clobbers everything not listed) is deliberately not modeled. Both sides of the grant compose through ValueFromRef: a GcpServiceAccount's `member` output and a GcpIamCustomRole's `name` output wire directly into the member and role fields.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Project IAM Member Binding** -- a `projects.IAMMember` merging the (role, member) pair into the target project's IAM policy, with an optional IAM Condition attached

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials permitted to set IAM policy on the target project (e.g. `roles/resourcemanager.projectIamAdmin`). Map it as the default for your environment, or specify it explicitly.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** whose IAM policy receives the grant. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **The identity** receiving the grant must already exist (deleted principals are not grantable).

## Deploy

### Console

Open the deployment store, find **GCP Project IAM Member**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the grant definition. Start from the **Service Account Grant (Predefined Role)** preset in the [Presets](#presets) tab for the most common shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpProjectIamMember
metadata:
  name: worker-log-writer
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  role:
    value: roles/logging.logWriter
  member:
    value: serviceAccount:worker@acme-prod-12345.iam.gserviceaccount.com
```

```shell
planton apply -f gcp-project-iam-member.yaml
```

This merges one binding into the project policy. An Infra Job tracks the provisioning in real time.

### InfraChart

The composed form is where this kind shines — wire the grant to resources deployed in the same InfraPipeline:

```yaml
spec:
  role:
    valueFrom:
      kind: GcpIamCustomRole
      name: log-bucket-writer
      fieldPath: status.outputs.name
  member:
    valueFrom:
      kind: GcpServiceAccount
      name: orders-api-worker
      fieldPath: status.outputs.member
```

The InfraPipeline resolves the dependency graph, deploys the role and service account first, then provisions the grant with the resolved values.

## Key Configuration

These are the most important decisions when configuring a grant. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Member format** -- the prefix declares the identity type: `serviceAccount:<email>`, `user:<email>`, `group:<email>`, `domain:<domain>`, `principal://`/`principalSet://` federation principals, or the bare public principals `allUsers` / `allAuthenticatedUsers` (grant with extreme care -- these expose the role's permissions to the world). Format validation happens at deploy time because values usually arrive through references.

**Role** -- a predefined role (`roles/...`) or a custom role's fully-qualified name (`projects/<project>/roles/<role_id>`). Referencing a GcpIamCustomRole means permission edits on the role propagate to this grant with the grant untouched.

**IAM Condition** -- an optional CEL expression scoping WHEN the grant applies (expiry dates, resource-name prefixes). The condition is part of the grant's identity: the same role with and without a condition are two independent grants.

**Everything replaces atomically** -- an IAM grant has no update. Changing role, member, project, or condition replaces the grant (a brief moment where it does not exist); the replacement destroys nothing and is GCP's own designed change workflow.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpIamCustomRole** (optional) | `role` | `status.outputs.name` |
| **GcpServiceAccount** (optional) | `member` | `status.outputs.member` |

### What This Kind Provides

This kind has no outputs that downstream Infra Components consume. `status.outputs` echoes the fully resolved grant tuple (`project_id`, `role`, `member`) plus the project policy `etag` at the moment the grant merged — audit visibility for values that usually arrive through references, not composition inputs. Downstream resources reference the service account or custom role directly, never the grant.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Service account grant** -- a predefined role to a workload identity; the everyday IaC shape. Start from the **Service Account Grant (Predefined Role)** preset.

**Custom role grant** -- both sides composed by reference: a GcpIamCustomRole's `name` and a GcpServiceAccount's `member`. Start from the **Custom Role Grant (Fully Composed)** preset.

**Conditional grant** -- a human identity with a time-boxed or resource-scoped condition; the access that removes itself. Start from the **Conditional Grant (Time-Bound Access)** preset.

## Works With

- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- its `member` output feeds the member field; its inline role lists are the alternative for account-owned grants
- [**GCP IAM Custom Role**](/infra-catalog/gcp-iam-custom-role) -- its `name` output feeds the role field for least-privilege bundles
- [**GCP Project**](/infra-catalog/gcp-project) -- provides the project whose policy receives the grant
- [**GCP Service Account IAM Member**](/infra-catalog/gcp-service-account-iam-member) -- the account-scoped sibling: grants ON a service account (impersonation, actAs) instead of on a project
