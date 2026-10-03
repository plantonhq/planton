# GCP Cert Manager Trust Config

Creates one Certificate Manager trust config — the root and intermediate CAs, plus individually allowlisted certificates, that a Google Cloud load balancer trusts when it validates client certificates for mutual TLS. A server TLS policy names the trust config and the target HTTPS proxy attaches that policy; a backend authentication config uses one the other way round, to validate the certificates backends present. Updates are in place, so rotating a CA never replaces the config.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Certificate Manager API enablement** (`certificatemanager.googleapis.com`) on the target project (never disabled on destroy)
- **Certificate Manager Trust Config** -- a `google_certificate_manager_trust_config` holding one trust store (trust anchors and intermediate CAs) and any allowlisted certificates

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.

### GCP Project

- **The CA certificates clients chain to**, PEM-encoded — your private root (a GcpPrivateCaPool's CA, a corporate PKI root) or a partner's. Certificates are public material; no private key is ever needed.

## Deploy

### Console

Open the deployment store, find **GCP Cert Manager Trust Config**, and click **Deploy**. The wizard walks the trust config's envelope (project, name, location), then the trust store and any allowlisted certificates. Start from the **Private Root Trust Store** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCertManagerTrustConfig
metadata:
  name: partner-mtls-trust
  org: acme-corp
  env: prod
spec:
  projectId:
    value: acme-prod-12345
  description: CAs that sign partner client certificates
  trustStores:
    - trustAnchors:
        - |
          -----BEGIN CERTIFICATE-----
          <partner-root-ca-pem-body>
          -----END CERTIFICATE-----
  deletionPolicy: PREVENT
```

```shell
planton apply -f trust-config.yaml
```

This creates a global trust config whose `trust_config_id` output a server TLS policy takes. An Infra Job tracks the provisioning in real time.

### InfraChart

In a chart, the project usually arrives by reference:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  trustStores:
    - trustAnchors:
        - |
          -----BEGIN CERTIFICATE-----
          <root-ca-pem-body>
          -----END CERTIFICATE-----
```

The InfraPipeline provisions the project first, then the trust config; a server TLS policy outside the catalog takes the `trust_config_id` output.

## Key Configuration

These are the most important decisions when configuring a trust config. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Trust store** (`trustStores`) -- at most one, as Google allows today. `trustAnchors` are the roots a client chain must build up to; `intermediateCas` complete chains when clients do not send their intermediates. Each entry is one PEM certificate (`-----BEGIN CERTIFICATE-----`).

**Allowlisted certificates** -- individual certificates accepted even when they chain to no anchor: self-signed device certificates, or one partner certificate you do not want to trust a whole CA for. A matching certificate must still parse, prove possession of its private key, and satisfy its SAN constraints.

**Location** -- empty means `global`, which serves global external and cross-region internal Application Load Balancers. A regional load balancer needs a trust config in its own region. Immutable.

**Rotation** -- every change is an in-place update. Rotate a root by adding the new anchor beside the old one, moving clients, then removing the old anchor.

**Deletion policy** -- `PREVENT` makes destroy fail, the guard rail for a trust config live mTLS traffic depends on. Google also refuses deletion while a TLS policy still references the config.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `trust_config_id` | Full resource name (`projects/*/locations/*/trustConfigs/*`) | A server TLS policy's mTLS client-validation trust config, or a backend authentication config |
| `trust_config_name` | The trust config's name in GCP | Auditing |
| `location` | The Certificate Manager location (`global` unless set) | Matching the load balancer's scope |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Private root trust store** -- one root CA (and optionally its intermediates) that signs every client certificate your load balancer should accept. Start from the **Private Root Trust Store** preset.

**Allowlisted device certificates** -- a handful of self-signed device or partner certificates accepted individually, with no CA trusted at all. Start from the **Allowlisted Device Certificates** preset.

## Works With

- [**GCP Target HTTPS Proxy**](/infra-catalog/gcp-target-https-proxy) -- attaches the server TLS policy that names this trust config
- [**GCP Cert Manager Cert**](/infra-catalog/gcp-cert-manager-cert) -- the server certificate the load balancer presents; the trust config validates the client side
- [**GCP Private CA Pool**](/infra-catalog/gcp-private-ca-pool) -- a private CA whose root certificate can be this config's trust anchor
