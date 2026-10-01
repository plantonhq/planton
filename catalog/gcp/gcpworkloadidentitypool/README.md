# GCP Workload Identity Pool

Deploys a Workload Identity Pool (`google_iam_workload_identity_pool`) — the trust boundary GCP uses to accept external identities (GitHub Actions, GitLab CI, AWS workloads, enterprise SAML/X.509 estates) without any service-account keys.

## What Gets Created

When you deploy a GcpWorkloadIdentityPool resource, Planton provisions:

- **Workload Identity Pool** — a `google_iam_workload_identity_pool` in the specified project, addressable as `projects/<number>/locations/global/workloadIdentityPools/<poolId>`

The pool itself holds no issuer configuration. It is the namespace for federated principals: attach one GcpWorkloadIdentityPoolProvider per external issuer, then authorize the pool's principals on service accounts or directly in IAM policies.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing GCP project** — referenced via `projectId`
- **IAM permissions** — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying principal needs

## Quick Start

Create a file `pool.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpWorkloadIdentityPool
metadata:
  name: github-actions-pool
spec:
  projectId:
    value: my-gcp-project-123
  workloadIdentityPoolId: github-actions
  displayName: GitHub Actions
  description: Keyless federation for the engineering org's CI pipelines
```

Deploy:

```shell
planton apply -f pool.yaml
```

This creates the pool, ready for a GcpWorkloadIdentityPoolProvider to attach the issuer.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `workloadIdentityPoolId` | `string` | Pool ID; becomes the final component of the resource name. | 4-32 chars; lowercase letters, digits, hyphens; `gcp-` prefix reserved. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default project | GCP project that owns this pool. Can reference a GcpProject resource. Immutable. |
| `displayName` | `string` | `""` | Console-visible name. Max 32 chars. Mutable. |
| `description` | `string` | `""` | What this pool federates and who owns it. Max 256 chars. Mutable. |
| `disabled` | `bool` | `false` | Kill switch: a disabled pool rejects all token exchanges; re-enabling restores access. Mutable. |
| `mode` | `string` | `FEDERATION_ONLY` | `FEDERATION_ONLY` (keyless federation) or `TRUST_DOMAIN` (managed workload identities; such pools cannot hold providers). Immutable. |
| `inlineCertificateIssuanceConfig` | `object` | — | mTLS workload-certificate issuance for TRUST_DOMAIN pools: exactly one CA source — region→CA-pool map (`caPools`, each value a `GcpPrivateCaPool` reference or the pool's full name) or the GCP-provisioned default (`useDefaultSharedCa`) — plus key algorithm, lifetime, rotation window. |
| `inlineTrustConfig` | `object` | — | Additional trust domains (PEM trust anchors per foreign domain) whose certificates this pool's trust domain accepts; each bundle can also trust the GCP shared CA roots (`trustDefaultSharedCa`, additive to its required anchors). |
| `attestationRules` | `list` | `[]` | Which Google Cloud workloads may receive a managed identity from this pool (full resource names, wildcard-capable; max 50). Applied through a second API call after pool create — a failed apply can leave a pool without its rules; re-apply converges. Mutable. |
| `deletionPolicy` | `string` | `DELETE` | Destroy behavior: `DELETE` soft-deletes the pool (restorable ~30 days, ID reserved), `PREVENT` fails the destroy, `ABANDON` unmanages while the pool keeps federating. Mutable. |

## Why Pools Are First-Class

Every principal built from the pool embeds its resource name — `principal://iam.googleapis.com/<pool name>/subject/<subject>` — so the pool is the stable identity boundary that IAM bindings, providers, and provider configurations all reference. One pool typically serves many issuers (a GitHub org and an AWS account can federate into the same boundary), which is exactly why the pool and its providers are separate composable nodes.

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | Full resource name (`projects/<number>/locations/global/workloadIdentityPools/<poolId>`) — the handle principals and providers are built from |
| `workload_identity_pool_id` | `string` | The bare pool ID; a GcpWorkloadIdentityPoolProvider's `workloadIdentityPoolId` field references exactly this |
| `state` | `string` | `ACTIVE`, or `DELETED` while soft-deleted |

## Deployment Methods

Planton supports two deployment methods:

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Immutability**: `workloadIdentityPoolId`, `projectId`, and `mode` are ForceNew — changing any of them destroys and recreates the pool, invalidating every principal built from the old pool name. (The API rejects `mode` updates outright, even though a plan may show one.)
- **Soft delete without undelete-on-create**: GCP retains deleted pools for ~30 days, during which the pool ID cannot be reused — and unlike custom roles, creating a pool with a soft-deleted ID fails rather than undeleting it. Treat pool IDs as long-lived; prefer `disabled: true` over deletion for temporary shutoffs.
- **Providers are separate resources**: issuer configuration (OIDC/AWS/SAML/X.509, attribute mappings, conditions) lives on GcpWorkloadIdentityPoolProvider, one per issuer.
- **Managed workload identities**: `TRUST_DOMAIN` pools additionally support namespaces and managed identities (SPIFFE-style workload identity). Those are separate resources with their own lifecycles; the pool-side surface — `mode`, attestation rules, certificate issuance (own CA pools or the GCP shared CA), and trust config — is fully modeled so the pool spec stays honest.

## Related Components

- [GcpWorkloadIdentityPoolProvider](/docs/catalog/gcp/gcpworkloadidentitypoolprovider) — attaches an external issuer to this pool (references this pool's `workload_identity_pool_id` output)
- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — the identity federated principals typically impersonate
- [GcpProjectIamMember](/docs/catalog/gcp/gcpprojectiammember) — grants roles to `principal://`/`principalSet://` members built from this pool
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project that owns the pool

## Additional Resources

- [Workload Identity Federation](https://cloud.google.com/iam/docs/workload-identity-federation)
- [Manage Workload Identity Pools and Providers](https://cloud.google.com/iam/docs/manage-workload-identity-pools-providers)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
