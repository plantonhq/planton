# GCP Private CA Certificate Authority

The signing authority inside a private CA pool -- a root you trust directly, or a subordinate chained to one -- with its key held in Cloud HSM or your own Cloud KMS key, published certificate and CRLs, and a guarded, recoverable teardown.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Certificate authority** -- a `privateca_certificate_authority` in the named pool, activated and enabled on create

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with CA Service admin permissions (`roles/privateca.caManager`) on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpPrivateCaPool`** -- the pool the authority lives in.

## Deploy

### Console

Open the deployment store, find **GCP Private CA Certificate Authority**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Self-Signed Root** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPrivateCaCertificateAuthority
metadata:
  name: root-ca
  org: acme-corp
  env: prod
spec:
  location: us-central1
  pool:
    valueFrom:
      kind: GcpPrivateCaPool
      name: root-pool
  config:
    subjectConfig:
      subject:
        commonName: Acme Root CA
        organization: Acme
    x509Config:
      caOptions:
        isCa: true
      keyUsage:
        baseKeyUsage:
          certSign: true
          crlSign: true
  keySpec:
    algorithm: EC_P384_SHA384
```

```shell
planton apply -f root-ca.yaml
```

This creates a ten-year root with a P-384 HSM key in `root-pool`, enabled and ready to sign. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference the pool from `pool`; a subordinate references its root from `subordinateConfig.certificateAuthority`, and certificates reference the authority that signs them.

## Key Configuration

These are the most important decisions when configuring this component. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Root or subordinate** -- a root signs itself and becomes the trust anchor; a subordinate is signed by another authority (in CA Service, by reference) or by your existing enterprise CA.

**The key** -- `algorithm` gives a Google-managed Cloud HSM key; `cloudKmsKeyVersion` uses a key you own and audit (Enterprise pools).

**The certificate** -- subject, CA constraints (`isCa`, `maxIssuerPathLength`), key usage, name constraints, and lifetime -- all immutable.

**Teardown** -- `deletionProtection`, the 30-day soft delete, and `skipGracePeriod` decide how hard it is to lose a CA by accident.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpPrivateCaPool** | `pool` | `status.outputs.name` |
| **GcpPrivateCaCertificateAuthority** | `subordinateConfig.certificateAuthority` | `status.outputs.name` |
| **GcpKmsKey** | `keySpec.cloudKmsKeyVersion` | `status.outputs.primary_version_name` |
| **GcpGcsBucket** | `gcsBucket` | `status.outputs.bucket_name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `name` | The authority's full resource name | A subordinate's `subordinateConfig.certificateAuthority`; `GcpPrivateCaCertificate.certificateAuthority` |
| `pem_ca_certificate` | The CA certificate | Trust stores of relying services |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Self-signed root** -- the trust anchor of a hierarchy. Start from the **Self-Signed Root** preset.

**Subordinate by reference** -- an issuing authority chained to a root in another pool. Start from the **Subordinate By Reference** preset.

**Your own key** -- a root whose key lives in a Cloud KMS key you control. Start from the **Root With Own KMS Key** preset.

## Works With

- [**GCP Private CA Pool**](/cloud-catalog/gcp-private-ca-pool) -- the pool it lives in
- [**GCP Private CA Certificate**](/cloud-catalog/gcp-private-ca-certificate) -- certificates it signs
- [**GCP KMS Key**](/cloud-catalog/gcp-kms-key) -- your own signing key
