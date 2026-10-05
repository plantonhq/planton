# GcpCertManagerTrustConfig

## Overview

`GcpCertManagerTrustConfig` manages one Certificate Manager trust config — the set of certificate authorities a Google Cloud load balancer trusts when it validates CLIENT certificates (mutual TLS), plus individual certificates it accepts outright. It also serves the other direction: a backend authentication config uses a trust config to validate the certificates BACKENDS present.

## Purpose

Mutual TLS asks the client to prove who it is with a certificate. The load balancer needs to know which certificates to believe, and the trust config is that list:

- **Trust anchors** — the root CAs a client certificate chain must build up to.
- **Intermediate CAs** — used to complete a chain when the client does not send its intermediates.
- **Allowlisted certificates** — individual certificates accepted even when they chain to no anchor (self-signed device certificates, one partner's certificate).

A trust config never attaches to a certificate. A server TLS policy names it in its mTLS client-validation settings, and the target HTTPS proxy attaches that policy. Neither of those is a catalog kind yet, so they take this kind's `trust_config_id` output by name.

## Key Features

- One trust store of root and intermediate CAs (Google currently allows one per config)
- Allowlisted certificates for clients that chain to no CA
- Global by default — serves global external and cross-region internal Application Load Balancers; regional locations for regional load balancers
- In-place updates: rotating a CA is a spec change, never a replacement
- Trust-store certificates kept out of plans and logs on both engines (the provider marks them sensitive), even though certificates are public material
- User labels merged beneath the platform's attribution labels, identically on both engines
- `deletionPolicy` (DELETE / PREVENT / ABANDON) controls what a destroy does — PREVENT guards a trust config live mTLS traffic depends on
- Enables the Certificate Manager API on the target project (never disabled on destroy)

## Example Usage

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCertManagerTrustConfig
metadata:
  name: partner-mtls-trust
spec:
  projectId:
    value: my-gcp-project
  description: CAs that sign partner client certificates
  trustStores:
    - trustAnchors:
        - |
          -----BEGIN CERTIFICATE-----
          <partner-root-ca-pem-body>
          -----END CERTIFICATE-----
```

The certificates are public material — never put a private key here.

## Outputs

| Output | Description |
|--------|-------------|
| `trust_config_id` | Full resource name (`projects/{project}/locations/{location}/trustConfigs/{name}`) — what a server TLS policy and a backend authentication config take |
| `trust_config_name` | The trust config's name in GCP |
| `location` | The Certificate Manager location (`global` unless set) |

## Best Practices

1. **Trust a CA, not a fleet of certificates** — an anchor admits every certificate the CA signs; allowlist individual certificates only for clients that chain to nothing.
2. **Keep the location aligned** with the load balancer that validates clients: `global` for global load balancers, the region for regional ones.
3. **Rotate by overlap** — add the new root beside the old one, move clients, then remove the old root. Each step is an in-place update.
4. **Set `deletionPolicy: PREVENT`** on a trust config serving live mTLS traffic; Google also refuses deletion while a TLS policy still references it.

## Related Kinds

- **GcpCertManagerCert** — the server certificate the load balancer presents (a trust config validates the client side)
- **GcpPrivateCaPool** — a private CA whose root can be this config's trust anchor
- **GcpTargetHttpsProxy** — attaches the server TLS policy that names this trust config

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
