# GCP Workload Identity Pool

Deploys the trust boundary of keyless authentication: a Workload Identity Pool lets external identities — GitHub Actions, GitLab CI, AWS workloads, on-prem SAML/X.509 estates — act in Google Cloud with no service-account key anywhere. The pool holds no issuer configuration itself; it is the boundary and the namespace for principals. Attach one GcpWorkloadIdentityPoolProvider per external issuer, then authorize the pool's principals through GcpServiceAccountIamMember grants. GCP soft-deletes pools: a deleted pool's ID stays reserved for about 30 days and cannot be recreated, so pool IDs are long-lived identifiers.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Workload Identity Pool** -- an `iam.WorkloadIdentityPool` in the target project, in FEDERATION_ONLY mode (the default) or TRUST_DOMAIN mode
- **Certificate Issuance** -- created only when `inlineCertificateIssuanceConfig` is specified; wires Certificate Authority Service CA pools for mTLS workload certificates (trust-domain pools)
- **Trust Bundles** -- created only when `inlineTrustConfig` lists foreign trust domains whose certificates this pool accepts
- **Attestation Rules** -- created only when `attestationRules` names the Google Cloud workloads permitted to receive a managed identity (at most 50); GCP applies them through a second API call after the pool create, so a failed apply can leave a pool without its rules — re-apply converges

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials permitted to manage workload identity pools (e.g. `roles/iam.workloadIdentityPoolAdmin`). Map it as the default for your environment, or specify it explicitly.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** to own the pool — federation quotas and IAM principals scope to it. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **GCP Workload Identity Pool**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, the pool's permanent identity, and its operating mode. Start from the **CI Federation Pool** preset in the [Presets](#presets) tab for the most common shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpWorkloadIdentityPool
metadata:
  name: github-actions
  org: acme-corp
  env: prod
spec:
  workloadIdentityPoolId: github-actions
  projectId:
    value: "acme-security-12345"
  displayName: GitHub Actions federation
  description: Federates CI workloads from the acme GitHub org
```

```shell
planton apply -f gcp-workload-identity-pool.yaml
```

This creates the pool in FEDERATION_ONLY mode (GCP's default when `mode` is unset). An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the pool to a GCP project deployed in the same InfraPipeline:

```yaml
spec:
  workloadIdentityPoolId: github-actions
  projectId:
    valueFrom:
      kind: GcpProject
      name: security-project
      fieldPath: status.outputs.project_id
```

The InfraPipeline resolves the dependency graph and deploys the project first. The keyless-CI composition builds on this node: a GcpWorkloadIdentityPoolProvider references the pool's `workload_identity_pool_id` output, and a GcpServiceAccountIamMember grants the pool's principals impersonation — the whole federation story in one chart.

## Key Configuration

These are the most important decisions when configuring a pool. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Pool ID is effectively permanent** -- it appears in every `principal://` string and every grant; recreating the pool invalidates all of them. GCP soft-deletes pools: a deleted pool lingers ~30 days, its ID cannot be reused, and creating against a soft-deleted ID fails outright. Name it like a long-lived identifier (`github-actions`, `partner-aws`).

**Mode** -- unset applies FEDERATION_ONLY: external identities federate into Google Cloud, the mode for keyless CI/CD and cross-cloud auth. TRUST_DOMAIN instead assigns SPIFFE-style managed identities to Google Cloud workloads with mTLS certificates — pools in this mode cannot hold providers. Immutable at the API (updates fail server-side).

**Disabled is the kill switch** -- a disabled pool rejects all token exchanges and existing tokens stop granting access; re-enabling restores them. Prefer disabling over deleting when rotating or investigating.

**One pool per trust boundary** -- "our CI systems", "the partner's AWS estate" — not one per repository. Providers and their attribute conditions do the fine-grained scoping inside the boundary.

**Destroy is a soft delete** -- destroying the pool stops token exchanges immediately, but the pool lingers DELETED for ~30 days and its ID stays reserved. Set `deletionPolicy: PREVENT` on the pool every keyless-CI grant in the project references, or `ABANDON` to hand a live trust boundary to another management plane.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | Full resource name (`projects/<number>/locations/global/workloadIdentityPools/<id>`) | The root of every `principal://`/`principalSet://` IAM member string; the parent providers attach under |
| `workload_identity_pool_id` | The bare pool ID | GcpWorkloadIdentityPoolProvider `workloadIdentityPoolId` field |
| `state` | `ACTIVE` or `DELETED` (soft-deleted, restorable ~30 days) | Health checks on the federation path |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**CI federation pool** -- the everyday shape: a FEDERATION_ONLY pool for keyless CI, ready for a GitHub/GitLab OIDC provider. Start from the **CI Federation Pool** preset.

**Locked-down pool** -- the land-first, arm-later pattern: the pool deploys disabled and a later change enables it once providers and grants are reviewed. Start from the **Locked-Down (Staged) Pool** preset.

## Works With

- [**GCP Workload Identity Pool Provider**](/infra-catalog/gcp-workload-identity-pool-provider) -- attaches one external issuer (GitHub OIDC, AWS, SAML, X.509) to this pool
- [**GCP Service Account IAM Member**](/infra-catalog/gcp-service-account-iam-member) -- grants the pool's principalSet impersonation of a service account (`roles/iam.workloadIdentityUser`)
- [**GCP Project**](/infra-catalog/gcp-project) -- provides the project that owns the pool
