# GcpWorkloadIdentityPool

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpWorkloadIdentityPoolSpec defines a Workload Identity Pool — the container
GCP uses to trust external identities (GitHub Actions, GitLab CI, AWS
workloads, on-prem SAML/X.509 estates) without any service-account keys.

The pool itself holds no issuer configuration; it is the trust boundary and
the namespace for principals. Attach one GcpWorkloadIdentityPoolProvider per
external issuer, then authorize the pool's principals on service accounts or
directly in IAM policies. Keeping the pool a first-class node means one
trust boundary can serve many issuers, and grants reference the pool's
stable resource name.

Lifecycle note: GCP soft-deletes pools. A deleted pool remains for ~30 days
(restorable via UndeleteWorkloadIdentityPool) and its ID cannot be reused
until permanent deletion. Unlike custom roles, creating a pool with a
soft-deleted ID does NOT undelete it — the create fails outright, so plan
pool IDs as long-lived, stable identifiers.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpWorkloadIdentityPool
metadata:
  name: my-sample-workload-identity-pool
spec:
  # Pool ID (4-32 chars; lowercase letters, digits, hyphens; the "gcp-"
  # prefix is reserved). Becomes the final component of the resource name:
  # projects/<number>/locations/global/workloadIdentityPools/<pool_id>
  workloadIdentityPoolId: github-actions

  # GCP project that owns this pool.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # Name shown in the GCP console
  displayName: GitHub Actions

  # What this pool federates and who owns it
  description: Keyless federation for the engineering org's CI pipelines

  # Kill switch — a disabled pool rejects all token exchanges (default false)
  disabled: false

  # Operating mode (default FEDERATION_ONLY; immutable)
  mode: FEDERATION_ONLY

  # Soft-delete the pool on destroy (GCP's default, made explicit): token
  # exchanges stop immediately, the pool is restorable for ~30 days, and its
  # ID stays reserved until permanent deletion.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.workloadIdentityPoolId` | `string` | yes |  |  |
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.displayName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.disabled` | `bool` |  |  |  |
| `spec.mode` | `string` |  | `FEDERATION_ONLY` |  |
| `spec.inlineCertificateIssuanceConfig` | `GcpWorkloadIdentityPoolCertificateIssuance` |  |  |  |
| `spec.inlineCertificateIssuanceConfig.caPools` | `map<string, string>` |  |  |  |
| `spec.inlineCertificateIssuanceConfig.keyAlgorithm` | `string` |  |  |  |
| `spec.inlineCertificateIssuanceConfig.lifetime` | `string` |  |  |  |
| `spec.inlineCertificateIssuanceConfig.rotationWindowPercentage` | `int32` |  |  |  |
| `spec.inlineCertificateIssuanceConfig.useDefaultSharedCa` | `bool` |  |  |  |
| `spec.inlineTrustConfig` | `GcpWorkloadIdentityPoolTrustConfig` |  |  |  |
| `spec.inlineTrustConfig.additionalTrustBundles` | `[]GcpWorkloadIdentityPoolTrustBundle` | yes |  |  |
| `spec.inlineTrustConfig.additionalTrustBundles[].trustDomain` | `string` | yes |  |  |
| `spec.inlineTrustConfig.additionalTrustBundles[].trustAnchors` | `[]GcpWorkloadIdentityPoolTrustAnchor` | yes |  |  |
| `spec.inlineTrustConfig.additionalTrustBundles[].trustAnchors[].pemCertificate` | `string` | yes |  |  |
| `spec.inlineTrustConfig.additionalTrustBundles[].trustDefaultSharedCa` | `bool` |  |  |  |
| `spec.attestationRules` | `[]GcpWorkloadIdentityPoolAttestationRule` |  |  |  |
| `spec.attestationRules[].googleCloudResource` | `string` | yes |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.workloadIdentityPoolId

`string` · required

The ID for the pool, which becomes the final component of its resource
name (projects/<number>/locations/global/workloadIdentityPools/<id>).
4-32 characters of lowercase letters, digits, and hyphens; the prefix
"gcp-" is reserved by Google. Immutable: changing it destroys and
recreates the pool, invalidating every principal and grant that
references the old pool name.

- rule: the prefix 'gcp-' is reserved by Google — choose a pool ID that does not start with it
- rule: {"required":true,"string":{"pattern":"^[a-z0-9-]{4,32}$"}}

### spec.projectId

`string | valueFrom`

The GCP project that owns this pool. Federation quotas and IAM principals
are scoped to this project.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable: changing it destroys and recreates the pool.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.displayName

`string`

Human-readable name shown in the GCP console (max 32 characters). Mutable.

- rule: {"string":{"maxLen":"32"}}

### spec.description

`string`

What this pool federates and who owns it — write it for the operator
auditing trust boundaries later (max 256 characters). Mutable.

- rule: {"string":{"maxLen":"256"}}

### spec.disabled

`bool`

Emergency kill switch: a disabled pool rejects all token exchanges and
existing tokens stop granting access; re-enabling restores them. Prefer
disabling over deleting when rotating or investigating — deletion starts
the 30-day soft-delete clock and blocks ID reuse. Mutable.

### spec.mode

`string` · optional (explicit presence)

How the pool operates. FEDERATION_ONLY (the default) federates external
identities into Google Cloud and is the right mode for keyless CI/CD and
cross-cloud auth. TRUST_DOMAIN assigns managed identities to Google Cloud
workloads (SPIFFE-style ns/<namespace>/sa/<workload> subjects) — pools in
this mode cannot hold providers. (A third mode, SYSTEM_TRUST_DOMAIN,
exists only for pools Google itself manages and cannot be created, so it
is not accepted here.)
Immutable in the API: the console may accept an edit attempt but the
update fails server-side — create a new pool to change modes.

- default: `FEDERATION_ONLY`
- rule: mode must be FEDERATION_ONLY or TRUST_DOMAIN (SYSTEM_TRUST_DOMAIN pools are Google-managed and cannot be created)

### spec.inlineCertificateIssuanceConfig

`GcpWorkloadIdentityPoolCertificateIssuance`

Configuration for issuing mutual-TLS (mTLS) workload certificates to the
identities in this pool — the certificate half of a TRUST_DOMAIN pool.
Leave unset for token-exchange federation (FEDERATION_ONLY pools).

- rule: choose exactly one certificate authority source: ca_pools (your own CA Service pools) or use_default_shared_ca

### spec.inlineCertificateIssuanceConfig.caPools

`map<string, string>`

Maps a cloud region to the Certificate Authority Service CA pool (full
resource path projects/<project>/locations/<location>/caPools/<pool>)
that issues certificates for workloads in that region. The region in the
key must match the CA pool's own region. Exactly one of ca_pools or
use_default_shared_ca supplies the signing authority.

### spec.inlineCertificateIssuanceConfig.keyAlgorithm

`string` · optional (explicit presence)

Key algorithm for the generated certificate key pairs. Defaults
server-side to ECDSA_P256 — the right choice unless a legacy verifier
requires RSA.

- rule: key_algorithm must be one of RSA_2048, RSA_3072, RSA_4096, ECDSA_P256, or ECDSA_P384

### spec.inlineCertificateIssuanceConfig.lifetime

`string` · optional (explicit presence)

Lifetime of issued workload certificates in seconds, formatted like
"86400s". Must be between 86400s (24 hours) and 2592000s (30 days);
defaults server-side to 86400s. Shorter lifetimes shrink the blast radius
of a leaked certificate at the cost of more rotation traffic.

- rule: {"string":{"pattern":"^[0-9]+s$"}}

### spec.inlineCertificateIssuanceConfig.rotationWindowPercentage

`int32` · optional (explicit presence)

Percentage of remaining certificate lifetime at which rotation begins.
Must be between 50 and 80; defaults server-side to 50 (rotate at half
life). Raise it only if workloads tolerate very tight rotation windows.

- rule: {"int32":{"lte":80,"gte":50}}

### spec.inlineCertificateIssuanceConfig.useDefaultSharedCa

`bool`

Issue certificates from the GCP-provisioned default shared CA in the
workload's own region instead of your own CA Service pools — the
zero-setup path for managed-identity trust domains. Exactly one of
ca_pools or use_default_shared_ca supplies the signing authority.

### spec.inlineTrustConfig

`GcpWorkloadIdentityPoolTrustConfig`

Additional trust domains whose certificates this pool's trust domain
accepts. A trust domain always trusts itself; list only foreign domains.

### spec.inlineTrustConfig.additionalTrustBundles

`[]GcpWorkloadIdentityPoolTrustBundle` · required

Trust bundles keyed by foreign trust domain (e.g. "example.com").
Maximum 10 entries. A trust domain automatically trusts itself and must
not be listed here.

- rule: {"repeated":{"minItems":"1","maxItems":"10"}}

### spec.inlineTrustConfig.additionalTrustBundles[].trustDomain

`string` · required

The foreign trust domain being trusted (e.g. "example.com").

- rule: {"required":true}

### spec.inlineTrustConfig.additionalTrustBundles[].trustAnchors

`[]GcpWorkloadIdentityPoolTrustAnchor` · required

Trust anchors for the domain: incoming end-entity certificates must chain
up to one of these. GCP requires at least one PEM anchor per bundle even
when trust_default_shared_ca is enabled — the shared CA is ADDED to the
bundle, never a substitute for it.

- rule: {"repeated":{"minItems":"1"}}

### spec.inlineTrustConfig.additionalTrustBundles[].trustAnchors[].pemCertificate

`string` · required

PEM certificate of the PKI used for validation. Must contain exactly one
CA certificate (root or intermediate). This is public key material, not a
secret.

- rule: {"required":true}

### spec.inlineTrustConfig.additionalTrustBundles[].trustDefaultSharedCa

`bool`

Additionally include the GCP-managed regional root certificates (the
default shared CA) in this bundle's trust set, so certificates issued
by use_default_shared_ca pools in the foreign domain are accepted.
Only meaningful for managed-identity (TRUST_DOMAIN) pools.

### spec.attestationRules

`[]GcpWorkloadIdentityPoolAttestationRule`

Which workloads may RECEIVE a managed identity from this pool: each
rule names a single Google Cloud workload resource (for example
"//run.googleapis.com/projects/123/type/Service/*"), and matching
workloads are issued the identity. GCP caps a pool at 50 rules.
Mutable — but note GCP applies the rules through a separate API call
after the pool itself is created, so a failed apply can leave a pool
without its rules; re-apply converges.

- rule: {"repeated":{"maxItems":"50"}}

### spec.attestationRules[].googleCloudResource

`string` · required

A single workload operating on Google Cloud, as a full resource name
with optional trailing wildcard — for example
"//run.googleapis.com/projects/123/type/Service/*".

- rule: {"required":true}

### spec.deletionPolicy

`string`

What happens to the pool in GCP when this resource is destroyed.
  "DELETE"  -- (GCP's default when unset) the pool is soft-deleted:
               token exchanges stop immediately, the pool is
               restorable for ~30 days, and its ID stays reserved
               until permanent deletion
  "PREVENT" -- destroy FAILS; protects the trust boundary every
               keyless-CI grant in the project references
  "ABANDON" -- the pool is removed from management but keeps
               federating in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpWorkloadIdentityPool, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | The full resource name: projects/<project_number>/locations/global/workloadIdentityPools/<pool_id>. This is the handle IAM principals are built from — principal://iam.googleapis.com/<name>/subject/<subject> and principalSet://iam.googleapis.com/<name>/attribute.<attr>/<value> — and the parent under which providers are created. |
| `status.outputs.workload_identity_pool_id` | `string` | The bare pool ID (the spec's workload_identity_pool_id, echoed for composing tooling that addresses the pool by short ID — providers reference the pool through this output). |
| `status.outputs.state` | `string` | The pool lifecycle state: ACTIVE, or DELETED while soft-deleted (GCP retains deleted pools for ~30 days; a DELETED pool rejects all token exchanges and blocks reuse of its ID). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpGkeFleetFeature | `spec.workloadidentity.scopeTenancyPool` | `status.outputs.name` |
| GcpWorkloadIdentityPoolProvider | `spec.workloadIdentityPoolId` | `status.outputs.workload_identity_pool_id` |

## See Also

- [Overview](../README.md)
