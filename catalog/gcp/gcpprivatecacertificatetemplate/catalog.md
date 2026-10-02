# GCP Private CA Certificate Template

Define a certificate shape once -- a TLS server, an mTLS client, a SPIFFE workload -- and let every team issue certificates that match it from any private CA pool in the region, without repeating the X.509 details or trusting each request to get them right.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `privateca.googleapis.com` on the project (never disabled on destroy)
- **Certificate template** -- a `privateca_certificate_template` with its predefined values, identity constraints, passthrough extensions, and maximum lifetime

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with CA Service template admin permissions (`roles/privateca.templateAdmin`) on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- None.

## Deploy

### Console

Open the deployment store, find **GCP Private CA Certificate Template**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **TLS Server Leaf** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPrivateCaCertificateTemplate
metadata:
  name: tls-server
  org: acme-corp
  env: prod
spec:
  location: us-central1
  maximumLifetime: 2592000s
  predefinedValues:
    caOptions:
      isCa: false
    keyUsage:
      baseKeyUsage:
        digitalSignature: true
        keyEncipherment: true
      extendedKeyUsage:
        serverAuth: true
```

```shell
planton apply -f certificate-template.yaml
```

This creates a 30-day TLS server template. A Stack Job tracks the provisioning in real time.

### InfraChart

Certificates reference the template from `certificateTemplate`.

## Key Configuration

These are the most important decisions when configuring this component. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Predefined values** -- the X.509 fields every certificate issued with the template carries, whatever the request says.

**Identity constraints** -- whether the request's subject and SANs pass through, and a CEL rule they must satisfy.

**Passthrough extensions** -- which extensions a request may add; everything else is dropped.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `name` | The template's full resource name | `GcpPrivateCaCertificate.certificateTemplate` |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**TLS server leaf** -- server authentication with the requested DNS names. Start from the **TLS Server Leaf** preset.

**mTLS client** -- client authentication with SPIFFE URIs only. Start from the **mTLS Client** preset.

## Works With

- [**GCP Private CA Certificate**](/cloud-catalog/gcp-private-ca-certificate) -- certificates issued with it
- [**GCP Private CA Pool**](/cloud-catalog/gcp-private-ca-pool) -- the pools it issues from
