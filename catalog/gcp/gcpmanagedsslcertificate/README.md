# GCP Managed SSL Certificate

Deploys a Google-managed SSL certificate (`google_compute_managed_ssl_certificate`) — a global Compute Engine SSL certificate whose private key and issuance are handled entirely by Google. Attach its `self_link` to a target HTTPS proxy to terminate TLS at a global external Application Load Balancer without ever handling key material yourself.

## What Gets Created

A single global Google-managed SSL certificate covering the domains you specify. Google provisions and renews the certificate asynchronously once DNS for each domain points at the load balancer's IP.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing GCP project** — referenced via `projectId` (or the provider's default project)
- **IAM permissions** — any role carrying `compute.sslCertificates.*` on the target project
- **DNS control** — each domain must eventually point at the load balancer's IP for provisioning to complete (the certificate object can be created before DNS is ready)

## Quick Start

Create a file `cert.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpManagedSslCertificate
metadata:
  name: app-cert
spec:
  projectId:
    value: my-gcp-project-123
  domains:
    - app.example.com
  description: TLS for the production app load balancer
```

Deploy:

```shell
planton apply -f cert.yaml
```

Reference the certificate's `self_link` from a target HTTPS proxy's `sslCertificates` list to terminate TLS at the load balancer.

## Configuration Reference

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default project | GCP project. Can reference a GcpProject. Immutable. |
| `certificateName` | `string` | `metadata.name` | Cloud-side name (RFC1035). Immutable. |
| `description` | `string` | `""` | What this certificate secures. Immutable. |
| `domains` | `string[]` | — (required, 1-100) | FQDNs the certificate is valid for. No wildcards. Immutable. |
| `deletionPolicy` | `string` | `DELETE` | What happens on destroy: `DELETE`, `PREVENT`, or `ABANDON` (leave in GCP — useful mid-rotation handoff). |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `self_link` | `string` | Self-link URI — the value a target HTTPS proxy references in `ssl_certificates` |
| `certificate_name` | `string` | Name of the SSL certificate in GCP |
| `certificate_id` | `string` | Server-assigned numeric ID of the certificate |
| `expire_time` | `string` | Expiry time in RFC3339 format; empty until provisioning completes |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md).

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md).

## Important Notes

- **Fully immutable**: name and domains are ForceNew — any change destroys and recreates the certificate. Because a cert attached to a proxy cannot be deleted, rotate create-before-destroy (create the replacement, repoint the proxy, then destroy the old cert).
- **DNS-gated provisioning**: creation returns immediately but the certificate stays PROVISIONING until each domain's DNS points at the load balancer. Until then the domain serves Google's default certificate and `expire_time` stays empty.
- **No wildcards**: Google-managed certificates do not support `*.` wildcard domains.
- **No sensitive fields**: Google manages the private key; nothing in this spec is marked sensitive.

## Related Kinds

- [GcpUrlMap](/docs/catalog/gcp/gcpurlmap) — routes traffic behind the HTTPS proxy
- [GcpBackendService](/docs/catalog/gcp/gcpbackendservice) — backends the load balancer targets
- [GcpGlobalAddress](/docs/catalog/gcp/gcpglobaladdress) — static IP for DNS A records
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project that owns the certificate

## Additional Resources

- [Google-managed SSL certificates overview](https://cloud.google.com/load-balancing/docs/ssl-certificates/google-managed-ssl)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
