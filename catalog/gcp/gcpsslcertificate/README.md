# GCP SSL Certificate (Self-Managed)

Deploys a self-managed Compute Engine SSL certificate (`google_compute_ssl_certificate` / `google_compute_region_ssl_certificate`) — you bring the PEM certificate chain and private key, and the load balancer presents them to clients. Attach its `self_link` to a target HTTPS (or SSL) proxy exactly like a Google-managed certificate; the two share one API collection and name namespace.

## What Gets Created

A single SSL certificate uploaded from your PEM material. Leave `region` empty for a **global** certificate (global external HTTPS load balancers); set it for a **regional** one (regional external and internal Application Load Balancers).

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing GCP project** — referenced via `projectId` (or the provider's default project)
- **IAM permissions** — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying principal needs
- **A PEM certificate chain and unencrypted private key** — issued by your CA or purchased commercially

## Quick Start

Create a file `cert.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSslCertificate
metadata:
  name: prod-app-cert
spec:
  projectId:
    value: my-gcp-project-123
  description: Wildcard cert from the corporate CA
  certificate: |
    -----BEGIN CERTIFICATE-----
    ...leaf, then intermediates...
    -----END CERTIFICATE-----
  privateKey: |
    -----BEGIN PRIVATE KEY-----
    ...matching unencrypted key...
    -----END PRIVATE KEY-----
```

Deploy:

```shell
planton apply -f cert.yaml
```

Reference the certificate's `self_link` from a target HTTPS proxy's `sslCertificates` list (with an explicit `valueFrom` kind — the list defaults to Google-managed certificates).

## Configuration Reference

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default project | GCP project. Can reference a GcpProject. Immutable. |
| `certificateName` | `string` | `metadata.name` | Cloud-side name (RFC1035); shares one namespace with managed certificates. Immutable. |
| `region` | `string` | `""` (global) | Region for a regional certificate; empty means global. Immutable. |
| `description` | `string` | `""` | What this certificate secures and where it came from. Immutable. |
| `certificate` | `string` | — (required) | PEM chain: leaf first, then intermediates (max 5 certs). Public material. Immutable. |
| `privateKey` | `string` (sensitive) | — (required) | Matching unencrypted PEM key. Write-only in GCP; never in outputs. Immutable. |
| `deletionPolicy` | `string` | `DELETE` | What happens on destroy: `DELETE`, `PREVENT`, or `ABANDON` (leave in GCP — useful mid-rotation handoff). |

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `self_link` | `string` | Self-link URI — the value a target HTTPS proxy references in `ssl_certificates` |
| `certificate_name` | `string` | Name of the SSL certificate in GCP |
| `certificate_id` | `string` | Server-assigned numeric ID of the certificate |
| `expire_time` | `string` | Expiry in RFC3339, parsed from the uploaded chain — plan rotation off this |
| `region` | `string` | Region of a regional certificate; empty for global |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md).

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md).

## Important Notes

- **Nothing renews itself**: unlike a Google-managed certificate, expiry is your responsibility — watch the `expire_time` output.
- **Fully immutable**: every field is ForceNew. Rotation is create-before-destroy — create the replacement under a new name, repoint the proxy's `sslCertificates`, then destroy the old one. GCP blocks deleting a certificate a proxy still references (`resourceInUseByAnotherResource`), so the destroy fails rather than dropping TLS.
- **The private key is the only secret**: it is marked sensitive (Pulumi encrypts it in state; OpenTofu encrypts it with the whole state file when given an encryption key), write-only in GCP, and never appears in outputs. The certificate chain is public handshake material presented to every client and is deliberately not treated as a secret.
- **When to prefer this over a managed certificate**: wildcard domains, EV/OV or private-CA issuance, internal load balancers (no public DNS for managed validation), or serving TLS before DNS cutover.

### Deliberately not modeled (recorded reasons)

- **`private_key_wo` / `private_key_wo_version`** (Terraform write-only argument flow) — the same key material as `privateKey` with Terraform-state-only plumbing; adopting it would raise this module's toolchain floor to Terraform/OpenTofu ≥ 1.11, a divergence from the catalog-wide assumption. Re-evaluate when the catalog declares a ≥ 1.11 floor.
- **`name_prefix`** — a Terraform-side create-before-destroy naming trick; Planton's metadata-driven naming owns resource names, and the rotation pattern it serves is expressed as a new Planton resource with a versioned `certificateName` instead (see the rotation preset).

## Related Components

- [GcpTargetHttpsProxy](/docs/catalog/gcp/gcptargethttpsproxy) — presents this certificate to clients
- [GcpManagedSslCertificate](/docs/catalog/gcp/gcpmanagedsslcertificate) — the Google-managed alternative when hands-off issuance fits
- [GcpSslPolicy](/docs/catalog/gcp/gcpsslpolicy) — hardens TLS versions and ciphers on the same proxy
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project that owns the certificate

## Additional Resources

- [Self-managed SSL certificates overview](https://cloud.google.com/load-balancing/docs/ssl-certificates/self-managed-certs)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
