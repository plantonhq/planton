# GCP Service Account IAM Member

Grants one role, to one identity, ON a service account — controlling who may USE or MANAGE the account itself. A GCP service account is both an identity (it holds roles elsewhere) and a resource (identities hold roles ON it); this kind covers the resource side: `roles/iam.workloadIdentityUser` (federation — the keyless-authentication hop), `roles/iam.serviceAccountTokenCreator` (minting short-lived tokens), and `roles/iam.serviceAccountUser` (actAs — deploying workloads that attach the account). Additive like every Planton grant: the (role, member) pair merges without touching anyone else's bindings, and removal subtracts only that pair. Prefer these account-scoped grants over their project-level equivalents — a project-level serviceAccountUser grant allows acting as EVERY service account in the project.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Service Account IAM Member Binding** -- a `serviceaccount.IAMMember` merging the (role, member) pair into the target service account's IAM policy, with an optional IAM Condition attached

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials permitted to set IAM policy on the target service account (e.g. `roles/iam.serviceAccountAdmin`). Map it as the default for your environment, or specify it explicitly.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **The target service account** must exist — reference a GcpServiceAccount Infra Component or provide its full resource name (`projects/<project>/serviceAccounts/<email>`). There is no separate project field: the account's project is embedded in the name.
- **For federation grants**: a workload identity pool and provider whose `principalSet://` subject you are granting.

## Deploy

### Console

Open the deployment store, find **GCP Service Account IAM Member**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the grant definition. Start from the **GitHub Workload Identity Impersonation** preset in the [Presets](#presets) tab for keyless CI/CD.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpServiceAccountIamMember
metadata:
  name: github-deploy-impersonation
  org: acme-corp
  env: prod
spec:
  serviceAccountId:
    value: projects/acme-prod-12345/serviceAccounts/deployer@acme-prod-12345.iam.gserviceaccount.com
  role:
    value: roles/iam.workloadIdentityUser
  member:
    value: principalSet://iam.googleapis.com/projects/123456/locations/global/workloadIdentityPools/github/attribute.repository/acme/orders-api
```

```shell
planton apply -f gcp-service-account-iam-member.yaml
```

This lets the GitHub repository impersonate the deploy account with no exported key anywhere. An Infra Job tracks the provisioning in real time.

### InfraChart

The composed form is where this kind shines — wire the grant to resources deployed in the same InfraPipeline:

```yaml
spec:
  serviceAccountId:
    valueFrom:
      kind: GcpServiceAccount
      name: runtime-sa
      fieldPath: status.outputs.name
  role:
    value: roles/iam.serviceAccountUser
  member:
    valueFrom:
      kind: GcpServiceAccount
      name: deployer-sa
      fieldPath: status.outputs.member
```

The InfraPipeline resolves the dependency graph, deploys both accounts first, then provisions the actAs grant with the resolved values.

## Key Configuration

These are the most important decisions when configuring a grant. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The target account** -- the FULL resource name (`projects/<project>/serviceAccounts/<email>`); a GcpServiceAccount's `name` output is exactly this value (not the bare email — that is the `email` output).

**The usage role** -- `workloadIdentityUser` for federation (external identities impersonating the account), `serviceAccountTokenCreator` for token minting (impersonation chains, privileged brokers), `serviceAccountUser` for actAs (deployers attaching the account to Cloud Run/GCE/Cloud Functions workloads). Custom roles are also grantable by their fully-qualified name.

**Member format** -- `principalSet://`/`principal://` federation subjects, `serviceAccount:<email>` (another account's `member` output), `user:`/`group:`/`domain:` literals. Deleted principals are not grantable; format validation happens at deploy time because values usually arrive through references.

**IAM Condition** -- an optional CEL expression scoping WHEN the grant applies; a time-boxed expiry is the everyday case for impersonation (break-glass access that removes itself). The condition is part of the grant's identity.

**Everything replaces atomically** -- an IAM grant has no update. Any change replaces the grant (a brief moment where impersonation fails); the replacement destroys nothing and is GCP's own designed change workflow.

**GKE workloads** -- for Kubernetes service accounts, use GcpGkeWorkloadIdentityBinding instead: it models the KSA-to-GSA pair end-to-end.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpServiceAccount** (target) | `serviceAccountId` | `status.outputs.name` |
| **GcpIamCustomRole** (optional) | `role` | `status.outputs.name` |
| **GcpServiceAccount** (member) | `member` | `status.outputs.member` |

### What This Kind Provides

This kind has no outputs that downstream Infra Components consume. `status.outputs` echoes the fully resolved grant tuple (`service_account_id`, `role`, `member`) plus the account policy `etag` at the moment the grant merged — audit visibility for values that usually arrive through references, not composition inputs. Downstream resources reference the service account itself, never the grant.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**GitHub Workload Identity impersonation** -- keyless CI/CD: the repository's `principalSet://` subject gains `workloadIdentityUser` on the deploy account. No key exists anywhere. Start from the **GitHub Workload Identity Impersonation** preset.

**Token creator grant** -- a broker or CI account gains `serviceAccountTokenCreator` on a target account, minting short-lived credentials instead of holding long-lived ones. Start from the **Token Creator Grant (Cross-Account Impersonation)** preset.

**Deployer act-as** -- the deploy identity gains `serviceAccountUser` on the runtime account, unblocking Cloud Run/GCE deployments that attach it. Start from the **Deployer actAs Grant** preset.

## Works With

- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- the target (its `name` output) and often the member (its `member` output)
- [**GCP IAM Custom Role**](/infra-catalog/gcp-iam-custom-role) -- its `name` output feeds the role field for custom usage bundles
- [**GCP Project IAM Member**](/infra-catalog/gcp-project-iam-member) -- the project-scoped sibling: grants on a project instead of on an account
- [**GCP GKE Workload Identity Binding**](/infra-catalog/gcp-gke-workload-identity-binding) -- the purpose-built alternative for GKE workload federation
