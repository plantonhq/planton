# GcpCertManagerCert

## Overview

`GcpCertManagerCert` manages one Certificate Manager certificate — the
modern GCP certificate resource that external Application Load Balancers
consume through a target HTTPS proxy's `certificateManagerCertificates`
list or a certificate map.

Exactly one of two arms is configured:

- **managed** — Google provisions and RENEWS the certificate for the
  listed domains. Domain control is proven through referenced
  `GcpCertManagerDnsAuthorization` resources (required for public
  wildcards, and the only way to issue before traffic serves), through a
  referenced `GcpCertManagerIssuanceConfig` that has your private CA pool
  issue it, or — when neither is set — through load-balancer authorization
  once traffic reaches the proxy.
- **selfManaged** — you upload a PEM chain and private key; the key is
  secret material (masked everywhere), and rotation is an in-place update.

The classic compute certificates are separate kinds:
`GcpManagedSslCertificate` (Google-managed classic) and `GcpSslCertificate`
(self-managed classic).

## Key Features

- Managed XOR self-managed arms with pre-deploy coherence validation
  (wildcards require DNS authorization or an issuance config; DNS
  authorizations and issuance config are mutually exclusive; swapped PEM
  material is rejected).
- First-class DNS authorizations: the certificate references
  `GcpCertManagerDnsAuthorization` resources instead of creating hidden
  ones — the authorization, its validation `GcpDnsRecord`, and the
  certificate compose as three explicit nodes.
- `location` (regional certificates) and `scope` (DEFAULT / EDGE_CACHE /
  ALL_REGIONS / CLIENT_AUTH).
- User labels merged beneath platform attribution labels, identically on
  both engines.
- `deletionPolicy` (DELETE / PREVENT / ABANDON) controls what a destroy
  does — PREVENT guards a certificate whose replacement is not yet
  serving; ABANDON hands the certificate off mid-rotation.

## Example Usage

### Managed Certificate, Validated Before Serving

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCertManagerCert
metadata:
  name: web-cert
spec:
  managed:
    domains:
      - app.example.com
    dnsAuthorizations:
      - valueFrom:
          kind: GcpCertManagerDnsAuthorization
          name: app-example-com-auth
          fieldPath: status.outputs.authorization_id
```

Compose the authorization's validation record into the zone:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDnsRecord
metadata:
  name: app-cert-validation
spec:
  managedZone:
    valueFrom:
      kind: GcpDnsZone
      name: example-com
      fieldPath: status.outputs.zone_name
  type: CNAME
  name: "_acme-challenge.app.example.com."
  values:
    - <dns_record_data from the authorization outputs>
```

(Or resolve `name`/`values` from the authorization's `dns_record_name` /
`dns_record_data` outputs via `valueFrom`.)

### Deploy with CLI

```bash
planton pulumi up --manifest certificate.yaml
# or
planton tofu apply --manifest certificate.yaml
```

## Best Practices

1. **Prefer DNS authorization** for production: the certificate reaches
   ACTIVE before any cutover, so migrations are zero-downtime.
2. **One authorization per distinct domain** — an authorization covers its
   domain and that domain's wildcard.
3. **A managed certificate stays PROVISIONING** until validation completes;
   attach it to the proxy only after `managed_state` reports ACTIVE when
   downtime matters.
4. **Rotate self-managed material in place** — consumers reference the
   certificate by name and never notice the swap.

## Related Kinds

- **GcpCertManagerDnsAuthorization** — the domain-control proof this
  certificate references
- **GcpDnsRecord** / **GcpDnsZone** — serve the validation record
- **GcpTargetHttpsProxy** — consumes the certificate by name
- **GcpManagedSslCertificate** / **GcpSslCertificate** — the classic
  compute certificate kinds

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
