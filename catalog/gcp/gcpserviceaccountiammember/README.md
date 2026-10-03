# GCP Service Account IAM Member

Deploys a single ADDITIVE IAM grant ON a service account (`google_service_account_iam_member`) — one role, to one member, on one service account resource. This is the least-privilege unit of service-account access control: the grant merges into the account's IAM policy without touching any other member's bindings, and removal subtracts only this exact pair.

A service account is both an identity and a resource. This kind covers the resource side — who may USE or MANAGE the account itself: workload identity federation impersonation (`roles/iam.workloadIdentityUser`), short-lived token minting (`roles/iam.serviceAccountTokenCreator`), and deploy-as/actAs (`roles/iam.serviceAccountUser`).

## What Gets Created

When you deploy a GcpServiceAccountIamMember resource, Planton provisions:

- **Service Account IAM Member** — one (role, member[, condition]) entry merged into the target service account's IAM policy

Nothing else in the policy is read as owned or modified — grants made by other charts, teams, or tools are never clobbered.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing service account** — referenced via `serviceAccountId`
- **IAM permissions** — the deploying principal's permissions are listed in [`iac/permissions.yaml`](iac/permissions.yaml)
- **The role and the member must exist** — reference a GcpServiceAccount and/or GcpIamCustomRole, or use literal values

## Quick Start

Create a file `impersonation-grant.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpServiceAccountIamMember
metadata:
  name: github-deployer-impersonation
spec:
  serviceAccountId:
    value: projects/my-gcp-project-123/serviceAccounts/deployer@my-gcp-project-123.iam.gserviceaccount.com
  role:
    value: roles/iam.workloadIdentityUser
  member:
    value: principalSet://iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/github/attribute.repository/my-org/my-repo
```

Deploy:

```shell
planton apply -f impersonation-grant.yaml
```

Or compose with the service account as a first-class node:

```yaml
spec:
  serviceAccountId:
    valueFrom:
      kind: GcpServiceAccount
      name: deployer
      fieldPath: status.outputs.name
  role:
    value: roles/iam.serviceAccountTokenCreator
  member:
    valueFrom:
      kind: GcpServiceAccount
      name: token-broker
      fieldPath: status.outputs.member
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `serviceAccountId` | `StringValueOrRef` | The service account whose IAM policy receives the grant, as its fully-qualified resource name (`projects/<project>/serviceAccounts/<email>`). References a GcpServiceAccount's `name` output by default. Immutable. |
| `role` | `StringValueOrRef` | The role to grant ON the account: a predefined role (typically `roles/iam.workloadIdentityUser`, `roles/iam.serviceAccountTokenCreator`, or `roles/iam.serviceAccountUser`) or a custom role's fully-qualified name. References a GcpIamCustomRole's `name` output by default. Immutable. |
| `member` | `StringValueOrRef` | The identity receiving the grant, in IAM member format. References a GcpServiceAccount's `member` output by default. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `condition` | `object` | — | IAM Condition (`title`, `expression`, optional `description`) restricting when the grant applies. Part of the grant's identity. Immutable. |

There is no project field: the account's project is embedded in its fully-qualified resource name.

### Member Formats

| Format | Grants to |
|--------|-----------|
| `serviceAccount:<email>` | Another service account (cross-SA impersonation) |
| `principal://...` / `principalSet://...` | Workload identity federation principals — the keyless CI/CD and external-workload path |
| `user:<email>` | A Google account |
| `group:<email>` | A Google group |
| `domain:<domain>` | Everyone in a Workspace/Cloud Identity domain |
| `allUsers` / `allAuthenticatedUsers` | Public access — never appropriate on a service account in practice |

Grants to deleted principals (`deleted:...`) are rejected at deploy time.

## Account-Scoped vs Project-Scoped

The same usage roles can be granted at project level (via GcpProjectIamMember), but a project-level `roles/iam.serviceAccountUser` grant allows acting as EVERY service account in the project. Granting on the specific account — this component — is the least-privilege posture and makes the access relationship an explicit edge in the resource graph.

For the GKE-specific impersonation pattern (Kubernetes ServiceAccount → GCP service account via Workload Identity), prefer GcpGkeWorkloadIdentityBinding: it derives the workload-identity principal from the cluster project, namespace, and KSA name so you never assemble the principal string by hand. Reach for this generic component for every non-GKE principal: GitHub Actions federation, cross-SA impersonation, users, and groups.

## Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `service_account_id` | `string` | The account whose policy received the grant (after reference resolution) |
| `role` | `string` | The granted role (after reference resolution) |
| `member` | `string` | The granted member (after reference resolution) |
| `etag` | `string` | The account IAM policy etag after the grant — useful for audit correlation |

## Deployment Methods

Planton supports two deployment methods:

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Everything is immutable**: IAM grants have no update — changing account, role, member, or condition replaces the grant atomically (destroy the old pair, create the new one), which mirrors the underlying API exactly.
- **Conditions are part of the grant's identity**: the same role granted with and without a condition are two independent grants that do not interfere.
- **Only additive grants are modeled**: authoritative per-role bindings and whole-policy writes clobber every grant they do not list and are deliberately not modeled.

## Related Kinds

- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — the account being granted on (its `name` output) and the most common member (its `member` output)
- [GcpWorkloadIdentityPoolProvider](/docs/catalog/gcp/gcpworkloadidentitypoolprovider) — issues the federated principals this grant authorizes for impersonation
- [GcpProjectIamMember](/docs/catalog/gcp/gcpprojectiammember) — the project-scoped counterpart for roles on the project itself
- [GcpGkeWorkloadIdentityBinding](/docs/catalog/gcp/gcpgkeworkloadidentitybinding) — the GKE-specific impersonation convenience

## Additional Resources

- [Service Account Impersonation](https://cloud.google.com/iam/docs/service-account-impersonation)
- [Workload Identity Federation](https://cloud.google.com/iam/docs/workload-identity-federation)
- [IAM Conditions](https://cloud.google.com/iam/docs/conditions-overview)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
