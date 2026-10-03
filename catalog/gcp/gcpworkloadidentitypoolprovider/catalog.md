# GCP Workload Identity Pool Provider

Attaches one external issuer to a Workload Identity Pool — the piece that turns the trust boundary into a working keyless-auth path. Each provider trusts exactly one identity system (a GitHub org's OIDC tokens, an AWS account, a SAML IdP, an X.509 certificate estate), translates its claims into Google attributes through the attribute mapping, and gates which otherwise-valid credentials are accepted through the attribute condition. A pool holds many providers, one per issuer — revoking one never touches the others. Like the pool, GCP soft-deletes providers for about 30 days and blocks ID reuse while soft-deleted, so prefer disabling over deleting during rotation or investigation.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Workload Identity Pool Provider** -- an `iam.WorkloadIdentityPoolProvider` under the target pool, configured with exactly one issuer type (OIDC, AWS, SAML, or X.509), its attribute mapping, and its attribute condition

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials permitted to manage workload identity pool providers (e.g. `roles/iam.workloadIdentityPoolAdmin`). Map it as the default for your environment, or specify it explicitly.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A workload identity pool** for the provider to attach to. Provide the bare pool ID directly or reference a GcpWorkloadIdentityPool Infra Component via ValueFromRef.
- **The issuer's verification material** -- an OIDC issuer URI (or JWKS for private issuers), a 12-digit AWS account ID, SAML IdP metadata XML, or PEM trust anchors. All of it is public verification material, never a secret.

## Deploy

### Console

Open the deployment store, find **GCP Workload Identity Pool Provider**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, the provider identity, the issuer choice, and the attribute mapping + condition. Start from the **GitHub Actions OIDC Provider** preset in the [Presets](#presets) tab for the most common shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpWorkloadIdentityPoolProvider
metadata:
  name: github-oidc
  org: acme-corp
  env: prod
spec:
  workloadIdentityPoolId:
    value: github-actions
  workloadIdentityPoolProviderId: github-oidc
  displayName: GitHub Actions OIDC
  attributeMapping:
    google.subject: assertion.sub
    attribute.repository: assertion.repository
  attributeCondition: assertion.repository_owner == "acme"
  oidc:
    issuerUri: https://token.actions.githubusercontent.com
```

```shell
planton apply -f gcp-workload-identity-pool-provider.yaml
```

This attaches a GitHub Actions OIDC issuer to the `github-actions` pool, scoped so only workflows in the `acme` GitHub org can federate. An Infra Job tracks the provisioning in real time.

### InfraChart

The keyless-CI composition wires pool → provider → grant in one InfraPipeline:

```yaml
spec:
  workloadIdentityPoolId:
    valueFrom:
      kind: GcpWorkloadIdentityPool
      name: github-actions
      fieldPath: status.outputs.workload_identity_pool_id
```

The InfraPipeline deploys the pool first, then this provider, then the GcpServiceAccountIamMember grant that lets the pool's principals impersonate a service account.

## Key Configuration

These are the most important decisions when configuring a provider. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The issuer is exactly one type, fixed at design time** -- `oidc` (the keyless-CI workhorse: GitHub Actions, GitLab CI, Kubernetes clusters), `aws` (a 12-digit account), `saml` (an enterprise IdP's metadata XML), or `x509` (certificates chaining to your trust store). Changing types replaces the provider and invalidates tokens minted for the old audience.

**The attribute condition is the security boundary** -- a CEL gate over `assertion`, `google`, and `attribute` deciding which otherwise-valid credentials federate. Without one, ANY identity the issuer vouches for can attempt federation — on multi-tenant issuers (GitHub, GitLab) that means every one of their customers. Always scope them: `assertion.repository_owner == "acme"`.

**Attribute mapping drives grants** -- mapped attributes become the targets IAM can bind: `principalSet://iam.googleapis.com/<pool>/attribute.<name>/<value>`. OIDC providers must map `google.subject` explicitly (e.g. `assertion.sub`); AWS, SAML, and X.509 fall back to issuer-specific defaults.

**Empty allowed audiences are the safest** (OIDC) -- incoming tokens must then carry the provider's own canonical resource name as `aud`; tokens minted for anything else are rejected.

**Disable, don't delete** -- `disabled: true` rejects new token exchanges immediately (already-issued Google credentials remain valid until they expire) and is fully reversible. Deletion starts the ~30-day soft-delete clock and blocks the ID from reuse. `deletionPolicy: PREVENT` protects the keyless-auth path every pipeline federating through this issuer depends on.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpWorkloadIdentityPool** | `workloadIdentityPoolId` | `status.outputs.workload_identity_pool_id` |
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream consumers use:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | Full resource name (`projects/<number>/…/providers/<id>`) | Prefixed with `//iam.googleapis.com/`, THE AUDIENCE external tokens carry — what a GitHub workflow's `workload_identity_provider` input takes |
| `workload_identity_pool_provider_id` | The bare provider ID | Display, logging |
| `state` | `ACTIVE` or `DELETED` (soft-deleted, restorable ~30 days) | Health checks on the federation path |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**GitHub Actions OIDC** -- issuer `https://token.actions.githubusercontent.com`, subject/repository mapping, and an org-scoping condition. Start from the **GitHub Actions OIDC Provider** preset.

**AWS account** -- a 12-digit account whose workloads federate, with a role-scoping condition over `attribute.aws_role`. Start from the **AWS Account Provider** preset.

**GitLab CI OIDC** -- issuer `https://gitlab.com` with project-path mapping and a namespace-scoping condition. Start from the **GitLab CI OIDC Provider** preset.

## Works With

- [**GCP Workload Identity Pool**](/infra-catalog/gcp-workload-identity-pool) -- the trust boundary this provider attaches to; its `workload_identity_pool_id` output feeds the pool field
- [**GCP Service Account IAM Member**](/infra-catalog/gcp-service-account-iam-member) -- grants the provider's federated principals impersonation of a service account (`roles/iam.workloadIdentityUser`)
- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- the identity federated workloads act as, keylessly
- [**GCP Project**](/infra-catalog/gcp-project) -- provides the project that owns the pool and this provider
