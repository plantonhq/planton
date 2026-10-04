# GcpManagedSslCertificate

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpManagedSslCertificateSpec defines a Google-managed SSL certificate — a
global Compute Engine SSL certificate whose private key and issuance are
handled entirely by Google. Attach it to a target HTTPS proxy to terminate
TLS at a global external Application Load Balancer without ever handling key
material yourself.

Provisioning is asynchronous and DNS-gated: creation returns immediately,
but the certificate stays PROVISIONING until each domain's DNS points at the
load balancer's IP (the same forwarding rule the proxy this cert is attached
to serves). Until then the domain serves Google's default certificate.

The whole resource is immutable in GCP: name and domains are ForceNew, so
changing the domain list destroys and recreates the certificate. Because a
cert attached to a proxy cannot be deleted, rotate by creating the
replacement first and swapping the proxy's reference before destroying the
old one (create-before-destroy) — otherwise the destroy fails and TLS drops
during the gap.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpManagedSslCertificate
metadata:
  name: my-sample-managed-ssl-cert
spec:
  # GCP project that owns the certificate.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # Cloud-side name; omit to default to metadata.name.
  certificateName: app-cert

  # What this certificate secures — shown in the GCP console.
  description: TLS for the production app load balancer

  # Domains the certificate is valid for (no wildcards).
  domains:
    - app.example.com

  # What a destroy does: DELETE (default), PREVENT, or ABANDON.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.certificateName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.domains` | `[]string` | yes |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the certificate.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable: changing it destroys and recreates the certificate.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.certificateName

`string`

Name of the certificate in GCP. Must be 1-63 characters: lowercase
letters, digits, and hyphens; must start with a letter and end with a
letter or digit. If not specified, defaults to metadata.name.
Immutable: changing it destroys and recreates the certificate, briefly
breaking every target HTTPS proxy that references the old self_link.

- rule: certificate_name must be RFC1035-compliant: 1-63 lowercase letters, digits, or hyphens; must start with a letter and end with a letter or digit

### spec.description

`string`

What this certificate secures and which proxy uses it — write it for the
operator debugging a TLS provisioning issue later. Immutable.

- rule: {"string":{"maxLen":"2048"}}

### spec.domains

`[]string` · required

The domains the certificate is valid for (1-100). Each must be a
fully-qualified domain name; a leading "*." wildcard is NOT supported by
Google-managed certificates. Every domain must have DNS pointing at the
load balancer before the certificate can finish provisioning. Immutable:
changing the list destroys and recreates the certificate.

- rule: each domain must be a fully-qualified domain name (no wildcards) such as app.example.com
- rule: {"repeated":{"minItems":"1","maxItems":"100"}}

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the certificate is deleted (GCP refuses while any
               proxy still references it, so destroy fails rather
               than dropping TLS)
  "PREVENT" -- destroy FAILS; a guard rail for a certificate whose
               replacement is not yet provisioned and serving
  "ABANDON" -- the certificate is removed from management but left
               in GCP (useful mid-rotation handoff)

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpManagedSslCertificate, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.self_link` | `string` | Self-link URI of the SSL certificate. This is the value a target HTTPS proxy references in its ssl_certificates list — the composition handle that terminates TLS at the load balancer. Format: https://www.googleapis.com/compute/v1/projects/{project}/global/sslCertificates/{name} |
| `status.outputs.certificate_name` | `string` | Name of the SSL certificate as it exists in GCP. |
| `status.outputs.certificate_id` | `string` | Server-assigned numeric ID of the certificate. |
| `status.outputs.expire_time` | `string` | Expiry time of the certificate in RFC3339 format. Empty until the certificate finishes provisioning (Google issues it only after DNS points at the load balancer). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpTargetHttpsProxy | `spec.sslCertificates` | `status.outputs.self_link` |

## See Also

- [Overview](../README.md)
