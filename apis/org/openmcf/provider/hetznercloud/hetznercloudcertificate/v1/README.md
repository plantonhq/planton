# HetznerCloudCertificate

The **HetznerCloudCertificate** resource provisions a TLS certificate in Hetzner Cloud for use by load balancer HTTPS services. The component supports two mutually exclusive certificate types through a proto `oneof`: **uploaded** (you provide the PEM certificate and private key) or **managed** (Hetzner Cloud obtains and renews a Let's Encrypt certificate automatically). Exactly one type must be specified per manifest.

## What It Represents

A [Hetzner Cloud Certificate](https://docs.hetzner.cloud/#certificates) is a TLS certificate stored in the Hetzner Cloud certificate store. Load balancers reference certificates by ID when terminating HTTPS connections. Certificates exist independently of load balancers — a single certificate can be referenced by multiple load balancer HTTPS services, and deleting a load balancer does not delete its referenced certificates.

Hetzner Cloud supports two certificate types:

- **Uploaded**: You provide a PEM-encoded certificate chain and private key. Hetzner Cloud stores them and makes the certificate available for HTTPS listeners. You are responsible for renewal and rotation. Changing the certificate or private key forces replacement of the resource (ForceNew).

- **Managed**: You specify one or more domain names and Hetzner Cloud obtains a Let's Encrypt certificate automatically. Hetzner Cloud handles renewal. The domains must resolve to a Hetzner Cloud load balancer with an HTTPS service referencing this certificate so the ACME HTTP-01 challenge can succeed. Changing the domain list forces replacement (ForceNew).

## Bundled Resources

| Terraform Resource | Count | Created When | Purpose |
|---|---|---|---|
| `hcloud_uploaded_certificate` | 0 or 1 | `uploaded` variant is set | Stores a user-provided PEM certificate chain and private key. |
| `hcloud_managed_certificate` | 0 or 1 | `managed` variant is set | Requests a Let's Encrypt certificate for the specified domains. |

Exactly one of the two resources is created per deployment. The proto `oneof` enforces mutual exclusivity at the schema level.

## Key Features

### Uploaded Certificates

The `uploaded` variant accepts a PEM-encoded certificate chain (`certificate` field) and a PEM-encoded private key (`privateKey` field). Both fields are required and immutable — changing either forces replacement of the certificate resource.

The private key is treated as sensitive material. The Pulumi module wraps it with `pulumi.ToSecret`; the Terraform module marks the variable as `sensitive = true`. The private key never appears in plan output or stack state in cleartext.

Use uploaded certificates for wildcard certificates, certificates from a specific CA (e.g., EV or OV certificates), or certificates obtained through an external ACME client.

### Managed Certificates (Let's Encrypt)

The `managed` variant takes a list of domain names (`domainNames` field, minimum one). Hetzner Cloud issues a single SAN certificate covering all listed domains via Let's Encrypt. Renewal is automatic.

Successful provisioning requires:
1. Each domain must have a DNS A or AAAA record pointing to a Hetzner Cloud load balancer.
2. The load balancer must have an HTTPS service configured that references this certificate.

Without both prerequisites, the ACME HTTP-01 challenge fails and the certificate remains in a pending state.

### Immutability

All substantive fields are immutable (ForceNew). For uploaded certificates, changing the certificate chain or private key destroys the old certificate and creates a new one. For managed certificates, changing the domain list has the same effect. Only the certificate name and labels can be updated in place.

### Automatic Labeling

Standard labels (`resource`, `name`, `kind`, `org`, `env`, `id`) are applied from metadata following the CG01 pattern. User-specified `metadata.labels` are merged in, with standard labels taking precedence on key conflicts.

## Upstream Dependencies (What This Resource Needs)

This component has no upstream dependencies. It does not reference other OpenMCF components via `StringValueOrRef`.

## Downstream Dependents (What References This Resource)

| Dependent | Field | Purpose |
|---|---|---|
| `HetznerCloudLoadBalancer` | HTTPS service certificate references | Load balancer HTTPS services reference `certificate_id` to terminate TLS connections. |

The `certificate_id` output is the primary integration point. Load balancers reference it when configuring HTTPS listeners.

## Stack Outputs

| Output | Description |
|---|---|
| `certificate_id` | The Hetzner Cloud numeric ID of the created certificate (as a string). Referenced by load balancer HTTPS services. |
| `type` | Certificate type: `"uploaded"` or `"managed"`. Computed by Hetzner Cloud. |
| `fingerprint` | SHA256 fingerprint of the certificate. Computed by Hetzner Cloud. |
| `not_valid_before` | Point in time when the certificate becomes valid (ISO-8601). |
| `not_valid_after` | Point in time when the certificate stops being valid (ISO-8601). |

## References

- [Hetzner Cloud Certificates Documentation](https://docs.hetzner.cloud/#certificates)
- [Hetzner Cloud Let's Encrypt Integration](https://docs.hetzner.cloud/#certificates-get-a-certificate)
- [Terraform hcloud_uploaded_certificate Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/uploaded_certificate)
- [Terraform hcloud_managed_certificate Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/managed_certificate)
- [Pulumi hcloud.UploadedCertificate Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/uploadedcertificate/)
- [Pulumi hcloud.ManagedCertificate Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/managedcertificate/)
