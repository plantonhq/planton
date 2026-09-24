# GCP Private CA Certificate

Issue a certificate from your private CA as code -- for a load balancer, a server, a device, or a client -- from a signing request or from structured fields, with the signed certificate and its chain as outputs and revocation on destroy.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Certificate** -- a `privateca_certificate` signed by the pool (or the named authority in it), optionally through a template

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with certificate permissions (`roles/privateca.certificateManager`) on the pool. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpPrivateCaPool`** (Enterprise tier) with an enabled **`GcpPrivateCaCertificateAuthority`**.

## Deploy

### Console

Open the deployment store, find **GCP Private CA Certificate**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **TLS Server From Config** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPrivateCaCertificate
metadata:
  name: api-server
  org: acme-corp
  env: prod
spec:
  location: us-central1
  pool:
    valueFrom:
      kind: GcpPrivateCaPool
      name: internal-servers
  lifetime: 2592000s
  pemCsr: |
    -----BEGIN CERTIFICATE REQUEST-----
    ...
    -----END CERTIFICATE REQUEST-----
```

```shell
planton apply -f certificate.yaml
```

This has the `internal-servers` pool sign a 30-day certificate for the CSR. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference the pool from `pool`, optionally the authority from `certificateAuthority` and a template from `certificateTemplate`; downstream resources read `pem_certificate` and `pem_certificate_chain`.

## Key Configuration

These are the most important decisions when configuring this component. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**CSR or config** -- a CSR when the workload already produces one; structured config (subject, SANs, key usage, public key) when you want the certificate's contents in the manifest.

**Template** -- issue through a template to inherit its shape and limits.

**Lifetime** -- how long it is valid, capped by the pool, the template, and the signing authority's own expiry.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpPrivateCaPool** | `pool` | `status.outputs.name` |
| **GcpPrivateCaCertificateAuthority** | `certificateAuthority` | `status.outputs.name` |
| **GcpPrivateCaCertificateTemplate** | `certificateTemplate` | `status.outputs.name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `pem_certificate` | The signed certificate | A server's or load balancer's certificate, beside its private key |
| `pem_certificate_chain` | The chain that verifies it | The intermediate chain served with the certificate |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Server certificate from structured fields** -- start from the **TLS Server From Config** preset.

**Certificate from a CSR** -- start from the **From CSR** preset.

## Works With

- [**GCP Private CA Pool**](/cloud-catalog/gcp-private-ca-pool) -- the pool that issues it
- [**GCP Private CA Certificate Authority**](/cloud-catalog/gcp-private-ca-certificate-authority) -- the authority that signs it
- [**GCP Private CA Certificate Template**](/cloud-catalog/gcp-private-ca-certificate-template) -- the shape it is issued with
