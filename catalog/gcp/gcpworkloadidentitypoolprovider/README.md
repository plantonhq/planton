# GCP Workload Identity Pool Provider

Deploys a Workload Identity Pool Provider (`google_iam_workload_identity_pool_provider`) — one external issuer trusted by a Workload Identity Pool. This is the piece that turns a pool into a working keyless-auth path: tokens minted by the issuer are exchanged for Google credentials with no service-account keys anywhere.

## What Gets Created

When you deploy a GcpWorkloadIdentityPoolProvider resource, Planton provisions:

- **Workload Identity Pool Provider** — a `google_iam_workload_identity_pool_provider` inside the referenced pool, addressable as `projects/<number>/locations/global/workloadIdentityPools/<poolId>/providers/<providerId>`

One pool holds many providers — one per issuer (a GitHub org, an AWS account, a SAML IdP, an X.509 CA), each with its own attribute mapping and trust condition.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing Workload Identity Pool** — referenced via `workloadIdentityPoolId` (typically a GcpWorkloadIdentityPool resource)
- **IAM permissions** — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying principal needs

## Quick Start

Create a file `provider.yaml` (GitHub Actions, the most common issuer):

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpWorkloadIdentityPoolProvider
metadata:
  name: github-oidc
spec:
  workloadIdentityPoolId:
    valueFrom:
      kind: GcpWorkloadIdentityPool
      name: github-actions-pool
      fieldPath: status.outputs.workload_identity_pool_id
  workloadIdentityPoolProviderId: github-oidc
  displayName: GitHub Actions OIDC
  attributeMapping:
    google.subject: assertion.sub
    attribute.repository: assertion.repository
  attributeCondition: assertion.repository_owner == "my-org"
  oidc:
    issuerUri: https://token.actions.githubusercontent.com
```

Deploy:

```shell
planton apply -f provider.yaml
```

The provider's `name` output is the audience your CI mints tokens for.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `workloadIdentityPoolId` | `StringValueOrRef` | The pool this provider belongs to (bare pool ID). References a GcpWorkloadIdentityPool's `workload_identity_pool_id` output. | Required. Immutable. |
| `workloadIdentityPoolProviderId` | `string` | Provider ID; becomes the final component of the resource name. | 4-32 chars; lowercase letters, digits, hyphens; `gcp-` prefix reserved. Immutable. |
| exactly one of `aws` / `oidc` / `saml` / `x509` | `object` | The external issuer this provider trusts. | The issuer type cannot change on a live provider. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default project | GCP project that owns the pool. Immutable. |
| `displayName` | `string` | `""` | Console-visible name. Max 32 chars. Mutable. |
| `description` | `string` | `""` | Which issuer this trusts and any operational notes. Max 256 chars. Mutable. |
| `disabled` | `bool` | `false` | Kill switch: a disabled provider rejects new token exchanges (already-issued Google credentials remain valid until expiry). Mutable. |
| `attributeMapping` | `map(string)` | issuer defaults | Claim → Google-attribute CEL mappings. **Required for OIDC** (must include `google.subject`); AWS/SAML/X.509 have server-side defaults. |
| `attributeCondition` | `string` | accept all | CEL gate over `assertion`/`google`/`attribute` restricting which otherwise valid credentials are accepted. Max 4096 chars. |
| `deletionPolicy` | `string` | `DELETE` | What destroy does: `DELETE` removes the provider (starting the ~30-day soft-delete/ID-reuse clock), `PREVENT` fails the destroy to protect a live federation path, `ABANDON` leaves it exchanging tokens unmanaged. |

### Issuer Arms

| Arm | Fields | Notes |
|-----|--------|-------|
| `aws` | `accountId` (12 digits) | Workloads presenting AWS credentials from this account can federate. |
| `oidc` | `issuerUri` (required), `allowedAudiences` (max 10), `jwksJson` | Empty audiences means the audience must equal the provider's full canonical resource name — the safest default. `jwksJson` only for issuers unreachable via `.well-known` discovery. |
| `saml` | `idpMetadataXml` (required) | The IdP's configuration metadata XML. |
| `x509` | `trustStore.trustAnchors` (min 1), `trustStore.intermediateCas` | Clients presenting certificates chaining to the anchors can federate. |

## Locking Down Multi-Tenant Issuers

For shared issuers like GitHub Actions (`https://token.actions.githubusercontent.com`), the issuer vouches for *every* GitHub repository on the planet. `attributeCondition` is what scopes trust to yours:

```yaml
attributeCondition: assertion.repository_owner == "my-org"
```

Without a condition on a multi-tenant issuer, any repository could mint tokens your pool accepts — always set one.

## Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | Full provider resource name — **the audience for token exchange**: OIDC tokens set `aud` to this value (with an `//iam.googleapis.com/` prefix), and web-identity provider configurations consume exactly this string |
| `workload_identity_pool_provider_id` | `string` | The bare provider ID |
| `state` | `string` | `ACTIVE`, or `DELETED` while soft-deleted |

## Deployment Methods

Planton supports two deployment methods:

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Immutability**: `workloadIdentityPoolId`, `workloadIdentityPoolProviderId`, and `projectId` are ForceNew, and the issuer TYPE cannot change on a live provider — switching from OIDC to SAML means a new provider. The chosen arm's contents, mappings, and condition update in place.
- **Soft delete without undelete-on-create**: GCP retains deleted providers for ~30 days, during which the provider ID cannot be reused; a create against a soft-deleted ID fails. Prefer `disabled: true` for temporary shutoffs.
- **OIDC mapping requirement**: OIDC providers must map `google.subject` explicitly (`{"google.subject": "assertion.sub"}` is the common form); the spec validates this before deploy.

## Related Kinds

- [GcpWorkloadIdentityPool](/docs/catalog/gcp/gcpworkloadidentitypool) — the pool this provider attaches to
- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — grant `roles/iam.workloadIdentityUser` to this provider's principals to enable impersonation
- [GcpProjectIamMember](/docs/catalog/gcp/gcpprojectiammember) — grants roles directly to `principal://`/`principalSet://` members mapped by this provider

## Additional Resources

- [Workload Identity Federation](https://cloud.google.com/iam/docs/workload-identity-federation)
- [Configure Workload Identity Federation with deployment pipelines](https://cloud.google.com/iam/docs/workload-identity-federation-with-deployment-pipelines)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
